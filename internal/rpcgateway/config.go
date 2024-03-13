package rpcgateway

import (
	"github.com/sygmaprotocol/rpc-gateway/internal/metrics"
	"github.com/sygmaprotocol/rpc-gateway/internal/proxy"
)

type RPCGatewayConfig struct { //nolint:revive
	Name         string                     `yaml:"name"`
	Metrics      metrics.Config             `yaml:"metrics"`
	Proxy        proxy.ProxyConfig          `yaml:"proxy"`
	HealthChecks proxy.HealthCheckConfig    `yaml:"healthChecks"`
	Targets      []proxy.NodeProviderConfig `yaml:"targets"`
}
