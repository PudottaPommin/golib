package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
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

func NewCookieValue(id uuid.UUID, username string, securityStamp []byte) *CookieValue {
	return &CookieValue{ID: id, Username: username, SecurityStamp: securityStamp}
}

func (cv *CookieValue) Digest() []byte {
	bytes := make([]byte, 49)
	copy(bytes, cv.ID[:])
	bytes[16] = byte(';')
	copy(bytes[17:], cv.SecurityStamp)
	bytes[33] = byte(';')
	tb, _ := cv.Timestamp.MarshalBinary()
	copy(bytes[34:], tb)
	return bytes
}
func (cv *CookieValue) WriteToRequest(w http.ResponseWriter, cfg *Config) error {
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
	http.SetCookie(w, cookie)
	return nil
}

func GetCookie(r *http.Request, cfg *Config) (cv *CookieValue, err error) {
	cookie, err := r.Cookie(cfg.cookieName)
	if err != nil {
		return cv, ErrorAuthCookieMissing
	}

	err = cookie.Valid()
	if err != nil {
		return
	}

	cv, err = decodeAuthToken(cfg.SigningKey(), cookie.Value)
	if err == nil && cfg.isSliding && cv.Timestamp.UTC().Sub(time.Now().UTC()) < (time.Minute*15) {
		// @todo
		// err = WithCookie(c, cfg, *cv)
	}
	return
}

func DeleteCookie(r *http.Request, w http.ResponseWriter, cfg *Config) error {
	cookie, err := r.Cookie(cfg.cookieName)
	if err != nil {
		return err
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	return nil
}
