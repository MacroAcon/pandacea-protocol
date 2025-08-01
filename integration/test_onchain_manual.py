#!/usr/bin/env python3
"""
Manual test script for on-chain lease creation functionality.
This script can be run directly without pytest to verify the integration works.
"""

import os
import sys
import time
from web3 import Web3

# Add the builder-sdk to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'builder-sdk'))

from pandacea_sdk.client import PandaceaClient

def test_onchain_integration():
    """Test the on-chain lease creation functionality."""
    
    print("🚀 Manual E2E Test: On-chain Lease Creation")
    print("=" * 50)
    
    # Check environment variables
    rpc_url = os.getenv("RPC_URL", "http://127.0.0.1:8545")
    contract_address = os.getenv("CONTRACT_ADDRESS")
    spender_private_key = os.getenv("SPENDER_PRIVATE_KEY")
    
    if not contract_address:
        print("❌ CONTRACT_ADDRESS environment variable not set")
        print("   Please set it to the deployed LeaseAgreement contract address")
        return False
    
    if not spender_private_key:
        print("❌ SPENDER_PRIVATE_KEY environment variable not set")
        print("   Please set it to the private key of the spender account")
        return False
    
    print(f"✅ RPC URL: {rpc_url}")
    print(f"✅ Contract Address: {contract_address}")
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
        
        # Define test parameters
        earner_address = "0x" + "2" * 40  # A valid, but likely unused address
        data_product_id = b'test-data-product-id'.ljust(32, b'\0')  # Must be 32 bytes
        max_price_wei = Web3.to_wei(0.01, 'ether')
        payment_wei = Web3.to_wei(0.001, 'ether')
        
        print(f"\n📝 Test parameters:")
        print(f"   Earner: {earner_address}")
        print(f"   Data Product ID: {data_product_id.hex()}")
        print(f"   Max Price: {max_price_wei} wei ({max_price_wei / 1e18:.4f} ETH)")
        print(f"   Payment: {payment_wei} wei ({payment_wei / 1e18:.4f} ETH)")
        
        # Get the latest block number before the transaction
        latest_block = client.w3.eth.block_number
        print(f"   Starting block: {latest_block}")
        
        # Execute the on-chain lease creation
        print(f"\n🔗 Submitting createLease transaction...")
        tx_hash = client.execute_lease_on_chain(
            earner=earner_address,
            data_product_id=data_product_id,
            max_price=max_price_wei,
            payment_in_wei=payment_wei
        )
        print(f"✅ Transaction successful with hash: {tx_hash}")
        
        # Verify the LeaseCreated Event
        print("\n🔍 Verifying LeaseCreated event on the blockchain...")
        
        # Give the node a moment, then check for the event in recent blocks
        time.sleep(2)
        
        event_filter = client.contract.events.LeaseCreated.create_filter(
            fromBlock=latest_block + 1
        )
        logs = event_filter.get_all_entries()
        
        if len(logs) != 1:
            print(f"❌ Expected exactly one LeaseCreated event, but found {len(logs)}")
            return False
        
        event_data = logs[0]['args']
        spender_account = client.w3.eth.account.from_key(client.spender_private_key)
        
        # Assert Event Parameters
        print(f"✅ Found event. Validating parameters...")
        
        if event_data['spender'] != spender_account.address:
            print(f"❌ Spender mismatch: expected {spender_account.address}, got {event_data['spender']}")
            return False
        
        if event_data['earner'] != Web3.to_checksum_address(earner_address):
            print(f"❌ Earner mismatch: expected {earner_address}, got {event_data['earner']}")
            return False
        
        if event_data['price'] != payment_wei:
            print(f"❌ Price mismatch: expected {payment_wei}, got {event_data['price']}")
            return False
        
        # The leaseId is dynamically generated, so we just check it exists
        if event_data['leaseId'] is None:
            print("❌ LeaseId is None")
            return False
        
        print(f"✅ Event validation successful!")
        print(f"   LeaseId: {event_data['leaseId']}")
        print(f"   Spender: {event_data['spender']}")
        print(f"   Earner: {event_data['earner']}")
        print(f"   Price: {event_data['price']} wei")
        
        print("\n🎉 E2E test passed: On-chain lease created and event verified successfully!")
        return True
        
    except Exception as e:
        print(f"❌ Error: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    success = test_onchain_integration()
    sys.exit(0 if success else 1) 