// Package preflight provides system prerequisite checks for servctl.
// It verifies OS compatibility, user privileges, hardware capabilities,
// network connectivity, and required dependencies before server setup.
package preflight

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

// CheckResult represents the result of a preflight check
type CheckResult struct {
	Name    string
	Status  Status
	Message string
	Details []string
}

// Status represents the status of a check
type Status int

const (
	StatusPass Status = iota
	StatusWarn
	StatusFail
	StatusSkip
)

func (s Status) String() string {
	switch s {
	case StatusPass:
		return "PASS"
	case StatusWarn:
		return "WARN"
	case StatusFail:
		return "FAIL"
	case StatusSkip:
		return "SKIP"
	default:
		return "UNKNOWN"
	}
}

// OSInfo contains information about the operating system
type OSInfo struct {
	ID              string
	VersionID       string
	Name            string
	PrettyName      string
	VersionCodename string
}

// CheckOS verifies the system is running Ubuntu 22.04 LTS or later
func CheckOS() CheckResult {
	result := CheckResult{
		Name: "Operating System Check",
	}

	osInfo, err := parseOSRelease()
	if err != nil {
		result.Status = StatusFail
		result.Message = "Failed to read OS information"
		result.Details = append(result.Details, err.Error())
		return result
	}

	result.Details = append(result.Details, fmt.Sprintf("Detected: %s", osInfo.PrettyName))

	// Check if it's Ubuntu
	if strings.ToLower(osInfo.ID) != "ubuntu" {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Unsupported OS: %s. servctl requires Ubuntu 22.04 LTS or later.", osInfo.ID)
		return result
	}

	// Parse version number
	versionParts := strings.Split(osInfo.VersionID, ".")
	if len(versionParts) < 1 {
		result.Status = StatusFail
		result.Message = "Could not parse Ubuntu version"
		return result
	}

	majorVersion, err := strconv.Atoi(versionParts[0])
	if err != nil {
		result.Status = StatusFail
		result.Message = "Could not parse Ubuntu major version"
		return result
	}

	// Require Ubuntu 22.04 or later
	if majorVersion < 22 {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Ubuntu %s is not supported. Please upgrade to Ubuntu 22.04 LTS or later.", osInfo.VersionID)
		return result
	}

	result.Status = StatusPass
	result.Message = fmt.Sprintf("Ubuntu %s is supported", osInfo.VersionID)
	return result
}

// parseOSRelease reads and parses /etc/os-release
func parseOSRelease() (*OSInfo, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("cannot open /etc/os-release: %w", err)
	}
	defer file.Close()

	info := &OSInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"")

		switch key {
		case "ID":
			info.ID = value
		case "VERSION_ID":
			info.VersionID = value
		case "NAME":
			info.Name = value
		case "PRETTY_NAME":
			info.PrettyName = value
		case "VERSION_CODENAME":
			info.VersionCodename = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /etc/os-release: %w", err)
	}

	return info, nil
}

// CheckPrivileges verifies the user is not root but has sudo access
func CheckPrivileges() CheckResult {
	result := CheckResult{
		Name: "User Privileges Check",
	}

	currentUser, err := user.Current()
	if err != nil {
		result.Status = StatusFail
		result.Message = "Failed to get current user"
		result.Details = append(result.Details, err.Error())
		return result
	}

	result.Details = append(result.Details, fmt.Sprintf("Running as: %s (UID: %s)", currentUser.Username, currentUser.Uid))

	// Check if running as root
	if currentUser.Uid == "0" {
		result.Status = StatusWarn
		result.Message = "Running as root is not recommended"
		result.Details = append(result.Details, "Consider running as a regular user with sudo access")
		result.Details = append(result.Details, "This helps prevent accidental system damage")
		return result
	}

	// Check sudo access
	cmd := exec.Command("sudo", "-n", "true")
	err = cmd.Run()
	if err != nil {
		// Try with password prompt check
		cmd = exec.Command("sudo", "-v")
		cmd.Stdin = nil
		err = cmd.Run()
		if err != nil {
			result.Status = StatusFail
			result.Message = "User does not have sudo access"
			result.Details = append(result.Details, "servctl requires sudo privileges for system configuration")
			result.Details = append(result.Details, "Add user to sudoers or run: sudo usermod -aG sudo "+currentUser.Username)
			return result
		}
	}

	result.Status = StatusPass
	result.Message = fmt.Sprintf("User '%s' has sudo access", currentUser.Username)
	return result
}

