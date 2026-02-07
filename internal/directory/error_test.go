package directory

import (
	"testing"
)

// =============================================================================
// Error Path Tests - Verify graceful handling of failure scenarios
// =============================================================================

// TestGetDirectoriesForServices_EmptyPaths tests with empty paths
func TestGetDirectoriesForServices_EmptyPaths(t *testing.T) {
	sel := DefaultServiceSelection()

	dirs := GetDirectoriesForServices(sel, "", "")

	// Empty base paths may result in directories with relative paths
	for _, d := range dirs {
		if d.Path == "" {
			t.Logf("Warning: Directory with empty path found (empty base paths)")
		}
	}
}

// TestGetDirectoriesForServices_NoneSelected tests with all services disabled
func TestGetDirectoriesForServices_NoneSelected(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: false,
		Immich:    false,
		Databases: false,
		Glances:   false,
	}

	dirs := GetDirectoriesForServices(sel, "/home/user", "/mnt/data")

	// Should still have core infrastructure directories
	t.Logf("Empty selection returned %d directories", len(dirs))
}

// TestGetDirectoriesForServices_AllSelected tests with all services enabled
func TestGetDirectoriesForServices_AllSelected(t *testing.T) {
	sel := ServiceSelection{
		Nextcloud: true,
		Immich:    true,
		Databases: true,
		Glances:   true,
	}

	dirs := GetDirectoriesForServices(sel, "/home/user", "/mnt/data")

	if len(dirs) == 0 {
		t.Error("All services selected should return directories")
	}
}

// TestDefaultServiceSelection_HasDefaults tests default selection values
func TestDefaultServiceSelection_HasDefaults(t *testing.T) {
	sel := DefaultServiceSelection()

	// Should have at least some services enabled
	hasEnabled := sel.Nextcloud || sel.Immich || sel.Databases || sel.Glances
	if !hasEnabled {
		t.Error("Default selection should have some services enabled")
	}
}

// TestGetDirectoriesForServices_PathSanitization tests path handling
func TestGetDirectoriesForServices_PathSanitization(t *testing.T) {
	sel := DefaultServiceSelection()

	// Test with trailing slashes - should be sanitized properly now
	dirs := GetDirectoriesForServices(sel, "/home/user/", "/mnt/data/")

	for _, d := range dirs {
		// Bug fix verification: no paths should have consecutive slashes
		if containsConsecutiveSlashes(d.Path) {
			t.Errorf("Path has consecutive slashes (bug not fixed): %s", d.Path)
		}
	}
}

// TestGetDirectoriesForServices_SpecialCharPaths tests with special characters
func TestGetDirectoriesForServices_SpecialCharPaths(t *testing.T) {
	sel := DefaultServiceSelection()

	// Test with space in path
	dirs := GetDirectoriesForServices(sel, "/home/user name", "/mnt/data drive")

	// Should work (paths with spaces are valid on Linux)
	if len(dirs) == 0 {
		t.Error("Should generate directories even with spaces in paths")
	}
}

// TestDirectory_HasRequiredFields tests directory struct fields
func TestDirectory_HasRequiredFields(t *testing.T) {
	sel := DefaultServiceSelection()
	dirs := GetDirectoriesForServices(sel, "/home/user", "/mnt/data")

	for _, d := range dirs {
		if d.Path == "" {
			t.Error("Directory has empty path")
		}
		// Description may be optional
		t.Logf("Directory: %s (%s)", d.Path, d.Description)
	}
}

// Helper function
func containsConsecutiveSlashes(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '/' && s[i+1] == '/' {
			return true
		}
	}
	return false
}
