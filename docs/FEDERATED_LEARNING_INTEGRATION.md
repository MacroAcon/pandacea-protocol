# Federated Learning Integration with PySyft

This document describes the integration of OpenMined's PySyft into the Pandacea Protocol to enable privacy-preserving federated learning computations.

## Overview

The Pandacea Protocol now includes a **Privacy Layer** that allows data owners (Earners) to participate in federated learning without ever sharing their raw data. Instead, computations are performed in isolated, sandboxed environments, and only the results (such as model weights) are returned to the data consumer (Spender).

## Architecture

### Privacy Service

The core of the federated learning integration is the `PrivacyService` in the Go backend:

- **Location**: `agent-backend/internal/privacy/service.go`
- **Interface**: `PrivacyService` with methods for executing computations and verifying leases
- **Implementation**: `privacyService` that manages PySyft Datasite lifecycle

### Key Components

1. **PrivacyService**: Manages the lifecycle of PySyft Datasites
2. **Docker Sandboxing**: Isolated execution environment for security
3. **Lease Verification**: Blockchain-based authorization for computations
4. **Artifact Management**: Secure handling of computation results

## API Endpoint

### POST /api/v1/privacy/execute

Executes a privacy-preserving computation on an Earner's data.

#### Request Body

```json
{
  "lease_id": "0x123abc...",
  "code": "import torch\n...",
  "inputs": [
    {"asset_id": "earner-data-asset-123", "variable_name": "features"},
    {"asset_id": "earner-data-asset-456", "variable_name": "labels"}
  ]
}
```

#### Response Body

```json
{
  "success": true,
  "results": {
    "output": "...",
    "artifacts": {
      "model_weights.pth": "BASE64_ENCODED_BYTES"
    }
  }
}
```

## SDK Integration

The Python SDK includes new methods for federated learning:

### execute_computation()

```python
def execute_computation(self, lease_id: str, code_path: str, inputs: list[dict]) -> dict:
    """
    Executes a privacy-preserving computation on an Earner's agent.
    
    Args:
        lease_id: The on-chain ID of the approved data lease
        code_path: The local file path to the Python script to execute
        inputs: A list of dicts specifying the data assets and their variable names
    
    Returns:
        A dictionary containing the computation results
    """
```

### decode_artifact()

```python
def decode_artifact(self, encoded_artifact: str) -> bytes:
    """
    Decode a base64-encoded artifact back into bytes.
    
    Args:
        encoded_artifact: Base64-encoded artifact string
    
    Returns:
        The decoded bytes
    """
```

## Usage Example

### 1. Basic Federated Learning

```python
from pandacea_sdk.client import PandaceaClient

# Create client
client = PandaceaClient("http://localhost:8080")

# Define computation inputs
inputs = [
    {"asset_id": "earner-data-asset-123", "variable_name": "features"},
    {"asset_id": "earner-data-asset-456", "variable_name": "labels"}
]

# Execute federated learning
result = client.execute_computation(
    lease_id="0x123abc...",
    code_path="federated_learning_script.py",
    inputs=inputs
)

# Decode model weights
if result['success'] and 'model_weights.pth' in result['results']['artifacts']:
    weights_bytes = client.decode_artifact(
        result['results']['artifacts']['model_weights.pth']
    )
    # Load model weights into PyTorch model
    model.load_state_dict(torch.load(io.BytesIO(weights_bytes)))
```

### 2. PyTorch Federated Learning Script

```python
# federated_learning_script.py
import torch
import torch.nn as nn
import torch.optim as optim

# Load data (provided by the privacy service)
features_tensor = torch.tensor(features.values, dtype=torch.float32)
labels_tensor = torch.tensor(labels.values, dtype=torch.float32).squeeze()

# Create model
model = nn.Linear(features_tensor.shape[1], 1)
criterion = nn.MSELoss()
optimizer = optim.SGD(model.parameters(), lr=0.01)

# Training loop
for epoch in range(10):
    outputs = model(features_tensor)
    loss = criterion(outputs.squeeze(), labels_tensor)
    
    optimizer.zero_grad()
    loss.backward()
    optimizer.step()

# Save model weights
torch.save(model.state_dict(), '/workspace/model_weights.pth')
```

