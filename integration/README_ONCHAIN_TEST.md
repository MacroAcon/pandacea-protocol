# On-Chain Integration Test

This document describes how to run the on-chain lease creation integration test.

## Overview

The integration test verifies the complete on-chain lease creation process by:
1. Connecting to a local Anvil blockchain node
2. Using the PandaceaClient to execute a lease creation transaction
3. Verifying that the `LeaseCreated` event was emitted correctly
4. Validating all event parameters match the transaction data

## Prerequisites

1. **Anvil Node**: A local Anvil node running on `http://127.0.0.1:8545`
2. **Deployed Contract**: The LeaseAgreement contract must be deployed
3. **Environment Variables**: Required configuration must be set

## Setup

### 1. Start Anvil Node

```bash
anvil
```

### 2. Deploy the LeaseAgreement Contract

Deploy the contract using Foundry and note the deployed address.

### 3. Set Environment Variables

```bash
export CONTRACT_ADDRESS="0x1234567890123456789012345678901234567890"  # Your deployed contract address
export SPENDER_PRIVATE_KEY="0x1234567890123456789012345678901234567890123456789012345678901234"  # Spender private key
export RPC_URL="http://127.0.0.1:8545"  # Optional, defaults to this value
```

## Running the Test

### Option 1: Manual Test Script (Recommended)

The manual test script bypasses pytest and can be run directly:

```bash
cd integration
python test_onchain_manual.py
```

This script provides detailed output and handles errors gracefully.

### Option 2: Pytest Integration Test

**Note**: There may be compatibility issues with the web3 pytest plugin. If you encounter import errors, use the manual test script instead.

```bash
cd integration
python -m pytest test_onchain_interaction.py -v
```

## Expected Output

When successful, you should see output similar to:

```
🚀 Manual E2E Test: On-chain Lease Creation
==================================================
✅ RPC URL: http://127.0.0.1:8545
✅ Contract Address: 0x1234567890123456789012345678901234567890
✅ Spender Private Key: 0x12345678...

📡 Initializing PandaceaClient...
✅ Blockchain connection established
✅ Contract loaded successfully

📝 Test parameters:
   Earner: 0x2222222222222222222222222222222222222222
   Data Product ID: 746573742d646174612d70726f647563742d6964000000000000000000000000
   Max Price: 10000000000000000 wei (0.0100 ETH)
   Payment: 1000000000000000 wei (0.0010 ETH)
   Starting block: 123

🔗 Submitting createLease transaction...
✅ Transaction successful with hash: 0xabc123...

🔍 Verifying LeaseCreated event on the blockchain...
✅ Found event. Validating parameters...
✅ Event validation successful!
   LeaseId: 1
   Spender: 0x1234567890123456789012345678901234567890
   Earner: 0x2222222222222222222222222222222222222222
   Price: 1000000000000000 wei

🎉 E2E test passed: On-chain lease created and event verified successfully!
```

## Troubleshooting

### Common Issues

1. **"CONTRACT_ADDRESS environment variable not set"**
   - Deploy the contract and set the CONTRACT_ADDRESS environment variable

2. **"SPENDER_PRIVATE_KEY environment variable not set"**
   - Set the SPENDER_PRIVATE_KEY environment variable with a valid private key

3. **"Web3 connection failed"**
   - Ensure Anvil is running on the correct port
   - Check the RPC_URL environment variable

4. **"Contract not loaded"**
   - Verify the contract address is correct
   - Ensure the ABI file exists at `../contracts/LeaseAgreement.abi.json`

5. **"Transaction failed"**
   - Check that the spender account has sufficient ETH
   - Verify the contract address is correct
   - Ensure the transaction parameters are valid

6. **"Expected exactly one LeaseCreated event, but found 0"**
   - The transaction may have failed or the event was not emitted
   - Check the transaction receipt for errors

### Pytest Plugin Issues

If you encounter import errors related to the web3 pytest plugin, use the manual test script instead:

```bash
python test_onchain_manual.py
```

## Test Parameters

The test uses the following parameters:

- **Earner Address**: `0x2222222222222222222222222222222222222222` (dummy address)
- **Data Product ID**: `test-data-product-id` (padded to 32 bytes)
- **Max Price**: 0.01 ETH (10,000,000,000,000,000 wei)
- **Payment**: 0.001 ETH (1,000,000,000,000,000 wei)

These parameters can be modified in the test script if needed.

## Integration with CI/CD

For automated testing, ensure the following are available in your CI environment:

1. Anvil node running
2. Contract deployed
3. Environment variables set
4. Sufficient ETH in the spender account

The test will automatically skip if required environment variables are not set. 