package maintenance

import (
	"testing"
)

// =============================================================================
// ScriptSelection Tests
// =============================================================================

func TestDefaultScriptSelection(t *testing.T) {
	sel := DefaultScriptSelection()

	if !sel.DailyBackup {
		t.Error("DailyBackup should be enabled by default")
	}
	if !sel.DiskAlert {
		t.Error("DiskAlert should be enabled by default")
	}
	if sel.SmartAlert {
		t.Error("SmartAlert should be DISABLED by default (requires smartctl)")
	}
	if !sel.WeeklyCleanup {
		t.Error("WeeklyCleanup should be enabled by default")
	}
}

func TestScriptSelection_SelectedNames(t *testing.T) {
	tests := []struct {
		name     string
		sel      ScriptSelection
		expected []string
	}{
		{
			name:     "all enabled",
			sel:      ScriptSelection{DailyBackup: true, DiskAlert: true, SmartAlert: true, WeeklyCleanup: true},
			expected: []string{"Daily Backup", "Disk Alert", "SMART Monitor", "Weekly Cleanup"},
		},
		{
			name:     "default selection",
			sel:      DefaultScriptSelection(),
			expected: []string{"Daily Backup", "Disk Alert", "Weekly Cleanup"},
		},
		{
			name:     "backup only",
			sel:      ScriptSelection{DailyBackup: true, DiskAlert: false, SmartAlert: false, WeeklyCleanup: false},
			expected: []string{"Daily Backup"},
		},
		{
			name:     "none selected",
			sel:      ScriptSelection{DailyBackup: false, DiskAlert: false, SmartAlert: false, WeeklyCleanup: false},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sel.SelectedNames()
			if len(result) != len(tt.expected) {
				t.Errorf("SelectedNames() returned %d items, want %d", len(result), len(tt.expected))
				return
			}
			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("SelectedNames()[%d] = %s, want %s", i, name, tt.expected[i])
				}
			}
		})
	}
}

// =============================================================================
// GetScriptsForSelection Tests
// =============================================================================

func TestGetScriptsForSelection_AllEnabled(t *testing.T) {
	sel := ScriptSelection{
		DailyBackup:   true,
		DiskAlert:     true,
		SmartAlert:    true,
		WeeklyCleanup: true,
	}
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(scripts) != 4 {
		t.Errorf("Expected 4 scripts, got %d", len(scripts))
	}

	// Check script names
	names := make(map[string]bool)
	for _, s := range scripts {
		names[s.Name] = true
	}
	if !names["Daily Backup"] {
		t.Error("Missing Daily Backup script")
	}
	if !names["Disk Alert"] {
		t.Error("Missing Disk Alert script")
	}
	if !names["SMART Monitor"] {
		t.Error("Missing SMART Monitor script")
	}
	if !names["Weekly Cleanup"] {
		t.Error("Missing Weekly Cleanup script")
	}
}

func TestGetScriptsForSelection_Default(t *testing.T) {
	sel := DefaultScriptSelection()
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Default has DailyBackup, DiskAlert, WeeklyCleanup enabled
	if len(scripts) != 3 {
		t.Errorf("Expected 3 scripts for default selection, got %d", len(scripts))
	}

	// Check that SmartAlert is NOT included
	for _, s := range scripts {
		if s.Name == "SMART Monitor" {
			t.Error("SMART Monitor should not be included in default selection")
		}
	}
}

func TestGetScriptsForSelection_BackupOnly(t *testing.T) {
	sel := ScriptSelection{
		DailyBackup:   true,
		DiskAlert:     false,
		SmartAlert:    false,
		WeeklyCleanup: false,
	}
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(scripts))
	}
	if scripts[0].Name != "Daily Backup" {
		t.Errorf("Expected Daily Backup script, got %s", scripts[0].Name)
	}
}

func TestGetScriptsForSelection_NoneSelected(t *testing.T) {
	sel := ScriptSelection{
		DailyBackup:   false,
		DiskAlert:     false,
		SmartAlert:    false,
		WeeklyCleanup: false,
	}
	config := DefaultScriptConfig()

	scripts, err := GetScriptsForSelection(sel, config)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("Expected 0 scripts when none selected, got %d", len(scripts))
	}
}

// =============================================================================
// ScriptInfo Tests
// =============================================================================

func TestGetScriptsForSelection_ScriptInfo(t *testing.T) {
	sel := ScriptSelection{DailyBackup: true}
	config := DefaultScriptConfig()

	scripts, _ := GetScriptsForSelection(sel, config)

	if len(scripts) == 0 {
		t.Fatal("Expected at least one script")
	}

	script := scripts[0]

	// Check ScriptInfo fields
	if script.Name == "" {
		t.Error("Script name should not be empty")
	}
	if script.Filename == "" {
		t.Error("Script filename should not be empty")
	}
	if script.Description == "" {
		t.Error("Script description should not be empty")
	}
	if script.Schedule == "" {
		t.Error("Script schedule should not be empty")
	}
	if script.Content == "" {
		t.Error("Script content should not be empty")
	}
	if script.Filename != "daily-backup.sh" {
		t.Errorf("Expected filename 'daily-backup.sh', got '%s'", script.Filename)
	}
}

// =============================================================================
// Script Content Tests
// =============================================================================

func TestGetScriptsForSelection_BackupContent(t *testing.T) {
	sel := ScriptSelection{DailyBackup: true}
	config := DefaultScriptConfig()
	config.DataRoot = "/mnt/data"
	config.BackupDest = "/mnt/backup"

	scripts, _ := GetScriptsForSelection(sel, config)

	if len(scripts) == 0 {
		t.Fatal("Expected backup script")
	}

	content := scripts[0].Content

	// Script should contain rsync command
	if !containsStr(content, "rsync") {
		t.Error("Backup script should contain rsync command")
	}
	// Script should be a bash script
	if !containsStr(content, "#!/bin/bash") {
		t.Error("Script should start with bash shebang")
	}
}

func TestGetScriptsForSelection_DiskAlertContent(t *testing.T) {
	sel := ScriptSelection{DiskAlert: true}
	config := DefaultScriptConfig()
	config.DiskAlertThreshold = 90

	scripts, _ := GetScriptsForSelection(sel, config)

	if len(scripts) == 0 {
		t.Fatal("Expected disk alert script")
	}

	content := scripts[0].Content

	// Script should check disk usage
	if !containsStr(content, "df") {
		t.Error("Disk alert script should use df command")
	}
}

func TestGetScriptsForSelection_WeeklyCleanupContent(t *testing.T) {
	sel := ScriptSelection{WeeklyCleanup: true}
	config := DefaultScriptConfig()

	scripts, _ := GetScriptsForSelection(sel, config)

	if len(scripts) == 0 {
		t.Fatal("Expected weekly cleanup script")
	}

	content := scripts[0].Content

	// Script should contain docker cleanup
	if !containsStr(content, "docker") {
		t.Error("Weekly cleanup should include docker cleanup")
	}
	// Script should contain apt cleanup
	if !containsStr(content, "apt") {
		t.Error("Weekly cleanup should include apt cleanup")
	}
}

// =============================================================================
// DefaultScriptConfig Tests
// =============================================================================

func TestDefaultScriptConfig_Selection(t *testing.T) {
	config := DefaultScriptConfig()

	if config.DiskAlertThreshold != 90 {
		t.Errorf("Expected DiskAlertThreshold 90, got %d", config.DiskAlertThreshold)
	}
	if config.BackupRetentionDays != 7 {
		t.Errorf("Expected BackupRetentionDays 7, got %d", config.BackupRetentionDays)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
