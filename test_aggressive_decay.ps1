#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Test script for the Pandacea Protocol Aggressive Reputation Decay implementation.

.DESCRIPTION
    This script helps test and validate the new aggressive reputation decay model
    that doubles the daily decay rate from 1 to 2 points per day and makes it
    DAO-configurable. It includes functions for:
    - Compiling and deploying smart contracts
    - Running integration tests
    - Validating decay calculations
    - Testing DAO configuration updates

.PARAMETER Action
    The action to perform: "compile", "test", "deploy", "validate", or "all"

.PARAMETER Environment
    The environment to use: "local" (default) or "testnet"

.EXAMPLE
    .\test_aggressive_decay.ps1 -Action "all"
    .\test_aggressive_decay.ps1 -Action "test" -Environment "local"
#>

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("compile", "test", "deploy", "validate", "all")]
    [string]$Action = "all",
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("local", "testnet")]
    [string]$Environment = "local"
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Test-Prerequisites {
    <#
    .SYNOPSIS
        Check if all required tools are installed.
    #>
    Write-ColorOutput "Checking prerequisites..." $Blue
    
    $tools = @{
        "Go" = "go version"
        "Python" = "python --version"
        "Poetry" = "poetry --version"
        "Foundry" = "forge --version"
        "Docker" = "docker --version"
    }
    
    $missing = @()
    
    foreach ($tool in $tools.GetEnumerator()) {
        try {
            $null = Invoke-Expression $tool.Value 2>$null
            Write-ColorOutput "✓ $($tool.Key) is installed" $Green
        }
        catch {
            Write-ColorOutput "✗ $($tool.Key) is not installed" $Red
            $missing += $tool.Key
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-ColorOutput "Missing tools: $($missing -join ', ')" $Red
        Write-ColorOutput "Please install the missing tools before proceeding." $Yellow
        return $false
    }
    
    return $true
}

function Set-EnvironmentVariables {
    <#
    .SYNOPSIS
        Set environment variables for testing.
    #>
    Write-ColorOutput "Setting environment variables..." $Blue
    
    if ($Environment -eq "local") {
        $env:RPC_URL = "http://127.0.0.1:8545"
        $env:AGENT_URL = "http://localhost:8080"
    } else {
        $env:RPC_URL = "https://polygon-mumbai.infura.io/v3/$env:INFURA_PROJECT_ID"
        $env:AGENT_URL = "https://agent.pandacea.io"
    }
    
    # Check for required environment variables
    $required = @("SPENDER_PRIVATE_KEY", "EARNER_PRIVATE_KEY", "CONTRACT_ADDRESS", "PGT_TOKEN_ADDRESS")
    $missing = @()
    
    foreach ($var in $required) {
        if (-not (Get-Variable -Name "env:$var" -ErrorAction SilentlyContinue)) {
            $missing += $var
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-ColorOutput "Missing environment variables: $($missing -join ', ')" $Red
        Write-ColorOutput "Please set these variables in your environment:" $Yellow
        foreach ($var in $missing) {
            Write-ColorOutput "  $var" $Yellow
        }
        return $false
    }
    
    Write-ColorOutput "✓ Environment variables set" $Green
    return $true
}

function Compile-Contracts {
    <#
    .SYNOPSIS
        Compile the smart contracts using Foundry.
    #>
    Write-ColorOutput "Compiling smart contracts..." $Blue
    
    try {
        Push-Location "contracts"
        
        # Clean previous builds
        Write-ColorOutput "Cleaning previous builds..." $Yellow
        forge clean
        
        # Compile contracts
        Write-ColorOutput "Compiling contracts..." $Yellow
        forge build
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Contracts compiled successfully" $Green
            return $true
        } else {
            Write-ColorOutput "✗ Contract compilation failed" $Red
            return $false
        }
    }
    catch {
        Write-ColorOutput "✗ Error during compilation: $($_.Exception.Message)" $Red
        return $false
    }
    finally {
        Pop-Location
    }
}

function Deploy-Contracts {
    <#
    .SYNOPSIS
        Deploy the smart contracts to the specified environment.
    #>
    Write-ColorOutput "Deploying smart contracts..." $Blue
    
    try {
        Push-Location "contracts"
        
        if ($Environment -eq "local") {
            # Deploy to local Anvil
            Write-ColorOutput "Deploying to local Anvil..." $Yellow
            forge script scripts/deploy.sol --rpc-url $env:RPC_URL --broadcast --verify
        } else {
            # Deploy to testnet
            Write-ColorOutput "Deploying to Polygon Mumbai testnet..." $Yellow
            forge script scripts/deploy.sol --rpc-url $env:RPC_URL --broadcast --verify --etherscan-api-key $env:ETHERSCAN_API_KEY
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Contracts deployed successfully" $Green
            return $true
        } else {
            Write-ColorOutput "✗ Contract deployment failed" $Red
            return $false
        }
    }
    catch {
        Write-ColorOutput "✗ Error during deployment: $($_.Exception.Message)" $Red
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-Integration {
    <#
    .SYNOPSIS
        Run the integration tests for aggressive reputation decay.
    #>
    Write-ColorOutput "Running integration tests..." $Blue
    
    try {
        Push-Location "integration"
        
        # Install Python dependencies if needed
        if (-not (Test-Path "venv")) {
            Write-ColorOutput "Creating Python virtual environment..." $Yellow
            python -m venv venv
        }
        
        # Activate virtual environment
        Write-ColorOutput "Activating virtual environment..." $Yellow
        & ".\venv\Scripts\Activate.ps1"
        
        # Install requirements
        Write-ColorOutput "Installing Python dependencies..." $Yellow
        pip install -r requirements.txt
        
        # Run the dispute system tests
        Write-ColorOutput "Running dispute system tests..." $Yellow
        python test_dispute_system.py
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Integration tests passed" $Green
            return $true
        } else {
            Write-ColorOutput "✗ Integration tests failed" $Red
            return $false
        }
    }
    catch {
        Write-ColorOutput "✗ Error during testing: $($_.Exception.Message)" $Red
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-DecayCalculations {
    <#
    .SYNOPSIS
        Validate the new decay calculations manually.
    #>
    Write-ColorOutput "Validating decay calculations..." $Blue
    
    try {
        # Test scenarios
        $testCases = @(
            @{
                InitialReputation = 800
                DaysPassed = 30
                OldDecayRate = 1
                NewDecayRate = 2
                ExpectedOldDecay = 30
                ExpectedNewDecay = 60
                ExpectedOldReputation = 770
                ExpectedNewReputation = 740
            },
            @{
                InitialReputation = 1000
                DaysPassed = 10
                OldDecayRate = 1
                NewDecayRate = 5
                ExpectedOldDecay = 10
                ExpectedNewDecay = 50
                ExpectedOldReputation = 990
                ExpectedNewReputation = 950
            }
        )
        
        foreach ($test in $testCases) {
            Write-ColorOutput "Testing decay calculation:" $Yellow
            Write-ColorOutput "  Initial reputation: $($test.InitialReputation)" $Yellow
            Write-ColorOutput "  Days passed: $($test.DaysPassed)" $Yellow
            Write-ColorOutput "  Old decay rate: $($test.OldDecayRate) points/day" $Yellow
            Write-ColorOutput "  New decay rate: $($test.NewDecayRate) points/day" $Yellow
            Write-ColorOutput "  Expected old decay: $($test.ExpectedOldDecay) points" $Yellow
            Write-ColorOutput "  Expected new decay: $($test.ExpectedNewDecay) points" $Yellow
            Write-ColorOutput "  Expected old reputation: $($test.ExpectedOldReputation)" $Yellow
            Write-ColorOutput "  Expected new reputation: $($test.ExpectedNewReputation)" $Yellow
            
            # Validate calculations
            $actualOldDecay = $test.DaysPassed * $test.OldDecayRate
            $actualNewDecay = $test.DaysPassed * $test.NewDecayRate
            $actualOldReputation = [Math]::Max(0, $test.InitialReputation - $actualOldDecay)
            $actualNewReputation = [Math]::Max(0, $test.InitialReputation - $actualNewDecay)
            
            if ($actualOldDecay -eq $test.ExpectedOldDecay -and 
                $actualNewDecay -eq $test.ExpectedNewDecay -and
                $actualOldReputation -eq $test.ExpectedOldReputation -and
                $actualNewReputation -eq $test.ExpectedNewReputation) {
                Write-ColorOutput "  ✓ Calculations are correct" $Green
            } else {
                Write-ColorOutput "  ✗ Calculations are incorrect" $Red
                return $false
            }
        }
        
        Write-ColorOutput "✓ All decay calculations validated" $Green
        return $true
    }
    catch {
        Write-ColorOutput "✗ Error during validation: $($_.Exception.Message)" $Red
        return $false
    }
}

function Show-Summary {
    <#
    .SYNOPSIS
        Show a summary of the aggressive reputation decay implementation.
    #>
    Write-ColorOutput "`n" $Blue
    Write-ColorOutput "=" * 60 $Blue
    Write-ColorOutput "AGGRESSIVE REPUTATION DECAY IMPLEMENTATION SUMMARY" $Blue
    Write-ColorOutput "=" * 60 $Blue
    
    Write-ColorOutput "`nKey Changes:" $Yellow
    Write-ColorOutput "• Decay rate increased from 1 to 2 points per day (doubled)" $Yellow
    Write-ColorOutput "• Decay rate is now DAO-configurable via setReputationDecayRate()" $Yellow
    Write-ColorOutput "• Added ReputationDecayRateUpdated event for transparency" $Yellow
    Write-ColorOutput "• Enhanced integration tests for decay validation" $Yellow
    
    Write-ColorOutput "`nBenefits:" $Yellow
    Write-ColorOutput "• More aggressive decay prevents reputation farming" $Yellow
    Write-ColorOutput "• DAO can tune decay rate based on market dynamics" $Yellow
    Write-ColorOutput "• Increased cost of maintaining stagnant high reputation" $Yellow
    Write-ColorOutput "• Better reflects current, active participation" $Yellow
    
    Write-ColorOutput "`nTesting:" $Yellow
    Write-ColorOutput "• Automated decay test with 30-day simulation" $Yellow
    Write-ColorOutput "• DAO configuration test with rate change to 5 points/day" $Yellow
    Write-ColorOutput "• Manual calculation validation" $Yellow
    
    Write-ColorOutput "`nNext Steps:" $Yellow
    Write-ColorOutput "• Deploy to testnet for community testing" $Yellow
    Write-ColorOutput "• Run agent-based simulations" $Yellow
    Write-ColorOutput "• Monitor decay impact on network behavior" $Yellow
    Write-ColorOutput "• Adjust rate based on simulation results" $Yellow
}

# Main execution
function Main {
    Write-ColorOutput "Pandacea Protocol - Aggressive Reputation Decay Test Script" $Blue
    Write-ColorOutput "Environment: $Environment" $Yellow
    Write-ColorOutput "Action: $Action" $Yellow
    Write-ColorOutput ""
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        exit 1
    }
    
    # Set environment variables
    if (-not (Set-EnvironmentVariables)) {
        exit 1
    }
    
    $success = $true
    
    # Execute requested actions
    switch ($Action) {
        "compile" {
            $success = Compile-Contracts
        }
        "deploy" {
            $success = Compile-Contracts -and Deploy-Contracts
        }
        "test" {
            $success = Test-Integration
        }
        "validate" {
            $success = Test-DecayCalculations
        }
        "all" {
            $success = Compile-Contracts -and Deploy-Contracts -and Test-Integration -and Test-DecayCalculations
        }
    }
    
    # Show summary
    Show-Summary
    
    if ($success) {
        Write-ColorOutput "`n🎉 All operations completed successfully!" $Green
        exit 0
    } else {
        Write-ColorOutput "`n❌ Some operations failed. Please review the errors above." $Red
        exit 1
    }
}

# Run main function
Main
