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
		ips:            make(map[string]struct{}),
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

func WithNext(next func(http.ResponseWriter, *http.Request) bool) OptsFn {
	return func(m *Middleware) {
		m.Next = next
	}
}

func WithLogger(logger *slog.Logger) OptsFn {
	return func(m *Middleware) {
		m.Logger = logger
	}
}

func WithTrustedProxies(proxies []string) OptsFn {
	return func(m *Middleware) {
		m.TrustedProxies = proxies
	}
}

func WithFromHeaders(headers []string) OptsFn {
	return func(m *Middleware) {
		m.FromHeaders = headers
	}
}
