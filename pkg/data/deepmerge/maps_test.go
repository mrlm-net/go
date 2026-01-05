package deepmerge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeableMap_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		mapVal   MergeableMap[string, string]
		expected bool
	}{
		{
			name:     "nil items is zero",
			mapVal:   MergeableMap[string, string]{items: nil},
			expected: true,
		},
		{
			name:     "empty map is not zero",
			mapVal:   NewMap(map[string]string{}),
			expected: false,
		},
		{
			name:     "populated map is not zero",
			mapVal:   NewMap(map[string]string{"a": "1", "b": "2"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mapVal.IsZero())
		})
	}
}

func TestMergeableMap_Items(t *testing.T) {
	items := map[string]string{"a": "1", "b": "2"}
	mapVal := NewMap(items)

	result := mapVal.Items()
	assert.Equal(t, items, result)
}

func TestMergeableMap_WithStrategy(t *testing.T) {
	mapVal := NewMap(map[string]string{"a": "1"})

	// Default strategy should be MapOverride
	assert.Equal(t, MapOverride, mapVal.strategy)

	// WithStrategy should return new instance with updated strategy
	updated := mapVal.WithStrategy(MapMerge)
	assert.Equal(t, MapMerge, updated.strategy)

	// Original should be unchanged
	assert.Equal(t, MapOverride, mapVal.strategy)
}

func TestMergeableMap_Merge_Override(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		expected map[string]string
	}{
		{
			name:     "override replaces base",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"c": "3", "d": "4"},
			expected: map[string]string{"c": "3", "d": "4"},
		},
		{
			name:     "nil override keeps base",
			base:     map[string]string{"a": "1", "b": "2"},
			override: nil,
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "empty override replaces base",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "nil base with override",
			base:     nil,
			override: map[string]string{"a": "1", "b": "2"},
			expected: map[string]string{"a": "1", "b": "2"},
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
			base := NewMap(tt.base).WithStrategy(MapOverride)
			override := NewMap(tt.override).WithStrategy(MapOverride)

			result := base.Merge(override)
			assert.Equal(t, tt.expected, result.Items())
		})
	}
}

func TestMergeableMap_Merge_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		expected map[string]string
	}{
		{
			name:     "merge combines maps",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"c": "3", "d": "4"},
			expected: map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"},
		},
		{
			name:     "override wins on conflict",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"b": "99", "c": "3"},
			expected: map[string]string{"a": "1", "b": "99", "c": "3"},
		},
		{
			name:     "nil override keeps base",
			base:     map[string]string{"a": "1", "b": "2"},
			override: nil,
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "empty override keeps base",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "nil base with override",
			base:     nil,
			override: map[string]string{"a": "1", "b": "2"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "both empty",
			base:     map[string]string{},
			override: map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "all keys conflict",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"a": "10", "b": "20"},
			expected: map[string]string{"a": "10", "b": "20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewMap(tt.base).WithStrategy(MapMerge)
			override := NewMap(tt.override).WithStrategy(MapMerge)

			result := base.Merge(override)
			assert.Equal(t, tt.expected, result.Items())
		})
	}
}

func TestMergeableMap_DifferentTypes(t *testing.T) {
	t.Run("int keys and values", func(t *testing.T) {
		base := NewMap(map[int]int{1: 10, 2: 20}).WithStrategy(MapMerge)
		override := NewMap(map[int]int{2: 99, 3: 30}).WithStrategy(MapMerge)

		result := base.Merge(override)
		expected := map[int]int{1: 10, 2: 99, 3: 30}
		assert.Equal(t, expected, result.Items())
	})

	t.Run("string keys with int values", func(t *testing.T) {
		base := NewMap(map[string]int{"a": 1, "b": 2}).WithStrategy(MapMerge)
		override := NewMap(map[string]int{"b": 99, "c": 3}).WithStrategy(MapMerge)

		result := base.Merge(override)
		expected := map[string]int{"a": 1, "b": 99, "c": 3}
		assert.Equal(t, expected, result.Items())
	})

	t.Run("struct values", func(t *testing.T) {
		type Value struct {
			Name string
			Age  int
		}

		base := NewMap(map[string]Value{
			"user1": {Name: "Alice", Age: 30},
		}).WithStrategy(MapMerge)

		override := NewMap(map[string]Value{
			"user2": {Name: "Bob", Age: 25},
		}).WithStrategy(MapMerge)

		result := base.Merge(override)
		expected := map[string]Value{
			"user1": {Name: "Alice", Age: 30},
			"user2": {Name: "Bob", Age: 25},
		}
		assert.Equal(t, expected, result.Items())
	})
}

func TestMergeMaps_ConvenienceFunction(t *testing.T) {
	tests := []struct {
		name     string
		strategy MapStrategy
		base     map[string]string
		override map[string]string
		expected map[string]string
	}{
		{
			name:     "override strategy",
			strategy: MapOverride,
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"c": "3", "d": "4"},
			expected: map[string]string{"c": "3", "d": "4"},
		},
		{
			name:     "merge strategy",
			strategy: MapMerge,
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"b": "99", "c": "3"},
			expected: map[string]string{"a": "1", "b": "99", "c": "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeMaps(tt.strategy, tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkMergeableMap_Merge(b *testing.B) {
	base := make(map[string]string, 1000)
	override := make(map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		key := string(rune('a' + (i % 26)))
		base[key] = "value"
		override[key] = "value"
	}

	baseMap := NewMap(base).WithStrategy(MapMerge)
	overrideMap := NewMap(override).WithStrategy(MapMerge)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseMap.Merge(overrideMap)
	}
}

func BenchmarkMergeableMap_Override(b *testing.B) {
	base := make(map[string]string, 1000)
	override := make(map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		key := string(rune('a' + (i % 26)))
		base[key] = "value"
		override[key] = "value"
	}

	baseMap := NewMap(base).WithStrategy(MapOverride)
	overrideMap := NewMap(override).WithStrategy(MapOverride)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseMap.Merge(overrideMap)
	}
}
