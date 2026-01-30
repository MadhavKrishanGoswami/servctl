package compose

import (
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check sensible defaults
	if config.PUID != 1000 {
		t.Errorf("PUID = %d, want 1000", config.PUID)
	}
	if config.PGID != 1000 {
		t.Errorf("PGID = %d, want 1000", config.PGID)
	}
	if config.DataRoot != "/mnt/data" {
		t.Errorf("DataRoot = %s, want /mnt/data", config.DataRoot)
	}
	if config.ImmichPort != 2283 {
		t.Errorf("ImmichPort = %d, want 2283", config.ImmichPort)
	}
	if config.NextcloudPort != 8080 {
		t.Errorf("NextcloudPort = %d, want 8080", config.NextcloudPort)
	}
	if config.GlancesPort != 61208 {
		t.Errorf("GlancesPort = %d, want 61208", config.GlancesPort)
	}
}

func TestGeneratePassword(t *testing.T) {
	// Generate multiple passwords and check uniqueness
	passwords := make(map[string]bool)

	for i := 0; i < 10; i++ {
		pass := GeneratePassword(16)

		if len(pass) < 16 {
			t.Errorf("Password too short: %d chars", len(pass))
		}

		if passwords[pass] {
			t.Error("Generated duplicate password")
		}
		passwords[pass] = true
	}
}

