# Resilience Drills Runbook (Windows-first)

This document provides step-by-step instructions for running resilience drills and chaos engineering scenarios on the Pandacea Protocol. These drills help validate system resilience, test recovery mechanisms, and ensure the system can handle various failure modes.

## Prerequisites

- Windows 10/11 with PowerShell
- Docker Desktop for Windows
- Python 3.11+ installed
- Git for Windows
- Access to the Pandacea Protocol repository

## Environment Setup

### 1. Clone and Setup Repository

```powershell
# Clone the repository
git clone https://github.com/your-org/pandacea-protocol.git
cd pandacea-protocol

# Install Python dependencies
cd integration
pip install -r requirements.txt
cd ..
```

### 2. Environment Configuration

Create or update your `.env` file with resilience-specific settings:

```bash
# Backend Resilience Configuration
BACKEND_QUEUE_SIZE=100

# SDK Reliability Configuration
SDK_MAX_RETRIES=5
SDK_BASE_DELAY_MS=100
SDK_CIRCUIT_FAIL_THRESHOLD=10
SDK_CIRCUIT_RESET_SEC=30

# Observability Configuration
LOG_LEVEL=INFO
PANDACEA_OTEL=1
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

### 3. Start Core Services

```powershell
# Start the core services needed for chaos testing
docker-compose up -d anvil ipfs agent-backend

# Wait for services to be ready (30 seconds)
Start-Sleep -Seconds 30

# Verify services are healthy
Write-Host "Checking service health..." -ForegroundColor Cyan

# Check Anvil (EVM RPC)
$anvilResponse = Invoke-RestMethod -Uri "http://localhost:8545" -Method POST -ContentType "application/json" -Body '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -ErrorAction SilentlyContinue
if ($anvilResponse.result) {
    Write-Host "✅ Anvil is healthy" -ForegroundColor Green
} else {
    Write-Host "❌ Anvil is not responding" -ForegroundColor Red
    exit 1
}

# Check IPFS
try {
    $ipfsResponse = Invoke-RestMethod -Uri "http://localhost:5001/api/v0/version" -ErrorAction Stop
    Write-Host "✅ IPFS is healthy" -ForegroundColor Green
} catch {
    Write-Host "❌ IPFS is not responding" -ForegroundColor Red
    exit 1
}

# Check agent-backend
try {
    $backendResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -ErrorAction Stop
    if ($backendResponse.status -eq "healthy") {
        Write-Host "✅ Agent backend is healthy" -ForegroundColor Green
    } else {
        Write-Host "❌ Agent backend is not responding" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Agent backend is not responding" -ForegroundColor Red
    exit 1
}
```

## Chaos Scenarios

### Scenario 1: EVM RPC Flap

**Purpose**: Test system resilience when the EVM RPC (Anvil) becomes intermittently unavailable.

**Expected Behavior**: 
- SDK should implement exponential backoff with jitter
- Circuit breaker should open after threshold failures
- System should recover within 30 seconds after RPC restoration

**RTO Target**: 30 seconds

#### Running the Scenario

```powershell
# Navigate to chaos directory
cd integration/chaos

# Run the EVM RPC flap scenario
python run_chaos.py --scenario evm_rpc_flap

# Check the results
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ EVM RPC flap scenario completed successfully" -ForegroundColor Green
} else {
    Write-Host "⚠️ EVM RPC flap scenario had issues (expected for resilience testing)" -ForegroundColor Yellow
}
```

#### Success Criteria
- [ ] SDK backoff behavior detected (takes >2s due to retries)
- [ ] Circuit breaker opens after threshold failures
- [ ] System recovers within 30 seconds after RPC restoration
- [ ] No data loss during the scenario

#### Monitoring Dashboard
- **Grafana**: `http://localhost:3000` → Pandacea Protocol → EVM RPC Health
- **Key Metrics**: 
  - RPC response time
  - Circuit breaker state
  - Retry attempts
  - Error rates

### Scenario 2: IPFS Slowdown

**Purpose**: Test system resilience when IPFS becomes slow or unavailable.

**Expected Behavior**:
- SDK should implement exponential backoff for IPFS operations
- Upload/download operations should retry with increasing delays
- System should gracefully handle IPFS unavailability

**RTO Target**: 45 seconds

#### Running the Scenario

```powershell
# Run the IPFS slowdown scenario
python run_chaos.py --scenario ipfs_slowdown

# Check the results
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ IPFS slowdown scenario completed successfully" -ForegroundColor Green
} else {
    Write-Host "⚠️ IPFS slowdown scenario had issues (expected for resilience testing)" -ForegroundColor Yellow
}
```

#### Success Criteria
- [ ] SDK backoff behavior detected for IPFS operations
- [ ] Upload/download operations retry with increasing delays
- [ ] System handles IPFS unavailability gracefully
- [ ] No data corruption during the scenario

#### Monitoring Dashboard
- **Grafana**: `http://localhost:3000` → Pandacea Protocol → IPFS Health
- **Key Metrics**:
  - IPFS response time
  - Upload/download success rates
  - Retry attempts
  - Circuit breaker state

### Scenario 3: PySyft Crash

**Purpose**: Test system resilience when the PySyft computation service crashes during a job.

**Expected Behavior**:
- Circuit breaker should open quickly after PySyft failures
- Computation jobs should be retried or queued
- System should recover when PySyft restarts

