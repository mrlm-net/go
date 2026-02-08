package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mrlm-net/go/pkg/routing"
)

// User represents a user in our system.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// In-memory user store for demo purposes.
var users = map[string]User{
	"1": {ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()},
	"2": {ID: "2", Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now()},
}

func main() {
	r := routing.New()

	// Global middleware: Request ID (applied to all routes)
	r.Use(requestIDMiddleware())

	// Global middleware: Logging (applied to all routes)
	r.Use(loggingMiddleware())

	// Public routes (no auth required)
	r.GETContext("/health", healthHandler)

	// API routes with auth middleware applied per-route
	r.GETContext("/api/users", withAuth(listUsers))
	r.POSTContext("/api/users", withAuth(createUser))
	r.GETContext("/api/users/:id", withAuth(getUser))
	r.PUTContext("/api/users/:id", withAuth(updateUser))
	r.DELETEContext("/api/users/:id", withAuth(deleteUser))

	// Build middleware chain (optional, for performance)
	r.Build()

	log.Println("Starting server on :8080")
	log.Println("Try: curl http://localhost:8080/health")
	log.Println("Try: curl -H \"Authorization: Bearer token\" http://localhost:8080/api/users")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// requestIDMiddleware generates a unique request ID for each request.
func requestIDMiddleware() routing.Middleware {
	return func(next routing.Handler) routing.Handler {
		return func(ctx *routing.Context) {
			requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
			ctx.Set("request_id", requestID)
			ctx.Writer.Header().Set("X-Request-ID", requestID)
			next(ctx)
		}
	}
}

// loggingMiddleware logs each request with method, path, and duration.
func loggingMiddleware() routing.Middleware {
	return func(next routing.Handler) routing.Handler {
		return func(ctx *routing.Context) {
			start := time.Now()
			requestID, _ := ctx.Get("request_id")

			next(ctx)

			duration := time.Since(start)
			log.Printf("[%v] %s %s - %v", requestID, ctx.Request.Method, ctx.Request.URL.Path, duration)
		}
	}
}

// withAuth wraps a handler with authentication middleware.
// This demonstrates per-route middleware application.
func withAuth(handler routing.Handler) routing.Handler {
	return func(ctx *routing.Context) {
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing authorization header",
			})
			return
		}

		// Simple token validation (in real app, validate JWT/token properly)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid authorization format",
			})
			return
		}

		// Store authenticated user info in context
		ctx.Set("user_id", "authenticated-user")
		ctx.Set("token", strings.TrimPrefix(authHeader, "Bearer "))

		handler(ctx)
	}
}

// healthHandler returns the health status of the service.
func healthHandler(ctx *routing.Context) {
	ctx.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// listUsers returns all users.
func listUsers(ctx *routing.Context) {
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	ctx.JSON(http.StatusOK, map[string]any{
		"users": userList,
		"count": len(userList),
	})
}

// createUser creates a new user.
func createUser(ctx *routing.Context) {
	newID := fmt.Sprintf("%d", time.Now().Unix())
	user := User{
		ID:        newID,
		Name:      "New User",
		Email:     "new@example.com",
		CreatedAt: time.Now(),
	}
	users[newID] = user

	ctx.JSON(http.StatusCreated, map[string]any{
		"user":    user,
		"message": "user created successfully",
	})
}

// getUser retrieves a user by ID.
func getUser(ctx *routing.Context) {
	id := ctx.Param("id")
	user, exists := users[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"user": user,
	})
}

// updateUser updates an existing user.
func updateUser(ctx *routing.Context) {
	id := ctx.Param("id")
	user, exists := users[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	// In a real app, parse request body and update fields
	user.Name = "Updated Name"
	users[id] = user

	ctx.JSON(http.StatusOK, map[string]any{
		"user":    user,
		"message": "user updated successfully",
	})
}

// deleteUser deletes a user by ID.
func deleteUser(ctx *routing.Context) {
	id := ctx.Param("id")
	if _, exists := users[id]; !exists {
		ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	delete(users, id)
	ctx.JSON(http.StatusOK, map[string]string{
		"message": "user deleted successfully",
	})
}
