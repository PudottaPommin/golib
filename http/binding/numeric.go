package binding

import (
	"fmt"
	"reflect"
	"strconv"
)

type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type UnsignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type Numeric interface {
	SignedInteger | UnsignedInteger | Float
}

var _ Binder[int] = (*NumericBinder[int])(nil)

// NumericBinder binds string values to any numeric type (signed int, unsigned int, float).
type NumericBinder[T Numeric] struct{}

func (m NumericBinder[T]) Bind(s string, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if s == "" {
		return ErrValueIsZero
	}
	v, err := m.parseNumber(s)
	if err != nil {
		return err
	}
	*dst = v
	return nil
}

func (m NumericBinder[T]) BindMany(s []string, dst *[]T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if cap(*dst) < len(s) {
		*dst = make([]T, len(s))
	} else {
		*dst = (*dst)[:len(s)]
	}
	for i, v := range s {
		val, err := m.parseNumber(v)
		if err != nil {
			return err
		}
		(*dst)[i] = val
	}
	return nil
}

// BindT is provided for backward compatibility.
func (m NumericBinder[T]) BindT(src string, dst *T) error {
	return m.Bind(src, dst)
}

// BindManyT is provided for backward compatibility.
func (m NumericBinder[T]) BindManyT(src []string, dst *[]T) error {
	return m.BindMany(src, dst)
}

func (m NumericBinder[T]) parseNumber(s string) (T, error) {
	var zero T
	t := reflect.TypeFor[T]()
	kind := t.Kind()
	bitSize := t.Bits()

	switch kind {
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, bitSize)
		if err != nil {
			return zero, fmt.Errorf("failed to bind value to %s", t.String())
		}
		return T(v), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 10, bitSize)
		if err != nil {
			return zero, fmt.Errorf("failed to bind value to %s", t.String())
		}
		return T(v), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 10, bitSize)
		if err != nil {
			return zero, fmt.Errorf("failed to bind value to %s", t.String())
		}
		return T(v), nil
	default:
		return zero, fmt.Errorf("failed to bind value to %s", t.String())
	}
}

// IntBinder and aliases below provide backward-compatible type aliases for all integer, unsigned integer, and float types.
type (
	IntBinder     = NumericBinder[int]
	Int8Binder    = NumericBinder[int8]
	Int16Binder   = NumericBinder[int16]
	Int32Binder   = NumericBinder[int32]
	Int64Binder   = NumericBinder[int64]
	UintBinder    = NumericBinder[uint]
	Uint8Binder   = NumericBinder[uint8]
	Uint16Binder  = NumericBinder[uint16]
	Uint32Binder  = NumericBinder[uint32]
	Uint64Binder  = NumericBinder[uint64]
	Float32Binder = NumericBinder[float32]
	Float64Binder = NumericBinder[float64]

	SignedIntBinder[T SignedInteger]     = NumericBinder[T]
	UnsignedIntBinder[T UnsignedInteger] = NumericBinder[T]
	FloatBinder[T Float]                 = NumericBinder[T]
)
