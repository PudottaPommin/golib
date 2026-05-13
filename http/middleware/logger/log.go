package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
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
			switch logger := m.logger.(type) {
			case *slog.Logger:
				level := statusCodeToLogLevel[ww.Status()]
				logger.Log(r.Context(), level, "response",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", ww.Status()),
					slog.String("statusText", statusLabel(ww.Status())),
					slog.String("reqId", r.Header.Get("X-Request-Id")),
					slog.String("remoteAddr", r.RemoteAddr),
					slog.String("proto", r.Proto),
					slog.Duration("latency", time.Since(t1)),
					slog.Int("size", ww.BytesWritten()))
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
