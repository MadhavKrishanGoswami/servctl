package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madhav/servctl/internal/compose"
	"github.com/madhav/servctl/internal/directory"
	"github.com/madhav/servctl/internal/maintenance"
	"github.com/madhav/servctl/internal/preflight"
	"github.com/madhav/servctl/internal/report"
	"github.com/madhav/servctl/internal/storage"
	"github.com/madhav/servctl/internal/tui"
	"github.com/madhav/servctl/internal/utils"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8B5CF6")).
			MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(1, 2)
)

func main() {
	// Command line flags
	startSetup := flag.Bool("start-setup", false, "Launch interactive installation wizard")
	status := flag.Bool("status", false, "Display current system status")
	getConfig := flag.Bool("get-config", false, "Display current configuration")
	getArch := flag.Bool("get-architecture", false, "Display folder structure and disk mapping")
	manualBackup := flag.Bool("manual-backup", false, "Trigger immediate backup")
	logs := flag.Bool("logs", false, "Display service logs")
	version := flag.Bool("version", false, "Display version information")
	preflightOnly := flag.Bool("preflight", false, "Run preflight checks only")
	dryRun := flag.Bool("dry-run", false, "Preview changes without making them")

	flag.Parse()

	// Handle version flag
	if *version {
		printVersion()
		return
	}

	// Handle preflight only
	if *preflightOnly {
		runPreflightChecks()
		return
	}

	// Handle start-setup (main wizard)
	if *startSetup {
		runSetupWizard(*dryRun)
		return
	}

	// Handle status
	if *status {
		runStatusCommand()
		return
	}

	// Handle get-config
	if *getConfig {
		runGetConfigCommand()
		return
	}

	// Handle get-architecture
	if *getArch {
		runGetArchitectureCommand()
		return
	}

	// Handle manual-backup
	if *manualBackup {
		runManualBackupCommand()
		return
	}

	// Handle logs
	if *logs {
		runLogsCommand()
		return
	}

	// No flags provided, show help
	printUsage()
}

