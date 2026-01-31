# 🔄 servctl Execution Flow

> **This document details the exact execution flow of `servctl -start-setup`.**  
> Edit this to modify the behavior of the setup wizard.

---

## 📍 Entry Point

**File:** `cmd/servctl/main.go`  
**Function:** `runSetupWizard(dryRun bool)`

```
servctl -start-setup
        │
        ▼
┌─────────────────┐
│  Parse Flags    │ 
│  -dry-run       │
│  -start-setup   │
└────────┬────────┘
         │
         ▼
   runSetupWizard()
```

---

## 📋 Phase 1: System Preparation

**Location:** Lines 226-285 in `main.go`

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 1: SYSTEM PREPARATION                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Step 1.1: Check Missing Dependencies                           │
│  ───────────────────────────────────────                        │
│  Function: preflight.GetMissingDependencies()                   │
│  Returns: []Dependency (list of missing packages)               │
│                                                                 │
│          ┌──── missing > 0? ────┐                               │
│          │                      │                               │
│          ▼ YES                  ▼ NO                            │
│    Install each dep       Continue to Step 1.2                  │
│    preflight.InstallDependency(dep)                             │
│                                                                 │
│  Step 1.2: Run Preflight Checks                                 │
│  ─────────────────────────────────                              │
│  Function: preflight.RunPreflightWithAutoFix(dryRun)            │
│  Returns: ([]CheckResult, []InstallResult, error)               │
│                                                                 │
│    Checks performed (in order):                                 │
│    ┌────────────────────────────────────────────────────────┐   │
│    │ 1. CheckOS()           → Ubuntu 22.04+?                │   │
│    │ 2. CheckPrivileges()   → Sudo access? Not root?        │   │
│    │ 3. CheckHardware()     → VT-x? Secure Boot?            │   │
│    │ 4. CheckConnectivity() → Ping, DNS, HTTPS              │   │
│    │ 5. CheckStaticIP()     → Static or DHCP?               │   │
│    │ 6. CheckAllDependencies() → All packages installed?    │   │
│    │ 7. CheckDockerRunning()   → Docker daemon active?      │   │
│    └────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Step 1.3: Display Results                                      │
│  ─────────────────────────────                                  │
│  Function: tui.RenderPreflightResults(results)                  │
│                                                                 │
│          ┌──── HasBlockers(results)? ────┐                      │
│          │                               │                      │
│          ▼ YES                           ▼ NO                   │
│    Display errors              Continue to Step 1.4             │
│    EXIT with error                                              │
│                                                                 │
│  Step 1.4: Interactive Static IP Setup (NEW!)                   │
│  ──────────────────────────────────────────────                 │
│  Function: preflight.PromptStaticIPSetup(reader, dryRun)        │
│                                                                 │
│          ┌──── DHCP detected? ────┐                             │
│          │                        │                             │
│          ▼ YES                    ▼ NO                          │
│    Show prompt:                   Skip (already static)         │
│    "Configure static IP? [y/N]"                                 │
│                                                                 │
│          ┌──── User says yes? ────┐                             │
│          │                        │                             │
│          ▼ YES                    ▼ NO                          │
│    Prompt for:                    Continue to Phase 2           │
│    • Gateway IP (auto-detected)                                 │
│    • DNS servers (defaults: 8.8.8.8, 1.1.1.1)                   │
│    Generate netplan config                                      │
│    Apply with: sudo netplan apply                               │
│                                                                 │
│  Step 1.5: User Confirmation                                    │
│  ──────────────────────────────                                 │
│  promptContinue("Continue to disk selection?")                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Functions Called

| Order | Function | File | Purpose |
|-------|----------|------|---------|
| 1 | `GetMissingDependencies()` | `preflight.go:1009` | List uninstalled deps |
| 2 | `InstallDependency(dep)` | `preflight.go:925` | Install single package |
| 3 | `RunPreflightWithAutoFix()` | `preflight.go:1096` | Run all checks + auto-fix |
| 4 | `CheckOS()` | `preflight.go:62` | Verify Ubuntu version |
| 5 | `CheckPrivileges()` | `preflight.go:115` | Check sudo/root |
| 6 | `CheckHardware()` | `preflight.go:167` | Check VT-x, Secure Boot |
| 7 | `CheckConnectivity()` | `preflight.go:314` | Test network |
| 8 | `CheckStaticIP()` | `preflight.go:413` | Check static/DHCP |
| 9 | `CheckAllDependencies()` | `preflight.go:749` | Verify all packages |
| 10 | `CheckDockerRunning()` | `preflight.go:770` | Docker daemon status |
| 11 | `RenderPreflightResults()` | `tui/preflight.go` | Display results |
| 12 | `HasBlockers()` | `preflight.go:982` | Check for failures |

