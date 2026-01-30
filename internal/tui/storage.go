package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/storage"
)

// Storage-specific styles
var (
	DiskStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorHighlight).
			Padding(0, 1).
			MarginBottom(1)

	DiskHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	DiskTypeStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	SSDStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")). // Green
			Bold(true)

	HDDStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")). // Amber
			Bold(true)

	NVMeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")). // Purple
			Bold(true)

	WarningBannerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#F59E0B")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 2)

	CriticalBannerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#EF4444")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 2)

	RecommendedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#10B981")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1)

	RankBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2).
			Width(60)

	SelectedRankBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2).
				Width(60)
)

// RenderDiskInfo renders information about a single disk
func RenderDiskInfo(disk storage.Disk) string {
	var b strings.Builder

	// Disk type icon and style
	var typeIcon string
	var typeStyle lipgloss.Style
	switch disk.Type {
	case storage.DiskTypeSSD:
		typeIcon = "⚡"
		typeStyle = SSDStyle
	case storage.DiskTypeHDD:
		typeIcon = "💿"
		typeStyle = HDDStyle
	case storage.DiskTypeNVMe:
		typeIcon = "🚀"
		typeStyle = NVMeStyle
	default:
		typeIcon = "💾"
		typeStyle = DiskTypeStyle
	}

	// Header line
	headerLine := fmt.Sprintf("%s %s  %s  %s",
		typeIcon,
		typeStyle.Render(disk.Type.String()),
		DiskHeaderStyle.Render(disk.Path),
		disk.SizeHuman,
	)
	b.WriteString(headerLine + "\n")

	// Model info
	if disk.Model != "" {
		b.WriteString(DetailStyle.Render("  Model: "+disk.Model) + "\n")
	}

	// Partitions
	if len(disk.Partitions) > 0 {
		b.WriteString(DetailStyle.Render("  Partitions:") + "\n")
		for _, part := range disk.Partitions {
			partInfo := fmt.Sprintf("    • %s: %s", part.Name, part.SizeHuman)
			if part.MountPoint != "" {
				partInfo += fmt.Sprintf(" → %s", part.MountPoint)
			}
			if part.Filesystem != "" {
				partInfo += fmt.Sprintf(" (%s)", part.Filesystem)
			}
			b.WriteString(DetailStyle.Render(partInfo) + "\n")
		}
	}

	// Status badges
	var badges []string
	if disk.IsOSDisk {
		badges = append(badges, lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render(" OS "))
	}
	if disk.Removable {
		badges = append(badges, lipgloss.NewStyle().
			Background(lipgloss.Color("#6B7280")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render(" USB "))
	}
	if len(badges) > 0 {
		b.WriteString("  " + strings.Join(badges, " ") + "\n")
	}

	return DiskStyle.Render(b.String())
}

// RenderDiskDiscovery renders the disk discovery results
func RenderDiskDiscovery(disks []storage.Disk) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("💾 Discovered Storage Devices") + "\n\n")

	if len(disks) == 0 {
		b.WriteString(WarnStyle.Render("No disks detected!") + "\n")
		return b.String()
	}

	// Summary
	var ssdCount, hddCount, nvmeCount int
	var totalSize uint64
	for _, disk := range disks {
		totalSize += disk.Size
		switch disk.Type {
		case storage.DiskTypeSSD:
			ssdCount++
		case storage.DiskTypeHDD:
			hddCount++
		case storage.DiskTypeNVMe:
			nvmeCount++
		}
	}

	summaryLine := fmt.Sprintf("Found %d disk(s): ", len(disks))
	var typeSummary []string
	if nvmeCount > 0 {
		typeSummary = append(typeSummary, NVMeStyle.Render(fmt.Sprintf("%d NVMe", nvmeCount)))
	}
	if ssdCount > 0 {
		typeSummary = append(typeSummary, SSDStyle.Render(fmt.Sprintf("%d SSD", ssdCount)))
	}
	if hddCount > 0 {
		typeSummary = append(typeSummary, HDDStyle.Render(fmt.Sprintf("%d HDD", hddCount)))
	}
	b.WriteString(summaryLine + strings.Join(typeSummary, ", ") + "\n\n")

	// Render each disk
	for _, disk := range disks {
		b.WriteString(RenderDiskInfo(disk))
	}

	return b.String()
}

// RenderStorageScenario renders the detected storage scenario
func RenderStorageScenario(result *storage.ClassificationResult) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📊 Storage Scenario Analysis") + "\n\n")

	// Scenario type
	scenarioStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)
	b.WriteString(fmt.Sprintf("Detected Scenario: %s\n\n", scenarioStyle.Render(result.Scenario.String())))

	// OS Disk info
	if result.OSDisk != nil {
		b.WriteString(DetailStyle.Render("OS Disk: "+result.OSDisk.Path+" ("+result.OSDisk.SizeHuman+")") + "\n")
	}

	// Available disks
	b.WriteString(DetailStyle.Render(fmt.Sprintf("Available for data: %d disk(s)", len(result.AvailableDisks))) + "\n\n")

	// Single disk warning
	if result.Scenario == storage.ScenarioSingleDisk {
		warning := WarningBannerStyle.Render(" ⚠️  SINGLE POINT OF FAILURE ")
		b.WriteString(warning + "\n\n")
		b.WriteString(WarnStyle.Render("With only one disk, a failure will result in data loss.") + "\n")
		b.WriteString(DetailStyle.Render("Consider adding a backup disk in the future.") + "\n\n")
	}

	return b.String()
}

