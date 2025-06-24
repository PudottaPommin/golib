package binding

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type IBindable interface {
	Parse(string) error
}

// BindingError represents an error that occurred while binding request data.
type BindingError struct {
	// Field is the field name where value binding failed
	Field string `json:"field"`
	*HTTPError
	// Values of parameter that failed to bind.
	Values []string `json:"-"`
}

// NewBindingError creates a new instance of binding error
func NewBindingError(sourceParam string, values []string, message interface{}, internalError error) error {
	return &BindingError{
		Field:  sourceParam,
		Values: values,
		HTTPError: &HTTPError{
			Code:     http.StatusBadRequest,
			Message:  message,
			Internal: internalError,
		},
	}
}

// Error returns error message
func (be *BindingError) Error() string {
	return fmt.Sprintf("%s, field=%s", be.HTTPError.Error(), be.Field)
}

type ValueBinder struct {
	// ValueFunc is used to get a single parameter (first) value from a request
	ValueFunc func(sourceParam string) string
	// ValuesFunc is used to get all values for parameter from request. i.e. `/api/search?ids=1&ids=2`
	ValuesFunc func(sourceParam string) []string
	// ErrorFunc is used to create errors. Allows you to use your own error type, that for example marshals to your specific json response
	ErrorFunc func(sourceParam string, values []string, message interface{}, internalError error) error
	errors    []error
	// failFast is a flag for binding methods to return without attempting to bind when previous binding already failed
	failFast bool
}

// FormFieldBinder creates form field value binder
// For all requests, FormFieldBinder parses the raw query from the URL and uses query params as form fields
//
// For POST, PUT, and PATCH requests, it also reads the request body, parses it
// as a form and uses query params as form fields. Request body parameters take precedence over URL query
// string values in r.Form.
//
// NB: when binding forms take note that this implementation uses standard library form parsing
// which parses form data from BOTH URL and BODY if content type is not MIMEMultipartForm
// See https://golang.org/pkg/net/http/#Request.ParseForm
func FormFieldBinder(c *gin.Context) *ValueBinder {
	vb := &ValueBinder{
		failFast: true,
		ValueFunc: func(sourceParam string) string {
			return c.Request.FormValue(sourceParam)
		},
		ErrorFunc: NewBindingError,
	}
	vb.ValuesFunc = func(sourceParam string) []string {
		if c.Request.Form == nil {
			// this is same as `Request().FormValue()` does internally
			_ = c.Request.ParseMultipartForm(32 << 20)
		}
		values, ok := c.Request.Form[sourceParam]
		if !ok {
			return nil
		}
		return values
	}

	return vb
}

// QueryParamsBinder creates query parameter value binder
func QueryParamsBinder(c *gin.Context) *ValueBinder {
	return &ValueBinder{
		failFast:  true,
		ValueFunc: c.Query,
		ValuesFunc: func(sourceParam string) []string {
			values, ok := c.GetQueryArray(sourceParam)
			if !ok {
				return nil
			}
			return values
		},
		ErrorFunc: NewBindingError,
	}
}

// PathParamsBinder creates path parameter value binder
func PathParamsBinder(c *gin.Context) *ValueBinder {
	return &ValueBinder{
		failFast:  true,
		ValueFunc: c.Param,
		ValuesFunc: func(sourceParam string) []string {
			value := c.Param(sourceParam)
			if value == "" {
				return nil
			}
			return []string{value}
		},
		ErrorFunc: NewBindingError,
	}
}

// BindError returns first seen bind error and resets/empties binder errors for further calls
func (b *ValueBinder) BindError() error {
	if b.errors == nil {
		return nil
	}
	err := b.errors[0]
	b.errors = nil // reset errors so next chain will start from zero
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
			b.setError(b.ErrorFunc(key, []string{value}, "required field value is empty", nil))
		}
		return b
	}

	id, err := uuid.FromString(value)
	if err != nil {
		b.setError(b.ErrorFunc(key, []string{value}, "invalid uuid value", nil))
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
			b.setError(b.ErrorFunc(key, []string{value}, "required field value is empty", nil))
		}
		return b
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		b.setError(b.ErrorFunc(key, []string{value}, "failed to bind field to value Time", err))
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
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "required field value is empty", nil))
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
			b.setError(b.ErrorFunc(sourceParam, []string{}, "required field value is empty", nil))
		}
		return b
	}
	return b.bool(sourceParam, value, dest)
}

func (b *ValueBinder) bool(sourceParam string, value string, dest *bool) *ValueBinder {
	n, err := strconv.ParseBool(value)
	if err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "failed to bind field value to bool", err))
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
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "failed to bind field value to IBindable", err))
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
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "required field value is empty", nil))
		return b
	}
	if err := dest.Parse(value); err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "failed to bind field value to IBindable", err))
	}
	return b
}

func (b *ValueBinder) CustomFunc(sourceParam string, fn func(string) error) *ValueBinder {
	if b.failFast && b.errors != nil {
		return b
	}

	value := b.ValueFunc(sourceParam)
	if value == "" {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "required field value is empty", nil))
		return b
	}
	if err := fn(value); err != nil {
		b.setError(b.ErrorFunc(sourceParam, []string{value}, "failed to bind field value to IBindable", err))
	}
	return b
}
