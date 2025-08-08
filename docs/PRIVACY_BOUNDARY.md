# Privacy Boundary Documentation

This document describes the privacy boundary implementation in the Pandacea Protocol, including local training, differential privacy parameters, consent flow, and artifact integrity verification.

## Overview

The Pandacea Protocol implements a privacy-first approach where:
1. **Local Training**: All training occurs locally on the earner's infrastructure
2. **Differential Privacy**: DP-SGD ensures privacy guarantees
3. **Consent Management**: Explicit consent manifests with integrity verification
4. **Budget Enforcement**: Per-user epsilon budgets prevent privacy budget exhaustion

## Local Training Architecture

### Training Worker
The `train_worker.py` performs all machine learning training locally:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Job Request   │───▶│  Privacy Check   │───▶│  PySyft Training │
│   (JSON stdin)  │    │  (Budget Check)  │    │  (DP-SGD)       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Budget Update   │    │  Artifact Gen   │
                       │  (Accountant)    │    │  (Schema Valid) │
                       └──────────────────┘    └─────────────────┘
```

### Key Components

1. **Job Configuration**: JSON input specifying dataset, task, and DP parameters
2. **Privacy Accountant**: Tracks per-user epsilon consumption and enforces budgets
3. **PySyft Integration**: Real DP-SGD training with synthetic datasets
4. **Mock Fallback**: Development mode when PySyft is unavailable
5. **Artifact Generation**: Schema-validated training artifacts with integrity hashes

## Differential Privacy Implementation

### DP-SGD Parameters

The training worker supports the following differential privacy parameters:

```json
{
  "dp": {
    "enabled": true,
    "epsilon": 1.0,           // Privacy budget (ε)
    "delta": 1e-5,           // Privacy parameter (δ)
    "clip": 1.0,             // Gradient clipping norm
    "noise_multiplier": 0.5  // Noise multiplier for DP-SGD
  }
}
```

### Privacy Guarantees

- **ε (Epsilon)**: Measures privacy loss. Lower values = stronger privacy
- **δ (Delta)**: Probability of privacy failure. Typically set to 1/n where n = dataset size
- **Gradient Clipping**: Bounds sensitivity of gradient updates
- **Noise Addition**: Adds calibrated noise to gradients

### Epsilon Composition

The privacy accountant tracks epsilon consumption across multiple training runs:

```
Total Privacy Loss = Σ(ε₁ + ε₂ + ... + εₙ)
```

Users are blocked from training when their total epsilon consumption exceeds their budget.

## Privacy Accountant

### Budget Management

The privacy accountant enforces per-user epsilon budgets:

```python
# Check if user has sufficient budget
has_budget, error_msg = accountant.check_budget(user_id, required_epsilon)

# Consume epsilon after successful training
accountant.consume_epsilon(user_id, consumed_epsilon, job_id)
```

### State Persistence

Accountant state is persisted to JSON files:
- `privacy_accountant_state.json`: Main state file
- User-specific budgets and consumption history
- Job-level tracking for audit trails

### Budget Enforcement

1. **Pre-training Check**: Verify sufficient budget before training
2. **Post-training Consumption**: Deduct actual epsilon used
3. **Fail-closed**: Reject jobs that would exceed budget
4. **Audit Trail**: Complete history of all privacy consumption

## Running with Docker on Windows

For Windows users, the privacy boundary can be run using Docker containers to avoid PySyft installation issues:

### Prerequisites
- Docker Desktop installed and running
- PowerShell with execution policy allowing scripts

### Quick Setup
```powershell
# Navigate to the project directory
cd C:\Users\thnxt\Documents\pandacea-protocol

# Build the PySyft Docker image
make pysyft-build

# Start the PySyft worker container
make pysyft-up

# Set environment variable for Docker execution
setx USE_DOCKER 1

# Close and reopen PowerShell to load the environment variable
# Then run the demo with real PySyft
make demo-real-docker
```

### Docker Configuration
The Docker setup includes:
- **Dockerfile.pysyft**: Python 3.11 with PySyft dependencies
- **docker-compose.pysyft.yml**: Container orchestration
- **Volume mounting**: Shared data directory for artifacts
- **Environment variables**: MOCK_DP=0 for real PySyft execution

### Troubleshooting
- Ensure Docker Desktop is running before executing commands
- Check container logs: `docker logs pandacea-protocol-pysyft-worker-1`
- Restart containers: `docker compose -f docker-compose.pysyft.yml restart`

## Consent Management

### Consent Manifest Schema

Consent manifests follow a strict schema (`docs/schemas/consent.schema.json`):

```json
{
  "subject_id": "user_123",
  "data_product_id": "product_456",
  "terms": {
    "purpose": "Machine learning model training",
    "retention": "Until consent revocation",
    "revocation": "Contact support@example.com"
  },
  "hash": "sha256_hash_of_manifest"
}
```

### Integrity Verification

1. **Hash Generation**: SHA-256 hash of consent manifest
2. **On-chain Reference**: Hash stored on blockchain in lease flow
3. **Verification**: Hash verified before data processing
4. **Immutable Record**: Hash provides tamper-evident consent record

### Consent Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌─────────────┐
│ User Consent│───▶│ Hash Manifest│───▶│ Store Hash  │───▶│ Verify Hash │
│   (Web UI)  │    │   (SHA-256)  │    │  (On-chain) │    │ (Pre-train) │
└─────────────┘    └──────────────┘    └─────────────┘    └─────────────┘
```

