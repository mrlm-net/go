package deepmerge

import "sync"

// dedupMapPool provides reusable maps for SliceUnion deduplication to reduce allocations.
var dedupMapPool = sync.Pool{
	New: func() any {
		return make(map[any]struct{}, 16)
	},
}

// MergeableSlice wraps a slice with merge capabilities.
type MergeableSlice[T comparable] struct {
	items    []T
	strategy SliceStrategy
}

// NewSlice creates a MergeableSlice with the given items.
// Uses SliceOverride strategy by default.
func NewSlice[T comparable](items []T) MergeableSlice[T] {
	return MergeableSlice[T]{
		items:    items,
		strategy: SliceOverride,
	}
}

// WithStrategy returns a copy with the specified strategy.
func (s MergeableSlice[T]) WithStrategy(strategy SliceStrategy) MergeableSlice[T] {
	return MergeableSlice[T]{
		items:    s.items,
		strategy: strategy,
	}
}

// Items returns the underlying slice.
func (s MergeableSlice[T]) Items() []T {
	return s.items
}

// IsZero implements Mergeable - returns true if items is nil.
func (s MergeableSlice[T]) IsZero() bool {
	return s.items == nil
}

// Merge implements Mergeable with strategy-based merging.
func (s MergeableSlice[T]) Merge(other MergeableSlice[T]) MergeableSlice[T] {
	// If other is zero (nil), keep base
	if other.IsZero() {
		return s
	}

	// Apply merge strategy
	var result []T
	switch s.strategy {
	case SliceOverride:
		result = other.items
	case SliceAppend:
		result = s.mergeAppend(other.items)
	case SliceUnion:
		result = s.mergeUnion(other.items)
	default:
		result = other.items
	}

	return MergeableSlice[T]{
		items:    result,
		strategy: s.strategy,
	}
}

// mergeAppend concatenates base and override slices with pre-allocation.
func (s MergeableSlice[T]) mergeAppend(override []T) []T {
	// If base is nil, return override
	if s.items == nil {
		return override
	}

	// Pre-allocate capacity for optimal performance
	result := make([]T, 0, len(s.items)+len(override))
	result = append(result, s.items...)
	result = append(result, override...)
	return result
}

// mergeUnion creates a slice with unique values using O(n) map-based deduplication.
// Preserves order: base items first, then new override items.
// Uses sync.Pool to reduce allocations for the deduplication map.
func (s MergeableSlice[T]) mergeUnion(override []T) []T {
	// If base is nil, return override
	if s.items == nil {
		return override
	}

	totalHint := len(s.items) + len(override)

	// Get map from pool for deduplication (type-erased to work with generics).
	// Use defensive type assertion in case pool is contaminated.
	seenRaw, ok := dedupMapPool.Get().(map[any]struct{})
	if !ok || totalHint > 16 {
		// Pool returned wrong type, or pooled map capacity (16) is too small.
		// Allocate a right-sized map instead.
		seenRaw = make(map[any]struct{}, totalHint)
	}
	defer func() {
		clear(seenRaw)
		dedupMapPool.Put(seenRaw)
	}()

	result := make([]T, 0, totalHint)

	// Add base items first (preserves order).
	// Note: the any(item) boxing is unavoidable with generics since the pooled map
	// must use map[any]struct{} to be shared across all type instantiations.
	for _, item := range s.items {
		key := any(item)
		if _, exists := seenRaw[key]; !exists {
			seenRaw[key] = struct{}{}
			result = append(result, item)
		}
	}

	// Add override items that aren't duplicates
	for _, item := range override {
		key := any(item)
		if _, exists := seenRaw[key]; !exists {
			seenRaw[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// MergeSlices is a convenience function for merging two slices with a given strategy.
func MergeSlices[T comparable](strategy SliceStrategy, base, override []T) []T {
	baseSlice := NewSlice(base).WithStrategy(strategy)
	overrideSlice := NewSlice(override).WithStrategy(strategy)
	return baseSlice.Merge(overrideSlice).Items()
}
