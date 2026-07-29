package principal

import (
	"context"
	"errors"
	"net/http"

	principalpkg "github.com/pudottapommin/golib/pkg/principal"
)

// Handler is the middleware handler function.
func (m *Authentication[T]) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		identity, err := m.Service.Authenticate(r)
		if err != nil {
			if m.Revoker != nil && !errors.Is(err, principalpkg.ErrNoIdentity) {
				_ = m.Revoker.Revoke(w, r)
			}
			if m.AllowAnonymous && errors.Is(err, principalpkg.ErrNoIdentity) {
				next.ServeHTTP(w, r)
				return
			}
			m.OnFailure(w, r, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), ContextKey, identity))
		if m.OnSuccess != nil {
			m.OnSuccess(w, r, identity)
		}
		next.ServeHTTP(w, r)
	})
}
