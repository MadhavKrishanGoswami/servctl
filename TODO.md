# 📋 servctl - Complete Development TODO List

> **Project:** servctl - Home Server Provisioning CLI Tool  
> **Target OS:** Ubuntu 22.04 LTS+  
> **Language:** Go 1.21+ with Bubble Tea TUI  
> **Created:** 2026-01-30

---

## 🏗️ Phase 0: Project Setup & Foundation

### 0.1 Repository Setup
- [ ] Initialize Go module: `go mod init github.com/madhav/servctl`
- [ ] Create project directory structure:
  ```
  servctl/
  ├── cmd/
  │   └── servctl/
  │       └── main.go
  ├── internal/
  │   ├── tui/           # Bubble Tea UI components
  │   ├── discovery/     # System discovery logic
  │   ├── executor/      # Command execution
  │   ├── generator/     # Config file generation
  │   └── models/        # Data structures
  ├── templates/         # Embedded templates
  │   ├── docker-compose.yml.tmpl
  │   ├── env.tmpl
  │   └── scripts/
  └── go.mod
  ```

### 0.2 Dependencies Installation
- [ ] Add Bubble Tea: `go get github.com/charmbracelet/bubbletea`
- [ ] Add Lip Gloss: `go get github.com/charmbracelet/lipgloss`
- [ ] Add Bubbles components: `go get github.com/charmbracelet/bubbles`
- [ ] Add YAML library: `go get gopkg.in/yaml.v3`

