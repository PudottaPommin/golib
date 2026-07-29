package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[uint32] = (*Uint32Binder)(nil)

type Uint32Binder struct{}

func (m Uint32Binder) guard(a any) bool {
	switch a.(type) {
	case *uint32, *[]uint32:
		return true
	default:
		return false
	}
}

func (m Uint32Binder) Mappable(a any) bool {
	switch a.(type) {
	case *uint32, *[]uint32:
		return true
	default:
		return false
	}
}

func (m Uint32Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*uint32))
}

func (m Uint32Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]uint32))
}

func (m Uint32Binder) BindT(src string, dst *uint32) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if src == "" {
		return ErrValueIsZero
	}
	v, err := strconv.ParseUint(src, 10, 32)
	if err != nil {
		return errors.New("failed to bind value to uint32")
	}
	*dst = uint32(v)
	return nil
}

func (m Uint32Binder) BindManyT(src []string, dst *[]uint32) error {
	if dst == nil {
		return ErrDestinationNil
	}
	arr := make([]uint32, len(src))
	for idx, v := range src {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return errors.New("failed to bind value to uint32")
		}
		arr[idx] = uint32(i)
	}
	*dst = arr
	return nil
}
