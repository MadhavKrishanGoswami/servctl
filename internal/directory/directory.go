// Package directory handles creation of the servctl directory structure.
// This includes user-space directories (~/infra/) and data-space directories (/mnt/data/).
package directory

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

// DirectoryType represents the category of directory
type DirectoryType int

const (
	DirTypeUserSpace DirectoryType = iota // ~/infra/
	DirTypeDataSpace                      // /mnt/data/
)

func (d DirectoryType) String() string {
	switch d {
	case DirTypeUserSpace:
		return "User Space"
	case DirTypeDataSpace:
		return "Data Space"
	default:
		return "Unknown"
	}
}

// DirectorySpec defines a directory to be created
type DirectorySpec struct {
	Path        string        // Absolute path
	Type        DirectoryType // User or Data space
	Service     string        // Which service owns this (e.g., "immich", "nextcloud")
	Description string        // Human-readable description
	Mode        os.FileMode   // Permissions (e.g., 0755)
}

// DirectoryResult represents the outcome of creating a directory
type DirectoryResult struct {
	Spec    DirectorySpec
	Created bool  // True if newly created, false if already exists
	Error   error // Any error that occurred
}

// PermissionInfo holds user/group IDs for ownership
type PermissionInfo struct {
	UID      int    // User ID
	GID      int    // Group ID
	Username string // Username
	HomeDir  string // User's home directory
}

// GetCurrentUserInfo returns the current user's UID, GID, and home directory
func GetCurrentUserInfo() (*PermissionInfo, error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UID: %w", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GID: %w", err)
	}

	return &PermissionInfo{
		UID:      uid,
		GID:      gid,
		Username: u.Username,
		HomeDir:  u.HomeDir,
	}, nil
}

// GetUserSpaceDirectories returns the list of user-space directories to create
func GetUserSpaceDirectories(homeDir string) []DirectorySpec {
	infraRoot := filepath.Join(homeDir, "infra")

	return []DirectorySpec{
		{
			Path:        infraRoot,
			Type:        DirTypeUserSpace,
			Service:     "core",
			Description: "Root directory for servctl infrastructure",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(infraRoot, "scripts"),
			Type:        DirTypeUserSpace,
			Service:     "core",
			Description: "Maintenance and backup scripts",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(infraRoot, "logs"),
			Type:        DirTypeUserSpace,
			Service:     "core",
			Description: "Centralized logging directory",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(infraRoot, "compose"),
			Type:        DirTypeUserSpace,
			Service:     "docker",
			Description: "Docker Compose files",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(infraRoot, "config"),
			Type:        DirTypeUserSpace,
			Service:     "core",
			Description: "Service configuration files",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(infraRoot, "backups"),
			Type:        DirTypeUserSpace,
			Service:     "backup",
			Description: "Local backup staging area",
			Mode:        0755,
		},
	}
}

