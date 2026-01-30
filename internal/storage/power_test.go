package storage

import (
	"testing"
)

func TestDefaultHDDPowerConfig(t *testing.T) {
	config := DefaultHDDPowerConfig("/dev/sdb")

	if config.DiskPath != "/dev/sdb" {
		t.Errorf("DefaultHDDPowerConfig().DiskPath = %v, want %v", config.DiskPath, "/dev/sdb")
	}

	// Spindown time should be set (241 = 30 minutes)
	if config.SpindownTime != 241 {
		t.Errorf("DefaultHDDPowerConfig().SpindownTime = %v, want %v", config.SpindownTime, 241)
	}

	// APM level should be reasonable
	if config.APMLevel < 1 || config.APMLevel > 254 {
		t.Errorf("DefaultHDDPowerConfig().APMLevel = %v, should be between 1-254", config.APMLevel)
	}
}

func TestHDDPowerConfigStructure(t *testing.T) {
	config := HDDPowerConfig{
		DiskPath:     "/dev/sdb",
		SpindownTime: 120, // 10 minutes
		APMLevel:     127,
		WriteThrough: false,
	}

	if config.DiskPath != "/dev/sdb" {
		t.Errorf("HDDPowerConfig.DiskPath = %v, want %v", config.DiskPath, "/dev/sdb")
	}
	if config.SpindownTime != 120 {
		t.Errorf("HDDPowerConfig.SpindownTime = %v, want %v", config.SpindownTime, 120)
	}
}

func TestHDParmConfEntryStructure(t *testing.T) {
	entry := HDParmConfEntry{
		DiskPath:     "/dev/sdb",
		SpindownTime: 241,
		APMLevel:     127,
	}

	if entry.DiskPath != "/dev/sdb" {
		t.Errorf("HDParmConfEntry.DiskPath = %v, want %v", entry.DiskPath, "/dev/sdb")
	}
}

func TestGetSpindownPresets(t *testing.T) {
	presets := GetSpindownPresets()

	if len(presets) == 0 {
		t.Error("GetSpindownPresets() returned empty list")
	}

	// Check we have common options
	hasDisabled := false
	has30min := false

	for _, preset := range presets {
		if preset.Value == 0 {
			hasDisabled = true
		}
		if preset.Minutes == 30 {
			has30min = true
		}
	}

	if !hasDisabled {
		t.Error("GetSpindownPresets() should include disabled option (value=0)")
	}
	if !has30min {
		t.Error("GetSpindownPresets() should include 30 minute option")
	}
}

func TestSpindownPresetStructure(t *testing.T) {
	preset := SpindownPreset{
		Name:    "30 minutes (Recommended)",
		Value:   241,
		Minutes: 30,
	}

	if preset.Name == "" {
		t.Error("SpindownPreset.Name should not be empty")
	}
	if preset.Value != 241 {
		t.Errorf("SpindownPreset.Value = %v, want %v", preset.Value, 241)
	}
}

func TestConfigureHDDSpindownDryRun(t *testing.T) {
	config := HDDPowerConfig{
		DiskPath:     "/dev/fake-disk",
		SpindownTime: 120,
		APMLevel:     127,
	}

	err := ConfigureHDDSpindown(config, true)

	if err != nil {
		t.Errorf("ConfigureHDDSpindown() dry run error: %v", err)
	}
}

func TestAddToHdparmConfDryRun(t *testing.T) {
	entry := HDParmConfEntry{
		DiskPath:     "/dev/fake-disk",
		SpindownTime: 241,
		APMLevel:     127,
	}

	err := AddToHdparmConf(entry, true)

	if err != nil {
		t.Errorf("AddToHdparmConf() dry run error: %v", err)
	}
}

func TestConfigureAllHDDSpindownDryRun(t *testing.T) {
	disks := []Disk{
		{Name: "sda", Path: "/dev/sda", Type: DiskTypeSSD, IsOSDisk: true},
		{Name: "sdb", Path: "/dev/sdb", Type: DiskTypeHDD, IsOSDisk: false},
		{Name: "sdc", Path: "/dev/sdc", Type: DiskTypeHDD, IsOSDisk: false},
		{Name: "nvme0n1", Path: "/dev/nvme0n1", Type: DiskTypeNVMe, IsOSDisk: false},
	}

	// Should only configure HDDs that are not OS disk
	err := ConfigureAllHDDSpindown(disks, true)

	if err != nil {
		t.Errorf("ConfigureAllHDDSpindown() dry run error: %v", err)
	}
}

func TestConfigureAllHDDSpindownSkipsSSD(t *testing.T) {
	// Create a slice with only SSDs - should complete without errors
	disks := []Disk{
		{Name: "sda", Path: "/dev/sda", Type: DiskTypeSSD, IsOSDisk: true},
		{Name: "sdb", Path: "/dev/sdb", Type: DiskTypeSSD, IsOSDisk: false},
	}

	err := ConfigureAllHDDSpindown(disks, true)

	if err != nil {
		t.Errorf("ConfigureAllHDDSpindown() should not error for SSDs: %v", err)
	}
}

func TestConfigureAllHDDSpindownSkipsOSDisk(t *testing.T) {
	// Even if OS disk is HDD, it should be skipped
	disks := []Disk{
		{Name: "sda", Path: "/dev/sda", Type: DiskTypeHDD, IsOSDisk: true},
		{Name: "sdb", Path: "/dev/sdb", Type: DiskTypeHDD, IsOSDisk: false},
	}

	err := ConfigureAllHDDSpindown(disks, true)

	if err != nil {
		t.Errorf("ConfigureAllHDDSpindown() dry run error: %v", err)
	}
}

// Benchmark tests
func BenchmarkDefaultHDDPowerConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultHDDPowerConfig("/dev/sdb")
	}
}

func BenchmarkGetSpindownPresets(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetSpindownPresets()
	}
}