// CheckHardware verifies hardware capabilities
func CheckHardware() CheckResult {
	result := CheckResult{
		Name: "Hardware Capabilities Check",
	}

	var warnings []string

	// Check for virtualization support (VT-x/AMD-V)
	virtSupported, virtDetails := checkVirtualization()
	result.Details = append(result.Details, virtDetails...)
	if !virtSupported {
		warnings = append(warnings, "Virtualization (VT-x/AMD-V) not detected")
	}

	// Check for Secure Boot
	secureBootEnabled, sbDetails := checkSecureBoot()
	result.Details = append(result.Details, sbDetails...)
	if secureBootEnabled {
		warnings = append(warnings, "Secure Boot is enabled - may cause issues with some drivers")
	}

	// Determine overall status
	if len(warnings) > 0 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("%d hardware warning(s) detected", len(warnings))
		result.Details = append(result.Details, "")
		result.Details = append(result.Details, "Warnings:")
		for _, w := range warnings {
			result.Details = append(result.Details, "  • "+w)
		}
	} else {
		result.Status = StatusPass
		result.Message = "Hardware checks passed"
	}

	return result
}

// checkVirtualization checks for VT-x/AMD-V support
func checkVirtualization() (bool, []string) {
	var details []string

	cmd := exec.Command("lscpu")
	output, err := cmd.Output()
	if err != nil {
		details = append(details, "Could not run lscpu to check virtualization")
		return false, details
	}

	outputStr := string(output)

	// Check for virtualization flags
	if strings.Contains(outputStr, "Virtualization:") {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Virtualization:") {
				details = append(details, strings.TrimSpace(line))
				return true, details
			}
		}
	}

	// Check /proc/cpuinfo for vmx (Intel) or svm (AMD)
	cpuInfo, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		cpuInfoStr := string(cpuInfo)
		if strings.Contains(cpuInfoStr, "vmx") {
			details = append(details, "Intel VT-x detected in CPU flags")
			return true, details
		}
		if strings.Contains(cpuInfoStr, "svm") {
			details = append(details, "AMD-V detected in CPU flags")
			return true, details
		}
	}

	details = append(details, "No virtualization support detected")
	return false, details
}

// checkSecureBoot checks if Secure Boot is enabled
func checkSecureBoot() (bool, []string) {
	var details []string

	// Check mokutil
	cmd := exec.Command("mokutil", "--sb-state")
	output, err := cmd.Output()
	if err == nil {
		state := strings.TrimSpace(string(output))
		details = append(details, fmt.Sprintf("Secure Boot: %s", state))
		return strings.Contains(strings.ToLower(state), "enabled"), details
	}

	// Fallback: check /sys/firmware/efi/efivars
	if _, err := os.Stat("/sys/firmware/efi/efivars"); err == nil {
		// System is UEFI, try to read SecureBoot variable
		sbData, err := os.ReadFile("/sys/firmware/efi/efivars/SecureBoot-*")
		if err == nil && len(sbData) > 0 {
			// Last byte indicates state
			if sbData[len(sbData)-1] == 1 {
				details = append(details, "Secure Boot: Enabled (detected via EFI vars)")
				return true, details
			}
		}
		details = append(details, "Secure Boot: Likely disabled (UEFI system)")
		return false, details
	}

	details = append(details, "Secure Boot: Not applicable (Legacy BIOS)")
	return false, details
}

// CheckConnectivity verifies network connectivity
func CheckConnectivity() CheckResult {
	result := CheckResult{
		Name: "Network Connectivity Check",
	}

	var passed int
	var failed int

	// Test 1: Ping external IP (Google DNS)
	pingOk, pingDetails := testPing("8.8.8.8")
	result.Details = append(result.Details, pingDetails)
	if pingOk {
		passed++
	} else {
		failed++
	}

	// Test 2: DNS Resolution
	dnsOk, dnsDetails := testDNS("google.com")
	result.Details = append(result.Details, dnsDetails)
	if dnsOk {
		passed++
	} else {
		failed++
	}

	// Test 3: HTTPS connectivity
	httpsOk, httpsDetails := testHTTPS("https://github.com")
	result.Details = append(result.Details, httpsDetails)
	if httpsOk {
		passed++
	} else {
		failed++
	}

	if failed > 0 {
		if passed == 0 {
			result.Status = StatusFail
			result.Message = "No network connectivity detected"
			result.Details = append(result.Details, "")
			result.Details = append(result.Details, "servctl requires internet access to download packages and Docker images")
		} else {
			result.Status = StatusWarn
			result.Message = fmt.Sprintf("Partial connectivity (%d/%d tests passed)", passed, passed+failed)
		}
	} else {
		result.Status = StatusPass
		result.Message = "Full network connectivity confirmed"
	}

	return result
}

// testPing tests ICMP connectivity to an IP
func testPing(ip string) (bool, string) {
	cmd := exec.Command("ping", "-c", "1", "-W", "3", ip)
	err := cmd.Run()
	if err != nil {
		return false, fmt.Sprintf("✗ Ping to %s failed", ip)
	}
	return true, fmt.Sprintf("✓ Ping to %s successful", ip)
}

