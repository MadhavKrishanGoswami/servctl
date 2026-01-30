package maintenance

import (
	"strings"
	"testing"
)

func TestDefaultScriptConfig(t *testing.T) {
	config := DefaultScriptConfig()

	if config == nil {
		t.Fatal("DefaultScriptConfig() returned nil")
	}

	if config.DataRoot != "/mnt/data" {
		t.Errorf("DataRoot = %s, want /mnt/data", config.DataRoot)
	}

	if config.DiskAlertThreshold != 90 {
		t.Errorf("DiskAlertThreshold = %d, want 90", config.DiskAlertThreshold)
	}

	if config.BackupRetentionDays != 7 {
		t.Errorf("BackupRetentionDays = %d, want 7", config.BackupRetentionDays)
	}

	if len(config.Drives) == 0 {
		t.Error("Drives should have at least one default")
	}
}

func TestGenerateDailyBackup(t *testing.T) {
	config := &ScriptConfig{
		DataRoot:   "/mnt/data",
		BackupDest: "/mnt/backup",
		LogDir:     "/home/user/infra/logs",
		WebhookURL: "https://discord.com/api/webhooks/123/abc",
	}

	content, err := GenerateDailyBackup(config)

	if err != nil {
		t.Fatalf("GenerateDailyBackup() error: %v", err)
	}

	// Check for expected content
	expectedParts := []string{
		"#!/bin/bash",
		"/mnt/data",
		"/mnt/backup",
		"rsync -av --delete",
		"NAS Guardian",
		"curl",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("Daily backup script missing: %s", part)
		}
	}
}

func TestGenerateDiskAlert(t *testing.T) {
	config := &ScriptConfig{
		DataRoot:           "/mnt/data",
		DiskAlertThreshold: 85,
		WebhookURL:         "https://discord.com/api/webhooks/123/abc",
	}

	content, err := GenerateDiskAlert(config)

	if err != nil {
		t.Fatalf("GenerateDiskAlert() error: %v", err)
	}

	expectedParts := []string{
		"#!/bin/bash",
		"THRESHOLD=85",
		"/mnt/data",
		"DISK FULL",
		"df -h",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("Disk alert script missing: %s", part)
		}
	}
}

func TestGenerateSmartAlert(t *testing.T) {
	config := &ScriptConfig{
		Drives:     []string{"/dev/sda", "/dev/sdb"},
		WebhookURL: "https://discord.com/api/webhooks/123/abc",
	}

	content, err := GenerateSmartAlert(config)

	if err != nil {
		t.Fatalf("GenerateSmartAlert() error: %v", err)
	}

	expectedParts := []string{
		"#!/bin/bash",
		"/dev/sda",
		"/dev/sdb",
		"smartctl",
		"DRIVE FAILURE",
		"Disk Doctor",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("SMART alert script missing: %s", part)
		}
	}
}

func TestGenerateWeeklyCleanup(t *testing.T) {
	config := &ScriptConfig{
		DataRoot:            "/mnt/data",
		BackupDest:          "/mnt/backup",
		LogDir:              "/home/user/infra/logs",
		BackupRetentionDays: 14,
		WebhookURL:          "https://discord.com/api/webhooks/123/abc",
	}

	content, err := GenerateWeeklyCleanup(config)

	if err != nil {
		t.Fatalf("GenerateWeeklyCleanup() error: %v", err)
	}

	expectedParts := []string{
		"#!/bin/bash",
		"apt-get clean",
		"docker image prune",
		"Janitor",
		"Weekly Cleanup Complete",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("Weekly cleanup script missing: %s", part)
		}
	}
}

