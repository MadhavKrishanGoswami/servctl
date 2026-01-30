package storage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FilesystemType represents supported filesystem types
type FilesystemType int

const (
	FSTypeExt4 FilesystemType = iota
	FSTypeXFS
	FSTypeBtrfs
	FSTypeZFS
)

func (f FilesystemType) String() string {
	switch f {
	case FSTypeExt4:
		return "ext4"
	case FSTypeXFS:
		return "xfs"
	case FSTypeBtrfs:
		return "btrfs"
	case FSTypeZFS:
		return "zfs"
	default:
		return "unknown"
	}
}

// FilesystemOption represents a filesystem choice with pros/cons
type FilesystemOption struct {
	Type        FilesystemType
	Name        string
	Description string
	Pros        []string
	Cons        []string
	IsDefault   bool
	MinRAMPerTB int    // Minimum RAM in GB per TB of storage (for ZFS)
	MkfsCommand string // Command template for formatting
}

// GetFilesystemOptions returns all available filesystem options
func GetFilesystemOptions() []FilesystemOption {
	return []FilesystemOption{
		{
			Type:        FSTypeExt4,
			Name:        "ext4 (Recommended)",
			Description: "The most stable and widely-used Linux filesystem.",
			Pros: []string{
				"Best stability & compatibility",
				"Native Linux, proven for 15+ years",
				"Excellent for SSDs and HDDs",
				"Fast fsck recovery",
			},
			Cons: []string{
				"No built-in snapshots",
				"No compression",
			},
			IsDefault:   true,
			MkfsCommand: "mkfs.ext4 -L %s %s",
		},
		{
			Type:        FSTypeXFS,
			Name:        "XFS (High Performance)",
			Description: "High-performance filesystem, excellent for large files.",
			Pros: []string{
				"Better for large files (media/video)",
				"Excellent parallel I/O",
				"Great for databases",
				"Online defragmentation",
			},
			Cons: []string{
				"Cannot shrink partitions",
				"Slower for small files",
			},
			IsDefault:   false,
			MkfsCommand: "mkfs.xfs -L %s %s",
		},
		{
			Type:        FSTypeBtrfs,
			Name:        "Btrfs (Advanced Features)",
			Description: "Modern copy-on-write filesystem with advanced features.",
			Pros: []string{
				"Snapshots & rollback",
				"Transparent compression",
				"Checksums for data integrity",
				"Built-in RAID support",
			},
			Cons: []string{
				"More complex to manage",
				"Higher CPU/memory overhead",
				"RAID 5/6 still experimental",
			},
			IsDefault:   false,
			MkfsCommand: "mkfs.btrfs -L %s %s",
		},
		{
			Type:        FSTypeZFS,
			Name:        "ZFS (Enterprise Grade)",
			Description: "Enterprise-grade filesystem with maximum data integrity.",
			Pros: []string{
				"Maximum data integrity",
				"Advanced caching (ARC/L2ARC)",
				"Deduplication",
				"Best-in-class snapshots",
			},
			Cons: []string{
				"Requires more RAM (1GB per TB)",
				"Not in Linux kernel (license)",
				"Complex to configure",
			},
			IsDefault:   false,
			MinRAMPerTB: 1,
			MkfsCommand: "zpool create -m %s %s %s", // mountpoint, pool name, device
		},
	}
}

// GetDefaultFilesystem returns ext4 as the default
func GetDefaultFilesystem() FilesystemOption {
	options := GetFilesystemOptions()
	for _, opt := range options {
		if opt.IsDefault {
			return opt
		}
	}
	return options[0]
}

// FormatResult represents the result of a format operation
type FormatResult struct {
	Success    bool
	DiskPath   string
	Filesystem FilesystemType
	Label      string
	Error      string
}

// FormatDisk formats a disk with the specified filesystem
func FormatDisk(diskPath string, fsType FilesystemType, label string, dryRun bool) (*FormatResult, error) {
	result := &FormatResult{
		DiskPath:   diskPath,
		Filesystem: fsType,
		Label:      label,
	}

	// Build the command based on filesystem type
	var cmd *exec.Cmd
	switch fsType {
	case FSTypeExt4:
		if dryRun {
			fmt.Printf("[DRY RUN] Would execute: mkfs.ext4 -L %s %s\n", label, diskPath)
			result.Success = true
			return result, nil
		}
		cmd = exec.Command("sudo", "mkfs.ext4", "-L", label, diskPath)

	case FSTypeXFS:
		if dryRun {
			fmt.Printf("[DRY RUN] Would execute: mkfs.xfs -L %s %s\n", label, diskPath)
			result.Success = true
			return result, nil
		}
		cmd = exec.Command("sudo", "mkfs.xfs", "-f", "-L", label, diskPath)

	case FSTypeBtrfs:
		if dryRun {
			fmt.Printf("[DRY RUN] Would execute: mkfs.btrfs -L %s %s\n", label, diskPath)
			result.Success = true
			return result, nil
		}
		cmd = exec.Command("sudo", "mkfs.btrfs", "-L", label, diskPath)

	case FSTypeZFS:
		// ZFS is special - we create a pool instead
		if dryRun {
			fmt.Printf("[DRY RUN] Would execute: zpool create %s %s\n", label, diskPath)
			result.Success = true
			return result, nil
		}
		// First check if zfs is available
		if _, err := exec.LookPath("zpool"); err != nil {
			result.Error = "ZFS is not installed. Install with: sudo apt install zfsutils-linux"
			return result, fmt.Errorf("zfs not available")
		}
		cmd = exec.Command("sudo", "zpool", "create", "-f", label, diskPath)

	default:
		result.Error = fmt.Sprintf("Unsupported filesystem type: %s", fsType)
		return result, fmt.Errorf("unsupported filesystem")
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = string(output)
		return result, fmt.Errorf("format failed: %w", err)
	}

	result.Success = true
	return result, nil
}

