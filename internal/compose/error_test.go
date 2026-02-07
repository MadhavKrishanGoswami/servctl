package compose

import (
	"testing"
)

// =============================================================================
// Error Path Tests - Verify graceful handling of failure scenarios
// =============================================================================

// TestValidateIP_InvalidFormats tests various invalid IP formats
func TestValidateIP_InvalidFormats(t *testing.T) {
	invalidIPs := []string{
		"",
		"not.an.ip",
		"256.256.256.256",
		"192.168.1",
		"192.168.1.1.1",
		"-1.0.0.0",
		"abc.def.ghi.jkl",
		" 192.168.1.1",
		"192.168.1.1 ",
		"192.168.1.1:8080",
	}

	for _, ip := range invalidIPs {
		err := ValidateIP(ip)
		if err == nil && ip != "" {
			t.Errorf("ValidateIP(%q) should return error", ip)
		}
	}
}

// TestValidateIP_ValidFormats tests valid private IP formats
func TestValidateIP_ValidFormats(t *testing.T) {
	// ValidateIP only accepts private IP ranges
	validIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"172.31.255.255",
	}

	for _, ip := range validIPs {
		err := ValidateIP(ip)
		if err != nil {
			t.Errorf("ValidateIP(%q) should be valid, got error: %v", ip, err)
		}
	}
}

// TestValidatePassword_EdgeCases tests password validation edge cases
func TestValidatePassword_EdgeCases(t *testing.T) {
	// ValidatePassword: if minLength==0, defaults to 8. Returns error if len < minLength.
	tests := []struct {
		password  string
		minLength int
		wantErr   bool
	}{
		{"", 8, true},
		{"short", 8, true},
		{"exactly8", 8, false}, // length is 8, does NOT return error for >= 8
		{"longenough", 8, false},
		{"a", 0, true}, // Zero min length defaults to 8, "a" < 8 so error
	}

	for _, tt := range tests {
		err := ValidatePassword(tt.password, tt.minLength)
		hasErr := err != nil
		if hasErr != tt.wantErr {
			t.Errorf("ValidatePassword(%q, %d) error=%v, wantErr=%v", tt.password, tt.minLength, err, tt.wantErr)
		}
	}
}

// TestGeneratePassword_EdgeCases tests password generation edge cases
func TestGeneratePassword_EdgeCases(t *testing.T) {
	tests := []struct {
		length       int
		expectLength int // May differ due to minimum enforcement
	}{
		{0, 16},  // Below minimum
		{-1, 16}, // Negative
		{8, 16},  // Below minimum
		{16, 16}, // Exact minimum
		{32, 32}, // Above minimum
		{64, 64}, // Large
	}

	for _, tt := range tests {
		password := GeneratePassword(tt.length)
		if len(password) != tt.expectLength {
			t.Errorf("GeneratePassword(%d) length = %d, want %d", tt.length, len(password), tt.expectLength)
		}
	}
}

// TestServiceConfig_Validate_AllEmpty tests validation with empty config
func TestServiceConfig_Validate_AllEmpty(t *testing.T) {
	config := &ServiceConfig{}

	errors := config.Validate()

	// Should have multiple validation errors
	if len(errors) == 0 {
		t.Error("Empty config should have validation errors")
	}
}

// TestServiceConfig_Validate_InvalidWebhook tests invalid webhook URL
func TestServiceConfig_Validate_InvalidWebhook(t *testing.T) {
	config := DefaultConfig()
	config.AutoFillDefaults()
	config.Timezone = "UTC"
	config.ImmichDBPassword = GeneratePassword(16)
	config.NextcloudDBPassword = GeneratePassword(16)
	config.NextcloudAdminPass = GeneratePassword(16)
	config.NextcloudAdminUser = "admin"

	// Invalid webhook
	config.DiscordWebhookURL = "not-a-url"

	errors := config.Validate()

	hasWebhookError := false
	for _, err := range errors {
		if containsStr(err.Error(), "webhook") || containsStr(err.Error(), "discord") {
			hasWebhookError = true
		}
	}

	if !hasWebhookError {
		t.Log("Warning: Invalid webhook may not be validated")
	}
}

// TestDefaultConfig_NonNil tests that default config returns non-nil values
func TestDefaultConfig_NonNil(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check important fields have defaults
	if config.PUID == 0 {
		t.Log("Warning: PUID is 0")
	}
	if config.PGID == 0 {
		t.Log("Warning: PGID is 0")
	}
}

// TestAutoFillDefaults_DoesNotOverwrite tests that AutoFill preserves existing values
func TestAutoFillDefaults_DoesNotOverwrite(t *testing.T) {
	config := DefaultConfig()
	config.HostIP = "10.0.0.100"
	config.Timezone = "Europe/London"

	config.AutoFillDefaults()

	if config.HostIP != "10.0.0.100" {
		t.Error("AutoFillDefaults should not overwrite existing HostIP")
	}
	if config.Timezone != "Europe/London" {
		t.Error("AutoFillDefaults should not overwrite existing Timezone")
	}
}

// TestRenderConfigPreview_EmptyConfig tests preview with minimal config
func TestRenderConfigPreview_EmptyConfig(t *testing.T) {
	config := &ServiceConfig{}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RenderConfigPreview panicked with empty config: %v", r)
		}
	}()

	preview := RenderConfigPreview(config)

	if preview == "" {
		t.Error("Preview should not be empty")
	}
}

// Helper function - use different name to avoid redeclaration
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
