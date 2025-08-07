"""
Test file to demonstrate Bandit security scanning functionality.
This file contains intentional security issues for testing purposes.
"""

import os
import subprocess
import tempfile
import yaml


def test_security_scan_demo():
    """Test function to demonstrate security scanning."""
    # TODO: This is a fake security issue for testing Bandit
    # password = "admin123"  # Bandit should flag this hardcoded password
    
    # TODO: Another fake security issue for testing
    # api_key = "sk-1234567890abcdef"  # Bandit should flag this hardcoded secret
    
    # TODO: Fake SQL injection vulnerability for testing
    # query = f"SELECT * FROM users WHERE id = {user_id}"  # Bandit should flag this
    
    # TODO: Fake subprocess with shell=True for testing
    # result = subprocess.run(f"echo {user_input}", shell=True)  # Bandit should flag this
    
    print("Security scan test completed")


def test_clean_code():
    """Test function that should pass security scanning."""
    # This function should pass security scanning
    config = {
        "host": "localhost",
        "port": 8080,
        "protocol": "http"
    }
    
    # Safe subprocess usage
    result = subprocess.run(["echo", "hello"], capture_output=True, text=True)
    
    # Safe file operations
    with tempfile.NamedTemporaryFile() as f:
        f.write(b"test data")
        f.flush()
    
    assert config["host"] == "localhost"
    assert result.returncode == 0


def test_yaml_safe_load():
    """Test function demonstrating safe YAML loading."""
    # Safe YAML loading
    safe_yaml = """
    host: localhost
    port: 8080
    """
    data = yaml.safe_load(safe_yaml)
    assert data["host"] == "localhost"


if __name__ == "__main__":
    test_security_scan_demo()
    test_clean_code()
    test_yaml_safe_load() 