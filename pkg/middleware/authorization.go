package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pudottapommin/golib/pkg/auth"
)

type (
	AuthorizationOptFn  OptionsFn[AuthorizationConfig]
	AuthorizationConfig struct {
		Key            string
		LoginPath      string
		BeforeFunc     func(echo.Context)
		SuccessHandler func(echo.Context)
		// ErrorHandler   func(echo.Context, error) error

		UnauthorizedHandler func(echo.Context) error
		AuthorizeHandler    func(echo.Context, auth.Identity) bool
	}
)

const LoginPath = "/login"

func NewAuthorizationConfig(opts ...AuthorizationOptFn) *AuthorizationConfig {
	cfg := &AuthorizationConfig{
		Key:       string(AuthenticationKey),
		LoginPath: LoginPath,
		// @todo: Check what conditions would I like to do compare by default
		AuthorizeHandler: func(c echo.Context, i auth.Identity) bool { return true },
		UnauthorizedHandler: func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, LoginPath)
		},
	}
	cfg.applyOptions(opts...)
	return cfg
}

func AuthorizationWithOptions(opts ...AuthorizationOptFn) echo.MiddlewareFunc {
	cfg := NewAuthorizationConfig(opts...)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.BeforeFunc != nil {
				cfg.BeforeFunc(c)
			}

			if identity, ok := c.Get(cfg.Key).(auth.Identity); ok && identity != nil && cfg.AuthorizeHandler(c, identity) {
				if cfg.SuccessHandler != nil {
					cfg.SuccessHandler(c)
				}
				return next(c)
			}

			return cfg.UnauthorizedHandler(c)
		}
	}
}

func (c *AuthorizationConfig) applyOptions(opts ...AuthorizationOptFn) {
	for _, opt := range opts {
		opt(c)
	}
}

func WithAuthorizationBefore(fn func(c echo.Context)) AuthorizationOptFn {
	return func(cfg *AuthorizationConfig) {
		cfg.BeforeFunc = fn
	}
}

func WithAuthorizationKey(key string) AuthorizationOptFn {
	return func(cfg *AuthorizationConfig) {
		cfg.Key = key
	}
}

// func WithErrorHandler(fn func(c echo.Context, err error) error) AuthorizationOptFn {
// 	return func(cfg *AuthorizationConfig) {
// 		cfg.ErrorHandler = fn
// 	}
// }

func WithAuthorizationHandler(fn func(echo.Context, auth.Identity) bool) AuthorizationOptFn {
	return func(cfg *AuthorizationConfig) {
		cfg.AuthorizeHandler = fn
	}
}

func WithUnauthorizedHandler(fn func(echo.Context) error) AuthorizationOptFn {
	return func(cfg *AuthorizationConfig) {
		cfg.UnauthorizedHandler = fn
	}
}
