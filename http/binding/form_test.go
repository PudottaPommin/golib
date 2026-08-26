package binding_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/pudottapommin/golib/http/binding"
)

type FormTestStruct struct {
	ID   int    `form:"id"`
	Name string `form:"name"`
	Age  int    `form:"age"`
}

func TestFormBinder_Bind(t *testing.T) {
	binder := NewFormBinder[FormTestStruct]()

	form := url.Values{}
	form.Add("id", "123")
	form.Add("name", "John Doe")
	form.Add("age", "30")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	t.Run("BindTo", func(t *testing.T) {
		var dst FormTestStruct
		err := binder.BindTo(req, &dst)
		if err != nil {
			t.Fatalf("BindTo failed: %v", err)
		}

		if dst.ID != 123 {
			t.Errorf("Expected ID 123, got %d", dst.ID)
		}
		if dst.Name != "John Doe" {
			t.Errorf("Expected Name 'John Doe', got '%s'", dst.Name)
		}
		if dst.Age != 30 {
			t.Errorf("Expected Age 30, got %d", dst.Age)
		}
	})

	t.Run("Bind", func(t *testing.T) {
		dst, err := binder.Bind(req)
		if err != nil {
			t.Fatalf("Bind failed: %v", err)
		}

		if dst.ID != 123 {
			t.Errorf("Expected ID 123, got %d", dst.ID)
		}
		if dst.Name != "John Doe" {
			t.Errorf("Expected Name 'John Doe', got '%s'", dst.Name)
		}
		if dst.Age != 30 {
			t.Errorf("Expected Age 30, got %d", dst.Age)
		}
	})
}

type FormManyTestStruct struct {
	IDs   []int    `form:"ids"`
	Names []string `form:"names"`
}

func TestFormBinder_BindMany(t *testing.T) {
	binder := NewFormBinder[FormManyTestStruct]()

	form := url.Values{}
	form.Add("ids", "1")
	form.Add("ids", "2")
	form.Add("ids", "3")
	form.Add("names", "Alice")
	form.Add("names", "Bob")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	var dst FormManyTestStruct
	err := binder.BindTo(req, &dst)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	expectedIDs := []int{1, 2, 3}
	if len(dst.IDs) != len(expectedIDs) {
		t.Errorf("Expected %d IDs, got %d", len(expectedIDs), len(dst.IDs))
	} else {
		for i, v := range dst.IDs {
			if v != expectedIDs[i] {
				t.Errorf("Expected ID[%d] %d, got %d", i, expectedIDs[i], v)
			}
		}
	}

	expectedNames := []string{"Alice", "Bob"}
	if len(dst.Names) != len(expectedNames) {
		t.Errorf("Expected %d Names, got %d", len(expectedNames), len(dst.Names))
	} else {
		for i, v := range dst.Names {
			if v != expectedNames[i] {
				t.Errorf("Expected Name[%d] %s, got %s", i, expectedNames[i], v)
			}
		}
	}
}

func TestFormBinder_SingularMultiple(t *testing.T) {
	binder := NewFormBinder[FormTestStruct]()

	form := url.Values{}
	form.Add("id", "100")
	form.Add("id", "200") // last should win
	form.Add("name", "First")
	form.Add("name", "Last") // last should win

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	var dst FormTestStruct
	err := binder.BindTo(req, &dst)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if dst.ID != 200 {
		t.Errorf("Expected ID 200 (last), got %d", dst.ID)
	}
	if dst.Name != "Last" {
		t.Errorf("Expected Name 'Last' (last), got '%s'", dst.Name)
	}
}

func TestFormBinder_SliceSingle(t *testing.T) {
	binder := NewFormBinder[FormManyTestStruct]()

	form := url.Values{}
	form.Add("ids", "42")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	var dst FormManyTestStruct
	err := binder.BindTo(req, &dst)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if len(dst.IDs) != 1 || dst.IDs[0] != 42 {
		t.Errorf("Expected IDs [42], got %v", dst.IDs)
	}
}

