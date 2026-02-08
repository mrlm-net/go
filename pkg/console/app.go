package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

// App is the root entry point for a CLI application.
// It holds global middleware, the root command tree, and handles argument parsing.
type App struct {
	name        string
	description string
	root        *Command
	middlewares []pipeline.Middleware
}

// New creates a new App with the given name and description.
func New(name, description string) *App {
	return &App{
		name:        name,
		description: description,
		root: &Command{
			Name:        name,
			subcommands: make(map[string]*Command),
		},
	}
}

// Use adds global middleware that applies to all commands.
func (a *App) Use(mw ...pipeline.Middleware) *App {
	a.middlewares = append(a.middlewares, mw...)
	return a
}

// Command registers a top-level command and returns it for configuration.
func (a *App) Command(name string, handler pipeline.Handler) *Command {
	cmd := NewCommand(name, handler)
	a.root.subcommands[name] = cmd
	return cmd
}

// Run parses the given arguments, resolves the command, and executes the pipeline.
func (a *App) Run(args []string) error {
	cmd, remaining := a.resolve(args)

	if cmd.handler == nil && len(cmd.subcommands) > 0 {
		return fmt.Errorf("%w: %s", ErrCommandNotFound, strings.Join(args, " "))
	}
	if cmd.handler == nil {
		return ErrNoHandler
	}

	// Parse flags and positional args from remaining.
	ctx := pipeline.NewContext()
	positional, err := a.parseFlags(ctx, cmd, remaining)
	if err != nil {
		return err
	}
	ctx.Set(argsKey, positional)

	// Build pipeline: global middleware + command middleware + handler.
	p := pipeline.New(cmd.handler)

	// Command middleware applied first (inner), then global (outer).
	p.Use(cmd.middlewares...)
	p.Use(a.middlewares...)

	return p.Run(ctx)
}

// resolve walks the command tree consuming args that match subcommand names.
func (a *App) resolve(args []string) (*Command, []string) {
	cmd := a.root
	i := 0
	for i < len(args) {
		sub := cmd.findSubcommand(args[i])
		if sub == nil {
			break
		}
		cmd = sub
		i++
	}
	return cmd, args[i:]
}

// parseFlags extracts flags from args and sets defaults, returning remaining positional args.
func (a *App) parseFlags(ctx *pipeline.Context, cmd *Command, args []string) ([]string, error) {
	// Build lookup maps.
	flagByName := make(map[string]*Flag, len(cmd.Flags))
	flagByShort := make(map[string]*Flag, len(cmd.Flags))
	for i := range cmd.Flags {
		f := &cmd.Flags[i]
		flagByName[f.Name] = f
		if f.Short != "" {
			flagByShort[f.Short] = f
		}
	}

	// Track which flags were explicitly set (via env var or CLI arg).
	explicitlySet := make(map[string]bool, len(cmd.Flags))

	// Set defaults, checking env vars first.
	for i := range cmd.Flags {
		f := &cmd.Flags[i]
		defaultVal := f.Default

		// Check environment variable if specified.
		if f.EnvVar != "" {
			if envVal := os.Getenv(f.EnvVar); envVal != "" {
				parsed, err := parseFlag(envVal, f.Default)
				if err != nil {
					return nil, fmt.Errorf("invalid environment variable %s: %w", f.EnvVar, err)
				}
				defaultVal = parsed
				explicitlySet[f.Name] = true
			}
		}

		setFlag(ctx, f.Name, defaultVal)
	}

	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check for "--" separator — stop flag parsing.
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}

		// Long flag: --name=value or --name value
		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			var value string
			if idx := strings.IndexByte(name, '='); idx >= 0 {
				value = name[idx+1:]
				name = name[:idx]
			} else if i+1 < len(args) {
				// Check if this is a bool flag with no value.
				f := flagByName[name]
				if f != nil {
					if _, ok := f.Default.(bool); ok {
						setFlag(ctx, name, true)
						explicitlySet[name] = true
						continue
					}
				}
				i++
				value = args[i]
			} else {
				// Bool flag at end of args.
				f := flagByName[name]
				if f != nil {
					if _, ok := f.Default.(bool); ok {
						setFlag(ctx, name, true)
						explicitlySet[name] = true
						continue
					}
				}
				return nil, fmt.Errorf("%w: missing value for --%s", ErrFlagParse, name)
			}

			f := flagByName[name]
			if f == nil {
				return nil, fmt.Errorf("%w: --%s", ErrFlagNotFound, name)
			}
			parsed, err := parseFlag(value, f.Default)
			if err != nil {
				return nil, err
			}
			setFlag(ctx, name, parsed)
			explicitlySet[name] = true
			continue
		}

		// Short flag: -n value or -n=value
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]
			var value string
			if idx := strings.IndexByte(short, '='); idx >= 0 {
				value = short[idx+1:]
				short = short[:idx]
			} else if i+1 < len(args) {
				f := flagByShort[short]
				if f != nil {
					if _, ok := f.Default.(bool); ok {
						setFlag(ctx, f.Name, true)
						explicitlySet[f.Name] = true
						continue
					}
				}
				i++
				value = args[i]
			} else {
				f := flagByShort[short]
				if f != nil {
					if _, ok := f.Default.(bool); ok {
						setFlag(ctx, f.Name, true)
						explicitlySet[f.Name] = true
						continue
					}
				}
				return nil, fmt.Errorf("%w: missing value for -%s", ErrFlagParse, short)
			}

			f := flagByShort[short]
			if f == nil {
				return nil, fmt.Errorf("%w: -%s", ErrFlagNotFound, short)
			}
			parsed, err := parseFlag(value, f.Default)
			if err != nil {
				return nil, err
			}
			setFlag(ctx, f.Name, parsed)
			explicitlySet[f.Name] = true
			continue
		}

		positional = append(positional, arg)
	}

	// Validate required flags.
	var missingFlags []string
	for i := range cmd.Flags {
		f := &cmd.Flags[i]
		if f.Required && !explicitlySet[f.Name] {
			missingFlags = append(missingFlags, f.Name)
		}
	}

	if len(missingFlags) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrFlagRequired, missingFlags)
	}

	return positional, nil
}

// Recovery returns a middleware that recovers from panics and converts them to errors.
func Recovery() pipeline.Middleware {
	return func(next pipeline.Handler) pipeline.Handler {
		return func(ctx *pipeline.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("panic: %v", r)
					}
				}
			}()
			return next(ctx)
		}
	}
}
