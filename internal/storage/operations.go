// Package storage provides intelligent storage orchestration for servctl.
// This file implements high-level storage strategy operations.
package storage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OperationResult represents the result of a storage operation
type OperationResult struct {
	Success bool
	Message string
	Error   error
}

// PromptEraseConfirmation asks user to type ERASE to confirm destructive operation
func PromptEraseConfirmation(reader *bufio.Reader, disk Disk) bool {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│  ⚠️  WARNING: DESTRUCTIVE OPERATION                         │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Printf("  You are about to FORMAT: %s (%s)\n", disk.Path, disk.SizeHuman)
	if disk.Model != "" {
		fmt.Printf("  Model: %s\n", disk.Model)
	}
	fmt.Println()
	fmt.Println("  This will PERMANENTLY ERASE all data on this disk!")
	fmt.Println()
	fmt.Print("  Type 'ERASE' to confirm, or press Enter to cancel: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	return strings.TrimSpace(response) == "ERASE"
}

// PromptStrategySelection prompts user to select a storage strategy
func PromptStrategySelection(reader *bufio.Reader, strategies []Strategy) (Strategy, bool) {
	if len(strategies) == 0 {
		return Strategy{}, false
	}

	fmt.Print("Select strategy [1-" + fmt.Sprintf("%d", len(strategies)) + ", or 's' to skip]: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return Strategy{}, false
	}

	response = strings.TrimSpace(strings.ToLower(response))

	if response == "s" || response == "skip" {
		return Strategy{}, false
	}

	var choice int
	if _, err := fmt.Sscanf(response, "%d", &choice); err != nil {
		fmt.Println("  Invalid selection.")
		return Strategy{}, false
	}

	if choice < 1 || choice > len(strategies) {
		fmt.Println("  Invalid selection.")
		return Strategy{}, false
	}

	return strategies[choice-1], true
}

// StrategyConfig holds customizable options for a strategy
type StrategyConfig struct {
	MountPoint     string
	BackupMount    string
	ScratchMount   string
	FastMount      string
	Filesystem     string
	Label          string
	BackupSchedule string
	MergerFSPolicy string
}

// DefaultStrategyConfig returns sensible defaults
func DefaultStrategyConfig() StrategyConfig {
	return StrategyConfig{
		MountPoint:     "/mnt/data",
		BackupMount:    "/mnt/backup",
		ScratchMount:   "/mnt/scratch",
		FastMount:      "/mnt/fast",
		Filesystem:     "ext4",
		Label:          "servctl_data",
		BackupSchedule: "daily",
		MergerFSPolicy: "epmfs",
	}
}

// RenderStrategyPreview shows the strategy config before applying
func RenderStrategyPreview(strategy Strategy, config StrategyConfig) string {
	var b strings.Builder

	b.WriteString("┌─────────────────────────────────────────┐\n")
	b.WriteString(fmt.Sprintf("│  %s\n", strategy.Name))
	b.WriteString("└─────────────────────────────────────────┘\n\n")

	// Show disk assignments based on strategy
	switch strategy.ID {
	case StrategyBackup:
		if len(strategy.Disks) >= 2 {
			b.WriteString(fmt.Sprintf("  Primary: %s → %s\n", strategy.Disks[0].Path, config.MountPoint))
			b.WriteString(fmt.Sprintf("  Backup:  %s → %s\n", strategy.Disks[1].Path, config.BackupMount))
			b.WriteString(fmt.Sprintf("  Schedule: %s\n", formatSchedule(config.BackupSchedule)))
		}
	case StrategyScratchVault:
		if len(strategy.Disks) >= 2 {
			b.WriteString(fmt.Sprintf("  Vault:   %s → %s\n", strategy.Disks[0].Path, config.MountPoint))
			b.WriteString(fmt.Sprintf("  Scratch: %s → %s\n", strategy.Disks[1].Path, config.ScratchMount))
		}
	case StrategyMergerFS:
		for i, d := range strategy.Disks {
			b.WriteString(fmt.Sprintf("  Disk %d:  %s → /mnt/disk%d\n", i+1, d.Path, i+1))
		}
		b.WriteString(fmt.Sprintf("  Pool:    → %s\n", config.MountPoint))
		b.WriteString(fmt.Sprintf("  Policy:  %s\n", config.MergerFSPolicy))
	default:
		if len(strategy.Disks) > 0 {
			b.WriteString(fmt.Sprintf("  Disk: %s → %s\n", strategy.Disks[0].Path, config.MountPoint))
		} else {
			b.WriteString(fmt.Sprintf("  Mount: %s\n", config.MountPoint))
		}
	}

	b.WriteString(fmt.Sprintf("\n  Filesystem: %s\n", config.Filesystem))
	b.WriteString(fmt.Sprintf("  Label: %s\n", config.Label))

	return b.String()
}

func formatSchedule(schedule string) string {
	switch schedule {
	case "daily":
		return "Daily at 3:00 AM"
	case "6h":
		return "Every 6 hours"
	case "12h":
		return "Every 12 hours"
	case "weekly":
		return "Weekly (Sunday 3 AM)"
	default:
		return schedule
	}
}

// PromptStrategyConfirmation shows preview and offers customization
func PromptStrategyConfirmation(reader *bufio.Reader, strategy Strategy) (StrategyConfig, bool) {
	config := DefaultStrategyConfig()

	fmt.Println()
	fmt.Print(RenderStrategyPreview(strategy, config))
	fmt.Println()
	fmt.Print("Press [Enter] to apply, [C] to customize, or [S] to skip: ")

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "c":
		return PromptStrategyCustomization(reader, strategy, config), true
	case "s":
		return config, false
	default:
		return config, true
	}
}

