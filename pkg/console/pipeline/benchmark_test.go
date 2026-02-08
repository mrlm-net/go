package pipeline

import "testing"

func BenchmarkPipeline_NoMiddleware(b *testing.B) {
	p := New(func(ctx *Context) error { return nil })
	ctx := NewContext()
	b.ResetTimer()
	for b.Loop() {
		_ = p.Run(ctx)
	}
}

func BenchmarkPipeline_1Middleware(b *testing.B) {
	benchmarkPipelineMiddleware(b, 1)
}

func BenchmarkPipeline_5Middleware(b *testing.B) {
	benchmarkPipelineMiddleware(b, 5)
}

func BenchmarkPipeline_10Middleware(b *testing.B) {
	benchmarkPipelineMiddleware(b, 10)
}

func BenchmarkPipeline_50Middleware(b *testing.B) {
	benchmarkPipelineMiddleware(b, 50)
}

func benchmarkPipelineMiddleware(b *testing.B, n int) {
	mw := func(next Handler) Handler {
		return func(ctx *Context) error {
			return next(ctx)
		}
	}

	p := New(func(ctx *Context) error { return nil })
	for range n {
		p.Use(mw)
	}
	ctx := NewContext()
	b.ResetTimer()
	for b.Loop() {
		_ = p.Run(ctx)
	}
}

// Benchmarks with Build() - zero allocations for middleware chain
func BenchmarkPipeline_Built_1Middleware(b *testing.B) {
	benchmarkPipelineMiddlewareBuilt(b, 1)
}

func BenchmarkPipeline_Built_5Middleware(b *testing.B) {
	benchmarkPipelineMiddlewareBuilt(b, 5)
}

func BenchmarkPipeline_Built_10Middleware(b *testing.B) {
	benchmarkPipelineMiddlewareBuilt(b, 10)
}

func BenchmarkPipeline_Built_50Middleware(b *testing.B) {
	benchmarkPipelineMiddlewareBuilt(b, 50)
}

func benchmarkPipelineMiddlewareBuilt(b *testing.B, n int) {
	mw := func(next Handler) Handler {
		return func(ctx *Context) error {
			return next(ctx)
		}
	}

	p := New(func(ctx *Context) error { return nil })
	for range n {
		p.Use(mw)
	}
	p.Build() // Pre-build the chain
	ctx := NewContext()
	b.ResetTimer()
	for b.Loop() {
		_ = p.Run(ctx)
	}
}

func BenchmarkGroup_1Handler(b *testing.B) {
	benchmarkGroup(b, 1)
}

func BenchmarkGroup_10Handlers(b *testing.B) {
	benchmarkGroup(b, 10)
}

func BenchmarkGroup_100Handlers(b *testing.B) {
	benchmarkGroup(b, 100)
}

func benchmarkGroup(b *testing.B, n int) {
	handlers := make([]Handler, n)
	for i := range handlers {
		handlers[i] = func(ctx *Context) error { return nil }
	}
	g := NewGroup(handlers...)
	ctx := NewContext()
	b.ResetTimer()
	for b.Loop() {
		_ = g.Run(ctx)
	}
}

func BenchmarkContext_SetGet(b *testing.B) {
	ctx := NewContext()
	ctx.Set("key", "value")
	b.ResetTimer()
	for b.Loop() {
		ctx.Get("key")
	}
}

func BenchmarkContext_Value_Typed(b *testing.B) {
	ctx := NewContext()
	ctx.Set("key", "value")
	b.ResetTimer()
	for b.Loop() {
		Value[string](ctx, "key")
	}
}
