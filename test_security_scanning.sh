#!/bin/bash

# Test script for Pandacea Protocol Security Scanning
# This script demonstrates how to run security scans locally

set -e

echo "🔒 Pandacea Protocol Security Scanning Test"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test Go Agent Backend Security Scanning
echo ""
echo "Testing Go Agent Backend Security Scanning..."
echo "---------------------------------------------"

if [ -d "agent-backend" ]; then
    cd agent-backend
    
    # Check if gosec is installed
    if command -v gosec &> /dev/null; then
        print_status "gosec is installed"
        
        # Run security scan
        echo "Running gosec security scan..."
        if gosec -fmt=text -config=.gosec ./...; then
            print_status "gosec scan completed successfully"
        else
            print_warning "gosec found security issues (this is expected with demo files)"
        fi
        
        # Generate JSON report
        echo "Generating JSON security report..."
        gosec -fmt=json -out=security-report.json -config=.gosec ./... || true
        if [ -f "security-report.json" ]; then
            print_status "Security report generated: security-report.json"
        fi
        
    else
        print_error "gosec is not installed. Install with:"
        echo "curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v2.18.2"
    fi
    
    cd ..
else
    print_error "agent-backend directory not found"
fi

# Test Python Builder SDK Security Scanning
echo ""
echo "Testing Python Builder SDK Security Scanning..."
echo "-----------------------------------------------"

if [ -d "builder-sdk" ]; then
    cd builder-sdk
    
    # Check if bandit is installed
    if command -v bandit &> /dev/null; then
        print_status "bandit is installed"
        
        # Run security scan
        echo "Running bandit security scan..."
        if bandit -r pandacea_sdk/ -f txt -c .bandit; then
            print_status "bandit scan completed successfully"
        else
            print_warning "bandit found security issues (this is expected with demo files)"
        fi
        
        # Generate JSON report
        echo "Generating JSON security report..."
        bandit -r pandacea_sdk/ -f json -o security-report.json -c .bandit || true
        if [ -f "security-report.json" ]; then
            print_status "Security report generated: security-report.json"
        fi
        
    else
        print_error "bandit is not installed. Install with: pip install bandit"
    fi
    
    cd ..
else
    print_error "builder-sdk directory not found"
fi

echo ""
echo "🎯 Security Scanning Test Complete!"
echo ""
echo "Next Steps:"
echo "1. Review any security issues found in the reports"
echo "2. Fix high and medium severity issues"
echo "3. Commit the security scanning implementation"
echo "4. Push to trigger CI/CD security scanning"
echo ""
echo "For CI/CD testing:"
echo "- Create a branch with security issues to test failure scenarios"
echo "- Create a clean branch to test successful scans"
echo "- Check GitHub Actions for automated security scanning results"