func TestStringBinder_Direct(t *testing.T) {
	m := StringBinder{}
	var s string
	err := m.Bind("test", &s)
	if err != nil {
		t.Fatal(err)
	}
	if s != "test" {
		t.Errorf("got %q want test", s)
	}
}

func TestIntBinder_Direct(t *testing.T) {
	m := IntBinder{}
	var i int
	err := m.Bind("123", &i)
	if err != nil {
		t.Fatal(err)
	}
	if i != 123 {
		t.Errorf("got %d want 123", i)
	}
}

type RequiredTestStruct struct {
	Name string `form:"name,required"`
	Opt  string `form:"opt"`
}

func TestFormBinder_Required(t *testing.T) {
	t.Run("missing required", func(t *testing.T) {
		binder := NewFormBinder[RequiredTestStruct]()
		form := url.Values{}
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Form = form

		var dst RequiredTestStruct
		err := binder.BindTo(req, &dst)
		if err == nil {
			t.Error("expected error for missing required field")
		}
		if !strings.Contains(err.Error(), "missing required field \"name\"") {
			t.Errorf("expected specific error, got %v", err)
		}
	})

	t.Run("required present", func(t *testing.T) {
		binder := NewFormBinder[RequiredTestStruct]()
		form := url.Values{}
		form.Set("name", "test")
		form.Set("opt", "optional")
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Form = form

		var dst RequiredTestStruct
		err := binder.BindTo(req, &dst)
		if err != nil {
			t.Fatalf("Bind failed: %v", err)
		}
		if dst.Name != "test" {
			t.Errorf("expected Name test, got %q", dst.Name)
		}
		if dst.Opt != "optional" {
			t.Errorf("expected Opt optional, got %q", dst.Opt)
		}
	})

	t.Run("optional missing", func(t *testing.T) {
		binder := NewFormBinder[RequiredTestStruct]()
		form := url.Values{"name": []string{"test"}}
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Form = form

		var dst RequiredTestStruct
		err := binder.BindTo(req, &dst)
		if err != nil {
			t.Fatalf("Bind failed: %v", err)
		}
		if dst.Opt != "" {
			t.Errorf("expected Opt empty, got %q", dst.Opt)
		}
	})
}

type PointerFormTestStruct struct {
	ID     *int    `form:"id"`
	Name   *string `form:"name"`
	Active *bool   `form:"active"`
}

func TestFormBinder_Pointers(t *testing.T) {
	binder := NewFormBinder[PointerFormTestStruct]()

	t.Run("populated fields", func(t *testing.T) {
		form := url.Values{}
		form.Set("id", "42")
		form.Set("name", "Jane")
		form.Set("active", "true")

		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Form = form

		var dst PointerFormTestStruct
		err := binder.BindTo(req, &dst)
		if err != nil {
			t.Fatalf("Bind failed: %v", err)
		}

		if dst.ID == nil || *dst.ID != 42 {
			t.Errorf("expected ID 42, got %v", dst.ID)
		}
		if dst.Name == nil || *dst.Name != "Jane" {
			t.Errorf("expected Name 'Jane', got %v", dst.Name)
		}
		if dst.Active == nil || *dst.Active != true {
			t.Errorf("expected Active true, got %v", dst.Active)
		}
	})

	t.Run("empty fields remain nil", func(t *testing.T) {
		form := url.Values{}
		form.Set("id", "")
		form.Set("name", "")
		form.Set("active", "")

		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Form = form

		var dst PointerFormTestStruct
		err := binder.BindTo(req, &dst)
		if err != nil {
			t.Fatalf("Bind failed: %v", err)
		}

		if dst.ID != nil {
			t.Errorf("expected ID nil, got %v", *dst.ID)
		}
		if dst.Name != nil {
			t.Errorf("expected Name nil, got %v", *dst.Name)
		}
		if dst.Active != nil {
			t.Errorf("expected Active nil, got %v", *dst.Active)
		}
	})
}
