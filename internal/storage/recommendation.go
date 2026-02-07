// Package storage provides intelligent storage orchestration for servctl.
// This file implements the storage recommendation engine.
package storage

import (
	"os/exec"
	"strings"
)

// StrategyID represents a storage configuration strategy
type StrategyID int

const (
	StrategyPartition    StrategyID = iota // 0: Single drive partitioning
	StrategyMergerFS                       // 1: MergerFS pool
	StrategyMirror                         // 2: ZFS/RAID mirror
	StrategyBackup                         // 3: Primary + backup
	StrategyScratchVault                   // 4: Scratch + vault
	StrategySpeedTiered                    // Edge: Speed-tiered pools
)

func (s StrategyID) String() string {
	switch s {
	case StrategyPartition:
		return "Partition Plan"
	case StrategyMergerFS:
		return "MergerFS Pool"
	case StrategyMirror:
		return "Mirror (Redundant)"
	case StrategyBackup:
		return "Primary + Backup"
	case StrategyScratchVault:
		return "Scratch + Vault"
	case StrategySpeedTiered:
		return "Speed-Tiered Pools"
	default:
		return "Unknown"
	}
}

// SpeedClass represents the speed tier of a disk
type SpeedClass int

const (
	SpeedClassSlow SpeedClass = iota // HDD
	SpeedClassFast                   // SSD, NVMe
)

// Strategy represents a possible storage configuration
type Strategy struct {
	ID          StrategyID
	Name        string
	Description string
	Capacity    string   // e.g., "4TB usable"
	Protection  string   // e.g., "1-disk fault tolerance"
	BestFor     string   // e.g., "Media libraries"
	Warning     string   // Optional warning message
	Score       int      // Higher = more recommended
	Recommended bool     // True for the top-ranked option
	Disks       []Disk   // Which disks this strategy uses
	MountPoints []string // Resulting mount points
}

// SystemInfo contains hardware info for recommendation decisions
type SystemInfo struct {
	TotalRAM        uint64 // Total RAM in bytes
	HasHardwareRAID bool   // True if virtual disks detected
}

// GetSystemInfo gathers system information for recommendations
func GetSystemInfo() SystemInfo {
	info := SystemInfo{}

	// Get RAM from /proc/meminfo
	cmd := exec.Command("grep", "MemTotal", "/proc/meminfo")
	output, err := cmd.Output()
	if err == nil {
		// Parse: "MemTotal:       16384000 kB"
		fields := strings.Fields(string(output))
		if len(fields) >= 2 {
			var kb uint64
			if _, err := exec.Command("sh", "-c",
				"grep MemTotal /proc/meminfo | awk '{print $2}'").Output(); err == nil {
				// Rough conversion
				if n, ok := parseUint(fields[1]); ok {
					info.TotalRAM = n * 1024 // KB to bytes
				}
			}
			_ = kb
		}
	}

	return info
}

// parseUint is a helper to parse uint64
func parseUint(s string) (uint64, bool) {
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + uint64(c-'0')
	}
	return n, true
}

// GetDiskSpeedClass returns the speed classification of a disk
func GetDiskSpeedClass(disk Disk) SpeedClass {
	switch disk.Type {
	case DiskTypeNVMe, DiskTypeSSD:
		return SpeedClassFast
	default:
		return SpeedClassSlow
	}
}

// IsHardwareRAID checks if a disk appears to be a hardware RAID virtual disk
func IsHardwareRAID(disk Disk) bool {
	model := strings.ToLower(disk.Model)
	raidIndicators := []string{
		"virtual disk",
		"perc",
		"megaraid",
		"smartarray",
		"raid",
		"logical volume",
	}
	for _, indicator := range raidIndicators {
		if strings.Contains(model, indicator) {
			return true
		}
	}
	return false
}

// AreSizesSimilar checks if two sizes are within a threshold (default 10%)
func AreSizesSimilar(size1, size2 uint64, thresholdPct float64) bool {
	if size1 == 0 || size2 == 0 {
		return false
	}
	larger := size1
	smaller := size2
	if size2 > size1 {
		larger = size2
		smaller = size1
	}
	diff := float64(larger-smaller) / float64(larger)
	return diff <= thresholdPct
}

