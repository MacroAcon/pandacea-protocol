#!/usr/bin/env python3
"""
Example demonstrating the Differentiated Dispute Stakes functionality.

This example shows how the new dynamic staking system works:
1. Creates leases of different values
2. Shows how stake amounts are calculated based on lease value
3. Demonstrates the new raise_dispute API without stake_amount parameter
4. Shows how to get required stake amounts before disputing
"""

import os
import sys
import logging
from pathlib import Path

# Add the SDK to the path
sys.path.insert(0, str(Path(__file__).parent.parent))

from pandacea_sdk.client import PandaceaClient
from pandacea_sdk.exceptions import PandaceaException

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def setup_environment():
    """Set up the environment variables for the example."""
    required_vars = {
        "AGENT_URL": "http://localhost:8080",
        "SPENDER_PRIVATE_KEY": "path/to/spender/private/key",
        "RPC_URL": "http://127.0.0.1:8545",
        "CONTRACT_ADDRESS": "0x...",  # LeaseAgreement contract address
        "PGT_TOKEN_ADDRESS": "0x...",  # PGT token contract address
    }
    
    # Check if environment variables are set
    missing_vars = []
    for var, default in required_vars.items():
        if not os.getenv(var):
            missing_vars.append(var)
            logger.warning(f"{var} not set, using default: {default}")
    
    if missing_vars:
        logger.info("Please set the following environment variables for full functionality:")
        for var in missing_vars:
            logger.info(f"  export {var}=<value>")
        logger.info("")
    
    return required_vars

def demonstrate_dynamic_staking():
    """Demonstrate the Differentiated Dispute Stakes functionality."""
    logger.info("🚀 Pandacea Protocol - Differentiated Dispute Stakes Example")
    logger.info("=" * 60)
    
    # Set up environment
    config = setup_environment()
    
    try:
        # Initialize the client
        client = PandaceaClient(
            base_url=config["AGENT_URL"],
            private_key_path=config["SPENDER_PRIVATE_KEY"],
            timeout=30.0
        )
        
        logger.info("✅ Client initialized successfully")
        
        # Example 1: Low-value lease (0.5 ETH)
        logger.info("\n📋 Example 1: Low-Value Lease (0.5 ETH)")
        logger.info("-" * 40)
        
        # Simulate lease creation (in real scenario, this would create an actual lease)
        low_value_lease_id = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
        lease_value_eth = 0.5
        
        logger.info(f"Lease ID: {low_value_lease_id}")
        logger.info(f"Lease Value: {lease_value_eth} ETH")
        
        # Get required stake (10% of lease value)
        try:
            required_stake = client.get_required_stake(low_value_lease_id)
            required_stake_eth = required_stake / 1e18
            logger.info(f"Required Stake: {required_stake_eth:.3f} ETH ({required_stake} wei)")
            logger.info(f"Stake Rate: 10% of lease value")
        except PandaceaException as e:
            logger.warning(f"Could not get required stake (blockchain not connected): {e}")
            # Calculate expected stake manually
            expected_stake = int(lease_value_eth * 1e18 * 0.1)
            logger.info(f"Expected Stake: {lease_value_eth * 0.1:.3f} ETH ({expected_stake} wei)")
        
        # Example 2: High-value lease (20 ETH)
        logger.info("\n📋 Example 2: High-Value Lease (20 ETH)")
        logger.info("-" * 40)
        
        high_value_lease_id = "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
        lease_value_eth = 20
        
        logger.info(f"Lease ID: {high_value_lease_id}")
        logger.info(f"Lease Value: {lease_value_eth} ETH")
        
        # Get required stake (10% of lease value)
        try:
            required_stake = client.get_required_stake(high_value_lease_id)
            required_stake_eth = required_stake / 1e18
            logger.info(f"Required Stake: {required_stake_eth:.1f} ETH ({required_stake} wei)")
            logger.info(f"Stake Rate: 10% of lease value")
        except PandaceaException as e:
            logger.warning(f"Could not get required stake (blockchain not connected): {e}")
            # Calculate expected stake manually
            expected_stake = int(lease_value_eth * 1e18 * 0.1)
            logger.info(f"Expected Stake: {lease_value_eth * 0.1:.1f} ETH ({expected_stake} wei)")
        
        # Example 3: Demonstrating the new raise_dispute API
        logger.info("\n📋 Example 3: New raise_dispute API")
        logger.info("-" * 40)
        
        logger.info("Old API (removed):")
        logger.info("  client.raise_dispute(lease_id, reason, stake_amount)")
        logger.info("")
        logger.info("New API (dynamic staking):")
        logger.info("  client.raise_dispute(lease_id, reason)")
        logger.info("")
        logger.info("The stake amount is now calculated automatically based on:")
        logger.info("  - Lease value")
        logger.info("  - Current dispute stake rate (10% by default)")
        logger.info("")
        
        # Example 4: Stake rate changes
        logger.info("\n📋 Example 4: Stake Rate Changes")
        logger.info("-" * 40)
        
        logger.info("DAO can change the stake rate using setDisputeStakeRate():")
        logger.info("  - Current rate: 10%")
        logger.info("  - New rate: 20%")
        logger.info("")
        
        # Show impact of rate change
        lease_values = [0.5, 1, 5, 20]
        old_rate = 0.1
        new_rate = 0.2
        
        logger.info("Impact of stake rate change:")
        logger.info("Lease Value | Old Stake (10%) | New Stake (20%) | Increase")
        logger.info("-" * 55)
        
        for value in lease_values:
            old_stake = value * old_rate
            new_stake = value * new_rate
            increase = new_stake / old_stake
            logger.info(f"{value:>10.1f} ETH | {old_stake:>12.3f} ETH | {new_stake:>12.3f} ETH | {increase:>7.1f}x")
        
        # Example 5: Best practices
        logger.info("\n📋 Example 5: Best Practices")
        logger.info("-" * 40)
        
        logger.info("1. Always check required stake before disputing:")
        logger.info("   required_stake = client.get_required_stake(lease_id)")
        logger.info("   print(f'You need {required_stake} PGT tokens to dispute this lease')")
        logger.info("")
        
        logger.info("2. Ensure sufficient PGT balance and allowance:")
        logger.info("   # Check PGT balance")
        logger.info("   # Approve PGT tokens for LeaseAgreement contract")
        logger.info("   client.approve_pgt_tokens(contract_address, required_stake)")
        logger.info("")
        
        logger.info("3. Raise dispute with new simplified API:")
        logger.info("   dispute_id = client.raise_dispute(lease_id, 'Data quality issues')")
        logger.info("")
        
        logger.info("4. Monitor stake rate changes:")
        logger.info("   # DAO can change rates to tune the economy")
        logger.info("   # Higher rates = more expensive disputes")
        logger.info("   # Lower rates = cheaper disputes")
        
        logger.info("\n✅ Differentiated Dispute Stakes demonstration completed!")
        logger.info("The new system provides better economic security by scaling")
        logger.info("dispute costs with lease values, preventing frivolous disputes")
        logger.info("on high-value leases while keeping low-value disputes accessible.")
        
    except Exception as e:
        logger.error(f"❌ Example failed: {e}")
        return False
    finally:
        try:
            client.close()
        except:
            pass
    
    return True

def main():
    """Main entry point for the example."""
    try:
        success = demonstrate_dynamic_staking()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        logger.info("Example interrupted by user")
        sys.exit(1)
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
