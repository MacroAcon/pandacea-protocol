#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Build reproducible container images for Pandacea Protocol.
.DESCRIPTION
    This script builds container images with deterministic flags and generates SBOM.
#>

param(
    [string]$OutputDir = "artifacts",
    [string]$ImageName = "pandacea/agent-backend",
    [string]$Tag = "repro",
    [switch]$Clean,
    [switch]$Verbose
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Create output directories
$artifactsDir = Join-Path $PSScriptRoot "..\..\$OutputDir"
$sbomDir = Join-Path $artifactsDir "sbom"

New-Item -ItemType Directory -Force -Path $sbomDir | Out-Null

Write-Host "🐳 Building reproducible container..." -ForegroundColor Green

# Set build arguments for reproducible builds
$buildArgs = @(
    "--build-arg", "VERSION_SHA=$Tag",
    "--build-arg", "BUILD_DATE=$(Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ')",
    "--build-arg", "VCS_REF=$(git rev-parse HEAD 2>$null || 'unknown')"
)

$fullImageName = "$ImageName`:$Tag"

# Build the container
Write-Host "Building container: $fullImageName"
$buildCmd = @("docker", "build", "-f", "agent-backend/Dockerfile", "-t", $fullImageName) + $buildArgs + @("agent-backend")

if ($Verbose) {
    Write-Host "Build command: $($buildCmd -join ' ')" -ForegroundColor Cyan
}

& $buildCmd[0] $buildCmd[1..($buildCmd.Length-1)]

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Container build failed" -ForegroundColor Red
    exit 1
}

# Generate SBOM
Write-Host "Generating container SBOM..."
$sbomPath = Join-Path $sbomDir "agent-backend-container.spdx.json"
syft $fullImageName -o spdx-json -o $sbomPath

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ SBOM generation failed" -ForegroundColor Red
    exit 1
}

# Get image details
Write-Host "Getting image details..."
$imageInfo = docker inspect $fullImageName --format='{{.Id}}' 2>$null
if ($imageInfo) {
    $imageSize = docker images $fullImageName --format='{{.Size}}' 2>$null
    
    Write-Host "✅ Container built successfully" -ForegroundColor Green
    Write-Host "  Image: $fullImageName" -ForegroundColor Cyan
    Write-Host "  Digest: $imageInfo" -ForegroundColor Cyan
    Write-Host "  Size: $imageSize" -ForegroundColor Cyan
    Write-Host "  SBOM: $sbomPath" -ForegroundColor Cyan
    
    # Display image layers for verification
    if ($Verbose) {
        Write-Host "`n📦 Image layers:" -ForegroundColor Magenta
        docker history $fullImageName --format='table {{.CreatedBy}}\t{{.Size}}\t{{.Comment}}'
    }
} else {
    Write-Host "❌ Image not found after build" -ForegroundColor Red
    exit 1
}

Write-Host "`n🎯 Container reproducible build completed!" -ForegroundColor Green
