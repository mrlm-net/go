package pipeline

import (
	"errors"
	"sync/atomic"
	"testing"
)

func TestGroup_AllSucceed(t *testing.T) {
	var count atomic.Int32
	g := NewGroup(
		func(ctx *Context) error { count.Add(1); return nil },
		func(ctx *Context) error { count.Add(1); return nil },
		func(ctx *Context) error { count.Add(1); return nil },
	)

	if err := g.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := count.Load(); got != 3 {
		t.Errorf("handlers called %d times, want 3", got)
	}
}

func TestGroup_SingleError(t *testing.T) {
	want := errors.New("fail")
	g := NewGroup(
		func(ctx *Context) error { return nil },
		func(ctx *Context) error { return want },
		func(ctx *Context) error { return nil },
	)

	got := g.Run(NewContext())
	if got != want {
		t.Errorf("Run() error = %v, want %v", got, want)
	}
}

func TestGroup_MultipleErrors(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	g := NewGroup(
		func(ctx *Context) error { return err1 },
		func(ctx *Context) error { return err2 },
	)

	got := g.Run(NewContext())
	var ge *GroupError
	if !errors.As(got, &ge) {
		t.Fatalf("Run() error type = %T, want *GroupError", got)
	}
	if len(ge.Errors) != 2 {
		t.Errorf("GroupError has %d errors, want 2", len(ge.Errors))
	}
}

func TestGroup_Empty(t *testing.T) {
	g := NewGroup()
	if err := g.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestGroup_NilContext(t *testing.T) {
	called := false
	g := NewGroup(func(ctx *Context) error {
		called = true
		return nil
	})
	if err := g.Run(nil); err != nil {
		t.Fatalf("Run(nil) error = %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestGroup_NilHandler(t *testing.T) {
	g := NewGroup(nil, func(ctx *Context) error { return nil })
	if err := g.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestGroup_SharedContext(t *testing.T) {
	ctx := NewContext()
	var done atomic.Int32

	g := NewGroup(
		func(ctx *Context) error {
			ctx.Set("a", 1)
			done.Add(1)
			return nil
		},
		func(ctx *Context) error {
			ctx.Set("b", 2)
			done.Add(1)
			return nil
		},
	)

	if err := g.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Both handlers ran and set values (context is thread-safe).
	if got := done.Load(); got != 2 {
		t.Errorf("handlers completed = %d, want 2", got)
	}
}

func TestGroup_Add(t *testing.T) {
	var count atomic.Int32
	g := NewGroup()
	g.Add(
		func(ctx *Context) error { count.Add(1); return nil },
		func(ctx *Context) error { count.Add(1); return nil },
	)

	if err := g.Run(NewContext()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := count.Load(); got != 2 {
		t.Errorf("handlers called %d times, want 2", got)
	}
}

func TestGroupError_Unwrap(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	ge := &GroupError{Errors: []error{err1, err2}}

	if !errors.Is(ge, err1) {
		t.Error("errors.Is should find err1")
	}
	if !errors.Is(ge, err2) {
		t.Error("errors.Is should find err2")
	}
}
