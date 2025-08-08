#!/usr/bin/env python3
"""
Privacy Accountant Tests
Tests for the privacy accountant functionality with budget enforcement.
"""

import json
import os
import sys
import tempfile
import unittest
from unittest.mock import patch, MagicMock

# Add the worker directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from privacy_accountant import PrivacyAccountant, ACCOUNTANT_BUDGET_EXCEEDED


class TestPrivacyAccountant(unittest.TestCase):
    """Test cases for the PrivacyAccountant class."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.state_file = os.path.join(self.temp_dir, "privacy_accountant_state.json")
        self.accountant = PrivacyAccountant(self.state_file)
        
        # Set up a test user with 5.0 epsilon budget
        self.test_user = "test_user_123"
        self.accountant.set_user_budget(self.test_user, 5.0)
    
    def tearDown(self):
        """Clean up test fixtures."""
        if os.path.exists(self.state_file):
            os.remove(self.state_file)
        if os.path.exists(self.temp_dir):
            os.rmdir(self.temp_dir)
    
    def test_initial_budget_setup(self):
        """Test that initial budget is set correctly."""
        budget = self.accountant.get_remaining_budget(self.test_user)
        self.assertEqual(budget, 5.0)
    
    def test_successful_job_submissions(self):
        """Test that jobs within budget limits are accepted."""
        # Submit jobs with epsilons [2.0, 2.0, 1.0]
        epsilons = [2.0, 2.0, 1.0]
        
        for i, epsilon in enumerate(epsilons):
            with self.subTest(epsilon=epsilon, job_num=i+1):
                result = self.accountant.check_budget(self.test_user, epsilon)
                self.assertIsNone(result)  # No error should be raised
                
                # Consume the budget
                self.accountant.consume_budget(self.test_user, epsilon)
                
                # Check remaining budget
                remaining = self.accountant.get_remaining_budget(self.test_user)
                expected_remaining = 5.0 - sum(epsilons[:i+1])
                self.assertEqual(remaining, expected_remaining)
    
    def test_budget_exceeded_error(self):
        """Test that exceeding budget raises the correct error."""
        # First, consume most of the budget
        self.accountant.consume_budget(self.test_user, 2.0)
        self.accountant.consume_budget(self.test_user, 2.0)
        
        # Remaining budget should be 1.0
        remaining = self.accountant.get_remaining_budget(self.test_user)
        self.assertEqual(remaining, 1.0)
        
        # Try to submit a job that exceeds the remaining budget
        with self.assertRaises(Exception) as context:
            self.accountant.check_budget(self.test_user, 1.5)
        
        # Check that the error has the correct code
        self.assertIn(ACCOUNTANT_BUDGET_EXCEEDED, str(context.exception))
        
        # Check that the error message is clear
        error_message = str(context.exception)
        self.assertIn("budget exceeded", error_message.lower())
        self.assertIn("1.5", error_message)  # Requested amount
        self.assertIn("1.0", error_message)  # Available amount
    
    def test_exact_budget_consumption(self):
        """Test that exact budget consumption works correctly."""
        # Consume exactly the available budget
        self.accountant.consume_budget(self.test_user, 5.0)
        
        remaining = self.accountant.get_remaining_budget(self.test_user)
        self.assertEqual(remaining, 0.0)
        
        # Any additional request should fail
        with self.assertRaises(Exception) as context:
            self.accountant.check_budget(self.test_user, 0.1)
        
        self.assertIn(ACCOUNTANT_BUDGET_EXCEEDED, str(context.exception))
    
    def test_state_persistence(self):
        """Test that budget state is persisted correctly."""
        # Consume some budget
        self.accountant.consume_budget(self.test_user, 2.0)
        
        # Create a new accountant instance (simulating restart)
        new_accountant = PrivacyAccountant(self.state_file)
        
        # Check that the budget state is preserved
        remaining = new_accountant.get_remaining_budget(self.test_user)
        self.assertEqual(remaining, 3.0)
    
    def test_multiple_users(self):
        """Test that budget tracking works for multiple users."""
        user1 = "user_1"
        user2 = "user_2"
        
        # Set different budgets for each user
        self.accountant.set_user_budget(user1, 3.0)
        self.accountant.set_user_budget(user2, 7.0)
        
        # Consume budget for user1
        self.accountant.consume_budget(user1, 2.0)
        
        # Check that user2's budget is unaffected
        remaining_user2 = self.accountant.get_remaining_budget(user2)
        self.assertEqual(remaining_user2, 7.0)
        
        # Check user1's remaining budget
        remaining_user1 = self.accountant.get_remaining_budget(user1)
        self.assertEqual(remaining_user1, 1.0)
    
    def test_invalid_epsilon_values(self):
        """Test that invalid epsilon values are handled correctly."""
        # Test negative epsilon
        with self.assertRaises(ValueError):
            self.accountant.check_budget(self.test_user, -1.0)
        
        # Test zero epsilon
        with self.assertRaises(ValueError):
            self.accountant.check_budget(self.test_user, 0.0)
        
        # Test very large epsilon
        with self.assertRaises(ValueError):
            self.accountant.check_budget(self.test_user, 1000000.0)
    
    def test_nonexistent_user(self):
        """Test behavior with nonexistent users."""
        nonexistent_user = "nonexistent_user"
        
        # Should return 0 budget for nonexistent user
        budget = self.accountant.get_remaining_budget(nonexistent_user)
        self.assertEqual(budget, 0.0)
        
        # Should fail to consume budget
        with self.assertRaises(Exception) as context:
            self.accountant.check_budget(nonexistent_user, 1.0)
        
        self.assertIn(ACCOUNTANT_BUDGET_EXCEEDED, str(context.exception))
    
    def test_budget_reset(self):
        """Test that budget can be reset for a user."""
        # Consume some budget
        self.accountant.consume_budget(self.test_user, 2.0)
        
        # Reset the budget
        self.accountant.set_user_budget(self.test_user, 5.0)
        
        # Check that budget is reset
        remaining = self.accountant.get_remaining_budget(self.test_user)
        self.assertEqual(remaining, 5.0)
    
    def test_concurrent_access_simulation(self):
        """Test that the accountant handles concurrent access scenarios."""
        # Simulate multiple budget checks and consumptions
        for i in range(10):
            epsilon = 0.5
            try:
                self.accountant.check_budget(self.test_user, epsilon)
                self.accountant.consume_budget(self.test_user, epsilon)
            except Exception as e:
                # Should fail after consuming 5.0 epsilon (10 * 0.5)
                self.assertIn(ACCOUNTANT_BUDGET_EXCEEDED, str(e))
                break
        
        # Check final budget state
        remaining = self.accountant.get_remaining_budget(self.test_user)
        self.assertLessEqual(remaining, 0.0)


class TestPrivacyAccountantIntegration(unittest.TestCase):
    """Integration tests for the privacy accountant."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.state_file = os.path.join(self.temp_dir, "privacy_accountant_state.json")
    
    def tearDown(self):
        """Clean up test fixtures."""
        if os.path.exists(self.state_file):
            os.remove(self.state_file)
        if os.path.exists(self.temp_dir):
            os.rmdir(self.temp_dir)
    
    def test_full_workflow(self):
        """Test the complete workflow from budget setup to exhaustion."""
        accountant = PrivacyAccountant(self.state_file)
        user = "integration_test_user"
        
        # Set initial budget
        accountant.set_user_budget(user, 5.0)
        
        # Submit jobs with epsilons [2.0, 2.0, 1.0] - all should pass
        epsilons = [2.0, 2.0, 1.0]
        for epsilon in epsilons:
            accountant.check_budget(user, epsilon)
            accountant.consume_budget(user, epsilon)
        
        # Next submission should fail
        with self.assertRaises(Exception) as context:
            accountant.check_budget(user, 1.0)
        
        self.assertIn(ACCOUNTANT_BUDGET_EXCEEDED, str(context.exception))
        
        # Verify final state
        remaining = accountant.get_remaining_budget(user)
        self.assertEqual(remaining, 0.0)
    
    def test_error_code_constant(self):
        """Test that the error code constant is defined correctly."""
        from privacy_accountant import ACCOUNTANT_BUDGET_EXCEEDED
        self.assertEqual(ACCOUNTANT_BUDGET_EXCEEDED, "ACCOUNTANT_BUDGET_EXCEEDED")


if __name__ == "__main__":
    unittest.main()
