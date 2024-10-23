package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"

	"github.com/sygmaprotocol/rpc-gateway/internal/auth"
	"github.com/sygmaprotocol/rpc-gateway/internal/metrics"
	"github.com/sygmaprotocol/rpc-gateway/internal/util"

	"github.com/pkg/errors"
	"github.com/sygmaprotocol/rpc-gateway/internal/rpcgateway"
	"github.com/urfave/cli/v2"
)

// Config represents the application configuration structure,
// including metrics and gateway configurations.
type Config struct {
	Metrics  MetricsConfig   `json:"metrics"`
	Port     uint            `json:"port"`
	Gateways []GatewayConfig `json:"gateways"`
}

type MetricsConfig struct {
	Port uint `json:"port"`
}

type GatewayConfig struct {
	ConfigFile string `json:"configFile"`
	Name       string `json:"name"`
}

func main() {
	c, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println(strings.Join(os.Args, ","))

	app := &cli.App{
		Name:  "rpc-gateway",
		Usage: "The failover proxy for node providers.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "The JSON configuration file path with gateway configurations.",
				Value: "config.json", // Default configuration file name
			},
			&cli.BoolFlag{
				Name:  "env",
				Usage: "Load configuration from environment variable named GATEWAY_CONFIG.",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "auth",
				Usage: "Enable basic authentication.",
				Value: false,
			},
		},
		Action: func(cc *cli.Context) error {
			configPath := resolveConfigPath(cc.String("config"), cc.Bool("env"))
			config, err := util.LoadJSONFile[Config](configPath)
			if err != nil {
				return errors.Wrap(err, "failed to load config")
			}

			logger := configureLogger()
			startMetricsServer(config.Metrics.Port)

			r := chi.NewRouter()
			r.Use(httplog.RequestLogger(logger))
			r.Use(middleware.Recoverer)
			r.Use(middleware.Heartbeat("/health"))
			// Add basic auth middleware
			if cc.Bool("auth") {
				tokenMapJSON := os.Getenv("GATEWAY_TOKEN_MAP")
				if tokenMapJSON == "" {
					return errors.New("GATEWAY_TOKEN_MAP environment variable must be set for authentication")
				}

				var tokenMap map[string]auth.TokenInfo

				if err := json.Unmarshal([]byte(tokenMapJSON), &tokenMap); err != nil {
					return errors.Wrap(err, "failed to parse GATEWAY_TOKEN_MAP")
				}

				for _, details := range tokenMap {
					if details.NumOfRequestPerSec <= 0 {
						return errors.New("numOfRequestPerSec must be a positive number")
					}
				}

				r.Use(auth.URLTokenAuth(tokenMap))
				fmt.Println("Authentication configured on gateway")
			}

			server := &http.Server{
				Addr:              fmt.Sprintf(":%d", config.Port),
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

			fmt.Printf("Starting RPC Gateway server on port: %d\n", config.Port)
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

func resolveConfigPath(config string, isENV bool) string {
	if isENV {
		return "GATEWAY_CONFIG"
	}

	return config
}

func configureLogger() *httplog.Logger {
	logLevel := slog.LevelWarn
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}

	return httplog.NewLogger("rpc-gateway", httplog.Options{
		JSON:           true,
		RequestHeaders: true,
		LogLevel:       logLevel,
	})
}

func startMetricsServer(port uint) {
	metricsServer := metrics.NewServer(metrics.Config{Port: port})
	go func() {
		err := metricsServer.Start()
		defer func(metricsServer *metrics.Server) {
			err := metricsServer.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error stopping metrics server: %v\n", err)
			}
		}(metricsServer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error starting metrics server: %v\n", err)
		}
	}()
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
