package pipeline

import "sync"

// Pipeline is an ordered chain of middleware wrapping a final handler.
// The zero value is not useful; use New to create a Pipeline.
type Pipeline struct {
	middlewares []Middleware
	handler     Handler
	builtChain  Handler      // cached built chain
	mu          sync.RWMutex // protects builtChain
}

// New creates a Pipeline with the given final handler.
// If handler is nil, a no-op handler is used.
func New(handler Handler) *Pipeline {
	if handler == nil {
		handler = func(ctx *Context) error { return nil }
	}
	return &Pipeline{handler: handler}
}

// Use appends one or more middleware to the pipeline.
// Middleware executes in the order added: first added is outermost.
// Adding middleware invalidates any previously built chain.
func (p *Pipeline) Use(mw ...Middleware) *Pipeline {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.middlewares = append(p.middlewares, mw...)
	p.builtChain = nil // invalidate cache
	return p
}

// Build constructs the middleware chain and caches it for subsequent Run calls.
// This is optional but improves performance when running the same pipeline multiple times.
// Build is thread-safe and returns the pipeline for method chaining.
func (p *Pipeline) Build() *Pipeline {
	p.mu.Lock()
	defer p.mu.Unlock()

	h := p.handler
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		h = p.middlewares[i](h)
	}
	p.builtChain = h
	return p
}

// Run executes the pipeline with the given context.
// If Build() was called previously, it uses the cached chain for better performance.
// Otherwise, it builds the chain on-demand. If the context signals abort
// at any point, execution stops and ErrAborted is returned (unless a more
// specific error was set via AbortWithError).
func (p *Pipeline) Run(ctx *Context) error {
	if ctx == nil {
		ctx = NewContext()
	}

	// Try to use cached built chain
	p.mu.RLock()
	h := p.builtChain
	p.mu.RUnlock()

	// Fallback: build on-demand if not cached
	if h == nil {
		h = p.handler
		for i := len(p.middlewares) - 1; i >= 0; i-- {
			h = p.middlewares[i](h)
		}
	}

	err := h(ctx)
	if err != nil {
		return err
	}

	if ctx.IsAborted() {
		if e := ctx.Err(); e != nil {
			return e
		}
		return ErrAborted
	}

	return nil
}
