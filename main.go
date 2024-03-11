package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0xProject/rpc-gateway/internal/metrics"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/0xProject/rpc-gateway/internal/rpcgateway"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Gateways []GatewayConfig `json:"gateways"`
}

type GatewayConfig struct {
	ConfigFile string `json:"config-file"`
	Name       string `json:"name"`
}

func main() {
	c, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := &cli.App{
		Name:  "rpc-gateway",
		Usage: "The failover proxy for node providers.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "The JSON configuration file path with gateway configurations.",
				DefaultText: "config.json",
			},
		},
		Action: func(cc *cli.Context) error {
			// configPath := cc.String("config")
			config, err := loadConfig("./config.json")
			if err != nil {
				return errors.Wrap(err, "failed to load config")
			}

			// Instantiate the metrics server based on the config before creating the RPCGateway instance.
			metricsServer := metrics.NewServer(metrics.Config{Port: 9000})

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

func loadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func startGateway(ctx context.Context, config GatewayConfig, server *metrics.Server) error {
	service, err := rpcgateway.NewRPCGatewayFromConfigFile(config.ConfigFile, server)
	if err != nil {
		return errors.Wrap(err, "rpc-gateway failed")
	}

	err = service.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot start service")
	}

	<-ctx.Done()
	return errors.Wrap(service.Stop(ctx), "cannot stop service")
}