func printVersion() {
	fmt.Println()
	fmt.Println(titleStyle.Render("servctl") + " - Home Server Provisioning CLI")
	fmt.Printf("  Version:    %s\n", Version)
	fmt.Printf("  Built:      %s\n", BuildTime)
	fmt.Printf("  Go version: %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}

func printUsage() {
	fmt.Println()
	fmt.Println(titleStyle.Render("servctl") + " - Home Server Provisioning CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s    %s\n", cmdStyle.Render("servctl -start-setup"), descStyle.Render("Launch interactive installation wizard"))
	fmt.Printf("  %s          %s\n", cmdStyle.Render("servctl -status"), descStyle.Render("Display current system status"))
	fmt.Printf("  %s       %s\n", cmdStyle.Render("servctl -preflight"), descStyle.Render("Run pre-flight checks only"))
	fmt.Printf("  %s      %s\n", cmdStyle.Render("servctl -get-config"), descStyle.Render("Display current configuration"))
	fmt.Printf("  %s %s\n", cmdStyle.Render("servctl -get-architecture"), descStyle.Render("Display folder structure"))
	fmt.Printf("  %s   %s\n", cmdStyle.Render("servctl -manual-backup"), descStyle.Render("Trigger immediate backup"))
	fmt.Printf("  %s            %s\n", cmdStyle.Render("servctl -logs"), descStyle.Render("Display service logs"))
	fmt.Printf("  %s         %s\n", cmdStyle.Render("servctl -version"), descStyle.Render("Display version info"))
	fmt.Println()
	fmt.Println("Options:")
	fmt.Printf("  %s         %s\n", cmdStyle.Render("-dry-run"), descStyle.Render("Preview changes without making them"))
	fmt.Println()
}

func runPreflightChecks() {
	fmt.Println()

	// Check if running on Linux
	if runtime.GOOS != "linux" {
		msg := boxStyle.Render(
			warningStyle.Render("⚠️  Platform Warning") + "\n\n" +
				"servctl is designed for Ubuntu Linux.\n" +
				"Current platform: " + runtime.GOOS + "/" + runtime.GOARCH + "\n\n" +
				"Some checks may not work correctly on this platform.")
		fmt.Println(msg)
		fmt.Println()
	}

	// Run all preflight checks
	results := preflight.RunAllPreflightChecks()

	// Render results
	fmt.Print(tui.RenderPreflightResults(results))
	fmt.Println()

	// Exit with appropriate code
	if preflight.HasBlockers(results) {
		os.Exit(1)
	}
}

func runSetupWizard(dryRun bool) {
	fmt.Println()

	// Get current user and paths
	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	infraRoot := filepath.Join(homeDir, "infra")

	// Initialize logger
	var logger *utils.Logger
	if !dryRun {
		var err error
		logger, err = utils.NewLogger(filepath.Join(infraRoot, "logs"))
		if err != nil {
			fmt.Println(warningStyle.Render("Warning: Could not initialize logger: " + err.Error()))
		} else {
			defer logger.Close()
			logger.Info("Starting servctl setup wizard")
		}
	}

	// Banner
	banner := `
   _____ ______ ______     _______ _     
  / ____|  ____|  __ \ \   / / ____| |    
 | (___ | |__  | |__) \ \_/ / |    | |    
  \___ \|  __| |  _  / \   /| |    | |    
  ____) | |____| | \ \  | | | |____| |____
 |_____/|______|_|  \_\ |_|  \_____|______|
                                           
   Home Server Provisioning CLI
`
	fmt.Println(titleStyle.Render(banner))

	if dryRun {
		fmt.Println(warningStyle.Render("🔍 DRY RUN MODE - No changes will be made"))
		fmt.Println()
	}

	// Phase 1: Preflight checks with auto-installation
	fmt.Println(sectionStyle.Render("📋 Phase 1: System Preparation"))
	fmt.Println()

	// Check for missing dependencies first
	missing := preflight.GetMissingDependencies()
	if len(missing) > 0 {
		fmt.Println(descStyle.Render("Found missing dependencies, installing..."))
		fmt.Println()

		for _, dep := range missing {
			fmt.Printf("  📦 Installing %s...", dep.Name)
			if dryRun {
				fmt.Println(successStyle.Render(" [DRY RUN]"))
			} else {
				err := preflight.InstallDependency(dep)
				if err != nil {
					fmt.Println(errorStyle.Render(" FAILED"))
					fmt.Printf("    Error: %v\n", err)
				} else {
					fmt.Println(successStyle.Render(" ✓"))
				}
			}
		}
		fmt.Println()
	}

	// Run preflight checks with auto-fix
	results, installResults, _ := preflight.RunPreflightWithAutoFix(dryRun)
	fmt.Print(tui.RenderPreflightResults(results))
	fmt.Println()

	// Show installation summary if any dependencies were installed
	if len(installResults) > 0 {
		successCount := 0
		for _, r := range installResults {
			if r.Success {
				successCount++
			}
		}
		fmt.Printf("  %s Installed %d/%d dependencies\n\n",
			successStyle.Render("✓"),
			successCount,
			len(installResults))
	}

	if preflight.HasBlockers(results) {
		fmt.Println(errorStyle.Render("Critical issues remain. Please resolve manually:"))
		for _, r := range results {
			if r.Status == preflight.StatusFail {
				fmt.Printf("  ✗ %s: %s\n", r.Name, r.Message)
				for _, d := range r.Details {
					if d != "" {
						fmt.Printf("    %s\n", d)
					}
				}
			}
		}
		os.Exit(1)
	}

	// Interactive: Prompt for static IP configuration if DHCP detected
	reader := bufio.NewReader(os.Stdin)
	preflight.PromptStaticIPSetup(reader, dryRun)

	if !promptContinue("Continue to disk selection?") {
		fmt.Println("Setup cancelled.")
		return
	}

	// Phase 2: Disk Selection
	fmt.Println()
	fmt.Println(sectionStyle.Render("💾 Phase 2: Storage Configuration"))
	fmt.Println()

	disks, err := storage.DiscoverDisks()
	if err != nil {
		fmt.Println(warningStyle.Render("Error discovering disks: " + err.Error()))
	}

	// Show discovered disks first
	if len(disks) > 0 {
		fmt.Print(tui.RenderDiskDiscovery(disks))
		fmt.Println()
	}

	// Generate and display storage strategy recommendations
	sysInfo := storage.GetSystemInfo()
	strategies := storage.GenerateStrategies(disks, sysInfo)

	if len(strategies) > 0 {
		fmt.Print(tui.RenderStrategies(strategies))
		fmt.Println()

		// Interactive strategy selection
		selectedStrategy, ok := storage.PromptStrategySelection(reader, strategies)
		if !ok {
			fmt.Println(descStyle.Render("  Skipping storage configuration."))
		} else {
			fmt.Println()
			fmt.Printf("  Selected: %s\n", successStyle.Render(selectedStrategy.Name))

			// Show preview and offer customization
			strategyConfig, proceed := storage.PromptStrategyConfirmation(reader, selectedStrategy)
			if !proceed {
				fmt.Println(descStyle.Render("  Skipping storage configuration."))
			} else {
				// Confirm destructive operation
				needsConfirmation := len(selectedStrategy.Disks) > 0
				if needsConfirmation && !dryRun {
					confirmed := true
					for _, disk := range selectedStrategy.Disks {
						if !storage.PromptEraseConfirmation(reader, disk) {
							confirmed = false
							fmt.Println(warningStyle.Render("  Operation cancelled."))
							break
						}
					}

					if confirmed {
						// Apply the strategy with user config
						results := storage.ApplyStrategy(selectedStrategy, strategyConfig.ToConfigMap(), dryRun)
						fmt.Println()
						for _, r := range results {
							if r.Success {
								fmt.Println(successStyle.Render("  ✓ " + r.Message))
							} else {
								fmt.Println(errorStyle.Render("  ✗ " + r.Message))
							}
						}
					}
				} else if dryRun {
					// Dry run - show what would happen
					results := storage.ApplyStrategy(selectedStrategy, strategyConfig.ToConfigMap(), true)
					fmt.Println()
					fmt.Println(descStyle.Render("  [Dry Run] Operations that would be performed:"))
					for _, r := range results {
						fmt.Println("    → " + r.Message)
					}
				}
			}
		}
	} else {
		fmt.Println(warningStyle.Render("No storage strategies available for your hardware."))
	}

	if !promptContinue("Continue to directory setup?") {
		fmt.Println("Setup cancelled.")
		return
	}

	// Phase 3: Directory Structure
	fmt.Println()
	fmt.Println(sectionStyle.Render("📁 Phase 3: Directory Structure"))
	fmt.Println()

	// Interactive service selection
	serviceSelection := directory.PromptServiceSelection(reader)
	fmt.Println()

	// Allow customization of data root
	dataRoot := "/mnt/data"
	fmt.Print("Press Enter to use default paths, or 'c' to customize: ")
	customInput, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(customInput)) == "c" {
		dataRoot = directory.PromptCustomDataRoot(reader, dataRoot)
	}

	// Generate directories based on selection
	allDirs := directory.GetDirectoriesForServices(serviceSelection, homeDir, dataRoot)

	fmt.Println()
	fmt.Printf("Creating directories for: %s\n", strings.Join(serviceSelection.SelectedNames(), ", "))
	fmt.Println()
	fmt.Print(tui.RenderDirectoryPlan(allDirs))
	fmt.Println()

	if !dryRun {
		fmt.Println(descStyle.Render("Creating directories..."))
		var results []directory.DirectoryResult
		for _, spec := range allDirs {
			results = append(results, directory.CreateDirectory(spec, dryRun))
		}
		fmt.Print(tui.RenderDirectoryComplete(results, nil))
	} else {
		fmt.Println(warningStyle.Render("[DRY RUN] Would create directories listed above"))
	}

	if !promptContinue("Continue to service configuration?") {
		fmt.Println("Setup cancelled.")
		return
	}

	// Phase 4: Service Composition
	fmt.Println()
	fmt.Println(sectionStyle.Render("🐳 Phase 4: Service Configuration"))
	fmt.Println()

	config := compose.DefaultConfig()
	config.AutoFillDefaults()
	config.InfraRoot = filepath.Join(homeDir, "infra")
	config.DataRoot = dataRoot

	// Detect host IP
	if ip, err := compose.DetectHostIP(); err == nil {
		config.HostIP = ip
		fmt.Printf("Detected Host IP: %s\n", successStyle.Render(ip))
	}

	// Generate credentials
	config.NextcloudAdminPass = compose.GenerateDBPassword()

	// Interactive config confirmation
	config, proceed := compose.PromptConfigConfirmation(reader, config)
	if !proceed {
		fmt.Println(descStyle.Render("  Skipping Docker Compose generation."))
	} else {
		composeDir := filepath.Join(homeDir, "infra", "compose")
		if !dryRun {
			fmt.Println(descStyle.Render("Generating Docker Compose files..."))
			if err := compose.WriteAllConfigFiles(config, composeDir, dryRun); err != nil {
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
			} else {
				fmt.Println(tui.RenderComposeGenerated(composeDir))
			}
		} else {
			fmt.Println(warningStyle.Render("[DRY RUN] Would generate Docker Compose files"))
			compose.WriteAllConfigFiles(config, composeDir, dryRun)
		}
	}

	if !promptContinue("Continue to maintenance setup?") {
		fmt.Println("Setup cancelled.")
		return
	}

	// Phase 5: Maintenance
	fmt.Println()
	fmt.Println(sectionStyle.Render("🔧 Phase 5: Maintenance Scripts"))
	fmt.Println()

	// Interactive script selection
	scriptSelection := maintenance.PromptScriptSelection(reader)
	fmt.Println()

	mConfig := maintenance.DefaultScriptConfig()
	mConfig.LogDir = filepath.Join(homeDir, "infra", "logs")
	mConfig.InfraRoot = filepath.Join(homeDir, "infra")
	mConfig.DataRoot = dataRoot

	// Prompt for backup schedule if backup selected
	if scriptSelection.DailyBackup {
		schedule := maintenance.PromptBackupSchedule(reader)
		fmt.Printf("  Backup schedule: %s\n", schedule)
	}

	// Prompt for webhook URL
	webhookURL := maintenance.PromptWebhookURL(reader)
	if webhookURL != "" {
		mConfig.WebhookURL = webhookURL
		fmt.Println(successStyle.Render("  ✓ Webhook configured"))
	}
	fmt.Println()

	// Generate selected scripts only
	scripts, _ := maintenance.GetScriptsForSelection(scriptSelection, mConfig)
	if len(scripts) > 0 {
		fmt.Print(tui.RenderAllScripts(scripts))
		fmt.Println()

		scriptsDir := filepath.Join(homeDir, "infra", "scripts")
		if !dryRun {
			fmt.Println(descStyle.Render("Generating maintenance scripts..."))
			for _, script := range scripts {
				maintenance.WriteScript(script, scriptsDir, dryRun)
			}
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ Generated %d scripts in %s", len(scripts), scriptsDir)))
		} else {
			fmt.Println(warningStyle.Render("[DRY RUN] Would generate scripts in " + scriptsDir))
		}
	} else {
		fmt.Println(descStyle.Render("  No scripts selected."))
	}

	// Final Summary - Mission Report
	fmt.Println()

	missionReport := report.NewMissionReport(config, infraRoot)
	missionReport.DirsCreated = len(allDirs)
	missionReport.ScriptsGen = len(scripts)

	if dryRun {
		fmt.Print(report.RenderCompactReport(missionReport))
		fmt.Println()
		fmt.Println(warningStyle.Render("DRY RUN complete. No actual changes were made."))
	} else {
		fmt.Print(report.RenderMissionReport(missionReport))
	}

	// Log completion
	if logger != nil {
		logger.Info("Setup completed successfully")
	}
}

func runStatusCommand() {
	fmt.Println()
	fmt.Println(sectionStyle.Render("📊 System Status"))
	fmt.Println()

	// Docker status
	fmt.Println(titleStyle.Render("Docker Containers:"))
	fmt.Println()

	cmd := exec.Command("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(warningStyle.Render("Docker not available or no containers running"))
	} else {
		fmt.Println(string(output))
	}

	// Disk usage
	fmt.Println(titleStyle.Render("Disk Usage:"))
	fmt.Println()

	paths := []string{"/mnt/data", "/mnt/backup", "/"}
	for _, path := range paths {
		var stat struct {
			Total uint64
			Free  uint64
		}
		// Simple check if path exists
		if _, err := os.Stat(path); err == nil {
			cmd := exec.Command("df", "-h", path)
			output, _ := cmd.Output()
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				fmt.Printf("  %s: %s\n", path, strings.TrimSpace(lines[1]))
			}
		}
		_ = stat // avoid unused warning
	}
	fmt.Println()

	// SMART status (if available)
	fmt.Println(titleStyle.Render("Drive Health:"))
	fmt.Println()

	smartCmd := exec.Command("sudo", "smartctl", "--scan")
	smartOutput, err := smartCmd.Output()
	if err != nil {
		fmt.Println(descStyle.Render("SMART not available (try running with sudo)"))
	} else {
		drives := strings.Split(strings.TrimSpace(string(smartOutput)), "\n")
		for _, drive := range drives {
			parts := strings.Fields(drive)
			if len(parts) > 0 {
				healthCmd := exec.Command("sudo", "smartctl", "-H", parts[0])
				healthOutput, _ := healthCmd.Output()
				if strings.Contains(string(healthOutput), "PASSED") {
					fmt.Printf("  %s: %s\n", parts[0], successStyle.Render("PASSED"))
				} else {
					fmt.Printf("  %s: %s\n", parts[0], warningStyle.Render("CHECK REQUIRED"))
				}
			}
		}
	}
	fmt.Println()
}

