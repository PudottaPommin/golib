package requestid

import (
	"net/http"
)

func (m *mw) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		rid := r.Header.Get(m.Header)
		if rid == "" {
			rid = m.Generator()
			r.Header.Set(m.Header, rid)
		}
		// Set the id to ensure that the requestid is in the response
		w.Header().Set(m.Header, rid)
		next.ServeHTTP(w, r)
	})
}

func Get(r *http.Request) string {
	return r.Header.Get(headerXRequestID)
}
