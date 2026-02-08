package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup_PrefixRouting(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("users"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "users", w.Body.String())
}

func TestGroup_NestedGroups(t *testing.T) {
	r := New()
	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("v1-users"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v1-users", w.Body.String())
}

func TestGroup_Middleware(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			ctx.Writer.Header().Set("X-Group-MW", "api")
			next(ctx)
		}
	})
	api.GETContext("/data", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Route without group middleware
	r.GETContext("/health", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Group route should have middleware
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "api", w.Header().Get("X-Group-MW"))

	// Non-group route should NOT have group middleware
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Header().Get("X-Group-MW"))
}

func TestGroup_NestedMiddleware(t *testing.T) {
	r := New()
	var order []string

	api := r.Group("/api")
	api.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			order = append(order, "api-before")
			next(ctx)
			order = append(order, "api-after")
		}
	})

	v1 := api.Group("/v1")
	v1.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			order = append(order, "v1-before")
			next(ctx)
			order = append(order, "v1-after")
		}
	})

	v1.GETContext("/items", func(ctx *Context) {
		order = append(order, "handler")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{
		"api-before",
		"v1-before",
		"handler",
		"v1-after",
		"api-after",
	}, order)
}

func TestGroup_MethodShortcuts(t *testing.T) {
	tests := []struct {
		name   string
		method string
		setup  func(g *Group)
	}{
		{"GET", http.MethodGet, func(g *Group) {
			g.GETContext("/test", func(ctx *Context) { ctx.Writer.WriteHeader(200) })
		}},
		{"POST", http.MethodPost, func(g *Group) {
			g.POSTContext("/test", func(ctx *Context) { ctx.Writer.WriteHeader(200) })
		}},
		{"PUT", http.MethodPut, func(g *Group) {
			g.PUTContext("/test", func(ctx *Context) { ctx.Writer.WriteHeader(200) })
		}},
		{"DELETE", http.MethodDelete, func(g *Group) {
			g.DELETEContext("/test", func(ctx *Context) { ctx.Writer.WriteHeader(200) })
		}},
		{"PATCH", http.MethodPatch, func(g *Group) {
			g.PATCHContext("/test", func(ctx *Context) { ctx.Writer.WriteHeader(200) })
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			g := r.Group("/api")
			tt.setup(g)

			req := httptest.NewRequest(tt.method, "/api/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestGroup_WithParams(t *testing.T) {
	r := New()
	api := r.Group("/api/v1")
	api.GETContext("/users/:id", func(ctx *Context) {
		ctx.Writer.Write([]byte(ctx.Param("id")))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "42", w.Body.String())
}

func TestGroup_UseChaining(t *testing.T) {
	r := New()
	g := r.Group("/api").Use(func(next Handler) Handler {
		return func(ctx *Context) {
			ctx.Writer.Header().Set("X-First", "1")
			next(ctx)
		}
	}).Use(func(next Handler) Handler {
		return func(ctx *Context) {
			ctx.Writer.Header().Set("X-Second", "2")
			next(ctx)
		}
	})

	g.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1", w.Header().Get("X-First"))
	assert.Equal(t, "2", w.Header().Get("X-Second"))
}

func TestGroup_Handle_HttpHandler(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.Handle(http.MethodGet, "/legacy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("legacy"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/legacy", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "legacy", w.Body.String())
}

func BenchmarkGroup_ZeroAllocs(b *testing.B) {
	r := New()
	api := r.Group("/api")
	api.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			next(ctx)
		}
	})
	api.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
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