**RTO Target**: 60 seconds

#### Running the Scenario

```powershell
# Run the PySyft crash scenario
python run_chaos.py --scenario pysyft_crash

# Check the results
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ PySyft crash scenario completed successfully" -ForegroundColor Green
} else {
    Write-Host "⚠️ PySyft crash scenario had issues (expected for resilience testing)" -ForegroundColor Yellow
}
```

#### Success Criteria
- [ ] Circuit breaker opens quickly after PySyft failures
- [ ] Computation jobs are retried or queued
- [ ] System recovers when PySyft restarts
- [ ] No job data loss during the scenario

#### Monitoring Dashboard
- **Grafana**: `http://localhost:3000` → Pandacea Protocol → PySyft Health
- **Key Metrics**:
  - PySyft service status
  - Computation job success rates
  - Circuit breaker state
  - Job queue depth

## Observability and Monitoring

### Telemetry Setup

1. **Start Observability Stack**:
```powershell
# Start OpenTelemetry collector, Prometheus, and Grafana
make obs-up
```

2. **Access Dashboards**:
- **Grafana**: `http://localhost:3000` (admin/admin)
- **Prometheus**: `http://localhost:9091`

### Key Metrics to Monitor

#### System Health Metrics
- Service availability (Anvil, IPFS, Agent Backend)
- Response times and latency
- Error rates and failure counts
- Circuit breaker states

#### Resilience Metrics
- Retry attempts and backoff delays
- Queue depths and capacity utilization
- Recovery times and success rates
- Refused request counts

#### Business Metrics
- Transaction success rates
- Data product availability
- Computation job completion rates
- User experience metrics

### Log Analysis

#### Structured Logs
All resilience events are logged in structured JSON format:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "WARN",
  "event": "request_refused",
  "ip": "192.168.1.100",
  "identity": "user123",
  "route": "/api/v1/train",
  "method": "POST",
  "reason": "queue_full",
  "queue_depth": 100,
  "queue_capacity": 100,
  "rate_limited": false,
  "backpressure": false,
  "trace_id": "abc123def456"
}
```

#### Key Log Events
- `request_refused`: When requests are refused due to queue full, rate limiting, or backpressure
- `circuit_breaker_opened`: When circuit breakers open due to failures
- `retry_attempt`: When SDK retries are attempted
- `recovery_achieved`: When system recovery is detected

## Troubleshooting

### Common Issues

#### 1. Docker Services Not Starting
```powershell
# Check Docker service status
Get-Service docker

# Restart Docker Desktop if needed
# Check available ports
netstat -an | findstr ":8545\|:5001\|:8080"
```

#### 2. Python Dependencies Missing
```powershell
# Install missing dependencies
cd integration
pip install -r requirements.txt

# Verify chaos harness
python -c "import scenarios.evm_rpc_flap; print('Chaos scenarios available')"
```

#### 3. Permission Issues
```powershell
# Run PowerShell as Administrator if needed
# Check file permissions
Get-Acl integration/chaos/run_chaos.py
```

#### 4. Network Connectivity Issues
```powershell
# Test service connectivity
Test-NetConnection -ComputerName localhost -Port 8545
Test-NetConnection -ComputerName localhost -Port 5001
Test-NetConnection -ComputerName localhost -Port 8080
```

### Recovery Procedures

#### 1. Service Recovery
```powershell
# Restart specific services
docker-compose restart anvil
docker-compose restart ipfs
docker-compose restart agent-backend

# Check service health
docker-compose ps
```

#### 2. Circuit Breaker Reset
```powershell
# Circuit breakers reset automatically after the configured timeout
# Check circuit breaker status via logs or metrics
```

#### 3. Queue Recovery
```powershell
# Queue depth resets automatically as requests complete
# Monitor queue metrics in Grafana dashboard
```

## Performance Benchmarks

### Baseline Performance
- **EVM RPC**: <100ms response time
- **IPFS**: <500ms upload/download
- **Agent Backend**: <200ms API response
- **Queue Depth**: <50% capacity under normal load

### Resilience Targets
- **RTO (Recovery Time Objective)**: 30-60 seconds depending on scenario
- **RPO (Recovery Point Objective)**: 0 data loss
- **Circuit Breaker**: Opens after 10 failures, resets after 30 seconds
- **Backoff**: Exponential with jitter, max 5 retries

## Continuous Improvement

### Regular Drills
- Run chaos scenarios weekly during development
- Run daily in CI/CD pipeline
- Document lessons learned and improvements

### Metrics Review
- Review resilience metrics monthly
- Update RTO/RPO targets based on performance
- Optimize circuit breaker and backoff parameters

### Documentation Updates
- Update this runbook based on new scenarios
- Document new failure modes and recovery procedures
- Share lessons learned with the team

## Support and Resources

### Documentation
- [Observability Guide](observability.md)
- [Windows Quickstart](windows_quickstart.md)
- [API Specification](../docs/Pandacea%20Protocol%20-%20API%20Specification%20(v1.1).pdf)

### Tools and Scripts
- Chaos harness: `integration/chaos/run_chaos.py`
- Windows runner: `scripts/windows/run-chaos.ps1`
- Make targets: `chaos-one`, `chaos-test`

### Contact
For questions or issues with resilience drills:
- Create an issue in the repository
- Contact the platform team
- Check the troubleshooting section above
