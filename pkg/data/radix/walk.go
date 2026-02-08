package radix

// WalkFunc is called for each key-value pair during tree traversal.
// Return false to stop the walk early, true to continue.
type WalkFunc[V any] func(key string, value V) bool

// walkState holds the shared buffer and callback for allocation-free walking.
type walkState[V any] struct {
	buf []byte
	fn  WalkFunc[V]
}

// walk recursively visits a node and its children, reusing buf for path construction.
func (ws *walkState[V]) walk(n *node[V]) bool {
	if n.isLeaf {
		if !ws.fn(string(ws.buf), n.value) {
			return false
		}
	}
	for _, child := range n.children {
		prevLen := len(ws.buf)
		ws.buf = append(ws.buf, child.prefix...)
		if !ws.walk(child) {
			return false
		}
		ws.buf = ws.buf[:prevLen]
	}
	return true
}

// WalkPrefix traverses all keys with the given prefix, calling fn for each.
// Returns true if the walk completed fully, false if stopped early by fn.
//
// Keys are visited in lexicographic order. The walk stops early if fn
// returns false.
func (t *Tree[V]) WalkPrefix(prefix string, fn WalkFunc[V]) bool {
	if t.root == nil {
		return true
	}

	// Find the node that matches the prefix
	n := t.root
	search := prefix
	pathLen := 0

	// Navigate to the prefix node
	for len(search) > 0 {
		found := false
		for _, child := range n.children {
			if hasPrefix(search, child.prefix) {
				n = child
				pathLen += len(child.prefix)
				search = search[len(child.prefix):]
				found = true
				break
			} else if hasPrefix(child.prefix, search) {
				// Prefix ends in the middle of a node - no exact match possible
				// but we can walk this subtree
				n = child
				pathLen += len(child.prefix)
				search = ""
				found = true
				break
			}
		}
		if !found {
			// Prefix not found in tree
			return true
		}
	}

	// Build the initial path buffer from prefix navigation
	ws := walkState[V]{
		buf: make([]byte, 0, pathLen+64),
		fn:  fn,
	}

	// Reconstruct path to current node by walking from root
	// We tracked pathLen, but we need actual bytes. Re-walk to fill buf.
	ws.buf = ws.buf[:0]
	cur := t.root
	remaining := prefix
	for len(remaining) > 0 {
		for _, child := range cur.children {
			if hasPrefix(remaining, child.prefix) {
				ws.buf = append(ws.buf, child.prefix...)
				remaining = remaining[len(child.prefix):]
				cur = child
				break
			} else if hasPrefix(child.prefix, remaining) {
				ws.buf = append(ws.buf, child.prefix...)
				remaining = ""
				cur = child
				break
			}
		}
	}

	return ws.walk(n)
}

// Walk traverses all keys in the tree, calling fn for each.
// Returns true if the walk completed fully, false if stopped early by fn.
//
// This is equivalent to WalkPrefix("", fn).
func (t *Tree[V]) Walk(fn WalkFunc[V]) bool {
	return t.WalkPrefix("", fn)
}
