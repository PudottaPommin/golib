package binding

import (
	"errors"
	"fmt"
)

var (
	ErrRequired        = errors.New("required field value is empty")
	ErrInvalidUUID     = errors.New("invalid uuid value")
	ErrInvalidTime     = errors.New("failed to bind field to value Time")
	ErrInvalidBool     = errors.New("failed to bind field value to bool")
	ErrInvalidBindable = errors.New("failed to bind field value to IBindable")
)

type HTTPError struct {
	Internal error `json:"-"` // Stores the error returned by an external dependency
	Code     int   `json:"-"`
	Message  any   `json:"message"`
}

// Error makes it compatible with the `error` interface.
func (he *HTTPError) Error() string {
	if he.Internal == nil {
		return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
	}
	return fmt.Sprintf("code=%d, message=%v, internal=%v", he.Code, he.Message, he.Internal)
}
