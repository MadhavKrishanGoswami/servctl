package directory

import (
	"testing"
)

// =============================================================================
// ServiceSelection Tests
// =============================================================================

func TestDefaultServiceSelection(t *testing.T) {
	sel := DefaultServiceSelection()

	if !sel.Nextcloud {
		t.Error("Nextcloud should be enabled by default")
	}
	if !sel.Immich {
		t.Error("Immich should be enabled by default")
	}
	if !sel.Databases {
		t.Error("Databases should be enabled by default")
	}
	if !sel.Glances {
		t.Error("Glances should be enabled by default")
	}
}

func TestServiceSelection_CountSelectedServices(t *testing.T) {
	tests := []struct {
		sel      ServiceSelection
		expected int
	}{
		{ServiceSelection{Nextcloud: true, Immich: true, Databases: true, Glances: true}, 4},
		{ServiceSelection{Nextcloud: true, Immich: false, Databases: true, Glances: false}, 2},
		{ServiceSelection{Nextcloud: false, Immich: false, Databases: false, Glances: false}, 0},
		{ServiceSelection{Nextcloud: true, Immich: false, Databases: false, Glances: false}, 1},
	}

	for _, tt := range tests {
		result := tt.sel.CountSelectedServices()
		if result != tt.expected {
			t.Errorf("CountSelectedServices() = %d, want %d", result, tt.expected)
		}
	}
}

func TestServiceSelection_SelectedNames(t *testing.T) {
	tests := []struct {
		sel      ServiceSelection
		expected []string
	}{
		{
			ServiceSelection{Nextcloud: true, Immich: true, Databases: true, Glances: true},
			[]string{"Nextcloud", "Immich", "Databases", "Glances"},
		},
		{
			ServiceSelection{Nextcloud: true, Immich: false, Databases: false, Glances: false},
			[]string{"Nextcloud"},
		},
		{
			ServiceSelection{Nextcloud: false, Immich: true, Databases: false, Glances: true},
			[]string{"Immich", "Glances"},
		},
		{
			ServiceSelection{Nextcloud: false, Immich: false, Databases: false, Glances: false},
			[]string{},
		},
	}

	for _, tt := range tests {
		result := tt.sel.SelectedNames()
		if len(result) != len(tt.expected) {
			t.Errorf("SelectedNames() returned %d items, want %d", len(result), len(tt.expected))
			continue
		}
		for i, name := range result {
			if name != tt.expected[i] {
				t.Errorf("SelectedNames()[%d] = %s, want %s", i, name, tt.expected[i])
			}
		}
	}
}

// =============================================================================
// GetDirectoriesForServices Tests
// =============================================================================

func TestGetDirectoriesForServices_AllServices(t *testing.T) {
	sel := DefaultServiceSelection()
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Should have core + nextcloud + immich + databases + glances directories
	if len(dirs) < 10 {
		t.Errorf("Expected at least 10 directories with all services, got %d", len(dirs))
	}

	// Check for core directories
	hasInfra := false
	hasCompose := false
	hasScripts := false
	for _, d := range dirs {
		if d.Path == homeDir+"/infra" {
			hasInfra = true
		}
		if d.Path == homeDir+"/infra/compose" {
			hasCompose = true
		}
		if d.Path == homeDir+"/infra/scripts" {
			hasScripts = true
		}
	}
	if !hasInfra {
		t.Error("Missing /home/testuser/infra directory")
	}
	if !hasCompose {
		t.Error("Missing /home/testuser/infra/compose directory")
	}
	if !hasScripts {
		t.Error("Missing /home/testuser/infra/scripts directory")
	}

	// Check for service-specific directories
	hasNextcloudData := false
	hasImmichUpload := false
	hasPostgres := false
	for _, d := range dirs {
		if d.Path == dataRoot+"/nextcloud/data" {
			hasNextcloudData = true
		}
		if d.Path == dataRoot+"/immich/upload" {
			hasImmichUpload = true
		}
		if d.Path == dataRoot+"/databases/postgres" {
			hasPostgres = true
		}
	}
	if !hasNextcloudData {
		t.Error("Missing nextcloud data directory")
	}
	if !hasImmichUpload {
		t.Error("Missing immich upload directory")
	}
	if !hasPostgres {
		t.Error("Missing postgres directory")
	}
}

func TestGetDirectoriesForServices_NextcloudOnly(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: true,
		Immich:    false,
		Databases: false,
		Glances:   false,
	}
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Should have core + nextcloud directories only
	hasNextcloud := false
	hasImmich := false
	hasDatabases := false

	for _, d := range dirs {
		if d.Service == "nextcloud" {
			hasNextcloud = true
		}
		if d.Service == "immich" {
			hasImmich = true
		}
		if d.Service == "databases" {
			hasDatabases = true
		}
	}

	if !hasNextcloud {
		t.Error("Should have Nextcloud directories when Nextcloud is selected")
	}
	if hasImmich {
		t.Error("Should NOT have Immich directories when Immich is not selected")
	}
	if hasDatabases {
		t.Error("Should NOT have Databases directories when Databases is not selected")
	}
}

