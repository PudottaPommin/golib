package logger

import (
	"net/http"

	"go.uber.org/zap"
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
		// Logger defines logger for middleware either [zap.Logger] or [zap.SugaredLogger]
		logger any
	}
)

func New(opts ...OptsFn) *mw {
	m := &mw{}
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
func WithLogger(logger any, name string) OptsFn {
	switch logger := logger.(type) {
	case *zap.Logger:
		return func(m *mw) {
			logger = zap.New(logger.Core(), zap.AddCallerSkip(1)).Named(name)
			logger.Debug("zap.logger detected for HTTP")
			m.logger = logger
		}
	case *zap.SugaredLogger:
		return func(m *mw) {
			logger = zap.New(logger.Desugar().Core(), zap.AddCallerSkip(1)).Sugar().Named(name)
			logger.Debug("zap.SugaredLogger logger detected for HTTP")
			m.logger = logger
		}
	default:
		panic("logger must be *zap.Logger or *zap.SuggarLogged")
	}
}
