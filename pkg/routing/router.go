package routing

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// Router is an HTTP request router that uses a radix trie for efficient
// path matching and supports path parameters and wildcards.
type Router struct {
	trees                 map[string]*radixNode
	maxParams             int // track max params across all routes for pre-allocation
	paramsPool            sync.Pool
	contextPool           sync.Pool
	middlewares           []Middleware
	chain                 Handler      // pre-built middleware chain, nil until Build/first request
	chainOnce             sync.Once    // ensures chain is built exactly once
	methodNotAllowed      http.Handler // custom 405 handler, nil uses default
	HandleOPTIONS         bool         // auto-respond to OPTIONS with Allow header (default true)
	HandleHEAD            bool         // auto-serve HEAD from GET handler without body (default true)
	RedirectTrailingSlash bool         // redirect /path/ to /path or vice versa (default false)
}

// New creates a new Router instance.
func New() *Router {
	return &Router{
		trees:         make(map[string]*radixNode),
		HandleOPTIONS: true,
		HandleHEAD:    true,
	}
}

// HandleContext registers a zero-allocation Handler for the given HTTP method and path pattern.
// This is the preferred method for high-performance routing with zero allocations.
// Path patterns can include:
//   - Parameters: /users/:id matches /users/123, extracts id=123
//   - Wildcards: /files/*filepath matches /files/a/b/c, extracts filepath=a/b/c
//
// Panics if the pattern conflicts with an existing route.
func (r *Router) HandleContext(method, pattern string, handler Handler) {
	if handler == nil {
		panic("routing: nil handler")
	}
	if pattern == "" {
		panic("routing: empty pattern")
	}

	// Ensure pattern starts with /
	if pattern[0] != '/' {
		panic(fmt.Sprintf("routing: pattern must start with /: %s", pattern))
	}

	// Get or create tree for this method
	tree, ok := r.trees[method]
	if !ok {
		tree = &radixNode{}
		r.trees[method] = tree
	}

	// Insert the route with Handler (radix tree handles conflicts)
	tree.addRoute(pattern, handler)

	// Update maxParams for pre-allocation optimization
	paramCount := int(countParams(pattern))
	if paramCount > r.maxParams {
		r.maxParams = paramCount
		// Reinitialize context pool with new maxParams
		r.initContextPool()
	} else if r.contextPool.New == nil {
		r.initContextPool()
	}
}

// Handle registers a handler for the given HTTP method and path pattern.
// Path patterns can include:
//   - Parameters: /users/:id matches /users/123, extracts id=123
//   - Wildcards: /files/*filepath matches /files/a/b/c, extracts filepath=a/b/c
//
// Panics if the pattern conflicts with an existing route.
func (r *Router) Handle(method, pattern string, handler http.Handler) {
	if handler == nil {
		panic("routing: nil handler")
	}
	if pattern == "" {
		panic("routing: empty pattern")
	}

	// Ensure pattern starts with /
	if pattern[0] != '/' {
		panic(fmt.Sprintf("routing: pattern must start with /: %s", pattern))
	}

	// Get or create tree for this method
	tree, ok := r.trees[method]
	if !ok {
		tree = &radixNode{}
		r.trees[method] = tree
	}

	// For http.Handler, wrap in a Handler that injects params into request context
	wrappedHandler := func(ctx *Context) {
		// Inject params into request context if any exist
		if ctx.paramCount > 0 {
			reqCtx := contextWithParams(ctx.Request.Context(), ctx.Params[:ctx.paramCount])
			ctx.Request = ctx.Request.WithContext(reqCtx)
		}
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}

	// Insert the route (radix tree handles conflicts)
	tree.addRoute(pattern, wrappedHandler)

	// Update maxParams for pre-allocation optimization
	paramCount := int(countParams(pattern))
	if paramCount > r.maxParams {
		r.maxParams = paramCount
		// Reinitialize context pool with new maxParams
		r.initContextPool()
	} else if r.contextPool.New == nil {
		r.initContextPool()
	}
}

// initContextPool initializes the context pool with the current maxParams.
func (r *Router) initContextPool() {
	maxParams := r.maxParams
	r.contextPool = sync.Pool{
		New: func() any {
			return &Context{
				Params: make(Params, maxParams),
			}
		},
	}
}

// HandleFunc registers a handler function for the given HTTP method and path pattern.
// This is a convenience wrapper around Handle.
func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) {
	r.Handle(method, pattern, fn)
}

// GET registers a handler for GET requests.
func (r *Router) GET(pattern string, handler http.Handler) {
	r.Handle(http.MethodGet, pattern, handler)
}

// POST registers a handler for POST requests.
func (r *Router) POST(pattern string, handler http.Handler) {
	r.Handle(http.MethodPost, pattern, handler)
}

// PUT registers a handler for PUT requests.
func (r *Router) PUT(pattern string, handler http.Handler) {
	r.Handle(http.MethodPut, pattern, handler)
}

// DELETE registers a handler for DELETE requests.
func (r *Router) DELETE(pattern string, handler http.Handler) {
	r.Handle(http.MethodDelete, pattern, handler)
}

