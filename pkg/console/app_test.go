package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

func TestApp_BasicCommand(t *testing.T) {
	called := false
	app := New("test", "test app")
	app.Command("hello", func(ctx *pipeline.Context) error {
		called = true
		return nil
	})

	if err := app.Run([]string{"hello"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestApp_CommandNotFound(t *testing.T) {
	app := New("test", "test app")
	app.Command("hello", func(ctx *pipeline.Context) error { return nil })

	err := app.Run([]string{"goodbye"})
	if !errors.Is(err, ErrCommandNotFound) {
		t.Errorf("Run() error = %v, want ErrCommandNotFound", err)
	}
}

func TestApp_Subcommand(t *testing.T) {
	called := false
	app := New("test", "test app")
	cmd := app.Command("db", func(ctx *pipeline.Context) error { return nil })
	cmd.Subcommand("migrate", func(ctx *pipeline.Context) error {
		called = true
		return nil
	})

	if err := app.Run([]string{"db", "migrate"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Error("subcommand handler was not called")
	}
}

func TestApp_FlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantName string
		wantPort int
	}{
		{"long flag", []string{"serve", "--host", "localhost", "--port", "8080"}, "localhost", 8080},
		{"long flag equals", []string{"serve", "--host=localhost", "--port=8080"}, "localhost", 8080},
		{"default values", []string{"serve"}, "0.0.0.0", 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotHost string
			var gotPort int

			app := New("test", "test app")
			cmd := app.Command("serve", func(ctx *pipeline.Context) error {
				gotHost = GetFlag[string](ctx, "host")
				gotPort = GetFlag[int](ctx, "port")
				return nil
			})
			cmd.Flag("host", "Host to bind", "0.0.0.0")
			cmd.Flag("port", "Port to bind", 3000)

			if err := app.Run(tt.args); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if gotHost != tt.wantName {
				t.Errorf("host = %q, want %q", gotHost, tt.wantName)
			}
			if gotPort != tt.wantPort {
				t.Errorf("port = %d, want %d", gotPort, tt.wantPort)
			}
		})
	}
}

func TestApp_ShortFlag(t *testing.T) {
	var got string
	app := New("test", "test app")
	app.Command("greet", func(ctx *pipeline.Context) error {
		got = GetFlag[string](ctx, "name")
		return nil
	}).ShortFlag("name", "n", "Name", "World")

	if err := app.Run([]string{"greet", "-n", "Alice"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != "Alice" {
		t.Errorf("name = %q, want %q", got, "Alice")
	}
}

func TestApp_BoolFlag(t *testing.T) {
	var got bool
	app := New("test", "test app")
	app.Command("run", func(ctx *pipeline.Context) error {
		got = GetFlag[bool](ctx, "verbose")
		return nil
	}).Flag("verbose", "Verbose output", false)

	if err := app.Run([]string{"run", "--verbose"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !got {
		t.Error("verbose should be true")
	}
}

func TestApp_PositionalArgs(t *testing.T) {
	var got []string
	app := New("test", "test app")
	app.Command("copy", func(ctx *pipeline.Context) error {
		got = GetArgs(ctx)
		return nil
	})

	if err := app.Run([]string{"copy", "src", "dst"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(got) != 2 || got[0] != "src" || got[1] != "dst" {
		t.Errorf("args = %v, want [src dst]", got)
	}
}

func TestApp_GlobalMiddleware(t *testing.T) {
	var order []string
	app := New("test", "test app")
	app.Use(func(next pipeline.Handler) pipeline.Handler {
		return func(ctx *pipeline.Context) error {
			order = append(order, "before")
			err := next(ctx)
			order = append(order, "after")
			return err
		}
	})
	app.Command("run", func(ctx *pipeline.Context) error {
		order = append(order, "handler")
		return nil
	})

	if err := app.Run([]string{"run"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	expected := []string{"before", "handler", "after"}
	if len(order) != len(expected) {
		t.Fatalf("order = %v, want %v", order, expected)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], expected[i])
		}
	}
}

func TestApp_Recovery(t *testing.T) {
	app := New("test", "test app")
	app.Use(Recovery())
	app.Command("panic", func(ctx *pipeline.Context) error {
		panic("something went wrong")
	})

	err := app.Run([]string{"panic"})
	if err == nil {
		t.Fatal("Run() should have returned an error")
	}
	if err.Error() != "panic: something went wrong" {
		t.Errorf("error = %q, want %q", err.Error(), "panic: something went wrong")
	}
}

func TestApp_UnknownFlag(t *testing.T) {
	app := New("test", "test app")
	app.Command("run", func(ctx *pipeline.Context) error { return nil })

	err := app.Run([]string{"run", "--unknown", "val"})
	if !errors.Is(err, ErrFlagNotFound) {
		t.Errorf("Run() error = %v, want ErrFlagNotFound", err)
	}
}

func TestApp_InvalidFlagValue(t *testing.T) {
	app := New("test", "test app")
	app.Command("run", func(ctx *pipeline.Context) error { return nil }).
		Flag("port", "Port", 0)

	err := app.Run([]string{"run", "--port", "abc"})
	if !errors.Is(err, ErrFlagParse) {
		t.Errorf("Run() error = %v, want ErrFlagParse", err)
	}
}

func TestApp_NoArgs(t *testing.T) {
	app := New("test", "test app")
	app.Command("run", func(ctx *pipeline.Context) error { return nil })

	err := app.Run([]string{})
	if err == nil {
		t.Error("Run() with no args should return error")
	}
}

func TestApp_Float64Flag(t *testing.T) {
	var got float64
	app := New("test", "test app")
	cmd := app.Command("calc", func(ctx *pipeline.Context) error {
		got = GetFlag[float64](ctx, "rate")
		return nil
	})
	cmd.Flag("rate", "Rate", 1.0)

	if err := app.Run([]string{"calc", "--rate", "3.14"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != 3.14 {
		t.Errorf("rate = %f, want 3.14", got)
	}
}

// CONSOLE-001: Environment Variable Flag Binding tests

func TestEnvFlagPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		args   []string
		want   int
		setEnv bool
	}{
		{"default only", "", []string{"run"}, 3000, false},
		{"env overrides default", "5000", []string{"run"}, 5000, true},
		{"cli overrides env", "5000", []string{"run", "--port", "8080"}, 8080, true},
		{"cli overrides default", "", []string{"run", "--port", "8080"}, 8080, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("TEST_PORT", tt.envVal)
			}

			var got int
			app := New("test", "test app")
			cmd := app.Command("run", func(ctx *pipeline.Context) error {
				got = GetFlag[int](ctx, "port")
				return nil
			})
			cmd.EnvFlag("port", "TEST_PORT", "Port", 3000)

			if err := app.Run(tt.args); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("port = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEnvFlagTypeConversion(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   any
	}{
		{"int from env", "42", 42},
		{"bool true from env", "true", true},
		{"bool false from env", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_VAL", tt.envVal)

			var gotInt int
			var gotBool bool
			app := New("test", "test app")

			switch tt.want.(type) {
			case int:
				cmd := app.Command("run", func(ctx *pipeline.Context) error {
					gotInt = GetFlag[int](ctx, "val")
					return nil
				})
				cmd.EnvFlag("val", "TEST_VAL", "Value", 0)

				if err := app.Run([]string{"run"}); err != nil {
					t.Fatalf("Run() error = %v", err)
				}
				if gotInt != tt.want.(int) {
					t.Errorf("val = %d, want %d", gotInt, tt.want.(int))
				}
			case bool:
				cmd := app.Command("run", func(ctx *pipeline.Context) error {
					gotBool = GetFlag[bool](ctx, "val")
					return nil
				})
				cmd.EnvFlag("val", "TEST_VAL", "Value", false)

				if err := app.Run([]string{"run"}); err != nil {
					t.Fatalf("Run() error = %v", err)
				}
				if gotBool != tt.want.(bool) {
					t.Errorf("val = %v, want %v", gotBool, tt.want.(bool))
				}
			}
		})
	}
}

func TestEnvFlagInvalidType(t *testing.T) {
	t.Setenv("TEST_PORT", "notanumber")

	app := New("test", "test app")
	cmd := app.Command("run", func(ctx *pipeline.Context) error { return nil })
	cmd.EnvFlag("port", "TEST_PORT", "Port", 0)

	err := app.Run([]string{"run"})
	if !errors.Is(err, ErrFlagParse) {
		t.Errorf("Run() error = %v, want ErrFlagParse", err)
	}
}

// CONSOLE-002: Required Flag Validation tests

func TestRequiredFlag(t *testing.T) {
	app := New("test", "test app")
	cmd := app.Command("deploy", func(ctx *pipeline.Context) error { return nil })
	cmd.Flag("target", "Target environment", "").Required()

	err := app.Run([]string{"deploy"})
	if !errors.Is(err, ErrFlagRequired) {
		t.Errorf("Run() error = %v, want ErrFlagRequired", err)
	}
}

func TestRequiredFlagProvided(t *testing.T) {
	var got string
	app := New("test", "test app")
	cmd := app.Command("deploy", func(ctx *pipeline.Context) error {
		got = GetFlag[string](ctx, "target")
		return nil
	})
	cmd.Flag("target", "Target environment", "").Required()

	if err := app.Run([]string{"deploy", "--target", "production"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != "production" {
		t.Errorf("target = %q, want %q", got, "production")
	}
}

func TestRequiredFlagWithEnvVar(t *testing.T) {
	t.Setenv("DEPLOY_TARGET", "staging")

	var got string
	app := New("test", "test app")
	cmd := app.Command("deploy", func(ctx *pipeline.Context) error {
		got = GetFlag[string](ctx, "target")
		return nil
	})
	cmd.EnvFlag("target", "DEPLOY_TARGET", "Target environment", "").Required()

	if err := app.Run([]string{"deploy"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != "staging" {
		t.Errorf("target = %q, want %q", got, "staging")
	}
}

func TestMultipleMissingRequiredFlags(t *testing.T) {
	app := New("test", "test app")
	cmd := app.Command("deploy", func(ctx *pipeline.Context) error { return nil })
	cmd.Flag("target", "Target", "").Required()
	cmd.Flag("version", "Version", "").Required()

	err := app.Run([]string{"deploy"})
	if !errors.Is(err, ErrFlagRequired) {
		t.Fatalf("Run() error = %v, want ErrFlagRequired", err)
	}

	// Both flags should be listed in the error.
	errMsg := err.Error()
	if !strings.Contains(errMsg, "target") || !strings.Contains(errMsg, "version") {
		t.Errorf("error message %q should contain both 'target' and 'version'", errMsg)
	}
}

// CONSOLE-003: Command Aliases tests

func TestCommandAliases(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{"primary name", "list"},
		{"first alias", "ls"},
		{"second alias", "l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			app := New("test", "test app")
			cmd := app.Command("list", func(ctx *pipeline.Context) error {
				called = true
				return nil
			})
			cmd.Alias("ls", "l")

			if err := app.Run([]string{tt.arg}); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if !called {
				t.Error("handler was not called")
			}
		})
	}
}

func TestAliasDoesNotConflict(t *testing.T) {
	calledList := false
	calledLs := false

	app := New("test", "test app")
	app.Command("list", func(ctx *pipeline.Context) error {
		calledList = true
		return nil
	}).Alias("l")

	app.Command("ls", func(ctx *pipeline.Context) error {
		calledLs = true
		return nil
	})

	// Exact name "ls" should take precedence over alias "l".
	if err := app.Run([]string{"ls"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !calledLs {
		t.Error("ls command was not called")
	}
	if calledList {
		t.Error("list command should not have been called")
	}
}

// CONSOLE-004: -- Separator tests

func TestDoubleDashSeparator(t *testing.T) {
	var gotVerbose bool
	var gotArgs []string

	app := New("test", "test app")
	cmd := app.Command("run", func(ctx *pipeline.Context) error {
		gotVerbose = GetFlag[bool](ctx, "verbose")
		gotArgs = GetArgs(ctx)
		return nil
	})
	cmd.Flag("verbose", "Verbose", false)

	err := app.Run([]string{"run", "--verbose", "--", "--not-a-flag", "arg"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !gotVerbose {
		t.Error("verbose should be true")
	}
	if len(gotArgs) != 2 || gotArgs[0] != "--not-a-flag" || gotArgs[1] != "arg" {
		t.Errorf("args = %v, want [--not-a-flag arg]", gotArgs)
	}
}

func TestDoubleDashWithNoArgs(t *testing.T) {
	var gotArgs []string

	app := New("test", "test app")
	app.Command("run", func(ctx *pipeline.Context) error {
		gotArgs = GetArgs(ctx)
		return nil
	})

	if err := app.Run([]string{"run", "--"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(gotArgs) != 0 {
		t.Errorf("args = %v, want []", gotArgs)
	}
}

func TestDoubleDashInMiddle(t *testing.T) {
	var gotPort int
	var gotArgs []string

	app := New("test", "test app")
	cmd := app.Command("run", func(ctx *pipeline.Context) error {
		gotPort = GetFlag[int](ctx, "port")
		gotArgs = GetArgs(ctx)
		return nil
	})
	cmd.Flag("port", "Port", 0)

	err := app.Run([]string{"run", "--port", "8080", "--", "--verbose", "file.txt"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if gotPort != 8080 {
		t.Errorf("port = %d, want 8080", gotPort)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "--verbose" || gotArgs[1] != "file.txt" {
		t.Errorf("args = %v, want [--verbose file.txt]", gotArgs)
	}
}

// Benchmarks

func BenchmarkEnvFlag(b *testing.B) {
	b.Setenv("BENCH_PORT", "8080")

	app := New("test", "test app")
	cmd := app.Command("run", func(ctx *pipeline.Context) error {
		_ = GetFlag[int](ctx, "port")
		return nil
	})
	cmd.EnvFlag("port", "BENCH_PORT", "Port", 3000)

	b.ResetTimer()
	for b.Loop() {
		if err := app.Run([]string{"run"}); err != nil {
			b.Fatal(err)
		}
	}
}
