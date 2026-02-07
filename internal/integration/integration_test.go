package integration

import (
	"testing"

	"github.com/madhav/servctl/internal/compose"
	"github.com/madhav/servctl/internal/directory"
	"github.com/madhav/servctl/internal/maintenance"
	"github.com/madhav/servctl/internal/storage"
)

// =============================================================================
// Full Wizard Flow Integration Tests
// =============================================================================

// TestWizardFlow_SingleDisk simulates the wizard flow for a single disk setup
func TestWizardFlow_SingleDisk(t *testing.T) {
	// Phase 2: Storage Strategy
	disks := []storage.Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true},
	}
	sysInfo := storage.SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := storage.GenerateStrategies(disks, sysInfo)
	if len(strategies) != 1 {
		t.Errorf("Single disk should generate 1 strategy, got %d", len(strategies))
	}

	// Phase 3: Directory Structure
	serviceSel := directory.DefaultServiceSelection()
	dirs := directory.GetDirectoriesForServices(serviceSel, "/home/user", "/mnt/data")
	if len(dirs) < 10 {
		t.Errorf("Expected at least 10 directories, got %d", len(dirs))
	}

	// Phase 4: Compose Config
	config := compose.DefaultConfig()
	config.AutoFillDefaults()
	if config.NextcloudPort == 0 {
		t.Error("Compose config should have ports set")
	}

	// Phase 5: Maintenance Scripts
	scriptSel := maintenance.DefaultScriptSelection()
	scripts, _ := maintenance.GetScriptsForSelection(scriptSel, maintenance.DefaultScriptConfig())
	if len(scripts) != 3 {
		t.Errorf("Default script selection should generate 3 scripts, got %d", len(scripts))
	}
}

// TestWizardFlow_TwoSimilarDisks simulates mirror/backup recommendations
func TestWizardFlow_TwoSimilarDisks(t *testing.T) {
	disks := []storage.Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: storage.DiskTypeHDD},
		{Path: "/dev/sdc", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: storage.DiskTypeHDD},
	}
	sysInfo := storage.SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := storage.GenerateStrategies(disks, sysInfo)

	// Should recommend Mirror
	if len(strategies) < 3 {
		t.Errorf("Two similar disks should generate at least 3 strategies, got %d", len(strategies))
	}

	// First should be Mirror (ZFS with 16GB RAM)
	if strategies[0].ID != storage.StrategyMirror {
		t.Errorf("First recommendation should be Mirror, got %v", strategies[0].ID)
	}
	if !strategies[0].Recommended {
		t.Error("First strategy should be marked as recommended")
	}
}

// TestWizardFlow_MixedSpeedDisks simulates NVMe + HDD setup
func TestWizardFlow_MixedSpeedDisks(t *testing.T) {
	disks := []storage.Disk{
		{Path: "/dev/nvme0n1", Size: 1 * 1024 * 1024 * 1024 * 1024, SizeHuman: "1TB", IsAvailable: true, Type: storage.DiskTypeNVMe, Transport: "nvme"},
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: storage.DiskTypeHDD, Rotational: true},
	}
	sysInfo := storage.SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := storage.GenerateStrategies(disks, sysInfo)

	// Should include speed-tiered option
	hasSpeedTiered := false
	for _, s := range strategies {
		if s.ID == storage.StrategySpeedTiered {
			hasSpeedTiered = true
		}
	}
	if !hasSpeedTiered {
		t.Error("Should recommend speed-tiered pools for NVMe + HDD")
	}
}

// TestWizardFlow_CustomSelection tests non-default service selection
func TestWizardFlow_CustomSelection(t *testing.T) {
	// User only wants Nextcloud
	serviceSel := directory.ServiceSelection{
		Nextcloud: true,
		Immich:    false,
		Databases: false,
		Glances:   false,
	}

	dirs := directory.GetDirectoriesForServices(serviceSel, "/home/user", "/mnt/data")

	// Should only have core + nextcloud directories
	for _, d := range dirs {
		if d.Service == "immich" {
			t.Error("Should not create Immich directories when not selected")
		}
		if d.Service == "databases" {
			t.Error("Should not create database directories when not selected")
		}
	}

	// Should still have Nextcloud
	hasNextcloud := false
	for _, d := range dirs {
		if d.Service == "nextcloud" {
			hasNextcloud = true
		}
	}
	if !hasNextcloud {
		t.Error("Should have Nextcloud directories when selected")
	}
}

