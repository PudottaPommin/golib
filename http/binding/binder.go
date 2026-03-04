// Package binding with inspiration from GIN Framework
package binding

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

type IBindable interface {
	Parse(string) error
}

type Error struct {
	*HTTPError
	Field  string   `json:"field"`
	Values []string `json:"-"`
}

// NewBindingError creates a new instance of binding error
func NewBindingError(sourceParam string, values []string, message any, internalError error) error {
	return &Error{
		Field:  sourceParam,
		Values: values,
		HTTPError: &HTTPError{
			Code:     http.StatusBadRequest,
			Message:  message,
			Internal: internalError,
		},
	}
}

type ValueBinder struct {
	// ValueFunc is used to get a single parameter (first) value from a request
	ValueFunc func(sourceParam string) string
	// ValuesFunc is used to get all values for parameter from request. i.e. `/api/search?ids=1&ids=2`
	ValuesFunc func(sourceParam string) []string
	// ErrorFunc is used to create errors. Allows you to use your own error type that, for example, marshals to your specific json response
	ErrorFunc func(sourceParam string, values []string, message any, internalError error) error
	errors    []error
	// failFast is a flag for binding methods to return without attempting to bind when the previous binding already failed
	failFast bool
}

func FormFieldBinder(r *http.Request) *ValueBinder {
	return &ValueBinder{
		failFast:  true,
		ErrorFunc: NewBindingError,
		ValueFunc: r.FormValue,
		ValuesFunc: func(sourceParam string) []string {
			if r.Form == nil {
				_ = r.ParseMultipartForm(32 << 20)
			}
			values, ok := r.Form[sourceParam]
			if !ok {
				return nil
			}
			return values
		},
	}
}

// BindError returns first seen bind error and resets/empties binder errors for further calls
func (b *ValueBinder) BindError() error {
	if b.errors == nil {
		return nil
	}
	err := b.errors[0]
	b.errors = nil // reset errors so the next chain will start from zero
	return err
}

func (b *ValueBinder) setError(err error) {
	if b.errors == nil {
		b.errors = []error{err}
		return
	}
	b.errors = append(b.errors, err)
}

func (b *ValueBinder) ShouldUUID(key string, dest *uuid.UUID) *ValueBinder {
	return b.uuid(key, dest, false)
}

func (b *ValueBinder) UUID(key string, dest *uuid.UUID) *ValueBinder {
	return b.uuid(key, dest, true)
}

func (b *ValueBinder) uuid(key string, dest *uuid.UUID, valueMustExist bool) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(key)
	if value == "" {
		if valueMustExist {
			b.setError(b.ErrorFunc(key, []string{value}, ErrRequired.Error(), nil))
		}
		return b
	}

	id, err := uuid.FromString(value)
	if err != nil {
		b.setError(b.ErrorFunc(key, []string{value}, ErrInvalidUUID.Error(), nil))
		return b
	}
	*dest = id
	return b
}

func (b *ValueBinder) ShouldTime(key string, dest *time.Time, layout string) *ValueBinder {
	return b.time(key, dest, layout, false)
}

func (b *ValueBinder) Time(key string, dest *time.Time, layout string) *ValueBinder {
	return b.time(key, dest, layout, true)
}

func (b *ValueBinder) ShouldDate(key string, dest *time.Time) *ValueBinder {
	return b.time(key, dest, time.DateOnly, false)
}

func (b *ValueBinder) Date(key string, dest *time.Time) *ValueBinder {
	return b.time(key, dest, time.DateOnly, true)
}

func (b *ValueBinder) ShouldTimeOnly(key string, dest *time.Time) *ValueBinder {
	return b.time(key, dest, time.TimeOnly, false)
}

func (b *ValueBinder) TimeOnly(key string, dest *time.Time) *ValueBinder {
	return b.time(key, dest, time.TimeOnly, true)
}

func (b *ValueBinder) time(key string, dest *time.Time, layout string, valueMustExist bool) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(key)
	if value == "" {
		if valueMustExist {
			b.setError(b.ErrorFunc(key, []string{value}, ErrRequired.Error(), nil))
		}
		return b
	}
	if layout == time.TimeOnly && strings.Count(value, ":") == 1 {
		value += ":00"
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		b.setError(b.ErrorFunc(key, []string{value}, ErrInvalidTime.Error(), err))
		return b
	}
	*dest = t
	return b
}

// ShouldString binds parameter to string variable
func (b *ValueBinder) ShouldString(sourceParam string, dest *string) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		return b
	}
	*dest = value
	return b
}

// String requires parameter value to exist to bind to string variable. Returns error when value does not exist
func (b *ValueBinder) String(sourceParam string, dest *string) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrRequired.Error(), nil))
		return b
	}
	*dest = value
	return b
}

// ShouldBool binds parameter to bool variable
func (b *ValueBinder) ShouldBool(sourceParam string, dest *bool) *ValueBinder {
	return b.boolValue(sourceParam, dest, false)
}

// Bool requires parameter value to exist to bind to bool variable. Returns error when value does not exist
func (b *ValueBinder) Bool(sourceParam string, dest *bool) *ValueBinder {
	return b.boolValue(sourceParam, dest, true)
}

func (b *ValueBinder) boolValue(sourceParam string, dest *bool, valueMustExist bool) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		if valueMustExist {
			b.setError(b.ErrorFunc(sourceParam, []string{}, ErrRequired.Error(), nil))
		}
		return b
	}
	return b.bool(sourceParam, value, dest)
}

func (b *ValueBinder) bool(sourceParam string, value string, dest *bool) *ValueBinder {
	n, err := strconv.ParseBool(value)
	if err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrInvalidBool.Error(), err))
		return b
	}

	*dest = n
	return b
}

// ShouldCustom binds parameter to IBindable variable
func (b *ValueBinder) ShouldCustom(sourceParam string, dest IBindable) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		return b
	}
	if err := dest.Parse(value); err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrInvalidBindable.Error(), err))
	}
	return b
}

// Custom requires parameter value to exist to bind to IBindable variable. Returns error when value does not exist
func (b *ValueBinder) Custom(sourceParam string, dest IBindable) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrRequired.Error(), nil))
		return b
	}
	if err := dest.Parse(value); err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrInvalidBindable.Error(), err))
	}
	return b
}

func (b *ValueBinder) CustomFunc(sourceParam string, fn func(string) error) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrRequired.Error(), nil))
		return b
	}
	if err := fn(value); err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, ErrInvalidBindable.Error(), err))
	}
	return b
}
