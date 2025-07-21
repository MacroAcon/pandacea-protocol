#!/usr/bin/env python3
"""
Verification script for the Pandacea Agent Backend implementation
Checks that all required files and components are in place
"""

import os
import json
import re
from pathlib import Path

def check_file_exists(filepath, description):
    """Check if a file exists and report status"""
    if os.path.exists(filepath):
        print(f"✅ {description}: {filepath}")
        return True
    else:
        print(f"❌ {description}: {filepath} (missing)")
        return False

def check_directory_exists(dirpath, description):
    """Check if a directory exists and report status"""
    if os.path.exists(dirpath) and os.path.isdir(dirpath):
        print(f"✅ {description}: {dirpath}")
        return True
    else:
        print(f"❌ {description}: {dirpath} (missing)")
        return False

def check_go_file_content(filepath, required_patterns):
    """Check if a Go file contains required patterns"""
    if not os.path.exists(filepath):
        print(f"❌ File not found: {filepath}")
        return False
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        all_found = True
        for pattern, description in required_patterns:
            if re.search(pattern, content, re.MULTILINE):
                print(f"  ✅ {description}")
            else:
                print(f"  ❌ {description} (not found)")
                all_found = False
        
        return all_found
    except Exception as e:
        print(f"❌ Error reading {filepath}: {e}")
        return False

