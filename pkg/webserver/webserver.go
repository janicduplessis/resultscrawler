// Package webserver implements a json api for the client to be able to
// access results from the web.
package webserver

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/sessions"
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
		SessionKey         string
	}

	// Webserver serves as a global context for the server.
	Webserver struct {
		userStore          user.Store
		crawlerConfigStore crawlerconfig.Store
		userResultsStore   results.Store
		router             *ws.Router
		sessions           *sessions.CookieStore
	}

	key int
)

const (
	userKey          key = 1
	sessionUserIDKey     = "userid"
	sessionName          = "rc-session"

	// Status for register and login.
	statusOK           = 0 // Everything is ok.
	statusInvalidLogin = 1 // Invalid username or password.
	statusTooMany      = 2 // Too many invalid logins attempts.
	statusInvalidEmail = 3 // The requested email is already used.
	statusInvalidInfos = 4 // The registration infos are invalid.
)

// ErrUnauthorized happens when an unauthorized access occur.
var ErrUnauthorized = errors.New("Unauthorized access")

// NewWebserver creates a new webserver object.
func NewWebserver(config *Config) *Webserver {
	router := ws.NewRouter()
	cs := sessions.NewCookieStore([]byte(config.SessionKey))
	cs.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
	}

	webserver := &Webserver{
		config.UserStore,
		config.CrawlerConfigStore,
		config.UserResultsStore,
		router,
		cs,
	}

	// Define middleware groups
	commonHandlers := ws.NewMiddlewareGroup(webserver.errorMiddleware, webserver.logMiddleware)
	registeredHandlers := commonHandlers.Append(webserver.authMiddleware)

	// Static files
	router.ServeFiles("/app/*filepath", http.Dir("public"))

	// Register routes
	router.GET("/", commonHandlers.Then(webserver.homeHandler))
	router.GET("/api/v1/results/:year", registeredHandlers.Then(webserver.resultsHandler))

	router.GET("/api/v1/crawler/config", registeredHandlers.Then(webserver.crawlerGetConfigHandler))
	router.POST("/api/v1/crawler/config", registeredHandlers.Then(webserver.crawlerSaveConfigHandler))

	router.GET("/api/v1/crawler/class", registeredHandlers.Then(webserver.crawlerGetClassesHandler))
	router.POST("/api/v1/crawler/class", registeredHandlers.Then(webserver.crawlerCreateClassHandler))
	router.PUT("/api/v1/crawler/class/:classId", registeredHandlers.Then(webserver.crawlerEditClassHandler))
	router.DELETE("/api/v1/crawler/class/:classId", registeredHandlers.Then(webserver.crawlerDeleteClassHandler))

	router.POST("/api/v1/auth/login", commonHandlers.Then(webserver.loginHandler))
	router.POST("/api/v1/auth/register", commonHandlers.Then(webserver.registerHandler))

	return webserver
}

// Start starts the server at address.
func (server *Webserver) Start(address string) error {
	return http.ListenAndServe(address, server.router)
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
	err = server.createSession(w, r, user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &loginResponse{
		Status: statusOK,
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
	err = server.createSession(w, r, user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Returns a status ok response with info about the user.
	response := &registerResponse{
		Status: statusOK,
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
		Year: year,
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
	s, err := server.sessions.Get(r, sessionName)
	if err != nil {
		return "", err
	}

	if s.Values[sessionUserIDKey] == nil {
		return "", ErrUnauthorized
	}

	userID, ok := s.Values[sessionUserIDKey].(string)
	if !ok {
		return "", ErrUnauthorized
	}
	return userID, nil
}

func (server *Webserver) createSession(w http.ResponseWriter, r *http.Request, userID string) error {
	session, err := server.sessions.Get(r, sessionName)
	if err != nil {
		return err
	}

	session.Values[sessionUserIDKey] = userID

	return session.Save(r, w)
}

func (server *Webserver) endSession(w http.ResponseWriter, r *http.Request) error {
	session, err := server.sessions.Get(r, sessionName)
	if err != nil {
		return err
	}

	delete(session.Values, sessionUserIDKey)
	return session.Save(r, w)
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
