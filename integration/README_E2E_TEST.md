# End-to-End Integration Test: Asynchronous Lease Approval Flow

This document describes the comprehensive end-to-end integration test that verifies the entire asynchronous lease approval flow across all components of the Pandacea protocol.

## Overview

The `test_full_asynchronous_lease_flow` test is the final integration test that ties together all components of the protocol:

1. **Python Builder SDK** - For API interactions and blockchain transactions
2. **Go Agent Backend** - For lease proposal creation and status tracking
3. **Smart Contract** - For on-chain lease creation and event emission
4. **Blockchain Node** - For transaction processing and event broadcasting

## Test Flow

The test simulates the complete user journey for creating a lease:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Step 1:       │    │   Step 2:       │    │   Step 3:       │
│   Discover      │───▶│   Create        │───▶│   Execute       │
│   Products      │    │   Proposal      │    │   On-Chain      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Agent API     │    │   Agent API     │    │   Smart         │
│   /products     │    │   /leases       │    │   Contract      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │   Step 4:       │
                                              │   Poll Status   │
                                              └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │   Agent API     │
                                              │   /leases/{id}  │
                                              └─────────────────┘
```

### Step-by-Step Breakdown

#### Step 1: Product Discovery
- **Action**: Call `client.discover_products()` via SDK
- **Target**: Agent API endpoint `/api/v1/products`
- **Purpose**: Get a valid product ID to lease
- **Validation**: Ensure at least one product is available

#### Step 2: Lease Proposal Creation
- **Action**: Call `client.request_lease()` via SDK
- **Target**: Agent API endpoint `/api/v1/leases` (POST)
- **Purpose**: Create an off-chain lease proposal
- **Validation**: Receive a valid `leaseProposalId`
- **Agent State**: Creates initial lease state with status "pending"

#### Step 3: On-Chain Execution
- **Action**: Call `client.execute_lease_on_chain()` via SDK
- **Target**: Smart contract `createLease` function
- **Purpose**: Execute the lease on the blockchain
- **Validation**: Transaction hash returned successfully
- **Event**: Smart contract emits `LeaseCreated` event

#### Step 4: Status Polling
- **Action**: Poll agent status endpoint
- **Target**: Agent API endpoint `/api/v1/leases/{leaseProposalId}` (GET)
- **Purpose**: Wait for agent to detect blockchain event
- **Validation**: Status changes from "pending" to "approved"
- **Agent State**: Event listener updates lease state

## Prerequisites

### Environment Setup

1. **Anvil Node**: Local blockchain node running
   ```bash
   anvil
   ```

2. **Smart Contract**: LeaseAgreement contract deployed
   ```bash
   # Deploy using Foundry
   forge script Deploy --rpc-url http://127.0.0.1:8545 --broadcast
   ```

3. **Agent Backend**: Go agent running with blockchain configuration
   ```bash
   export CONTRACT_ADDRESS="0x1234567890123456789012345678901234567890"
   export RPC_URL="http://127.0.0.1:8545"
   go run cmd/agent/main.go
   ```

### Environment Variables

```bash
# Required for blockchain interaction
export CONTRACT_ADDRESS="0x1234567890123456789012345678901234567890"
export SPENDER_PRIVATE_KEY="0x1234567890123456789012345678901234567890123456789012345678901234"

# Optional (defaults provided)
export RPC_URL="http://127.0.0.1:8545"
export AGENT_API_URL="http://localhost:8080"
```

## Running the Test

### Option 1: Run with Pytest

```bash
cd integration
python -m pytest test_onchain_interaction.py::TestOnChainInteraction::test_full_asynchronous_lease_flow -v
```

### Option 2: Run All Integration Tests

```bash
cd integration
python -m pytest test_onchain_interaction.py -v
```

### Option 3: Run with Manual Test Script

```bash
cd integration
python test_onchain_manual.py
```

## Expected Output

When successful, you should see output similar to:

```
E2E Test: Verifying full asynchronous lease state machine...
 Found product to lease: Novel Package 3D Scans - Warehouse A
