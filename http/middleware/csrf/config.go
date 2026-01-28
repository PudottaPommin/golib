package csrf

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	ghttp "github.com/pudottapommin/golib/http"
	"github.com/pudottapommin/golib/http/cookie"
)

type (
	key        string
	OptsFn     func(*Middleware)
	Middleware struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(http.ResponseWriter, *http.Request) bool
		// Optional, Default: cookie.GenerateRandomKey
		Generator func() []byte
		// Store is the csrf storage backend
		Store Store

		// HeaderName specifies the HTTP header used to fetch the CSRF token from incoming requests.
		// Optional, Default: "X-CSRF-Token"
		HeaderName string
		// FormFieldName specifies the form field name used to fetch the CSRF token from form submissions.
		// Optional, Default: ""
		FormFieldName string
		// CookieName is the name of the CSRF cookie.
		// Optional, Default: "_csrf"
		CookieName string
		// CookiePath defines the path attribute of the CSRF cookie.
		// Optional, Default: "/"
		CookiePath string
		// CookieDomain defines the domain attribute of the CSRF cookie.
		// Optional, Default: ""
		CookieDomain string
		// CookieSameSite defines the SameSite attribute of the CSRF cookie.
		// Optional, Default: http.SameSiteStrictMode
		CookieSameSite http.SameSite
		// MaxAge sets the expiration duration for the CSRF cookie.
		// Optional, Default: 0 session-only
		MaxAge time.Duration
		// CookieHttpOnly defines if the HttpOnly flag should be set on the CSRF cookie.
		// Optional, Default: true
		CookieHttpOnly bool
		// CookieSecure defines if the Secure flag should be set on the CSRF cookie.
		// Optional, Default: true
		CookieSecure bool
	}
)

const (
	ContextKey          key = "golib.csrf"
	FormFieldContextKey     = "golib.csrf.form"

	FormFieldName = "_csrf"
	CookieName    = "_csrf"
	HeaderName    = ghttp.HeaderXCSRFToken

	defaultTokenLength = 32
)

func New(sc *cookie.Cookie, opts ...OptsFn) *Middleware {
	m := &Middleware{
		CookieName:     CookieName,
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSameSite: http.SameSiteStrictMode,
		MaxAge:         0,
		CookieHttpOnly: true,
		CookieSecure:   true,
		FormFieldName:  FormFieldName,
		HeaderName:     HeaderName,
		Generator:      func() []byte { return cookie.GenerateRandomKey(defaultTokenLength) },
	}
	for i := range opts {
		opts[i](m)
	}
	if m.Store == nil {
		m.Store = &cookieStore{
			sc:       sc,
			name:     m.CookieName,
			secure:   m.CookieSecure,
			httpOnly: m.CookieHttpOnly,
			sameSite: m.CookieSameSite,
			maxAge:   m.MaxAge,
			path:     m.CookiePath,
			domain:   m.CookieDomain,
		}
	}
	return m
}

func WithNext(fn func(http.ResponseWriter, *http.Request) bool) OptsFn {
	return func(c *Middleware) {
		c.Next = fn
	}
}

func WithGenerator(generator func() []byte) OptsFn {
	return func(c *Middleware) {
		c.Generator = generator
	}
}

func WithCookieName(value string) OptsFn {
	return func(c *Middleware) {
		c.CookieName = value
	}
}

func WithCookiePath(value string) OptsFn {
	return func(c *Middleware) {
		c.CookiePath = value
	}
}

func WithCookieSameSite(value http.SameSite) OptsFn {
	return func(c *Middleware) {
		c.CookieSameSite = value
	}
}

func WithCookieExpiration(value time.Duration) OptsFn {
	return func(c *Middleware) {
		c.MaxAge = value
	}
}

func WithCookieHttpOnly(value bool) OptsFn {
	return func(c *Middleware) {
		c.CookieHttpOnly = value
	}
}

func WithCookieSecure(value bool) OptsFn {
	return func(c *Middleware) {
		c.CookieSecure = value
	}
}

func WithHeaderName(value string) OptsFn {
	return func(c *Middleware) {
		c.HeaderName = value
	}
}

func WithStore(store Store) OptsFn {
	return func(c *Middleware) {
		c.Store = store
	}
}

func FromContextFieldName(ctx context.Context) string {
	v, ok := ctx.Value(FormFieldContextKey).(string)
	if !ok {
		return ""
	}
	return v
}

func FromContext(ctx context.Context) []byte {
	v, ok := ctx.Value(ContextKey).([]byte)
	if !ok {
		return nil
	}
	return v
}

func FromContextStringed(ctx context.Context) string {
	v, ok := ctx.Value(ContextKey).([]byte)
	if !ok {
		return ""
	}
	return base64.StdEncoding.EncodeToString(v)
}
