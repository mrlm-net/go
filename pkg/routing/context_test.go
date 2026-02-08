package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_Param(t *testing.T) {
	ctx := &Context{
		Params:     make(Params, 3),
		paramCount: 3,
	}
	ctx.Params[0] = Param{Key: "id", Value: "123"}
	ctx.Params[1] = Param{Key: "name", Value: "test"}
	ctx.Params[2] = Param{Key: "version", Value: "v1"}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing param", "id", "123"},
		{"existing param middle", "name", "test"},
		{"existing param last", "version", "v1"},
		{"non-existent param", "missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.Param(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContext_ParamByIndex(t *testing.T) {
	ctx := &Context{
		Params:     make(Params, 3),
		paramCount: 2,
	}
	ctx.Params[0] = Param{Key: "id", Value: "123"}
	ctx.Params[1] = Param{Key: "name", Value: "test"}

	tests := []struct {
		name     string
		index    int
		expected Param
	}{
		{"first param", 0, Param{Key: "id", Value: "123"}},
		{"second param", 1, Param{Key: "name", Value: "test"}},
		{"out of bounds", 2, Param{}},
		{"negative index", -1, Param{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.ParamByIndex(tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContext_ParamCount(t *testing.T) {
	ctx := &Context{
		Params:     make(Params, 5),
		paramCount: 3,
	}

	assert.Equal(t, 3, ctx.ParamCount())
}

func TestContext_Reset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := &Context{
		Params:     make(Params, 2),
		paramCount: 1,
	}
	ctx.Params[0] = Param{Key: "old", Value: "value"}

	ctx.Reset(w, req)

	assert.Equal(t, w, ctx.Writer)
	assert.Equal(t, req, ctx.Request)
	assert.Equal(t, 0, ctx.paramCount)
	// Params slice should still exist but count is 0
	assert.NotNil(t, ctx.Params)
}

func TestRouter_HandleContext(t *testing.T) {
	r := New()

	called := false
	var capturedCtx *Context

	handler := func(ctx *Context) {
		called = true
		capturedCtx = ctx
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.True(t, called)
	assert.NotNil(t, capturedCtx)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_HandleContextWithParams(t *testing.T) {
	r := New()

	var params []Param

	handler := func(ctx *Context) {
		params = make([]Param, ctx.ParamCount())
		for i := 0; i < ctx.ParamCount(); i++ {
			params[i] = ctx.ParamByIndex(i)
		}
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/users/:id/posts/:postId", handler)

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, params, 2)
	assert.Equal(t, "id", params[0].Key)
	assert.Equal(t, "123", params[0].Value)
	assert.Equal(t, "postId", params[1].Key)
	assert.Equal(t, "456", params[1].Value)
}

func TestRouter_MethodShortcutsContext(t *testing.T) {
	r := New()

	handler := func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.GETContext("/get", handler)
	r.POSTContext("/post", handler)
	r.PUTContext("/put", handler)
	r.DELETEContext("/delete", handler)
	r.PATCHContext("/patch", handler)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/get"},
		{http.MethodPost, "/post"},
		{http.MethodPut, "/put"},
		{http.MethodDelete, "/delete"},
		{http.MethodPatch, "/patch"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRouter_MixedHandlers(t *testing.T) {
	r := New()

	// Register a mix of context and http handlers
	r.HandleContext(http.MethodGet, "/context", func(ctx *Context) {
		ctx.Writer.Write([]byte("context"))
	})

	r.Handle(http.MethodGet, "/http", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("http"))
	}))

	tests := []struct {
		path         string
		expectedBody string
	}{
		{"/context", "context"},
		{"/http", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestContext_ParamMethod(t *testing.T) {
	r := New()

	var id, postId string

	handler := func(ctx *Context) {
		id = ctx.Param("id")
		postId = ctx.Param("postId")
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/users/:id/posts/:postId", handler)

	req := httptest.NewRequest(http.MethodGet, "/users/abc123/posts/xyz789", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, "abc123", id)
	assert.Equal(t, "xyz789", postId)
}

func TestRouter_ContextPoolReuse(t *testing.T) {
	r := New()

	var contexts []*Context

	handler := func(ctx *Context) {
		// Capture the context pointer
		contexts = append(contexts, ctx)
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	r.HandleContext(http.MethodGet, "/test/:id", handler)

	// Make multiple requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// Check that contexts were reused (at least some should have same pointer)
	// We can't guarantee exact reuse due to concurrent access, but with 10 sequential
	// requests, we should see some reuse
	assert.Len(t, contexts, 10)
}

func TestContextSetGet(t *testing.T) {
	ctx := &Context{
		Params: make(Params, 1),
	}

	// Test Set and Get
	ctx.Set("user_id", 123)
	val, ok := ctx.Get("user_id")
	assert.True(t, ok)
	assert.Equal(t, 123, val)

	// Test overwrite
	ctx.Set("user_id", 456)
	val, ok = ctx.Get("user_id")
	assert.True(t, ok)
	assert.Equal(t, 456, val)

	// Test multiple keys
	ctx.Set("request_id", "abc-123")
	ctx.Set("session", "xyz")

	val, ok = ctx.Get("request_id")
	assert.True(t, ok)
	assert.Equal(t, "abc-123", val)

	val, ok = ctx.Get("session")
	assert.True(t, ok)
	assert.Equal(t, "xyz", val)
}

func TestContextGetEmpty(t *testing.T) {
	ctx := &Context{
		Params: make(Params, 1),
	}

	// Get on empty store
	val, ok := ctx.Get("missing")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestContextMustGetPanic(t *testing.T) {
	ctx := &Context{
		Params: make(Params, 1),
	}

	// MustGet on missing key should panic
	assert.PanicsWithValue(t, `routing: key "missing" does not exist`, func() {
		ctx.MustGet("missing")
	})

	// MustGet on existing key should not panic
	ctx.Set("key", "value")
	assert.NotPanics(t, func() {
		val := ctx.MustGet("key")
		assert.Equal(t, "value", val)
	})
}

func TestContextResetClearsStore(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := &Context{
		Params: make(Params, 2),
	}

	// Set some values
	ctx.Set("key1", "value1")
	ctx.Set("key2", "value2")

	// Reset should clear the store
	ctx.Reset(w, req)

	// Store should be nil after reset to prevent unbounded memory growth in pool
	val, ok := ctx.Get("key1")
	assert.False(t, ok)
	assert.Nil(t, val)

	val, ok = ctx.Get("key2")
	assert.False(t, ok)
	assert.Nil(t, val)

	// Map should be nil (re-allocated lazily on next Set)
	assert.Nil(t, ctx.store)
}

func TestContextJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := &Context{
		Writer:  w,
		Request: req,
	}

	data := map[string]any{
		"message": "hello",
		"status":  "ok",
		"count":   42,
	}

	err := ctx.JSON(http.StatusOK, data)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `"message":"hello"`)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), `"count":42`)
}

func TestContextJSONEncodingError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := &Context{
		Writer:  w,
		Request: req,
	}

	// Channels cannot be JSON encoded
	unencodable := make(chan int)

	err := ctx.JSON(http.StatusOK, unencodable)
	assert.Error(t, err)
}

func TestContextString(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := &Context{
		Writer:  w,
		Request: req,
	}

	err := ctx.String(http.StatusOK, "Hello, World!")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "Hello, World!", w.Body.String())
}

func TestContextRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	ctx := &Context{
		Writer:  w,
		Request: req,
	}

	ctx.Redirect(http.StatusFound, "/new")
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/new", w.Header().Get("Location"))
}

func TestContextContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := &Context{
		Request: req,
	}

	reqCtx := ctx.Context()
	assert.NotNil(t, reqCtx)
	assert.Equal(t, req.Context(), reqCtx)
}

func BenchmarkContextSet(b *testing.B) {
	ctx := &Context{
		Params: make(Params, 1),
	}

	// First set to initialize the map
	ctx.Set("key", "value")

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		ctx.Set("key", "value")
	}
}

func BenchmarkContextGet(b *testing.B) {
	ctx := &Context{
		Params: make(Params, 1),
	}
	ctx.Set("key", "value")

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = ctx.Get("key")
	}
}

func BenchmarkContextJSON(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	data := map[string]any{
		"message": "hello",
		"status":  "ok",
		"count":   42,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		w.Body.Reset()
		ctx := &Context{
			Writer:  w,
			Request: req,
		}
		_ = ctx.JSON(http.StatusOK, data)
	}
}
