package routing

import (
	"context"
	"net/http"
)

// Param is a single URL parameter (key-value pair).
type Param struct {
	Key   string
	Value string
}

// Params is a slice of Param, avoiding map allocation.
// Use Get to retrieve parameter values by key.
type Params []Param

// Get returns the value of the first Param with the given key, or empty string.
// For most use cases with few parameters, linear search is faster than a map lookup.
func (ps Params) Get(key string) string {
	for i := range ps {
		if ps[i].Key == key {
			return ps[i].Value
		}
	}
	return ""
}

// contextKey is the type used for context keys to avoid collisions.
type contextKey int

const (
	paramsKey contextKey = iota
)

// ParamsFromContext extracts path parameters from the request context.
// Returns an empty Params slice if no parameters are found.
func ParamsFromContext(ctx context.Context) Params {
	if params, ok := ctx.Value(paramsKey).(Params); ok {
		return params
	}
	return nil
}

// ParamsFromRequest extracts path parameters from the request.
// This is a convenience wrapper around ParamsFromContext.
func ParamsFromRequest(r *http.Request) Params {
	return ParamsFromContext(r.Context())
}

// contextWithParams returns a new context with the given params attached.
func contextWithParams(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, paramsKey, params)
}
