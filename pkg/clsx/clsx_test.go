package clsx

import (
	"strings"
	"testing"
)

func TestMake(t *testing.T) {
	t.Run("stores boolean values", func(t *testing.T) {
		m := Make(
			Boolean("active", true),
			Boolean("disabled", false),
		)
		if len(m) != 2 {
			t.Fatalf("expected map length 2, got %d", len(m))
		}
		if v, ok := m["active"]; !ok || v.kind != kindBoolean || v.b != true {
			t.Fatal("expected active to be a true boolean value")
		}
		if v, ok := m["disabled"]; !ok || v.kind != kindBoolean || v.b != false {
			t.Fatal("expected disabled to be a false boolean value")
		}
	})

	t.Run("stores handler values", func(t *testing.T) {
		m := Make(
			Handler("dynamic", func() bool { return true }),
		)
		if len(m) != 1 {
			t.Fatalf("expected map length 1, got %d", len(m))
		}
		if v, ok := m["dynamic"]; !ok || v.kind != kindFn || v.h == nil {
			t.Fatal("expected dynamic to be a handler value")
		}
	})

	t.Run("ignores invalid kinds", func(t *testing.T) {
		m := Make(
			Value{kind: valueKind(99), cls: "invalid"},
			Boolean("valid", true),
		)
		if len(m) != 1 {
			t.Fatalf("expected map length 1, got %d", len(m))
		}
		if _, ok := m["valid"]; !ok {
			t.Fatal("expected valid key to be present")
		}
		if _, ok := m["invalid"]; ok {
			t.Fatal("expected invalid key to be ignored")
		}
	})
}

func TestMapString(t *testing.T) {
	t.Run("includes only truthy entries", func(t *testing.T) {
		m := Make(
			Boolean("active", true),
			Boolean("disabled", false),
			Handler("dynamic", func() bool { return true }),
			Handler("hidden", func() bool { return false }),
		)
		result := m.String()
		set := classSet(result)

		if _, ok := set["active"]; !ok {
			t.Error("expected result to contain active")
		}
		if _, ok := set["dynamic"]; !ok {
			t.Error("expected result to contain dynamic")
		}
		if _, ok := set["disabled"]; ok {
			t.Error("expected result to not contain disabled")
		}
		if _, ok := set["hidden"]; ok {
			t.Error("expected result to not contain hidden")
		}
	})

	t.Run("returns empty string for empty map", func(t *testing.T) {
		m := Map{}
		result := m.String()
		if result != "" {
			t.Fatalf("expected empty string, got %q", result)
		}
	})

	t.Run("returns empty string when all conditions false", func(t *testing.T) {
		m := Make(
			Boolean("one", false),
			Handler("two", func() bool { return false }),
		)
		result := m.String()
		if result != "" {
			t.Fatalf("expected empty string, got %q", result)
		}
	})
}

func TestValueConstructors(t *testing.T) {
	t.Run("String maps to true boolean", func(t *testing.T) {
		v := String("class")
		if v.kind != kindBoolean || v.cls != "class" || v.b != true {
			t.Fatalf("unexpected value from String: %#v", v)
		}
	})

	t.Run("Boolean sets kind and value", func(t *testing.T) {
		v := Boolean("class", false)
		if v.kind != kindBoolean || v.cls != "class" || v.b != false {
			t.Fatalf("unexpected value from Boolean: %#v", v)
		}
	})

	t.Run("Handler sets kind and function", func(t *testing.T) {
		called := false
		v := Handler("class", func() bool {
			called = true
			return true
		})
		if v.kind != kindFn || v.cls != "class" || v.h == nil {
			t.Fatalf("unexpected value from Handler: %#v", v)
		}
		_ = v.h()
		if !called {
			t.Fatal("expected handler to be callable")
		}
	})
}

func classSet(s string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range strings.Fields(s) {
		set[item] = struct{}{}
	}
	return set
}
