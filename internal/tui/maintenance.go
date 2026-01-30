package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/maintenance"
)

// Maintenance-specific styles
var (
	ScriptCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2).
			Width(55)

	CronBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#6366F1")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	ScheduleBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#10B981")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	LogPathStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)

// RenderMaintenanceIntro renders the introduction for maintenance setup
func RenderMaintenanceIntro() string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🔧 Maintenance & Reliability Setup") + "\n\n")

	b.WriteString("Servctl will configure automated maintenance tasks:\n\n")

	tasks := []struct {
		icon string
		name string
		desc string
	}{
		{"💾", "Daily Backup", "Sync data to backup drive with rsync"},
		{"📊", "Disk Alert", "Notify when disk usage exceeds threshold"},
		{"🔍", "SMART Check", "Monitor drive health"},
		{"🧹", "Weekly Cleanup", "Clean Docker, apt cache, and old logs"},
	}

	for _, task := range tasks {
		b.WriteString(fmt.Sprintf("  %s %s\n     %s\n\n",
			task.icon,
			TitleStyle.Render(task.name),
			DetailStyle.Render(task.desc),
		))
	}

	return b.String()
}

// RenderScriptPreview renders a preview of a maintenance script
func RenderScriptPreview(script maintenance.ScriptInfo) string {
	var b strings.Builder

	var card strings.Builder
	card.WriteString(fmt.Sprintf("%s\n", TitleStyle.Render(script.Name)))
	card.WriteString(fmt.Sprintf("%s\n\n", DetailStyle.Render(script.Description)))
	card.WriteString(fmt.Sprintf("Schedule: %s\n", ScheduleBadgeStyle.Render(script.Schedule)))
	card.WriteString(fmt.Sprintf("File: %s", LogPathStyle.Render(script.Filename)))

	b.WriteString(ScriptCardStyle.Render(card.String()))

	return b.String()
}

// RenderAllScripts renders all maintenance scripts summary
func RenderAllScripts(scripts []maintenance.ScriptInfo) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📜 Maintenance Scripts") + "\n\n")

	for _, script := range scripts {
		b.WriteString(RenderScriptPreview(script))
		b.WriteString("\n")
	}

	return b.String()
}

// RenderCronSchedule renders the cron schedule configuration
func RenderCronSchedule(jobs []maintenance.CronJob) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("⏰ Cron Schedule") + "\n\n")

	// Warning about root
	b.WriteString(WarningBannerStyle.Render(" ⚠️  ROOT REQUIRED ") + "\n\n")
	b.WriteString(WarnStyle.Render("These jobs require root privileges for:") + "\n")
	b.WriteString(DetailStyle.Render("  • Docker commands\n  • smartctl (raw disk access)\n  • System log access\n\n"))

	// Jobs table
	b.WriteString(TitleStyle.Render("Scheduled Jobs:") + "\n\n")

	for _, job := range jobs {
		schedule := CronBadgeStyle.Render(job.Schedule.String())
		humanTime := ScheduleBadgeStyle.Render(job.Schedule.HumanReadable())
		b.WriteString(fmt.Sprintf("  %s  %s\n", schedule, job.Name))
		b.WriteString(fmt.Sprintf("      %s\n", humanTime))
		b.WriteString(fmt.Sprintf("      %s\n\n", DetailStyle.Render(job.Description)))
	}

	// Cron file location
	b.WriteString(SectionStyle.Render("Installation:") + "\n")
	b.WriteString(fmt.Sprintf("  Cron file: %s\n", LogPathStyle.Render("/etc/cron.d/servctl")))

	return b.String()
}

// RenderScheduleInput renders schedule input for a specific job
func RenderScheduleInput(jobName string, currentSchedule maintenance.CronSchedule, options []string, selectedIndex int) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render(fmt.Sprintf("⏰ Configure: %s", jobName)) + "\n\n")
	b.WriteString(fmt.Sprintf("Current: %s\n\n", InputValueStyle.Render(currentSchedule.HumanReadable())))

	b.WriteString(TitleStyle.Render("Select schedule:") + "\n\n")

	for i, opt := range options {
		prefix := "  "
		if i == selectedIndex {
			prefix = "▶ "
			b.WriteString(InputValueStyle.Render(prefix+opt) + "\n")
		} else {
			b.WriteString(prefix + opt + "\n")
		}
	}

	b.WriteString("\n" + DetailStyle.Render("Use ↑/↓ to select, Enter to confirm") + "\n")

	return b.String()
}

