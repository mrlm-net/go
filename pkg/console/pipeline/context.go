package pipeline

import "sync"

// Context carries request data and controls flow through the pipeline.
// It provides a typed key-value store and abort/skip signals.
type Context struct {
	mu      sync.RWMutex
	values  map[any]any
	aborted bool
	err     error
}

// NewContext creates a new empty Context.
func NewContext() *Context {
	return &Context{
		values: make(map[any]any),
	}
}

// Set stores a value in the context under the given key.
func (c *Context) Set(key any, value any) {
	c.mu.Lock()
	c.values[key] = value
	c.mu.Unlock()
}

// Get retrieves a value from the context. Returns the value and true if found,
// or the zero value and false if not found.
func (c *Context) Get(key any) (any, bool) {
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()
	return v, ok
}

// Value returns a typed value from the context.
// Returns the zero value of T if the key is not found or the type does not match.
func Value[T any](ctx *Context, key any) T {
	ctx.mu.RLock()
	v, ok := ctx.values[key]
	ctx.mu.RUnlock()
	if !ok {
		var zero T
		return zero
	}
	typed, ok := v.(T)
	if !ok {
		var zero T
		return zero
	}
	return typed
}

// Abort signals the pipeline to stop execution after the current handler.
func (c *Context) Abort() {
	c.mu.Lock()
	c.aborted = true
	c.mu.Unlock()
}

// AbortWithError signals the pipeline to stop and records an error.
func (c *Context) AbortWithError(err error) {
	c.mu.Lock()
	c.aborted = true
	c.err = err
	c.mu.Unlock()
}

// IsAborted returns true if Abort or AbortWithError has been called.
func (c *Context) IsAborted() bool {
	c.mu.RLock()
	a := c.aborted
	c.mu.RUnlock()
	return a
}

// Err returns the error set by AbortWithError, or nil.
func (c *Context) Err() error {
	c.mu.RLock()
	e := c.err
	c.mu.RUnlock()
	return e
}
