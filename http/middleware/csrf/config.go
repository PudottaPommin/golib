package csrf

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pudottapommin/golib/pkg/id"
)

type (
	OptsFn func(*Config)
	Config struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(c *gin.Context) bool
		// Optional, Default: id.New
		Generator func() string
		// CookieName is the name of the CSRF cookie.
		// Optional, Default: "_csrf"
		CookieName string
		// CookiePath defines the path attribute of the CSRF cookie.
		// Optional, Default: "/"
		CookiePath string
		// CookieSameSite defines the SameSite attribute of the CSRF cookie.
		// Optional, Default: http.SameSiteStrictMode
		CookieSameSite http.SameSite
		// CookieExpiration sets the expiration duration for the CSRF cookie.
		// Optional, Default: 30 minutes
		CookieExpiration time.Duration
		// CookieHttpOnly defines if the HttpOnly flag should be set on the CSRF cookie.
		// Optional, Default: true
		CookieHttpOnly bool
		// CookieSecure defines if the Secure flag should be set on the CSRF cookie.
		// Optional, Default: true
		CookieSecure bool
	}
)

const ContextKey = "csrf"

var (
	defaultConfig = Config{
		Next:             nil,
		CookieName:       "_csrf",
		CookiePath:       "/",
		CookieSameSite:   http.SameSiteStrictMode,
		CookieExpiration: 30 * time.Minute,
		CookieHttpOnly:   true,
		CookieSecure:     true,
		Generator:        func() string { return id.New().String() },
	}
)

func WithNext(fn func(*gin.Context) bool) OptsFn {
	return func(c *Config) {
		c.Next = fn
	}
}

func WithGenerator(generator func() string) OptsFn {
	return func(c *Config) {
		c.Generator = generator
	}
}

func WithCookieName(value string) OptsFn {
	return func(c *Config) {
		c.CookieName = value
	}
}

func WithCookiePath(value string) OptsFn {
	return func(c *Config) {
		c.CookiePath = value
	}
}

func WithCookieSameSite(value http.SameSite) OptsFn {
	return func(c *Config) {
		c.CookieSameSite = value
	}
}

func WithCookieExpiration(value time.Duration) OptsFn {
	return func(c *Config) {
		c.CookieExpiration = value
	}
}

func WithCookieHttpOnly(value bool) OptsFn {
	return func(c *Config) {
		c.CookieHttpOnly = value
	}
}

func WithCookieSecure(value bool) OptsFn {
	return func(c *Config) {
		c.CookieSecure = value
	}
}
