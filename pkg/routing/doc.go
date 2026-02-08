// Package routing provides a high-performance HTTP router with zero-allocation routing.
//
// # Overview
//
// The routing package implements a fast segment-tree-based HTTP router that supports:
//   - Static routes: /users, /api/v1/health
//   - Path parameters: /users/:id, /posts/:postId/comments/:commentId
//   - Wildcard routes: /static/*filepath
//   - Zero-allocation routing with Context API (Gin-like performance)
//   - Standard http.Handler compatibility
//
// # Performance
//
// The router achieves zero allocations per request when using the Context API:
//   - BenchmarkRouter_ContextZeroAlloc: ~226-427 ns/op, 0 B/op, 0 allocs/op
//   - BenchmarkTree_LookupContext: ~125-202 ns/op, 0 B/op, 0 allocs/op
//
// This is comparable to Gin's performance while maintaining idiomatic Go code.
//
// # Basic Usage (Standard API)
//
//	router := routing.New()
//
//	// Register handlers using standard http.Handler
//	router.GET("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprintf(w, "List users\n")
//	}))
//
//	router.GET("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		params := routing.ParamsFromContext(r.Context())
//		id := params.Get("id")
//		fmt.Fprintf(w, "User ID: %s\n", id)
//	}))
//
//	http.ListenAndServe(":8080", router)
//
// # Zero-Allocation Usage (Context API)
//
// For maximum performance with zero allocations, use the Context API:
//
//	router := routing.New()
//
//	// Register handlers using the zero-allocation Context API
//	router.GETContext("/users/:id", func(ctx *routing.Context) {
//		id := ctx.Param("id")
//		fmt.Fprintf(ctx.Writer, "User ID: %s\n", id)
//	})
//
//	router.POSTContext("/users", func(ctx *routing.Context) {
//		// Access request and writer through context
//		body, _ := io.ReadAll(ctx.Request.Body)
//		fmt.Fprintf(ctx.Writer, "Created user\n")
//	})
//
//	http.ListenAndServe(":8080", router)
//
// # Path Parameters
//
// Parameters are defined with a colon prefix:
//
//	router.GET("/users/:id/posts/:postId", handler)
//
// Access parameters via Context.Param() or ParamsFromContext():
//
//	// Context API (zero-allocation)
//	id := ctx.Param("id")
//	postId := ctx.Param("postId")
//
//	// Standard API (uses request context)
//	params := routing.ParamsFromContext(r.Context())
//	id := params.Get("id")
//	postId := params.Get("postId")
//
// # Wildcard Routes
//
// Wildcard routes capture the rest of the path:
//
//	router.GET("/static/*filepath", handler)
//
// Matching /static/css/main.css extracts filepath=css/main.css.
//
// # Conflict Detection
//
// The router detects conflicting routes at registration time:
//
//	router.GET("/users/:id", handler1)
//	router.GET("/users/:userId", handler2) // PANIC: parameter name conflict
//
//	router.GET("/files/*path", handler1)
//	router.GET("/files/:id", handler2) // PANIC: wildcard vs parameter conflict
//
// # Method Shortcuts
//
// Convenience methods for common HTTP methods:
//
//	// Standard API
//	router.GET(pattern, handler)
//	router.POST(pattern, handler)
//	router.PUT(pattern, handler)
//	router.DELETE(pattern, handler)
//	router.PATCH(pattern, handler)
//
//	// Context API (zero-allocation)
//	router.GETContext(pattern, handler)
//	router.POSTContext(pattern, handler)
//	router.PUTContext(pattern, handler)
//	router.DELETEContext(pattern, handler)
//	router.PATCHContext(pattern, handler)
//
// # Architecture
//
// The router uses a segment tree data structure where:
//   - Each node represents a path segment
//   - Static segments are stored in a map for O(1) lookup
//   - Parameter segments have a single child node
//   - Wildcard segments capture remaining path
//   - Lookup priority: static > parameter > wildcard
//
// Context pooling enables zero allocations:
//   - Contexts are pre-allocated with params slice
//   - Retrieved from sync.Pool before request
//   - Returned to pool after request
//   - No heap allocations per request
//
// # Thread Safety
//
// Router is not thread-safe during route registration. All routes should be
// registered before serving requests. Once registered, the router is safe
// for concurrent requests via ServeHTTP.
//
// # Examples
//
// See examples/routing and examples/routing-zeroalloc for complete working examples.
package routing
