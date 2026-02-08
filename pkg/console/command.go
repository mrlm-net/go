package console

import (
	"github.com/mrlm-net/go/pkg/console/pipeline"
)

// argsKey is the context key for positional arguments.
const argsKey contextKey = "args"

// Command represents a CLI command with optional subcommands, flags, and a handler.
type Command struct {
	Name        string
	Description string
	Flags       []Flag
	Aliases     []string
	handler     pipeline.Handler
	subcommands map[string]*Command
	middlewares []pipeline.Middleware
}

// NewCommand creates a new command with the given name and handler.
func NewCommand(name string, handler pipeline.Handler) *Command {
	return &Command{
		Name:        name,
		handler:     handler,
		subcommands: make(map[string]*Command),
	}
}

// Flag registers a flag on the command and returns a FlagBuilder for chaining.
func (c *Command) Flag(name, description string, defaultVal any) *FlagBuilder {
	c.Flags = append(c.Flags, Flag{
		Name:        name,
		Description: description,
		Default:     defaultVal,
	})
	return &FlagBuilder{
		cmd:  c,
		flag: &c.Flags[len(c.Flags)-1],
	}
}

// ShortFlag registers a flag with a short alias and returns a FlagBuilder.
func (c *Command) ShortFlag(name, short, description string, defaultVal any) *FlagBuilder {
	c.Flags = append(c.Flags, Flag{
		Name:        name,
		Short:       short,
		Description: description,
		Default:     defaultVal,
	})
	return &FlagBuilder{
		cmd:  c,
		flag: &c.Flags[len(c.Flags)-1],
	}
}

// Subcommand registers a subcommand and returns it for further configuration.
func (c *Command) Subcommand(name string, handler pipeline.Handler) *Command {
	sub := NewCommand(name, handler)
	c.subcommands[name] = sub
	return sub
}

// Use adds middleware to this command.
func (c *Command) Use(mw ...pipeline.Middleware) *Command {
	c.middlewares = append(c.middlewares, mw...)
	return c
}

// Describe sets the command description and returns the command for chaining.
func (c *Command) Describe(desc string) *Command {
	c.Description = desc
	return c
}

// Alias adds one or more aliases to the command.
func (c *Command) Alias(names ...string) *Command {
	c.Aliases = append(c.Aliases, names...)
	return c
}

// EnvFlag registers a flag with environment variable binding and returns a FlagBuilder.
func (c *Command) EnvFlag(name, envVar, description string, defaultVal any) *FlagBuilder {
	c.Flags = append(c.Flags, Flag{
		Name:        name,
		Description: description,
		Default:     defaultVal,
		EnvVar:      envVar,
	})
	return &FlagBuilder{
		cmd:  c,
		flag: &c.Flags[len(c.Flags)-1],
	}
}

// findSubcommand looks up a subcommand by name or alias. Returns nil if not found.
func (c *Command) findSubcommand(name string) *Command {
	// Check exact name first.
	if sub, ok := c.subcommands[name]; ok {
		return sub
	}
	// Check aliases.
	for _, sub := range c.subcommands {
		for _, alias := range sub.Aliases {
			if alias == name {
				return sub
			}
		}
	}
	return nil
}

// GetArgs retrieves positional arguments from the context.
func GetArgs(ctx *pipeline.Context) []string {
	return pipeline.Value[[]string](ctx, argsKey)
}
