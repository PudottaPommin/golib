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
			cfg.UnauthorizedHandler(c, *new(T))
			return
		}
		iden, ok := cIden.(T)
		if !ok {
			cfg.UnauthorizedHandler(c, iden)
			return
		}
		if !cfg.AuthorizeHandler(c, iden) {
			cfg.UnauthorizedHandler(c, iden)
			return
		}
		if cfg.AuthorizedHandler != nil {
			cfg.AuthorizedHandler(c, iden)
		}
		c.Next()
	}
}
