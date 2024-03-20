package rpcgateway

import (
	"github.com/sygmaprotocol/rpc-gateway/internal/metrics"
	"github.com/sygmaprotocol/rpc-gateway/internal/proxy"
)

type RPCGatewayConfig struct { //nolint:revive
	Name         string                     `json:"name"`
	Metrics      metrics.Config             `json:"metrics"`
	Proxy        proxy.ProxyConfig          `json:"proxy"`
	HealthChecks proxy.HealthCheckConfig    `json:"healthChecks"`
	Targets      []proxy.NodeProviderConfig `json:"targets"`
}
