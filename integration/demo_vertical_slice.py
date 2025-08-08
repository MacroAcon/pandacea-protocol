#!/usr/bin/env python3
"""
Pandacea Protocol Vertical Slice Demo

This script demonstrates the complete end-to-end flow:
1. Posts a federated learning job to the agent backend
2. Polls for completion 
3. Mints a lease via the Python SDK using local blockchain
4. Fetches the produced data product manifest

Prerequisites:
- Agent backend running on localhost:8080
- Anvil blockchain running on localhost:8545
- Environment variables set in integration/.env
"""

import os
import sys
import time
import json
import requests
from pathlib import Path
from dotenv import load_dotenv

# Add the parent directory to Python path to import SDK
sys.path.insert(0, str(Path(__file__).parent.parent / "builder-sdk"))

from pandacea_sdk.client import PandaceaClient
from pandacea_sdk.exceptions import PandaceaException, AgentConnectionError, APIResponseError


def load_environment():
    """Load environment variables from .env file"""
    env_path = Path(__file__).parent / ".env"
    if not env_path.exists():
        print(f"❌ Environment file not found: {env_path}")
        print("Please copy integration/.env.example to integration/.env and configure it")
        sys.exit(1)
    
    load_dotenv(env_path)
    
    # Validate required environment variables
    required_vars = ["AGENT_URL", "RPC_URL", "CHAIN_ID", "PGT_ADDRESS", "LEASE_ADDRESS"]
    missing_vars = []
    
    for var in required_vars:
        if not os.getenv(var):
            missing_vars.append(var)
    
    if missing_vars:
        print(f"❌ Missing required environment variables: {', '.join(missing_vars)}")
        sys.exit(1)
    
    return {
        "agent_url": os.getenv("AGENT_URL"),
        "rpc_url": os.getenv("RPC_URL"),
        "chain_id": int(os.getenv("CHAIN_ID")),
        "pgt_address": os.getenv("PGT_ADDRESS"),
        "lease_address": os.getenv("LEASE_ADDRESS"),
        "timeout": int(os.getenv("TIMEOUT_SECONDS", "300")),
        "poll_interval": float(os.getenv("POLL_INTERVAL_SECONDS", "2"))
    }


def post_training_job(agent_url, timeout=30):
    """
    Post a federated learning job to the agent backend
    
    Args:
        agent_url: URL of the agent backend
        timeout: Request timeout in seconds
        
    Returns:
        job_id: The ID of the created job
    """
    print("📊 Posting federated learning job...")
    
    payload = {
        "dataset": "toy_telemetry",
        "task": "logreg", 
        "dp": {
            "enabled": True,
            "epsilon": 5.0
        }
    }
    
    try:
        response = requests.post(
            f"{agent_url}/train",
            json=payload,
            timeout=timeout,
            headers={"Content-Type": "application/json"}
        )
        response.raise_for_status()
        
        data = response.json()
        job_id = data.get("job_id")
        
        if not job_id:
            print(f"❌ No job_id in response: {data}")
            sys.exit(1)
            
        print(f"✅ Training job created with ID: {job_id}")
        return job_id
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Failed to post training job: {e}")
        sys.exit(1)


def wait_for_completion(agent_url, job_id, timeout=300, poll_interval=2):
    """
    Poll the aggregate endpoint until the job completes
    
    Args:
        agent_url: URL of the agent backend
        job_id: ID of the job to poll
        timeout: Maximum time to wait in seconds
        poll_interval: Time between polls in seconds
        
    Returns:
        result: The final job result
    """
    print(f"⏳ Waiting for job {job_id} to complete...")
    
    start_time = time.time()
    
    while time.time() - start_time < timeout:
        try:
            response = requests.get(
                f"{agent_url}/aggregate/{job_id}",
                timeout=30
            )
            response.raise_for_status()
            
            result = response.json()
            status = result.get("status")
            
            print(f"   Status: {status}")
            
            if status == "complete":
                print("✅ Job completed successfully!")
                return result
            elif status == "failed":
                error = result.get("error", "Unknown error")
                print(f"❌ Job failed: {error}")
                sys.exit(1)
            elif status in ["pending", "running"]:
                time.sleep(poll_interval)
                continue
            else:
                print(f"❌ Unknown status: {status}")
                sys.exit(1)
                
        except requests.exceptions.RequestException as e:
            print(f"⚠️ Error polling job status: {e}")
            time.sleep(poll_interval)
            continue
    
    print(f"❌ Job timed out after {timeout} seconds")
    sys.exit(1)


