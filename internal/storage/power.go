package storage

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// HDDPowerConfig represents power management configuration for an HDD
type HDDPowerConfig struct {
	DiskPath     string
	SpindownTime int  // -S value for hdparm (units of 5 seconds, 0=disabled)
	APMLevel     int  // -B value (1-127=allow spindown, 128-254=no spindown, 255=disabled)
	WriteThrough bool // Enable write-through caching
}

// DefaultHDDPowerConfig returns sensible defaults for HDD power management
func DefaultHDDPowerConfig(diskPath string) HDDPowerConfig {
	return HDDPowerConfig{
		DiskPath:     diskPath,
		SpindownTime: 241, // 30 minutes (241 = 30 minutes in hdparm units)
		APMLevel:     127, // Maximum power saving while allowing spindown
		WriteThrough: false,
	}
}

// ConfigureHDDSpindown configures HDD spindown using hdparm
func ConfigureHDDSpindown(config HDDPowerConfig, dryRun bool) error {
	if dryRun {
		fmt.Printf("[DRY RUN] Would configure HDD power management for %s:\n", config.DiskPath)
		fmt.Printf("  Spindown time: %d (hdparm -S)\n", config.SpindownTime)
		fmt.Printf("  APM level: %d (hdparm -B)\n", config.APMLevel)
		return nil
	}

	// Check if hdparm is available
	if _, err := exec.LookPath("hdparm"); err != nil {
		return fmt.Errorf("hdparm not installed: %w", err)
	}

	// Set spindown time
	spindownCmd := exec.Command("sudo", "hdparm", "-S", strconv.Itoa(config.SpindownTime), config.DiskPath)
	if output, err := spindownCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set spindown: %s - %w", string(output), err)
	}

	// Set APM level
	apmCmd := exec.Command("sudo", "hdparm", "-B", strconv.Itoa(config.APMLevel), config.DiskPath)
	if output, err := apmCmd.CombinedOutput(); err != nil {
		// APM might not be supported, just warn
		fmt.Printf("Warning: APM not supported on %s: %s\n", config.DiskPath, string(output))
	}

	return nil
}

// HDParmConfEntry represents an entry in /etc/hdparm.conf
type HDParmConfEntry struct {
	DiskPath     string
	SpindownTime int
	APMLevel     int
}

// AddToHdparmConf adds persistent HDD power settings to /etc/hdparm.conf
func AddToHdparmConf(entry HDParmConfEntry, dryRun bool) error {
	hdparmConf := "/etc/hdparm.conf"

	// Check if entry already exists
	content, _ := exec.Command("cat", hdparmConf).Output()
	if strings.Contains(string(content), entry.DiskPath) {
		fmt.Printf("Entry for %s already exists in hdparm.conf, skipping.\n", entry.DiskPath)
		return nil
	}

	// Format the hdparm.conf entry
	configEntry := fmt.Sprintf(`
%s {
	spindown_time = %d
	apm = %d
}
`, entry.DiskPath, entry.SpindownTime, entry.APMLevel)

	if dryRun {
		fmt.Printf("[DRY RUN] Would add to /etc/hdparm.conf:\n%s", configEntry)
		return nil
	}

	// Append to hdparm.conf using tee
	cmd := exec.Command("sudo", "tee", "-a", hdparmConf)
	cmd.Stdin = strings.NewReader(configEntry)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add hdparm.conf entry: %w", err)
	}

	return nil
}

// ConfigureAllHDDSpindown configures spindown for all HDDs
func ConfigureAllHDDSpindown(disks []Disk, dryRun bool) error {
	for _, disk := range disks {
		// Only configure HDDs
		if disk.Type != DiskTypeHDD {
			continue
		}

		// Skip OS disk
		if disk.IsOSDisk {
			continue
		}

		config := DefaultHDDPowerConfig(disk.Path)

		// Apply runtime config
		if err := ConfigureHDDSpindown(config, dryRun); err != nil {
			fmt.Printf("Warning: Failed to configure %s: %v\n", disk.Path, err)
			continue
		}

		// Add to hdparm.conf for persistence
		entry := HDParmConfEntry{
			DiskPath:     disk.Path,
			SpindownTime: config.SpindownTime,
			APMLevel:     config.APMLevel,
		}
		if err := AddToHdparmConf(entry, dryRun); err != nil {
			fmt.Printf("Warning: Failed to persist config for %s: %v\n", disk.Path, err)
		}

		fmt.Printf("Configured power management for %s (%s)\n", disk.Path, disk.SizeHuman)
	}

	return nil
}

// SpindownTimeOptions provides common spindown time presets
type SpindownPreset struct {
	Name    string
	Value   int // hdparm -S value
	Minutes int // Equivalent minutes
}

// GetSpindownPresets returns common spindown time options
func GetSpindownPresets() []SpindownPreset {
	return []SpindownPreset{
		{Name: "5 minutes", Value: 60, Minutes: 5},
		{Name: "10 minutes", Value: 120, Minutes: 10},
		{Name: "20 minutes", Value: 240, Minutes: 20},
		{Name: "30 minutes (Recommended)", Value: 241, Minutes: 30},
		{Name: "1 hour", Value: 242, Minutes: 60},
		{Name: "2 hours", Value: 244, Minutes: 120},
		{Name: "Disabled", Value: 0, Minutes: 0},
	}
}

// GetDiskPowerStatus gets current power/spindown status of a disk
func GetDiskPowerStatus(diskPath string) (string, error) {
	cmd := exec.Command("sudo", "hdparm", "-C", diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get power status: %w", err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "standby") {
		return "standby (spun down)", nil
	} else if strings.Contains(outputStr, "active/idle") {
		return "active/idle (spinning)", nil
	}

	return "unknown", nil
}
