// Package tui provides Bubble Tea TUI components for servctl.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/preflight"
)

// Color palette for the TUI
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorError     = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#6B7280") // Gray
	ColorHighlight = lipgloss.Color("#3B82F6") // Blue
)

// Styles for different elements
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			MarginTop(1)

	PassStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarnStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	FailStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	SkipStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true)

	DetailStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingLeft(4)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2)
)

// StatusIcon returns an icon for the given status
func StatusIcon(status preflight.Status) string {
	switch status {
	case preflight.StatusPass:
		return PassStyle.Render("✓")
	case preflight.StatusWarn:
		return WarnStyle.Render("⚠")
	case preflight.StatusFail:
		return FailStyle.Render("✗")
	case preflight.StatusSkip:
		return SkipStyle.Render("○")
	default:
		return "?"
	}
}

// StatusLabel returns a styled label for the status
func StatusLabel(status preflight.Status) string {
	switch status {
	case preflight.StatusPass:
		return PassStyle.Render("[PASS]")
	case preflight.StatusWarn:
		return WarnStyle.Render("[WARN]")
	case preflight.StatusFail:
		return FailStyle.Render("[FAIL]")
	case preflight.StatusSkip:
		return SkipStyle.Render("[SKIP]")
	default:
		return "[????]"
	}
}

// RenderCheckResult renders a single check result
func RenderCheckResult(result preflight.CheckResult) string {
	var b strings.Builder

	// Main line with icon and name
	b.WriteString(fmt.Sprintf("%s %s %s\n",
		StatusIcon(result.Status),
		StatusLabel(result.Status),
		result.Name,
	))

	// Message
	if result.Message != "" {
		messageStyle := DetailStyle
		switch result.Status {
		case preflight.StatusPass:
			messageStyle = messageStyle.Foreground(ColorSuccess)
		case preflight.StatusWarn:
			messageStyle = messageStyle.Foreground(ColorWarning)
		case preflight.StatusFail:
			messageStyle = messageStyle.Foreground(ColorError)
		}
		b.WriteString(messageStyle.Render("→ "+result.Message) + "\n")
	}

	// Details
	for _, detail := range result.Details {
		if detail == "" {
			b.WriteString("\n")
		} else {
			b.WriteString(DetailStyle.Render(detail) + "\n")
		}
	}

	return b.String()
}

// RenderPreflightResults renders all preflight check results
func RenderPreflightResults(results []preflight.CheckResult) string {
	var b strings.Builder

	// Header
	header := `
███████╗███████╗██████╗ ██╗   ██╗ ██████╗████████╗██╗     
██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝██║     
███████╗█████╗  ██████╔╝██║   ██║██║        ██║   ██║     
╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██║        ██║   ██║     
███████║███████╗██║  ██║ ╚████╔╝ ╚██████╗   ██║   ███████╗
╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝   ╚═════╝   ╚═╝   ╚══════╝
`
	b.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(header))
	b.WriteString("\n")
	b.WriteString(TitleStyle.Render("Pre-flight System Checks"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 60) + "\n\n")

	// Group results by category
	systemChecks := []preflight.CheckResult{}
	depChecks := []preflight.CheckResult{}
	serviceChecks := []preflight.CheckResult{}

	for _, r := range results {
		switch {
		case strings.HasPrefix(r.Name, "Dependency:"):
			depChecks = append(depChecks, r)
		case strings.Contains(r.Name, "Service") || strings.Contains(r.Name, "Docker Service"):
			serviceChecks = append(serviceChecks, r)
		default:
			systemChecks = append(systemChecks, r)
		}
	}

	// Render system checks
	if len(systemChecks) > 0 {
		b.WriteString(SectionStyle.Render("📋 System Prerequisites") + "\n\n")
		for _, r := range systemChecks {
			b.WriteString(RenderCheckResult(r))
			b.WriteString("\n")
		}
	}

	// Render dependency checks
	if len(depChecks) > 0 {
		b.WriteString(SectionStyle.Render("📦 Dependencies") + "\n\n")
		for _, r := range depChecks {
			b.WriteString(RenderCheckResult(r))
		}
		b.WriteString("\n")
	}

	// Render service checks
	if len(serviceChecks) > 0 {
		b.WriteString(SectionStyle.Render("🐳 Services") + "\n\n")
		for _, r := range serviceChecks {
			b.WriteString(RenderCheckResult(r))
			b.WriteString("\n")
		}
	}

	// Summary
	b.WriteString(strings.Repeat("─", 60) + "\n")
	b.WriteString(RenderSummary(results))

	return b.String()
}

// RenderSummary renders a summary of all check results
func RenderSummary(results []preflight.CheckResult) string {
	counts := preflight.CountByStatus(results)
	total := len(results)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(SectionStyle.Render("📊 Summary") + "\n\n")

	// Status counts
	if counts[preflight.StatusPass] > 0 {
		b.WriteString(fmt.Sprintf("  %s Passed: %d\n",
			PassStyle.Render("✓"),
			counts[preflight.StatusPass]))
	}
	if counts[preflight.StatusWarn] > 0 {
		b.WriteString(fmt.Sprintf("  %s Warnings: %d\n",
			WarnStyle.Render("⚠"),
			counts[preflight.StatusWarn]))
	}
	if counts[preflight.StatusFail] > 0 {
		b.WriteString(fmt.Sprintf("  %s Failed: %d\n",
			FailStyle.Render("✗"),
			counts[preflight.StatusFail]))
	}
	if counts[preflight.StatusSkip] > 0 {
		b.WriteString(fmt.Sprintf("  %s Skipped: %d\n",
			SkipStyle.Render("○"),
			counts[preflight.StatusSkip]))
	}

	b.WriteString(fmt.Sprintf("\n  Total checks: %d\n\n", total))

	// Overall status
	if preflight.HasBlockers(results) {
		msg := ErrorBoxStyle.Render(
			FailStyle.Render("❌ BLOCKED") + "\n\n" +
				"Critical issues must be resolved before continuing.\n" +
				"Please address the failed checks above and re-run servctl.")
		b.WriteString(msg + "\n")
	} else if counts[preflight.StatusWarn] > 0 {
		msg := BoxStyle.Render(
			WarnStyle.Render("⚠️  WARNINGS") + "\n\n" +
				"The system is ready but has some warnings.\n" +
				"Review the warnings above. You may continue,\n" +
				"but it's recommended to address them.")
		b.WriteString(msg + "\n")
	} else {
		msg := SuccessBoxStyle.Render(
			PassStyle.Render("✅ ALL CLEAR") + "\n\n" +
				"All pre-flight checks passed!\n" +
				"The system is ready for server setup.")
		b.WriteString(msg + "\n")
	}

	return b.String()
}

// RenderSpinner returns a simple spinner animation frame
func RenderSpinner(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return lipgloss.NewStyle().Foreground(ColorPrimary).Render(spinners[frame%len(spinners)])
}

// RenderProgress renders a progress bar
func RenderProgress(current, total int, width int) string {
	if width <= 0 {
		width = 40
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentStr := fmt.Sprintf("%3d%%", int(percent*100))

	return fmt.Sprintf("[%s] %s",
		lipgloss.NewStyle().Foreground(ColorPrimary).Render(bar),
		percentStr)
}
