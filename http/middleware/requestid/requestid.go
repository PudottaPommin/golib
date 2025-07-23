package requestid

import (
	"net/http"
)

func New(opts ...OptsFn) func(http.Handler) http.Handler {
	cfg := defaultConfig
	for i := range opts {
		opts[i](&cfg)
	}
	headerXRequestID = cfg.Header
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(w, r) {
				next.ServeHTTP(w, r)
				return
			}

			rid := r.Header.Get(cfg.Header)
			if rid == "" {
				rid = cfg.Generator()
				r.Header.Set(cfg.Header, rid)
			}
			// Set the id to ensure that the requestid is in the response
			w.Header().Set(cfg.Header, rid)
			next.ServeHTTP(w, r)
		})
	}
}

func Get(r *http.Request) string {
	return r.Header.Get(headerXRequestID)
}
