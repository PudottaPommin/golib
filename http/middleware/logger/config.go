package logger

import (
	"log/slog"
	"net/http"
)

type (
	OptsFn func(*mw)
	mw     struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(http.ResponseWriter, *http.Request) bool
		// Name defines name of the logger
		name string
		// Logger defines logger for middleware [slog.Logger]
		logger any
	}
)

func New(opts ...OptsFn) (m *mw) {
	m = new(mw)
	for i := range opts {
		opts[i](m)
	}
	return m
}

// WithNext sets the Next function for the middleware
func WithNext(next func(http.ResponseWriter, *http.Request) bool) OptsFn {
	return func(m *mw) {
		m.Next = next
	}
}

// WithLogger sets the logger for the middleware
func WithLogger(l any) OptsFn {
	switch l := l.(type) {
	case *slog.Logger:
		return func(m *mw) {
			l.Debug("slog.Logger detected for HTTP")
			m.logger = l
		}
	case slog.Logger:
		return func(m *mw) {
			l.Debug("slog.Logger detected for HTTP")
			m.logger = &l
		}
	default:
		panic("logger must be *slog.Logger")
	}
}

func WithNamedLogger(l any, name string) OptsFn {
	switch l := l.(type) {
	case *slog.Logger:
		return WithLogger(l.WithGroup(name))
	case slog.Logger:
		return WithLogger(l.WithGroup(name))
	default:
		panic("logger must be *slog.Logger")
	}
}
