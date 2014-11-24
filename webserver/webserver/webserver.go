// Package webserver implements a json api for the client to be able to
// access results from the web.
package webserver

import (
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"labix.org/v2/mgo/bson"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/sessions"

	"github.com/janicduplessis/resultscrawler/lib"
	"github.com/janicduplessis/resultscrawler/lib/ws"
)

type (
	// Config contains parameters to initialize the webserver.
	Config struct {
		UserStore  lib.UserStore
		Crypto     lib.Crypto
		Logger     lib.Logger
		SessionKey string
	}

	// Webserver serves as a global context for the server.
	Webserver struct {
		userStore lib.UserStore
		crypto    lib.Crypto
		logger    lib.Logger
		router    *ws.Router
		sessions  *sessions.CookieStore
	}

	key int
)

const (
	userKey          key = 1
	sessionUserIDKey     = "userid"
	sessionName          = "rc-session"
)

// NewWebserver creates a new webserver object.
func NewWebserver(config *Config) *Webserver {
	router := ws.NewRouter()
	sessions := sessions.NewCookieStore([]byte(config.SessionKey))

	webserver := &Webserver{
		config.UserStore,
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

func (server *Webserver) homeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
}

func (server *Webserver) loginHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//user, err := server.userStore.FindByID()
}

func (server *Webserver) registerHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func (server *Webserver) resultsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := ws.Params(ctx)
	year := params.ByName("year")
	class := params.ByName("class")
	user := getUser(ctx)
	server.logger.Logf("Getting classes for user %s, year %s and class %s", user.UserName, year, class)
}

func (server *Webserver) authMiddleware(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		s, err := server.sessions.Get(r, sessionName)
		if err != nil {
			server.logger.Error(err)
			server.authError(w)
			return
		}

		if s.Values[sessionUserIDKey] == nil {
			server.authError(w)
			return
		}

		userID, ok := s.Values[sessionUserIDKey].(bson.ObjectId)
		if !ok {
			server.logger.Error(err)
			server.authError(w)
			return
		}

		server.logger.Logf("%v", userID)

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
		next.ServeHTTP(ctx, w, r)
		elapsed := time.Since(start)
		server.logger.Logf("Request end %s. Took %vms.", r.URL.String(), elapsed.Seconds()*1000)
	}

	return ws.HandlerFunc(fn)
}

func (server *Webserver) authError(w http.ResponseWriter) {
	server.logger.Log("Unauthorized request attempt")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
