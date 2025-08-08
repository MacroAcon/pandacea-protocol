#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Build reproducible Go binary for Pandacea Agent Backend.
.DESCRIPTION
    This script builds the Go binary with deterministic flags and generates SBOM.
#>

param(
    [string]$OutputDir = "artifacts",
    [string]$Platform = "windows",
    [string]$Arch = "amd64"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Create output directories
$artifactsDir = Join-Path $PSScriptRoot "..\..\$OutputDir"
$binDir = Join-Path $artifactsDir "bin"
$sbomDir = Join-Path $artifactsDir "sbom"

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
New-Item -ItemType Directory -Force -Path $sbomDir | Out-Null

Write-Host "🔧 Building reproducible Go binary..." -ForegroundColor Green

# Set environment variables for reproducible builds
$env:CGO_ENABLED = "0"
$env:GOOS = $Platform
$env:GOARCH = $Arch

# Build flags for reproducibility
$buildFlags = @(
    "-trimpath",
    "-buildvcs=false",
    "-ldflags=-s -w"
)

$outputPath = Join-Path $binDir "agent-backend"
$sourcePath = Join-Path $PSScriptRoot "..\..\agent-backend"

Push-Location $sourcePath
try {
    # Build the binary
    Write-Host "Building Go binary with deterministic flags..."
    $buildArgs = $buildFlags + @("-o", $outputPath, "./cmd/agent")
    go build @buildArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Go build failed" -ForegroundColor Red
        exit 1
    }
    
    # Generate SBOM
    Write-Host "Generating Go SBOM..."
    $sbomPath = Join-Path $sbomDir "agent-backend.spdx.json"
    syft . -o spdx-json -o $sbomPath
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ SBOM generation failed" -ForegroundColor Red
        exit 1
    }
    
    # Verify binary exists
    if (Test-Path $outputPath) {
        $fileSize = (Get-Item $outputPath).Length
        $fileHash = (Get-FileHash -Path $outputPath -Algorithm SHA256).Hash.ToLower()
        
        Write-Host "✅ Go binary built successfully" -ForegroundColor Green
        Write-Host "  Path: $outputPath" -ForegroundColor Cyan
        Write-Host "  Size: $($fileSize / 1MB) MB" -ForegroundColor Cyan
        Write-Host "  SHA256: $fileHash" -ForegroundColor Cyan
        Write-Host "  SBOM: $sbomPath" -ForegroundColor Cyan
    } else {
        Write-Host "❌ Binary not found at expected location" -ForegroundColor Red
        exit 1
    }
    
} finally {
    Pop-Location
}

Write-Host "`n🎯 Go reproducible build completed!" -ForegroundColor Green
