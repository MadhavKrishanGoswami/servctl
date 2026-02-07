# Storage Recommendation Engine Design

**Objective**: Intelligently analyze server hardware, calculate all viable storage layouts, rank them, and present clear choices to the user.

---

## 1. Input Parameters (Hardware Detection)

Captured via `lsblk`, `smartctl`, and system stats. **No user questionnaire.**

| Parameter | Description | Impact |
|-----------|-------------|--------|
| **Disk Count** | Available disks (excluding OS) | Pool vs single drive vs partitioning |
| **Disk Types** | SSD, HDD, NVMe, USB | Speed tiering decisions |
| **Disk Sizes** | Capacity in Bytes | Mirror vs mismatched strategies |
| **Model Names** | Vendor/Model string | Hardware RAID detection |
| **System RAM** | Total Memory | ZFS eligibility (≥8GB) |

---

## 2. Core Strategy Profiles

### 💿 Strategy 0: Single Drive Partitioning
*   **Condition**: Only 1 physical drive total.
*   **Implementation**: Separate partitions (Root 50GB + Data remainder).
*   **Pros**: OS reinstall without data loss.
*   **Warning**: ⚠️ No redundancy. Drive failure = total loss.

### 🧩 Strategy 1: MergerFS Pool
*   **Condition**: 2+ HDDs of any size (same speed class).
*   **Implementation**: Individual ext4 disks combined via MergerFS.
*   **Pros**: 100% capacity, partial failure survivability.
*   **Best For**: Media collections, mismatched drive sizes.

### 🛡️ Strategy 2: ZFS/RAID Mirror
*   **Condition**: 2+ disks of similar size (within ~10%).
*   **Implementation**: ZFS Mirror (if RAM ≥8GB) or MDADM RAID1 (if RAM <8GB).
*   **Pros**: Drive failure protection, read speed boost.
*   **Best For**: Critical data, Nextcloud, databases.

### 📸 Strategy 3: Primary + Snapshot
*   **Condition**: 2 drives where Backup >= Primary.
*   **Implementation**: Drive A = `/mnt/data`, Drive B = `/mnt/backup` with daily rsync.
*   **Pros**: Protects against hardware failure + user error + ransomware.
*   **Best For**: Home users wanting "set and forget" safety.

### 🗑️ Strategy 4: Scratch + Vault
*   **Condition**: 1 small + 1 large drive (size diff > 50%).
*   **Implementation**: Large = `/mnt/data`, Small = `/mnt/scratch`.
*   **Pros**: Uses old drives effectively, protects main storage.
*   **Best For**: Users with legacy drives.

---

## 3. Edge Case Handling

### 🍳 Edge Case A: Mixed-Speed Data Drives
**Scenario**: NVMe + HDD available.  
**Solution**: Create speed-tiered pools:
- `/mnt/fast` → NVMe (databases, active projects)
- `/mnt/data` → HDD (media, archives)

### 🖥️ Edge Case B: Hardware RAID Detected
**Detection**: Model contains "Virtual Disk", "PERC", "MegaRAID", "SmartArray".  
**Solution**: Force ext4/XFS only. Block software RAID. Display warning.

### 📦 Edge Case C: Pre-Partitioned Drive
**Scenario**: OS drive has unmounted partitions.  
**Solution**: Offer to mount existing partitions instead of "no disks found."

---

## 4. Implementation Defaults (Design Decisions)

| Decision | Choice | Reasoning |
|----------|--------|-----------|
| **Default Filesystem** | **ext4** | Most reliable, universal. XFS only for >4TB media drives. |
| **MergerFS Policy** | **epmfs** | "Existing path, most free space" - keeps related files together. |
| **Backup Tool** | **rsync** | Universal, no dependencies. Borg is overkill for v1.0. |
| **ZFS vs MDADM** | **ZFS if RAM ≥8GB**, else MDADM | ZFS needs RAM for ARC cache. MDADM is lighter. |
| **USB Drives** | **Excluded** | Too risky for permanent storage. User can mount manually. |
| **Encryption** | **Skipped for v1.0** | Adds complexity. Can add LUKS support later. |
| **Mount Points** | **Fixed names** | Simplifies scripts and Docker volumes. |