// TestWizardFlow_StrategyConfig tests customization flow
func TestWizardFlow_StrategyConfig(t *testing.T) {
	// Default config
	config := storage.DefaultStrategyConfig()
	if config.MountPoint != "/mnt/data" {
		t.Error("Default mount point should be /mnt/data")
	}

	// Convert to map for ApplyStrategy
	configMap := config.ToConfigMap()
	if configMap["mountpoint"] != "/mnt/data" {
		t.Error("Config map should have mountpoint")
	}
	if configMap["filesystem"] != "ext4" {
		t.Error("Config map should have filesystem")
	}

	// Custom config
	customConfig := storage.StrategyConfig{
		MountPoint:     "/custom/data",
		BackupMount:    "/custom/backup",
		Filesystem:     "xfs",
		Label:          "my_data",
		BackupSchedule: "6h",
	}
	customMap := customConfig.ToConfigMap()
	if customMap["mountpoint"] != "/custom/data" {
		t.Error("Custom mount point should be preserved")
	}
	if customMap["filesystem"] != "xfs" {
		t.Error("Custom filesystem should be preserved")
	}
}

// TestWizardFlow_ScriptGeneration tests maintenance script generation
func TestWizardFlow_ScriptGeneration(t *testing.T) {
	// All scripts enabled
	allEnabled := maintenance.ScriptSelection{
		DailyBackup:   true,
		DiskAlert:     true,
		SmartAlert:    true,
		WeeklyCleanup: true,
	}
	config := maintenance.DefaultScriptConfig()
	config.DataRoot = "/mnt/data"
	config.BackupDest = "/mnt/backup"
	config.WebhookURL = "https://discord.com/api/webhooks/test"

	scripts, err := maintenance.GetScriptsForSelection(allEnabled, config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(scripts) != 4 {
		t.Errorf("Expected 4 scripts, got %d", len(scripts))
	}

	// Verify script content includes config values
	for _, s := range scripts {
		if s.Content == "" {
			t.Errorf("Script %s has empty content", s.Name)
		}
	}
}

// =============================================================================
// Edge Case Integration Tests
// =============================================================================

// TestEdgeCase_NoDisks tests the no-disk scenario
func TestEdgeCase_NoDisks(t *testing.T) {
	disks := []storage.Disk{}
	sysInfo := storage.SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := storage.GenerateStrategies(disks, sysInfo)

	if len(strategies) != 1 {
		t.Errorf("No disks should generate partition strategy, got %d", len(strategies))
	}
	if strategies[0].ID != storage.StrategyPartition {
		t.Error("No disks should recommend partition strategy")
	}
}

// TestEdgeCase_HardwareRAID tests hardware RAID detection
func TestEdgeCase_HardwareRAID(t *testing.T) {
	disks := []storage.Disk{
		{Path: "/dev/sdb", Model: "PERC H730 Virtual Disk", IsAvailable: true},
		{Path: "/dev/sdc", Model: "PERC H730 Virtual Disk", IsAvailable: true},
	}
	sysInfo := storage.SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	strategies := storage.GenerateStrategies(disks, sysInfo)

	// Should NOT suggest Mirror or MergerFS
	for _, s := range strategies {
		if s.ID == storage.StrategyMirror {
			t.Error("Should not suggest software mirror on hardware RAID")
		}
		if s.ID == storage.StrategyMergerFS {
			t.Error("Should not suggest MergerFS on hardware RAID")
		}
	}
}

// TestEdgeCase_LowRAM tests low RAM ZFS detection
func TestEdgeCase_LowRAM(t *testing.T) {
	disks := []storage.Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: storage.DiskTypeHDD},
		{Path: "/dev/sdc", Size: 4 * 1024 * 1024 * 1024 * 1024, SizeHuman: "4TB", IsAvailable: true, Type: storage.DiskTypeHDD},
	}
	sysInfo := storage.SystemInfo{TotalRAM: 4 * 1024 * 1024 * 1024} // 4GB - too low for ZFS

	strategies := storage.GenerateStrategies(disks, sysInfo)

	// Mirror should suggest MDADM, not ZFS
	for _, s := range strategies {
		if s.ID == storage.StrategyMirror {
			if containsString(s.Name, "ZFS") {
				t.Error("With 4GB RAM, should suggest MDADM not ZFS")
			}
		}
	}
}

// TestEdgeCase_AllServicesDisabled tests minimal setup
func TestEdgeCase_AllServicesDisabled(t *testing.T) {
	serviceSel := directory.ServiceSelection{
		Nextcloud: false,
		Immich:    false,
		Databases: false,
		Glances:   false,
	}

	dirs := directory.GetDirectoriesForServices(serviceSel, "/home/user", "/mnt/data")

	// Should still have core infrastructure
	hasCore := false
	for _, d := range dirs {
		if d.Service == "core" {
			hasCore = true
		}
	}
	if !hasCore {
		t.Error("Should always create core directories")
	}
}

// TestEdgeCase_AllScriptsDisabled tests no scripts
func TestEdgeCase_AllScriptsDisabled(t *testing.T) {
	scriptSel := maintenance.ScriptSelection{
		DailyBackup:   false,
		DiskAlert:     false,
		SmartAlert:    false,
		WeeklyCleanup: false,
	}

	scripts, _ := maintenance.GetScriptsForSelection(scriptSel, maintenance.DefaultScriptConfig())

	if len(scripts) != 0 {
		t.Errorf("No scripts selected should return empty, got %d", len(scripts))
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
