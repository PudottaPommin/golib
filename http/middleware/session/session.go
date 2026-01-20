package session

import (
	"context"
	"net/http"

	ghttp "github.com/pudottapommin/golib/http"
)

// Handler is the middleware handler function
func (mw *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.Next != nil && mw.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		sessId, err := mw.Store.Get(r)
		if err != nil || len(sessId) != defaultTokenLength {
			sessId = mw.Generator()
			if err = mw.Store.Save(w, sessId); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set(ghttp.HeaderVary, "Cookie")
			w.Header().Add(ghttp.HeaderCacheControl, `no-cache="Set-Cookie"`)
		}

		r = r.WithContext(context.WithValue(r.Context(), ContextKey, sessId))
		next.ServeHTTP(w, r)
	})
}
