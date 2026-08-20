package binding

import (
	"errors"
	"strconv"
)

var _ Binder[bool] = (*BoolBinder)(nil)

// BoolGenericBinder binds string inputs to boolean types (including custom ~bool types).
type BoolGenericBinder[T ~bool] struct{}

func (m BoolGenericBinder[T]) Bind(s string, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return errors.New("failed to bind value to bool")
	}
	*dst = T(v)
	return nil
}

func (m BoolGenericBinder[T]) BindMany(s []string, dst *[]T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if cap(*dst) < len(s) {
		*dst = make([]T, len(s))
	} else {
		*dst = (*dst)[:len(s)]
	}
	for i, v := range s {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return errors.New("failed to bind value to bool")
		}
		(*dst)[i] = T(b)
	}
	return nil
}

// BindT is provided for backward compatibility.
func (m BoolGenericBinder[T]) BindT(src string, dst *T) error {
	return m.Bind(src, dst)
}

// BindManyT is provided for backward compatibility.
func (m BoolGenericBinder[T]) BindManyT(src []string, dst *[]T) error {
	return m.BindMany(src, dst)
}

// BoolBinder is an alias for BoolGenericBinder[bool].
type BoolBinder = BoolGenericBinder[bool]
