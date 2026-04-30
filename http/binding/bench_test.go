package binding

import (
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
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
	binder := NewFormBinder()
	form := url.Values{}
	form.Set("name", "John Doe")
	form.Set("age", "30")
	form.Set("active", "true")
	form.Set("score", "95.5")
	form.Add("items", "1")
	form.Add("items", "2")
	form.Add("items", "3")
	form.Set("address", "123 Main St")

	req, _ := http.NewRequest("POST", "/", nil)
	req.Form = form

	b.Run("Cached", func(b *testing.B) {
		var dst BenchStruct
		// Warm up the cache
		_ = binder.Bind(req, &dst)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = binder.Bind(req, &dst)
		}
	})

	b.Run("NoCache_MetadataOnly", func(b *testing.B) {
		var dst BenchStruct
		rt := reflect.TypeOf(&dst).Elem()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Manually clear the metadata cache for this type
			metadataCache.Delete(rt)
			_ = binder.Bind(req, &dst)
		}
	})

	b.Run("NoCache_Full", func(b *testing.B) {
		var dst BenchStruct
		rt := reflect.TypeOf(&dst).Elem()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Clear both caches
			metadataCache.Delete(rt)
			binder.cache = sync.Map{}
			_ = binder.Bind(req, &dst)
		}
	})
}
