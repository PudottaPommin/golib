package authorization

import (
	"net/http"

	"github.com/pudottapommin/golib"
	"github.com/pudottapommin/golib/http/middleware/authentication"
	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

type (
	AuthorizeHandlerFn[T gAuth.Identity]    func(http.ResponseWriter, *http.Request, *T) bool
	AuthorizedHandlerFn[T gAuth.Identity]   func(http.ResponseWriter, *http.Request, *T)
	UnauthorizedHandlerFn[T gAuth.Identity] func(http.ResponseWriter, *http.Request, *T)

	OptsFn[T gAuth.Identity] func(*mw[T])
	mw[T gAuth.Identity]     struct {
		// Optional, Default: authentication.ContextKey
		ContextKey string
		// Default: if T not null, then TRUE
		AuthorizeHandler  AuthorizeHandlerFn[T]
		AuthorizedHandler AuthorizedHandlerFn[T]
		// Optional, Default: AbortWithStatus
		UnauthorizedHandler UnauthorizedHandlerFn[T]
	}
)

func New[T gAuth.Identity](opts ...OptsFn[T]) *mw[T] {
	m := &mw[T]{
		ContextKey: authentication.ContextKey,
		AuthorizeHandler: func(w http.ResponseWriter, r *http.Request, t *T) bool {
			return golib.ToPointer(t) != nil
		},
		AuthorizedHandler: nil,
		UnauthorizedHandler: func(w http.ResponseWriter, r *http.Request, _ *T) {
			w.WriteHeader(http.StatusUnauthorized)
		},
	}
	for i := range opts {
		opts[i](m)
	}
	return m
}

func WithContextKey[T gAuth.Identity](key string) OptsFn[T] {
	return func(c *mw[T]) {
		c.ContextKey = key
	}
}

func WithAuthorizeHandler[T gAuth.Identity](h AuthorizeHandlerFn[T]) OptsFn[T] {
	return func(c *mw[T]) {
		c.AuthorizeHandler = h
	}
}

func WithAuthorizedHandler[T gAuth.Identity](h AuthorizedHandlerFn[T]) OptsFn[T] {
	return func(c *mw[T]) {
		c.AuthorizedHandler = h
	}
}

func WithUnauthorizedHandler[T gAuth.Identity](h UnauthorizedHandlerFn[T]) OptsFn[T] {
	return func(c *mw[T]) {
		c.UnauthorizedHandler = h
	}
}
