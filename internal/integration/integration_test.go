// Package integration provides integration tests for servctl.
// These tests verify the full workflow works correctly.
package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madhav/servctl/internal/compose"
	"github.com/madhav/servctl/internal/directory"
	"github.com/madhav/servctl/internal/maintenance"
	"github.com/madhav/servctl/internal/preflight"
	"github.com/madhav/servctl/internal/report"
	"github.com/madhav/servctl/internal/utils"
)

// TestFullWorkflow tests the complete setup workflow in dry-run mode
func TestFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full workflow test in short mode")
	}

	// Create temp directories
	tmpHome := filepath.Join(os.TempDir(), "servctl-integration-home")
	tmpData := filepath.Join(os.TempDir(), "servctl-integration-data")
	defer os.RemoveAll(tmpHome)
	defer os.RemoveAll(tmpData)

	infraRoot := filepath.Join(tmpHome, "infra")

	// Phase 1: Preflight (just validate it runs)
	t.Run("Phase1_Preflight", func(t *testing.T) {
		results := preflight.RunAllPreflightChecks()
		if results == nil {
			t.Fatal("Preflight checks returned nil")
		}
		// We don't check for blockers since this could be running on Mac
		t.Logf("Preflight checks completed: %d results", len(results))
	})

	// Phase 3: Directory Structure
	t.Run("Phase3_Directories", func(t *testing.T) {
		userDirs := directory.GetUserSpaceDirectories(tmpHome)
		dataDirs := directory.GetDataSpaceDirectories(tmpData)
		allDirs := append(userDirs, dataDirs...)

		if len(allDirs) == 0 {
			t.Fatal("No directories returned")
		}

		// Create directories for real
		results := directory.CreateAllDirectories(tmpHome, tmpData, false)

		created := directory.CountCreated(results)
		failed := directory.CountFailed(results)

		t.Logf("Created: %d, Failed: %d", created, failed)

		if failed > 0 {
			t.Errorf("Expected 0 failures, got %d", failed)
		}

		// Verify some directories exist
		expectedDirs := []string{
			filepath.Join(tmpHome, "infra"),
			filepath.Join(tmpHome, "infra", "scripts"),
			filepath.Join(tmpHome, "infra", "logs"),
			filepath.Join(tmpData, "gallery"),
			filepath.Join(tmpData, "cloud"),
		}

		for _, dir := range expectedDirs {
			if !utils.DirExists(dir) {
				t.Errorf("Directory not created: %s", dir)
			}
		}
	})

	// Phase 4: Service Composition
	t.Run("Phase4_Compose", func(t *testing.T) {
		config := compose.DefaultConfig()
		config.HostIP = "192.168.1.100"
		config.InfraRoot = infraRoot
		config.DataRoot = tmpData
		config.AutoFillDefaults()

		// Validate config
		if err := config.Validate(); err != nil {
			t.Fatalf("Config validation failed: %v", err)
		}

		// Generate files
		composeDir := filepath.Join(infraRoot, "compose")
		err := compose.WriteAllConfigFiles(config, composeDir, false)
		if err != nil {
			t.Fatalf("WriteAllConfigFiles failed: %v", err)
		}

		// Verify files exist
		composePath := filepath.Join(composeDir, "docker-compose.yml")
		envPath := filepath.Join(composeDir, ".env")

		if !utils.FileExists(composePath) {
			t.Error("docker-compose.yml not created")
		}
		if !utils.FileExists(envPath) {
			t.Error(".env not created")
		}

		// Check docker-compose.yml content
		content, _ := os.ReadFile(composePath)
		if !strings.Contains(string(content), "immich") {
			t.Error("docker-compose.yml should contain immich")
		}
		if !strings.Contains(string(content), "nextcloud") {
			t.Error("docker-compose.yml should contain nextcloud")
		}

		// Check .env content
		envContent, _ := os.ReadFile(envPath)
		if !strings.Contains(string(envContent), "TZ=") {
			t.Error(".env should contain TZ")
		}
		if !strings.Contains(string(envContent), "DB_PASSWORD=") {
			t.Error(".env should contain DB_PASSWORD")
		}
	})

	// Phase 5: Maintenance
	t.Run("Phase5_Maintenance", func(t *testing.T) {
		mConfig := maintenance.DefaultScriptConfig()
		mConfig.DataRoot = tmpData
		mConfig.LogDir = filepath.Join(infraRoot, "logs")
		mConfig.InfraRoot = infraRoot
		mConfig.WebhookURL = "https://discord.com/api/webhooks/test/test"

		scriptsDir := filepath.Join(infraRoot, "scripts")
		scripts, err := maintenance.WriteAllScripts(mConfig, scriptsDir, false)
		if err != nil {
			t.Fatalf("WriteAllScripts failed: %v", err)
		}

		if len(scripts) != 4 {
			t.Errorf("Expected 4 scripts, got %d", len(scripts))
		}

		// Verify scripts exist and are executable
		expectedScripts := []string{
			"daily_backup.sh",
			"disk_alert.sh",
			"smart_alert.sh",
			"weekly_cleanup.sh",
		}

		for _, script := range expectedScripts {
			scriptPath := filepath.Join(scriptsDir, script)
			if !utils.FileExists(scriptPath) {
				t.Errorf("Script not created: %s", script)
			}

			// Check it's executable
			info, err := os.Stat(scriptPath)
			if err != nil {
				t.Errorf("Cannot stat script: %s", script)
				continue
			}
			if info.Mode().Perm()&0100 == 0 {
				t.Errorf("Script not executable: %s", script)
			}
		}
	})

	// Phase 8: Mission Report
	t.Run("Phase8_MissionReport", func(t *testing.T) {
		config := compose.DefaultConfig()
		config.HostIP = "192.168.1.100"
		config.NextcloudAdminPass = "testpass123"
		config.ImmichDBPassword = "immichdb123"
		config.NextcloudDBPassword = "ncdb123"
		config.DataRoot = tmpData

		missionReport := report.NewMissionReport(config, infraRoot)
		output := report.RenderMissionReport(missionReport)

		if output == "" {
			t.Error("Mission report is empty")
		}

		// Check essential sections
		essentials := []string{
			"192.168.1.100",
			"2283",
			"8080",
			"61208",
			"testpass123",
			"docker compose",
		}

		for _, essential := range essentials {
			if !strings.Contains(output, essential) {
				t.Errorf("Mission report missing: %s", essential)
			}
		}
	})
}

