package healthcheck

import (
	"github.com/gin-gonic/gin"
)

type (
	OptsFn func(*Config)
	Config struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next  func(c *gin.Context) bool
		Probe HealthCheck
	}
)

const (
	DefaultHealthzEndpoint = "/healthz"
)

var defaultConfig = Config{
	Next:  nil,
	Probe: defaultProbe,
}

func defaultProbe(_ *gin.Context) bool { return true }

func WithProbe(probe HealthCheck) OptsFn {
	return func(c *Config) {
		c.Probe = probe
	}
}
