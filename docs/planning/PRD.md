# Product Requirements Document: servctl

| Metadata | Details |
| --- | --- |
| **Product Name** | servctl (CLI Tool) |
| **Version** | 1.0.0-draft |
| **Target OS** | Ubuntu 22.04 LTS+ (Server) |
| **Primary User** | HomeLab/Self-Hosted Enthusiasts (Madhav Goswami) |
| **Objective** | Automate the end-to-end provisioning of a robust, self-healing home server (NAS, Media, Cloud) with an interactive, "opinionated" CLI. |

---

## 1. System Prerequisites

The agent must verify the following environment constraints before execution. If critical checks fail, execution aborts.

* **OS:** Ubuntu 22.04 LTS or later.
* **Privileges:** Non-root user with `sudo` access.
* **Hardware Config:**
* Virtualization (VT-x/AMD-V) enabled (Check via `lscpu`).
* Secure Boot disabled (Optional, warn if enabled).
* Serial Ports disabled (Optional, warn if enabled).


* **Connectivity:** Active internet connection (for package fetching).

---

## 2. Functional Requirements

### 2.1 Module A: "Pre-flight" & Dependency Management

**Goal:** Ensure the host is compliant with the software stack.

* **System Update:** Auto-execute `apt update && apt upgrade -y`.
* **Dependency Audit:** Check for existence of tools; install if missing:
* `curl`, `net-tools` (Network verification)
* `docker`, `docker-compose` (Container orchestration)
* `hdparm` (Power management)
* `smartmontools` (Disk health monitoring)
* `crontab` (Scheduling)



### 2.2 Module B: Intelligent Storage Orchestration

**Goal:** Analyze physical hardware and enforce an optimal disk layout based on "The 5 Ranks."

* **Discovery:** Run `lsblk` and `fdisk -l` to detect physical drives.
* **Logic Branching:**
* **Case: 1 Disk Found** → Warn user (Single Point of Failure). Configure partitioning (OS + Data on same drive).
* **Case: 2 Disks Found** → Prompt user to select configuration mode:
* *Rank 1 (Hybrid):* SSD (OS/Apps) + HDD (Bulk Data). **Recommended.**
* *Rank 2 (Speed Demon):* SSD (OS) + SSD (Active DBs).
* *Rank 3 (Mirror):* 2x SSD (RAID 1).
* *Rank 4 (Data Hoarder):* 2x HDD (RAID 1) — *Trigger Performance Warning.*
* *Rank 5 (Kamikaze):* RAID 0 — *Trigger Critical Risk Warning.*


* **Case: 3+ Disks Found** → Propose OS (SSD) + Storage (HDD) + Backup (HDD) setup.


* **Power Optimization:** Configure `hdparm` spindown rules for identified mechanical drives.

### 2.3 Module C: Directory & Infra Structure

**Goal:** Create a standardized, immutable filesystem structure.

* **User Space:**
* `~/infra/` (Root for configs)
* `~/infra/scripts/` (Maintenance executables)
* `~/infra/logs/` (Centralized logging)


* **Data Space:**
* `/mnt/data/gallery/{library, upload, profile, video}` (Immich)
* `/mnt/data/cloud/data` (Nextcloud)


* **Permission Enforcement:** Recursively set ownership (`chown`) and permissions (`chmod`) to prevent "Permission Denied" boot loops.

### 2.4 Module D: Service Composition

**Goal:** Dynamic generation of `docker-compose.yml` and `.env`.

* **Input:** Interactive prompts for user variables (Passwords, Timezone, Notification Webhooks).
* **Output:** Generate `docker-compose.yml` containing:
1. **Immich:** Server, Microservices, ML.
2. **Nextcloud:** App container + Network fixes (`OVERWRITEPROTOCOL`, `TRUSTED_DOMAINS`).
3. **Persistence Layer:** Shared Postgres instance (Vector enabled for Immich), Redis, Valkey.
4. **Observability:** Glances (Host mode, reading SMART data), Diun (Update notification).


