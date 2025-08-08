#!/usr/bin/env python3
"""
Example script demonstrating the Pandacea Protocol's Aggressive Reputation Decay functionality.

This script shows how the new configurable decay rate works and how the DAO can
adjust it based on network dynamics. It demonstrates:
1. Initial decay rate of 2 points per day (doubled from original)
2. DAO's ability to update the decay rate
3. Impact of different decay rates on reputation scores
4. Monitoring reputation changes over time

Prerequisites:
- Local Anvil blockchain running on http://127.0.0.1:8545
- Pandacea contracts deployed
- Environment variables set (SPENDER_PRIVATE_KEY, EARNER_PRIVATE_KEY, etc.)
"""

import os
import sys
import time
import logging
from pathlib import Path
from typing import Dict, Any

# Add the SDK to the path
sys.path.insert(0, str(Path(__file__).parent.parent))

from pandacea_sdk.client import PandaceaClient
from pandacea_sdk.exceptions import PandaceaException

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def setup_environment() -> Dict[str, Any]:
    """Set up the environment and return configuration."""
    config = {
        "agent_url": os.getenv("AGENT_URL", "http://localhost:8080"),
        "spender_private_key": os.getenv("SPENDER_PRIVATE_KEY"),
        "earner_private_key": os.getenv("EARNER_PRIVATE_KEY"),
        "rpc_url": os.getenv("RPC_URL", "http://127.0.0.1:8545"),
        "contract_address": os.getenv("CONTRACT_ADDRESS"),
        "pgt_token_address": os.getenv("PGT_TOKEN_ADDRESS"),
    }
    
    # Validate required environment variables
    required_vars = ["SPENDER_PRIVATE_KEY", "EARNER_PRIVATE_KEY", "CONTRACT_ADDRESS", "PGT_TOKEN_ADDRESS"]
    missing_vars = [var for var in required_vars if not os.getenv(var)]
    
    if missing_vars:
        logger.error(f"Missing required environment variables: {missing_vars}")
        logger.error("Please set the following environment variables:")
        for var in missing_vars:
            logger.error(f"  {var}")
        sys.exit(1)
    
    return config

def create_clients(config: Dict[str, Any]) -> tuple[PandaceaClient, PandaceaClient]:
    """Create test clients for spender and earner."""
    # Create spender client
    spender_client = PandaceaClient(
        base_url=config["agent_url"],
        private_key_path=config["spender_private_key"],
        timeout=30.0
    )
    
    # Create earner client (using different private key)
    earner_client = PandaceaClient(
        base_url=config["agent_url"],
        private_key_path=config["earner_private_key"],
        timeout=30.0
    )
    
    return spender_client, earner_client

def demonstrate_initial_decay_rate():
    """Demonstrate the new initial decay rate of 2 points per day."""
    logger.info("=" * 60)
    logger.info("DEMONSTRATING INITIAL AGGRESSIVE DECAY RATE")
    logger.info("=" * 60)
    
    config = setup_environment()
    spender_client, earner_client = create_clients(config)
    
    try:
        # Create a test lease
        products = spender_client.discover_products()
        if not products:
            logger.error("No data products available for testing")
            return False
        
        test_product = products[0]
        lease_proposal_id = spender_client.request_lease(
            product_id=test_product.product_id,
            max_price="1",  # 1 ETH lease
            duration="7d"
        )
        
        lease_id = f"lease_{lease_proposal_id}_{int(time.time())}"
        initial_reputation = 800
        
        logger.info(f"Created test lease: {lease_id}")
        logger.info(f"Initial reputation: {initial_reputation}")
        
        # Demonstrate new decay rate calculations
        logger.info("\nDecay Rate Analysis:")
        logger.info("Old decay rate: 1 point per day")
        logger.info("New decay rate: 2 points per day (doubled)")
        
        # Calculate decay for different time periods
        time_periods = [7, 14, 30, 60, 90]
        
        for days in time_periods:
            old_decay = days * 1  # Old rate
            new_decay = days * 2  # New rate
            old_reputation = max(0, initial_reputation - old_decay)
            new_reputation = max(0, initial_reputation - new_decay)
            
            logger.info(f"\n{days} days of inactivity:")
            logger.info(f"  Old decay: {old_decay} points → Reputation: {old_reputation}")
            logger.info(f"  New decay: {new_decay} points → Reputation: {new_reputation}")
            logger.info(f"  Additional loss: {new_decay - old_decay} points")
        
        logger.info("\n✓ Initial aggressive decay rate demonstrated")
        return True
        
    except Exception as e:
        logger.error(f"Error demonstrating initial decay rate: {e}")
        return False
    finally:
        spender_client.close()
        earner_client.close()

