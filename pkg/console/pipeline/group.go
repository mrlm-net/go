package pipeline

import "sync"

// Group runs multiple handlers concurrently and collects errors.
// The zero value is ready to use.
type Group struct {
	handlers []Handler
}

// NewGroup creates a Group with the given handlers.
func NewGroup(handlers ...Handler) *Group {
	return &Group{handlers: handlers}
}

// Add appends one or more handlers to the group.
func (g *Group) Add(handlers ...Handler) *Group {
	g.handlers = append(g.handlers, handlers...)
	return g
}

// Run executes all handlers concurrently, sharing the same context.
// It waits for all handlers to complete and returns a combined error
// containing all non-nil errors. If only one error occurs, it is returned
// directly. If no errors occur, nil is returned.
func (g *Group) Run(ctx *Context) error {
	if ctx == nil {
		ctx = NewContext()
	}
	if len(g.handlers) == 0 {
		return nil
	}

	errs := make([]error, len(g.handlers))
	var wg sync.WaitGroup
	wg.Add(len(g.handlers))

	for i, h := range g.handlers {
		go func(idx int, handler Handler) {
			defer wg.Done()
			if handler == nil {
				return
			}
			errs[idx] = handler(ctx)
		}(i, h)
	}

	wg.Wait()

	// Collect non-nil errors.
	var collected []error
	for _, e := range errs {
		if e != nil {
			collected = append(collected, e)
		}
	}

	switch len(collected) {
	case 0:
		return nil
	case 1:
		return collected[0]
	default:
		return &GroupError{Errors: collected}
	}
}

// GroupError holds multiple errors from concurrent handler execution.
type GroupError struct {
	Errors []error
}

// Error returns a string representation of all collected errors.
func (e *GroupError) Error() string {
	msg := "pipeline group: multiple errors:"
	for _, err := range e.Errors {
		msg += "\n  - " + err.Error()
	}
	return msg
}

// Unwrap returns the list of errors for use with errors.Is/As.
func (e *GroupError) Unwrap() []error {
	return e.Errors
}
