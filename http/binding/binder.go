package binding

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

type (
	// Binder is the generic interface for binding string and slice inputs to type T.
	Binder[T any] interface {
		Bind(s string, dst *T) error
		BindMany(s []string, dst *[]T) error
	}

	// GenericBinder is an alias for Binder[T].
	GenericBinder[T any] = Binder[T]

	// FieldBinder is the reflection interface used by FormBinder for dynamic dispatch.
	FieldBinder interface {
		Mappable(any) bool
		MappableType(reflect.Type) bool
		Bind(src string, dst any) error
		BindMany(src []string, dst any) error
	}

	// BindUnmarshaler is implemented by types that can unmarshal themselves from a form string.
	BindUnmarshaler interface {
		UnmarshalBind(string) error
	}
)

var (
	ErrDestinationTypeInvalid = errors.New("binder: invalid destination type for binder")
	ErrDestinationNil         = errors.New("binder: destination cannot be nil")
	ErrValueIsZero            = errors.New("binder: value is zero")
	ErrEmptySource            = errors.New("binder: empty source")
)

var unmarshalerType = reflect.TypeFor[BindUnmarshaler]()

func implementsUnmarshaler(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if t.Implements(unmarshalerType) {
		return true
	}
	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(unmarshalerType) {
		return true
	}
	return false
}

type binderAdapter[T any] struct {
	binder    Binder[T]
	ptrType   reflect.Type
	sliceType reflect.Type
	valType   reflect.Type
}

func (a binderAdapter[T]) MappableType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if t == a.ptrType || t == a.sliceType || t == a.valType || (t.Kind() == reflect.Slice && t.Elem() == a.valType) {
		return true
	}

	target := t
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}

	if target.Kind() == reflect.Slice {
		elem := target.Elem()
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if implementsUnmarshaler(elem) {
			return false
		}
		return elem.Kind() == a.valType.Kind()
	}

	if implementsUnmarshaler(target) {
		return false
	}
	return target.Kind() == a.valType.Kind()
}

func (a binderAdapter[T]) Mappable(dst any) bool {
	if dst == nil {
		return false
	}
	dt := reflect.TypeOf(dst)
	if dt == a.ptrType || dt == a.sliceType {
		return true
	}
	if dt.Kind() != reflect.Pointer {
		return false
	}

	elem := dt.Elem()
	if elem.Kind() == reflect.Slice {
		sliceElem := elem.Elem()
		if sliceElem.Kind() == reflect.Pointer {
			sliceElem = sliceElem.Elem()
		}
		if implementsUnmarshaler(sliceElem) {
			return false
		}
		return sliceElem.Kind() == a.valType.Kind()
	}

	if elem.Kind() == reflect.Pointer {
		ptrElem := elem.Elem()
		if implementsUnmarshaler(ptrElem) {
			return false
		}
		return ptrElem.Kind() == a.valType.Kind()
	}

	if implementsUnmarshaler(elem) {
		return false
	}
	return elem.Kind() == a.valType.Kind()
}

// Mappable returns true if the binder can map into destination type D.
func Mappable[D any](b FieldBinder, _ D) bool {
	return b.MappableType(reflect.TypeFor[D]())
}

func (a binderAdapter[T]) Bind(src string, dst any) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if v, ok := dst.(*T); ok {
		return a.binder.Bind(src, v)
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrDestinationTypeInvalid
	}

	targetVal := rv.Elem()
	if targetVal.Kind() == reflect.Pointer {
		if targetVal.IsNil() {
			targetVal.Set(reflect.New(targetVal.Type().Elem()))
		}
		targetVal = targetVal.Elem()
	}

	var tmp T
	if err := a.binder.Bind(src, &tmp); err != nil {
		return err
	}

	tmpVal := reflect.ValueOf(tmp)
	if !tmpVal.Type().ConvertibleTo(targetVal.Type()) {
		return ErrDestinationTypeInvalid
	}

	targetVal.Set(tmpVal.Convert(targetVal.Type()))
	return nil
}

