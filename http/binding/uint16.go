package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[uint16] = (*Uint16Binder)(nil)

type Uint16Binder struct{}

func (m Uint16Binder) guard(a any) bool {
	switch a.(type) {
	case *uint16, *[]uint16:
		return true
	default:
		return false
	}
}

func (m Uint16Binder) Mappable(a any) bool {
	switch a.(type) {
	case uint16, *uint16, []uint16, *[]uint16:
		return true
	default:
		return false
	}
}

func (m Uint16Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*uint16))
}

func (m Uint16Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]uint16))
}

func (m Uint16Binder) BindT(src string, dst *uint16) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseUint(src, 10, 16)
	if err != nil {
		return errors.New("failed to bind value to uint16")
	}
	*dst = uint16(v)
	return nil
}

func (m Uint16Binder) BindManyT(src []string, dst *[]uint16) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]uint16, len(src))
	for idx, v := range src {
		i, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return errors.New("failed to bind value to uint16")
		}
		arr[idx] = uint16(i)
	}
	*dst = arr
	return nil
}
