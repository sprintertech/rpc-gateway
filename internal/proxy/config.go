package proxy

import (
	"github.com/sygmaprotocol/rpc-gateway/internal/util"
)

type HealthCheckConfig struct {
	Interval         util.DurationUnmarshalled `json:"interval"`
	Timeout          util.DurationUnmarshalled `json:"timeout"`
	FailureThreshold uint                      `json:"failureThreshold"`
	SuccessThreshold uint                      `json:"successThreshold"`
}

type ProxyConfig struct { // nolint:revive
	Path            string                    `json:"path"`
	UpstreamTimeout util.DurationUnmarshalled `json:"upstreamTimeout"`
}

// This struct is temporary. It's about to keep the input interface clean and simple.
type Config struct {
	Proxy              ProxyConfig
	Targets            []NodeProviderConfig
	HealthChecks       HealthCheckConfig
	HealthcheckManager *HealthCheckManager
	Name               string
}
