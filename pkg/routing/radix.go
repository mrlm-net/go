package routing

import (
	"fmt"
	"strings"
)

// nodeType represents the type of a radix tree node.
type nodeType uint8

const (
	static   nodeType = iota // static path segment
	root                     // root node
	param                    // :param segment
	catchAll                 // *wildcard segment
)

// radixNode represents a node in the compressed radix trie.
// It stores common prefixes in a single node, not character-by-character.
type radixNode struct {
	path      string       // edge label - the path segment from parent to this node
	indices   string       // first byte of each child's path for fast lookup
	children  []*radixNode // child nodes
	handler   Handler      // handler if this is an endpoint
	priority  uint32       // priority for child reordering (access frequency)
	nType     nodeType     // node type
	maxParams uint8        // max params in sub-tree
	wildChild bool         // true if a child is param or catchAll
	paramName string       // parameter name (for param/catchAll nodes)
}

// addRoute adds a route to the radix tree with prefix compression.
func (n *radixNode) addRoute(path string, handler Handler) {
	fullPath := path
	n.priority++

	// Empty tree - insert as root
	if n.path == "" && len(n.children) == 0 {
		n.insertChild(path, fullPath, handler)
		n.nType = root
		return
	}

	parentFullPathIndex := 0

walk:
	for {
		// Find longest common prefix
		i := longestCommonPrefix(path, n.path)

		// Split edge if necessary
		if i < len(n.path) {
			child := &radixNode{
				path:      n.path[i:],
				wildChild: n.wildChild,
				indices:   n.indices,
				children:  n.children,
				handler:   n.handler,
				priority:  n.priority - 1,
				nType:     static,
				maxParams: n.maxParams,
				paramName: n.paramName,
			}

			n.children = []*radixNode{child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.handler = nil
			n.wildChild = false
			n.paramName = ""
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// Check if a child with the next path byte exists
			if n.nType == param && c == '/' && len(n.children) == 1 {
				parentFullPathIndex += len(n.path)
				n = n.children[0]
				n.priority++
				continue walk
			}

			// Check for wildcard child
			for i, idx := range []byte(n.indices) {
				if c == idx {
					parentFullPathIndex += len(n.path)
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Insert new child
			if c != ':' && c != '*' && n.nType != catchAll {
				n.indices += string([]byte{c})
				child := &radixNode{
					maxParams: countParams(fullPath),
				}
				n.children = append(n.children, child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			} else if n.wildChild {
				// Wildcard child already exists
				n = n.children[len(n.children)-1]
				n.priority++

				// Check if wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					n.nType != catchAll &&
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				}

				// Wildcard conflict
				pathSeg := path
				if n.nType != catchAll {
					pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
				}
				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
				panic(fmt.Sprintf("pattern '%s' conflicts with existing pattern '%s'",
					fullPath, prefix))
			}

			n.insertChild(path, fullPath, handler)
			return
		}

		// Exact match - update handler (allow replacement)
		n.handler = handler
		return
	}
}

// insertChild inserts a child node with wildcard handling.
func (n *radixNode) insertChild(path, fullPath string, handler Handler) {
	for {
		// Find wildcard in path
		wildcard, i, valid := findWildcard(path)
		if i < 0 { // No wildcard found
			break
		}

		if !valid {
			panic(fmt.Sprintf("only one wildcard per path segment is allowed in path '%s'", fullPath))
		}

		// Check if this node has existing children
		if len(n.children) > 0 {
			panic(fmt.Sprintf("wildcard segment '%s' conflicts with existing children in path '%s'", wildcard, fullPath))
		}

		// Check if wildcard has a name
		if len(wildcard) < 2 {
			panic(fmt.Sprintf("wildcards must be named with a non-empty name in path '%s'", fullPath))
		}

		// Handle param (:name)
		if wildcard[0] == ':' {
			// Insert prefix before wildcard
			if i > 0 {
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &radixNode{
				nType:     param,
				path:      wildcard,
				paramName: wildcard[1:],
				maxParams: countParams(fullPath),
			}
			n.children = []*radixNode{child}
			n = child
			n.priority++

			// If this is the end of the path, set handler
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]

				child := &radixNode{
					maxParams: countParams(fullPath),
					priority:  1,
				}
				n.children = []*radixNode{child}
				n = child
				continue
			}

			n.handler = handler
			return
		}

		// Handle catch-all (*name)
		if wildcard[0] == '*' {
			// Catch-all must be at the end
			if i+len(wildcard) != len(path) {
				panic(fmt.Sprintf("catch-all routes are only allowed at the end of the path in path '%s'", fullPath))
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic(fmt.Sprintf("catch-all conflicts with existing handle for the path segment root in path '%s'", fullPath))
			}

			// Insert prefix before wildcard
			if i > 0 {
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &radixNode{
				nType:     catchAll,
				path:      path,
				paramName: wildcard[1:],
				maxParams: countParams(fullPath),
				priority:  1,
			}
			n.children = []*radixNode{child}
			n = child
			n.handler = handler
			return
		}
	}

	// No wildcard - insert as static node
	n.path = path
	n.handler = handler
	n.maxParams = countParams(fullPath)
}

// findWildcard finds the first wildcard in the path and returns:
// - wildcard: the wildcard segment (including : or *)
// - i: the index where the wildcard starts
// - valid: whether the wildcard is valid (only one wildcard per segment)
func findWildcard(path string) (wildcard string, i int, valid bool) {
	// Find start
	for start, c := range []byte(path) {
		// Wildcard start
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid wildcard
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

// incrementChildPrio increments the priority of the child at index i
// and reorders children by priority.
func (n *radixNode) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Reorder if necessary - move high-priority children to front
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Update indices if position changed
	if newPos != pos {
		n.indices = n.indices[:newPos] + n.indices[pos:pos+1] +
			n.indices[newPos:pos] + n.indices[pos+1:]
	}

	return newPos
}

// getValue retrieves a handler and extracts params in a single traversal.
func (n *radixNode) getValue(path string, ctx *Context) (handler Handler, found bool) {
	ctx.paramCount = 0

walk:
	for {
		prefix := n.path

		if len(path) > len(prefix) {
			if hasPrefix(path, prefix) {
				path = path[len(prefix):]

				// If path is now empty after matching prefix, check for handler
				if len(path) == 0 {
					if handler = n.handler; handler != nil {
						return handler, true
					}
					// Check if there's a catch-all child that can match empty
					if n.wildChild && len(n.children) > 0 {
						child := n.children[len(n.children)-1]
						if child.nType == catchAll {
							ctx.Params[ctx.paramCount].Key = child.paramName
							ctx.Params[ctx.paramCount].Value = ""
							ctx.paramCount++
							return child.handler, true
						}
					}
					return nil, false
				}

				// Try to find a child node
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}

				// Nothing found - check wildcard child
				if !n.wildChild {
					return nil, false
				}

				n = n.children[len(n.children)-1]

				// Handle param
				if n.nType == param {
					// Find end of param (next '/' or end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// Save param
					ctx.Params[ctx.paramCount].Key = n.paramName
					ctx.Params[ctx.paramCount].Value = path[:end]
					ctx.paramCount++

					// Continue with child if exists
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						return nil, false
					}

					if handler = n.handler; handler != nil {
						return handler, true
					}

					if len(n.children) == 1 {
						// No handler, but maybe trailing slash handler
						n = n.children[0]
						if n.path == "/" && n.handler != nil {
							return n.handler, true
						}
					}

					return nil, false
				}

				// Handle catch-all
				if n.nType == catchAll {
					ctx.Params[ctx.paramCount].Key = n.paramName
					ctx.Params[ctx.paramCount].Value = path
					ctx.paramCount++
					return n.handler, true
				}

				return nil, false
			}
		}

		// Exact match
		if len(path) == len(prefix) && hasPrefix(path, prefix) {
			if handler = n.handler; handler != nil {
				return handler, true
			}

			// Try to find index/default handler
			if path == "/" && n.wildChild && n.nType != root {
				return nil, false
			}

			// No handler - check for trailing slash
			for i, c := range []byte(n.indices) {
				if c == '/' {
					n = n.children[i]
					if n.path == "/" && n.handler != nil {
						return n.handler, true
					}
				}
			}

			// Check if there's a catch-all child that can match (for paths ending with /)
			if n.wildChild && len(n.children) > 0 {
				child := n.children[len(n.children)-1]
				if child.nType == catchAll {
					ctx.Params[ctx.paramCount].Key = child.paramName
					ctx.Params[ctx.paramCount].Value = ""
					ctx.paramCount++
					return child.handler, true
				}
			}

			return nil, false
		}

		return nil, false
	}
}

// hasPrefix checks if s starts with prefix using byte-by-byte comparison
// without creating a substring, avoiding allocations on the hot path.
func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

// longestCommonPrefix finds the longest common prefix of two strings.
func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}
