package etag

import (
	"github.com/gin-gonic/gin"
	"github.com/pudottapommin/golib/http"
)

type (
	OptsFn func(*Config)
	Config struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(c *gin.Context) bool
		// Weak enables the use of weak ETags by prefixing them with 'W/'.
		// When true, the ETag is considered weak for comparison, allowing content to be
		// semantically equivalent but not byte-for-byte identical.
		Weak bool
	}
)

var (
	defaultConfig = Config{
		Next: nil,
		Weak: false,
	}
	headerETag        = http.HeaderETag
	headerIfNoneMatch = http.HeaderIfNoneMatch
	weakPrefix        = []byte("W/")
)

func WithWeak() OptsFn {
	return func(c *Config) {
		c.Weak = true
	}
}

func WithNext(fn func(c *gin.Context) bool) OptsFn {
	return func(c *Config) {
		c.Next = fn
	}
}
