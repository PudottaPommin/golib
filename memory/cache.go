package memory

import (
	"sync"
	"weak"
)

type (
	Cache[T any] interface {
		Get(key string) (v *T, ok bool)
		Set(key string, value T)
		Delete(key string)
		Count() int
	}
	cache[T any] struct {
		store sync.Map
	}
)

var _ Cache[int] = (*cache[int])(nil)

func New[T any]() Cache[T] {
	return &cache[T]{}
}

func (c *cache[T]) Get(key string) (v *T, ok bool) {
	ptr, ok := c.store.Load(key)
	if !ok {
		return v, false
	}
	v = ptr.(weak.Pointer[T]).Value()
	if v == nil {
		c.store.Delete(key)
		return v, false
	}
	return v, true
}

func (c *cache[T]) Set(key string, value T) {
	c.store.Store(key, weak.Make(&value))
}

func (c *cache[T]) Delete(key string) {
	c.store.Delete(key)
}

func (c *cache[T]) Count() (i int) {
	c.store.Range(func(_, _ any) bool {
		i++
		return true
	})
	return
}
