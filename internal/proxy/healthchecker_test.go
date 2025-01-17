package proxy

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sygmaprotocol/rpc-gateway/internal/util"

	"github.com/caitlinelfring/go-env-default"
	"github.com/stretchr/testify/assert"
)

// TestBasicHealthchecker checks if it runs with default options.
func TestBasicHealthchecker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healtcheckConfig := HealthCheckerConfig{
		URL:              env.GetDefault("RPC_GATEWAY_NODE_URL_1", "https://lodestar-holeskyrpc.chainsafe.io/"),
		Interval:         util.DurationUnmarshalled(2 * time.Second),
		Timeout:          util.DurationUnmarshalled(3 * time.Second),
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Logger:           slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}

	healthchecker, err := NewHealthChecker(healtcheckConfig, "")
	assert.NoError(t, err)

	healthchecker.Start(ctx)

	assert.NotZero(t, healthchecker.BlockNumber())

	// TODO: can be flaky due to cloudflare-eth endpoint
	assert.True(t, healthchecker.IsHealthy())

	healthchecker.isHealthy = false
	assert.False(t, healthchecker.IsHealthy())

	healthchecker.isHealthy = true
	assert.True(t, healthchecker.IsHealthy())
}
