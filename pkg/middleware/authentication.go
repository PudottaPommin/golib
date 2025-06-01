package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pudottapommin/golib/pkg/auth"
)

type (
	AuthenticationConfigFn OptionsFn[AuthenticationConfig]
	AuthenticationConfig   struct {
		Key             string
		LoginRouteName  string
		Config          auth.Config
		Skipper         middleware.Skipper
		BeforeFunc      func(echo.Context)
		AfterFunc       func(echo.Context)
		IdentityFactory func(echo.Context, *auth.CookieValue) (auth.Identity, error)
	}
)

func NewAuthenticationConfig(cfg auth.Config) AuthenticationConfig {
	return AuthenticationConfig{
		Key:    string(AuthenticationKey),
		Config: cfg,
		IdentityFactory: func(c echo.Context, cv *auth.CookieValue) (auth.Identity, error) {
			return auth.NewIdentity(cv)
		},
	}
}

func Authentication(cfg auth.Config) echo.MiddlewareFunc {
	return AuthenticationWithConfig(NewAuthenticationConfig(cfg))
}

func AuthenticationWithOptions(cfg auth.Config, opts ...AuthenticationConfigFn) echo.MiddlewareFunc {
	return AuthenticationWithConfig(NewAuthenticationConfig(cfg).applyOptions(opts...))
}

func AuthenticationWithConfig(cfg AuthenticationConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil {
				if cfg.Skipper(c) {
					return next(c)
				}
			}

			if cfg.BeforeFunc != nil {
				cfg.BeforeFunc(c)
			}

			cv, err := auth.GetCookie(c, &cfg.Config)

			if errors.Is(err, auth.ErrorAuthCookieMissing) {
				return c.Redirect(http.StatusTemporaryRedirect, c.Echo().Reverse(cfg.LoginRouteName))
			}

			if err == nil {
				identity, err := cfg.IdentityFactory(c, cv)
				if err == nil {
					c.Set(cfg.Key, identity)
				}
			}

			if err != nil {
				c.Error(err)
			}

			if cfg.AfterFunc != nil {
				cfg.AfterFunc(c)
			}

			return next(c)
		}
	}
}

func (c AuthenticationConfig) applyOptions(opts ...AuthenticationConfigFn) AuthenticationConfig {
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func WithAuthenticationIdentityFactory(factory func(echo.Context, *auth.CookieValue) (auth.Identity, error)) func(*AuthenticationConfig) {
	return func(cfg *AuthenticationConfig) {
		cfg.IdentityFactory = factory
	}
}

func WithAuthenticationKey(key string) func(*AuthenticationConfig) {
	return func(cfg *AuthenticationConfig) {
		cfg.Key = key
	}
}

func WithAuthenticationLoginRouteName(routeName string) func(*AuthenticationConfig) {
	return func(cfg *AuthenticationConfig) {
		cfg.LoginRouteName = routeName
	}
}

func WithAuthenticationSkipper(skipper middleware.Skipper) func(*AuthenticationConfig) {
	return func(cfg *AuthenticationConfig) {
		cfg.Skipper = skipper
	}
}
