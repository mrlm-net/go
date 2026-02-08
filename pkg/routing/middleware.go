package routing

import "sync"

// Middleware wraps a Handler to add pre/post processing.
// Middleware must not allocate per-request to maintain zero-alloc routing.
type Middleware func(Handler) Handler

// Use adds global middleware to the router.
// Middleware is applied in the order added: the first middleware added
// is the outermost wrapper (executes first on the way in, last on the way out).
func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
	r.chain = nil              // Invalidate pre-built chain
	r.chainOnce = sync.Once{} // Reset so chain rebuilds on next request
}

// Build pre-builds the middleware chain so the first request
// does not pay the build cost. Call after all routes and
// middleware have been registered.
func (r *Router) Build() {
	if len(r.middlewares) == 0 {
		return
	}
	r.chainOnce.Do(func() {
		r.chain = r.buildChain(r.dispatch)
	})
}

// buildChain wraps the given handler with all registered middleware.
// The first middleware in the slice is the outermost wrapper.
func (r *Router) buildChain(handler Handler) Handler {
	h := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}
