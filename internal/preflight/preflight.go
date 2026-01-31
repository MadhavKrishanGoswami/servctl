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

// NetworkConfig represents the IP configuration status
type NetworkConfig struct {
	Interface    string
	IPAddress    string
	IsStatic     bool
	IsDHCP       bool
	Gateway      string
	ConfigFile   string
	ConfigMethod string // "netplan", "networkd", "interfaces", "unknown"
}

// CheckStaticIP checks if the system has a static IP configuration
func CheckStaticIP() CheckResult {
	result := CheckResult{
		Name: "Static IP Configuration",
	}

	// Get current IP and interface
	config, err := detectNetworkConfig()
	if err != nil {
		result.Status = StatusWarn
		result.Message = "Could not determine IP configuration"
		result.Details = append(result.Details, err.Error())
		result.Details = append(result.Details, "Consider configuring a static IP for server stability")
		return result
	}

	result.Details = append(result.Details, fmt.Sprintf("Interface: %s", config.Interface))
	result.Details = append(result.Details, fmt.Sprintf("IP Address: %s", config.IPAddress))
	result.Details = append(result.Details, fmt.Sprintf("Config Method: %s", config.ConfigMethod))

	if config.IsStatic {
		result.Status = StatusPass
		result.Message = "Static IP is configured"
		if config.Gateway != "" {
			result.Details = append(result.Details, fmt.Sprintf("Gateway: %s", config.Gateway))
		}
		return result
	}

	if config.IsDHCP {
		result.Status = StatusWarn
		result.Message = "DHCP detected - Static IP recommended for servers"
		result.Details = append(result.Details, "")
		result.Details = append(result.Details, "⚠️  DHCP may cause IP changes that break Nextcloud access")
		result.Details = append(result.Details, "")
		result.Details = append(result.Details, "To configure static IP via Netplan:")
		result.Details = append(result.Details, fmt.Sprintf("  1. Edit: sudo nano /etc/netplan/01-static.yaml"))
		result.Details = append(result.Details, "  2. Add the following configuration:")
		result.Details = append(result.Details, "")
		result.Details = append(result.Details, "     network:")
		result.Details = append(result.Details, "       version: 2")
		result.Details = append(result.Details, fmt.Sprintf("       ethernets:"))
		result.Details = append(result.Details, fmt.Sprintf("         %s:", config.Interface))
		result.Details = append(result.Details, "           dhcp4: false")
		result.Details = append(result.Details, fmt.Sprintf("           addresses: [%s/24]", config.IPAddress))
		result.Details = append(result.Details, "           routes:")
		result.Details = append(result.Details, "             - to: default")
		result.Details = append(result.Details, "               via: <YOUR_GATEWAY_IP>")
		result.Details = append(result.Details, "           nameservers:")
		result.Details = append(result.Details, "             addresses: [8.8.8.8, 1.1.1.1]")
		result.Details = append(result.Details, "")
		result.Details = append(result.Details, "  3. Apply: sudo netplan apply")
		return result
	}

	result.Status = StatusWarn
	result.Message = "Could not determine if IP is static or DHCP"
	result.Details = append(result.Details, "Please verify your network configuration manually")
	return result
}

// StaticIPConfig holds the configuration for static IP setup
type StaticIPConfig struct {
	Interface  string
	IPAddress  string
	Subnet     string // e.g., "24"
	Gateway    string
	DNS1       string
	DNS2       string
	ConfigPath string
}

