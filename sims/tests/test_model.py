"""
Unit tests for the Pandacea economic simulation model.
Tests basic invariants and model correctness.
"""

import unittest
import sys
import os
import tempfile
import shutil
from pathlib import Path

# Add the sims directory to the path
sys.path.insert(0, str(Path(__file__).parent.parent))

from engine.model import (
    HonestEarner, Colluder, Griefer, Hoarder,
    EconomicModel, SimulationEngine
)


class TestAgentClasses(unittest.TestCase):
    """Test individual agent classes for basic functionality."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.honest = HonestEarner("honest_1", 1.0, 0.8)
        self.colluder = Colluder("colluder_1", 1.0, 0.6)
        self.griefer = Griefer("griefer_1", 1.0, 0.4)
        self.hoarder = Hoarder("hoarder_1", 1.0, 0.7)
    
    def test_agent_initialization(self):
        """Test that agents initialize with correct attributes."""
        self.assertEqual(self.honest.identity, "honest_1")
        self.assertEqual(self.honest.stake, 1.0)
        self.assertEqual(self.honest.reputation, 0.8)
        self.assertEqual(self.honest.balance, 0.0)
        
        self.assertEqual(self.colluder.identity, "colluder_1")
        self.assertEqual(self.colluder.stake, 1.0)
        self.assertEqual(self.colluder.reputation, 0.6)
    
    def test_agent_balance_never_negative(self):
        """Test that agent balances never go negative."""
        # Try to spend more than balance
        initial_balance = self.honest.balance
        self.honest.spend(100.0)
        self.assertGreaterEqual(self.honest.balance, 0.0)
        
        # Verify balance didn't change if insufficient funds
        self.assertEqual(self.honest.balance, initial_balance)
    
    def test_reputation_bounds(self):
        """Test that reputation stays within [0, 1] bounds."""
        # Test reputation increase
        self.honest.update_reputation(0.5)
        self.assertLessEqual(self.honest.reputation, 1.0)
        
        # Test reputation decrease
        self.honest.update_reputation(-2.0)
        self.assertGreaterEqual(self.honest.reputation, 0.0)
    
    def test_stake_never_negative(self):
        """Test that stake never goes negative."""
        initial_stake = self.honest.stake
        self.honest.adjust_stake(-2.0)
        self.assertGreaterEqual(self.honest.stake, 0.0)
        
        # Verify stake didn't change if would go negative
        self.assertEqual(self.honest.stake, initial_stake)


class TestEconomicModel(unittest.TestCase):
    """Test the economic model for correctness."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.model = EconomicModel()
        self.agents = [
            HonestEarner("honest_1", 1.0, 0.8),
            HonestEarner("honest_2", 1.5, 0.9),
            Colluder("colluder_1", 1.0, 0.6),
            Griefer("griefer_1", 0.5, 0.4),
            Hoarder("hoarder_1", 2.0, 0.7)
        ]
    
    def test_revenue_distribution(self):
        """Test that all revenue is accounted for."""
        total_revenue = 100.0
        distribution = self.model.calculate_revenue_distribution(
            self.agents, total_revenue
        )
        
        # Check that all revenue is distributed
        total_distributed = sum(distribution.values())
        self.assertAlmostEqual(total_distributed, total_revenue, places=2)
    
    def test_stake_weighted_distribution(self):
        """Test that distribution is stake-weighted."""
        # Create agents with different stakes
        agents = [
            HonestEarner("low_stake", 1.0, 0.8),
            HonestEarner("high_stake", 3.0, 0.8)
        ]
        
        total_revenue = 100.0
        distribution = self.model.calculate_revenue_distribution(
            agents, total_revenue
        )
        
        # High stake agent should get more revenue
        self.assertGreater(
            distribution["high_stake"],
            distribution["low_stake"]
        )
    
    def test_reputation_multiplier(self):
        """Test that reputation affects revenue distribution."""
        # Create agents with same stake but different reputation
        agents = [
            HonestEarner("low_rep", 1.0, 0.5),
            HonestEarner("high_rep", 1.0, 0.9)
        ]
        
        total_revenue = 100.0
        distribution = self.model.calculate_revenue_distribution(
            agents, total_revenue
        )
        
        # High reputation agent should get more revenue
        self.assertGreater(
            distribution["high_rep"],
            distribution["low_rep"]
        )
    
    def test_dispute_resolution(self):
        """Test dispute resolution mechanics."""
        honest = HonestEarner("honest", 1.0, 0.8)
        griefer = Griefer("griefer", 1.0, 0.4)
        
        # Simulate a dispute
        dispute_cost = 0.1
        penalty = 0.5
        
        result = self.model.resolve_dispute(
            honest, griefer, dispute_cost, penalty
        )
        
        # Check that costs are applied
        self.assertLess(honest.balance, 0.0)  # Honest pays dispute cost
        self.assertLess(griefer.balance, 0.0)  # Griefer pays penalty


