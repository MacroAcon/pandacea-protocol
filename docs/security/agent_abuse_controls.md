# Agent Abuse Controls

This document describes the security controls implemented in the Pandacea agent backend to prevent abuse, DoS attacks, and ensure fair resource usage.

## Overview

The agent backend implements multiple layers of security controls:

1. **Rate Limiting** - Prevents request flooding
2. **Authentication** - Identity binding and challenge-response
3. **Concurrency Quotas** - Limits simultaneous operations
4. **Backpressure** - Graceful degradation under load
5. **Request Size Limits** - Prevents large payload attacks
6. **Greylisting/Banning** - Temporary restrictions for violations

## Configuration

Security settings are configured in `agent-backend/config/security.yaml`:

```yaml
rate_limits:
  per_ip_rps: 5          # Requests per second per IP address
  per_identity_rps: 2    # Requests per second per authenticated identity
  burst: 10              # Burst allowance for rate limiting

quotas:
  concurrent_jobs_per_identity: 2  # Maximum concurrent training jobs per identity

backpressure:
  cpu_high_watermark: 85           # CPU usage percentage to trigger backpressure
  mem_high_watermark_mb: 2048      # Memory usage in MB to trigger backpressure

bans:
  greylist_seconds: 600            # Duration to greylist IPs that exceed limits
  temp_ban_seconds: 1800          # Duration to temporarily ban IPs for violations

request_limits:
  max_body_size_mb: 10            # Maximum request body size in MB
  max_header_size_kb: 8           # Maximum header size in KB

auth:
  challenge_timeout_seconds: 300   # Timeout for authentication challenges
  nonce_length: 32                # Length of authentication nonces
```

## Rate Limiting

### Token Bucket Algorithm

The system uses a token bucket algorithm for rate limiting:

- **IP-based**: Each client IP gets a token bucket
- **Identity-based**: Each authenticated identity gets a separate bucket
- **Burst allowance**: Initial burst capacity for legitimate traffic spikes
- **Refill rate**: Tokens refill at the configured RPS rate

### Rate Limit Headers

When rate limits are exceeded, the response includes:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 600
Content-Type: application/json

{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Rate limit exceeded",
    "requestId": "req_123456"
  }
}
```

### Rate Limit Recovery

1. **Greylisting**: IPs that exceed limits are greylisted for 10 minutes
2. **Automatic cleanup**: Expired greylist entries are automatically removed
3. **Progressive backoff**: Repeated violations lead to longer restrictions

## Authentication

### Challenge-Response Flow

1. **Request Challenge**:
   ```http
   POST /api/v1/auth/challenge
   Content-Type: application/json
   
   {
     "address": "0x1234567890123456789012345678901234567890"
   }
   ```

   **Response**:
   ```json
   {
     "nonce": "a1b2c3d4e5f6...",
     "address": "0x1234567890123456789012345678901234567890",
     "expires_at": "2024-01-01T12:00:00Z"
   }
   ```

2. **Sign Challenge**: Client signs the nonce with their private key

3. **Verify Challenge**:
   ```http
   POST /api/v1/auth/verify
   Content-Type: application/json
   
   {
     "nonce": "a1b2c3d4e5f6...",
     "signature": "0xabcd..."
   }
   ```

   **Response**:
   ```json
   {
     "address": "0x1234567890123456789012345678901234567890",
     "valid": true
   }
   ```

### Identity Binding

- Maps libp2p peer keys to on-chain addresses
- Requires signed nonce challenges for sensitive routes
- Prevents identity spoofing and Sybil attacks

## Concurrency Quotas

### Training Job Limits

Each authenticated identity is limited to a maximum number of concurrent training jobs:

- **Default**: 2 concurrent jobs per identity
- **Enforcement**: Applied to `/api/v1/train` endpoint
- **Error Response**: HTTP 409 Conflict with `QUOTA_EXCEEDED` code

### Quota Management

- **Automatic release**: Quotas are released when jobs complete
- **Graceful handling**: Jobs in progress are not affected by quota changes
- **Monitoring**: Quota usage is logged for analysis

## Backpressure

### System Load Monitoring

The system monitors:

- **CPU usage**: Percentage of CPU utilization
- **Memory usage**: Current memory allocation in MB
- **Goroutine count**: Number of active goroutines

### Backpressure Response

When system load exceeds thresholds:

```http
HTTP/1.1 503 Service Unavailable
Retry-After: 30
Content-Type: application/json

