package deepmerge

// OptionalValue wraps a value with explicit presence tracking for any type,
// including non-comparable types like slices, maps, and structs.
//
// Unlike [Optional], OptionalValue has no comparable constraint, so it
// cannot be used as a map key. Use [Optional] for comparable types and
// OptionalValue when you need to wrap slices, maps, or complex structs.
//
// The zero value of OptionalValue[T] is equivalent to NoneValue[T]().
type OptionalValue[T any] struct {
	value T
	set   bool
}

// SomeValue creates an OptionalValue containing a value.
// The value is explicitly set, even if it equals the zero value of T.
func SomeValue[T any](v T) OptionalValue[T] {
	return OptionalValue[T]{
		value: v,
		set:   true,
	}
}

// NoneValue creates an empty OptionalValue with no value set.
// This is equivalent to the zero value OptionalValue[T]{}.
func NoneValue[T any]() OptionalValue[T] {
	return OptionalValue[T]{
		set: false,
	}
}

// Value returns the contained value.
// If the OptionalValue is not set, returns the zero value of T.
func (o OptionalValue[T]) Value() T {
	return o.value
}

// ValueOr returns the contained value if set, otherwise returns the provided default.
func (o OptionalValue[T]) ValueOr(def T) T {
	if o.set {
		return o.value
	}
	return def
}

// IsSet reports whether a value has been explicitly set.
func (o OptionalValue[T]) IsSet() bool {
	return o.set
}

// IsZero implements Mergeable - returns true if no value is set.
func (o OptionalValue[T]) IsZero() bool {
	return !o.set
}

// Merge implements Mergeable - combines two OptionalValue values.
// If the override is set, it wins. Otherwise the base is preserved.
func (o OptionalValue[T]) Merge(other OptionalValue[T]) OptionalValue[T] {
	if other.set {
		return other
	}
	return o
}