func TestGetDirectoriesForServices_NoServices(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: false,
		Immich:    false,
		Databases: false,
		Glances:   false,
	}
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Should still have core directories
	hasCore := false
	for _, d := range dirs {
		if d.Service == "core" {
			hasCore = true
		}
	}
	if !hasCore {
		t.Error("Should always have core directories")
	}

	// Should not have any service-specific directories
	for _, d := range dirs {
		if d.Service != "core" {
			t.Errorf("Should not have service directory: %s (%s)", d.Path, d.Service)
		}
	}
}

func TestGetDirectoriesForServices_ImmichOnly(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: false,
		Immich:    true,
		Databases: false,
		Glances:   false,
	}
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Check for Immich-specific directories
	hasImmich := false
	hasUpload := false
	hasLibrary := false
	hasThumbs := false

	for _, d := range dirs {
		if d.Path == dataRoot+"/immich" {
			hasImmich = true
		}
		if d.Path == dataRoot+"/immich/upload" {
			hasUpload = true
		}
		if d.Path == dataRoot+"/immich/library" {
			hasLibrary = true
		}
		if d.Path == dataRoot+"/immich/thumbs" {
			hasThumbs = true
		}
	}

	if !hasImmich {
		t.Error("Missing immich root directory")
	}
	if !hasUpload {
		t.Error("Missing immich upload directory")
	}
	if !hasLibrary {
		t.Error("Missing immich library directory")
	}
	if !hasThumbs {
		t.Error("Missing immich thumbs directory")
	}
}

func TestGetDirectoriesForServices_DatabasesOnly(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: false,
		Immich:    false,
		Databases: true,
		Glances:   false,
	}
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Check for database directories
	hasDatabases := false
	hasPostgres := false
	hasRedis := false

	for _, d := range dirs {
		if d.Path == dataRoot+"/databases" {
			hasDatabases = true
		}
		if d.Path == dataRoot+"/databases/postgres" {
			hasPostgres = true
		}
		if d.Path == dataRoot+"/databases/redis" {
			hasRedis = true
		}
	}

	if !hasDatabases {
		t.Error("Missing databases root directory")
	}
	if !hasPostgres {
		t.Error("Missing postgres directory")
	}
	if !hasRedis {
		t.Error("Missing redis directory")
	}
}

func TestGetDirectoriesForServices_GlancesOnly(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: false,
		Immich:    false,
		Databases: false,
		Glances:   true,
	}
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Check for Glances config directory
	hasGlances := false
	for _, d := range dirs {
		if d.Path == homeDir+"/infra/glances" && d.Service == "glances" {
			hasGlances = true
		}
	}

	if !hasGlances {
		t.Error("Missing glances config directory")
	}
}

// =============================================================================
// DirectorySpec Tests
// =============================================================================

func TestDirectorySpec_Permissions(t *testing.T) {
	sel := DefaultServiceSelection()
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	// Check that database directories have restricted permissions
	for _, d := range dirs {
		if d.Service == "databases" && d.Path != dataRoot+"/databases" {
			if d.Mode != 0700 {
				t.Errorf("Database directory %s should have mode 0700, got %o", d.Path, d.Mode)
			}
		}
	}

	// Check that data directories have appropriate permissions
	for _, d := range dirs {
		if d.Path == dataRoot+"/nextcloud/data" || d.Path == dataRoot+"/immich/upload" {
			if d.Mode != 0770 {
				t.Errorf("Data directory %s should have mode 0770, got %o", d.Path, d.Mode)
			}
		}
	}
}

func TestDirectorySpec_Types(t *testing.T) {
	sel := DefaultServiceSelection()
	homeDir := "/home/testuser"
	dataRoot := "/mnt/data"

	dirs := GetDirectoriesForServices(sel, homeDir, dataRoot)

	for _, d := range dirs {
		// Infra directories should be UserSpace
		if containsPath(d.Path, homeDir+"/infra") {
			if d.Type != DirTypeUserSpace {
				t.Errorf("Directory %s should be UserSpace type", d.Path)
			}
		}
		// Data directories should be DataSpace
		if containsPath(d.Path, dataRoot) && d.Path != dataRoot {
			if d.Type != DirTypeDataSpace {
				t.Errorf("Directory %s should be DataSpace type", d.Path)
			}
		}
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsPath(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}
