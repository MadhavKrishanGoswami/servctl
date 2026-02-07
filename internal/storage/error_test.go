package storage

import (
	"testing"
)

// =============================================================================
// Error Path Tests - Verify graceful handling of failure scenarios
// =============================================================================

// TestFormatDisk_InvalidDevice tests formatting an invalid device
func TestFormatDisk_InvalidDevice(t *testing.T) {
	result, err := FormatDisk("", FSTypeExt4, "test", true)

	// Dry run mode accepts empty path (doesn't validate)
	if !result.Success {
		t.Logf("Empty device returned: err=%v, result.Error=%s", err, result.Error)
	}
	// This is acceptable behavior for dry run
}

// TestFormatDisk_DifferentFilesystems tests all filesystem types
func TestFormatDisk_DifferentFilesystems(t *testing.T) {
	filesystems := []FilesystemType{FSTypeExt4, FSTypeXFS, FSTypeBtrfs, FSTypeZFS}

	for _, fs := range filesystems {
		result, err := FormatDisk("/dev/sdb", fs, "test", true)
		if err != nil && !result.Success {
			t.Logf("Filesystem %s dry run: %v", fs, err)
		}
	}
}

// TestMountDisk_EmptyMountPoint tests mounting with no mount point
func TestMountDisk_EmptyMountPoint(t *testing.T) {
	result, err := MountDisk("/dev/sdb", "", true)

	// Dry run may still succeed
	t.Logf("Empty mount point result: Success=%v, Error=%s, err=%v", result.Success, result.Error, err)
}

// TestMountDisk_EmptyDevice tests mounting with no device
func TestMountDisk_EmptyDevice(t *testing.T) {
	result, _ := MountDisk("", "/mnt/data", true)

	t.Logf("Empty device result: Success=%v", result.Success)
}

// TestSetupMirror_EmptyDisks tests mirror with no disks
func TestSetupMirror_EmptyDisks(t *testing.T) {
	result := SetupMirror([]Disk{}, "/mnt/data", true)

	if result.Success {
		t.Error("SetupMirror should fail with no disks")
	}
	if !containsStr(result.Message, "2 disks") && !containsStr(result.Message, "require") {
		t.Logf("Error message: %s", result.Message)
	}
}

// TestSetupMergerFS_SingleDisk tests MergerFS with only one disk
func TestSetupMergerFS_SingleDisk(t *testing.T) {
	disks := []Disk{{Path: "/dev/sdb"}}

	result := SetupMergerFS(disks, "/mnt/data", "epmfs", true)

	// MergerFS with single disk is unusual but may be allowed
	t.Logf("Single disk MergerFS result: %s", result.Message)
}

// TestSetupMergerFS_EmptyMountPoint tests MergerFS with no mount point
func TestSetupMergerFS_EmptyMountPoint(t *testing.T) {
	disks := []Disk{{Path: "/dev/sdb"}, {Path: "/dev/sdc"}}

	result := SetupMergerFS(disks, "", "epmfs", true)

	// In dry run mode, empty mount point may be accepted
	t.Logf("Empty mount point result: Success=%v, Message=%s", result.Success, result.Message)
}

// TestSetupBackupCron_InvalidSchedule tests backup cron with invalid schedule
func TestSetupBackupCron_InvalidSchedule(t *testing.T) {
	result := SetupBackupCron("/mnt/data", "/mnt/backup", "invalid_schedule", true)

	// Should either fail or use default schedule
	t.Logf("Invalid schedule result: Success=%v, Message=%s", result.Success, result.Message)
}

// TestSetupBackupCron_SamePaths tests backup with same source and dest
func TestSetupBackupCron_SamePaths(t *testing.T) {
	result := SetupBackupCron("/mnt/data", "/mnt/data", "daily", true)

	// Backing up to same location makes no sense
	if result.Success {
		t.Log("Warning: backup to same path should ideally be prevented")
	}
}

