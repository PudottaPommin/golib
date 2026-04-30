package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[int64] = (*Int64Binder)(nil)

type Int64Binder struct{}

func (m Int64Binder) guard(a any) bool {
	switch a.(type) {
	case *int64, *[]int64:
		return true
	default:
		return false
	}
}

func (m Int64Binder) Mappable(a any) bool {
	switch a.(type) {
	case *int64, *[]int64:
		return true
	default:
		return false
	}
}

func (m Int64Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*int64))
}

func (m Int64Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]int64))
}

func (m Int64Binder) BindT(src string, dst *int64) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseInt(src, 10, 64)
	if err != nil {
		return errors.New("failed to bind value to int64")
	}
	*dst = v
	return nil
}

func (m Int64Binder) BindManyT(src []string, dst *[]int64) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]int64, len(src))
	for idx, v := range src {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return errors.New("failed to bind value to int64")
		}
		arr[idx] = i
	}
	*dst = arr
	return nil
}
