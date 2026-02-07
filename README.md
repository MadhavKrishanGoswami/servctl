# servctl

**Home Server Provisioning CLI** — Transform any Ubuntu machine into a fully configured home server with Nextcloud, Immich, monitoring, and automated backups.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://github.com/madhav/servctl/actions/workflows/test.yml/badge.svg)](https://github.com/madhav/servctl/actions)

---

## ✨ Features

- **🔧 5-Phase Setup Wizard** — Interactive CLI guides you through complete server setup
- **💾 Smart Storage Management** — Auto-detects disks and recommends RAID/pooling strategies
- **🐳 Docker Compose Generation** — Creates production-ready configs for all services
- **📊 System Monitoring** — Pre-configured Glances dashboard
- **🔔 Discord Notifications** — Automated alerts for backups, disk health, and updates
- **🛡️ Preflight Checks** — Validates system requirements with auto-fix capabilities

---

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [CLI Reference](#-cli-reference)
- [Setup Wizard](#-setup-wizard)
- [Services Included](#-services-included)
- [Storage Strategies](#-storage-strategies)
- [Directory Structure](#-directory-structure)
- [Maintenance Scripts](#-maintenance-scripts)
- [Development](#-development)
- [Contributing](#-contributing)

---

## 🚀 Quick Start

```bash
# Download and run
curl -LO https://github.com/madhav/servctl/releases/latest/download/servctl
chmod +x servctl

# Preview what will happen
sudo ./servctl -dry-run -start-setup

# Run the setup wizard
sudo ./servctl -start-setup
```

---

## 📦 Installation

### From Source

```bash
git clone https://github.com/madhav/servctl.git
cd servctl
make build-local
./bin/servctl-local -version
```

### From Release

Download the latest binary from [Releases](https://github.com/madhav/servctl/releases).

### Requirements

- **OS**: Ubuntu 22.04+ (tested on Ubuntu 24.04)
- **RAM**: 4GB minimum, 8GB+ recommended
- **Disk**: 20GB+ for OS, additional storage for data
- **Docker**: Installed automatically if missing

---

## 📖 CLI Reference

### Commands

| Command | Description |
|---------|-------------|
| `servctl -start-setup` | Launch interactive 5-phase setup wizard |
| `servctl -preflight` | Run system checks without making changes |
| `servctl -status` | Display Docker containers, disk usage, SMART health |
| `servctl -get-config` | Show current .env configuration (passwords masked) |
| `servctl -get-architecture` | Display directory structure and service diagram |
| `servctl -manual-backup` | Trigger immediate backup sync |
| `servctl -logs` | Tail Docker Compose logs (Ctrl+C to exit) |
| `servctl -version` | Display version, build time, and system info |

### Options

| Option | Description |
|--------|-------------|
| `-dry-run` | Preview all changes without executing them |

### Examples

```bash
# Check system readiness
sudo servctl -preflight

# Preview full setup
sudo servctl -dry-run -start-setup

# Run complete installation
sudo servctl -start-setup

# Monitor system
servctl -status

# View service logs
servctl -logs
```

---

## 🧙 Setup Wizard

The `-start-setup` command launches an interactive wizard with 5 phases:

### Phase 1: System Preparation
- Checks for root/sudo access
- Validates Docker installation
- Detects network configuration
- Offers static IP setup for DHCP systems
- Auto-installs missing dependencies

### Phase 2: Storage Configuration
- Discovers all connected disks (HDD, SSD, NVMe)
- Analyzes disk sizes, types, and current usage
- Recommends optimal storage strategies:
  - **Simple Partition** — Single disk, ext4 formatted
  - **MergerFS Pool** — Combine multiple disks into one mount
  - **Mirror (RAID1)** — ZFS or MDADM mirroring for redundancy
- Configures automatic disk mounting via `/etc/fstab`

### Phase 3: Directory Structure
- Creates organized folder hierarchy:
  ```
  ~/infra/           # Configuration files
  /mnt/data/         # User data (Nextcloud, Immich, etc.)
  ```
- Sets proper ownership and permissions
- Supports customization of paths

### Phase 4: Service Configuration
- Generates `docker-compose.yml` with all services
- Creates `.env` file with secure random passwords
- Configures networking and volume mounts
- Detects host IP for service URLs

### Phase 5: Maintenance Scripts
- Generates shell scripts for:
  - Daily backup (rsync with Discord notifications)
  - Disk space alerts (threshold-based)
  - SMART health monitoring
  - Weekly Docker cleanup
- Sets up cron jobs for automation

---

## 🐳 Services Included

| Service | Port | Description |
|---------|------|-------------|
| **Nextcloud** | 8080 | File sync, calendar, office suite |
| **Immich** | 2283 | Photo/video library (Google Photos alternative) |
| **PostgreSQL** | - | Database for Nextcloud and Immich |
| **Redis** | - | Caching layer |
| **Glances** | 61208 | Real-time system monitoring |
| **Diun** | - | Docker image update notifications |

---

## 💾 Storage Strategies

servctl analyzes your hardware and recommends the best storage approach:

| Strategy | Disks | Use Case | Features |
|----------|-------|----------|----------|
| **Simple** | 1 | Basic setup | ext4, mount to `/mnt/data` |
| **MergerFS** | 2+ | Maximum capacity | Pools disks, expandable |
| **Mirror** | 2 | Data protection | ZFS/MDADM RAID1, 50% capacity |
| **Tiered** | Mixed SSD+HDD | Performance + capacity | SSD cache, HDD storage |

Each strategy includes:
- Automatic formatting and mounting
- HDD spindown configuration (30-minute default)
- Backup path configuration
- fstab entries for boot persistence

---

## 📁 Directory Structure

After setup, your server will have:

```
~/infra/
├── compose/
│   ├── docker-compose.yml    # Main service definitions
│   └── .env                  # Environment configuration
├── scripts/
│   ├── daily_backup.sh       # Rsync backup script
│   ├── disk_alert.sh         # Space monitoring
│   ├── smart_alert.sh        # Drive health checks
│   └── weekly_cleanup.sh     # Docker/system cleanup
└── logs/
    └── servctl.log           # Setup and operation logs

/mnt/data/
├── nextcloud/
│   ├── data/                 # User files
│   └── config/               # App configuration
├── immich/
│   ├── upload/               # Photo uploads
│   ├── library/              # Processed photos
│   └── thumbs/               # Thumbnails
└── databases/
    ├── postgres/             # PostgreSQL data
    └── redis/                # Redis persistence
```

---

## 🔧 Maintenance Scripts

All scripts support Discord webhook notifications:

### Daily Backup (`daily_backup.sh`)
```bash
# Runs at 2 AM daily via cron
# Syncs /mnt/data → /mnt/backup with rsync
# Sends success/failure notification to Discord
```

### Disk Alert (`disk_alert.sh`)
```bash
# Runs every 6 hours
# Alerts when disk usage > 90%
```

### SMART Monitor (`smart_alert.sh`)
```bash
# Runs daily
# Checks S.M.A.R.T. status of all drives
# Alerts on failing health status
```

### Weekly Cleanup (`weekly_cleanup.sh`)
```bash
# Runs Sunday at 3 AM
# Cleans apt cache
# Prunes dangling Docker images
# Truncates large log files
```

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Docker (for integration tests)

### Build

```bash
# Build for current OS
make build-local

# Build for Linux (production)
make build

# Run tests
make test-short

# Run with coverage
make test-coverage
```

### Testing

```bash
# Unit tests (fast, works on macOS)
go test ./internal/... -short

# Full tests in Docker (Linux environment)
make docker-test

# Interactive Docker shell
make docker-shell
```

### Project Structure

```
servctl/
├── cmd/servctl/        # CLI entry point
├── internal/
│   ├── compose/        # Docker Compose generation
│   ├── directory/      # Directory structure creation
│   ├── maintenance/    # Maintenance script generation
│   ├── preflight/      # System requirement checks
│   ├── report/         # Mission report rendering
│   ├── storage/        # Disk discovery and configuration
│   ├── tui/            # Terminal UI components
│   └── utils/          # Logging and helpers
├── templates/          # Compose and script templates
├── scripts/            # Development/test scripts
├── build/              # Docker configs
└── docs/               # Documentation
```

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing`
3. Make changes with tests
4. Run: `make test-short && go vet ./...`
5. Commit: `git commit -m "feat: add amazing feature"`
6. Push: `git push origin feature/amazing`
7. Open a Pull Request

---

## 📜 License

MIT License — see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [Charm](https://charm.sh) — Beautiful terminal UI library
- [Immich](https://immich.app) — Self-hosted photo management
- [Nextcloud](https://nextcloud.com) — File sync and collaboration
- [MergerFS](https://github.com/trapexit/mergerfs) — Union filesystem
