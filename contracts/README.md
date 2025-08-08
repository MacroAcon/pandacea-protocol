# Pandacea Protocol - Smart Contracts

This directory contains the smart contracts for the Pandacea Protocol MVP.

## Implementation Summary

### LeaseAgreement.sol
- **Interface**: `ILeaseAgreement` - Defines the core functions and events for data lease management
- **Implementation**: `LeaseAgreement` - Full implementation with Dynamic Minimum Pricing (DMP) logic

### Key Features Implemented

#### Dynamic Minimum Pricing (DMP)
- **MIN_PRICE**: Set to 0.001 ETH (1,000,000,000,000,000 wei)
- **Validation**: All lease creations must meet or exceed the minimum price
- **Error Message**: "LeaseAgreement: Insufficient payment - below minimum price"

#### Core Functions
1. **createLease**: Creates a new lease agreement with DMP validation
2. **approveLease**: Allows designated earner to approve a lease
3. **executeLease**: Executes an approved lease
4. **raiseDispute**: Allows spender or earner to raise disputes

#### Security Features
- **ReentrancyGuard**: Prevents reentrant attacks on all state-changing functions
- **Ownable**: Provides owner-only functions for contract management
- **Input Validation**: Comprehensive validation for all function parameters

### Test Coverage

The test suite (`LeaseAgreement.t.sol`) covers:

#### Success Cases ✅
- User can create lease with exact MIN_PRICE
- User can create lease with more than MIN_PRICE
- LeaseCreated event is correctly emitted
- Designated earner can approve lease
- Multiple leases can be created

#### Failure Cases ✅
- createLease reverts when payment < MIN_PRICE
- createLease reverts when payment > maxPrice
- createLease reverts with invalid earner address
- approveLease reverts when called by non-earner
- approveLease reverts for non-existent lease
- approveLease reverts when already approved
- approveLease reverts for disputed lease

#### Edge Cases ✅
- MIN_PRICE constant verification
- Lease counter increments correctly
- Dispute functionality for both spender and earner
- All revert conditions properly tested

### TODO Items (Future Implementation)
- Data product validation logic
- Reputation system integration
- Escrow mechanism for payments
- Data transfer initiation
- Payment escrow release conditions
- Access control for execution
- Data access verification
- Payment release logic
- Dispute resolution mechanism
- Arbitration system
- Refund logic for disputed leases
- Minimum price update logic
- Emergency pause functionality

## Project Structure
```
contracts/
├── src/
│   └── LeaseAgreement.sol          # Main contract implementation
├── test/
│   └── LeaseAgreement.t.sol        # Comprehensive test suite
├── lib/
│   ├── openzeppelin-contracts/     # OpenZeppelin dependencies
│   └── forge-std/                  # Foundry testing framework
├── foundry.toml                    # Foundry configuration
├── remappings.txt                  # Import mappings
└── README.md                       # This file
```

## Setup Instructions

1. **Install Foundry** (if not already installed):
   ```bash
   curl -L https://foundry.paradigm.xyz | bash
   foundryup
   ```

2. **Install Dependencies**:
   ```bash
   forge install OpenZeppelin/openzeppelin-contracts
   ```

3. **Run Tests**:
   ```bash
   forge test
   ```

## Windows Quickstart

For Windows users, see the [Windows Quickstart Guide](../docs/windows_quickstart.md) for detailed setup instructions including:

- Foundry installation on Windows
- PowerShell commands
- Docker setup for PySyft
- Troubleshooting common issues
   ```

4. **Compile Contracts**:
   ```bash
   forge build
   ```

## Contract Addresses
After deployment, contract addresses will be stored in `/config/deployments.json` for use by the agent backend and builder SDK.

## Security Considerations
- All state-changing functions are protected with `nonReentrant` modifier
- Owner-only functions for contract management
- Comprehensive input validation
- Clear error messages for debugging
- DMP ensures minimum economic value for all transactions 