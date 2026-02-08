package radix

// Tree is a generic radix tree (prefix tree) for fast prefix-based lookups.
// It provides O(k) performance where k is the key length.
type Tree[V any] struct {
	root *node[V]
	size int
}

// node represents an internal node in the radix tree.
type node[V any] struct {
	// prefix is the edge label from parent to this node
	prefix string

	// value is the stored value if this node represents a complete key
	value V

	// isLeaf indicates if this node stores a value
	isLeaf bool

	// children are the outgoing edges
	children []*node[V]
}

// New creates a new empty radix tree.
func New[V any]() *Tree[V] {
	return &Tree[V]{
		root: &node[V]{},
	}
}

// Insert adds a key-value pair to the tree. If the key already exists,
// the old value is returned along with replaced=true.
func (t *Tree[V]) Insert(key string, value V) (old V, replaced bool) {
	if t.root == nil {
		t.root = &node[V]{}
	}

	n := t.root
	search := key

	for {
		// If search string is empty, we're at the target node
		if len(search) == 0 {
			if n.isLeaf {
				old = n.value
				replaced = true
			} else {
				t.size++
			}
			n.value = value
			n.isLeaf = true
			return old, replaced
		}

		// Find a child that shares a prefix with search
		childIdx := -1
		for i, child := range n.children {
			if len(child.prefix) > 0 && child.prefix[0] == search[0] {
				childIdx = i
				break
			}
		}

		// No matching child, create a new one
		if childIdx == -1 {
			newChild := &node[V]{
				prefix: search,
				value:  value,
				isLeaf: true,
			}
			n.children = append(n.children, newChild)
			t.size++
			return old, false
		}

		child := n.children[childIdx]

		// Find common prefix length
		commonLen := longestCommonPrefix(search, child.prefix)

		// If we match the entire child prefix, recurse down
		if commonLen == len(child.prefix) {
			n = child
			search = search[commonLen:]
			continue
		}

		// Need to split the child node
		splitNode := &node[V]{
			prefix:   child.prefix[:commonLen],
			children: make([]*node[V], 0, 2),
		}

		// Update child to have remaining prefix
		child.prefix = child.prefix[commonLen:]
		splitNode.children = append(splitNode.children, child)

		// Create new leaf for our search string or mark split as leaf
		if commonLen == len(search) {
			// The split point is our target
			splitNode.value = value
			splitNode.isLeaf = true
			t.size++
		} else {
			// Create a new child for remaining search
			newChild := &node[V]{
				prefix: search[commonLen:],
				value:  value,
				isLeaf: true,
			}
			splitNode.children = append(splitNode.children, newChild)
			t.size++
		}

		// Replace child with split node
		n.children[childIdx] = splitNode
		return old, false
	}
}

// Get retrieves the value for the given key. Returns (value, true) if found,
// or (zero, false) if not found.
func (t *Tree[V]) Get(key string) (value V, found bool) {
	if t.root == nil {
		return value, false
	}

	n := t.root
	search := key

	for len(search) > 0 {
		// Find child with matching prefix
		childFound := false
		for _, child := range n.children {
			if hasPrefix(search, child.prefix) {
				n = child
				search = search[len(child.prefix):]
				childFound = true
				break
			}
		}

		if !childFound {
			return value, false
		}
	}

	// If search is empty and this node is a leaf, we found it
	if n.isLeaf {
		return n.value, true
	}

	return value, false
}

// Delete removes the key from the tree. Returns (old value, true) if the key
// was found and deleted, or (zero, false) if not found.
func (t *Tree[V]) Delete(key string) (old V, deleted bool) {
	if t.root == nil {
		return old, false
	}

	var zero V
	var path []*node[V]
	var childIndices []int

	n := t.root
	search := key
	path = append(path, n)

	// Find the node to delete
	for len(search) > 0 {
		found := false
		for i, child := range n.children {
			if hasPrefix(search, child.prefix) {
				n = child
				path = append(path, n)
				childIndices = append(childIndices, i)
				search = search[len(child.prefix):]
				found = true
				break
			}
		}
		if !found {
			return old, false
		}
	}

	// Check if target node has a value
	if !n.isLeaf {
		return old, false
	}

	old = n.value
	n.value = zero
	n.isLeaf = false
	t.size--

	// Clean up the tree if needed
	// If the node has no children, remove it
	if len(n.children) == 0 && len(path) > 1 {
		parent := path[len(path)-2]
		idx := childIndices[len(childIndices)-1]
		parent.children = append(parent.children[:idx], parent.children[idx+1:]...)
	} else if len(n.children) == 1 && !n.isLeaf {
		// Merge with single child
		child := n.children[0]
		n.prefix = n.prefix + child.prefix
		n.value = child.value
		n.isLeaf = child.isLeaf
		n.children = child.children
	}

	return old, true
}

// Has returns true if the key exists in the tree.
func (t *Tree[V]) Has(key string) bool {
	_, found := t.Get(key)
	return found
}

// Len returns the number of key-value pairs stored in the tree.
func (t *Tree[V]) Len() int {
	return t.size
}

// LongestPrefix finds the longest prefix match for the given key.
// Returns (prefix, value, true) if a match is found, or ("", zero, false)
// if no prefix matches.
//
// This is useful for routing where you want to find the best matching route.
func (t *Tree[V]) LongestPrefix(key string) (prefix string, value V, found bool) {
	if t.root == nil {
		return "", value, false
	}

	n := t.root
	search := key
	var lastLeaf *node[V]
	var lastPrefix string

	for {
		// Check if current node is a leaf
		if n.isLeaf {
			lastLeaf = n
			lastPrefix = key[:len(key)-len(search)]
		}

		// If search is empty, return last found leaf
		if len(search) == 0 {
			break
		}

		// Find child with matching prefix
		foundChild := false
		for _, child := range n.children {
			if hasPrefix(search, child.prefix) {
				n = child
				search = search[len(child.prefix):]
				foundChild = true
				break
			}
		}

		if !foundChild {
			break
		}
	}

	if lastLeaf != nil {
		return lastPrefix, lastLeaf.value, true
	}

	return "", value, false
}

// longestCommonPrefix returns the length of the longest common prefix
// between two strings.
func longestCommonPrefix(a, b string) int {
	minLen := min(len(a), len(b))

	for i := range minLen {
		if a[i] != b[i] {
			return i
		}
	}

	return minLen
}

// hasPrefix checks if string s starts with prefix p.
func hasPrefix(s, p string) bool {
	if len(p) > len(s) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if s[i] != p[i] {
			return false
		}
	}
	return true
}