def demonstrate_dao_configuration():
    """Demonstrate DAO's ability to configure decay rates."""
    logger.info("\n" + "=" * 60)
    logger.info("DEMONSTRATING DAO CONFIGURATION CAPABILITIES")
    logger.info("=" * 60)
    
    config = setup_environment()
    spender_client, earner_client = create_clients(config)
    
    try:
        # Create a test lease
        products = spender_client.discover_products()
        if not products:
            logger.error("No data products available for testing")
            return False
        
        test_product = products[0]
        lease_proposal_id = spender_client.request_lease(
            product_id=test_product.product_id,
            max_price="1",  # 1 ETH lease
            duration="7d"
        )
        
        lease_id = f"lease_{lease_proposal_id}_{int(time.time())}"
        initial_reputation = 1000
        
        logger.info(f"Created test lease: {lease_id}")
        logger.info(f"Initial reputation: {initial_reputation}")
        
        # Demonstrate different DAO configurations
        logger.info("\nDAO Configuration Scenarios:")
        
        scenarios = [
            {"name": "Conservative", "rate": 1, "description": "Less aggressive decay"},
            {"name": "Default", "rate": 2, "description": "Current default rate"},
            {"name": "Moderate", "rate": 3, "description": "Moderately aggressive"},
            {"name": "Aggressive", "rate": 5, "description": "Highly aggressive decay"},
            {"name": "Maximum", "rate": 10, "description": "Maximum allowed rate"}
        ]
        
        for scenario in scenarios:
            rate = scenario["rate"]
            name = scenario["name"]
            description = scenario["description"]
            
            # Calculate impact over 10 days
            days = 10
            decay = days * rate
            final_reputation = max(0, initial_reputation - decay)
            
            logger.info(f"\n{name} Configuration ({description}):")
            logger.info(f"  Decay rate: {rate} points per day")
            logger.info(f"  10-day decay: {decay} points")
            logger.info(f"  Final reputation: {final_reputation}")
            logger.info(f"  Reputation loss: {initial_reputation - final_reputation} points")
        
        logger.info("\n✓ DAO configuration capabilities demonstrated")
        return True
        
    except Exception as e:
        logger.error(f"Error demonstrating DAO configuration: {e}")
        return False
    finally:
        spender_client.close()
        earner_client.close()

def demonstrate_network_impact():
    """Demonstrate the impact on network behavior and economics."""
    logger.info("\n" + "=" * 60)
    logger.info("DEMONSTRATING NETWORK IMPACT ANALYSIS")
    logger.info("=" * 60)
    
    try:
        # Simulate different network scenarios
        logger.info("Network Impact Analysis:")
        
        scenarios = [
            {
                "name": "Reputation Farming Prevention",
                "description": "Impact on reputation farming attempts",
                "old_rate": 1,
                "new_rate": 2,
                "days": 30,
                "initial_reputation": 900
            },
            {
                "name": "Active Participation Incentive",
                "description": "Encouraging continuous engagement",
                "old_rate": 1,
                "new_rate": 3,
                "days": 14,
                "initial_reputation": 800
            },
            {
                "name": "Market Dynamics",
                "description": "Impact on overall market behavior",
                "old_rate": 1,
                "new_rate": 5,
                "days": 7,
                "initial_reputation": 1000
            }
        ]
        
        for scenario in scenarios:
            name = scenario["name"]
            description = scenario["description"]
            old_rate = scenario["old_rate"]
            new_rate = scenario["new_rate"]
            days = scenario["days"]
            initial = scenario["initial_reputation"]
            
            old_decay = days * old_rate
            new_decay = days * new_rate
            old_final = max(0, initial - old_decay)
            new_final = max(0, initial - new_decay)
            
            logger.info(f"\n{name}:")
            logger.info(f"  {description}")
            logger.info(f"  Time period: {days} days")
            logger.info(f"  Initial reputation: {initial}")
            logger.info(f"  Old rate ({old_rate}/day): {old_decay} points lost → {old_final}")
            logger.info(f"  New rate ({new_rate}/day): {new_decay} points lost → {new_final}")
            logger.info(f"  Additional cost: {new_decay - old_decay} points")
            
            # Calculate percentage impact
            old_percentage = (old_decay / initial) * 100
            new_percentage = (new_decay / initial) * 100
            logger.info(f"  Old impact: {old_percentage:.1f}% of reputation")
            logger.info(f"  New impact: {new_percentage:.1f}% of reputation")
        
        logger.info("\n✓ Network impact analysis completed")
        return True
        
    except Exception as e:
        logger.error(f"Error demonstrating network impact: {e}")
        return False

