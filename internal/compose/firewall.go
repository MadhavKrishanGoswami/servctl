package compose

import (
	"fmt"
	"os/exec"
	"strings"
)

// FirewallRule represents a UFW firewall rule
type FirewallRule struct {
	Port        int
	Protocol    string // "tcp" or "udp"
	Service     string // Human-readable service name
	Description string
	Required    bool // If true, failure is critical
}

// GetDefaultFirewallRules returns the firewall rules for servctl services
func GetDefaultFirewallRules() []FirewallRule {
	return []FirewallRule{
		{
			Port:        22,
			Protocol:    "tcp",
			Service:     "SSH",
			Description: "Secure Shell access (CRITICAL - must be first!)",
			Required:    true,
		},
		{
			Port:        2283,
			Protocol:    "tcp",
			Service:     "Immich",
			Description: "Photo & video management web UI",
			Required:    true,
		},
		{
			Port:        8080,
			Protocol:    "tcp",
			Service:     "Nextcloud",
			Description: "File sync & share web UI",
			Required:    true,
		},
		{
			Port:        61208,
			Protocol:    "tcp",
			Service:     "Glances",
			Description: "System monitoring (consider limiting to local network)",
			Required:    false,
		},
	}
}

// IsUFWInstalled checks if UFW is available
func IsUFWInstalled() bool {
	_, err := exec.LookPath("ufw")
	return err == nil
}

// IsUFWEnabled checks if UFW is currently enabled
func IsUFWEnabled() (bool, error) {
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check UFW status: %w", err)
	}
	return strings.Contains(string(output), "Status: active"), nil
}

// AllowPort adds a UFW allow rule for a port
func AllowPort(port int, protocol string, dryRun bool) error {
	rule := fmt.Sprintf("%d/%s", port, protocol)

	if dryRun {
		fmt.Printf("[DRY RUN] Would execute: ufw allow %s\n", rule)
		return nil
	}

	cmd := exec.Command("ufw", "allow", rule)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to allow port %d: %s: %w", port, string(output), err)
	}

	return nil
}

// AllowSSH ensures SSH access is allowed (CRITICAL: must be done first!)
func AllowSSH(dryRun bool) error {
	if dryRun {
		fmt.Println("[DRY RUN] Would execute: ufw allow ssh")
		return nil
	}

	// Try both 'ssh' and explicit port 22
	cmd := exec.Command("ufw", "allow", "ssh")
	if _, err := cmd.CombinedOutput(); err != nil {
		// Fallback to explicit port
		cmd = exec.Command("ufw", "allow", "22/tcp")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to allow SSH: %s: %w", string(output), err)
		}
	}

	return nil
}

// EnableUFW enables the firewall (DANGER: ensure SSH is allowed first!)
func EnableUFW(dryRun bool) error {
	if dryRun {
		fmt.Println("[DRY RUN] Would execute: ufw --force enable")
		return nil
	}

	cmd := exec.Command("ufw", "--force", "enable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable UFW: %s: %w", string(output), err)
	}

	return nil
}

// GetUFWStatus returns the current UFW status
func GetUFWStatus() (string, error) {
	cmd := exec.Command("ufw", "status", "verbose")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get UFW status: %w", err)
	}
	return string(output), nil
}

// ConfigureFirewallResult holds the result of firewall configuration
type ConfigureFirewallResult struct {
	RulesApplied []FirewallRule
	Enabled      bool
	Errors       []error
}

// ConfigureFirewall sets up UFW with all required rules
// IMPORTANT: This follows a strict order to prevent lockout:
// 1. Check if UFW is installed
// 2. Allow SSH FIRST (lockout prevention)
// 3. Add service rules
// 4. Enable UFW only after rules are verified
func ConfigureFirewall(rules []FirewallRule, enable bool, dryRun bool) ConfigureFirewallResult {
	result := ConfigureFirewallResult{
		RulesApplied: make([]FirewallRule, 0),
	}

	// Step 1: Check if UFW is installed
	if !IsUFWInstalled() {
		result.Errors = append(result.Errors, fmt.Errorf("UFW is not installed"))
		return result
	}

	// Step 2: CRITICAL - Allow SSH first
	if dryRun {
		fmt.Println("\n⚠️  LOCKOUT PREVENTION: Ensuring SSH is allowed FIRST")
	}
	if err := AllowSSH(dryRun); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("CRITICAL: SSH rule failed: %w", err))
		return result // Do NOT continue if SSH fails
	}

	// Step 3: Apply all rules
	for _, rule := range rules {
		if err := AllowPort(rule.Port, rule.Protocol, dryRun); err != nil {
			if rule.Required {
				result.Errors = append(result.Errors, err)
			}
			continue
		}
		result.RulesApplied = append(result.RulesApplied, rule)
		if !dryRun {
			fmt.Printf("✓ Allowed %s (port %d/%s)\n", rule.Service, rule.Port, rule.Protocol)
		}
	}

	// Step 4: Enable UFW only if all required rules succeeded
	if enable {
		hasRequiredError := false
		for _, rule := range rules {
			if rule.Required {
				found := false
				for _, applied := range result.RulesApplied {
					if applied.Port == rule.Port {
						found = true
						break
					}
				}
				if !found {
					hasRequiredError = true
					break
				}
			}
		}

		if hasRequiredError {
			result.Errors = append(result.Errors,
				fmt.Errorf("not enabling UFW: required rule(s) failed"))
			return result
		}

		if err := EnableUFW(dryRun); err != nil {
			result.Errors = append(result.Errors, err)
			return result
		}
		result.Enabled = true

		if !dryRun {
			fmt.Println("\n✅ UFW firewall enabled")
		}
	}

	return result
}

// HasCriticalErrors checks if any errors are critical
func (r *ConfigureFirewallResult) HasCriticalErrors() bool {
	for _, err := range r.Errors {
		if strings.Contains(err.Error(), "CRITICAL") {
			return true
		}
	}
	return false
}
