package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathJoin(t *testing.T) {
	t.Parallel()
	pairs := []struct {
		elements []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{""}, ""},
		{[]string{".."}, "."},
		{[]string{"za"}, "za"},
		{[]string{"/za/"}, "za"},
		{[]string{"a", "b", "c"}, "a/b/c"},
		{[]string{"a", "b", ""}, "a/b"},
	}

	for _, p := range pairs {
		p := p
		t.Run(p.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, p.expected, PathJoinRel(p.elements...), "elements %s", p.elements)
		})
		t.Run(p.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, p.expected, PathJoinRelX(p.elements...), "X elements %s", p.elements)
		})
	}
}
