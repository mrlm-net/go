package console

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

func TestApp_Help(t *testing.T) {
	app := New("myapp", "A test application")
	app.Command("serve", func(ctx *pipeline.Context) error { return nil }).
		Describe("Start the server")
	app.Command("migrate", func(ctx *pipeline.Context) error { return nil }).
		Describe("Run database migrations")

	help := app.Help()

	// Check header.
	if !strings.Contains(help, "myapp - A test application") {
		t.Errorf("help missing header, got:\n%s", help)
	}

	// Check usage.
	if !strings.Contains(help, "Usage:") {
		t.Errorf("help missing usage section, got:\n%s", help)
	}
	if !strings.Contains(help, "myapp <command> [flags]") {
		t.Errorf("help missing usage line, got:\n%s", help)
	}

	// Check commands are listed.
	if !strings.Contains(help, "Commands:") {
		t.Errorf("help missing commands section, got:\n%s", help)
	}
	if !strings.Contains(help, "serve") {
		t.Errorf("help missing serve command, got:\n%s", help)
	}
	if !strings.Contains(help, "migrate") {
		t.Errorf("help missing migrate command, got:\n%s", help)
	}

	// Check footer.
	if !strings.Contains(help, "Use \"myapp <command> --help\"") {
		t.Errorf("help missing footer, got:\n%s", help)
	}
}

func TestCommand_Help(t *testing.T) {
	cmd := NewCommand("serve", func(ctx *pipeline.Context) error { return nil })
	cmd.Describe("Start the HTTP server")
	cmd.Flag("host", "Host to bind", "0.0.0.0")
	cmd.ShortFlag("port", "p", "Port to bind", 8080)
	cmd.Flag("verbose", "Enable verbose logging", false)

	help := cmd.Help()

	// Check header.
	if !strings.Contains(help, "serve - Start the HTTP server") {
		t.Errorf("help missing header, got:\n%s", help)
	}

	// Check usage.
	if !strings.Contains(help, "Usage:") {
		t.Errorf("help missing usage section, got:\n%s", help)
	}

	// Check flags section.
	if !strings.Contains(help, "Flags:") {
		t.Errorf("help missing flags section, got:\n%s", help)
	}

	// Check individual flags.
	if !strings.Contains(help, "--host") {
		t.Errorf("help missing --host flag, got:\n%s", help)
	}
	if !strings.Contains(help, "--port, -p") {
		t.Errorf("help missing --port, -p flag, got:\n%s", help)
	}
	if !strings.Contains(help, "(default: \"0.0.0.0\")") {
		t.Errorf("help missing host default, got:\n%s", help)
	}
	if !strings.Contains(help, "(default: 8080)") {
		t.Errorf("help missing port default, got:\n%s", help)
	}

	// Bool flags should not show default.
	if strings.Contains(help, "(default: false)") {
		t.Errorf("help should not show bool default, got:\n%s", help)
	}
}

func TestCommand_HelpWithPath(t *testing.T) {
	cmd := NewCommand("migrate", func(ctx *pipeline.Context) error { return nil })
	cmd.Describe("Run migrations")

	help := cmd.HelpWithPath("myapp db migrate")

	if !strings.Contains(help, "myapp db migrate - Run migrations") {
		t.Errorf("help should use full path, got:\n%s", help)
	}
}

func TestCommand_HelpWithSubcommands(t *testing.T) {
	cmd := NewCommand("db", nil)
	cmd.Describe("Database operations")
	cmd.Subcommand("migrate", func(ctx *pipeline.Context) error { return nil }).
		Describe("Run migrations")
	cmd.Subcommand("seed", func(ctx *pipeline.Context) error { return nil }).
		Describe("Seed database")

	help := cmd.Help()

	if !strings.Contains(help, "Commands:") {
		t.Errorf("help missing commands section, got:\n%s", help)
	}
	if !strings.Contains(help, "migrate") {
		t.Errorf("help missing migrate subcommand, got:\n%s", help)
	}
	if !strings.Contains(help, "seed") {
		t.Errorf("help missing seed subcommand, got:\n%s", help)
	}
}

func TestApp_RunWithHelp_AppLevel(t *testing.T) {
	var buf bytes.Buffer
	oldWriter := HelpWriter
	HelpWriter = &buf
	defer func() { HelpWriter = oldWriter }()

	app := New("myapp", "Test app")
	app.Command("serve", func(ctx *pipeline.Context) error { return nil }).
		Describe("Start server")

	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"--help flag", []string{"--help"}},
		{"-h flag", []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := app.RunWithHelp(tt.args)
			if err != nil {
				t.Fatalf("RunWithHelp() error = %v", err)
			}
			output := buf.String()
			if !strings.Contains(output, "myapp - Test app") {
				t.Errorf("expected app help, got:\n%s", output)
			}
		})
	}
}

