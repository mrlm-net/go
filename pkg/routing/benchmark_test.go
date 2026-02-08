package routing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Benchmarks for routing performance

func BenchmarkRouter_StaticRoutes(b *testing.B) {
	r := New()
	handler := mockHandler("test")

	// Register static routes
	routes := []string{
		"/",
		"/api",
		"/api/users",
		"/api/users/list",
		"/api/posts",
		"/api/posts/list",
		"/api/comments",
		"/health",
		"/metrics",
		"/status",
	}

	for _, route := range routes {
		r.GET(route, handler)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/list", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouter_ParameterRoutes(b *testing.B) {
	r := New()
	handler := mockHandlerWithParams()

	// Register routes with parameters
	r.GET("/users/:id", handler)
	r.GET("/users/:id/posts/:postId", handler)
	r.GET("/api/v1/users/:id/profile", handler)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouter_WildcardRoutes(b *testing.B) {
	r := New()
	handler := mockHandlerWithParams()

	r.GET("/static/*filepath", handler)

	req := httptest.NewRequest(http.MethodGet, "/static/css/main.css", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouter_MixedRoutes(b *testing.B) {
	r := New()
	handler := mockHandlerWithParams()

	// Mix of static, parameter, and wildcard routes
	r.GET("/", handler)
	r.GET("/health", handler)
	r.GET("/users", handler)
	r.GET("/users/:id", handler)
	r.GET("/users/:id/edit", handler)
	r.GET("/posts", handler)
	r.GET("/posts/:id", handler)
	r.GET("/api/v1/users/:id/posts/:postId", handler)
	r.GET("/static/*filepath", handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouter_NotFound(b *testing.B) {
	r := New()
	handler := mockHandler("test")

	r.GET("/users", handler)
	r.GET("/posts", handler)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouter_LongPath(b *testing.B) {
	r := New()
	handler := mockHandlerWithParams()

	r.GET("/api/v1/organizations/:orgId/teams/:teamId/members/:memberId/projects/:projectId", handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org1/teams/team2/members/member3/projects/proj4", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkRouter_MultipleMethodsSamePath(b *testing.B) {
	r := New()
	handler := mockHandler("test")

	// Same path, different methods
	r.GET("/users/:id", handler)
	r.POST("/users/:id", handler)
	r.PUT("/users/:id", handler)
	r.DELETE("/users/:id", handler)
	r.PATCH("/users/:id", handler)

	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

// Benchmark parameter extraction specifically (using radix tree)
func BenchmarkExtractParams_SingleParam(b *testing.B) {
	r := New()
	handler := func(ctx *Context) {}
	r.HandleContext(http.MethodGet, "/users/:id", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, _ = r.Lookup(http.MethodGet, "/users/123")
	}
}

func BenchmarkExtractParams_MultipleParams(b *testing.B) {
	r := New()
	handler := func(ctx *Context) {}
	r.HandleContext(http.MethodGet, "/users/:id/posts/:postId/comments/:commentId", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, _ = r.Lookup(http.MethodGet, "/users/123/posts/456/comments/789")
	}
}

func BenchmarkExtractParams_Wildcard(b *testing.B) {
	r := New()
	handler := func(ctx *Context) {}
	r.HandleContext(http.MethodGet, "/static/*filepath", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, _ = r.Lookup(http.MethodGet, "/static/css/components/button.css")
	}
}

// Benchmark route registration
func BenchmarkRouter_Registration(b *testing.B) {
	handler := mockHandler("test")

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r := New()
		r.GET("/users", handler)
		r.GET("/users/:id", handler)
		r.POST("/users", handler)
		r.PUT("/users/:id", handler)
		r.DELETE("/users/:id", handler)
	}
}

// BenchmarkRouter_ScalingByRouteCount proves O(k) - lookup time is constant regardless of route count.
// If the algorithm were O(n), lookup time would increase with route count.
func BenchmarkRouter_ScalingByRouteCount(b *testing.B) {
	handler := mockHandler("test")

	for _, routeCount := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("routes=%d", routeCount), func(b *testing.B) {
			r := New()

			// Register routeCount unique routes
			for i := range routeCount {
				pattern := fmt.Sprintf("/api/v1/resource%d", i)
				r.GET(pattern, handler)
			}

			// Always lookup the same route (first one registered)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/resource0", nil)
			w := httptest.NewRecorder()

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				r.ServeHTTP(w, req)
				w.Body.Reset()
			}
		})
	}
}

// BenchmarkRouter_ScalingByPathDepth proves O(k) - lookup time scales linearly with path segments.
func BenchmarkRouter_ScalingByPathDepth(b *testing.B) {
	handler := mockHandler("test")

	for _, depth := range []int{1, 3, 5, 10} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			r := New()

			// Build pattern and path with 'depth' segments
			segments := make([]string, depth)
			for i := range depth {
				segments[i] = fmt.Sprintf("seg%d", i)
			}
			pattern := "/" + strings.Join(segments, "/")
			r.GET(pattern, handler)

			req := httptest.NewRequest(http.MethodGet, pattern, nil)
			w := httptest.NewRecorder()

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				r.ServeHTTP(w, req)
				w.Body.Reset()
			}
		})
	}
}

// BenchmarkRouter_ScalingByRouteCountWithParams tests parameter routes scaling.
func BenchmarkRouter_ScalingByRouteCountWithParams(b *testing.B) {
	handler := mockHandlerWithParams()

	for _, routeCount := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("routes=%d", routeCount), func(b *testing.B) {
			r := New()

			// Register routeCount parameter routes under different prefixes
			for i := range routeCount {
				pattern := fmt.Sprintf("/api/v%d/users/:id", i)
				r.GET(pattern, handler)
			}

			// Lookup a specific route
			req := httptest.NewRequest(http.MethodGet, "/api/v0/users/123", nil)
			w := httptest.NewRecorder()

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				r.ServeHTTP(w, req)
				w.Body.Reset()
			}
		})
	}
}

