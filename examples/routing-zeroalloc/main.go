// Package main demonstrates zero-allocation routing with the Context API.
// This example achieves Gin-like performance with zero allocations per request.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mrlm-net/go/pkg/routing"
)

func main() {
	// Create a new router
	router := routing.New()

	// Register routes using the zero-allocation Context API
	// These handlers achieve 0 allocs/op for maximum performance

	// Static routes (no parameters)
	router.GETContext("/", handleHomeZeroAlloc)
	router.GETContext("/health", handleHealthZeroAlloc)

	// Routes with path parameters
	router.GETContext("/users/:id", handleGetUserZeroAlloc)
	router.POSTContext("/users", handleCreateUserZeroAlloc)
	router.PUTContext("/users/:id", handleUpdateUserZeroAlloc)
	router.DELETEContext("/users/:id", handleDeleteUserZeroAlloc)

	// Nested routes with multiple parameters
	router.GETContext("/users/:id/posts/:postId", handleGetPostZeroAlloc)
	router.POSTContext("/users/:id/posts", handleCreatePostZeroAlloc)

	// Wildcard route for static files
	router.GETContext("/static/*filepath", handleStaticFilesZeroAlloc)

	// Start the server
	fmt.Println("Starting zero-allocation server on :8080")
	fmt.Println("This server achieves 0 allocs/op for optimal performance")
	fmt.Println("\nTry these endpoints:")
	fmt.Println("  GET  http://localhost:8080/")
	fmt.Println("  GET  http://localhost:8080/health")
	fmt.Println("  GET  http://localhost:8080/users/123")
	fmt.Println("  GET  http://localhost:8080/users/123/posts/456")
	fmt.Println("  GET  http://localhost:8080/static/css/main.css")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Zero-allocation handlers using the Context API

func handleHomeZeroAlloc(ctx *routing.Context) {
	fmt.Fprintf(ctx.Writer, "Welcome to the zero-allocation routing example!\n")
}

func handleHealthZeroAlloc(ctx *routing.Context) {
	fmt.Fprintf(ctx.Writer, "OK\n")
}

func handleGetUserZeroAlloc(ctx *routing.Context) {
	// Access params with zero allocations
	userID := ctx.Param("id")
	fmt.Fprintf(ctx.Writer, "Getting user: %s\n", userID)
}

func handleCreateUserZeroAlloc(ctx *routing.Context) {
	fmt.Fprintf(ctx.Writer, "Creating new user\n")
}

func handleUpdateUserZeroAlloc(ctx *routing.Context) {
	userID := ctx.Param("id")
	fmt.Fprintf(ctx.Writer, "Updating user: %s\n", userID)
}

func handleDeleteUserZeroAlloc(ctx *routing.Context) {
	userID := ctx.Param("id")
	fmt.Fprintf(ctx.Writer, "Deleting user: %s\n", userID)
}

func handleGetPostZeroAlloc(ctx *routing.Context) {
	// Access multiple params with zero allocations
	userID := ctx.Param("id")
	postID := ctx.Param("postId")
	fmt.Fprintf(ctx.Writer, "Getting post %s for user %s\n", postID, userID)
}

func handleCreatePostZeroAlloc(ctx *routing.Context) {
	userID := ctx.Param("id")
	fmt.Fprintf(ctx.Writer, "Creating new post for user %s\n", userID)
}

func handleStaticFilesZeroAlloc(ctx *routing.Context) {
	filepath := ctx.Param("filepath")
	fmt.Fprintf(ctx.Writer, "Serving static file: %s\n", filepath)
}
