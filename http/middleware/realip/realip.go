package realip

import (
	"net"
	"net/http"
	"strings"
)

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
