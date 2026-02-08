package console

import (
	"fmt"
	"strconv"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const flagPrefix contextKey = "flag:"

// Flag defines a command-line flag with a typed default value.
type Flag struct {
	Name        string
	Short       string
	Description string
	Default     any
	EnvVar      string // environment variable name (empty = no binding)
	Required    bool   // whether this flag must be provided
}

// GetFlag retrieves a typed flag value from the pipeline context.
// Returns the zero value of T if the flag is not set or the type does not match.
func GetFlag[T any](ctx *pipeline.Context, name string) T {
	return pipeline.Value[T](ctx, flagPrefix+contextKey(name))
}

// setFlag stores a flag value in the pipeline context.
func setFlag(ctx *pipeline.Context, name string, value any) {
	ctx.Set(flagPrefix+contextKey(name), value)
}

// parseFlag converts a string value to the type matching the default value.
func parseFlag(raw string, defaultVal any) (any, error) {
	switch defaultVal.(type) {
	case string:
		return raw, nil
	case int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: %q is not a valid int", ErrFlagParse, raw)
		}
		return v, nil
	case bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: %q is not a valid bool", ErrFlagParse, raw)
		}
		return v, nil
	case float64:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %q is not a valid float64", ErrFlagParse, raw)
		}
		return v, nil
	default:
		return raw, nil
	}
}

// FlagBuilder enables chainable flag configuration.
type FlagBuilder struct {
	cmd  *Command
	flag *Flag
}

// Required marks the flag as required.
func (fb *FlagBuilder) Required() *Command {
	fb.flag.Required = true
	return fb.cmd
}