{
  "error": {
    "code": "BACKPRESSURE",
    "message": "Service temporarily unavailable due to high load",
    "requestId": "req_123456"
  }
}
```

### Degradation Strategy

1. **Reduced RPS**: Accept 50% fewer requests when under pressure
2. **Priority queuing**: Prioritize authenticated requests
3. **Graceful degradation**: Return 503 with retry advice

## Request Size Limits

### Body Size Limits

- **Maximum**: 10MB per request
- **Enforcement**: Applied to all POST/PUT endpoints
- **Error**: HTTP 413 Request Entity Too Large

### Header Size Limits

- **Maximum**: 8KB total header size
- **Enforcement**: Applied to all requests
- **Error**: HTTP 431 Request Header Fields Too Large

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `RATE_LIMITED` | 429 | Rate limit exceeded |
| `QUOTA_EXCEEDED` | 409 | Concurrency quota exceeded |
| `BACKPRESSURE` | 503 | System under high load |
| `INVALID_REQUEST` | 400 | Invalid request format |
| `MISSING_ADDRESS` | 400 | Address required for challenge |
| `MISSING_FIELDS` | 400 | Required fields missing |
| `CHALLENGE_CREATION_FAILED` | 500 | Failed to create auth challenge |
| `REQUEST_ENTITY_TOO_LARGE` | 413 | Request body too large |

## Security Headers

### Response Headers

- `X-API-Version: v1` - API version identifier
- `X-API-Deprecation-Warning` - Deprecation notices for legacy endpoints
- `Retry-After` - Backoff advice for rate limits and backpressure

### Request Headers

- `X-Signature` - Cryptographic signature for authentication
- `X-Forwarded-For` - Client IP (when behind proxy)
- `X-Real-IP` - Real client IP (when behind proxy)

## Security Event Logging

All security events are logged with structured data:

```json
{
  "ts": "2024-01-01T12:00:00Z",
  "identity": "0x1234...",
  "ip": "192.168.1.1",
  "route": "/api/v1/train",
  "decision": "rate_limited",
  "reason": "IP rate limit exceeded",
  "counters": {
    "ip_rps": 5
  }
}
```

### Event Types

- `rate_limited` - Rate limit enforcement
- `quota_exceeded` - Concurrency quota enforcement
- `backpressure` - System load protection
- `auth_success` - Successful authentication
- `auth_failure` - Failed authentication attempt
- `banned` - IP banned for violations
- `greylisted` - IP greylisted for violations

## Monitoring and Alerting

### Key Metrics

- **Rate limit violations**: Number of 429 responses
- **Quota violations**: Number of 409 responses
- **Backpressure events**: Number of 503 responses
- **Authentication failures**: Failed auth attempts
- **Banned IPs**: Number of currently banned IPs

### Recommended Alerts

- High rate of 429 responses (>10% of requests)
- Sustained backpressure (>5 minutes)
- Unusual authentication failure patterns
- Large number of banned IPs (>100)

## Best Practices

### For Clients

1. **Implement exponential backoff** when receiving 429/503 responses
2. **Respect Retry-After headers** for proper backoff timing
3. **Use authentication** for sensitive operations
4. **Handle large responses** gracefully
5. **Monitor rate limits** and adjust request patterns

### For Operators

1. **Monitor security logs** for unusual patterns
2. **Adjust rate limits** based on legitimate usage patterns
3. **Review banned IPs** regularly for false positives
4. **Scale resources** when backpressure is frequent
5. **Update security config** as usage patterns evolve

## Troubleshooting

### Common Issues

1. **False positives**: Adjust rate limits or whitelist legitimate IPs
2. **Performance impact**: Monitor backpressure and scale resources
3. **Authentication failures**: Check signature generation and timing
4. **Quota exhaustion**: Review concurrent job limits

### Debug Commands

```bash
# Check security service status
curl -X GET http://localhost:8080/api/v1/health

# Test rate limiting
for i in {1..10}; do curl http://localhost:8080/api/v1/products; done

# Test authentication
curl -X POST http://localhost:8080/api/v1/auth/challenge \
  -H "Content-Type: application/json" \
  -d '{"address":"0x1234..."}'
```

## Security Considerations

1. **Token bucket implementation** is in-memory and resets on restart
2. **Authentication challenges** expire after 5 minutes
3. **Greylist/ban durations** are configurable but should be reasonable
4. **Rate limits** should be tuned based on legitimate usage patterns
5. **Backpressure thresholds** should account for normal system load

## Future Enhancements

1. **Persistent rate limiting** with Redis/database backend
2. **Advanced authentication** with JWT tokens
3. **Machine learning** for anomaly detection
4. **Geographic rate limiting** based on IP geolocation
5. **API key management** for different client tiers
