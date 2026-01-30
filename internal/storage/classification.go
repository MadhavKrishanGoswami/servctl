package storage

import (
	"fmt"
	"strings"
)

// StorageRank represents one of the "5 Ranks" storage configurations
type StorageRank int

const (
	RankHybrid      StorageRank = iota + 1 // SSD (OS/Apps) + HDD (Bulk Data)
	RankSpeedDemon                         // SSD (OS) + SSD (Active DBs)
	RankMirror                             // 2x SSD RAID 1
	RankDataHoarder                        // 2x HDD RAID 1
	RankKamikaze                           // RAID 0 (any combination)
)

func (r StorageRank) String() string {
	switch r {
	case RankHybrid:
		return "Rank 1: Hybrid"
	case RankSpeedDemon:
		return "Rank 2: Speed Demon"
	case RankMirror:
		return "Rank 3: Mirror"
	case RankDataHoarder:
		return "Rank 4: Data Hoarder"
	case RankKamikaze:
		return "Rank 5: Kamikaze"
	default:
		return "Unknown Rank"
	}
}

// StorageRecommendation represents a storage configuration recommendation
type StorageRecommendation struct {
	Rank        StorageRank
	Name        string
	Description string
	Pros        []string
	Cons        []string
	Warning     string // Critical warning if any
	IsDefault   bool   // Recommended default choice
	Disks       []DiskAssignment
}

// DiskAssignment represents how a disk should be used
type DiskAssignment struct {
	Disk  *Disk
	Role  string // "os", "apps", "data", "backup", "raid"
	Label string // Filesystem label
	Mount string // Mount point
}

// DiskScenario represents the detected disk configuration scenario
type DiskScenario int

const (
	ScenarioSingleDisk DiskScenario = iota
	ScenarioTwoDisk
	ScenarioMultiDisk
)

func (s DiskScenario) String() string {
	switch s {
	case ScenarioSingleDisk:
		return "Single Disk"
	case ScenarioTwoDisk:
		return "Two Disks"
	case ScenarioMultiDisk:
		return "Multi-Disk (3+)"
	default:
		return "Unknown"
	}
}

// ClassificationResult contains the results of disk classification
type ClassificationResult struct {
	Scenario        DiskScenario
	OSDisk          *Disk
	AvailableDisks  []Disk
	SSDs            []Disk
	HDDs            []Disk
	NVMes           []Disk
	Recommendations []StorageRecommendation
	SelectedRank    *StorageRecommendation
}

// ClassifyDisks analyzes available disks and generates recommendations
func ClassifyDisks(disks []Disk) *ClassificationResult {
	result := &ClassificationResult{}

	// Find OS disk
	for i := range disks {
		if disks[i].IsOSDisk {
			result.OSDisk = &disks[i]
			break
		}
	}

	// Filter available disks (non-OS, non-removable)
	for _, disk := range disks {
		if disk.IsOSDisk || disk.Removable {
			continue
		}
		result.AvailableDisks = append(result.AvailableDisks, disk)

		// Categorize by type
		switch disk.Type {
		case DiskTypeSSD:
			result.SSDs = append(result.SSDs, disk)
		case DiskTypeHDD:
			result.HDDs = append(result.HDDs, disk)
		case DiskTypeNVMe:
			result.NVMes = append(result.NVMes, disk)
		}
	}

	// Determine scenario
	availableCount := len(result.AvailableDisks)
	switch {
	case availableCount == 0:
		// Only OS disk available
		result.Scenario = ScenarioSingleDisk
	case availableCount == 1:
		result.Scenario = ScenarioSingleDisk
	case availableCount == 2:
		result.Scenario = ScenarioTwoDisk
	default:
		result.Scenario = ScenarioMultiDisk
	}

	// Generate recommendations based on scenario
	result.Recommendations = generateRecommendations(result)

	return result
}

// generateRecommendations creates storage recommendations based on available disks
func generateRecommendations(result *ClassificationResult) []StorageRecommendation {
	var recommendations []StorageRecommendation

	switch result.Scenario {
	case ScenarioSingleDisk:
		recommendations = generateSingleDiskRecommendations(result)
	case ScenarioTwoDisk:
		recommendations = generateTwoDiskRecommendations(result)
	case ScenarioMultiDisk:
		recommendations = generateMultiDiskRecommendations(result)
	}

	return recommendations
}

