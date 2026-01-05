package deepmerge

// Optional wraps a value with explicit presence tracking.
// This solves the zero-value detection problem for primitives.
//
// The zero value of Optional[T] is equivalent to None[T]() - an unset value.
// This makes the zero value useful according to Go principles.
type Optional[T comparable] struct {
	value T
	set   bool
}

// Some creates an Optional containing a value.
// The value is explicitly set, even if it equals the zero value of T.
func Some[T comparable](v T) Optional[T] {
	return Optional[T]{
		value: v,
		set:   true,
	}
}

// None creates an empty Optional with no value set.
// This is equivalent to the zero value Optional[T]{}.
func None[T comparable]() Optional[T] {
	return Optional[T]{
		set: false,
	}
}

// Value returns the contained value.
// If the Optional is not set, returns the zero value of T.
func (o Optional[T]) Value() T {
	return o.value
}

// ValueOr returns the contained value if set, otherwise returns the provided default.
// This is useful for providing fallback values.
func (o Optional[T]) ValueOr(def T) T {
	if o.set {
		return o.value
	}
	return def
}

// IsSet reports whether a value has been explicitly set.
// Returns true for Some(v) regardless of whether v is the zero value.
// Returns false for None() or the zero value Optional[T]{}.
func (o Optional[T]) IsSet() bool {
	return o.set
}

// IsZero implements Mergeable - returns true if no value is set.
// This is the inverse of IsSet() and allows Optional to participate
// in the deepmerge system.
func (o Optional[T]) IsZero() bool {
	return !o.set
}

// Merge implements Mergeable - combines two Optional values.
// If the override is set, it wins and replaces the base.
// If the override is not set, the base is preserved.
// This allows explicit overrides (including zero values) while
// preserving base values when no override is specified.
func (o Optional[T]) Merge(other Optional[T]) Optional[T] {
	if other.set {
		return other
	}
	return o
}
