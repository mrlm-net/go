package router

type NodeType uint8

const (
	STATIC NodeType = iota
	PARAM
	WILDCARD
)

type node[T any] struct {
	path          string
	indices       string
	children      []*node[T]
	handlers      map[string]HandlerFunc[T] // Handlers per execution method
	wildcardChild *node[T]                  // pointer to wildcard child, if any
	nType         NodeType
	priority      uint32
}

// addRoute adds a route to the node
func (n *node[T]) addRoute(path, method string, handler HandlerFunc[T]) {
	fullPath := path
	n.priority++

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, method, handler)
		n.nType = STATIC
		return
	}

walk:
	for {
		// Find longest common prefix
		i := longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := &node[T]{
				path:          n.path[i:],
				wildcardChild: n.wildcardChild,
				nType:         STATIC,
				indices:       n.indices,
				children:      n.children,
				handlers:      n.handlers,
				priority:      n.priority - 1,
			}

			n.children = []*node[T]{child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.handlers = make(map[string]HandlerFunc[T])
			n.wildcardChild = nil
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]

			c := path[0]

			// '/' after param
			if n.nType == PARAM && c == '/' && len(n.children) == 1 {
				n = n.children[0]
				n.priority++
				continue walk
			}

			// Check if a child with the next path byte exists
			for i, index := range []byte(n.indices) {
				if c == index {
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Otherwise insert it
			if c != ':' && c != '*' {
				n.indices += string([]byte{c})
				child := &node[T]{
					handlers: make(map[string]HandlerFunc[T]),
				}
				n.children = append(n.children, child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			}

			n.insertChild(path, method, handler)
			return
		}

		// Set handler for this node
		if n.handlers[method] != nil {
			panic("handler already registered for path '" + fullPath + "'")
		}
		n.handlers[method] = handler
		return
	}
}

// insertChild inserts a child node
func (n *node[T]) insertChild(path, method string, handler HandlerFunc[T]) {
	for {
		// Find prefix until first wildcard
		wildcard, i, valid := findWildcard(path)
		if i < 0 { // No wildcard found
			break
		}

		if !valid {
			panic("only one wildcard per path segment is allowed")
		}

		// Check if wildcard has a name
		if len(wildcard) < 2 {
			panic("wildcards must be named")
		}

		// Param
		if wildcard[0] == ':' {
			if i > 0 {
				n.path = path[:i]
				path = path[i:]
			}

			n.wildcardChild = nil
			child := &node[T]{
				nType:    PARAM,
				path:     wildcard,
				handlers: make(map[string]HandlerFunc[T]),
			}
			n.children = []*node[T]{child}
			n = child
			n.priority++

			// If the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node[T]{
					priority: 1,
					handlers: make(map[string]HandlerFunc[T]),
				}
				n.children = []*node[T]{child}
				n = child
				continue
			}

			// Otherwise we're done
			n.handlers[method] = handler
			return
		}

		// CatchAll
		if i+len(wildcard) != len(path) {
			panic("catch-all routes are only allowed at the end of the path")
		}

		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
			panic("catch-all conflicts with existing handle for the path segment root")
		}

		// Currently fixed width 1 for '/'
		i--
		if path[i] != '/' {
			panic("no / before catch-all")
		}

		n.path = path[:i]
		child := &node[T]{
			wildcardChild: nil,
			nType:         WILDCARD,
			priority:      1,
		}
		n.children = []*node[T]{child}
		n.indices = string('/')
		n = child
		n.priority++

		child = &node[T]{
			path:     path[i:],
			nType:    WILDCARD,
			handlers: make(map[string]HandlerFunc[T]),
			priority: 1,
		}
		n.children = []*node[T]{child}
		n = child

		n.handlers[method] = handler
		return
	}

	// No wildcard found, just set the handler
	n.path = path
	n.handlers[method] = handler
}

// incrementChildPrio increments the priority of the child at position i
func (n *node[T]) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Update indices
	if newPos != pos {
		n.indices = n.indices[:newPos] +
			n.indices[pos:pos+1] +
			n.indices[newPos:pos] + n.indices[pos+1:]
	}

	return newPos
}