// PromptStrategyCustomization prompts user to customize strategy options
func PromptStrategyCustomization(reader *bufio.Reader, strategy Strategy, config StrategyConfig) StrategyConfig {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Println("│         Customize Settings              │")
	fmt.Println("└─────────────────────────────────────────┘")
	fmt.Println()

	// Mount point
	fmt.Printf("  1. Data mount point [%s]: ", config.MountPoint)
	if mp := readLine(reader); mp != "" {
		config.MountPoint = mp
	}

	// Strategy-specific options
	switch strategy.ID {
	case StrategyBackup:
		fmt.Printf("  2. Backup mount point [%s]: ", config.BackupMount)
		if mp := readLine(reader); mp != "" {
			config.BackupMount = mp
		}

		fmt.Println("  3. Backup schedule:")
		fmt.Println("     [1] Daily at 3 AM")
		fmt.Println("     [2] Every 6 hours")
		fmt.Println("     [3] Every 12 hours")
		fmt.Println("     [4] Weekly (Sunday)")
		fmt.Print("     Select [1-4]: ")
		switch readLine(reader) {
		case "2":
			config.BackupSchedule = "6h"
		case "3":
			config.BackupSchedule = "12h"
		case "4":
			config.BackupSchedule = "weekly"
		default:
			config.BackupSchedule = "daily"
		}

	case StrategyScratchVault:
		fmt.Printf("  2. Scratch mount point [%s]: ", config.ScratchMount)
		if mp := readLine(reader); mp != "" {
			config.ScratchMount = mp
		}

	case StrategyMergerFS:
		fmt.Println("  2. MergerFS policy:")
		fmt.Println("     [1] epmfs - Existing path, most free space (default)")
		fmt.Println("     [2] mfs   - Most free space")
		fmt.Println("     [3] lfs   - Least free space")
		fmt.Print("     Select [1-3]: ")
		switch readLine(reader) {
		case "2":
			config.MergerFSPolicy = "mfs"
		case "3":
			config.MergerFSPolicy = "lfs"
		default:
			config.MergerFSPolicy = "epmfs"
		}
	}

	// Filesystem
	fmt.Println()
	fmt.Println("  Filesystem:")
	fmt.Println("     [1] ext4 (default, most compatible)")
	fmt.Println("     [2] XFS (better for large files)")
	fmt.Print("  Select [1-2]: ")
	if readLine(reader) == "2" {
		config.Filesystem = "xfs"
	}

	// Label
	fmt.Printf("  Label [%s]: ", config.Label)
	if l := readLine(reader); l != "" {
		config.Label = l
	}

	fmt.Println()
	fmt.Print(RenderStrategyPreview(strategy, config))

	return config
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// ToConfigMap converts StrategyConfig to map[string]string for ApplyStrategy
func (c StrategyConfig) ToConfigMap() map[string]string {
	return map[string]string{
		"mountpoint":      c.MountPoint,
		"backup_mount":    c.BackupMount,
		"scratch_mount":   c.ScratchMount,
		"fast_mount":      c.FastMount,
		"filesystem":      c.Filesystem,
		"label":           c.Label,
		"backup_schedule": c.BackupSchedule,
		"mergerfs_policy": c.MergerFSPolicy,
	}
}

// ApplyStrategy applies the selected storage strategy
func ApplyStrategy(strategy Strategy, config map[string]string, dryRun bool) []OperationResult {
	var results []OperationResult

	fsType := FSTypeExt4
	if fs, ok := config["filesystem"]; ok && fs == "xfs" {
		fsType = FSTypeXFS
	}

	label := "servctl_data"
	if l, ok := config["label"]; ok {
		label = l
	}

	mountPoint := "/mnt/data"
	if mp, ok := config["mountpoint"]; ok {
		mountPoint = mp
	}

	switch strategy.ID {
	case StrategyPartition:
		// Single disk - simple format and mount
		if len(strategy.Disks) > 0 {
			disk := strategy.Disks[0]
			results = append(results, formatDiskWrapper(disk.Path, fsType, label, dryRun))
			results = append(results, createMountPointWrapper(mountPoint, dryRun))
			results = append(results, mountDiskWrapper(disk.Path, mountPoint, dryRun))
			results = append(results, addToFstabWrapper(disk.Path, mountPoint, fsType.String(), dryRun))
		}

	case StrategyMergerFS:
		// Format each disk individually, then setup MergerFS
		for i, disk := range strategy.Disks {
			diskLabel := fmt.Sprintf("%s_%d", label, i+1)
			diskMount := filepath.Join("/mnt", fmt.Sprintf("disk%d", i+1))
			results = append(results, formatDiskWrapper(disk.Path, fsType, diskLabel, dryRun))
			results = append(results, createMountPointWrapper(diskMount, dryRun))
			results = append(results, mountDiskWrapper(disk.Path, diskMount, dryRun))
			results = append(results, addToFstabWrapper(disk.Path, diskMount, fsType.String(), dryRun))
		}
		results = append(results, createMountPointWrapper(mountPoint, dryRun))
		results = append(results, SetupMergerFS(strategy.Disks, mountPoint, "epmfs", dryRun))

	case StrategyMirror:
		results = append(results, SetupMirror(strategy.Disks, mountPoint, dryRun))

	case StrategyBackup:
		if len(strategy.Disks) >= 2 {
			primary := strategy.Disks[0]
			backup := strategy.Disks[1]

			// Primary disk
			results = append(results, formatDiskWrapper(primary.Path, fsType, label, dryRun))
			results = append(results, createMountPointWrapper(mountPoint, dryRun))
			results = append(results, mountDiskWrapper(primary.Path, mountPoint, dryRun))
			results = append(results, addToFstabWrapper(primary.Path, mountPoint, fsType.String(), dryRun))

			// Backup disk
			backupMount := "/mnt/backup"
			results = append(results, formatDiskWrapper(backup.Path, fsType, label+"_backup", dryRun))
			results = append(results, createMountPointWrapper(backupMount, dryRun))
			results = append(results, mountDiskWrapper(backup.Path, backupMount, dryRun))
			results = append(results, addToFstabWrapper(backup.Path, backupMount, fsType.String(), dryRun))

			// Setup backup cron
			schedule := "daily"
			if s, ok := config["backup_schedule"]; ok {
				schedule = s
			}
			results = append(results, SetupBackupCron(mountPoint, backupMount, schedule, dryRun))
		}

	case StrategyScratchVault:
		if len(strategy.Disks) >= 2 {
			large := strategy.Disks[0]
			small := strategy.Disks[1]
			if small.Size > large.Size {
				large, small = small, large
			}

			// Vault (large disk)
			results = append(results, formatDiskWrapper(large.Path, fsType, "vault", dryRun))
			results = append(results, createMountPointWrapper(mountPoint, dryRun))
			results = append(results, mountDiskWrapper(large.Path, mountPoint, dryRun))
			results = append(results, addToFstabWrapper(large.Path, mountPoint, fsType.String(), dryRun))

			// Scratch (small disk)
			scratchMount := "/mnt/scratch"
			results = append(results, formatDiskWrapper(small.Path, fsType, "scratch", dryRun))
			results = append(results, createMountPointWrapper(scratchMount, dryRun))
			results = append(results, mountDiskWrapper(small.Path, scratchMount, dryRun))
			results = append(results, addToFstabWrapper(small.Path, scratchMount, fsType.String(), dryRun))
		}

	case StrategySpeedTiered:
		var fastDisks, slowDisks []Disk
		for _, d := range strategy.Disks {
			if GetDiskSpeedClass(d) == SpeedClassFast {
				fastDisks = append(fastDisks, d)
			} else {
				slowDisks = append(slowDisks, d)
			}
		}

		// Fast tier
		for i, disk := range fastDisks {
			diskLabel := fmt.Sprintf("fast_%d", i+1)
			diskMount := fmt.Sprintf("/mnt/fast%d", i+1)
			results = append(results, formatDiskWrapper(disk.Path, fsType, diskLabel, dryRun))
			results = append(results, createMountPointWrapper(diskMount, dryRun))
			results = append(results, mountDiskWrapper(disk.Path, diskMount, dryRun))
			results = append(results, addToFstabWrapper(disk.Path, diskMount, fsType.String(), dryRun))
		}
		results = append(results, createMountPointWrapper("/mnt/fast", dryRun))

		// Slow tier
		for i, disk := range slowDisks {
			diskLabel := fmt.Sprintf("data_%d", i+1)
			diskMount := fmt.Sprintf("/mnt/slow%d", i+1)
			results = append(results, formatDiskWrapper(disk.Path, fsType, diskLabel, dryRun))
			results = append(results, createMountPointWrapper(diskMount, dryRun))
			results = append(results, mountDiskWrapper(disk.Path, diskMount, dryRun))
			results = append(results, addToFstabWrapper(disk.Path, diskMount, fsType.String(), dryRun))
		}
		results = append(results, createMountPointWrapper(mountPoint, dryRun))
	}

	return results
}

// Wrapper functions to adapt format.go functions to OperationResult

func formatDiskWrapper(diskPath string, fsType FilesystemType, label string, dryRun bool) OperationResult {
	result, err := FormatDisk(diskPath, fsType, label, dryRun)
	if err != nil {
		return OperationResult{Success: false, Message: err.Error(), Error: err}
	}
	return OperationResult{Success: result.Success, Message: fmt.Sprintf("Formatted %s as %s", diskPath, fsType)}
}

func createMountPointWrapper(mountPoint string, dryRun bool) OperationResult {
	if dryRun {
		return OperationResult{Success: true, Message: fmt.Sprintf("[Dry Run] Would create: %s", mountPoint)}
	}
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return OperationResult{Success: false, Message: err.Error(), Error: err}
	}
	return OperationResult{Success: true, Message: fmt.Sprintf("Created: %s", mountPoint)}
}

