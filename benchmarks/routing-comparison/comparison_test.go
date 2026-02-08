package comparison_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mrlm-net/go/pkg/routing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// Comparison benchmarks between mrlm/routing and gin-gonic/gin.
// These benchmarks measure the same operations to ensure fair comparison.

// BenchmarkComparison_Gin_StaticRoute benchmarks Gin with a static route.
func BenchmarkComparison_Gin_StaticRoute(b *testing.B) {
	r := gin.New()
	r.GET("/api/v1/users/list", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

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

// BenchmarkComparison_MRLM_StaticRoute benchmarks mrlm/routing with a static route.
func BenchmarkComparison_MRLM_StaticRoute(b *testing.B) {
	r := routing.New()
	r.GETContext("/api/v1/users/list", func(ctx *routing.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

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

// BenchmarkComparison_Gin_SingleParam benchmarks Gin with a single parameter.
func BenchmarkComparison_Gin_SingleParam(b *testing.B) {
	r := gin.New()
	r.GET("/users/:id", func(c *gin.Context) {
		_ = c.Param("id")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_MRLM_SingleParam benchmarks mrlm/routing with a single parameter.
func BenchmarkComparison_MRLM_SingleParam(b *testing.B) {
	r := routing.New()
	r.GETContext("/users/:id", func(ctx *routing.Context) {
		_ = ctx.Param("id")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_Gin_TwoParams benchmarks Gin with two parameters.
func BenchmarkComparison_Gin_TwoParams(b *testing.B) {
	r := gin.New()
	r.GET("/users/:id/posts/:postId", func(c *gin.Context) {
		_ = c.Param("id")
		_ = c.Param("postId")
		c.Status(http.StatusOK)
	})

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

// BenchmarkComparison_MRLM_TwoParams benchmarks mrlm/routing with two parameters.
func BenchmarkComparison_MRLM_TwoParams(b *testing.B) {
	r := routing.New()
	r.GETContext("/users/:id/posts/:postId", func(ctx *routing.Context) {
		_ = ctx.Param("id")
		_ = ctx.Param("postId")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

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

// BenchmarkComparison_Gin_FourParams benchmarks Gin with four parameters.
func BenchmarkComparison_Gin_FourParams(b *testing.B) {
	r := gin.New()
	r.GET("/api/v1/orgs/:orgId/teams/:teamId/members/:memberId/projects/:projectId", func(c *gin.Context) {
		_ = c.Param("orgId")
		_ = c.Param("teamId")
		_ = c.Param("memberId")
		_ = c.Param("projectId")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/org1/teams/team2/members/member3/projects/proj4", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_MRLM_FourParams benchmarks mrlm/routing with four parameters.
func BenchmarkComparison_MRLM_FourParams(b *testing.B) {
	r := routing.New()
	r.GETContext("/api/v1/orgs/:orgId/teams/:teamId/members/:memberId/projects/:projectId", func(ctx *routing.Context) {
		_ = ctx.Param("orgId")
		_ = ctx.Param("teamId")
		_ = ctx.Param("memberId")
		_ = ctx.Param("projectId")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/org1/teams/team2/members/member3/projects/proj4", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_Gin_Wildcard benchmarks Gin with a wildcard parameter.
func BenchmarkComparison_Gin_Wildcard(b *testing.B) {
	r := gin.New()
	r.GET("/static/*filepath", func(c *gin.Context) {
		_ = c.Param("filepath")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/static/css/components/button.css", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_MRLM_Wildcard benchmarks mrlm/routing with a wildcard parameter.
func BenchmarkComparison_MRLM_Wildcard(b *testing.B) {
	r := routing.New()
	r.GETContext("/static/*filepath", func(ctx *routing.Context) {
		_ = ctx.Param("filepath")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/static/css/components/button.css", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_Gin_ManyRoutes benchmarks Gin with many registered routes.
func BenchmarkComparison_Gin_ManyRoutes(b *testing.B) {
	r := gin.New()

	r.GET("/api/v1/health", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/users", func(c *gin.Context) { c.Status(200) })
	r.POST("/api/v1/users", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/users/:id", func(c *gin.Context) { c.Status(200) })
	r.PUT("/api/v1/users/:id", func(c *gin.Context) { c.Status(200) })
	r.DELETE("/api/v1/users/:id", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/users/:id/posts", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/users/:id/posts/:postId", func(c *gin.Context) {
		_ = c.Param("id")
		_ = c.Param("postId")
		c.Status(200)
	})
	r.POST("/api/v1/users/:id/posts", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/posts", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/posts/:id", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/posts/:id/comments", func(c *gin.Context) { c.Status(200) })
	r.POST("/api/v1/posts/:id/comments", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/v1/search", func(c *gin.Context) { c.Status(200) })
	r.GET("/static/*filepath", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}

// BenchmarkComparison_MRLM_ManyRoutes benchmarks mrlm/routing with many registered routes.
func BenchmarkComparison_MRLM_ManyRoutes(b *testing.B) {
	r := routing.New()

	r.GETContext("/api/v1/health", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/users", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.POSTContext("/api/v1/users", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/users/:id", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.PUTContext("/api/v1/users/:id", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.DELETEContext("/api/v1/users/:id", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/users/:id/posts", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/users/:id/posts/:postId", func(ctx *routing.Context) {
		_ = ctx.Param("id")
		_ = ctx.Param("postId")
		ctx.Writer.WriteHeader(200)
	})
	r.POSTContext("/api/v1/users/:id/posts", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/posts", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/posts/:id", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/posts/:id/comments", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.POSTContext("/api/v1/posts/:id/comments", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/api/v1/search", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })
	r.GETContext("/static/*filepath", func(ctx *routing.Context) { ctx.Writer.WriteHeader(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
		w.Body.Reset()
		w.Code = 0
	}
}