* **Network Config:**
* Detect Host IP automatically.
* Configure UFW (Firewall): Allow `2283` (Immich) and `8080` (Nextcloud).



### 2.5 Module E: Maintenance & Reliability (The "Antigravity")

**Goal:** Self-healing and automated reporting.

* **Script Generation:** Create executable bash scripts in `~/infra/scripts/`:
* `disk_alert.sh`: Monitor usage threshold.
* `smart_alert.sh`: Monitor drive health.
* `daily_backup.sh`: Backup specific volumes/DB dumps.
* `weekly_cleanup.sh`: Prune docker images/logs.


* **Cron Orchestration:** Interactive scheduling (Defaulting to the user's detailed timetable: Backups @ 4:00 AM, Cleanup @ Sundays).
* **Notification Integration:** Inject user's webhook (Discord/Telegram) into all scripts for success/failure alerts.

---

## 3. CLI Command Interface (`servctl`)

| Command | Description |
| --- | --- |
| `servctl -start-setup` | Initiates the interactive installation wizard (Modules A-E). |
| `servctl -status` | Returns health status of Docker containers and disk usage. |
| `servctl -get-config` | Dumps the current `.env` and `docker-compose.yml` to stdout. |
| `servctl -get-architecture` | Visualizes the current folder structure and disk mapping. |
| `servctl -manual-backup` | Triggers immediate execution of `daily_backup.sh`. |

---

## 4. Future Roadmap (v1.1+)

* **Web Dashboard:** A lightweight GUI wrapping `servctl -status` and log viewing.
* **One-Click Restore:** `servctl -restore <backup_file>` to rebuild infra from a backup.
* **Remote Access:** Interactive setup of Cloudflare Tunnels or Tailscale.

---

## Critical Analysis & "Brutal Truth" Feedback

As requested, here is the challenge to your logic/assumptions:

**1. The "Shared Database" Risk**

* **Your Plan:** You are running Immich and Nextcloud on a *single* shared Postgres container (`postgres:14-vectorchord...`).
* **The Flaw:** Immich requires a highly specific Postgres version with `pgvecto.rs`. Nextcloud is generally version-agnostic but picky about configuration. If Immich updates and requires a DB upgrade (e.g., to PG15 or a newer vector extension), upgrading that container might break Nextcloud, or vice-versa.
* **Correction:** Use **isolated database containers** for each service. It costs minimal RAM but saves you from "dependency hell" later. Do not couple your photo storage to your file cloud.

**2. IP Binding Hardcoding**

* **Your Plan:** The prompt suggests `NEXTCLOUD_TRUSTED_DOMAINS=192.168.1.24`.
* **The Flaw:** If the user's router assigns a new DHCP IP, Nextcloud will lock the user out ("Untrusted Domain").
* **Correction:** The script must either (a) Force the user to set a Static IP during setup, or (b) Dynamically fetch the current IP at runtime, but (a) is the only robust server solution.

**3. Permissions on `/mnt/data**`

* **The Flaw:** You are mounting host directories (`/mnt/data/gallery`) into containers. Docker containers often run as `root` (0) or `www-data` (33), while your host user is `madhav` (1000).
* **The Risk:** Without explicit PUID/PGID handling in the compose file or `chown` commands in the script, the containers will either fail to write to the disk or write files that you (the host user) cannot delete/move later.
* **Correction:** Pass `PUID=1000` and `PGID=1000` into the containers (where supported) or have the script explicitly `chown -R 1000:1000 /mnt/data`.

**4. The Backup "Gap"**

* **The Flaw:** You mentioned `daily_backup.sh`.
* **The Risk:** A script that copies data from `Drive A` to `Drive B` inside the same machine is **not a backup**; it's redundancy. If the power supply surges and fries the motherboard and connected drives, you lose everything.
* **Correction:** The PRD should include an optional "Offsite Push" (e.g., Rclone to Google Drive/S3) for critical config/DB dumps.

**5. Firewall (UFW) Sequence**

* **The Flaw:** Enabling UFW on a remote SSH server is dangerous. If you run `ufw enable` before `ufw allow ssh`, you lock yourself out of the server instantly.
* **Correction:** The script must explicitly ensure port 22 (SSH) is allowed *before* enabling the firewall.

### 3.7 Security & Access Control (Firewall)

**Gap:** The current PRD opens ports but doesn't secure them.
**Requirement:** `servctl` must configure `ufw` (Uncomplicated Firewall) to secure the server while preventing lockout.

* **Lockout Prevention:** Explicitly allow SSH (Port 22) *before* enabling the firewall.
* **Service Rules:**
* Allow 2283/tcp (Immich).
* Allow 8080/tcp (Nextcloud).
* Allow 61208/tcp (Glances - Optional/Local only).


* **Action:** Execute `ufw enable` only after rule verification.

### 3.8 Input & Environment Management

**Gap:** The PRD mentions generating `.env` but not how it gets the data.
**Requirement:** An interactive "Wizard" module to capture and validate user inputs before execution.

* **User Inputs:**
* `TZ` (Timezone) - Default: Auto-detect or "UTC".
* `PUID/PGID` - Default: Current user (1000).
* `UPLOAD_LOCATION` - Default: `/mnt/data`.
* `DB_PASSWORD` - Auto-generate strong alphanumeric string if left blank.


* **Validation:** Ensure passwords are not empty; ensure paths exist or are creatable.

### 3.9 Log Rotation Policy

**Gap:** You created a `~/infra/logs/` directory, but without rotation, these logs will eventually consume 100% of the OS disk (Disk Pressure).
**Requirement:** Configure standard Linux `logrotate` for the generated logs.

* **Config:** Create `/etc/logrotate.d/servctl` to target `~/infra/logs/*.log`.
* **Policy:** Rotate weekly, keep 4 weeks of history, compress old logs.

### 3.10 Database Architecture (Critical Fix)

**Gap:** Section 3.4 implies a shared Postgres instance. This creates "Dependency Hell" (e.g., Immich needs Postgres 14 with Vectors, Nextcloud needs standard Postgres 15).
**Requirement:** Enforce **Container Isolation**.

* **Immich DB:** Dedicated `postgres` container (with `pgvecto.rs` extension).
* **Nextcloud DB:** Dedicated `mariadb` or `postgres` container.
* **Rationale:** Decoupling ensures that upgrading one app doesn't break the other. Resources (RAM) are cheap; stability is expensive.

### 3.11 Reverse Proxy & HTTPS (The "Make it Perfect" Module)

**Gap:** In your raw notes, you mentioned being "tired of non-https". The current PRD leaves services on raw IPs (`http://192.168.x.x:2283`).
**Requirement:** Optional setup for **Caddy** or **Nginx Proxy Manager**.

* **Function:** Map `immich.local` or `cloud.local` to the internal ports.
* **Zero-Config SSL:** If the user owns a domain, use Caddy for automatic Let's Encrypt certificates.
* **Local SSL:** Generate self-signed certs for local LAN access (removing the "Not Secure" browser warning).

### 4. Non-Functional Requirements

* **Idempotency:** The tool must be re-runnable. If `servctl` runs a second time, it should detect existing configs/mounts and skip them rather than overwriting or duplicating entries in `/etc/fstab`.
* **Error Handling:** "Fail Fast." If a critical step (e.g., Disk Formatting) fails, stop execution immediately and output the `stderr` log. Do not proceed to install Docker on a broken filesystem.

### 5. The "Handover" (UX)

**Gap:** After the script finishes, the user is left staring at a terminal prompt.
**Requirement:** `servctl` must output a "Mission Report" at completion.

* **Dashboard URLs:** `http://<HOST_IP>:2283`, `http://<HOST_IP>:8080`.
* **Credentials:** Display the generated Admin Usernames/Passwords for one-time capture.
* **Next Steps:** Instructions on how to log in and change defaults immediately.
