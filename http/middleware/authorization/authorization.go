package authorization

import (
	"net/http"
)

func (m *mw[T]) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iden, ok := r.Context().Value(m.ContextKey).(T)
		if !ok {
			m.UnauthorizedHandler(w, r, *new(T))
			return
		}

		if iden == nil {
			m.UnauthorizedHandler(w, r, iden)
			return
		}
		if !m.AuthorizeHandler(w, r, iden) {
			m.UnauthorizedHandler(w, r, iden)
			return
		}
		if m.AuthorizedHandler != nil {
			m.AuthorizedHandler(w, r, iden)
		}
		next.ServeHTTP(w, r)
	})
}
