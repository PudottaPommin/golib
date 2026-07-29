package cookie

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	gcookie "github.com/pudottapommin/golib/http/cookie"
	"github.com/pudottapommin/golib/pkg/principal"
)

const defaultMaxAge = time.Hour * 8

var (
	// ErrHashKeyRequired is returned by NewCookieStore when hashKey is empty. Without a hash key
	// http/cookie.Cookie's HMAC is computed with an empty key, which anyone can forge without
	// knowing any secret.
	ErrHashKeyRequired = errors.New("principal/cookie: hashKey is required")
	// ErrBlockKeyRequired is returned by NewCookieStore when blockKey is nil. Without a block key
	// http/cookie.Cookie installs no encryptor, leaving the identity (including the username) as
	// readable base64 in the browser.
	ErrBlockKeyRequired = errors.New("principal/cookie: blockKey is required")
)

type cookieIdentity[T comparable] struct {
	ID       T      `json:"id"`
	Username string `json:"username"`
	Stamp    []byte `json:"stamp,omitempty"`
}

type Factory func(name, value string) *http.Cookie

type Store[T comparable] struct {
	cookieName    string
	sc            *gcookie.Cookie
	cookieFactory Factory

	cookiePath     string
	cookieDomain   string
	cookieSameSite http.SameSite
	cookieHTTPOnly bool
	cookieSecure   bool
}

var _ principal.IdentityStore[string] = (*Store[string])(nil)

func NewCookieStore[T comparable](
	cookieName string,
	hashKey, blockKey []byte,
	options ...gcookie.OptFn,
) (*Store[T], error) {
	if len(hashKey) == 0 {
		return nil, ErrHashKeyRequired
	}
	if blockKey == nil {
		return nil, ErrBlockKeyRequired
	}

	opts := append([]gcookie.OptFn{gcookie.WithMaxAge(int64(defaultMaxAge.Seconds()))}, options...)
	sc, err := gcookie.New(hashKey, blockKey, opts...)
	if err != nil {
		return nil, fmt.Errorf("[Principal] failed to create cookie store: %w", err)
	}

	cs := &Store[T]{
		cookieName:     cookieName,
		sc:             sc,
		cookieFactory:  nil,
		cookiePath:     "/",
		cookieDomain:   "",
		cookieSameSite: http.SameSiteStrictMode,
		cookieHTTPOnly: true,
		cookieSecure:   true,
	}
	cs.cookieFactory = cs.defaultCookieFactory
	return cs, nil
}

// WithCookieFactory overrides how the identity's [net/http.Cookie] is built entirely, taking
// precedence over WithCookiePath/WithCookieDomain/WithCookieSameSite/WithCookieHTTPOnly/
// WithCookieSecure.
func (cs *Store[T]) WithCookieFactory(factory Factory) *Store[T] {
	cs.cookieFactory = factory
	return cs
}

// WithCookiePath sets the Path attribute of the identity cookie. Default: "/".
func (cs *Store[T]) WithCookiePath(path string) *Store[T] {
	cs.cookiePath = path
	return cs
}

// WithCookieDomain sets the Domain attribute of the identity cookie. Default: "".
func (cs *Store[T]) WithCookieDomain(domain string) *Store[T] {
	cs.cookieDomain = domain
	return cs
}

// WithCookieSameSite sets the SameSite attribute of the identity cookie. Default:
// [net/http.SameSiteStrictMode].
func (cs *Store[T]) WithCookieSameSite(sameSite http.SameSite) *Store[T] {
	cs.cookieSameSite = sameSite
	return cs
}

// WithCookieHTTPOnly sets the HttpOnly attribute of the identity cookie. Default: true.
func (cs *Store[T]) WithCookieHTTPOnly(httpOnly bool) *Store[T] {
	cs.cookieHTTPOnly = httpOnly
	return cs
}

// WithCookieSecure sets the Secure attribute of the identity cookie. Default: true.
func (cs *Store[T]) WithCookieSecure(secure bool) *Store[T] {
	cs.cookieSecure = secure
	return cs
}

func (cs *Store[T]) Resolve(r *http.Request) (principal.Identity[T], error) {
	c, err := r.Cookie(cs.cookieName)
	if err != nil {
		return nil, principal.ErrNoIdentity
	}
	if err = c.Valid(); err != nil {
		return nil, fmt.Errorf("[Principal] failed to validate cookie: %w: %w", principal.ErrInvalidIdentity, err)
	}

	var identity cookieIdentity[T]
	if err = cs.sc.Decrypt(cs.cookieName, c.Value, &identity); err != nil {
		return nil, fmt.Errorf("[Principal] failed to decrypt cookie: %w: %w", principal.ErrInvalidIdentity, err)
	}

	return principal.NewUser[T](identity.ID, identity.Username, identity.Stamp), nil
}

func (cs *Store[T]) Store(w http.ResponseWriter, _ *http.Request, identity principal.Identity[T]) error {
	ci := cookieIdentity[T]{
		ID:       identity.ID(),
		Username: identity.Username(),
		Stamp:    identity.SecurityStamp(),
	}

	secured, err := cs.sc.Secure(cs.cookieName, ci)
	if err != nil {
		return fmt.Errorf("[Principal] failed to secure identity: %w", err)
	}

	//nolint:gosec // attributes are set by cookieFactory
	c := cs.cookieFactory(cs.cookieName, string(secured))
	if cs.sc.MaxAge() > 0 {
		c.MaxAge = int(cs.sc.MaxAge())
		c.Expires = time.Now().Add(time.Duration(cs.sc.MaxAge()) * time.Second)
	}

	http.SetCookie(w, c)

	return nil
}

func (cs *Store[T]) Revoke(w http.ResponseWriter, _ *http.Request) error {
	//nolint:gosec // attributes are set by cookieFactory
	c := cs.cookieFactory(cs.cookieName, "")
	c.MaxAge = -1
	c.Expires = time.Unix(0, 0)

	http.SetCookie(w, c)

	return nil
}

func (cs *Store[T]) defaultCookieFactory(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cs.cookiePath,
		Domain:   cs.cookieDomain,
		HttpOnly: cs.cookieHTTPOnly,
		Secure:   cs.cookieSecure,
		SameSite: cs.cookieSameSite,
	}
}
