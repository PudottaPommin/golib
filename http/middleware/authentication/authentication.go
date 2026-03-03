package authentication

import (
	"context"
	"errors"
	"net/http"

	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

func (m *mw[T]) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cv, err := gAuth.GetCookie(r, m.AuthConfig)
		if err != nil {
			if !errors.Is(err, gAuth.ErrorAuthCookieMissing) {
				_ = gAuth.DeleteCookie(r, w, m.AuthConfig)
			}
			if m.NotAuthenticatedHandler != nil {
				m.NotAuthenticatedHandler(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var identity T
		if identity, err = m.Factory(w, r, cv); err == nil {
			*r = *r.WithContext(context.WithValue(r.Context(), m.ContextKey, &identity))
		}
		if m.AfterHandler != nil {
			m.AfterHandler(w, r, &identity)
		}
		next.ServeHTTP(w, r)
	})
}
