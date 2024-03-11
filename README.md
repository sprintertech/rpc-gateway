Based on the given details and the previous code modification instructions, here's how you could update the README for the rpc-gateway project:

---

# RPC Gateway

The rpc-gateway is a failover proxy designed for node providers. It ensures high availability and reliability by automatically rerouting requests to a backup node provider when health checks indicate the primary provider is down. This process ensures uninterrupted service even in the event of node provider failures.

## Caution

> :warning: The rpc-gateway is currently in development mode. It is not considered stable and should be used with caution in production environments.

## Overview

The rpc-gateway operates by continuously performing health checks on configured node providers. If the primary node provider fails these checks, the gateway will automatically attempt to route requests to the next available provider based on a predefined failover sequence.

```mermaid
sequenceDiagram
Alice->>RPC Gateway: eth_call
loop Healthcheck
    RPC Gateway->>Alchemy: Check health
    RPC Gateway->>Infura: Check health
end
Note right of RPC Gateway: Routes only to healthy targets
loop Configurable Retries
RPC Gateway->>Alchemy: eth_call?
Alchemy-->>RPC Gateway: ERROR
end
Note right of RPC Gateway: RPC Call is rerouted after failing retries
RPC Gateway->>Infura: eth_call?
Infura-->>RPC Gateway: {"result":[...]}
RPC Gateway-->>Alice: {"result":[...]}
```

## Development

To contribute to the development of rpc-gateway, ensure that you have Go installed and the project set up locally. Start by running tests to ensure everything is working as expected.

```console
go test -v ./...
```

For local development and testing, you can run the application with:

```console
DEBUG=true go run . --config config.json
```

The above command assumes you have a `config.json` file configured to start multiple gateways, each with its own `yml` configuration file as described previously.

## Configuration

The rpc-gateway is highly configurable to meet different operational requirements. Below is an example configuration (`config.json`) that specifies multiple gateways, each with its own `.yml` configuration file:

```json
{
  "gateways": [
    {
      "config-file": "config1.yml",
      "name": "Chain A gateway"
    },
    {
      "config-file": "config2.yml",
      "name": "Chain B gateway"
    }
    // Add more gateways as needed
  ]
}
```

Each `.yml` configuration file can specify detailed settings for metrics, proxy behavior, health checks, and target node providers. Here is an example `.yml` configuration:

```yaml
metrics:
  port: "9090" # Port for Prometheus metrics, served on /metrics and /

proxy:
  port: "3000" # Port for RPC gateway
  upstreamTimeout: "1s" # When is a request considered timed out

healthChecks:
  interval: "5s" # How often to perform health checks
  timeout: "1s" # Timeout duration for health checks
  failureThreshold: 2 # Failed checks until a target is marked unhealthy
  successThreshold: 1 # Successes required to mark a target healthy again

targets: # Failover order is determined by the list order
  - name: "Cloudflare"
    connection:
      http:
        url: "https://cloudflare-eth.com"
  - name: "Alchemy"
    connection:
      http:
        url: "https://alchemy.com/rpc/<apikey>"
```

This setup allows for a flexible and robust configuration, ensuring your RPC gateway can effectively manage multiple node providers and maintain service availability.

--- 