// PATCH registers a handler for PATCH requests.
func (r *Router) PATCH(pattern string, handler http.Handler) {
	r.Handle(http.MethodPatch, pattern, handler)
}

// GETContext registers a zero-allocation handler for GET requests.
func (r *Router) GETContext(pattern string, handler Handler) {
	r.HandleContext(http.MethodGet, pattern, handler)
}

// POSTContext registers a zero-allocation handler for POST requests.
func (r *Router) POSTContext(pattern string, handler Handler) {
	r.HandleContext(http.MethodPost, pattern, handler)
}

// PUTContext registers a zero-allocation handler for PUT requests.
func (r *Router) PUTContext(pattern string, handler Handler) {
	r.HandleContext(http.MethodPut, pattern, handler)
}

// DELETEContext registers a zero-allocation handler for DELETE requests.
func (r *Router) DELETEContext(pattern string, handler Handler) {
	r.HandleContext(http.MethodDelete, pattern, handler)
}

// PATCHContext registers a zero-allocation handler for PATCH requests.
func (r *Router) PATCHContext(pattern string, handler Handler) {
	r.HandleContext(http.MethodPatch, pattern, handler)
}

// ServeHTTP implements http.Handler, making Router compatible with the standard library.
// It routes incoming requests to the appropriate handler based on method and path.
// For zero-allocation routing, use HandleContext to register handlers.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Build middleware chain on first request (lazy init via sync.Once).
	if len(r.middlewares) > 0 {
		r.chainOnce.Do(func() {
			r.chain = r.buildChain(r.dispatch)
		})
	}

	if r.chain != nil {
		ctx := r.getContext(w, req)
		r.chain(ctx)
		r.putContext(ctx)
		return
	}

	r.dispatch(&Context{Writer: w, Request: req})
}

