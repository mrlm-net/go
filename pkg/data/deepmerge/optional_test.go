package deepmerge

import (
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, expected, actual any, msgAndArgs ...string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := ""
		if len(msgAndArgs) > 0 {
			msg = ": " + msgAndArgs[0]
		}
		t.Errorf("expected %v, got %v%s", expected, actual, msg)
	}
}

func assertTrue(t *testing.T, val bool, msgAndArgs ...string) {
	t.Helper()
	if !val {
		msg := "expected true, got false"
		if len(msgAndArgs) > 0 {
			msg += ": " + msgAndArgs[0]
		}
		t.Error(msg)
	}
}

func assertFalse(t *testing.T, val bool, msgAndArgs ...string) {
	t.Helper()
	if val {
		msg := "expected false, got true"
		if len(msgAndArgs) > 0 {
			msg += ": " + msgAndArgs[0]
		}
		t.Error(msg)
	}
}

func TestSome(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected any
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
				assertTrue(t, opt.IsSet(), "Some() should create set Optional")
				assertEqual(t, v, opt.Value())
			case int:
				opt := Some(v)
				assertTrue(t, opt.IsSet(), "Some() should create set Optional")
				assertEqual(t, v, opt.Value())
			case bool:
				opt := Some(v)
				assertTrue(t, opt.IsSet(), "Some() should create set Optional")
				assertEqual(t, v, opt.Value())
			case float64:
				opt := Some(v)
				assertTrue(t, opt.IsSet(), "Some() should create set Optional")
				assertEqual(t, v, opt.Value())
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
				assertFalse(t, opt.IsSet(), "None() should create unset Optional")
				assertEqual(t, "", opt.Value(), "None() should return zero value")
			},
		},
		{
			name: "int none",
			testFunc: func(t *testing.T) {
				opt := None[int]()
				assertFalse(t, opt.IsSet(), "None() should create unset Optional")
				assertEqual(t, 0, opt.Value(), "None() should return zero value")
			},
		},
		{
			name: "bool none",
			testFunc: func(t *testing.T) {
				opt := None[bool]()
				assertFalse(t, opt.IsSet(), "None() should create unset Optional")
				assertEqual(t, false, opt.Value(), "None() should return zero value")
			},
		},
		{
			name: "float64 none",
			testFunc: func(t *testing.T) {
				opt := None[float64]()
				assertFalse(t, opt.IsSet(), "None() should create unset Optional")
				assertEqual(t, 0.0, opt.Value(), "None() should return zero value")
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
				assertEqual(t, "test", opt.Value())
			},
		},
		{
			name: "unset string returns zero",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assertEqual(t, "", opt.Value())
			},
		},
		{
			name: "set int returns value",
			testFunc: func(t *testing.T) {
				opt := Some(42)
				assertEqual(t, 42, opt.Value())
			},
		},
		{
			name: "unset int returns zero",
			testFunc: func(t *testing.T) {
				opt := None[int]()
				assertEqual(t, 0, opt.Value())
			},
		},
		{
			name: "explicit zero is returned",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assertTrue(t, opt.IsSet(), "Some(0) should be set")
				assertEqual(t, 0, opt.Value())
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
				assertEqual(t, "actual", opt.ValueOr("default"))
			},
		},
		{
			name: "unset value returns default",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assertEqual(t, "default", opt.ValueOr("default"))
			},
		},
		{
			name: "set zero value returns zero not default",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assertEqual(t, 0, opt.ValueOr(999))
			},
		},
		{
			name: "set empty string returns empty not default",
			testFunc: func(t *testing.T) {
				opt := Some("")
				assertEqual(t, "", opt.ValueOr("default"))
			},
		},
		{
			name: "set false returns false not default",
			testFunc: func(t *testing.T) {
				opt := Some(false)
				assertEqual(t, false, opt.ValueOr(true))
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
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "None creates unset optional",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assertFalse(t, opt.IsSet())
			},
		},
		{
			name: "zero value struct is unset",
			testFunc: func(t *testing.T) {
				var opt Optional[int]
				assertFalse(t, opt.IsSet())
			},
		},
		{
			name: "Some with zero value is set",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "Some with empty string is set",
			testFunc: func(t *testing.T) {
				opt := Some("")
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "Some with false is set",
			testFunc: func(t *testing.T) {
				opt := Some(false)
				assertTrue(t, opt.IsSet())
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
				assertFalse(t, opt.IsZero())
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "IsZero is inverse of IsSet for None",
			testFunc: func(t *testing.T) {
				opt := None[string]()
				assertTrue(t, opt.IsZero())
				assertFalse(t, opt.IsSet())
			},
		},
		{
			name: "IsZero for zero value struct",
			testFunc: func(t *testing.T) {
				var opt Optional[int]
				assertTrue(t, opt.IsZero())
			},
		},
		{
			name: "IsZero for Some(0) is false",
			testFunc: func(t *testing.T) {
				opt := Some(0)
				assertFalse(t, opt.IsZero(), "Some(0) is explicitly set, not zero")
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
				assertEqual(t, "override", result.Value())
				assertTrue(t, result.IsSet())
			},
		},
		{
			name: "base set, override unset - base preserved",
			testFunc: func(t *testing.T) {
				base := Some("base")
				override := None[string]()
				result := base.Merge(override)
				assertEqual(t, "base", result.Value())
				assertTrue(t, result.IsSet())
			},
		},
		{
			name: "base unset, override set - override wins",
			testFunc: func(t *testing.T) {
				base := None[string]()
				override := Some("override")
				result := base.Merge(override)
				assertEqual(t, "override", result.Value())
				assertTrue(t, result.IsSet())
			},
		},
		{
			name: "base unset, override unset - stays unset",
			testFunc: func(t *testing.T) {
				base := None[string]()
				override := None[string]()
				result := base.Merge(override)
				assertFalse(t, result.IsSet())
				assertEqual(t, "", result.Value())
			},
		},
		{
			name: "merge int values",
			testFunc: func(t *testing.T) {
				base := Some(10)
				override := Some(20)
				result := base.Merge(override)
				assertEqual(t, 20, result.Value())
			},
		},
		{
			name: "merge bool values",
			testFunc: func(t *testing.T) {
				base := Some(true)
				override := Some(false)
				result := base.Merge(override)
				assertEqual(t, false, result.Value())
				assertTrue(t, result.IsSet())
			},
		},
		{
			name: "merge float64 values",
			testFunc: func(t *testing.T) {
				base := Some(1.5)
				override := Some(2.5)
				result := base.Merge(override)
				assertEqual(t, 2.5, result.Value())
			},
		},
		{
			name: "override with explicit zero wins",
			testFunc: func(t *testing.T) {
				base := Some(42)
				override := Some(0)
				result := base.Merge(override)
				assertEqual(t, 0, result.Value())
				assertTrue(t, result.IsSet(), "Explicit zero should be set")
			},
		},
		{
			name: "base with explicit zero preserved when override unset",
			testFunc: func(t *testing.T) {
				base := Some(0)
				override := None[int]()
				result := base.Merge(override)
				assertEqual(t, 0, result.Value())
				assertTrue(t, result.IsSet(), "Explicit zero should be preserved")
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

		assertEqual(t, none.IsSet(), zero.IsSet())
		assertEqual(t, none.Value(), zero.Value())
		assertEqual(t, none.IsZero(), zero.IsZero())
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
				assertEqual(t, "hello", opt.Value())
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "int type",
			testFunc: func(t *testing.T) {
				opt := Some(123)
				assertEqual(t, 123, opt.Value())
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "bool type",
			testFunc: func(t *testing.T) {
				opt := Some(true)
				assertEqual(t, true, opt.Value())
				assertTrue(t, opt.IsSet())
			},
		},
		{
			name: "float64 type",
			testFunc: func(t *testing.T) {
				opt := Some(99.99)
				assertEqual(t, 99.99, opt.Value())
				assertTrue(t, opt.IsSet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
