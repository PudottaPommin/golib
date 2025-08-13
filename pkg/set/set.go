package set

import (
	"iter"
	"maps"
)

type Set[T comparable] map[T]struct{}

func SetOf[T comparable](values ...T) Set[T] {
	s := make(Set[T])
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) IsSubset(values ...T) bool {
	for _, v := range values {
		if !s.Contains(v) {
			return false
		}
	}
	return true
}

func (s Set[T]) Add(v T) bool {
	if !s.Contains(v) {
		s[v] = struct{}{}
		return true
	}
	return false
}

func (s Set[T]) AddMultiple(values ...T) []bool {
	added := make([]bool, len(values))
	for i := range values {
		added[i] = s.Add(values[i])
	}
	return added
}

func (s Set[T]) Remove(v T) bool {
	if s.Contains(v) {
		delete(s, v)
		return true
	}
	return false
}

func (s Set[T]) Values() []T {
	values := make([]T, 0, len(s))
	for v := range s {
		values = append(values, v)
	}
	return values
}

func (s Set[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

func (s Set[T]) Clone() Set[T] {
	return maps.Clone(s)
}
