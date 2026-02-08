package radix

import (
	"fmt"
	"testing"
)

// Benchmark scaling tests with different data sizes
// Tests insert, get, delete, and walk performance across 10, 100, 1000, and 10000 keys

// BenchmarkInsert_10 benchmarks insert with 10 keys
func BenchmarkInsert_10(b *testing.B) {
	benchmarkInsert(b, 10)
}

// BenchmarkInsert_100 benchmarks insert with 100 keys
func BenchmarkInsert_100(b *testing.B) {
	benchmarkInsert(b, 100)
}

// BenchmarkInsert_1000 benchmarks insert with 1000 keys
func BenchmarkInsert_1000(b *testing.B) {
	benchmarkInsert(b, 1000)
}

// BenchmarkInsert_10000 benchmarks insert with 10000 keys
func BenchmarkInsert_10000(b *testing.B) {
	benchmarkInsert(b, 10000)
}

func benchmarkInsert(b *testing.B, size int) {
	keys := generateKeys(size)

	b.ResetTimer()
	for b.Loop() {
		tree := New[int]()
		for j, key := range keys {
			tree.Insert(key, j)
		}
	}
}

// BenchmarkGet_10 benchmarks get with 10 keys
func BenchmarkGet_10(b *testing.B) {
	benchmarkGet(b, 10)
}

// BenchmarkGet_100 benchmarks get with 100 keys
func BenchmarkGet_100(b *testing.B) {
	benchmarkGet(b, 100)
}

// BenchmarkGet_1000 benchmarks get with 1000 keys
func BenchmarkGet_1000(b *testing.B) {
	benchmarkGet(b, 1000)
}

// BenchmarkGet_10000 benchmarks get with 10000 keys
func BenchmarkGet_10000(b *testing.B) {
	benchmarkGet(b, 10000)
}

func benchmarkGet(b *testing.B, size int) {
	keys := generateKeys(size)
	tree := New[int]()
	for i, key := range keys {
		tree.Insert(key, i)
	}

	i := 0
	b.ResetTimer()
	for b.Loop() {
		key := keys[i%len(keys)]
		_, _ = tree.Get(key)
		i++
	}
}

// BenchmarkDelete_10 benchmarks delete with 10 keys
func BenchmarkDelete_10(b *testing.B) {
	benchmarkDelete(b, 10)
}

// BenchmarkDelete_100 benchmarks delete with 100 keys
func BenchmarkDelete_100(b *testing.B) {
	benchmarkDelete(b, 100)
}

// BenchmarkDelete_1000 benchmarks delete with 1000 keys
func BenchmarkDelete_1000(b *testing.B) {
	benchmarkDelete(b, 1000)
}

// BenchmarkDelete_10000 benchmarks delete with 10000 keys
func BenchmarkDelete_10000(b *testing.B) {
	benchmarkDelete(b, 10000)
}

func benchmarkDelete(b *testing.B, size int) {
	keys := generateKeys(size)

	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		tree := New[int]()
		for j, key := range keys {
			tree.Insert(key, j)
		}
		b.StartTimer()

		for _, key := range keys {
			tree.Delete(key)
		}
	}
}

// BenchmarkLongestPrefix_10 benchmarks longest prefix with 10 keys
func BenchmarkLongestPrefix_10(b *testing.B) {
	benchmarkLongestPrefix(b, 10)
}

// BenchmarkLongestPrefix_100 benchmarks longest prefix with 100 keys
func BenchmarkLongestPrefix_100(b *testing.B) {
	benchmarkLongestPrefix(b, 100)
}

// BenchmarkLongestPrefix_1000 benchmarks longest prefix with 1000 keys
func BenchmarkLongestPrefix_1000(b *testing.B) {
	benchmarkLongestPrefix(b, 1000)
}

// BenchmarkLongestPrefix_10000 benchmarks longest prefix with 10000 keys
func BenchmarkLongestPrefix_10000(b *testing.B) {
	benchmarkLongestPrefix(b, 10000)
}

func benchmarkLongestPrefix(b *testing.B, size int) {
	keys := generateKeys(size)
	tree := New[int]()
	for i, key := range keys {
		tree.Insert(key, i)
	}

	// Generate lookup keys with extra suffix
	lookupKeys := make([]string, len(keys))
	for i, key := range keys {
		lookupKeys[i] = key + "/extra/path"
	}

	i := 0
	b.ResetTimer()
	for b.Loop() {
		key := lookupKeys[i%len(lookupKeys)]
		_, _, _ = tree.LongestPrefix(key)
		i++
	}
}

// BenchmarkWalkPrefix_10 benchmarks walk prefix with 10 keys
func BenchmarkWalkPrefix_10(b *testing.B) {
	benchmarkWalkPrefix(b, 10)
}

// BenchmarkWalkPrefix_100 benchmarks walk prefix with 100 keys
func BenchmarkWalkPrefix_100(b *testing.B) {
	benchmarkWalkPrefix(b, 100)
}

// BenchmarkWalkPrefix_1000 benchmarks walk prefix with 1000 keys
func BenchmarkWalkPrefix_1000(b *testing.B) {
	benchmarkWalkPrefix(b, 1000)
}

