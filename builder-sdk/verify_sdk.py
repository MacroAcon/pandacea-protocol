#!/usr/bin/env python3
"""
Verification script for the Pandacea SDK implementation.

This script checks that all required components are present and functional.
"""

import os
import sys
import importlib
import subprocess
from pathlib import Path


def check_file_exists(file_path, description):
    """Check if a file exists and print status."""
    if os.path.exists(file_path):
        print(f"✅ {description}: {file_path}")
        return True
    else:
        print(f"❌ {description}: {file_path} (MISSING)")
        return False


def check_import(module_name, description):
    """Check if a module can be imported."""
    try:
        importlib.import_module(module_name)
        print(f"✅ {description}: {module_name}")
        return True
    except ImportError as e:
        print(f"❌ {description}: {module_name} (IMPORT ERROR: {e})")
        return False


def check_package_structure():
    """Check the package structure."""
    print("=== Package Structure Check ===\n")
    
    required_files = [
        ("pyproject.toml", "Poetry configuration"),
        ("README.md", "SDK documentation"),
        ("pandacea_sdk/__init__.py", "Main package init"),
        ("pandacea_sdk/client.py", "Main client implementation"),
        ("pandacea_sdk/models.py", "Data models"),
        ("pandacea_sdk/exceptions.py", "Custom exceptions"),
        ("tests/__init__.py", "Tests package init"),
        ("tests/test_client.py", "Client unit tests"),
        ("tests/test_models.py", "Model unit tests"),
        ("tests/test_exceptions.py", "Exception unit tests"),
        ("examples/basic_usage.py", "Usage example"),
    ]
    
    all_exist = True
    for file_path, description in required_files:
        if not check_file_exists(file_path, description):
            all_exist = False
    
    return all_exist


def check_imports():
    """Check that all modules can be imported."""
    print("\n=== Import Check ===\n")
    
    required_imports = [
        ("pandacea_sdk", "Main SDK package"),
        ("pandacea_sdk.client", "Client module"),
        ("pandacea_sdk.models", "Models module"),
        ("pandacea_sdk.exceptions", "Exceptions module"),
        ("pydantic", "Pydantic dependency"),
        ("requests", "Requests dependency"),
    ]
    
    all_importable = True
    for module_name, description in required_imports:
        if not check_import(module_name, description):
            all_importable = False
    
    return all_importable


def check_sdk_functionality():
    """Check basic SDK functionality."""
    print("\n=== SDK Functionality Check ===\n")
    
    try:
        # Test imports
        from pandacea_sdk import PandaceaClient, DataProduct, PandaceaException, AgentConnectionError, APIResponseError
        print("✅ SDK classes imported successfully")
        
        # Test DataProduct creation
        product = DataProduct(
            product_id="did:pandacea:earner:123/abc-456",
            name="Test Product",
            data_type="RoboticSensorData",
            keywords=["robotics", "sensor"]
        )
        print("✅ DataProduct creation successful")
        
        # Test client initialization
        client = PandaceaClient("http://localhost:8080")
        print("✅ PandaceaClient initialization successful")
        
        # Test client cleanup
        client.close()
        print("✅ Client cleanup successful")
        
        return True
        
    except Exception as e:
        print(f"❌ SDK functionality test failed: {e}")
        return False


def check_tests():
    """Check that tests can be run."""
    print("\n=== Test Check ===\n")
    
    try:
        # Try to run pytest to check if tests are working
        result = subprocess.run(
            [sys.executable, "-m", "pytest", "--collect-only", "-q"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ Tests can be collected successfully")
            return True
        else:
            print(f"❌ Test collection failed: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("❌ Test collection timed out")
        return False
    except FileNotFoundError:
        print("❌ pytest not found")
        return False
    except Exception as e:
        print(f"❌ Test check failed: {e}")
        return False


def check_poetry():
    """Check Poetry configuration."""
    print("\n=== Poetry Configuration Check ===\n")
    
    try:
        # Check if pyproject.toml exists and is valid
        if not os.path.exists("pyproject.toml"):
            print("❌ pyproject.toml not found")
            return False
        
        # Try to validate the Poetry configuration
        result = subprocess.run(
            ["poetry", "check"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ Poetry configuration is valid")
            return True
        else:
            print(f"❌ Poetry configuration error: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("❌ Poetry check timed out")
        return False
    except FileNotFoundError:
        print("⚠️  Poetry not found (this is okay if using pip)")
        return True
    except Exception as e:
        print(f"❌ Poetry check failed: {e}")
        return False


def main():
    """Main verification function."""
    print("=== Pandacea SDK Implementation Verification ===\n")
    
    # Change to the builder-sdk directory
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    
    checks = [
        ("Package Structure", check_package_structure),
        ("Imports", check_imports),
        ("SDK Functionality", check_sdk_functionality),
        ("Tests", check_tests),
        ("Poetry Configuration", check_poetry),
    ]
    
    results = {}
    for check_name, check_func in checks:
        try:
            results[check_name] = check_func()
        except Exception as e:
            print(f"❌ {check_name} check failed with exception: {e}")
            results[check_name] = False
    
    # Summary
    print("\n=== Verification Summary ===")
    all_passed = True
    for check_name, passed in results.items():
        status = "✅ PASS" if passed else "❌ FAIL"
        print(f"{check_name}: {status}")
        if not passed:
            all_passed = False
    
    print()
    if all_passed:
        print("🎉 All checks passed! The SDK implementation is complete and functional.")
    else:
        print("⚠️  Some checks failed. Please review the issues above.")
    
    return all_passed


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1) 