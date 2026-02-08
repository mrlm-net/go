// Package deepmerge provides generic deep merging for maps, slices, and
// custom types with configurable strategies and explicit presence tracking.
//
// The core abstraction is the [Mergeable] interface. Any type implementing
// Merge and IsZero can participate in the merge system:
//
//	type Mergeable[T any] interface {
//	    Merge(other T) T
//	    IsZero() bool
//	}
//
// Built-in mergeable types include [Optional], [MergeableSlice], and
// [MergeableMap]. Each supports configurable merge strategies.
//
// # Strategies
//
// Slice strategies: [SliceOverride] (default, replace), [SliceAppend]
// (concatenate), [SliceUnion] (deduplicate).
//
// Map strategies: [MapOverride] (default, replace), [MapMerge] (recursive
// key-by-key merge).
//
// Nil behavior: [NilSkip] (default, preserve base), [NilOverride] (nil
// replaces base).
//
// # Optional
//
// [Optional] solves the zero-value detection problem for comparable types.
// Some(0) is distinguishable from None[int](), enabling layered configuration
// where explicit zero values override defaults.
//
// [OptionalValue] provides the same semantics for non-comparable types
// (slices, maps, structs containing slices, etc.) at the cost of not being
// usable as map keys.
//
// # Quick Start
//
//	// Merge slices with union strategy
//	result := deepmerge.MergeSlices(deepmerge.SliceUnion,
//	    []string{"a", "b"}, []string{"b", "c"})
//	// result: ["a", "b", "c"]
//
//	// Optional for config layering
//	base := deepmerge.Some(8080)
//	override := deepmerge.None[int]()
//	port := base.Merge(override).Value() // 8080 — base preserved
package deepmerge
