package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	ErrorAuthCookieMissing = errors.New("auth: Cookie is missing")
)

type CookieValue struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	SecurityStamp []byte    `json:"securityStamp"`
	Timestamp     time.Time `json:"timestamp"`
	Signature     []byte    `json:"signature"`
}

func (cv *CookieValue) Digest() []byte {
	bytes := make([]byte, 49)
	copy(bytes, cv.ID[:])
	bytes[16] = byte(';')
	copy(bytes[17:], cv.SecurityStamp)
	bytes[33] = byte(';')
	timebytes, _ := cv.Timestamp.MarshalBinary()
	copy(bytes[34:], timebytes)
	return bytes
}

func NewCookieValue(id uuid.UUID, securityStamp []byte) CookieValue {
	return CookieValue{ID: id, SecurityStamp: securityStamp}
}

func WithCookie(c echo.Context, cfg *Config, cv CookieValue) error {
	cv.Timestamp = time.Now().UTC().Add(cfg.expiration)
	token, err := encodeAuthToken(cfg.SigningKey(), cv)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = cfg.cookieName
	cookie.Value = token
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Expires = cv.Timestamp
	c.SetCookie(cookie)

	return nil
}

func GetCookie(c echo.Context, cfg *Config) (cv *CookieValue, err error) {
	cookie, err := c.Cookie(cfg.cookieName)
	if err != nil {
		return cv, ErrorAuthCookieMissing
	}

	err = cookie.Valid()
	if err != nil {
		return
	}

	cv, err = decodeAuthToken(cfg.SigningKey(), cookie.Value)
	if err == nil && cfg.isSliding && cv.Timestamp.UTC().Sub(time.Now().UTC()) < (time.Minute*15) {
		err = WithCookie(c, cfg, *cv)
	}
	return
}

func DeleteCookie(c echo.Context, cfg *Config) error {
	cookie, err := c.Cookie(cfg.cookieName)
	if err != nil {
		return err
	}
	cookie.MaxAge = -1
	c.SetCookie(cookie)
	return nil
}
