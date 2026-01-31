// Package storage provides intelligent storage orchestration for servctl.
// It handles disk discovery, classification, formatting, and mounting.
package storage

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// DiskType represents the type of storage device
type DiskType int

const (
	DiskTypeUnknown DiskType = iota
	DiskTypeSSD
	DiskTypeHDD
	DiskTypeNVMe
	DiskTypeUSB
)

func (t DiskType) String() string {
	switch t {
	case DiskTypeSSD:
		return "SSD"
	case DiskTypeHDD:
		return "HDD"
	case DiskTypeNVMe:
		return "NVMe"
	case DiskTypeUSB:
		return "USB"
	default:
		return "Unknown"
	}
}

// DiskSize represents size category
type DiskSize int

const (
	DiskSizeSmall  DiskSize = iota // < 256GB
	DiskSizeMedium                 // 256GB - 1TB
	DiskSizeLarge                  // > 1TB
)

func (s DiskSize) String() string {
	switch s {
	case DiskSizeSmall:
		return "Small (<256GB)"
	case DiskSizeMedium:
		return "Medium (256GB-1TB)"
	case DiskSizeLarge:
		return "Large (>1TB)"
	default:
		return "Unknown"
	}
}

// Partition represents a disk partition
type Partition struct {
	Name       string `json:"name"`
	Size       uint64 `json:"size"`       // Size in bytes
	SizeHuman  string `json:"size_human"` // Human readable size
	Filesystem string `json:"fstype"`
	MountPoint string `json:"mountpoint"`
	Label      string `json:"label"`
	UUID       string `json:"uuid"`
}

// Disk represents a physical disk device
type Disk struct {
	Name         string      `json:"name"`       // e.g., "sda", "nvme0n1"
	Path         string      `json:"path"`       // e.g., "/dev/sda"
	Size         uint64      `json:"size"`       // Size in bytes
	SizeHuman    string      `json:"size_human"` // Human readable size
	Model        string      `json:"model"`      // Disk model
	Serial       string      `json:"serial"`     // Serial number
	Type         DiskType    `json:"type"`       // SSD, HDD, NVMe, USB
	SizeCategory DiskSize    `json:"size_category"`
	Rotational   bool        `json:"rotational"` // True for HDD
	Removable    bool        `json:"removable"`  // True for USB/removable
	Transport    string      `json:"transport"`  // sata, nvme, usb, etc.
	Partitions   []Partition `json:"partitions"`
	IsOSDisk     bool        `json:"is_os_disk"`   // Contains root filesystem
	IsAvailable  bool        `json:"is_available"` // Available for use
	SMARTHealth  string      `json:"smart_health"` // SMART health status
}

// lsblkOutput represents the JSON output from lsblk
type lsblkOutput struct {
	BlockDevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name       string        `json:"name"`
	Size       interface{}   `json:"size"` // Can be string or int depending on lsblk version
	Type       string        `json:"type"`
	Model      interface{}   `json:"model"`  // Can be null
	Serial     interface{}   `json:"serial"` // Can be null
	Rota       interface{}   `json:"rota"`   // Can be bool or string
	RM         interface{}   `json:"rm"`     // Can be bool or string
	Tran       interface{}   `json:"tran"`   // Can be null
	Mountpoint interface{}   `json:"mountpoint"`
	Fstype     interface{}   `json:"fstype"`
	Label      interface{}   `json:"label"`
	UUID       interface{}   `json:"uuid"`
	Children   []lsblkDevice `json:"children"`
}

// Helper functions to safely extract values
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func getBoolValue(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "1" || val == "true"
	default:
		return false
	}
}

func getUint64Value(v interface{}) uint64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return uint64(val)
	case int:
		return uint64(val)
	case int64:
		return uint64(val)
	case string:
		if n, err := strconv.ParseUint(val, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

// DiscoverDisks discovers all block devices on the system
func DiscoverDisks() ([]Disk, error) {
	// Run lsblk with JSON output
	cmd := exec.Command("lsblk", "-J", "-b", "-o",
		"NAME,SIZE,TYPE,MODEL,SERIAL,ROTA,RM,TRAN,MOUNTPOINT,FSTYPE,LABEL,UUID")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk: %w", err)
	}

	var lsblk lsblkOutput
	if err := json.Unmarshal(output, &lsblk); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	var disks []Disk
	for _, device := range lsblk.BlockDevices {
		// Process disk devices and loop devices (for testing with virtual disks)
		if device.Type != "disk" && device.Type != "loop" {
			continue
		}

		// Skip small loop devices (less than 100MB) - likely system loops
		if device.Type == "loop" {
			size := getUint64Value(device.Size)
			if size < 100*1024*1024 { // 100MB minimum
				continue
			}
		}

		disk := parseLsblkDevice(device)

		// Mark loop devices appropriately
		if device.Type == "loop" {
			disk.Model = "Virtual Disk (loopback)"
			disk.Transport = "loop"
			disk.IsAvailable = true // Loop devices are always available for testing
		}

		disks = append(disks, disk)
	}

	return disks, nil
}

