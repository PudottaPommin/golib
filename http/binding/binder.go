package binding

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
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

	BindUnmarshaler interface {
		UnmarshalBind(string) error
	}
)

var (
	ErrDestinationTypeInvalid = errors.New("binder: invalid destination type for binder")
	ErrDestinationNil         = errors.New("binder: destination cannot be nil")
	ErrValueIsZero            = errors.New("binder: value is zero")
)

var (
	bindersMu      sync.RWMutex
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
		AnyBinder{},
	}
)

func GetDefaultBinders() []Binder {
	bindersMu.RLock()
	defer bindersMu.RUnlock()
	res := make([]Binder, len(defaultBinders))
	copy(res, defaultBinders)
	return res
}

func AddDefaultBinder(b Binder) {
	bindersMu.Lock()
	defer bindersMu.Unlock()
	defaultBinders = append([]Binder{b}, defaultBinders...)
}

type fieldInfo struct {
	index    int
	name     string
	required bool
	// marks field as trimmable (mainly for strings)
	trim bool
}

type structMetadata struct {
	fields []fieldInfo
}

var metadataCache sync.Map // map[reflect.Type]*structMetadata

type (
	FormBinderOptsFn func(binder *FormBinder)
	FormBinder       struct {
		skipDefaults bool
		binders      []Binder
		cache        sync.Map // map[reflect.Type]Binder
	}
)

func FormWithBinders(binders ...Binder) FormBinderOptsFn {
	return func(b *FormBinder) {
		b.AddBinder(binders...)
	}
}

func FormSkipDefaultBinders(skipDefaults bool) FormBinderOptsFn {
	return func(b *FormBinder) {
		b.skipDefaults = skipDefaults
	}
}

// NewFormBinder initializes and returns a new FormBinder instance with default binders.
func NewFormBinder(opts ...FormBinderOptsFn) *FormBinder {
	binder := &FormBinder{binders: GetDefaultBinders()}
	for _, opt := range opts {
		opt(binder)
	}
	if !binder.skipDefaults {
		binder.binders = append(binder.binders, GetDefaultBinders()...)
	}
	return binder
}

// AddBinder appends one or more binders to the FormBinder and clears the resolution cache. Returns the updated FormBinder.
func (fb *FormBinder) AddBinder(b ...Binder) *FormBinder {
	fb.binders = append(b, fb.binders...)
	fb.cache = sync.Map{} // Clear resolution cache when binders change
	return fb
}

// Bind maps form or multipart form values from an [http.Request] to a destination struct pointer.
// The destination must be a pointer to a struct; otherwise, an error is returned.
// Fields in the destination struct must be annotated with `form` tags for binding.
// Required fields are validated based on the `required` attribute in the tag.
// Any parse or binding error encountered during this process is returned.
func (fb *FormBinder) Bind(r *http.Request, dst any) error {
	if r.Form == nil {
		const maxFormSize = 32 << 20 // 32 MB
		if err := r.ParseMultipartForm(maxFormSize); err != nil {
			return err
		}
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return errors.New("dst must be a pointer to a struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	meta, ok := metadataCache.Load(rt)
	if !ok {
		meta = parseStructMetadata(rt)
		metadataCache.Store(rt, meta)
	}

	sMeta := meta.(*structMetadata)

	for _, f := range sMeta.fields {
		values, ok := r.Form[f.name]
		if f.required {
			hasValue := false
			if ok {
				for _, v := range values {
					if v != "" {
						hasValue = true
						break
					}
				}
			}
			if !hasValue {
				return fmt.Errorf("missing required field %q", f.name)
			}
		}
		if f.trim {
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
		}

		if !ok || len(values) == 0 {
			continue
		}

		fieldVal := rv.Field(f.index)
		if !fieldVal.CanSet() {
			continue
		}

		if err := fb.bindField(fieldVal, values); err != nil {
			if errors.Is(err, ErrValueIsZero) {
				continue
			}
			return err
		}
	}

	return nil
}

// bindField resolves a Binder for the given field and delegates the binding of values to it. Returns an error on failure.
func (fb *FormBinder) bindField(field reflect.Value, values []string) error {
	ft := field.Type()
	if b, ok := fb.cache.Load(ft); ok {
		return fb.executeBind(b.(Binder), field, values)
	}

	addr := field.Addr().Interface()
	for _, b := range fb.binders {
		if b.Mappable(addr) {
			fb.cache.Store(ft, b)
			return fb.executeBind(b, field, values)
		}
	}
	return nil
}

// executeBind binds values to a field using the provided Binder. Handles singular and slice type bindings. Returns an error.
func (fb *FormBinder) executeBind(b Binder, field reflect.Value, values []string) error {
	addr := field.Addr().Interface()
	if field.Type().Kind() == reflect.Slice {
		return b.BindMany(values, addr)
	}
	return b.Bind(values[len(values)-1], addr)
}

// parseStructMetadata parses the struct type information to extract metadata for fields tagged with `form` tags.
// Returns a structMetadata instance containing details about fields, including their indexes, names, and requirements.
func parseStructMetadata(rt reflect.Type) *structMetadata {
	meta := new(structMetadata)
	for i := range rt.NumField() {
		field := rt.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		name := strings.TrimSpace(parts[0])
		required := false
		trim := false
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			if part == "required" {
				required = true
			} else if part == "trim" {
				trim = true
			}
		}
		meta.fields = append(meta.fields, fieldInfo{
			index:    i,
			name:     name,
			required: required,
			trim:     trim,
		})
	}
	return meta
}
