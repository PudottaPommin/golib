package secure

import (
	"fmt"
	"net/http"

	ghttp "github.com/pudottapommin/golib/http"
)

func (m *mw) Handler(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Next != nil && m.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}

		if m.XSSProtection != "" {
			w.Header().Set(ghttp.HeaderXXSSProtection, m.XSSProtection)
		}
		if m.ContentTypeNosniff != "" {
			w.Header().Set(ghttp.HeaderXContentTypeOptions, m.ContentTypeNosniff)
		}
		if m.XFrameOptions != "" {
			w.Header().Set(ghttp.HeaderXFrameOptions, m.XFrameOptions)
		}
		if m.ContentSecurityPolicy != "" {
			if m.CSPReportOnly {
				w.Header().Set(ghttp.HeaderContentSecurityPolicyReportOnly, m.ContentSecurityPolicy)
			} else {
				w.Header().Set(ghttp.HeaderContentSecurityPolicy, m.ContentSecurityPolicy)
			}
		}
		if m.ReferrerPolicy != "" {
			w.Header().Set(ghttp.HeaderReferrerPolicy, m.ReferrerPolicy)
		}

		if (r.TLS != nil || (r.Header.Get(ghttp.HeaderXForwardedProto) == "https")) && m.HSTSMaxAge > 0 {
			var subdomains string
			if !m.HSTSExcludeSubdomains {
				subdomains = "; includeSubdomains"
			}
			if m.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}
			w.Header().Set(ghttp.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", m.HSTSMaxAge, subdomains))
		}

		next.ServeHTTP(w, r)
	})
}
