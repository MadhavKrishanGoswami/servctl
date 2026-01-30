package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Test with temp file
	tmpFile, err := os.CreateTemp("", "test-file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if !FileExists(tmpFile.Name()) {
		t.Error("FileExists should return true for existing file")
	}

	if FileExists("/nonexistent/path/file.txt") {
		t.Error("FileExists should return false for non-existent file")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if !DirExists(tmpDir) {
		t.Error("DirExists should return true for existing directory")
	}

	if DirExists("/nonexistent/directory") {
		t.Error("DirExists should return false for non-existent directory")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "servctl-test-ensure")
	defer os.RemoveAll(tmpDir)

	// First call should create
	created, err := EnsureDir(tmpDir, 0755)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	if !created {
		t.Error("EnsureDir should return created=true for new directory")
	}

	// Second call should not create
	created, err = EnsureDir(tmpDir, 0755)
	if err != nil {
		t.Fatalf("EnsureDir failed on second call: %v", err)
	}
	if created {
		t.Error("EnsureDir should return created=false for existing directory")
	}
}

func TestContainsLine(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-contains")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "line one\nline two\nline three\n"
	tmpFile.WriteString(content)
	tmpFile.Close()

	tests := []struct {
		line     string
		expected bool
	}{
		{"line one", true},
		{"line two", true},
		{"line three", true},
		{"line four", false},
		{"one", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			exists, err := ContainsLine(tmpFile.Name(), tt.line)
			if err != nil {
				t.Fatalf("ContainsLine error: %v", err)
			}
			if exists != tt.expected {
				t.Errorf("ContainsLine(%q) = %v, want %v", tt.line, exists, tt.expected)
			}
		})
	}
}

func TestAppendLineIfMissing(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-append")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("existing line\n")
	tmpFile.Close()

	// First append should succeed
	added, err := AppendLineIfMissing(tmpFile.Name(), "new line")
	if err != nil {
		t.Fatalf("AppendLineIfMissing failed: %v", err)
	}
	if !added {
		t.Error("Should return true when adding new line")
	}

	// Second append should be idempotent
	added, err = AppendLineIfMissing(tmpFile.Name(), "new line")
	if err != nil {
		t.Fatalf("AppendLineIfMissing failed: %v", err)
	}
	if added {
		t.Error("Should return false when line already exists")
	}

	// Verify content
	content, _ := os.ReadFile(tmpFile.Name())
	if count := countOccurrences(string(content), "new line"); count != 1 {
		t.Errorf("Line should appear exactly once, found %d times", count)
	}
}

func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}

func TestBackupFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-backup")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.WriteString("original content")
	tmpFile.Close()

	backupPath, err := BackupFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("BackupFile failed: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer os.Remove(backupPath)

	if backupPath == "" {
		t.Error("BackupFile should return backup path")
	}

	// Verify backup content
	content, _ := os.ReadFile(backupPath)
	if string(content) != "original content" {
		t.Error("Backup content should match original")
	}
}

func TestBackupFileNonExistent(t *testing.T) {
	backupPath, err := BackupFile("/nonexistent/file.txt")
	if err != nil {
		t.Error("BackupFile should not error for non-existent file")
	}
	if backupPath != "" {
		t.Error("BackupFile should return empty string for non-existent file")
	}
}

func TestServctlError(t *testing.T) {
	err := NewCriticalError(
		"Storage",
		"Format disk",
		os.ErrPermission,
		"Run with sudo",
		"Check disk permissions",
	)

	if !err.IsCritical {
		t.Error("Should be marked as critical")
	}

	errorStr := err.Error()
	if errorStr == "" {
		t.Error("Error string should not be empty")
	}

	formatted := FormatError(err)
	if formatted == "" {
		t.Error("Formatted error should not be empty")
	}

	// Check it contains expected parts
	if !containsString(formatted, "CRITICAL") {
		t.Error("Formatted error should contain 'CRITICAL'")
	}
	if !containsString(formatted, "Storage") {
		t.Error("Formatted error should contain phase")
	}
}

func TestWarningError(t *testing.T) {
	err := NewWarningError(
		"Preflight",
		"Check Docker",
		os.ErrNotExist,
		"Install Docker",
	)

	if err.IsCritical {
		t.Error("Should not be marked as critical")
	}

	formatted := FormatError(err)
	if !containsString(formatted, "WARNING") {
		t.Error("Formatted error should contain 'WARNING'")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"~/foo/bar", filepath.Join(home, "foo/bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExpandHome(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	// Test with empty log dir
	logger, err := NewLogger("")
	if err != nil {
		t.Fatalf("NewLogger with empty dir failed: %v", err)
	}
	if logger == nil {
		t.Error("Logger should not be nil")
	}
	logger.Close()

	// Test with temp dir
	tmpDir := filepath.Join(os.TempDir(), "servctl-log-test")
	defer os.RemoveAll(tmpDir)

	logger, err = NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("Test info message")
	logger.Warn("Test warning")
	logger.Error("Test error")

	// Verify log file exists
	if !FileExists(filepath.Join(tmpDir, "servctl.log")) {
		t.Error("Log file should exist")
	}
}

// Benchmark
func BenchmarkEnsureDir(b *testing.B) {
	tmpDir := filepath.Join(os.TempDir(), "servctl-bench-ensure")
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EnsureDir(tmpDir, 0755)
	}
}
