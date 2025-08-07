# Test script for Pandacea Protocol Security Scanning
# This script demonstrates how to run security scans locally

Write-Host "🔒 Pandacea Protocol Security Scanning Test" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

# Test Go Agent Backend Security Scanning
Write-Host ""
Write-Host "Testing Go Agent Backend Security Scanning..." -ForegroundColor Cyan
Write-Host "---------------------------------------------" -ForegroundColor Cyan

if (Test-Path "agent-backend") {
    Set-Location "agent-backend"
    
    # Check if gosec is installed
    try {
        $gosecVersion = gosec --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Status "gosec is installed"
            
            # Run security scan
            Write-Host "Running gosec security scan..." -ForegroundColor White
            gosec -fmt=text -config=.gosec ./...
            if ($LASTEXITCODE -eq 0) {
                Write-Status "gosec scan completed successfully"
            } else {
                Write-Warning "gosec found security issues (this is expected with demo files)"
            }
            
            # Generate JSON report
            Write-Host "Generating JSON security report..." -ForegroundColor White
            gosec -fmt=json -out=security-report.json -config=.gosec ./...
            if (Test-Path "security-report.json") {
                Write-Status "Security report generated: security-report.json"
            }
            
        } else {
            throw "gosec not found"
        }
    } catch {
        Write-Error "gosec is not installed. Install with:"
        Write-Host "curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b `$(go env GOPATH)/bin v2.18.2" -ForegroundColor Gray
    }
    
    Set-Location ".."
} else {
    Write-Error "agent-backend directory not found"
}

# Test Python Builder SDK Security Scanning
Write-Host ""
Write-Host "Testing Python Builder SDK Security Scanning..." -ForegroundColor Cyan
Write-Host "-----------------------------------------------" -ForegroundColor Cyan

if (Test-Path "builder-sdk") {
    Set-Location "builder-sdk"
    
    # Check if bandit is installed
    try {
        $banditVersion = bandit --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Status "bandit is installed"
            
            # Run security scan
            Write-Host "Running bandit security scan..." -ForegroundColor White
            bandit -r pandacea_sdk/ -f txt -c .bandit
            if ($LASTEXITCODE -eq 0) {
                Write-Status "bandit scan completed successfully"
            } else {
                Write-Warning "bandit found security issues (this is expected with demo files)"
            }
            
            # Generate JSON report
            Write-Host "Generating JSON security report..." -ForegroundColor White
            bandit -r pandacea_sdk/ -f json -o security-report.json -c .bandit
            if (Test-Path "security-report.json") {
                Write-Status "Security report generated: security-report.json"
            }
            
        } else {
            throw "bandit not found"
        }
    } catch {
        Write-Error "bandit is not installed. Install with: pip install bandit"
    }
    
    Set-Location ".."
} else {
    Write-Error "builder-sdk directory not found"
}

Write-Host ""
Write-Host "🎯 Security Scanning Test Complete!" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor White
Write-Host "1. Review any security issues found in the reports" -ForegroundColor Gray
Write-Host "2. Fix high and medium severity issues" -ForegroundColor Gray
Write-Host "3. Commit the security scanning implementation" -ForegroundColor Gray
Write-Host "4. Push to trigger CI/CD security scanning" -ForegroundColor Gray
Write-Host ""
Write-Host "For CI/CD testing:" -ForegroundColor White
Write-Host "- Create a branch with security issues to test failure scenarios" -ForegroundColor Gray
Write-Host "- Create a clean branch to test successful scans" -ForegroundColor Gray
Write-Host "- Check GitHub Actions for automated security scanning results" -ForegroundColor Gray
