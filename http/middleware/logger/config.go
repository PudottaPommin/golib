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

var statusCodeToLogLevel = map[int]slog.Level{
	// 1xx Informational
	http.StatusContinue:           slog.LevelDebug,
	http.StatusSwitchingProtocols: slog.LevelDebug,
	http.StatusProcessing:         slog.LevelDebug,
	http.StatusEarlyHints:         slog.LevelDebug,

	// 2xx Success
	http.StatusOK:                   slog.LevelInfo,
	http.StatusCreated:              slog.LevelInfo,
	http.StatusAccepted:             slog.LevelInfo,
	http.StatusNonAuthoritativeInfo: slog.LevelInfo,
	http.StatusNoContent:            slog.LevelInfo,
	http.StatusResetContent:         slog.LevelInfo,
	http.StatusPartialContent:       slog.LevelInfo,
	http.StatusMultiStatus:          slog.LevelInfo,
	http.StatusAlreadyReported:      slog.LevelInfo,
	http.StatusIMUsed:               slog.LevelInfo,

	// 3xx Redirection
	http.StatusMultipleChoices:   slog.LevelInfo,
	http.StatusMovedPermanently:  slog.LevelInfo,
	http.StatusFound:             slog.LevelInfo,
	http.StatusSeeOther:          slog.LevelInfo,
	http.StatusNotModified:       slog.LevelDebug,
	http.StatusUseProxy:          slog.LevelInfo,
	http.StatusTemporaryRedirect: slog.LevelInfo,
	http.StatusPermanentRedirect: slog.LevelInfo,

	// 4xx Client Error
	http.StatusBadRequest:                   slog.LevelWarn,
	http.StatusUnauthorized:                 slog.LevelWarn,
	http.StatusPaymentRequired:              slog.LevelWarn,
	http.StatusForbidden:                    slog.LevelWarn,
	http.StatusNotFound:                     slog.LevelWarn,
	http.StatusMethodNotAllowed:             slog.LevelWarn,
	http.StatusNotAcceptable:                slog.LevelWarn,
	http.StatusProxyAuthRequired:            slog.LevelWarn,
	http.StatusRequestTimeout:               slog.LevelWarn,
	http.StatusConflict:                     slog.LevelWarn,
	http.StatusGone:                         slog.LevelWarn,
	http.StatusLengthRequired:               slog.LevelWarn,
	http.StatusPreconditionFailed:           slog.LevelWarn,
	http.StatusRequestEntityTooLarge:        slog.LevelWarn,
	http.StatusRequestURITooLong:            slog.LevelWarn,
	http.StatusUnsupportedMediaType:         slog.LevelWarn,
	http.StatusRequestedRangeNotSatisfiable: slog.LevelWarn,
	http.StatusExpectationFailed:            slog.LevelWarn,
	http.StatusTeapot:                       slog.LevelWarn,
	http.StatusMisdirectedRequest:           slog.LevelWarn,
	http.StatusUnprocessableEntity:          slog.LevelWarn,
	http.StatusLocked:                       slog.LevelWarn,
	http.StatusFailedDependency:             slog.LevelWarn,
	http.StatusTooEarly:                     slog.LevelWarn,
	http.StatusUpgradeRequired:              slog.LevelWarn,
	http.StatusPreconditionRequired:         slog.LevelWarn,
	http.StatusTooManyRequests:              slog.LevelWarn,
	http.StatusRequestHeaderFieldsTooLarge:  slog.LevelWarn,
	http.StatusUnavailableForLegalReasons:   slog.LevelWarn,

	// 5xx Server Error
	http.StatusInternalServerError:           slog.LevelError,
	http.StatusNotImplemented:                slog.LevelError,
	http.StatusBadGateway:                    slog.LevelError,
	http.StatusServiceUnavailable:            slog.LevelError,
	http.StatusGatewayTimeout:                slog.LevelError,
	http.StatusHTTPVersionNotSupported:       slog.LevelError,
	http.StatusVariantAlsoNegotiates:         slog.LevelError,
	http.StatusInsufficientStorage:           slog.LevelError,
	http.StatusLoopDetected:                  slog.LevelError,
	http.StatusNotExtended:                   slog.LevelError,
	http.StatusNetworkAuthenticationRequired: slog.LevelError,
}

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
