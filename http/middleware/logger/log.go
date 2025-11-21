package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func (m *mw) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		if m.logger == nil {
			next.ServeHTTP(w, r)
			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			if l, ok := m.logger.(*zerolog.Logger); ok {
				l.Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", ww.Status()).
					Str("statusText", statusLabel(ww.Status())).
					Str("reqId", middleware.GetReqID(r.Context())).
					Str("remoteAddr", r.RemoteAddr).
					Str("proto", r.Proto).
					Dur("latency", time.Since(t1)).
					Int("size", ww.BytesWritten()).
					Msg("served")
			}
		}()
		next.ServeHTTP(ww, r)
	})
}

func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return fmt.Sprintf("%d OK", status)
	case status >= 300 && status < 400:
		return fmt.Sprintf("%d Redirect", status)
	case status >= 400 && status < 500:
		return fmt.Sprintf("%d Client Error", status)
	case status >= 500:
		return fmt.Sprintf("%d Server Error", status)
	default:
		return fmt.Sprintf("%d Unknown", status)
	}
}
