package binding

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type (
	Binder interface {
		Mappable(any) bool
		Bind(src string, dst any) error
		BindMany(src []string, dst any) error
	}
	GenericBinder[T any] interface {
		Binder
		BindT(src string, dst *T) error
		BindManyT(src []string, dst *[]T) error
	}
)

var (
	ErrorDestinationTypeInvalid = errors.New("invalid destination type for binder")
	ErrorDestinationNil         = errors.New("destination cannot be nil")
)

type FormBinder struct {
	binders []Binder
}

func NewFormBinder() *FormBinder {
	return &FormBinder{binders: DefaultBinders}
}

func (fb *FormBinder) AddBinder(b ...Binder) *FormBinder {
	fb.binders = append(fb.binders, b...)
	return fb
}

func (fb *FormBinder) Bind(r *http.Request, dst any) error {
	if r.Form == nil {
		if err := r.ParseMultipartForm(20 << 32); err != nil {
			return err
		}
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("dst must be a pointer to a struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		name := strings.TrimSpace(parts[0])
		required := false
		for _, part := range parts[1:] {
			if strings.TrimSpace(part) == "required" {
				required = true
			}
		}
		values, ok := r.Form[name]
		if required && (!ok || len(values) == 0) {
			return fmt.Errorf("missing required field %q", name)
		}
		if !ok || len(values) == 0 {
			continue
		}

		fieldVal := rv.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		if err := fb.bindField(fieldVal, values); err != nil {
			return err
		}
	}

	return nil
}

func (fb *FormBinder) bindField(field reflect.Value, values []string) error {
	for _, b := range fb.binders {
		addr := field.Addr().Interface()
		if b.Mappable(addr) {
			if field.Type().Kind() == reflect.Slice {
				return b.BindMany(values, addr)
			}
			return b.Bind(values[len(values)-1], addr)
		}
	}
	return nil
}

var (
	defaultBinders = []Binder{
		StringBinder{},
		BoolBinder{},
		Float32Binder{},
		Float64Binder{},
		IntBinder{},
		Int8Binder{},
		Int16Binder{},
		Int32Binder{},
		Int64Binder{},
		UintBinder{},
		Uint8Binder{},
		Uint16Binder{},
		Uint32Binder{},
		Uint64Binder{},
	}
	DefaultBinders = defaultBinders
)

func AddBinder(b Binder) {
	DefaultBinders = append(DefaultBinders, b)
}
