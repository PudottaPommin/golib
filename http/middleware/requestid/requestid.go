package requestid

import (
	"github.com/gin-gonic/gin"
)

func New(opts ...OptsFn) gin.HandlerFunc {
	cfg := defaultConfig
	for i := range opts {
		opts[i](&cfg)
	}
	headerXRequestID = cfg.Header
	return func(c *gin.Context) {
		if cfg.Next != nil && cfg.Next(c) {
			c.Next()
			return
		}

		rid := c.GetHeader(cfg.Header)
		if rid == "" {
			rid = cfg.Generator()
			c.Request.Header.Set(cfg.Header, rid)
		}
		// Set the id to ensure that the requestid is in the response
		c.Header(cfg.Header, rid)
		c.Next()
	}
}

func Get(c *gin.Context) string {
	return c.GetHeader(headerXRequestID)
}