func TestGenerateDBPassword(t *testing.T) {
	pass := GenerateDBPassword()

	if len(pass) != 24 {
		t.Errorf("DB password length = %d, want 24", len(pass))
	}

	// Check alphanumeric only
	for _, c := range pass {
		if !isAlphanumeric(c) {
			t.Errorf("DB password contains non-alphanumeric: %c", c)
		}
	}
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		ip      string
		wantErr bool
	}{
		{"192.168.1.100", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"8.8.8.8", true},       // Public IP
		{"invalid", true},       // Invalid format
		{"192.168.1.256", true}, // Invalid octet
		{"", true},              // Empty
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := ValidateIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIP(%s) error = %v, wantErr = %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"", false}, // Empty is valid (optional)
		{"https://discord.com/api/webhooks/123456789/abcdef", false},
		{"https://hooks.slack.com/services/T00/B00/xxx", false},
		{"http://discord.com/api/webhooks/123", true}, // HTTP not allowed
		{"not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWebhookURL(%s) error = %v, wantErr = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password  string
		minLength int
		wantErr   bool
	}{
		{"password123", 8, false},
		{"short", 8, true},
		{"exactly8", 8, false},
		{"", 8, true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := ValidatePassword(tt.password, tt.minLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceConfigValidate(t *testing.T) {
	config := DefaultConfig()
	config.Timezone = "UTC"
	config.HostIP = "192.168.1.100"
	config.ImmichDBPassword = "password123456"
	config.NextcloudDBPassword = "password123456"
	config.NextcloudAdminUser = "admin"
	config.NextcloudAdminPass = "adminpass123"

	errors := config.Validate()

	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Validation error: %v", err)
		}
	}
}

func TestServiceConfigValidateErrors(t *testing.T) {
	config := &ServiceConfig{
		Timezone:            "",        // Missing
		HostIP:              "8.8.8.8", // Public IP
		ImmichDBPassword:    "short",   // Too short
		NextcloudDBPassword: "short",   // Too short
		NextcloudAdminUser:  "",        // Missing
		NextcloudAdminPass:  "short",   // Too short
	}

	errors := config.Validate()

	if len(errors) == 0 {
		t.Error("Expected validation errors, got none")
	}

	// Should have at least 5 errors
	if len(errors) < 5 {
		t.Errorf("Expected at least 5 errors, got %d", len(errors))
	}
}

func TestAutoFillDefaults(t *testing.T) {
	config := &ServiceConfig{}
	config.AutoFillDefaults()

	if config.PUID == 0 {
		t.Error("AutoFillDefaults did not set PUID")
	}
	if config.PGID == 0 {
		t.Error("AutoFillDefaults did not set PGID")
	}
	if config.DataRoot == "" {
		t.Error("AutoFillDefaults did not set DataRoot")
	}
	if config.ImmichDBPassword == "" {
		t.Error("AutoFillDefaults did not generate ImmichDBPassword")
	}
	if config.NextcloudDBPassword == "" {
		t.Error("AutoFillDefaults did not generate NextcloudDBPassword")
	}
}

func TestGetTimezoneOptions(t *testing.T) {
	options := GetTimezoneOptions()

	if len(options) == 0 {
		t.Error("GetTimezoneOptions() returned empty list")
	}

	// Should include UTC
	hasUTC := false
	for _, tz := range options {
		if tz == "UTC" {
			hasUTC = true
			break
		}
	}
	if !hasUTC {
		t.Error("GetTimezoneOptions() should include UTC")
	}
}

func TestGenerateDockerCompose(t *testing.T) {
	config := DefaultConfig()
	config.Timezone = "UTC"
	config.HostIP = "192.168.1.100"
	config.ImmichDBPassword = "testpass123"
	config.NextcloudDBPassword = "testpass456"
	config.NextcloudAdminUser = "admin"
	config.NextcloudAdminPass = "adminpass"

	content, err := GenerateDockerCompose(config)

	if err != nil {
		t.Fatalf("GenerateDockerCompose() error: %v", err)
	}

	// Check for expected services
	expectedServices := []string{
		"immich-server:",
		"immich-machine-learning:",
		"immich-redis:",
		"immich-postgres:",
		"nextcloud:",
		"nextcloud-mariadb:",
		"glances:",
		"diun:",
	}

	for _, svc := range expectedServices {
		if !strings.Contains(content, svc) {
			t.Errorf("Docker Compose missing service: %s", svc)
		}
	}

	// Check for network
	if !strings.Contains(content, "servctl-network:") {
		t.Error("Docker Compose missing servctl-network")
	}
}

func TestGenerateEnvFile(t *testing.T) {
	config := DefaultConfig()
	config.Timezone = "Asia/Kolkata"
	config.HostIP = "192.168.1.100"
	config.ImmichDBPassword = "testpass"
	config.NextcloudDBPassword = "ncpass"

	content, err := GenerateEnvFile(config)

	if err != nil {
		t.Fatalf("GenerateEnvFile() error: %v", err)
	}

	// Check for expected variables
	expectedVars := []string{
		"TZ=Asia/Kolkata",
		"PUID=1000",
		"PGID=1000",
		"HOST_IP=192.168.1.100",
		"DATA_ROOT=/mnt/data",
	}

	for _, v := range expectedVars {
		if !strings.Contains(content, v) {
			t.Errorf("ENV file missing variable: %s", v)
		}
	}
}

func TestGetDefaultFirewallRules(t *testing.T) {
	rules := GetDefaultFirewallRules()

	if len(rules) == 0 {
		t.Fatal("GetDefaultFirewallRules() returned empty")
	}

	// SSH should be first and required
	if rules[0].Port != 22 {
		t.Error("First rule should be SSH (port 22)")
	}
	if !rules[0].Required {
		t.Error("SSH rule should be required")
	}

	// Check for required services
	ports := make(map[int]bool)
	for _, rule := range rules {
		ports[rule.Port] = true
	}

	expectedPorts := []int{22, 2283, 8080, 61208}
	for _, port := range expectedPorts {
		if !ports[port] {
			t.Errorf("Missing firewall rule for port %d", port)
		}
	}
}

func TestDetectHostIP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	ip, err := DetectHostIP()

	// This may fail in CI/containers, so we just check it doesn't panic
	if err != nil {
		t.Logf("DetectHostIP() note: %v (may be expected in container)", err)
		return
	}

	if ip == "" {
		t.Error("DetectHostIP() returned empty IP")
	}

	// Validate the returned IP
	if err := ValidateIP(ip); err != nil {
		// Public IPs are okay for detection
		t.Logf("Detected IP: %s (validation: %v)", ip, err)
	} else {
		t.Logf("Detected IP: %s (private range)", ip)
	}
}

// Benchmark tests
func BenchmarkGenerateDockerCompose(b *testing.B) {
	config := DefaultConfig()
	config.Timezone = "UTC"
	config.HostIP = "192.168.1.100"
	config.ImmichDBPassword = "testpass"
	config.NextcloudDBPassword = "testpass"
	config.NextcloudAdminUser = "admin"
	config.NextcloudAdminPass = "adminpass"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateDockerCompose(config)
	}
}

func BenchmarkGeneratePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GeneratePassword(24)
	}
}
