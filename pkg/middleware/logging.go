package middleware

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	eMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/pudottapommin/golib/pkg/log"
	"github.com/rs/zerolog"
)

type (
	// LoggingConfig is the configuration for the middleware.
	LoggingConfig struct {
		// Logger is a custom instance of the logger to use.
		Logger *log.Logger
		// Skipper defines a function to skip middleware.
		Skipper eMiddleware.Skipper
		// AfterNextSkipper defines a function to skip middleware after the next handler is called.
		AfterNextSkipper eMiddleware.Skipper
		// BeforeNext is a function executed before the next handler is called.
		BeforeNext eMiddleware.BeforeFunc
		// Enricher is a function that can be used to enrich the logger with additional information.
		Enricher Enricher
		// RequestIDHeader is the header name to use for the request ID in a log record.
		RequestIDHeader string
		// RequestIDKey is the MiddlewareKey name to use for the request ID in a log record.
		RequestIDKey string
		// NestKey is the MiddlewareKey name to use for the nested logger in a log record.
		NestKey string
		// HandleError indicates whether to propagate errors up the middleware chain, so the global error handler can decide appropriate status code.
		HandleError bool
		// For long-running requests that take longer than this limit, log at a different level.  Ignored by default
		RequestLatencyLimit time.Duration
		// The level to log at if RequestLatencyLimit is exceeded
		RequestLatencyLevel zerolog.Level
	}

	// Enricher is a function that can be used to enrich the logger with additional information.
	Enricher func(c echo.Context, logger zerolog.Context) zerolog.Context
)

// Logging Middleware returns a middleware which logs HTTP requests.
func Logging(config LoggingConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = eMiddleware.DefaultSkipper
	}

	if config.AfterNextSkipper == nil {
		config.AfterNextSkipper = eMiddleware.DefaultSkipper
	}

	if config.Logger == nil {
		config.Logger = log.New(os.Stdout, log.WithTimestamp())
	}

	if config.RequestIDKey == "" {
		config.RequestIDKey = "id"
	}

	if config.RequestIDHeader == "" {
		config.RequestIDHeader = echo.HeaderXRequestID
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var err error
			req := c.Request()
			res := c.Response()
			start := time.Now()

			id := req.Header.Get(config.RequestIDHeader)

			if id == "" {
				id = res.Header().Get(config.RequestIDHeader)
			}

			cloned := false
			logger := config.Logger

			if id != "" {
				logger = log.From(logger.Log, log.WithLevel(config.Logger.Level()), log.WithField(config.RequestIDKey, id))
				cloned = true
			}

			if config.Enricher != nil {
				// to avoid mutation of shared instance
				if !cloned {
					logger = log.From(logger.Log, log.WithLevel(config.Logger.Level()))
					cloned = true
				}

				logger.Log = config.Enricher(c, logger.Log.With()).Logger()
			}

			ctx := req.Context()

			if ctx == nil {
				ctx = context.Background()
			}

			// Pass logger down to request context
			c.SetRequest(req.WithContext(logger.WithContext(ctx)))
			// c = NewContext(c, logger)

			if config.BeforeNext != nil {
				config.BeforeNext(c)
			}

			if err = next(c); err != nil {
				if config.HandleError {
					c.Error(err)
				}
			}

			if config.AfterNextSkipper(c) {
				return err
			}

			stop := time.Now()
			latency := stop.Sub(start)
			var mainEvt *zerolog.Event
			if err != nil {
				mainEvt = logger.Log.Err(err)
			} else if config.RequestLatencyLimit != 0 && latency > config.RequestLatencyLimit {
				mainEvt = logger.Log.WithLevel(config.RequestLatencyLevel)
			} else {
				mainEvt = logger.Log.WithLevel(logger.Log.GetLevel())
			}

			var evt *zerolog.Event
			if config.NestKey != "" { // Start a new event (dict) if there's a nest MiddlewareKey.
				evt = zerolog.Dict()
			} else {
				evt = mainEvt
			}

			evt.Str("remote_ip", c.RealIP()).
				Str("host", req.Host).
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("user_agent", req.UserAgent()).
				Int("status", res.Status).
				Str("referer", req.Referer()).
				Dur("latency", latency).
				Str("latency_human", latency.String())

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}

			evt.Str("bytes_in", cl).
				Str("bytes_out", strconv.FormatInt(res.Size, 10))

			if config.NestKey != "" { // Nest the new event (dict) under the nest MiddlewareKey.
				mainEvt.Dict(config.NestKey, evt)
			}
			mainEvt.Send()
			return err
		}
	}
}
