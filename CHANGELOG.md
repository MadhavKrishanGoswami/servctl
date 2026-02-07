# Changelog

All notable changes to servctl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Comprehensive documentation (README, CONTRIBUTING, LICENSE)
- Error path tests for all modules
- Fuzz tests for storage operations
- Benchmark tests for performance-critical functions
- Snapshot tests for output stability
- GitHub Actions CI workflow
- Docker testing infrastructure

### Fixed
- Path sanitization bug in directory generation
- Shellcheck warnings in all scripts

### Changed
- Reorganized project structure (moved build files to `build/`, planning docs to `docs/planning/`)
- Enhanced `.gitignore` with comprehensive patterns
- Improved test assertions to match actual API behavior

---

## [0.1.0] - 2024-XX-XX

### Added
- Initial release
- 5-phase interactive setup wizard
- Disk discovery and smart strategy recommendations
- Docker Compose generation for Nextcloud, Immich, Glances
- Maintenance script generation with Discord notifications
- Preflight checks with auto-fix capabilities
- Beautiful terminal UI with Lipgloss
- Dry-run mode for safe previews
- System status command
- Configuration display command
- Architecture visualization
- Manual backup trigger
- Service log viewer

### Supported Storage Strategies
- Simple single-disk setup
- MergerFS disk pooling
- ZFS/MDADM mirroring (RAID1)
- Tiered SSD + HDD configurations

### Supported Services
- Nextcloud (file sync, office)
- Immich (photo/video library)
- PostgreSQL (database)
- Redis (caching)
- Glances (monitoring)
- Diun (update notifications)

---

## Future Roadmap

### v0.2.0 (Planned)
- [ ] Caddy reverse proxy integration
- [ ] Let's Encrypt SSL automation
- [ ] Backup to remote storage (S3, B2)
- [ ] Health check dashboard

### v0.3.0 (Planned)
- [ ] Portainer integration
- [ ] Watchtower auto-updates
- [ ] Multi-user management
- [ ] Web UI for configuration

---

[Unreleased]: https://github.com/madhav/servctl/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/madhav/servctl/releases/tag/v0.1.0
