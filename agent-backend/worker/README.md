# Pandacea Protocol Training Worker

This module contains the PySyft-based training worker for performing differential privacy training in the Pandacea Protocol.

## Components

### `train_worker.py`
The main training worker that performs DP-SGD training using PySyft.

**Features:**
- Real PySyft DP-SGD training with synthetic datasets
- Mock training fallback for development
- Privacy budget enforcement
- Schema-validated artifact generation
- Integrity hashing for artifacts

**Usage:**
```bash
# Real PySyft training
echo '{"job_id": "test_123", "dp": {"epsilon": 1.0}}' | \
  python train_worker.py --user-id user_456

# Mock training (development)
echo '{"job_id": "test_123", "dp": {"epsilon": 1.0}}' | \
  python train_worker.py --mock --user-id user_456
```

### `privacy_accountant.py`
Privacy accountant that tracks per-user epsilon consumption and enforces budgets.

**Features:**
- Per-user epsilon budget tracking
- Budget enforcement with fail-closed policy
- Persistent state storage
- Audit trail for all privacy consumption

**Usage:**
```python
from privacy_accountant import PrivacyAccountant

# Initialize accountant
accountant = PrivacyAccountant(default_budget=10.0)

# Check budget
has_budget, error = accountant.check_budget("user_123", 1.0)

# Consume budget
success = accountant.consume_epsilon("user_123", 1.0, "job_456")
```

## Installation

1. Install Python dependencies:
```bash
pip install -r requirements.txt
```

2. For Windows users, PySyft may have compatibility issues. Use Docker as fallback:
```bash
docker run -it pysyft/pysyft:latest
```

## Configuration

### Environment Variables
- `MOCK_DP=1`: Force mock training mode
- `PYTHONPATH`: Add worker directory to Python path

### Privacy Accountant Configuration
- `default_budget`: Default epsilon budget per user (default: 10.0)
- `state_file`: Path to state persistence file (default: privacy_accountant_state.json)

## Job Configuration Schema

Training jobs are configured via JSON input:

```json
{
  "job_id": "unique_job_identifier",
  "dataset": "synthetic",
  "task": "classification",
  "dp": {
    "enabled": true,
    "epsilon": 1.0,
    "delta": 1e-5,
    "clip": 1.0,
    "noise_multiplier": 0.5
  },
  "epochs": 10,
  "batch_size": 32,
  "learning_rate": 0.01,
  "seed": 42
}
```

## Artifact Schema

Training artifacts follow the schema in `docs/schemas/artifact.schema.json`:

```json
{
  "job_id": "job_789",
  "model": "base64_encoded_model",
  "epsilon": 1.0,
  "accuracy": 0.85,
  "n": 1000,
  "dp": {
    "enabled": true,
    "epsilon": 1.0,
    "clip": 1.0,
    "noise_multiplier": 0.5,
    "delta": 1e-5
  },
  "hash": "sha256_hash_of_artifact",
  "seed": 42,
  "created_at": "2024-01-01T00:00:00Z"
}
```

## Privacy Guarantees

The worker implements differential privacy with the following guarantees:

- **ε (Epsilon)**: Privacy budget consumption tracked per user
- **δ (Delta)**: Privacy failure probability (typically 1/n)
- **Gradient Clipping**: Bounds sensitivity of gradient updates
- **Noise Addition**: Calibrated noise added to gradients

## Security Considerations

1. **Local Training**: All training occurs locally on earner infrastructure
2. **Budget Enforcement**: Strict per-user epsilon limits prevent privacy budget exhaustion
3. **Fail-closed**: Default to denying access when uncertain
4. **Audit Trail**: Complete record of all privacy consumption

## Testing

Run the test suite:
```bash
# Test privacy accountant
make test-accountant

# Validate schemas
make validate-schemas

# Run demo
make demo-real
make demo-mock
```

## Troubleshooting

### PySyft Import Errors
If PySyft fails to import:
1. Use mock training mode: `--mock` flag
2. Use Docker: `docker run -it pysyft/pysyft:latest`
3. Check PySyft compatibility with your Python version

### Privacy Budget Errors
If jobs fail due to insufficient budget:
1. Check current budget: `accountant.get_user_budget(user_id)`
2. Reset budget if needed: `accountant.reset_user_budget(user_id)`
3. Increase default budget in configuration

### Schema Validation Errors
If artifacts fail schema validation:
1. Check required fields are present
2. Verify data types match schema
3. Ensure hash is computed correctly

## Development

### Adding New Training Tasks
1. Extend the `PySyftTrainer` class
2. Add new task types to the job configuration
3. Update artifact schema if needed
4. Add tests for new functionality

### Modifying Privacy Parameters
1. Update DP parameter validation
2. Adjust epsilon computation if needed
3. Test with various parameter combinations
4. Update documentation

## Production Deployment

1. **Dependencies**: Pin all versions in requirements.txt
2. **Security**: Run in isolated containers
3. **Monitoring**: Log all privacy consumption
4. **Backup**: Regular backup of privacy accountant state
5. **Updates**: Test thoroughly before updating PySyft versions
