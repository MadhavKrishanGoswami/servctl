package storage

import (
	"testing"
)

// =============================================================================
// GenerateStrategies Tests
// =============================================================================

func TestGenerateStrategies_NoDisks(t *testing.T) {
	disks := []Disk{}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := GenerateStrategies(disks, info)

	if len(strategies) != 1 {
		t.Errorf("Expected 1 strategy for no disks, got %d", len(strategies))
	}
	if strategies[0].ID != StrategyPartition {
		t.Errorf("Expected StrategyPartition for no disks, got %v", strategies[0].ID)
	}
	if strategies[0].Name != "Create Data Partition" {
		t.Errorf("Expected 'Create Data Partition' name, got %s", strategies[0].Name)
	}
}

func TestGenerateStrategies_SingleDisk(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true},
	}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := GenerateStrategies(disks, info)

	if len(strategies) != 1 {
		t.Errorf("Expected 1 strategy for single disk, got %d", len(strategies))
	}
	if strategies[0].ID != StrategyPartition {
		t.Errorf("Expected StrategyPartition for single disk, got %v", strategies[0].ID)
	}
	if len(strategies[0].Disks) != 1 {
		t.Errorf("Expected strategy to have 1 disk, got %d", len(strategies[0].Disks))
	}
}

func TestGenerateStrategies_TwoSimilarDisks(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: DiskTypeHDD},
		{Path: "/dev/sdc", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: DiskTypeHDD},
	}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024} // 16GB RAM - eligible for ZFS

	strategies := GenerateStrategies(disks, info)

	// Should include Mirror, Backup, and MergerFS strategies
	if len(strategies) < 3 {
		t.Errorf("Expected at least 3 strategies for two similar disks, got %d", len(strategies))
	}

	// Check for Mirror strategy
	hasMirror := false
	for _, s := range strategies {
		if s.ID == StrategyMirror {
			hasMirror = true
			if !contains(s.Name, "ZFS") {
				t.Errorf("Expected ZFS mirror with 16GB RAM, got %s", s.Name)
			}
		}
	}
	if !hasMirror {
		t.Error("Expected Mirror strategy for two similar disks")
	}
}

func TestGenerateStrategies_TwoSimilarDisks_LowRAM(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: DiskTypeHDD},
		{Path: "/dev/sdc", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: DiskTypeHDD},
	}
	info := SystemInfo{TotalRAM: 4 * 1024 * 1024 * 1024} // 4GB RAM - MDADM only

	strategies := GenerateStrategies(disks, info)

	// Check for MDADM mirror (not ZFS due to low RAM)
	for _, s := range strategies {
		if s.ID == StrategyMirror {
			if contains(s.Name, "ZFS") {
				t.Errorf("Expected MDADM mirror with 4GB RAM, got %s", s.Name)
			}
		}
	}
}

func TestGenerateStrategies_MismatchedSizes(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: DiskTypeHDD},
		{Path: "/dev/sdc", Size: 500 * 1024 * 1024 * 1024, SizeHuman: "500GB", IsAvailable: true, Type: DiskTypeHDD},
	}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := GenerateStrategies(disks, info)

	// Should include ScratchVault for mismatched sizes
	hasScratchVault := false
	for _, s := range strategies {
		if s.ID == StrategyScratchVault {
			hasScratchVault = true
		}
	}
	if !hasScratchVault {
		t.Error("Expected ScratchVault strategy for mismatched disk sizes")
	}
}

func TestGenerateStrategies_MixedSpeedClasses(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/nvme0n1", Size: 1 * 1024 * 1024 * 1024 * 1024, SizeHuman: "1TB", IsAvailable: true, Type: DiskTypeNVMe, Transport: "nvme"},
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: DiskTypeHDD, Rotational: true},
	}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := GenerateStrategies(disks, info)

	// Should include SpeedTiered for mixed NVMe + HDD
	hasSpeedTiered := false
	for _, s := range strategies {
		if s.ID == StrategySpeedTiered {
			hasSpeedTiered = true
		}
	}
	if !hasSpeedTiered {
		t.Error("Expected SpeedTiered strategy for mixed NVMe + HDD")
	}
}

func TestGenerateStrategies_HardwareRAID(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 1 * 1024 * 1024 * 1024 * 1024, SizeHuman: "1TB", IsAvailable: true, Model: "PERC H730 Virtual Disk"},
		{Path: "/dev/sdc", Size: 1 * 1024 * 1024 * 1024 * 1024, SizeHuman: "1TB", IsAvailable: true, Model: "PERC H730 Virtual Disk"},
	}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := GenerateStrategies(disks, info)

	// Should NOT include Mirror or MergerFS for hardware RAID
	for _, s := range strategies {
		if s.ID == StrategyMirror {
			t.Error("Should not suggest software mirror for hardware RAID")
		}
		if s.ID == StrategyMergerFS {
			t.Error("Should not suggest MergerFS for hardware RAID")
		}
	}

	// Should suggest simple format
	hasPartition := false
	for _, s := range strategies {
		if s.ID == StrategyPartition {
			hasPartition = true
			if !contains(s.Warning, "Hardware RAID") {
				t.Error("Expected Hardware RAID warning")
			}
		}
	}
	if !hasPartition {
		t.Error("Expected simple format strategy for hardware RAID")
	}
}

