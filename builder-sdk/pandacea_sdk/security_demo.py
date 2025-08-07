"""
This file contains intentional security issues for testing Bandit.
DO NOT USE IN PRODUCTION - FOR TESTING ONLY
"""

import os
import subprocess
import hashlib
import sqlite3
from cryptography.fernet import Fernet


def demo_security_issues():
    """Demonstrate various security issues that Bandit should detect."""
    
    # SECURITY ISSUE: Hardcoded password (Bandit B105)
    password = "admin123"
    
    # SECURITY ISSUE: Hardcoded API key (Bandit B105)
    api_key = "sk-1234567890abcdef"
    
    # SECURITY ISSUE: Weak crypto algorithm (Bandit B303)
    weak_hash = hashlib.md5(password.encode()).hexdigest()
    
    # SECURITY ISSUE: SQL injection vulnerability (Bandit B608)
    user_id = "123"
    query = f"SELECT * FROM users WHERE id = {user_id}"
    
    # SECURITY ISSUE: Command injection with shell=True (Bandit B602)
    user_input = "malicious_command"
    result = subprocess.run(f"echo {user_input}", shell=True, capture_output=True)
    
    # SECURITY ISSUE: Hardcoded encryption key (Bandit B105)
    key = b"my-secret-key-32-bytes-long!!"
    cipher = Fernet(key)
    
    # SECURITY ISSUE: Unsafe YAML loading (Bandit B506)
    import yaml
    yaml_data = yaml.load("some: data")  # Should use yaml.safe_load()
    
    # This would normally be a security issue, but we're just demonstrating
    _ = weak_hash
    _ = query
    _ = result
    _ = cipher
    _ = yaml_data


def demo_file_operations():
    """Demonstrate unsafe file operations."""
    
    # SECURITY ISSUE: Potential path traversal (Bandit B108)
    user_filename = "../../../etc/passwd"
    file_path = f"/tmp/{user_filename}"
    
    # SECURITY ISSUE: Unsafe file permissions (Bandit B103)
    with open(file_path, "w") as f:
        f.write("sensitive data")
    os.chmod(file_path, 0o777)  # Too permissive
    
    _ = file_path


if __name__ == "__main__":
    demo_security_issues()
    demo_file_operations() 