package binding_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	. "github.com/pudottapommin/golib/http/binding"
	"github.com/stretchr/testify/assert"
)

type GenericUserForm struct {
	ID      int64    `form:"id,required"`
	Name    string   `form:"name,required,trim"`
	Age     int      `form:"age"`
	Active  bool     `form:"active"`
	Score   float64  `form:"score"`
	Tags    []string `form:"tags"`
	Ratings []int    `form:"ratings"`
}

type CustomUserID string
type CustomAge int
type CustomFlag bool

type CustomTypeForm struct {
	ID     CustomUserID `form:"id"`
	Age    CustomAge    `form:"age"`
	Active CustomFlag   `form:"active"`
}

func TestGenericBind(t *testing.T) {
	form := url.Values{}
	form.Set("id", "9876543210")
	form.Set("name", "  Jane Doe  ")
	form.Set("age", "28")
	form.Set("active", "true")
	form.Set("score", "98.5")
	form.Add("tags", "go")
	form.Add("tags", "web")
	form.Add("ratings", "5")
	form.Add("ratings", "4")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	t.Run("Bind[T] Success", func(t *testing.T) {
		user, err := Bind[GenericUserForm](req)
		assert.NoError(t, err)
		assert.Equal(t, int64(9876543210), user.ID)
		assert.Equal(t, "Jane Doe", user.Name)
		assert.Equal(t, 28, user.Age)
		assert.True(t, user.Active)
		assert.InDelta(t, 98.5, user.Score, 0.001)
		assert.Equal(t, []string{"go", "web"}, user.Tags)
		assert.Equal(t, []int{5, 4}, user.Ratings)
	})

	t.Run("BindTo[T] Success", func(t *testing.T) {
		var user GenericUserForm
		err := BindTo(req, &user)
		assert.NoError(t, err)
		assert.Equal(t, int64(9876543210), user.ID)
		assert.Equal(t, "Jane Doe", user.Name)
	})

	t.Run("FormBinder[T].Bind Instance Method", func(t *testing.T) {
		fb := NewFormBinder[GenericUserForm]()
		user, err := fb.Bind(req)
		assert.NoError(t, err)
		assert.Equal(t, int64(9876543210), user.ID)
		assert.Equal(t, "Jane Doe", user.Name)
	})

	t.Run("FormBinder[T].BindTo Instance Method", func(t *testing.T) {
		fb := NewFormBinder[GenericUserForm]()
		var user GenericUserForm
		err := fb.BindTo(req, &user)
		assert.NoError(t, err)
		assert.Equal(t, int64(9876543210), user.ID)
	})

	t.Run("For[T] Helper", func(t *testing.T) {
		tb := For[GenericUserForm]()
		user, err := tb.Bind(req)
		assert.NoError(t, err)
		assert.Equal(t, int64(9876543210), user.ID)
	})

	t.Run("Bind[T] Required Field Error", func(t *testing.T) {
		badForm := url.Values{}
		badForm.Set("name", "Jane")
		badReq, _ := http.NewRequest(http.MethodPost, "/", nil)
		badReq.Form = badForm

		_, err := Bind[GenericUserForm](badReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field \"id\"")
	})
}

func TestCustomApproximateTypes(t *testing.T) {
	form := url.Values{}
	form.Set("id", "user_abc_123")
	form.Set("age", "35")
	form.Set("active", "true")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	t.Run("Default Auto Match", func(t *testing.T) {
		res, err := Bind[CustomTypeForm](req)
		assert.NoError(t, err)
		assert.Equal(t, CustomUserID("user_abc_123"), res.ID)
		assert.Equal(t, CustomAge(35), res.Age)
		assert.Equal(t, CustomFlag(true), res.Active)
	})

	t.Run("With Explicit Generic Binders", func(t *testing.T) {
		res, err := Bind[CustomTypeForm](req,
			FormWithGenericBinder[CustomTypeForm](StringGenericBinder[CustomUserID]{}),
			FormWithGenericBinder[CustomTypeForm](NumericBinder[CustomAge]{}),
			FormWithGenericBinder[CustomTypeForm](BoolGenericBinder[CustomFlag]{}),
		)
		assert.NoError(t, err)
		assert.Equal(t, CustomUserID("user_abc_123"), res.ID)
		assert.Equal(t, CustomAge(35), res.Age)
		assert.Equal(t, CustomFlag(true), res.Active)
	})
}

