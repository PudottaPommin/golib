package csrf

import (
	"context"
	"encoding/base64"
	"net/http"

	ghttp "github.com/pudottapommin/golib/http"
	"github.com/pudottapommin/golib/http/extractors"
)

func (mw *Middleware) WithHTTPSecHandler(next http.Handler) http.Handler {
	return http.NewCrossOriginProtection().Handler(mw.Handler(next))
}

func (mw *Middleware) Handler(next http.Handler) http.Handler {
	requestExtractor := extractors.Chain(extractors.FromHeader(mw.HeaderName), extractors.FromForm(mw.FormFieldName))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.Next != nil && mw.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		realToken, err := mw.getRealToken(r)
		if err != nil || len(realToken) != defaultTokenLength {
			realToken = mw.Generator()
			if err = mw.Store.Save(w, realToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set(ghttp.HeaderVary, "Cookie")
			w.Header().Add(ghttp.HeaderCacheControl, `no-cache="Set-Cookie"`)
		}

		ctx := context.WithValue(r.Context(), ContextKey, maskToken(realToken))
		ctx = context.WithValue(ctx, FormFieldContextKey, mw.FormFieldName)
		r = r.WithContext(ctx)

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			maskedRequestToken, err := mw.getRequestToken(requestExtractor, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if maskedRequestToken == nil {
				http.Error(w, "missing CSRF token", http.StatusBadRequest)
				return
			}
			requestToken := unmaskToken(maskedRequestToken)
			if !validateToken(realToken, requestToken) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (mw *Middleware) getRealToken(r *http.Request) ([]byte, error) {
	return mw.Store.Get(r)
}

func (mw *Middleware) getRequestToken(extractor extractors.Extractor, r *http.Request) ([]byte, error) {
	token, err := extractor.Extract(r)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
