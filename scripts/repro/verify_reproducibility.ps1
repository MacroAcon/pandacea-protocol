#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Verify reproducible builds for Pandacea Protocol components.
.DESCRIPTION
    This script builds Go binaries, Python wheels, and containers twice
    and compares checksums/digests to ensure reproducibility.
#>

param(
    [switch]$Verbose,
    [string]$OutputDir = "artifacts"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Create output directories
$artifactsDir = Join-Path $PSScriptRoot "..\..\$OutputDir"
$binDir = Join-Path $artifactsDir "bin"
$distDir = Join-Path $artifactsDir "dist"
$sbomDir = Join-Path $artifactsDir "sbom"

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
New-Item -ItemType Directory -Force -Path $distDir | Out-Null
New-Item -ItemType Directory -Force -Path $sbomDir | Out-Null

Write-Host "🔍 Starting reproducibility verification..." -ForegroundColor Green

# Function to get file hash
function Get-FileHash {
    param([string]$FilePath)
    return (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
}

# Function to compare two builds
function Compare-Builds {
    param(
        [string]$Name,
        [string]$Build1Path,
        [string]$Build2Path
    )
    
    if (-not (Test-Path $Build1Path) -or -not (Test-Path $Build2Path)) {
        Write-Host "❌ $Name build files not found" -ForegroundColor Red
        return $false
    }
    
    $hash1 = Get-FileHash $Build1Path
    $hash2 = Get-FileHash $Build2Path
    
    if ($hash1 -eq $hash2) {
        Write-Host "✅ $Name builds are reproducible" -ForegroundColor Green
        return $true
    } else {
        Write-Host "❌ $Name builds are NOT reproducible" -ForegroundColor Red
        Write-Host "  Build 1: $hash1" -ForegroundColor Yellow
        Write-Host "  Build 2: $hash2" -ForegroundColor Yellow
        return $false
    }
}

# Go reproducible build verification
Write-Host "`n🔧 Verifying Go reproducible builds..." -ForegroundColor Cyan
$goAgentPath = Join-Path $PSScriptRoot "..\..\agent-backend"
$goBinary1 = Join-Path $binDir "agent-backend-1"
$goBinary2 = Join-Path $binDir "agent-backend-2"

# Build Go binary twice with deterministic flags
Push-Location $goAgentPath
try {
    # First build
    Write-Host "Building Go binary (attempt 1)..."
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o $goBinary1 ./cmd/agent
    
    # Second build
    Write-Host "Building Go binary (attempt 2)..."
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o $goBinary2 ./cmd/agent
    
    # Generate SBOM
    Write-Host "Generating Go SBOM..."
    syft . -o spdx-json -o (Join-Path $sbomDir "agent-backend.spdx.json")
    
} finally {
    Pop-Location
}

$goRepro = Compare-Builds "Go binary" $goBinary1 $goBinary2

# Python reproducible build verification
Write-Host "`n🐍 Verifying Python reproducible builds..." -ForegroundColor Cyan
$pythonSdkPath = Join-Path $PSScriptRoot "..\..\builder-sdk"
$wheel1 = Join-Path $distDir "pandacea_sdk-0.3.0-py3-none-any.whl"
$wheel2 = Join-Path $distDir "pandacea_sdk-0.3.0-py3-none-any.whl.2"

Push-Location $pythonSdkPath
try {
    # Clean previous builds
    if (Test-Path "dist") { Remove-Item -Recurse -Force "dist" }
    
    # First build
    Write-Host "Building Python wheel (attempt 1)..."
    poetry build --format wheel
    
    # Copy first wheel
    $firstWheel = Get-ChildItem "dist\*.whl" | Select-Object -First 1
    Copy-Item $firstWheel.FullName $wheel1
    
    # Clean and rebuild
    Remove-Item -Recurse -Force "dist"
    
    # Second build
    Write-Host "Building Python wheel (attempt 2)..."
    poetry build --format wheel
    
    # Copy second wheel
    $secondWheel = Get-ChildItem "dist\*.whl" | Select-Object -First 1
    Copy-Item $secondWheel.FullName $wheel2
    
    # Generate SBOM
    Write-Host "Generating Python SBOM..."
    syft . -o spdx-json -o (Join-Path $sbomDir "pandacea-sdk.spdx.json")
    
} finally {
    Pop-Location
}

$pythonRepro = Compare-Builds "Python wheel" $wheel1 $wheel2

# Container reproducible build verification
Write-Host "`n🐳 Verifying container reproducible builds..." -ForegroundColor Cyan
$containerImage1 = "pandacea/agent-backend:repro-test-1"
$containerImage2 = "pandacea/agent-backend:repro-test-2"

# Build container twice
Write-Host "Building container (attempt 1)..."
docker build -f agent-backend/Dockerfile -t $containerImage1 agent-backend

Write-Host "Building container (attempt 2)..."
docker build -f agent-backend/Dockerfile -t $containerImage2 agent-backend

# Get image digests
$digest1 = docker inspect $containerImage1 --format='{{.Id}}'
$digest2 = docker inspect $containerImage2 --format='{{.Id}}'

if ($digest1 -eq $digest2) {
    Write-Host "✅ Container builds are reproducible" -ForegroundColor Green
    $containerRepro = $true
} else {
    Write-Host "❌ Container builds are NOT reproducible" -ForegroundColor Red
    Write-Host "  Build 1: $digest1" -ForegroundColor Yellow
    Write-Host "  Build 2: $digest2" -ForegroundColor Yellow
    $containerRepro = $false
}

# Generate container SBOM
Write-Host "Generating container SBOM..."
syft $containerImage1 -o spdx-json -o (Join-Path $sbomDir "agent-backend-container.spdx.json")

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Container SBOM generation failed" -ForegroundColor Red
    exit 1
}

# Cleanup test images
docker rmi $containerImage1 $containerImage2 2>$null

# Summary
Write-Host "`n📊 Reproducibility Summary:" -ForegroundColor Magenta
Write-Host "  Go binary: $(if ($goRepro) { '✅' } else { '❌' })" -ForegroundColor $(if ($goRepro) { 'Green' } else { 'Red' })
Write-Host "  Python wheel: $(if ($pythonRepro) { '✅' } else { '❌' })" -ForegroundColor $(if ($pythonRepro) { 'Green' } else { 'Red' })
Write-Host "  Container: $(if ($containerRepro) { '✅' } else { '❌' })" -ForegroundColor $(if ($containerRepro) { 'Green' } else { 'Red' })

# Exit with failure if any build is not reproducible
if (-not ($goRepro -and $pythonRepro -and $containerRepro)) {
    Write-Host "`n❌ Some builds are not reproducible!" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ All builds are reproducible!" -ForegroundColor Green
Write-Host "Artifacts saved to: $artifactsDir" -ForegroundColor Cyan
