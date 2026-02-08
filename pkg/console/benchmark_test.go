package console

import (
	"fmt"
	"testing"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

func BenchmarkApp_Run_SimpleCommand(b *testing.B) {
	app := New("test", "Test app")
	app.Command("serve", func(ctx *pipeline.Context) error { return nil })
	args := []string{"serve"}
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run(args)
	}
}

func BenchmarkApp_Run_WithFlags_1(b *testing.B) {
	benchmarkAppRunWithFlags(b, 1)
}

func BenchmarkApp_Run_WithFlags_5(b *testing.B) {
	benchmarkAppRunWithFlags(b, 5)
}

func BenchmarkApp_Run_WithFlags_10(b *testing.B) {
	benchmarkAppRunWithFlags(b, 10)
}

func benchmarkAppRunWithFlags(b *testing.B, n int) {
	app := New("test", "Test app")
	cmd := app.Command("serve", func(ctx *pipeline.Context) error { return nil })
	args := []string{"serve"}
	for i := range n {
		name := fmt.Sprintf("flag%d", i)
		cmd.Flag(name, "test flag", "default")
		args = append(args, fmt.Sprintf("--%s=value%d", name, i))
	}
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run(args)
	}
}

func BenchmarkApp_CommandResolution_Depth1(b *testing.B) {
	benchmarkCommandResolution(b, 1)
}

func BenchmarkApp_CommandResolution_Depth3(b *testing.B) {
	benchmarkCommandResolution(b, 3)
}

func BenchmarkApp_CommandResolution_Depth5(b *testing.B) {
	benchmarkCommandResolution(b, 5)
}

func benchmarkCommandResolution(b *testing.B, depth int) {
	app := New("test", "Test app")
	handler := func(ctx *pipeline.Context) error { return nil }

	// Build a chain of subcommands at the specified depth.
	var args []string
	var parent *Command
	for i := range depth {
		name := fmt.Sprintf("cmd%d", i)
		args = append(args, name)
		if i == 0 {
			parent = app.Command(name, handler)
		} else {
			parent = parent.Subcommand(name, handler)
		}
	}
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run(args)
	}
}

func BenchmarkApp_Run_WithMiddleware_1(b *testing.B) {
	benchmarkAppRunWithMiddleware(b, 1)
}

func BenchmarkApp_Run_WithMiddleware_5(b *testing.B) {
	benchmarkAppRunWithMiddleware(b, 5)
}

func BenchmarkApp_Run_WithMiddleware_10(b *testing.B) {
	benchmarkAppRunWithMiddleware(b, 10)
}

func benchmarkAppRunWithMiddleware(b *testing.B, n int) {
	mw := func(next pipeline.Handler) pipeline.Handler {
		return func(ctx *pipeline.Context) error {
			return next(ctx)
		}
	}

	app := New("test", "Test app")
	app.Command("serve", func(ctx *pipeline.Context) error { return nil })
	for range n {
		app.Use(mw)
	}
	args := []string{"serve"}
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run(args)
	}
}

func BenchmarkFlagParsing_1Flag(b *testing.B) {
	benchmarkFlagParsing(b, 1)
}

func BenchmarkFlagParsing_5Flags(b *testing.B) {
	benchmarkFlagParsing(b, 5)
}

func BenchmarkFlagParsing_10Flags(b *testing.B) {
	benchmarkFlagParsing(b, 10)
}

func benchmarkFlagParsing(b *testing.B, n int) {
	app := New("test", "Test app")
	cmd := app.Command("run", func(ctx *pipeline.Context) error { return nil })

	// Register flags of different types.
	args := []string{"run"}
	for i := range n {
		name := fmt.Sprintf("flag%d", i)
		switch i % 3 {
		case 0:
			cmd.Flag(name, "string flag", "default")
			args = append(args, fmt.Sprintf("--%s=value", name))
		case 1:
			cmd.Flag(name, "int flag", 0)
			args = append(args, fmt.Sprintf("--%s=%d", name, i))
		case 2:
			cmd.Flag(name, "bool flag", false)
			args = append(args, fmt.Sprintf("--%s=true", name))
		}
	}
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run(args)
	}
}
