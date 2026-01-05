package main

import (
	"fmt"

	"github.com/mrlm-net/go/pkg/data/deepmerge"
)

// ServerConfig demonstrates a realistic configuration struct using deepmerge types.
// This pattern allows layered configuration (defaults -> env -> user overrides).
type ServerConfig struct {
	Host     deepmerge.Optional[string]
	Port     deepmerge.Optional[int]
	TLS      deepmerge.Optional[bool]
	Timeout  deepmerge.Optional[int]
	Tags     deepmerge.MergeableSlice[string]
	Settings deepmerge.MergeableMap[string, string]
}

// IsZero returns true if all fields are zero.
// Required by deepmerge.Mergeable interface.
func (c ServerConfig) IsZero() bool {
	return c.Host.IsZero() &&
		c.Port.IsZero() &&
		c.TLS.IsZero() &&
		c.Timeout.IsZero() &&
		c.Tags.IsZero() &&
		c.Settings.IsZero()
}

// Merge combines two ServerConfig instances field by field.
// Required by deepmerge.Mergeable interface.
func (c ServerConfig) Merge(other ServerConfig) ServerConfig {
	return ServerConfig{
		Host:     c.Host.Merge(other.Host),
		Port:     c.Port.Merge(other.Port),
		TLS:      c.TLS.Merge(other.TLS),
		Timeout:  c.Timeout.Merge(other.Timeout),
		Tags:     c.Tags.Merge(other.Tags),
		Settings: c.Settings.Merge(other.Settings),
	}
}

func main() {
	fmt.Println("=== Deep Merge Example: Layered Configuration ===")

	// 1. Default configuration (application defaults)
	defaults := ServerConfig{
		Host:     deepmerge.Some("localhost"),
		Port:     deepmerge.Some(8080),
		TLS:      deepmerge.Some(false),
		Timeout:  deepmerge.Some(30),
		Tags:     deepmerge.NewSlice([]string{"dev"}).WithStrategy(deepmerge.SliceUnion),
		Settings: deepmerge.NewMap(map[string]string{"env": "development"}).WithStrategy(deepmerge.MapMerge),
	}

	fmt.Println("1. Default Configuration:")
	printConfig(defaults)

	// 2. Environment-specific overrides (from env vars or config file)
	envConfig := ServerConfig{
		Host:     deepmerge.Some("api.example.com"),
		Port:     deepmerge.Some(443),
		TLS:      deepmerge.Some(true),
		Tags:     deepmerge.NewSlice([]string{"production", "web"}).WithStrategy(deepmerge.SliceUnion),
		Settings: deepmerge.NewMap(map[string]string{"env": "production", "region": "us-east-1"}).WithStrategy(deepmerge.MapMerge),
	}

	fmt.Println("\n2. Environment Configuration Override:")
	printConfig(envConfig)

	// 3. User-specific overrides (runtime flags or user preferences)
	userConfig := ServerConfig{
		Timeout:  deepmerge.Some(60),
		Tags:     deepmerge.NewSlice([]string{"monitoring", "alerts"}).WithStrategy(deepmerge.SliceUnion),
		Settings: deepmerge.NewMap(map[string]string{"tier": "premium"}).WithStrategy(deepmerge.MapMerge),
	}

	fmt.Println("\n3. User Configuration Override:")
	printConfig(userConfig)

	// 4. Merge all configurations (defaults -> env -> user)
	finalConfig := deepmerge.Merge(defaults, envConfig, userConfig)

	fmt.Println("\n4. Final Merged Configuration:")
	printConfig(finalConfig)

	// Demonstrate different slice strategies
	fmt.Println("\n=== Slice Merge Strategies ===")
	demonstrateSliceStrategies()

	// Demonstrate different map strategies
	fmt.Println("\n=== Map Merge Strategies ===")
	demonstrateMapStrategies()

	// Demonstrate Optional behavior with zero values
	fmt.Println("\n=== Optional: Explicit Zero Values ===")
	demonstrateOptionalZeroValues()
}

func printConfig(config ServerConfig) {
	fmt.Printf("  Host:     %v (set: %v)\n", config.Host.ValueOr("<not set>"), config.Host.IsSet())
	fmt.Printf("  Port:     %v (set: %v)\n", config.Port.ValueOr(0), config.Port.IsSet())
	fmt.Printf("  TLS:      %v (set: %v)\n", config.TLS.ValueOr(false), config.TLS.IsSet())
	fmt.Printf("  Timeout:  %v (set: %v)\n", config.Timeout.ValueOr(0), config.Timeout.IsSet())

	if !config.Tags.IsZero() {
		fmt.Printf("  Tags:     %v\n", config.Tags.Items())
	} else {
		fmt.Printf("  Tags:     <not set>\n")
	}

	if !config.Settings.IsZero() {
		fmt.Printf("  Settings: %v\n", config.Settings.Items())
	} else {
		fmt.Printf("  Settings: <not set>\n")
	}
}

