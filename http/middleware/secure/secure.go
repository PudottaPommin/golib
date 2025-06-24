package secure

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pudottapommin/golib/http"
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

		if cfg.XSSProtection != "" {
			c.Header(http.HeaderXXSSProtection, cfg.XSSProtection)
		}
		if cfg.ContentTypeNosniff != "" {
			c.Header(http.HeaderXContentTypeOptions, cfg.ContentTypeNosniff)
		}
		if cfg.XFrameOptions != "" {
			c.Header(http.HeaderXFrameOptions, cfg.XFrameOptions)
		}
		if cfg.ContentSecurityPolicy != "" {
			if cfg.CSPReportOnly {
				c.Header(http.HeaderContentSecurityPolicyReportOnly, cfg.ContentSecurityPolicy)
			} else {
				c.Header(http.HeaderContentSecurityPolicy, cfg.ContentSecurityPolicy)
			}
		}
		if cfg.ReferrerPolicy != "" {
			c.Header(http.HeaderReferrerPolicy, cfg.ReferrerPolicy)
		}

		if (c.Request.TLS != nil || (c.Request.Header.Get(http.HeaderXForwardedProto) == "https")) && cfg.HSTSMaxAge > 0 {
			var subdomains string
			if !cfg.HSTSExcludeSubdomains {
				subdomains = "; includeSubdomains"
			}
			if cfg.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}
			c.Header(http.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", cfg.HSTSMaxAge, subdomains))
		}

		c.Next()
	}
}