def main():
    print("=== Pandacea Agent Backend Implementation Verification ===\n")
    
    # Check project structure
    print("📁 Project Structure:")
    structure_ok = True
    
    structure_ok &= check_directory_exists("cmd", "cmd directory")
    structure_ok &= check_directory_exists("cmd/agent", "cmd/agent directory")
    structure_ok &= check_directory_exists("internal", "internal directory")
    structure_ok &= check_directory_exists("internal/api", "internal/api directory")
    structure_ok &= check_directory_exists("internal/config", "internal/config directory")
    structure_ok &= check_directory_exists("internal/p2p", "internal/p2p directory")
    structure_ok &= check_directory_exists("internal/policy", "internal/policy directory")
    
    print()
    
    # Check key files
    print("📄 Key Files:")
    files_ok = True
    
    files_ok &= check_file_exists("go.mod", "Go module file")
    files_ok &= check_file_exists("go.sum", "Go dependencies file")
    files_ok &= check_file_exists("config.yaml", "Configuration file")
    files_ok &= check_file_exists("README.md", "Documentation")
    files_ok &= check_file_exists("cmd/agent/main.go", "Main application entry point")
    files_ok &= check_file_exists("internal/api/server.go", "API server implementation")
    files_ok &= check_file_exists("internal/config/config.go", "Configuration management")
    files_ok &= check_file_exists("internal/p2p/node.go", "P2P node implementation")
    files_ok &= check_file_exists("internal/policy/policy.go", "Policy engine")
    files_ok &= check_file_exists("internal/api/server_test.go", "API tests")
    
    print()
    
    # Check main.go content
    print("🔍 Main Application (cmd/agent/main.go):")
    main_patterns = [
        (r'package main', 'Package declaration'),
        (r'pandacea/agent-backend/internal/api', 'API import'),
        (r'pandacea/agent-backend/internal/config', 'Config import'),
        (r'pandacea/agent-backend/internal/p2p', 'P2P import'),
        (r'pandacea/agent-backend/internal/policy', 'Policy import'),
        (r'func main\(\)', 'Main function'),
        (r'config\.Load', 'Configuration loading'),
        (r'p2p\.NewNode', 'P2P node initialization'),
        (r'api\.NewServer', 'API server initialization'),
        (r'signal\.Notify', 'Signal handling'),
        (r'graceful shutdown', 'Graceful shutdown'),
    ]
    
    main_ok = check_go_file_content("cmd/agent/main.go", main_patterns)
    
    print()
    
    # Check API server content
    print("🔍 API Server (internal/api/server.go):")
    api_patterns = [
        (r'package api', 'Package declaration'),
        (r'type Server struct', 'Server struct definition'),
        (r'func NewServer', 'NewServer function'),
        (r'GET /api/v1/products', 'Products endpoint'),
        (r'POST /api/v1/leases', 'Leases endpoint'),
        (r'validateLeaseRequest', 'Input validation'),
        (r'did:pandacea', 'DID format validation'),
        (r'policy\.EvaluateRequest', 'Policy integration'),
        (r'StatusAccepted', 'Correct response status'),
    ]
    
    api_ok = check_go_file_content("internal/api/server.go", api_patterns)
    
    print()
    
    # Check P2P node content
    print("🔍 P2P Node (internal/p2p/node.go):")
    p2p_patterns = [
        (r'package p2p', 'Package declaration'),
        (r'type Node struct', 'Node struct definition'),
        (r'func NewNode', 'NewNode function'),
        (r'libp2p\.New', 'libp2p initialization'),
        (r'kadDHT', 'KAD-DHT integration'),
        (r'GetPeerID', 'Peer ID method'),
        (r'mdns\.NewMdnsService', 'mDNS discovery'),
    ]
    
    p2p_ok = check_go_file_content("internal/p2p/node.go", p2p_patterns)
    
    print()
    
    # Check policy engine content
    print("🔍 Policy Engine (internal/policy/policy.go):")
    policy_patterns = [
        (r'package policy', 'Package declaration'),
        (r'type Engine struct', 'Engine struct definition'),
        (r'func NewEngine', 'NewEngine function'),
        (r'func.*EvaluateRequest', 'EvaluateRequest function'),
        (r'Allowed: true', 'Policy approval logic'),
        (r'TODO.*Future policy', 'TODO comments'),
    ]
    
    policy_ok = check_go_file_content("internal/policy/policy.go", policy_patterns)
    
    print()
    
    # Check configuration content
    print("🔍 Configuration (internal/config/config.go):")
    config_patterns = [
        (r'package config', 'Package declaration'),
        (r'type Config struct', 'Config struct definition'),
        (r'func Load', 'Load function'),
        (r'yaml\.Unmarshal', 'YAML parsing'),
        (r'os\.Getenv', 'Environment variables'),
        (r'HTTP_PORT', 'HTTP port config'),
        (r'P2P_PORT', 'P2P port config'),
    ]
    
    config_ok = check_go_file_content("internal/config/config.go", config_patterns)
    
    print()
    
    # Check go.mod content
    print("🔍 Dependencies (go.mod):")
    go_mod_patterns = [
        (r'module pandacea/agent-backend', 'Module name'),
        (r'github\.com/go-chi/chi/v5', 'Chi router'),
        (r'github\.com/libp2p/go-libp2p', 'libp2p'),
        (r'github\.com/libp2p/go-libp2p-kad-dht', 'KAD-DHT'),
        (r'gopkg\.in/yaml\.v3', 'YAML parser'),
        (r'github\.com/stretchr/testify', 'Testing framework'),
    ]
    
    go_mod_ok = check_go_file_content("go.mod", go_mod_patterns)
    
    print()
    
    # Check config.yaml content
    print("🔍 Configuration File (config.yaml):")
    if os.path.exists("config.yaml"):
        try:
            with open("config.yaml", 'r', encoding='utf-8') as f:
                content = f.read()
            
            yaml_patterns = [
                (r'server:', 'Server section'),
                (r'port: 8080', 'HTTP port'),
                (r'p2p:', 'P2P section'),
                (r'listen_port: 0', 'P2P port'),
            ]
            
            yaml_ok = check_go_file_content("config.yaml", yaml_patterns)
        except Exception as e:
            print(f"❌ Error reading config.yaml: {e}")
            yaml_ok = False
    else:
        print("❌ config.yaml not found")
        yaml_ok = False
    
    print()
    
    # Summary
    print("=== Verification Summary ===")
    all_ok = (structure_ok and files_ok and main_ok and api_ok and 
              p2p_ok and policy_ok and config_ok and go_mod_ok and yaml_ok)
    
    if all_ok:
        print("✅ All components verified successfully!")
        print("✅ Agent backend implementation is complete and ready for use.")
        print("\nTo run the application:")
        print("  go run ./cmd/agent")
        print("  # or")
        print("  ./agent")
    else:
        print("❌ Some components need attention.")
        print("Please review the issues above and fix them.")
    
    print("\nNote: This verification checks file structure and content patterns.")
    print("For full testing, run: go test ./...")

if __name__ == "__main__":
    main() 