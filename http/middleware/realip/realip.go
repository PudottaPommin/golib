package realip

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/pudottapommin/golib/http/extractors"
)

type (
	OptsFn     func(middleware *Middleware)
	Middleware struct {
		ipExtractor extractors.Extractor
		ips         map[string]struct{}
		ranges      []*net.IPNet

		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next           func(http.ResponseWriter, *http.Request) bool
		Logger         *slog.Logger
		TrustedProxies []string
		FromHeaders    []string
	}
)

func New(opts ...OptsFn) *Middleware {
	m := &Middleware{
		TrustedProxies: []string{"127.0.0.1/32", "::1/128"},
		FromHeaders:    []string{"True-Client-IP", "X-Forwarded-For", "X-Real-IP"},
	}
	for _, opt := range opts {
		opt(m)
	}

	for _, ipAddress := range m.TrustedProxies {
		if strings.ContainsRune(ipAddress, '/') {
			_, ipNet, err := net.ParseCIDR(ipAddress)
			if err != nil {
				// return error?
				continue
			}
			m.ranges = append(m.ranges, ipNet)
		} else {
			ip := net.ParseIP(ipAddress)
			if ip == nil {
				// return error?
				continue
			}
			m.ips[ipAddress] = struct{}{}
		}
	}

	chain := make([]extractors.Extractor, len(m.FromHeaders))
	for i, header := range m.FromHeaders {
		chain[i] = extractors.FromHeader(header)
	}
	m.ipExtractor = extractors.Chain(chain...)

	return m
}

func (mw *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.Next != nil && mw.Next(w, r) {
			next.ServeHTTP(w, r)
			return
		}
		ip := mw.getRealIP(r)
		if !strings.EqualFold(r.RemoteAddr, ip) {
			r.RemoteAddr = ip
		}
		next.ServeHTTP(w, r)
	})
}

func (mw *Middleware) getRealIP(r *http.Request) string {
	if mw.Logger != nil {
		mw.Logger.Debug("Getting real IP", "remoteAddr", r.RemoteAddr)
	}
	remAddr := r.RemoteAddr
	if split := strings.Split(remAddr, ":"); len(split) > 0 {
		remAddr = split[0]
	}
	remoteIP := net.ParseIP(remAddr)
	if !mw.isProxyTrusted(remoteIP) {
		return r.RemoteAddr
	}

	if ip, err := mw.ipExtractor.Extract(r); err == nil && ip != "" {
		if strings.ContainsRune(ip, ',') {
			if rip, _, found := strings.Cut(ip, ","); found {
				ip = strings.TrimSpace(rip)
			}
		}
		if ip != "" && net.ParseIP(ip) != nil {
			r.RemoteAddr = ip
		}
	}
	return r.RemoteAddr
}

func (mw *Middleware) isProxyTrusted(ip net.IP) bool {
	if _, trusted := mw.ips[ip.String()]; trusted {
		return true
	}

	for _, ipNet := range mw.ranges {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
