#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Run chaos engineering scenarios for Pandacea Protocol

.DESCRIPTION
    Windows PowerShell script to run chaos engineering scenarios locally.
    This script provides a Windows-first interface to the chaos harness.

.PARAMETER Scenario
    Name of the scenario to run (e.g., "evm_rpc_flap")

.PARAMETER List
    List available scenarios

.PARAMETER ScenariosDir
    Directory containing scenario modules (default: "integration/chaos/scenarios")

.EXAMPLE
    .\run-chaos.ps1 -Scenario evm_rpc_flap

.EXAMPLE
    .\run-chaos.ps1 -List

.NOTES
    Requires Python 3.8+ and the chaos harness to be available.
#>

param(
    [Parameter(Position=0)]
    [string]$Scenario,
    
    [switch]$List,
    
    [string]$ScenariosDir = "integration/chaos/scenarios"
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Function to check if Python is available
function Test-Python {
    try {
        $pythonVersion = python --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Python found: $pythonVersion" -ForegroundColor Green
            return $true
        }
    }
    catch {
        Write-Host "❌ Python not found in PATH" -ForegroundColor Red
        return $false
    }
    return $false
}

# Function to check if chaos harness exists
function Test-ChaosHarness {
    $harnessPath = "integration/chaos/run_chaos.py"
    if (Test-Path $harnessPath) {
        Write-Host "✅ Chaos harness found at $harnessPath" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "❌ Chaos harness not found at $harnessPath" -ForegroundColor Red
        return $false
    }
}

# Function to run chaos scenario
function Invoke-ChaosScenario {
    param([string]$ScenarioName)
    
    Write-Host "🚀 Running chaos scenario: $ScenarioName" -ForegroundColor Yellow
    
    # Set environment variable for the scenario
    $env:SCENARIO = $ScenarioName
    
    # Run the chaos harness
    $harnessPath = "integration/chaos/run_chaos.py"
    $scenariosDir = "integration/chaos/scenarios"
    
    try {
        python $harnessPath --scenario $ScenarioName --scenarios-dir $scenariosDir
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-Host "✅ Scenario $ScenarioName completed successfully" -ForegroundColor Green
        }
        else {
            Write-Host "❌ Scenario $ScenarioName failed with exit code $exitCode" -ForegroundColor Red
        }
        
        return $exitCode -eq 0
    }
    catch {
        Write-Host "❌ Failed to run scenario $ScenarioName : $_" -ForegroundColor Red
        return $false
    }
}

# Function to list available scenarios
function Get-ChaosScenarios {
    $scenariosDir = "integration/chaos/scenarios"
    if (Test-Path $scenariosDir) {
        $scenarios = Get-ChildItem -Path $scenariosDir -Filter "*.py" | 
                    Where-Object { $_.Name -ne "__init__.py" } |
                    ForEach-Object { $_.BaseName }
        
        if ($scenarios) {
            Write-Host "Available scenarios:" -ForegroundColor Cyan
            foreach ($scenario in $scenarios) {
                Write-Host "  - $scenario" -ForegroundColor White
            }
        }
        else {
            Write-Host "No scenarios found in $scenariosDir" -ForegroundColor Yellow
        }
    }
    else {
        Write-Host "Scenarios directory not found: $scenariosDir" -ForegroundColor Red
    }
}

# Main execution
Write-Host "🔧 Pandacea Protocol Chaos Engineering Runner" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# Check prerequisites
if (-not (Test-Python)) {
    Write-Host "Please install Python 3.8+ and ensure it's in your PATH" -ForegroundColor Red
    exit 1
}

if (-not (Test-ChaosHarness)) {
    Write-Host "Please ensure the chaos harness is available" -ForegroundColor Red
    exit 1
}

# Handle list command
if ($List) {
    Get-ChaosScenarios
    exit 0
}

# Handle scenario execution
if ($Scenario) {
    $success = Invoke-ChaosScenario -ScenarioName $Scenario
    exit $(if ($success) { 0 } else { 1 })
}
else {
    Write-Host "No scenario specified. Use -Scenario <name> or -List to see available scenarios" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor White
    Write-Host "  .\run-chaos.ps1 -Scenario evm_rpc_flap" -ForegroundColor Gray
    Write-Host "  .\run-chaos.ps1 -List" -ForegroundColor Gray
    exit 1
}