Requesting off-chain lease proposal from agent...
 Received leaseProposalId: lease_prop_1703123456789012345
Executing the lease on-chain...
 On-chain transaction sent with hash: 0xabc123...
Polling agent for status update on lease: lease_prop_1703123456789012345
✅ Lease status successfully updated to 'approved'!
✅ E2E test passed: Full asynchronous lease flow verified successfully!
```

## Test Configuration

### Timeout Settings

- **Polling Timeout**: 30 seconds (configurable in test)
- **Polling Interval**: 2 seconds between requests
- **Transaction Timeout**: Handled by Web3.py defaults

### Test Parameters

- **Product**: First available product from discovery
- **Max Price**: 0.01 ETH (10,000,000,000,000,000 wei)
- **Payment**: 0.001 ETH (1,000,000,000,000,000 wei)
- **Duration**: 24 hours
- **Earner Address**: Dummy address for testing

## Error Handling

### Common Failure Scenarios

1. **Agent Not Running**
   - **Symptom**: Connection refused on API calls
   - **Solution**: Start the Go agent backend

2. **Anvil Not Running**
   - **Symptom**: Connection refused on blockchain calls
   - **Solution**: Start Anvil node

3. **Contract Not Deployed**
   - **Symptom**: Invalid contract address errors
   - **Solution**: Deploy contract and set CONTRACT_ADDRESS

4. **Insufficient ETH**
   - **Symptom**: Transaction fails with "insufficient funds"
   - **Solution**: Fund the spender account with ETH

5. **Event Listener Not Working**
   - **Symptom**: Status never changes to "approved"
   - **Solution**: Check agent logs for event listener errors

### Debugging Tips

1. **Check Agent Logs**: Look for event listener messages
2. **Check Anvil Logs**: Verify transaction processing
3. **Verify Environment**: Ensure all env vars are set correctly
4. **Test Components Individually**: Run unit tests first

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: E2E Integration Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
          
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Install Foundry
        run: |
          curl -L https://foundry.paradigm.xyz | bash
          foundryup
          
      - name: Start Anvil
        run: anvil &
        
      - name: Deploy Contract
        run: |
          forge script Deploy --rpc-url http://127.0.0.1:8545 --broadcast
          # Extract contract address and set env var
          
      - name: Start Agent
        run: |
          export CONTRACT_ADDRESS=${{ env.CONTRACT_ADDRESS }}
          export RPC_URL="http://127.0.0.1:8545"
          go run cmd/agent/main.go &
          
      - name: Run E2E Test
        run: |
          cd integration
          pip install -r requirements.txt
          python -m pytest test_onchain_interaction.py::TestOnChainInteraction::test_full_asynchronous_lease_flow -v
```

## Performance Considerations

### Test Duration

- **Typical Runtime**: 30-60 seconds
- **Polling Overhead**: ~15 seconds (30 seconds timeout, 2-second intervals)
- **Blockchain Confirmation**: ~2-5 seconds
- **Agent Processing**: ~1-2 seconds

### Resource Usage

- **Memory**: Minimal (in-memory state)
- **CPU**: Low (polling and event processing)
- **Network**: Moderate (API calls and blockchain interaction)

## Future Enhancements

1. **Multiple Lease Types**: Test different lease configurations
2. **Concurrent Leases**: Test multiple simultaneous lease operations
3. **Failure Scenarios**: Test rollback and error recovery
4. **Performance Testing**: Measure latency and throughput
5. **Load Testing**: Test with high transaction volumes

## Security Considerations

1. **Private Key Management**: Test uses environment variables
2. **Network Isolation**: Test runs on local Anvil node
3. **API Authentication**: All requests require valid signatures
4. **Transaction Validation**: Smart contract enforces business rules

This E2E test provides a critical regression guard for the entire protocol and ensures all components work together correctly in the asynchronous lease approval flow. 