// GetDataSpaceDirectories returns the list of data-space directories to create
func GetDataSpaceDirectories(dataRoot string) []DirectorySpec {
	if dataRoot == "" {
		dataRoot = "/mnt/data"
	}

	return []DirectorySpec{
		// Root data directory
		{
			Path:        dataRoot,
			Type:        DirTypeDataSpace,
			Service:     "core",
			Description: "Root data directory for all services",
			Mode:        0755,
		},

		// Immich (Photo Gallery) directories
		{
			Path:        filepath.Join(dataRoot, "gallery"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich photo gallery root",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "gallery", "library"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich photo library storage",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "gallery", "upload"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich upload staging area",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "gallery", "profile"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich user profiles",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "gallery", "video"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich video transcodes",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "gallery", "thumbs"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich thumbnail cache",
			Mode:        0755,
		},

		// Nextcloud directories
		{
			Path:        filepath.Join(dataRoot, "cloud"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud root directory",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "cloud", "data"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud user data storage",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "cloud", "config"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud configuration",
			Mode:        0755,
		},

		// Database directories (isolated per service)
		{
			Path:        filepath.Join(dataRoot, "databases"),
			Type:        DirTypeDataSpace,
			Service:     "database",
			Description: "Database storage root",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "databases", "immich-postgres"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich PostgreSQL data",
			Mode:        0755,
		},
		{
			Path:        filepath.Join(dataRoot, "databases", "nextcloud-mariadb"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud MariaDB data",
			Mode:        0755,
		},

		// Redis/Cache
		{
			Path:        filepath.Join(dataRoot, "cache"),
			Type:        DirTypeDataSpace,
			Service:     "redis",
			Description: "Redis/Valkey cache storage",
			Mode:        0755,
		},
	}
}

// GetAllDirectories returns all directories to create
func GetAllDirectories(homeDir, dataRoot string) []DirectorySpec {
	all := make([]DirectorySpec, 0)
	all = append(all, GetUserSpaceDirectories(homeDir)...)
	all = append(all, GetDataSpaceDirectories(dataRoot)...)
	return all
}

// CreateDirectory creates a single directory with the specified permissions
func CreateDirectory(spec DirectorySpec, dryRun bool) DirectoryResult {
	result := DirectoryResult{Spec: spec}

	// Check if directory already exists
	if info, err := os.Stat(spec.Path); err == nil {
		if info.IsDir() {
			// Already exists
			result.Created = false
			return result
		}
		// Exists but is not a directory
		result.Error = fmt.Errorf("path exists but is not a directory: %s", spec.Path)
		return result
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would create directory: %s (%s)\n", spec.Path, spec.Description)
		result.Created = true
		return result
	}

	// Create the directory with all parents
	if err := os.MkdirAll(spec.Path, spec.Mode); err != nil {
		result.Error = fmt.Errorf("failed to create directory %s: %w", spec.Path, err)
		return result
	}

	result.Created = true
	return result
}

// CreateAllDirectories creates all directories for servctl
func CreateAllDirectories(homeDir, dataRoot string, dryRun bool) []DirectoryResult {
	specs := GetAllDirectories(homeDir, dataRoot)
	results := make([]DirectoryResult, 0, len(specs))

	for _, spec := range specs {
		result := CreateDirectory(spec, dryRun)
		results = append(results, result)
	}

	return results
}

// SetPermissions sets ownership and permissions on the data directory
func SetPermissions(dataRoot string, perm *PermissionInfo, dryRun bool) error {
	if dataRoot == "" {
		dataRoot = "/mnt/data"
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would set ownership of %s to %d:%d (%s)\n",
			dataRoot, perm.UID, perm.GID, perm.Username)
		return nil
	}

	// Use chown command for recursive ownership change
	// We use the command because os.Chown doesn't do recursive
	cmd := exec.Command("chown", "-R",
		fmt.Sprintf("%d:%d", perm.UID, perm.GID),
		dataRoot)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set ownership: %s: %w", string(output), err)
	}

	fmt.Printf("Set ownership of %s to %d:%d (%s)\n",
		dataRoot, perm.UID, perm.GID, perm.Username)
	return nil
}

// SetDirectoryPermissions sets the correct file mode on directories
func SetDirectoryPermissions(dataRoot string, dryRun bool) error {
	if dataRoot == "" {
		dataRoot = "/mnt/data"
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would set directory permissions (755) on %s\n", dataRoot)
		return nil
	}

	// Set directories to 755
	cmd := exec.Command("find", dataRoot, "-type", "d", "-exec", "chmod", "755", "{}", ";")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set directory permissions: %s: %w", string(output), err)
	}

	// Set files to 644
	cmd = exec.Command("find", dataRoot, "-type", "f", "-exec", "chmod", "644", "{}", ";")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set file permissions: %s: %w", string(output), err)
	}

	fmt.Printf("Set permissions on %s (dirs: 755, files: 644)\n", dataRoot)
	return nil
}

// CountByService returns a map of service -> directory count
func CountByService(specs []DirectorySpec) map[string]int {
	counts := make(map[string]int)
	for _, spec := range specs {
		counts[spec.Service]++
	}
	return counts
}

// CountByType returns a map of directory type -> count
func CountByType(specs []DirectorySpec) map[DirectoryType]int {
	counts := make(map[DirectoryType]int)
	for _, spec := range specs {
		counts[spec.Type]++
	}
	return counts
}

// HasErrors checks if any results have errors
func HasErrors(results []DirectoryResult) bool {
	for _, r := range results {
		if r.Error != nil {
			return true
		}
	}
	return false
}

// CountCreated returns the number of newly created directories
func CountCreated(results []DirectoryResult) int {
	count := 0
	for _, r := range results {
		if r.Created && r.Error == nil {
			count++
		}
	}
	return count
}

// CountExisting returns the number of directories that already existed
func CountExisting(results []DirectoryResult) int {
	count := 0
	for _, r := range results {
		if !r.Created && r.Error == nil {
			count++
		}
	}
	return count
}

// CountFailed returns the number of directories that failed to create
func CountFailed(results []DirectoryResult) int {
	count := 0
	for _, r := range results {
		if r.Error != nil {
			count++
		}
	}
	return count
}
