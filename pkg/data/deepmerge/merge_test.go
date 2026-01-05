package deepmerge

import (
	"testing"
)

// TestConfig is a test struct that implements Mergeable.
type TestConfig struct {
	Host     Optional[string]
	Port     Optional[int]
	Debug    Optional[bool]
	Tags     MergeableSlice[string]
	Settings MergeableMap[string, string]
}

// IsZero returns true if all fields are zero.
func (c TestConfig) IsZero() bool {
	return c.Host.IsZero() &&
		c.Port.IsZero() &&
		c.Debug.IsZero() &&
		c.Tags.IsZero() &&
		c.Settings.IsZero()
}

// Merge combines two TestConfig instances.
func (c TestConfig) Merge(other TestConfig) TestConfig {
	return TestConfig{
		Host:     c.Host.Merge(other.Host),
		Port:     c.Port.Merge(other.Port),
		Debug:    c.Debug.Merge(other.Debug),
		Tags:     c.Tags.Merge(other.Tags),
		Settings: c.Settings.Merge(other.Settings),
	}
}

func TestMerger_EmptyBaseAndOverride(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{}
	override := TestConfig{}

	result := merger.Merge(base, override)

	if !result.IsZero() {
		t.Error("Expected zero result when merging two empty configs")
	}
}

func TestMerger_SetBaseEmptyOverride(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override := TestConfig{}

	result := merger.Merge(base, override)

	if result.Host.Value() != "localhost" {
		t.Errorf("Expected Host=localhost, got %v", result.Host.Value())
	}
	if result.Port.Value() != 8080 {
		t.Errorf("Expected Port=8080, got %v", result.Port.Value())
	}
	if !result.Debug.Value() {
		t.Error("Expected Debug=true")
	}
}

