package compose

import (
	"testing"
)

// =============================================================================
// RenderConfigPreview Tests
// =============================================================================

func TestRenderConfigPreview_OutputFormat(t *testing.T) {
	config := &ServiceConfig{
		HostIP:        "192.168.1.100",
		Timezone:      "America/New_York",
		DataRoot:      "/mnt/data",
		NextcloudPort: 8080,
		ImmichPort:    2283,
		GlancesPort:   61208,
	}

	preview := RenderConfigPreview(config)

	// Check for key elements
	if !containsString(preview, "Configuration Preview") {
		t.Error("Preview should contain header")
	}
	if !containsString(preview, "192.168.1.100") {
		t.Error("Preview should contain host IP")
	}
	if !containsString(preview, "America/New_York") {
		t.Error("Preview should contain timezone")
	}
	if !containsString(preview, "/mnt/data") {
		t.Error("Preview should contain data root")
	}
	if !containsString(preview, "8080") {
		t.Error("Preview should contain Nextcloud port")
	}
	if !containsString(preview, "2283") {
		t.Error("Preview should contain Immich port")
	}
	if !containsString(preview, "61208") {
		t.Error("Preview should contain Glances port")
	}
}

func TestRenderConfigPreview_Structure(t *testing.T) {
	config := DefaultConfig()
	config.AutoFillDefaults()

	preview := RenderConfigPreview(config)

	// Should have visual structure
	if !containsString(preview, "┌") {
		t.Error("Preview should have box border")
	}
	if !containsString(preview, "Service Ports") {
		t.Error("Preview should have Service Ports section")
	}
}

// =============================================================================
// DefaultConfig Tests
// =============================================================================

func TestDefaultConfig_Selection(t *testing.T) {
	config := DefaultConfig()

	if config.PUID != 1000 {
		t.Errorf("Expected PUID 1000, got %d", config.PUID)
	}
	if config.PGID != 1000 {
		t.Errorf("Expected PGID 1000, got %d", config.PGID)
	}
}

func TestDefaultConfig_AutoFill(t *testing.T) {
	config := DefaultConfig()
	config.AutoFillDefaults()

	if config.NextcloudPort == 0 {
		t.Error("AutoFillDefaults should set NextcloudPort")
	}
	if config.ImmichPort == 0 {
		t.Error("AutoFillDefaults should set ImmichPort")
	}
	if config.GlancesPort == 0 {
		t.Error("AutoFillDefaults should set GlancesPort")
	}
	if config.Timezone == "" {
		t.Error("AutoFillDefaults should set Timezone")
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestValidateIP_Valid(t *testing.T) {
	validIPs := []string{
		"192.168.1.100",
		"10.0.0.1",
		"172.16.0.1",
	}

	for _, ip := range validIPs {
		err := ValidateIP(ip)
		if err != nil {
			t.Errorf("ValidateIP(%s) should be valid, got error: %v", ip, err)
		}
	}
}

func TestValidateIP_Invalid(t *testing.T) {
	invalidIPs := []string{
		"",
		"not-an-ip",
		"256.256.256.256",
	}

	for _, ip := range invalidIPs {
		err := ValidateIP(ip)
		if err == nil {
			t.Errorf("ValidateIP(%s) should be invalid", ip)
		}
	}
}

func TestValidatePassword_Selection(t *testing.T) {
	tests := []struct {
		password  string
		minLength int
		valid     bool
	}{
		{"password123", 8, true},
		{"short", 8, false},
		{"", 1, false},
		{"a", 1, true},
	}

	for _, tt := range tests {
		err := ValidatePassword(tt.password, tt.minLength)
		if (err == nil) != tt.valid {
			t.Errorf("ValidatePassword(%s, %d) valid=%v, want %v", tt.password, tt.minLength, err == nil, tt.valid)
		}
	}
}

// =============================================================================
// Password Generation Tests
// =============================================================================

func TestGeneratePassword_Length(t *testing.T) {
	// GeneratePassword enforces minimum of 16 characters
	lengths := []int{16, 24, 32}

	for _, length := range lengths {
		password := GeneratePassword(length)
		if len(password) != length {
			t.Errorf("GeneratePassword(%d) returned length %d", length, len(password))
		}
	}

	// Test that values below 16 get promoted to 16
	password := GeneratePassword(8)
	if len(password) != 16 {
		t.Errorf("GeneratePassword(8) should return length 16 (minimum), got %d", len(password))
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	passwords := make(map[string]bool)

	for i := 0; i < 10; i++ {
		password := GeneratePassword(16)
		if passwords[password] {
			t.Error("GeneratePassword should generate unique passwords")
		}
		passwords[password] = true
	}
}

func TestGenerateDBPassword_AlphaNumeric(t *testing.T) {
	password := GenerateDBPassword()

	// Should only contain alphanumeric characters
	for _, char := range password {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			t.Errorf("GenerateDBPassword should be alphanumeric only, got character: %c", char)
		}
	}

	if len(password) == 0 {
		t.Error("GenerateDBPassword should not return empty string")
	}
}

// =============================================================================
// Network Detection Tests
// =============================================================================

func TestDetectHostIP_Selection(t *testing.T) {
	ip, err := DetectHostIP()

	// This test may fail if there's no network interface
	if err != nil {
		t.Logf("DetectHostIP returned error (may be expected): %v", err)
		return
	}

	if ip == "" {
		t.Error("DetectHostIP should return an IP address")
	}

	// Should be valid IP
	if err := ValidateIP(ip); err != nil {
		t.Errorf("DetectHostIP returned invalid IP: %s", ip)
	}
}

// =============================================================================
// Timezone Tests
// =============================================================================

func TestGetTimezoneOptions_Selection(t *testing.T) {
	options := GetTimezoneOptions()

	if len(options) == 0 {
		t.Error("GetTimezoneOptions should return timezone options")
	}

	// Check for common timezones
	hasUTC := false
	for _, tz := range options {
		if tz == "UTC" {
			hasUTC = true
		}
	}
	if !hasUTC {
		t.Error("GetTimezoneOptions should include UTC")
	}
}

// =============================================================================
// ServiceConfig Validation Tests
// =============================================================================

func TestServiceConfig_Validate_Valid(t *testing.T) {
	config := DefaultConfig()
	config.AutoFillDefaults()
	config.HostIP = "192.168.1.100"
	config.DataRoot = "/mnt/data"
	config.InfraRoot = "/home/user/infra"
	config.Timezone = "UTC"
	// Ensure all required passwords are set
	config.ImmichDBPassword = GeneratePassword(16)
	config.NextcloudDBPassword = GeneratePassword(16)
	config.NextcloudAdminPass = GeneratePassword(16)
	config.NextcloudAdminUser = "admin"
	config.DiscordWebhookURL = "" // Empty is valid (optional)

	errors := config.Validate()

	if len(errors) > 0 {
		t.Errorf("Valid config should have no errors, got: %v", errors)
	}
}

func TestServiceConfig_Validate_MissingPassword(t *testing.T) {
	config := DefaultConfig()
	config.AutoFillDefaults()
	// Missing required passwords
	config.Timezone = "UTC"
	config.NextcloudAdminPass = ""

	errors := config.Validate()

	hasPassError := false
	for _, err := range errors {
		if containsString(err.Error(), "password") {
			hasPassError = true
		}
	}
	if !hasPassError {
		t.Error("Validation should report missing/short password")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
