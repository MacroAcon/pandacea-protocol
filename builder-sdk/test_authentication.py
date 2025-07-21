#!/usr/bin/env python3
"""
Test script to verify the authentication implementation.

This script tests the cryptographic signing and verification functionality
of the Pandacea SDK.
"""

import base64
import json
import os
import sys
import tempfile

# Add the parent directory to the path so we can import the SDK
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.backends import default_backend
import multibase
import multihash

from pandacea_sdk import PandaceaClient


def generate_test_key_pair():
    """Generate a test RSA key pair for testing."""
    # Generate private key
    private_key = rsa.generate_private_key(
        public_exponent=65537,
        key_size=2048,
        backend=default_backend()
    )
    
    # Get public key
    public_key = private_key.public_key()
    
    return private_key, public_key


def save_private_key(private_key, filepath):
    """Save private key to PEM file."""
    pem = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption()
    )
    
    with open(filepath, 'wb') as f:
        f.write(pem)


def test_peer_id_generation():
    """Test that peer ID generation works correctly."""
    print("Testing peer ID generation...")
    
    private_key, public_key = generate_test_key_pair()
    
    # Extract public key bytes
    public_key_bytes = public_key.public_bytes(
        encoding=serialization.Encoding.DER,
        format=serialization.PublicFormat.SubjectPublicKeyInfo
    )
    
    # Create multihash
    mh = multihash.encode(public_key_bytes, 'sha2-256')
    
    # Create peer ID (base58-encoded multihash without multibase prefix)
    import base58
    peer_id = base58.b58encode(mh).decode('ascii')
    
    print(f"Generated peer ID: {peer_id}")
    print(f"Peer ID length: {len(peer_id)}")
    print(f"Peer ID starts with 'q': {peer_id.startswith('q')}")
    
    assert peer_id.startswith('q'), "Peer ID should start with 'q'"
    print("✅ Peer ID generation test passed\n")
    
    return private_key, public_key, peer_id


def test_client_initialization():
    """Test client initialization with private key."""
    print("Testing client initialization...")
    
    private_key, _ = generate_test_key_pair()
    
    # Create temporary key file
    with tempfile.NamedTemporaryFile(mode='wb', suffix='.pem', delete=False) as f:
        save_private_key(private_key, f.name)
        key_path = f.name
    
    try:
        # Initialize client
        client = PandaceaClient(
            base_url="http://localhost:8080",
            private_key_path=key_path,
            timeout=5.0
        )
        
        print(f"Client initialized successfully")
        print(f"Peer ID: {client.peer_id}")
        print(f"Private key loaded: {client.private_key is not None}")
        
        assert client.peer_id is not None, "Peer ID should be set"
        assert client.private_key is not None, "Private key should be loaded"
        
        print("✅ Client initialization test passed\n")
        
    finally:
        # Clean up
        os.unlink(key_path)


def test_request_signing():
    """Test request signing functionality."""
    print("Testing request signing...")
    
    private_key, public_key = generate_test_key_pair()
    
    # Create temporary key file
    with tempfile.NamedTemporaryFile(mode='wb', suffix='.pem', delete=False) as f:
        save_private_key(private_key, f.name)
        key_path = f.name
    
    try:
        # Initialize client
        client = PandaceaClient(
            base_url="http://localhost:8080",
            private_key_path=key_path,
            timeout=5.0
        )
        
        # Test data to sign
        test_data = b"GET /api/v1/products"
        
        # Sign the data
        signature = client._sign_request(test_data)
        print(f"Generated signature: {signature[:20]}...")
        
        # Verify the signature
        signature_bytes = base64.b64decode(signature)
        try:
            public_key.verify(
                signature_bytes,
                test_data,
                padding.PKCS1v15(),
                hashes.SHA256()
            )
            print("✅ Signature verification successful")
        except Exception as e:
            print(f"❌ Signature verification failed: {e}")
            raise
        
        print("✅ Request signing test passed\n")
        
    finally:
        # Clean up
        os.unlink(key_path)


def test_header_preparation():
    """Test header preparation for authenticated requests."""
    print("Testing header preparation...")
    
    private_key, _ = generate_test_key_pair()
    
    # Create temporary key file
    with tempfile.NamedTemporaryFile(mode='wb', suffix='.pem', delete=False) as f:
        save_private_key(private_key, f.name)
        key_path = f.name
    
    try:
        # Initialize client
        client = PandaceaClient(
            base_url="http://localhost:8080",
            private_key_path=key_path,
            timeout=5.0
        )
        
        # Test data
        test_data = b'{"productId":"test","maxPrice":"10.50","duration":"24h"}'
        
        # Prepare headers
        headers = client._prepare_headers(test_data)
        
        print(f"Generated headers: {headers}")
        
        # Check required headers
        assert 'X-Pandacea-Peer-ID' in headers, "Peer ID header should be present"
        assert 'X-Pandacea-Signature' in headers, "Signature header should be present"
        
        print(f"Peer ID header: {headers['X-Pandacea-Peer-ID']}")
        print(f"Signature header: {headers['X-Pandacea-Signature'][:20]}...")
        
        print("✅ Header preparation test passed\n")
        
    finally:
        # Clean up
        os.unlink(key_path)


def main():
    """Run all authentication tests."""
    print("=== Pandacea SDK Authentication Tests ===\n")
    
    try:
        test_peer_id_generation()
        test_client_initialization()
        test_request_signing()
        test_header_preparation()
        
        print("🎉 All authentication tests passed!")
        print("\nThe authentication implementation is working correctly.")
        print("You can now use the SDK with cryptographic signing for secure agent communication.")
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main() 