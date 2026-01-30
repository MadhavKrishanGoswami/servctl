// Package report handles mission report generation and final output display.
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/compose"
)

// Colors
var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorError     = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorHighlight = lipgloss.Color("#3B82F6")
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	URLStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Underline(true)

	CredentialStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#F9FAFB")).
			Padding(0, 1)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2)

	CredentialBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorWarning).
				Padding(1, 2)
)

// MissionReport contains all information for the final report
type MissionReport struct {
	// System info
	HostIP   string
	Timezone string
	PUID     int
	PGID     int

	// Service URLs
	ImmichURL    string
	NextcloudURL string
	GlancesURL   string

	// Credentials
	NextcloudAdminUser  string
	NextcloudAdminPass  string
	ImmichDBPassword    string
	NextcloudDBPassword string

	// Paths
	InfraRoot  string
	ComposeDir string
	ScriptsDir string
	DataRoot   string

	// Stats
	Duration    time.Duration
	DirsCreated int
	ScriptsGen  int
}

// NewMissionReport creates a mission report from config
func NewMissionReport(config *compose.ServiceConfig, infraRoot string) *MissionReport {
	return &MissionReport{
		HostIP:              config.HostIP,
		Timezone:            config.Timezone,
		PUID:                config.PUID,
		PGID:                config.PGID,
		ImmichURL:           fmt.Sprintf("http://%s:%d", config.HostIP, config.ImmichPort),
		NextcloudURL:        fmt.Sprintf("http://%s:%d", config.HostIP, config.NextcloudPort),
		GlancesURL:          fmt.Sprintf("http://%s:%d", config.HostIP, config.GlancesPort),
		NextcloudAdminUser:  config.NextcloudAdminUser,
		NextcloudAdminPass:  config.NextcloudAdminPass,
		ImmichDBPassword:    config.ImmichDBPassword,
		NextcloudDBPassword: config.NextcloudDBPassword,
		InfraRoot:           infraRoot,
		ComposeDir:          infraRoot + "/compose",
		ScriptsDir:          infraRoot + "/scripts",
		DataRoot:            config.DataRoot,
	}
}

// RenderMissionReport generates the complete mission report
func RenderMissionReport(report *MissionReport) string {
	var b strings.Builder

	// Success banner
	successBanner := `
╔═══════════════════════════════════════════════════════════════════════╗
║                                                                       ║
║   ██████╗ ███████╗██████╗ ██╗   ██╗ ██████╗████████╗██╗               ║
║  ██╔════╝ ██╔════╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝██║               ║
║  ╚█████╗  █████╗  ██████╔╝██║   ██║██║        ██║   ██║               ║
║   ╚═══██╗ ██╔══╝  ██╔══██╗╚██╗ ██╔╝██║        ██║   ██║               ║
║  ██████╔╝ ███████╗██║  ██║ ╚████╔╝ ╚██████╗   ██║   ███████╗          ║
║  ╚═════╝  ╚══════╝╚═╝  ╚═╝  ╚═══╝   ╚═════╝   ╚═╝   ╚══════╝          ║
║                                                                       ║
║                    ✅ SETUP COMPLETE                                  ║
║                                                                       ║
╚═══════════════════════════════════════════════════════════════════════╝
`
	b.WriteString(SuccessStyle.Render(successBanner))
	b.WriteString("\n\n")

	// Dashboard URLs
	b.WriteString(RenderDashboardURLs(report))
	b.WriteString("\n\n")

	// Credentials (one-time display)
	b.WriteString(RenderCredentials(report))
	b.WriteString("\n\n")

	// Quick Start
	b.WriteString(RenderQuickStart(report))
	b.WriteString("\n\n")

	// Next Steps
	b.WriteString(RenderNextSteps(report))
	b.WriteString("\n")

	return b.String()
}

// RenderDashboardURLs renders the service dashboard URLs
func RenderDashboardURLs(report *MissionReport) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🌐 Dashboard URLs") + "\n\n")

	services := []struct {
		name    string
		url     string
		desc    string
		hasApp  bool
		appInfo string
	}{
		{
			name:    "📷 Immich",
			url:     report.ImmichURL,
			desc:    "Photo & Video Management",
			hasApp:  true,
			appInfo: "Mobile app: iOS/Android - Enter this URL in the app",
		},
		{
			name:    "☁️ Nextcloud",
			url:     report.NextcloudURL,
			desc:    "File Sync & Share",
			hasApp:  true,
			appInfo: "Mobile/Desktop apps available - Use this URL",
		},
		{
			name:    "📊 Glances",
			url:     report.GlancesURL,
			desc:    "System Monitoring",
			hasApp:  false,
			appInfo: "Browser only - No mobile app",
		},
	}

	for _, svc := range services {
		b.WriteString(fmt.Sprintf("  %s\n", TitleStyle.Render(svc.name)))
		b.WriteString(fmt.Sprintf("    URL: %s\n", URLStyle.Render(svc.url)))
		b.WriteString(fmt.Sprintf("    %s\n", MutedStyle.Render(svc.desc)))
		if svc.hasApp {
			b.WriteString(fmt.Sprintf("    📱 %s\n", MutedStyle.Render(svc.appInfo)))
		} else {
			b.WriteString(fmt.Sprintf("    💻 %s\n", MutedStyle.Render(svc.appInfo)))
		}
		b.WriteString("\n")
	}

	return BoxStyle.Render(b.String())
}