func runGetConfigCommand() {
	fmt.Println()
	fmt.Println(sectionStyle.Render("⚙️  Current Configuration"))
	fmt.Println()

	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	composeDir := filepath.Join(homeDir, "infra", "compose")

	// Read .env
	envPath := filepath.Join(composeDir, ".env")
	if content, err := os.ReadFile(envPath); err == nil {
		fmt.Println(titleStyle.Render(".env Configuration:"))
		fmt.Println()

		// Mask passwords
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToUpper(line), "PASSWORD") ||
				strings.Contains(strings.ToUpper(line), "SECRET") ||
				strings.Contains(strings.ToUpper(line), "TOKEN") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					fmt.Printf("  %s=%s\n", parts[0], strings.Repeat("*", len(parts[1])))
				}
			} else if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
				fmt.Printf("  %s\n", line)
			}
		}
		fmt.Println()
	} else {
		fmt.Println(warningStyle.Render("No .env file found at " + envPath))
	}

	// Check docker-compose.yml
	composePath := filepath.Join(composeDir, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		fmt.Println(successStyle.Render("✓ docker-compose.yml exists"))
		fmt.Printf("  Path: %s\n", composePath)
	} else {
		fmt.Println(warningStyle.Render("No docker-compose.yml found"))
	}
	fmt.Println()
}

