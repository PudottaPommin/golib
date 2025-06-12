package golib

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	t.Run("String", func(t *testing.T) {

		sp := NewPool[string](func() string {
			return "test"
		})
		v := sp.Get()
		assert.Equal(t, "test", v)
		sp.Put(v)
	})

	t.Run("Buffer", func(t *testing.T) {
		sp := NewPool[*bytes.Buffer](func() *bytes.Buffer {
			return new(bytes.Buffer)
		})
		v := sp.Get()

		v.WriteByte(1)
		assert.Equal(t, byte(1), v.Bytes()[0])
		assert.Equal(t, 1, v.Len())

		v.Reset()
		sp.Put(v)

		assert.Equal(t, 0, v.Len())
	})
}
