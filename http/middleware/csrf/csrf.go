package csrf

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ghttp "github.com/pudottapommin/golib/http"
)

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

		var token string
		if v, err := c.Request.Cookie(cfg.CookieName); err != nil {
			token = cfg.Generator()
		} else {
			token = v.Value
		}

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			clientToken, _ := c.Cookie(cfg.CookieName)
			if !validateToken(token, clientToken) {
				setCookie(c, &cfg, "", -1)
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		setCookie(c, &cfg, token, int(cfg.CookieExpiration.Seconds()))
		c.Set(ContextKey, token)
		c.Next()
	}
}

func setCookie(c *gin.Context, cfg *Config, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     cfg.CookiePath,
		Expires:  time.Now().Add(cfg.CookieExpiration),
		MaxAge:   maxAge,
		Secure:   cfg.CookieSecure,
		HttpOnly: cfg.CookieHttpOnly,
		SameSite: cfg.CookieSameSite,
	})
	c.Header(ghttp.HeaderVary, "Cookie")
}

func validateToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}
