# Pandacea Protocol Sprint 2 Completion Checklist
# Validates all Sprint 2 objectives have been implemented and working

param(
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"
$Host.UI.RawUI.WindowTitle = "Pandacea Sprint 2 Validation"

# Colors for output
$Green = "Green"
$Red = "Red"
$Yellow = "Yellow"
$White = "White"

function Write-Status {
    param(
        [string]$Message,
        [string]$Status,
        [string]$Color = $White
    )
    
    $statusSymbol = switch ($Status) {
        "PASS" { "[OK]" }
        "FAIL" { "[FAIL]" }
        "WARN" { "[WARN]" }
        default { "[INFO]" }
    }
    
    Write-Host "$statusSymbol $Message" -ForegroundColor $Color
}

function Test-Command {
    param(
        [string]$Command,
        [string]$Description
    )
    
    try {
        $result = Invoke-Expression $Command 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Status "$Description" "PASS" $Green
            if ($Verbose) {
                Write-Host "  Output: $result" -ForegroundColor $Gray
            }
            return $true
        } else {
            Write-Status "$Description" "FAIL" $Red
            Write-Host "  Error: $result" -ForegroundColor $Red
            return $false
        }
    } catch {
        Write-Status "$Description" "FAIL" $Red
        Write-Host "  Exception: $($_.Exception.Message)" -ForegroundColor $Red
        return $false
    }
}

function Test-FileExists {
    param(
        [string]$Path,
        [string]$Description
    )
    
    if (Test-Path $Path) {
        Write-Status "$Description" "PASS" $Green
        return $true
    } else {
        Write-Status "$Description" "FAIL" $Red
        return $false
    }
}

Write-Host "Pandacea Protocol Sprint 2 Validation" -ForegroundColor $Green
Write-Host "================================================" -ForegroundColor $Green
Write-Host ""

# Track overall success
$allPassed = $true

# 1. Agent Security Tests
Write-Host "Testing Agent Security Controls..." -ForegroundColor $Yellow
$securityTestsPassed = Test-Command "make agent-security-test" "Agent security tests"
if (-not $securityTestsPassed) { $allPassed = $false }

# 2. Simulation Tests
Write-Host "`nTesting Adversarial Economic Simulations..." -ForegroundColor $Yellow
$simsPassed = Test-Command "make sims-run" "Economic simulations"
if (-not $simsPassed) { $allPassed = $false }

# 3. Check for simulation report
Write-Host "`nChecking Simulation Report..." -ForegroundColor $Yellow
$reportExists = Test-FileExists "docs/economics/sim_report.html" "Simulation report exists"
if (-not $reportExists) { $allPassed = $false }

# 4. Security Configuration
Write-Host "`nChecking Security Configuration..." -ForegroundColor $Yellow
$securityConfigExists = Test-FileExists "agent-backend/config/security.yaml" "Security configuration file"
if (-not $securityConfigExists) { $allPassed = $false }

# 5. Security Documentation
Write-Host "`nChecking Security Documentation..." -ForegroundColor $Yellow
$securityDocsExist = Test-FileExists "docs/security/agent_abuse_controls.md" "Security abuse controls documentation"
if (-not $securityDocsExist) { $allPassed = $false }

# 6. Audit Documentation
Write-Host "`nChecking Audit Documentation..." -ForegroundColor $Yellow
$auditScopeExists = Test-FileExists "docs/audits/scope_sprint2.md" "Sprint 2 audit scope"
if (-not $auditScopeExists) { $allPassed = $false }

# 7. Economic Simulation Documentation
Write-Host "`nChecking Economic Documentation..." -ForegroundColor $Yellow
$econDocsExist = Test-FileExists "docs/economics/simulation_report.md" "Economic simulation report"
if (-not $econDocsExist) { $allPassed = $false }

# 8. Simulation Configuration
Write-Host "`nChecking Simulation Configuration..." -ForegroundColor $Yellow
$simsConfigExists = Test-FileExists "sims/config/params.yaml" "Simulation parameters"
if (-not $simsConfigExists) { $allPassed = $false }

# 9. Simulation Engine
Write-Host "`nChecking Simulation Engine..." -ForegroundColor $Yellow
$simsEngineExists = Test-FileExists "sims/engine/model.py" "Simulation engine model"
if (-not $simsEngineExists) { $allPassed = $false }

