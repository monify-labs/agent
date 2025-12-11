# Changelog

All notable changes to the Monify Agent will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.3] - 2025-12-11

### Fixed
- Fixed metric accuracy for CPU and Memory by implementing continuous sampling (1s interval) with averaging. This aligns agent metrics with cloud provider dashboards (like DigitalOcean) by capturing spikes and fluctuating loads that were previously missed by snapshot-based collection.
- Refactored `Collector` interface to support graceful shutdown of background sampling goroutines.


## [0.2.2] - 2025-12-11

### Fixed
- Corrected CPU core and thread counting logic for accurate hardware reporting on all architectures
- Filtered out disk partitions with 0 total size (e.g., pseudo-filesystems, empty mounts) to prevent noise in metrics
- Fixed `monify status` showing "Not authenticated" by correctly loading environment variables
- Added check in installation script to prevent installing over a running agent instance

## [0.2.0] - 2025-12-11

### Added
- Initial release of Monify Agent
- Linux-only support (Ubuntu, Debian, CentOS, RHEL, Fedora, etc.)
- System metrics collection (CPU, Memory, Disk, Network, System)
- Token-based authentication
- Systemd service integration
- Automatic installation script
- Port scanning capability
- Remote command execution (refresh, scan_ports, update_config, stop, uninstall)
- Gzip compression for efficient data transfer (~70-80% compression)
- Single-instance lock mechanism
- Comprehensive logging with configurable levels
- Configuration via YAML and environment variables
- Comprehensive documentation (README, CHANGELOG, CONTRIBUTING, SECURITY, API)
- Enhanced Makefile with multiple build targets
- Multi-architecture support (AMD64, ARM64)

### Performance
- CPU Usage: < 1%
- Memory Usage: < 50 MB
- Bandwidth: ~10-20 MB/month
- Optimized metric collection with 30s default interval

### Security
- HTTPS-only communication
- Token-based authentication
- No sensitive data collection
- Open source and auditable code
- Secure configuration file permissions

[Unreleased]: https://github.com/monify-labs/agent/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/monify-labs/agent/releases/tag/v0.2.0
