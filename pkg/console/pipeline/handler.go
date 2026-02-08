package pipeline

// Handler is a processing step that operates on a Context.
type Handler func(ctx *Context) error

// Middleware wraps a Handler to add cross-cutting behavior.
// It receives the next handler in the chain and returns a new handler.
type Middleware func(next Handler) Handler