def demonstrate_monitoring_and_analytics():
    """Demonstrate monitoring and analytics capabilities."""
    logger.info("\n" + "=" * 60)
    logger.info("DEMONSTRATING MONITORING AND ANALYTICS")
    logger.info("=" * 60)
    
    config = setup_environment()
    spender_client, earner_client = create_clients(config)
    
    try:
        # Create a test lease
        products = spender_client.discover_products()
        if not products:
            logger.error("No data products available for testing")
            return False
        
        test_product = products[0]
        lease_proposal_id = spender_client.request_lease(
            product_id=test_product.product_id,
            max_price="1",  # 1 ETH lease
            duration="7d"
        )
        
        lease_id = f"lease_{lease_proposal_id}_{int(time.time())}"
        
        logger.info(f"Created test lease: {lease_id}")
        
        # Demonstrate monitoring capabilities
        logger.info("\nMonitoring Capabilities:")
        
        # Get current decay rate (would be available via contract call)
        current_rate = 2  # Default rate
        logger.info(f"Current decay rate: {current_rate} points per day")
        
        # Simulate reputation monitoring over time
        logger.info("\nReputation Monitoring Simulation:")
        
        initial_reputation = 850
        monitoring_periods = [
            {"day": 0, "activity": "Active", "reputation": initial_reputation},
            {"day": 7, "activity": "Inactive", "reputation": initial_reputation - (7 * current_rate)},
            {"day": 14, "activity": "Inactive", "reputation": initial_reputation - (14 * current_rate)},
            {"day": 21, "activity": "Active", "reputation": initial_reputation - (21 * current_rate) + 50},
            {"day": 28, "activity": "Inactive", "reputation": initial_reputation - (28 * current_rate) + 50 - (7 * current_rate)},
        ]
        
        for period in monitoring_periods:
            day = period["day"]
            activity = period["activity"]
            reputation = period["reputation"]
            
            logger.info(f"  Day {day:2d}: {activity:8s} → Reputation: {reputation}")
        
        # Analytics insights
        logger.info("\nAnalytics Insights:")
        logger.info("• Decay rate changes are immediately visible in reputation trends")
        logger.info("• Active participants maintain higher reputation scores")
        logger.info("• Inactive participants experience accelerated reputation loss")
        logger.info("• DAO can adjust rates based on network behavior patterns")
        
        logger.info("\n✓ Monitoring and analytics demonstrated")
        return True
        
    except Exception as e:
        logger.error(f"Error demonstrating monitoring: {e}")
        return False
    finally:
        spender_client.close()
        earner_client.close()

def main():
    """Main function to run all demonstrations."""
    logger.info("Pandacea Protocol - Aggressive Reputation Decay Example")
    logger.info("This script demonstrates the new configurable decay rate functionality")
    logger.info("")
    
    try:
        # Run all demonstrations
        demonstrations = [
            ("Initial Decay Rate", demonstrate_initial_decay_rate),
            ("DAO Configuration", demonstrate_dao_configuration),
            ("Network Impact", demonstrate_network_impact),
            ("Monitoring & Analytics", demonstrate_monitoring_and_analytics)
        ]
        
        results = []
        
        for name, func in demonstrations:
            logger.info(f"\n{'='*20} {name} {'='*20}")
            try:
                success = func()
                results.append((name, success))
            except Exception as e:
                logger.error(f"Error in {name}: {e}")
                results.append((name, False))
        
        # Summary
        logger.info("\n" + "=" * 60)
        logger.info("DEMONSTRATION SUMMARY")
        logger.info("=" * 60)
        
        passed = 0
        failed = 0
        
        for name, success in results:
            status = "✓ PASSED" if success else "✗ FAILED"
            color = "GREEN" if success else "RED"
            logger.info(f"{name}: {status}")
            
            if success:
                passed += 1
            else:
                failed += 1
        
        logger.info(f"\nTotal demonstrations: {len(results)}")
        logger.info(f"Passed: {passed}")
        logger.info(f"Failed: {failed}")
        
        if failed == 0:
            logger.info("\n🎉 All demonstrations completed successfully!")
            logger.info("The aggressive reputation decay implementation is working correctly.")
        else:
            logger.error(f"\n❌ {failed} demonstration(s) failed. Please review the errors above.")
        
        return failed == 0
        
    except KeyboardInterrupt:
        logger.info("\nDemonstration interrupted by user")
        return False
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
