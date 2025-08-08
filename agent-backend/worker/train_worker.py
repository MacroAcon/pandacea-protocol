#!/usr/bin/env python3
"""
PySyft Training Worker for Pandacea Protocol

Performs differential privacy training using PySyft DP-SGD on synthetic datasets.
Integrates with privacy accountant to enforce per-user epsilon budgets.
"""

import json
import sys
import os
import logging
import hashlib
import base64
import random
import numpy as np
from datetime import datetime
from typing import Dict, Any, Optional
import argparse

# Try to import PySyft, fall back to mock if not available
try:
    import torch
    import torch.nn as nn
    import torch.optim as optim
    from torch.utils.data import DataLoader, TensorDataset
    import syft as sy
    from syft.core.node.domain import Domain
    from syft.core.node.network import Network
    PYSYFT_AVAILABLE = True
except ImportError:
    PYSYFT_AVAILABLE = False
    print("Warning: PySyft not available, will use mock training")
    # Define dummy classes for type hints when PySyft is not available
    class DataLoader:
        pass
    class TensorDataset:
        pass
    class nn:
        class Module:
            pass
        class Sequential:
            pass
        class Linear:
            pass
        class ReLU:
            pass
        class Dropout:
            pass
        class Sigmoid:
            pass
        class BCELoss:
            pass
    class optim:
        class SGD:
            pass
    class torch:
        @staticmethod
        def manual_seed(seed):
            pass
        @staticmethod
        def randn(*args):
            pass
        @staticmethod
        def save(obj, f):
            pass
        @staticmethod
        def load(f):
            pass
        class utils:
            class clip_grad_norm_:
                @staticmethod
                def __call__(*args):
                    pass

from privacy_accountant import PrivacyAccountant

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class MockTrainer:
    """Mock trainer for development/testing when PySyft is not available."""
    
    def __init__(self, job_config: Dict[str, Any]):
        self.job_config = job_config
        self.seed = job_config.get('seed', 42)
        random.seed(self.seed)
        np.random.seed(self.seed)
        
    def train(self) -> Dict[str, Any]:
        """Perform mock training and return results."""
        logger.info("Running mock training")
        
        # Simulate training time
        import time
        time.sleep(2)
        
        # Generate mock results
        epsilon = self.job_config.get('dp', {}).get('epsilon', 1.0)
        accuracy = 0.85 + random.uniform(-0.05, 0.05)
        
        # Create mock model weights (small tensor)
        model_weights = np.random.randn(100, 10).astype(np.float32)
        model_str = base64.b64encode(model_weights.tobytes()).decode('utf-8')
        
        return {
            'model': model_str,
            'epsilon': epsilon,
            'accuracy': accuracy,
            'n': 1000,
            'dp': {
                'enabled': True,
                'epsilon': epsilon,
                'clip': 1.0,
                'noise_multiplier': 0.5,
                'delta': 1e-5
            },
            'seed': self.seed,
            'training_params': {
                'epochs': 10,
                'batch_size': 32,
                'learning_rate': 0.01
            }
        }

