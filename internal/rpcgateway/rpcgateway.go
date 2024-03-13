package rpcgateway

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sygmaprotocol/rpc-gateway/internal/util"

	"github.com/carlmjohnson/flowmatic"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/sygmaprotocol/rpc-gateway/internal/proxy"
)

type RPCGateway struct {
	config RPCGatewayConfig
	proxy  *proxy.Proxy
	hcm    *proxy.HealthCheckManager
}

func (r *RPCGateway) Start(c context.Context) error {
	return flowmatic.Do(
		func() error {
			return errors.Wrap(r.hcm.Start(c), "failed to start health check manager")
		},
	)
}

func (r *RPCGateway) Stop(c context.Context) error {
	return flowmatic.Do(
		func() error {
			return errors.Wrap(r.hcm.Stop(c), "failed to stop health check manager")
		},
	)
}

func NewRPCGateway(config RPCGatewayConfig, router *chi.Mux) (*RPCGateway, error) {
	logLevel := slog.LevelWarn
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}

	hcm, err := proxy.NewHealthCheckManager(
		proxy.HealthCheckManagerConfig{
			Targets: config.Targets,
			Config:  config.HealthChecks,
			Logger: slog.New(
				slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
					Level: logLevel,
				})),
		}, config.Name)
	if err != nil {
		return nil, errors.Wrap(err, "healthcheckmanager failed")
	}

	proxy, err := proxy.NewProxy(
		proxy.Config{
			Proxy:              config.Proxy,
			Targets:            config.Targets,
			HealthChecks:       config.HealthChecks,
			HealthcheckManager: hcm,
			Name:               config.Name,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "proxy failed")
	}

	router.Handle(fmt.Sprintf("/%s", config.Proxy.Path), proxy)

	return &RPCGateway{
		config: config,
		proxy:  proxy,
		hcm:    hcm,
	}, nil
}

// NewRPCGatewayFromConfigFile creates an instance of RPCGateway from provided
// configuration file.
func NewRPCGatewayFromConfigFile(fileOrUrl string, router *chi.Mux) (*RPCGateway, error) {
	config, err := util.LoadYamlFile[RPCGatewayConfig](fileOrUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	fmt.Println("Starting RPC Gateway for " + config.Name + " on path: /" + config.Proxy.Path)

	// Pass the metrics router as an argument to NewRPCGateway.
	return NewRPCGateway(*config, router)
}
