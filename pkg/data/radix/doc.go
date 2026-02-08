// Package radix provides a generic radix tree (prefix tree) implementation
// optimized for fast prefix-based lookups and routing use cases.
//
// A radix tree is a space-optimized trie where nodes with a single child are
// merged with their parent. This provides O(k) performance where k is the key
// length, making it ideal for routing, autocomplete, and prefix matching.
//
// # Core Operations
//
// The [Tree] type provides standard map-like operations plus specialized
// prefix operations:
//
//   - Insert: Add or update a key-value pair
//   - Get: Retrieve a value by exact key match
//   - Delete: Remove a key-value pair
//   - Has: Check if a key exists
//   - Len: Count of stored key-value pairs
//   - LongestPrefix: Find the longest matching prefix
//   - WalkPrefix: Iterate over all keys with a given prefix
//
// # Quick Start
//
//	// Create a new tree for routing
//	routes := radix.New[http.Handler]()
//
//	// Insert routes
//	routes.Insert("/api/users", usersHandler)
//	routes.Insert("/api/users/:id", userHandler)
//	routes.Insert("/api/products", productsHandler)
//
//	// Find exact match
//	if handler, found := routes.Get("/api/users"); found {
//	    handler.ServeHTTP(w, r)
//	}
//
//	// Find longest prefix match for routing
//	prefix, handler, found := routes.LongestPrefix("/api/users/123")
//	// prefix: "/api/users/:id", handler: userHandler, found: true
//
//	// Walk all routes under /api
//	routes.WalkPrefix("/api", func(key string, handler http.Handler) bool {
//	    fmt.Printf("Route: %s\n", key)
//	    return true // continue walking
//	})
//
// # Performance
//
// All operations have O(k) time complexity where k is the key length.
// Space complexity is O(n*k) where n is the number of keys, but the radix
// structure provides significant compression for keys with common prefixes.
//
// The implementation is optimized for:
//   - Fast lookups with minimal allocations
//   - Efficient prefix matching for routing
//   - Low memory overhead through path compression
//
// # Thread Safety
//
// Tree is not thread-safe. External synchronization is required for
// concurrent access.
package radix
