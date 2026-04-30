package binding

import (
	"errors"
	"strconv"
)

var _ GenericBinder[float32] = (*Float32Binder)(nil)

type Float32Binder struct{}

func (m Float32Binder) guard(a any) bool {
	switch a.(type) {
	case *float32, *[]float32:
		return true
	default:
		return false
	}
}

func (m Float32Binder) Mappable(a any) bool {
	switch a.(type) {
	case *float32, *[]float32:
		return true
	default:
		return false
	}
}

func (m Float32Binder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*float32))
}

func (m Float32Binder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]float32))
}

func (m Float32Binder) BindT(src string, dst *float32) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	v, err := strconv.ParseFloat(src, 32)
	if err != nil {
		return errors.New("failed to bind value to float32")
	}
	*dst = float32(v)
	return nil
}

func (m Float32Binder) BindManyT(src []string, dst *[]float32) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	arr := make([]float32, len(src))
	for idx, v := range src {
		i, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return errors.New("failed to bind value to float32")
		}
		arr[idx] = float32(i)
	}
	*dst = arr
	return nil
}
