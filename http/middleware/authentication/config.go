package authentication

import (
	"net/http"

	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

type (
	OptsFn[T gAuth.Identity] func(*mw[T])
	mw[T gAuth.Identity]     struct {
		// Optional, Default: "auth/context"
		ContextKey string
		// Optional, Default: nil
		NotAuthenticatedHandler func(http.ResponseWriter, *http.Request)
		AuthConfig              *gAuth.Config
		Factory                 func(http.ResponseWriter, *http.Request, *gAuth.CookieValue) (T, error)
		// Optional, Default: nil
		AfterHandler func(http.ResponseWriter, *http.Request, T)
	}
)

const ContextKey = "auth/context"

func New[T gAuth.Identity](opts ...OptsFn[T]) *mw[T] {
	m := &mw[T]{
		ContextKey:              ContextKey,
		NotAuthenticatedHandler: nil,
		AuthConfig:              nil,
		Factory:                 nil,
		AfterHandler:            nil,
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

func WithAuthConfig[T gAuth.Identity](cfg *gAuth.Config) OptsFn[T] {
	return func(c *mw[T]) {
		c.AuthConfig = cfg
	}
}

func WithNotAuthenticatedHandler[T gAuth.Identity](handler func(http.ResponseWriter, *http.Request)) OptsFn[T] {
	return func(c *mw[T]) {
		c.NotAuthenticatedHandler = handler
	}
}

func WithFactory[T gAuth.Identity](factory func(http.ResponseWriter, *http.Request, *gAuth.CookieValue) (T, error)) OptsFn[T] {
	return func(c *mw[T]) {
		c.Factory = factory
	}
}

func WithAfterHandler[T gAuth.Identity](handler func(http.ResponseWriter, *http.Request, T)) OptsFn[T] {
	return func(c *mw[T]) {
		c.AfterHandler = handler
	}
}
