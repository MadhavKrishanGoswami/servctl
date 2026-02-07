// Package maintenance handles generation of maintenance scripts and cron configuration.
// This file implements interactive script selection prompts.
package maintenance

import (
	"bufio"
	"fmt"
	"strings"
)

// ScriptSelection represents which scripts to generate
type ScriptSelection struct {
	DailyBackup   bool
	DiskAlert     bool
	SmartAlert    bool
	WeeklyCleanup bool
}

// DefaultScriptSelection returns all scripts enabled
func DefaultScriptSelection() ScriptSelection {
	return ScriptSelection{
		DailyBackup:   true,
		DiskAlert:     true,
		SmartAlert:    false, // Requires smartctl
		WeeklyCleanup: true,
	}
}

// PromptScriptSelection prompts user to select which scripts to generate
func PromptScriptSelection(reader *bufio.Reader) ScriptSelection {
	selection := DefaultScriptSelection()

	fmt.Println("Select maintenance scripts to generate:")
	fmt.Println()

	renderSelection := func() {
		checkbox := func(enabled bool) string {
			if enabled {
				return "[x]"
			}
			return "[ ]"
		}
		fmt.Printf("  1. %s Daily Backup    - rsync data to backup drive\n", checkbox(selection.DailyBackup))
		fmt.Printf("  2. %s Disk Alert      - Alert when disk >90%% full\n", checkbox(selection.DiskAlert))
		fmt.Printf("  3. %s SMART Monitor   - Drive health monitoring\n", checkbox(selection.SmartAlert))
		fmt.Printf("  4. %s Weekly Cleanup  - Docker/apt/log cleanup\n", checkbox(selection.WeeklyCleanup))
		fmt.Println()
	}

	renderSelection()
	fmt.Print("Toggle (e.g., '1 3' to toggle), or Enter to continue: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return selection
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return selection
	}

	// Parse toggles
	for _, char := range strings.Fields(response) {
		switch char {
		case "1":
			selection.DailyBackup = !selection.DailyBackup
		case "2":
			selection.DiskAlert = !selection.DiskAlert
		case "3":
			selection.SmartAlert = !selection.SmartAlert
		case "4":
			selection.WeeklyCleanup = !selection.WeeklyCleanup
		}
	}

	fmt.Println("Selected scripts:")
	renderSelection()

	return selection
}

// PromptBackupSchedule prompts user to select backup schedule
func PromptBackupSchedule(reader *bufio.Reader) string {
	fmt.Println("Backup schedule:")
	fmt.Println("  1. Daily at 3 AM")
	fmt.Println("  2. Every 6 hours")
	fmt.Println("  3. Every 12 hours")
	fmt.Println("  4. Weekly (Sunday 3 AM)")
	fmt.Print("Select [1-4, default: 1]: ")

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	switch response {
	case "2":
		return "6h"
	case "3":
		return "12h"
	case "4":
		return "weekly"
	default:
		return "daily"
	}
}

// PromptWebhookURL prompts for optional Discord/Telegram webhook
func PromptWebhookURL(reader *bufio.Reader) string {
	fmt.Print("Discord/Telegram webhook URL (Enter to skip): ")

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	return response
}

// GetScriptsForSelection filters scripts based on selection
func GetScriptsForSelection(sel ScriptSelection, config *ScriptConfig) ([]ScriptInfo, error) {
	var scripts []ScriptInfo

	if sel.DailyBackup {
		script, err := GenerateDailyBackup(config)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, ScriptInfo{
			Name:        "Daily Backup",
			Filename:    "daily-backup.sh",
			Description: "Syncs data to backup drive",
			Schedule:    "3 AM daily",
			Content:     script,
		})
	}

	if sel.DiskAlert {
		script, err := GenerateDiskAlert(config)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, ScriptInfo{
			Name:        "Disk Alert",
			Filename:    "disk-alert.sh",
			Description: "Alerts when disk usage exceeds threshold",
			Schedule:    "Hourly",
			Content:     script,
		})
	}

	if sel.SmartAlert {
		script, err := GenerateSmartAlert(config)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, ScriptInfo{
			Name:        "SMART Monitor",
			Filename:    "smart-monitor.sh",
			Description: "Monitors drive health using smartctl",
			Schedule:    "Daily",
			Content:     script,
		})
	}

	if sel.WeeklyCleanup {
		script, err := GenerateWeeklyCleanup(config)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, ScriptInfo{
			Name:        "Weekly Cleanup",
			Filename:    "weekly-cleanup.sh",
			Description: "Cleans up Docker, apt, and logs",
			Schedule:    "Sunday 3 AM",
			Content:     script,
		})
	}

	return scripts, nil
}

// SelectedNames returns names of selected scripts
func (s ScriptSelection) SelectedNames() []string {
	var names []string
	if s.DailyBackup {
		names = append(names, "Daily Backup")
	}
	if s.DiskAlert {
		names = append(names, "Disk Alert")
	}
	if s.SmartAlert {
		names = append(names, "SMART Monitor")
	}
	if s.WeeklyCleanup {
		names = append(names, "Weekly Cleanup")
	}
	return names
}
