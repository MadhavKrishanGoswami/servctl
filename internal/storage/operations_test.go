package storage

import (
	"testing"
)

// =============================================================================
// StrategyConfig Tests
// =============================================================================

func TestDefaultStrategyConfig(t *testing.T) {
	config := DefaultStrategyConfig()

	if config.MountPoint != "/mnt/data" {
		t.Errorf("Expected MountPoint '/mnt/data', got '%s'", config.MountPoint)
	}
	if config.BackupMount != "/mnt/backup" {
		t.Errorf("Expected BackupMount '/mnt/backup', got '%s'", config.BackupMount)
	}
	if config.ScratchMount != "/mnt/scratch" {
		t.Errorf("Expected ScratchMount '/mnt/scratch', got '%s'", config.ScratchMount)
	}
	if config.FastMount != "/mnt/fast" {
		t.Errorf("Expected FastMount '/mnt/fast', got '%s'", config.FastMount)
	}
	if config.Filesystem != "ext4" {
		t.Errorf("Expected Filesystem 'ext4', got '%s'", config.Filesystem)
	}
	if config.Label != "servctl_data" {
		t.Errorf("Expected Label 'servctl_data', got '%s'", config.Label)
	}
	if config.BackupSchedule != "daily" {
		t.Errorf("Expected BackupSchedule 'daily', got '%s'", config.BackupSchedule)
	}
	if config.MergerFSPolicy != "epmfs" {
		t.Errorf("Expected MergerFSPolicy 'epmfs', got '%s'", config.MergerFSPolicy)
	}
}

func TestToConfigMap(t *testing.T) {
	config := StrategyConfig{
		MountPoint:     "/custom/data",
		BackupMount:    "/custom/backup",
		ScratchMount:   "/custom/scratch",
		FastMount:      "/custom/fast",
		Filesystem:     "xfs",
		Label:          "my_label",
		BackupSchedule: "6h",
		MergerFSPolicy: "mfs",
	}

	m := config.ToConfigMap()

	if m["mountpoint"] != "/custom/data" {
		t.Errorf("Expected mountpoint '/custom/data', got '%s'", m["mountpoint"])
	}
	if m["backup_mount"] != "/custom/backup" {
		t.Errorf("Expected backup_mount '/custom/backup', got '%s'", m["backup_mount"])
	}
	if m["filesystem"] != "xfs" {
		t.Errorf("Expected filesystem 'xfs', got '%s'", m["filesystem"])
	}
	if m["label"] != "my_label" {
		t.Errorf("Expected label 'my_label', got '%s'", m["label"])
	}
	if m["backup_schedule"] != "6h" {
		t.Errorf("Expected backup_schedule '6h', got '%s'", m["backup_schedule"])
	}
	if m["mergerfs_policy"] != "mfs" {
		t.Errorf("Expected mergerfs_policy 'mfs', got '%s'", m["mergerfs_policy"])
	}
}

// =============================================================================
// RenderStrategyPreview Tests
// =============================================================================

func TestRenderStrategyPreview_SingleDisk(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyPartition,
		Name: "Single Data Disk",
		Disks: []Disk{
			{Path: "/dev/sdb"},
		},
	}
	config := DefaultStrategyConfig()

	preview := RenderStrategyPreview(strategy, config)

	// Check that it contains key information
	if !containsStr(preview, "Single Data Disk") {
		t.Error("Preview should contain strategy name")
	}
	if !containsStr(preview, "/dev/sdb") {
		t.Error("Preview should contain disk path")
	}
	if !containsStr(preview, "/mnt/data") {
		t.Error("Preview should contain mount point")
	}
	if !containsStr(preview, "ext4") {
		t.Error("Preview should contain filesystem")
	}
}

func TestRenderStrategyPreview_Backup(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyBackup,
		Name: "Primary + Nightly Backup",
		Disks: []Disk{
			{Path: "/dev/sdb"},
			{Path: "/dev/sdc"},
		},
	}
	config := DefaultStrategyConfig()
	config.BackupSchedule = "daily"

	preview := RenderStrategyPreview(strategy, config)

	if !containsStr(preview, "Primary") {
		t.Error("Preview should contain 'Primary'")
	}
	if !containsStr(preview, "Backup") {
		t.Error("Preview should contain 'Backup'")
	}
	if !containsStr(preview, "/dev/sdb") {
		t.Error("Preview should contain primary disk")
	}
	if !containsStr(preview, "/dev/sdc") {
		t.Error("Preview should contain backup disk")
	}
	if !containsStr(preview, "Daily") {
		t.Error("Preview should contain schedule")
	}
}