// PromptStaticIPSetup checks if DHCP and prompts user to configure static IP
// Returns true if static IP was configured, false otherwise
func PromptStaticIPSetup(reader *bufio.Reader, dryRun bool) bool {
	config, err := detectNetworkConfig()
	if err != nil {
		return false
	}

	// If already static, nothing to do
	if config.IsStatic {
		return false
	}

	// Only prompt if DHCP is detected
	if !config.IsDHCP {
		return false
	}

	fmt.Println()
	fmt.Println("┌────────────────────────────────────────────────────────────┐")
	fmt.Println("│  ⚠️  DHCP Detected - Static IP Recommended                 │")
	fmt.Println("└────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Printf("  Current Interface: %s\n", config.Interface)
	fmt.Printf("  Current IP:        %s\n", config.IPAddress)
	fmt.Println()
	fmt.Println("  A static IP ensures your server address never changes.")
	fmt.Println("  This is important for Nextcloud and mobile app access.")
	fmt.Println()
	fmt.Print("Would you like to configure a static IP now? [y/N]: ")

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("  Skipping static IP configuration.")
		fmt.Println("  You can configure it later using: sudo nano /etc/netplan/01-static.yaml")
		return false
	}

	// Get gateway IP
	fmt.Println()
	fmt.Println("  Enter your network details:")
	fmt.Println()

	// Try to detect gateway
	defaultGateway := detectDefaultGateway()
	if defaultGateway != "" {
		fmt.Printf("  Gateway IP [%s]: ", defaultGateway)
	} else {
		fmt.Print("  Gateway IP (e.g., 192.168.1.1): ")
	}
	gateway, _ := reader.ReadString('\n')
	gateway = strings.TrimSpace(gateway)
	if gateway == "" && defaultGateway != "" {
		gateway = defaultGateway
	}

	if gateway == "" {
		fmt.Println("  ✗ Gateway IP is required. Skipping static IP configuration.")
		return false
	}

	// Validate gateway
	if net.ParseIP(gateway) == nil {
		fmt.Println("  ✗ Invalid gateway IP. Skipping static IP configuration.")
		return false
	}

	// DNS servers
	fmt.Print("  Primary DNS [8.8.8.8]: ")
	dns1, _ := reader.ReadString('\n')
	dns1 = strings.TrimSpace(dns1)
	if dns1 == "" {
		dns1 = "8.8.8.8"
	}

	fmt.Print("  Secondary DNS [1.1.1.1]: ")
	dns2, _ := reader.ReadString('\n')
	dns2 = strings.TrimSpace(dns2)
	if dns2 == "" {
		dns2 = "1.1.1.1"
	}

	// Create config
	staticConfig := StaticIPConfig{
		Interface:  config.Interface,
		IPAddress:  config.IPAddress,
		Subnet:     "24",
		Gateway:    gateway,
		DNS1:       dns1,
		DNS2:       dns2,
		ConfigPath: "/etc/netplan/01-servctl-static.yaml",
	}

	// Show preview
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │  📋 Static IP Configuration Preview     │")
	fmt.Println("  ├─────────────────────────────────────────┤")
	fmt.Printf("  │  Interface:  %-26s │\n", staticConfig.Interface)
	fmt.Printf("  │  IP Address: %-26s │\n", staticConfig.IPAddress+"/"+staticConfig.Subnet)
	fmt.Printf("  │  Gateway:    %-26s │\n", staticConfig.Gateway)
	fmt.Printf("  │  DNS:        %-26s │\n", staticConfig.DNS1+", "+staticConfig.DNS2)
	fmt.Println("  └─────────────────────────────────────────┘")
	fmt.Println()

	fmt.Print("Apply this configuration? [y/N]: ")
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("  Configuration cancelled.")
		return false
	}

	if dryRun {
		fmt.Println("  [DRY RUN] Would create: " + staticConfig.ConfigPath)
		fmt.Println("  [DRY RUN] Would run: sudo netplan apply")
		return true
	}

	// Apply configuration
	err = applyStaticIPConfig(staticConfig)
	if err != nil {
		fmt.Printf("  ✗ Failed to apply configuration: %v\n", err)
		return false
	}

	fmt.Println()
	fmt.Println("  ✓ Static IP configured successfully!")
	fmt.Println("  ✓ Network configuration applied.")
	fmt.Println()

	return true
}

