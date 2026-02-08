package pipeline

import "testing"

func TestContext_SetGet(t *testing.T) {
	tests := []struct {
		name   string
		key    any
		value  any
		wantOK bool
	}{
		{"string value", "key", "value", true},
		{"int value", "count", 42, true},
		{"missing key", "missing", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			if tt.wantOK {
				ctx.Set(tt.key, tt.value)
			}
			got, ok := ctx.Get(tt.key)
			if ok != tt.wantOK {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && got != tt.value {
				t.Errorf("Get() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestValue_Typed(t *testing.T) {
	ctx := NewContext()
	ctx.Set("name", "alice")
	ctx.Set("age", 30)

	if got := Value[string](ctx, "name"); got != "alice" {
		t.Errorf("Value[string] = %q, want %q", got, "alice")
	}
	if got := Value[int](ctx, "age"); got != 30 {
		t.Errorf("Value[int] = %d, want %d", got, 30)
	}

	// Wrong type returns zero value.
	if got := Value[int](ctx, "name"); got != 0 {
		t.Errorf("Value[int] for string key = %d, want 0", got)
	}

	// Missing key returns zero value.
	if got := Value[string](ctx, "missing"); got != "" {
		t.Errorf("Value[string] for missing key = %q, want empty", got)
	}
}

func TestContext_Abort(t *testing.T) {
	ctx := NewContext()

	if ctx.IsAborted() {
		t.Error("new context should not be aborted")
	}

	ctx.Abort()
	if !ctx.IsAborted() {
		t.Error("context should be aborted after Abort()")
	}
	if ctx.Err() != nil {
		t.Error("Err() should be nil after Abort() without error")
	}
}

func TestContext_AbortWithError(t *testing.T) {
	ctx := NewContext()
	err := ErrAborted

	ctx.AbortWithError(err)
	if !ctx.IsAborted() {
		t.Error("context should be aborted")
	}
	if ctx.Err() != err {
		t.Errorf("Err() = %v, want %v", ctx.Err(), err)
	}
}
