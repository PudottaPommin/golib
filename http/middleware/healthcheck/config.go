package healthcheck

import (
	"net/http"
)

type (
	OptsFn func(*mw)
	mw     struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next  func(http.ResponseWriter, *http.Request) bool
		Probe HealthCheck
	}
)

const (
	DefaultHealthzEndpoint = "/healthz"
)

func New(opts ...OptsFn) *mw {
	m := &mw{Probe: defaultProbe}
	for i := range opts {
		opts[i](m)
	}
	return m
}

func defaultProbe(_ http.ResponseWriter, _ *http.Request) bool { return true }

func WithProbe(probe HealthCheck) OptsFn {
	return func(c *mw) {
		c.Probe = probe
	}
}
