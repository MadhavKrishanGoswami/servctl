package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/compose"
)

// Compose-specific styles
var (
	InputLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	InputValueStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight)

	InputHintStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	ServiceCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMuted).
				Padding(1, 2).
				Width(50)

	ServiceActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Padding(1, 2).
				Width(50)

	PortBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	PasswordMaskStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	NetworkInfoStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight).
				Padding(1, 2)
)

// RenderConfigWizardIntro renders the intro for the configuration wizard
func RenderConfigWizardIntro() string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("⚙️  Service Configuration") + "\n\n")

	b.WriteString("We'll now collect the settings needed to deploy your services.\n\n")

	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(1, 2)

	info := `
Servctl is an opinionated tool. The following are fixed:
  • Data location: /mnt/data/
  • Upload location: /mnt/data/gallery/
  • Config location: ~/infra/

You can customize:
  • Timezone
  • Database passwords (auto-generated if blank)
  • Nextcloud admin credentials
  • Notification webhooks
`
	b.WriteString(infoBox.Render(info))

	return b.String()
}

// RenderInputPrompt renders a single input prompt
func RenderInputPrompt(label, hint, currentValue string, isPassword bool) string {
	var b strings.Builder

	b.WriteString(InputLabelStyle.Render(label) + "\n")

	if hint != "" {
		b.WriteString(InputHintStyle.Render(hint) + "\n")
	}

	if currentValue != "" {
		displayValue := currentValue
		if isPassword {
			displayValue = strings.Repeat("•", len(currentValue))
		}
		b.WriteString(fmt.Sprintf("Current: %s\n", InputValueStyle.Render(displayValue)))
	}

	b.WriteString("\n> ")

	return b.String()
}

// RenderTimezoneSelection renders timezone selection
func RenderTimezoneSelection(options []string, selectedIndex int, currentTZ string) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🌍 Timezone Selection") + "\n\n")
	b.WriteString(fmt.Sprintf("Detected timezone: %s\n\n", InputValueStyle.Render(currentTZ)))

	for i, tz := range options {
		prefix := "  "
		if i == selectedIndex {
			prefix = "▶ "
			b.WriteString(InputValueStyle.Render(prefix+tz) + "\n")
		} else {
			b.WriteString(prefix + tz + "\n")
		}
	}

	b.WriteString("\n" + DetailStyle.Render("Use ↑/↓ to select, Enter to confirm") + "\n")

	return b.String()
}

// RenderNetworkConfig renders network configuration screen
func RenderNetworkConfig(info *compose.NetworkInfo, staticIP string, warning string) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🌐 Network Configuration") + "\n\n")

	// Current network info
	var netInfo strings.Builder
	netInfo.WriteString(TitleStyle.Render("Current Network:") + "\n\n")
	if info != nil {
		netInfo.WriteString(fmt.Sprintf("  IP Address: %s\n", InputValueStyle.Render(info.CurrentIP)))
		if info.Gateway != "" {
			netInfo.WriteString(fmt.Sprintf("  Gateway:    %s\n", info.Gateway))
		}
		if info.Interface != "" {
			netInfo.WriteString(fmt.Sprintf("  Interface:  %s\n", info.Interface))
		}
		if info.IsDHCP {
			netInfo.WriteString(fmt.Sprintf("  Type:       %s\n", WarnStyle.Render("DHCP (Dynamic)")))
		} else {
			netInfo.WriteString(fmt.Sprintf("  Type:       %s\n", PassStyle.Render("Static")))
		}
	}
	b.WriteString(NetworkInfoStyle.Render(netInfo.String()))
	b.WriteString("\n\n")

	// Static IP input
	if staticIP != "" {
		b.WriteString(fmt.Sprintf("Static IP: %s\n", InputValueStyle.Render(staticIP)))
	}

	// Warning
	if warning != "" {
		b.WriteString("\n" + WarnStyle.Render("⚠️  "+warning) + "\n")
	}

	// DHCP warning
	dhcpWarning := `
IMPORTANT: Nextcloud requires a static IP address!

If your router assigns a new IP via DHCP, you may get locked out
of Nextcloud. Please ensure you:
  1. Set a static IP in your router, or
  2. Configure a static IP on this server
`
	b.WriteString("\n" + WarningBannerStyle.Render(" STATIC IP RECOMMENDED ") + "\n")
	b.WriteString(DetailStyle.Render(dhcpWarning))

	return b.String()
}

// RenderGeneratedCredentials renders auto-generated credentials
func RenderGeneratedCredentials(config *compose.ServiceConfig) string {
	var b strings.Builder

	credBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorWarning).
		Padding(1, 2)

	b.WriteString(SectionStyle.Render("🔐 Generated Credentials") + "\n\n")

	b.WriteString(WarnStyle.Render("SAVE THESE CREDENTIALS SECURELY!") + "\n\n")

	// Show passwords (this is intentional - one-time display)
	creds := fmt.Sprintf(`
Immich Database:
  Password: %s

Nextcloud Admin:
  Username: %s
  Password: %s

Nextcloud Database:
  Password: %s
`,
		config.ImmichDBPassword,
		config.NextcloudAdminUser,
		config.NextcloudAdminPass,
		config.NextcloudDBPassword,
	)

	b.WriteString(credBox.Render(creds))
	b.WriteString("\n\n")
	b.WriteString(DetailStyle.Render("These are saved in ~/infra/compose/.env (mode 0600)"))

	return b.String()
}

