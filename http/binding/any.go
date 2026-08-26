package binding

import (
	"fmt"
	"reflect"
)

// AnyBinder is a fallback binder that uses reflection to detect BindUnmarshaler at runtime.
type AnyBinder struct{}

func (m AnyBinder) MappableType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	target := t
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	if target.Kind() == reflect.Slice {
		target = target.Elem()
	}
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	return implementsUnmarshaler(target)
}

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
		elem := et.Elem()
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		return implementsUnmarshaler(elem)
	}
	if et.Kind() == reflect.Pointer {
		return implementsUnmarshaler(et.Elem())
	}

	return implementsUnmarshaler(et)
}

func (m AnyBinder) Bind(src string, dst any) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if bu, ok := dst.(BindUnmarshaler); ok {
		return bu.UnmarshalBind(src)
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Pointer {
		elem := rv.Elem()
		if elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		if bu, ok := elem.Interface().(BindUnmarshaler); ok {
			return bu.UnmarshalBind(src)
		}
	}
	return ErrDestinationTypeInvalid
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