func TestMerger_EmptyBaseSetOverride(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{}
	override := TestConfig{
		Host:  Some("example.com"),
		Port:  Some(443),
		Debug: Some(false),
	}

	result := merger.Merge(base, override)

	if result.Host.Value() != "example.com" {
		t.Errorf("Expected Host=example.com, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}
	if result.Debug.Value() {
		t.Error("Expected Debug=false")
	}
}

func TestMerger_SetBaseSetOverride(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override := TestConfig{
		Host: Some("example.com"),
		Port: Some(443),
	}

	result := merger.Merge(base, override)

	if result.Host.Value() != "example.com" {
		t.Errorf("Expected Host=example.com, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}
	if !result.Debug.Value() {
		t.Error("Expected Debug=true (preserved from base)")
	}
}

func TestMerger_MultipleOverrides(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override1 := TestConfig{
		Host: Some("staging.com"),
	}
	override2 := TestConfig{
		Port: Some(443),
	}
	override3 := TestConfig{
		Debug: Some(false),
	}

	result := merger.Merge(base, override1, override2, override3)

	if result.Host.Value() != "staging.com" {
		t.Errorf("Expected Host=staging.com, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}
	if result.Debug.Value() {
		t.Error("Expected Debug=false")
	}
}

func TestMerger_NestedMerging(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{
		Host:     Some("localhost"),
		Tags:     NewSlice([]string{"dev", "local"}),
		Settings: NewMap(map[string]string{"env": "dev", "region": "us"}),
	}
	override := TestConfig{
		Port:     Some(443),
		Tags:     NewSlice([]string{"prod", "remote"}),
		Settings: NewMap(map[string]string{"env": "prod"}),
	}

	result := merger.Merge(base, override)

	// Check Optional fields
	if result.Host.Value() != "localhost" {
		t.Errorf("Expected Host=localhost, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}

	// Check Slice - should override by default
	tags := result.Tags.Items()
	if len(tags) != 2 || tags[0] != "prod" || tags[1] != "remote" {
		t.Errorf("Expected Tags=[prod, remote], got %v", tags)
	}

	// Check Map - should override by default
	settings := result.Settings.Items()
	if len(settings) != 1 || settings["env"] != "prod" {
		t.Errorf("Expected Settings={env:prod}, got %v", settings)
	}
}

func TestMerger_WithOptions(t *testing.T) {
	merger := New[TestConfig](
		WithSliceStrategy(SliceAppend),
		WithMapStrategy(MapMerge),
	)

	base := TestConfig{
		Tags:     NewSlice([]string{"dev"}).WithStrategy(SliceAppend),
		Settings: NewMap(map[string]string{"env": "dev", "region": "us"}).WithStrategy(MapMerge),
	}
	override := TestConfig{
		Tags:     NewSlice([]string{"prod"}).WithStrategy(SliceAppend),
		Settings: NewMap(map[string]string{"env": "prod"}).WithStrategy(MapMerge),
	}

	result := merger.Merge(base, override)

	// Check Slice - should append
	tags := result.Tags.Items()
	if len(tags) != 2 || tags[0] != "dev" || tags[1] != "prod" {
		t.Errorf("Expected Tags=[dev, prod], got %v", tags)
	}

	// Check Map - should merge
	settings := result.Settings.Items()
	if len(settings) != 2 || settings["env"] != "prod" || settings["region"] != "us" {
		t.Errorf("Expected Settings={env:prod, region:us}, got %v", settings)
	}
}

func TestMerger_Config(t *testing.T) {
	merger := New[TestConfig](
		WithNilBehavior(NilOverride),
		WithSliceStrategy(SliceAppend),
		WithMapStrategy(MapMerge),
	)

	config := merger.Config()

	if config.NilBehavior != NilOverride {
		t.Errorf("Expected NilBehavior=NilOverride, got %v", config.NilBehavior)
	}
	if config.SliceStrategy != SliceAppend {
		t.Errorf("Expected SliceStrategy=SliceAppend, got %v", config.SliceStrategy)
	}
	if config.MapStrategy != MapMerge {
		t.Errorf("Expected MapStrategy=MapMerge, got %v", config.MapStrategy)
	}
}

func TestMerge_ConvenienceFunction(t *testing.T) {
	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override := TestConfig{
		Port: Some(443),
	}

	result := Merge(base, override)

	if result.Host.Value() != "localhost" {
		t.Errorf("Expected Host=localhost, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}
	if !result.Debug.Value() {
		t.Error("Expected Debug=true")
	}
}

func TestMerger_ZeroValue(t *testing.T) {
	var merger Merger[TestConfig]

	base := TestConfig{
		Host: Some("localhost"),
		Port: Some(8080),
	}
	override := TestConfig{
		Port: Some(443),
	}

	result := merger.Merge(base, override)

	if result.Host.Value() != "localhost" {
		t.Errorf("Expected Host=localhost, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}
}

func TestMerger_AllZeroOverrides(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}

	result := merger.Merge(base, TestConfig{}, TestConfig{}, TestConfig{})

	if result.Host.Value() != "localhost" {
		t.Errorf("Expected Host=localhost, got %v", result.Host.Value())
	}
	if result.Port.Value() != 8080 {
		t.Errorf("Expected Port=8080, got %v", result.Port.Value())
	}
	if !result.Debug.Value() {
		t.Error("Expected Debug=true")
	}
}

func TestMerger_MixedZeroAndNonZeroOverrides(t *testing.T) {
	merger := New[TestConfig]()

	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override1 := TestConfig{}
	override2 := TestConfig{
		Port: Some(443),
	}
	override3 := TestConfig{}

	result := merger.Merge(base, override1, override2, override3)

	if result.Host.Value() != "localhost" {
		t.Errorf("Expected Host=localhost, got %v", result.Host.Value())
	}
	if result.Port.Value() != 443 {
		t.Errorf("Expected Port=443, got %v", result.Port.Value())
	}
	if !result.Debug.Value() {
		t.Error("Expected Debug=true")
	}
}

// Benchmarks

func BenchmarkMerger_SimpleConfig(b *testing.B) {
	merger := New[TestConfig]()
	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override := TestConfig{
		Port: Some(443),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merger.Merge(base, override)
	}
}

func BenchmarkMerger_ComplexConfig(b *testing.B) {
	merger := New[TestConfig]()
	base := TestConfig{
		Host:     Some("localhost"),
		Port:     Some(8080),
		Debug:    Some(true),
		Tags:     NewSlice([]string{"dev", "local", "debug"}),
		Settings: NewMap(map[string]string{"env": "dev", "region": "us", "tier": "free"}),
	}
	override := TestConfig{
		Port:     Some(443),
		Tags:     NewSlice([]string{"prod", "remote"}),
		Settings: NewMap(map[string]string{"env": "prod", "region": "eu"}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merger.Merge(base, override)
	}
}

func BenchmarkMerger_MultipleOverrides(b *testing.B) {
	merger := New[TestConfig]()
	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override1 := TestConfig{Host: Some("staging.com")}
	override2 := TestConfig{Port: Some(443)}
	override3 := TestConfig{Debug: Some(false)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merger.Merge(base, override1, override2, override3)
	}
}

func BenchmarkMerge_ConvenienceFunction(b *testing.B) {
	base := TestConfig{
		Host:  Some("localhost"),
		Port:  Some(8080),
		Debug: Some(true),
	}
	override := TestConfig{
		Port: Some(443),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Merge(base, override)
	}
}