// detectDefaultGateway tries to detect the current gateway
func detectDefaultGateway() string {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse: "default via 192.168.1.1 dev eth0"
	fields := strings.Fields(string(output))
	for i, field := range fields {
		if field == "via" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

// applyStaticIPConfig creates the netplan config and applies it
func applyStaticIPConfig(config StaticIPConfig) error {
	// Generate netplan YAML
	netplanConfig := fmt.Sprintf(`# Generated by servctl - Static IP Configuration
# Do not edit manually unless you know what you're doing
network:
  version: 2
  renderer: networkd
  ethernets:
    %s:
      dhcp4: false
      addresses:
        - %s/%s
      routes:
        - to: default
          via: %s
      nameservers:
        addresses:
          - %s
          - %s
`, config.Interface, config.IPAddress, config.Subnet, config.Gateway, config.DNS1, config.DNS2)

	// Write to file (requires root)
	err := os.WriteFile(config.ConfigPath, []byte(netplanConfig), 0644)
	if err != nil {
		// Try with sudo
		cmd := exec.Command("sudo", "tee", config.ConfigPath)
		cmd.Stdin = strings.NewReader(netplanConfig)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to write netplan config: %w", err)
		}
	}

	fmt.Println("  → Created: " + config.ConfigPath)

	// Apply netplan
	fmt.Println("  → Applying netplan configuration...")
	cmd := exec.Command("sudo", "netplan", "apply")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply netplan: %w", err)
	}

	return nil
}

// detectNetworkConfig detects the current network configuration
func detectNetworkConfig() (*NetworkConfig, error) {
	config := &NetworkConfig{
		ConfigMethod: "unknown",
	}

	// Get the primary interface and IP
	iface, ip, err := getPrimaryInterface()
	if err != nil {
		return nil, fmt.Errorf("could not detect primary interface: %w", err)
	}
	config.Interface = iface
	config.IPAddress = ip

	// Check Netplan configuration (Ubuntu 17.10+)
	if checkNetplanConfig(config) {
		config.ConfigMethod = "netplan"
		return config, nil
	}

	// Check /etc/network/interfaces (older systems)
	if checkNetworkInterfaces(config) {
		config.ConfigMethod = "interfaces"
		return config, nil
	}

	// Check if NetworkManager is managing
	if checkNetworkManager(config) {
		config.ConfigMethod = "networkmanager"
		return config, nil
	}

	// Default: assume DHCP if we can't determine
	config.IsDHCP = true
	return config, nil
}

// getPrimaryInterface returns the primary network interface and its IP
func getPrimaryInterface() (string, string, error) {
	// Get the default route interface
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	// Parse: "default via 192.168.1.1 dev eth0 proto dhcp src 192.168.1.100"
	fields := strings.Fields(string(output))
	var iface string
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			iface = fields[i+1]
			break
		}
	}

	if iface == "" {
		return "", "", fmt.Errorf("could not find default interface")
	}

	// Get the IP address for this interface
	cmd = exec.Command("ip", "-4", "addr", "show", iface)
	output, err = cmd.Output()
	if err != nil {
		return iface, "", err
	}

	// Parse: "inet 192.168.1.100/24 brd ..."
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Remove CIDR suffix
				ip := strings.Split(parts[1], "/")[0]
				return iface, ip, nil
			}
		}
	}

	return iface, "", fmt.Errorf("could not find IP for interface %s", iface)
}

// checkNetplanConfig checks Netplan configuration files
func checkNetplanConfig(config *NetworkConfig) bool {
	// Check /etc/netplan/*.yaml files
	files, err := os.ReadDir("/etc/netplan")
	if err != nil {
		return false
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		path := "/etc/netplan/" + file.Name()
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		contentStr := string(content)
		config.ConfigFile = path

		// Check if this file configures our interface
		if !strings.Contains(contentStr, config.Interface) {
			continue
		}

		// Check for dhcp4: true/yes or dhcp4: false/no
		if strings.Contains(contentStr, "dhcp4: true") || strings.Contains(contentStr, "dhcp4: yes") {
			config.IsDHCP = true
			return true
		}

		if strings.Contains(contentStr, "dhcp4: false") || strings.Contains(contentStr, "dhcp4: no") {
			// Check if there are static addresses
			if strings.Contains(contentStr, "addresses:") {
				config.IsStatic = true
				return true
			}
		}

		// Check for static addresses without explicit dhcp4: false
		if strings.Contains(contentStr, "addresses:") && strings.Contains(contentStr, config.Interface) {
			config.IsStatic = true
			return true
		}
	}

	return false
}