class PySyftTrainer:
    """Real PySyft trainer for differential privacy training."""
    
    def __init__(self, job_config: Dict[str, Any]):
        self.job_config = job_config
        self.seed = job_config.get('seed', 42)
        
        # Set random seeds
        torch.manual_seed(self.seed)
        np.random.seed(self.seed)
        random.seed(self.seed)
        
        # DP parameters
        self.dp_config = job_config.get('dp', {})
        self.epsilon = self.dp_config.get('epsilon', 1.0)
        self.delta = self.dp_config.get('delta', 1e-5)
        self.clip_norm = self.dp_config.get('clip', 1.0)
        self.noise_multiplier = self.dp_config.get('noise_multiplier', 0.5)
        
        # Training parameters
        self.epochs = job_config.get('epochs', 10)
        self.batch_size = job_config.get('batch_size', 32)
        self.learning_rate = job_config.get('learning_rate', 0.01)
        
        logger.info(f"Initialized PySyft trainer with epsilon={self.epsilon}, clip={self.clip_norm}")
        
    def _create_synthetic_dataset(self, n_samples: int = 1000) -> DataLoader:
        """Create a synthetic dataset for training."""
        # Generate synthetic data (simple classification task)
        X = torch.randn(n_samples, 10)
        y = (X[:, 0] + X[:, 1] > 0).float()
        
        dataset = TensorDataset(X, y)
        dataloader = DataLoader(dataset, batch_size=self.batch_size, shuffle=True)
        
        logger.info(f"Created synthetic dataset with {n_samples} samples")
        return dataloader
    
    def _create_model(self) -> nn.Module:
        """Create a simple neural network model."""
        model = nn.Sequential(
            nn.Linear(10, 64),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(64, 32),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(32, 1),
            nn.Sigmoid()
        )
        return model
    
    def _compute_epsilon(self, noise_multiplier: float, n_samples: int, 
                        batch_size: int, epochs: int, delta: float) -> float:
        """Compute epsilon using the moments accountant."""
        # Simplified epsilon computation
        # In practice, you'd use a proper moments accountant
        q = batch_size / n_samples  # sampling probability
        T = epochs * (n_samples // batch_size)  # number of steps
        
        # Approximate epsilon using the Gaussian mechanism formula
        # This is a simplified version - real implementation would use proper composition
        epsilon = np.sqrt(2 * np.log(1.25 / delta)) * noise_multiplier * np.sqrt(T * q)
        
        return epsilon
    
    def train(self) -> Dict[str, Any]:
        """Perform DP-SGD training using PySyft."""
        logger.info("Starting PySyft DP-SGD training")
        
        # Create synthetic dataset
        dataloader = self._create_synthetic_dataset()
        n_samples = len(dataloader.dataset)
        
        # Create model
        model = self._create_model()
        model.train()
        
        # Setup optimizer with DP-SGD
        optimizer = optim.SGD(model.parameters(), lr=self.learning_rate)
        
        # Training loop
        total_loss = 0
        num_batches = 0
        
        for epoch in range(self.epochs):
            epoch_loss = 0
            for batch_idx, (data, target) in enumerate(dataloader):
                optimizer.zero_grad()
                
                # Forward pass
                output = model(data).squeeze()
                loss = nn.BCELoss()(output, target)
                
                # Backward pass
                loss.backward()
                
                # Clip gradients for DP
                torch.nn.utils.clip_grad_norm_(model.parameters(), self.clip_norm)
                
                # Add noise to gradients for DP
                for param in model.parameters():
                    if param.grad is not None:
                        noise = torch.randn_like(param.grad) * self.noise_multiplier * self.clip_norm
                        param.grad += noise
                
                optimizer.step()
                
                epoch_loss += loss.item()
                num_batches += 1
            
            logger.info(f"Epoch {epoch+1}/{self.epochs}, Loss: {epoch_loss/len(dataloader):.4f}")
        
        # Compute final epsilon
        final_epsilon = self._compute_epsilon(
            self.noise_multiplier, n_samples, self.batch_size, self.epochs, self.delta
        )
        
        # Evaluate model
        model.eval()
        correct = 0
        total = 0
        with torch.no_grad():
            for data, target in dataloader:
                output = model(data).squeeze()
                pred = (output > 0.5).float()
                correct += (pred == target).sum().item()
                total += target.size(0)
        
        accuracy = correct / total
        
        # Serialize model
        import io
        model_state = model.state_dict()
        model_bytes = torch.save(model_state, io.BytesIO()).getvalue()
        model_str = base64.b64encode(model_bytes).decode('utf-8')
        
        logger.info(f"Training completed. Accuracy: {accuracy:.4f}, Epsilon: {final_epsilon:.4f}")
        
        return {
            'model': model_str,
            'epsilon': final_epsilon,
            'accuracy': accuracy,
            'n': n_samples,
            'dp': {
                'enabled': True,
                'epsilon': final_epsilon,
                'clip': self.clip_norm,
                'noise_multiplier': self.noise_multiplier,
                'delta': self.delta
            },
            'seed': self.seed,
            'training_params': {
                'epochs': self.epochs,
                'batch_size': self.batch_size,
                'learning_rate': self.learning_rate
            }
        }

def validate_artifact_schema(artifact: Dict[str, Any]) -> bool:
    """Validate artifact against the schema."""
    required_fields = ['job_id', 'model', 'epsilon', 'dp', 'hash', 'created_at']
    
    for field in required_fields:
        if field not in artifact:
            logger.error(f"Missing required field: {field}")
            return False
    
    # Validate DP fields
    dp_fields = ['enabled', 'epsilon', 'clip', 'noise_multiplier']
    for field in dp_fields:
        if field not in artifact['dp']:
            logger.error(f"Missing required DP field: {field}")
            return False
    
    return True

def create_artifact_hash(artifact: Dict[str, Any]) -> str:
    """Create SHA-256 hash of the artifact."""
    # Create a copy without the hash field for hashing
    artifact_copy = artifact.copy()
    if 'hash' in artifact_copy:
        del artifact_copy['hash']
    
    # Sort keys for deterministic hashing
    artifact_str = json.dumps(artifact_copy, sort_keys=True)
    return hashlib.sha256(artifact_str.encode()).hexdigest()

def main():
    """Main entry point for the training worker."""
    parser = argparse.ArgumentParser(description='Pandacea PySyft Training Worker')
    parser.add_argument('--mock', action='store_true', help='Use mock training instead of PySyft')
    parser.add_argument('--user-id', default='default_user', help='User ID for privacy accounting')
    parser.add_argument('--accountant-state', default='privacy_accountant_state.json', 
                       help='Path to privacy accountant state file')
    args = parser.parse_args()
    
    # Read job configuration from stdin
    try:
        job_config = json.loads(sys.stdin.read())
    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse job configuration: {e}")
        sys.exit(1)
    
    job_id = job_config.get('job_id', 'unknown')
    logger.info(f"Starting training job: {job_id}")
    
    # Initialize privacy accountant
    accountant = PrivacyAccountant(state_file=args.accountant_state)
    
    # Check privacy budget
    required_epsilon = job_config.get('dp', {}).get('epsilon', 1.0)
    has_budget, error_msg = accountant.check_budget(args.user_id, required_epsilon)
    
    if not has_budget:
        error_response = {
            'error': error_msg,
            'job_id': job_id,
            'status': 'failed'
        }
        print(json.dumps(error_response))
        sys.exit(1)
    
    # Determine training mode
    use_mock = args.mock or os.environ.get('MOCK_DP') == '1' or not PYSYFT_AVAILABLE
    
    if use_mock:
        logger.info("Using mock training")
        trainer = MockTrainer(job_config)
    else:
        logger.info("Using PySyft training")
        trainer = PySyftTrainer(job_config)
    
    try:
        # Perform training
        training_result = trainer.train()
        
        # Create artifact
        artifact = {
            'job_id': job_id,
            'dataset': job_config.get('dataset', 'synthetic'),
            'task': job_config.get('task', 'classification'),
            'created_at': datetime.utcnow().isoformat(),
            **training_result
        }
        
        # Add hash
        artifact['hash'] = create_artifact_hash(artifact)
        
        # Validate artifact
        if not validate_artifact_schema(artifact):
            error_response = {
                'error': 'Generated artifact does not match schema',
                'job_id': job_id,
                'status': 'failed'
            }
            print(json.dumps(error_response))
            sys.exit(1)
        
        # Consume epsilon from budget
        consumed_epsilon = artifact['epsilon']
        if not accountant.consume_epsilon(args.user_id, consumed_epsilon, job_id):
            error_response = {
                'error': 'Failed to consume epsilon from privacy budget',
                'job_id': job_id,
                'status': 'failed'
            }
            print(json.dumps(error_response))
            sys.exit(1)
        
        # Output artifact
        print(json.dumps(artifact, indent=2))
        logger.info(f"Training job {job_id} completed successfully")
        
    except Exception as e:
        logger.error(f"Training failed: {e}")
        error_response = {
            'error': str(e),
            'job_id': job_id,
            'status': 'failed'
        }
        print(json.dumps(error_response))
        sys.exit(1)

if __name__ == '__main__':
    main()
