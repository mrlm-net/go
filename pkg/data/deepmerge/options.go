package deepmerge

// NilBehavior defines how nil values should be handled during merge.
type NilBehavior uint8

const (
	NilSkip     NilBehavior = iota // Skip nil, keep base (default)
	NilOverride                    // Nil overwrites base
)

// Config holds merge configuration.
type Config struct {
	NilBehavior   NilBehavior
	SliceStrategy SliceStrategy
	MapStrategy   MapStrategy
}

// Option is a functional option for configuring merge behavior.
type Option func(*Config)

// WithNilBehavior sets how nil values are handled during merge.
func WithNilBehavior(b NilBehavior) Option {
	return func(c *Config) {
		c.NilBehavior = b
	}
}

// WithSliceStrategy sets how slices are merged.
func WithSliceStrategy(s SliceStrategy) Option {
	return func(c *Config) {
		c.SliceStrategy = s
	}
}

// WithMapStrategy sets how maps are merged.
func WithMapStrategy(m MapStrategy) Option {
	return func(c *Config) {
		c.MapStrategy = m
	}
}