// dispatch performs the tree lookup and calls the matched route handler.
// This is the innermost handler in the middleware chain.
func (r *Router) dispatch(ctx *Context) {
	path := ctx.Request.URL.Path
	method := ctx.Request.Method

	tree, ok := r.trees[method]
	if ok {
		// If coming through the middleware chain, ctx is from the pool and already reset.
		// If called directly (no middleware), we need a pooled context for params.
		if ctx.Params == nil {
			poolCtx := r.getContext(ctx.Writer, ctx.Request)
			handler, found := tree.getValue(path, poolCtx)
			if found {
				handler(poolCtx)
				r.putContext(poolCtx)
				return
			}
			r.putContext(poolCtx)
		} else {
			handler, found := tree.getValue(path, ctx)
			if found {
				handler(ctx)
				return
			}
		}
	}

	// Path not found in the requested method's tree.

	// Trailing slash redirect: try alternate path with/without trailing slash.
	// Uses probeRoute to avoid allocating a temporary Context on every 404.
	if r.RedirectTrailingSlash {
		var redirectPath string
		if len(path) > 1 && path[len(path)-1] == '/' {
			redirectPath = path[:len(path)-1]
		} else {
			redirectPath = path + "/"
		}
		if tree, ok := r.trees[method]; ok {
			if r.probeRoute(tree, redirectPath) {
				http.Redirect(ctx.Writer, ctx.Request, redirectPath, http.StatusMovedPermanently)
				return
			}
		}
	}

	// Auto-OPTIONS: respond with Allow header listing all methods for this path.
	if method == http.MethodOptions && r.HandleOPTIONS {
		if allowed := r.allMethods(path); len(allowed) > 0 {
			ctx.Writer.Header().Set("Allow", strings.Join(allowed, ", "))
			ctx.Writer.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Auto-HEAD: serve GET handler with body suppressed.
	// Both pool and non-pool paths use the same pattern: wrap the writer
	// with bodylessResponseWriter and call the GET handler.
	if method == http.MethodHead && r.HandleHEAD {
		if getTree, hasGet := r.trees[http.MethodGet]; hasGet {
			if r.serveHEAD(ctx, getTree, path) {
				return
			}
		}
	}

	// Check other method trees for a 405 response.
	if allowed := r.allowedMethods(path, method); len(allowed) > 0 {
		r.handleMethodNotAllowed(ctx.Writer, ctx.Request, allowed)
		return
	}

	http.NotFound(ctx.Writer, ctx.Request)
}

// handleMethodNotAllowed writes a 405 response with the Allow header.
func (r *Router) handleMethodNotAllowed(w http.ResponseWriter, req *http.Request, allowed []string) {
	allow := strings.Join(allowed, ", ")
	w.Header().Set("Allow", allow)
	if r.methodNotAllowed != nil {
		r.methodNotAllowed.ServeHTTP(w, req)
		return
	}
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// getContext retrieves a context from the pool and resets it for the request.
func (r *Router) getContext(w http.ResponseWriter, req *http.Request) *Context {
	ctx := r.contextPool.Get().(*Context)
	ctx.Reset(w, req)
	return ctx
}

// putContext returns a context to the pool.
func (r *Router) putContext(ctx *Context) {
	ctx.Writer = nil
	ctx.Request = nil
	ctx.store = nil
	r.contextPool.Put(ctx)
}

// probeRoute checks if a route exists in the tree without allocating a new Context.
// Uses a pooled context for the probe and returns it immediately.
func (r *Router) probeRoute(tree *radixNode, path string) bool {
	probeCtx := r.contextPool.Get().(*Context)
	probeCtx.paramCount = 0
	_, found := tree.getValue(path, probeCtx)
	r.contextPool.Put(probeCtx)
	return found
}

// serveHEAD looks up the GET handler for the given path and serves it with
// the response body suppressed. Returns true if a handler was found and served.
// Handles both pool and non-pool context paths with the same writer-wrapping pattern.
func (r *Router) serveHEAD(ctx *Context, getTree *radixNode, path string) bool {
	if ctx.Params == nil {
		poolCtx := r.getContext(ctx.Writer, ctx.Request)
		handler, found := getTree.getValue(path, poolCtx)
		if found {
			origWriter := poolCtx.Writer
			poolCtx.Writer = &bodylessResponseWriter{ResponseWriter: origWriter}
			handler(poolCtx)
			poolCtx.Writer = origWriter
		}
		r.putContext(poolCtx)
		return found
	}
	handler, found := getTree.getValue(path, ctx)
	if found {
		origWriter := ctx.Writer
		ctx.Writer = &bodylessResponseWriter{ResponseWriter: origWriter}
		handler(ctx)
		ctx.Writer = origWriter
	}
	return found
}

// SetMethodNotAllowed sets a custom handler for 405 Method Not Allowed responses.
// If not set, a default handler is used that returns a plain text 405 response.
func (r *Router) SetMethodNotAllowed(handler http.Handler) {
	r.methodNotAllowed = handler
}

// allowedMethods returns the HTTP methods that have a handler registered
// for the given path, excluding the specified method.
func (r *Router) allowedMethods(path, excludeMethod string) []string {
	var allowed []string
	ctx := &Context{
		Params: make(Params, max(r.maxParams, 1)),
	}
	for method, tree := range r.trees {
		if method == excludeMethod {
			continue
		}
		ctx.paramCount = 0
		if _, found := tree.getValue(path, ctx); found {
			allowed = append(allowed, method)
		}
	}
	return allowed
}

// allMethods returns all HTTP methods that have a handler registered for the given path.
func (r *Router) allMethods(path string) []string {
	var methods []string
	ctx := &Context{
		Params: make(Params, max(r.maxParams, 1)),
	}
	for method, tree := range r.trees {
		ctx.paramCount = 0
		if _, found := tree.getValue(path, ctx); found {
			methods = append(methods, method)
		}
	}
	return methods
}

// bodylessResponseWriter wraps http.ResponseWriter to suppress the response body.
// Used for auto HEAD handling to serve GET handlers without sending body content.
type bodylessResponseWriter struct {
	http.ResponseWriter
}

func (w *bodylessResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil // Discard body, report success
}

// HandlerFunc is a function type for handlers that receive params directly,
// avoiding context allocation overhead.
type HandlerFunc func(w http.ResponseWriter, req *http.Request, params Params)

// fastHandler wraps HandlerFunc to implement http.Handler with zero-alloc param passing.
type fastHandler struct {
	fn     HandlerFunc
	router *Router
}

func (h *fastHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Get params from context (already set by ServeHTTP)
	params := ParamsFromContext(req.Context())
	h.fn(w, req, params)
}

// Lookup performs a route lookup without handling, returning the handler and params.
// This is useful for custom routing logic or zero-allocation scenarios.
// Returns (handler, params, true) if found, or (nil, nil, false) if not found.
func (r *Router) Lookup(method, path string) (Handler, Params, bool) {
	tree, ok := r.trees[method]
	if !ok {
		return nil, nil, false
	}

	// Create temporary context for lookup
	ctx := &Context{
		Params: make(Params, r.maxParams),
	}

	handler, found := tree.getValue(path, ctx)
	if !found {
		return nil, nil, false
	}

	return handler, ctx.Params[:ctx.paramCount], true
}

// countParams counts the number of parameters in a pattern.
func countParams(pattern string) uint8 {
	var n uint8
	for i := range len(pattern) {
		if pattern[i] == ':' || pattern[i] == '*' {
			n++
		}
	}
	return n
}

// RouteInfo contains information about a registered route.
type RouteInfo struct {
	Method  string
	Pattern string
}

// Routes returns a list of all registered routes, sorted by method then pattern.
func (r *Router) Routes() []RouteInfo {
	var routes []RouteInfo
	for method, tree := range r.trees {
		collectRoutes(tree, "", &routes, method)
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Method != routes[j].Method {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Pattern < routes[j].Pattern
	})
	return routes
}

// collectRoutes recursively traverses the radix tree and collects all routes.
func collectRoutes(n *radixNode, prefix string, routes *[]RouteInfo, method string) {
	if n == nil {
		return
	}
	path := prefix + n.path
	if n.handler != nil {
		*routes = append(*routes, RouteInfo{
			Method:  method,
			Pattern: path,
		})
	}
	for _, child := range n.children {
		collectRoutes(child, path, routes, method)
	}
}
