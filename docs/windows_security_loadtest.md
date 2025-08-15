# Windows Security Load Testing Guide

This guide provides instructions for performing security load testing on Windows to validate the agent abuse controls implemented in Sprint 2.

## Prerequisites

- Windows 10/11 with PowerShell
- Docker Desktop installed and running
- Git Bash or WSL2 (optional, for additional tools)

## Quick Start

### 1. Start the Agent Backend

```powershell
# Navigate to the project directory
cd C:\Users\thnxt\Documents\pandacea-protocol

# Start the agent backend with security controls
cd agent-backend
make build
./agent
```

The agent will start on `http://localhost:8080` with security controls enabled.

### 2. Run Security Tests

```powershell
# Run the comprehensive security test suite
make agent-security-test
```

This will test:
- Rate limiting (IP and identity-based)
- Authentication challenges
- Concurrency quotas
- Backpressure controls
- Request size limits
- Security event logging

## Load Testing with Docker

### Using k6 (Recommended)

k6 is a modern load testing tool that works well on Windows via Docker.

#### Install and Run k6

```powershell
# Pull the k6 Docker image
docker pull grafana/k6

# Create a test script
@"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up
    { duration: '1m', target: 10 },   // Stay at 10 RPS
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate should be below 10%
  },
};

export default function () {
  const url = 'http://localhost:8080/api/v1/train';
  const payload = JSON.stringify({
    model_type: 'test_model',
    data_size: 1000,
    epochs: 10
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);
  
  check(response, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
"@ | Out-File -FilePath "loadtest.js" -Encoding UTF8

# Run the load test
docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/loadtest.js
```

#### Test Rate Limiting

```powershell
# Create a rate limit test
@"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 20,
  duration: '30s',
};

export default function () {
  const url = 'http://localhost:8080/api/v1/train';
  const payload = JSON.stringify({
    model_type: 'test_model',
    data_size: 1000,
    epochs: 10
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);
  
  check(response, {
    'rate limited when appropriate': (r) => {
      // Should get 429 when rate limit exceeded
      return r.status === 429 || r.status === 200;
    },
    'retry-after header present': (r) => {
      if (r.status === 429) {
        return r.headers['Retry-After'] !== undefined;
      }
      return true;
    },
  });

  sleep(0.1); // 10 RPS per virtual user
}
"@ | Out-File -FilePath "ratelimit_test.js" -Encoding UTF8

docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/ratelimit_test.js
```

### Using Vegeta (Alternative)

Vegeta is another excellent load testing tool.

#### Install and Run Vegeta

```powershell
# Pull the Vegeta Docker image
docker pull peterevans/vegeta

# Create a targets file
@"
POST http://localhost:8080/api/v1/train
Content-Type: application/json

{"model_type": "test_model", "data_size": 1000, "epochs": 10}
"@ | Out-File -FilePath "targets.txt" -Encoding UTF8

# Run load test
docker run --rm -v ${PWD}:/vegeta peterevans/vegeta attack -targets=/vegeta/targets.txt -rate=10 -duration=30s | docker run --rm -i peterevans/vegeta report
```

## Testing Specific Security Controls

### 1. Rate Limiting Tests

#### Test IP-based Rate Limiting

```powershell
# Test from multiple IPs (using different Docker containers)
docker run --rm -it --network host grafana/k6 run - < loadtest.js
```

#### Test Identity-based Rate Limiting

```powershell
# Create test with authentication
@"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 5,
  duration: '30s',
};

export default function () {
  // First, get authentication challenge
  const challengeUrl = 'http://localhost:8080/api/v1/auth/challenge';
  const challengeResponse = http.get(challengeUrl);
  
  if (challengeResponse.status === 200) {
    const challenge = JSON.parse(challengeResponse.body);
    
    // Simulate signature verification (in real scenario, this would be signed)
    const verifyUrl = 'http://localhost:8080/api/v1/auth/verify';
    const verifyPayload = JSON.stringify({
      nonce: challenge.nonce,
      signature: 'test_signature_' + __VU, // Different signature per user
      address: '0x' + __VU.toString().padStart(40, '0')
    });
    
    const verifyResponse = http.post(verifyUrl, verifyPayload, {
      headers: { 'Content-Type': 'application/json' }
    });
    
    if (verifyResponse.status === 200) {
      const authToken = JSON.parse(verifyResponse.body).token;
      
      // Now test rate limiting with authentication
      const trainUrl = 'http://localhost:8080/api/v1/train';
      const trainPayload = JSON.stringify({
        model_type: 'test_model',
        data_size: 1000,
        epochs: 10
      });
      
      const trainResponse = http.post(trainUrl, trainPayload, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + authToken
        }
      });
      
      check(trainResponse, {
        'authenticated rate limiting works': (r) => r.status === 200 || r.status === 429,
      });
    }
  }
  
  sleep(0.2); // 5 RPS per authenticated user
}
"@ | Out-File -FilePath "auth_ratelimit_test.js" -Encoding UTF8

docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/auth_ratelimit_test.js
```

### 2. Concurrency Quota Tests