// =============================================================================
// ScoreStrategies Tests
// =============================================================================

func TestScoreStrategies_Ranking(t *testing.T) {
	strategies := []Strategy{
		{ID: StrategyMergerFS, Name: "MergerFS", Score: 65},
		{ID: StrategyMirror, Name: "Mirror", Score: 80},
		{ID: StrategyBackup, Name: "Backup", Score: 75},
	}

	scored := ScoreStrategies(strategies)

	// Highest score should be first
	if scored[0].Score != 80 {
		t.Errorf("Expected highest score first, got %d", scored[0].Score)
	}
	if scored[0].ID != StrategyMirror {
		t.Errorf("Expected Mirror to be first, got %v", scored[0].ID)
	}

	// Highest should be marked recommended
	if !scored[0].Recommended {
		t.Error("Highest scoring strategy should be marked Recommended")
	}

	// Others should not be recommended
	for i := 1; i < len(scored); i++ {
		if scored[i].Recommended {
			t.Errorf("Strategy %d should not be recommended", i)
		}
	}
}

func TestScoreStrategies_Empty(t *testing.T) {
	strategies := []Strategy{}

	scored := ScoreStrategies(strategies)

	if len(scored) != 0 {
		t.Errorf("Expected empty result, got %d", len(scored))
	}
}

// =============================================================================
// FilterAvailableDisks Tests
// =============================================================================

func TestFilterAvailableDisks_RecommendationModule(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sda", IsOSDisk: true},  // OS disk - excluded
		{Path: "/dev/sdb", IsOSDisk: false}, // Available
		{Path: "/dev/sdc", IsOSDisk: false}, // Available
		{Path: "/dev/sdd", Removable: true}, // Removable - excluded
	}

	available := FilterAvailableDisks(disks)

	if len(available) != 2 {
		t.Errorf("Expected 2 available disks, got %d", len(available))
	}
	for _, d := range available {
		if d.IsOSDisk || d.Removable {
			t.Errorf("Disk %s should not be OS or removable", d.Path)
		}
	}
}

// =============================================================================
// Size Comparison Tests
// =============================================================================

func TestAreSizesSimilar(t *testing.T) {
	tests := []struct {
		a, b      uint64
		threshold float64
		expected  bool
	}{
		{1000, 1000, 0.10, true},  // Exact match
		{1000, 1050, 0.10, true},  // 5% diff
		{1000, 1100, 0.10, true},  // 10% diff (edge)
		{1000, 1200, 0.10, false}, // 20% diff
		{4000, 500, 0.10, false},  // Way different
	}

	for _, tt := range tests {
		result := AreSizesSimilar(tt.a, tt.b, tt.threshold)
		if result != tt.expected {
			t.Errorf("AreSizesSimilar(%d, %d, %.2f) = %v, want %v", tt.a, tt.b, tt.threshold, result, tt.expected)
		}
	}
}

func TestIsSizeMismatchLarge(t *testing.T) {
	tests := []struct {
		a, b      uint64
		threshold float64
		expected  bool
	}{
		{4000, 500, 0.50, true},   // >50% diff
		{1000, 600, 0.50, false},  // 40% diff
		{1000, 1000, 0.50, false}, // Same
	}

	for _, tt := range tests {
		result := IsSizeMismatchLarge(tt.a, tt.b, tt.threshold)
		if result != tt.expected {
			t.Errorf("IsSizeMismatchLarge(%d, %d, %.2f) = %v, want %v", tt.a, tt.b, tt.threshold, result, tt.expected)
		}
	}
}

// =============================================================================
// Hardware Detection Tests
// =============================================================================

func TestIsHardwareRAID(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"PERC H730 Virtual Disk", true},
		{"MegaRAID LD 0", true},
		{"HP LOGICAL VOLUME", true},
		{"Samsung SSD 870", false},
		{"WDC WD40EFRX", false},
		{"", false},
	}

	for _, tt := range tests {
		disk := Disk{Model: tt.model}
		result := IsHardwareRAID(disk)
		if result != tt.expected {
			t.Errorf("IsHardwareRAID(%s) = %v, want %v", tt.model, result, tt.expected)
		}
	}
}

func TestGetDiskSpeedClass(t *testing.T) {
	tests := []struct {
		disk     Disk
		expected SpeedClass
	}{
		{Disk{Type: DiskTypeNVMe}, SpeedClassFast},
		{Disk{Type: DiskTypeSSD}, SpeedClassFast},
		{Disk{Type: DiskTypeHDD}, SpeedClassSlow},
		{Disk{Type: DiskTypeUnknown}, SpeedClassSlow}, // Default to slow
		{Disk{Rotational: true}, SpeedClassSlow},      // Type not set
	}

	for _, tt := range tests {
		result := GetDiskSpeedClass(tt.disk)
		if result != tt.expected {
			t.Errorf("GetDiskSpeedClass(%+v) = %v, want %v", tt.disk, result, tt.expected)
		}
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
