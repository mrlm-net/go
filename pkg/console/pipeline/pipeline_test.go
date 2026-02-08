package pipeline

import (
	"errors"
	"testing"
)

func TestPipeline_BasicExecution(t *testing.T) {
	called := false
	p := New(func(ctx *Context) error {
		called = true
		return nil
	})

	if err := p.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestPipeline_MiddlewareOrder(t *testing.T) {
	var order []int

	mw := func(id int) Middleware {
		return func(next Handler) Handler {
			return func(ctx *Context) error {
				order = append(order, id)
				err := next(ctx)
				order = append(order, -id)
				return err
			}
		}
	}

	p := New(func(ctx *Context) error {
		order = append(order, 0)
		return nil
	})
	p.Use(mw(1), mw(2), mw(3))

	if err := p.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expected := []int{1, 2, 3, 0, -3, -2, -1}
	if len(order) != len(expected) {
		t.Fatalf("order = %v, want %v", order, expected)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Errorf("order[%d] = %d, want %d", i, order[i], expected[i])
		}
	}
}

func TestPipeline_HandlerError(t *testing.T) {
	want := errors.New("handler failed")
	p := New(func(ctx *Context) error {
		return want
	})

	got := p.Run(NewContext())
	if got != want {
		t.Errorf("Run() error = %v, want %v", got, want)
	}
}

func TestPipeline_AbortInMiddleware(t *testing.T) {
	handlerCalled := false

	p := New(func(ctx *Context) error {
		handlerCalled = true
		return nil
	})
	p.Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			ctx.Abort()
			return ErrAborted
		}
	})

	err := p.Run(NewContext())
	if !errors.Is(err, ErrAborted) {
		t.Errorf("Run() error = %v, want ErrAborted", err)
	}
	if handlerCalled {
		t.Error("handler should not have been called after abort")
	}
}

func TestPipeline_NilContext(t *testing.T) {
	called := false
	p := New(func(ctx *Context) error {
		called = true
		return nil
	})

	if err := p.Run(nil); err != nil {
		t.Fatalf("Run(nil) error = %v", err)
	}
	if !called {
		t.Error("handler was not called with nil context")
	}
}

func TestPipeline_NilHandler(t *testing.T) {
	p := New(nil)
	if err := p.Run(NewContext()); err != nil {
		t.Fatalf("Run() with nil handler error = %v", err)
	}
}

func TestPipeline_AbortWithErrorInHandler(t *testing.T) {
	want := errors.New("custom abort error")
	p := New(func(ctx *Context) error {
		ctx.AbortWithError(want)
		return nil
	})

	got := p.Run(NewContext())
	if got != want {
		t.Errorf("Run() error = %v, want %v", got, want)
	}
}

func TestPipeline_AbortWithoutError(t *testing.T) {
	p := New(func(ctx *Context) error {
		ctx.Abort()
		return nil
	})

	got := p.Run(NewContext())
	if !errors.Is(got, ErrAborted) {
		t.Errorf("Run() error = %v, want ErrAborted", got)
	}
}

func TestPipeline_ChainedUse(t *testing.T) {
	count := 0
	mw := func(next Handler) Handler {
		return func(ctx *Context) error {
			count++
			return next(ctx)
		}
	}

	p := New(func(ctx *Context) error { return nil })
	p.Use(mw).Use(mw).Use(mw)

	if err := p.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if count != 3 {
		t.Errorf("middleware called %d times, want 3", count)
	}
}