// TestIdempotency tests that running twice doesn't create duplicates
func TestIdempotency(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "servctl-idempotency")
	defer os.RemoveAll(tmpDir)

	// First run
	created1, err := utils.EnsureDir(tmpDir, 0755)
	if err != nil {
		t.Fatalf("First EnsureDir failed: %v", err)
	}
	if !created1 {
		t.Error("First run should create directory")
	}

	// Second run (should be idempotent)
	created2, err := utils.EnsureDir(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Second EnsureDir failed: %v", err)
	}
	if created2 {
		t.Error("Second run should NOT create directory (already exists)")
	}

	// Test AppendLineIfMissing
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte(""), 0644)

	added1, _ := utils.AppendLineIfMissing(testFile, "test line")
	if !added1 {
		t.Error("First append should add line")
	}

	added2, _ := utils.AppendLineIfMissing(testFile, "test line")
	if added2 {
		t.Error("Second append should NOT add line (already exists)")
	}

	// Verify file has only one occurrence
	content, _ := os.ReadFile(testFile)
	count := strings.Count(string(content), "test line")
	if count != 1 {
		t.Errorf("Line should appear exactly once, found %d times", count)
	}
}

// TestDirectoryIdempotency tests directory creation is idempotent
func TestDirectoryIdempotency(t *testing.T) {
	tmpHome := filepath.Join(os.TempDir(), "servctl-dir-idem")
	tmpData := filepath.Join(os.TempDir(), "servctl-data-idem")
	defer os.RemoveAll(tmpHome)
	defer os.RemoveAll(tmpData)

	// First run
	results1 := directory.CreateAllDirectories(tmpHome, tmpData, false)
	created1 := directory.CountCreated(results1)

	// Second run
	results2 := directory.CreateAllDirectories(tmpHome, tmpData, false)
	created2 := directory.CountCreated(results2)
	existing2 := directory.CountExisting(results2)

	t.Logf("Run 1: created=%d", created1)
	t.Logf("Run 2: created=%d, existing=%d", created2, existing2)

	// Second run should create nothing (all should exist)
	if created2 > 0 {
		t.Errorf("Second run created %d directories (should be 0)", created2)
	}

	// All should be marked as existing
	if existing2 != len(results2) {
		t.Errorf("Expected all directories to be marked existing, got %d/%d", existing2, len(results2))
	}
}

