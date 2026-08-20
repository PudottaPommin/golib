package binding

import (
	"fmt"
)

var _ Binder[string] = (*FuncBinder[string])(nil)

// FuncBinder is a generic binder backed by a custom parsing function.
type FuncBinder[T any] struct {
	fn func(string) (T, error)
}

// NewFuncBinder creates a new FuncBinder for type T using the provided parsing function.
func NewFuncBinder[T any](fn func(string) (T, error)) FuncBinder[T] {
	return FuncBinder[T]{fn: fn}
}

func (m FuncBinder[T]) Bind(s string, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	val, err := m.fn(s)
	if err != nil {
		return err
	}
	*dst = val
	return nil
}

func (m FuncBinder[T]) BindMany(s []string, dst *[]T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if cap(*dst) < len(s) {
		*dst = make([]T, len(s))
	} else {
		*dst = (*dst)[:len(s)]
	}
	for i, v := range s {
		val, err := m.fn(v)
		if err != nil {
			return fmt.Errorf("failed to bind value at index %d: %w", i, err)
		}
		(*dst)[i] = val
	}
	return nil
}

// BindT is provided for backward compatibility.
func (m FuncBinder[T]) BindT(src string, dst *T) error {
	return m.Bind(src, dst)
}

// BindManyT is provided for backward compatibility.
func (m FuncBinder[T]) BindManyT(src []string, dst *[]T) error {
	return m.BindMany(src, dst)
}
