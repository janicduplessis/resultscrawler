// Package webserver implements a json api for the client to be able to
// access results from the web.
package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/dgrijalva/jwt-go"
	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/crawlerconfig"
	"github.com/janicduplessis/resultscrawler/pkg/store/results"
	"github.com/janicduplessis/resultscrawler/pkg/store/user"
	"github.com/janicduplessis/resultscrawler/pkg/ws"
)

type (
	// Config contains parameters to initialize the webserver.
	Config struct {
		UserStore          user.Store
		CrawlerConfigStore crawlerconfig.Store
		UserResultsStore   results.Store
		RSAPublic          []byte
		RSAPrivate         []byte
		CrawlerClient      api.Crawler
	}

	// Webserver serves as a global context for the server.
	Webserver struct {
		userStore          user.Store
		crawlerConfigStore crawlerconfig.Store
		userResultsStore   results.Store
		rsaPublic          []byte
		rsaPrivate         []byte
		router             *ws.Router
		crawlerClient      api.Crawler
		httpPort           string
		httpsPort          string
	}

	key int
)

const (
	urlBase           = "/api/v1"
	urlResults        = urlBase + "/results"
	urlCrawlerConfig  = urlBase + "/crawler/config"
	urlCrawlerClass   = urlBase + "/crawler/class"
	urlCrawlerRefresh = urlBase + "/crawler/refresh"
	urlLogin          = urlBase + "/auth/login"
	urlRegister       = urlBase + "/auth/register"
	urlLogout         = urlBase + "/auth/logout"

	userKey          key = 1
	sessionUserIDKey     = "userid"
	headerName           = "X-Access-Token"
)

const (
	// Status for register and login.
	statusOK           = iota // Everything is ok.
	statusInvalidLogin        // Invalid username or password.
	statusTooMany             // Too many invalid logins attempts.
	statusInvalidEmail        // The requested email is already used.
	statusInvalidInfos        // The registration infos are invalid.
)

// ErrUnauthorized happens when an unauthorized access occur.
var ErrUnauthorized = errors.New("Unauthorized access")

// NewWebserver creates a new webserver object.
func NewWebserver(config *Config) *Webserver {
	router := ws.NewRouter()

	webserver := &Webserver{
		userStore:          config.UserStore,
		crawlerConfigStore: config.CrawlerConfigStore,
		userResultsStore:   config.UserResultsStore,
		rsaPublic:          config.RSAPublic,
		rsaPrivate:         config.RSAPrivate,
		router:             router,
		crawlerClient:      config.CrawlerClient,
	}

	// Define middleware groups
	commonHandlers := ws.NewMiddlewareGroup(webserver.errorMiddleware, webserver.logMiddleware)
	registeredHandlers := commonHandlers.Append(webserver.authMiddleware)

	// Static files
	router.ServeFiles("/app/*filepath", http.Dir("public"))

	// Register routes
	router.GET("/", commonHandlers.Then(webserver.homeHandler))
	router.GET(urlResults+"/:year", registeredHandlers.Then(webserver.resultsHandler))

	router.GET(urlCrawlerConfig, registeredHandlers.Then(webserver.crawlerGetConfigHandler))
	router.POST(urlCrawlerConfig, registeredHandlers.Then(webserver.crawlerSaveConfigHandler))

	router.GET(urlCrawlerClass, registeredHandlers.Then(webserver.crawlerGetClassesHandler))
	router.POST(urlCrawlerClass, registeredHandlers.Then(webserver.crawlerCreateClassHandler))
	router.PUT(urlCrawlerClass+"/:classId", registeredHandlers.Then(webserver.crawlerEditClassHandler))
	router.DELETE(urlCrawlerClass+"/:classId", registeredHandlers.Then(webserver.crawlerDeleteClassHandler))

	router.POST(urlCrawlerRefresh, registeredHandlers.Then(webserver.crawlerRefreshHandler))

	router.POST(urlLogin, commonHandlers.Then(webserver.loginHandler))
	router.POST(urlRegister, commonHandlers.Then(webserver.registerHandler))
	router.POST(urlLogout, registeredHandlers.Then(webserver.logoutHandler))

	return webserver
}

// Start starts the server at address. Panics if there is an error.
func (server *Webserver) Start(httpPort, httpsPort, cert, key string) {
	server.httpPort = httpPort
	server.httpsPort = httpsPort
	if len(cert) > 0 && len(key) > 0 {
		log.Println("Server running using https on port " + httpsPort + ", " + httpPort)
		// If we are using https, we will serve on the http port too and redirect to https.
		go func() {
			log.Panic(http.ListenAndServeTLS(":"+httpsPort, cert, key, server.router))
		}()
		log.Panic(http.ListenAndServe(":"+httpPort, http.HandlerFunc(server.redirectHTTPS)))
	} else {
		log.Panic(http.ListenAndServe(":"+httpPort, server.router))
	}
}

