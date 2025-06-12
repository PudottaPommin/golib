package hasher

import (
	"testing"
)

func BenchmarkHasher(b *testing.B) {
	b.Run("Argon2ID:hash", func(b *testing.B) {
		h := NewArgon2id()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = h.Hash("abcdefghijklmnopqrstuvwqyz")
			}
		})
	})

	b.Run("Pbkdf2:hash", func(b *testing.B) {
		h := NewPbkdf2()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = h.Hash("abcdefghijklmnopqrstuvwqyz")
			}
		})
	})
}
