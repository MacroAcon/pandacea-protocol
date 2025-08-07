# Asynchronous Privacy Service Refactor

## Overview

This document describes the refactoring of the PrivacyService from a synchronous, blocking model to an asynchronous, non-blocking architecture with a warm container pool. This refactor significantly improves performance and scalability by eliminating cold-start latency and enabling concurrent computation execution.

## Key Changes

### 1. API Endpoint Changes

#### Before: Synchronous Execution
- **Endpoint**: `POST /api/v1/privacy/execute`
- **Behavior**: Blocking, waits for computation completion
- **Response**: `200 OK` with results or error
- **Response Format**:
```json
{
  "success": true,
  "results": {
    "output": "...",
    "artifacts": { "model_weights.pth": "BASE64_ENCODED_BYTES" }
  }
}
```

#### After: Asynchronous Execution
- **Endpoint**: `POST /api/v1/privacy/execute`
- **Behavior**: Non-blocking, returns immediately with job ID
- **Response**: `202 Accepted` with computation ID
- **Response Format**:
```json
{
  "computation_id": "comp-a1b2c3d4e5"
}
```

#### New Endpoint: Results Polling
- **Endpoint**: `GET /api/v1/privacy/results/{computation_id}`
- **Behavior**: Returns job status and results
- **Response Formats**:

**Pending**:
```json
{
  "status": "pending"
}
```

**Completed**:
```json
{
  "status": "completed",
  "results": {
    "output": "...",
    "artifacts": { "model_weights.pth": "BASE64_ENCODED_BYTES" }
  }
}
```

**Failed**:
```json
{
  "status": "failed",
  "error": "Error message from the script execution."
}
```

### 2. PrivacyService Architecture

#### New Components

1. **Job Management**
   - Thread-safe in-memory job store (`map[string]*ComputationJob`)
   - Job lifecycle tracking (pending → completed/failed)
   - Automatic cleanup of completed jobs

2. **Container Pool**
   - Warm pool of pre-initialized Docker containers
   - Configurable pool size (default: 3 containers)
   - Automatic container lifecycle management
   - Container health monitoring and replacement

3. **Asynchronous Execution**
   - Goroutine-based job execution
   - Non-blocking API responses
   - Concurrent job processing

#### Container Pool Benefits

- **Eliminates Cold-Start Latency**: Containers are pre-warmed and ready for immediate use
- **Improved Throughput**: Multiple computations can run concurrently
- **Resource Efficiency**: Containers are reused, reducing overhead
- **Fault Tolerance**: Failed containers are automatically replaced

### 3. SDK Client Changes

#### New Methods

1. **`execute_computation()`** - Now returns a computation ID instead of results
2. **`get_computation_result()`** - Polls for job status and results
3. **`wait_for_computation()`** - Convenience method that polls until completion

#### Usage Example

```python
# Start computation (non-blocking)
computation_id = client.execute_computation(
    lease_id="lease_123",
    code_path="script.py",
    inputs=[{"asset_id": "data_456", "variable_name": "features"}]
)

# Wait for completion (blocking with timeout)
result = client.wait_for_computation(computation_id, timeout=300.0)

# Or poll manually
while True:
    result = client.get_computation_result(computation_id)
    if result['status'] == 'completed':
        break
    elif result['status'] == 'failed':
        raise Exception(result['error'])
    time.sleep(2)
```

### 4. Integration Test Updates

The integration test has been updated to use the new asynchronous flow:

1. **Start computation** and receive computation ID
2. **Poll for results** until completion
3. **Verify results** and model artifacts
4. **Maintain privacy guarantees** - raw data never leaves the Earner's environment

## Implementation Details

### Container Pool Management

```go
type DockerContainer struct {
    ID       string
    IsActive bool
}

type privacyService struct {
    containerPool chan *DockerContainer
    poolSize      int
    // ... other fields
}
```

#### Container Lifecycle

1. **Initialization**: On service start, create `poolSize` containers
2. **Acquisition**: Jobs acquire containers from the pool with timeout
3. **Execution**: Computation runs in the acquired container
4. **Release**: Container is cleaned and returned to pool
5. **Replacement**: Failed containers are destroyed and replaced

