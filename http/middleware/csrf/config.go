package csrf

import (
	"net/http"
	"time"

	"github.com/pudottapommin/golib/pkg/id"
)

type (
	OptsFn func(*mw)
	mw     struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(http.ResponseWriter, *http.Request) bool
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

func New(opts ...OptsFn) *mw {
	m := &mw{
		Next:             nil,
		CookieName:       "_csrf",
		CookiePath:       "/",
		CookieSameSite:   http.SameSiteStrictMode,
		CookieExpiration: 30 * time.Minute,
		CookieHttpOnly:   true,
		CookieSecure:     true,
		Generator:        func() string { return id.New().String() },
	}
	for i := range opts {
		opts[i](m)
	}
	return m
}

func WithNext(fn func(http.ResponseWriter, *http.Request) bool) OptsFn {
	return func(c *mw) {
		c.Next = fn
	}
}

func WithGenerator(generator func() string) OptsFn {
	return func(c *mw) {
		c.Generator = generator
	}
}

func WithCookieName(value string) OptsFn {
	return func(c *mw) {
		c.CookieName = value
	}
}

func WithCookiePath(value string) OptsFn {
	return func(c *mw) {
		c.CookiePath = value
	}
}

func WithCookieSameSite(value http.SameSite) OptsFn {
	return func(c *mw) {
		c.CookieSameSite = value
	}
}

func WithCookieExpiration(value time.Duration) OptsFn {
	return func(c *mw) {
		c.CookieExpiration = value
	}
}

func WithCookieHttpOnly(value bool) OptsFn {
	return func(c *mw) {
		c.CookieHttpOnly = value
	}
}

func WithCookieSecure(value bool) OptsFn {
	return func(c *mw) {
		c.CookieSecure = value
	}
}
