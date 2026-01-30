package storage

import (
	"testing"
)

func TestStorageRankString(t *testing.T) {
	tests := []struct {
		rank     StorageRank
		expected string
	}{
		{RankHybrid, "Rank 1: Hybrid"},
		{RankSpeedDemon, "Rank 2: Speed Demon"},
		{RankMirror, "Rank 3: Mirror"},
		{RankDataHoarder, "Rank 4: Data Hoarder"},
		{RankKamikaze, "Rank 5: Kamikaze"},
		{StorageRank(99), "Unknown Rank"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.rank.String(); got != tt.expected {
				t.Errorf("StorageRank.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiskScenarioString(t *testing.T) {
	tests := []struct {
		scenario DiskScenario
		expected string
	}{
		{ScenarioSingleDisk, "Single Disk"},
		{ScenarioTwoDisk, "Two Disks"},
		{ScenarioMultiDisk, "Multi-Disk (3+)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.scenario.String(); got != tt.expected {
				t.Errorf("DiskScenario.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClassifyDisksSingleDisk(t *testing.T) {
	disks := []Disk{
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: true, Size: 500000000000},
		{Name: "sdb", Type: DiskTypeHDD, IsOSDisk: false, Size: 1000000000000},
	}

	result := ClassifyDisks(disks)

	if result.Scenario != ScenarioSingleDisk {
		t.Errorf("ClassifyDisks() scenario = %v, want %v", result.Scenario, ScenarioSingleDisk)
	}

	if result.OSDisk == nil {
		t.Error("ClassifyDisks() OSDisk should not be nil")
	}

	if len(result.AvailableDisks) != 1 {
		t.Errorf("ClassifyDisks() available = %d, want %d", len(result.AvailableDisks), 1)
	}
}

func TestClassifyDisksTwoDisk(t *testing.T) {
	disks := []Disk{
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: true, Size: 500000000000},
		{Name: "sdb", Type: DiskTypeSSD, IsOSDisk: false, Size: 500000000000},
		{Name: "sdc", Type: DiskTypeHDD, IsOSDisk: false, Size: 2000000000000},
	}

	result := ClassifyDisks(disks)

	if result.Scenario != ScenarioTwoDisk {
		t.Errorf("ClassifyDisks() scenario = %v, want %v", result.Scenario, ScenarioTwoDisk)
	}

	if len(result.AvailableDisks) != 2 {
		t.Errorf("ClassifyDisks() available = %d, want %d", len(result.AvailableDisks), 2)
	}

	// Should have recommendations
	if len(result.Recommendations) == 0 {
		t.Error("ClassifyDisks() should provide recommendations for two disks")
	}
}

func TestClassifyDisksMultiDisk(t *testing.T) {
	disks := []Disk{
		{Name: "nvme0n1", Type: DiskTypeNVMe, IsOSDisk: true, Size: 500000000000},
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: false, Size: 500000000000},
		{Name: "sdb", Type: DiskTypeHDD, IsOSDisk: false, Size: 2000000000000},
		{Name: "sdc", Type: DiskTypeHDD, IsOSDisk: false, Size: 2000000000000},
	}

	result := ClassifyDisks(disks)

	if result.Scenario != ScenarioMultiDisk {
		t.Errorf("ClassifyDisks() scenario = %v, want %v", result.Scenario, ScenarioMultiDisk)
	}

	if len(result.AvailableDisks) != 3 {
		t.Errorf("ClassifyDisks() available = %d, want %d", len(result.AvailableDisks), 3)
	}
}

func TestClassifyDisksCategorizesTypes(t *testing.T) {
	disks := []Disk{
		{Name: "nvme0n1", Type: DiskTypeNVMe, IsOSDisk: true},
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: false},
		{Name: "sdb", Type: DiskTypeHDD, IsOSDisk: false},
		{Name: "sdc", Type: DiskTypeHDD, IsOSDisk: false},
	}

	result := ClassifyDisks(disks)

	if len(result.SSDs) != 1 {
		t.Errorf("ClassifyDisks() SSDs = %d, want %d", len(result.SSDs), 1)
	}
	if len(result.HDDs) != 2 {
		t.Errorf("ClassifyDisks() HDDs = %d, want %d", len(result.HDDs), 2)
	}
	if len(result.NVMes) != 0 { // NVMe is OS disk, so not in available
		t.Errorf("ClassifyDisks() NVMes = %d, want %d", len(result.NVMes), 0)
	}
}

func TestGetDefaultRecommendation(t *testing.T) {
	recommendations := []StorageRecommendation{
		{Rank: RankSpeedDemon, IsDefault: false},
		{Rank: RankHybrid, IsDefault: true},
		{Rank: RankMirror, IsDefault: false},
	}

	def := GetDefaultRecommendation(recommendations)

	if def == nil {
		t.Fatal("GetDefaultRecommendation() returned nil")
	}

	if def.Rank != RankHybrid {
		t.Errorf("GetDefaultRecommendation() = %v, want %v", def.Rank, RankHybrid)
	}
}

func TestGetDefaultRecommendationNoDefault(t *testing.T) {
	recommendations := []StorageRecommendation{
		{Rank: RankSpeedDemon, IsDefault: false},
		{Rank: RankMirror, IsDefault: false},
	}

	def := GetDefaultRecommendation(recommendations)

	// Should return first item when no default
	if def == nil {
		t.Fatal("GetDefaultRecommendation() returned nil")
	}

	if def.Rank != RankSpeedDemon {
		t.Errorf("GetDefaultRecommendation() = %v, want first item %v", def.Rank, RankSpeedDemon)
	}
}

func TestGetDefaultRecommendationEmpty(t *testing.T) {
	recommendations := []StorageRecommendation{}

	def := GetDefaultRecommendation(recommendations)

	if def != nil {
		t.Error("GetDefaultRecommendation() should return nil for empty list")
	}
}

func TestStorageRecommendationStructure(t *testing.T) {
	rec := StorageRecommendation{
		Rank:        RankHybrid,
		Name:        "Hybrid: SSD + HDD",
		Description: "Best of both worlds",
		Pros:        []string{"Fast apps", "Cheap storage"},
		Cons:        []string{"No redundancy"},
		Warning:     "",
		IsDefault:   true,
		Disks: []DiskAssignment{
			{Role: "apps", Mount: "/mnt/apps"},
			{Role: "data", Mount: "/mnt/data"},
		},
	}

	if rec.Rank != RankHybrid {
		t.Errorf("StorageRecommendation.Rank = %v, want %v", rec.Rank, RankHybrid)
	}
	if !rec.IsDefault {
		t.Error("StorageRecommendation.IsDefault should be true")
	}
	if len(rec.Disks) != 2 {
		t.Errorf("len(StorageRecommendation.Disks) = %v, want %v", len(rec.Disks), 2)
	}
}

func TestFormatRecommendationSummary(t *testing.T) {
	ssd := Disk{Name: "sda", Path: "/dev/sda", Type: DiskTypeSSD, SizeHuman: "500 GB"}
	hdd := Disk{Name: "sdb", Path: "/dev/sdb", Type: DiskTypeHDD, SizeHuman: "2 TB"}

	rec := StorageRecommendation{
		Rank:        RankHybrid,
		Name:        "Hybrid: SSD + HDD",
		Description: "SSD for apps, HDD for data",
		Disks: []DiskAssignment{
			{Disk: &ssd, Role: "apps", Mount: "/mnt/apps"},
			{Disk: &hdd, Role: "data", Mount: "/mnt/data"},
		},
	}

	summary := FormatRecommendationSummary(&rec)

	if summary == "" {
		t.Error("FormatRecommendationSummary() returned empty string")
	}

	// Check that key info is present
	if !containsString(summary, "Hybrid") {
		t.Error("Summary should contain 'Hybrid'")
	}
	// FormatRecommendationSummary uses Disk.Name, not Disk.Path
	if !containsString(summary, "sda") {
		t.Error("Summary should contain disk name 'sda'")
	}
}

func TestHybridRecommendationGeneration(t *testing.T) {
	// Simulate SSD + HDD scenario
	disks := []Disk{
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: true, Size: 500000000000},
		{Name: "sdb", Type: DiskTypeSSD, IsOSDisk: false, Size: 500000000000},
		{Name: "sdc", Type: DiskTypeHDD, IsOSDisk: false, Size: 2000000000000},
	}

	result := ClassifyDisks(disks)

	// Should have hybrid recommendation
	hasHybrid := false
	for _, rec := range result.Recommendations {
		if rec.Rank == RankHybrid {
			hasHybrid = true
			if !rec.IsDefault {
				t.Error("Hybrid should be default when both SSD and HDD available")
			}
			break
		}
	}

	if !hasHybrid {
		t.Error("Should generate Hybrid recommendation when SSD + HDD available")
	}
}

func TestKamikazeWarning(t *testing.T) {
	disks := []Disk{
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: true, Size: 500000000000},
		{Name: "sdb", Type: DiskTypeSSD, IsOSDisk: false, Size: 500000000000},
		{Name: "sdc", Type: DiskTypeSSD, IsOSDisk: false, Size: 500000000000},
	}

	result := ClassifyDisks(disks)

	// Find Kamikaze recommendation and verify warning
	for _, rec := range result.Recommendations {
		if rec.Rank == RankKamikaze {
			if rec.Warning == "" {
				t.Error("Kamikaze should have a critical warning")
			}
			return
		}
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkClassifyDisks(b *testing.B) {
	disks := []Disk{
		{Name: "sda", Type: DiskTypeSSD, IsOSDisk: true, Size: 500000000000},
		{Name: "sdb", Type: DiskTypeSSD, IsOSDisk: false, Size: 500000000000},
		{Name: "sdc", Type: DiskTypeHDD, IsOSDisk: false, Size: 2000000000000},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClassifyDisks(disks)
	}
}

func BenchmarkGetDefaultRecommendation(b *testing.B) {
	recommendations := []StorageRecommendation{
		{Rank: RankSpeedDemon, IsDefault: false},
		{Rank: RankHybrid, IsDefault: true},
		{Rank: RankMirror, IsDefault: false},
		{Rank: RankDataHoarder, IsDefault: false},
		{Rank: RankKamikaze, IsDefault: false},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDefaultRecommendation(recommendations)
	}
}
