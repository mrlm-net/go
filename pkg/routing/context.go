package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Context carries request-scoped data through the handler chain.
// It is pooled and reused across requests for zero allocations.
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Params     Params
	paramCount int            // Number of valid params (avoid slice ops)
	store      map[string]any // Lazily initialized key-value store
}

// Handler is the zero-allocation handler function type.
// Handlers receive a pooled Context that is automatically returned to the pool.
type Handler func(ctx *Context)

// Reset prepares the context for reuse from the pool.
// It clears the Writer, Request, and resets the param count to 0.
func (c *Context) Reset(w http.ResponseWriter, req *http.Request) {
	c.Writer = w
	c.Request = req
	c.paramCount = 0
	c.store = nil
}

// Param returns the value of the URL param by name.
// Returns empty string if the parameter is not found.
func (c *Context) Param(key string) string {
	for i := 0; i < c.paramCount; i++ {
		if c.Params[i].Key == key {
			return c.Params[i].Value
		}
	}
	return ""
}

// ParamByIndex returns the parameter at the given index.
// Returns empty Param if index is out of bounds.
// This is faster than Param for sequential access.
func (c *Context) ParamByIndex(idx int) Param {
	if idx >= 0 && idx < c.paramCount {
		return c.Params[idx]
	}
	return Param{}
}

// ParamCount returns the number of parameters extracted from the route.
func (c *Context) ParamCount() int {
	return c.paramCount
}

// Set stores a key-value pair in the context's store.
// The store is lazily initialized on the first Set call.
func (c *Context) Set(key string, val any) {
	if c.store == nil {
		c.store = make(map[string]any)
	}
	c.store[key] = val
}

// Get retrieves a value from the context's store.
// Returns the value and true if found, or nil and false if not found.
func (c *Context) Get(key string) (any, bool) {
	if c.store == nil {
		return nil, false
	}
	val, ok := c.store[key]
	return val, ok
}

// MustGet retrieves a value from the context's store.
// Panics if the key does not exist.
func (c *Context) MustGet(key string) any {
	val, ok := c.Get(key)
	if !ok {
		panic(fmt.Sprintf("routing: key %q does not exist", key))
	}
	return val
}

// JSON writes a JSON response with the given status code and data.
// Sets Content-Type to application/json and encodes the data as JSON.
func (c *Context) JSON(status int, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(data)
}

// String writes a plain text response with the given status code and message.
// Sets Content-Type to text/plain; charset=utf-8.
func (c *Context) String(status int, msg string) error {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write([]byte(msg))
	return err
}

// Redirect sends an HTTP redirect to the client with the given status code and URL.
func (c *Context) Redirect(status int, url string) {
	http.Redirect(c.Writer, c.Request, url, status)
}

// Context returns the request's context.Context.
func (c *Context) Context() context.Context {
	return c.Request.Context()
}
