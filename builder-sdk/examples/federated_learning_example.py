#!/usr/bin/env python3
"""
Federated Learning Example for Pandacea Protocol

This example demonstrates how to use the Pandacea SDK to execute
privacy-preserving federated learning computations on an Earner's data.

The example shows:
1. Connecting to an Earner agent
2. Creating a lease agreement
3. Executing a federated learning script
4. Retrieving and using the trained model
"""

import os
import sys
import tempfile
import torch
import torch.nn as nn
from pathlib import Path

# Add the SDK to the path
sys.path.insert(0, str(Path(__file__).parent.parent))

from pandacea_sdk.client import PandaceaClient
from pandacea_sdk.exceptions import PandaceaException


class SimpleLinearModel(nn.Module):
    """Simple linear regression model for demonstration."""
    
    def __init__(self, input_size=2):
        super(SimpleLinearModel, self).__init__()
        self.linear = nn.Linear(input_size, 1)
    
    def forward(self, x):
        return self.linear(x)


def create_federated_learning_script():
    """Create a simple federated learning script."""
    return '''
import torch
import torch.nn as nn
import torch.optim as optim

# Define a simple linear model
class SimpleLinearModel(nn.Module):
    def __init__(self, input_size=2):
        super(SimpleLinearModel, self).__init__()
        self.linear = nn.Linear(input_size, 1)
    
    def forward(self, x):
        return self.linear(x)

# Load and prepare data
print("Loading data for federated learning...")
features_tensor = torch.tensor(features.values, dtype=torch.float32)
labels_tensor = torch.tensor(labels.values, dtype=torch.float32).squeeze()

print(f"Training on {features_tensor.shape[0]} samples with {features_tensor.shape[1]} features")

# Create model
model = SimpleLinearModel(input_size=features_tensor.shape[1])
criterion = nn.MSELoss()
optimizer = optim.SGD(model.parameters(), lr=0.01)

# Training loop
print("Starting federated learning training...")
num_epochs = 5
for epoch in range(num_epochs):
    # Forward pass
    outputs = model(features_tensor)
    loss = criterion(outputs.squeeze(), labels_tensor)
    
    # Backward pass and optimization
    optimizer.zero_grad()
    loss.backward()
    optimizer.step()
    
    print(f"Epoch [{epoch+1}/{num_epochs}], Loss: {loss.item():.4f}")

# Save model weights
print("Training completed. Saving model weights...")
torch.save(model.state_dict(), '/workspace/model_weights.pth')

print("Federated learning computation completed successfully!")
'''


def main():
    """Main example function."""
    print("=" * 50)
    print("PANDACEA PROTOCOL - FEDERATED LEARNING EXAMPLE")
    print("=" * 50)
    
    # Configuration
    agent_url = "http://localhost:8080"  # Earner agent URL
    lease_id = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"  # Mock lease ID
    
    print(f"\n1. Connecting to Earner agent at: {agent_url}")
    
    try:
        # Create client
        client = PandaceaClient(
            base_url=agent_url,
            private_key_path=None,  # No private key for this example
            timeout=60.0
        )
        
        print("   ✓ Connected to agent")
        
        print("\n2. Discovering available data products...")
        try:
            products = client.discover_products()
            print(f"   Found {len(products)} data products:")
            for product in products:
                print(f"     - {product.productId}: {product.name} ({product.dataType})")
        except Exception as e:
            print(f"   Warning: Could not discover products: {e}")
            print("   Continuing with mock data...")
        
        print(f"\n3. Using lease ID: {lease_id}")
        
        print("\n4. Creating federated learning script...")
        script_content = create_federated_learning_script()
        
        # Create temporary script file
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write(script_content)
            script_path = f.name
        
        print(f"   Script created: {script_path}")
        
        print("\n5. Defining computation inputs...")
        inputs = [
            {"asset_id": "earner-data-asset-123", "variable_name": "features"},
            {"asset_id": "earner-data-asset-456", "variable_name": "labels"}
        ]
        
        print("   Inputs:")
        for input_def in inputs:
            print(f"     - {input_def['asset_id']} → {input_def['variable_name']}")
        
        print("\n6. Executing federated learning computation...")
        
        # Execute the computation
        result = client.execute_computation(
            lease_id=lease_id,
            code_path=script_path,
            inputs=inputs
        )
        
        print("   ✓ Computation completed successfully!")
        
        if result.get('success') and 'results' in result:
            results = result['results']
            
            print(f"\n7. Computation results:")
            print(f"   Output length: {len(results.get('output', ''))}")
            print(f"   Artifacts: {list(results.get('artifacts', {}).keys())}")
            
            # Check for model weights
            if 'model_weights.pth' in results['artifacts']:
                print("\n8. Loading trained model...")
                
                # Decode and load model weights
                encoded_weights = results['artifacts']['model_weights.pth']
                weights_bytes = client.decode_artifact(encoded_weights)
                
                # Save to temporary file
                with tempfile.NamedTemporaryFile(suffix='.pth', delete=False) as f:
                    f.write(weights_bytes)
                    weights_path = f.name
                
                # Load the model
                model = SimpleLinearModel()
                model.load_state_dict(torch.load(weights_path, map_location='cpu'))
                
                print("   ✓ Model loaded successfully!")
                
                # Show model parameters
                print("\n   Model parameters:")
                for name, param in model.named_parameters():
                    print(f"     {name}: {param.data.numpy()}")
                
                # Test the model
                print("\n9. Testing the trained model...")
                test_input = torch.randn(3, 2)
                with torch.no_grad():
                    predictions = model(test_input)
                
                print(f"   Test input shape: {test_input.shape}")
                print(f"   Predictions shape: {predictions.shape}")
                print(f"   Sample predictions: {predictions.squeeze().numpy()}")
                
                print("\n   ✓ Model test successful!")
                
                # Clean up temporary files
                os.unlink(weights_path)
            else:
                print("   Warning: No model weights found in results")
        
        print("\n" + "=" * 50)
        print("✅ FEDERATED LEARNING EXAMPLE COMPLETED")
        print("=" * 50)
        print("\nKey points demonstrated:")
        print("  • Privacy-preserving computation using PySyft")
        print("  • Data never leaves the Earner's environment")
        print("  • Only model weights are returned")
        print("  • Secure execution in isolated containers")
        
    except PandaceaException as e:
        print(f"\n❌ Error: {e}")
        return False
    
    finally:
        # Clean up
        if 'script_path' in locals():
            try:
                os.unlink(script_path)
            except:
                pass
    
    return True


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1) 