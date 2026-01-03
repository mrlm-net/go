package router

func New[T any]() *Router[T] {
	router := &Router[T]{
		// Radix tree nodes
		nodeTree: &node[T]{
			handlers: make(map[string]HandlerFunc[T]),
		},
		// Static hash map for quick static route lookup
		static: make(map[string]*node[T]),
	}

	router.pool.New = func() any {
		var zero T
		// Return pointer to zero value of T
		return &zero
	}

	return router
}

// Params represents path parameters that can be embedded in custom contexts
type Params []Param

// Param represents a path parameter
type Param struct {
	Key   string
	Value string
}