func runGetArchitectureCommand() {
	fmt.Println()
	fmt.Println(sectionStyle.Render("🏗️  System Architecture"))
	fmt.Println()

	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir

	// Directory tree
	fmt.Print(tui.RenderDirectoryTree(homeDir, "/mnt/data"))
	fmt.Println()

	// Service relationships
	fmt.Println(titleStyle.Render("Service Relationships:"))
	fmt.Println()

	services := `
  ┌─────────────────────────────────────────────────────────────┐
  │                     servctl-network                         │
  ├─────────────────────────────────────────────────────────────┤
  │                                                             │
  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
  │  │   Immich    │───▶│    Redis    │    │  Nextcloud  │     │
  │  │   :2283     │    │   (cache)   │    │   :8080     │     │
  │  └──────┬──────┘    └─────────────┘    └──────┬──────┘     │
  │         │                                      │            │
  │         ▼                                      ▼            │
  │  ┌─────────────┐                        ┌─────────────┐     │
  │  │ PostgreSQL  │                        │  MariaDB    │     │
  │  │  (Immich)   │                        │ (Nextcloud) │     │
  │  └─────────────┘                        └─────────────┘     │
  │                                                             │
  │  ┌─────────────┐    ┌─────────────┐                        │
  │  │   Glances   │    │    Diun     │                        │
  │  │   :61208    │    │  (updates)  │                        │
  │  └─────────────┘    └─────────────┘                        │
  │                                                             │
  └─────────────────────────────────────────────────────────────┘
`
	fmt.Println(services)
}

