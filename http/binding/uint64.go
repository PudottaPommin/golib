package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[uint64] = (*Uint64Binder)(nil)

type Uint64Binder struct{}

func (m Uint64Binder) guard(a any) bool {
	switch a.(type) {
	case *uint64, *[]uint64:
		return true
	default:
		return false
	}
}

func (m Uint64Binder) Mappable(a any) bool {
	switch a.(type) {
	case uint64, *uint64, []uint64, *[]uint64:
		return true
	default:
		return false
	}
}

func (m Uint64Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*uint64))
}

func (m Uint64Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]uint64))
}

func (m Uint64Binder) BindT(src string, dst *uint64) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseUint(src, 10, 64)
	if err != nil {
		return errors.New("failed to bind value to uint64")
	}
	*dst = v
	return nil
}

func (m Uint64Binder) BindManyT(src []string, dst *[]uint64) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]uint64, len(src))
	for idx, v := range src {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return errors.New("failed to bind value to uint64")
		}
		arr[idx] = i
	}
	*dst = arr
	return nil
}
