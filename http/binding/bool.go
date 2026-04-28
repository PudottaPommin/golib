package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[bool] = (*BoolBinder)(nil)

type BoolBinder struct{}

func (m BoolBinder) guard(a any) bool {
	switch a.(type) {
	case *bool, *[]bool:
		return true
	default:
		return false
	}
}

func (m BoolBinder) Mappable(a any) bool {
	switch a.(type) {
	case bool, *bool, []bool, *[]bool:
		return true
	default:
		return false
	}
}

func (m BoolBinder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*bool))
}

func (m BoolBinder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]bool))
}

func (m BoolBinder) BindT(src string, dst *bool) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseBool(src)
	if err != nil {
		return errors.New("failed to bind value to bool")
	}
	*dst = v
	return nil
}

func (m BoolBinder) BindManyT(src []string, dst *[]bool) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]bool, len(src))
	for idx, v := range src {
		i, err := strconv.ParseBool(v)
		if err != nil {
			return errors.New("failed to bind value to bool")
		}
		arr[idx] = i
	}
	*dst = arr
	return nil
}