### 0.3 Build System
- [ ] Create `Makefile` with build targets
- [ ] Set up static binary build: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`
- [ ] Add version injection at build time

---

## 🚀 Phase 1: Module A - Pre-flight & Dependency Management

### 1.1 System Prerequisite Checks
- [ ] Create `CheckOS()` function
  - [ ] Verify Ubuntu 22.04 LTS or later
  - [ ] Parse `/etc/os-release` for version info
  - [ ] Abort with clear error if OS check fails

- [ ] Create `CheckPrivileges()` function
  - [ ] Verify non-root user with sudo access
  - [ ] Test sudo capability without password caching issues
  - [ ] Display warning if running as root directly

- [ ] Create `CheckHardware()` function
  - [ ] Check VT-x/AMD-V via `lscpu` (virtualization)
  - [ ] Warn if Secure Boot is enabled
  - [ ] Warn if Serial Ports are enabled

- [ ] Create `CheckConnectivity()` function
  - [ ] Ping external endpoint (e.g., `8.8.8.8`)
  - [ ] Verify DNS resolution
  - [ ] Abort if no internet connection

### 1.2 System Update Module
- [ ] Create `SystemUpdate()` function
  - [ ] Execute `sudo apt update`
  - [ ] Execute `sudo apt upgrade -y`
  - [ ] Show TUI progress spinner during execution
  - [ ] Parse and display packages updated

### 1.3 Dependency Audit & Installation
- [ ] Create `DependencyAudit()` function
  - [ ] Check for `curl` → Install if missing
  - [ ] Check for `net-tools` → Install if missing
  - [ ] Check for `docker` → Install via official script if missing
  - [ ] Check for `docker-compose` → Install if missing
  - [ ] Check for `hdparm` → Install if missing
  - [ ] Check for `smartmontools` → Install if missing
  - [ ] Check for `cron` → Install if missing
  - [ ] Check for `ufw` → Install if missing

- [ ] Create `VerifyDockerRunning()` function
  - [ ] Check `systemctl status docker`
  - [ ] Start docker service if not running
  - [ ] Add current user to docker group

### 1.4 Critical Blocker Checks (From TechStack.md)
- [ ] Verify `docker` binary exists → **Blocker** (abort if missing)
- [ ] Verify `lsblk` binary exists → **Blocker** (abort if missing)
- [ ] Verify `mkfs.ext4` binary exists → **Blocker** (abort if missing)
- [ ] Verify `hdparm` binary exists → **Recommended** (warn if missing)
- [ ] Verify `smartctl` binary exists → **Recommended** (warn if missing)
- [ ] Verify `ufw` binary exists → **High** (warn if missing)

---

## 💾 Phase 2: Module B - Intelligent Storage Orchestration

### 2.1 Disk Discovery
- [ ] Create `DiscoverDisks()` function
  - [ ] Execute `lsblk -J` (JSON output)
  - [ ] Parse JSON into Go structs
  - [ ] Identify disk type (SSD vs HDD) via `lsblk -d -o NAME,ROTA`
  - [ ] Get disk sizes and current mount points
  - [ ] Execute `fdisk -l` for additional partition info

### 2.2 Disk Classification & Recommendation Engine
- [ ] Create `ClassifyDisks()` function
  - [ ] Categorize disks by type (SSD/HDD)
  - [ ] Categorize by size (small/medium/large)
  - [ ] Identify OS disk vs available disks

### 2.3 Single Disk Scenario (Case: 1 Disk Found)
- [ ] Display **Single Point of Failure Warning**
- [ ] Configure partitioning:
  - [ ] OS partition
  - [ ] Data partition on same drive
- [ ] Show TUI confirmation prompt

### 2.4 Two Disk Scenario (Case: 2 Disks Found)
- [ ] Create TUI selection menu for "The 5 Ranks":
  - [ ] **Rank 1 (Hybrid):** SSD (OS/Apps) + HDD (Bulk Data) - *Recommended*
  - [ ] **Rank 2 (Speed Demon):** SSD (OS) + SSD (Active DBs)
  - [ ] **Rank 3 (Mirror):** 2x SSD (RAID 1)
  - [ ] **Rank 4 (Data Hoarder):** 2x HDD (RAID 1) - *Trigger Performance Warning*
  - [ ] **Rank 5 (Kamikaze):** RAID 0 - *Trigger Critical Risk Warning*

### 2.5 Multi-Disk Scenario (Case: 3+ Disks Found)
- [ ] Propose optimal configuration:
  - [ ] OS disk (SSD)
  - [ ] Storage disk (HDD)
  - [ ] Backup disk (HDD)
- [ ] Allow user customization via TUI

### 2.6 Disk Formatting & Mounting
- [ ] Create `FormatDisk()` function
  - [ ] Create TUI filesystem format selection menu (ordered by recommendation):
    ```
    ┌─────────────────────────────────────────────────────────────┐
    │  Select Filesystem Format                                   │
    ├─────────────────────────────────────────────────────────────┤
    │  [1] ext4 (Recommended - Default)    ⭐ PRESS ENTER TO USE  │
    │      ✓ Best stability & compatibility                       │
    │      ✓ Native Linux, proven for 15+ years                   │
    │      ✓ Excellent for SSDs and HDDs                          │
    │                                                             │
    │  [2] XFS (High Performance)                                 │
    │      ✓ Better for large files (media/video)                 │
    │      ✓ Excellent parallel I/O                               │
    │      ⚠ Cannot shrink partitions                             │
    │                                                             │
    │  [3] Btrfs (Advanced Features)                              │
    │      ✓ Snapshots, compression, checksums                    │
    │      ✓ Built-in RAID support                                │
    │      ⚠ More complex, higher overhead                        │
    │                                                             │
    │  [4] ZFS (Enterprise Grade)                                 │
    │      ✓ Maximum data integrity                               │
    │      ✓ Advanced caching & deduplication                     │
    │      ⚠ Requires more RAM (1GB per TB)                       │
    │      ⚠ Not in Linux kernel (license issues)                 │
    └─────────────────────────────────────────────────────────────┘
    ```
  - [ ] **Default Behavior:** Auto-select `ext4` if user presses Enter
  - [ ] Execute appropriate `mkfs` command based on selection:
    - [ ] `mkfs.ext4 -L <label>` for ext4
    - [ ] `mkfs.xfs -L <label>` for XFS
    - [ ] `mkfs.btrfs -L <label>` for Btrfs
    - [ ] `zpool create` for ZFS (if installed)
  - [ ] Confirm destructive action with user (⚠️ ALL DATA WILL BE ERASED)
  - [ ] Implement "Fail Fast" - abort on any error
  - [ ] Store selected filesystem in config for future reference

- [ ] Create `MountDisk()` function
  - [ ] Create mount point directory
  - [ ] Add entry to `/etc/fstab`
  - [ ] **Idempotency:** Check if entry already exists before adding
  - [ ] Execute `mount -a`

### 2.7 Power Optimization
- [ ] Create `ConfigureHDDSpindown()` function
  - [ ] Identify mechanical drives (ROTA=1)
  - [ ] Configure `hdparm` spindown rules
  - [ ] Add to `/etc/hdparm.conf` for persistence

---

## 📁 Phase 3: Module C - Directory & Infrastructure Structure

### 3.1 User Space Directory Creation
- [ ] Create `~/infra/` (Root for configs)
- [ ] Create `~/infra/scripts/` (Maintenance executables)
- [ ] Create `~/infra/logs/` (Centralized logging)

### 3.2 Data Space Directory Creation
- [ ] Create `/mnt/data/` root
- [ ] Create Immich directories:
  - [ ] `/mnt/data/gallery/library`
  - [ ] `/mnt/data/gallery/upload`
  - [ ] `/mnt/data/gallery/profile`
  - [ ] `/mnt/data/gallery/video`
- [ ] Create Nextcloud directory:
  - [ ] `/mnt/data/cloud/data`

### 3.3 Permission Enforcement (Critical Fix from PRD)
- [ ] Create `SetPermissions()` function
  - [ ] Get current user PUID (typically 1000)
  - [ ] Get current user PGID (typically 1000)
  - [ ] Execute `chown -R ${PUID}:${PGID} /mnt/data`
  - [ ] Set appropriate `chmod` (755 for dirs, 644 for files)
  - [ ] Prevent "Permission Denied" boot loops

---

## 🐳 Phase 4: Module D - Service Composition

### 4.1 Input Wizard (Interactive TUI)
- [ ] Create `InputWizard()` TUI component
- [ ] Collect user inputs:
  - [ ] `TZ` (Timezone) - Default: Auto-detect or "UTC"
  - [ ] `PUID/PGID` - Default: Current user (1000/1000)
  - [ ] `DB_PASSWORD` - Auto-generate strong alphanumeric if blank
  - [ ] Discord/Telegram Webhook URLs (for notifications)
  - [ ] Nextcloud Admin Username
  - [ ] Nextcloud Admin Password
  
  > **Note:** `UPLOAD_LOCATION` and data paths are **NOT user-configurable**.  
  > They are auto-derived from Module B (disk selection) and hardcoded to `/mnt/data/`  
  > for consistency. This is intentional — servctl is an **opinionated** tool.

- [ ] Input Validation:
  - [ ] Ensure passwords are not empty
  - [ ] Ensure paths exist or are creatable
  - [ ] Validate webhook URL format

### 4.2 Network Configuration
- [ ] Create `DetectHostIP()` function
  - [ ] Auto-detect primary interface IP
  - [ ] Prompt user to set **Static IP** (Critical from PRD)
  - [ ] Validate IP format
  - [ ] Validate IP range (192.168.x.x or 10.x.x.x)
  - [ ] Validate IP is not already in use
  - [ ] Warn about DHCP lockout risk for Nextcloud

### 4.3 Docker Compose Generation
- [ ] Create `docker-compose.yml.tmpl` template
- [ ] Generate services:
  - [ ] **Immich Server** (Port 2283)
    - [ ] Volume mounts for uploads
    - [ ] ENV file reference
    - [ ] Depends on: redis, database
  - [ ] **Immich Machine Learning**
    - [ ] Model cache volume
    - [ ] ENV file reference
  - [ ] **Nextcloud** (Port 8080)
    - [ ] Configure `OVERWRITEPROTOCOL`
    - [ ] Configure `TRUSTED_DOMAINS` with static IP
    - [ ] Configure `OVERWRITEHOST`
    - [ ] Admin auto-setup via ENV
  - [ ] **Redis/Valkey** (For Immich)
  - [ ] **Glances** (Port 61208)
    - [ ] Host mode networking
    - [ ] SMART data reading capabilities
    - [ ] `SYS_ADMIN` and `SYS_RAWIO` capabilities
  - [ ] **Diun** (Update notifier)
    - [ ] Discord/Telegram webhook integration
    - [ ] Bi-monthly check schedule

### 4.4 Database Architecture (Critical Fix from PRD)
- [ ] Create **ISOLATED** database containers:
  - [ ] **Immich DB:** Dedicated `postgres:14-vectorchord` container
  - [ ] **Nextcloud DB:** Dedicated `mariadb` or `postgres` container
  - [ ] Separate credentials for each database
- [ ] **Rationale:** Prevent "Dependency Hell" during upgrades

### 4.5 ENV File Generation
- [ ] Create `.env.tmpl` template
- [ ] Generate sections:
  - [ ] Shared/Global settings (TZ, PUID, PGID)
  - [ ] Immich DB credentials
  - [ ] Nextcloud DB credentials
  - [ ] Path configurations
  - [ ] Webhook URLs

### 4.6 Firewall Configuration (UFW)
- [ ] Create `ConfigureFirewall()` function
- [ ] **CRITICAL: Lockout Prevention**
  - [ ] Execute `ufw allow ssh` (Port 22) **FIRST**
  - [ ] Verify SSH rule is active
- [ ] Add service rules:
  - [ ] `ufw allow 2283/tcp` (Immich)
  - [ ] `ufw allow 8080/tcp` (Nextcloud)
  - [ ] `ufw allow 61208/tcp` (Glances - Optional/Local only)
- [ ] Execute `ufw enable` only after rule verification
- [ ] Display firewall status to user

---

## 🔧 Phase 5: Module E - Maintenance & Reliability

### 5.1 Script Generation
- [ ] Create `disk_alert.sh` template
  - [ ] Monitor disk usage threshold (e.g., 80%)
  - [ ] Send notification if threshold exceeded
  - [ ] Inject user's webhook URL

- [ ] Create `smart_alert.sh` template
  - [ ] Run `smartctl` health check on all drives
  - [ ] Parse SMART attributes for warnings
  - [ ] Send notification on failure indicators

- [ ] Create `daily_backup.sh` template
  - [ ] Backup specific Docker volumes
  - [ ] Create PostgreSQL/MariaDB dumps
  - [ ] Compress backups with timestamp
  - [ ] Rotate old backups (keep last N)
  - [ ] Send success/failure notification with the % of storage occupied and amount of storage used in the notification

- [ ] Create `weekly_cleanup.sh` template
  - [ ] Prune unused Docker images: `docker image prune -f`
  - [ ] Prune Docker system: `docker system prune -f`
  - [ ] Clean old log files
  - [ ] Send completion notification with the % of storage occupied and amount of storage used in the notification

### 5.2 Cron Orchestration
- [ ] Create `ConfigureCron()` function
- [ ] Interactive scheduling with TUI:
  - [ ] Daily backup time (Default: 4:00 AM)
  - [ ] Weekly cleanup day (Default: Sunday)
  - [ ] Disk alert frequency (Default: Every 6 hours)
  - [ ] SMART check frequency (Default: Daily)
- [ ] **⚠️ SUDO/ROOT REQUIRED:** All cron jobs run with elevated privileges
  - [ ] Use `/etc/cron.d/servctl` (system cron, runs as root)
  - [ ] **NOT** user crontab (`crontab -e`) — insufficient permissions
  - [ ] Scripts require root for:
    - [ ] `docker` commands (unless user is in docker group)
    - [ ] `smartctl` (raw disk access)
    - [ ] Reading `/mnt/data` stats
    - [ ] Writing to system logs
  - [ ] Set proper permissions: `chmod 644 /etc/cron.d/servctl`
  - [ ] Cron file format: `SHELL=/bin/bash` + `PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin`
- [ ] **Idempotency:** Check for existing entries before adding

### 5.3 Notification Integration
- [ ] Create `TestWebhook()` function
  - [ ] Send test notification to Discord/Telegram
  - [ ] Verify webhook is working before finalizing
- [ ] Inject webhook URLs into all maintenance scripts

### 5.4 Log Rotation Configuration (From PRD Gap)
- [ ] Create `/etc/logrotate.d/servctl` config
- [ ] Configure rotation policy:
  - [ ] Rotate weekly
  - [ ] Keep 4 weeks of history
  - [ ] Compress old logs (gzip)
  - [ ] Target: `~/infra/logs/*.log`
- tell them about these configrations

<!-- Future Featurer  -->

<!-- ### 5.5 Offsite Backup (Optional - From PRD Feedback) 
- [ ] Create optional Rclone integration
  - [ ] Configure remote (Google Drive/S3/Backblaze)
  - [ ] Daily push of critical configs/DB dumps
  - [ ] **Rationale:** Local backup ≠ True backup -->

---

## 🖥️ Phase 6: CLI Command Interface

### 6.1 Main Entry Point
- [ ] Create `cmd/servctl/main.go`
- [ ] Implement command routing

### 6.2 Command Implementations
- [ ] `servctl -start-setup`
  - [ ] Launch interactive installation wizard
  - [ ] Execute Modules A through E sequentially
  - [ ] Use Bubble Tea for TUI navigation and make it very user-friendly and pretty

- [ ] `servctl -status`
  - [ ] Display Docker container health status
  - [ ] Show disk usage statistics
  - [ ] Display SMART health summary
  - [ ] Format output with Lip Gloss styling

- [ ] `servctl -get-config`
  - [ ] Read and display current `.env`
  - [ ] Read and display current `docker-compose.yml`
  - [ ] Mask sensitive passwords in output

- [ ] `servctl -get-architecture`
  - [ ] Visualize folder structure tree
  - [ ] Display disk mapping diagram
  - [ ] Show service relationships

- [ ] `servctl -manual-backup`
  - [ ] Trigger immediate `daily_backup.sh` execution
  - [ ] Display real-time progress
  - [ ] Show completion status
- [ ] `servctl -logs`
  - [ ] Display logs of all the services
  - [ ] Make it interactive and pretty

---

## 🌐 Phase 7: Reverse Proxy & HTTPS (Optional Module)

### 7.1 Proxy Selection
- [ ] Create TUI prompt for proxy choice:
  - [ ] **Caddy** (Recommended - Zero-config SSL)
  - [ ] **Nginx Proxy Manager** (GUI-based)
  - [ ] **Skip** (Use raw IPs)

### 7.2 Caddy Configuration
- [ ] Generate `Caddyfile` template
- [ ] Map custom domains:
  - [ ] `immich.local` → `localhost:2283`
  - [ ] `cloud.local` → `localhost:8080`
  - [ ] `monitor.local` → `localhost:61208`
- [ ] Configure Let's Encrypt (if public domain)
- [ ] Generate self-signed certs (for LAN-only)

### 7.3 Local DNS / Hosts File
- [ ] Optionally update `/etc/hosts`
- [ ] Provide instructions for client-side hosts file

---

## 🎯 Phase 8: Non-Functional Requirements

### 8.1 Idempotency Implementation
- [ ] All operations must be re-runnable
- [ ] Check for existing configs before writing
- [ ] Skip duplicate `/etc/fstab` entries
- [ ] Version existing files before overwriting

### 8.2 Error Handling ("Fail Fast")
- [ ] Implement consistent error returns
- [ ] On critical failure:
  - [ ] Stop execution immediately
  - [ ] Display `stderr` log
  - [ ] Suggest remediation steps
- [ ] Log all operations to `~/infra/logs/servctl.log`

### 8.3 Mission Report (UX Handover)
- [ ] Create `DisplayMissionReport()` function
- [ ] Display at completion:
  - [ ] Dashboard URLs:
    - [ ] Immich: `http://<HOST_IP>:2283` (telll them where to put it in the browser/app)
    - [ ] Nextcloud: `http://<HOST_IP>:8080` (telll them where to put it in the browser/app)
    - [ ] Glances: `http://<HOST_IP>:61208` (telll them where to put it in the browser/no app)
  - [ ] Generated credentials (one-time display)
  - [ ] Next steps instructions
  - [ ] How to change default passwords

---

## 🧪 Phase 9: Testing & Validation

### 9.1 Unit Tests
- [ ] Test `DiscoverDisks()` with mock `lsblk` output
- [ ] Test `ClassifyDisks()` logic
- [ ] Test template generation
- [ ] Test input validation

### 9.2 Integration Tests
- [ ] Test full workflow on VM
- [ ] Test idempotency (run twice, verify no duplicates)
- [ ] Test error handling (simulate failures)

### 9.3 Manual Testing Scenarios
- [ ] Test on fresh Ubuntu 22.04 install
- [ ] Test with 1 disk, 2 disks, 3+ disks
- [ ] Test firewall configuration (verify no lockout)
- [ ] Test notification webhooks

---

## 📦 Phase 10: Distribution & Documentation

### 10.1 Build & Release
- [ ] Set up GitHub Actions for CI/CD
- [ ] Build static binary for Linux AMD64
- [ ] Create GitHub Release with binary
- [ ] Generate SHA256 checksum

### 10.2 Documentation
- [ ] Write comprehensive README.md
- [ ] Document all CLI commands
- [ ] Create troubleshooting guide
- [ ] Add architecture diagrams

### 10.3 Installation Script
- [ ] Create one-liner installation:
  ```bash
  wget https://github.com/madhav/servctl/releases/latest/download/servctl
  chmod +x servctl
  sudo ./servctl
  ```

---

## 🗓️ Future Roadmap (v1.1+)

### Web Dashboard
- [ ] Lightweight GUI wrapping `servctl -status`
- [ ] Log viewing interface
- [ ] Container management

### One-Click Restore
- [ ] Implement `servctl -restore <backup_file>`
- [ ] Rebuild infrastructure from backup
- [ ] Restore database dumps

### Remote Access
- [ ] Interactive Cloudflare Tunnels setup
- [ ] Tailscale VPN integration
- [ ] WireGuard configuration

---

## 📊 Progress Tracker

| Phase | Status | Progress |
|-------|--------|----------|
| Phase 0: Project Setup | ⬜ Not Started | 0% |
| Phase 1: Pre-flight | ⬜ Not Started | 0% |
| Phase 2: Storage | ⬜ Not Started | 0% |
| Phase 3: Directory Structure | ⬜ Not Started | 0% |
| Phase 4: Service Composition | ⬜ Not Started | 0% |
| Phase 5: Maintenance | ⬜ Not Started | 0% |
| Phase 6: CLI Interface | ⬜ Not Started | 0% |
| Phase 7: Reverse Proxy | ⬜ Not Started | 0% |
| Phase 8: Non-Functional | ⬜ Not Started | 0% |
| Phase 9: Testing | ⬜ Not Started | 0% |
| Phase 10: Distribution | ⬜ Not Started | 0% |

---

> **Legend:**  
> ⬜ Not Started | 🟡 In Progress | ✅ Complete | ❌ Blocked
