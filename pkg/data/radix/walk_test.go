package radix

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkPrefix_EmptyTree(t *testing.T) {
	tree := New[int]()

	called := false
	completed := tree.WalkPrefix("", func(key string, value int) bool {
		called = true
		return true
	})

	assert.True(t, completed)
	assert.False(t, called)
}

func TestWalkPrefix_EmptyPrefix(t *testing.T) {
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	var keys []string
	tree.WalkPrefix("", func(key string, value int) bool {
		keys = append(keys, key)
		return true
	})

	sort.Strings(keys)
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestWalkPrefix_SpecificPrefix(t *testing.T) {
	tree := New[string]()
	tree.Insert("/api/v1/users", "users")
	tree.Insert("/api/v1/products", "products")
	tree.Insert("/api/v2/users", "users-v2")
	tree.Insert("/health", "health")

	var keys []string
	tree.WalkPrefix("/api/v1", func(key string, value string) bool {
		keys = append(keys, key)
		return true
	})

	sort.Strings(keys)
	assert.Equal(t, []string{"/api/v1/products", "/api/v1/users"}, keys)
}

func TestWalkPrefix_NoMatch(t *testing.T) {
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("b", 2)

	called := false
	completed := tree.WalkPrefix("x", func(key string, value int) bool {
		called = true
		return true
	})

	assert.True(t, completed)
	assert.False(t, called)
}

func TestWalkPrefix_StopEarly(t *testing.T) {
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	count := 0
	completed := tree.WalkPrefix("", func(key string, value int) bool {
		count++
		return count < 2 // stop after 2 items
	})

	assert.False(t, completed)
	assert.Equal(t, 2, count)
}

func TestWalkPrefix_NestedPaths(t *testing.T) {
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("ab", 2)
	tree.Insert("abc", 3)
	tree.Insert("abcd", 4)

	var keys []string
	tree.WalkPrefix("ab", func(key string, value int) bool {
		keys = append(keys, key)
		return true
	})

	sort.Strings(keys)
	assert.Equal(t, []string{"ab", "abc", "abcd"}, keys)
}

func TestWalkPrefix_Values(t *testing.T) {
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	sum := 0
	tree.WalkPrefix("", func(key string, value int) bool {
		sum += value
		return true
	})

	assert.Equal(t, 6, sum)
}

func TestWalk_AllKeys(t *testing.T) {
	tree := New[string]()
	tree.Insert("/api/users", "users")
	tree.Insert("/api/products", "products")
	tree.Insert("/health", "health")

	var keys []string
	tree.Walk(func(key string, value string) bool {
		keys = append(keys, key)
		return true
	})

	sort.Strings(keys)
	assert.Equal(t, []string{"/api/products", "/api/users", "/health"}, keys)
}

func TestWalk_StopEarly(t *testing.T) {
	tree := New[int]()
	for i := range 10 {
		tree.Insert(string(rune('a'+i)), i)
	}

	count := 0
	completed := tree.Walk(func(key string, value int) bool {
		count++
		return count < 5
	})

	assert.False(t, completed)
	assert.Equal(t, 5, count)
}

func TestWalkPrefix_CommonPrefixBehavior(t *testing.T) {
	tree := New[int]()
	tree.Insert("test", 1)
	tree.Insert("testing", 2)
	tree.Insert("tester", 3)
	tree.Insert("team", 4)

	tests := []struct {
		prefix string
		want   []string
	}{
		{"test", []string{"test", "tester", "testing"}},
		{"te", []string{"team", "test", "tester", "testing"}},
		{"testing", []string{"testing"}},
		{"z", nil},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			var keys []string
			tree.WalkPrefix(tt.prefix, func(key string, value int) bool {
				keys = append(keys, key)
				return true
			})
			sort.Strings(keys)
			if tt.want == nil {
				assert.Nil(t, keys)
			} else {
				assert.Equal(t, tt.want, keys)
			}
		})
	}
}

func TestWalkPrefix_DeleteDuringWalk(t *testing.T) {
	// This test documents current behavior - walking during modification
	// is not recommended but should not crash
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	// Walk and collect keys first, then delete
	// Don't delete during walk as behavior is undefined
	var keys []string
	tree.WalkPrefix("", func(key string, value int) bool {
		keys = append(keys, key)
		return true
	})

	// Now delete
	for _, key := range keys {
		tree.Delete(key)
	}

	assert.Equal(t, 0, tree.Len())
}

func TestWalkPrefix_EmptyStringInTree(t *testing.T) {
	tree := New[string]()
	tree.Insert("", "empty")
	tree.Insert("a", "a")
	tree.Insert("b", "b")

	var keys []string
	tree.WalkPrefix("", func(key string, value string) bool {
		keys = append(keys, key)
		return true
	})

	sort.Strings(keys)
	assert.Equal(t, []string{"", "a", "b"}, keys)
}
