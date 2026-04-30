package binding

import (
	"fmt"
	"reflect"
)

// AnyBinder is a fallback binder that uses reflection to detect BindUnmarshaler at runtime.
type AnyBinder struct{}

func (m AnyBinder) Mappable(a any) bool {
	if a == nil {
		return false
	}
	t := reflect.TypeOf(a)
	if t.Kind() != reflect.Ptr {
		return false
	}

	et := t.Elem()
	if et.Kind() == reflect.Slice {
		return reflect.PointerTo(et.Elem()).Implements(reflect.TypeOf((*BindUnmarshaler)(nil)).Elem())
	}

	return t.Implements(reflect.TypeOf((*BindUnmarshaler)(nil)).Elem())
}

func (m AnyBinder) Bind(src string, dst any) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	bu, ok := dst.(BindUnmarshaler)
	if !ok {
		return ErrorDestinationTypeInvalid
	}
	return bu.UnmarshalBind(src)
}

func (m AnyBinder) BindMany(src []string, dst any) error {
	if dst == nil {
		return ErrorDestinationNil
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return ErrorDestinationTypeInvalid
	}

	slice := rv.Elem()
	elemType := slice.Type().Elem()
	ptrElemType := reflect.PointerTo(elemType)

	if !ptrElemType.Implements(reflect.TypeOf((*BindUnmarshaler)(nil)).Elem()) {
		return ErrorDestinationTypeInvalid
	}

	newSlice := reflect.MakeSlice(slice.Type(), len(src), len(src))
	for i, v := range src {
		elem := newSlice.Index(i)
		var bu BindUnmarshaler
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				elem.Set(reflect.New(elem.Type().Elem()))
			}
			bu = elem.Interface().(BindUnmarshaler)
		} else {
			bu = elem.Addr().Interface().(BindUnmarshaler)
		}

		if err := bu.UnmarshalBind(v); err != nil {
			return fmt.Errorf("failed to bind value at index %d: %w", i, err)
		}
	}

	slice.Set(newSlice)
	return nil
}
