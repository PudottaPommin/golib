package authentication

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	gAuth "github.com/pudottapommin/golib/pkg/auth"
)

func New[T gAuth.Identity](opts ...OptsFn[T]) gin.HandlerFunc {
	cfg := newDefaultConfig[T]()
	for i := range opts {
		opts[i](&cfg)
	}
	return func(c *gin.Context) {
		cv, err := gAuth.GetCookie(c.Request, cfg.AuthConfig)
		if errors.Is(err, gAuth.ErrorAuthCookieMissing) {
			if cfg.NotAuthenticatedHandler != nil {
				cfg.NotAuthenticatedHandler(c)
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if err != nil {
			c.Error(err)
		}
		var identity *T
		if identity, err = cfg.Factory(c, cv); err == nil {
			c.Set(cfg.ContextKey, identity)
		}

		if cfg.AfterHandler != nil {
			cfg.AfterHandler(c, identity)
		}

		c.Next()
	}
}
