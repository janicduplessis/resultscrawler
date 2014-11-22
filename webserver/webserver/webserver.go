package webserver

import (
	"log"
	"net/http"

	"code.google.com/p/go.net/context"

	"github.com/janicduplessis/resultscrawler/lib"
	"github.com/janicduplessis/resultscrawler/lib/ws"
)

// Webserver serves as a global context for the server.
type Webserver struct {
	store  lib.UserStore
	crypto lib.Crypto
	logger lib.Logger
	router *ws.Router
}

type key int

var userKey key = 1

// NewWebserver creates a new webserver object.
func NewWebserver(userStore lib.UserStore, crypto lib.Crypto, logger lib.Logger) *Webserver {
	router := ws.NewRouter()

	webserver := &Webserver{
		userStore,
		crypto,
		logger,
		router,
	}

	// Define middleware groups
	commonHandlers := ws.NewMiddlewareGroup(webserver.errorHandler)
	registeredHandlers := commonHandlers.Append(webserver.authHandler)

	router.GET("/home", commonHandlers.Then(webserver.homeHandler))
	router.GET("/results", registeredHandlers.Then(webserver.homeHandler))

	return webserver
}

// Start starts the server at address.
func (server *Webserver) Start(address string) error {
	return http.ListenAndServe(address, server.router)
}

func (server *Webserver) authHandler(next ws.Handler) ws.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, userKey, lib.User{})

		next.ServeHTTP(ctx, w, r)
	}

	return ws.HandlerFunc(fn)
}

func (server *Webserver) errorHandler(next ws.Handler) ws.Handler {
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

func (server *Webserver) homeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ws.Params(ctx)
}

func (server *Webserver) resultsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}
