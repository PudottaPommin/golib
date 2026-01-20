package csrf

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pudottapommin/golib/http/cookie"
)

type Store interface {
	Get(*http.Request) ([]byte, error)
	Save(w http.ResponseWriter, token []byte) error
}

type cookieStore struct {
	sc *cookie.Cookie

	name     string
	secure   bool
	httpOnly bool
	sameSite http.SameSite
	maxAge   time.Duration
	path     string
	domain   string
}

func (cs *cookieStore) Get(r *http.Request) ([]byte, error) {
	ck, err := r.Cookie(cs.name)
	if err != nil {
		return nil, fmt.Errorf("[CSRF] failed to get cookie: %w", err)
	}
	if err = ck.Valid(); err != nil {
		return nil, fmt.Errorf("[CSRF] cookie is invalid: %w", err)
	}

	token := make([]byte, defaultTokenLength)
	if err = cs.sc.Decrypt(cs.name, ck.Value, &token); err != nil {
		return nil, fmt.Errorf("[CSRF] failed to decrypt cookie: %w", err)
	}
	return token, nil
}

func (cs *cookieStore) Save(w http.ResponseWriter, token []byte) error {
	encoded, err := cs.sc.Secure(cs.name, token)
	if err != nil {
		return fmt.Errorf("[CSRF] failed to secure token: %w", err)
	}
	ck := &http.Cookie{
		Name:     cs.name,
		Value:    string(encoded),
		Path:     cs.path,
		Domain:   cs.domain,
		MaxAge:   int(cs.maxAge.Seconds()),
		Secure:   cs.secure,
		HttpOnly: cs.httpOnly,
		SameSite: cs.sameSite,
	}
	if cs.maxAge > 0 {
		ck.Expires = time.Now().Add(cs.maxAge)
	}
	http.SetCookie(w, ck)
	return nil
}
