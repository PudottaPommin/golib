package log

import (
	"context"

	"github.com/rs/zerolog"
)

// WithContext returns a new context with the provided logger.
func (l *Logger) WithContext(ctx context.Context) context.Context {
	zl := l.Unwrap()
	return zl.WithContext(ctx)
}

// Ctx returns a logger from the provided context.
// If no logger is found in the context, a new one is created.
func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