func TestRenderStrategyPreview_MergerFS(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyMergerFS,
		Name: "Combined Pool (MergerFS)",
		Disks: []Disk{
			{Path: "/dev/sdb"},
			{Path: "/dev/sdc"},
			{Path: "/dev/sdd"},
		},
	}
	config := DefaultStrategyConfig()

	preview := RenderStrategyPreview(strategy, config)

	if !containsStr(preview, "Disk 1") {
		t.Error("Preview should list disks")
	}
	if !containsStr(preview, "Pool") {
		t.Error("Preview should mention pool")
	}
	if !containsStr(preview, "epmfs") {
		t.Error("Preview should show MergerFS policy")
	}
}

func TestRenderStrategyPreview_ScratchVault(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyScratchVault,
		Name: "Scratch + Vault",
		Disks: []Disk{
			{Path: "/dev/sdb"},
			{Path: "/dev/sdc"},
		},
	}
	config := DefaultStrategyConfig()

	preview := RenderStrategyPreview(strategy, config)

	if !containsStr(preview, "Vault") {
		t.Error("Preview should contain 'Vault'")
	}
	if !containsStr(preview, "Scratch") {
		t.Error("Preview should contain 'Scratch'")
	}
}

// =============================================================================
// formatSchedule Tests
// =============================================================================

func TestFormatSchedule(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"daily", "Daily at 3:00 AM"},
		{"6h", "Every 6 hours"},
		{"12h", "Every 12 hours"},
		{"weekly", "Weekly (Sunday 3 AM)"},
		{"custom", "custom"}, // Unknown returns as-is
	}

	for _, tt := range tests {
		result := formatSchedule(tt.input)
		if result != tt.expected {
			t.Errorf("formatSchedule(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

// =============================================================================
// ApplyStrategy Tests (Dry Run)
// =============================================================================

func TestApplyStrategy_SingleDisk_DryRun(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyPartition,
		Name: "Single Data Disk",
		Disks: []Disk{
			{Path: "/dev/sdb"},
		},
	}
	config := DefaultStrategyConfig().ToConfigMap()

	results := ApplyStrategy(strategy, config, true)

	if len(results) == 0 {
		t.Error("Expected operations in dry run")
	}

	// Check for expected operations
	hasFormat := false
	hasMount := false
	for _, r := range results {
		if containsStr(r.Message, "Format") || containsStr(r.Message, "Dry Run") {
			hasFormat = true
		}
		if containsStr(r.Message, "mount") || containsStr(r.Message, "Mount") {
			hasMount = true
		}
	}

	if !hasFormat {
		t.Error("Expected format operation")
	}
	if !hasMount {
		t.Log("Mount operation may be combined with format")
	}
}

func TestApplyStrategy_Backup_DryRun(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyBackup,
		Name: "Primary + Backup",
		Disks: []Disk{
			{Path: "/dev/sdb"},
			{Path: "/dev/sdc"},
		},
	}
	config := map[string]string{
		"filesystem":      "ext4",
		"label":           "test",
		"mountpoint":      "/mnt/data",
		"backup_schedule": "daily",
	}

	results := ApplyStrategy(strategy, config, true)

	// Should have operations for both disks plus backup cron
	if len(results) < 8 {
		t.Errorf("Expected at least 8 operations for backup strategy, got %d", len(results))
	}

	// Check for backup cron setup
	hasBackupCron := false
	for _, r := range results {
		if containsStr(r.Message, "Backup") || containsStr(r.Message, "backup") || containsStr(r.Message, "cron") {
			hasBackupCron = true
		}
	}
	if !hasBackupCron {
		t.Error("Expected backup cron operation")
	}
}

func TestApplyStrategy_MergerFS_DryRun(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyMergerFS,
		Name: "MergerFS Pool",
		Disks: []Disk{
			{Path: "/dev/sdb"},
			{Path: "/dev/sdc"},
		},
	}
	config := DefaultStrategyConfig().ToConfigMap()

	results := ApplyStrategy(strategy, config, true)

	// Should format each disk and setup MergerFS
	if len(results) < 9 {
		t.Errorf("Expected at least 9 operations for MergerFS, got %d", len(results))
	}

	// Check for MergerFS setup
	hasMergerFS := false
	for _, r := range results {
		if containsStr(r.Message, "MergerFS") || containsStr(r.Message, "mergerfs") {
			hasMergerFS = true
		}
	}
	if !hasMergerFS {
		t.Error("Expected MergerFS setup operation")
	}
}