func (a binderAdapter[T]) BindMany(src []string, dst any) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if v, ok := dst.(*[]T); ok {
		return a.binder.BindMany(src, v)
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Slice {
		return ErrDestinationTypeInvalid
	}

	sliceVal := rv.Elem()
	sliceType := sliceVal.Type()
	elemType := sliceType.Elem()

	var tmp []T
	if cap := sliceVal.Cap(); cap >= len(src) {
		tmp = make([]T, 0, cap)
	}
	if err := a.binder.BindMany(src, &tmp); err != nil {
		return err
	}

	targetElemType := elemType
	isPtrElem := elemType.Kind() == reflect.Pointer
	if isPtrElem {
		targetElemType = elemType.Elem()
	}

	tmpType := reflect.TypeFor[T]()
	if !tmpType.ConvertibleTo(targetElemType) {
		return ErrDestinationTypeInvalid
	}

	if sliceVal.Cap() >= len(tmp) {
		sliceVal.SetLen(len(tmp))
		for i, v := range tmp {
			val := reflect.ValueOf(v).Convert(targetElemType)
			if isPtrElem {
				elem := sliceVal.Index(i)
				if elem.IsNil() {
					elem.Set(reflect.New(targetElemType))
				}
				elem.Elem().Set(val)
			} else {
				sliceVal.Index(i).Set(val)
			}
		}
	} else {
		newSlice := reflect.MakeSlice(sliceType, len(tmp), len(tmp))
		for i, v := range tmp {
			val := reflect.ValueOf(v).Convert(targetElemType)
			if isPtrElem {
				ptr := reflect.New(targetElemType)
				ptr.Elem().Set(val)
				newSlice.Index(i).Set(ptr)
			} else {
				newSlice.Index(i).Set(val)
			}
		}
		sliceVal.Set(newSlice)
	}

	return nil
}

// WrapBinder converts a typed Binder[T] into a FieldBinder for use with FormBinder.
func WrapBinder[T any](b Binder[T]) FieldBinder {
	return binderAdapter[T]{
		binder:    b,
		ptrType:   reflect.TypeFor[*T](),
		sliceType: reflect.TypeFor[*[]T](),
		valType:   reflect.TypeFor[T](),
	}
}

var (
	bindersMu      sync.RWMutex
	defaultBinders = []FieldBinder{
		WrapBinder(StringBinder{}),
		WrapBinder(BoolBinder{}),
		WrapBinder(Float32Binder{}),
		WrapBinder(Float64Binder{}),
		WrapBinder(IntBinder{}),
		WrapBinder(Int8Binder{}),
		WrapBinder(Int16Binder{}),
		WrapBinder(Int32Binder{}),
		WrapBinder(Int64Binder{}),
		WrapBinder(UintBinder{}),
		WrapBinder(Uint8Binder{}),
		WrapBinder(Uint16Binder{}),
		WrapBinder(Uint32Binder{}),
		WrapBinder(Uint64Binder{}),
		AnyBinder{},
	}
)

func GetDefaultBinders() []FieldBinder {
	bindersMu.RLock()
	defer bindersMu.RUnlock()
	res := make([]FieldBinder, len(defaultBinders))
	copy(res, defaultBinders)
	return res
}

func AddDefaultBinder(b FieldBinder) {
	bindersMu.Lock()
	defer bindersMu.Unlock()
	defaultBinders = append([]FieldBinder{b}, defaultBinders...)
}

func AddDefaultGenericBinder[T any](b Binder[T]) {
	AddDefaultBinder(WrapBinder(b))
}

type boundField struct {
	index    int
	name     string
	required bool
	trim     bool
	isSlice  bool
	binder   FieldBinder
}

type structMetadata struct {
	fields []boundField
}

func compileStructMetadata(rt reflect.Type, binders []FieldBinder) *structMetadata {
	meta := new(structMetadata)
	for i := range rt.NumField() {
		field := rt.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}
		info := parseFieldTag(i, tag)
		ft := field.Type
		isSlice := ft.Kind() == reflect.Slice

		var matched FieldBinder
		for _, b := range binders {
			if b.MappableType(ft) {
				matched = b
				break
			}
		}

		meta.fields = append(meta.fields, boundField{
			index:    info.index,
			name:     info.name,
			required: info.required,
			trim:     info.trim,
			isSlice:  isSlice,
			binder:   matched,
		})
	}
	return meta
}

