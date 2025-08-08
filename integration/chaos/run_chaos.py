#!/usr/bin/env python3
"""
Chaos Engineering Harness for Pandacea Protocol
Runs fault injection scenarios and asserts observable recovery.
"""

import os
import sys
import time
import argparse
import logging
import importlib.util
from pathlib import Path
from typing import Optional, Dict, Any

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class ChaosHarness:
    """Minimal chaos engineering harness for Pandacea Protocol."""
    
    def __init__(self, scenarios_dir: str = "scenarios"):
        self.scenarios_dir = Path(scenarios_dir)
        self.scenarios_dir.mkdir(exist_ok=True)
    
    def load_scenario(self, scenario_name: str) -> Optional[Any]:
        """Load a scenario module by name."""
        scenario_path = self.scenarios_dir / f"{scenario_name}.py"
        
        if not scenario_path.exists():
            logger.error(f"Scenario {scenario_name} not found at {scenario_path}")
            return None
        
        try:
            spec = importlib.util.spec_from_file_location(scenario_name, scenario_path)
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            return module
        except Exception as e:
            logger.error(f"Failed to load scenario {scenario_name}: {e}")
            return None
    
    def run_scenario(self, scenario_name: str) -> bool:
        """Run a single scenario and assert recovery."""
        logger.info(f"Running chaos scenario: {scenario_name}")
        
        # Load scenario module
        scenario_module = self.load_scenario(scenario_name)
        if not scenario_module:
            return False
        
        # Check if scenario has required interface
        if not hasattr(scenario_module, 'run_scenario'):
            logger.error(f"Scenario {scenario_name} missing required 'run_scenario' function")
            return False
        
        try:
            # Run scenario and capture result
            start_time = time.time()
            result = scenario_module.run_scenario()
            duration = time.time() - start_time
            
            if result:
                logger.info(f"✅ Scenario {scenario_name} completed successfully in {duration:.2f}s")
                return True
            else:
                logger.error(f"❌ Scenario {scenario_name} failed")
                return False
                
        except Exception as e:
            logger.error(f"❌ Scenario {scenario_name} threw exception: {e}")
            return False
    
    def list_scenarios(self) -> list:
        """List available scenarios."""
        scenarios = []
        for scenario_file in self.scenarios_dir.glob("*.py"):
            if scenario_file.name != "__init__.py":
                scenarios.append(scenario_file.stem)
        return scenarios

def main():
    parser = argparse.ArgumentParser(description="Pandacea Protocol Chaos Harness")
    parser.add_argument("--scenario", "-s", 
                       help="Scenario name to run (or set SCENARIO env var)")
    parser.add_argument("--list", "-l", action="store_true",
                       help="List available scenarios")
    parser.add_argument("--scenarios-dir", default="scenarios",
                       help="Directory containing scenario modules")
    
    args = parser.parse_args()
    
    # Initialize harness
    harness = ChaosHarness(args.scenarios_dir)
    
    # List scenarios if requested
    if args.list:
        scenarios = harness.list_scenarios()
        if scenarios:
            print("Available scenarios:")
            for scenario in scenarios:
                print(f"  - {scenario}")
        else:
            print("No scenarios found")
        return 0
    
    # Determine scenario to run
    scenario_name = args.scenario or os.getenv("SCENARIO")
    if not scenario_name:
        logger.error("No scenario specified. Use --scenario or set SCENARIO env var")
        logger.info("Available scenarios:")
        for scenario in harness.list_scenarios():
            logger.info(f"  - {scenario}")
        return 1
    
    # Run scenario
    success = harness.run_scenario(scenario_name)
    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())
