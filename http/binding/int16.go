package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[int16] = (*Int16Binder)(nil)

type Int16Binder struct{}

func (m Int16Binder) guard(a any) bool {
	switch a.(type) {
	case *int16, *[]int16:
		return true
	default:
		return false
	}
}

func (m Int16Binder) Mappable(a any) bool {
	switch a.(type) {
	case int16, *int16, []int16, *[]int16:
		return true
	default:
		return false
	}
}

func (m Int16Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*int16))
}

func (m Int16Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]int16))
}

func (m Int16Binder) BindT(src string, dst *int16) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseInt(src, 10, 16)
	if err != nil {
		return errors.New("failed to bind value to int16")
	}
	*dst = int16(v)
	return nil
}

func (m Int16Binder) BindManyT(src []string, dst *[]int16) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]int16, len(src))
	for idx, v := range src {
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return errors.New("failed to bind value to int16")
		}
		arr[idx] = int16(i)
	}
	*dst = arr
	return nil
}
