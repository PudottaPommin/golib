package requestid

import (
	"net/http"

	ghttp "github.com/pudottapommin/golib/http"
	"github.com/pudottapommin/golib/pkg/id"
)

type (
	OptsFn func(*Config)
	Config struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(http.ResponseWriter, *http.Request) bool
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
		Header: ghttp.HeaderXRequestID,
	}
	headerXRequestID = ghttp.HeaderXRequestID
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

func WithNext(fn func(w http.ResponseWriter, r *http.Request) bool) OptsFn {
	return func(c *Config) {
		c.Next = fn
	}
}