func (server *Webserver) redirectHTTPS(w http.ResponseWriter, r *http.Request) {
	// Si on n'est pas sur le port 80 ou 443 on va l'inclure explicitement.
	portStr := ""
	// TODO: figure out why it doesnt work for prod servers
	if strings.HasSuffix(r.Host, ":"+server.httpPort) {
		portStr = ":" + server.httpsPort
	}
	redirectURL := fmt.Sprintf("https://%s%s%s", strings.TrimSuffix(r.Host, ":"+server.httpPort), portStr, r.RequestURI)
	log.Printf("Redirecting to https url: %s from %s", redirectURL, r.Host+r.RequestURI)
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

// Handlers
func (server *Webserver) homeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
}

func (server *Webserver) loginHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Watch out this thing has many exit points!
	// Went with the return on error technique here. So if
	// you get to the bottom of this function you are logged in.
	request := &loginRequest{}
	err := readJSON(r, request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	//TODO: prevent login spam by ip.

	// Check if the user exists.
	user, passHash, err := server.userStore.GetUserForLogin(request.Email)
	if err != nil {
		server.serverError(w, err)
		return
	}
	if user == nil {
		// If the user is not found returns an invalid login status.
		response := &loginResponse{
			Status: statusInvalidLogin,
			User:   nil,
		}
		err = sendJSON(w, response)
		if err != nil {
			server.serverError(w, err)
		}
		log.Printf("Invalid login attempt. Email: %s, IP: %s", request.Email, r.RemoteAddr)
		return
	}

	// At this point we have a valid email, check the password.
	res, err := crypto.CompareHashAndPassword(passHash, request.Password)
	if err != nil {
		server.serverError(w, err)
		return
	}
	if !res {
		// Bad password :( that was close. Returns an invalid login status.
		response := &loginResponse{
			Status: statusInvalidLogin,
			User:   nil,
		}
		err = sendJSON(w, response)
		if err != nil {
			server.serverError(w, err)
		}

		log.Println("Invalid password.")
		return
	}

	// Good password, start the session and returns user info.
	token, err := server.createSession(w, r, user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &loginResponse{
		Status: statusOK,
		Token:  token,
		User: &userModel{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}

	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}

	log.Printf("Succesful login for user %s", user.Email)
}

func (server *Webserver) registerHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	request := &registerRequest{}
	err := readJSON(r, request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Make sure the email is not already used.
	user, _, err := server.userStore.GetUserForLogin(request.Email)
	if err != nil {
		server.serverError(w, err)
		return
	}
	if user != nil {
		// Send an invalid email response.
		response := &registerResponse{
			Status: statusInvalidEmail,
		}

		err = sendJSON(w, response)
		if err != nil {
			server.serverError(w, err)
		}
		return
	}

	//TODO: validate fields

	// Here all the registration infos are good, create the user.
	// Hash the password for storage.
	passwordHash, err := crypto.GenerateFromPassword(request.Password)
	if err != nil {
		server.serverError(w, err)
		return
	}

	user = &api.User{
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	}

	// Create the user in the datastore.
	err = server.userStore.CreateUser(user, passwordHash)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Once registration is succesful create a session.
	token, err := server.createSession(w, r, user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Returns a status ok response with info about the user.
	response := &registerResponse{
		Status: statusOK,
		Token:  token,
		User: &userModel{
			user.Email,
			user.FirstName,
			user.LastName,
		},
	}
	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}

	log.Printf("Succesful registration for user %s", user.Email)
}

func (server *Webserver) logoutHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//userID := getUserID(ctx)

}

