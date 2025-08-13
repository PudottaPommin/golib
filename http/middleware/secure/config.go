package secure

import (
	"net/http"
)

type (
	OptsFn func(*mw)
	mw     struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(w http.ResponseWriter, r *http.Request) bool
		// XSSProtection provides protection against cross-site scripting attacks
		// X-XSS-Protection:
		//
		// Optional, Default: "1; mode=block"
		XSSProtection string
		// ContentTypeNosniff prevents MIME-sniffing a response away from declared content-type
		// X-Content-Type-Options: nosniff
		//
		// Optional, Default: "nosniff"
		ContentTypeNosniff string
		// XFrameOptions can be used to indicate whether or not a browser should be allowed to render a page in a <frame>, <iframe> or <object>
		// X-Frame-Options: DENY, SAMEORIGIN
		//
		// Optional, Default: "SAMEORIGIN"
		XFrameOptions string
		// ContentSecurityPolicy helps prevent XSS, clickjacking and other code injection attacks
		// Content-Security-Policy: <policy-directive>; <policy-directive>
		//
		// Optional, Default: ""
		ContentSecurityPolicy string
		// CSPReportOnly would use the `Content-Security-Policy-Report-Only` header instead
		// of the `Content-Security-Policy` header. This allows iterative updates of the
		// content security policy by only reporting the violations that would
		// have occurred instead of blocking the resource.
		//
		// Optional, Default: false
		CSPReportOnly bool
		// ReferrerPolicy sets the `Referrer-Policy` header providing security against
		// leaking potentially sensitive request paths to third parties.
		//
		// Optional, Default: ""
		ReferrerPolicy string
		// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
		// long (in seconds) browsers should remember that this site is only to
		// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
		// man-in-the-middle (MITM) attacks.
		//
		// Optional, Default: 0
		HSTSMaxAge int
		// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
		// header, excluding all subdomains from security policy. It has no effect
		// unless HSTSMaxAge is set to a non-zero value.
		//
		// Optional, Default: false
		HSTSExcludeSubdomains bool
		// HSTSPreloadEnabled will add the preload tag in the `Strict Transport Security`
		// header, which enables the domain to be included in the HSTS preload list
		// maintained by Chrome (and used by Firefox and Safari): https://hstspreload.org/
		//
		// Optional, Default: false
		HSTSPreloadEnabled bool
	}
)

func New(opts ...OptsFn) *mw {
	m := &mw{
		Next:                  nil,
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		ContentSecurityPolicy: "",
		ReferrerPolicy:        "",
	}
	for i := range opts {
		opts[i](m)
	}
	return m
}

func WithNext(fn func(w http.ResponseWriter, r *http.Request) bool) OptsFn {
	return func(c *mw) {
		c.Next = fn
	}
}

func WithXSSProtection(value string) OptsFn {
	return func(c *mw) {
		c.XSSProtection = value
	}
}

func WithContentTypeNosniff(value string) OptsFn {
	return func(c *mw) {
		c.ContentTypeNosniff = value
	}
}

func WithXFrameOptions(value string) OptsFn {
	return func(c *mw) {
		c.XFrameOptions = value
	}
}

func WithContentSecurityPolicy(value string) OptsFn {
	return func(c *mw) {
		c.ContentSecurityPolicy = value
	}
}

func WithReferrerPolicy(value string) OptsFn {
	return func(c *mw) {
		c.ReferrerPolicy = value
	}
}