// RenderServiceSummary renders a summary of services to be deployed
func RenderServiceSummary(config *compose.ServiceConfig) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🐳 Services to Deploy") + "\n\n")

	// Services list
	services := []struct {
		name        string
		port        int
		description string
		icon        string
	}{
		{"Immich", config.ImmichPort, "Photo & Video Management", "📷"},
		{"Immich ML", 0, "Machine Learning (Internal)", "🤖"},
		{"Nextcloud", config.NextcloudPort, "File Sync & Share", "☁️"},
		{"Glances", config.GlancesPort, "System Monitoring", "📊"},
		{"Diun", 0, "Update Notifications", "🔔"},
		{"Redis", 0, "Cache (Internal)", "⚡"},
		{"PostgreSQL", 0, "Immich Database (Isolated)", "🐘"},
		{"MariaDB", 0, "Nextcloud Database (Isolated)", "🐬"},
	}

	for _, svc := range services {
		var card strings.Builder

		card.WriteString(fmt.Sprintf("%s %s\n", svc.icon, TitleStyle.Render(svc.name)))
		card.WriteString(DetailStyle.Render(svc.description) + "\n")

		if svc.port > 0 {
			card.WriteString("Port: " + PortBadgeStyle.Render(fmt.Sprintf("%d", svc.port)))
		} else {
			card.WriteString(DetailStyle.Render("(No external port)"))
		}

		b.WriteString(ServiceCardStyle.Render(card.String()))
		b.WriteString("\n")
	}

	// Access URLs
	if config.HostIP != "" {
		b.WriteString("\n" + SectionStyle.Render("Access URLs:") + "\n")
		b.WriteString(fmt.Sprintf("  Immich:    %s\n",
			InputValueStyle.Render(fmt.Sprintf("http://%s:%d", config.HostIP, config.ImmichPort))))
		b.WriteString(fmt.Sprintf("  Nextcloud: %s\n",
			InputValueStyle.Render(fmt.Sprintf("http://%s:%d", config.HostIP, config.NextcloudPort))))
		b.WriteString(fmt.Sprintf("  Glances:   %s\n",
			InputValueStyle.Render(fmt.Sprintf("http://%s:%d", config.HostIP, config.GlancesPort))))
	}

	return b.String()
}

// RenderFirewallConfig renders firewall configuration screen
func RenderFirewallConfig(rules []compose.FirewallRule) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🔥 Firewall Configuration (UFW)") + "\n\n")

	// Lockout warning
	b.WriteString(WarningBannerStyle.Render(" ⚠️  LOCKOUT PREVENTION ") + "\n\n")
	b.WriteString(WarnStyle.Render("SSH (port 22) will be allowed FIRST to prevent lockout.") + "\n\n")

	b.WriteString(TitleStyle.Render("Rules to be applied:") + "\n\n")

	for _, rule := range rules {
		status := PassStyle.Render("✓")
		if !rule.Required {
			status = DetailStyle.Render("○")
		}

		portInfo := fmt.Sprintf("%d/%s", rule.Port, rule.Protocol)
		b.WriteString(fmt.Sprintf("  %s %s - %s\n",
			status,
			PortBadgeStyle.Render(portInfo),
			rule.Service,
		))
		b.WriteString(fmt.Sprintf("      %s\n", DetailStyle.Render(rule.Description)))
	}

	return b.String()
}

// RenderComposeGenerated renders compose file generation confirmation
func RenderComposeGenerated(outputDir string) string {
	var b strings.Builder

	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(1, 2)

	b.WriteString(PassStyle.Render("✅ Docker Compose Generated") + "\n\n")

	files := fmt.Sprintf(`
Generated files in %s:

  📄 docker-compose.yml   - Service definitions
  🔐 .env                 - Configuration (mode 0600)

To start services:
  cd %s
  docker compose up -d

To view logs:
  docker compose logs -f
`, outputDir, outputDir)

	b.WriteString(successBox.Render(files))

	return b.String()
}

// RenderConfigSummary renders the final configuration summary
func RenderConfigSummary(config *compose.ServiceConfig) string {
	var b strings.Builder

	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2)

	b.WriteString(SectionStyle.Render("📋 Configuration Summary") + "\n\n")

	summary := fmt.Sprintf(`
System:
  Timezone:    %s
  PUID/PGID:   %d/%d
  Host IP:     %s

Paths:
  Data Root:   %s
  Upload Dir:  %s

Services:
  Immich:      :%d
  Nextcloud:   :%d (admin: %s)
  Glances:     :%d

Notifications:
  Discord:     %s
  Telegram:    %s
`,
		config.Timezone,
		config.PUID, config.PGID,
		config.HostIP,
		config.DataRoot,
		config.UploadPath,
		config.ImmichPort,
		config.NextcloudPort, config.NextcloudAdminUser,
		config.GlancesPort,
		maskEmpty(config.DiscordWebhookURL),
		maskEmpty(config.TelegramBotToken),
	)

	b.WriteString(summaryBox.Render(summary))

	return b.String()
}

// maskEmpty returns "Not configured" for empty strings, or "Configured" otherwise
func maskEmpty(s string) string {
	if s == "" {
		return DetailStyle.Render("Not configured")
	}
	return PassStyle.Render("Configured")
}