class TestSimulationEngine(unittest.TestCase):
    """Test the simulation engine for correctness."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.engine = SimulationEngine(
            total_agents=10,
            honest_percentage=0.6,
            colluder_percentage=0.3,
            griefer_percentage=0.1,
            hoarder_percentage=0.0,
            output_dir=self.temp_dir
        )
    
    def tearDown(self):
        """Clean up test fixtures."""
        shutil.rmtree(self.temp_dir)
    
    def test_agent_population(self):
        """Test that agent population matches configuration."""
        self.assertEqual(len(self.engine.agents), 10)
        
        # Count agent types
        honest_count = sum(1 for a in self.engine.agents if isinstance(a, HonestEarner))
        colluder_count = sum(1 for a in self.engine.agents if isinstance(a, Colluder))
        griefer_count = sum(1 for a in self.engine.agents if isinstance(a, Griefer))
        
        self.assertEqual(honest_count, 6)  # 60% of 10
        self.assertEqual(colluder_count, 3)  # 30% of 10
        self.assertEqual(griefer_count, 1)  # 10% of 10
    
    def test_epoch_execution(self):
        """Test that epochs execute without errors."""
        # Run a single epoch
        results = self.engine.run_epoch()
        
        # Check that results contain expected keys
        expected_keys = ['total_revenue', 'honest_revenue', 'attack_revenue', 'metrics']
        for key in expected_keys:
            self.assertIn(key, results)
        
        # Check that revenue is non-negative
        self.assertGreaterEqual(results['total_revenue'], 0.0)
        self.assertGreaterEqual(results['honest_revenue'], 0.0)
        self.assertGreaterEqual(results['attack_revenue'], 0.0)
    
    def test_simulation_run(self):
        """Test that full simulation runs without errors."""
        # Run a short simulation
        results = self.engine.run_simulation(epochs=5)
        
        # Check that results contain expected structure
        self.assertIn('epochs', results)
        self.assertIn('final_metrics', results)
        self.assertIn('agent_states', results)
        
        # Check that we have results for each epoch
        self.assertEqual(len(results['epochs']), 5)
    
    def test_metrics_calculation(self):
        """Test that metrics are calculated correctly."""
        # Run a simulation and check metrics
        results = self.engine.run_simulation(epochs=3)
        metrics = results['final_metrics']
        
        # Check that metrics are within expected bounds
        self.assertGreaterEqual(metrics['honest_share'], 0.0)
        self.assertLessEqual(metrics['honest_share'], 1.0)
        self.assertGreaterEqual(metrics['expected_loss'], 0.0)
        self.assertGreaterEqual(metrics['liveness'], 0.0)
        self.assertLessEqual(metrics['liveness'], 1.0)
    
    def test_output_generation(self):
        """Test that simulation outputs are generated."""
        # Run simulation
        results = self.engine.run_simulation(epochs=2)
        
        # Check that output files exist
        output_files = list(Path(self.temp_dir).glob("*.csv"))
        self.assertGreater(len(output_files), 0)
        
        # Check that CSV files are not empty
        for file_path in output_files:
            with open(file_path, 'r') as f:
                content = f.read()
                self.assertGreater(len(content), 0)


class TestModelInvariants(unittest.TestCase):
    """Test that the model maintains important invariants."""
    
    def test_total_wealth_conservation(self):
        """Test that total wealth is conserved (no money created/destroyed)."""
        model = EconomicModel()
        agents = [
            HonestEarner("a1", 1.0, 0.8),
            HonestEarner("a2", 1.0, 0.8),
            Colluder("a3", 1.0, 0.6)
        ]
        
        # Set initial balances
        for agent in agents:
            agent.balance = 10.0
        
        initial_total = sum(agent.balance for agent in agents)
        
        # Run some economic activity
        revenue = 50.0
        distribution = model.calculate_revenue_distribution(agents, revenue)
        
        # Apply distribution
        for agent in agents:
            agent.balance += distribution.get(agent.identity, 0.0)
        
        final_total = sum(agent.balance for agent in agents)
        
        # Total wealth should be conserved
        self.assertAlmostEqual(final_total, initial_total + revenue, places=2)
    
    def test_reputation_bounds_invariant(self):
        """Test that reputation always stays within [0, 1] bounds."""
        engine = SimulationEngine(total_agents=5, honest_percentage=1.0)
        
        # Run simulation with reputation changes
        results = engine.run_simulation(epochs=10)
        
        # Check all agents have valid reputation
        for agent in engine.agents:
            self.assertGreaterEqual(agent.reputation, 0.0)
            self.assertLessEqual(agent.reputation, 1.0)
    
    def test_stake_non_negative_invariant(self):
        """Test that stake never goes negative."""
        engine = SimulationEngine(total_agents=5, honest_percentage=1.0)
        
        # Run simulation
        results = engine.run_simulation(epochs=10)
        
        # Check all agents have non-negative stake
        for agent in engine.agents:
            self.assertGreaterEqual(agent.stake, 0.0)


if __name__ == '__main__':
    unittest.main()
