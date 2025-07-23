package authorization

import (
	"github.com/gin-gonic/gin"
	"github.com/pudottapommin/golib/http/middleware/authentication"
	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

type (
	AuthorizeHandlerFn[T gAuth.Identity]    func(*gin.Context, T) bool
	AuthorizedHandlerFn[T gAuth.Identity]   func(*gin.Context, T)
	UnauthorizedHandlerFn[T gAuth.Identity] func(*gin.Context, T)

	OptsFn[T gAuth.Identity] func(*Config[T])
	Config[T gAuth.Identity] struct {
		// Optional, Default: authentication.ContextKey
		ContextKey string
		// Default: if T not null, then TRUE
		AuthorizeHandler  AuthorizeHandlerFn[T]
		AuthorizedHandler AuthorizedHandlerFn[T]
		// Optional, Default: AbortWithStatus
		UnauthorizedHandler UnauthorizedHandlerFn[T]
	}
)

func newDefaultConfig[T gAuth.Identity]() Config[T] {
	return Config[T]{
		ContextKey: authentication.ContextKey,
		AuthorizeHandler: func(c *gin.Context, i T) bool {
			var x = &i
			return x != nil
		},
		AuthorizedHandler: nil,
		UnauthorizedHandler: func(c *gin.Context, _ T) {
			c.AbortWithStatus(401)
		},
	}
}

func WithContextKey[T gAuth.Identity](key string) OptsFn[T] {
	return func(c *Config[T]) {
		c.ContextKey = key
	}
}

func WithAuthorizeHandler[T gAuth.Identity](h AuthorizeHandlerFn[T]) OptsFn[T] {
	return func(c *Config[T]) {
		c.AuthorizeHandler = h
	}
}

func WithAuthorizedHandler[T gAuth.Identity](h AuthorizedHandlerFn[T]) OptsFn[T] {
	return func(c *Config[T]) {
		c.AuthorizedHandler = h
	}
}

func WithUnauthorizedHandler[T gAuth.Identity](h UnauthorizedHandlerFn[T]) OptsFn[T] {
	return func(c *Config[T]) {
		c.UnauthorizedHandler = h
	}
}