def mint_local_lease(config, data_product_manifest):
    """
    Mint a lease via the Python SDK using the local blockchain
    
    Args:
        config: Environment configuration
        data_product_manifest: The data product information
        
    Returns:
        lease_info: Information about the created lease
    """
    print("🔗 Minting lease on local blockchain...")
    
    try:
        # Set up environment variables for the SDK
        os.environ["RPC_URL"] = config["rpc_url"]
        os.environ["CONTRACT_ADDRESS"] = config["lease_address"]
        os.environ["PGT_TOKEN_ADDRESS"] = config["pgt_address"]
        # For demo purposes, use a test private key (in production, this would be secure)
        os.environ["SPENDER_PRIVATE_KEY"] = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
        
        # Initialize the Pandacea client
        client = PandaceaClient(config["agent_url"])
        
        # Create a mock lease for demo purposes
        # In a real scenario, this would interact with the actual data product
        earner_address = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"  # Test address
        data_product_id = b"toy_telemetry_product_12345678901234"  # 32 bytes
        max_price = 1000000000000000000  # 1 ETH in wei
        payment_in_wei = 1000000000000000  # 0.001 ETH (minimum)
        
        artifact_path = data_product_manifest.get('artifact_path', 'N/A')
        print(f"   Creating lease for data product: {artifact_path}")
        print(f"   Earner: {earner_address}")
        print(f"   Payment: {payment_in_wei / 1e18} ETH")
        
        # Test blockchain connectivity before executing lease
        print("   Testing blockchain connectivity...")
        
        # Execute the lease on-chain
        tx_hash = client.execute_lease_on_chain(
            earner=earner_address,
            data_product_id=data_product_id,
            max_price=max_price,
            payment_in_wei=payment_in_wei
        )
        
        print(f"✅ Lease created! Transaction hash: {tx_hash}")
        
        return {
            "tx_hash": tx_hash,
            "earner": earner_address,
            "payment_eth": payment_in_wei / 1e18,
            "data_product_id": data_product_id.hex()
        }
        
    except Exception as e:
        print(f"❌ Failed to mint lease: {e}")
        print("   This might be due to:")
        print("   - Anvil blockchain not running on localhost:8545")
        print("   - Contracts not deployed or wrong addresses")
        print("   - Network connectivity issues")
        sys.exit(1)


def print_summary(job_result, lease_info):
    """
    Print a summary of the demo results
    
    Args:
        job_result: Result from the federated learning job
        lease_info: Information about the created lease
    """
    print("\n" + "="*60)
    print("🎉 DEMO COMPLETED SUCCESSFULLY!")
    print("="*60)
    
    print("\n📊 Federated Learning Job:")
    print(f"   Job ID: {job_result.get('job_id', 'N/A')}")
    print(f"   Status: {job_result.get('status', 'N/A')}")
    print(f"   Epsilon Used: {job_result.get('epsilon', 'N/A')}")
    print(f"   Artifact Path: {job_result.get('artifact_path', 'N/A')}")
    
    print("\n🔗 Blockchain Lease:")
    print(f"   Transaction Hash: {lease_info.get('tx_hash', 'N/A')}")
    print(f"   Earner Address: {lease_info.get('earner', 'N/A')}")
    print(f"   Payment Amount: {lease_info.get('payment_eth', 'N/A')} ETH")
    print(f"   Data Product ID: {lease_info.get('data_product_id', 'N/A')}")
    
    print(f"\n📁 Data Product Manifest:")
    artifact_path = job_result.get('artifact_path', 'N/A')
    epsilon = job_result.get('epsilon', 'N/A')
    print(f"   Path: {artifact_path}")
    print(f"   Privacy Budget (ε): {epsilon}")
    
    print("\n✨ End-to-end demo completed successfully!")


def main():
    """Main demo function"""
    print("🚀 Starting Pandacea Protocol Vertical Slice Demo")
    print("=" * 60)
    
    # Load environment configuration
    config = load_environment()
    print(f"🔧 Configuration loaded from environment")
    print(f"   Agent URL: {config['agent_url']}")
    print(f"   RPC URL: {config['rpc_url']}")
    print(f"   Chain ID: {config['chain_id']}")
    
    try:
        # Step 1: Post training job
        job_id = post_training_job(config["agent_url"])
        
        # Step 2: Wait for completion
        job_result = wait_for_completion(
            config["agent_url"], 
            job_id, 
            config["timeout"], 
            config["poll_interval"]
        )
        
        # Step 3: Mint local lease
        lease_info = mint_local_lease(config, job_result)
        
        # Step 4: Print summary
        print_summary(job_result, lease_info)
        
        return 0
        
    except KeyboardInterrupt:
        print("\n❌ Demo interrupted by user")
        return 1
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
