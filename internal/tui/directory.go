package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/directory"
)

// Directory-specific styles
var (
	DirStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight)

	DirCreatedStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	DirExistsStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	DirFailedStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	ServiceBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3B82F6")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	ImmichBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#8B5CF6")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	NextcloudBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#0082C9")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	DatabaseBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#F59E0B")).
				Foreground(lipgloss.Color("#000000")).
				Padding(0, 1)

	DirectoryTreeStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMuted).
				Padding(1, 2)
)

// getServiceBadge returns an appropriate badge for a service
func getServiceBadge(service string) string {
	switch service {
	case "immich":
		return ImmichBadgeStyle.Render("IMMICH")
	case "nextcloud":
		return NextcloudBadgeStyle.Render("NEXTCLOUD")
	case "database":
		return DatabaseBadgeStyle.Render("DATABASE")
	case "core":
		return ServiceBadgeStyle.Render("CORE")
	case "docker":
		return ServiceBadgeStyle.Render("DOCKER")
	case "backup":
		return ServiceBadgeStyle.Render("BACKUP")
	case "redis":
		return ServiceBadgeStyle.Render("REDIS")
	default:
		return ServiceBadgeStyle.Render(strings.ToUpper(service))
	}
}

// RenderDirectoryPlan renders the directory creation plan
func RenderDirectoryPlan(specs []directory.DirectorySpec) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📁 Directory Structure Plan") + "\n\n")

	// Group by type
	userSpaceDirs := make([]directory.DirectorySpec, 0)
	dataSpaceDirs := make([]directory.DirectorySpec, 0)

	for _, spec := range specs {
		switch spec.Type {
		case directory.DirTypeUserSpace:
			userSpaceDirs = append(userSpaceDirs, spec)
		case directory.DirTypeDataSpace:
			dataSpaceDirs = append(dataSpaceDirs, spec)
		}
	}

	// User Space directories
	if len(userSpaceDirs) > 0 {
		b.WriteString(TitleStyle.Render("🏠 User Space (~/infra/)") + "\n")
		b.WriteString(DetailStyle.Render("Configuration files and scripts") + "\n\n")

		for _, spec := range userSpaceDirs {
			b.WriteString(fmt.Sprintf("  %s %s\n",
				DirStyle.Render("📂 "+spec.Path),
				getServiceBadge(spec.Service),
			))
			b.WriteString(fmt.Sprintf("     %s\n", DetailStyle.Render(spec.Description)))
		}
		b.WriteString("\n")
	}

	// Data Space directories
	if len(dataSpaceDirs) > 0 {
		b.WriteString(TitleStyle.Render("💾 Data Space (/mnt/data/)") + "\n")
		b.WriteString(DetailStyle.Render("Service data and databases") + "\n\n")

		for _, spec := range dataSpaceDirs {
			b.WriteString(fmt.Sprintf("  %s %s\n",
				DirStyle.Render("📂 "+spec.Path),
				getServiceBadge(spec.Service),
			))
			b.WriteString(fmt.Sprintf("     %s\n", DetailStyle.Render(spec.Description)))
		}
	}

	// Summary
	counts := directory.CountByService(specs)
	b.WriteString("\n" + SectionStyle.Render("Summary") + "\n")
	b.WriteString(fmt.Sprintf("  Total directories: %d\n", len(specs)))
	b.WriteString("  By service:\n")
	for service, count := range counts {
		b.WriteString(fmt.Sprintf("    • %s: %d\n", service, count))
	}

	return b.String()
}

