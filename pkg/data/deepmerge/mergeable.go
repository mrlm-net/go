package deepmerge

// Mergeable defines the contract for types that can be deeply merged.
type Mergeable[T any] interface {
	Merge(other T) T // Combine this with other, return result
	IsZero() bool    // True if value is unset/empty
}
