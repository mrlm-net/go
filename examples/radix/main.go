package main

import (
	"fmt"

	"github.com/mrlm-net/go/pkg/data/radix"
)

func main() {
	fmt.Println("Radix Tree Example")
	fmt.Println("==================")

	basicExample()
	routingExample()
	prefixWalkExample()
}

func basicExample() {
	fmt.Println("\n1. Basic Operations")
	fmt.Println("-------------------")

	tree := radix.New[string]()

	// Insert values
	tree.Insert("apple", "🍎")
	tree.Insert("application", "📱")
	tree.Insert("apply", "✓")
	tree.Insert("banana", "🍌")

	fmt.Printf("Tree size: %d\n", tree.Len())

	// Get values
	if val, found := tree.Get("apple"); found {
		fmt.Printf("apple: %s\n", val)
	}

	// Check existence
	fmt.Printf("Has 'app': %v\n", tree.Has("app"))
	fmt.Printf("Has 'apple': %v\n", tree.Has("apple"))

	// Delete
	if old, deleted := tree.Delete("apply"); deleted {
		fmt.Printf("Deleted 'apply' with value: %s\n", old)
	}

	fmt.Printf("Final size: %d\n", tree.Len())
}

func routingExample() {
	fmt.Println("\n2. HTTP Routing Simulation")
	fmt.Println("---------------------------")

	type handler struct {
		method string
		name   string
	}

	routes := radix.New[handler]()

	// Register routes
	routes.Insert("/", handler{"GET", "home"})
	routes.Insert("/api", handler{"GET", "api-root"})
	routes.Insert("/api/v1/users", handler{"GET", "list-users"})
	routes.Insert("/api/v1/products", handler{"GET", "list-products"})
	routes.Insert("/health", handler{"GET", "health-check"})
	routes.Insert("/static", handler{"GET", "static-files"})

	// Route incoming requests
	requests := []string{
		"/",
		"/api/v1/users",
		"/api/v1/users/123",
		"/api/v1/products/widget",
		"/health",
		"/static/css/app.css",
		"/unknown",
	}

	for _, path := range requests {
		if prefix, h, found := routes.LongestPrefix(path); found {
			fmt.Printf("%-30s -> %-15s (matched: %s)\n", path, h.name, prefix)
		} else {
			fmt.Printf("%-30s -> NOT FOUND\n", path)
		}
	}
}

func prefixWalkExample() {
	fmt.Println("\n3. Prefix Walking")
	fmt.Println("-----------------")

	dict := radix.New[string]()

	// Build a dictionary
	words := map[string]string{
		"test":     "a procedure for critical evaluation",
		"testing":  "the act of conducting tests",
		"tester":   "a person who tests",
		"tested":   "tried and proved",
		"team":     "a group of people",
		"tea":      "a beverage",
		"teach":    "to impart knowledge",
		"teacher":  "one who teaches",
		"teaching": "the profession of a teacher",
	}

	for word, def := range words {
		dict.Insert(word, def)
	}

	// Find all words starting with "test"
	fmt.Println("Words starting with 'test':")
	dict.WalkPrefix("test", func(word string, def string) bool {
		fmt.Printf("  %-10s - %s\n", word, def)
		return true
	})

	// Find all words starting with "tea"
	fmt.Println("\nWords starting with 'tea':")
	dict.WalkPrefix("tea", func(word string, def string) bool {
		fmt.Printf("  %-10s - %s\n", word, def)
		return true
	})

	// Count all words
	count := 0
	dict.Walk(func(word string, def string) bool {
		count++
		return true
	})
	fmt.Printf("\nTotal words in dictionary: %d\n", count)
}
