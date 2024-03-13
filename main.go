package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"

	"github.com/sygmaprotocol/rpc-gateway/internal/metrics"
	"github.com/sygmaprotocol/rpc-gateway/internal/util"

	"github.com/pkg/errors"
	"github.com/sygmaprotocol/rpc-gateway/internal/rpcgateway"
	"github.com/urfave/cli/v2"
)

type MetricsConfig struct {
	Port int `yaml:"port"`
}

type Config struct {
	Metrics  MetricsConfig   `yaml:"metrics"`
	Port     string          `yaml:"port"`
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

			logLevel := slog.LevelWarn
			if os.Getenv("DEBUG") == "true" {
				logLevel = slog.LevelDebug
			}

			logger := httplog.NewLogger("rpc-gateway", httplog.Options{
				JSON:           true,
				RequestHeaders: true,
				LogLevel:       logLevel,
			})

			metricsServer := metrics.NewServer(metrics.Config{Port: uint(config.Metrics.Port)})
			go func() {
				err = metricsServer.Start()
				defer metricsServer.Stop()
				if err != nil {
					fmt.Fprintf(os.Stderr, "error starting metrics server: %v\n", err)
				}
			}()

			r := chi.NewRouter()
			r.Use(httplog.RequestLogger(logger))
			r.Use(middleware.Recoverer)
			server := &http.Server{
				Addr:              fmt.Sprintf(":%s", config.Port),
				Handler:           r,
				WriteTimeout:      time.Second * 15,
				ReadTimeout:       time.Second * 15,
				ReadHeaderTimeout: time.Second * 5,
			}
			defer server.Close()

			var wg sync.WaitGroup
			for _, gatewayConfig := range config.Gateways {
				wg.Add(1)
				go func(gwConfig GatewayConfig) {
					defer wg.Done()
					err := startGateway(c, gwConfig, r)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error starting gateway '%s': %v\n", gwConfig.Name, err)
					}
				}(gatewayConfig)
			}

			fmt.Println("Starting RPC Gateway server on port: " + config.Port)
			err = server.ListenAndServe()
			if err != nil {
				return err
			}

			wg.Wait()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}

func startGateway(ctx context.Context, config GatewayConfig, router *chi.Mux) error {
	service, err := rpcgateway.NewRPCGatewayFromConfigFile(config.ConfigFile, router)
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
