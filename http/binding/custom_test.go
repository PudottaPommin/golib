package binding

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type customType struct {
	val string
}

func (c *customType) UnmarshalBind(src string) error {
	if src == "error" {
		return fmt.Errorf("custom error")
	}
	c.val = "custom:" + src
	return nil
}

func TestUnmarshalerBinder(t *testing.T) {
	binder := CustomBinder[customType, *customType]{}

	t.Run("Single", func(t *testing.T) {
		var dst customType
		err := binder.Bind("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, "custom:foo", dst.val)
	})

	t.Run("Slice", func(t *testing.T) {
		var dst []customType
		err := binder.BindMany([]string{"foo", "bar"}, &dst)
		assert.NoError(t, err)
		assert.Len(t, dst, 2)
		assert.Equal(t, "custom:foo", dst[0].val)
		assert.Equal(t, "custom:bar", dst[1].val)
	})

	t.Run("Error", func(t *testing.T) {
		var dst customType
		err := binder.Bind("error", &dst)
		assert.Error(t, err)
	})
}

func TestAnyUnmarshalerBinder(t *testing.T) {
	binder := AnyBinder{}

	t.Run("Mappable", func(t *testing.T) {
		assert.True(t, binder.Mappable(new(customType)))
		assert.True(t, binder.Mappable(new([]customType)))
		assert.False(t, binder.Mappable(customType{}))
		assert.False(t, binder.Mappable(123))
	})

	t.Run("Bind", func(t *testing.T) {
		var dst customType
		err := binder.Bind("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, "custom:foo", dst.val)
	})

	t.Run("BindMany", func(t *testing.T) {
		var dst []customType
		err := binder.BindMany([]string{"foo", "bar"}, &dst)
		assert.NoError(t, err)
		assert.Len(t, dst, 2)
		assert.Equal(t, "custom:foo", dst[0].val)
		assert.Equal(t, "custom:bar", dst[1].val)
	})
}

func BenchmarkUnmarshalerBinder(b *testing.B) {
	generic := CustomBinder[customType, *customType]{}
	any := AnyBinder{}

	src := "bench"
	var dst customType

	b.Run("Generic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = generic.BindT(src, &dst)
		}
	})

	b.Run("Any", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = any.Bind(src, &dst)
		}
	})
}

type BenchUnmarshalerStruct struct {
	Custom customType `form:"custom"`
}

func BenchmarkFormBinder_Unmarshaler(b *testing.B) {
	b.Run("Any", func(b *testing.B) {
		// Use the default which includes AnyUnmarshalerBinder
		binder := NewFormBinder()
		form := url.Values{}
		form.Set("custom", "data")
		req, _ := http.NewRequest("POST", "/", nil)
		req.Form = form

		var dst BenchUnmarshalerStruct
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = binder.Bind(req, &dst)
		}
	})

	b.Run("Generic", func(b *testing.B) {
		// Explicitly add the generic binder
		binder := NewFormBinder()
		binder.AddBinder(CustomBinder[customType, *customType]{})
		form := url.Values{}
		form.Set("custom", "data")
		req, _ := http.NewRequest("POST", "/", nil)
		req.Form = form

		var dst BenchUnmarshalerStruct
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = binder.Bind(req, &dst)
		}
	})
}