// RenderDirectoryTree renders a tree view of directories
func RenderDirectoryTree(homeDir, dataRoot string) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📂 Directory Tree") + "\n\n")

	// User space tree
	b.WriteString(TitleStyle.Render("~/infra/") + "\n")
	b.WriteString("├── " + DirStyle.Render("scripts/") + "     # Maintenance scripts\n")
	b.WriteString("├── " + DirStyle.Render("logs/") + "        # Centralized logs\n")
	b.WriteString("├── " + DirStyle.Render("compose/") + "     # Docker Compose files\n")
	b.WriteString("├── " + DirStyle.Render("config/") + "      # Service configs\n")
	b.WriteString("└── " + DirStyle.Render("backups/") + "     # Backup staging\n")
	b.WriteString("\n")

	// Data space tree
	b.WriteString(TitleStyle.Render("/mnt/data/") + "\n")
	b.WriteString("├── " + ImmichBadgeStyle.Render("gallery/") + "\n")
	b.WriteString("│   ├── library/         # Photo storage\n")
	b.WriteString("│   ├── upload/          # Upload staging\n")
	b.WriteString("│   ├── profile/         # User profiles\n")
	b.WriteString("│   ├── video/           # Video transcodes\n")
	b.WriteString("│   └── thumbs/          # Thumbnails\n")
	b.WriteString("├── " + NextcloudBadgeStyle.Render("cloud/") + "\n")
	b.WriteString("│   ├── data/            # User files\n")
	b.WriteString("│   └── config/          # NC config\n")
	b.WriteString("├── " + DatabaseBadgeStyle.Render("databases/") + "\n")
	b.WriteString("│   ├── immich-postgres/ # Immich DB\n")
	b.WriteString("│   └── nextcloud-mariadb/ # NC DB\n")
	b.WriteString("└── cache/               # Redis data\n")

	return DirectoryTreeStyle.Render(b.String())
}

// RenderDirectoryProgress renders directory creation progress
func RenderDirectoryProgress(results []directory.DirectoryResult) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🔧 Creating Directories") + "\n\n")

	for _, result := range results {
		var statusIcon string
		var statusStyle lipgloss.Style

		if result.Error != nil {
			statusIcon = "✗"
			statusStyle = DirFailedStyle
		} else if result.Created {
			statusIcon = "✓"
			statusStyle = DirCreatedStyle
		} else {
			statusIcon = "○"
			statusStyle = DirExistsStyle
		}

		b.WriteString(fmt.Sprintf("  %s %s\n",
			statusStyle.Render(statusIcon),
			statusStyle.Render(result.Spec.Path),
		))

		if result.Error != nil {
			b.WriteString(fmt.Sprintf("     %s\n",
				FailStyle.Render("Error: "+result.Error.Error()),
			))
		}
	}

	return b.String()
}

// RenderDirectoryComplete renders the directory creation summary
func RenderDirectoryComplete(results []directory.DirectoryResult, perm *directory.PermissionInfo) string {
	var b strings.Builder

	created := directory.CountCreated(results)
	existing := directory.CountExisting(results)
	failed := directory.CountFailed(results)

	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(1, 2)

	if failed > 0 {
		successBox = successBox.BorderForeground(ColorError)
	}

	b.WriteString(PassStyle.Render("📁 Directory Structure Created") + "\n\n")

	b.WriteString(fmt.Sprintf("  %s %d directories created\n",
		DirCreatedStyle.Render("✓"),
		created,
	))
	b.WriteString(fmt.Sprintf("  %s %d directories already existed\n",
		DirExistsStyle.Render("○"),
		existing,
	))

	if failed > 0 {
		b.WriteString(fmt.Sprintf("  %s %d directories failed\n",
			DirFailedStyle.Render("✗"),
			failed,
		))
	}

	if perm != nil {
		b.WriteString("\n")
		b.WriteString(SectionStyle.Render("Permissions:") + "\n")
		b.WriteString(fmt.Sprintf("  Owner: %s (UID: %d, GID: %d)\n",
			perm.Username, perm.UID, perm.GID,
		))
	}

	return successBox.Render(b.String())
}

// RenderPermissionConfirmation renders permission setting confirmation
func RenderPermissionConfirmation(dataRoot string, perm *directory.PermissionInfo) string {
	var b strings.Builder

	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorHighlight).
		Padding(1, 2)

	b.WriteString(SectionStyle.Render("🔐 Setting Permissions") + "\n\n")
	b.WriteString(fmt.Sprintf("Setting ownership of %s to:\n\n", dataRoot))
	b.WriteString(fmt.Sprintf("  User:  %s (UID: %d)\n", perm.Username, perm.UID))
	b.WriteString(fmt.Sprintf("  Group: %d\n", perm.GID))
	b.WriteString("\n")
	b.WriteString(DetailStyle.Render("This ensures services can read/write to data directories.") + "\n")
	b.WriteString(DetailStyle.Render("Directories: 755 (rwxr-xr-x)") + "\n")
	b.WriteString(DetailStyle.Render("Files: 644 (rw-r--r--)") + "\n")

	return infoBox.Render(b.String())
}