## Security Features

### 1. Input Validation

- **Code Size Limits**: Maximum 100KB for computation scripts
- **Dangerous Import Detection**: Blocks imports of `os`, `subprocess`, `sys`, etc.
- **Lease Verification**: Validates blockchain-based authorization

### 2. Sandboxed Execution

- **Docker Isolation**: Each computation runs in an isolated container
- **Network Isolation**: Containers have no network access
- **Resource Limits**: Memory and CPU limits prevent abuse
- **Non-root User**: Containers run as non-privileged user

### 3. Data Privacy

- **No Raw Data Transfer**: Data never leaves the Earner's environment
- **Temporary Containers**: Containers are destroyed after each computation
- **Artifact Encoding**: Results are base64-encoded for secure transfer

## Docker Setup

### PySyft Datasite Image

The integration includes a custom Docker image for PySyft computations:

```dockerfile
# agent-backend/Dockerfile.pysyft
FROM python:3.9-slim

# Install PySyft and dependencies
RUN pip install syft==0.5.0 torch==1.13.1 pandas==1.5.3

# Create non-root user
RUN useradd --create-home --shell /bin/bash pysyft
USER pysyft

# Set up workspace
WORKDIR /workspace
```

### Building the Image

```bash
cd agent-backend
docker build -f Dockerfile.pysyft -t pandacea/pysyft-datasite:latest .
```

## Testing

### Integration Test

Run the comprehensive federated learning test:

```bash
cd integration
python test_federated_learning.py
```

### Example Script

Run the SDK example:

```bash
cd builder-sdk
python examples/federated_learning_example.py
```

## Configuration

### Environment Variables

The privacy service can be configured with these environment variables:

- `RPC_URL`: Ethereum RPC endpoint for lease verification
- `CONTRACT_ADDRESS`: Address of the LeaseAgreement contract
- `DATA_DIR`: Directory containing data assets (default: `./data`)

### Agent Configuration

Add privacy service configuration to your agent:

```yaml
# config.yaml
blockchain:
  rpc_url: "http://localhost:8545"
  contract_address: "0x..."

privacy:
  data_dir: "./data"
  docker_image: "pandacea/pysyft-datasite:latest"
```

## Limitations and Considerations

### Current Limitations

1. **Single Node**: Currently supports single-node federated learning
2. **PyTorch Only**: Limited to PyTorch-based computations
3. **Synchronous**: Computations are synchronous (no async support yet)
4. **Resource Constraints**: Limited by Docker container resources

### Future Enhancements

1. **Multi-Node Federated Learning**: Support for distributed training
2. **Additional Frameworks**: Support for TensorFlow, scikit-learn, etc.
3. **Asynchronous Execution**: Background computation support
4. **Advanced Privacy**: Differential privacy and secure multi-party computation
5. **Model Aggregation**: Built-in federated averaging algorithms

## Troubleshooting

### Common Issues

1. **Docker Not Available**: Ensure Docker is installed and running
2. **PySyft Image Missing**: Build the PySyft image before running computations
3. **Lease Verification Failed**: Check blockchain connection and contract address
4. **Data Assets Not Found**: Verify data files exist in the configured data directory

### Debug Mode

Enable debug logging for the privacy service:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

## Contributing

To contribute to the federated learning integration:

1. Follow the existing code patterns in `agent-backend/internal/privacy/`
2. Add tests for new functionality
3. Update documentation for any API changes
4. Ensure security best practices are followed

## References

- [OpenMined PySyft Documentation](https://docs.openmined.org/)
- [Federated Learning Overview](https://ai.googleblog.com/2017/04/federated-learning-collaborative.html)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/) 