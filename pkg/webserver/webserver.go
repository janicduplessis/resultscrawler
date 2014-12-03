// Package webserver implements a json api for the client to be able to
// access results from the web.
package webserver

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/sessions"
	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/logger"
	"github.com/janicduplessis/resultscrawler/pkg/store"
	"github.com/janicduplessis/resultscrawler/pkg/ws"
)

type (
	// Config contains parameters to initialize the webserver.
	Config struct {
		UserStore          store.UserStore
		CrawlerConfigStore store.CrawlerConfigStore
		UserResultsStore   store.UserResultsStore
		Crypto             crypto.Crypto
		Logger             logger.Logger
		SessionKey         string
	}

	// Webserver serves as a global context for the server.
	Webserver struct {
		userStore          store.UserStore
		crawlerConfigStore store.CrawlerConfigStore
		userResultsStore   store.UserResultsStore
		crypto             crypto.Crypto
		logger             logger.Logger
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
	sessions := sessions.NewCookieStore([]byte(config.SessionKey))

	webserver := &Webserver{
		config.UserStore,
		config.CrawlerConfigStore,
		config.UserResultsStore,
		config.Crypto,
		config.Logger,
		router,
		sessions,
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

func (server *Webserver) appHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var tmpUser *userModel

	userID, err := server.getSessionUserID(r)
	if err != ErrUnauthorized {
		if err != nil {
			server.serverError(w, err)
			return
		}

		user, err := server.userStore.FindByID(userID)
		if err != nil {
			server.serverError(w, err)
			return
		}

		tmpUser = &userModel{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}
	}

	userJSON, err := json.Marshal(tmpUser)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user-tmp",
		Value:    string(userJSON),
		Expires:  time.Now().Add(1 * time.Minute),
		HttpOnly: false,
	})
	http.ServeFile(w, r, "public/index.html")
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
	user, err := server.userStore.FindByEmail(request.Email)
	if err != nil {
		if err != store.ErrNotFound {
			server.serverError(w, err)
			return
		}

		// If the user is not found returns an invalid login status.
		response := &loginResponse{
			Status: statusInvalidLogin,
			User:   nil,
		}
		err = sendJSON(w, response)
		if err != nil {
			server.serverError(w, err)
		}
		server.logger.Logf("Invalid login attempt. Email: %s, IP: %s", request.Email, r.RemoteAddr)
		return
	}

	// At this point we have a valid email, check the password.
	res, err := server.crypto.CompareHashAndPassword(user.PasswordHash, request.Password)
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

		server.logger.Log("Invalid password.")
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

	server.logger.Logf("Succesful login for user %s", user.Email)
}

