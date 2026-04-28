package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[uint8] = (*Uint8Binder)(nil)

type Uint8Binder struct{}

func (m Uint8Binder) guard(a any) bool {
	switch a.(type) {
	case *uint8, *[]uint8:
		return true
	default:
		return false
	}
}

func (m Uint8Binder) Mappable(a any) bool {
	switch a.(type) {
	case uint8, *uint8, []uint8, *[]uint8:
		return true
	default:
		return false
	}
}

func (m Uint8Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*uint8))
}

func (m Uint8Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]uint8))
}

func (m Uint8Binder) BindT(src string, dst *uint8) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseUint(src, 10, 8)
	if err != nil {
		return errors.New("failed to bind value to uint8")
	}
	*dst = uint8(v)
	return nil
}

func (m Uint8Binder) BindManyT(src []string, dst *[]uint8) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]uint8, len(src))
	for idx, v := range src {
		i, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return errors.New("failed to bind value to uint8")
		}
		arr[idx] = uint8(i)
	}
	*dst = arr
	return nil
}
