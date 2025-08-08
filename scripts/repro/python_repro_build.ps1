#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Build reproducible Python wheel for Pandacea SDK.
.DESCRIPTION
    This script builds the Python wheel with deterministic flags and generates SBOM.
#>

param(
    [string]$OutputDir = "artifacts",
    [switch]$Clean,
    [switch]$Verbose
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Create output directories
$artifactsDir = Join-Path $PSScriptRoot "..\..\$OutputDir"
$distDir = Join-Path $artifactsDir "dist"
$sbomDir = Join-Path $artifactsDir "sbom"

New-Item -ItemType Directory -Force -Path $distDir | Out-Null
New-Item -ItemType Directory -Force -Path $sbomDir | Out-Null

Write-Host "🐍 Building reproducible Python wheel..." -ForegroundColor Green

$pythonSdkPath = Join-Path $PSScriptRoot "..\..\builder-sdk"

Push-Location $pythonSdkPath
try {
    # Clean previous builds if requested
    if ($Clean -or (Test-Path "dist")) {
        Write-Host "Cleaning previous builds..."
        Remove-Item -Recurse -Force "dist" -ErrorAction SilentlyContinue
        Remove-Item -Recurse -Force "build" -ErrorAction SilentlyContinue
        Remove-Item -Recurse -Force "*.egg-info" -ErrorAction SilentlyContinue
    }
    
    # Verify Poetry is installed
    try {
        $poetryVersion = poetry --version
        Write-Host "Using Poetry: $poetryVersion" -ForegroundColor Cyan
    } catch {
        Write-Host "❌ Poetry not found. Please install Poetry first:" -ForegroundColor Red
        Write-Host "  pip install poetry" -ForegroundColor Yellow
        exit 1
    }
    
    # Install dependencies with exact versions
    Write-Host "Installing dependencies with exact versions..."
    poetry install --no-dev --no-interaction
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Dependency installation failed" -ForegroundColor Red
        exit 1
    }
    
    # Build wheel with deterministic flags
    Write-Host "Building Python wheel..."
    $env:PYTHONHASHSEED = "0"
    $env:SOURCE_DATE_EPOCH = "0"
    
    poetry build --format wheel --no-interaction
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Wheel build failed" -ForegroundColor Red
        exit 1
    }
    
    # Copy wheel to artifacts
    $wheelFile = Get-ChildItem "dist\*.whl" | Select-Object -First 1
    if ($wheelFile) {
        $targetPath = Join-Path $distDir $wheelFile.Name
        Copy-Item $wheelFile.FullName $targetPath -Force
        
        # Generate SBOM
        Write-Host "Generating Python SBOM..."
        $sbomPath = Join-Path $sbomDir "pandacea-sdk.spdx.json"
        syft . -o spdx-json -o $sbomPath
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "❌ SBOM generation failed" -ForegroundColor Red
            exit 1
        }
        
        # Verify wheel exists and get details
        if (Test-Path $targetPath) {
            $fileSize = (Get-Item $targetPath).Length
            $fileHash = (Get-FileHash -Path $targetPath -Algorithm SHA256).Hash.ToLower()
            
            Write-Host "✅ Python wheel built successfully" -ForegroundColor Green
            Write-Host "  Path: $targetPath" -ForegroundColor Cyan
            Write-Host "  Size: $([math]::Round($fileSize / 1KB, 2)) KB" -ForegroundColor Cyan
            Write-Host "  SHA256: $fileHash" -ForegroundColor Cyan
            Write-Host "  SBOM: $sbomPath" -ForegroundColor Cyan
            
            # Display wheel contents for verification
            if ($Verbose) {
                Write-Host "`n📦 Wheel contents:" -ForegroundColor Magenta
                $tempDir = Join-Path $env:TEMP "wheel-inspect-$(Get-Random)"
                New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
                try {
                    # Extract wheel contents for inspection
                    Add-Type -AssemblyName System.IO.Compression.FileSystem
                    [System.IO.Compression.ZipFile]::ExtractToDirectory($targetPath, $tempDir)
                    Get-ChildItem -Path $tempDir -Recurse -File | ForEach-Object {
                        Write-Host "  $($_.FullName.Replace($tempDir, ''))" -ForegroundColor Gray
                    }
                } finally {
                    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
                }
            }
        } else {
            Write-Host "❌ Wheel not found at expected location" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "❌ No wheel file generated" -ForegroundColor Red
        exit 1
    }
    
} finally {
    Pop-Location
}

Write-Host "`n🎯 Python reproducible build completed!" -ForegroundColor Green