type parseStrategy int

const (
	parseStrategyMultipart parseStrategy = iota
	parseStrategyFormOnly
	parseStrategySkip
)

const defaultMaxMemory = 32 << 20 // 32 MB

type (
	FormBinderOption[T any] func(binder *FormBinder[T])
	FormBinder[T any]       struct {
		skipDefaults  bool
		parseStrategy parseStrategy
		maxMemory     int64
		binders       []FieldBinder
		meta          *structMetadata
	}
)

func FormWithBinders[T any](binders ...FieldBinder) FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.AddBinder(binders...)
	}
}

func FormWithGenericBinder[T any, F any](binder Binder[F]) FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.AddBinder(WrapBinder(binder))
	}
}

func FormWithMaxMemory[T any](maxMemory int64) FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.maxMemory = maxMemory
	}
}

func FormWithParseForm[T any]() FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.parseStrategy = parseStrategyFormOnly
	}
}

func FormWithParseMultipart[T any](maxMemory int64) FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.parseStrategy = parseStrategyMultipart
		b.maxMemory = maxMemory
	}
}

func FormWithSkipParse[T any]() FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.parseStrategy = parseStrategySkip
	}
}

func FormSkipDefaultBinders[T any](skipDefaults bool) FormBinderOption[T] {
	return func(b *FormBinder[T]) {
		b.skipDefaults = skipDefaults
	}
}

func NewFormBinder[T any](opts ...FormBinderOption[T]) *FormBinder[T] {
	fb := &FormBinder[T]{
		binders:       nil,
		parseStrategy: parseStrategyMultipart,
		maxMemory:     defaultMaxMemory,
	}
	for _, opt := range opts {
		opt(fb)
	}
	if !fb.skipDefaults {
		fb.binders = append(fb.binders, GetDefaultBinders()...)
	}
	fb.recompile()
	return fb
}

func (fb *FormBinder[T]) recompile() {
	rt := reflect.TypeFor[T]()
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() == reflect.Struct {
		fb.meta = compileStructMetadata(rt, fb.binders)
	}
}

func (fb *FormBinder[T]) AddBinder(b ...FieldBinder) *FormBinder[T] {
	fb.binders = append(b, fb.binders...)
	fb.recompile()
	return fb
}

func (fb *FormBinder[T]) parseRequest(r *http.Request) error {
	if r.Form != nil || fb.parseStrategy == parseStrategySkip {
		return nil
	}
	if fb.parseStrategy == parseStrategyFormOnly {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("binder: failed to parse form: %w", err)
		}
		return nil
	}
	maxMem := fb.maxMemory
	if maxMem <= 0 {
		maxMem = defaultMaxMemory
	}
	if err := r.ParseMultipartForm(maxMem); err != nil {
		return fmt.Errorf("binder: failed to parse multipart form: %w", err)
	}
	return nil
}

func hasNonEmptyValue(values []string) bool {
	for _, v := range values {
		if v != "" {
			return true
		}
	}
	return false
}

func validateAndExtractValues(form url.Values, name string, required, trim bool) ([]string, error) {
	values, ok := form[name]
	if required && !hasNonEmptyValue(values) {
		return nil, fmt.Errorf("missing required field %q", name)
	}
	if trim {
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
	}
	if !ok || len(values) == 0 {
		return nil, nil
	}
	return values, nil
}

// Bind extracts form data and returns an instantiated and populated value of type T.
func (fb *FormBinder[T]) Bind(r *http.Request) (T, error) {
	var dst T
	if err := fb.BindTo(r, &dst); err != nil {
		var zero T
		return zero, err
	}
	return dst, nil
}

// BindTo populates dst with values from r.Form using pre-compiled metadata for T.
func (fb *FormBinder[T]) BindTo(r *http.Request, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if fb.meta == nil {
		return errors.New("dst must be a pointer to a struct")
	}
	if err := fb.parseRequest(r); err != nil {
		return err
	}
	return fb.bindStructWithMeta(reflect.ValueOf(dst).Elem(), r.Form)
}

