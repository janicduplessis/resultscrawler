package ws

import (
	"net/http"

	"code.google.com/p/go.net/context"

	"github.com/julienschmidt/httprouter"
)

type (
	key int

	// Router handles routes for the application.
	// It is a simple wrapper around the httprouter lib.
	Router struct {
		router *httprouter.Router
	}
)

const paramsKey key = 1

// NewRouter creates a new router object.
func NewRouter() *Router {
	return &Router{
		httprouter.New(),
	}
}

// GET registers a GET route handler.
func (r *Router) GET(path string, handle Handler) {
	r.router.GET(path, wrapHandler(handle))
}

// POST registers a POST route handler.
func (r *Router) POST(path string, handle Handler) {
	r.router.POST(path, wrapHandler(handle))
}

// PUT registers a PUT route handler.
func (r *Router) PUT(path string, handle Handler) {
	r.router.PUT(path, wrapHandler(handle))
}

// DELETE registers a DELETE route handler.
func (r *Router) DELETE(path string, handle Handler) {
	r.router.DELETE(path, wrapHandler(handle))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// ServeFiles serves files from the filesystem.
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	r.router.ServeFiles(path, root)
}

// Params returns the route parameters from the context
func Params(ctx context.Context) httprouter.Params {
	p, ok := ctx.Value(paramsKey).(httprouter.Params)
	if !ok {
		panic("No params object for request. This should never happen.")
	}
	return p
}

func wrapHandler(handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(context.Background(), paramsKey, ps)
		handler.ServeHTTP(ctx, w, r)
	}
}