func TestGenerateAllScripts(t *testing.T) {
	config := &ScriptConfig{
		DataRoot:            "/mnt/data",
		BackupDest:          "/mnt/backup",
		LogDir:              "/home/user/infra/logs",
		Drives:              []string{"/dev/sda"},
		DiskAlertThreshold:  90,
		BackupRetentionDays: 7,
		WebhookURL:          "https://discord.com/api/webhooks/123/abc",
	}

	scripts, err := GenerateAllScripts(config)

	if err != nil {
		t.Fatalf("GenerateAllScripts() error: %v", err)
	}

	if len(scripts) != 4 {
		t.Errorf("GenerateAllScripts() returned %d scripts, want 4", len(scripts))
	}

	expectedScripts := []string{
		"daily_backup.sh",
		"disk_alert.sh",
		"smart_alert.sh",
		"weekly_cleanup.sh",
	}

	for _, expected := range expectedScripts {
		found := false
		for _, script := range scripts {
			if script.Filename == expected {
				found = true
				if script.Content == "" {
					t.Errorf("Script %s has empty content", expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("Missing script: %s", expected)
		}
	}
}

func TestCronScheduleString(t *testing.T) {
	tests := []struct {
		schedule CronSchedule
		expected string
	}{
		{
			CronSchedule{"0", "4", "*", "*", "*"},
			"0 4 * * *",
		},
		{
			CronSchedule{"30", "*/6", "*", "*", "*"},
			"30 */6 * * *",
		},
		{
			CronSchedule{"0", "3", "*", "*", "0"},
			"0 3 * * 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.schedule.String(); got != tt.expected {
				t.Errorf("CronSchedule.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCronScheduleHumanReadable(t *testing.T) {
	tests := []struct {
		schedule CronSchedule
		contains string
	}{
		{
			CronSchedule{"0", "4", "*", "*", "*"},
			"Daily at 4:00",
		},
		{
			CronSchedule{"0", "3", "*", "*", "0"},
			"Sunday",
		},
		{
			CronSchedule{"0", "*/6", "*", "*", "*"},
			"Every 6 hours",
		},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			result := tt.schedule.HumanReadable()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("HumanReadable() = %v, should contain %v", result, tt.contains)
			}
		})
	}
}

func TestDefaultCronJobs(t *testing.T) {
	jobs := DefaultCronJobs("/home/user/infra/scripts")

	if len(jobs) != 4 {
		t.Errorf("DefaultCronJobs() returned %d jobs, want 4", len(jobs))
	}

	expectedJobs := []string{
		"daily_backup",
		"disk_alert",
		"smart_alert",
		"weekly_cleanup",
	}

	for _, expected := range expectedJobs {
		found := false
		for _, job := range jobs {
			if job.Name == expected {
				found = true
				if job.User != "root" {
					t.Errorf("Job %s should run as root", expected)
				}
				if job.Command == "" {
					t.Errorf("Job %s has empty command", expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("Missing cron job: %s", expected)
		}
	}
}

func TestGenerateCronFile(t *testing.T) {
	jobs := DefaultCronJobs("/home/user/infra/scripts")

	content, err := GenerateCronFile(jobs)

	if err != nil {
		t.Fatalf("GenerateCronFile() error: %v", err)
	}

	expectedParts := []string{
		"SHELL=/bin/bash",
		"PATH=",
		"daily_backup.sh",
		"disk_alert.sh",
		"smart_alert.sh",
		"weekly_cleanup.sh",
		"root",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("Cron file missing: %s", part)
		}
	}
}

func TestGenerateLogrotateConfig(t *testing.T) {
	content := GenerateLogrotateConfig("/home/user/infra/logs", "madhav")

	expectedParts := []string{
		"/home/user/infra/logs/*.log",
		"weekly",
		"rotate 4",
		"compress",
		"madhav",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("Logrotate config missing: %s", part)
		}
	}
}

func TestScriptInfoStructure(t *testing.T) {
	info := ScriptInfo{
		Name:        "Test Script",
		Filename:    "test.sh",
		Description: "A test script",
		Schedule:    "Daily",
		Content:     "#!/bin/bash\necho test",
	}

	if info.Name != "Test Script" {
		t.Errorf("Name = %s, want Test Script", info.Name)
	}
	if info.Filename != "test.sh" {
		t.Errorf("Filename = %s, want test.sh", info.Filename)
	}
}

func TestCronJobStructure(t *testing.T) {
	job := CronJob{
		Name:        "test_job",
		Schedule:    CronSchedule{"0", "4", "*", "*", "*"},
		Command:     "/path/to/script.sh",
		Description: "Test job",
		User:        "root",
	}

	if job.Name != "test_job" {
		t.Errorf("Name = %s, want test_job", job.Name)
	}
	if job.User != "root" {
		t.Errorf("User = %s, want root", job.User)
	}
	if job.Schedule.String() != "0 4 * * *" {
		t.Errorf("Schedule = %s, want 0 4 * * *", job.Schedule.String())
	}
}

// Test without webhook (scripts should still generate)
func TestGenerateScriptsWithoutWebhook(t *testing.T) {
	config := &ScriptConfig{
		DataRoot:   "/mnt/data",
		BackupDest: "/mnt/backup",
		LogDir:     "/home/user/logs",
		Drives:     []string{"/dev/sda"},
		WebhookURL: "", // No webhook
	}

	scripts, err := GenerateAllScripts(config)

	if err != nil {
		t.Fatalf("GenerateAllScripts() without webhook error: %v", err)
	}

	if len(scripts) != 4 {
		t.Errorf("Should still generate 4 scripts without webhook")
	}

	// Check that curl is NOT in the output (no webhook)
	for _, script := range scripts {
		if strings.Contains(script.Content, "curl -s -H") {
			t.Errorf("Script %s contains curl when webhook is empty", script.Filename)
		}
	}
}

// Benchmark
func BenchmarkGenerateAllScripts(b *testing.B) {
	config := DefaultScriptConfig()
	config.WebhookURL = "https://discord.com/api/webhooks/123/abc"
	config.LogDir = "/home/user/logs"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateAllScripts(config)
	}
}
