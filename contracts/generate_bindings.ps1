# PowerShell script to generate Go bindings for LeaseAgreement smart contract
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "=== Generating Go Bindings with abigen ===" -ForegroundColor Green

# Define paths
$ABI_PATH = "./LeaseAgreement.abi.json"
$OUTPUT_PATH = "../agent-backend/internal/contracts/LeaseAgreement.go"
$PKG_NAME = "contracts"
$CONTRACT_NAME = "LeaseAgreement"

# Check if abigen is installed
try {
    $null = Get-Command abigen -ErrorAction Stop
} catch {
    Write-Host "abigen could not be found. Please install it first." -ForegroundColor Red
    Write-Host "Run: go install github.com/ethereum/go-ethereum/cmd/abigen@latest" -ForegroundColor Yellow
    exit 1
}

# Check if ABI file exists
if (!(Test-Path $ABI_PATH)) {
    Write-Host "ABI file not found at: $ABI_PATH" -ForegroundColor Red
    Write-Host "Please ensure the LeaseAgreement.abi.json file exists." -ForegroundColor Yellow
    exit 1
}

# Create the output directory if it doesn't exist
$outputDir = Split-Path $OUTPUT_PATH -Parent
if (!(Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    Write-Host "Created output directory: $outputDir" -ForegroundColor Green
}

# Generate the Go bindings using the ABI file
try {
    abigen --abi $ABI_PATH --pkg $PKG_NAME --type $CONTRACT_NAME --out $OUTPUT_PATH
    
    Write-Host "Go bindings generated successfully at: $OUTPUT_PATH" -ForegroundColor Green
    Write-Host "Package: $PKG_NAME" -ForegroundColor Cyan
    Write-Host "Contract: $CONTRACT_NAME" -ForegroundColor Cyan
} catch {
    Write-Host "Failed to generate Go bindings: $_" -ForegroundColor Red
    exit 1
} 