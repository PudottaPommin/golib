package csrf

import (
	"context"
	"crypto/subtle"
	"net/http"
	"time"

	ghttp "github.com/pudottapommin/golib/http"
)

func (m *mw) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		var token string
		if v, err := r.Cookie(m.CookieName); err != nil {
			token = m.Generator()
		} else {
			token = v.Value
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			clientToken, _ := r.Cookie(m.CookieName)
			if clientToken == nil || !validateToken(token, clientToken.Value) {
				setCookie(w, r, m, "", -1)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		}
		setCookie(w, r, m, token, int(m.CookieExpiration.Seconds()))
		r.WithContext(context.WithValue(r.Context(), ContextKey, token))
		next.ServeHTTP(w, r)
	})
}

func setCookie(w http.ResponseWriter, r *http.Request, cfg *mw, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     cfg.CookiePath,
		Expires:  time.Now().Add(cfg.CookieExpiration),
		MaxAge:   maxAge,
		Secure:   cfg.CookieSecure,
		HttpOnly: cfg.CookieHttpOnly,
		SameSite: cfg.CookieSameSite,
	})
	r.Header.Set(ghttp.HeaderVary, "Cookie")
}

func validateToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}