func TestFormParsingOptions(t *testing.T) {
	t.Run("FormWithParseForm", func(t *testing.T) {
		body := strings.NewReader("id=777&name=Alice&age=22")
		req, _ := http.NewRequest(http.MethodPost, "/?extra=1", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		user, err := Bind[GenericUserForm](req, FormWithParseForm[GenericUserForm]())
		assert.NoError(t, err)
		assert.Equal(t, int64(777), user.ID)
		assert.Equal(t, "Alice", user.Name)
		assert.Equal(t, 22, user.Age)
	})

	t.Run("FormWithParseMultipart", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		_ = w.WriteField("id", "888")
		_ = w.WriteField("name", "Bob")
		_ = w.Close()

		req, _ := http.NewRequest(http.MethodPost, "/", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		user, err := Bind[GenericUserForm](req, FormWithParseMultipart[GenericUserForm](10<<20))
		assert.NoError(t, err)
		assert.Equal(t, int64(888), user.ID)
		assert.Equal(t, "Bob", user.Name)
	})

	t.Run("FormWithMaxMemory", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		_ = w.WriteField("id", "999")
		_ = w.WriteField("name", "Charlie")
		_ = w.Close()

		req, _ := http.NewRequest(http.MethodPost, "/", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		user, err := Bind[GenericUserForm](req, FormWithMaxMemory[GenericUserForm](5<<20))
		assert.NoError(t, err)
		assert.Equal(t, int64(999), user.ID)
		assert.Equal(t, "Charlie", user.Name)
	})

	t.Run("FormWithSkipParse", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Form = url.Values{"id": []string{"111"}, "name": []string{"Dave"}}

		user, err := Bind[GenericUserForm](req, FormWithSkipParse[GenericUserForm]())
		assert.NoError(t, err)
		assert.Equal(t, int64(111), user.ID)
		assert.Equal(t, "Dave", user.Name)
	})
}

func TestGenericMappable(t *testing.T) {
	strBinder := WrapBinder(StringBinder{})
	intBinder := WrapBinder(IntBinder{})

	var s string
	var i int
	var slice []string

	assert.True(t, Mappable(strBinder, &s))
	assert.True(t, Mappable(strBinder, s))
	assert.True(t, Mappable(strBinder, &slice))
	assert.False(t, Mappable(strBinder, &i))

	assert.True(t, Mappable(intBinder, &i))
	assert.True(t, Mappable(intBinder, i))
	assert.False(t, Mappable(intBinder, &s))
}

func TestSliceBufferReuse(t *testing.T) {
	t.Run("NumericBinder Slice Reuse", func(t *testing.T) {
		b := NumericBinder[int]{}
		buf := make([]int, 5)
		err := b.BindMany([]string{"10", "20", "30"}, &buf)
		assert.NoError(t, err)
		assert.Equal(t, []int{10, 20, 30}, buf)
	})

	t.Run("StringBinder Slice Reuse", func(t *testing.T) {
		b := StringBinder{}
		buf := make([]string, 4)
		err := b.BindMany([]string{"alpha", "beta"}, &buf)
		assert.NoError(t, err)
		assert.Equal(t, []string{"alpha", "beta"}, buf)
	})

	t.Run("BoolBinder Slice Reuse", func(t *testing.T) {
		b := BoolBinder{}
		buf := make([]bool, 3)
		err := b.BindMany([]string{"true", "false"}, &buf)
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, false}, buf)
	})
}

func TestGenericValueExtraction(t *testing.T) {
	form := url.Values{}
	form.Set("id", "456")
	form.Set("pi", "3.14159")
	form.Set("active", "true")
	form.Set("title", " Go Library ")
	form.Add("scores", "10")
	form.Add("scores", "20")
	form.Add("scores", "30")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	t.Run("Value[T] int", func(t *testing.T) {
		val, err := Value[int](req, "id")
		assert.NoError(t, err)
		assert.Equal(t, 456, val)
	})

	t.Run("Value[T] float64", func(t *testing.T) {
		val, err := Value[float64](req, "pi")
		assert.NoError(t, err)
		assert.InDelta(t, 3.14159, val, 0.00001)
	})

	t.Run("Value[T] bool", func(t *testing.T) {
		val, err := Value[bool](req, "active")
		assert.NoError(t, err)
		assert.True(t, val)
	})

	t.Run("Value[T] string", func(t *testing.T) {
		val, err := Value[string](req, "title")
		assert.NoError(t, err)
		assert.Equal(t, " Go Library ", val)
	})

	t.Run("Value[T] missing field", func(t *testing.T) {
		_, err := Value[int](req, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing field \"nonexistent\"")
	})

	t.Run("Values[T] slice", func(t *testing.T) {
		scores, err := Values[int](req, "scores")
		assert.NoError(t, err)
		assert.Equal(t, []int{10, 20, 30}, scores)
	})

	t.Run("Values[T] missing field returns nil", func(t *testing.T) {
		vals, err := Values[int](req, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, vals)
	})

	t.Run("ValueOrDefault[T]", func(t *testing.T) {
		assert.Equal(t, 456, ValueOrDefault[int](req, "id", 999))
		assert.Equal(t, 999, ValueOrDefault[int](req, "nonexistent", 999))
		assert.Equal(t, 999, ValueOrDefault[int](req, "title", 999)) // parse error yields default
	})
}

