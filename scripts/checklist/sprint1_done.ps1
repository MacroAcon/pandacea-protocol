#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Sprint 1 Completion Checklist
.DESCRIPTION
    Validates that all Sprint 1 objectives have been completed successfully.
    This script runs all verification checks and provides a final status report.
#>

param(
    [switch]$Verbose
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
$Green = "Green"
$Red = "Red"
$Yellow = "Yellow"
$White = "White"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = $White
    )
    Write-Host $Message -ForegroundColor $Color
}

function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

function Test-FoundryCoverage {
    Write-ColorOutput "🔍 Checking Foundry coverage thresholds..." $Yellow
    
    try {
        # Change to contracts directory
        Push-Location "contracts"
        
        # Run coverage check
        $coverageOutput = & python "scripts/coverage/check_coverage.py" 2>&1
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-ColorOutput "✅ Foundry coverage check passed" $Green
            return $true
        } else {
            Write-ColorOutput "❌ Foundry coverage check failed" $Red
            Write-ColorOutput $coverageOutput $Red
            return $false
        }
    }
    catch {
        Write-ColorOutput "❌ Error running Foundry coverage check: $_" $Red
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-GoAgentTests {
    Write-ColorOutput "🔍 Running Go agent backend tests..." $Yellow
    
    try {
        # Change to agent-backend directory
        Push-Location "agent-backend"
        
        # Run Go tests
        $testOutput = & go test ./... -v 2>&1
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-ColorOutput "✅ Go agent backend tests passed" $Green
            return $true
        } else {
            Write-ColorOutput "❌ Go agent backend tests failed" $Red
            Write-ColorOutput $testOutput $Red
            return $false
        }
    }
    catch {
        Write-ColorOutput "❌ Error running Go agent backend tests: $_" $Red
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-PythonAccountantTests {
    Write-ColorOutput "🔍 Running Python privacy accountant tests..." $Yellow
    
    try {
        # Change to agent-backend/worker directory
        Push-Location "agent-backend/worker"
        
        # Check if pytest is available
        if (-not (Test-Command "pytest")) {
            Write-ColorOutput "⚠️  pytest not found, trying python -m pytest" $Yellow
            $pytestCmd = "python -m pytest"
        } else {
            $pytestCmd = "pytest"
        }
        
        # Run privacy accountant tests
        $testOutput = & $pytestCmd "tests/test_accountant.py" -v 2>&1
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-ColorOutput "✅ Python privacy accountant tests passed" $Green
            return $true
        } else {
            Write-ColorOutput "❌ Python privacy accountant tests failed" $Red
            Write-ColorOutput $testOutput $Red
            return $false
        }
    }
    catch {
        Write-ColorOutput "❌ Error running Python privacy accountant tests: $_" $Red
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-DockerPySyft {
    Write-ColorOutput "🔍 Testing Docker PySyft setup..." $Yellow
    
    try {
        # Check if Docker is running
        $dockerInfo = & docker info 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-ColorOutput "❌ Docker is not running or not accessible" $Red
            return $false
        }
        
        # Check if docker-compose.pysyft.yml exists
        if (-not (Test-Path "docker-compose.pysyft.yml")) {
            Write-ColorOutput "❌ docker-compose.pysyft.yml not found" $Red
            return $false
        }
        
        # Check if Dockerfile.pysyft exists
        if (-not (Test-Path "Dockerfile.pysyft")) {
            Write-ColorOutput "❌ Dockerfile.pysyft not found" $Red
            return $false
        }
        
        Write-ColorOutput "✅ Docker PySyft setup appears correct" $Green
        return $true
    }
    catch {
        Write-ColorOutput "❌ Error testing Docker PySyft setup: $_" $Red
        return $false
    }
}

function Test-MakefileTargets {
    Write-ColorOutput "🔍 Testing Makefile targets..." $Yellow
    
    $requiredTargets = @(
        "contracts-test",
        "contracts-coverage", 
        "verify",
        "pysyft-build",
        "pysyft-up",
        "demo-real-docker"
    )
    
    $missingTargets = @()
    
    foreach ($target in $requiredTargets) {
        $helpOutput = & make help 2>&1
        if ($helpOutput -notmatch $target) {
            $missingTargets += $target
        }
    }
    
    if ($missingTargets.Count -eq 0) {
        Write-ColorOutput "✅ All required Makefile targets found" $Green
        return $true
    } else {
        Write-ColorOutput "❌ Missing Makefile targets: $($missingTargets -join ', ')" $Red
        return $false
    }
}

function Test-Documentation {
    Write-ColorOutput "🔍 Checking documentation files..." $Yellow
    
    $requiredDocs = @(
        "docs/windows_quickstart.md",
        "contracts/scripts/coverage/README.md",
        ".github/workflows/verify.yml"
    )
    
    $missingDocs = @()
    
    foreach ($doc in $requiredDocs) {
        if (-not (Test-Path $doc)) {
            $missingDocs += $doc
        }
    }
    
    if ($missingDocs.Count -eq 0) {
        Write-ColorOutput "✅ All required documentation files found" $Green
        return $true
    } else {
        Write-ColorOutput "❌ Missing documentation files: $($missingDocs -join ', ')" $Red
        return $false
    }
}

# Main execution
Write-ColorOutput "🚀 Pandacea Protocol - Sprint 1 Completion Checklist" $White
Write-ColorOutput "==================================================" $White
Write-ColorOutput ""

# Initialize results
$results = @{}

# Check prerequisites
Write-ColorOutput "📋 Checking prerequisites..." $Yellow

if (-not (Test-Command "forge")) {
    Write-ColorOutput "❌ Foundry (forge) not found. Please install Foundry first." $Red
    exit 1
}

if (-not (Test-Command "docker")) {
    Write-ColorOutput "❌ Docker not found. Please install Docker Desktop first." $Red
    exit 1
}

if (-not (Test-Command "go")) {
    Write-ColorOutput "❌ Go not found. Please install Go first." $Red
    exit 1
}

Write-ColorOutput "✅ Prerequisites check passed" $Green
Write-ColorOutput ""

# Run all checks
$results["Foundry Coverage"] = Test-FoundryCoverage
$results["Go Agent Tests"] = Test-GoAgentTests
$results["Python Accountant Tests"] = Test-PythonAccountantTests
$results["Docker PySyft"] = Test-DockerPySyft
$results["Makefile Targets"] = Test-MakefileTargets
$results["Documentation"] = Test-Documentation

# Summary
Write-ColorOutput ""
Write-ColorOutput "📊 Sprint 1 Completion Summary" $White
Write-ColorOutput "=============================" $White

$passed = 0
$total = $results.Count

foreach ($check in $results.GetEnumerator()) {
    $status = if ($check.Value) { "✅ PASS" } else { "❌ FAIL" }
    $color = if ($check.Value) { $Green } else { $Red }
    Write-ColorOutput "$status $($check.Key)" $color
    
    if ($check.Value) {
        $passed++
    }
}

Write-ColorOutput ""
Write-ColorOutput "Results: $passed/$total checks passed" $(if ($passed -eq $total) { $Green } else { $Red })

if ($passed -eq $total) {
    Write-ColorOutput ""
    Write-ColorOutput "🎉 Sprint 1 is complete! All objectives have been successfully implemented." $Green
    Write-ColorOutput ""
    Write-ColorOutput "Next steps for the developer:" $White
    Write-ColorOutput "1. Run: cd C:\Users\thnxt\Documents\pandacea-protocol" $White
    Write-ColorOutput "2. Run: iwr https://foundry.paradigm.xyz | Invoke-Expression" $White
    Write-ColorOutput "3. Run: foundryup" $White
    Write-ColorOutput "4. Run: forge test -vvv" $White
    Write-ColorOutput "5. Run: forge coverage --report summary" $White
    Write-ColorOutput "6. Run: make verify" $White
    Write-ColorOutput "7. Run: make pysyft-up" $White
    Write-ColorOutput "8. Run: setx USE_DOCKER 1" $White
    Write-ColorOutput "9. Close and reopen PowerShell" $White
    Write-ColorOutput "10. Run: make demo-real-docker" $White
    exit 0
} else {
    Write-ColorOutput ""
    Write-ColorOutput "❌ Sprint 1 is not complete. Please address the failing checks above." $Red
    exit 1
}
