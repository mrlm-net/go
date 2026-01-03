package router

import "sync"

type Router[T any] struct {
	pool     sync.Pool
	nodeTree *node[T]
	static   map[string]*node[T]
}

func (r *Router[T]) AddRoute(method, path string, handler HandlerFunc[T]) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}

	r.nodeTree.addRoute(path, method, handler)
}
