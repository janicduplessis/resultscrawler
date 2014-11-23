package ws

type (
	// Middleware is a function that implements additional functionnality
	// for a handler.
	Middleware func(Handler) Handler

	// MiddlewareGroup allows simple chaining of middlewares.
	// Based on alice (https://github.com/justinas/alice).
	MiddlewareGroup struct {
		handlers []Middleware
	}
)

// NewMiddlewareGroup creates a new middleware group.
func NewMiddlewareGroup(handlers ...Middleware) MiddlewareGroup {
	group := MiddlewareGroup{}
	group.handlers = append(group.handlers, handlers...)

	return group
}

// Append creates a new middleware group from an existing group.
func (g MiddlewareGroup) Append(handlers ...Middleware) MiddlewareGroup {
	newGroup := make([]Middleware, len(g.handlers))
	copy(newGroup, g.handlers)
	newGroup = append(newGroup, handlers...)
	return g
}

// Then returns a handler that executes the middleware chain then calls handler.
func (g MiddlewareGroup) Then(handlerFunc HandlerFunc) Handler {
	final := Handler(handlerFunc)
	for i := len(g.handlers) - 1; i >= 0; i-- {
		final = g.handlers[i](final)
	}

	return final
}