func (server *Webserver) resultsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	year := params.ByName("year")
	userID := getUserID(ctx)
	results, err := server.userResultsStore.GetResults(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &resultsResponse{
		Year:       year,
		LastUpdate: results.LastUpdate,
	}

	for _, c := range results.Classes {
		if c.Year == year {
			response.Classes = append(response.Classes, c)
		}
	}

	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerGetConfigHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userID := getUserID(ctx)
	config, err := server.crawlerConfigStore.GetCrawlerConfig(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	err = sendJSON(w, config)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerSaveConfigHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	request := &api.CrawlerConfig{}
	err := readJSON(r, request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// TODO: validate config

	userID := getUserID(ctx)
	config, err := server.crawlerConfigStore.GetCrawlerConfig(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	config.Code = request.Code
	config.Nip = request.Nip
	config.NotificationEmail = request.NotificationEmail
	config.Status = request.Status

	err = server.crawlerConfigStore.UpdateCrawlerConfig(config)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerGetClassesHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userID := getUserID(ctx)
	results, err := server.userResultsStore.GetResults(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := getClassesModel(results.Classes)
	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerCreateClassHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	request := crawlerConfigClassModel{}
	err := readJSON(r, &request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	userID := getUserID(ctx)
	results, err := server.userResultsStore.GetResults(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}
	log.Printf("%+v", results)

	classID := bson.NewObjectId().Hex()

	results.Classes = append(results.Classes, api.Class{
		ID:    classID,
		Name:  request.Name,
		Group: request.Group,
		Year:  request.Year,
	})

	err = server.userResultsStore.UpdateResults(results)
	if err != nil {
		server.serverError(w, err)
		return
	}
	request.ID = classID
	err = sendJSON(w, &request)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerEditClassHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	classID := params.ByName("classId")

	request := crawlerConfigClassModel{}
	err := readJSON(r, &request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	userID := getUserID(ctx)
	results, err := server.userResultsStore.GetResults(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	for i, c := range results.Classes {
		if c.ID == classID {
			results.Classes[i] = api.Class{
				ID:    c.ID,
				Name:  request.Name,
				Group: request.Group,
				Year:  request.Year,
			}
			break
		}
	}

	err = server.userResultsStore.UpdateResults(results)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerDeleteClassHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	classID := params.ByName("classId")

	userID := getUserID(ctx)
	results, err := server.userResultsStore.GetResults(userID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	log.Printf("before: %v", results.Classes)

	for i, c := range results.Classes {
		if c.ID == classID {
			results.Classes = append(results.Classes[:i], results.Classes[i+1:]...)
			break
		}
	}

	log.Printf("after: %v", results.Classes)

	err = server.userResultsStore.UpdateResults(results)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerRefreshHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userID := getUserID(ctx)
	if err := server.crawlerClient.Refresh(userID); err != nil {
		server.serverError(w, err)
	}
}

// Middlewares
func (server *Webserver) authMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		userID, err := server.getSessionUserID(r)
		if err != nil {
			log.Println(err)
			server.authError(w)
			return
		}

		ctx = context.WithValue(ctx, userKey, userID)

		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

func (server *Webserver) errorMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v\n%s", err, debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

func (server *Webserver) logMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		log.Printf("Request start %s", r.URL.String())
		start := time.Now()
		defer func() {
			elapsed := time.Since(start)
			log.Printf("Request end %s. Took %.f ms.", r.URL.String(), elapsed.Seconds()*1000)
		}()
		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

// Session helpers
func (server *Webserver) getSessionUserID(r *http.Request) (string, error) {
	tokenHeader := r.Header.Get(headerName)
	// validate the token
	token, err := jwt.Parse(tokenHeader, func(token *jwt.Token) (interface{}, error) {
		return server.rsaPublic, nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrUnauthorized
	}

	userID, ok := token.Claims[sessionUserIDKey].(string)
	if !ok || len(userID) == 0 {
		return "", ErrUnauthorized
	}

	return userID, nil
}

func (server *Webserver) createSession(w http.ResponseWriter, r *http.Request, userID string) (string, error) {
	// Create the token
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	// Set some claims
	token.Claims[sessionUserIDKey] = userID
	// TODO: expire session?
	// Sign and get the complete encoded token as a string
	return token.SignedString(server.rsaPrivate)
}

func (server *Webserver) endSession(w http.ResponseWriter, r *http.Request) error {
	// Nothing to do here anymore.
	return nil
}

// Error helpers
func (server *Webserver) authError(w http.ResponseWriter) {
	log.Println("Unauthorized request attempt")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (server *Webserver) serverError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// JSON helpers
func readJSON(r *http.Request, obj interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return err
	}
	err = r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, obj)
}

func sendJSON(w http.ResponseWriter, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)

	return err
}

func getUserID(ctx context.Context) string {
	userID, ok := ctx.Value(userKey).(string)
	if !ok {
		panic("No user in context. Make sure the handler is authentified")
	}
	return userID
}

// Model helpers
func getClassesModel(classes []api.Class) []*crawlerConfigClassModel {
	result := make([]*crawlerConfigClassModel, len(classes))
	for i, c := range classes {
		result[i] = &crawlerConfigClassModel{
			ID:    c.ID,
			Name:  c.Name,
			Group: c.Group,
			Year:  c.Year,
		}
	}
	return result
}
