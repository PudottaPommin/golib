package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthCheck func(*gin.Context) bool

func New(opts ...OptsFn) gin.HandlerFunc {
	cfg := defaultConfig
	for i := range opts {
		opts[i](&cfg)
	}

	return func(c *gin.Context) {
		if cfg.Next != nil && cfg.Next(c) {
			c.Next()
			return
		}

		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		if cfg.Probe(c) {
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusServiceUnavailable)
	}
}