// WipeFilesystem removes all filesystem signatures from a disk
func WipeFilesystem(diskPath string, dryRun bool) error {
	if dryRun {
		fmt.Printf("[DRY RUN] Would execute: wipefs -a %s\n", diskPath)
		return nil
	}

	cmd := exec.Command("sudo", "wipefs", "-a", diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wipefs failed: %s - %w", string(output), err)
	}
	return nil
}

// MountResult represents the result of a mount operation
type MountResult struct {
	Success    bool
	DiskPath   string
	MountPoint string
	Error      string
}

// MountDisk mounts a disk at the specified mount point
func MountDisk(diskPath string, mountPoint string, dryRun bool) (*MountResult, error) {
	result := &MountResult{
		DiskPath:   diskPath,
		MountPoint: mountPoint,
	}

	// Create mount point directory
	if dryRun {
		fmt.Printf("[DRY RUN] Would create directory: %s\n", mountPoint)
	} else {
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			// Try with sudo
			cmd := exec.Command("sudo", "mkdir", "-p", mountPoint)
			if err := cmd.Run(); err != nil {
				result.Error = fmt.Sprintf("Failed to create mount point: %v", err)
				return result, err
			}
		}
	}

	// Mount the disk
	if dryRun {
		fmt.Printf("[DRY RUN] Would execute: mount %s %s\n", diskPath, mountPoint)
	} else {
		cmd := exec.Command("sudo", "mount", diskPath, mountPoint)
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Error = string(output)
			return result, fmt.Errorf("mount failed: %w", err)
		}
	}

	result.Success = true
	return result, nil
}

// FstabEntry represents an entry in /etc/fstab
type FstabEntry struct {
	Device     string
	MountPoint string
	Filesystem string
	Options    string
	Dump       int
	Pass       int
}

// AddToFstab adds an entry to /etc/fstab if it doesn't already exist
func AddToFstab(entry FstabEntry, dryRun bool) error {
	fstabPath := "/etc/fstab"

	// Check if entry already exists (idempotency)
	exists, err := fstabEntryExists(fstabPath, entry.MountPoint)
	if err != nil {
		return fmt.Errorf("failed to check fstab: %w", err)
	}
	if exists {
		fmt.Printf("Entry for %s already exists in fstab, skipping.\n", entry.MountPoint)
		return nil
	}

	// Format the fstab line
	fstabLine := fmt.Sprintf("%s  %s  %s  %s  %d  %d\n",
		entry.Device,
		entry.MountPoint,
		entry.Filesystem,
		entry.Options,
		entry.Dump,
		entry.Pass,
	)

	if dryRun {
		fmt.Printf("[DRY RUN] Would add to /etc/fstab:\n  %s", fstabLine)
		return nil
	}

	// Create backup of fstab
	backupCmd := exec.Command("sudo", "cp", fstabPath, fstabPath+".bak")
	if err := backupCmd.Run(); err != nil {
		return fmt.Errorf("failed to backup fstab: %w", err)
	}

	// Append to fstab using tee (for sudo)
	cmd := exec.Command("sudo", "tee", "-a", fstabPath)
	cmd.Stdin = strings.NewReader(fstabLine)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add fstab entry: %w", err)
	}

	return nil
}

// fstabEntryExists checks if a mount point already exists in fstab
func fstabEntryExists(fstabPath string, mountPoint string) (bool, error) {
	file, err := os.Open(fstabPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == mountPoint {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// MountAll runs mount -a to mount all fstab entries
func MountAll(dryRun bool) error {
	if dryRun {
		fmt.Println("[DRY RUN] Would execute: mount -a")
		return nil
	}

	cmd := exec.Command("sudo", "mount", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount -a failed: %s - %w", string(output), err)
	}
	return nil
}

// GetDiskByUUID gets the UUID of a disk/partition
func GetDiskByUUID(diskPath string) (string, error) {
	cmd := exec.Command("sudo", "blkid", "-s", "UUID", "-o", "value", diskPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get UUID: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
