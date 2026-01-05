package deepmerge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSome(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"string value", "hello", "hello"},
		{"int value", 42, 42},
		{"zero int", 0, 0},
		{"bool true", true, true},
		{"bool false", false, false},
		{"float64", 3.14, 3.14},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.value.(type) {
			case string:
				opt := Some(v)
				assert.True(t, opt.IsSet(), "Some() should create set Optional")
				assert.Equal(t, v, opt.Value())
			case int:
				opt := Some(v)
				assert.True(t, opt.IsSet(), "Some() should create set Optional")
				assert.Equal(t, v, opt.Value())
			case bool:
				opt := Some(v)
				assert.True(t, opt.IsSet(), "Some() should create set Optional")
				assert.Equal(t, v, opt.Value())
			case float64:
				opt := Some(v)
				assert.True(t, opt.IsSet(), "Some() should create set Optional")
				assert.Equal(t, v, opt.Value())
			}
		})
	}
}

func TestNone(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "string none",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assert.False(t, opt.IsSet(), "None() should create unset Optional")
				assert.Equal(t, "", opt.Value(), "None() should return zero value")
			},
		},
		{
			name: "int none",
			testFunc: func(t *testing.T) {
				opt := None[int]()
				assert.False(t, opt.IsSet(), "None() should create unset Optional")
				assert.Equal(t, 0, opt.Value(), "None() should return zero value")
			},
		},
		{
			name: "bool none",
			testFunc: func(t *testing.T) {
				opt := None[bool]()
				assert.False(t, opt.IsSet(), "None() should create unset Optional")
				assert.Equal(t, false, opt.Value(), "None() should return zero value")
			},
		},
		{
			name: "float64 none",
			testFunc: func(t *testing.T) {
				opt := None[float64]()
				assert.False(t, opt.IsSet(), "None() should create unset Optional")
				assert.Equal(t, 0.0, opt.Value(), "None() should return zero value")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestOptional_Value(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "set string returns value",
			testFunc: func(t *testing.T) {
				opt := Some("test")
				assert.Equal(t, "test", opt.Value())
			},
		},
		{
			name: "unset string returns zero",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assert.Equal(t, "", opt.Value())
			},
		},
		{
			name: "set int returns value",
			testFunc: func(t *testing.T) {
				opt := Some(42)
				assert.Equal(t, 42, opt.Value())
			},
		},
		{
			name: "unset int returns zero",
			testFunc: func(t *testing.T) {
				opt := None[int]()
				assert.Equal(t, 0, opt.Value())
			},
		},
		{
			name: "explicit zero is returned",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assert.True(t, opt.IsSet(), "Some(0) should be set")
				assert.Equal(t, 0, opt.Value())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestOptional_ValueOr(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "set value returns value",
			testFunc: func(t *testing.T) {
				opt := Some("actual")
				assert.Equal(t, "actual", opt.ValueOr("default"))
			},
		},
		{
			name: "unset value returns default",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assert.Equal(t, "default", opt.ValueOr("default"))
			},
		},
		{
			name: "set zero value returns zero not default",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assert.Equal(t, 0, opt.ValueOr(999))
			},
		},
		{
			name: "set empty string returns empty not default",
			testFunc: func(t *testing.T) {
				opt := Some("")
				assert.Equal(t, "", opt.ValueOr("default"))
			},
		},
		{
			name: "set false returns false not default",
			testFunc: func(t *testing.T) {
				opt := Some(false)
				assert.Equal(t, false, opt.ValueOr(true))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestOptional_IsSet(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Some creates set optional",
			testFunc: func(t *testing.T) {
				opt := Some("value")
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "None creates unset optional",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assert.False(t, opt.IsSet())
			},
		},
		{
			name: "zero value struct is unset",
			testFunc: func(t *testing.T) {
				var opt Optional[int]
				assert.False(t, opt.IsSet())
			},
		},
		{
			name: "Some with zero value is set",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "Some with empty string is set",
			testFunc: func(t *testing.T) {
				opt := Some("")
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "Some with false is set",
			testFunc: func(t *testing.T) {
				opt := Some(false)
				assert.True(t, opt.IsSet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestOptional_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "IsZero is inverse of IsSet for Some",
			testFunc: func(t *testing.T) {
				opt := Some("value")
				assert.False(t, opt.IsZero())
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "IsZero is inverse of IsSet for None",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assert.True(t, opt.IsZero())
				assert.False(t, opt.IsSet())
			},
		},
		{
			name: "IsZero for zero value struct",
			testFunc: func(t *testing.T) {
				var opt Optional[int]
				assert.True(t, opt.IsZero())
			},
		},
		{
			name: "IsZero for Some(0) is false",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assert.False(t, opt.IsZero(), "Some(0) is explicitly set, not zero")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestOptional_Merge(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "base set, override set - override wins",
			testFunc: func(t *testing.T) {
				base := Some("base")
				override := Some("override")
				result := base.Merge(override)
				assert.Equal(t, "override", result.Value())
				assert.True(t, result.IsSet())
			},
		},
		{
			name: "base set, override unset - base preserved",
			testFunc: func(t *testing.T) {
				base := Some("base")
				override := None[string]()
				result := base.Merge(override)
				assert.Equal(t, "base", result.Value())
				assert.True(t, result.IsSet())
			},
		},
		{
			name: "base unset, override set - override wins",
			testFunc: func(t *testing.T) {
				base := None[string]()
				override := Some("override")
				result := base.Merge(override)
				assert.Equal(t, "override", result.Value())
				assert.True(t, result.IsSet())
			},
		},
		{
			name: "base unset, override unset - stays unset",
			testFunc: func(t *testing.T) {
				base := None[string]()
				override := None[string]()
				result := base.Merge(override)
				assert.False(t, result.IsSet())
				assert.Equal(t, "", result.Value())
			},
		},
		{
			name: "merge int values",
			testFunc: func(t *testing.T) {
				base := Some(10)
				override := Some(20)
				result := base.Merge(override)
				assert.Equal(t, 20, result.Value())
			},
		},
		{
			name: "merge bool values",
			testFunc: func(t *testing.T) {
				base := Some(true)
				override := Some(false)
				result := base.Merge(override)
				assert.Equal(t, false, result.Value())
				assert.True(t, result.IsSet())
			},
		},
		{
			name: "merge float64 values",
			testFunc: func(t *testing.T) {
				base := Some(1.5)
				override := Some(2.5)
				result := base.Merge(override)
				assert.Equal(t, 2.5, result.Value())
			},
		},
		{
			name: "override with explicit zero wins",
			testFunc: func(t *testing.T) {
				base := Some(42)
				override := Some(0)
				result := base.Merge(override)
				assert.Equal(t, 0, result.Value())
				assert.True(t, result.IsSet(), "Explicit zero should be set")
			},
		},
		{
			name: "base with explicit zero preserved when override unset",
			testFunc: func(t *testing.T) {
				base := Some(0)
				override := None[int]()
				result := base.Merge(override)
				assert.Equal(t, 0, result.Value())
				assert.True(t, result.IsSet(), "Explicit zero should be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestOptional_ZeroValueUseful(t *testing.T) {
	t.Run("zero value struct equals None", func(t *testing.T) {
		var zero Optional[string]
		none := None[string]()

		assert.Equal(t, none.IsSet(), zero.IsSet())
		assert.Equal(t, none.Value(), zero.Value())
		assert.Equal(t, none.IsZero(), zero.IsZero())
	})
}

func TestOptional_DifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "string type",
			testFunc: func(t *testing.T) {
				opt := Some("hello")
				assert.Equal(t, "hello", opt.Value())
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "int type",
			testFunc: func(t *testing.T) {
				opt := Some(123)
				assert.Equal(t, 123, opt.Value())
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "bool type",
			testFunc: func(t *testing.T) {
				opt := Some(true)
				assert.Equal(t, true, opt.Value())
				assert.True(t, opt.IsSet())
			},
		},
		{
			name: "float64 type",
			testFunc: func(t *testing.T) {
				opt := Some(99.99)
				assert.Equal(t, 99.99, opt.Value())
				assert.True(t, opt.IsSet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
