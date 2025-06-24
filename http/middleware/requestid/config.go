package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/pudottapommin/golib/http"
	"github.com/pudottapommin/golib/pkg/id"
)

type (
	OptsFn func(*Config)
	Config struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(c *gin.Context) bool
		// Generator defines function which returns new request id
		//
		// Optional, Default: github.com/pudottapommin/golib/pkg/id
		Generator func() string
		// Header is the header key where to get/set the unique request ID
		//
		// Optional. Default: "X-Request-ID"
		Header string
	}
)

var (
	defaultConfig = Config{
		Generator: func() string {
			return id.New().String()
		},
		Header: http.HeaderXRequestID,
	}
	headerXRequestID = http.HeaderXRequestID
)

func WithGenerator(fn func() string) OptsFn {
	return func(c *Config) {
		c.Generator = fn
	}
}

func WithHeader(header string) OptsFn {
	return func(c *Config) {
		c.Header = header
	}
}

func WithNext(fn func(c *gin.Context) bool) OptsFn {
	return func(c *Config) {
		c.Next = fn
	}
}