// RenderStorageRank renders a single storage rank recommendation
func RenderStorageRank(rec storage.StorageRecommendation, isSelected bool, index int) string {
	var b strings.Builder

	// Choose box style based on selection
	boxStyle := RankBoxStyle
	if isSelected {
		boxStyle = SelectedRankBoxStyle
	}

	// Header with rank number
	header := fmt.Sprintf("[%d] %s", index+1, rec.Name)
	if rec.IsDefault {
		header += " " + RecommendedStyle.Render("RECOMMENDED")
	}
	b.WriteString(TitleStyle.Render(header) + "\n\n")

	// Description
	b.WriteString(rec.Description + "\n\n")

	// Pros
	if len(rec.Pros) > 0 {
		b.WriteString(PassStyle.Render("✓ Pros:") + "\n")
		for _, pro := range rec.Pros {
			b.WriteString(fmt.Sprintf("  • %s\n", pro))
		}
	}

	// Cons
	if len(rec.Cons) > 0 {
		b.WriteString(WarnStyle.Render("⚠ Cons:") + "\n")
		for _, con := range rec.Cons {
			b.WriteString(fmt.Sprintf("  • %s\n", con))
		}
	}

	// Warning banner
	if rec.Warning != "" {
		b.WriteString("\n" + CriticalBannerStyle.Render(rec.Warning) + "\n")
	}

	// Disk assignments
	if len(rec.Disks) > 0 {
		b.WriteString("\n" + SectionStyle.Render("Disk Assignment:") + "\n")
		for _, assignment := range rec.Disks {
			if assignment.Disk != nil {
				assignLine := fmt.Sprintf("  %s (%s) → %s [%s]",
					assignment.Disk.Path,
					assignment.Disk.SizeHuman,
					assignment.Role,
					assignment.Mount,
				)
				b.WriteString(DetailStyle.Render(assignLine) + "\n")
			}
		}
	}

	return boxStyle.Render(b.String())
}

// RenderStorageRecommendations renders all storage recommendations
func RenderStorageRecommendations(recommendations []storage.StorageRecommendation, selectedIndex int) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("🏆 The 5 Ranks - Storage Configurations") + "\n\n")

	if len(recommendations) == 0 {
		b.WriteString(WarnStyle.Render("No recommendations available for current disk configuration.") + "\n")
		return b.String()
	}

	b.WriteString(DetailStyle.Render("Choose a storage configuration:") + "\n\n")

	for i, rec := range recommendations {
		isSelected := i == selectedIndex
		b.WriteString(RenderStorageRank(rec, isSelected, i))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n" + DetailStyle.Render("Use ↑/↓ to navigate, Enter to select, or press the number key.") + "\n")

	return b.String()
}

// RenderFilesystemSelection renders filesystem selection menu
func RenderFilesystemSelection(options []storage.FilesystemOption, selectedIndex int) string {
	var b strings.Builder

	b.WriteString(SectionStyle.Render("📁 Select Filesystem Format") + "\n\n")

	for i, opt := range options {
		isSelected := i == selectedIndex
		b.WriteString(RenderFilesystemOption(opt, isSelected, i))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n" + DetailStyle.Render("Press Enter for default (ext4), or select another option.") + "\n")

	return b.String()
}

// RenderFilesystemOption renders a single filesystem option
func RenderFilesystemOption(opt storage.FilesystemOption, isSelected bool, index int) string {
	var b strings.Builder

	boxStyle := RankBoxStyle
	if isSelected {
		boxStyle = SelectedRankBoxStyle
	}

	// Header
	header := fmt.Sprintf("[%d] %s", index+1, opt.Name)
	if opt.IsDefault {
		header += " " + RecommendedStyle.Render("⭐ PRESS ENTER TO USE")
	}
	b.WriteString(TitleStyle.Render(header) + "\n\n")

	// Pros
	for _, pro := range opt.Pros {
		b.WriteString(PassStyle.Render("✓ ") + pro + "\n")
	}

	// Cons
	for _, con := range opt.Cons {
		b.WriteString(WarnStyle.Render("⚠ ") + con + "\n")
	}

	return boxStyle.Render(b.String())
}

// RenderFormatConfirmation renders the destructive action confirmation
func RenderFormatConfirmation(disk storage.Disk, fsType storage.FilesystemType) string {
	var b strings.Builder

	warningBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#EF4444")).
		Padding(1, 2)

	b.WriteString(CriticalBannerStyle.Render(" ⚠️  WARNING: DESTRUCTIVE OPERATION ") + "\n\n")
	b.WriteString(fmt.Sprintf("This will format %s with %s filesystem.\n", disk.Path, fsType.String()))
	b.WriteString(FailStyle.Render("ALL DATA ON THIS DISK WILL BE PERMANENTLY ERASED!") + "\n\n")
	b.WriteString(fmt.Sprintf("Disk: %s\n", disk.Path))
	b.WriteString(fmt.Sprintf("Size: %s\n", disk.SizeHuman))
	if disk.Model != "" {
		b.WriteString(fmt.Sprintf("Model: %s\n", disk.Model))
	}
	b.WriteString("\n")
	b.WriteString(WarnStyle.Render("Type 'YES' to confirm, or 'no' to cancel:") + "\n")

	return warningBox.Render(b.String())
}

// RenderStorageComplete renders the storage setup completion summary
func RenderStorageComplete(assignments []storage.DiskAssignment) string {
	var b strings.Builder

	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(1, 2)

	b.WriteString(PassStyle.Render("✅ Storage Configuration Complete") + "\n\n")

	b.WriteString(SectionStyle.Render("Configured Mounts:") + "\n")
	for _, assignment := range assignments {
		if assignment.Disk != nil {
			b.WriteString(fmt.Sprintf("  • %s → %s (%s)\n",
				assignment.Disk.Path,
				assignment.Mount,
				assignment.Role,
			))
		}
	}

	return successBox.Render(b.String())
}
