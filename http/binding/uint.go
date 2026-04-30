package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[uint] = (*UintBinder)(nil)

type UintBinder struct{}

func (m UintBinder) guard(a any) bool {
	switch a.(type) {
	case *uint, *[]uint:
		return true
	default:
		return false
	}
}

func (m UintBinder) Mappable(a any) bool {
	switch a.(type) {
	case *uint, *[]uint:
		return true
	default:
		return false
	}
}

func (m UintBinder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*uint))
}

func (m UintBinder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]uint))
}

func (m UintBinder) BindT(src string, dst *uint) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseUint(src, 10, 0)
	if err != nil {
		return errors.New("failed to bind value to uint")
	}
	*dst = uint(v)
	return nil
}

func (m UintBinder) BindManyT(src []string, dst *[]uint) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]uint, len(src))
	for idx, v := range src {
		i, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return errors.New("failed to bind value to uint")
		}
		arr[idx] = uint(i)
	}
	*dst = arr
	return nil
}