func TestGenericParse(t *testing.T) {
	t.Run("Parse[int]", func(t *testing.T) {
		v, err := Parse[int]("100")
		assert.NoError(t, err)
		assert.Equal(t, 100, v)
	})

	t.Run("Parse[float64]", func(t *testing.T) {
		v, err := Parse[float64]("2.718")
		assert.NoError(t, err)
		assert.InDelta(t, 2.718, v, 0.001)
	})

	t.Run("Parse[bool]", func(t *testing.T) {
		v, err := Parse[bool]("true")
		assert.NoError(t, err)
		assert.True(t, v)
	})

	t.Run("Parse custom FuncBinder", func(t *testing.T) {
		timeBinder := NewFuncBinder(func(s string) (time.Time, error) {
			return time.Parse("2006-01-02", s)
		})
		v, err := Parse[time.Time]("2026-08-20", WrapBinder(timeBinder))
		assert.NoError(t, err)
		assert.Equal(t, 2026, v.Year())
	})

	t.Run("ParseSlice[int]", func(t *testing.T) {
		v, err := ParseSlice[int]([]string{"1", "2", "3"})
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, v)
	})

	t.Run("Parse Invalid Type", func(t *testing.T) {
		type unhandledType struct{ X string }
		_, err := Parse[unhandledType]("foo")
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

type CustomFloat float64
type CustomUint uint32
type CustomSliceItem int

type ComprehensiveCustomForm struct {
	ID       CustomAge          `form:"id"`
	Score    CustomFloat        `form:"score"`
	Port     CustomUint         `form:"port"`
	Name     CustomUserID       `form:"name"`
	Active   CustomFlag         `form:"active"`
	PtrID    *CustomAge         `form:"ptr_id"`
	PtrName  *CustomUserID      `form:"ptr_name"`
	Items    []CustomSliceItem  `form:"items"`
	PtrItems []*CustomSliceItem `form:"ptr_items"`
}

type CustomWithUnmarshaler int

func (c *CustomWithUnmarshaler) UnmarshalBind(s string) error {
	*c = 9999 // custom logic to verify UnmarshalBind is called over default int parser
	return nil
}

type FormWithCustomUnmarshaler struct {
	Normal CustomAge             `form:"normal"`
	Custom CustomWithUnmarshaler `form:"custom"`
}

func TestCustomTypesAutoMatch(t *testing.T) {
	form := url.Values{}
	form.Set("id", "42")
	form.Set("score", "99.75")
	form.Set("port", "8080")
	form.Set("name", "Gopher")
	form.Set("active", "true")
	form.Set("ptr_id", "100")
	form.Set("ptr_name", "PointerGopher")
	form.Add("items", "10")
	form.Add("items", "20")
	form.Add("ptr_items", "30")
	form.Add("ptr_items", "40")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	res, err := Bind[ComprehensiveCustomForm](req)
	assert.NoError(t, err)
	assert.Equal(t, CustomAge(42), res.ID)
	assert.InDelta(t, 99.75, float64(res.Score), 0.001)
	assert.Equal(t, CustomUint(8080), res.Port)
	assert.Equal(t, CustomUserID("Gopher"), res.Name)
	assert.True(t, bool(res.Active))

	assert.NotNil(t, res.PtrID)
	assert.Equal(t, CustomAge(100), *res.PtrID)
	assert.NotNil(t, res.PtrName)
	assert.Equal(t, CustomUserID("PointerGopher"), *res.PtrName)

	assert.Equal(t, []CustomSliceItem{10, 20}, res.Items)
	assert.Len(t, res.PtrItems, 2)
	assert.Equal(t, CustomSliceItem(30), *res.PtrItems[0])
	assert.Equal(t, CustomSliceItem(40), *res.PtrItems[1])
}

func TestCustomTypeWithBindUnmarshalerPriority(t *testing.T) {
	form := url.Values{}
	form.Set("normal", "123")
	form.Set("custom", "123")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	res, err := Bind[FormWithCustomUnmarshaler](req)
	assert.NoError(t, err)
	assert.Equal(t, CustomAge(123), res.Normal)
	assert.Equal(t, CustomWithUnmarshaler(9999), res.Custom) // UnmarshalBind was invoked!

	// Test Parse with custom unmarshaler vs normal custom type
	parsedNormal, err := Parse[CustomAge]("456")
	assert.NoError(t, err)
	assert.Equal(t, CustomAge(456), parsedNormal)

	parsedCustom, err := Parse[CustomWithUnmarshaler]("456")
	assert.NoError(t, err)
	assert.Equal(t, CustomWithUnmarshaler(9999), parsedCustom)

	// Test Value / Values helpers
	val, err := Value[CustomAge](req, "normal")
	assert.NoError(t, err)
	assert.Equal(t, CustomAge(123), val)
}

func BenchmarkGenericBind(b *testing.B) {
	form := url.Values{}
	form.Set("id", "9876543210")
	form.Set("name", "Jane Doe")
	form.Set("age", "28")
	form.Set("active", "true")
	form.Set("score", "98.5")
	form.Add("tags", "go")
	form.Add("ratings", "5")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	b.Run("Bind[T]", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Bind[GenericUserForm](req)
		}
	})

	b.Run("For[T].Bind", func(b *testing.B) {
		tb := For[GenericUserForm]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tb.Bind(req)
		}
	})

	b.Run("Value[int]", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Value[int](req, "age")
		}
	})
}
