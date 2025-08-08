# SDK Stability and Versioning Policy

## Overview

The Pandacea SDK follows [Semantic Versioning](https://semver.org/) (SemVer) to ensure predictable and stable releases for our users.

## Version Format

Versions follow the format: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes that require code modifications
- **MINOR**: New features added in a backward-compatible manner
- **PATCH**: Bug fixes and minor improvements

## Public API Surface

### Stable APIs (v1.0+)

The following APIs are considered stable and will maintain backward compatibility:

#### Core Client
```python
from pandacea_sdk import PandaceaClient

# Core client initialization
client = PandaceaClient(base_url="http://localhost:8080")
```

#### Data Product Operations
```python
# Data product creation and management
data_product = client.create_data_product(...)
products = client.list_data_products()
product = client.get_data_product(product_id)
```

#### Error Handling
```python
from pandacea_sdk import PandaceaException, AgentConnectionError, APIResponseError
```

#### Telemetry
```python
from pandacea_sdk import telemetry_init
telemetry_init()
```

### Experimental APIs

APIs marked as experimental may change between minor versions. These are clearly documented with `@experimental` decorators or warnings.

## Deprecation Policy

1. **Deprecation Notice**: APIs will be marked as deprecated for at least one major version before removal
2. **Migration Guide**: Clear migration paths will be provided for deprecated APIs
3. **Breaking Changes**: Only introduced in major versions with comprehensive migration documentation

## Version Support

| Version | Status | Support Until |
|---------|--------|---------------|
| 0.3.x   | Current | TBD |
| 0.2.x   | Deprecated | 2025-06-30 |
| 0.1.x   | EOL | 2024-12-31 |

## Reproducible Builds

Starting with v0.3.0, all SDK releases are built with reproducible build techniques:

- **Deterministic Dependencies**: All dependencies are pinned to exact versions
- **SBOM Generation**: Software Bill of Materials generated for each release
- **Build Verification**: Automated checks ensure reproducible artifacts

### Building from Source

```powershell
# Windows (PowerShell)
.\scripts\repro\python_repro_build.ps1

# Or using Make
make repro-build-python
```

## Migration Guide

### From v0.1.x to v0.3.0

#### Breaking Changes
- All dependencies are now pinned to exact versions
- Poetry is required for development builds

#### Required Actions
1. Update your `pyproject.toml` or `requirements.txt` to use exact versions
2. Install Poetry if not already installed: `pip install poetry`
3. Run `poetry install` to install dependencies

#### Code Changes
No code changes required - this is a patch-level update focused on build reproducibility.

## Support and Feedback

- **Issues**: [GitHub Issues](https://github.com/pandacea/pandacea-protocol/issues)
- **Discussions**: [GitHub Discussions](https://github.com/pandacea/pandacea-protocol/discussions)
- **Documentation**: [SDK Documentation](https://docs.pandacea.org/sdk)

## Release Process

1. **Development**: Features developed in feature branches
2. **Testing**: Comprehensive test suite runs on all platforms
3. **Release Candidate**: RC builds tested by community
4. **Release**: Tagged and published to PyPI
5. **Documentation**: Updated docs and migration guides

## Security

- **Vulnerability Reports**: security@pandacea.org
- **SBOM**: Generated for every release in `artifacts/sbom/`
- **Dependency Scanning**: Automated vulnerability scanning in CI/CD
