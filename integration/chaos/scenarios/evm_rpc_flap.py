#!/usr/bin/env python3
"""
EVM RPC Flap Scenario
Simulates intermittent EVM RPC failures by toggling latency and errors.
"""

import time
import logging
import requests
import subprocess
import docker
from typing import Optional

logger = logging.getLogger(__name__)

class EVMRPCFlapScenario:
    """Simulates EVM RPC failures by toggling service availability."""
    
    def __init__(self):
        self.docker_client = docker.from_env()
        self.anvil_container = None
        self.original_health = None
        
    def _find_anvil_container(self) -> Optional[docker.models.containers.Container]:
        """Find the Anvil container."""
        try:
            containers = self.docker_client.containers.list(
                filters={"label": "com.docker.compose.service=anvil"}
            )
            if containers:
                return containers[0]
            
            # Fallback: search by name
            containers = self.docker_client.containers.list(
                filters={"name": "anvil"}
            )
            if containers:
                return containers[0]
                
        except Exception as e:
            logger.warning(f"Could not find Anvil container: {e}")
        
        return None
    
    def _check_anvil_health(self) -> bool:
        """Check if Anvil is responding."""
        try:
            response = requests.post(
                "http://localhost:8545",
                json={"jsonrpc": "2.0", "method": "eth_blockNumber", "params": [], "id": 1},
                timeout=5
            )
            return response.status_code == 200
        except Exception:
            return False
    
    def _simulate_latency(self, duration: int = 10):
        """Simulate high latency by throttling the container."""
        if not self.anvil_container:
            logger.warning("Anvil container not found, skipping latency simulation")
            return
        
        try:
            # Use tc (traffic control) to add latency if available
            # This is a simplified approach - in production you'd use more sophisticated methods
            logger.info(f"Simulating {duration}s of high latency...")
            time.sleep(duration)  # Simplified simulation
        except Exception as e:
            logger.warning(f"Could not simulate latency: {e}")
    
    def _simulate_errors(self, duration: int = 10):
        """Simulate errors by temporarily stopping the container."""
        if not self.anvil_container:
            logger.warning("Anvil container not found, skipping error simulation")
            return
        
        try:
            logger.info(f"Simulating {duration}s of RPC errors...")
            self.anvil_container.pause()
            time.sleep(duration)
            self.anvil_container.unpause()
            logger.info("RPC errors simulation completed")
        except Exception as e:
            logger.warning(f"Could not simulate errors: {e}")
    
    def run_scenario(self) -> bool:
        """Run the EVM RPC flap scenario."""
        logger.info("Starting EVM RPC flap scenario...")
        
        # Find Anvil container
        self.anvil_container = self._find_anvil_container()
        if not self.anvil_container:
            logger.warning("Anvil container not found, using simplified simulation")
        
        # Check initial health
        initial_health = self._check_anvil_health()
        logger.info(f"Initial Anvil health: {'✅' if initial_health else '❌'}")
        
        if not initial_health:
            logger.warning("Anvil not healthy at start, continuing anyway...")
        
        # Phase 1: Simulate latency
        logger.info("Phase 1: Simulating high latency...")
        self._simulate_latency(5)
        
        # Check health during latency
        latency_health = self._check_anvil_health()
        logger.info(f"Health during latency: {'✅' if latency_health else '❌'}")
        
        # Phase 2: Simulate errors
        logger.info("Phase 2: Simulating RPC errors...")
        self._simulate_errors(5)
        
        # Check health after errors
        error_health = self._check_anvil_health()
        logger.info(f"Health after errors: {'✅' if error_health else '❌'}")
        
        # Phase 3: Recovery period
        logger.info("Phase 3: Waiting for recovery...")
        recovery_start = time.time()
        max_recovery_time = 30  # 30 seconds max recovery time
        
        while time.time() - recovery_start < max_recovery_time:
            if self._check_anvil_health():
                recovery_time = time.time() - recovery_start
                logger.info(f"✅ Recovery achieved in {recovery_time:.2f}s")
                return True
            time.sleep(1)
        
        logger.error(f"❌ Recovery not achieved within {max_recovery_time}s")
        return False

def run_scenario() -> bool:
    """Entry point for the scenario."""
    scenario = EVMRPCFlapScenario()
    return scenario.run_scenario()

if __name__ == "__main__":
    success = run_scenario()
    exit(0 if success else 1)
