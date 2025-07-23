package authentication

import (
	"github.com/gin-gonic/gin"
	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

type (
	OptsFn[T gAuth.Identity] func(*Config[T])
	Config[T gAuth.Identity] struct {
		// Optional, Default: "auth/context"
		ContextKey string
		// Optional, Default: nil
		NotAuthenticatedHandler func(*gin.Context)
		AuthConfig              *gAuth.Config
		Factory                 func(*gin.Context, *gAuth.CookieValue) (T, error)
		// Optional, Default: nil
		AfterHandler func(*gin.Context, T)
	}
)

const ContextKey = "auth/context"

func newDefaultConfig[T gAuth.Identity]() Config[T] {
	return Config[T]{
		ContextKey:              ContextKey,
		NotAuthenticatedHandler: nil,
		AuthConfig:              nil,
		Factory:                 nil,
		AfterHandler:            nil,
	}
}

func WithContextKey[T gAuth.Identity](key string) OptsFn[T] {
	return func(c *Config[T]) {
		c.ContextKey = key
	}
}

func WithAuthConfig[T gAuth.Identity](cfg *gAuth.Config) OptsFn[T] {
	return func(c *Config[T]) {
		c.AuthConfig = cfg
	}
}

func WithNotAuthenticatedHandler[T gAuth.Identity](handler func(*gin.Context)) OptsFn[T] {
	return func(c *Config[T]) {
		c.NotAuthenticatedHandler = handler
	}
}

func WithFactory[T gAuth.Identity](factory func(*gin.Context, *gAuth.CookieValue) (T, error)) OptsFn[T] {
	return func(c *Config[T]) {
		c.Factory = factory
	}
}

func WithAfterHandler[T gAuth.Identity](handler func(*gin.Context, T)) OptsFn[T] {
	return func(c *Config[T]) {
		c.AfterHandler = handler
	}
}
