// Package pipeline provides a middleware-based execution engine.
//
// A [Pipeline] chains [Middleware] around a final [Handler] using the onion
// model: the first middleware added via [Pipeline.Use] is the outermost layer.
//
//	p := pipeline.New(handler)
//	p.Use(logging, auth, rateLimit)
//	// execution order: logging -> auth -> rateLimit -> handler
//
// # Context
//
// [Context] is a thread-safe key-value store shared across the pipeline.
// Use [Context.Set] and [Context.Get] for untyped access, or the generic
// [Value] function for type-safe retrieval.
//
// Context also carries abort signals. Call [Context.Abort] or
// [Context.AbortWithError] to stop pipeline execution after the current
// handler returns.
//
// # Groups
//
// [Group] runs multiple handlers concurrently, sharing the same Context.
// Errors from all handlers are collected; if multiple handlers fail, a
// [GroupError] is returned that supports [errors.Is] and [errors.As] via
// its Unwrap method.
//
// # Handler and Middleware
//
// [Handler] is a function that processes a Context and returns an error.
// [Middleware] wraps a Handler to add cross-cutting behavior (logging,
// recovery, authentication, etc.).
//
//	func timing(next pipeline.Handler) pipeline.Handler {
//	    return func(ctx *pipeline.Context) error {
//	        start := time.Now()
//	        err := next(ctx)
//	        log.Printf("took %v", time.Since(start))
//	        return err
//	    }
//	}
package pipeline