// testDNS tests DNS resolution
func testDNS(hostname string) (bool, string) {
	_, err := net.LookupHost(hostname)
	if err != nil {
		return false, fmt.Sprintf("✗ DNS resolution for %s failed", hostname)
	}
	return true, fmt.Sprintf("✓ DNS resolution for %s successful", hostname)
}

// testHTTPS tests HTTPS connectivity
func testHTTPS(url string) (bool, string) {
	// Use curl with timeout
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "5", url)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Sprintf("✗ HTTPS connection to %s failed", url)
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode == "200" || statusCode == "301" || statusCode == "302" {
		return true, fmt.Sprintf("✓ HTTPS connection to %s successful (HTTP %s)", url, statusCode)
	}

	return false, fmt.Sprintf("✗ HTTPS connection to %s returned status %s", url, statusCode)
}

// Dependency represents a system dependency
type Dependency struct {
	Name        string
	Binary      string
	Package     string
	Criticality string // "blocker", "high", "recommended"
	InstallCmd  string
}

// GetRequiredDependencies returns the list of required dependencies
func GetRequiredDependencies() []Dependency {
	return []Dependency{
		{Name: "curl", Binary: "curl", Package: "curl", Criticality: "blocker", InstallCmd: "apt install -y curl"},
		{Name: "net-tools", Binary: "ifconfig", Package: "net-tools", Criticality: "recommended", InstallCmd: "apt install -y net-tools"},
		{Name: "Docker", Binary: "docker", Package: "docker-ce", Criticality: "blocker", InstallCmd: "curl -fsSL https://get.docker.com | sh"},
		{Name: "Docker Compose", Binary: "docker-compose", Package: "docker-compose-plugin", Criticality: "blocker", InstallCmd: "apt install -y docker-compose-plugin"},
		{Name: "hdparm", Binary: "hdparm", Package: "hdparm", Criticality: "recommended", InstallCmd: "apt install -y hdparm"},
		{Name: "smartmontools", Binary: "smartctl", Package: "smartmontools", Criticality: "recommended", InstallCmd: "apt install -y smartmontools"},
		{Name: "cron", Binary: "crontab", Package: "cron", Criticality: "high", InstallCmd: "apt install -y cron"},
		{Name: "UFW Firewall", Binary: "ufw", Package: "ufw", Criticality: "high", InstallCmd: "apt install -y ufw"},
		{Name: "lsblk", Binary: "lsblk", Package: "util-linux", Criticality: "blocker", InstallCmd: "apt install -y util-linux"},
		{Name: "mkfs.ext4", Binary: "mkfs.ext4", Package: "e2fsprogs", Criticality: "blocker", InstallCmd: "apt install -y e2fsprogs"},
	}
}

// CheckDependency checks if a single dependency is installed
func CheckDependency(dep Dependency) CheckResult {
	result := CheckResult{
		Name: fmt.Sprintf("Dependency: %s", dep.Name),
	}

	// Check if binary exists
	path, err := exec.LookPath(dep.Binary)
	if err != nil {
		switch dep.Criticality {
		case "blocker":
			result.Status = StatusFail
			result.Message = fmt.Sprintf("%s is REQUIRED but not installed", dep.Name)
		case "high":
			result.Status = StatusWarn
			result.Message = fmt.Sprintf("%s is highly recommended but not installed", dep.Name)
		default:
			result.Status = StatusWarn
			result.Message = fmt.Sprintf("%s is recommended but not installed", dep.Name)
		}
		result.Details = append(result.Details, fmt.Sprintf("Install with: sudo %s", dep.InstallCmd))
		return result
	}

	result.Status = StatusPass
	result.Message = fmt.Sprintf("%s is installed", dep.Name)
	result.Details = append(result.Details, fmt.Sprintf("Found at: %s", path))

	// Get version if possible
	version := getVersion(dep.Binary)
	if version != "" {
		result.Details = append(result.Details, fmt.Sprintf("Version: %s", version))
	}

	return result
}

// getVersion attempts to get the version of a binary
func getVersion(binary string) string {
	// Try common version flags
	for _, flag := range []string{"--version", "-v", "version"} {
		cmd := exec.Command(binary, flag)
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	}
	return ""
}

// CheckAllDependencies checks all required dependencies
func CheckAllDependencies() []CheckResult {
	var results []CheckResult
	deps := GetRequiredDependencies()

	for _, dep := range deps {
		results = append(results, CheckDependency(dep))
	}

	return results
}