// BenchmarkWalkPrefix_10000 benchmarks walk prefix with 10000 keys
func BenchmarkWalkPrefix_10000(b *testing.B) {
	benchmarkWalkPrefix(b, 10000)
}

func benchmarkWalkPrefix(b *testing.B, size int) {
	keys := generateKeys(size)
	tree := New[int]()
	for i, key := range keys {
		tree.Insert(key, i)
	}

	b.ResetTimer()
	for b.Loop() {
		tree.WalkPrefix("", func(key string, value int) bool {
			return true
		})
	}
}

// BenchmarkRouting benchmarks realistic routing scenario
func BenchmarkRouting(b *testing.B) {
	// Create realistic route tree
	routes := []string{
		"/",
		"/api",
		"/api/v1",
		"/api/v1/users",
		"/api/v1/users/:id",
		"/api/v1/users/:id/posts",
		"/api/v1/users/:id/posts/:post_id",
		"/api/v1/products",
		"/api/v1/products/:id",
		"/api/v2",
		"/api/v2/users",
		"/health",
		"/metrics",
		"/static",
		"/static/js",
		"/static/css",
	}

	tree := New[string]()
	for _, route := range routes {
		tree.Insert(route, route)
	}

	// Realistic request paths
	paths := []string{
		"/api/v1/users/123",
		"/api/v1/users/456/posts",
		"/api/v1/users/789/posts/abc",
		"/api/v1/products/widget",
		"/health",
		"/static/js/app.js",
		"/api/v2/users",
	}

	i := 0
	b.ResetTimer()
	for b.Loop() {
		path := paths[i%len(paths)]
		_, _, _ = tree.LongestPrefix(path)
		i++
	}
}

// BenchmarkMixedOperations benchmarks realistic mixed workload
func BenchmarkMixedOperations(b *testing.B) {
	tree := New[int]()
	keys := generateKeys(100)

	// Pre-populate
	for i, key := range keys {
		tree.Insert(key, i)
	}

	i := 0
	b.ResetTimer()
	for b.Loop() {
		// 70% reads, 20% writes, 10% deletes
		op := i % 10
		key := keys[i%len(keys)]

		if op < 7 {
			// Read
			_, _ = tree.Get(key)
		} else if op < 9 {
			// Write
			tree.Insert(key, i)
		} else {
			// Delete then re-insert to maintain size
			tree.Delete(key)
			tree.Insert(key, i)
		}
		i++
	}
}

// BenchmarkCommonPrefixScenario benchmarks trees with many common prefixes
func BenchmarkCommonPrefixScenario(b *testing.B) {
	tree := New[int]()

	// Create keys with common prefixes like URL paths
	for i := range 100 {
		tree.Insert(fmt.Sprintf("/api/v1/users/%d", i), i)
		tree.Insert(fmt.Sprintf("/api/v1/products/%d", i), i)
		tree.Insert(fmt.Sprintf("/api/v2/users/%d", i), i)
	}

	lookupKey := "/api/v1/users/50"

	b.ResetTimer()
	for b.Loop() {
		_, _ = tree.Get(lookupKey)
	}
}

// BenchmarkMemoryEfficiency benchmarks space efficiency with common prefixes
func BenchmarkMemoryEfficiency(b *testing.B) {
	keys := make([]string, 1000)
	for i := range 1000 {
		// All keys share common prefix
		keys[i] = fmt.Sprintf("/very/long/common/prefix/that/compresses/well/item_%d", i)
	}

	b.ResetTimer()
	for b.Loop() {
		tree := New[int]()
		for j, key := range keys {
			tree.Insert(key, j)
		}
	}
}

// generateKeys creates a set of keys for benchmarking
func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := range n {
		// Generate keys with some common prefixes to simulate realistic usage
		prefix := fmt.Sprintf("/api/v%d", i%5)
		suffix := fmt.Sprintf("/resource/%d", i)
		keys[i] = prefix + suffix
	}
	return keys
}

// BenchmarkWorstCase benchmarks worst-case scenario with no common prefixes
func BenchmarkWorstCase_Insert_1000(b *testing.B) {
	// Generate completely unique keys with no common prefixes
	keys := make([]string, 1000)
	for i := range 1000 {
		keys[i] = fmt.Sprintf("%c%d", rune('a'+(i%26)), i)
	}

	b.ResetTimer()
	for b.Loop() {
		tree := New[int]()
		for j, key := range keys {
			tree.Insert(key, j)
		}
	}
}

// BenchmarkBestCase benchmarks best-case scenario with maximum prefix sharing
func BenchmarkBestCase_Insert_1000(b *testing.B) {
	// Generate keys with maximum prefix sharing
	keys := make([]string, 1000)
	basePrefix := "/very/long/shared/prefix/that/compresses/extremely/well"
	for i := range 1000 {
		keys[i] = fmt.Sprintf("%s/%d", basePrefix, i)
	}

	b.ResetTimer()
	for b.Loop() {
		tree := New[int]()
		for j, key := range keys {
			tree.Insert(key, j)
		}
	}
}
