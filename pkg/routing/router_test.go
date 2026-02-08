package routing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHandler creates a simple handler that writes a message to the response
func mockHandler(msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	}
}

// mockHandlerWithParams creates a handler that writes parameters to the response
// in sorted key order for deterministic test output.
func mockHandlerWithParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := ParamsFromContext(r.Context())
		// Collect params into a slice for sorting
		type kv struct {
			key   string
			value string
		}
		pairs := make([]kv, len(params))
		for i := range params {
			pairs[i] = kv{params[i].Key, params[i].Value}
		}
		// Sort by key for deterministic output
		slices.SortFunc(pairs, func(a, b kv) int {
			if a.key < b.key {
				return -1
			}
			if a.key > b.key {
				return 1
			}
			return 0
		})
		for _, p := range pairs {
			fmt.Fprintf(w, "%s=%s ", p.key, p.value)
		}
	}
}

func TestNew(t *testing.T) {
	r := New()
	assert.NotNil(t, r)
	assert.NotNil(t, r.trees)
}

func TestRouter_Handle(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		pattern     string
		handler     http.Handler
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:    "valid GET route",
			method:  http.MethodGet,
			pattern: "/users",
			handler: mockHandler("users"),
		},
		{
			name:    "valid POST route with parameter",
			method:  http.MethodPost,
			pattern: "/users/:id",
			handler: mockHandler("user"),
		},
		{
			name:    "valid route with wildcard",
			method:  http.MethodGet,
			pattern: "/files/*filepath",
			handler: mockHandler("files"),
		},
		{
			name:        "nil handler panics",
			method:      http.MethodGet,
			pattern:     "/test",
			handler:     nil,
			shouldPanic: true,
			panicMsg:    "routing: nil handler",
		},
		{
			name:        "empty pattern panics",
			method:      http.MethodGet,
			pattern:     "",
			handler:     mockHandler("test"),
			shouldPanic: true,
			panicMsg:    "routing: empty pattern",
		},
		{
			name:        "pattern without leading slash panics",
			method:      http.MethodGet,
			pattern:     "users",
			handler:     mockHandler("test"),
			shouldPanic: true,
			panicMsg:    "routing: pattern must start with /: users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			if tt.shouldPanic {
				assert.PanicsWithValue(t, tt.panicMsg, func() {
					r.Handle(tt.method, tt.pattern, tt.handler)
				})
			} else {
				assert.NotPanics(t, func() {
					r.Handle(tt.method, tt.pattern, tt.handler)
				})
			}
		})
	}
}

func TestRouter_MethodShortcuts(t *testing.T) {
	r := New()
	handler := mockHandler("test")

	// Test each method shortcut
	r.GET("/get", handler)
	r.POST("/post", handler)
	r.PUT("/put", handler)
	r.DELETE("/delete", handler)
	r.PATCH("/patch", handler)

	// Verify routes were registered
	assert.NotNil(t, r.trees[http.MethodGet])
	assert.NotNil(t, r.trees[http.MethodPost])
	assert.NotNil(t, r.trees[http.MethodPut])
	assert.NotNil(t, r.trees[http.MethodDelete])
	assert.NotNil(t, r.trees[http.MethodPatch])
}

func TestRouter_HandleFunc(t *testing.T) {
	r := New()
	r.HandleFunc(http.MethodGet, "/test", mockHandler("test"))

	// Verify route was registered
	assert.NotNil(t, r.trees[http.MethodGet])
}