## Artifact Integrity

### Artifact Schema

Training artifacts follow a strict schema (`docs/schemas/artifact.schema.json`):

```json
{
  "job_id": "job_789",
  "model": "base64_encoded_model",
  "epsilon": 1.0,
  "dp": {
    "enabled": true,
    "epsilon": 1.0,
    "clip": 1.0,
    "noise_multiplier": 0.5
  },
  "hash": "sha256_hash_of_artifact"
}
```

### Hash Generation

Artifact hashes are computed deterministically:

```python
def create_artifact_hash(artifact):
    # Remove hash field for hashing
    artifact_copy = artifact.copy()
    del artifact_copy['hash']
    
    # Sort keys for deterministic hashing
    artifact_str = json.dumps(artifact_copy, sort_keys=True)
    return hashlib.sha256(artifact_str.encode()).hexdigest()
```

### Integrity Verification

1. **Pre-hash Validation**: Verify artifact matches schema
2. **Hash Computation**: Generate SHA-256 hash
3. **Post-hash Validation**: Verify hash is present and valid
4. **Storage**: Store artifact with hash for later verification

## Security Considerations

### Data Isolation

- **Local Processing**: All training occurs on earner infrastructure
- **No Data Export**: Raw data never leaves the earner's control
- **Synthetic Datasets**: Training uses synthetic data for privacy
- **Model Only**: Only model artifacts are shared, never raw data

### Privacy Budget Attacks

- **Budget Enforcement**: Strict per-user epsilon limits
- **Rate Limiting**: Prevent rapid successive training attempts
- **Audit Logging**: Complete record of all privacy consumption
- **Fail-closed**: Default to denying access when uncertain

### Consent Verification

- **Hash Verification**: Verify consent manifest integrity
- **On-chain Storage**: Immutable consent records
- **Revocation Support**: Clear process for consent withdrawal
- **Audit Trail**: Complete history of consent changes

## Implementation Details

### Training Worker Usage

```bash
# Run with real PySyft training
echo '{"job_id": "test_123", "dp": {"epsilon": 1.0}}' | \
  python train_worker.py --user-id user_456

# Run with mock training (development)
echo '{"job_id": "test_123", "dp": {"epsilon": 1.0}}' | \
  python train_worker.py --mock --user-id user_456
```

### Privacy Accountant Usage

```python
from privacy_accountant import PrivacyAccountant

# Initialize accountant
accountant = PrivacyAccountant(default_budget=10.0)

# Check budget
has_budget, error = accountant.check_budget("user_123", 1.0)

# Consume budget
success = accountant.consume_epsilon("user_123", 1.0, "job_456")
```

### Schema Validation

```python
import json
from jsonschema import validate

# Load schema
with open('docs/schemas/artifact.schema.json') as f:
    schema = json.load(f)

# Validate artifact
validate(instance=artifact, schema=schema)
```

## Monitoring and Auditing

### Privacy Metrics

- **Epsilon Consumption**: Track per-user privacy budget usage
- **Training Success Rate**: Monitor successful vs failed training jobs
- **Budget Violations**: Alert on attempted budget overruns
- **Consent Verification**: Track consent manifest verification success

### Audit Logs

- **Training Jobs**: Complete record of all training attempts
- **Privacy Consumption**: Detailed epsilon consumption history
- **Consent Changes**: Track all consent manifest modifications
- **Integrity Checks**: Log all hash verification results

### Compliance Reporting

- **Privacy Budget Reports**: Per-user epsilon consumption summaries
- **Consent Compliance**: Verification of consent manifest integrity
- **Training Audits**: Complete training job audit trails
- **Security Incidents**: Logging of privacy boundary violations

## Future Enhancements

### Advanced DP Techniques

- **Moments Accountant**: More precise epsilon composition
- **Adaptive Clipping**: Dynamic gradient clipping based on data
- **Rényi DP**: Alternative privacy measure for better composition
- **Local DP**: Additional privacy guarantees at data level

### Enhanced Consent

- **Granular Consent**: Per-field consent management
- **Temporal Consent**: Time-limited consent with automatic expiration
- **Purpose Limitation**: Purpose-specific consent tracking
- **Third-party Consent**: Consent for data sharing with third parties

### Improved Integrity

- **Multi-signature Artifacts**: Multiple signatures for critical artifacts
- **Blockchain Verification**: On-chain artifact integrity verification
- **Provenance Tracking**: Complete data lineage tracking
- **Tamper Detection**: Advanced tamper detection mechanisms
