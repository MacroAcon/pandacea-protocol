# Lease State Machine and On-Chain Event Listener

This document describes the lease state machine and on-chain event listener functionality implemented in the Go Agent Backend.

## Overview

The agent backend now includes a complete lease state management system that:

1. **Tracks lease proposals** in an in-memory state store
2. **Listens for blockchain events** from the LeaseAgreement smart contract
3. **Updates lease status** automatically when on-chain events are detected
4. **Exposes lease status** via a REST API endpoint for client polling

## Architecture

### Components

1. **LeaseProposalState**: In-memory data structure for tracking lease status
2. **Server.pendingLeases**: Thread-safe map storing lease proposal states
3. **startEventListener**: Goroutine that listens for blockchain events
4. **handleLeaseCreatedEvent**: Event handler that processes LeaseCreated events
5. **UpdateLeaseStatus**: Method for updating lease state (used by event handler)
6. **handleGetLeaseStatus**: API endpoint for querying lease status

### Data Flow

```
1. Client creates lease proposal → POST /api/v1/leases
2. Agent creates initial state → status: "pending"
3. Client executes lease on-chain → SDK calls execute_lease_on_chain()
4. Smart contract emits LeaseCreated event
5. Agent event listener detects event → updates state to "approved"
6. Client polls status → GET /api/v1/leases/{leaseProposalId}
7. Agent returns current state → status: "approved"
```

## Configuration

### Environment Variables

```bash
# Required for blockchain integration
export RPC_URL="http://127.0.0.1:8545"                    # Anvil RPC URL
export CONTRACT_ADDRESS="0x1234..."                       # Deployed contract address

# Optional (defaults provided)
export HTTP_PORT=8080                                     # Agent HTTP port
export P2P_PORT=0                                         # P2P port (0 = random)
```

### Configuration File (config.yaml)

```yaml
server:
  port: 8080
  min_price: "0.001"
  # ... other server config

p2p:
  listen_port: 0
  key_file_path: ""

blockchain:
  rpc_url: "http://127.0.0.1:8545"
  contract_address: "0x1234..."  # Must be set
```

## API Endpoints

### Create Lease Proposal

```http
POST /api/v1/leases
Content-Type: application/json
X-Pandacea-Signature: <base64-signature>
X-Pandacea-Peer-ID: <peer-id>

{
  "productId": "did:pandacea:earner:123/abc-456",
  "maxPrice": "0.01",
  "duration": "24h"
}
```

**Response:**
```json
{
  "leaseProposalId": "lease_prop_1703123456789012345"
}
```

### Get Lease Status

```http
GET /api/v1/leases/{leaseProposalId}
X-Pandacea-Signature: <base64-signature>
X-Pandacea-Peer-ID: <peer-id>
```

**Response:**
```json
{
  "status": "approved",
  "createdAt": "2023-12-21T10:30:45.123Z",
  "updatedAt": "2023-12-21T10:31:15.456Z",
  "leaseId": 1,
  "spenderAddr": "0x1234567890123456789012345678901234567890",
  "earnerAddr": "0x0987654321098765432109876543210987654321",
  "price": "1000000000000000000"
}
```

## Lease States

### Status Values

- **"pending"**: Lease proposal created, waiting for on-chain execution
- **"approved"**: Lease created on-chain, LeaseCreated event detected
- **"executed"**: Lease executed (future state)
- **"disputed"**: Lease disputed (future state)

### State Transitions

```
pending → approved (when LeaseCreated event detected)
approved → executed (when LeaseExecuted event detected)
approved → disputed (when LeaseDisputed event detected)
```

## Event Listener

### LeaseCreated Event

The event listener subscribes to `LeaseCreated` events from the smart contract:

```solidity
event LeaseCreated(
    bytes32 indexed leaseId,
    address indexed spender,
    address indexed earner,
    uint256 price
);
```

### Event Processing

When a `LeaseCreated` event is detected:

1. **Log event details** for debugging
2. **Generate lease proposal ID** from lease ID
3. **Update lease state** with event data:
   - Status: "approved"
   - Lease ID: from event
   - Spender address: from event
   - Earner address: from event
   - Price: from event
   - Timestamps: updated

### Error Handling

- **Connection failures**: Logged as warnings, agent continues without blockchain
- **Subscription errors**: Logged and event listener restarts
- **Event processing errors**: Logged but don't crash the listener

## Usage Examples

### Starting the Agent

```bash
# Set environment variables
export CONTRACT_ADDRESS="0x1234567890123456789012345678901234567890"
export RPC_URL="http://127.0.0.1:8545"

# Start the agent
go run cmd/agent/main.go
```

### Client Integration

```python
# Using the Python SDK
from pandacea_sdk import PandaceaClient

# Create client
client = PandaceaClient("http://localhost:8080")

# Create lease proposal
response = client.create_lease(
    product_id="did:pandacea:earner:123/abc-456",
    max_price="0.01",
    duration="24h"
)

lease_proposal_id = response.lease_proposal_id

# Execute lease on-chain
tx_hash = client.execute_lease_on_chain(
    earner="0x2222222222222222222222222222222222222222",
    data_product_id=b'test-data-product-id'.ljust(32, b'\0'),
    max_price=10000000000000000,  # 0.01 ETH in wei
    payment_in_wei=1000000000000000  # 0.001 ETH in wei
)

# Poll for status updates
import time
while True:
    status = client.get_lease_status(lease_proposal_id)
    if status.status == "approved":
        print(f"Lease approved! Lease ID: {status.lease_id}")
        break
    time.sleep(1)
```

## Testing

### Unit Tests

```bash
# Run all API tests
go test ./internal/api -v

# Run specific test
go test ./internal/api -v -run TestServer_handleGetLeaseStatus
```

### Integration Testing

1. **Start Anvil node**: `anvil`
2. **Deploy contract**: Use Foundry to deploy LeaseAgreement
3. **Set environment variables**: CONTRACT_ADDRESS and RPC_URL
4. **Start agent**: `go run cmd/agent/main.go`
5. **Run SDK tests**: Use the integration tests from the Python SDK

## Security Considerations

1. **Signature verification**: All API endpoints require valid signatures
2. **Environment variables**: Sensitive data (contract addresses) via env vars
3. **Thread safety**: All state updates use proper mutex locking
4. **Error handling**: Graceful degradation when blockchain is unavailable

## Future Enhancements

1. **Database persistence**: Replace in-memory storage with database
2. **Event replay**: Handle missed events on startup
3. **Multiple event types**: Support LeaseExecuted, LeaseDisputed events
4. **Webhook notifications**: Push status updates to clients
5. **Metrics and monitoring**: Add Prometheus metrics for event processing 