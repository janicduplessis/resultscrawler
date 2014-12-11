package ws

import (
	"net/http"

	"code.google.com/p/go.net/context"
)

type (
	// Handler is an interface for processing an http request.
	// Based on the standard http lib but with an additional context object.
	Handler interface {
		ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
	}

	// HandlerFunc defines a handler function for a request.
	HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)
)

// ServeHTTP allows the HandlerFunc to be used as an handler.
func (f HandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}
