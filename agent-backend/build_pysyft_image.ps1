# Build PySyft Datasite Docker Image for Pandacea Protocol
# This script builds the Docker image used in the warm container pool

Write-Host "🔧 Building PySyft Datasite Docker Image" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Check if Docker is available
try {
    $dockerVersion = docker --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Docker not found"
    }
    Write-Host "✅ Docker found: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

# Build the image
Write-Host "📦 Building Docker image..." -ForegroundColor White
docker build -f Dockerfile.pysyft -t pandacea/pysyft-datasite:latest .

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Docker image built successfully!" -ForegroundColor Green
    Write-Host "   Image: pandacea/pysyft-datasite:latest" -ForegroundColor Gray
    Write-Host ""
    Write-Host "🚀 The image is ready for use in the container pool." -ForegroundColor White
    Write-Host "   The PrivacyService will automatically use this image." -ForegroundColor Gray
} else {
    Write-Host "❌ Failed to build Docker image" -ForegroundColor Red
    exit 1
}
