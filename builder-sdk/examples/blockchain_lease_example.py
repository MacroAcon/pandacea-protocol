#!/usr/bin/env python3
"""
Example script demonstrating on-chain lease creation with the Pandacea SDK.

This example shows how to:
1. Initialize the PandaceaClient with blockchain integration
2. Execute a lease creation transaction on the blockchain
3. Handle various error conditions

Prerequisites:
- Anvil node running on http://127.0.0.1:8545
- Environment variables set:
  - CONTRACT_ADDRESS: The deployed LeaseAgreement contract address
  - SPENDER_PRIVATE_KEY: The private key of the spender account
  - RPC_URL: The RPC URL (defaults to http://127.0.0.1:8545)
"""

import os
import sys
from pandacea_sdk.client import PandaceaClient

def main():
    """Demonstrate on-chain lease creation."""
    
    print("🚀 Pandacea Blockchain Lease Creation Example")
    print("=" * 50)
    
    # Check environment variables
    contract_address = os.getenv("CONTRACT_ADDRESS")
    spender_private_key = os.getenv("SPENDER_PRIVATE_KEY")
    rpc_url = os.getenv("RPC_URL", "http://127.0.0.1:8545")
    
    if not contract_address:
        print("❌ CONTRACT_ADDRESS environment variable not set")
        print("   Please set it to the deployed LeaseAgreement contract address")
        return False
    
    if not spender_private_key:
        print("❌ SPENDER_PRIVATE_KEY environment variable not set")
        print("   Please set it to the private key of the spender account")
        return False
    
    print(f"✅ Contract Address: {contract_address}")
    print(f"✅ RPC URL: {rpc_url}")
    print(f"✅ Spender Private Key: {spender_private_key[:10]}...")
    
    try:
        # Initialize the client
        print("\n📡 Initializing PandaceaClient...")
        client = PandaceaClient("http://localhost:8080")
        
        # Check blockchain connection
        if not client.w3:
            print("❌ Web3 connection failed. Please ensure Anvil is running.")
            return False
        
        if not client.contract:
            print("❌ Contract not loaded. Please check CONTRACT_ADDRESS and ABI file.")
            return False
        
        print("✅ Blockchain connection established")
        print("✅ Contract loaded successfully")
        
        # Example parameters for lease creation
        earner_address = "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"  # Example earner address
        data_product_id = b"example_data_product_id_32_bytes_long"  # 32-byte product ID
        max_price_wei = 1000000000000000000  # 1 ETH in wei
        payment_wei = 500000000000000000  # 0.5 ETH in wei
        
        print(f"\n📝 Creating lease with parameters:")
        print(f"   Earner: {earner_address}")
        print(f"   Data Product ID: {data_product_id.hex()}")
        print(f"   Max Price: {max_price_wei} wei ({max_price_wei / 1e18:.2f} ETH)")
        print(f"   Payment: {payment_wei} wei ({payment_wei / 1e18:.2f} ETH)")
        
        # Execute the lease creation transaction
        print("\n🔗 Executing lease creation transaction...")
        tx_hash = client.execute_lease_on_chain(
            earner=earner_address,
            data_product_id=data_product_id,
            max_price=max_price_wei,
            payment_in_wei=payment_wei
        )
        
        print(f"✅ Transaction successful!")
        print(f"   Transaction Hash: {tx_hash}")
        
        return True
        
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1) 