// parseLsblkDevice converts lsblk output to our Disk struct
func parseLsblkDevice(device lsblkDevice) Disk {
	disk := Disk{
		Name:      device.Name,
		Path:      "/dev/" + device.Name,
		Model:     strings.TrimSpace(getStringValue(device.Model)),
		Serial:    strings.TrimSpace(getStringValue(device.Serial)),
		Transport: getStringValue(device.Tran),
	}

	// Parse size
	size := getUint64Value(device.Size)
	if size > 0 {
		disk.Size = size
		disk.SizeHuman = formatBytes(size)
		disk.SizeCategory = categorizeDiskSize(size)
	}

	// Determine if rotational (HDD)
	disk.Rotational = getBoolValue(device.Rota)

	// Determine if removable
	disk.Removable = getBoolValue(device.RM)

	// Classify disk type
	disk.Type = classifyDiskType(device, disk.Rotational, disk.Removable)

	// Parse partitions
	for _, child := range device.Children {
		if child.Type == "part" {
			partition := Partition{
				Name:       child.Name,
				Filesystem: getStringValue(child.Fstype),
				MountPoint: getStringValue(child.Mountpoint),
				Label:      getStringValue(child.Label),
				UUID:       getStringValue(child.UUID),
			}
			childSize := getUint64Value(child.Size)
			if childSize > 0 {
				partition.Size = childSize
				partition.SizeHuman = formatBytes(childSize)
			}
			disk.Partitions = append(disk.Partitions, partition)

			// Check if this is the OS disk
			if getStringValue(child.Mountpoint) == "/" {
				disk.IsOSDisk = true
			}
		}
	}

	// Determine if disk is available for use
	disk.IsAvailable = !disk.IsOSDisk && !disk.Removable && len(disk.Partitions) == 0

	return disk
}

// classifyDiskType determines the type of disk
func classifyDiskType(device lsblkDevice, rotational, removable bool) DiskType {
	tran := getStringValue(device.Tran)

	// NVMe drives
	if strings.HasPrefix(device.Name, "nvme") || tran == "nvme" {
		return DiskTypeNVMe
	}

	// USB drives
	if tran == "usb" || removable {
		return DiskTypeUSB
	}

	// HDD vs SSD based on rotational
	if rotational {
		return DiskTypeHDD
	}

	return DiskTypeSSD
}

// categorizeDiskSize categorizes disk by size
func categorizeDiskSize(bytes uint64) DiskSize {
	const (
		GB = 1024 * 1024 * 1024
		TB = 1024 * GB
	)

	switch {
	case bytes < 256*GB:
		return DiskSizeSmall
	case bytes < TB:
		return DiskSizeMedium
	default:
		return DiskSizeLarge
	}
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetDiskSMARTHealth gets SMART health status for a disk
func GetDiskSMARTHealth(diskPath string) (string, error) {
	cmd := exec.Command("sudo", "smartctl", "-H", diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// smartctl might not be available or disk doesn't support SMART
		return "Unknown", nil
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "PASSED") {
		return "PASSED", nil
	} else if strings.Contains(outputStr, "FAILED") {
		return "FAILED", nil
	}

	return "Unknown", nil
}

// FilterAvailableDisks returns only disks available for use
func FilterAvailableDisks(disks []Disk) []Disk {
	var available []Disk
	for _, disk := range disks {
		if !disk.IsOSDisk && !disk.Removable {
			available = append(available, disk)
		}
	}
	return available
}

// FilterByType returns disks of a specific type
func FilterByType(disks []Disk, diskType DiskType) []Disk {
	var filtered []Disk
	for _, disk := range disks {
		if disk.Type == diskType {
			filtered = append(filtered, disk)
		}
	}
	return filtered
}

// GetOSDisk returns the disk containing the root filesystem
func GetOSDisk(disks []Disk) *Disk {
	for i, disk := range disks {
		if disk.IsOSDisk {
			return &disks[i]
		}
	}
	return nil
}

// SortDisksBySize sorts disks by size (largest first)
func SortDisksBySize(disks []Disk) []Disk {
	sorted := make([]Disk, len(disks))
	copy(sorted, disks)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Size > sorted[i].Size {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