### Required Dependencies

| Package | Binary | Criticality | Install Command |
|---------|--------|-------------|-----------------|
| curl | `curl` | Blocker | `apt install -y curl` |
| net-tools | `ifconfig` | Recommended | `apt install -y net-tools` |
| Docker | `docker` | Blocker | `curl -fsSL https://get.docker.com \| sh` |
| Docker Compose | `docker compose` | Blocker | `apt install -y docker-compose` |
| hdparm | `hdparm` | Recommended | `apt install -y hdparm` |
| smartmontools | `smartctl` | Recommended | `apt install -y smartmontools` |
| cron | `crontab` | High | `apt install -y cron` |
| UFW | `ufw` | High | `apt install -y ufw` |
| lsblk | `lsblk` | Blocker | `apt install -y util-linux` |
| mkfs.ext4 | `mkfs.ext4` | Blocker | `apt install -y e2fsprogs` |

---

## 💾 Phase 2: Storage Configuration

**Location:** Lines 292-310 in `main.go`

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 2: STORAGE CONFIGURATION              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Step 2.1: Discover Disks                                       │
│  ───────────────────────────                                    │
│  Function: storage.DiscoverDisks()                              │
│  Returns: ([]Disk, error)                                       │
│                                                                 │
│    Runs: lsblk -J -b -o NAME,SIZE,TYPE,MODEL,...               │
│    Parses JSON output into Disk structs                         │
│                                                                 │
│    Disk types detected:                                         │
│    • SSD (rotational=false)                                     │
│    • HDD (rotational=true)                                      │
│    • NVMe (name starts with "nvme")                             │
│    • USB (transport=usb OR removable=true)                      │
│    • Loop (type=loop, for testing)                              │
│                                                                 │
│  Step 2.2: Display Discovered Disks                             │
│  ────────────────────────────────────                           │
│  Function: tui.RenderDiskDiscovery(disks)                       │
│                                                                 │
│          ┌──── len(disks) == 0? ────┐                           │
│          │                          │                           │
│          ▼ YES                      ▼ NO                        │
│    "No suitable disks"        Display disk cards                │
│    Skip to Phase 3            Allow selection (future)          │
│                                                                 │
│  Step 2.3: User Confirmation                                    │
│  ──────────────────────────────                                 │
│  promptContinue("Continue to directory setup?")                 │
│                                                                 │
│  NOTE: Disk formatting/mounting is NOT yet interactive.         │
│  Currently just displays discovered disks.                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Functions Called

| Order | Function | File | Purpose |
|-------|----------|------|---------|
| 1 | `DiscoverDisks()` | `storage/discovery.go:158` | List all block devices |
| 2 | `parseLsblkDevice()` | `storage/discovery.go:204` | Parse JSON to struct |
| 3 | `classifyDiskType()` | `storage/discovery.go:260` | SSD/HDD/NVMe/USB |
| 4 | `RenderDiskDiscovery()` | `tui/storage.go` | Display disk cards |

### Disk Struct Fields

```go
type Disk struct {
    Name         string      // e.g., "sda", "nvme0n1"
    Path         string      // e.g., "/dev/sda"
    Size         uint64      // Size in bytes
    SizeHuman    string      // "500.00 GB"
    Model        string      // Disk model name
    Type         DiskType    // SSD, HDD, NVMe, USB
    Rotational   bool        // true = HDD
    Removable    bool        // true = USB/removable
    Partitions   []Partition // Existing partitions
    IsOSDisk     bool        // Contains root filesystem
    IsAvailable  bool        // Available for use
}
```

---

## 📁 Phase 3: Directory Structure

**Location:** Lines 312-336 in `main.go`

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 3: DIRECTORY STRUCTURE                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Step 3.1: Get Directory Lists                                  │
│  ───────────────────────────────                                │
│  userDirs = directory.GetUserSpaceDirectories(homeDir)          │
│  dataDirs = directory.GetDataSpaceDirectories("/mnt/data")      │
│                                                                 │
│  Step 3.2: Display Plan                                         │
│  ─────────────────────────                                      │
│  Function: tui.RenderDirectoryPlan(allDirs)                     │
│                                                                 │
│  Step 3.3: Create Directories                                   │
│  ───────────────────────────────                                │
│                                                                 │
│          ┌──── dryRun? ────┐                                    │
│          │                 │                                    │
│          ▼ YES             ▼ NO                                 │
│    Print "[DRY RUN]"  CreateAllDirectories()                    │
│                       SetPermissions()                          │
│                                                                 │
│  Step 3.4: Display Results                                      │
│  ─────────────────────────────                                  │
│  Function: tui.RenderDirectoryComplete(results)                 │
│                                                                 │
│  Step 3.5: User Confirmation                                    │
│  ──────────────────────────────                                 │
│  promptContinue("Continue to service configuration?")           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Directories Created

