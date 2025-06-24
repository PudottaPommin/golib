package authorization

import (
	"github.com/gin-gonic/gin"

	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

func New[T gAuth.Identity](opts ...OptsFn[T]) gin.HandlerFunc {
	cfg := newDefaultConfig[T]()
	for i := range opts {
		opts[i](&cfg)
	}
	return func(c *gin.Context) {
		cIden, ok := c.Get(cfg.ContextKey)
		if !ok {
			cfg.UnauthorizedHandler(c, nil)
			return
		}
		iden, ok := cIden.(*T)
		if !ok {
			cfg.UnauthorizedHandler(c, nil)
			return
		}
		if !cfg.AuthorizeHandler(c, iden) {
			cfg.UnauthorizedHandler(c, nil)
			return
		}
		cfg.AuthorizedHandler(c, iden)
		c.Next()
	}
}
