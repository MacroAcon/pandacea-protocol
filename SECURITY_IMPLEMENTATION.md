# Pandacea Protocol Security Implementation

## Overview

This document describes the implementation of cryptographic authentication for the Pandacea Protocol MVP, addressing the critical security vulnerability where API requests were not being authenticated.

## What Was Fixed

### Critical Vulnerability
- **Issue**: The original implementation allowed any unauthenticated actor to interact with an agent's API
- **Impact**: This violated the protocol's "Secure by Design" principle and could allow spoofing and unauthorized access
- **Solution**: Implemented cryptographic signing for all agent-to-agent API communication

### Security Requirements Met
- ✅ Each request must be signed by the originating agent's private key
- ✅ Signatures must be included in request headers
- ✅ Only legitimate, identifiable agents can communicate
- ✅ Prevents spoofing and unauthorized access

## Implementation Details

### Go Agent Backend Changes

#### 1. Signature Verification Middleware
**File**: `agent-backend/internal/api/server.go`

- Added `verifySignatureMiddleware` function that:
  - Extracts signature from `X-Pandacea-Signature` header
  - Extracts peer ID from `X-Pandacea-Peer-ID` header
  - Verifies the signature using the peer's public key
  - Returns appropriate HTTP error codes for authentication failures

#### 2. Server Updates
- Modified `NewServer` to accept a P2P node for public key lookup
- Updated `setupRoutes` to apply signature verification to all `/api/v1` routes
- Health check endpoint remains unauthenticated for monitoring purposes

#### 3. Main Application Updates
**File**: `agent-backend/cmd/agent/main.go`
- Updated to pass the P2P node to the API server constructor

### Python SDK Changes

#### 1. Client Authentication
**File**: `builder-sdk/pandacea_sdk/client.py`

- Added `private_key_path` parameter to `PandaceaClient.__init__`
- Implemented `_load_private_key()` method for key loading and peer ID derivation
- Added `_sign_request()` method for cryptographic signing
- Added `_prepare_headers()` method for authenticated request preparation

#### 2. Request Signing
- **GET requests**: Sign canonical representation `"GET /api/v1/products"`
- **POST requests**: Sign the JSON-serialized request body
- All signatures are base64-encoded and included in `X-Pandacea-Signature` header
- Peer ID is included in `X-Pandacea-Peer-ID` header

#### 3. Dependencies
**File**: `builder-sdk/pyproject.toml`
- Added `cryptography` for RSA signing
- Added `py-multibase` and `py-multihash` for peer ID generation

## Usage Examples

### Python SDK Usage

#### Authenticated Client
```python
from pandacea_sdk import PandaceaClient

# Initialize with private key for authentication
client = PandaceaClient(
    base_url="http://localhost:8080",
    private_key_path="~/.pandacea/agent.key",
    timeout=30.0
)

# All requests are now automatically signed
products = client.discover_products()
lease_id = client.request_lease("did:pandacea:product:123", "10.50", "24h")
```

#### Backward Compatibility
```python
# Unauthenticated client (limited functionality)
client = PandaceaClient("http://localhost:8080")
# Note: This will only work if the agent allows unauthenticated requests
```

### Go Agent Usage

The Go agent automatically applies signature verification to all API requests. No changes needed to existing agent code.

## Security Headers

### Request Headers
- `X-Pandacea-Signature`: Base64-encoded RSA signature of request data
- `X-Pandacea-Peer-ID`: Peer ID (base58-encoded multihash of public key)

### Error Responses
- `401 Unauthorized`: Missing signature or peer ID headers
- `403 Forbidden`: Invalid signature or signature verification failure

## Key Management

### Private Key Format
- RSA private keys in PEM format
- No password protection (for simplicity in MVP)
- Recommended location: `~/.pandacea/agent.key`

### Peer ID Generation
- Public key is hashed using SHA2-256
- Multihash is base58-encoded to create peer ID
- Format: `Qm...` (standard libp2p peer ID format)

## Testing

### Python SDK Tests
Run the authentication test suite:
```bash
cd builder-sdk
python test_authentication.py
```

### Integration Testing
1. Start the Go agent backend
2. Use the Python SDK with a private key to make authenticated requests
3. Verify that unauthenticated requests are rejected

## Security Considerations

### Current Implementation
- ✅ Cryptographic signing prevents request tampering
- ✅ Peer ID verification ensures request origin
- ✅ RSA-2048 provides strong cryptographic security
- ✅ Base64 encoding ensures safe header transmission

### Future Enhancements
- Add support for password-protected private keys
- Implement key rotation mechanisms
- Add support for different signature algorithms
- Implement certificate-based authentication
- Add rate limiting and request replay protection

## Migration Guide

### For Existing Users
1. Generate a private key pair for your agent
2. Update SDK initialization to include `private_key_path`
3. Test authentication with the provided examples
4. Remove any hardcoded authentication bypasses

### For Developers
1. Update agent configuration to include private key path
2. Test signature verification with known good/bad signatures
3. Monitor logs for authentication failures
4. Update documentation and examples

## Compliance

This implementation satisfies the security requirements specified in:
- Pandacea Protocol - System Design Document (SDD) (v1.2).md
- Security Threat Model requirements
- "Secure by Design" principle compliance

## Support

For issues with the authentication implementation:
1. Check the test suite output
2. Verify private key format and permissions
3. Review server logs for authentication errors
4. Ensure all required headers are present in requests 