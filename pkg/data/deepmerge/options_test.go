package deepmerge

import "testing"

func TestConfigZeroValue(t *testing.T) {
	var cfg Config

	// Zero value should have sensible defaults
	if cfg.NilBehavior != NilSkip {
		t.Errorf("expected NilBehavior to be NilSkip (0), got %v", cfg.NilBehavior)
	}
	if cfg.SliceStrategy != SliceOverride {
		t.Errorf("expected SliceStrategy to be SliceOverride (0), got %v", cfg.SliceStrategy)
	}
	if cfg.MapStrategy != MapOverride {
		t.Errorf("expected MapStrategy to be MapOverride (0), got %v", cfg.MapStrategy)
	}
}

func TestWithNilBehavior(t *testing.T) {
	tests := []struct {
		name     string
		behavior NilBehavior
	}{
		{"NilSkip", NilSkip},
		{"NilOverride", NilOverride},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			opt := WithNilBehavior(tt.behavior)
			opt(&cfg)

			if cfg.NilBehavior != tt.behavior {
				t.Errorf("expected NilBehavior %v, got %v", tt.behavior, cfg.NilBehavior)
			}
		})
	}
}

func TestWithSliceStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy SliceStrategy
	}{
		{"SliceOverride", SliceOverride},
		{"SliceAppend", SliceAppend},
		{"SliceUnion", SliceUnion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			opt := WithSliceStrategy(tt.strategy)
			opt(&cfg)

			if cfg.SliceStrategy != tt.strategy {
				t.Errorf("expected SliceStrategy %v, got %v", tt.strategy, cfg.SliceStrategy)
			}
		})
	}
}

func TestWithMapStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy MapStrategy
	}{
		{"MapOverride", MapOverride},
		{"MapMerge", MapMerge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			opt := WithMapStrategy(tt.strategy)
			opt(&cfg)

			if cfg.MapStrategy != tt.strategy {
				t.Errorf("expected MapStrategy %v, got %v", tt.strategy, cfg.MapStrategy)
			}
		})
	}
}

func TestOptionChaining(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		expected Config
	}{
		{
			name:    "single option",
			options: []Option{WithNilBehavior(NilOverride)},
			expected: Config{
				NilBehavior:   NilOverride,
				SliceStrategy: SliceOverride,
				MapStrategy:   MapOverride,
			},
		},
		{
			name: "multiple options",
			options: []Option{
				WithNilBehavior(NilOverride),
				WithSliceStrategy(SliceAppend),
				WithMapStrategy(MapMerge),
			},
			expected: Config{
				NilBehavior:   NilOverride,
				SliceStrategy: SliceAppend,
				MapStrategy:   MapMerge,
			},
		},
		{
			name: "overriding options",
			options: []Option{
				WithSliceStrategy(SliceAppend),
				WithSliceStrategy(SliceUnion),
			},
			expected: Config{
				NilBehavior:   NilSkip,
				SliceStrategy: SliceUnion,
				MapStrategy:   MapOverride,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			for _, opt := range tt.options {
				opt(&cfg)
			}

			if cfg.NilBehavior != tt.expected.NilBehavior {
				t.Errorf("NilBehavior: expected %v, got %v", tt.expected.NilBehavior, cfg.NilBehavior)
			}
			if cfg.SliceStrategy != tt.expected.SliceStrategy {
				t.Errorf("SliceStrategy: expected %v, got %v", tt.expected.SliceStrategy, cfg.SliceStrategy)
			}
			if cfg.MapStrategy != tt.expected.MapStrategy {
				t.Errorf("MapStrategy: expected %v, got %v", tt.expected.MapStrategy, cfg.MapStrategy)
			}
		})
	}
}

func TestSliceStrategyString(t *testing.T) {
	tests := []struct {
		strategy SliceStrategy
		expected string
	}{
		{SliceOverride, "SliceOverride"},
		{SliceAppend, "SliceAppend"},
		{SliceUnion, "SliceUnion"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestMapStrategyString(t *testing.T) {
	tests := []struct {
		strategy MapStrategy
		expected string
	}{
		{MapOverride, "MapOverride"},
		{MapMerge, "MapMerge"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