// Benchmark just the tree lookup without context or response handling
func BenchmarkTree_LookupStatic(b *testing.B) {
	tree := &radixNode{}
	handler := func(ctx *Context) {}

	tree.addRoute("/api/v1/users/list", handler)

	ctx := &Context{
		Params: make(Params, 4),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = tree.getValue("/api/v1/users/list", ctx)
	}
}

func BenchmarkTree_LookupWithParams(b *testing.B) {
	tree := &radixNode{}
	handler := func(ctx *Context) {}

	tree.addRoute("/users/:id/posts/:postId", handler)

	ctx := &Context{
		Params: make(Params, 2),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = tree.getValue("/users/123/posts/456", ctx)
	}
}

func BenchmarkTree_LookupWithWildcard(b *testing.B) {
	tree := &radixNode{}
	handler := func(ctx *Context) {}

	tree.addRoute("/static/*filepath", handler)

	ctx := &Context{
		Params: make(Params, 1),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = tree.getValue("/static/css/main.css", ctx)
	}
}

// Benchmark ServeHTTP without context allocation (using no-params handler)
func BenchmarkRouter_ServeHTTPNoContext(b *testing.B) {
	r := New()
	handler := mockHandler("test")

	r.GET("/users/:id", handler)
	r.GET("/users/:id/posts/:postId", handler)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

// Benchmark Lookup method (for custom routing, avoids http.ResponseWriter)
func BenchmarkRouter_Lookup(b *testing.B) {
	r := New()
	handler := mockHandler("test")

	r.GET("/users/:id", handler)
	r.GET("/users/:id/posts/:postId", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, _ = r.Lookup(http.MethodGet, "/users/123/posts/456")
	}
}

// Benchmark zero-allocation context-based routing
func BenchmarkRouter_ContextZeroAlloc(b *testing.B) {
	r := New()

	handler := func(ctx *Context) {
		// Access params to ensure they're extracted
		_ = ctx.Param("id")
		_ = ctx.Param("postId")
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/users/:id/posts/:postId", handler)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// Benchmark context-based static routes
func BenchmarkRouter_ContextStatic(b *testing.B) {
	r := New()

	handler := func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/api/v1/users/list", handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/list", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// Benchmark context-based wildcard routes
func BenchmarkRouter_ContextWildcard(b *testing.B) {
	r := New()

	handler := func(ctx *Context) {
		_ = ctx.Param("filepath")
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/static/*filepath", handler)

	req := httptest.NewRequest(http.MethodGet, "/static/css/main.css", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// Benchmark tree lookup with context (no HTTP overhead)
func BenchmarkTree_LookupContext(b *testing.B) {
	tree := &radixNode{}

	handler := func(ctx *Context) {}
	tree.addRoute("/users/:id/posts/:postId", handler)

	ctx := &Context{
		Params: make(Params, 2),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = tree.getValue("/users/123/posts/456", ctx)
	}
}

// Benchmark realistic API scenario
func BenchmarkRouter_RealisticAPI(b *testing.B) {
	r := New()
	handler := mockHandlerWithParams()

	// Typical REST API routes
	r.GET("/api/v1/health", handler)
	r.GET("/api/v1/users", handler)
	r.POST("/api/v1/users", handler)
	r.GET("/api/v1/users/:id", handler)
	r.PUT("/api/v1/users/:id", handler)
	r.DELETE("/api/v1/users/:id", handler)
	r.GET("/api/v1/users/:id/posts", handler)
	r.GET("/api/v1/users/:id/posts/:postId", handler)
	r.POST("/api/v1/users/:id/posts", handler)
	r.GET("/api/v1/posts", handler)
	r.GET("/api/v1/posts/:id", handler)
	r.GET("/api/v1/posts/:id/comments", handler)
	r.POST("/api/v1/posts/:id/comments", handler)
	r.GET("/api/v1/search", handler)
	r.GET("/static/*filepath", handler)

	paths := []string{
		"/api/v1/health",
		"/api/v1/users",
		"/api/v1/users/123",
		"/api/v1/users/123/posts",
		"/api/v1/users/123/posts/456",
		"/api/v1/posts/789/comments",
		"/static/css/main.css",
	}

	requests := make([]*http.Request, len(paths))
	for i, path := range paths {
		requests[i] = httptest.NewRequest(http.MethodGet, path, nil)
	}

	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		for _, req := range requests {
			r.ServeHTTP(w, req)
			w.Body.Reset()
		}
	}
}