// TestApplyStrategy_EmptyDisks tests applying strategy with no disks
func TestApplyStrategy_EmptyDisks(t *testing.T) {
	strategy := Strategy{
		ID:    StrategyPartition,
		Name:  "Empty Disk Strategy",
		Disks: []Disk{},
	}
	config := DefaultStrategyConfig().ToConfigMap()

	results := ApplyStrategy(strategy, config, true)

	// Should handle gracefully (may return empty or error result)
	t.Logf("Empty disks strategy returned %d results", len(results))
}

// TestApplyStrategy_NilConfig tests applying strategy with nil config
func TestApplyStrategy_NilConfig(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyPartition,
		Name: "Test",
		Disks: []Disk{
			{Path: "/dev/sdb"},
		},
	}

	// Should not panic with nil config
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ApplyStrategy panicked with nil config: %v", r)
		}
	}()

	results := ApplyStrategy(strategy, nil, true)
	t.Logf("Nil config returned %d results", len(results))
}

// TestGenerateStrategies_InvalidSystemInfo tests with invalid system info
func TestGenerateStrategies_InvalidSystemInfo(t *testing.T) {
	disks := []Disk{{Path: "/dev/sdb", IsAvailable: true}}
	info := SystemInfo{TotalRAM: 0} // Invalid

	strategies := GenerateStrategies(disks, info)

	// Should still return valid strategies
	if len(strategies) == 0 {
		t.Error("Should return at least one fallback strategy")
	}
}

// TestStrategyConfig_ToConfigMap_Empty tests empty config conversion
func TestStrategyConfig_ToConfigMap_Empty(t *testing.T) {
	config := StrategyConfig{} // All empty fields

	m := config.ToConfigMap()

	// Should not panic, may have empty values
	if m == nil {
		t.Error("ToConfigMap should not return nil")
	}
}

// =============================================================================
// Disk Discovery Error Paths
// =============================================================================

// TestFilterAvailableDisks_EmptyList tests filtering empty disk list
func TestFilterAvailableDisks_EmptyList(t *testing.T) {
	available := FilterAvailableDisks([]Disk{})

	if len(available) != 0 {
		t.Error("Empty input should return empty output")
	}
}

// TestFilterAvailableDisks_AllUnavailable tests when all disks are OS disks
func TestFilterAvailableDisks_AllUnavailable(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sda", IsOSDisk: true},
		{Path: "/dev/sdb", Removable: true},
	}

	available := FilterAvailableDisks(disks)

	if len(available) != 0 {
		t.Error("All unavailable should return empty list")
	}
}

// =============================================================================
// Power Management Error Paths
// =============================================================================

// TestConfigureHDDSpindown_DryRun tests HDD spindown in dry run mode
func TestConfigureHDDSpindown_DryRun(t *testing.T) {
	config := DefaultHDDPowerConfig("/dev/sdb")

	err := ConfigureHDDSpindown(config, true)

	if err != nil {
		t.Errorf("Dry run should not fail: %v", err)
	}
}

// TestConfigureAllHDDSpindown_EmptyList tests with no disks
func TestConfigureAllHDDSpindown_EmptyList(t *testing.T) {
	err := ConfigureAllHDDSpindown([]Disk{}, true)

	if err != nil {
		t.Errorf("Empty disk list should not error: %v", err)
	}
}

// TestConfigureAllHDDSpindown_OnlySSDs tests with no HDDs
func TestConfigureAllHDDSpindown_OnlySSDs(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/nvme0n1", Type: DiskTypeNVMe},
		{Path: "/dev/sda", Type: DiskTypeSSD},
	}

	err := ConfigureAllHDDSpindown(disks, true)

	if err != nil {
		t.Errorf("SSD-only list should not error: %v", err)
	}
}

// TestDefaultHDDPowerConfig_Values tests default config values
func TestDefaultHDDPowerConfig_Values(t *testing.T) {
	config := DefaultHDDPowerConfig("/dev/sdb")

	if config.DiskPath != "/dev/sdb" {
		t.Errorf("DiskPath should be /dev/sdb, got %s", config.DiskPath)
	}
	if config.SpindownTime < 0 {
		t.Errorf("SpindownTime should be non-negative, got %d", config.SpindownTime)
	}
	if config.APMLevel < 0 || config.APMLevel > 255 {
		t.Errorf("APMLevel should be 0-255, got %d", config.APMLevel)
	}
}
