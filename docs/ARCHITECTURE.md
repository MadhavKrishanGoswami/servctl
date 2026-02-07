# Architecture Overview

This document describes the high-level architecture of servctl.

---

## System Design

```
┌──────────────────────────────────────────────────────────────────────┐
│                              servctl                                  │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────┐                                                     │
│  │  cmd/main   │ ◄── CLI Entry Point                                 │
│  │  (753 LOC)  │     Flag parsing, command routing                   │
│  └──────┬──────┘                                                     │
│         │                                                            │
│         ▼                                                            │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Internal Packages                          │   │
│  ├──────────────────────────────────────────────────────────────┤   │
│  │                                                               │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐ │   │
│  │  │ preflight │  │  storage  │  │ directory │  │  compose  │ │   │
│  │  │   Phase 1 │  │   Phase 2 │  │   Phase 3 │  │   Phase 4 │ │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────┘ │   │
│  │                                                               │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐ │   │
│  │  │maintenance│  │   report  │  │    tui    │  │   utils   │ │   │
│  │  │   Phase 5 │  │  Summary  │  │  Renderer │  │  Helpers  │ │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────┘ │   │
│  │                                                               │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Package Descriptions

### cmd/servctl

The main entry point. Handles:
- Flag parsing with Go's `flag` package
- Command routing based on flags
- Setup wizard orchestration (5 phases)
- Terminal output styling via Lipgloss

### internal/preflight

System requirement validation:
- Docker installation check
- Sudo/root access verification
- Network configuration detection
- Dependency auto-installation

### internal/storage

Disk management core:

```
storage/
├── discovery.go      # Disk detection via lsblk
├── classification.go # HDD/SSD/NVMe detection
├── recommendation.go # Strategy generation algorithm
├── operations.go     # RAID setup, MergerFS pooling
├── format.go         # Filesystem formatting (ext4, xfs, btrfs, zfs)
└── power.go          # HDD spindown via hdparm
```

**Strategy Generation Algorithm:**
1. Discover all block devices
2. Filter out OS disk and removable media
3. Classify by type (HDD, SSD, NVMe)
4. Group by size similarity (within 20%)
5. Generate applicable strategies based on disk count and types

### internal/directory

Directory structure management:
- Service-specific directory creation
- Ownership and permission setting
- Path sanitization using `filepath.Join`

### internal/compose

Docker Compose generation:
- Template-based YAML generation
- .env file creation with secure passwords
- Service configuration (Nextcloud, Immich, etc.)

### internal/maintenance

Maintenance script generation:
- Bash script templates
- Cron job configuration
- Discord webhook integration

### internal/tui

Terminal UI rendering using Lipgloss:
- Styled output components
- Progress indicators
- Result tables and trees

### internal/utils

Shared utilities:
- File logging with rotation
- Path helpers
- Common functions

---

## Data Flow

```
User Input
    │
    ▼
┌─────────────────┐
│  cmd/main.go    │ ◄── Parse flags, route commands
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
[Wizard]  [Single Commands]
    │         │
    ▼         ▼
┌─────────────────────────────────────────────────────────┐
│                    Phase Execution                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Phase 1: preflight.RunPreflightWithAutoFix()           │
│     └─▶ Check system, install deps, configure network   │
│                                                          │
│  Phase 2: storage.DiscoverDisks()                       │
│     └─▶ storage.GenerateStrategies()                    │
│     └─▶ storage.ApplyStrategy()                         │
│                                                          │
│  Phase 3: directory.GetDirectoriesForServices()         │
│     └─▶ directory.CreateDirectory()                     │
│                                                          │
│  Phase 4: compose.WriteAllConfigFiles()                 │
│                                                          │
│  Phase 5: maintenance.GetScriptsForSelection()          │
│     └─▶ maintenance.WriteScript()                       │
│                                                          │
└─────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  report.Render  │ ◄── Mission summary
└─────────────────┘
```

---

## Key Design Decisions

### 1. Dry-Run First

All destructive operations support `dryRun bool` parameter:
```go
func FormatDisk(path string, fs FilesystemType, label string, dryRun bool) (*FormatResult, error)
```

### 2. Interactive by Default

Uses `bufio.Reader` for user prompts:
```go
reader := bufio.NewReader(os.Stdin)
selection := storage.PromptStrategySelection(reader, strategies)
```

### 3. Graceful Degradation

Works on both Linux and macOS:
- Linux: Full functionality with actual disk operations
- macOS: Simulated disk discovery for development

### 4. Template-Based Generation

Uses Go templates for composable output:
```go
tmpl := template.Must(template.ParseFiles("templates/docker-compose.yml.tmpl"))
```

### 5. Lipgloss Styling

Consistent terminal styling via Charm's Lipgloss:
```go
var successStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#10B981")).
    Bold(true)
```

---

## Testing Strategy

| Test Type | Location | Purpose |
|-----------|----------|---------|
| Unit | `*_test.go` | Function-level testing |
| Error Path | `error_test.go` | Edge cases, invalid inputs |
| Fuzz | `fuzz_test.go` | Random input discovery |
| Benchmark | `benchmark_test.go` | Performance measurement |
| Integration | `integration_test.go` | Linux-only system tests |

---

## Security Considerations

1. **Password Generation**: Cryptographically secure random passwords
2. **sudo Usage**: Minimal, only for disk operations
3. **No Remote Connections**: All operations are local
4. **Masked Output**: Passwords hidden in `-get-config`

---

## Future Considerations

- [ ] Plugin architecture for additional services
- [ ] Configuration file instead of flags
- [ ] Web-based setup wizard
- [ ] Backup to cloud storage