// generateSingleDiskRecommendations handles the single disk scenario
func generateSingleDiskRecommendations(result *ClassificationResult) []StorageRecommendation {
	var recommendations []StorageRecommendation

	var dataDisk *Disk
	if len(result.AvailableDisks) > 0 {
		dataDisk = &result.AvailableDisks[0]
	}

	rec := StorageRecommendation{
		Rank:        RankHybrid, // Using Rank 1 naming even for single disk
		Name:        "Single Disk Configuration",
		Description: "All data stored on one disk alongside the OS.",
		Pros: []string{
			"Simple setup",
			"Maximum usable space",
		},
		Cons: []string{
			"No redundancy - disk failure = data loss",
			"No performance separation",
		},
		Warning:   "⚠️ SINGLE POINT OF FAILURE: Consider adding a backup disk!",
		IsDefault: true,
	}

	if dataDisk != nil {
		rec.Disks = []DiskAssignment{
			{
				Disk:  dataDisk,
				Role:  "data",
				Label: "servctl-data",
				Mount: "/mnt/data",
			},
		}
	}

	recommendations = append(recommendations, rec)
	return recommendations
}

// generateTwoDiskRecommendations generates "The 5 Ranks" for two-disk scenarios
func generateTwoDiskRecommendations(result *ClassificationResult) []StorageRecommendation {
	var recommendations []StorageRecommendation

	disks := result.AvailableDisks
	hasSSD := len(result.SSDs) > 0 || len(result.NVMes) > 0
	hasHDD := len(result.HDDs) > 0

	// Combine NVMe and SSD for simplicity
	allSSDs := append(result.SSDs, result.NVMes...)

	// Rank 1: Hybrid (SSD + HDD)
	if hasSSD && hasHDD {
		var ssd, hdd *Disk
		if len(allSSDs) > 0 {
			ssd = &allSSDs[0]
		}
		if len(result.HDDs) > 0 {
			hdd = &result.HDDs[0]
		}

		rec := StorageRecommendation{
			Rank:        RankHybrid,
			Name:        "Hybrid: SSD + HDD",
			Description: "SSD for OS/Apps/Databases, HDD for bulk media storage.",
			Pros: []string{
				"Best of both worlds",
				"Fast access for critical data",
				"Cost-effective bulk storage",
			},
			Cons: []string{
				"No redundancy",
				"HDD failure loses media files",
			},
			IsDefault: true,
		}
		if ssd != nil && hdd != nil {
			rec.Disks = []DiskAssignment{
				{Disk: ssd, Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
				{Disk: hdd, Role: "data", Label: "servctl-data", Mount: "/mnt/data"},
			}
		}
		recommendations = append(recommendations, rec)
	}

	// Rank 2: Speed Demon (SSD + SSD)
	if len(allSSDs) >= 2 {
		rec := StorageRecommendation{
			Rank:        RankSpeedDemon,
			Name:        "Speed Demon: Dual SSD",
			Description: "One SSD for OS, another for active databases.",
			Pros: []string{
				"Maximum I/O performance",
				"Parallel access for DBs",
				"Low latency for all operations",
			},
			Cons: []string{
				"No redundancy",
				"Limited bulk storage capacity",
				"More expensive per GB",
			},
			IsDefault: !hasSSD || !hasHDD, // Default if no hybrid option
		}
		rec.Disks = []DiskAssignment{
			{Disk: &allSSDs[0], Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
			{Disk: &allSSDs[1], Role: "data", Label: "servctl-data", Mount: "/mnt/data"},
		}
		recommendations = append(recommendations, rec)
	}

	// Rank 3: Mirror (2x SSD RAID 1)
	if len(allSSDs) >= 2 {
		rec := StorageRecommendation{
			Rank:        RankMirror,
			Name:        "Mirror: SSD RAID 1",
			Description: "Two SSDs in RAID 1 for redundancy.",
			Pros: []string{
				"Data redundancy",
				"Survives single disk failure",
				"Fast read performance",
			},
			Cons: []string{
				"50% storage capacity loss",
				"Slower writes than single disk",
				"More complex setup",
			},
		}
		rec.Disks = []DiskAssignment{
			{Disk: &allSSDs[0], Role: "raid", Label: "servctl-raid", Mount: "/mnt/data"},
			{Disk: &allSSDs[1], Role: "raid", Label: "servctl-raid", Mount: "/mnt/data"},
		}
		recommendations = append(recommendations, rec)
	}

	// Rank 4: Data Hoarder (2x HDD RAID 1)
	if len(result.HDDs) >= 2 {
		rec := StorageRecommendation{
			Rank:        RankDataHoarder,
			Name:        "Data Hoarder: HDD RAID 1",
			Description: "Two HDDs in RAID 1 for redundant bulk storage.",
			Pros: []string{
				"Data redundancy",
				"Cost-effective for large storage",
				"Survives single disk failure",
			},
			Cons: []string{
				"Slower than SSD",
				"50% storage capacity loss",
			},
			Warning: "⚠️ PERFORMANCE WARNING: HDD RAID 1 will be slower than SSD configurations.",
		}
		rec.Disks = []DiskAssignment{
			{Disk: &result.HDDs[0], Role: "raid", Label: "servctl-raid", Mount: "/mnt/data"},
			{Disk: &result.HDDs[1], Role: "raid", Label: "servctl-raid", Mount: "/mnt/data"},
		}
		recommendations = append(recommendations, rec)
	}

	// Rank 5: Kamikaze (RAID 0 - any combination)
	if len(disks) >= 2 {
		rec := StorageRecommendation{
			Rank:        RankKamikaze,
			Name:        "Kamikaze: RAID 0",
			Description: "Stripe data across disks for maximum speed. NO REDUNDANCY!",
			Pros: []string{
				"Maximum combined capacity",
				"Fastest write speeds",
				"Best for temporary/replaceable data",
			},
			Cons: []string{
				"ANY disk failure = TOTAL DATA LOSS",
				"Not suitable for important data",
				"Higher failure probability",
			},
			Warning: "🚨 CRITICAL RISK: RAID 0 offers NO redundancy. Disk failure will result in COMPLETE data loss!",
		}
		rec.Disks = []DiskAssignment{
			{Disk: &disks[0], Role: "raid0", Label: "servctl-stripe", Mount: "/mnt/data"},
			{Disk: &disks[1], Role: "raid0", Label: "servctl-stripe", Mount: "/mnt/data"},
		}
		recommendations = append(recommendations, rec)
	}

	// If no recommendations yet, add a generic one
	if len(recommendations) == 0 && len(disks) >= 2 {
		rec := StorageRecommendation{
			Rank:        RankHybrid,
			Name:        "Standard: Separate Storage",
			Description: "Use disks for separate purposes.",
			Pros: []string{
				"Simple configuration",
				"Easy to manage",
			},
			Cons: []string{
				"No redundancy",
			},
			IsDefault: true,
		}
		rec.Disks = []DiskAssignment{
			{Disk: &disks[0], Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
			{Disk: &disks[1], Role: "data", Label: "servctl-data", Mount: "/mnt/data"},
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateMultiDiskRecommendations handles 3+ disk scenarios
func generateMultiDiskRecommendations(result *ClassificationResult) []StorageRecommendation {
	var recommendations []StorageRecommendation

	disks := result.AvailableDisks
	allSSDs := append(result.SSDs, result.NVMes...)

	// Optimal: SSD (apps) + HDD (data) + HDD (backup)
	if len(allSSDs) >= 1 && len(result.HDDs) >= 2 {
		rec := StorageRecommendation{
			Rank:        RankHybrid,
			Name:        "Optimal: SSD + HDD + Backup",
			Description: "SSD for apps/DBs, HDD for data, second HDD for backups.",
			Pros: []string{
				"Fast app performance",
				"Dedicated backup disk",
				"3-2-1 backup capability",
				"Cost-effective",
			},
			Cons: []string{
				"Backup requires manual sync",
				"HDDs are slower than SSDs",
			},
			IsDefault: true,
		}
		rec.Disks = []DiskAssignment{
			{Disk: &allSSDs[0], Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
			{Disk: &result.HDDs[0], Role: "data", Label: "servctl-data", Mount: "/mnt/data"},
			{Disk: &result.HDDs[1], Role: "backup", Label: "servctl-backup", Mount: "/mnt/backup"},
		}
		recommendations = append(recommendations, rec)
	}

	// Performance: SSD (apps) + SSD (data) + HDD (backup)
	if len(allSSDs) >= 2 && len(result.HDDs) >= 1 {
		rec := StorageRecommendation{
			Rank:        RankSpeedDemon,
			Name:        "Performance: Dual SSD + Backup",
			Description: "SSDs for all active data, HDD for backups.",
			Pros: []string{
				"Maximum performance",
				"Silent operation",
				"Dedicated backup",
			},
			Cons: []string{
				"More expensive",
				"Less bulk storage",
			},
		}
		rec.Disks = []DiskAssignment{
			{Disk: &allSSDs[0], Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
			{Disk: &allSSDs[1], Role: "data", Label: "servctl-data", Mount: "/mnt/data"},
			{Disk: &result.HDDs[0], Role: "backup", Label: "servctl-backup", Mount: "/mnt/backup"},
		}
		recommendations = append(recommendations, rec)
	}

	// Storage-focused: HDD RAID + SSD cache
	if len(result.HDDs) >= 2 && len(allSSDs) >= 1 {
		rec := StorageRecommendation{
			Rank:        RankDataHoarder,
			Name:        "Massive Storage: RAID + SSD Cache",
			Description: "HDD RAID for bulk storage, SSD for apps/cache.",
			Pros: []string{
				"Maximum storage with redundancy",
				"Fast app performance",
				"Protects against HDD failure",
			},
			Cons: []string{
				"Complex setup",
				"Slower bulk access",
			},
		}
		rec.Disks = []DiskAssignment{
			{Disk: &allSSDs[0], Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
			{Disk: &result.HDDs[0], Role: "raid", Label: "servctl-raid", Mount: "/mnt/data"},
			{Disk: &result.HDDs[1], Role: "raid", Label: "servctl-raid", Mount: "/mnt/data"},
		}
		recommendations = append(recommendations, rec)
	}

	// Generic fallback
	if len(recommendations) == 0 && len(disks) >= 3 {
		rec := StorageRecommendation{
			Rank:        RankHybrid,
			Name:        "Standard: Apps + Data + Backup",
			Description: "Separate disks for different purposes.",
			IsDefault:   true,
		}
		rec.Disks = []DiskAssignment{
			{Disk: &disks[0], Role: "apps", Label: "servctl-apps", Mount: "/mnt/apps"},
			{Disk: &disks[1], Role: "data", Label: "servctl-data", Mount: "/mnt/data"},
			{Disk: &disks[2], Role: "backup", Label: "servctl-backup", Mount: "/mnt/backup"},
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// GetDefaultRecommendation returns the default (recommended) storage configuration
func GetDefaultRecommendation(recommendations []StorageRecommendation) *StorageRecommendation {
	for i := range recommendations {
		if recommendations[i].IsDefault {
			return &recommendations[i]
		}
	}
	if len(recommendations) > 0 {
		return &recommendations[0]
	}
	return nil
}

// FormatRecommendationSummary returns a formatted summary of a recommendation
func FormatRecommendationSummary(rec *StorageRecommendation) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s: %s\n", rec.Rank.String(), rec.Name))
	b.WriteString(fmt.Sprintf("  %s\n", rec.Description))

	if len(rec.Disks) > 0 {
		b.WriteString("  Disk assignments:\n")
		for _, assignment := range rec.Disks {
			if assignment.Disk != nil {
				b.WriteString(fmt.Sprintf("    • %s (%s %s) → %s [%s]\n",
					assignment.Disk.Name,
					assignment.Disk.Type.String(),
					assignment.Disk.SizeHuman,
					assignment.Role,
					assignment.Mount,
				))
			}
		}
	}

	return b.String()
}