### Default Mount Point Names
| Purpose | Default Path | Configurable? |
|---------|--------------|---------------|
| Primary data | `/mnt/data` | Yes |
| Backup target | `/mnt/backup` | Yes |
| Scratch/temp | `/mnt/scratch` | Yes |
| Fast tier | `/mnt/fast` | Yes |

---

## 5. Configurability Philosophy

**Principle**: Smart defaults that just work, with optional depth for power users.

### How It Works in the TUI

After selecting a strategy, show:
```
✓ Using recommended settings. Press [Enter] to continue.
  Or press [C] to customize options.
```

### Configurable Options (Only if user presses C)

| Option | Default | Alternatives | When to Show |
|--------|---------|--------------|--------------|
| **Backup Frequency** | Daily 3 AM | Every 6h, Every 12h, Weekly | Strategy 3 |
| **Mount Point** | `/mnt/data` | User input | All strategies |
| **Filesystem** | ext4 | XFS | All strategies |
| **MergerFS Policy** | epmfs | mfs, lfs | Strategy 1 |
| **Label Name** | `servctl_data` | User input | All strategies |

### Example Flow (Strategy 3 Selected)

```
📸 Primary + Backup Configuration
─────────────────────────────────
Primary Drive: /dev/sdb (4TB) → /mnt/data
Backup Drive:  /dev/sdc (4TB) → /mnt/backup
Backup Schedule: Daily at 3:00 AM

✓ Press [Enter] to apply, or [C] to customize.
```

If user presses **C**:
```
Customize Settings:
  1. Mount point [/mnt/data]: _
  2. Backup frequency [daily 3am]: _
  3. Filesystem [ext4]: _
```

If user presses **Enter**: Apply defaults immediately.

---

## 6. Strategy Selection Logic

```
IF available_disks == 0:
    Check for unmounted partitions (Edge Case C)
    IF found: Offer to mount
    ELSE: Strategy 0 (partition OS drive)

ELSE IF available_disks == 1:
    Mount as /mnt/data (simple single disk)
    
ELSE IF available_disks >= 2:
    IF hardware_raid_detected:
        Force ext4 only (Edge Case B)
    ELSE IF mixed_speed_classes (NVMe + HDD):
        Speed-Tiered Pools (Edge Case A)
    ELSE IF sizes_similar (within 10%):
        Rank: Strategy 2 (Mirror) > Strategy 3 (Backup)
    ELSE IF one_much_smaller (diff > 50%):
        Strategy 4 (Scratch + Vault)
    ELSE:
        Strategy 1 (MergerFS Pool)
```

---

## 6. Safety Constraints

1. **Boot Protection**: Never touch `/` or `/boot` partitions.
2. **Destructive Confirmation**: Require typing `ERASE` for disks with existing data.
3. **Hardware RAID Block**: Prevent ZFS/MDADM on virtual disks.
4. **Dependency Check**: Verify `mergerfs`/`zfs`/`mdadm` before applying.

---

## 7. TUI Output Examples

### Single Disk Detected
```
💾 Storage Configuration
────────────────────────
Only 1 drive detected. Recommending partition-based setup.

[1] ⭐ Create Data Partition (Recommended)
    • Root: 50GB (OS + Apps)
    • Data: 950GB (/mnt/data)
    • Warning: No redundancy!
```

### Two Identical HDDs (2x 4TB)
```
💾 Storage Configuration
────────────────────────
[1] ⭐ Mirror (ZFS) - Recommended
    • Capacity: 4TB usable
    • Protection: 1-disk fault tolerance

[2] Primary + Nightly Backup
    • Capacity: 4TB usable  
    • Protection: Hardware + ransomware protection

[3] Combined Pool (MergerFS)
    • Capacity: 8TB usable
    • Protection: None
```

### Mismatched Drives (500GB + 4TB)
```
💾 Storage Configuration
────────────────────────
[1] ⭐ Scratch + Vault - Recommended
    • Vault (4TB): /mnt/data
    • Scratch (500GB): /mnt/scratch
    • Reason: Avoids wasting 3.5TB in a mirror

[2] Combined Pool (MergerFS)
    • Capacity: 4.5TB combined
```
