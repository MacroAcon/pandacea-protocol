# Windows Quickstart Guide

This guide provides step-by-step instructions for setting up and running the Pandacea Protocol on Windows 10/11.

## Prerequisites

- Windows 10 or 11
- PowerShell (included with Windows)
- Docker Desktop (download from [docker.com](https://www.docker.com/products/docker-desktop/))
- Git (download from [git-scm.com](https://git-scm.com/download/win))

## Step-by-Step Setup

### 1. Open PowerShell
Press `Win + X` and select "Windows PowerShell" or "Terminal"

### 2. Navigate to Project Directory
```powershell
cd C:\Users\thnxt\Documents\pandacea-protocol
```

### 3. Install Foundry
```powershell
iwr https://foundry.paradigm.xyz | Invoke-Expression
```

### 4. Update Foundry
```powershell
foundryup
```

### 5. Verify Foundry Installation
```powershell
forge --version
cast --version
```

### 6. Verify Docker Installation
```powershell
docker --version
```

### 7. Run Contract Tests
```powershell
make contracts-test
```

### 8. Run Coverage Check
```powershell
make contracts-coverage
```

### 9. Run Full Verification
```powershell
make verify
```

## Running Real PySyft via Docker

For users who want to run real PySyft training (instead of mock mode):

### 1. Start PySyft Worker Container
```powershell
cd C:\Users\thnxt\Documents\pandacea-protocol
make pysyft-up
```

### 2. Set Docker Environment Variable
```powershell
setx USE_DOCKER 1
```

### 3. Close and Reopen PowerShell
Close the current PowerShell window and open a new one to load the environment variable.

### 4. Run Demo with Real PySyft
```powershell
cd C:\Users\thnxt\Documents\pandacea-protocol
make demo-real-docker
```

## Troubleshooting

### Foundry Installation Issues
If Foundry installation fails:
```powershell
# Try running PowerShell as Administrator
# Or use the manual installation method:
git clone https://github.com/foundry-rs/foundry.git
cd foundry
cargo build --release
```

### Docker Issues
- Ensure Docker Desktop is running
- Check Docker service: `Get-Service docker`
- Restart Docker Desktop if needed

### Permission Issues
If you encounter permission errors:
```powershell
# Set execution policy (run as Administrator)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Path Issues
If commands are not found:
```powershell
# Refresh environment variables
refreshenv
# Or restart PowerShell
```

## Development Workflow

### Running Tests
```powershell
# Contract tests only
make contracts-test

# Coverage check
make contracts-coverage

# All tests
make verify
```

### Building Docker Images
```powershell
# Build PySyft image
make pysyft-build

# Build agent backend
cd agent-backend
make build
```

### Running Demos
```powershell
# Mock demo (fast)
make demo

# Real PySyft demo (requires Docker)
make demo-real-docker
```

## Environment Variables

Key environment variables for Windows:

- `USE_DOCKER=1`: Enable Docker execution for PySyft
- `MOCK_DP=1`: Use mock differential privacy (default)
- `MOCK_DP=0`: Use real PySyft (requires Docker)

## File Paths

Important Windows-specific paths:
- Project root: `C:\Users\thnxt\Documents\pandacea-protocol`
- Data directory: `C:\Users\thnxt\Documents\pandacea-protocol\data`
- Contract artifacts: `C:\Users\thnxt\Documents\pandacea-protocol\contracts\out`

## Next Steps

After completing the quickstart:
1. Read the [Privacy Boundary Documentation](PRIVACY_BOUNDARY.md)
2. Explore the [API Specification](../docs/Pandacea%20Protocol%20-%20API%20Specification%20(v1.1).pdf)
3. Review the [System Design Document](../docs/Pandacea%20Protocol%20-%20System%20Design%20Document%20(SDD)%20v1.4.md)

## Support

For issues specific to Windows:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review Docker Desktop logs
3. Check PowerShell execution policy
4. Ensure all prerequisites are installed correctly