func TestApp_RunWithHelp_CommandLevel(t *testing.T) {
	var buf bytes.Buffer
	oldWriter := HelpWriter
	HelpWriter = &buf
	defer func() { HelpWriter = oldWriter }()

	app := New("myapp", "Test app")
	app.Command("serve", func(ctx *pipeline.Context) error { return nil }).
		Describe("Start the server").
		Flag("port", "Port to bind", 8080)

	tests := []struct {
		name string
		args []string
	}{
		{"command --help", []string{"serve", "--help"}},
		{"command -h", []string{"serve", "-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := app.RunWithHelp(tt.args)
			if err != nil {
				t.Fatalf("RunWithHelp() error = %v", err)
			}
			output := buf.String()
			if !strings.Contains(output, "myapp serve - Start the server") {
				t.Errorf("expected command help with path, got:\n%s", output)
			}
			if !strings.Contains(output, "--port") {
				t.Errorf("expected port flag in help, got:\n%s", output)
			}
		})
	}
}

func TestApp_RunWithHelp_SubcommandLevel(t *testing.T) {
	var buf bytes.Buffer
	oldWriter := HelpWriter
	HelpWriter = &buf
	defer func() { HelpWriter = oldWriter }()

	app := New("myapp", "Test app")
	dbCmd := app.Command("db", func(ctx *pipeline.Context) error { return nil }).
		Describe("Database operations")
	dbCmd.Subcommand("migrate", func(ctx *pipeline.Context) error { return nil }).
		Describe("Run migrations").
		Flag("steps", "Number of steps", 1)

	buf.Reset()
	err := app.RunWithHelp([]string{"db", "migrate", "--help"})
	if err != nil {
		t.Fatalf("RunWithHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "myapp db migrate - Run migrations") {
		t.Errorf("expected subcommand help with full path, got:\n%s", output)
	}
}

func TestApp_RunWithHelp_NormalExecution(t *testing.T) {
	var buf bytes.Buffer
	oldWriter := HelpWriter
	HelpWriter = &buf
	defer func() { HelpWriter = oldWriter }()

	called := false
	app := New("myapp", "Test app")
	app.Command("serve", func(ctx *pipeline.Context) error {
		called = true
		return nil
	})

	err := app.RunWithHelp([]string{"serve"})
	if err != nil {
		t.Fatalf("RunWithHelp() error = %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
	if buf.Len() > 0 {
		t.Errorf("unexpected help output: %s", buf.String())
	}
}

func TestShowHelp(t *testing.T) {
	var buf bytes.Buffer
	oldWriter := HelpWriter
	HelpWriter = &buf
	defer func() { HelpWriter = oldWriter }()

	err := ShowHelp("Custom help text")
	if err != ErrHelpRequested {
		t.Errorf("ShowHelp() error = %v, want ErrHelpRequested", err)
	}
	if buf.String() != "Custom help text" {
		t.Errorf("ShowHelp() wrote %q, want %q", buf.String(), "Custom help text")
	}
}

func TestCheckHelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", []string{}, false},
		{"--help", []string{"--help"}, true},
		{"-h", []string{"-h"}, true},
		{"command --help", []string{"serve", "--help"}, true},
		{"command -h", []string{"serve", "-h"}, true},
		{"no help", []string{"serve", "--port", "8080"}, false},
		{"-- separator", []string{"--", "--help"}, false},
		{"help after separator", []string{"serve", "--", "-h"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkHelpFlag(tt.args); got != tt.want {
				t.Errorf("checkHelpFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelp_CommandsSorted(t *testing.T) {
	app := New("myapp", "Test app")
	app.Command("zebra", func(ctx *pipeline.Context) error { return nil })
	app.Command("alpha", func(ctx *pipeline.Context) error { return nil })
	app.Command("beta", func(ctx *pipeline.Context) error { return nil })

	help := app.Help()

	alphaIdx := strings.Index(help, "alpha")
	betaIdx := strings.Index(help, "beta")
	zebraIdx := strings.Index(help, "zebra")

	if alphaIdx > betaIdx || betaIdx > zebraIdx {
		t.Errorf("commands not sorted alphabetically, got:\n%s", help)
	}
}

func TestHelp_FlagAlignment(t *testing.T) {
	cmd := NewCommand("test", func(ctx *pipeline.Context) error { return nil })
	cmd.Flag("a", "Short flag", "")
	cmd.Flag("verylongflag", "Long flag", "")

	help := cmd.Help()

	// Both descriptions should start at the same column.
	// Find the column where "Short" and "Long" start.
	lines := strings.Split(help, "\n")
	var descColumns []int
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			// Find where "Short" or "Long" starts.
			if idx := strings.Index(line, "Short"); idx > 0 {
				descColumns = append(descColumns, idx)
			} else if idx := strings.Index(line, "Long"); idx > 0 {
				descColumns = append(descColumns, idx)
			}
		}
	}

	if len(descColumns) != 2 {
		t.Fatalf("expected 2 description columns, got %d", len(descColumns))
	}
	if descColumns[0] != descColumns[1] {
		t.Errorf("flag descriptions not aligned at same column, columns: %v, got:\n%s", descColumns, help)
	}
}
