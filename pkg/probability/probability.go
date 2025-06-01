package probability

import (
	"errors"
	"math/rand/v2"
	"sort"
	"time"
)

type Generator[T any] struct {
	values  []T
	weights []float64
	size    int
	r       *rand.Rand
}

type randSource struct {
	i uint64
}

func New[T any](v []T, w []float64) (*Generator[T], error) {
	return NewWithSeed(v, w, uint64(time.Now().UnixNano()))
}

func NewWithSeed[T any](v []T, w []float64, seed uint64) (*Generator[T], error) {
	if len(v) != len(w) {
		return nil, errors.New("generator: Weights and Values must have same len")
	}
	wsum := float64(0)
	wghts := make([]float64, len(w))
	for i, w := range w {
		wsum += w
		wghts[i] = wsum
	}
	if wsum-1 >= 1e-4 {
		return nil, errors.New("generator: Sum of weights must be 1.0")
	}

	vals := make([]T, len(v))
	copy(vals, v)

	gen := &Generator[T]{
		values:  vals,
		weights: wghts,
		size:    len(v),
		r:       rand.New(newRandSource(seed)),
	}

	sort.Sort(gen)

	return gen, nil
}

func newRandSource(seed uint64) rand.Source {
	return &randSource{seed}
}

func (g *Generator[T]) Len() int {
	return len(g.values)
}

func (g *Generator[T]) Swap(i, j int) {
	g.values[i], g.values[j] = g.values[j], g.values[i]
	g.weights[i], g.weights[j] = g.weights[j], g.weights[i]
}

func (g *Generator[T]) Less(i, j int) bool {
	return g.weights[i] < g.weights[j]
}

func (s *randSource) Uint64() uint64 {
	return s.i
}
