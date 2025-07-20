# Pandacea Agent Backend

A Go-based agent backend for the Pandacea Protocol that provides P2P discovery and REST API endpoints for data product management.

## Features

- **P2P Discovery**: libp2p-based peer-to-peer networking with KAD-DHT
- **REST API**: HTTP endpoints for product discovery and lease management
- **Policy Engine**: Request evaluation based on Guiding Principles
- **Structured Logging**: Privacy-compliant logging without PII
- **Graceful Shutdown**: Clean shutdown handling for SIGINT/SIGTERM
- **Configuration Management**: YAML config files with environment variable overrides

## Project Structure

```
agent-backend/
├── cmd/
│   └── agent/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── server.go            # HTTP API server
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── p2p/
│   │   └── node.go              # P2P node implementation
│   └── policy/
│       └── policy.go            # Policy evaluation engine
├── config.yaml                  # Sample configuration
├── go.mod                       # Go module file
└── README.md                    # This file
```

## API Endpoints

### GET /api/v1/products
Returns a list of available data products.

**Response:**
```json
{
  "data": [
    {
      "productId": "did:pandacea:earner:123/abc-456",
      "name": "Novel Package 3D Scans - Warehouse A",
      "dataType": "RoboticSensorData",
      "keywords": ["robotics", "3d-scan", "lidar"]
    }
  ],
  "nextCursor": "cursor_def456"
}
```

### POST /api/v1/leases
Creates a new lease request with strict input validation.

**Request:**
```json
{
  "productId": "did:pandacea:earner:123/abc-456",
  "maxPrice": "0.01",
  "duration": "24h"
}
```

**Response:**
```json
{
  "leaseProposalId": "lease_prop_123456789"
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

## Configuration

The agent can be configured via a YAML file and environment variables.

### Configuration File (config.yaml)
```yaml
server:
  port: 8080  # HTTP server port

p2p:
  listen_port: 0  # 0 means let libp2p choose a random port
```

### Environment Variables
- `HTTP_PORT`: Override HTTP server port
- `P2P_PORT`: Override P2P listen port

## Installation & Usage

### Prerequisites
- Go 1.21 or later

### Build
```bash
go build -o agent ./cmd/agent
```

### Run
```bash
# Run with default configuration
./agent

# Run with custom config file
./agent -config config.yaml

# Run with environment variables
HTTP_PORT=9090 P2P_PORT=4001 ./agent
```

### Development
```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run with hot reload (if using air)
air
```

## P2P Discovery

The agent automatically:
- Generates a unique peer ID on startup
- Joins the KAD-DHT network for discovery
- Logs its peer ID for external connections
- Handles peer discovery via mDNS

### Peer ID Logging
On startup, the agent logs its peer ID:
```
{"level":"INFO","msg":"P2P node initialized","peer_id":"QmXxxx...","listen_addrs":["/ip4/127.0.0.1/tcp/4001"]}
```

## Policy Engine

The policy engine evaluates lease requests according to the Pandacea Protocol's Guiding Principles. Currently implemented as a placeholder that accepts all requests.

### Future Policy Features
- Dynamic Minimum Pricing (DMP) validation
- Reputation system integration
- Data product availability checks
- Rate limiting and abuse prevention
- Compliance with regulatory requirements

## Input Validation

The lease endpoint performs strict schema-based validation:

- **productId**: Must conform to `did:pandacea` format
- **maxPrice**: Must be a valid decimal number
- **duration**: Must be in format `<number>[d|h|m|s]`

## Logging

The agent uses structured JSON logging that complies with privacy requirements:

- No PII or sensitive data in logs
- Event-based logging with metadata
- Request IDs for tracing
- Structured error reporting

### Log Levels
- `INFO`: Normal operation events
- `ERROR`: Error conditions
- `DEBUG`: Detailed debugging information (when enabled)

## Security & Privacy

- **No PII Logging**: Request bodies are not logged
- **Structured Events**: Logs contain event types and metadata only
- **Input Validation**: Strict schema validation for all inputs
- **Policy Enforcement**: All requests evaluated by policy engine

## Integration

### Contract Integration
The agent is designed to integrate with deployed smart contracts. Contract addresses should be read from `/config/deployments.json`.

### SDK Integration
The Python Builder SDK can discover and connect to this agent using the logged peer ID.

## Development

### Adding New Endpoints
1. Add route in `internal/api/server.go`
2. Implement handler function
3. Add validation if needed
4. Update tests

### Extending Policy Engine
1. Modify `internal/policy/policy.go`
2. Add new evaluation rules
3. Update request/response structures
4. Add tests

### Configuration Changes
1. Update `internal/config/config.go`
2. Add new fields to Config struct
3. Update YAML parsing
4. Add environment variable support

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/api
```

## Deployment

### Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o agent ./cmd/agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent .
COPY --from=builder /app/config.yaml .
CMD ["./agent"]
```

### Environment Variables for Production
- `HTTP_PORT`: Production HTTP port
- `P2P_PORT`: Production P2P port
- `LOG_LEVEL`: Logging level (INFO, DEBUG, ERROR)

## Troubleshooting

### Common Issues

1. **Port Already in Use**
   - Change HTTP_PORT or P2P_PORT
   - Check for other running instances

2. **P2P Connection Issues**
   - Ensure firewall allows P2P port
   - Check network connectivity

3. **Configuration Errors**
   - Validate YAML syntax
   - Check environment variable names

### Debug Mode
```bash
LOG_LEVEL=DEBUG ./agent
```

## Contributing

1. Follow Go coding standards
2. Add tests for new features
3. Update documentation
4. Ensure privacy compliance in logging 