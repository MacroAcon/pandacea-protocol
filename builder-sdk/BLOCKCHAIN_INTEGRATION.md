# Blockchain Integration in Pandacea SDK

This document describes the blockchain integration features added to the Pandacea Python SDK.

## Overview

The PandaceaClient now supports on-chain lease creation through integration with the LeaseAgreement smart contract. This allows users to execute lease proposals directly on the blockchain, which will emit events that the Go agent backend can listen for.

## Features

- **Web3 Integration**: Automatic connection to Ethereum nodes via Web3.py
- **Smart Contract Interaction**: Direct interaction with the LeaseAgreement contract
- **Transaction Management**: Building, signing, and sending transactions
- **Error Handling**: Comprehensive error handling for blockchain operations

## Prerequisites

1. **Anvil Node**: A local Anvil node running on `http://127.0.0.1:8545`
2. **Deployed Contract**: The LeaseAgreement contract must be deployed and its address known
3. **Private Key**: A spender account private key for signing transactions
4. **ABI File**: The `LeaseAgreement.abi.json` file must be available

## Environment Variables

The following environment variables are required for blockchain functionality:

- `CONTRACT_ADDRESS`: The deployed LeaseAgreement contract address
- `SPENDER_PRIVATE_KEY`: The private key of the spender account (with 0x prefix)
- `RPC_URL`: The RPC URL for the blockchain node (defaults to `http://127.0.0.1:8545`)

## Usage

### Basic Setup

```python
import os
from pandacea_sdk.client import PandaceaClient

# Set environment variables
os.environ["CONTRACT_ADDRESS"] = "0x1234567890123456789012345678901234567890"
os.environ["SPENDER_PRIVATE_KEY"] = "0x1234567890123456789012345678901234567890123456789012345678901234"
os.environ["RPC_URL"] = "http://127.0.0.1:8545"

# Initialize client
client = PandaceaClient("http://localhost:8080")
```

### Creating a Lease On-Chain

```python
# Example parameters
earner_address = "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
data_product_id = b"example_data_product_id_32_bytes_long"  # Must be 32 bytes
max_price_wei = 1000000000000000000  # 1 ETH in wei
payment_wei = 500000000000000000  # 0.5 ETH in wei

# Execute the transaction
tx_hash = client.execute_lease_on_chain(
    earner=earner_address,
    data_product_id=data_product_id,
    max_price=max_price_wei,
    payment_in_wei=payment_wei
)

print(f"Transaction hash: {tx_hash}")
```

## Method Reference

### `execute_lease_on_chain(earner, data_product_id, max_price, payment_in_wei)`

Creates a lease on the blockchain by calling the `createLease` function of the LeaseAgreement contract.

**Parameters:**
- `earner` (str): The blockchain address of the data earner
- `data_product_id` (bytes): The 32-byte ID of the data product
- `max_price` (int): The maximum price the spender is willing to pay, in wei
- `payment_in_wei` (int): The amount to send with the transaction, in wei

**Returns:**
- `str`: The transaction hash as a hex string

**Raises:**
- `PandaceaException`: If Web3 is not connected, contract is not initialized, or private key is not set
- `APIResponseError`: If the transaction fails

## Error Handling

The SDK provides comprehensive error handling for blockchain operations:

1. **Connection Errors**: Graceful handling when blockchain node is not available
2. **Contract Errors**: Clear error messages when contract is not properly initialized
3. **Transaction Errors**: Detailed error information when transactions fail
4. **Configuration Errors**: Helpful messages when required environment variables are missing

## Example Script

See `examples/blockchain_lease_example.py` for a complete working example.

## Testing

To test the blockchain integration:

1. Start an Anvil node: `anvil`
2. Deploy the LeaseAgreement contract
3. Set the required environment variables
4. Run the example script: `python examples/blockchain_lease_example.py`

## Integration with Go Agent Backend

When a lease is created on-chain, the LeaseAgreement contract will emit a `LeaseCreated` event. The Go agent backend should listen for these events to process the lease creation.

The event structure is:
```solidity
event LeaseCreated(
    address indexed spender,
    address indexed earner,
    bytes32 indexed dataProductId,
    uint256 maxPrice,
    uint256 payment
);
```

## Security Considerations

1. **Private Key Management**: Never hardcode private keys in your code
2. **Environment Variables**: Use secure methods to manage environment variables
3. **Network Security**: Ensure RPC endpoints are secure and trusted
4. **Transaction Validation**: Always validate transaction parameters before sending
5. **Gas Estimation**: Consider using dynamic gas estimation for production use

## Troubleshooting

### Common Issues

1. **"Failed to connect to blockchain node"**
   - Ensure Anvil is running on the correct port
   - Check the RPC_URL environment variable

2. **"Contract is not initialized"**
   - Verify CONTRACT_ADDRESS is set correctly
   - Ensure the ABI file exists and is readable

3. **"SPENDER_PRIVATE_KEY environment variable not set"**
   - Set the SPENDER_PRIVATE_KEY environment variable
   - Ensure the private key has the 0x prefix

4. **"Transaction failed"**
   - Check that the spender account has sufficient ETH
   - Verify the contract address is correct
   - Ensure the transaction parameters are valid 