// RenderCredentials renders the generated credentials (ONE-TIME DISPLAY)
func RenderCredentials(report *MissionReport) string {
	var b strings.Builder

	b.WriteString(WarningStyle.Render("🔐 SAVE THESE CREDENTIALS NOW!") + "\n")
	b.WriteString(MutedStyle.Render("This is the only time they will be displayed.") + "\n\n")

	// Nextcloud Admin
	b.WriteString(SectionStyle.Render("Nextcloud Admin:") + "\n")
	b.WriteString(fmt.Sprintf("  Username: %s\n", CredentialStyle.Render(report.NextcloudAdminUser)))
	b.WriteString(fmt.Sprintf("  Password: %s\n\n", CredentialStyle.Render(report.NextcloudAdminPass)))

	// Database passwords
	b.WriteString(SectionStyle.Render("Database Passwords:") + "\n")
	b.WriteString(fmt.Sprintf("  Immich (PostgreSQL):    %s\n", CredentialStyle.Render(report.ImmichDBPassword)))
	b.WriteString(fmt.Sprintf("  Nextcloud (MariaDB):    %s\n\n", CredentialStyle.Render(report.NextcloudDBPassword)))

	// File location
	b.WriteString(MutedStyle.Render(fmt.Sprintf("Stored in: %s/.env (mode 0600)", report.ComposeDir)))

	return CredentialBoxStyle.Render(b.String())
}

// RenderQuickStart renders quick start commands
func RenderQuickStart(report *MissionReport) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🚀 Quick Start") + "\n\n")

	commands := []struct {
		desc string
		cmd  string
	}{
		{"Start all services:", fmt.Sprintf("cd %s && docker compose up -d", report.ComposeDir)},
		{"View logs:", fmt.Sprintf("cd %s && docker compose logs -f", report.ComposeDir)},
		{"Stop all services:", fmt.Sprintf("cd %s && docker compose down", report.ComposeDir)},
		{"Check status:", "servctl -status"},
		{"View architecture:", "servctl -get-architecture"},
		{"Manual backup:", "servctl -manual-backup"},
	}

	for _, c := range commands {
		b.WriteString(fmt.Sprintf("  %s\n", MutedStyle.Render(c.desc)))
		b.WriteString(fmt.Sprintf("  $ %s\n\n", SuccessStyle.Render(c.cmd)))
	}

	return BoxStyle.Render(b.String())
}

// RenderNextSteps renders next steps and recommendations
func RenderNextSteps(report *MissionReport) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📋 Next Steps") + "\n\n")

	steps := []string{
		"1. Start services: Run the docker compose command above",
		"2. Access Immich: Open " + report.ImmichURL + " and create your first account",
		"3. Access Nextcloud: Login with the admin credentials above",
		"4. Mobile Apps: Download Immich/Nextcloud apps and enter the URLs",
		"5. Set Static IP: Configure your router to give this server a static IP",
		"6. Backup: Verify daily_backup.sh runs at 4:00 AM (check logs tomorrow)",
	}

	for _, step := range steps {
		b.WriteString(fmt.Sprintf("  %s\n", step))
	}

	b.WriteString("\n")
	b.WriteString(SectionStyle.Render("🔑 To Change Passwords:") + "\n\n")
	b.WriteString(fmt.Sprintf("  1. Edit: %s/.env\n", report.ComposeDir))
	b.WriteString("  2. Update the PASSWORD variables\n")
	b.WriteString(fmt.Sprintf("  3. Restart: cd %s && docker compose down && docker compose up -d\n", report.ComposeDir))

	return BoxStyle.Render(b.String())
}

// RenderCompactReport renders a compact version of the report
func RenderCompactReport(report *MissionReport) string {
	var b strings.Builder

	b.WriteString(SuccessStyle.Render("✅ Setup Complete!") + "\n\n")

	b.WriteString(fmt.Sprintf("Immich:    %s\n", URLStyle.Render(report.ImmichURL)))
	b.WriteString(fmt.Sprintf("Nextcloud: %s (admin/%s)\n", URLStyle.Render(report.NextcloudURL), report.NextcloudAdminPass[:4]+"..."))
	b.WriteString(fmt.Sprintf("Glances:   %s\n\n", URLStyle.Render(report.GlancesURL)))

	b.WriteString(fmt.Sprintf("Config: %s\n", report.ComposeDir))
	b.WriteString(fmt.Sprintf("Start:  cd %s && docker compose up -d\n", report.ComposeDir))

	return b.String()
}