func mountDiskWrapper(diskPath, mountPoint string, dryRun bool) OperationResult {
	result, err := MountDisk(diskPath, mountPoint, dryRun)
	if err != nil {
		return OperationResult{Success: false, Message: err.Error(), Error: err}
	}
	return OperationResult{Success: result.Success, Message: fmt.Sprintf("Mounted %s at %s", diskPath, mountPoint)}
}

func addToFstabWrapper(diskPath, mountPoint, filesystem string, dryRun bool) OperationResult {
	entry := FstabEntry{
		Device:     diskPath,
		MountPoint: mountPoint,
		Filesystem: filesystem,
		Options:    "defaults,noatime",
		Dump:       0,
		Pass:       2,
	}
	if err := AddToFstab(entry, dryRun); err != nil {
		return OperationResult{Success: false, Message: err.Error(), Error: err}
	}
	return OperationResult{Success: true, Message: fmt.Sprintf("Added %s to /etc/fstab", mountPoint)}
}

// SetupMergerFS configures MergerFS to combine multiple disks
func SetupMergerFS(disks []Disk, mountPoint, policy string, dryRun bool) OperationResult {
	result := OperationResult{Success: false}

	var sources []string
	for i := range disks {
		sources = append(sources, fmt.Sprintf("/mnt/disk%d", i+1))
	}
	sourcePath := strings.Join(sources, ":")

	fstabLine := fmt.Sprintf("%s %s fuse.mergerfs defaults,allow_other,use_ino,cache.files=partial,dropcacheonclose=true,category.create=%s 0 0\n",
		sourcePath, mountPoint, policy)

	if dryRun {
		result.Success = true
		result.Message = fmt.Sprintf("[Dry Run] Would add MergerFS entry:\n  %s", fstabLine)
		return result
	}

	if _, err := exec.LookPath("mergerfs"); err != nil {
		result.Error = fmt.Errorf("mergerfs not installed. Run: sudo apt install mergerfs")
		result.Message = result.Error.Error()
		return result
	}

	f, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		result.Error = err
		result.Message = err.Error()
		return result
	}
	defer f.Close()

	if _, err := f.WriteString(fstabLine); err != nil {
		result.Error = err
		result.Message = err.Error()
		return result
	}

	cmd := exec.Command("mount", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Error = fmt.Errorf("mount failed: %w - %s", err, string(output))
		result.Message = result.Error.Error()
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("MergerFS: %s → %s", sourcePath, mountPoint)
	return result
}

