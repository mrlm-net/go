package deepmerge

// Merger provides configurable deep merge operations.
type Merger[T Mergeable[T]] struct {
	config Config
}

// New creates a new Merger with the given options.
// Zero value config provides sensible defaults:
// - NilSkip: nil values are skipped
// - SliceOverride: override replaces base slice
// - MapOverride: override replaces base map
func New[T Mergeable[T]](opts ...Option) *Merger[T] {
	config := Config{
		NilBehavior:   NilSkip,
		SliceStrategy: SliceOverride,
		MapStrategy:   MapOverride,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &Merger[T]{
		config: config,
	}
}

// Merge combines base with one or more overrides, applying them in order.
// Returns a new merged value without modifying inputs.
// If override IsZero(), it is skipped and base is preserved.
func (m *Merger[T]) Merge(base T, overrides ...T) T {
	result := base

	for _, override := range overrides {
		// Skip zero overrides - preserve base
		if override.IsZero() {
			continue
		}

		// Merge non-zero override with current result
		result = result.Merge(override)
	}

	return result
}

// Config returns the current configuration (for inspection).
func (m *Merger[T]) Config() Config {
	return m.config
}

// Merge is a package-level convenience function using default options.
func Merge[T Mergeable[T]](base T, overrides ...T) T {
	merger := New[T]()
	return merger.Merge(base, overrides...)
}
