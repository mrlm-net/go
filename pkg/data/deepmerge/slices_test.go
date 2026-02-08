package deepmerge

import (
	"reflect"
	"testing"
)

func TestMergeableSlice_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		slice    MergeableSlice[string]
		expected bool
	}{
		{
			name:     "nil items is zero",
			slice:    MergeableSlice[string]{items: nil},
			expected: true,
		},
		{
			name:     "empty slice is not zero",
			slice:    NewSlice([]string{}),
			expected: false,
		},
		{
			name:     "populated slice is not zero",
			slice:    NewSlice([]string{"a", "b"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.slice.IsZero(); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestMergeableSlice_Items(t *testing.T) {
	items := []string{"a", "b", "c"}
	slice := NewSlice(items)

	result := slice.Items()
	if !reflect.DeepEqual(items, result) {
		t.Errorf("expected %v, got %v", items, result)
	}
}

func TestMergeableSlice_WithStrategy(t *testing.T) {
	slice := NewSlice([]string{"a"})

	// Default strategy should be SliceOverride
	if slice.strategy != SliceOverride {
		t.Errorf("expected %v, got %v", SliceOverride, slice.strategy)
	}

	// WithStrategy should return new instance with updated strategy
	updated := slice.WithStrategy(SliceAppend)
	if updated.strategy != SliceAppend {
		t.Errorf("expected %v, got %v", SliceAppend, updated.strategy)
	}

	// Original should be unchanged
	if slice.strategy != SliceOverride {
		t.Errorf("expected %v, got %v", SliceOverride, slice.strategy)
	}
}

func TestMergeableSlice_Merge_Override(t *testing.T) {
	tests := []struct {
		name     string
		base     []string
		override []string
		expected []string
	}{
		{
			name:     "override replaces base",
			base:     []string{"a", "b"},
			override: []string{"c", "d"},
			expected: []string{"c", "d"},
		},
		{
			name:     "nil override keeps base",
			base:     []string{"a", "b"},
			override: nil,
			expected: []string{"a", "b"},
		},
		{
			name:     "empty override replaces base",
			base:     []string{"a", "b"},
			override: []string{},
			expected: []string{},
		},
		{
			name:     "nil base with override",
			base:     nil,
			override: []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "both nil",
			base:     nil,
			override: nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewSlice(tt.base).WithStrategy(SliceOverride)
			override := NewSlice(tt.override).WithStrategy(SliceOverride)

			result := base.Merge(override)
			if !reflect.DeepEqual(tt.expected, result.Items()) {
				t.Errorf("expected %v, got %v", tt.expected, result.Items())
			}
		})
	}
}

func TestMergeableSlice_Merge_Append(t *testing.T) {
	tests := []struct {
		name     string
		base     []string
		override []string
		expected []string
	}{
		{
			name:     "append concatenates slices",
			base:     []string{"a", "b"},
			override: []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "nil override keeps base",
			base:     []string{"a", "b"},
			override: nil,
			expected: []string{"a", "b"},
		},
		{
			name:     "empty override keeps base",
			base:     []string{"a", "b"},
			override: []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "nil base with override",
			base:     nil,
			override: []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "both empty",
			base:     []string{},
			override: []string{},
			expected: []string{},
		},
		{
			name:     "append with duplicates",
			base:     []string{"a", "b"},
			override: []string{"b", "c"},
			expected: []string{"a", "b", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewSlice(tt.base).WithStrategy(SliceAppend)
			override := NewSlice(tt.override).WithStrategy(SliceAppend)

			result := base.Merge(override)
			if !reflect.DeepEqual(tt.expected, result.Items()) {
				t.Errorf("expected %v, got %v", tt.expected, result.Items())
			}
		})
	}
}

func TestMergeableSlice_Merge_Union(t *testing.T) {
	tests := []struct {
		name     string
		base     []string
		override []string
		expected []string
	}{
		{
			name:     "union removes duplicates",
			base:     []string{"a", "b"},
			override: []string{"b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "nil override keeps base",
			base:     []string{"a", "b"},
			override: nil,
			expected: []string{"a", "b"},
		},
		{
			name:     "empty override keeps base",
			base:     []string{"a", "b"},
			override: []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "nil base with override",
			base:     nil,
			override: []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "both empty",
			base:     []string{},
			override: []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			base:     []string{"a", "b"},
			override: []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "preserves order - base first",
			base:     []string{"c", "a"},
			override: []string{"b", "a"},
			expected: []string{"c", "a", "b"},
		},
		{
			name:     "duplicates within base",
			base:     []string{"a", "b", "a"},
			override: []string{"c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "duplicates within override",
			base:     []string{"a"},
			override: []string{"b", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewSlice(tt.base).WithStrategy(SliceUnion)
			override := NewSlice(tt.override).WithStrategy(SliceUnion)

			result := base.Merge(override)
			if !reflect.DeepEqual(tt.expected, result.Items()) {
				t.Errorf("expected %v, got %v", tt.expected, result.Items())
			}
		})
	}
}

func TestMergeableSlice_DifferentTypes(t *testing.T) {
	t.Run("int slices", func(t *testing.T) {
		base := NewSlice([]int{1, 2}).WithStrategy(SliceAppend)
		override := NewSlice([]int{3, 4}).WithStrategy(SliceAppend)

		result := base.Merge(override)
		if !reflect.DeepEqual([]int{1, 2, 3, 4}, result.Items()) {
			t.Errorf("expected %v, got %v", []int{1, 2, 3, 4}, result.Items())
		}
	})

	t.Run("int union", func(t *testing.T) {
		base := NewSlice([]int{1, 2}).WithStrategy(SliceUnion)
		override := NewSlice([]int{2, 3}).WithStrategy(SliceUnion)

		result := base.Merge(override)
		if !reflect.DeepEqual([]int{1, 2, 3}, result.Items()) {
			t.Errorf("expected %v, got %v", []int{1, 2, 3}, result.Items())
		}
	})
}

func TestMergeSlices_ConvenienceFunction(t *testing.T) {
	tests := []struct {
		name     string
		strategy SliceStrategy
		base     []string
		override []string
		expected []string
	}{
		{
			name:     "override strategy",
			strategy: SliceOverride,
			base:     []string{"a", "b"},
			override: []string{"c", "d"},
			expected: []string{"c", "d"},
		},
		{
			name:     "append strategy",
			strategy: SliceAppend,
			base:     []string{"a", "b"},
			override: []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "union strategy",
			strategy: SliceUnion,
			base:     []string{"a", "b"},
			override: []string{"b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeSlices(tt.strategy, tt.base, tt.override)
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func BenchmarkMergeableSlice_Append(b *testing.B) {
	base := make([]string, 1000)
	override := make([]string, 1000)
	for i := range base {
		base[i] = "item"
		override[i] = "item"
	}

	baseSlice := NewSlice(base).WithStrategy(SliceAppend)
	overrideSlice := NewSlice(override).WithStrategy(SliceAppend)

	b.ResetTimer()
	for b.Loop() {
		_ = baseSlice.Merge(overrideSlice)
	}
}

func BenchmarkMergeableSlice_Union(b *testing.B) {
	base := make([]string, 1000)
	override := make([]string, 1000)
	for i := range base {
		base[i] = "item"
		override[i] = "item"
	}

	baseSlice := NewSlice(base).WithStrategy(SliceUnion)
	overrideSlice := NewSlice(override).WithStrategy(SliceUnion)

	b.ResetTimer()
	for b.Loop() {
		_ = baseSlice.Merge(overrideSlice)
	}
}
