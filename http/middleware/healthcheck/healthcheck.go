package healthcheck

import (
	"net/http"
)

type HealthCheck func(http.ResponseWriter, *http.Request) bool

func (m *mw) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		if m.Probe(w, r) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}
