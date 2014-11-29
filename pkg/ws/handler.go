// Package ws profides utilities to create webservices. It is
// a lightweight framework around the stardard net/http and
// the httprouter libraries. It uses go.net/context to store
// values for the lifetime of a request.
// The design of the framework is based on this excellent article on
// web frameworks in go.
// http://nicolasmerouze.com/build-web-framework-golang/
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
