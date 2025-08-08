# Reproducible Builds Implementation

## Overview

This document describes the reproducible builds implementation for the Pandacea Protocol, ensuring deterministic builds across Go, Python, and container components.

## Architecture

### Components

1. **Go Agent Backend** - Deterministic binary builds with SBOM generation
2. **Python SDK** - Reproducible wheel builds with pinned dependencies
3. **Container Images** - Deterministic Docker builds with OCI labels
4. **CI/CD Integration** - Automated reproducibility verification
5. **SBOM Generation** - Software Bill of Materials for security compliance

## Implementation Details

### Go Reproducible Builds

**Key Features:**
- Pinned Go toolchain version (1.24.2)
- Deterministic build flags: `-trimpath -buildvcs=false`
- Stripped binary with `-ldflags="-s -w"`
- Cross-platform builds (Windows, Linux, macOS)

**Build Process:**
```bash
# Reproducible Go build
make repro-build-go

# Manual build with deterministic flags
cd agent-backend
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
go build -trimpath -buildvcs=false -ldflags="-s -w" \
-o ../artifacts/bin/agent-backend ./cmd/agent
```

**SBOM Generation:**
```bash
# Generate Go SBOM
cd agent-backend
syft . -o spdx-json -o ../artifacts/sbom/agent-backend.spdx.json
```

### Python SDK Reproducible Builds

**Key Features:**
- Pinned dependencies to exact versions
- Deterministic wheel generation with `PYTHONHASHSEED=0`
- Poetry-based dependency management
- Version freeze at v0.3.0

**Build Process:**
```bash
# Reproducible Python build
make repro-build-python

# Manual build with deterministic flags
cd builder-sdk
PYTHONHASHSEED=0 SOURCE_DATE_EPOCH=0 \
poetry build --format wheel --no-interaction
```

**Dependency Management:**
```toml
[tool.poetry.dependencies]
python = "^3.8"
requests = "2.31.0"  # Exact version
pydantic = "2.5.0"   # Exact version
cryptography = "41.0.0"  # Exact version
# ... other dependencies pinned
```

### Container Reproducible Builds

**Key Features:**
- Pinned base image digests
- OCI labels for metadata
- Build arguments for versioning
- Multi-stage builds for optimization

**Build Process:**
```bash
# Reproducible container build
make repro-build-container

# Manual build with deterministic flags
docker build \
  --build-arg VERSION_SHA=repro \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  --build-arg VCS_REF=$(git rev-parse HEAD) \
  -f agent-backend/Dockerfile \
  -t pandacea/agent-backend:repro \
  agent-backend
```

**OCI Labels:**
```dockerfile
LABEL org.opencontainers.image.title="Pandacea Agent Backend" \
      org.opencontainers.image.description="P2P node and HTTP API server for the Pandacea Protocol" \
      org.opencontainers.image.vendor="Pandacea Protocol" \
      org.opencontainers.image.version="${VERSION_SHA:-latest}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/pandacea/pandacea-protocol" \
      org.opencontainers.image.licenses="MIT"
```

## CI/CD Integration

### Reproducibility Workflow

The `.github/workflows/reproducibility.yml` workflow provides:

1. **Automated Verification** - Builds components twice and compares checksums
2. **SBOM Drift Detection** - Monitors for unexpected changes in dependencies
3. **Cross-Platform Testing** - Ensures reproducibility on Windows, Linux, macOS
4. **Artifact Storage** - Preserves build artifacts for audit trails

### Verification Process

```bash
# Run full reproducibility verification
make repro-verify

# Individual component verification
powershell -ExecutionPolicy Bypass -File scripts/repro/verify_reproducibility.ps1
```

## Usage

### Quick Start

1. **Build all reproducible artifacts:**
   ```bash
   make repro-build
   ```

2. **Verify reproducibility:**
   ```bash
   make repro-verify
   ```

3. **Generate SBOMs:**
   ```bash
   make sbom
   ```

### Individual Components

**Go Binary:**
```bash
make repro-build-go
```

**Python Wheel:**
```bash
make repro-build-python
```

**Container Image:**
```bash
make repro-build-container
```

### PowerShell Scripts

**Go Build:**
```powershell
.\scripts\repro\go_repro_build.ps1
```

**Python Build:**
```powershell
.\scripts\repro\python_repro_build.ps1
```

**Container Build:**
```powershell
.\scripts\repro\container_repro_build.ps1
```

**Full Verification:**
```powershell
.\scripts\repro\verify_reproducibility.ps1
```

## Artifacts Structure

```
artifacts/
├── bin/
│   └── agent-backend          # Reproducible Go binary
├── dist/
│   └── pandacea_sdk-0.3.0-py3-none-any.whl  # Reproducible Python wheel
└── sbom/
    ├── agent-backend.spdx.json              # Go SBOM
    ├── pandacea-sdk.spdx.json               # Python SBOM
    └── agent-backend-container.spdx.json    # Container SBOM
```

## Security Considerations

### SBOM Benefits

1. **Dependency Transparency** - Complete visibility into all dependencies
2. **Vulnerability Scanning** - Automated security scanning of dependencies
3. **License Compliance** - Tracking of open source licenses
4. **Audit Trail** - Immutable record of build components

### Reproducibility Security

1. **Deterministic Builds** - Ensures identical outputs from identical inputs
2. **Pinned Dependencies** - Prevents supply chain attacks
3. **Digest Verification** - Validates artifact integrity
4. **CI/CD Integration** - Automated security checks

## Troubleshooting

### Common Issues

**Go Build Issues:**
- Ensure Go toolchain version matches `go.mod`
- Verify `CGO_ENABLED=0` for cross-platform builds
- Check for non-deterministic build flags

**Python Build Issues:**
- Ensure Poetry is installed and configured
- Verify `PYTHONHASHSEED=0` environment variable
- Check for unpinned dependencies in `pyproject.toml`

**Container Build Issues:**
- Verify base image digests are correct
- Ensure build arguments are properly set
- Check for non-deterministic build steps

### Debugging

**Enable Verbose Output:**
```bash
make repro-verify VERBOSE=1
```

**Check Build Logs:**
```bash
# Go build logs
cd agent-backend && go build -v -trimpath -buildvcs=false -ldflags="-s -w" ./cmd/agent

# Python build logs
cd builder-sdk && poetry build --format wheel -v

# Container build logs
docker build -f agent-backend/Dockerfile -t test:latest agent-backend --progress=plain
```

## Future Enhancements

1. **Multi-Architecture Support** - ARM64, ARMv7 builds
2. **Signing Integration** - GPG signing of artifacts
3. **Distributed Builds** - Remote build verification
4. **Advanced SBOM** - CycloneDX format support
5. **Compliance Reporting** - Automated compliance reports

## References

- [Reproducible Builds](https://reproducible-builds.org/)
- [SPDX Specification](https://spdx.dev/)
- [OCI Image Specification](https://github.com/opencontainers/image-spec)
- [Go Build Constraints](https://golang.org/pkg/go/build/)
- [Poetry Documentation](https://python-poetry.org/docs/)
