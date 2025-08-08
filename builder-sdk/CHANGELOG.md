# Changelog

All notable changes to the Pandacea SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2024-12-19

### Added
- Reproducible build support with deterministic wheel generation
- SBOM (Software Bill of Materials) generation for security compliance
- Enhanced telemetry and observability integration
- Comprehensive test coverage for all SDK components
- Windows-first PowerShell build scripts

### Changed
- **BREAKING**: Pinned all dependencies to exact versions for reproducible builds
- Updated to use Poetry for dependency management
- Improved error handling and exception types
- Enhanced documentation and type hints

### Fixed
- Resolved dependency version conflicts
- Fixed Windows path handling in build scripts
- Improved error messages for better debugging

### Security
- Added SBOM generation for supply chain security
- Pinned all dependencies to prevent supply chain attacks
- Enhanced cryptographic operations

### Technical
- Migrated to Poetry for dependency management
- Added reproducible build infrastructure
- Implemented deterministic wheel generation
- Added comprehensive CI/CD pipeline support

## [0.1.0] - 2024-11-01

### Added
- Initial SDK release
- Basic client implementation for Pandacea Agent API
- Support for data product operations
- Telemetry integration
- Basic error handling and exceptions

[0.3.0]: https://github.com/pandacea/pandacea-protocol/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/pandacea/pandacea-protocol/releases/tag/v0.1.0
