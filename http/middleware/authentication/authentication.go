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
		if errors.Is(err, gAuth.ErrorAuthCookieMissing) {
			if m.NotAuthenticatedHandler != nil {
				m.NotAuthenticatedHandler(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var identity T
		if identity, err = m.Factory(w, r, cv); err == nil {
			r.WithContext(context.WithValue(r.Context(), m.ContextKey, identity))
		}
		if m.AfterHandler != nil {
			m.AfterHandler(w, r, &identity)
		}
		next.ServeHTTP(w, r)
	})
}
