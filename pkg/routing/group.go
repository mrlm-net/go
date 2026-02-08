package routing

import "net/http"

// Group represents a set of routes with a common prefix and shared middleware.
// Middleware is applied at registration time, not at request time, to preserve
// zero-alloc routing.
type Group struct {
	prefix      string
	router      *Router
	middlewares []Middleware
}

// Group creates a new route group with the given prefix.
func (r *Router) Group(prefix string) *Group {
	return &Group{
		prefix: prefix,
		router: r,
	}
}

// Group creates a nested group with an additional prefix.
func (g *Group) Group(prefix string) *Group {
	return &Group{
		prefix:      g.prefix + prefix,
		router:      g.router,
		middlewares: g.middlewares,
	}
}

// Use adds middleware to the group. Returns the group for chaining.
// Middleware is applied to handlers at registration time.
func (g *Group) Use(mw ...Middleware) *Group {
	g.middlewares = append(g.middlewares, mw...)
	return g
}

// wrapHandler applies group middleware to the handler at registration time.
func (g *Group) wrapHandler(handler Handler) Handler {
	h := handler
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		h = g.middlewares[i](h)
	}
	return h
}

// HandleContext registers a zero-allocation Handler for the given method and pattern.
// The group prefix is prepended to the pattern.
func (g *Group) HandleContext(method, pattern string, handler Handler) {
	g.router.HandleContext(method, g.prefix+pattern, g.wrapHandler(handler))
}

// Handle registers an http.Handler for the given method and pattern.
// The group prefix is prepended to the pattern.
// Note: group middleware is not applied to http.Handler registrations because
// they operate on (http.ResponseWriter, *http.Request), not *Context.
// Use HandleContext with Context-based middleware for consistent behavior.
func (g *Group) Handle(method, pattern string, handler http.Handler) {
	g.router.Handle(method, g.prefix+pattern, handler)
}

// GETContext registers a zero-allocation handler for GET requests.
func (g *Group) GETContext(pattern string, handler Handler) {
	g.HandleContext(http.MethodGet, pattern, handler)
}

// POSTContext registers a zero-allocation handler for POST requests.
func (g *Group) POSTContext(pattern string, handler Handler) {
	g.HandleContext(http.MethodPost, pattern, handler)
}

// PUTContext registers a zero-allocation handler for PUT requests.
func (g *Group) PUTContext(pattern string, handler Handler) {
	g.HandleContext(http.MethodPut, pattern, handler)
}

// DELETEContext registers a zero-allocation handler for DELETE requests.
func (g *Group) DELETEContext(pattern string, handler Handler) {
	g.HandleContext(http.MethodDelete, pattern, handler)
}

// PATCHContext registers a zero-allocation handler for PATCH requests.
func (g *Group) PATCHContext(pattern string, handler Handler) {
	g.HandleContext(http.MethodPatch, pattern, handler)
}
