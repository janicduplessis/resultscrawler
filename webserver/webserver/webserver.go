// Package webserver implements a json api for the client to be able to
// access results from the web.
package webserver

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/sessions"
	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/lib"
	"github.com/janicduplessis/resultscrawler/lib/ws"
)

type (
	// Config contains parameters to initialize the webserver.
	Config struct {
		UserStore        lib.UserStore
		UserInfoStore    lib.UserInfoStore
		UserResultsStore lib.UserResultsStore
		Crypto           lib.Crypto
		Logger           lib.Logger
		SessionKey       string
	}

	// Webserver serves as a global context for the server.
	Webserver struct {
		userStore        lib.UserStore
		userInfoStore    lib.UserInfoStore
		userResultsStore lib.UserResultsStore
		crypto           lib.Crypto
		logger           lib.Logger
		router           *ws.Router
		sessions         *sessions.CookieStore
	}

	key int
)

const (
	userKey          key = 1
	sessionUserIDKey     = "userid"
	sessionName          = "rc-session"

	// Status for register and login.
	statusOK              = 0 // Everything is ok.
	statusInvalidLogin    = 1 // Invalid username or password.
	statusTooMany         = 2 // Too many invalid logins attempts.
	statusInvalidUserName = 3 // The requested username already exists.
	statusInvalidInfos    = 4 // The registration infos are invalid.
)

// ErrUnauthorized happens when an unauthorized access occur.
var ErrUnauthorized = errors.New("Unauthorized access")

// NewWebserver creates a new webserver object.
func NewWebserver(config *Config) *Webserver {
	router := ws.NewRouter()
	sessions := sessions.NewCookieStore([]byte(config.SessionKey))

	webserver := &Webserver{
		config.UserStore,
		config.UserInfoStore,
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
	router.GET("/api/v1/results/:year/:class", registeredHandlers.Then(webserver.resultsHandler))

	router.POST("/api/v1/login", commonHandlers.Then(webserver.loginHandler))
	router.POST("/api/v1/register", commonHandlers.Then(webserver.registerHandler))

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

	// Check if the user exists.
	user, err := server.userStore.FindByEmail(request.Email)
	if err != nil {
		if err != lib.ErrNotFound {
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
		return
	}

	// Good password, start the session and returns user info.
	info, err := server.userInfoStore.FindByID(user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	err = server.createSession(w, r, user.ID)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &loginResponse{
		Status: statusOK,
		User: &userModel{
			Email:     user.Email,
			FirstName: info.FirstName,
			LastName:  info.LastName,
		},
	}

	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) registerHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	request := &registerRequest{}
	err := readJSON(r, request)
	if err != nil {
		server.serverError(w, err)
		return
	}

	//TODO: validate fields

	passwordHash, err := server.crypto.GenerateFromPassword(request.Password)
	if err != nil {
		server.serverError(w, err)
		return
	}

	user := &lib.User{
		Email:        request.Email,
		PasswordHash: passwordHash,
	}

	err = server.userStore.Insert(user)
	if err != nil {
		server.serverError(w, err)
		return
	}

	userInfo := &lib.UserInfo{
		UserID:    user.ID,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		CrawlerOn: false,
	}

	err = server.userInfoStore.Insert(userInfo)
	if err != nil {
		server.serverError(w, err)
		return
	}

	results := &lib.UserResults{
		UserID: user.ID,
	}

	server.userResultsStore.Insert(results)
	if err != nil {
		server.serverError(w, err)
		return
	}

	response := &registerResponse{
		Status: statusOK,
		User: &userModel{
			user.Email,
			userInfo.FirstName,
			userInfo.LastName,
		},
	}

	err = sendJSON(w, response)
	if err != nil {
		server.serverError(w, err)
	}
}

func (server *Webserver) resultsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	year := params.ByName("year")
	class := params.ByName("class")
	user := getUser(ctx)
	server.logger.Logf("Getting classes for user %s, year %s and class %s", user.Email, year, class)
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
			server.logger.Logf("Request end %s. Took %vms.", r.URL.String(), elapsed.Seconds()*1000)
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
	body, err := ioutil.ReadAll(r.Body)
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

func getUser(ctx context.Context) *lib.User {
	user, ok := ctx.Value(userKey).(*lib.User)
	if !ok {
		panic("No user in context. Make sure the handler is authentified")
	}
	return user
}

func init() {
	gob.Register(bson.ObjectId(""))
}
