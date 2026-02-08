package main

import (
	"fmt"
	"os"

	"github.com/mrlm-net/go/pkg/console"
	"github.com/mrlm-net/go/pkg/console/pipeline"
)

func main() {
	app := console.New("myapp", "A sample CLI app")
	app.Use(console.Recovery())

	// Simple command with flags.
	greetCmd := app.Command("greet", func(ctx *pipeline.Context) error {
		name := console.GetFlag[string](ctx, "name")
		fmt.Printf("Hello, %s!\n", name)
		return nil
	})
	greetCmd.Describe("Greet someone by name")
	greetCmd.ShortFlag("name", "n", "Person to greet", "World")

	// Command with multiple flags and environment variable binding.
	// Try: PORT=9000 myapp serve
	serveCmd := app.Command("serve", func(ctx *pipeline.Context) error {
		host := console.GetFlag[string](ctx, "host")
		port := console.GetFlag[int](ctx, "port")
		verbose := console.GetFlag[bool](ctx, "verbose")
		if verbose {
			fmt.Printf("Starting server with verbose logging...\n")
		}
		fmt.Printf("Serving on %s:%d\n", host, port)
		return nil
	})
	serveCmd.Describe("Start the HTTP server")
	serveCmd.Flag("host", "Host address to bind", "0.0.0.0")
	serveCmd.EnvFlag("port", "PORT", "Port number to bind", 8080)
	serveCmd.ShortFlag("verbose", "v", "Enable verbose logging", false)

	// Command with required flags.
	deployCmd := app.Command("deploy", func(ctx *pipeline.Context) error {
		target := console.GetFlag[string](ctx, "target")
		version := console.GetFlag[string](ctx, "version")
		args := console.GetArgs(ctx)
		fmt.Printf("Deploying version %s to %s\n", version, target)
		if len(args) > 0 {
			fmt.Printf("Additional args: %v\n", args)
		}
		return nil
	})
	deployCmd.Describe("Deploy application (supports -- separator)")
	deployCmd.ShortFlag("target", "t", "Deployment target", "").Required()
	deployCmd.ShortFlag("version", "v", "Version to deploy", "").Required()

	// Command with aliases.
	listCmd := app.Command("list", func(ctx *pipeline.Context) error {
		fmt.Println("Listing all items...")
		return nil
	})
	listCmd.Describe("List all items")
	listCmd.Alias("ls", "l")

	// Command with subcommands.
	dbCmd := app.Command("db", func(ctx *pipeline.Context) error {
		// This runs if no subcommand is given.
		return fmt.Errorf("please specify a subcommand: migrate, seed, or reset")
	}).Describe("Database operations")

	migrateCmd := dbCmd.Subcommand("migrate", func(ctx *pipeline.Context) error {
		steps := console.GetFlag[int](ctx, "steps")
		fmt.Printf("Running %d migration(s)...\n", steps)
		return nil
	})
	migrateCmd.Describe("Run database migrations")
	migrateCmd.Flag("steps", "Number of migrations to run", 0)

	seedCmd := dbCmd.Subcommand("seed", func(ctx *pipeline.Context) error {
		fmt.Println("Seeding database...")
		return nil
	})
	seedCmd.Describe("Seed database with sample data")

	resetCmd := dbCmd.Subcommand("reset", func(ctx *pipeline.Context) error {
		force := console.GetFlag[bool](ctx, "force")
		if !force {
			return fmt.Errorf("use --force to confirm database reset")
		}
		fmt.Println("Resetting database...")
		return nil
	})
	resetCmd.Describe("Reset database to initial state")
	resetCmd.ShortFlag("force", "f", "Confirm destructive action", false)

	// RunWithHelp handles --help and -h automatically.
	// Try: myapp --help
	//      myapp serve --help
	//      myapp db --help
	//      myapp db migrate --help
	if err := app.RunWithHelp(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
