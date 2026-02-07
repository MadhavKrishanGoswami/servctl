package maintenance

import (
	"testing"
)

// =============================================================================
// Error Path Tests - Verify graceful handling of failure scenarios
// =============================================================================

// TestGetScriptsForSelection_NoneSelected_ErrorPath tests with all scripts disabled
func TestGetScriptsForSelection_NoneSelected_ErrorPath(t *testing.T) {
	sel := ScriptSelection{
		DailyBackup:   false,
		DiskAlert:     false,
		SmartAlert:    false,
		WeeklyCleanup: false,
	}
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)

	if err != nil {
		t.Errorf("Error with empty selection: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("Empty selection should return no scripts, got %d", len(scripts))
	}
}

// TestGetScriptsForSelection_NilConfig tests with nil config
func TestGetScriptsForSelection_NilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetScriptsForSelection panicked with nil config: %v", r)
		}
	}()

	sel := DefaultScriptSelection()
	scripts, err := GetScriptsForSelection(sel, nil)

	if err != nil {
		t.Logf("Expected error with nil config: %v", err)
	}
	t.Logf("Nil config returned %d scripts", len(scripts))
}

// TestDefaultScriptConfig_ValidValues tests that default config has valid values
func TestDefaultScriptConfig_ValidValues(t *testing.T) {
	config := DefaultScriptConfig()

	if config == nil {
		t.Fatal("DefaultScriptConfig returned nil")
	}

	// Check valid ranges
	if config.DiskAlertThreshold < 0 || config.DiskAlertThreshold > 100 {
		t.Errorf("DiskAlertThreshold should be 0-100, got %d", config.DiskAlertThreshold)
	}
}

// TestScriptSelection_SelectedNames tests SelectedNames method
func TestScriptSelection_SelectedNames_Error(t *testing.T) {
	sel := ScriptSelection{
		DailyBackup:   true,
		DiskAlert:     true,
		SmartAlert:    false,
		WeeklyCleanup: false,
	}

	names := sel.SelectedNames()

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d: %v", len(names), names)
	}
}

// TestScriptSelection_SelectedNames_Empty tests with no selections
func TestScriptSelection_SelectedNames_Empty(t *testing.T) {
	sel := ScriptSelection{}

	names := sel.SelectedNames()

	if len(names) != 0 {
		t.Errorf("Empty selection should return no names, got: %v", names)
	}
}

// TestScriptContent_NotEmpty tests that generated scripts have content
func TestScriptContent_NotEmpty_Error(t *testing.T) {
	sel := DefaultScriptSelection()
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)
	if err != nil {
		t.Fatalf("GetScriptsForSelection failed: %v", err)
	}

	for _, script := range scripts {
		if script.Name == "" {
			t.Error("Script has empty name")
		}
		if script.Content == "" {
			t.Errorf("Script %s has empty content", script.Name)
		}
		if script.Filename == "" {
			t.Errorf("Script %s has empty filename", script.Name)
		}
	}
}

// TestScriptContent_ValidBashSyntax tests scripts start with shebang
func TestScriptContent_ValidBashSyntax_Error(t *testing.T) {
	sel := DefaultScriptSelection()
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)
	if err != nil {
		t.Fatalf("GetScriptsForSelection failed: %v", err)
	}

	for _, script := range scripts {
		if len(script.Content) < 2 {
			t.Errorf("Script %s content too short", script.Name)
			continue
		}

		// Check for shebang
		if script.Content[0:2] != "#!" {
			t.Errorf("Script %s missing shebang, starts with: %q", script.Name, script.Content[:min(10, len(script.Content))])
		}
	}
}

// TestDefaultScriptSelection_HasDefaults tests default selection has scripts
func TestDefaultScriptSelection_HasDefaults_Error(t *testing.T) {
	sel := DefaultScriptSelection()

	// At least some scripts should be enabled by default
	enabled := 0
	if sel.DailyBackup {
		enabled++
	}
	if sel.DiskAlert {
		enabled++
	}
	if sel.SmartAlert {
		enabled++
	}
	if sel.WeeklyCleanup {
		enabled++
	}

	if enabled == 0 {
		t.Error("Default selection should have at least one script enabled")
	}
}

// TestGetScriptsForSelection_AllSelected tests with all scripts enabled
func TestGetScriptsForSelection_AllSelected(t *testing.T) {
	sel := ScriptSelection{
		DailyBackup:   true,
		DiskAlert:     true,
		SmartAlert:    true,
		WeeklyCleanup: true,
	}
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)
	if err != nil {
		t.Fatalf("GetScriptsForSelection failed: %v", err)
	}

	if len(scripts) != 4 {
		t.Errorf("All 4 scripts enabled should return 4 scripts, got %d", len(scripts))
	}
}

// TestScriptInfo_HasRequiredFields tests ScriptInfo struct
func TestScriptInfo_HasRequiredFields(t *testing.T) {
	sel := DefaultScriptSelection()
	sel.DailyBackup = true
	sel.DiskAlert = false
	sel.SmartAlert = false
	sel.WeeklyCleanup = false

	config := DefaultScriptConfig()
	scripts, err := GetScriptsForSelection(sel, config)
	if err != nil {
		t.Fatalf("GetScriptsForSelection failed: %v", err)
	}

	if len(scripts) == 0 {
		t.Fatal("Expected at least one script")
	}

	script := scripts[0]
	if script.Name == "" {
		t.Error("Script Name should not be empty")
	}
	if script.Filename == "" {
		t.Error("Script Filename should not be empty")
	}
	if script.Schedule == "" {
		t.Error("Script Schedule should not be empty")
	}
}

// Helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
