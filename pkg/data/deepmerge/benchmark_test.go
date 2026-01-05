package deepmerge

import (
	"fmt"
	"testing"
)

// Benchmark scaling tests with different data sizes
// Tests slice append, union, and map merge performance across 10, 100, 1000, and 10000 items

// BenchmarkSliceAppend_10 benchmarks slice append with 10 items
func BenchmarkSliceAppend_10(b *testing.B) {
	benchmarkSliceAppend(b, 10)
}

// BenchmarkSliceAppend_100 benchmarks slice append with 100 items
func BenchmarkSliceAppend_100(b *testing.B) {
	benchmarkSliceAppend(b, 100)
}

// BenchmarkSliceAppend_1000 benchmarks slice append with 1000 items
func BenchmarkSliceAppend_1000(b *testing.B) {
	benchmarkSliceAppend(b, 1000)
}

// BenchmarkSliceAppend_10000 benchmarks slice append with 10000 items
func BenchmarkSliceAppend_10000(b *testing.B) {
	benchmarkSliceAppend(b, 10000)
}

func benchmarkSliceAppend(b *testing.B, size int) {
	base := make([]int, size)
	override := make([]int, size)
	for i := 0; i < size; i++ {
		base[i] = i
		override[i] = i + size
	}

	baseSlice := NewSlice(base).WithStrategy(SliceAppend)
	overrideSlice := NewSlice(override).WithStrategy(SliceAppend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseSlice.Merge(overrideSlice)
	}
}

// BenchmarkSliceUnion_10 benchmarks slice union with 10 items
func BenchmarkSliceUnion_10(b *testing.B) {
	benchmarkSliceUnion(b, 10)
}

// BenchmarkSliceUnion_100 benchmarks slice union with 100 items
func BenchmarkSliceUnion_100(b *testing.B) {
	benchmarkSliceUnion(b, 100)
}

// BenchmarkSliceUnion_1000 benchmarks slice union with 1000 items
func BenchmarkSliceUnion_1000(b *testing.B) {
	benchmarkSliceUnion(b, 1000)
}

// BenchmarkSliceUnion_10000 benchmarks slice union with 10000 items
func BenchmarkSliceUnion_10000(b *testing.B) {
	benchmarkSliceUnion(b, 10000)
}

func benchmarkSliceUnion(b *testing.B, size int) {
	base := make([]int, size)
	override := make([]int, size)
	// Create 50% overlap for realistic union scenario
	for i := 0; i < size; i++ {
		base[i] = i
		override[i] = i + size/2
	}

	baseSlice := NewSlice(base).WithStrategy(SliceUnion)
	overrideSlice := NewSlice(override).WithStrategy(SliceUnion)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseSlice.Merge(overrideSlice)
	}
}

// BenchmarkMapMerge_10 benchmarks map merge with 10 items
func BenchmarkMapMerge_10(b *testing.B) {
	benchmarkMapMerge(b, 10)
}

// BenchmarkMapMerge_100 benchmarks map merge with 100 items
func BenchmarkMapMerge_100(b *testing.B) {
	benchmarkMapMerge(b, 100)
}

// BenchmarkMapMerge_1000 benchmarks map merge with 1000 items
func BenchmarkMapMerge_1000(b *testing.B) {
	benchmarkMapMerge(b, 1000)
}

// BenchmarkMapMerge_10000 benchmarks map merge with 10000 items
func BenchmarkMapMerge_10000(b *testing.B) {
	benchmarkMapMerge(b, 10000)
}

func benchmarkMapMerge(b *testing.B, size int) {
	base := make(map[string]int, size)
	override := make(map[string]int, size)
	// Create 50% overlap for realistic merge scenario
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key_%d", i)
		base[key] = i
		if i >= size/2 {
			override[key] = i * 2
		}
	}

	baseMap := NewMap(base).WithStrategy(MapMerge)
	overrideMap := NewMap(override).WithStrategy(MapMerge)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseMap.Merge(overrideMap)
	}
}

