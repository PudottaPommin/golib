package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[int] = (*IntBinder)(nil)

type IntBinder struct{}

func (m IntBinder) guard(a any) bool {
	switch a.(type) {
	case *int, *[]int:
		return true
	default:
		return false
	}
}

func (m IntBinder) Mappable(a any) bool {
	switch a.(type) {
	case int, *int, []int, *[]int:
		return true
	default:
		return false
	}
}

func (m IntBinder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*int))
}

func (m IntBinder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]int))
}

func (m IntBinder) BindT(src string, dst *int) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.Atoi(src)
	if err != nil {
		return errors.New("failed to bind value to int")
	}
	*dst = v
	return nil
}

func (m IntBinder) BindManyT(src []string, dst *[]int) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]int, len(src))
	for idx, v := range src {
		i, err := strconv.Atoi(v)
		if err != nil {
			return errors.New("failed to bind value to int")
		}
		arr[idx] = i
	}
	*dst = arr
	return nil
}
