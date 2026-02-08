package radix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tree := New[string]()
	assert.NotNil(t, tree)
	assert.Equal(t, 0, tree.Len())
}

func TestInsert_SingleKey(t *testing.T) {
	tree := New[int]()

	old, replaced := tree.Insert("key", 42)
	assert.Equal(t, 0, old)
	assert.False(t, replaced)
	assert.Equal(t, 1, tree.Len())
}

func TestInsert_Replace(t *testing.T) {
	tree := New[int]()

	tree.Insert("key", 42)
	old, replaced := tree.Insert("key", 100)

	assert.Equal(t, 42, old)
	assert.True(t, replaced)
	assert.Equal(t, 1, tree.Len())

	val, found := tree.Get("key")
	assert.True(t, found)
	assert.Equal(t, 100, val)
}

func TestInsert_MultipleKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want int
	}{
		{
			name: "no common prefix",
			keys: []string{"apple", "banana", "cherry"},
			want: 3,
		},
		{
			name: "common prefix",
			keys: []string{"test", "testing", "tester"},
			want: 3,
		},
		{
			name: "nested prefix",
			keys: []string{"a", "ab", "abc", "abcd"},
			want: 4,
		},
		{
			name: "routing paths",
			keys: []string{"/api/users", "/api/users/:id", "/api/products", "/health"},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New[int]()
			for i, key := range tt.keys {
				tree.Insert(key, i)
			}
			assert.Equal(t, tt.want, tree.Len())

			// Verify all keys are retrievable
			for i, key := range tt.keys {
				val, found := tree.Get(key)
				assert.True(t, found, "key %s not found", key)
				assert.Equal(t, i, val, "wrong value for key %s", key)
			}
		})
	}
}

func TestInsert_EmptyKey(t *testing.T) {
	tree := New[string]()

	tree.Insert("", "empty")
	assert.Equal(t, 1, tree.Len())

	val, found := tree.Get("")
	assert.True(t, found)
	assert.Equal(t, "empty", val)
}

func TestGet_NotFound(t *testing.T) {
	tree := New[int]()
	tree.Insert("key", 42)

	val, found := tree.Get("notfound")
	assert.False(t, found)
	assert.Equal(t, 0, val)
}

func TestGet_EmptyTree(t *testing.T) {
	tree := New[int]()

	val, found := tree.Get("key")
	assert.False(t, found)
	assert.Equal(t, 0, val)
}

func TestGet_PartialMatch(t *testing.T) {
	tree := New[int]()
	tree.Insert("testing", 42)

	// "test" is a prefix but not a key
	val, found := tree.Get("test")
	assert.False(t, found)
	assert.Equal(t, 0, val)
}

func TestDelete_ExistingKey(t *testing.T) {
	tree := New[int]()
	tree.Insert("key", 42)

	old, deleted := tree.Delete("key")
	assert.True(t, deleted)
	assert.Equal(t, 42, old)
	assert.Equal(t, 0, tree.Len())

	_, found := tree.Get("key")
	assert.False(t, found)
}

func TestDelete_NonExistentKey(t *testing.T) {
	tree := New[int]()
	tree.Insert("key", 42)

	old, deleted := tree.Delete("notfound")
	assert.False(t, deleted)
	assert.Equal(t, 0, old)
	assert.Equal(t, 1, tree.Len())
}

func TestDelete_EmptyTree(t *testing.T) {
	tree := New[int]()

	old, deleted := tree.Delete("key")
	assert.False(t, deleted)
	assert.Equal(t, 0, old)
}

func TestDelete_WithCommonPrefix(t *testing.T) {
	tree := New[int]()
	tree.Insert("test", 1)
	tree.Insert("testing", 2)
	tree.Insert("tester", 3)

	old, deleted := tree.Delete("testing")
	assert.True(t, deleted)
	assert.Equal(t, 2, old)
	assert.Equal(t, 2, tree.Len())

	// Other keys should still exist
	val, found := tree.Get("test")
	assert.True(t, found)
	assert.Equal(t, 1, val)

	val, found = tree.Get("tester")
	assert.True(t, found)
	assert.Equal(t, 3, val)
}

func TestHas(t *testing.T) {
	tree := New[int]()
	tree.Insert("key", 42)

	assert.True(t, tree.Has("key"))
	assert.False(t, tree.Has("notfound"))
	assert.False(t, tree.Has("k"))
}

func TestLen(t *testing.T) {
	tree := New[int]()
	assert.Equal(t, 0, tree.Len())

	tree.Insert("a", 1)
	assert.Equal(t, 1, tree.Len())

	tree.Insert("b", 2)
	assert.Equal(t, 2, tree.Len())

	tree.Insert("a", 3) // replace
	assert.Equal(t, 2, tree.Len())

	tree.Delete("a")
	assert.Equal(t, 1, tree.Len())

	tree.Delete("b")
	assert.Equal(t, 0, tree.Len())
}

