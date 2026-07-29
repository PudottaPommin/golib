package principal

import "net/http"

// Handler is the middleware handler function.
func (m *Authorization[T]) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		identity := FromContext[T](r.Context())
		if identity == nil || !m.Authorize(identity) {
			m.OnFailure(w, r, identity)
			return
		}

		if m.OnSuccess != nil {
			m.OnSuccess(w, r, identity)
		}
		next.ServeHTTP(w, r)
	})
}
