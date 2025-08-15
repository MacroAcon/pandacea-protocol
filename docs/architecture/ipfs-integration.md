# IPFS Integration for Pandacea Protocol

## Overview

The Pandacea Protocol has been upgraded to support IPFS (InterPlanetary File System) for flexible and complex computation script delivery. This enhancement allows Spenders to submit an IPFS Content ID (CID) that points to their computation script, rather than sending the raw code directly in the API request.

## Key Benefits

1. **Flexibility**: Spenders can upload complex computation scripts to IPFS and reference them by CID
2. **Scalability**: Large computation scripts can be stored and retrieved efficiently
3. **Decentralization**: Computation scripts are stored in a decentralized manner
4. **Versioning**: IPFS CIDs provide immutable references to specific script versions
5. **Security**: Content validation and size limits prevent abuse

## Architecture Changes

### Go Backend (agent-backend)

#### 1. API Endpoint Modification

The `POST /api/v1/privacy/execute` endpoint now accepts `computationCid` instead of `code`:

```json
{
  "lease_id": "0x123abc...",
  "computationCid": "Qm...", // IPFS Content ID
  "inputs": [
    {"asset_id": "earner-data-asset-123", "variable_name": "features"}
  ]
}
```

#### 2. PrivacyService Updates

- **New IPFS Client**: Added HTTP client for IPFS API interactions
- **Content Fetching**: `fetchContentFromIPFS()` method retrieves scripts from IPFS
- **Validation**: Content size limits (1MB) and CID format validation
- **Error Handling**: Robust error handling for IPFS network issues

#### 3. Configuration

Added IPFS configuration to `config.yaml`:

```yaml
ipfs:
  api_url: "http://127.0.0.1:5001"  # IPFS API URL
```

### Python SDK (builder-sdk)

#### 1. Method Signature Changes

The `execute_computation()` method now accepts `computation_cid`:

```python
def execute_computation(self, lease_id: str, computation_cid: str, inputs: list[dict]) -> str:
```

#### 2. New IPFS Helper Method

Added `upload_code_to_ipfs()` method for uploading local files:

```python
def upload_code_to_ipfs(self, file_path: str) -> str:
    """
    Uploads a local file to IPFS and returns its CID.
    """
```

#### 3. Dependencies

Added `ipfshttpclient` dependency to `pyproject.toml`:

```toml
ipfshttpclient = "^0.8.0"
```

## Usage Examples

### Basic Usage

```python
from pandacea_sdk.client import PandaceaClient

# Initialize client
client = PandaceaClient("http://localhost:8080")

# Upload computation script to IPFS
computation_cid = client.upload_code_to_ipfs("my_script.py")

# Execute computation using IPFS CID
computation_id = client.execute_computation(
    lease_id="0x123...",
    computation_cid=computation_cid,
    inputs=[{"asset_id": "data-123", "variable_name": "features"}]
)

# Wait for results
result = client.wait_for_computation(computation_id)
```

### Integration Test

The `integration/test_federated_learning.py` test demonstrates the complete IPFS workflow:

1. Start local IPFS node
2. Upload PyTorch training script to IPFS
3. Execute computation using the returned CID
4. Verify results and privacy guarantees

## Configuration

### Environment Variables

- `IPFS_API_URL`: IPFS API endpoint (default: `http://127.0.0.1:5001`)

### Agent Configuration

```yaml
# agent-backend/config.yaml
ipfs:
  api_url: "http://127.0.0.1:5001"
```

## Security Considerations

### Content Validation

- **Size Limits**: Maximum 1MB per computation script
- **CID Format**: Validates IPFS CID format (46 characters, starts with 'Q')
- **Content Type**: Validates that content is text-based

### Network Security

- **Timeout**: 30-second timeout for IPFS API requests
- **Error Handling**: Graceful degradation when IPFS is unavailable
- **Logging**: Comprehensive logging for debugging and monitoring

## Error Handling

### Common Error Scenarios

1. **IPFS Node Unavailable**: Returns descriptive error message
2. **Invalid CID**: Validates CID format before attempting fetch
3. **Content Too Large**: Rejects scripts exceeding 1MB limit
4. **Network Timeout**: Handles slow IPFS responses gracefully

### Error Response Format

```json
{
  "error": {
    "code": "IPFS_FETCH_ERROR",
    "message": "Failed to fetch content from IPFS: connection refused",
    "requestId": "req-123"
  }
}
```

## Migration Guide

### For Existing Users

1. **Update SDK**: Install new version with IPFS support
2. **Modify Code**: Replace `code_path` parameter with `computation_cid`
3. **Upload Scripts**: Use `upload_code_to_ipfs()` before execution
4. **Test**: Verify functionality with integration tests

### Backward Compatibility

The API maintains backward compatibility during the transition period, but new features require the IPFS-based approach.

## Testing

### Local Development

1. **Install IPFS**: Download and install IPFS CLI
2. **Start IPFS Node**: Run `ipfs daemon` in background
3. **Run Tests**: Execute integration tests with IPFS support

### Test Commands

```bash
# Start IPFS daemon
ipfs daemon &

# Run integration test
python integration/test_federated_learning.py

# Stop IPFS daemon
pkill ipfs
```

## Performance Considerations

### Caching

- **IPFS Content**: Consider caching frequently used scripts
- **CID Resolution**: IPFS provides built-in content addressing
- **Network Latency**: Plan for IPFS network delays

### Optimization

- **Script Size**: Keep computation scripts under 1MB
- **Dependencies**: Minimize external dependencies in scripts
- **Parallel Execution**: Multiple computations can run concurrently

## Future Enhancements

### Planned Features

1. **IPFS Pinning**: Automatic pinning of frequently used scripts
2. **Content Verification**: Cryptographic verification of script integrity
3. **Distributed Storage**: Integration with multiple IPFS nodes
4. **Script Templates**: Pre-built computation templates on IPFS

### Advanced Use Cases

1. **Multi-Script Workflows**: Chaining multiple IPFS scripts
2. **Version Management**: Script versioning and rollback
3. **Collaborative Development**: Shared script repositories
4. **Audit Trail**: Immutable record of script execution

## Troubleshooting

### Common Issues

1. **IPFS Not Running**: Ensure IPFS daemon is started
2. **Network Connectivity**: Check IPFS API endpoint accessibility
3. **CID Not Found**: Verify script was uploaded successfully
4. **Timeout Errors**: Increase timeout for large scripts

### Debug Commands

```bash
# Check IPFS status
ipfs id

# Test IPFS API
curl -X POST "http://127.0.0.1:5001/api/v0/cat?arg=<CID>"

# View IPFS logs
ipfs log tail
```

## Conclusion

The IPFS integration significantly enhances the Pandacea Protocol's flexibility and scalability for complex computations. By leveraging IPFS's decentralized content addressing, the system can handle larger, more sophisticated computation scripts while maintaining security and performance standards.

This upgrade positions Pandacea as a leading platform for privacy-preserving, decentralized data computation with enterprise-grade capabilities.
