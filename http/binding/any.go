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
	if t.Kind() != reflect.Pointer {
		return false
	}

	et := t.Elem()
	if et.Kind() == reflect.Slice {
		return reflect.PointerTo(et.Elem()).Implements(reflect.TypeFor[BindUnmarshaler]())
	}

	return t.Implements(reflect.TypeFor[BindUnmarshaler]())
}

func (m AnyBinder) Bind(src string, dst any) error {
	if dst == nil {
		return ErrDestinationNil
	}
	bu, ok := dst.(BindUnmarshaler)
	if !ok {
		return ErrDestinationTypeInvalid
	}
	return bu.UnmarshalBind(src)
}

func (m AnyBinder) BindMany(src []string, dst any) error {
	if dst == nil {
		return ErrDestinationNil
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Slice {
		return ErrDestinationTypeInvalid
	}

	slice := rv.Elem()
	elemType := slice.Type().Elem()
	ptrElemType := reflect.PointerTo(elemType)

	if !ptrElemType.Implements(reflect.TypeFor[BindUnmarshaler]()) {
		return ErrDestinationTypeInvalid
	}

	newSlice := reflect.MakeSlice(slice.Type(), len(src), len(src))
	for i, v := range src {
		elem := newSlice.Index(i)
		var bu BindUnmarshaler
		if elem.Kind() == reflect.Pointer {
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
