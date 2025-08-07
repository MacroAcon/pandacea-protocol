#!/bin/bash

# Build PySyft Datasite Docker Image for Pandacea Protocol
# This script builds the Docker image used in the warm container pool

set -e

echo "🔧 Building PySyft Datasite Docker Image"
echo "========================================"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

# Build the image
echo "📦 Building Docker image..."
docker build -f Dockerfile.pysyft -t pandacea/pysyft-datasite:latest .

if [ $? -eq 0 ]; then
    echo "✅ Docker image built successfully!"
    echo "   Image: pandacea/pysyft-datasite:latest"
    echo ""
    echo "🚀 The image is ready for use in the container pool."
    echo "   The PrivacyService will automatically use this image."
else
    echo "❌ Failed to build Docker image"
    exit 1
fi
