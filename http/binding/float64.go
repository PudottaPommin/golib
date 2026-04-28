package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[float64] = (*Float64Binder)(nil)

type Float64Binder struct{}

func (m Float64Binder) guard(a any) bool {
	switch a.(type) {
	case *float64, *[]float64:
		return true
	default:
		return false
	}
}

func (m Float64Binder) Mappable(a any) bool {
	switch a.(type) {
	case float64, *float64, []float64, *[]float64:
		return true
	default:
		return false
	}
}

func (m Float64Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*float64))
}

func (m Float64Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]float64))
}

func (m Float64Binder) BindT(src string, dst *float64) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseFloat(src, 64)
	if err != nil {
		return errors.New("failed to bind value to float64")
	}
	*dst = v
	return nil
}

func (m Float64Binder) BindManyT(src []string, dst *[]float64) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]float64, len(src))
	for idx, v := range src {
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return errors.New("failed to bind value to float64")
		}
		arr[idx] = i
	}
	*dst = arr
	return nil
}
