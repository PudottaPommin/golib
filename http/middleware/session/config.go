package session

import (
	"context"
	"net/http"
	"time"

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
		// Store is the session storage backend
		Store      Store
		ContextKey string

		// CookieName is the name of the session cookie.
		// Optional, Default: "_sessionID"
		CookieName string
		// CookiePath defines the path attribute of the session cookie.
		// Optional, Default: "/"
		CookiePath string
		// CookieDomain defines the domain attribute of the CSRF cookie.
		// Optional, Default: ""
		CookieDomain string
		// CookieSameSite defines the SameSite attribute of the session cookie.
		// Optional, Default: http.SameSiteStrictMode
		CookieSameSite http.SameSite
		// MaxAge sets the expiration duration for the session cookie.
		// Optional, Default: 30 minutes
		MaxAge time.Duration
		// CookieHttpOnly defines if the HttpOnly flag should be set on the session cookie.
		// Optional, Default: true
		CookieHttpOnly bool
		// CookieSecure defines if the Secure flag should be set on the session cookie.
		// Optional, Default: true
		CookieSecure bool
	}
)

const (
	ContextKey key = "golib.session"

	CookieName = "_sessionID"

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
		Generator:      func() []byte { return cookie.GenerateRandomKey(defaultTokenLength) },
	}
	for _, opt := range opts {
		opt(m)
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

// WithNext sets the Next handler
func WithNext(next func(http.ResponseWriter, *http.Request) bool) OptsFn {
	return func(m *Middleware) {
		m.Next = next
	}
}

// WithCookieName sets the cookie name
func WithCookieName(name string) OptsFn {
	return func(m *Middleware) {
		m.CookieName = name
	}
}

// WithCookiePath sets the cookie path
func WithCookiePath(path string) OptsFn {
	return func(m *Middleware) {
		m.CookiePath = path
	}
}

// WithCookieSameSite sets the SameSite attribute
func WithCookieSameSite(sameSite http.SameSite) OptsFn {
	return func(m *Middleware) {
		m.CookieSameSite = sameSite
	}
}

// WithCookieExpiration sets the cookie expiration
func WithCookieExpiration(duration time.Duration) OptsFn {
	return func(m *Middleware) {
		m.MaxAge = duration
	}
}

// WithCookieHttpOnly sets the HttpOnly flag
func WithCookieHttpOnly(httpOnly bool) OptsFn {
	return func(m *Middleware) {
		m.CookieHttpOnly = httpOnly
	}
}

// WithCookieSecure sets the Secure flag
func WithCookieSecure(secure bool) OptsFn {
	return func(m *Middleware) {
		m.CookieSecure = secure
	}
}

func FromContext(ctx context.Context) []byte {
	v, ok := ctx.Value(ContextKey).([]byte)
	if !ok {
		return nil
	}
	return v
}
