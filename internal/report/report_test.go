package report

import (
	"strings"
	"testing"

	"github.com/madhav/servctl/internal/compose"
)

func TestNewMissionReport(t *testing.T) {
	config := compose.DefaultConfig()
	config.HostIP = "192.168.1.100"
	config.NextcloudAdminPass = "testpass123"
	config.ImmichDBPassword = "immichdb123"
	config.NextcloudDBPassword = "ncdb123"

	report := NewMissionReport(config, "/home/user/infra")

	if report == nil {
		t.Fatal("NewMissionReport returned nil")
	}

	if report.HostIP != "192.168.1.100" {
		t.Errorf("HostIP = %s, want 192.168.1.100", report.HostIP)
	}

	if report.ImmichURL != "http://192.168.1.100:2283" {
		t.Errorf("ImmichURL = %s, want http://192.168.1.100:2283", report.ImmichURL)
	}

	if report.NextcloudURL != "http://192.168.1.100:8080" {
		t.Errorf("NextcloudURL = %s, want http://192.168.1.100:8080", report.NextcloudURL)
	}

	if report.GlancesURL != "http://192.168.1.100:61208" {
		t.Errorf("GlancesURL = %s, want http://192.168.1.100:61208", report.GlancesURL)
	}

	if report.ComposeDir != "/home/user/infra/compose" {
		t.Errorf("ComposeDir = %s, want /home/user/infra/compose", report.ComposeDir)
	}
}

func TestRenderMissionReport(t *testing.T) {
	config := compose.DefaultConfig()
	config.HostIP = "192.168.1.100"
	config.NextcloudAdminUser = "admin"
	config.NextcloudAdminPass = "testpass123"
	config.ImmichDBPassword = "immichdb123"
	config.NextcloudDBPassword = "ncdb123"

	report := NewMissionReport(config, "/home/user/infra")
	output := RenderMissionReport(report)

	// Check for essential sections
	checks := []string{
		"SETUP COMPLETE",
		"Dashboard URLs",
		"Immich",
		"Nextcloud",
		"Glances",
		"192.168.1.100",
		"SAVE THESE CREDENTIALS",
		"Quick Start",
		"docker compose",
		"Next Steps",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Mission report should contain %q", check)
		}
	}
}

func TestRenderDashboardURLs(t *testing.T) {
	report := &MissionReport{
		ImmichURL:    "http://192.168.1.100:2283",
		NextcloudURL: "http://192.168.1.100:8080",
		GlancesURL:   "http://192.168.1.100:61208",
	}

	output := RenderDashboardURLs(report)

	if !strings.Contains(output, "2283") {
		t.Error("Should contain Immich port")
	}
	if !strings.Contains(output, "8080") {
		t.Error("Should contain Nextcloud port")
	}
	if !strings.Contains(output, "61208") {
		t.Error("Should contain Glances port")
	}
	if !strings.Contains(output, "Mobile app") {
		t.Error("Should mention mobile app")
	}
}

func TestRenderCredentials(t *testing.T) {
	report := &MissionReport{
		NextcloudAdminUser:  "admin",
		NextcloudAdminPass:  "secretpass",
		ImmichDBPassword:    "immichdb",
		NextcloudDBPassword: "ncdb",
		ComposeDir:          "/home/user/infra/compose",
	}

	output := RenderCredentials(report)

	if !strings.Contains(output, "SAVE THESE CREDENTIALS") {
		t.Error("Should warn about saving credentials")
	}
	if !strings.Contains(output, "admin") {
		t.Error("Should contain admin username")
	}
	if !strings.Contains(output, "secretpass") {
		t.Error("Should contain admin password")
	}
	if !strings.Contains(output, ".env") {
		t.Error("Should mention .env file")
	}
}

func TestRenderQuickStart(t *testing.T) {
	report := &MissionReport{
		ComposeDir: "/home/user/infra/compose",
	}

	output := RenderQuickStart(report)

	if !strings.Contains(output, "docker compose up -d") {
		t.Error("Should contain start command")
	}
	if !strings.Contains(output, "docker compose logs") {
		t.Error("Should contain logs command")
	}
	if !strings.Contains(output, "servctl -status") {
		t.Error("Should mention servctl status")
	}
}

func TestRenderNextSteps(t *testing.T) {
	report := &MissionReport{
		ImmichURL:  "http://192.168.1.100:2283",
		ComposeDir: "/home/user/infra/compose",
	}

	output := RenderNextSteps(report)

	if !strings.Contains(output, "Next Steps") {
		t.Error("Should have Next Steps header")
	}
	if !strings.Contains(output, "Static IP") {
		t.Error("Should mention static IP")
	}
	if !strings.Contains(output, "Change Passwords") {
		t.Error("Should explain how to change passwords")
	}
}

func TestRenderCompactReport(t *testing.T) {
	report := &MissionReport{
		ImmichURL:          "http://192.168.1.100:2283",
		NextcloudURL:       "http://192.168.1.100:8080",
		GlancesURL:         "http://192.168.1.100:61208",
		NextcloudAdminPass: "secretpass",
		ComposeDir:         "/home/user/infra/compose",
	}

	output := RenderCompactReport(report)

	if !strings.Contains(output, "Setup Complete") {
		t.Error("Should indicate completion")
	}
	if !strings.Contains(output, "Immich") {
		t.Error("Should mention Immich")
	}
	// Password should be partially masked
	if strings.Contains(output, "secretpass") {
		t.Error("Full password should not appear in compact report")
	}
}

// Benchmark
func BenchmarkRenderMissionReport(b *testing.B) {
	config := compose.DefaultConfig()
	config.HostIP = "192.168.1.100"
	config.NextcloudAdminPass = "testpass123"

	report := NewMissionReport(config, "/home/user/infra")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderMissionReport(report)
	}
}
