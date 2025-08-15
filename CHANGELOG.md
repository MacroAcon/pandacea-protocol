# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Pre-commit hooks** with ruff, black, mypy, and golangci-lint for code quality
- **Go HTTP server hardening** with healthz, readyz, metrics endpoints and graceful shutdown
- **Rate limiting** with token bucket algorithm for abuse prevention
- **Python SDK typing** with Pydantic models and structured error handling
- **Solidity custom errors** replacing require strings for gas optimization
- **ABI export system** with scripts for Windows and Unix systems
- **ABI compatibility testing** to prevent accidental contract breaks

### Changed
- Enhanced `.gitignore` to exclude build outputs while preserving ABI exports
- Updated Solidity contracts to use custom errors instead of require strings

### Development
- Added comprehensive development setup instructions in README
- Created `requirements-dev.txt` for Python development dependencies
- Added health endpoint documentation and testing
- Documented ABI management and compatibility testing procedures

## [0.1.0] - 2024-01-01
- Initial release