func (fb *FormBinder[T]) bindStructWithMeta(elem reflect.Value, form url.Values) error {
	for i := range fb.meta.fields {
		f := &fb.meta.fields[i]
		values, err := validateAndExtractValues(form, f.name, f.required, f.trim)
		if err != nil {
			return err
		}
		if len(values) == 0 || f.binder == nil {
			continue
		}

		fieldVal := elem.Field(f.index)
		if !fieldVal.CanSet() {
			continue
		}

		if f.isSlice {
			addr := fieldVal.Addr().Interface()
			err = f.binder.BindMany(values, addr)
		} else {
			lastVal := values[len(values)-1]
			if fieldVal.Kind() == reflect.Pointer {
				if lastVal == "" {
					fieldVal.Set(reflect.Zero(fieldVal.Type()))
					continue
				}
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
				err = f.binder.Bind(lastVal, fieldVal.Interface())
				if err != nil {
					if errors.Is(err, ErrValueIsZero) {
						fieldVal.Set(reflect.Zero(fieldVal.Type()))
						err = nil
					} else {
						fieldVal.Set(reflect.Zero(fieldVal.Type()))
					}
				}
			} else {
				addr := fieldVal.Addr().Interface()
				err = f.binder.Bind(lastVal, addr)
			}
		}
		if err != nil && !errors.Is(err, ErrValueIsZero) {
			return err
		}
	}
	return nil
}

func Bind[T any](r *http.Request, opts ...FormBinderOption[T]) (T, error) {
	return NewFormBinder[T](opts...).Bind(r)
}

func BindTo[T any](r *http.Request, dst *T, opts ...FormBinderOption[T]) error {
	return NewFormBinder[T](opts...).BindTo(r, dst)
}

func For[T any](opts ...FormBinderOption[T]) *FormBinder[T] {
	return NewFormBinder[T](opts...)
}

func parseRequestForm(r *http.Request) error {
	if r.Form != nil {
		return nil
	}
	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		if err = r.ParseForm(); err != nil {
			return fmt.Errorf("binder: failed to parse form: %w", err)
		}
	}
	return nil
}

func Value[T any](r *http.Request, name string) (T, error) {
	if err := parseRequestForm(r); err != nil {
		var zero T
		return zero, err
	}
	values, ok := r.Form[name]
	if !ok || len(values) == 0 {
		var zero T
		return zero, fmt.Errorf("missing field %q", name)
	}
	return Parse[T](values[len(values)-1])
}

func Values[T any](r *http.Request, name string) ([]T, error) {
	if err := parseRequestForm(r); err != nil {
		return nil, err
	}
	values, ok := r.Form[name]
	if !ok || len(values) == 0 {
		return nil, nil
	}
	return ParseSlice[T](values)
}

func ValueOrDefault[T any](r *http.Request, name string, defaultVal T) T {
	val, err := Value[T](r, name)
	if err != nil {
		return defaultVal
	}
	return val
}

func Parse[T any](src string, binders ...FieldBinder) (T, error) {
	var dst T
	bList := binders
	if len(bList) == 0 {
		bList = GetDefaultBinders()
	}
	ptr := &dst
	for _, b := range bList {
		if b.Mappable(ptr) {
			if err := b.Bind(src, ptr); err != nil {
				var zero T
				return zero, err
			}
			return dst, nil
		}
	}
	var zero T
	return zero, ErrDestinationTypeInvalid
}

func ParseSlice[T any](src []string, binders ...FieldBinder) ([]T, error) {
	var dst []T
	bList := binders
	if len(bList) == 0 {
		bList = GetDefaultBinders()
	}
	ptr := &dst
	for _, b := range bList {
		if b.Mappable(ptr) {
			if err := b.BindMany(src, ptr); err != nil {
				return nil, err
			}
			return dst, nil
		}
	}
	return nil, ErrDestinationTypeInvalid
}

type fieldInfo struct {
	index    int
	name     string
	required bool
	trim     bool
}

func parseFieldTag(index int, tag string) fieldInfo {
	parts := strings.Split(tag, ",")
	info := fieldInfo{
		index: index,
		name:  strings.TrimSpace(parts[0]),
	}
	for _, part := range parts[1:] {
		switch strings.TrimSpace(part) {
		case "required":
			info.required = true
		case "trim":
			info.trim = true
		}
	}
	return info
}
