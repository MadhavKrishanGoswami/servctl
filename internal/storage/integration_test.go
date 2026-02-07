//go:build integration

package storage

import (
	"os/exec"
	"strings"
	"testing"
)

// =============================================================================
// Linux-only Integration Tests
// These tests require actual Linux tools and may need sudo
// Run with: go test ./... -tags=integration
// =============================================================================

// TestDiscoverDisks_Real tests real lsblk output parsing
func TestDiscoverDisks_Real(t *testing.T) {
	// Skip if lsblk not available
	if _, err := exec.LookPath("lsblk"); err != nil {
		t.Skip("lsblk not available, skipping integration test")
	}

	disks, err := DiscoverDisks()
	if err != nil {
		t.Fatalf("DiscoverDisks failed: %v", err)
	}

	// Should find at least the system disk
	if len(disks) == 0 {
		t.Log("No disks found - this might be expected in CI")
	}

	// Verify each disk has required fields
	for _, disk := range disks {
		if disk.Path == "" {
			t.Error("Disk has empty path")
		}
		if disk.Size == 0 {
			t.Logf("Warning: Disk %s has size 0", disk.Path)
		}
	}
}

// TestGetSystemInfo_Real tests real system info gathering
func TestGetSystemInfo_Real(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.TotalRAM == 0 {
		t.Error("TotalRAM should not be 0")
	}

	t.Logf("System RAM: %d bytes", info.TotalRAM)
}

// TestFormatDisk_DryRun tests format with dry run on real system
func TestFormatDisk_DryRun_Integration(t *testing.T) {
	// Skip if mkfs not available
	if _, err := exec.LookPath("mkfs.ext4"); err != nil {
		t.Skip("mkfs.ext4 not available")
	}

	// Use a non-existent device to test dry run
	result, err := FormatDisk("/dev/fake-test-device", FSTypeExt4, "test", true)

	if err != nil {
		// Dry run should not fail
		t.Logf("Dry run returned error (may be expected): %v", err)
	}

	if !result.Success {
		t.Logf("Dry run result: %s", result.Message)
	}
}

// TestMountDisk_DryRun tests mount with dry run
func TestMountDisk_DryRun_Integration(t *testing.T) {
	result, err := MountDisk("/dev/fake-test-device", "/tmp/fake-mount", true)

	if err != nil {
		t.Logf("Dry run returned error (may be expected): %v", err)
	}

	if !result.Success {
		t.Logf("Dry run result: %+v", result)
	}
}

// TestLsblkOutputParsing tests parsing of actual lsblk JSON output
func TestLsblkOutputParsing_Integration(t *testing.T) {
	// Run lsblk and check output
	cmd := exec.Command("lsblk", "-J", "-b", "-o", "NAME,SIZE,TYPE,FSTYPE,MOUNTPOINT,MODEL,SERIAL,TRAN,ROTA,RM")
	output, err := cmd.Output()
	if err != nil {
		t.Skipf("lsblk failed: %v", err)
	}

	t.Logf("lsblk output length: %d bytes", len(output))

	// Verify we can parse the output
	if !strings.Contains(string(output), "blockdevices") {
		t.Error("lsblk output should contain 'blockdevices'")
	}
}

// TestMdadmAvailable checks if mdadm is installed
func TestMdadmAvailable_Integration(t *testing.T) {
	_, err := exec.LookPath("mdadm")
	if err != nil {
		t.Log("mdadm not installed - RAID1 tests will use MDADM fallback")
	} else {
		t.Log("mdadm available")
	}
}

// TestMergerFSAvailable checks if mergerfs is installed
func TestMergerFSAvailable_Integration(t *testing.T) {
	_, err := exec.LookPath("mergerfs")
	if err != nil {
		t.Log("mergerfs not installed - pool tests will be limited")
	} else {
		t.Log("mergerfs available")
	}
}

// TestZFSAvailable checks if ZFS is installed
func TestZFSAvailable_Integration(t *testing.T) {
	_, err := exec.LookPath("zpool")
	if err != nil {
		t.Log("ZFS not installed - ZFS mirror tests will use MDADM fallback")
	} else {
		t.Log("ZFS available")
	}
}

// TestHdparmAvailable checks if hdparm is installed
func TestHdparmAvailable_Integration(t *testing.T) {
	_, err := exec.LookPath("hdparm")
	if err != nil {
		t.Log("hdparm not installed - HDD power management tests will be limited")
	} else {
		t.Log("hdparm available")
	}
}
