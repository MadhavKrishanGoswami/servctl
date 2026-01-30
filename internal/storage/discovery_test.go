package storage

import (
	"testing"
)

func TestDiskTypeString(t *testing.T) {
	tests := []struct {
		diskType DiskType
		expected string
	}{
		{DiskTypeSSD, "SSD"},
		{DiskTypeHDD, "HDD"},
		{DiskTypeNVMe, "NVMe"},
		{DiskTypeUSB, "USB"},
		{DiskTypeUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.diskType.String(); got != tt.expected {
				t.Errorf("DiskType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiskSizeString(t *testing.T) {
	tests := []struct {
		size     DiskSize
		expected string
	}{
		{DiskSizeSmall, "Small (<256GB)"},
		{DiskSizeMedium, "Medium (256GB-1TB)"},
		{DiskSizeLarge, "Large (>1TB)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.size.String(); got != tt.expected {
				t.Errorf("DiskSize.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
		{2199023255552, "2.00 TB"},
		{549755813888, "512.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := formatBytes(tt.bytes); got != tt.expected {
				t.Errorf("formatBytes(%d) = %v, want %v", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestCategorizeDiskSize(t *testing.T) {
	const GB = 1024 * 1024 * 1024

	tests := []struct {
		name     string
		bytes    uint64
		expected DiskSize
	}{
		{"128GB SSD", 128 * GB, DiskSizeSmall},
		{"256GB SSD", 256 * GB, DiskSizeMedium},
		{"512GB SSD", 512 * GB, DiskSizeMedium},
		{"1TB HDD", 1024 * GB, DiskSizeLarge},
		{"2TB HDD", 2048 * GB, DiskSizeLarge},
		{"100GB", 100 * GB, DiskSizeSmall},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := categorizeDiskSize(tt.bytes); got != tt.expected {
				t.Errorf("categorizeDiskSize(%d) = %v, want %v", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestClassifyDiskType(t *testing.T) {
	tests := []struct {
		name       string
		device     lsblkDevice
		rotational bool
		removable  bool
		expected   DiskType
	}{
		{
			name:       "NVMe drive",
			device:     lsblkDevice{Name: "nvme0n1", Tran: "nvme"},
			rotational: false,
			removable:  false,
			expected:   DiskTypeNVMe,
		},
		{
			name:       "SATA SSD",
			device:     lsblkDevice{Name: "sda", Tran: "sata"},
			rotational: false,
			removable:  false,
			expected:   DiskTypeSSD,
		},
		{
			name:       "SATA HDD",
			device:     lsblkDevice{Name: "sdb", Tran: "sata"},
			rotational: true,
			removable:  false,
			expected:   DiskTypeHDD,
		},
		{
			name:       "USB drive",
			device:     lsblkDevice{Name: "sdc", Tran: "usb"},
			rotational: false,
			removable:  true,
			expected:   DiskTypeUSB,
		},
		{
			name:       "Removable drive",
			device:     lsblkDevice{Name: "sdd", Tran: "sata"},
			rotational: false,
			removable:  true,
			expected:   DiskTypeUSB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyDiskType(tt.device, tt.rotational, tt.removable)
			if got != tt.expected {
				t.Errorf("classifyDiskType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiskStructure(t *testing.T) {
	disk := Disk{
		Name:         "sda",
		Path:         "/dev/sda",
		Size:         500107862016,
		SizeHuman:    "500.00 GB",
		Model:        "Samsung SSD 870",
		Type:         DiskTypeSSD,
		SizeCategory: DiskSizeMedium,
		Rotational:   false,
		IsOSDisk:     true,
		Partitions: []Partition{
			{Name: "sda1", Size: 536870912, MountPoint: "/boot/efi"},
			{Name: "sda2", Size: 499571040256, MountPoint: "/"},
		},
	}

	if disk.Name != "sda" {
		t.Errorf("Disk.Name = %v, want %v", disk.Name, "sda")
	}
	if disk.Type != DiskTypeSSD {
		t.Errorf("Disk.Type = %v, want %v", disk.Type, DiskTypeSSD)
	}
	if len(disk.Partitions) != 2 {
		t.Errorf("len(Disk.Partitions) = %v, want %v", len(disk.Partitions), 2)
	}
	if !disk.IsOSDisk {
		t.Error("Disk.IsOSDisk should be true")
	}
}

func TestFilterAvailableDisks(t *testing.T) {
	disks := []Disk{
		{Name: "sda", IsOSDisk: true, Removable: false},
		{Name: "sdb", IsOSDisk: false, Removable: false},
		{Name: "sdc", IsOSDisk: false, Removable: true},
		{Name: "nvme0n1", IsOSDisk: false, Removable: false},
	}

	available := FilterAvailableDisks(disks)

	if len(available) != 2 {
		t.Errorf("FilterAvailableDisks() returned %d disks, want %d", len(available), 2)
	}

	// Check that OS disk and removable are excluded
	for _, disk := range available {
		if disk.IsOSDisk {
			t.Error("FilterAvailableDisks() included OS disk")
		}
		if disk.Removable {
			t.Error("FilterAvailableDisks() included removable disk")
		}
	}
}

func TestFilterByType(t *testing.T) {
	disks := []Disk{
		{Name: "sda", Type: DiskTypeSSD},
		{Name: "sdb", Type: DiskTypeHDD},
		{Name: "sdc", Type: DiskTypeHDD},
		{Name: "nvme0n1", Type: DiskTypeNVMe},
	}

	ssds := FilterByType(disks, DiskTypeSSD)
	hdds := FilterByType(disks, DiskTypeHDD)
	nvmes := FilterByType(disks, DiskTypeNVMe)

	if len(ssds) != 1 {
		t.Errorf("FilterByType(SSD) returned %d, want %d", len(ssds), 1)
	}
	if len(hdds) != 2 {
		t.Errorf("FilterByType(HDD) returned %d, want %d", len(hdds), 2)
	}
	if len(nvmes) != 1 {
		t.Errorf("FilterByType(NVMe) returned %d, want %d", len(nvmes), 1)
	}
}

func TestGetOSDisk(t *testing.T) {
	disks := []Disk{
		{Name: "sda", IsOSDisk: false},
		{Name: "sdb", IsOSDisk: true},
		{Name: "sdc", IsOSDisk: false},
	}

	osDisk := GetOSDisk(disks)

	if osDisk == nil {
		t.Error("GetOSDisk() returned nil")
	} else if osDisk.Name != "sdb" {
		t.Errorf("GetOSDisk() = %v, want %v", osDisk.Name, "sdb")
	}
}

func TestGetOSDiskNone(t *testing.T) {
	disks := []Disk{
		{Name: "sda", IsOSDisk: false},
		{Name: "sdb", IsOSDisk: false},
	}

	osDisk := GetOSDisk(disks)

	if osDisk != nil {
		t.Error("GetOSDisk() should return nil when no OS disk")
	}
}

func TestSortDisksBySize(t *testing.T) {
	disks := []Disk{
		{Name: "small", Size: 100},
		{Name: "large", Size: 1000},
		{Name: "medium", Size: 500},
	}

	sorted := SortDisksBySize(disks)

	if sorted[0].Name != "large" {
		t.Errorf("First disk should be 'large', got %v", sorted[0].Name)
	}
	if sorted[1].Name != "medium" {
		t.Errorf("Second disk should be 'medium', got %v", sorted[1].Name)
	}
	if sorted[2].Name != "small" {
		t.Errorf("Third disk should be 'small', got %v", sorted[2].Name)
	}

	// Verify original slice is not modified
	if disks[0].Name != "small" {
		t.Error("Original slice was modified")
	}
}

// Integration test for DiscoverDisks (requires Linux with lsblk)
func TestDiscoverDisks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DiscoverDisks in short mode (requires Linux lsblk)")
	}

	// This test may fail on systems without lsblk
	disks, err := DiscoverDisks()

	if err != nil {
		t.Skipf("DiscoverDisks() failed (may not have lsblk): %v", err)
	}

	t.Logf("DiscoverDisks() found %d disks", len(disks))
	for _, disk := range disks {
		t.Logf("  - %s: %s %s", disk.Path, disk.Type.String(), disk.SizeHuman)
	}
}

// Benchmark tests
func BenchmarkFormatBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatBytes(1099511627776)
	}
}

func BenchmarkCategorizeDiskSize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		categorizeDiskSize(500107862016)
	}
}

func BenchmarkFilterAvailableDisks(b *testing.B) {
	disks := []Disk{
		{Name: "sda", IsOSDisk: true, Removable: false},
		{Name: "sdb", IsOSDisk: false, Removable: false},
		{Name: "sdc", IsOSDisk: false, Removable: true},
		{Name: "nvme0n1", IsOSDisk: false, Removable: false},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterAvailableDisks(disks)
	}
}
