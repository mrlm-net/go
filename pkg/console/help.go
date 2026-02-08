package console

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// HelpWriter is the default output writer for help text.
// Override this in tests or to redirect help output.
var HelpWriter io.Writer = os.Stdout

// Help returns the formatted help text for the application.
func (a *App) Help() string {
	var b strings.Builder

	// Header
	b.WriteString(a.name)
	if a.description != "" {
		b.WriteString(" - ")
		b.WriteString(a.description)
	}
	b.WriteString("\n\n")

	// Usage
	b.WriteString("Usage:\n")
	b.WriteString("  ")
	b.WriteString(a.name)
	b.WriteString(" <command> [flags]\n\n")

	// Commands
	if len(a.root.subcommands) > 0 {
		b.WriteString("Commands:\n")
		writeCommands(&b, a.root.subcommands, "  ")
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("Use \"")
	b.WriteString(a.name)
	b.WriteString(" <command> --help\" for more information about a command.\n")

	return b.String()
}

// Help returns the formatted help text for the command.
func (c *Command) Help() string {
	return c.HelpWithPath(c.Name)
}

// HelpWithPath returns help text with the full command path (e.g., "app db migrate").
func (c *Command) HelpWithPath(path string) string {
	var b strings.Builder

	// Header
	b.WriteString(path)
	if c.Description != "" {
		b.WriteString(" - ")
		b.WriteString(c.Description)
	}
	b.WriteString("\n\n")

	// Usage
	b.WriteString("Usage:\n")
	b.WriteString("  ")
	b.WriteString(path)
	if len(c.subcommands) > 0 {
		b.WriteString(" <command>")
	}
	b.WriteString(" [flags]")
	if c.handler != nil {
		b.WriteString(" [args]")
	}
	b.WriteString("\n")

	// Subcommands
	if len(c.subcommands) > 0 {
		b.WriteString("\nCommands:\n")
		writeCommands(&b, c.subcommands, "  ")
	}

	// Flags
	if len(c.Flags) > 0 {
		b.WriteString("\nFlags:\n")
		writeFlags(&b, c.Flags)
	}

	return b.String()
}

// writeCommands formats a map of commands into the builder.
func writeCommands(b *strings.Builder, cmds map[string]*Command, indent string) {
	// Sort command names for consistent output.
	names := make([]string, 0, len(cmds))
	for name := range cmds {
		names = append(names, name)
	}
	sort.Strings(names)

	// Find longest name for alignment.
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	for _, name := range names {
		cmd := cmds[name]
		b.WriteString(indent)
		b.WriteString(name)

		// Show aliases if present.
		var aliasText string
		if len(cmd.Aliases) > 0 {
			aliasText = " (aliases: " + strings.Join(cmd.Aliases, ", ") + ")"
			b.WriteString(aliasText)
		}

		if cmd.Description != "" {
			// Pad to align descriptions.
			padding := maxLen - len(name) - len(aliasText)
			if padding > 0 {
				b.WriteString(strings.Repeat(" ", padding+3))
			} else {
				b.WriteString("   ")
			}
			b.WriteString(cmd.Description)
		}
		b.WriteString("\n")
	}
}

// writeFlags formats flags into the builder.
func writeFlags(b *strings.Builder, flags []Flag) {
	// Build flag entries with their formatted names.
	type entry struct {
		names string
		flag  Flag
	}
	entries := make([]entry, len(flags))

	maxLen := 0
	for i, f := range flags {
		var names string
		if f.Short != "" {
			names = fmt.Sprintf("--%s, -%s", f.Name, f.Short)
		} else {
			names = fmt.Sprintf("--%s", f.Name)
		}
		entries[i] = entry{names: names, flag: f}
		if len(names) > maxLen {
			maxLen = len(names)
		}
	}

	for _, e := range entries {
		b.WriteString("  ")
		b.WriteString(e.names)
		b.WriteString(strings.Repeat(" ", maxLen-len(e.names)+3))

		if e.flag.Description != "" {
			b.WriteString(e.flag.Description)
		}

		// Show default value.
		if e.flag.Default != nil {
			switch v := e.flag.Default.(type) {
			case string:
				if v != "" {
					fmt.Fprintf(b, " (default: %q)", v)
				}
			case bool:
				// Don't show default for bool flags, it's always false.
			default:
				fmt.Fprintf(b, " (default: %v)", v)
			}
		}
		b.WriteString("\n")
	}
}

// ShowHelp writes help to HelpWriter and returns ErrHelpRequested.
// Use this in command handlers that need custom help behavior.
func ShowHelp(help string) error {
	fmt.Fprint(HelpWriter, help)
	return ErrHelpRequested
}

// checkHelpFlag returns true if args contain --help or -h.
func checkHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
		// Stop at -- separator.
		if arg == "--" {
			return false
		}
	}
	return false
}

// RunWithHelp is like Run but handles --help and -h flags automatically.
// When help is requested, it prints help and returns nil (not an error).
func (a *App) RunWithHelp(args []string) error {
	// Check for app-level help (no command or help before command).
	if len(args) == 0 || checkHelpFlag(args) {
		// Check if first arg is a command.
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			// Resolve command and show its help.
			cmd, remaining := a.resolve(args)
			if cmd != a.root && checkHelpFlag(remaining) {
				path := buildCommandPath(a.name, args, cmd)
				fmt.Fprint(HelpWriter, cmd.HelpWithPath(path))
				return nil
			}
		}
		// Show app help.
		fmt.Fprint(HelpWriter, a.Help())
		return nil
	}

	// Resolve command first to check for help at command level.
	cmd, remaining := a.resolve(args)
	if checkHelpFlag(remaining) {
		path := buildCommandPath(a.name, args, cmd)
		fmt.Fprint(HelpWriter, cmd.HelpWithPath(path))
		return nil
	}

	// No help requested, run normally.
	return a.Run(args)
}

// buildCommandPath builds the full path to a command (e.g., "app db migrate").
func buildCommandPath(appName string, args []string, cmd *Command) string {
	parts := []string{appName}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		parts = append(parts, arg)
		if arg == cmd.Name {
			break
		}
	}
	return strings.Join(parts, " ")
}