// IsSizeMismatchLarge checks if size difference is greater than threshold
func IsSizeMismatchLarge(size1, size2 uint64, thresholdPct float64) bool {
	if size1 == 0 || size2 == 0 {
		return false
	}
	larger := size1
	smaller := size2
	if size2 > size1 {
		larger = size2
		smaller = size1
	}
	diff := float64(larger-smaller) / float64(larger)
	return diff > thresholdPct
}

// GenerateStrategies analyzes available disks and returns viable strategies
func GenerateStrategies(disks []Disk, info SystemInfo) []Strategy {
	var strategies []Strategy

	// Filter to available disks only
	available := FilterAvailableDisks(disks)

	// Check for hardware RAID
	hasHardwareRAID := false
	for _, d := range available {
		if IsHardwareRAID(d) {
			hasHardwareRAID = true
			break
		}
	}

	// Separate by speed class
	var fastDisks, slowDisks []Disk
	for _, d := range available {
		if GetDiskSpeedClass(d) == SpeedClassFast {
			fastDisks = append(fastDisks, d)
		} else {
			slowDisks = append(slowDisks, d)
		}
	}

	// Decision tree based on disk count and types
	switch len(available) {
	case 0:
		// No available disks - offer Strategy 0 (partition OS drive)
		strategies = append(strategies, Strategy{
			ID:          StrategyPartition,
			Name:        "Create Data Partition",
			Description: "Partition your OS drive to separate data from system",
			Capacity:    "Depends on available space",
			Protection:  "OS reinstall protection only",
			BestFor:     "Single-drive systems (NUC, laptop)",
			Warning:     "⚠️ No hardware redundancy!",
			Score:       50,
			MountPoints: []string{"/mnt/data"},
		})

	case 1:
		// Single data disk - simple mount
		disk := available[0]
		strategies = append(strategies, Strategy{
			ID:          StrategyPartition,
			Name:        "Single Data Disk",
			Description: "Format and mount as data drive",
			Capacity:    disk.SizeHuman,
			Protection:  "None",
			BestFor:     "Simple setup",
			Warning:     "⚠️ No redundancy",
			Score:       60,
			Disks:       []Disk{disk},
			MountPoints: []string{"/mnt/data"},
		})

	default:
		// Multiple disks - more options available

		// Edge Case: Hardware RAID detected
		if hasHardwareRAID {
			strategies = append(strategies, Strategy{
				ID:          StrategyPartition,
				Name:        "Simple Format (ext4)",
				Description: "Format drives individually. Your RAID card handles redundancy.",
				Capacity:    calculateTotalCapacity(available),
				Protection:  "Hardware RAID",
				BestFor:     "Enterprise servers with RAID controllers",
				Warning:     "Hardware RAID detected. Avoid software RAID.",
				Score:       90,
				Disks:       available,
				MountPoints: []string{"/mnt/data"},
			})
		}

		// Edge Case: Mixed speed classes (NVMe + HDD)
		if len(fastDisks) > 0 && len(slowDisks) > 0 {
			strategies = append(strategies, Strategy{
				ID:          StrategySpeedTiered,
				Name:        "Speed-Tiered Pools",
				Description: "Fast drives for active data, slow drives for archives",
				Capacity:    calculateTotalCapacity(available),
				Protection:  "None",
				BestFor:     "Mixed workloads (databases + media)",
				Score:       85,
				Disks:       available,
				MountPoints: []string{"/mnt/fast", "/mnt/data"},
			})
		}

		// Check if sizes are similar (within 10%)
		sizesSimilar := len(available) == 2 &&
			AreSizesSimilar(available[0].Size, available[1].Size, 0.10)

		// Check if sizes are wildly different (>50%)
		sizeMismatch := len(available) == 2 &&
			IsSizeMismatchLarge(available[0].Size, available[1].Size, 0.50)

		// Strategy 2: Mirror (if sizes similar and not hardware RAID)
		if sizesSimilar && !hasHardwareRAID {
			smallerSize := available[0].Size
			if available[1].Size < smallerSize {
				smallerSize = available[1].Size
			}

			fsType := "MDADM RAID1"
			if info.TotalRAM >= 8*1024*1024*1024 { // 8GB
				fsType = "ZFS Mirror"
			}

			strategies = append(strategies, Strategy{
				ID:          StrategyMirror,
				Name:        "Mirror (" + fsType + ")",
				Description: "Duplicate data across both drives for fault tolerance",
				Capacity:    formatBytes(smallerSize),
				Protection:  "1-disk fault tolerance",
				BestFor:     "Critical data, Nextcloud, databases",
				Score:       80,
				Disks:       available,
				MountPoints: []string{"/mnt/data"},
			})
		}

		// Strategy 3: Primary + Backup (if 2 disks, backupDisk >= primaryDisk)
		if len(available) == 2 {
			primary, backup := available[0], available[1]
			if backup.Size < primary.Size {
				primary, backup = backup, primary
			}

			strategies = append(strategies, Strategy{
				ID:          StrategyBackup,
				Name:        "Primary + Nightly Backup",
				Description: "One drive for data, one for automated backups",
				Capacity:    primary.SizeHuman,
				Protection:  "Hardware failure + user error protection",
				BestFor:     "Home users wanting 'set and forget' safety",
				Score:       75,
				Disks:       []Disk{primary, backup},
				MountPoints: []string{"/mnt/data", "/mnt/backup"},
			})
		}

		// Strategy 4: Scratch + Vault (if size mismatch > 50%)
		if sizeMismatch {
			large, small := available[0], available[1]
			if small.Size > large.Size {
				large, small = small, large
			}

			strategies = append(strategies, Strategy{
				ID:          StrategyScratchVault,
				Name:        "Scratch + Vault",
				Description: "Large drive for permanent data, small drive for downloads/temp",
				Capacity:    large.SizeHuman + " (vault) + " + small.SizeHuman + " (scratch)",
				Protection:  "None",
				BestFor:     "Optimizing mismatched drives",
				Score:       70,
				Disks:       []Disk{large, small},
				MountPoints: []string{"/mnt/data", "/mnt/scratch"},
			})
		}

		// Strategy 1: MergerFS Pool (always an option for 2+ disks)
		if !hasHardwareRAID {
			strategies = append(strategies, Strategy{
				ID:          StrategyMergerFS,
				Name:        "Combined Pool (MergerFS)",
				Description: "Combine all drives into one large pool",
				Capacity:    calculateTotalCapacity(available) + " (100% utilization)",
				Protection:  "Partial (only affected drive's files lost)",
				BestFor:     "Media libraries, maximum capacity",
				Score:       65,
				Disks:       available,
				MountPoints: []string{"/mnt/data"},
			})
		}
	}

	// Score and rank strategies
	strategies = ScoreStrategies(strategies)

	return strategies
}

// ScoreStrategies ranks strategies and marks the recommended one
func ScoreStrategies(strategies []Strategy) []Strategy {
	if len(strategies) == 0 {
		return strategies
	}

	// Find highest score
	maxScore := 0
	for _, s := range strategies {
		if s.Score > maxScore {
			maxScore = s.Score
		}
	}

	// Mark highest as recommended
	for i := range strategies {
		if strategies[i].Score == maxScore {
			strategies[i].Recommended = true
			break // Only one recommended
		}
	}

	// Sort by score (descending) - simple bubble sort
	for i := 0; i < len(strategies)-1; i++ {
		for j := 0; j < len(strategies)-i-1; j++ {
			if strategies[j].Score < strategies[j+1].Score {
				strategies[j], strategies[j+1] = strategies[j+1], strategies[j]
			}
		}
	}

	return strategies
}

// calculateTotalCapacity returns human-readable total capacity
func calculateTotalCapacity(disks []Disk) string {
	var total uint64
	for _, d := range disks {
		total += d.Size
	}
	return formatBytes(total)
}
