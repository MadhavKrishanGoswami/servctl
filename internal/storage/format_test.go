package storage

import (
	"runtime"
	"testing"
)

func TestFilesystemTypeString(t *testing.T) {
	tests := []struct {
		fsType   FilesystemType
		expected string
	}{
		{FSTypeExt4, "ext4"},
		{FSTypeXFS, "xfs"},
		{FSTypeBtrfs, "btrfs"},
		{FSTypeZFS, "zfs"},
		{FilesystemType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.fsType.String(); got != tt.expected {
				t.Errorf("FilesystemType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetFilesystemOptions(t *testing.T) {
	options := GetFilesystemOptions()

	if len(options) != 4 {
		t.Errorf("GetFilesystemOptions() returned %d options, want %d", len(options), 4)
	}

	// Verify all filesystem types are present
	types := make(map[FilesystemType]bool)
	for _, opt := range options {
		types[opt.Type] = true
	}

	expectedTypes := []FilesystemType{FSTypeExt4, FSTypeXFS, FSTypeBtrfs, FSTypeZFS}
	for _, expected := range expectedTypes {
		if !types[expected] {
			t.Errorf("Missing filesystem type: %v", expected)
		}
	}
}

func TestGetFilesystemOptionsHasProsAndCons(t *testing.T) {
	options := GetFilesystemOptions()

	for _, opt := range options {
		if len(opt.Pros) == 0 {
			t.Errorf("Filesystem %s has no pros", opt.Name)
		}
		if len(opt.Cons) == 0 {
			t.Errorf("Filesystem %s has no cons", opt.Name)
		}
		if opt.MkfsCommand == "" {
			t.Errorf("Filesystem %s has no mkfs command", opt.Name)
		}
	}
}

func TestGetDefaultFilesystem(t *testing.T) {
	def := GetDefaultFilesystem()

	if def.Type != FSTypeExt4 {
		t.Errorf("GetDefaultFilesystem() = %v, want %v", def.Type, FSTypeExt4)
	}

	if !def.IsDefault {
		t.Error("Default filesystem should have IsDefault = true")
	}
}

func TestFilesystemOptionStructure(t *testing.T) {
	opt := FilesystemOption{
		Type:        FSTypeExt4,
		Name:        "ext4 (Recommended)",
		Description: "The most stable filesystem",
		Pros:        []string{"Stable", "Fast"},
		Cons:        []string{"No snapshots"},
		IsDefault:   true,
		MkfsCommand: "mkfs.ext4 -L %s %s",
	}

	if opt.Type != FSTypeExt4 {
		t.Errorf("FilesystemOption.Type = %v, want %v", opt.Type, FSTypeExt4)
	}
	if len(opt.Pros) != 2 {
		t.Errorf("len(FilesystemOption.Pros) = %v, want %v", len(opt.Pros), 2)
	}
}

func TestFormatResultStructure(t *testing.T) {
	result := FormatResult{
		Success:    true,
		DiskPath:   "/dev/sdb",
		Filesystem: FSTypeExt4,
		Label:      "servctl-data",
		Error:      "",
	}

	if !result.Success {
		t.Error("FormatResult.Success should be true")
	}
	if result.DiskPath != "/dev/sdb" {
		t.Errorf("FormatResult.DiskPath = %v, want %v", result.DiskPath, "/dev/sdb")
	}
}

func TestMountResultStructure(t *testing.T) {
	result := MountResult{
		Success:    true,
		DiskPath:   "/dev/sdb1",
		MountPoint: "/mnt/data",
		Error:      "",
	}

	if !result.Success {
		t.Error("MountResult.Success should be true")
	}
	if result.MountPoint != "/mnt/data" {
		t.Errorf("MountResult.MountPoint = %v, want %v", result.MountPoint, "/mnt/data")
	}
}

func TestFstabEntryStructure(t *testing.T) {
	entry := FstabEntry{
		Device:     "UUID=12345678-1234-1234-1234-123456789012",
		MountPoint: "/mnt/data",
		Filesystem: "ext4",
		Options:    "defaults,noatime",
		Dump:       0,
		Pass:       2,
	}

	if entry.MountPoint != "/mnt/data" {
		t.Errorf("FstabEntry.MountPoint = %v, want %v", entry.MountPoint, "/mnt/data")
	}
	if entry.Options != "defaults,noatime" {
		t.Errorf("FstabEntry.Options = %v, want %v", entry.Options, "defaults,noatime")
	}
}

func TestFormatDiskDryRun(t *testing.T) {
	// Test dry run for each filesystem type
	filesystems := []FilesystemType{FSTypeExt4, FSTypeXFS, FSTypeBtrfs}

	for _, fs := range filesystems {
		t.Run(fs.String(), func(t *testing.T) {
			result, err := FormatDisk("/dev/fake", fs, "test-label", true)

			if err != nil {
				t.Errorf("FormatDisk() dry run error: %v", err)
			}
			if result == nil {
				t.Fatal("FormatDisk() returned nil result")
			}
			if !result.Success {
				t.Errorf("FormatDisk() dry run should succeed")
			}
			if result.Filesystem != fs {
				t.Errorf("FormatDisk() filesystem = %v, want %v", result.Filesystem, fs)
			}
		})
	}
}

func TestFormatDiskUnsupportedFilesystem(t *testing.T) {
	result, err := FormatDisk("/dev/fake", FilesystemType(99), "test", true)

	if err == nil {
		t.Error("FormatDisk() should error for unsupported filesystem")
	}
	if result.Success {
		t.Error("FormatDisk() should not succeed for unsupported filesystem")
	}
}

func TestWipeFilesystemDryRun(t *testing.T) {
	err := WipeFilesystem("/dev/fake", true)

	if err != nil {
		t.Errorf("WipeFilesystem() dry run error: %v", err)
	}
}

func TestMountDiskDryRun(t *testing.T) {
	result, err := MountDisk("/dev/fake", "/tmp/test-mount", true)

	if err != nil {
		t.Errorf("MountDisk() dry run error: %v", err)
	}
	if result == nil {
		t.Fatal("MountDisk() returned nil result")
	}
	if !result.Success {
		t.Error("MountDisk() dry run should succeed")
	}
}

func TestAddToFstabDryRun(t *testing.T) {
	// Skip on non-Linux systems since /etc/fstab doesn't exist
	if runtime.GOOS != "linux" {
		t.Skip("Skipping fstab test on non-Linux system")
	}

	entry := FstabEntry{
		Device:     "/dev/sdb1",
		MountPoint: "/mnt/test",
		Filesystem: "ext4",
		Options:    "defaults",
		Dump:       0,
		Pass:       2,
	}

	err := AddToFstab(entry, true)

	if err != nil {
		t.Errorf("AddToFstab() dry run error: %v", err)
	}
}

func TestMountAllDryRun(t *testing.T) {
	err := MountAll(true)

	if err != nil {
		t.Errorf("MountAll() dry run error: %v", err)
	}
}

// Benchmark tests
func BenchmarkGetFilesystemOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFilesystemOptions()
	}
}

func BenchmarkGetDefaultFilesystem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDefaultFilesystem()
	}
}
