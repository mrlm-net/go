package deepmerge

import "testing"

// TestOptionalImplementsMergeable verifies that Optional[T] implements Mergeable[Optional[T]]
// This is a compile-time check - if it compiles, the interface is implemented correctly.
func TestOptionalImplementsMergeable(t *testing.T) {
	var _ Mergeable[Optional[string]] = Optional[string]{}
	var _ Mergeable[Optional[int]] = Optional[int]{}
	var _ Mergeable[Optional[bool]] = Optional[bool]{}
	var _ Mergeable[Optional[float64]] = Optional[float64]{}

	// Runtime verification
	opt := Some("test")
	if opt.IsZero() {
		t.Error("Some() should not be zero")
	}

	none := None[string]()
	if !none.IsZero() {
		t.Error("None() should be zero")
	}

	merged := opt.Merge(none)
	if merged.Value() != "test" {
		t.Errorf("Expected 'test', got '%s'", merged.Value())
	}
}
