package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware_SingleModifiesResponse(t *testing.T) {
	r := New()
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			ctx.Writer.Header().Set("X-Middleware", "applied")
			next(ctx)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "applied", w.Header().Get("X-Middleware"))
}

func TestMiddleware_MultipleExecuteInOrder(t *testing.T) {
	r := New()
	var order []string

	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			order = append(order, "first-before")
			next(ctx)
			order = append(order, "first-after")
		}
	})
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			order = append(order, "second-before")
			next(ctx)
			order = append(order, "second-after")
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		order = append(order, "handler")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{
		"first-before",
		"second-before",
		"handler",
		"second-after",
		"first-after",
	}, order)
}

func TestMiddleware_ShortCircuit(t *testing.T) {
	r := New()
	handlerCalled := false

	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			// Short-circuit: don't call next
			ctx.Writer.WriteHeader(http.StatusForbidden)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		handlerCalled = true
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, handlerCalled)
}

func TestMiddleware_ZeroAllocs(b *testing.T) {
	r := New()
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Warm up to trigger lazy init
	r.ServeHTTP(w, req)
	w.Body.Reset()
	w.Code = 0

	// Verify the chain is built
	assert.NotNil(b, r.chain)
}

func BenchmarkMiddleware_ZeroAllocs(b *testing.B) {
	r := New()
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Warm up
	r.ServeHTTP(w, req)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		w.Body.Reset()
		w.Code = 0
		r.ServeHTTP(w, req)
	}
}

func TestMiddleware_NotFoundStillWorks(t *testing.T) {
	r := New()
	middlewareCalled := false

	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			middlewareCalled = true
			next(ctx)
		}
	})
	r.GETContext("/exists", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/not-exists", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Middleware runs even for not-found since it wraps the dispatch handler
	assert.True(t, middlewareCalled)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func BenchmarkMiddleware_1(b *testing.B) {
	benchmarkMiddlewareN(b, 1)
}

func BenchmarkMiddleware_5(b *testing.B) {
	benchmarkMiddlewareN(b, 5)
}

func BenchmarkMiddleware_10(b *testing.B) {
	benchmarkMiddlewareN(b, 10)
}

func benchmarkMiddlewareN(b *testing.B, n int) {
	r := New()
	mw := func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
		}
	}
	for range n {
		r.Use(mw)
	}
	r.GETContext("/test", func(ctx *Context) {})
	r.Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}

func TestMiddleware_WithParams(t *testing.T) {
	r := New()
	var paramFromMiddleware string
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
			paramFromMiddleware = ctx.Param("id")
		}
	})
	r.GETContext("/users/:id", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "42", paramFromMiddleware)
}

func TestMiddleware_Build(t *testing.T) {
	r := New()
	var called bool
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			called = true
			next(ctx)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	r.Build()

	assert.NotNil(t, r.chain)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_UseResetsChain(t *testing.T) {
	r := New()
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Trigger chain build
	r.ServeHTTP(w, req)
	assert.NotNil(t, r.chain)

	// Adding middleware resets the chain
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
		}
	})
	assert.Nil(t, r.chain)

	// Chain gets rebuilt on next request
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.NotNil(t, r.chain)
	assert.Equal(t, http.StatusOK, w.Code)
}
