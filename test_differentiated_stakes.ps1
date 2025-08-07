#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Test script for Differentiated Dispute Stakes implementation

.DESCRIPTION
    This script helps test the new Differentiated Dispute Stakes functionality
    by running integration tests and examples.

.PARAMETER TestType
    Type of test to run: "integration", "example", or "all"

.PARAMETER Environment
    Environment to test against: "local", "testnet", or "mainnet"

.EXAMPLE
    .\test_differentiated_stakes.ps1 -TestType "all" -Environment "local"
#>

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("integration", "example", "all")]
    [string]$TestType = "all",
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("local", "testnet", "mainnet")]
    [string]$Environment = "local"
)

# Set error action preference
$ErrorActionPreference = "Stop"

Write-Host "🚀 Pandacea Protocol - Differentiated Dispute Stakes Test Suite" -ForegroundColor Green
Write-Host "=" * 60 -ForegroundColor Green

# Function to check if command exists
function Test-Command($cmdname) {
    return [bool](Get-Command -Name $cmdname -ErrorAction SilentlyContinue)
}

# Function to check Python environment
function Test-PythonEnvironment {
    Write-Host "Checking Python environment..." -ForegroundColor Yellow
    
    if (-not (Test-Command "python")) {
        Write-Host "❌ Python not found. Please install Python 3.8+ and add it to PATH." -ForegroundColor Red
        return $false
    }
    
    $pythonVersion = python --version
    Write-Host "✅ Python found: $pythonVersion" -ForegroundColor Green
    
    # Check if required packages are installed
    try {
        python -c "import requests, web3, cryptography" 2>$null
        Write-Host "✅ Required Python packages found" -ForegroundColor Green
    }
    catch {
        Write-Host "❌ Missing required Python packages. Installing..." -ForegroundColor Red
        pip install requests web3 cryptography multibase multihash
    }
    
    return $true
}

# Function to set environment variables
function Set-TestEnvironment {
    Write-Host "Setting up test environment..." -ForegroundColor Yellow
    
    $envVars = @{
        "AGENT_URL" = "http://localhost:8080"
        "RPC_URL" = "http://127.0.0.1:8545"
    }
    
    switch ($Environment) {
        "local" {
            $envVars["CONTRACT_ADDRESS"] = "0x1234567890123456789012345678901234567890"
            $envVars["PGT_TOKEN_ADDRESS"] = "0x0987654321098765432109876543210987654321"
            Write-Host "Using local environment configuration" -ForegroundColor Cyan
        }
        "testnet" {
            $envVars["CONTRACT_ADDRESS"] = $env:TESTNET_CONTRACT_ADDRESS
            $envVars["PGT_TOKEN_ADDRESS"] = $env:TESTNET_PGT_TOKEN_ADDRESS
            Write-Host "Using testnet environment configuration" -ForegroundColor Cyan
        }
        "mainnet" {
            $envVars["CONTRACT_ADDRESS"] = $env:MAINNET_CONTRACT_ADDRESS
            $envVars["PGT_TOKEN_ADDRESS"] = $env:MAINNET_PGT_TOKEN_ADDRESS
            Write-Host "Using mainnet environment configuration" -ForegroundColor Cyan
        }
    }
    
    # Set environment variables
    foreach ($key in $envVars.Keys) {
        if ($envVars[$key]) {
            Set-Item -Path "env:$key" -Value $envVars[$key]
            Write-Host "Set $key = $($envVars[$key])" -ForegroundColor Gray
        }
    }
    
    # Check for required private keys
    if (-not $env:SPENDER_PRIVATE_KEY) {
        Write-Host "⚠️  SPENDER_PRIVATE_KEY not set. Some tests may fail." -ForegroundColor Yellow
    }
    
    if (-not $env:EARNER_PRIVATE_KEY) {
        Write-Host "⚠️  EARNER_PRIVATE_KEY not set. Some tests may fail." -ForegroundColor Yellow
    }
}

# Function to run integration tests
function Test-IntegrationTests {
    Write-Host "Running integration tests..." -ForegroundColor Yellow
    Write-Host "-" * 40 -ForegroundColor Gray
    
    $testPath = "integration\test_dispute_system.py"
    
    if (-not (Test-Path $testPath)) {
        Write-Host "❌ Integration test file not found: $testPath" -ForegroundColor Red
        return $false
    }
    
    try {
        $result = python $testPath 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Integration tests passed!" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ Integration tests failed!" -ForegroundColor Red
            Write-Host $result -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "❌ Error running integration tests: $_" -ForegroundColor Red
        return $false
    }
}

