package memory

import (
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type cacheItem struct {
	Id    int
	Value string
}

func TestCache_Set(t *testing.T) {
	t.Parallel()

	c := New[cacheItem]()
	var i int
	for range 10 {
		i++
		c.Set(strconv.Itoa(i), cacheItem{Id: i, Value: "test"})
	}
	v, ok := c.Get("10")
	require.True(t, ok)
	require.Equal(t, 10, v.Id)
	v, ok = c.Get("1")
	require.True(t, ok)
	// This GC should not remove a weak pointer (current V is a strong pointer)
	runtime.GC()
	require.Equal(t, 1, v.Id)
	require.Equal(t, "test", v.Value)

	// After this GC reference should be gone and done
	runtime.GC()
	v, ok = c.Get("1")
	require.False(t, ok)
	require.Nil(t, v)
}
