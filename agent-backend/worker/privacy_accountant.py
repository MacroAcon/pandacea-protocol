#!/usr/bin/env python3
"""
Privacy Accountant for Pandacea Protocol

Tracks per-user epsilon consumption and enforces privacy budgets.
Implements a simple in-memory accountant with persistence to JSON files.
"""

import json
import os
import hashlib
import time
from datetime import datetime
from typing import Dict, Optional, Tuple
import logging

logger = logging.getLogger(__name__)

class PrivacyAccountant:
    """
    Privacy accountant that tracks epsilon consumption per user and enforces budgets.
    """
    
    def __init__(self, state_file: str = "privacy_accountant_state.json", default_budget: float = 10.0):
        """
        Initialize the privacy accountant.
        
        Args:
            state_file: Path to JSON file for persisting state
            default_budget: Default epsilon budget per user
        """
        self.state_file = state_file
        self.default_budget = default_budget
        self.state = self._load_state()
        
    def _load_state(self) -> Dict:
        """Load accountant state from file."""
        if os.path.exists(self.state_file):
            try:
                with open(self.state_file, 'r') as f:
                    state = json.load(f)
                logger.info(f"Loaded privacy accountant state from {self.state_file}")
                return state
            except (json.JSONDecodeError, IOError) as e:
                logger.warning(f"Failed to load state from {self.state_file}: {e}")
                return self._get_default_state()
        else:
            logger.info(f"No existing state file found, starting fresh")
            return self._get_default_state()
    
    def _get_default_state(self) -> Dict:
        """Get default state structure."""
        return {
            "version": "1.0",
            "default_budget": self.default_budget,
            "users": {},
            "last_updated": datetime.utcnow().isoformat()
        }
    
    def _save_state(self):
        """Save current state to file."""
        try:
            self.state["last_updated"] = datetime.utcnow().isoformat()
            with open(self.state_file, 'w') as f:
                json.dump(self.state, f, indent=2)
            logger.debug(f"Saved privacy accountant state to {self.state_file}")
        except IOError as e:
            logger.error(f"Failed to save state to {self.state_file}: {e}")
    
    def get_user_budget(self, user_id: str) -> float:
        """
        Get the current epsilon budget for a user.
        
        Args:
            user_id: Unique identifier for the user
            
        Returns:
            Current available epsilon budget
        """
        if user_id not in self.state["users"]:
            return self.default_budget
        
        user_state = self.state["users"][user_id]
        consumed = user_state.get("consumed_epsilon", 0.0)
        budget = user_state.get("budget", self.default_budget)
        return max(0.0, budget - consumed)
    
    def check_budget(self, user_id: str, required_epsilon: float) -> Tuple[bool, str]:
        """
        Check if a user has sufficient budget for a training job.
        
        Args:
            user_id: Unique identifier for the user
            required_epsilon: Epsilon required for the training job
            
        Returns:
            Tuple of (has_sufficient_budget, error_message)
        """
        available_budget = self.get_user_budget(user_id)
        
        if required_epsilon > available_budget:
            error_msg = (
                f"Insufficient privacy budget for user {user_id}. "
                f"Required: {required_epsilon:.4f}, Available: {available_budget:.4f}"
            )
            logger.warning(error_msg)
            return False, error_msg
        
        return True, ""
    
    def consume_epsilon(self, user_id: str, epsilon: float, job_id: str) -> bool:
        """
        Consume epsilon from a user's budget.
        
        Args:
            user_id: Unique identifier for the user
            epsilon: Amount of epsilon to consume
            job_id: ID of the training job
            
        Returns:
            True if consumption was successful, False otherwise
        """
        if epsilon <= 0:
            logger.warning(f"Attempted to consume non-positive epsilon: {epsilon}")
            return False
        
        # Check if we have sufficient budget
        has_budget, error_msg = self.check_budget(user_id, epsilon)
        if not has_budget:
            return False
        
        # Initialize user state if not exists
        if user_id not in self.state["users"]:
            self.state["users"][user_id] = {
                "budget": self.default_budget,
                "consumed_epsilon": 0.0,
                "jobs": []
            }
        
        # Update user state
        user_state = self.state["users"][user_id]
        user_state["consumed_epsilon"] += epsilon
        
        # Record job details
        job_record = {
            "job_id": job_id,
            "epsilon": epsilon,
            "timestamp": datetime.utcnow().isoformat(),
            "remaining_budget": self.get_user_budget(user_id)
        }
        user_state["jobs"].append(job_record)
        
        # Keep only last 100 jobs to prevent unbounded growth
        if len(user_state["jobs"]) > 100:
            user_state["jobs"] = user_state["jobs"][-100:]
        
        logger.info(
            f"Consumed {epsilon:.4f} epsilon for user {user_id} (job {job_id}). "
            f"Remaining budget: {self.get_user_budget(user_id):.4f}"
        )
        
        # Save state
        self._save_state()
        return True
    
    def set_user_budget(self, user_id: str, budget: float) -> bool:
        """
        Set a custom budget for a specific user.
        
        Args:
            user_id: Unique identifier for the user
            budget: New epsilon budget
            
        Returns:
            True if successful, False otherwise
        """
        if budget < 0:
            logger.error(f"Attempted to set negative budget: {budget}")
            return False
        
        if user_id not in self.state["users"]:
            self.state["users"][user_id] = {
                "budget": budget,
                "consumed_epsilon": 0.0,
                "jobs": []
            }
        else:
            self.state["users"][user_id]["budget"] = budget
        
        logger.info(f"Set budget for user {user_id} to {budget:.4f}")
        self._save_state()
        return True
    
    def reset_user_budget(self, user_id: str) -> bool:
        """
        Reset a user's consumed epsilon to 0.
        
        Args:
            user_id: Unique identifier for the user
            
        Returns:
            True if successful, False otherwise
        """
        if user_id in self.state["users"]:
            self.state["users"][user_id]["consumed_epsilon"] = 0.0
            self.state["users"][user_id]["jobs"] = []
            logger.info(f"Reset budget for user {user_id}")
            self._save_state()
            return True
        else:
            logger.warning(f"User {user_id} not found in accountant state")
            return False
    
    def get_user_history(self, user_id: str) -> Optional[Dict]:
        """
        Get the privacy consumption history for a user.
        
        Args:
            user_id: Unique identifier for the user
            
        Returns:
            User state dictionary or None if user not found
        """
        return self.state["users"].get(user_id)
    
    def get_all_users(self) -> Dict[str, float]:
        """
        Get all users and their remaining budgets.
        
        Returns:
            Dictionary mapping user_id to remaining budget
        """
        return {
            user_id: self.get_user_budget(user_id)
            for user_id in self.state["users"].keys()
        }
    
    def get_stats(self) -> Dict:
        """
        Get overall statistics about the privacy accountant.
        
        Returns:
            Dictionary with statistics
        """
        total_users = len(self.state["users"])
        total_consumed = sum(
            user_state.get("consumed_epsilon", 0.0)
            for user_state in self.state["users"].values()
        )
        total_budget = sum(
            user_state.get("budget", self.default_budget)
            for user_state in self.state["users"].values()
        )
        
        return {
            "total_users": total_users,
            "total_consumed_epsilon": total_consumed,
            "total_budget": total_budget,
            "default_budget": self.default_budget,
            "last_updated": self.state.get("last_updated")
        }
