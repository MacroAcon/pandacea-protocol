# Pandacea Protocol Vertical Slice Demo

This demo showcases the complete end-to-end flow of the Pandacea Protocol:

1. **Federated Learning Job**: Posts a training job to the agent backend
2. **Differential Privacy**: Applies DP-SGD with configurable epsilon
3. **Blockchain Integration**: Mints a lease using the Python SDK
4. **Data Product**: Fetches the produced artifact with privacy guarantees

## Prerequisites

- **Foundry**: For deploying contracts and running Anvil
- **Python 3.8+**: For the SDK and demo script
- **Go 1.21+**: For building the agent backend

## Quick Start

### Option 1: Using the Makefile (Recommended)

```bash
# Run the complete demo
make demo
```

This will:
- Set up the environment and install dependencies
- Start Anvil blockchain on localhost:8545
- Deploy PGT and LeaseAgreement contracts
- Start the agent backend on localhost:8080
- Run the demo script
- Clean up all services

### Option 2: Manual Steps

1. **Setup Environment**:
   ```bash
   # Install Python dependencies
   cd builder-sdk && pip install -e .
   cd ../integration && pip install -r requirements.txt
   
   # Build agent backend
   cd ../agent-backend && make build
   ```

2. **Start Services**:
   ```bash
   # Start Anvil blockchain
   anvil --port 8545 &
   
   # Deploy contracts
   cd contracts
   PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
   forge script scripts/deploy.sol:DeployScript --rpc-url http://localhost:8545 --broadcast --legacy
   
   # Start agent backend
   cd ../agent-backend && ./agent &
   ```

3. **Configure Environment**:
   ```bash
   # Copy and edit environment file
   cp integration/.env.example integration/.env
   # Edit integration/.env with deployed contract addresses
   ```

4. **Run Demo**:
   ```bash
   cd integration
   python demo_vertical_slice.py
   ```

## Demo Flow

### 1. Federated Learning Job
The demo posts a training request:
```json
{
  "dataset": "toy_telemetry",
  "task": "logreg",
  "dp": {
    "enabled": true,
    "epsilon": 5.0
  }
}
```

### 2. Job Processing
The agent backend:
- Queues the job with a unique ID
- Runs a mock DP-SGD training process
- Produces an aggregate.json artifact with:
  - Model accuracy metrics
  - Privacy budget consumption (epsilon)
  - Training metadata

### 3. Blockchain Lease
The demo mints a lease on-chain:
- Uses local Anvil blockchain
- Deploys PGT tokens for payment
- Creates a LeaseAgreement contract
- Executes a lease transaction

### 4. Results
The demo outputs:
- Job completion status
- Privacy budget used (epsilon)
- Artifact file path
- Blockchain transaction hash
- Lease details

## Demo Output Example

```
🚀 Starting Pandacea Protocol Vertical Slice Demo
============================================================
🔧 Configuration loaded from environment
   Agent URL: http://localhost:8080
   RPC URL: http://127.0.0.1:8545
   Chain ID: 31337

📊 Posting federated learning job...
✅ Training job created with ID: job_1702123456789

⏳ Waiting for job job_1702123456789 to complete...
   Status: pending
   Status: running
   Status: complete
✅ Job completed successfully!

🔗 Minting lease on local blockchain...
   Creating lease for data product: ./data/products/job_1702123456789/aggregate.json
   Earner: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
   Payment: 0.001 ETH
✅ Lease created! Transaction hash: 0x1234567890abcdef...

============================================================
🎉 DEMO COMPLETED SUCCESSFULLY!
============================================================

📊 Federated Learning Job:
   Job ID: job_1702123456789
   Status: complete
   Epsilon Used: 5.0
   Artifact Path: ./data/products/job_1702123456789/aggregate.json

🔗 Blockchain Lease:
   Transaction Hash: 0x1234567890abcdef...
   Earner Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
   Payment Amount: 0.001 ETH
   Data Product ID: 746f795f74656c656d657472795f70726f647563745f3132333435363738393031323334

📁 Data Product Manifest:
   Path: ./data/products/job_1702123456789/aggregate.json
   Privacy Budget (ε): 5.0

✨ End-to-end demo completed successfully!
```

## Architecture

The demo demonstrates these key components:

- **Agent Backend** (Go): REST API for federated learning jobs
- **Python SDK**: Blockchain integration and client library
- **Smart Contracts**: PGT tokens and LeaseAgreement on Anvil
- **Privacy Engine**: Mock DP-SGD with configurable epsilon

## Configuration

Environment variables in `integration/.env`:

```bash
# Agent Backend
AGENT_URL=http://localhost:8080

# Blockchain
RPC_URL=http://127.0.0.1:8545
CHAIN_ID=31337

# Contracts (auto-populated by deploy script)
PGT_ADDRESS=0x5FbDB2315678afecb367f032d93F642f64180aa3
LEASE_ADDRESS=0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512

# Demo Timeouts
TIMEOUT_SECONDS=300
POLL_INTERVAL_SECONDS=2
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8545 and 8080 are available
2. **Missing dependencies**: Run `make demo-setup` to install all requirements
3. **Contract deployment fails**: Check Anvil is running and accessible
4. **Agent backend fails**: Ensure Go dependencies are installed

### Cleanup

```bash
# Stop all services and clean up
make demo-cleanup

# Force cleanup if processes are stuck
make demo-force-cleanup
```

### Logs

Check these files for debugging:
- `anvil.log`: Blockchain logs
- `agent.log`: Agent backend logs  
- `contracts/deploy.log`: Contract deployment logs

## Next Steps

This vertical slice demonstrates the core protocol flow. Future enhancements:

- Real PySyft integration for federated learning
- IPFS integration for decentralized storage
- Enhanced privacy guarantees and verification
- Production-ready error handling and monitoring
