#!/bin/bash
set -e

# Navigate to the contracts directory
cd "$(dirname "$0")"

echo "=== 1. Compiling Smart Contracts with Foundry ==="
forge build

echo "=== 2. Generating Go Bindings with abigen ==="

# Define paths
ABI_PATH="./out/LeaseAgreement.sol/LeaseAgreement.json"
OUTPUT_PATH="../agent-backend/internal/contracts/LeaseAgreement.go"
PKG_NAME="contracts"
CONTRACT_NAME="LeaseAgreement"

# Check if abigen is installed
if ! command -v abigen &> /dev/null
then
    echo "abigen could not be found. Please install it first."
    echo "Run: go install github.com/ethereum/go-ethereum/cmd/abigen@latest"
    exit 1
fi

# Create the output directory if it doesn't exist
mkdir -p "$(dirname "$OUTPUT_PATH")"

# Extract the ABI from the Foundry output file
ABI=$(jq -r '.abi | tojson' $ABI_PATH)

# Generate the Go bindings
abigen --abi - --pkg "$PKG_NAME" --type "$CONTRACT_NAME" --out "$OUTPUT_PATH" <<< "$ABI"

echo "✅ Go bindings generated successfully at: $OUTPUT_PATH" 