// RenderWebhookTest renders webhook test result
func RenderWebhookTest(success bool, webhookURL string) string {
	var b strings.Builder

	if success {
		successBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2)

		b.WriteString(successBox.Render(
			PassStyle.Render("✅ Webhook Test Successful") + "\n\n" +
				"Check your Discord/Telegram for a test notification.\n" +
				"You will receive alerts here.",
		))
	} else {
		failBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2)

		b.WriteString(failBox.Render(
			FailStyle.Render("❌ Webhook Test Failed") + "\n\n" +
				"Please check:\n" +
				"  • Webhook URL is correct\n" +
				"  • Network connectivity\n" +
				"  • Webhook is not deleted",
		))
	}

	return b.String()
}

// RenderLogrotateConfig renders logrotate configuration info
func RenderLogrotateConfig(logDir string) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📋 Log Rotation Configuration") + "\n\n")

	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(1, 2)

	info := fmt.Sprintf(`
Log rotation will be configured for:
  %s/*.log

Settings:
  • Rotate: Weekly
  • Keep: 4 weeks of history
  • Compression: gzip

Config file: /etc/logrotate.d/servctl
`, logDir)

	b.WriteString(infoBox.Render(info))

	return b.String()
}

// RenderMaintenanceComplete renders the maintenance setup completion summary
func RenderMaintenanceComplete(scripts []maintenance.ScriptInfo, jobs []maintenance.CronJob, scriptsDir string) string {
	var b strings.Builder

	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(1, 2)

	b.WriteString(PassStyle.Render("✅ Maintenance Configuration Complete") + "\n\n")

	// Scripts
	b.WriteString(SectionStyle.Render("Generated Scripts:") + "\n")
	for _, script := range scripts {
		b.WriteString(fmt.Sprintf("  • %s (%s)\n", script.Filename, script.Schedule))
	}
	b.WriteString(fmt.Sprintf("\n  Location: %s\n\n", LogPathStyle.Render(scriptsDir)))

	// Cron jobs
	b.WriteString(SectionStyle.Render("Scheduled Jobs:") + "\n")
	for _, job := range jobs {
		b.WriteString(fmt.Sprintf("  • %s → %s\n", job.Name, job.Schedule.HumanReadable()))
	}
	b.WriteString("\n  Cron file: " + LogPathStyle.Render("/etc/cron.d/servctl") + "\n")

	return successBox.Render(b.String())
}

// RenderMaintenanceSummary renders a summary of all maintenance configuration
func RenderMaintenanceSummary(config *maintenance.ScriptConfig) string {
	var b strings.Builder

	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2)

	b.WriteString(SectionStyle.Render("📋 Maintenance Summary") + "\n\n")

	summary := fmt.Sprintf(`
Paths:
  Data Root:   %s
  Backup Dest: %s
  Log Dir:     %s

Monitoring:
  Disk Alert:  %d%% threshold
  SMART Check: %d drive(s)
  Retention:   %d days

Notifications:
  Webhook:     %s
`,
		config.DataRoot,
		config.BackupDest,
		config.LogDir,
		config.DiskAlertThreshold,
		len(config.Drives),
		config.BackupRetentionDays,
		maskWebhook(config.WebhookURL),
	)

	b.WriteString(summaryBox.Render(summary))

	return b.String()
}

// maskWebhook masks a webhook URL for display
func maskWebhook(url string) string {
	if url == "" {
		return DetailStyle.Render("Not configured")
	}
	// Show only first 30 chars
	if len(url) > 35 {
		return PassStyle.Render(url[:30] + "...")
	}
	return PassStyle.Render(url)
}
