package id

import (
	"bytes"
	"sync"
	"testing"

	"github.com/gofrs/uuid/v5"
)

func TestID(t *testing.T) {
	t.Parallel()
	id := New()
	if len(id.Bytes()) != idSize {
		t.Errorf("id: size of id was not %d", idSize)
	}
	if bytes.Equal(id.Bytes(), New().Bytes()) {
		t.Error("collision")
	}
}

func TestIDUniqueness(t *testing.T) {
	t.Parallel()
	var (
		wg  sync.WaitGroup
		ids sync.Map
	)

	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			for i := 0; i < 10_000; i++ {
				id := New()
				if _, ok := ids.Load(id); ok {
					t.Errorf("collision: %s", id)
				}
				ids.Store(id, struct{}{})
			}
		}()
	}
	wg.Wait()
}

func BenchmarkID(b *testing.B) {
	b.Run("ID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New()
		}
		// b.RunParallel(func(pb *testing.PB) {
		// 	for pb.Next() {
		// 		_ = New()
		// 	}
		// })
	})
	b.Run("Gofrs/UUIDv7", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = uuid.NewV7()
		}
		// b.RunParallel(func(pb *testing.PB) {
		// 	for pb.Next() {
		// 		_, _ = uuid.NewV7()
		// 	}
		// })
	})
}
