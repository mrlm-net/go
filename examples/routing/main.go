// Package main demonstrates the routing package with a simple HTTP server.
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

	// Register static routes
	router.GET("/", http.HandlerFunc(handleHome))
	router.GET("/health", http.HandlerFunc(handleHealth))

	// Register routes with path parameters
	router.GET("/users/:id", http.HandlerFunc(handleGetUser))
	router.POST("/users", http.HandlerFunc(handleCreateUser))
	router.PUT("/users/:id", http.HandlerFunc(handleUpdateUser))
	router.DELETE("/users/:id", http.HandlerFunc(handleDeleteUser))

	// Register nested routes with multiple parameters
	router.GET("/users/:id/posts/:postId", http.HandlerFunc(handleGetPost))
	router.POST("/users/:id/posts", http.HandlerFunc(handleCreatePost))

	// Register wildcard route for static files
	router.GET("/static/*filepath", http.HandlerFunc(handleStaticFiles))

	// Start the server
	fmt.Println("Starting server on :8080")
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  http://localhost:8080/")
	fmt.Println("  GET  http://localhost:8080/health")
	fmt.Println("  GET  http://localhost:8080/users/123")
	fmt.Println("  GET  http://localhost:8080/users/123/posts/456")
	fmt.Println("  GET  http://localhost:8080/static/css/main.css")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the routing example!\n")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	params := routing.ParamsFromContext(r.Context())
	userID := params.Get("id")
	fmt.Fprintf(w, "Getting user: %s\n", userID)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Creating new user\n")
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	params := routing.ParamsFromContext(r.Context())
	userID := params.Get("id")
	fmt.Fprintf(w, "Updating user: %s\n", userID)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	params := routing.ParamsFromContext(r.Context())
	userID := params.Get("id")
	fmt.Fprintf(w, "Deleting user: %s\n", userID)
}

func handleGetPost(w http.ResponseWriter, r *http.Request) {
	params := routing.ParamsFromContext(r.Context())
	userID := params.Get("id")
	postID := params.Get("postId")
	fmt.Fprintf(w, "Getting post %s for user %s\n", postID, userID)
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	params := routing.ParamsFromContext(r.Context())
	userID := params.Get("id")
	fmt.Fprintf(w, "Creating new post for user %s\n", userID)
}

func handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	params := routing.ParamsFromContext(r.Context())
	filepath := params.Get("filepath")
	fmt.Fprintf(w, "Serving static file: %s\n", filepath)
}
