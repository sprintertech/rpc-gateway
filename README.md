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

Additionally, to load configuration from an environment variable, use the `--env` flag. Ensure the `GATEWAY_CONFIG` environment variable is set with the main configuration data.

```console
DEBUG=true go run . --env
```

## Configuration

The main configuration has been updated to use JSON format (`config.json`). It specifies the metrics server port and multiple gateways, each with its own JSON configuration file:

```json
{
  "metrics": {
    "port": 9090
  },
  "port": 4000,
  "gateways": [
    {
      "configFile": "config_holesky.json",
      "name": "Holesky gateway"
    },
    {
      "configFile": "config_sepolia.json",
      "name": "Sepolia gateway"
    }
  ]
}
```

Each JSON configuration file for the gateways can specify detailed settings for proxy behavior, health checks, and target node providers. Here is an example of what these individual gateway configuration files might contain:

```json
{
  "proxy": {
    "port": "3000",
    "upstreamTimeout": "1s"
  },
  "healthChecks": {
    "interval": "5s",
    "timeout": "1s",
    "failureThreshold": 2,
    "successThreshold": 1
  },
  "targets": [
    {
      "name": "Cloudflare",
      "connection": {
        "http": {
          "url": "https://cloudflare-eth.com"
        }
      }
    },
    {
      "name": "Alchemy",
      "connection": {
        "http": {
          "url": "https://alchemy.com/rpc/<apikey>"
        }
      }
    }
  ]
}
```

## Authentication
Basic authentication has been added to the RPC Gateway. You need to provide a username and password for access. These can be configured using CLI flags:

- `--username`: The username for basic authentication.
- `--password`: The password for basic authentication.

**The password must be provided; otherwise, the application will throw an error.**

### Running the Application
To run the application with authentication:

```
DEBUG=true go run . --config config.json --username myuser --password mypass
```
To use configuration from an environment variable:

```
DEBUG=true go run . --env --username myuser --password mypass
```