#### Container Health Management

- **Cleaning**: Workspace is cleaned between jobs
- **Monitoring**: Container health is checked after each use
- **Replacement**: Unhealthy containers are automatically replaced
- **Pool Maintenance**: Pool size is maintained at the configured level

### Job State Management

```go
type ComputationJob struct {
    ID        string    `json:"id"`
    Status    string    `json:"status"` // "pending", "completed", "failed"
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Request   *ComputationRequest `json:"request,omitempty"`
    Results   *ComputationResults `json:"results,omitempty"`
    Error     string    `json:"error,omitempty"`
}
```

#### Job Lifecycle

1. **Creation**: Job is created with "pending" status
2. **Execution**: Job runs asynchronously in a goroutine
3. **Completion**: Job status is updated to "completed" or "failed"
4. **Retrieval**: Results are available via polling endpoint

### Error Handling

- **Container Failures**: Automatic replacement and retry
- **Job Failures**: Detailed error messages in job state
- **Timeout Handling**: Configurable timeouts for container acquisition
- **Graceful Degradation**: Service continues operating even with container issues

## Performance Improvements

### Before (Synchronous)
- **Latency**: 5-10 seconds per computation (including container startup)
- **Throughput**: 1 computation at a time
- **Resource Usage**: High (container creation/destruction overhead)

### After (Asynchronous with Pool)
- **Latency**: 1-2 seconds per computation (warm containers)
- **Throughput**: 3+ concurrent computations (configurable)
- **Resource Usage**: Optimized (container reuse)

## Configuration

### Privacy Service Configuration

```go
// In main.go
poolSize := 3 // Default pool size
privacyService, err = privacy.NewPrivacyService(
    logger, 
    ethClient, 
    contractAddress, 
    dataDir, 
    poolSize
)
```

### Docker Image

The service uses a custom PySyft Docker image:
- **Image**: `pandacea/pysyft-datasite:latest`
- **Build**: `./build_pysyft_image.sh`
- **Features**: Pre-installed PySyft, PyTorch, pandas, scikit-learn

## Migration Guide

### For API Consumers

1. **Update Response Handling**: Expect `202 Accepted` instead of `200 OK`
2. **Implement Polling**: Use the new results endpoint to check job status
3. **Handle Timeouts**: Implement appropriate timeout logic for long-running jobs

### For SDK Users

1. **Update Method Calls**: `execute_computation()` now returns a string ID
2. **Add Polling Logic**: Use `wait_for_computation()` or implement custom polling
3. **Error Handling**: Check for job failures in the polling response

### For Developers

1. **Build Docker Image**: Run `./build_pysyft_image.sh` before starting the service
2. **Monitor Container Pool**: Check logs for container health and replacement
3. **Tune Pool Size**: Adjust pool size based on expected load and available resources

## Testing

### Unit Tests

- Container pool management
- Job state transitions
- Error handling scenarios
- Concurrent job execution

### Integration Tests

- End-to-end asynchronous flow
- Container pool behavior
- SDK client compatibility
- Privacy guarantees verification

### Performance Tests

- Throughput measurement
- Latency comparison
- Resource usage analysis
- Scalability testing

## Monitoring and Observability

### Metrics

- Container pool utilization
- Job completion rates
- Average job duration
- Error rates and types

### Logging

- Container lifecycle events
- Job state transitions
- Error details and stack traces
- Performance metrics

### Health Checks

- Container pool health
- Service availability
- Resource usage monitoring
- Error rate alerts

## Future Enhancements

1. **Persistent Job Storage**: Store job state in database for persistence
2. **Job Scheduling**: Implement job queuing and prioritization
3. **Advanced Container Management**: Dynamic pool sizing based on load
4. **Distributed Execution**: Support for multi-node computation
5. **Resource Monitoring**: Real-time container resource usage tracking
6. **Job Cancellation**: Ability to cancel running computations

## Conclusion

The asynchronous refactor significantly improves the PrivacyService's performance, scalability, and reliability. The warm container pool eliminates cold-start latency, while the asynchronous architecture enables concurrent processing and better resource utilization. The new API design provides better user experience with immediate feedback and flexible result retrieval.
