package storage

import (
	"strings"
	"testing"
)

// =============================================================================
// Snapshot Tests - Verify output format stability
// These tests lock down the exact output format to catch unintended changes
// =============================================================================

// TestRenderStrategyPreview_Snapshot_SingleDisk verifies single disk preview format
func TestRenderStrategyPreview_Snapshot_SingleDisk(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyPartition,
		Name: "Single Data Disk",
		Disks: []Disk{
			{Path: "/dev/sdb", SizeHuman: "4TB"},
		},
	}
	config := DefaultStrategyConfig()

	preview := RenderStrategyPreview(strategy, config)

	// Verify key structural elements present
	expectedElements := []string{
		"Single Data Disk",
		"/dev/sdb",
		"/mnt/data",
		"ext4",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(preview, elem) {
			t.Errorf("Preview missing expected element: %q\nPreview:\n%s", elem, preview)
		}
	}
}

// TestRenderStrategyPreview_Snapshot_Backup verifies backup preview includes schedule
func TestRenderStrategyPreview_Snapshot_Backup(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyBackup,
		Name: "Primary + Nightly Backup",
		Disks: []Disk{
			{Path: "/dev/sdb", SizeHuman: "4TB"},
			{Path: "/dev/sdc", SizeHuman: "4TB"},
		},
	}
	config := DefaultStrategyConfig()
	config.BackupSchedule = "daily"

	preview := RenderStrategyPreview(strategy, config)

	expectedElements := []string{
		"Primary",
		"Backup",
		"/dev/sdb",
		"/dev/sdc",
		"Daily",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(preview, elem) {
			t.Errorf("Backup preview missing: %q\nPreview:\n%s", elem, preview)
		}
	}
}

// TestRenderStrategyPreview_Snapshot_MergerFS verifies pool preview format
func TestRenderStrategyPreview_Snapshot_MergerFS(t *testing.T) {
	strategy := Strategy{
		ID:   StrategyMergerFS,
		Name: "Combined Pool (MergerFS)",
		Disks: []Disk{
			{Path: "/dev/sdb", SizeHuman: "4TB"},
			{Path: "/dev/sdc", SizeHuman: "4TB"},
			{Path: "/dev/sdd", SizeHuman: "2TB"},
		},
	}
	config := DefaultStrategyConfig()

	preview := RenderStrategyPreview(strategy, config)

	expectedElements := []string{
		"Pool",
		"Disk 1",
		"Disk 2",
		"Disk 3",
		"epmfs",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(preview, elem) {
			t.Errorf("MergerFS preview missing: %q\nPreview:\n%s", elem, preview)
		}
	}
}

// TestRenderStrategyPreview_Snapshot_NoDisks verifies empty disk handling
func TestRenderStrategyPreview_Snapshot_NoDisks(t *testing.T) {
	strategy := Strategy{
		ID:    StrategyPartition,
		Name:  "No Disks Test",
		Disks: []Disk{},
	}
	config := DefaultStrategyConfig()

	preview := RenderStrategyPreview(strategy, config)

	// Should not panic and should have strategy name
	if !strings.Contains(preview, "No Disks Test") {
		t.Errorf("Preview missing strategy name\nPreview:\n%s", preview)
	}
}

// TestFormatSchedule_Snapshot verifies all schedule formats
func TestFormatSchedule_Snapshot(t *testing.T) {
	snapshots := map[string]string{
		"daily":  "Daily at 3:00 AM",
		"6h":     "Every 6 hours",
		"12h":    "Every 12 hours",
		"weekly": "Weekly (Sunday 3 AM)",
	}

	for input, expected := range snapshots {
		result := formatSchedule(input)
		if result != expected {
			t.Errorf("formatSchedule(%q) = %q, want %q", input, result, expected)
		}
	}
}

// TestDefaultStrategyConfig_Snapshot verifies default values remain stable
func TestDefaultStrategyConfig_Snapshot(t *testing.T) {
	config := DefaultStrategyConfig()

	// These values should remain stable across versions
	snapshots := map[string]string{
		"MountPoint":     "/mnt/data",
		"BackupMount":    "/mnt/backup",
		"ScratchMount":   "/mnt/scratch",
		"FastMount":      "/mnt/fast",
		"Filesystem":     "ext4",
		"Label":          "servctl_data",
		"BackupSchedule": "daily",
		"MergerFSPolicy": "epmfs",
	}

	if config.MountPoint != snapshots["MountPoint"] {
		t.Errorf("MountPoint changed: got %q, want %q", config.MountPoint, snapshots["MountPoint"])
	}
	if config.BackupMount != snapshots["BackupMount"] {
		t.Errorf("BackupMount changed: got %q, want %q", config.BackupMount, snapshots["BackupMount"])
	}
	if config.Filesystem != snapshots["Filesystem"] {
		t.Errorf("Filesystem changed: got %q, want %q", config.Filesystem, snapshots["Filesystem"])
	}
	if config.MergerFSPolicy != snapshots["MergerFSPolicy"] {
		t.Errorf("MergerFSPolicy changed: got %q, want %q", config.MergerFSPolicy, snapshots["MergerFSPolicy"])
	}
}

// TestToConfigMap_Snapshot verifies config map keys remain stable
func TestToConfigMap_Snapshot(t *testing.T) {
	config := DefaultStrategyConfig()
	m := config.ToConfigMap()

	// These keys should remain stable for API compatibility
	requiredKeys := []string{
		"mountpoint",
		"backup_mount",
		"scratch_mount",
		"fast_mount",
		"filesystem",
		"label",
		"backup_schedule",
		"mergerfs_policy",
	}

	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("Config map missing required key: %q", key)
		}
	}
}
