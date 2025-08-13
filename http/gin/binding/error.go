package binding

import (
	"fmt"
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