# Function to run example
function Test-Example {
    Write-Host "Running dynamic staking example..." -ForegroundColor Yellow
    Write-Host "-" * 40 -ForegroundColor Gray
    
    $examplePath = "builder-sdk\examples\dynamic_staking_example.py"
    
    if (-not (Test-Path $examplePath)) {
        Write-Host "❌ Example file not found: $examplePath" -ForegroundColor Red
        return $false
    }
    
    try {
        $result = python $examplePath 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Example completed successfully!" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ Example failed!" -ForegroundColor Red
            Write-Host $result -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "❌ Error running example: $_" -ForegroundColor Red
        return $false
    }
}

# Function to compile smart contracts
function Test-SmartContracts {
    Write-Host "Compiling smart contracts..." -ForegroundColor Yellow
    Write-Host "-" * 40 -ForegroundColor Gray
    
    $contractsPath = "contracts"
    
    if (-not (Test-Path $contractsPath)) {
        Write-Host "❌ Contracts directory not found: $contractsPath" -ForegroundColor Red
        return $false
    }
    
    # Check if Foundry is installed
    if (-not (Test-Command "forge")) {
        Write-Host "⚠️  Foundry not found. Skipping contract compilation." -ForegroundColor Yellow
        Write-Host "   Install Foundry from: https://getfoundry.sh/" -ForegroundColor Gray
        return $true
    }
    
    try {
        Push-Location $contractsPath
        $result = forge build 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Smart contracts compiled successfully!" -ForegroundColor Green
            Pop-Location
            return $true
        } else {
            Write-Host "❌ Smart contract compilation failed!" -ForegroundColor Red
            Write-Host $result -ForegroundColor Red
            Pop-Location
            return $false
        }
    }
    catch {
        Write-Host "❌ Error compiling smart contracts: $_" -ForegroundColor Red
        Pop-Location
        return $false
    }
}

# Function to run Go backend tests
function Test-GoBackend {
    Write-Host "Testing Go backend..." -ForegroundColor Yellow
    Write-Host "-" * 40 -ForegroundColor Gray
    
    $backendPath = "agent-backend"
    
    if (-not (Test-Path $backendPath)) {
        Write-Host "❌ Backend directory not found: $backendPath" -ForegroundColor Red
        return $false
    }
    
    # Check if Go is installed
    if (-not (Test-Command "go")) {
        Write-Host "⚠️  Go not found. Skipping backend tests." -ForegroundColor Yellow
        Write-Host "   Install Go from: https://golang.org/dl/" -ForegroundColor Gray
        return $true
    }
    
    try {
        Push-Location $backendPath
        $result = go test ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Go backend tests passed!" -ForegroundColor Green
            Pop-Location
            return $true
        } else {
            Write-Host "❌ Go backend tests failed!" -ForegroundColor Red
            Write-Host $result -ForegroundColor Red
            Pop-Location
            return $false
        }
    }
    catch {
        Write-Host "❌ Error running Go backend tests: $_" -ForegroundColor Red
        Pop-Location
        return $false
    }
}

# Main execution
try {
    # Check Python environment
    if (-not (Test-PythonEnvironment)) {
        exit 1
    }
    
    # Set up test environment
    Set-TestEnvironment
    
    $results = @{}
    
    # Run tests based on TestType parameter
    switch ($TestType) {
        "integration" {
            $results["Integration Tests"] = Test-IntegrationTests
        }
        "example" {
            $results["Example"] = Test-Example
        }
        "all" {
            $results["Smart Contracts"] = Test-SmartContracts
            $results["Go Backend"] = Test-GoBackend
            $results["Integration Tests"] = Test-IntegrationTests
            $results["Example"] = Test-Example
        }
    }
    
    # Display results
    Write-Host "`n" + "=" * 60 -ForegroundColor Green
    Write-Host "TEST RESULTS" -ForegroundColor Green
    Write-Host "=" * 60 -ForegroundColor Green
    
    $passed = 0
    $failed = 0
    
    foreach ($test in $results.Keys) {
        if ($results[$test]) {
            Write-Host "✅ $test : PASSED" -ForegroundColor Green
            $passed++
        } else {
            Write-Host "❌ $test : FAILED" -ForegroundColor Red
            $failed++
        }
    }
    
    Write-Host "`nSummary:" -ForegroundColor Cyan
    Write-Host "  Passed: $passed" -ForegroundColor Green
    Write-Host "  Failed: $failed" -ForegroundColor Red
    
    if ($failed -eq 0) {
        Write-Host "`n🎉 All tests passed! Differentiated Dispute Stakes implementation is working correctly." -ForegroundColor Green
        exit 0
    } else {
        Write-Host "`n❌ Some tests failed. Please review the implementation." -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "❌ Unexpected error: $_" -ForegroundColor Red
    exit 1
}