func TestLongestPrefix_ExactMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert("/api/users", "users")
	tree.Insert("/api/products", "products")

	prefix, val, found := tree.LongestPrefix("/api/users")
	assert.True(t, found)
	assert.Equal(t, "/api/users", prefix)
	assert.Equal(t, "users", val)
}

func TestLongestPrefix_PartialMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert("/api", "api")
	tree.Insert("/api/users", "users")

	prefix, val, found := tree.LongestPrefix("/api/users/123")
	assert.True(t, found)
	assert.Equal(t, "/api/users", prefix)
	assert.Equal(t, "users", val)
}

func TestLongestPrefix_RootMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert("/", "root")
	tree.Insert("/api", "api")

	prefix, val, found := tree.LongestPrefix("/unknown/path")
	assert.True(t, found)
	assert.Equal(t, "/", prefix)
	assert.Equal(t, "root", val)
}

func TestLongestPrefix_NoMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert("/api", "api")

	prefix, val, found := tree.LongestPrefix("/other")
	assert.False(t, found)
	assert.Equal(t, "", prefix)
	assert.Equal(t, "", val)
}

func TestLongestPrefix_EmptyTree(t *testing.T) {
	tree := New[string]()

	prefix, val, found := tree.LongestPrefix("/api")
	assert.False(t, found)
	assert.Equal(t, "", prefix)
	assert.Equal(t, "", val)
}

func TestLongestPrefix_EmptyKey(t *testing.T) {
	tree := New[string]()
	tree.Insert("", "root")
	tree.Insert("a", "a")

	prefix, val, found := tree.LongestPrefix("abc")
	assert.True(t, found)
	assert.Equal(t, "a", prefix)
	assert.Equal(t, "a", val)
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"identical", "test", "test", 4},
		{"no common", "abc", "xyz", 0},
		{"partial", "testing", "tester", 4},
		{"one empty", "test", "", 0},
		{"both empty", "", "", 0},
		{"prefix", "test", "testing", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestCommonPrefix(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name string
		s    string
		p    string
		want bool
	}{
		{"exact match", "test", "test", true},
		{"has prefix", "testing", "test", true},
		{"no prefix", "test", "testing", false},
		{"empty prefix", "test", "", true},
		{"empty string", "", "test", false},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasPrefix(tt.s, tt.p)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRouting demonstrates real-world routing use case
func TestRouting(t *testing.T) {
	type route struct {
		pattern string
		handler string
	}

	routes := []route{
		{"/", "home"},
		{"/api", "api-root"},
		{"/api/v1", "api-v1"},
		{"/api/v1/users", "users-list"},
		{"/api/v1/products", "products-list"},
		{"/health", "health-check"},
	}

	tree := New[string]()
	for _, r := range routes {
		tree.Insert(r.pattern, r.handler)
	}

	tests := []struct {
		path   string
		want   string
		wantOk bool
	}{
		{"/", "home", true},
		{"/api/v1/users", "users-list", true},
		{"/api/v1/users/123", "users-list", true},         // longest prefix is /api/v1/users
		{"/api/v1/users/123/profile", "users-list", true}, // longest prefix is /api/v1/users
		{"/health", "health-check", true},
		{"/unknown", "home", true},    // falls back to root
		{"/api/v2", "api-root", true}, // falls back to /api
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			_, handler, found := tree.LongestPrefix(tt.path)
			require.Equal(t, tt.wantOk, found)
			if found {
				assert.Equal(t, tt.want, handler)
			}
		})
	}
}

// TestSequentialOperations ensures operations work correctly in sequence
func TestSequentialOperations(t *testing.T) {
	tree := New[int]()

	// Build tree with unique keys
	keys := make([]string, 100)
	for i := range 100 {
		key := string(rune('a'+i%26)) + string(rune('0'+i/26))
		keys[i] = key
		tree.Insert(key, i)
	}

	assert.Equal(t, 100, tree.Len())

	// Read from tree
	for i, key := range keys {
		val, found := tree.Get(key)
		assert.True(t, found, "key %s not found", key)
		assert.Equal(t, i, val)
	}

	// Delete first 50 from tree
	for i := range 50 {
		_, deleted := tree.Delete(keys[i])
		assert.True(t, deleted)
	}

	assert.Equal(t, 50, tree.Len())

	// Verify remaining 50 still exist
	for i := 50; i < 100; i++ {
		val, found := tree.Get(keys[i])
		assert.True(t, found)
		assert.Equal(t, i, val)
	}
}
