package directory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryTypeString(t *testing.T) {
	tests := []struct {
		dirType  DirectoryType
		expected string
	}{
		{DirTypeUserSpace, "User Space"},
		{DirTypeDataSpace, "Data Space"},
		{DirectoryType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.dirType.String(); got != tt.expected {
				t.Errorf("DirectoryType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetUserSpaceDirectories(t *testing.T) {
	homeDir := "/home/testuser"
	dirs := GetUserSpaceDirectories(homeDir)

	if len(dirs) == 0 {
		t.Error("GetUserSpaceDirectories() returned empty list")
	}

	// Check that all paths are under home directory
	for _, dir := range dirs {
		if !filepath.HasPrefix(dir.Path, homeDir) {
			t.Errorf("Directory %s is not under home directory %s", dir.Path, homeDir)
		}
		if dir.Type != DirTypeUserSpace {
			t.Errorf("Directory %s has wrong type %v", dir.Path, dir.Type)
		}
	}

	// Verify key directories exist
	expectedPaths := []string{
		filepath.Join(homeDir, "infra"),
		filepath.Join(homeDir, "infra", "scripts"),
		filepath.Join(homeDir, "infra", "logs"),
	}

	pathMap := make(map[string]bool)
	for _, dir := range dirs {
		pathMap[dir.Path] = true
	}

	for _, expected := range expectedPaths {
		if !pathMap[expected] {
			t.Errorf("Missing expected directory: %s", expected)
		}
	}
}

func TestGetDataSpaceDirectories(t *testing.T) {
	dataRoot := "/mnt/data"
	dirs := GetDataSpaceDirectories(dataRoot)

	if len(dirs) == 0 {
		t.Error("GetDataSpaceDirectories() returned empty list")
	}

	// Check that all paths are under data root
	for _, dir := range dirs {
		if !filepath.HasPrefix(dir.Path, dataRoot) {
			t.Errorf("Directory %s is not under data root %s", dir.Path, dataRoot)
		}
		if dir.Type != DirTypeDataSpace {
			t.Errorf("Directory %s has wrong type %v", dir.Path, dir.Type)
		}
	}

	// Verify key directories exist
	expectedPaths := []string{
		dataRoot,
		filepath.Join(dataRoot, "gallery"),
		filepath.Join(dataRoot, "gallery", "library"),
		filepath.Join(dataRoot, "cloud"),
		filepath.Join(dataRoot, "cloud", "data"),
		filepath.Join(dataRoot, "databases"),
		filepath.Join(dataRoot, "databases", "immich-postgres"),
		filepath.Join(dataRoot, "databases", "nextcloud-mariadb"),
	}

	pathMap := make(map[string]bool)
	for _, dir := range dirs {
		pathMap[dir.Path] = true
	}

	for _, expected := range expectedPaths {
		if !pathMap[expected] {
			t.Errorf("Missing expected directory: %s", expected)
		}
	}
}

func TestGetDataSpaceDirectoriesDefaultRoot(t *testing.T) {
	// Empty string should default to /mnt/data
	dirs := GetDataSpaceDirectories("")

	if len(dirs) == 0 {
		t.Error("GetDataSpaceDirectories('') returned empty list")
	}

	// First directory should be /mnt/data
	if dirs[0].Path != "/mnt/data" {
		t.Errorf("Default root should be /mnt/data, got %s", dirs[0].Path)
	}
}

func TestGetAllDirectories(t *testing.T) {
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	all := GetAllDirectories(homeDir, dataRoot)
	userDirs := GetUserSpaceDirectories(homeDir)
	dataDirs := GetDataSpaceDirectories(dataRoot)

	expectedCount := len(userDirs) + len(dataDirs)
	if len(all) != expectedCount {
		t.Errorf("GetAllDirectories() returned %d, expected %d", len(all), expectedCount)
	}
}

func TestCountByService(t *testing.T) {
	specs := []DirectorySpec{
		{Service: "immich"},
		{Service: "immich"},
		{Service: "nextcloud"},
		{Service: "core"},
	}

	counts := CountByService(specs)

	if counts["immich"] != 2 {
		t.Errorf("CountByService['immich'] = %d, want 2", counts["immich"])
	}
	if counts["nextcloud"] != 1 {
		t.Errorf("CountByService['nextcloud'] = %d, want 1", counts["nextcloud"])
	}
}

func TestCountByType(t *testing.T) {
	specs := []DirectorySpec{
		{Type: DirTypeUserSpace},
		{Type: DirTypeUserSpace},
		{Type: DirTypeDataSpace},
		{Type: DirTypeDataSpace},
		{Type: DirTypeDataSpace},
	}

	counts := CountByType(specs)

	if counts[DirTypeUserSpace] != 2 {
		t.Errorf("CountByType[UserSpace] = %d, want 2", counts[DirTypeUserSpace])
	}
	if counts[DirTypeDataSpace] != 3 {
		t.Errorf("CountByType[DataSpace] = %d, want 3", counts[DirTypeDataSpace])
	}
}

func TestDirectorySpecStructure(t *testing.T) {
	spec := DirectorySpec{
		Path:        "/mnt/data/gallery",
		Type:        DirTypeDataSpace,
		Service:     "immich",
		Description: "Immich photo gallery",
		Mode:        0755,
	}

	if spec.Path != "/mnt/data/gallery" {
		t.Errorf("spec.Path = %v, want /mnt/data/gallery", spec.Path)
	}
	if spec.Type != DirTypeDataSpace {
		t.Errorf("spec.Type = %v, want DirTypeDataSpace", spec.Type)
	}
	if spec.Mode != 0755 {
		t.Errorf("spec.Mode = %v, want 0755", spec.Mode)
	}
}

func TestCreateDirectoryDryRun(t *testing.T) {
	spec := DirectorySpec{
		Path:        "/tmp/servctl-test-dryrun",
		Type:        DirTypeDataSpace,
		Service:     "test",
		Description: "Test directory",
		Mode:        0755,
	}

	result := CreateDirectory(spec, true)

	if result.Error != nil {
		t.Errorf("CreateDirectory() dry run error: %v", result.Error)
	}
	if !result.Created {
		t.Error("CreateDirectory() dry run should report Created=true")
	}

	// Directory should NOT actually exist
	if _, err := os.Stat(spec.Path); err == nil {
		t.Error("Dry run should not create actual directory")
	}
}

func TestCreateDirectoryActual(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping directory creation in short mode")
	}

	// Create a unique temp directory for testing
	tempDir := filepath.Join(os.TempDir(), "servctl-test-create")
	defer os.RemoveAll(tempDir)

	spec := DirectorySpec{
		Path:        tempDir,
		Type:        DirTypeUserSpace,
		Service:     "test",
		Description: "Test directory",
		Mode:        0755,
	}

	result := CreateDirectory(spec, false)

	if result.Error != nil {
		t.Errorf("CreateDirectory() error: %v", result.Error)
	}
	if !result.Created {
		t.Error("CreateDirectory() should report Created=true")
	}

	// Verify directory exists
	if info, err := os.Stat(spec.Path); err != nil {
		t.Errorf("Directory was not created: %v", err)
	} else if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
}

func TestCreateDirectoryAlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Use temp directory which already exists
	tempDir := os.TempDir()

	spec := DirectorySpec{
		Path:        tempDir,
		Type:        DirTypeUserSpace,
		Service:     "test",
		Description: "Existing directory",
		Mode:        0755,
	}

	result := CreateDirectory(spec, false)

	if result.Error != nil {
		t.Errorf("CreateDirectory() error for existing dir: %v", result.Error)
	}
	if result.Created {
		t.Error("CreateDirectory() should report Created=false for existing dir")
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		results  []DirectoryResult
		expected bool
	}{
		{
			name:     "No errors",
			results:  []DirectoryResult{{Created: true}, {Created: false}},
			expected: false,
		},
		{
			name:     "Has error",
			results:  []DirectoryResult{{Created: true}, {Error: os.ErrPermission}},
			expected: true,
		},
		{
			name:     "Empty",
			results:  []DirectoryResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.results); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCountCreated(t *testing.T) {
	results := []DirectoryResult{
		{Created: true},
		{Created: true},
		{Created: false},
		{Created: true, Error: os.ErrPermission}, // Error, doesn't count
	}

	if got := CountCreated(results); got != 2 {
		t.Errorf("CountCreated() = %d, want 2", got)
	}
}

func TestCountExisting(t *testing.T) {
	results := []DirectoryResult{
		{Created: true},
		{Created: false},
		{Created: false},
	}

	if got := CountExisting(results); got != 2 {
		t.Errorf("CountExisting() = %d, want 2", got)
	}
}

func TestCountFailed(t *testing.T) {
	results := []DirectoryResult{
		{Created: true},
		{Error: os.ErrPermission},
		{Error: os.ErrNotExist},
	}

	if got := CountFailed(results); got != 2 {
		t.Errorf("CountFailed() = %d, want 2", got)
	}
}

func TestGetCurrentUserInfo(t *testing.T) {
	info, err := GetCurrentUserInfo()

	if err != nil {
		t.Errorf("GetCurrentUserInfo() error: %v", err)
	}

	if info == nil {
		t.Fatal("GetCurrentUserInfo() returned nil")
	}

	if info.UID < 0 {
		t.Errorf("Invalid UID: %d", info.UID)
	}
	if info.GID < 0 {
		t.Errorf("Invalid GID: %d", info.GID)
	}
	if info.Username == "" {
		t.Error("Username is empty")
	}
	if info.HomeDir == "" {
		t.Error("HomeDir is empty")
	}

	t.Logf("Current user: %s (UID=%d, GID=%d, Home=%s)",
		info.Username, info.UID, info.GID, info.HomeDir)
}

func TestSetPermissionsDryRun(t *testing.T) {
	info := &PermissionInfo{
		UID:      1000,
		GID:      1000,
		Username: "testuser",
		HomeDir:  "/home/testuser",
	}

	err := SetPermissions("/mnt/data", info, true)

	if err != nil {
		t.Errorf("SetPermissions() dry run error: %v", err)
	}
}

func TestSetDirectoryPermissionsDryRun(t *testing.T) {
	err := SetDirectoryPermissions("/mnt/data", true)

	if err != nil {
		t.Errorf("SetDirectoryPermissions() dry run error: %v", err)
	}
}

// Benchmark tests
func BenchmarkGetAllDirectories(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetAllDirectories("/home/user", "/mnt/data")
	}
}

func BenchmarkCountByService(b *testing.B) {
	specs := GetAllDirectories("/home/user", "/mnt/data")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountByService(specs)
	}
}
