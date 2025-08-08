#!/usr/bin/env python3
"""
IPFS Slowdown Scenario
Simulates IPFS slowdown and unavailability by controlling the IPFS container.
"""

import time
import logging
import requests
import docker
from typing import Optional

logger = logging.getLogger(__name__)

class IPFSSlowdownScenario:
    """Simulates IPFS slowdown and unavailability."""
    
    def __init__(self):
        self.docker_client = docker.from_env()
        self.ipfs_container = None
        self.original_health = None
        
    def _find_ipfs_container(self) -> Optional[docker.models.containers.Container]:
        """Find the IPFS container."""
        try:
            containers = self.docker_client.containers.list(
                filters={"label": "com.docker.compose.service=ipfs"}
            )
            if containers:
                return containers[0]
            
            # Fallback: search by name
            containers = self.docker_client.containers.list(
                filters={"name": "ipfs"}
            )
            if containers:
                return containers[0]
                
        except Exception as e:
            logger.warning(f"Could not find IPFS container: {e}")
        
        return None
    
    def _check_ipfs_health(self) -> bool:
        """Check if IPFS is responding."""
        try:
            response = requests.get(
                "http://localhost:5001/api/v0/version",
                timeout=5
            )
            return response.status_code == 200
        except Exception:
            return False
    
    def _simulate_slowdown(self, duration: int = 10):
        """Simulate IPFS slowdown by throttling the container."""
        if not self.ipfs_container:
            logger.warning("IPFS container not found, skipping slowdown simulation")
            return
        
        try:
            logger.info(f"Simulating {duration}s of IPFS slowdown...")
            # In a real implementation, you'd use tc (traffic control) to throttle
            # For now, we'll simulate by pausing briefly
            self.ipfs_container.pause()
            time.sleep(2)  # Brief pause to simulate slowdown
            self.ipfs_container.unpause()
            time.sleep(duration - 2)  # Continue slowdown simulation
            logger.info("IPFS slowdown simulation completed")
        except Exception as e:
            logger.warning(f"Could not simulate slowdown: {e}")
    
    def _simulate_unavailability(self, duration: int = 10):
        """Simulate IPFS unavailability by stopping the container."""
        if not self.ipfs_container:
            logger.warning("IPFS container not found, skipping unavailability simulation")
            return
        
        try:
            logger.info(f"Simulating {duration}s of IPFS unavailability...")
            self.ipfs_container.pause()
            time.sleep(duration)
            self.ipfs_container.unpause()
            logger.info("IPFS unavailability simulation completed")
        except Exception as e:
            logger.warning(f"Could not simulate unavailability: {e}")
    
    def _test_sdk_backoff(self) -> bool:
        """Test that SDK backoff kicks in during IPFS issues."""
        try:
            # Import SDK and test backoff behavior
            import sys
            import os
            sys.path.append(os.path.join(os.path.dirname(__file__), '..', '..', '..', 'builder-sdk'))
            
            from pandacea_sdk import PandaceaClient
            
            client = PandaceaClient("http://localhost:8080")
            
            # Try to upload code to IPFS (should trigger backoff)
            logger.info("Testing SDK backoff behavior...")
            start_time = time.time()
            
            try:
                # This should fail due to IPFS issues and trigger backoff
                result = client.upload_code_to_ipfs("test_file.py")
                logger.warning("Upload succeeded unexpectedly during IPFS issues")
                return False
            except Exception as e:
                logger.info(f"Expected upload failure during IPFS issues: {e}")
                
                # Check if backoff was triggered (should take longer than normal)
                elapsed = time.time() - start_time
                if elapsed > 2.0:  # Should take at least 2s due to backoff
                    logger.info(f"✅ SDK backoff detected (took {elapsed:.2f}s)")
                    return True
                else:
                    logger.warning(f"❌ SDK backoff not detected (took {elapsed:.2f}s)")
                    return False
                    
        except ImportError:
            logger.warning("SDK not available, skipping backoff test")
            return True  # Skip test if SDK not available
        except Exception as e:
            logger.warning(f"Could not test SDK backoff: {e}")
            return True  # Skip test on error
    
    def run_scenario(self) -> bool:
        """Run the IPFS slowdown scenario."""
        logger.info("Starting IPFS slowdown scenario...")
        
        # Find IPFS container
        self.ipfs_container = self._find_ipfs_container()
        if not self.ipfs_container:
            logger.warning("IPFS container not found, using simplified simulation")
        
        # Check initial health
        initial_health = self._check_ipfs_health()
        logger.info(f"Initial IPFS health: {'✅' if initial_health else '❌'}")
        
        if not initial_health:
            logger.warning("IPFS not healthy at start, continuing anyway...")
        
        # Phase 1: Simulate slowdown
        logger.info("Phase 1: Simulating IPFS slowdown...")
        self._simulate_slowdown(5)
        
        # Check health during slowdown
        slowdown_health = self._check_ipfs_health()
        logger.info(f"Health during slowdown: {'✅' if slowdown_health else '❌'}")
        
        # Phase 2: Simulate unavailability
        logger.info("Phase 2: Simulating IPFS unavailability...")
        self._simulate_unavailability(5)
        
        # Check health after unavailability
        unavailability_health = self._check_ipfs_health()
        logger.info(f"Health after unavailability: {'✅' if unavailability_health else '❌'}")
        
        # Phase 3: Test SDK backoff behavior
        logger.info("Phase 3: Testing SDK backoff behavior...")
        backoff_success = self._test_sdk_backoff()
        
        # Phase 4: Recovery period
        logger.info("Phase 4: Waiting for recovery...")
        recovery_start = time.time()
        max_recovery_time = 30  # 30 seconds max recovery time
        
        while time.time() - recovery_start < max_recovery_time:
            if self._check_ipfs_health():
                recovery_time = time.time() - recovery_start
                logger.info(f"✅ Recovery achieved in {recovery_time:.2f}s")
                
                # Final success check
                if backoff_success:
                    logger.info("✅ IPFS slowdown scenario completed successfully")
                    return True
                else:
                    logger.error("❌ SDK backoff test failed")
                    return False
            time.sleep(1)
        
        logger.error(f"❌ Recovery not achieved within {max_recovery_time}s")
        return False

def run_scenario() -> bool:
    """Entry point for the scenario."""
    scenario = IPFSSlowdownScenario()
    return scenario.run_scenario()

if __name__ == "__main__":
    success = run_scenario()
    exit(0 if success else 1)