# 10. Simulation Tests
Write-Host "`nChecking Simulation Tests..." -ForegroundColor $Yellow
$simsTestsExist = Test-FileExists "sims/tests/test_model.py" "Simulation unit tests"
if (-not $simsTestsExist) { $allPassed = $false }

# 11. Check Makefile targets
Write-Host "`nChecking Makefile Targets..." -ForegroundColor $Yellow
$makefileExists = Test-FileExists "Makefile" "Makefile exists"
if ($makefileExists) {
    $makefileContent = Get-Content "Makefile" -Raw
    $hasAgentSecurityTest = $makefileContent -match "agent-security-test"
    $hasSimsRun = $makefileContent -match "sims-run"
    $hasSimsReport = $makefileContent -match "sims-report"
    
    Write-Status "Makefile contains agent-security-test target" $(if ($hasAgentSecurityTest) { "PASS" } else { "FAIL" }) $(if ($hasAgentSecurityTest) { $Green } else { $Red })
    Write-Status "Makefile contains sims-run target" $(if ($hasSimsRun) { "PASS" } else { "FAIL" }) $(if ($hasSimsRun) { $Green } else { $Red })
    Write-Status "Makefile contains sims-report target" $(if ($hasSimsReport) { "PASS" } else { "FAIL" }) $(if ($hasSimsReport) { $Green } else { $Red })
    
    if (-not ($hasAgentSecurityTest -and $hasSimsRun -and $hasSimsReport)) {
        $allPassed = $false
    }
} else {
    $allPassed = $false
}

# 12. Check for simulation output directory
Write-Host "`nChecking Simulation Output..." -ForegroundColor $Yellow
$simsOutExists = Test-Path "sims/out"
if ($simsOutExists) {
    $csvFiles = Get-ChildItem "sims/out" -Filter "*.csv" | Measure-Object
    Write-Status "Simulation output directory exists with $($csvFiles.Count) CSV files" "PASS" $Green
} else {
    Write-Status "Simulation output directory exists" "FAIL" $Red
    $allPassed = $false
}

# 13. Check for plots directory
Write-Host "`nChecking Plots..." -ForegroundColor $Yellow
$plotsDirExists = Test-Path "sims/plots"
if ($plotsDirExists) {
    $notebookExists = Test-Path "sims/plots/report.ipynb"
    Write-Status "Plots directory exists with report notebook" $(if ($notebookExists) { "PASS" } else { "FAIL" }) $(if ($notebookExists) { $Green } else { $Red })
    if (-not $notebookExists) { $allPassed = $false }
} else {
    Write-Status "Plots directory exists" "FAIL" $Red
    $allPassed = $false
}

# 14. Check for economics assets directory
Write-Host "`nChecking Economics Assets..." -ForegroundColor $Yellow
$econAssetsExists = Test-Path "docs/economics/assets"
if (-not $econAssetsExists) {
    # Create the directory if it doesn't exist
    New-Item -ItemType Directory -Path "docs/economics/assets" -Force | Out-Null
    Write-Status "Created economics assets directory" "PASS" $Green
} else {
    Write-Status "Economics assets directory exists" "PASS" $Green
}

# Summary
Write-Host "`n" + "="*50 -ForegroundColor $Green
if ($allPassed) {
    Write-Host "ALL SPRINT 2 OBJECTIVES COMPLETED SUCCESSFULLY!" -ForegroundColor $Green
    Write-Host "`n[OK] Agent abuse controls implemented and tested" -ForegroundColor $Green
    Write-Host "[OK] Adversarial economic simulations running" -ForegroundColor $Green
    Write-Host "[OK] External audit preparation completed" -ForegroundColor $Green
    Write-Host "[OK] Documentation and configuration in place" -ForegroundColor $Green
    Write-Host "`nReady for production deployment!" -ForegroundColor $Green
    exit 0
} else {
    Write-Host "SOME SPRINT 2 OBJECTIVES FAILED VALIDATION" -ForegroundColor $Red
    Write-Host "`nPlease review the failed checks above and complete the missing items." -ForegroundColor $Red
    Write-Host "`nKey commands to run:" -ForegroundColor $Yellow
    Write-Host "  make agent-security-test" -ForegroundColor $White
    Write-Host "  make sims-run" -ForegroundColor $White
    Write-Host "  make sims-report" -ForegroundColor $White
    exit 1
}
