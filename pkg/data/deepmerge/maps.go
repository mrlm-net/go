package deepmerge

// MergeableMap wraps a map with merge capabilities.
type MergeableMap[K comparable, V any] struct {
	items    map[K]V
	strategy MapStrategy
}

// NewMap creates a MergeableMap with the given items.
// Uses MapOverride strategy by default.
func NewMap[K comparable, V any](items map[K]V) MergeableMap[K, V] {
	return MergeableMap[K, V]{
		items:    items,
		strategy: MapOverride,
	}
}

// WithStrategy returns a copy with the specified strategy.
func (m MergeableMap[K, V]) WithStrategy(strategy MapStrategy) MergeableMap[K, V] {
	return MergeableMap[K, V]{
		items:    m.items,
		strategy: strategy,
	}
}

// Items returns the underlying map.
func (m MergeableMap[K, V]) Items() map[K]V {
	return m.items
}

// IsZero implements Mergeable - returns true if items is nil.
func (m MergeableMap[K, V]) IsZero() bool {
	return m.items == nil
}

// Merge implements Mergeable with strategy-based merging.
func (m MergeableMap[K, V]) Merge(other MergeableMap[K, V]) MergeableMap[K, V] {
	// If other is zero (nil), keep base
	if other.IsZero() {
		return m
	}

	// Apply merge strategy
	var result map[K]V
	switch m.strategy {
	case MapOverride:
		result = other.items
	case MapMerge:
		result = m.mergeMaps(other.items)
	default:
		result = other.items
	}

	return MergeableMap[K, V]{
		items:    result,
		strategy: m.strategy,
	}
}

// mergeMaps combines base and override maps with pre-allocation.
// Override values win on key conflicts.
func (m MergeableMap[K, V]) mergeMaps(override map[K]V) map[K]V {
	// If base is nil, return override
	if m.items == nil {
		return override
	}

	// Pre-allocate capacity for optimal performance
	result := make(map[K]V, len(m.items)+len(override))

	// Copy all base keys
	for k, v := range m.items {
		result[k] = v
	}

	// Apply override keys (wins on conflict)
	for k, v := range override {
		result[k] = v
	}

	return result
}

// MergeMaps is a convenience function for merging two maps with a given strategy.
func MergeMaps[K comparable, V any](strategy MapStrategy, base, override map[K]V) map[K]V {
	baseMap := NewMap(base).WithStrategy(strategy)
	overrideMap := NewMap(override).WithStrategy(strategy)
	return baseMap.Merge(overrideMap).Items()
}
