# 📋 servctl - Complete Development TODO List

> **Project:** servctl - Home Server Provisioning CLI Tool  
> **Target OS:** Ubuntu 22.04 LTS+  
> **Language:** Go 1.21+ with Bubble Tea TUI  
> **Created:** 2026-01-30

---

## 🏗️ Phase 0: Project Setup & Foundation

### 0.1 Repository Setup
- [x] Initialize Go module: `go mod init github.com/madhav/servctl`
- [x] Create project directory structure:
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
- [x] Add Bubble Tea: `go get github.com/charmbracelet/bubbletea`
- [x] Add Lip Gloss: `go get github.com/charmbracelet/lipgloss`
- [x] Add Bubbles components: `go get github.com/charmbracelet/bubbles`
- [x] Add YAML library: `go get gopkg.in/yaml.v3`

### 0.3 Build System
- [x] Create `Makefile` with build targets
- [x] Set up static binary build: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`
- [x] Add version injection at build time

---

## 🚀 Phase 1: Module A - Pre-flight & Dependency Management

### 1.1 System Prerequisite Checks
- [x] Create `CheckOS()` function
  - [x] Verify Ubuntu 22.04 LTS or later
  - [x] Parse `/etc/os-release` for version info
  - [x] Abort with clear error if OS check fails

- [x] Create `CheckPrivileges()` function
  - [x] Verify non-root user with sudo access
  - [x] Test sudo capability without password caching issues
  - [x] Display warning if running as root directly

- [x] Create `CheckHardware()` function
  - [x] Check VT-x/AMD-V via `lscpu` (virtualization)
  - [x] Warn if Secure Boot is enabled
  - [x] Warn if Serial Ports are enabled

- [x] Create `CheckConnectivity()` function
  - [x] Ping external endpoint (e.g., `8.8.8.8`)
  - [x] Verify DNS resolution
  - [x] Abort if no internet connection

### 1.2 System Update Module
- [x] Create `SystemUpdate()` function
  - [x] Execute `sudo apt update`
  - [x] Execute `sudo apt upgrade -y`
  - [x] Show TUI progress spinner during execution
  - [x] Parse and display packages updated

### 1.3 Dependency Audit & Installation
- [x] Create `DependencyAudit()` function
  - [x] Check for `curl` → Install if missing
  - [x] Check for `net-tools` → Install if missing
  - [x] Check for `docker` → Install via official script if missing
  - [x] Check for `docker-compose` → Install if missing
  - [x] Check for `hdparm` → Install if missing
  - [x] Check for `smartmontools` → Install if missing
  - [x] Check for `cron` → Install if missing
  - [x] Check for `ufw` → Install if missing

- [x] Create `VerifyDockerRunning()` function
  - [x] Check `systemctl status docker`
  - [x] Start docker service if not running
  - [x] Add current user to docker group

### 1.4 Critical Blocker Checks (From TechStack.md)
- [x] Verify `docker` binary exists → **Blocker** (abort if missing)
- [x] Verify `lsblk` binary exists → **Blocker** (abort if missing)
- [x] Verify `mkfs.ext4` binary exists → **Blocker** (abort if missing)
- [x] Verify `hdparm` binary exists → **Recommended** (warn if missing)
- [x] Verify `smartctl` binary exists → **Recommended** (warn if missing)
- [x] Verify `ufw` binary exists → **High** (warn if missing)

---

## 💾 Phase 2: Module B - Intelligent Storage Orchestration

### 2.1 Disk Discovery
- [x] Create `DiscoverDisks()` function
  - [x] Execute `lsblk -J` (JSON output)
  - [x] Parse JSON into Go structs
  - [x] Identify disk type (SSD vs HDD) via `lsblk -d -o NAME,ROTA`
  - [x] Get disk sizes and current mount points
  - [x] Execute `fdisk -l` for additional partition info

### 2.2 Disk Classification & Recommendation Engine
- [x] Create `ClassifyDisks()` function
  - [x] Categorize disks by type (SSD/HDD)
  - [x] Categorize by size (small/medium/large)
  - [x] Identify OS disk vs available disks

### 2.3 Single Disk Scenario (Case: 1 Disk Found)
- [x] Display **Single Point of Failure Warning**
- [x] Configure partitioning:
  - [x] OS partition
  - [x] Data partition on same drive
- [x] Show TUI confirmation prompt

### 2.4 Two Disk Scenario (Case: 2 Disks Found)
- [x] Create TUI selection menu for "The 5 Ranks":
  - [x] **Rank 1 (Hybrid):** SSD (OS/Apps) + HDD (Bulk Data) - *Recommended*
  - [x] **Rank 2 (Speed Demon):** SSD (OS) + SSD (Active DBs)
  - [x] **Rank 3 (Mirror):** 2x SSD (RAID 1)
  - [x] **Rank 4 (Data Hoarder):** 2x HDD (RAID 1) - *Trigger Performance Warning*
  - [x] **Rank 5 (Kamikaze):** RAID 0 - *Trigger Critical Risk Warning*

### 2.5 Multi-Disk Scenario (Case: 3+ Disks Found)
- [x] Propose optimal configuration:
  - [x] OS disk (SSD)
  - [x] Storage disk (HDD)
  - [x] Backup disk (HDD)
- [x] Allow user customization via TUI

### 2.6 Disk Formatting & Mounting
- [x] Create `FormatDisk()` function
  - [x] Create TUI filesystem format selection menu (ordered by recommendation):
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
  - [x] **Default Behavior:** Auto-select `ext4` if user presses Enter
  - [x] Execute appropriate `mkfs` command based on selection:
    - [x] `mkfs.ext4 -L <label>` for ext4
    - [x] `mkfs.xfs -L <label>` for XFS
    - [x] `mkfs.btrfs -L <label>` for Btrfs
    - [x] `zpool create` for ZFS (if installed)
  - [x] Confirm destructive action with user (⚠️ ALL DATA WILL BE ERASED)
  - [x] Implement "Fail Fast" - abort on any error
  - [x] Store selected filesystem in config for future reference

- [x] Create `MountDisk()` function
  - [x] Create mount point directory
  - [x] Add entry to `/etc/fstab`
  - [x] **Idempotency:** Check if entry already exists before adding
  - [x] Execute `mount -a`

### 2.7 Power Optimization
- [x] Create `ConfigureHDDSpindown()` function
  - [x] Identify mechanical drives (ROTA=1)
  - [x] Configure `hdparm` spindown rules
  - [x] Add to `/etc/hdparm.conf` for persistence

---

## 📁 Phase 3: Module C - Directory & Infrastructure Structure ✅

### 3.1 User Space Directory Creation
- [x] Create `~/infra/` (Root for configs)
- [x] Create `~/infra/scripts/` (Maintenance executables)
- [x] Create `~/infra/logs/` (Centralized logging)
- [x] Create `~/infra/compose/` (Docker Compose files)
- [x] Create `~/infra/config/` (Service configurations)
- [x] Create `~/infra/backups/` (Backup staging area)

### 3.2 Data Space Directory Creation
- [x] Create `/mnt/data/` root
- [x] Create Immich directories:
  - [x] `/mnt/data/gallery/library`
  - [x] `/mnt/data/gallery/upload`
  - [x] `/mnt/data/gallery/profile`
  - [x] `/mnt/data/gallery/video`
  - [x] `/mnt/data/gallery/thumbs`
- [x] Create Nextcloud directories:
  - [x] `/mnt/data/cloud/data`
  - [x] `/mnt/data/cloud/config`
- [x] Create Database directories (ISOLATED per service):
  - [x] `/mnt/data/databases/immich-postgres`
  - [x] `/mnt/data/databases/nextcloud-mariadb`
- [x] Create Cache directory:
  - [x] `/mnt/data/cache` (Redis/Valkey)

### 3.3 Permission Enforcement (Critical Fix from PRD)
- [x] Create `SetPermissions()` function
  - [x] Get current user PUID (typically 1000)
  - [x] Get current user PGID (typically 1000)
  - [x] Execute `chown -R ${PUID}:${PGID} /mnt/data`
  - [x] Set appropriate `chmod` (755 for dirs, 644 for files)
  - [x] Prevent "Permission Denied" boot loops

### 3.4 TUI Components
- [x] Create `RenderDirectoryPlan()` - Shows planned directories
- [x] Create `RenderDirectoryTree()` - Visual tree view
- [x] Create `RenderDirectoryProgress()` - Creation progress
- [x] Create `RenderDirectoryComplete()` - Summary with stats
- [x] Create `RenderPermissionConfirmation()` - Permission info

---

## 🐳 Phase 4: Module D - Service Composition ✅

### 4.1 Input Wizard (Interactive TUI)
- [x] Create `InputWizard()` TUI component
- [x] Collect user inputs:
  - [x] `TZ` (Timezone) - Default: Auto-detect or "UTC"
  - [x] `PUID/PGID` - Default: Current user (1000/1000)
  - [x] `DB_PASSWORD` - Auto-generate strong alphanumeric if blank
  - [x] Discord/Telegram Webhook URLs (for notifications)
  - [x] Nextcloud Admin Username
  - [x] Nextcloud Admin Password
  
  > **Note:** `UPLOAD_LOCATION` and data paths are **NOT user-configurable**.  
  > They are auto-derived from Module B (disk selection) and hardcoded to `/mnt/data/`  
  > for consistency. This is intentional — servctl is an **opinionated** tool.

- [x] Input Validation:
  - [x] Ensure passwords are not empty
  - [x] Ensure paths exist or are creatable
  - [x] Validate webhook URL format

### 4.2 Network Configuration
- [x] Create `DetectHostIP()` function
  - [x] Auto-detect primary interface IP
  - [x] Prompt user to set **Static IP** (Critical from PRD)
  - [x] Validate IP format
  - [x] Validate IP range (192.168.x.x or 10.x.x.x)
  - [x] Validate IP is not already in use
  - [x] Warn about DHCP lockout risk for Nextcloud

### 4.3 Docker Compose Generation
- [x] Create `docker-compose.yml.tmpl` template
- [x] Generate services:
  - [x] **Immich Server** (Port 2283)
    - [x] Volume mounts for uploads
    - [x] ENV file reference
    - [x] Depends on: redis, database
  - [x] **Immich Machine Learning**
    - [x] Model cache volume
    - [x] ENV file reference
  - [x] **Nextcloud** (Port 8080)
    - [x] Configure `OVERWRITEPROTOCOL`
    - [x] Configure `TRUSTED_DOMAINS` with static IP
    - [x] Configure `OVERWRITEHOST`
    - [x] Admin auto-setup via ENV
  - [x] **Redis/Valkey** (For Immich)
  - [x] **Glances** (Port 61208)
    - [x] Host mode networking
    - [x] SMART data reading capabilities
    - [x] `SYS_ADMIN` and `SYS_RAWIO` capabilities
  - [x] **Diun** (Update notifier)
    - [x] Discord/Telegram webhook integration
    - [x] Bi-monthly check schedule

### 4.4 Database Architecture (Critical Fix from PRD)
- [x] Create **ISOLATED** database containers:
  - [x] **Immich DB:** Dedicated `postgres:14-vectorchord` container
  - [x] **Nextcloud DB:** Dedicated `mariadb` container
  - [x] Separate credentials for each database
- [x] **Rationale:** Prevent "Dependency Hell" during upgrades

### 4.5 ENV File Generation
- [x] Create `.env.tmpl` template (built into GenerateEnvFile)
- [x] Generate sections:
  - [x] Shared/Global settings (TZ, PUID, PGID)
  - [x] Immich DB credentials
  - [x] Nextcloud DB credentials
  - [x] Path configurations
  - [x] Webhook URLs

### 4.6 Firewall Configuration (UFW)
- [x] Create `ConfigureFirewall()` function
- [x] **CRITICAL: Lockout Prevention**
  - [x] Execute `ufw allow ssh` (Port 22) **FIRST**
  - [x] Verify SSH rule is active
- [x] Add service rules:
  - [x] `ufw allow 2283/tcp` (Immich)
  - [x] `ufw allow 8080/tcp` (Nextcloud)
  - [x] `ufw allow 61208/tcp` (Glances - Optional/Local only)
- [x] Execute `ufw enable` only after rule verification
- [x] Display firewall status to user

### 4.7 TUI Components
- [x] `RenderConfigWizardIntro()` - Configuration wizard intro
- [x] `RenderInputPrompt()` - Input field rendering
- [x] `RenderTimezoneSelection()` - Timezone picker
- [x] `RenderNetworkConfig()` - Network configuration screen
- [x] `RenderGeneratedCredentials()` - Credential display
- [x] `RenderServiceSummary()` - Service overview
- [x] `RenderFirewallConfig()` - Firewall rules display
- [x] `RenderComposeGenerated()` - Compose file confirmation
- [x] `RenderConfigSummary()` - Final configuration summary

---

## 🔧 Phase 5: Module E - Maintenance & Reliability ✅

### 5.1 Script Generation
- [x] Create `disk_alert.sh` template
  - [x] Monitor disk usage threshold (default: 90%)
  - [x] Send notification if threshold exceeded
  - [x] Inject user's webhook URL

- [x] Create `smart_alert.sh` template
  - [x] Run `smartctl` health check on all drives
  - [x] Parse SMART attributes for warnings
  - [x] Send notification on failure indicators

- [x] Create `daily_backup.sh` template
  - [x] Backup with rsync (--delete for sync)
  - [x] Track before/after disk usage
  - [x] Send success/failure notification with storage stats
  - [x] Color-coded Discord embeds (green=success, red=fail)

- [x] Create `weekly_cleanup.sh` template
  - [x] Prune unused Docker images: `docker image prune -f`
  - [x] Clean apt cache: `apt-get clean && autoremove`
  - [x] Truncate large log files (>50MB)
  - [x] Clean old backups (configurable retention days)
  - [x] Send completion notification with before/after usage

### 5.2 Cron Orchestration
- [x] Create `GenerateCronFile()` function
- [x] Default schedules:
  - [x] Daily backup: 4:00 AM
  - [x] Weekly cleanup: Sunday 3:00 AM
  - [x] Disk alert: Every 6 hours
  - [x] SMART check: Daily 5:00 AM
- [x] **⚠️ SUDO/ROOT REQUIRED:** All cron jobs run with elevated privileges
  - [x] Use `/etc/cron.d/servctl` (system cron, runs as root)
  - [x] Proper SHELL and PATH settings
  - [x] Set permissions: chmod 644
- [x] `CronSchedule` struct with `String()` and `HumanReadable()` methods

### 5.3 Notification Integration
- [x] Create `TestWebhook()` function
  - [x] Send test notification to Discord
  - [x] Verify webhook is working before finalizing
- [x] Webhook URLs templated into all scripts

### 5.4 Log Rotation Configuration
- [x] Create `GenerateLogrotateConfig()` function
- [x] Configure rotation policy:
  - [x] Rotate weekly
  - [x] Keep 4 weeks of history
  - [x] Compress old logs (gzip)
  - [x] Target: `~/infra/logs/*.log`
- [x] Output file: `/etc/logrotate.d/servctl`

### 5.5 TUI Components
- [x] `RenderMaintenanceIntro()` - Setup introduction
- [x] `RenderScriptPreview()` - Script preview cards
- [x] `RenderAllScripts()` - All scripts summary
- [x] `RenderCronSchedule()` - Cron configuration display
- [x] `RenderWebhookTest()` - Webhook test result
- [x] `RenderLogrotateConfig()` - Logrotate info display
- [x] `RenderMaintenanceComplete()` - Completion summary
- [x] `RenderMaintenanceSummary()` - Full configuration summary

<!-- Future Feature  -->

<!-- ### 5.6 Offsite Backup (Optional - From PRD Feedback) 
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
| Phase 0: Project Setup | ✅ Complete | 100% |
| Phase 1: Pre-flight | ✅ Complete | 100% |
| Phase 2: Storage | ✅ Complete | 100% |
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
