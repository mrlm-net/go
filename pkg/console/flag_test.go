package console

import (
	"errors"
	"testing"

	"github.com/mrlm-net/go/pkg/console/pipeline"
)

func TestParseFlag(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		defaultVal any
		want       any
		wantErr    bool
	}{
		{"string value", "hello", "", "hello", false},
		{"empty string", "", "", "", false},
		{"int value", "42", 0, 42, false},
		{"negative int", "-5", 0, -5, false},
		{"invalid int", "abc", 0, nil, true},
		{"bool true", "true", false, true, false},
		{"bool false", "false", false, false, false},
		{"bool 1", "1", false, true, false},
		{"bool 0", "0", false, false, false},
		{"invalid bool", "maybe", false, nil, true},
		{"float64 value", "3.14", 0.0, 3.14, false},
		{"float64 negative", "-1.5", 0.0, -1.5, false},
		{"invalid float64", "abc", 0.0, nil, true},
		{"unsupported type returns raw", "raw", struct{}{}, "raw", false},
		{"string with equals", "key=value", "", "key=value", false},
		{"unicode string", "\u00fc\u00f6\u00e4", "", "\u00fc\u00f6\u00e4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlag(tt.raw, tt.defaultVal)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, ErrFlagParse) {
					t.Errorf("error = %v, want ErrFlagParse", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestGetFlag_TypeMismatch(t *testing.T) {
	ctx := pipeline.NewContext()
	setFlag(ctx, "port", "not-an-int")

	got := GetFlag[int](ctx, "port")
	if got != 0 {
		t.Errorf("GetFlag with type mismatch = %d, want 0", got)
	}
}

func TestGetFlag_NotSet(t *testing.T) {
	ctx := pipeline.NewContext()

	got := GetFlag[string](ctx, "missing")
	if got != "" {
		t.Errorf("GetFlag for missing key = %q, want empty", got)
	}
}

func TestGetArgs_Empty(t *testing.T) {
	ctx := pipeline.NewContext()

	got := GetArgs(ctx)
	if got != nil {
		t.Errorf("GetArgs on empty context = %v, want nil", got)
	}
}

func TestGetArgs_Set(t *testing.T) {
	ctx := pipeline.NewContext()
	ctx.Set(argsKey, []string{"a", "b"})

	got := GetArgs(ctx)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("GetArgs = %v, want [a b]", got)
	}
}