// TestComposeIdempotency tests compose file generation is idempotent
func TestComposeIdempotency(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "servctl-compose-idem")
	defer os.RemoveAll(tmpDir)

	config := compose.DefaultConfig()
	config.HostIP = "192.168.1.100"

	// First run
	err := compose.WriteAllConfigFiles(config, tmpDir, false)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Get file contents
	content1, _ := os.ReadFile(filepath.Join(tmpDir, "docker-compose.yml"))

	// Second run
	err = compose.WriteAllConfigFiles(config, tmpDir, false)
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	// Get file contents again
	content2, _ := os.ReadFile(filepath.Join(tmpDir, "docker-compose.yml"))

	// Should be identical
	if string(content1) != string(content2) {
		t.Error("Files should be identical on repeated runs")
	}
}

// TestErrorHandling tests error handling works correctly
func TestErrorHandling(t *testing.T) {
	t.Run("CriticalError", func(t *testing.T) {
		err := utils.NewCriticalError(
			"Storage",
			"Format disk",
			os.ErrPermission,
			"Run with sudo",
		)

		if !err.IsCritical {
			t.Error("Should be critical")
		}

		formatted := utils.FormatError(err)
		if !strings.Contains(formatted, "CRITICAL") {
			t.Error("Should contain CRITICAL")
		}
		if !strings.Contains(formatted, "sudo") {
			t.Error("Should contain remediation")
		}
	})

	t.Run("WarningError", func(t *testing.T) {
		err := utils.NewWarningError(
			"Preflight",
			"Check Docker",
			os.ErrNotExist,
		)

		if err.IsCritical {
			t.Error("Should not be critical")
		}

		formatted := utils.FormatError(err)
		if !strings.Contains(formatted, "WARNING") {
			t.Error("Should contain WARNING")
		}
	})
}

// TestConfigValidation tests input validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*compose.ServiceConfig)
		expectErr bool
	}{
		{
			name: "Valid default config",
			modify: func(c *compose.ServiceConfig) {
				c.AutoFillDefaults()
				c.NextcloudAdminPass = "securepassword123" // Need 8+ chars
			},
			expectErr: false,
		},
		{
			name:      "Empty timezone",
			modify:    func(c *compose.ServiceConfig) { c.Timezone = "" },
			expectErr: true,
		},
		{
			name:      "Invalid port",
			modify:    func(c *compose.ServiceConfig) { c.ImmichPort = 0 },
			expectErr: true,
		},
		{
			name: "Valid IP",
			modify: func(c *compose.ServiceConfig) {
				c.HostIP = "192.168.1.1"
				c.AutoFillDefaults()
				c.NextcloudAdminPass = "securepassword123"
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := compose.DefaultConfig()
			tt.modify(config)

			err := config.Validate()
			hasErr := err != nil

			if hasErr != tt.expectErr {
				if tt.expectErr {
					t.Errorf("Expected error but got nil")
				} else {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestPasswordGeneration tests password generation
func TestPasswordGeneration(t *testing.T) {
	passwords := make(map[string]bool)

	// Generate multiple passwords and ensure uniqueness
	for i := 0; i < 100; i++ {
		pass := compose.GeneratePassword(32)

		if len(pass) < 16 {
			t.Errorf("Password too short: %d chars", len(pass))
		}

		if passwords[pass] {
			t.Errorf("Duplicate password generated: %s", pass)
		}
		passwords[pass] = true
	}

	// Test DB password (alphanumeric only)
	for i := 0; i < 100; i++ {
		pass := compose.GenerateDBPassword()

		if len(pass) < 24 {
			t.Errorf("DB password too short: %d chars", len(pass))
		}

		// Should only contain alphanumeric
		for _, r := range pass {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				t.Errorf("DB password contains invalid char: %c", r)
			}
		}
	}
}

// TestIPValidation tests IP address validation
func TestIPValidation(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"8.8.8.8", false},       // Public IP
		{"1.1.1.1", false},       // Public IP
		{"192.168.1.256", false}, // Invalid
		{"not-an-ip", false},     // Invalid
		{"", false},              // Empty
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := compose.ValidateIP(tt.ip)
			isValid := err == nil
			if isValid != tt.valid {
				t.Errorf("ValidateIP(%q) = %v, want %v", tt.ip, isValid, tt.valid)
			}
		})
	}
}

