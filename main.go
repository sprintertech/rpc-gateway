package main

import (
	"context"
	"fmt"
	"github.com/sygmaprotocol/rpc-gateway/internal/metrics"
	"github.com/sygmaprotocol/rpc-gateway/internal/util"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sygmaprotocol/rpc-gateway/internal/rpcgateway"
	"github.com/urfave/cli/v2"
)

type MetricsConfig struct {
	Port int `yaml:"port"`
}

type Config struct {
	Metrics  MetricsConfig   `yaml:"metrics"`
	Gateways []GatewayConfig `yaml:"gateways"`
}

type GatewayConfig struct {
	ConfigFile string `yaml:"configFile"`
	Name       string `yaml:"name"`
}

func main() {
	c, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := &cli.App{
		Name:  "rpc-gateway",
		Usage: "The failover proxy for node providers.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "The YAML configuration file path with gateway configurations.",
				Value: "config.yml", // Default configuration file name
			},
		},
		Action: func(cc *cli.Context) error {
			configPath := cc.String("config")
			config, err := util.LoadYamlFile[Config](configPath)
			if err != nil {
				return errors.Wrap(err, "failed to load config")
			}

			metricsServer := metrics.NewServer(metrics.Config{Port: uint(config.Metrics.Port)})

			var wg sync.WaitGroup
			for _, gatewayConfig := range config.Gateways {
				wg.Add(1)
				go func(gwConfig GatewayConfig) {
					defer wg.Done()
					err := startGateway(c, gwConfig, metricsServer)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error starting gateway '%s': %v\n", gwConfig.Name, err)
					}
				}(gatewayConfig)
			}

			wg.Wait()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}

func startGateway(ctx context.Context, config GatewayConfig, server *metrics.Server) error {
	service, err := rpcgateway.NewRPCGatewayFromConfigFile(config.ConfigFile, server)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%s rpc-gateway failed", config.Name))
	}

	err = service.Start(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot start %s rpc-gateway", config.Name))
	}

	<-ctx.Done()

	return errors.Wrap(service.Stop(ctx), fmt.Sprintf("cannot stop %s rpc-gateway", config.Name))
}