// CheckDockerRunning verifies Docker daemon is running
func CheckDockerRunning() CheckResult {
	result := CheckResult{
		Name: "Docker Service Status",
	}

	// First check if docker binary exists
	_, err := exec.LookPath("docker")
	if err != nil {
		result.Status = StatusSkip
		result.Message = "Docker is not installed"
		result.Details = append(result.Details, "Install Docker first, then run this check again")
		return result
	}

	// Check if docker daemon is running
	cmd := exec.Command("docker", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Status = StatusFail
		result.Message = "Docker daemon is not running"
		result.Details = append(result.Details, "Start Docker with: sudo systemctl start docker")
		result.Details = append(result.Details, "Enable on boot: sudo systemctl enable docker")

		// Check if it's a permission issue
		if strings.Contains(string(output), "permission denied") {
			result.Details = append(result.Details, "")
			result.Details = append(result.Details, "Or add user to docker group:")
			result.Details = append(result.Details, "  sudo usermod -aG docker $USER")
			result.Details = append(result.Details, "  Then log out and back in")
		}
		return result
	}

	result.Status = StatusPass
	result.Message = "Docker daemon is running"

	// Parse some useful info from docker info
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Server Version:") ||
			strings.HasPrefix(line, "Storage Driver:") ||
			strings.HasPrefix(line, "Cgroup Driver:") {
			result.Details = append(result.Details, line)
		}
	}

	return result
}

// SystemUpdateResult contains the result of system update
type SystemUpdateResult struct {
	Success         bool
	PackagesUpdated int
	Errors          []string
	Duration        time.Duration
}

// RunSystemUpdate executes apt update and upgrade
func RunSystemUpdate(dryRun bool) (*SystemUpdateResult, error) {
	result := &SystemUpdateResult{}
	startTime := time.Now()

	// Run apt update
	updateCmd := exec.Command("sudo", "apt", "update")
	updateOutput, err := updateCmd.CombinedOutput()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("apt update failed: %s", string(updateOutput)))
		return result, errors.New("apt update failed")
	}

	if dryRun {
		// Just simulate upgrade
		upgradeCmd := exec.Command("sudo", "apt", "upgrade", "-s")
		output, err := upgradeCmd.CombinedOutput()
		if err != nil {
			result.Errors = append(result.Errors, "Failed to simulate upgrade")
			return result, err
		}

		// Parse number of packages
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "upgraded") && strings.Contains(line, "newly installed") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					count, _ := strconv.Atoi(parts[0])
					result.PackagesUpdated = count
				}
				break
			}
		}
	} else {
		// Actually run upgrade
		upgradeCmd := exec.Command("sudo", "apt", "upgrade", "-y")
		output, err := upgradeCmd.CombinedOutput()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("apt upgrade failed: %s", string(output)))
			return result, errors.New("apt upgrade failed")
		}

		// Parse output for package count
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "upgraded") && strings.Contains(line, "newly installed") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					count, _ := strconv.Atoi(parts[0])
					result.PackagesUpdated = count
				}
				break
			}
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = true
	return result, nil
}

// InstallDependency installs a missing dependency
func InstallDependency(dep Dependency) error {
	// Handle special case for Docker (uses script)
	if dep.Binary == "docker" {
		cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | sh")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Handle Docker Compose v2 (plugin)
	if dep.Binary == "docker-compose" {
		// Check if docker compose (v2) works
		cmd := exec.Command("docker", "compose", "version")
		if err := cmd.Run(); err == nil {
			return nil // Already installed as plugin
		}
		// Install as plugin
		installCmd := exec.Command("sudo", "apt", "install", "-y", "docker-compose-plugin")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		return installCmd.Run()
	}

	// Standard apt install
	cmd := exec.Command("sudo", "apt", "install", "-y", dep.Package)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AddUserToDockerGroup adds the current user to the docker group
func AddUserToDockerGroup() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	cmd := exec.Command("sudo", "usermod", "-aG", "docker", currentUser.Username)
	return cmd.Run()
}

// RunAllPreflightChecks runs all preflight checks and returns the results
func RunAllPreflightChecks() []CheckResult {
	var results []CheckResult

	// System checks
	results = append(results, CheckOS())
	results = append(results, CheckPrivileges())
	results = append(results, CheckHardware())
	results = append(results, CheckConnectivity())

	// Dependency checks
	results = append(results, CheckAllDependencies()...)

	// Docker service check
	results = append(results, CheckDockerRunning())

	return results
}

// HasBlockers checks if any results have blocking failures
func HasBlockers(results []CheckResult) bool {
	for _, r := range results {
		if r.Status == StatusFail {
			return true
		}
	}
	return false
}

// CountByStatus counts results by status
func CountByStatus(results []CheckResult) map[Status]int {
	counts := make(map[Status]int)
	for _, r := range results {
		counts[r.Status]++
	}
	return counts
}
