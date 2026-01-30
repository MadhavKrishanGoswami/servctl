## 1. Core Technology

| Component | Technology | Version | Reasoning |
| --- | --- | --- | --- |
| **Language** | **Go (Golang)** | `1.21+` | Single binary distribution, strict typing, excellent system/IO performance. |
| **UI Framework** | **Bubble Tea** | `v0.25+` | The Elm Architecture (TEA) for TUI. Robust, event-driven, and "pretty" terminal UIs. |
| **Styling** | **Lip Gloss** | `v0.9+` | Style definitions for Bubble Tea (colors, margins, layouts) to make it look professional. |
| **Target OS** | **Ubuntu Server** | `22.04 LTS` | The industry standard for stability. `servctl` is tightly coupled to `apt` and `systemd`. |

## 2. Go Modules & Libraries

We will minimize external dependencies to keep the binary small and secure.

### **User Interface (TUI)**

* `github.com/charmbracelet/bubbletea`: The core TUI loop.
* `github.com/charmbracelet/lipgloss`: CSS-like styling for terminal output.
* `github.com/charmbracelet/bubbles`: Pre-made components (spinners for long tasks, text inputs for prompts, lists for disk selection).

### **System Interaction**

* `os/exec`: **Standard Library.** Used to execute shell commands (`apt`, `docker`, `lsblk`).
* `syscall`: **Standard Library.** Used for verifying sudo privileges and handling signals.
* `net`: **Standard Library.** Used for pre-flight connectivity checks and port conflict detection.

### **Data Parsing & Logic**

* **JSON Parsing:** Native `encoding/json` to parse the output of `lsblk -json`.
* **YAML Generation:** `gopkg.in/yaml.v3` to reliably generate and validate `docker-compose.yml`.
* **Templating:** Native `text/template` to inject user variables (passwords, paths) into bash scripts and `.env` files.

## 3. System Architecture

### **The "Executor" Pattern**

Since `servctl` is an infrastructure tool, it follows a strict Command/Execute pattern to ensure safety.

1. **Discovery Phase (Read-Only):**
* Runs `lsblk`, `free -m`, `lscpu`.
* Parses JSON output into internal Go Structs.
* *Risk:* Zero. No state change.


2. **Decision Phase (Interactive):**
* Bubble Tea Model displays options based on Discovery.
* User inputs config (Timezone, Passwords).
* *Risk:* Zero. In-memory only.


3. **Execution Phase (Root Required):**
* The "Dirty" phase.
* Writes files to disk.
* Formats drives.
* Restarts services.
* *Risk:* High. Protected by confirmation prompts and "Fail Fast" error handling.



## 4. External System Dependencies

`servctl` does not reinvent the wheel. It acts as an orchestrator for these standard Linux binaries. The tool *must* check for these presence at startup.

| Binary | Purpose | Criticality |
| --- | --- | --- |
| `docker` | Container Runtime | **Blocker** |
| `lsblk` | Disk Discovery | **Blocker** |
| `mkfs.ext4` | Disk Formatting | **Blocker** |
| `hdparm` | HDD Power Management | Recommended |
| `smartctl` | HDD Health Check | Recommended |
| `ufw` | Firewall Management | High |

## 5. File Generation Strategy

### **Embedded Templates**

To keep the tool as a **single binary**, all configuration files are stored as Go string constants or embedded files (`//go:embed`).

* `templates/docker-compose.yml.tmpl` -> Populated with `text/template`.
* `templates/scripts/daily_backup.sh.tmpl` -> Populated with Webhook URLs.

### **Directory Structure (On Host)**

The tool enforces this structure (hardcoded for consistency):

```text
/home/{user}/infra/
├── docker-compose.yml  (Generated)
├── .env                (Generated with secrets)
├── logs/               (Mounted to containers)
└── scripts/            (Executable bash scripts)

```

## 6. Build & Distribution

### **Build Command**

```bash
# Builds a static binary.
# CGO_ENABLED=0 ensures no dependency on host libc versions.
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o servctl main.go

```

### **Installation (User-Side)**

The user only needs to run:

```bash
wget https://github.com/madhav/servctl/releases/latest/download/servctl
chmod +x servctl
sudo ./servctl

```
