package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[int8] = (*Int8Binder)(nil)

type Int8Binder struct{}

func (m Int8Binder) guard(a any) bool {
	switch a.(type) {
	case *int8, *[]int8:
		return true
	default:
		return false
	}
}

func (m Int8Binder) Mappable(a any) bool {
	switch a.(type) {
	case int8, *int8, []int8, *[]int8:
		return true
	default:
		return false
	}
}

func (m Int8Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*int8))
}

func (m Int8Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]int8))
}

func (m Int8Binder) BindT(src string, dst *int8) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseInt(src, 10, 8)
	if err != nil {
		return errors.New("failed to bind value to int8")
	}
	*dst = int8(v)
	return nil
}

func (m Int8Binder) BindManyT(src []string, dst *[]int8) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]int8, len(src))
	for idx, v := range src {
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return errors.New("failed to bind value to int8")
		}
		arr[idx] = int8(i)
	}
	*dst = arr
	return nil
}
