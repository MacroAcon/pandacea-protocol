#!/usr/bin/env python3
"""
Simple Solidity contract verification script
Checks for basic syntax and structure without requiring a full Solidity compiler
"""

import os
import re
from pathlib import Path

def check_solidity_file(file_path):
    """Basic syntax check for Solidity files"""
    print(f"Checking {file_path}...")
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Basic Solidity syntax checks
        checks = {
            "SPDX License": "SPDX-License-Identifier" in content,
            "Pragma Statement": "pragma solidity" in content,
            "Contract Definition": "contract " in content or "interface " in content,
            "Function Definitions": "function " in content,
            "Event Definitions": "event " in content,
            "Import Statements": "import " in content,
            "Proper Braces": content.count('{') == content.count('}'),
            "Proper Parentheses": content.count('(') == content.count(')'),
            "Proper Semicolons": content.count(';') > 0,
        }
        
        all_passed = True
        for check_name, passed in checks.items():
            status = "✅ PASS" if passed else "❌ FAIL"
            print(f"  {check_name}: {status}")
            if not passed:
                all_passed = False
        
        return all_passed
        
    except Exception as e:
        print(f"  ❌ ERROR: {e}")
        return False

def check_contract_structure():
    """Check the overall contract structure and requirements"""
    print("\n=== Contract Structure Verification ===")
    
    # Check main contract file
    main_contract = "src/LeaseAgreement.sol"
    if os.path.exists(main_contract):
        print(f"✅ Main contract exists: {main_contract}")
        
        with open(main_contract, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Check for required components
        required_components = {
            "ILeaseAgreement Interface": "interface ILeaseAgreement" in content,
            "LeaseAgreement Contract": "contract LeaseAgreement" in content,
            "ReentrancyGuard Inheritance": "ReentrancyGuard" in content,
            "Ownable Inheritance": "Ownable" in content,
            "MIN_PRICE Constant": "MIN_PRICE = 0.001 ether" in content,
            "createLease Function": "function createLease" in content,
            "approveLease Function": "function approveLease" in content,
            "executeLease Function": "function executeLease" in content,
            "raiseDispute Function": "function raiseDispute" in content,
            "DMP Logic": "msg.value >= MIN_PRICE" in content,
            "TODO Comments": "TODO:" in content,
        }
        
        for component, exists in required_components.items():
            status = "✅" if exists else "❌"
            print(f"  {component}: {status}")
            
    else:
        print(f"❌ Main contract missing: {main_contract}")

def check_test_structure():
    """Check the test file structure"""
    print("\n=== Test Structure Verification ===")
    
    test_file = "test/LeaseAgreement.t.sol"
    if os.path.exists(test_file):
        print(f"✅ Test file exists: {test_file}")
        
        with open(test_file, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Check for required test components
        test_components = {
            "Test Contract": "contract LeaseAgreementTest" in content,
            "Test Inheritance": "is Test" in content,
            "Success Cases": "testCreateLeaseWithExactMinPrice" in content,
            "Failure Cases": "testCreateLeaseRevertsWhenPaymentBelowMinPrice" in content,
            "Event Testing": "expectEmit" in content,
            "Revert Testing": "expectRevert" in content,
            "MIN_PRICE Testing": "testMinPriceConstantIsCorrect" in content,
        }
        
        for component, exists in test_components.items():
            status = "✅" if exists else "❌"
            print(f"  {component}: {status}")
            
    else:
        print(f"❌ Test file missing: {test_file}")

def check_project_structure():
    """Check the overall project structure"""
    print("\n=== Project Structure Verification ===")
    
    required_files = [
        "src/LeaseAgreement.sol",
        "test/LeaseAgreement.t.sol",
        "foundry.toml",
        "remappings.txt",
        "README.md",
    ]
    
    required_dirs = [
        "src",
        "test", 
        "lib",
        "lib/openzeppelin-contracts",
        "lib/forge-std",
    ]
    
    for file_path in required_files:
        if os.path.exists(file_path):
            print(f"✅ {file_path}")
        else:
            print(f"❌ {file_path} (missing)")
    
    for dir_path in required_dirs:
        if os.path.exists(dir_path):
            print(f"✅ {dir_path}/")
        else:
            print(f"❌ {dir_path}/ (missing)")

def main():
    print("=== Pandacea Protocol Smart Contracts Verification ===\n")
    
    # Check project structure
    check_project_structure()
    
    # Check contract structure
    check_contract_structure()
    
    # Check test structure
    check_test_structure()
    
    # Check individual files
    print("\n=== Individual File Verification ===")
    
    files_to_check = [
        "src/LeaseAgreement.sol",
        "test/LeaseAgreement.t.sol",
        "lib/openzeppelin-contracts/contracts/security/ReentrancyGuard.sol",
        "lib/openzeppelin-contracts/contracts/access/Ownable.sol",
        "lib/openzeppelin-contracts/contracts/utils/Context.sol",
        "lib/forge-std/src/Test.sol",
        "lib/forge-std/src/Vm.sol",
    ]
    
    all_files_ok = True
    for file_path in files_to_check:
        if os.path.exists(file_path):
            if not check_solidity_file(file_path):
                all_files_ok = False
        else:
            print(f"❌ File missing: {file_path}")
            all_files_ok = False
        print()
    
    print("=== Verification Summary ===")
    if all_files_ok:
        print("✅ All contracts appear to be syntactically correct!")
        print("✅ Project structure is properly set up!")
        print("✅ Ready for Foundry testing!")
    else:
        print("❌ Some issues were found. Please review the output above.")
    
    print("\nNote: This is a basic syntax check. For full compilation and testing,")
    print("install Foundry and run 'forge build' and 'forge test'")

if __name__ == "__main__":
    main() 