// checkNetworkInterfaces checks /etc/network/interfaces
func checkNetworkInterfaces(config *NetworkConfig) bool {
	content, err := os.ReadFile("/etc/network/interfaces")
	if err != nil {
		return false
	}

	contentStr := string(content)
	config.ConfigFile = "/etc/network/interfaces"

	// Look for our interface
	if !strings.Contains(contentStr, config.Interface) {
		return false
	}

	// Check for "iface <interface> inet static" or "iface <interface> inet dhcp"
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "iface "+config.Interface) {
			if strings.Contains(line, "inet static") {
				config.IsStatic = true
				return true
			}
			if strings.Contains(line, "inet dhcp") {
				config.IsDHCP = true
				return true
			}
		}
	}

	return false
}

// checkNetworkManager checks if NetworkManager is managing the interface
func checkNetworkManager(config *NetworkConfig) bool {
	// Check if NetworkManager is running
	cmd := exec.Command("systemctl", "is-active", "NetworkManager")
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) != "active" {
		return false
	}

	// Check connection details using nmcli
	cmd = exec.Command("nmcli", "-t", "-f", "GENERAL.CONNECTION", "device", "show", config.Interface)
	output, err = cmd.Output()
	if err != nil {
		return false
	}

	connName := strings.TrimPrefix(strings.TrimSpace(string(output)), "GENERAL.CONNECTION:")
	if connName == "" || connName == "--" {
		return false
	}

	// Check if the connection uses DHCP
	cmd = exec.Command("nmcli", "-t", "-f", "ipv4.method", "connection", "show", connName)
	output, err = cmd.Output()
	if err != nil {
		return false
	}

	method := strings.TrimPrefix(strings.TrimSpace(string(output)), "ipv4.method:")
	config.ConfigFile = "NetworkManager: " + connName

	if method == "auto" {
		config.IsDHCP = true
		return true
	}
	if method == "manual" {
		config.IsStatic = true
		return true
	}

	return false
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
		{Name: "Docker Compose", Binary: "docker compose", Package: "docker-compose", Criticality: "blocker", InstallCmd: "apt install -y docker-compose"},
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

	var isInstalled bool
	var path string

	// Special handling for Docker Compose v2 (it's a plugin, not a standalone binary)
	if dep.Binary == "docker compose" {
		cmd := exec.Command("docker", "compose", "version")
		output, err := cmd.Output()
		if err == nil {
			isInstalled = true
			path = "docker compose (plugin)"
			// Extract version from output
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				result.Details = append(result.Details, fmt.Sprintf("Version: %s", strings.TrimSpace(lines[0])))
			}
		}
	} else {
		// Standard binary check
		var err error
		path, err = exec.LookPath(dep.Binary)
		isInstalled = err == nil
	}

	if !isInstalled {
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

	// Get version if possible (for non-plugin binaries)
	if dep.Binary != "docker compose" {
		version := getVersion(dep.Binary)
		if version != "" {
			result.Details = append(result.Details, fmt.Sprintf("Version: %s", version))
		}
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
	if dep.Binary == "docker compose" {
		// Check if docker compose (v2) works
		cmd := exec.Command("docker", "compose", "version")
		if err := cmd.Run(); err == nil {
			return nil // Already installed as plugin
		}
		// Install docker-compose
		installCmd := exec.Command("sudo", "apt", "install", "-y", "docker-compose")
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
	results = append(results, CheckStaticIP())

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

// InstallResult represents the result of installing a dependency
type InstallResult struct {
	Dependency Dependency
	Success    bool
	Error      error
	Duration   time.Duration
}

// GetMissingDependencies returns dependencies that are not installed
func GetMissingDependencies() []Dependency {
	var missing []Dependency
	deps := GetRequiredDependencies()

	for _, dep := range deps {
		// Special check for Docker Compose v2 (it's a plugin, not a standalone binary)
		if dep.Binary == "docker compose" {
			cmd := exec.Command("docker", "compose", "version")
			if cmd.Run() == nil {
				continue // docker compose v2 is installed
			}
			missing = append(missing, dep)
			continue
		}

		// Standard binary check
		_, err := exec.LookPath(dep.Binary)
		if err != nil {
			missing = append(missing, dep)
		}
	}

	return missing
}

// InstallAllMissingDependencies installs all missing dependencies
func InstallAllMissingDependencies(dryRun bool) []InstallResult {
	var results []InstallResult
	missing := GetMissingDependencies()

	// First run apt update
	if !dryRun && len(missing) > 0 {
		updateCmd := exec.Command("sudo", "apt", "update")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		updateCmd.Run() // Ignore error, continue anyway
	}

	// Install each missing dependency
	for _, dep := range missing {
		result := InstallResult{Dependency: dep}
		startTime := time.Now()

		if dryRun {
			result.Success = true
			result.Duration = 0
		} else {
			err := InstallDependency(dep)
			result.Error = err
			result.Success = err == nil
			result.Duration = time.Since(startTime)
		}

		results = append(results, result)
	}

	return results
}

// EnableAndStartDocker enables and starts the Docker service
func EnableAndStartDocker() error {
	// Enable Docker
	enableCmd := exec.Command("sudo", "systemctl", "enable", "docker")
	if err := enableCmd.Run(); err != nil {
		return fmt.Errorf("failed to enable docker: %w", err)
	}

	// Start Docker
	startCmd := exec.Command("sudo", "systemctl", "start", "docker")
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start docker: %w", err)
	}

	// Wait for Docker to be ready
	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("docker", "info")
		if checkCmd.Run() == nil {
			return nil // Docker is ready
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("docker did not start within 30 seconds")
}

// RunPreflightWithAutoFix runs preflight checks and automatically fixes issues
func RunPreflightWithAutoFix(dryRun bool) ([]CheckResult, []InstallResult, error) {
	var installResults []InstallResult

	// First run: check what's missing
	results := RunAllPreflightChecks()

	// Check for missing dependencies that need installation
	missing := GetMissingDependencies()

	if len(missing) > 0 {
		// Install missing dependencies
		installResults = InstallAllMissingDependencies(dryRun)

		// Re-run checks after installation
		if !dryRun {
			results = RunAllPreflightChecks()
		}
	}

	// Check if Docker needs to be started
	for _, r := range results {
		if r.Name == "Docker Service Status" && r.Status == StatusFail {
			if !dryRun {
				if err := EnableAndStartDocker(); err == nil {
					// Update the result
					for i := range results {
						if results[i].Name == "Docker Service Status" {
							results[i] = CheckDockerRunning()
						}
					}
				}
			}
		}
	}

	// Add user to docker group if needed
	dockerCheck := CheckDockerRunning()
	if dockerCheck.Status == StatusFail {
		for _, detail := range dockerCheck.Details {
			if strings.Contains(detail, "permission denied") || strings.Contains(strings.ToLower(detail), "permission") {
				if !dryRun {
					AddUserToDockerGroup()
				}
			}
		}
	}

	return results, installResults, nil
}

// SystemSetup performs complete system setup (update + install deps)
func SystemSetup(dryRun bool) error {
	// Step 1: Update system
	_, err := RunSystemUpdate(dryRun)
	if err != nil {
		return fmt.Errorf("system update failed: %w", err)
	}

	// Step 2: Install missing dependencies
	installResults := InstallAllMissingDependencies(dryRun)
	for _, r := range installResults {
		if !r.Success {
			return fmt.Errorf("failed to install %s: %w", r.Dependency.Name, r.Error)
		}
	}

	// Step 3: Enable and start Docker
	if !dryRun {
		// Check if docker was just installed
		if _, err := exec.LookPath("docker"); err == nil {
			if err := EnableAndStartDocker(); err != nil {
				return fmt.Errorf("failed to start docker: %w", err)
			}
			AddUserToDockerGroup()
		}
	}

	return nil
}