**User Space: `~/infra/`**
```
~/infra/
├── scripts/     # Maintenance executables
├── logs/        # Centralized logging
├── compose/     # Docker Compose files
├── config/      # Service configurations
└── backups/     # Backup staging area
```

**Data Space: `/mnt/data/`**
```
/mnt/data/
├── gallery/              # Immich
│   ├── library/          # Photo storage
│   ├── upload/           # Upload staging
│   ├── profile/          # User profiles
│   ├── video/            # Video transcodes
│   └── thumbs/           # Thumbnails
├── cloud/                # Nextcloud
│   ├── data/             # User files
│   └── config/           # Config
├── databases/            # Databases
│   ├── immich-postgres/  # Immich DB
│   └── nextcloud-mariadb/# Nextcloud DB
└── cache/                # Redis
```

### Functions Called

| Order | Function | File | Purpose |
|-------|----------|------|---------|
| 1 | `GetUserSpaceDirectories()` | `directory/structure.go` | Get ~/infra/* dirs |
| 2 | `GetDataSpaceDirectories()` | `directory/structure.go` | Get /mnt/data/* dirs |
| 3 | `RenderDirectoryPlan()` | `tui/directory.go` | Show planned dirs |
| 4 | `CreateAllDirectories()` | `directory/structure.go` | Create dirs |
| 5 | `SetPermissions()` | `directory/structure.go` | chown/chmod |
| 6 | `RenderDirectoryComplete()` | `tui/directory.go` | Show results |

---

## 🐳 Phase 4: Service Configuration

**Location:** Lines 339-398 in `main.go`

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 4: SERVICE CONFIGURATION              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Step 4.1: Create Default Config                                │
│  ─────────────────────────────────                              │
│  config = compose.DefaultConfig()                               │
│  config.AutoFillDefaults()  ← Generates passwords               │
│                                                                 │
│  Step 4.2: Detect Host IP                                       │
│  ───────────────────────────                                    │
│  ip, err := compose.DetectHostIP()                              │
│  config.HostIP = ip                                             │
│                                                                 │
│  Step 4.3: Display Configuration                                │
│  ─────────────────────────────────                              │
│  Function: tui.RenderConfigSummary(config)                      │
│                                                                 │
│  Step 4.4: Generate Files                                       │
│  ───────────────────────────                                    │
│                                                                 │
│          ┌──── dryRun? ────┐                                    │
│          │                 │                                    │
│          ▼ YES             ▼ NO                                 │
│    Print "[DRY RUN]"  GenerateDockerCompose()                   │
│                       GenerateEnvFile()                         │
│                                                                 │
│  Step 4.5: Display Generated Files                              │
│  ───────────────────────────────────                            │
│  Function: tui.RenderComposeGenerated(config)                   │
│                                                                 │
│  Step 4.6: User Confirmation                                    │
│  ──────────────────────────────                                 │
│  promptContinue("Continue to maintenance setup?")               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files Generated

**`~/infra/compose/docker-compose.yml`**
```yaml
services:
  immich-server:        # Port 2283
  immich-machine-learning:
  immich-postgres:      # PostgreSQL 14
  redis:
  nextcloud:           # Port 8080
  nextcloud-mariadb:   # MariaDB
  glances:             # Port 61208
  diun:                # Update notifier

networks:
  servctl-network:
    driver: bridge
```

**`~/infra/compose/.env`** (mode 0600)
```env
# System
TZ=UTC
PUID=1000
PGID=1000

# Immich
IMMICH_DB_PASSWORD=<generated>
UPLOAD_LOCATION=/mnt/data/gallery

# Nextcloud
NEXTCLOUD_ADMIN_USER=admin
NEXTCLOUD_ADMIN_PASSWORD=<generated>
MARIADB_PASSWORD=<generated>

# Notifications (optional)
DISCORD_WEBHOOK=
TELEGRAM_BOT_TOKEN=
```

### Functions Called

| Order | Function | File | Purpose |
|-------|----------|------|---------|
| 1 | `DefaultConfig()` | `compose/config.go` | Create config struct |
| 2 | `AutoFillDefaults()` | `compose/config.go` | Generate passwords |
| 3 | `DetectHostIP()` | `compose/network.go` | Get server IP |
| 4 | `RenderConfigSummary()` | `tui/compose.go` | Display config |
| 5 | `GenerateDockerCompose()` | `compose/generator.go` | Create YAML |
| 6 | `GenerateEnvFile()` | `compose/generator.go` | Create .env |
| 7 | `RenderComposeGenerated()` | `tui/compose.go` | Show file paths |

---

## 🔧 Phase 5: Maintenance Scripts

**Location:** Lines 400-440 in `main.go`

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 5: MAINTENANCE SCRIPTS                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Step 5.1: Display Script Previews                              │
│  ───────────────────────────────────                            │
│  Function: tui.RenderAllScripts(scripts)                        │
│                                                                 │
│  Step 5.2: Generate Scripts                                     │
│  ───────────────────────────                                    │
│                                                                 │
│  Scripts generated:                                             │
│  ┌────────────────────────────────────────────────────────┐     │
│  │ daily_backup.sh    → Daily 4:00 AM                     │     │
│  │ disk_alert.sh      → Every 6 hours                     │     │
│  │ smart_alert.sh     → Daily 5:00 AM                     │     │
│  │ weekly_cleanup.sh  → Sunday 3:00 AM                    │     │
│  └────────────────────────────────────────────────────────┘     │
│                                                                 │
│          ┌──── dryRun? ────┐                                    │
│          │                 │                                    │
│          ▼ YES             ▼ NO                                 │
│    Print "[DRY RUN]"  GenerateAllScripts()                      │
│                       InstallCronJobs()                         │
│                                                                 │
│  Step 5.3: Display Mission Report                               │
│  ───────────────────────────────────                            │
│  Function: report.RenderMissionReport()                         │
│                                                                 │
│    Includes:                                                    │
│    • Dashboard URLs (Immich, Nextcloud, Glances)                │
│    • Generated credentials (ONE-TIME DISPLAY)                   │
│    • Quick start commands                                       │
│    • Next steps guide                                           │
│                                                                 │
│  ─────────────────────── SETUP COMPLETE ───────────────────────│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Scripts Generated

| Script | Schedule | Purpose |
|--------|----------|---------|
| `daily_backup.sh` | `0 4 * * *` | rsync data + DB dumps |
| `disk_alert.sh` | `0 */6 * * *` | Alert if disk > 90% |
| `smart_alert.sh` | `0 5 * * *` | SMART health check |
| `weekly_cleanup.sh` | `0 3 * * 0` | Prune Docker, clean logs |

### Functions Called

| Order | Function | File | Purpose |
|-------|----------|------|---------|
| 1 | `RenderAllScripts()` | `tui/maintenance.go` | Preview scripts |
| 2 | `GenerateAllScripts()` | `maintenance/generator.go` | Create .sh files |
| 3 | `InstallCronJobs()` | `maintenance/cron.go` | Add to crontab |
| 4 | `RenderMissionReport()` | `report/mission.go` | Final summary |
| 5 | `RenderCredentials()` | `report/mission.go` | Show passwords |
| 6 | `RenderQuickStart()` | `report/mission.go` | docker compose cmds |
| 7 | `RenderNextSteps()` | `report/mission.go` | Post-setup guide |

---

## 🔀 Decision Points Summary

| Phase | Decision | YES Path | NO Path |
|-------|----------|----------|---------|
| 1 | Missing deps? | Install each | Continue |
| 1 | Has blockers? | EXIT with error | Continue |
| 1 | User confirms? | Continue | EXIT cancelled |
| 2 | Disks found? | Display disks | Skip phase |
| 2 | User confirms? | Continue | EXIT cancelled |
| 3 | Dry run? | Print plan only | Create dirs |
| 3 | User confirms? | Continue | EXIT cancelled |
| 4 | Dry run? | Print summary | Generate files |
| 4 | User confirms? | Continue | EXIT cancelled |
| 5 | Dry run? | Print scripts | Generate all |

---

## 📁 Key Files Reference

| File | Location | Purpose |
|------|----------|---------|
| `main.go` | `cmd/servctl/` | CLI entry, setup wizard |
| `preflight.go` | `internal/preflight/` | System checks |
| `discovery.go` | `internal/storage/` | Disk detection |
| `structure.go` | `internal/directory/` | Dir creation |
| `generator.go` | `internal/compose/` | Docker Compose gen |
| `cron.go` | `internal/maintenance/` | Cron job setup |
| `mission.go` | `internal/report/` | Final report |

---

## 🎨 TUI Rendering Functions

All UI is rendered via the `internal/tui/` package using Lip Gloss:

| Function | Purpose |
|----------|---------|
| `RenderPreflightResults()` | Preflight check results |
| `RenderDiskDiscovery()` | Disk cards |
| `RenderDirectoryPlan()` | Directory tree |
| `RenderDirectoryComplete()` | Creation summary |
| `RenderConfigSummary()` | Config table |
| `RenderComposeGenerated()` | File generation info |
| `RenderAllScripts()` | Script preview cards |
| `RenderMissionReport()` | Final dashboard |

---

> **Last Updated:** 2026-02-01  
> **Author:** servctl development team
