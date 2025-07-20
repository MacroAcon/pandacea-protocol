# Pandacea Protocol MVP

A revolutionary protocol for creating a human-centered data economy where individuals own and monetize their data while maintaining privacy and control.

## Overview

The Pandacea Protocol MVP demonstrates the core functionality of a decentralized data marketplace where:
- **Earners** (data owners) can securely monetize their data
- **Spenders** (data consumers) can discover and lease data products
- **Smart contracts** ensure fair pricing and secure transactions
- **P2P networking** enables decentralized discovery and communication

## Core Components

### 🏗️ Smart Contracts (`/contracts`)
- **LeaseAgreement.sol**: Core smart contract implementing Dynamic Minimum Pricing (DMP)
- **Foundry Testing**: Comprehensive unit tests with 100% coverage
- **Polygon Mumbai Deployment**: Ready for testnet deployment

### 🤖 Agent Backend (`/agent-backend`)
- **Go-based P2P Node**: libp2p networking with KAD-DHT discovery
- **HTTP API Server**: RESTful endpoints for data product discovery
- **Policy Engine**: Dynamic Minimum Pricing validation with decimal precision
- **Structured Logging**: Privacy-preserving event logging

### 🛠️ Builder SDK (`/builder-sdk`)
- **Python SDK**: Professional client library for developers
- **Pydantic Models**: Type-safe data validation
- **Comprehensive Testing**: Unit tests with mocking and error scenarios
- **Developer-Friendly**: Clear error messages and context manager support

## Visual Architecture

![Pandacea Protocol Logical Architecture](https://storage.googleapis.com/pandacea-public-assets/logical-architecture-diagram.png)

*The Pandacea Protocol enables direct peer-to-peer data transactions while maintaining privacy and ensuring fair pricing through smart contracts.*

## 🚀 Running the MVP Demo

Follow these simple steps to run the complete Pandacea Protocol MVP:

### 1. Clone the Repository
```bash
git clone https://github.com/pandacea/pandacea-protocol.git
cd pandacea-protocol
```

### 2. Set Up Environment
Create a `.env` file in the project root:
```bash
# Required for contract deployment
MUMBAI_RPC_URL=https://polygon-mumbai.infura.io/v3/YOUR_PROJECT_ID
DEPLOYER_PRIVATE_KEY=your_private_key_here

# Optional: For integration testing
AGENT_PEER_ID=your_agent_peer_id
```

### 3. Start the Development Environment
```bash
# Start all services (Agent Backend, Anvil blockchain, IPFS)
docker-compose up -d

# Or run the agent backend directly
cd agent-backend
make run
```

### 4. Test the SDK
```bash
# Run the basic usage example
cd builder-sdk
python examples/basic_usage.py
```

### 5. Verify Everything Works
```bash
# Run integration tests
cd integration
python -m pytest test_integration.py -v
```

## Project Verification

The project includes comprehensive verification scripts to ensure all components are working correctly:

- **Smart Contracts**: `contracts/verify_contracts.py`
- **Agent Backend**: `agent-backend/verify_implementation.py`
- **Builder SDK**: `builder-sdk/verify_sdk.py`

For detailed verification results, see [VERIFICATION.md](VERIFICATION.md).

## 📚 Project Documentation

All detailed technical, strategic, and planning documents are organized in the `/docs` directory for easy access and reference. This includes comprehensive documentation covering every aspect of the Pandacea Protocol.

### Key Documents

- **[Technical Whitepaper](docs/Pandacea%20Technical%20Whitepaper%20(v3.1).pdf)** - Comprehensive technical overview and protocol design
- **[System Design Document (SDD)](docs/Pandacea%20Protocol%20-%20System%20Design%20Document%20(SDD)%20(v1.1).pdf)** - Detailed system architecture and implementation specifications
- **[Comprehensive Roadmap](docs/ROADMAP.md)** - Complete development roadmap and technical plan for future sprints
- **[Engineering Handbook](docs/Pandacea%20Protocol%20-%20Engineering%20Handbook%20(v1).pdf)** - Development standards and best practices
- **[API Specification](docs/Pandacea%20Protocol%20-%20API%20Specification%20(v1.1).pdf)** - Complete API documentation and interface specifications

## Technology Stack

- **Smart Contracts**: Solidity with Foundry for testing
- **Agent Backend**: Go with libp2p for P2P networking
- **Builder SDK**: Python with Poetry and Pytest
- **Blockchain**: Polygon Mumbai testnet
- **Development**: Docker Compose for local environment

## Security & Privacy

- **Structured Logging**: No PII in logs, only event types and metadata
- **Environment Variables**: All secrets read from environment
- **Decimal Precision**: Currency calculations use high-precision decimal library
- **Input Validation**: Strict schema validation for all API requests

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and [Pandacea Protocol - Engineering Handbook (v1).pdf](docs/Pandacea%20Protocol%20-%20Engineering%20Handbook%20(v1).pdf) for detailed technical specifications.

## License

This project is licensed under the MIT License - see the LICENSE file for details.