func runManualBackupCommand() {
	fmt.Println()
	fmt.Println(sectionStyle.Render("💾 Manual Backup"))
	fmt.Println()

	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	scriptPath := filepath.Join(homeDir, "infra", "scripts", "daily_backup.sh")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Println(errorStyle.Render("Backup script not found: " + scriptPath))
		fmt.Println(descStyle.Render("Run 'servctl -start-setup' first to generate maintenance scripts."))
		return
	}

	fmt.Println("Running backup script...")
	fmt.Println(descStyle.Render("Script: " + scriptPath))
	fmt.Println()

	cmd := exec.Command("sudo", "bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(errorStyle.Render("Backup failed: " + err.Error()))
	} else {
		fmt.Println()
		fmt.Println(successStyle.Render("✅ Backup completed successfully!"))
	}
}

func runLogsCommand() {
	fmt.Println()
	fmt.Println(sectionStyle.Render("📋 Service Logs"))
	fmt.Println()

	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	composeDir := filepath.Join(homeDir, "infra", "compose")

	// Check if docker-compose.yml exists
	if _, err := os.Stat(filepath.Join(composeDir, "docker-compose.yml")); os.IsNotExist(err) {
		fmt.Println(warningStyle.Render("No docker-compose.yml found"))
		fmt.Println(descStyle.Render("Run 'servctl -start-setup' first."))
		return
	}

	fmt.Println("Showing last 50 lines (Ctrl+C to exit)...")
	fmt.Println()

	cmd := exec.Command("docker", "compose", "-f",
		filepath.Join(composeDir, "docker-compose.yml"),
		"logs", "--tail=50", "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run interactively
	cmd.Run()
}

// promptContinue asks user to continue and returns true if yes
func promptContinue(message string) bool {
	fmt.Printf("\n%s [Y/n]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}
