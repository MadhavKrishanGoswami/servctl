# Storage Recommendation Engine Design

**Objective**: To intelligently look at a server's detected hardware, calculate all viable storage layouts, rank them by utility, and present them as clear, actionable choices to the user.

---

## 1. Input Parameters (The "Senses")

The recommendation engine relies solely on hardware detection. It does **not** interrogate the user.

### 1.1 Hardware Detection (Automated)
Captured via `lsblk`, `smartctl`, and system stats.

| Parameter | Description | Impact on Recommendation |
|-----------|-------------|--------------------------|
| **Disk Count** | Number of *Available* physical disks (excluding OS disk) | Determines viability of pools vs single drives vs partitioning. |
| **Disk Types** | SSD, HDD, NVMe, USB | **Speed Tiering**. SSDs = OS/Apps/Cache. HDDs = Bulk Data. |
| **Disk Sizes** | Capacity in Bytes | **Symmetry Check**. Determines Mirror vs Mismatched strategies. |
| **Model Name** | "Virtual", "PERC", "RAID" | **HW RAID Check**. Disables ZFS/Soft-RAID if hardware controller detected. |
| **Partitions** | Unmounted partitions on boot drive | **Fallback**. Used if no empty physical disks are found. |

---

## 2. Decision Logic (The "Brain")

The engine generates **Strategy Options** based on the composition of *Available* disks.

### 2.1 Strategy Profiles

#### 💿 Strategy 0: Single Drive Partitioning (The "Minimalist")
*   **Condition**: Total system has only **1 physical drive**.
*   **Implementation**: Create separate partitions on the single drive.
    *   **P1**: EFI (Boot)
    *   **P2**: Root `/` (50GB for OS + Apps)
    *   **P3**: Data `/mnt/data` (Remainder of space)
*   **Edge Case "Zero Data Disk"**: If user already partitioned manually, detect the unused partition and offer to format/mount it as `/mnt/data` instead of re-partitioning.

#### 🧩 Strategy 1: MergerFS Pool (The "Safe" Combiner)
*   **Condition**: 2+ HDDs of **any** size.
*   **Implementation**: Disks formatted individually (ext4/xfs), combined via [MergerFS](https://github.com/trapexit/mergerfs) into `/mnt/data`. 
*   **Edge Case "The Kitchen Sink"**: If disks are mixed speeds (e.g., 1x NVMe + 1x HDD), **DO NOT MERGE**. Create separate pools: `/mnt/fast` (NVMe) and `/mnt/storage` (HDD).

#### 🛡️ Strategy 2: ZFS/RAID Mirror (The "Uptime" King)
*   **Condition**: 2+ HDDs of **similar** size (within ~10%).
*   **Implementation**: ZFS Mirror (if RAM > 4GB) or MDADM RAID1 (ext4).
*   **Edge Case "Hardware RAID"**: If disk model contains "Virtual", "RAID", or "PERC", force **Standard Formatting (ext4/xfs)**. Do NOT suggest ZFS/MDADM (double RAID is bad).
*   **Pros**: Redundancy. Drive failure = No downtime.

#### 📸 Strategy 3: Primary + Snapshot (The "Ransomware" Killer)
*   **Condition**: 2 HDDs where **Backup Drive >= Primary Drive**.
*   **Implementation**: Drive A → `/mnt/data` (Live). Drive B → `/mnt/backup` (Cold). Daily Rsync/Borg script.
*   **Pros**: Protects against hardware failure **AND** user error (deletion).

#### 🗑️ Strategy 4: Scratch + Vault (The "Mismatched" Fix)
*   **Condition**: 1 Small HDD + 1 Large HDD (e.g., 500GB + 4TB).
*   **Implementation**: Large Drive → `/mnt/data` (The Vault). Small Drive → `/mnt/scratch` (Torrents/Temp).
*   **Pros**: Keeps wear-and-tear off the main storage.

---

## 3. Recommendation Output Format (TUI)

The user sees a simplified menu of generated options based on the scoring.

### Example A: Hardware RAID Detected
*   **Analysis**: Disks named "DELL PERC H730".

```text
💾 Storage Configuration
────────────────────────
⚠️ Hardware RAID Controller Detected ("PERC H730")
   Skipping Software RAID/ZFS options to prevent conflicts.

[1] ⭐ Standard Mount (Recommended)
    • Structure:  Mount RAID Vol at /mnt/data (ext4)
    • Capacity:   10TB Usable
    • Protection: Managed by Hardware Controller
```

### Example B: Mixed Speed (NVMe + HDD)
*   **Analysis**: 1x 1TB NVMe, 1x 8TB HDD.

```text
💾 Storage Configuration
────────────────────────
[1] ⭐ Speed Tiering (Recommended)
    • Structure:  /mnt/fast (NVMe) + /mnt/storage (HDD)
    • Capacity:   1TB Fast + 8TB Bulk
    • Best For:   Docker/DBs on Fast, Movies on Slow

[2] Backup Setup
    • Structure:  NVMe (Data) -> HDD (Daily Backup)
    • Warning:    Wastes 7TB of HDD space!
```

---

## 4. Safety Constraints & Validation

Before applying any layout:

1.  **Boot Protection**: Never include the Boot/Root disk.
2.  **Destructive Confirmation**: Force user to type `ERASE` if partitions are detected.
3.  **HW RAID Safety**: Explicitly block ZFS/Btrfs on Hardware RAID volumes.
