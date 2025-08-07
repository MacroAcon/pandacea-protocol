#!/usr/bin/env python3
"""
Simple test script to verify IPFS integration functionality.
This script tests the basic IPFS upload and CID retrieval functionality.
"""

import os
import tempfile
import sys
from pathlib import Path

# Add the SDK to the path
sys.path.insert(0, str(Path(__file__).parent / "builder-sdk"))

from pandacea_sdk.client import PandaceaClient
from pandacea_sdk.exceptions import PandaceaException


def test_ipfs_upload():
    """Test IPFS upload functionality."""
    print("Testing IPFS upload functionality...")
    
    # Create a temporary test script
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        test_script = '''
import torch
import torch.nn as nn

class TestModel(nn.Module):
    def __init__(self):
        super().__init__()
        self.linear = nn.Linear(10, 1)
    
    def forward(self, x):
        return self.linear(x)

# Test computation
model = TestModel()
print("Model created successfully!")
'''
        f.write(test_script)
        script_path = f.name
    
    try:
        # Initialize client (without connecting to agent for this test)
        client = PandaceaClient("http://localhost:8080")
        
        print(f"Created test script: {script_path}")
        print("Attempting to upload to IPFS...")
        
        # Try to upload to IPFS
        try:
            cid = client.upload_code_to_ipfs(script_path)
            print(f"✅ Successfully uploaded to IPFS with CID: {cid}")
            return True
        except PandaceaException as e:
            print(f"❌ Failed to upload to IPFS: {e}")
            print("This is expected if IPFS is not running locally.")
            return False
        except ImportError as e:
            print(f"❌ IPFS client library not available: {e}")
            print("Please install ipfshttpclient: pip install ipfshttpclient")
            return False
            
    finally:
        # Clean up
        if os.path.exists(script_path):
            os.unlink(script_path)


def test_cid_validation():
    """Test CID format validation."""
    print("\nTesting CID format validation...")
    
    # Test valid CID format
    valid_cid = "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"
    if len(valid_cid) == 46 and valid_cid[0] == 'Q':
        print("✅ Valid CID format check passed")
    else:
        print("❌ Valid CID format check failed")
        return False
    
    # Test invalid CID format
    invalid_cids = [
        "invalid_cid",
        "QmShort",
        "1234567890abcdef",
        ""
    ]
    
    for invalid_cid in invalid_cids:
        if len(invalid_cid) != 46 or invalid_cid[0] != 'Q':
            print(f"✅ Correctly identified invalid CID: {invalid_cid}")
        else:
            print(f"❌ Failed to identify invalid CID: {invalid_cid}")
            return False
    
    return True


def test_api_structure():
    """Test that the API structure is correct."""
    print("\nTesting API structure...")
    
    # Test that the new method exists
    client = PandaceaClient("http://localhost:8080")
    
    # Check if upload_code_to_ipfs method exists
    if hasattr(client, 'upload_code_to_ipfs'):
        print("✅ upload_code_to_ipfs method exists")
    else:
        print("❌ upload_code_to_ipfs method not found")
        return False
    
    # Check if execute_computation method signature is updated
    import inspect
    sig = inspect.signature(client.execute_computation)
    params = list(sig.parameters.keys())
    
    if 'computation_cid' in params:
        print("✅ execute_computation accepts computation_cid parameter")
    else:
        print("❌ execute_computation does not accept computation_cid parameter")
        return False
    
    if 'code_path' not in params:
        print("✅ execute_computation no longer accepts code_path parameter")
    else:
        print("❌ execute_computation still accepts code_path parameter")
        return False
    
    return True


def main():
    """Main test function."""
    print("=" * 60)
    print("PANDACEA PROTOCOL - IPFS INTEGRATION TEST")
    print("=" * 60)
    
    tests = [
        ("API Structure", test_api_structure),
        ("CID Validation", test_cid_validation),
        ("IPFS Upload", test_ipfs_upload),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n--- {test_name} ---")
        try:
            if test_func():
                passed += 1
                print(f"✅ {test_name} PASSED")
            else:
                print(f"❌ {test_name} FAILED")
        except Exception as e:
            print(f"❌ {test_name} FAILED with exception: {e}")
    
    print(f"\n" + "=" * 60)
    print(f"TEST RESULTS: {passed}/{total} tests passed")
    print("=" * 60)
    
    if passed == total:
        print("🎉 All tests passed! IPFS integration is working correctly.")
        return 0
    else:
        print("⚠️  Some tests failed. Please check the output above.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
