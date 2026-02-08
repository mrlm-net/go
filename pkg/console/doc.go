// Package console provides a CLI application framework built on the
// [github.com/mrlm-net/go/pkg/console/pipeline] middleware engine.
//
// An [App] holds the command tree, global middleware, and flag definitions.
// Commands are registered with handlers and can nest arbitrarily via
// [Command.Subcommand].
//
//	app := console.New("myapp", "My CLI tool")
//	app.Use(console.Recovery())
//
//	app.Command("serve", serveHandler).
//	    Describe("Start the server").
//	    Flag("host", "Bind address", "0.0.0.0").
//	    ShortFlag("port", "p", "Port number", 3000)
//
//	app.RunWithHelp(os.Args[1:])
//
// # Command Resolution
//
// Arguments are matched left-to-right against the command tree. The first
// argument that does not match a subcommand name becomes a positional
// argument. Retrieve positional args with [GetArgs].
//
// # Flags
//
// Flags are typed by their default value. Supported types: string, int,
// bool, float64. Both long (--name) and short (-n) forms are supported,
// with value provided as --name=value or --name value.
//
// Bool flags can be used without a value (--verbose sets true).
// Retrieve flag values with the generic [GetFlag] function.
//
// # Help System
//
// Use [App.RunWithHelp] instead of [App.Run] to enable automatic help.
// When --help or -h is passed, help text is printed and nil is returned.
//
//	myapp --help           # Show app help
//	myapp serve --help     # Show command help
//	myapp db migrate -h    # Show subcommand help
//
// Add descriptions to commands with [Command.Describe] for better help output.
// Help text includes command descriptions, available flags with defaults,
// and subcommands when present.
//
// For custom help behavior, use [ShowHelp] in your handler:
//
//	if needsCustomHelp {
//	    return console.ShowHelp(cmd.Help())
//	}
//
// # Middleware
//
// Global middleware added via [App.Use] wraps all commands (outermost).
// Per-command middleware added via [Command.Use] wraps only that command
// (inner layer). Execution order: global middleware -> command middleware
// -> handler.
//
// The built-in [Recovery] middleware converts panics to errors.
package console
