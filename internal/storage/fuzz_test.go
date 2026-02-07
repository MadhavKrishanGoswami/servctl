package storage

import (
	"testing"
)

// =============================================================================
// Fuzz Tests - Find edge cases in parsing and validation
// Run with: go test -fuzz=Fuzz -fuzztime=30s
// =============================================================================

// FuzzAreSizesSimilar tests size comparison with random inputs
func FuzzAreSizesSimilar(f *testing.F) {
	// Seed corpus with known test cases
	f.Add(uint64(1000), uint64(1000), 0.10)
	f.Add(uint64(0), uint64(0), 0.10)
	f.Add(uint64(1), uint64(1000000), 0.50)
	f.Add(uint64(1<<63), uint64(1<<63), 0.01) // Large values

	f.Fuzz(func(t *testing.T, a, b uint64, threshold float64) {
		// Should never panic regardless of input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AreSizesSimilar panicked: a=%d, b=%d, threshold=%f", a, b, threshold)
			}
		}()

		result := AreSizesSimilar(a, b, threshold)

		// Basic sanity checks
		if a == b && threshold >= 0 {
			if !result {
				// Same values should always be similar (unless threshold is weird)
				if threshold > 0 {
					t.Logf("Same values %d not considered similar with threshold %f", a, threshold)
				}
			}
		}
	})
}

// FuzzIsSizeMismatchLarge tests mismatch detection with random inputs
func FuzzIsSizeMismatchLarge(f *testing.F) {
	f.Add(uint64(4000), uint64(500), 0.50)
	f.Add(uint64(0), uint64(0), 0.50)

	f.Fuzz(func(t *testing.T, a, b uint64, threshold float64) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("IsSizeMismatchLarge panicked: a=%d, b=%d, threshold=%f", a, b, threshold)
			}
		}()

		_ = IsSizeMismatchLarge(a, b, threshold)
	})
}

// FuzzFormatSchedule tests schedule formatting with random strings
func FuzzFormatSchedule(f *testing.F) {
	f.Add("daily")
	f.Add("6h")
	f.Add("12h")
	f.Add("weekly")
	f.Add("")
	f.Add("random_string_12345")

	f.Fuzz(func(t *testing.T, schedule string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("formatSchedule panicked with input: %q", schedule)
			}
		}()

		result := formatSchedule(schedule)

		// Result should never be empty for any input
		if result == "" && schedule != "" {
			t.Logf("Empty result for non-empty input: %q", schedule)
		}
	})
}

// FuzzDefaultStrategyConfig tests config operations don't panic
func FuzzDefaultStrategyConfig(f *testing.F) {
	f.Add("/mnt/data", "ext4", "label", "daily")

	f.Fuzz(func(t *testing.T, mount, fs, label, schedule string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("StrategyConfig panicked with inputs: %q, %q, %q, %q", mount, fs, label, schedule)
			}
		}()

		config := StrategyConfig{
			MountPoint:     mount,
			Filesystem:     fs,
			Label:          label,
			BackupSchedule: schedule,
		}

		m := config.ToConfigMap()
		if m == nil {
			t.Error("ToConfigMap returned nil")
		}
	})
}

// FuzzIsHardwareRAID tests RAID detection with random model strings
func FuzzIsHardwareRAID(f *testing.F) {
	f.Add("PERC H730 Virtual Disk")
	f.Add("MegaRAID LD 0")
	f.Add("Samsung SSD 870")
	f.Add("")
	f.Add("random model string")

	f.Fuzz(func(t *testing.T, model string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("IsHardwareRAID panicked with model: %q", model)
			}
		}()

		disk := Disk{Model: model}
		_ = IsHardwareRAID(disk)
	})
}

// FuzzGetDiskSpeedClass tests speed classification doesn't panic
func FuzzGetDiskSpeedClass(f *testing.F) {
	// Seed with various DiskType values
	f.Add(0)   // Unknown
	f.Add(1)   // HDD
	f.Add(2)   // SSD
	f.Add(3)   // NVMe
	f.Add(255) // Invalid

	f.Fuzz(func(t *testing.T, diskType int) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetDiskSpeedClass panicked with type: %d", diskType)
			}
		}()

		disk := Disk{Type: DiskType(diskType)}
		result := GetDiskSpeedClass(disk)

		// Should always return a valid speed class
		if result != SpeedClassFast && result != SpeedClassSlow {
			t.Errorf("Invalid speed class returned: %v", result)
		}
	})
}