// TestWebhookValidation tests webhook URL validation
func TestWebhookValidation(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://discord.com/api/webhooks/123/abc", true},
		{"https://hooks.slack.com/services/T00/B00/xxx", true},
		{"not-a-url", false},
		{"", true}, // Empty is valid (optional)
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := compose.ValidateWebhookURL(tt.url)
			isValid := err == nil
			if isValid != tt.valid {
				t.Errorf("ValidateWebhookURL(%q) = %v, want %v (err: %v)", tt.url, isValid, tt.valid, err)
			}
		})
	}
}

// TestTemplateGeneration tests Docker Compose template generation
func TestTemplateGeneration(t *testing.T) {
	config := compose.DefaultConfig()
	config.HostIP = "192.168.1.100"
	config.Timezone = "Asia/Kolkata"
	config.PUID = 1000
	config.PGID = 1000
	config.ImmichDBPassword = "testpass"
	config.NextcloudDBPassword = "ncpass"
	config.DataRoot = "/mnt/data"

	// Generate Docker Compose
	composeContent, err := compose.GenerateDockerCompose(config)
	if err != nil {
		t.Fatalf("GenerateDockerCompose failed: %v", err)
	}

	// Check for required services (adjust based on actual implementation)
	requiredServices := []string{
		"immich",
		"nextcloud",
		"postgres",
		"redis",
		"glances",
	}

	for _, svc := range requiredServices {
		if !strings.Contains(composeContent, svc) {
			t.Errorf("Missing service: %s", svc)
		}
	}

	// Generate ENV
	envContent, err := compose.GenerateEnvFile(config)
	if err != nil {
		t.Fatalf("GenerateEnvFile failed: %v", err)
	}

	// Check for required variables (adjust based on actual template)
	requiredVars := []string{
		"TZ=",
		"DB_PASSWORD=",
		"PUID=",
		"PGID=",
	}

	for _, v := range requiredVars {
		if !strings.Contains(envContent, v) {
			t.Errorf("Missing env var pattern: %s", v)
		}
	}
}

// Benchmark for full workflow
func BenchmarkFullWorkflow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmpHome := filepath.Join(os.TempDir(), "servctl-bench")
		tmpData := filepath.Join(os.TempDir(), "servctl-bench-data")

		// Directories
		directory.CreateAllDirectories(tmpHome, tmpData, true) // dry run

		// Compose
		config := compose.DefaultConfig()
		compose.GenerateDockerCompose(config)
		compose.GenerateEnvFile(config)

		// Maintenance
		mConfig := maintenance.DefaultScriptConfig()
		maintenance.GenerateAllScripts(mConfig)

		// Report
		missionReport := report.NewMissionReport(config, tmpHome)
		report.RenderMissionReport(missionReport)
	}
}