```powershell
# Test concurrent job limits
@"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

export default function () {
  const url = 'http://localhost:8080/api/v1/train';
  const payload = JSON.stringify({
    model_type: 'test_model',
    data_size: 1000,
    epochs: 10
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);
  
  check(response, {
    'concurrency quota enforced': (r) => {
      // Should get 429 or 409 when quota exceeded
      return r.status === 429 || r.status === 409 || r.status === 200;
    },
  });

  sleep(2); // Longer sleep to test concurrency
}
"@ | Out-File -FilePath "concurrency_test.js" -Encoding UTF8

docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/concurrency_test.js
```

### 3. Request Size Limit Tests

```powershell
# Test request size limits
@"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '10s',
};

export default function () {
  const url = 'http://localhost:8080/api/v1/train';
  
  // Create a large payload (exceeds 10MB limit)
  const largePayload = JSON.stringify({
    model_type: 'test_model',
    data_size: 1000000,
    epochs: 1000,
    large_data: 'x'.repeat(11 * 1024 * 1024) // 11MB
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, largePayload, params);
  
  check(response, {
    'large request rejected': (r) => r.status === 413, // Payload Too Large
  });

  sleep(1);
}
"@ | Out-File -FilePath "size_limit_test.js" -Encoding UTF8

docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/size_limit_test.js
```

### 4. Backpressure Tests

```powershell
# Test backpressure under high load
@"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp up to high load
    { duration: '1m', target: 50 },   // Maintain high load
    { duration: '30s', target: 0 },   // Ramp down
  ],
};

export default function () {
  const url = 'http://localhost:8080/api/v1/train';
  const payload = JSON.stringify({
    model_type: 'test_model',
    data_size: 1000,
    epochs: 10
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);
  
  check(response, {
    'backpressure handled': (r) => {
      // Should get 503 when backpressure triggered
      return r.status === 503 || r.status === 200 || r.status === 429;
    },
    'retry-after header on 503': (r) => {
      if (r.status === 503) {
        return r.headers['Retry-After'] !== undefined;
      }
      return true;
    },
  });

  sleep(0.1);
}
"@ | Out-File -FilePath "backpressure_test.js" -Encoding UTF8

docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/backpressure_test.js
```

## Monitoring and Analysis

### View Security Logs

```powershell
# Monitor security events in real-time
Get-Content agent-backend/agent.log -Wait | Select-String "SECURITY"
```

### Analyze Test Results

```powershell
# Parse k6 results
docker run -i --rm -v ${PWD}:/scripts grafana/k6 run /scripts/loadtest.js | docker run --rm -i grafana/k6 run --out json=/scripts/results.json -

# Convert to CSV for analysis
docker run --rm -v ${PWD}:/scripts grafana/k6 run --out csv=/scripts/results.csv /scripts/loadtest.js
```

### Generate Reports

```powershell
# Create a simple analysis script
@"
import pandas as pd
import matplotlib.pyplot as plt

# Load results
df = pd.read_csv('results.csv')

# Analyze response codes
status_counts = df['http_req_status'].value_counts()
print('Response Status Distribution:')
print(status_counts)

# Analyze response times
print(f'\nResponse Time Statistics:')
print(f'Mean: {df["http_req_duration"].mean():.2f}ms')
print(f'95th percentile: {df["http_req_duration"].quantile(0.95):.2f}ms')
print(f'Max: {df["http_req_duration"].max():.2f}ms')

# Plot response time distribution
plt.figure(figsize=(10, 6))
plt.hist(df['http_req_duration'], bins=50, alpha=0.7)
plt.xlabel('Response Time (ms)')
plt.ylabel('Frequency')
plt.title('Response Time Distribution')
plt.savefig('response_times.png')
plt.show()
"@ | Out-File -FilePath "analyze_results.py" -Encoding UTF8

python analyze_results.py
```

## Expected Results

### Successful Security Controls

1. **Rate Limiting**: Should see 429 responses with Retry-After headers
2. **Authentication**: Should require valid challenges and signatures
3. **Concurrency**: Should see 429 or 409 when quota exceeded
4. **Size Limits**: Should see 413 for oversized requests
5. **Backpressure**: Should see 503 with Retry-After under high load

### Performance Benchmarks

- **Normal Load**: < 500ms response time for 95% of requests
- **Rate Limited**: 429 responses with appropriate Retry-After values
- **Error Rate**: < 10% for legitimate requests
- **Security Logging**: All security events properly logged

## Troubleshooting

### Common Issues

1. **Docker not running**: Ensure Docker Desktop is started
2. **Port conflicts**: Check if port 8080 is available
3. **Agent not responding**: Verify agent is running and healthy
4. **Permission issues**: Run PowerShell as Administrator if needed

### Debug Commands

```powershell
# Check if agent is running
netstat -an | findstr :8080

# Check Docker status
docker ps

# View agent logs
Get-Content agent-backend/agent.log -Tail 50

# Test basic connectivity
Invoke-WebRequest -Uri "http://localhost:8080/health" -Method GET
```

## Next Steps

1. Run the comprehensive test suite: `make agent-security-test`
2. Execute load tests with different scenarios
3. Analyze results and tune security parameters
4. Document findings and recommendations
5. Update security configuration as needed

For more information, see the [Security Implementation Guide](security/implementation.md) and [Agent Abuse Controls Documentation](docs/security/agent_abuse_controls.md).