func (server *Webserver) registerHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	request := &registerRequest{}
	err := readJSON(r, request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Make sure the email is not already used.
	_, err = server.userStore.FindByEmail(request.Email)
	if err != store.ErrNotFound {
		if err != nil {
			server.serverError(w, err)
			return
		}

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
	passwordHash, err := server.crypto.GenerateFromPassword(request.Password)
	if err != nil {
		server.serverError(w, err)
		return
	}

	user := &store.User{
		Email:        request.Email,
		PasswordHash: passwordHash,
		FirstName:    request.FirstName,
		LastName:     request.LastName,
	}

	// Create the user in the datastore.
	err = server.userStore.Insert(user)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Create the crawler config in the datastore.
	crawlerConfig := &store.CrawlerConfig{
		UserID:            user.ID,
		Status:            false,
		NotificationEmail: request.Email,
	}
	err = server.crawlerConfigStore.Insert(crawlerConfig)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// Create empty results for the user.
	results := &store.UserResults{
		UserID: user.ID,
	}
	server.userResultsStore.Insert(results)
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

	server.logger.Logf("Succesful registration for user %s", user.Email)
}

func (server *Webserver) resultsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	year := params.ByName("year")
	user := getUser(ctx)
	results, err := server.userResultsStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &resultsResponse{
		Year: year,
	}

	for _, c := range results.Classes {
		if c.Year == year {
			newClass := &resultClassModel{
				Name:    c.Name,
				Group:   c.Group,
				Results: make([]*resultModel, len(c.Results)),
			}
			for i, result := range c.Results {
				newClass.Results[i] = &resultModel{
					Name:    result.Name,
					Result:  result.Result,
					Average: result.Average,
				}
			}
			response.Classes = append(response.Classes, newClass)
		}
	}

	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerGetConfigHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := getUser(ctx)
	config, err := server.crawlerConfigStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	userCode, err := server.crypto.AESDecrypt(config.Code)
	if err != nil {
		server.serverError(w, err)
		return
	}
	userNip, err := server.crypto.AESDecrypt(config.Nip)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &crawlerConfigModel{
		Status:            config.Status,
		Code:              string(userCode),
		Nip:               string(userNip),
		NotificationEmail: config.NotificationEmail,
	}
	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerSaveConfigHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	request := &crawlerConfigModel{}
	err := readJSON(r, request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	// TODO: validate config

	userCode, err := server.crypto.AESEncrypt([]byte(request.Code))
	if err != nil {
		server.serverError(w, err)
		return
	}
	userNip, err := server.crypto.AESEncrypt([]byte(request.Nip))
	if err != nil {
		server.serverError(w, err)
		return
	}

	user := getUser(ctx)
	config, err := server.crawlerConfigStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	results, err := server.userResultsStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	config.Code = userCode
	config.Nip = userNip
	config.NotificationEmail = request.NotificationEmail
	config.Status = request.Status

	err = server.crawlerConfigStore.Update(config)
	if err != nil {
		server.serverError(w, err)
	}

	err = server.userResultsStore.Update(results)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerGetClassesHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := getUser(ctx)
	results, err := server.userResultsStore.FindByID(user.ID)
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

	user := getUser(ctx)
	results, err := server.userResultsStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	id := bson.NewObjectId()

	results.Classes = append(results.Classes, store.Class{
		ID:    id,
		Name:  request.Name,
		Group: request.Group,
		Year:  request.Year,
	})

	err = server.userResultsStore.Update(results)
	if err != nil {
		server.serverError(w, err)
		return
	}
	request.ID = id.Hex()
	err = sendJSON(w, &request)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerEditClassHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	classID := bson.ObjectId(params.ByName("classId"))

	request := crawlerConfigClassModel{}
	err := readJSON(r, &request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	user := getUser(ctx)
	results, err := server.userResultsStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	for i, c := range results.Classes {
		if c.ID == classID {
			results.Classes[i] = store.Class{
				ID:    c.ID,
				Name:  request.Name,
				Group: request.Group,
				Year:  request.Year,
			}
			break
		}
	}

	err = server.userResultsStore.Update(results)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) crawlerDeleteClassHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	idHex := params.ByName("classId")
	if !bson.IsObjectIdHex(idHex) {
		server.serverError(w, store.ErrInvalidID)
		return
	}
	classID := bson.ObjectIdHex(idHex)

	user := getUser(ctx)
	results, err := server.userResultsStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	log.Printf("id: %s", classID.Hex())
	log.Printf("before: %v", results.Classes)

	for i, c := range results.Classes {
		if c.ID == classID {
			results.Classes = append(results.Classes[:i], results.Classes[i+1:]...)
			break
		}
	}

	log.Printf("after: %v", results.Classes)

	err = server.userResultsStore.Update(results)
	if err != nil {
		server.serverError(w, err)
	}
}

// Middlewares
func (server *Webserver) authMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		userID, err := server.getSessionUserID(r)
		if err != nil {
			server.logger.Error(err)
			server.authError(w)
			return
		}

		user, err := server.userStore.FindByID(userID)
		if err != nil {
			server.logger.Error(err)
			server.authError(w)
			return
		}

		ctx = context.WithValue(ctx, userKey, user)

		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

func (server *Webserver) errorMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

func (server *Webserver) logMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		server.logger.Logf("Request start %s", r.URL.String())
		start := time.Now()
		defer func() {
			elapsed := time.Since(start)
			server.logger.Logf("Request end %s. Took %.f ms.", r.URL.String(), elapsed.Seconds()*1000)
		}()
		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

// Session helpers
func (server *Webserver) getSessionUserID(r *http.Request) (bson.ObjectId, error) {
	s, err := server.sessions.Get(r, sessionName)
	if err != nil {
		return bson.ObjectId(""), err
	}

	if s.Values[sessionUserIDKey] == nil {
		return bson.ObjectId(""), ErrUnauthorized
	}

	userID, ok := s.Values[sessionUserIDKey].(bson.ObjectId)
	if !ok {
		return bson.ObjectId(""), ErrUnauthorized
	}
	return userID, nil
}

func (server *Webserver) createSession(w http.ResponseWriter, r *http.Request, userID bson.ObjectId) error {
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
	server.logger.Log("Unauthorized request attempt")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (server *Webserver) serverError(w http.ResponseWriter, err error) {
	server.logger.Error(err)
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

func getUser(ctx context.Context) *store.User {
	user, ok := ctx.Value(userKey).(*store.User)
	if !ok {
		panic("No user in context. Make sure the handler is authentified")
	}
	return user
}

// Model helpers
func getClassesModel(classes []store.Class) []*crawlerConfigClassModel {
	result := make([]*crawlerConfigClassModel, len(classes))
	for i, c := range classes {
		result[i] = &crawlerConfigClassModel{
			ID:    c.ID.Hex(),
			Name:  c.Name,
			Group: c.Group,
			Year:  c.Year,
		}
	}
	return result
}

func init() {
	gob.Register(bson.ObjectId(""))
}
