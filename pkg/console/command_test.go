package console

import (
	"testing"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

func TestNewCommand(t *testing.T) {
	handler := func(ctx *pipeline.Context) error { return nil }
	cmd := NewCommand("test", handler)

	if cmd.Name != "test" {
		t.Errorf("Name = %q, want %q", cmd.Name, "test")
	}
	if cmd.handler == nil {
		t.Error("handler should not be nil")
	}
	if cmd.subcommands == nil {
		t.Error("subcommands map should be initialized")
	}
}

func TestNewCommand_NilHandler(t *testing.T) {
	cmd := NewCommand("test", nil)
	if cmd.handler != nil {
		t.Error("handler should be nil")
	}
}

func TestCommand_Flag(t *testing.T) {
	cmd := NewCommand("test", nil)
	result := cmd.Flag("host", "Host address", "localhost")

	if result.cmd != cmd {
		t.Error("Flag() should return a FlagBuilder pointing to the command")
	}
	if len(cmd.Flags) != 1 {
		t.Fatalf("Flags length = %d, want 1", len(cmd.Flags))
	}
	f := cmd.Flags[0]
	if f.Name != "host" || f.Description != "Host address" || f.Default != "localhost" {
		t.Errorf("Flag = %+v, unexpected values", f)
	}
}

func TestCommand_ShortFlag(t *testing.T) {
	cmd := NewCommand("test", nil)
	cmd.ShortFlag("port", "p", "Port number", 3000)

	if len(cmd.Flags) != 1 {
		t.Fatalf("Flags length = %d, want 1", len(cmd.Flags))
	}
	f := cmd.Flags[0]
	if f.Short != "p" {
		t.Errorf("Short = %q, want %q", f.Short, "p")
	}
}

func TestCommand_Subcommand(t *testing.T) {
	parent := NewCommand("db", nil)
	sub := parent.Subcommand("migrate", func(ctx *pipeline.Context) error { return nil })

	if sub == nil {
		t.Fatal("Subcommand() returned nil")
	}
	if sub.Name != "migrate" {
		t.Errorf("Name = %q, want %q", sub.Name, "migrate")
	}
	if parent.findSubcommand("migrate") != sub {
		t.Error("subcommand should be registered in parent")
	}
}

func TestCommand_Describe(t *testing.T) {
	cmd := NewCommand("test", nil)
	result := cmd.Describe("A test command")

	if result != cmd {
		t.Error("Describe() should return the command for chaining")
	}
	if cmd.Description != "A test command" {
		t.Errorf("Description = %q, want %q", cmd.Description, "A test command")
	}
}

func TestCommand_Use(t *testing.T) {
	cmd := NewCommand("test", nil)
	mw := func(next pipeline.Handler) pipeline.Handler { return next }
	result := cmd.Use(mw)

	if result != cmd {
		t.Error("Use() should return the command for chaining")
	}
	if len(cmd.middlewares) != 1 {
		t.Errorf("middlewares length = %d, want 1", len(cmd.middlewares))
	}
}

func TestCommand_FindSubcommand(t *testing.T) {
	cmd := NewCommand("root", nil)
	cmd.Subcommand("child", nil)

	if cmd.findSubcommand("child") == nil {
		t.Error("findSubcommand should find registered subcommand")
	}
	if cmd.findSubcommand("nonexistent") != nil {
		t.Error("findSubcommand should return nil for unknown name")
	}
}

func TestCommand_FlagChaining(t *testing.T) {
	cmd := NewCommand("test", nil)
	// Flag() now returns *FlagBuilder, so we can't chain flags directly.
	// We can chain flag configuration with the builder, then get back *Command.
	cmd.Flag("a", "first", "")
	cmd.Flag("b", "second", 0)
	cmd.ShortFlag("c", "c", "third", false)

	if len(cmd.Flags) != 3 {
		t.Errorf("Flags length = %d, want 3", len(cmd.Flags))
	}
}

func TestFlagBuilder_Required(t *testing.T) {
	cmd := NewCommand("test", nil)
	result := cmd.Flag("target", "Target", "").Required()

	if result != cmd {
		t.Error("Required() should return the command for chaining")
	}
	if len(cmd.Flags) != 1 {
		t.Fatalf("Flags length = %d, want 1", len(cmd.Flags))
	}
	if !cmd.Flags[0].Required {
		t.Error("Flag should be marked as required")
	}
}