// BenchmarkNestedConfig tests nested struct performance with realistic config merging
func BenchmarkNestedConfig(b *testing.B) {
	type NestedConfig struct {
		Name     Optional[string]
		Enabled  Optional[bool]
		Timeout  Optional[int]
		Tags     MergeableSlice[string]
		Metadata MergeableMap[string, string]
	}

	base := NestedConfig{
		Name:     Some("base-config"),
		Enabled:  Some(true),
		Timeout:  Some(30),
		Tags:     NewSlice([]string{"prod", "web", "api"}).WithStrategy(SliceAppend),
		Metadata: NewMap(map[string]string{"env": "production", "region": "us-east-1", "version": "1.0.0"}).WithStrategy(MapMerge),
	}

	override := NestedConfig{
		Timeout:  Some(60),
		Tags:     NewSlice([]string{"monitoring", "alerts"}).WithStrategy(SliceAppend),
		Metadata: NewMap(map[string]string{"region": "us-west-2", "tier": "premium"}).WithStrategy(MapMerge),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := NestedConfig{
			Name:     base.Name.Merge(override.Name),
			Enabled:  base.Enabled.Merge(override.Enabled),
			Timeout:  base.Timeout.Merge(override.Timeout),
			Tags:     base.Tags.Merge(override.Tags),
			Metadata: base.Metadata.Merge(override.Metadata),
		}
		_ = result
	}
}

// BenchmarkDeepNesting tests 5 levels deep nested structure performance
func BenchmarkDeepNesting(b *testing.B) {
	type Level5 struct {
		Value Optional[string]
		Data  MergeableMap[string, string]
	}

	type Level4 struct {
		Value Optional[int]
		Level Level5
	}

	type Level3 struct {
		Value Optional[bool]
		Level Level4
	}

	type Level2 struct {
		Value Optional[string]
		Tags  MergeableSlice[string]
		Level Level3
	}

	type Level1 struct {
		Name  Optional[string]
		Level Level2
	}

	base := Level1{
		Name: Some("root"),
		Level: Level2{
			Value: Some("level2-base"),
			Tags:  NewSlice([]string{"tag1", "tag2"}).WithStrategy(SliceUnion),
			Level: Level3{
				Value: Some(true),
				Level: Level4{
					Value: Some(100),
					Level: Level5{
						Value: Some("deep-value"),
						Data:  NewMap(map[string]string{"key1": "val1", "key2": "val2"}).WithStrategy(MapMerge),
					},
				},
			},
		},
	}

	override := Level1{
		Level: Level2{
			Tags: NewSlice([]string{"tag3", "tag4"}).WithStrategy(SliceUnion),
			Level: Level3{
				Level: Level4{
					Value: Some(200),
					Level: Level5{
						Data: NewMap(map[string]string{"key3": "val3"}).WithStrategy(MapMerge),
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Level1{
			Name: base.Name.Merge(override.Name),
			Level: Level2{
				Value: base.Level.Value.Merge(override.Level.Value),
				Tags:  base.Level.Tags.Merge(override.Level.Tags),
				Level: Level3{
					Value: base.Level.Level.Value.Merge(override.Level.Level.Value),
					Level: Level4{
						Value: base.Level.Level.Level.Value.Merge(override.Level.Level.Level.Value),
						Level: Level5{
							Value: base.Level.Level.Level.Level.Value.Merge(override.Level.Level.Level.Level.Value),
							Data:  base.Level.Level.Level.Level.Data.Merge(override.Level.Level.Level.Level.Data),
						},
					},
				},
			},
		}
		_ = result
	}
}

// BenchmarkSliceOverride_Comparison benchmarks override strategy for comparison
func BenchmarkSliceOverride_1000(b *testing.B) {
	size := 1000
	base := make([]int, size)
	override := make([]int, size)
	for i := 0; i < size; i++ {
		base[i] = i
		override[i] = i + size
	}

	baseSlice := NewSlice(base).WithStrategy(SliceOverride)
	overrideSlice := NewSlice(override).WithStrategy(SliceOverride)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseSlice.Merge(overrideSlice)
	}
}

// BenchmarkMapOverride_Comparison benchmarks override strategy for comparison
func BenchmarkMapOverride_1000(b *testing.B) {
	size := 1000
	base := make(map[string]int, size)
	override := make(map[string]int, size)
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key_%d", i)
		base[key] = i
		override[key] = i * 2
	}

	baseMap := NewMap(base).WithStrategy(MapOverride)
	overrideMap := NewMap(override).WithStrategy(MapOverride)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseMap.Merge(overrideMap)
	}
}

// BenchmarkOptional_Merge benchmarks Optional merging performance
func BenchmarkOptional_Merge(b *testing.B) {
	base := Some("base-value")
	override := Some("override-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.Merge(override)
	}
}

// BenchmarkOptional_None benchmarks Optional with None values
func BenchmarkOptional_None(b *testing.B) {
	base := Some("base-value")
	override := None[string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.Merge(override)
	}
}
