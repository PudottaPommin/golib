package binding_test

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/pudottapommin/golib/http/binding"
)

type BenchStruct struct {
	Name    string  `form:"name"`
	Age     int     `form:"age"`
	Active  bool    `form:"active"`
	Score   float64 `form:"score"`
	Items   []int   `form:"items"`
	Address string  `form:"address"`
}

func BenchmarkFormBinder_Bind(b *testing.B) {
	binder := NewFormBinder[BenchStruct]()
	form := url.Values{}
	form.Set("name", "John Doe")
	form.Set("age", "30")
	form.Set("active", "true")
	form.Set("score", "95.5")
	form.Add("items", "1")
	form.Add("items", "2")
	form.Add("items", "3")
	form.Set("address", "123 Main St")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	b.Run("BindTo", func(b *testing.B) {
		var dst BenchStruct
		_ = binder.BindTo(req, &dst)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = binder.BindTo(req, &dst)
		}
	})

	b.Run("Bind", func(b *testing.B) {
		_, _ = binder.Bind(req)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = binder.Bind(req)
		}
	})

	b.Run("NewInstance", func(b *testing.B) {
		var dst BenchStruct
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fresh := NewFormBinder[BenchStruct]()
			_ = fresh.BindTo(req, &dst)
		}
	})
}