// SetupMirror configures a ZFS or MDADM mirror
func SetupMirror(disks []Disk, mountPoint string, dryRun bool) OperationResult {
	result := OperationResult{Success: false}

	if len(disks) < 2 {
		result.Error = fmt.Errorf("mirror requires at least 2 disks")
		result.Message = result.Error.Error()
		return result
	}

	// Check for ZFS
	if _, err := exec.LookPath("zpool"); err == nil {
		return setupZFSMirror(disks, mountPoint, dryRun)
	}

	// Fall back to MDADM
	if _, err := exec.LookPath("mdadm"); err == nil {
		return setupMDADMMirror(disks, mountPoint, dryRun)
	}

	result.Error = fmt.Errorf("neither ZFS nor MDADM installed")
	result.Message = result.Error.Error()
	return result
}

func setupZFSMirror(disks []Disk, mountPoint string, dryRun bool) OperationResult {
	result := OperationResult{Success: false}

	var diskPaths []string
	for _, d := range disks {
		diskPaths = append(diskPaths, d.Path)
	}

	args := append([]string{"create", "-f", "-m", mountPoint, "servctl_pool", "mirror"}, diskPaths...)

	if dryRun {
		result.Success = true
		result.Message = fmt.Sprintf("[Dry Run] Would run: zpool %s", strings.Join(args, " "))
		return result
	}

	cmd := exec.Command("zpool", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("zpool failed: %w - %s", err, string(output))
		result.Message = result.Error.Error()
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("ZFS mirror: %s → %s", strings.Join(diskPaths, "+"), mountPoint)
	return result
}

func setupMDADMMirror(disks []Disk, mountPoint string, dryRun bool) OperationResult {
	result := OperationResult{Success: false}

	var diskPaths []string
	for _, d := range disks {
		diskPaths = append(diskPaths, d.Path)
	}

	mdDevice := "/dev/md0"
	args := append([]string{"--create", mdDevice, "--level=1", fmt.Sprintf("--raid-devices=%d", len(disks))}, diskPaths...)

	if dryRun {
		result.Success = true
		result.Message = fmt.Sprintf("[Dry Run] Would run: mdadm %s", strings.Join(args, " "))
		return result
	}

	cmd := exec.Command("mdadm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("mdadm failed: %w - %s", err, string(output))
		result.Message = result.Error.Error()
		return result
	}

	// Format and mount the array
	if _, err := FormatDisk(mdDevice, FSTypeExt4, "servctl_data", false); err != nil {
		result.Error = err
		result.Message = err.Error()
		return result
	}

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		result.Error = err
		result.Message = err.Error()
		return result
	}

	mountCmd := exec.Command("mount", mdDevice, mountPoint)
	if output, err := mountCmd.CombinedOutput(); err != nil {
		result.Error = fmt.Errorf("mount failed: %w - %s", err, string(output))
		result.Message = result.Error.Error()
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("MDADM RAID1: %s → %s", strings.Join(diskPaths, "+"), mountPoint)
	return result
}

// SetupBackupCron creates a cron job for automated backups
func SetupBackupCron(source, dest, schedule string, dryRun bool) OperationResult {
	result := OperationResult{Success: false}

	var cronSchedule string
	switch schedule {
	case "daily":
		cronSchedule = "0 3 * * *"
	case "6h":
		cronSchedule = "0 */6 * * *"
	case "12h":
		cronSchedule = "0 */12 * * *"
	case "weekly":
		cronSchedule = "0 3 * * 0"
	default:
		cronSchedule = "0 3 * * *"
	}

	scriptPath := "/usr/local/bin/servctl-backup.sh"
	scriptContent := fmt.Sprintf(`#!/bin/bash
# servctl automated backup
rsync -av --delete %s/ %s/
echo "$(date): Backup completed" >> /var/log/servctl-backup.log
`, source, dest)

	if dryRun {
		result.Success = true
		result.Message = fmt.Sprintf("[Dry Run] Would create backup cron (%s)", schedule)
		return result
	}

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		result.Error = err
		result.Message = err.Error()
		return result
	}

	cronLine := fmt.Sprintf("%s root %s\n", cronSchedule, scriptPath)
	if err := os.WriteFile("/etc/cron.d/servctl-backup", []byte(cronLine), 0644); err != nil {
		result.Error = err
		result.Message = err.Error()
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("Backup cron: %s → %s (%s)", source, dest, schedule)
	return result
}
