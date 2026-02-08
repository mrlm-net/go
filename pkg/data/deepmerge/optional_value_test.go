package deepmerge

import (
	"testing"
)

func TestSomeValue(t *testing.T) {
	t.Run("slice value", func(t *testing.T) {
		opt := SomeValue([]string{"a", "b"})
		assertTrue(t, opt.IsSet(), "SomeValue should be set")
		assertEqual(t, []string{"a", "b"}, opt.Value())
	})

	t.Run("map value", func(t *testing.T) {
		opt := SomeValue(map[string]int{"a": 1})
		assertTrue(t, opt.IsSet(), "SomeValue should be set")
		assertEqual(t, map[string]int{"a": 1}, opt.Value())
	})

	t.Run("nil slice is set", func(t *testing.T) {
		opt := SomeValue([]int(nil))
		assertTrue(t, opt.IsSet(), "SomeValue(nil slice) should be set")
	})

	t.Run("comparable types work too", func(t *testing.T) {
		opt := SomeValue(42)
		assertTrue(t, opt.IsSet())
		assertEqual(t, 42, opt.Value())
	})
}

func TestNoneValue(t *testing.T) {
	t.Run("slice none", func(t *testing.T) {
		opt := NoneValue[[]string]()
		assertFalse(t, opt.IsSet(), "NoneValue should not be set")
		assertEqual(t, []string(nil), opt.Value())
	})

	t.Run("map none", func(t *testing.T) {
		opt := NoneValue[map[string]int]()
		assertFalse(t, opt.IsSet())
		assertEqual(t, map[string]int(nil), opt.Value())
	})
}

func TestOptionalValue_ValueOr(t *testing.T) {
	t.Run("set returns value", func(t *testing.T) {
		opt := SomeValue([]string{"a"})
		result := opt.ValueOr([]string{"default"})
		assertEqual(t, []string{"a"}, result)
	})

	t.Run("unset returns default", func(t *testing.T) {
		opt := NoneValue[[]string]()
		result := opt.ValueOr([]string{"default"})
		assertEqual(t, []string{"default"}, result)
	})
}

func TestOptionalValue_IsZero(t *testing.T) {
	t.Run("set is not zero", func(t *testing.T) {
		opt := SomeValue([]int{1, 2})
		assertFalse(t, opt.IsZero())
	})

	t.Run("none is zero", func(t *testing.T) {
		opt := NoneValue[[]int]()
		assertTrue(t, opt.IsZero())
	})

	t.Run("zero value struct is zero", func(t *testing.T) {
		var opt OptionalValue[[]int]
		assertTrue(t, opt.IsZero())
	})
}

func TestOptionalValue_Merge(t *testing.T) {
	t.Run("override set wins", func(t *testing.T) {
		base := SomeValue([]string{"a"})
		override := SomeValue([]string{"b"})
		result := base.Merge(override)
		assertEqual(t, []string{"b"}, result.Value())
		assertTrue(t, result.IsSet())
	})

	t.Run("override unset preserves base", func(t *testing.T) {
		base := SomeValue([]string{"a"})
		override := NoneValue[[]string]()
		result := base.Merge(override)
		assertEqual(t, []string{"a"}, result.Value())
		assertTrue(t, result.IsSet())
	})

	t.Run("base unset override set", func(t *testing.T) {
		base := NoneValue[[]string]()
		override := SomeValue([]string{"b"})
		result := base.Merge(override)
		assertEqual(t, []string{"b"}, result.Value())
		assertTrue(t, result.IsSet())
	})

	t.Run("both unset stays unset", func(t *testing.T) {
		base := NoneValue[[]string]()
		override := NoneValue[[]string]()
		result := base.Merge(override)
		assertFalse(t, result.IsSet())
	})

	t.Run("map merge", func(t *testing.T) {
		base := SomeValue(map[string]int{"a": 1})
		override := SomeValue(map[string]int{"b": 2})
		result := base.Merge(override)
		assertEqual(t, map[string]int{"b": 2}, result.Value())
	})
}

func TestOptionalValueImplementsMergeable(t *testing.T) {
	var _ Mergeable[OptionalValue[[]string]] = OptionalValue[[]string]{}
	var _ Mergeable[OptionalValue[map[string]int]] = OptionalValue[map[string]int]{}

	opt := SomeValue([]int{1, 2, 3})
	if opt.IsZero() {
		t.Error("SomeValue should not be zero")
	}

	none := NoneValue[[]int]()
	if !none.IsZero() {
		t.Error("NoneValue should be zero")
	}

	merged := opt.Merge(none)
	assertEqual(t, []int{1, 2, 3}, merged.Value())
}

func TestOptionalValue_ZeroValueUseful(t *testing.T) {
	var zero OptionalValue[[]string]
	none := NoneValue[[]string]()

	assertEqual(t, none.IsSet(), zero.IsSet())
	assertEqual(t, none.IsZero(), zero.IsZero())
}