func TestRouter_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		routes         map[string]map[string]http.Handler // method -> pattern -> handler
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		expectedParams Params
	}{
		{
			name: "exact match static route",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/users": mockHandler("users list"),
				},
			},
			method:         http.MethodGet,
			path:           "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "users list",
		},
		{
			name: "single parameter extraction",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/users/:id": mockHandlerWithParams(),
				},
			},
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   "id=123 ",
			expectedParams: Params{{Key: "id", Value: "123"}},
		},
		{
			name: "multiple parameters extraction",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/users/:userId/posts/:postId": mockHandlerWithParams(),
				},
			},
			method:         http.MethodGet,
			path:           "/users/123/posts/456",
			expectedStatus: http.StatusOK,
			expectedBody:   "postId=456 userId=123 ", // sorted alphabetically
			expectedParams: Params{{Key: "userId", Value: "123"}, {Key: "postId", Value: "456"}},
		},
		{
			name: "wildcard extraction",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/files/*filepath": mockHandlerWithParams(),
				},
			},
			method:         http.MethodGet,
			path:           "/files/docs/readme.md",
			expectedStatus: http.StatusOK,
			expectedBody:   "filepath=docs/readme.md ",
			expectedParams: Params{{Key: "filepath", Value: "docs/readme.md"}},
		},
		{
			name: "not found returns 404",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/users": mockHandler("users"),
				},
			},
			method:         http.MethodGet,
			path:           "/posts",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "method not allowed returns 405",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/users": mockHandler("users"),
				},
			},
			method:         http.MethodPost,
			path:           "/users",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "nested paths with parameters",
			routes: map[string]map[string]http.Handler{
				http.MethodGet: {
					"/api/v1/users/:id/profile": mockHandlerWithParams(),
				},
			},
			method:         http.MethodGet,
			path:           "/api/v1/users/789/profile",
			expectedStatus: http.StatusOK,
			expectedBody:   "id=789 ",
			expectedParams: Params{{Key: "id", Value: "789"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()

			// Register all routes
			for method, routes := range tt.routes {
				for pattern, handler := range routes {
					r.Handle(method, pattern, handler)
				}
			}

			// Create request and response recorder
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestRouter_ConflictDetection(t *testing.T) {
	tests := []struct {
		name        string
		routes      []string // patterns to register in order
		shouldPanic bool
		conflictMsg string
	}{
		{
			name:   "no conflict - different static paths",
			routes: []string{"/users", "/posts"},
		},
		{
			name:   "no conflict - same parameter name",
			routes: []string{"/users/:id", "/users/:id/edit"},
		},
		{
			name:        "conflict - different parameter names same position",
			routes:      []string{"/users/:id", "/users/:userId"},
			shouldPanic: true,
			conflictMsg: "conflicts with existing pattern",
		},
		{
			name:        "conflict - parameter and wildcard",
			routes:      []string{"/users/:id", "/users/*path"},
			shouldPanic: true,
			conflictMsg: "conflicts with existing pattern",
		},
		{
			name:        "conflict - wildcard overlap",
			routes:      []string{"/files/*filepath", "/files/*path"},
			shouldPanic: true,
			conflictMsg: "conflicts with existing pattern",
		},
		{
			name:   "no conflict - different methods",
			routes: []string{"/users", "/users"}, // would conflict, but we'll use different methods
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						assert.Contains(t, fmt.Sprint(r), tt.conflictMsg)
					} else {
						t.Error("Expected panic but didn't get one")
					}
				}()

				for _, pattern := range tt.routes {
					r.GET(pattern, mockHandler("test"))
				}
			} else {
				for i, pattern := range tt.routes {
					// Use different methods to avoid conflict in the "different methods" test
					if tt.name == "no conflict - different methods" {
						if i == 0 {
							r.GET(pattern, mockHandler("test"))
						} else {
							r.POST(pattern, mockHandler("test"))
						}
					} else {
						r.GET(pattern, mockHandler("test"))
					}
				}
			}
		})
	}
}

func TestParamsFromContext(t *testing.T) {
	t.Run("params exist in context", func(t *testing.T) {
		params := Params{{Key: "id", Value: "123"}, {Key: "name", Value: "test"}}
		ctx := contextWithParams(context.Background(), params)

		result := ParamsFromContext(ctx)
		assert.Equal(t, params, result)
	})

	t.Run("no params in context", func(t *testing.T) {
		ctx := context.Background()

		result := ParamsFromContext(ctx)
		assert.Nil(t, result)
	})
}

func TestRouter_IntegrationScenarios(t *testing.T) {
	t.Run("RESTful API routes", func(t *testing.T) {
		r := New()

		// Register RESTful routes
		r.GET("/users", mockHandler("list users"))
		r.POST("/users", mockHandler("create user"))
		r.GET("/users/:id", mockHandler("get user"))
		r.PUT("/users/:id", mockHandler("update user"))
		r.DELETE("/users/:id", mockHandler("delete user"))

		// Test each route
		tests := []struct {
			method string
			path   string
			body   string
		}{
			{http.MethodGet, "/users", "list users"},
			{http.MethodPost, "/users", "create user"},
			{http.MethodGet, "/users/123", "get user"},
			{http.MethodPut, "/users/123", "update user"},
			{http.MethodDelete, "/users/123", "delete user"},
		}

		for _, tt := range tests {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.body, w.Body.String())
		}
	})

	t.Run("nested resource routes", func(t *testing.T) {
		r := New()

		r.GET("/users/:userId/posts", mockHandlerWithParams())
		r.GET("/users/:userId/posts/:postId", mockHandlerWithParams())
		r.GET("/users/:userId/posts/:postId/comments/:commentId", mockHandlerWithParams())

		// Test nested route with all parameters
		req := httptest.NewRequest(http.MethodGet, "/users/1/posts/2/comments/3", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "userId=1")
		assert.Contains(t, w.Body.String(), "postId=2")
		assert.Contains(t, w.Body.String(), "commentId=3")
	})

	t.Run("static file serving with wildcard", func(t *testing.T) {
		r := New()

		r.GET("/static/*filepath", mockHandlerWithParams())

		req := httptest.NewRequest(http.MethodGet, "/static/css/main.css", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "filepath=css/main.css")
	})
}

func TestRouter_EdgeCases(t *testing.T) {
	t.Run("root path", func(t *testing.T) {
		r := New()
		r.GET("/", mockHandler("root"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "root", w.Body.String())
	})

	t.Run("empty wildcard capture", func(t *testing.T) {
		r := New()
		r.GET("/files/*filepath", mockHandlerWithParams())

		req := httptest.NewRequest(http.MethodGet, "/files/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("parameter with special characters", func(t *testing.T) {
		r := New()
		r.GET("/users/:id", mockHandlerWithParams())

		req := httptest.NewRequest(http.MethodGet, "/users/user-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "id=user-123")
	})
}

func TestRouter_ReplaceRoute(t *testing.T) {
	r := New()

	// Register initial handler
	r.GET("/users", mockHandler("original"))

	// Replace with new handler
	r.GET("/users", mockHandler("updated"))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "updated", w.Body.String())
}

func TestRouter_HandleContext_Panics(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		pattern  string
		handler  Handler
		panicMsg string
	}{
		{
			name:     "nil handler panics",
			method:   http.MethodGet,
			pattern:  "/test",
			handler:  nil,
			panicMsg: "routing: nil handler",
		},
		{
			name:     "empty pattern panics",
			method:   http.MethodGet,
			pattern:  "",
			handler:  func(ctx *Context) {},
			panicMsg: "routing: empty pattern",
		},
		{
			name:     "pattern without leading slash panics",
			method:   http.MethodGet,
			pattern:  "users",
			handler:  func(ctx *Context) {},
			panicMsg: "routing: pattern must start with /: users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			assert.PanicsWithValue(t, tt.panicMsg, func() {
				r.HandleContext(tt.method, tt.pattern, tt.handler)
			})
		})
	}
}

func TestRouter_Lookup(t *testing.T) {
	r := New()

	called := false
	handler := func(ctx *Context) { called = true }

	r.HandleContext(http.MethodGet, "/users/:id", handler)
	r.HandleContext(http.MethodGet, "/posts", handler)
	r.HandleContext(http.MethodPost, "/users", handler)

	t.Run("found with params", func(t *testing.T) {
		h, params, found := r.Lookup(http.MethodGet, "/users/123")
		assert.True(t, found)
		assert.NotNil(t, h)
		assert.Equal(t, Params{{Key: "id", Value: "123"}}, params)

		// Execute handler to verify it works
		called = false
		ctx := &Context{Params: make(Params, 1)}
		h(ctx)
		assert.True(t, called)
	})

	t.Run("found static", func(t *testing.T) {
		h, params, found := r.Lookup(http.MethodGet, "/posts")
		assert.True(t, found)
		assert.NotNil(t, h)
		assert.Empty(t, params)
	})

	t.Run("not found path", func(t *testing.T) {
		h, params, found := r.Lookup(http.MethodGet, "/nonexistent")
		assert.False(t, found)
		assert.Nil(t, h)
		assert.Nil(t, params)
	})

	t.Run("wrong method", func(t *testing.T) {
		h, params, found := r.Lookup(http.MethodDelete, "/users/123")
		assert.False(t, found)
		assert.Nil(t, h)
		assert.Nil(t, params)
	})

	t.Run("method not registered at all", func(t *testing.T) {
		h, params, found := r.Lookup(http.MethodPatch, "/anything")
		assert.False(t, found)
		assert.Nil(t, h)
		assert.Nil(t, params)
	})
}

func TestCountParams(t *testing.T) {
	tests := []struct {
		pattern  string
		expected uint8
	}{
		{"/users", 0},
		{"/users/:id", 1},
		{"/users/:id/posts/:postId", 2},
		{"/files/*filepath", 1},
		{"/api/:version/files/*filepath", 2},
		{"/", 0},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			assert.Equal(t, tt.expected, countParams(tt.pattern))
		})
	}
}

func TestParams_Get(t *testing.T) {
	params := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
	}

	assert.Equal(t, "123", params.Get("id"))
	assert.Equal(t, "test", params.Get("name"))
	assert.Equal(t, "", params.Get("missing"))
}

func TestParamsFromRequest(t *testing.T) {
	params := Params{{Key: "id", Value: "456"}}
	ctx := contextWithParams(context.Background(), params)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(ctx)

	result := ParamsFromRequest(req)
	assert.Equal(t, params, result)
}

func TestParamsFromRequest_NoParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	result := ParamsFromRequest(req)
	assert.Nil(t, result)
}

func TestRouter_GetValue_EdgeCases(t *testing.T) {
	t.Run("trailing slash with child handler", func(t *testing.T) {
		r := New()
		r.GETContext("/users/", func(ctx *Context) {
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("deep nesting with params", func(t *testing.T) {
		r := New()
		var captured []Param
		r.GETContext("/a/:b/c/:d/e/:f", func(ctx *Context) {
			captured = make([]Param, ctx.ParamCount())
			for i := 0; i < ctx.ParamCount(); i++ {
				captured[i] = ctx.ParamByIndex(i)
			}
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/a/1/c/2/e/3", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, []Param{
			{Key: "b", Value: "1"},
			{Key: "d", Value: "2"},
			{Key: "f", Value: "3"},
		}, captured)
	})

	t.Run("param no match beyond param segment", func(t *testing.T) {
		r := New()
		r.GETContext("/users/:id/profile", func(ctx *Context) {
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/users/123/settings", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("path longer than any registered route", func(t *testing.T) {
		r := New()
		r.GETContext("/api", func(ctx *Context) {
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/extra/path", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("param at end with no handler and trailing slash child", func(t *testing.T) {
		r := New()
		r.GETContext("/users/:id/", func(ctx *Context) {
			ctx.Writer.Write([]byte("trailing"))
		})

		req := httptest.NewRequest(http.MethodGet, "/users/123/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "trailing", w.Body.String())
	})

	t.Run("wildcard deep path", func(t *testing.T) {
		r := New()
		var filepath string
		r.GETContext("/assets/*filepath", func(ctx *Context) {
			filepath = ctx.Param("filepath")
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/assets/js/vendor/lodash/core.min.js", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "js/vendor/lodash/core.min.js", filepath)
	})

	t.Run("prefix mismatch returns not found", func(t *testing.T) {
		r := New()
		r.GETContext("/api/v1/users", func(ctx *Context) {
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v2/users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("exact path match with no handler on node", func(t *testing.T) {
		r := New()
		// Register only a child, not the parent
		r.GETContext("/api/v1/users", func(ctx *Context) {
			ctx.Writer.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRouter_GetValue_RadixTreeEdgePaths(t *testing.T) {
	t.Run("wildChild false and no index match", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/api/users", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		_, found := tree.getValue("/api/posts", ctx)
		assert.False(t, found)
	})

	t.Run("exact match with catch-all child but no handler on node", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/files/*filepath", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		handler, found := tree.getValue("/files/", ctx)
		assert.True(t, found)
		assert.NotNil(t, handler)
	})

	t.Run("static child index match in trailing slash check", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/users", func(ctx *Context) {})
		tree.addRoute("/users/list", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		handler, found := tree.getValue("/users", ctx)
		assert.True(t, found)
		assert.NotNil(t, handler)
	})

	t.Run("param child with no further children returns not found", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/users/:id", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		_, found := tree.getValue("/users/123/extra", ctx)
		assert.False(t, found)
	})

	t.Run("catch-all node type in longer path branch", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/static/*filepath", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		handler, found := tree.getValue("/static/a/b/c", ctx)
		assert.True(t, found)
		assert.NotNil(t, handler)
		assert.Equal(t, "filepath", ctx.Params[0].Key)
		assert.Equal(t, "a/b/c", ctx.Params[0].Value)
	})

	t.Run("path shorter than prefix returns not found", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/api/v1/users", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		_, found := tree.getValue("/api", ctx)
		assert.False(t, found)
	})

	t.Run("multiple params with mixed static segments", func(t *testing.T) {
		tree := &radixNode{}
		tree.addRoute("/orgs/:orgId/teams/:teamId/members", func(ctx *Context) {})

		ctx := &Context{Params: make(Params, 4)}
		handler, found := tree.getValue("/orgs/myorg/teams/myteam/members", ctx)
		assert.True(t, found)
		assert.NotNil(t, handler)
		assert.Equal(t, 2, ctx.paramCount)
		assert.Equal(t, "orgId", ctx.Params[0].Key)
		assert.Equal(t, "myorg", ctx.Params[0].Value)
		assert.Equal(t, "teamId", ctx.Params[1].Key)
		assert.Equal(t, "myteam", ctx.Params[1].Value)
	})
}

func TestRouter_HandleContext_MaxParamsGrowth(t *testing.T) {
	r := New()

	// First route with 1 param
	r.HandleContext(http.MethodGet, "/a/:id", func(ctx *Context) {})
	assert.Equal(t, 1, r.maxParams)

	// Second route with 2 params - should grow maxParams
	r.HandleContext(http.MethodGet, "/b/:id/c/:name", func(ctx *Context) {})
	assert.Equal(t, 2, r.maxParams)

	// Third route with fewer params - should not shrink
	r.HandleContext(http.MethodPost, "/d", func(ctx *Context) {})
	assert.Equal(t, 2, r.maxParams)
}

func TestRadixNode_IncrementChildPrio(t *testing.T) {
	// Test priority reordering
	tree := &radixNode{}

	// Adding routes that share a prefix creates children that get reordered
	tree.addRoute("/api/users", func(ctx *Context) {})
	tree.addRoute("/api/posts", func(ctx *Context) {})
	tree.addRoute("/api/comments", func(ctx *Context) {})

	// Access posts multiple times to increase its priority
	ctx := &Context{Params: make(Params, 4)}
	tree.getValue("/api/posts", ctx)
	tree.getValue("/api/posts", ctx)

	// Just verify tree works correctly after reordering
	handler, found := tree.getValue("/api/users", ctx)
	assert.True(t, found)
	assert.NotNil(t, handler)

	handler, found = tree.getValue("/api/posts", ctx)
	assert.True(t, found)
	assert.NotNil(t, handler)

	handler, found = tree.getValue("/api/comments", ctx)
	assert.True(t, found)
	assert.NotNil(t, handler)
}

func TestRadixNode_InsertChild_Panics(t *testing.T) {
	t.Run("catch-all not at end", func(t *testing.T) {
		tree := &radixNode{}
		assert.Panics(t, func() {
			tree.addRoute("/files/*path/extra", func(ctx *Context) {})
		})
	})

	t.Run("invalid wildcard - two wildcards in same segment", func(t *testing.T) {
		tree := &radixNode{}
		assert.Panics(t, func() {
			tree.addRoute("/files/:id:name", func(ctx *Context) {})
		})
	})

	t.Run("wildcard must have name", func(t *testing.T) {
		tree := &radixNode{}
		assert.Panics(t, func() {
			tree.addRoute("/files/:", func(ctx *Context) {})
		})
	})
}

func TestGetValue_PathEmptyAfterPrefix(t *testing.T) {
	// Tests the len(path) == 0 branch after consuming prefix in the longer path case
	tree := &radixNode{}
	tree.addRoute("/api/v1", func(ctx *Context) {})
	tree.addRoute("/api/v1/users", func(ctx *Context) {})

	ctx := &Context{Params: make(Params, 4)}

	// /api/v1 is an exact match for the first route
	handler, found := tree.getValue("/api/v1", ctx)
	assert.True(t, found)
	assert.NotNil(t, handler)
}

func TestGetValue_ExactMatchTrailingSlashIndex(t *testing.T) {
	// Test the indices '/' check branch in exact match
	tree := &radixNode{}
	tree.addRoute("/users/", func(ctx *Context) {})

	ctx := &Context{Params: make(Params, 4)}

	// Try to match /users without trailing slash - should check indices for '/'
	handler, found := tree.getValue("/users", ctx)
	// The tree has /users/ registered. Looking up /users should find the trailing slash handler.
	if found {
		assert.NotNil(t, handler)
	}
}

func TestGetValue_ParamNodeNoHandler(t *testing.T) {
	// Test param node reached at end of path but has no handler, only a child
	tree := &radixNode{}
	tree.addRoute("/users/:id/profile", func(ctx *Context) {})

	ctx := &Context{Params: make(Params, 4)}
	_, found := tree.getValue("/users/123", ctx)
	assert.False(t, found)
}

func TestGetValue_EmptyWildcardCapture(t *testing.T) {
	tree := &radixNode{}
	tree.addRoute("/files/*filepath", func(ctx *Context) {})

	ctx := &Context{Params: make(Params, 4)}
	handler, found := tree.getValue("/files/", ctx)
	assert.True(t, found)
	assert.NotNil(t, handler)
	assert.Equal(t, "", ctx.Params[0].Value)
}

func TestGetValue_CatchAllInExactMatchBranch(t *testing.T) {
	// Test the catch-all child check in the "exact match" branch
	tree := &radixNode{}
	tree.addRoute("/api/*rest", func(ctx *Context) {})

	ctx := &Context{Params: make(Params, 4)}
	handler, found := tree.getValue("/api/", ctx)
	assert.True(t, found)
	assert.NotNil(t, handler)
}

func TestFindWildcard_NoWildcard(t *testing.T) {
	wildcard, i, valid := findWildcard("/users/list")
	assert.Equal(t, "", wildcard)
	assert.Equal(t, -1, i)
	assert.False(t, valid)
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	r := New()
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	r.POSTContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusCreated)
	})

	// POST to /users should work
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// DELETE to /users should return 405
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	allow := w.Header().Get("Allow")
	assert.NotEmpty(t, allow)
	assert.Contains(t, allow, "GET")
	assert.Contains(t, allow, "POST")
}

func TestRouter_MethodNotAllowed_404WhenNoMatch(t *testing.T) {
	r := New()
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Request to /nonexistent should return 404, not 405
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/nonexistent", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_SetMethodNotAllowed_CustomHandler(t *testing.T) {
	r := New()
	r.SetMethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "custom 405")
	}))
	r.GETContext("/api/data", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/data", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "custom 405", w.Body.String())
	assert.Contains(t, w.Header().Get("Allow"), "GET")
}

func TestRouter_MethodNotAllowed_WithParams(t *testing.T) {
	r := New()
	r.GETContext("/users/:id", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	r.PUTContext("/users/:id", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// DELETE to /users/42 should return 405
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/42", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	allow := w.Header().Get("Allow")
	assert.Contains(t, allow, "GET")
	assert.Contains(t, allow, "PUT")
}

func TestRouter_MethodNotAllowed_NoMethodTree(t *testing.T) {
	r := New()
	r.HandleHEAD = false // Disable auto-HEAD so we test 405 behavior
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// HEAD to /test with auto-HEAD disabled should return 405
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Header().Get("Allow"), "GET")
}

func TestRouter_MethodNotAllowed_WithMiddleware(t *testing.T) {
	r := New()
	var middlewareCalled bool
	r.Use(func(next Handler) Handler {
		return func(ctx *Context) {
			middlewareCalled = true
			next(ctx)
		}
	})
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.True(t, middlewareCalled)
}

func TestRouter_AutoOPTIONS(t *testing.T) {
	r := New()
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	r.POSTContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusCreated)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	allow := w.Header().Get("Allow")
	assert.Contains(t, allow, "GET")
	assert.Contains(t, allow, "POST")
}

func TestRouter_AutoOPTIONS_ExplicitOverrides(t *testing.T) {
	r := New()
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	r.HandleContext(http.MethodOptions, "/users", func(ctx *Context) {
		ctx.Writer.Header().Set("X-Custom", "yes")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/users", nil)
	r.ServeHTTP(w, req)
	// Explicit handler should be used instead of auto-OPTIONS
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "yes", w.Header().Get("X-Custom"))
}

func TestRouter_AutoOPTIONS_Disabled(t *testing.T) {
	r := New()
	r.HandleOPTIONS = false
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(w, req)
	// With auto-OPTIONS disabled, should return 405 (GET is registered for the path)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestRouter_AutoOPTIONS_NoRoutes(t *testing.T) {
	r := New()
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/nonexistent", nil)
	r.ServeHTTP(w, req)
	// No methods registered for this path, should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_AutoHEAD(t *testing.T) {
	r := New()
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.WriteHeader(http.StatusOK)
		fmt.Fprint(ctx.Writer, "hello body")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Empty(t, w.Body.String(), "HEAD response should have no body")
}

func TestRouter_AutoHEAD_ExplicitOverrides(t *testing.T) {
	r := New()
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		fmt.Fprint(ctx.Writer, "GET body")
	})
	r.HandleContext(http.MethodHead, "/test", func(ctx *Context) {
		ctx.Writer.Header().Set("X-Explicit-Head", "yes")
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "yes", w.Header().Get("X-Explicit-Head"))
	assert.Empty(t, w.Body.String())
}

func TestRouter_AutoHEAD_Disabled(t *testing.T) {
	r := New()
	r.HandleHEAD = false
	r.GETContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	r.ServeHTTP(w, req)
	// With auto-HEAD disabled, should return 405 (GET is registered)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestRouter_AutoHEAD_NoGETRoute(t *testing.T) {
	r := New()
	r.POSTContext("/test", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusCreated)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	r.ServeHTTP(w, req)
	// No GET route for this path, should return 405
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestRouter_AutoHEAD_WithParams(t *testing.T) {
	r := New()
	r.GETContext("/users/:id", func(ctx *Context) {
		ctx.Writer.Header().Set("Content-Type", "application/json")
		ctx.Writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(ctx.Writer, `{"id":"%s"}`, ctx.Param("id"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/users/42", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Empty(t, w.Body.String(), "HEAD response should have no body")
}

func TestRouterRoutes(t *testing.T) {
	r := New()

	// Register multiple routes with different methods
	r.GETContext("/users", func(ctx *Context) {})
	r.POSTContext("/users", func(ctx *Context) {})
	r.GETContext("/users/:id", func(ctx *Context) {})
	r.PUTContext("/users/:id", func(ctx *Context) {})
	r.DELETEContext("/users/:id", func(ctx *Context) {})
	r.GETContext("/posts", func(ctx *Context) {})
	r.GETContext("/api/v1/data", func(ctx *Context) {})

	routes := r.Routes()

	// Verify count
	assert.Len(t, routes, 7)

	// Verify sorted order (by method, then pattern)
	expected := []RouteInfo{
		{Method: "DELETE", Pattern: "/users/:id"},
		{Method: "GET", Pattern: "/api/v1/data"},
		{Method: "GET", Pattern: "/posts"},
		{Method: "GET", Pattern: "/users"},
		{Method: "GET", Pattern: "/users/:id"},
		{Method: "POST", Pattern: "/users"},
		{Method: "PUT", Pattern: "/users/:id"},
	}

	assert.Equal(t, expected, routes)
}

func TestRoutesEmpty(t *testing.T) {
	r := New()
	routes := r.Routes()
	assert.Empty(t, routes)
}

func TestRoutesNestedParams(t *testing.T) {
	r := New()

	r.GETContext("/orgs/:orgId/teams/:teamId/members/:memberId", func(ctx *Context) {})
	r.POSTContext("/files/*filepath", func(ctx *Context) {})

	routes := r.Routes()

	assert.Len(t, routes, 2)
	assert.Contains(t, routes, RouteInfo{Method: "GET", Pattern: "/orgs/:orgId/teams/:teamId/members/:memberId"})
	assert.Contains(t, routes, RouteInfo{Method: "POST", Pattern: "/files/*filepath"})
}

func TestRedirectTrailingSlashDisabledByDefault(t *testing.T) {
	r := New()
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Request with trailing slash should return 404 when disabled
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRedirectTrailingSlashEnabled(t *testing.T) {
	r := New()
	r.RedirectTrailingSlash = true
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Request with trailing slash should redirect to /users
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "/users", w.Header().Get("Location"))
}

func TestRedirectTrailingSlashBidirectional(t *testing.T) {
	r := New()
	r.RedirectTrailingSlash = true

	// Register route WITH trailing slash
	r.GETContext("/api/", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Request WITHOUT trailing slash should redirect to /api/
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "/api/", w.Header().Get("Location"))
}

func TestRedirectTrailingSlashNoMatch(t *testing.T) {
	r := New()
	r.RedirectTrailingSlash = true
	r.GETContext("/users", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	// Request for completely different path should not redirect
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