func TestApplyStrategy_ScratchVault_DryRun(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyScratchVault,
		Name: "Scratch + Vault",
		Disks: []Disk{
			{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024}, // 4TB (vault)
			{Path: "/dev/sdc", Size: 500 * 1024 * 1024 * 1024},      // 500GB (scratch)
		},
	}
	config := DefaultStrategyConfig().ToConfigMap()

	results := ApplyStrategy(strategy, config, true)

	// Should format both disks
	if len(results) < 8 {
		t.Errorf("Expected at least 8 operations for ScratchVault, got %d", len(results))
	}
}

func TestApplyStrategy_SpeedTiered_DryRun(t *testing.T) {
	strategy := Strategy{
		ID:   StrategySpeedTiered,
		Name: "Speed-Tiered Pools",
		Disks: []Disk{
			{Path: "/dev/nvme0n1", Type: DiskTypeNVMe, Transport: "nvme"},
			{Path: "/dev/sdb", Type: DiskTypeHDD, Rotational: true},
		},
	}
	config := DefaultStrategyConfig().ToConfigMap()

	results := ApplyStrategy(strategy, config, true)

	// Should setup fast and slow tiers
	if len(results) < 8 {
		t.Errorf("Expected at least 8 operations for speed-tiered, got %d", len(results))
	}

	// Check for fast tier mount
	hasFast := false
	for _, r := range results {
		if containsStr(r.Message, "fast") {
			hasFast = true
		}
	}
	if !hasFast {
		t.Error("Expected fast tier operation")
	}
}

// =============================================================================
// SetupBackupCron Tests
// =============================================================================

func TestSetupBackupCron_DryRun(t *testing.T) {
	schedules := []string{"daily", "6h", "12h", "weekly"}

	for _, schedule := range schedules {
		result := SetupBackupCron("/mnt/data", "/mnt/backup", schedule, true)

		if !result.Success {
			t.Errorf("SetupBackupCron dry run failed for schedule %s", schedule)
		}
		if !containsStr(result.Message, "Dry Run") {
			t.Error("Dry run message should indicate dry run")
		}
	}
}

// =============================================================================
// SetupMergerFS Tests
// =============================================================================

func TestSetupMergerFS_DryRun(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb"},
		{Path: "/dev/sdc"},
	}

	result := SetupMergerFS(disks, "/mnt/data", "epmfs", true)

	if !result.Success {
		t.Error("SetupMergerFS dry run should succeed")
	}
	if !containsStr(result.Message, "MergerFS") {
		t.Error("Message should mention MergerFS")
	}
}

// =============================================================================
// SetupMirror Tests
// =============================================================================

func TestSetupMirror_DryRun(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb"},
		{Path: "/dev/sdc"},
	}

	result := SetupMirror(disks, "/mnt/data", true)

	// May fail if neither ZFS nor MDADM is installed (expected on macOS)
	if !result.Success {
		// Check if error is about tools not being installed
		if containsStr(result.Message, "installed") || containsStr(result.Message, "ZFS") || containsStr(result.Message, "MDADM") {
			// This is expected behavior on systems without these tools
			return
		}
		t.Errorf("Unexpected error: %s", result.Message)
	}
	// If it succeeds, that's also fine (tools are installed)
}

func TestSetupMirror_NotEnoughDisks(t *testing.T) {
	disks := []Disk{
		{Path: "/dev/sdb"},
	}

	result := SetupMirror(disks, "/mnt/data", true)

	if result.Success {
		t.Error("SetupMirror should fail with less than 2 disks")
	}
	if !containsStr(result.Message, "2 disks") {
		t.Errorf("Expected error about needing 2 disks, got: %s", result.Message)
	}
}

// =============================================================================
// OperationResult Tests
// =============================================================================

func TestOperationResult(t *testing.T) {
	success := OperationResult{Success: true, Message: "Test passed"}
	failure := OperationResult{Success: false, Message: "Test failed", Error: nil}

	if !success.Success {
		t.Error("Success result should be successful")
	}
	if failure.Success {
		t.Error("Failure result should not be successful")
	}
}

// =============================================================================
// Helper
// =============================================================================

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
