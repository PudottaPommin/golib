package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
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
			if l, ok := m.logger.(*zap.Logger); ok {
				l.Info("served",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", ww.Status()),
					zap.String("statusText", statusLabel(ww.Status())),
					zap.String("reqId", middleware.GetReqID(r.Context())),
					zap.String("remoteAddr", r.RemoteAddr),
					zap.String("proto", r.Proto),
					zap.Duration("latency", time.Since(t1)),
					zap.Int("size", ww.BytesWritten()))
			} else if s, ok := m.logger.(*zap.SugaredLogger); ok {
				s.Infow("served",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", ww.Status()),
					zap.String("statusText", statusLabel(ww.Status())),
					zap.String("reqId", middleware.GetReqID(r.Context())),
					zap.String("remoteAddr", r.RemoteAddr),
					zap.String("proto", r.Proto),
					zap.Duration("latency", time.Since(t1)),
					zap.Int("size", ww.BytesWritten()))
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