func demonstrateSliceStrategies() {
	base := []string{"apple", "banana"}
	override := []string{"banana", "cherry", "date"}

	// Strategy 1: Override (default) - replace base with override
	fmt.Println("1. SliceOverride (default):")
	fmt.Printf("   Base:     %v\n", base)
	fmt.Printf("   Override: %v\n", override)
	result := deepmerge.MergeSlices(deepmerge.SliceOverride, base, override)
	fmt.Printf("   Result:   %v\n", result)

	// Strategy 2: Append - concatenate slices (keeps duplicates)
	fmt.Println("\n2. SliceAppend:")
	fmt.Printf("   Base:     %v\n", base)
	fmt.Printf("   Override: %v\n", override)
	result = deepmerge.MergeSlices(deepmerge.SliceAppend, base, override)
	fmt.Printf("   Result:   %v\n", result)

	// Strategy 3: Union - unique values only
	fmt.Println("\n3. SliceUnion:")
	fmt.Printf("   Base:     %v\n", base)
	fmt.Printf("   Override: %v\n", override)
	result = deepmerge.MergeSlices(deepmerge.SliceUnion, base, override)
	fmt.Printf("   Result:   %v\n", result)
}

func demonstrateMapStrategies() {
	base := map[string]string{
		"env":    "development",
		"region": "us-west-1",
		"tier":   "free",
	}
	override := map[string]string{
		"env":     "production",
		"region":  "us-east-1",
		"version": "2.0.0",
	}

	// Strategy 1: Override (default) - replace base with override
	fmt.Println("1. MapOverride (default):")
	fmt.Printf("   Base:     %v\n", base)
	fmt.Printf("   Override: %v\n", override)
	result := deepmerge.MergeMaps(deepmerge.MapOverride, base, override)
	fmt.Printf("   Result:   %v\n", result)

	// Strategy 2: Merge - combine maps (override wins on conflicts)
	fmt.Println("\n2. MapMerge:")
	fmt.Printf("   Base:     %v\n", base)
	fmt.Printf("   Override: %v\n", override)
	result = deepmerge.MergeMaps(deepmerge.MapMerge, base, override)
	fmt.Printf("   Result:   %v\n", result)
	fmt.Printf("   Note: 'tier' preserved from base, 'env' and 'region' overridden, 'version' added\n")
}

func demonstrateOptionalZeroValues() {
	// Problem: How to distinguish "not set" from "set to zero value"?
	// Solution: Use Optional[T]

	type FeatureFlags struct {
		EnableCache   deepmerge.Optional[bool]
		EnableLogging deepmerge.Optional[bool]
		MaxRetries    deepmerge.Optional[int]
	}

	// Defaults: cache enabled, logging enabled, 3 retries
	defaults := FeatureFlags{
		EnableCache:   deepmerge.Some(true),
		EnableLogging: deepmerge.Some(true),
		MaxRetries:    deepmerge.Some(3),
	}

	// User wants to explicitly disable cache and set retries to 0
	userOverride := FeatureFlags{
		EnableCache: deepmerge.Some(false), // Explicitly set to false
		MaxRetries:  deepmerge.Some(0),     // Explicitly set to 0
		// EnableLogging not set - should preserve default
	}

	finalFlags := FeatureFlags{
		EnableCache:   defaults.EnableCache.Merge(userOverride.EnableCache),
		EnableLogging: defaults.EnableLogging.Merge(userOverride.EnableLogging),
		MaxRetries:    defaults.MaxRetries.Merge(userOverride.MaxRetries),
	}

	fmt.Println("Defaults:")
	fmt.Printf("  EnableCache:   %v (set: %v)\n", defaults.EnableCache.Value(), defaults.EnableCache.IsSet())
	fmt.Printf("  EnableLogging: %v (set: %v)\n", defaults.EnableLogging.Value(), defaults.EnableLogging.IsSet())
	fmt.Printf("  MaxRetries:    %v (set: %v)\n", defaults.MaxRetries.Value(), defaults.MaxRetries.IsSet())

	fmt.Println("\nUser Override:")
	fmt.Printf("  EnableCache:   %v (set: %v)\n", userOverride.EnableCache.Value(), userOverride.EnableCache.IsSet())
	fmt.Printf("  EnableLogging: %v (set: %v)\n", userOverride.EnableLogging.Value(), userOverride.EnableLogging.IsSet())
	fmt.Printf("  MaxRetries:    %v (set: %v)\n", userOverride.MaxRetries.Value(), userOverride.MaxRetries.IsSet())

	fmt.Println("\nFinal Configuration:")
	fmt.Printf("  EnableCache:   %v (explicitly disabled by user)\n", finalFlags.EnableCache.Value())
	fmt.Printf("  EnableLogging: %v (preserved from defaults)\n", finalFlags.EnableLogging.Value())
	fmt.Printf("  MaxRetries:    %v (explicitly set to 0 by user)\n", finalFlags.MaxRetries.Value())

	fmt.Println("\nKey Insight:")
	fmt.Println("  Optional allows distinguishing 'not set' from 'set to zero value'")
	fmt.Println("  - Some(false) means 'explicitly set to false'")
	fmt.Println("  - None() means 'not set, use default'")
	fmt.Println("  This is critical for configuration merging!")
}
