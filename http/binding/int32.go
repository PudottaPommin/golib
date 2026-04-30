package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[int32] = (*Int32Binder)(nil)

type Int32Binder struct{}

func (m Int32Binder) guard(a any) bool {
	switch a.(type) {
	case *int32, *[]int32:
		return true
	default:
		return false
	}
}

func (m Int32Binder) Mappable(a any) bool {
	switch a.(type) {
	case *int32, *[]int32:
		return true
	default:
		return false
	}
}

func (m Int32Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*int32))
}

func (m Int32Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]int32))
}

func (m Int32Binder) BindT(src string, dst *int32) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseInt(src, 10, 32)
	if err != nil {
		return errors.New("failed to bind value to int32")
	}
	*dst = int32(v)
	return nil
}

func (m Int32Binder) BindManyT(src []string, dst *[]int32) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]int32, len(src))
	for idx, v := range src {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return errors.New("failed to bind value to int32")
		}
		arr[idx] = int32(i)
	}
	*dst = arr
	return nil
}
