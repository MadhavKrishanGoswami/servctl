// Package compose handles Docker Compose generation and service configuration.
// It generates docker-compose.yml and .env files for all services.
package compose

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ServiceConfig holds all configuration for servctl services
type ServiceConfig struct {
	// System settings
	Timezone string // TZ (e.g., "Asia/Kolkata")
	PUID     int    // Process User ID
	PGID     int    // Process Group ID
	HostIP   string // Static IP address of the host

	// Paths (opinionated, not user-configurable)
	DataRoot   string // /mnt/data
	InfraRoot  string // ~/infra
	UploadPath string // /mnt/data/gallery (Immich uploads)

	// Immich settings
	ImmichDBPassword string // Postgres password for Immich

	// Nextcloud settings
	NextcloudAdminUser      string // Admin username
	NextcloudAdminPass      string // Admin password
	NextcloudDBPassword     string // MariaDB password for Nextcloud
	NextcloudTrustedDomains string // Comma-separated trusted domains

	// Notification webhooks
	DiscordWebhookURL string // Discord webhook for notifications
	TelegramBotToken  string // Telegram bot token
	TelegramChatID    string // Telegram chat ID

	// Service ports (with sensible defaults)
	ImmichPort    int // Default: 2283
	NextcloudPort int // Default: 8080
	GlancesPort   int // Default: 61208
}

// DefaultConfig returns a ServiceConfig with sensible defaults
func DefaultConfig() *ServiceConfig {
	return &ServiceConfig{
		Timezone:           detectTimezone(),
		PUID:               1000,
		PGID:               1000,
		DataRoot:           "/mnt/data",
		InfraRoot:          "",
		UploadPath:         "/mnt/data/gallery",
		ImmichPort:         2283,
		NextcloudPort:      8080,
		GlancesPort:        61208,
		NextcloudAdminUser: "admin",
	}
}

// detectTimezone attempts to detect the system timezone
func detectTimezone() string {
	// Try reading from /etc/timezone
	if data, err := os.ReadFile("/etc/timezone"); err == nil {
		tz := strings.TrimSpace(string(data))
		if tz != "" {
			return tz
		}
	}

	// Try TZ environment variable
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}

	// Default to UTC
	return "UTC"
}

// GeneratePassword creates a secure random password
func GeneratePassword(length int) string {
	if length < 16 {
		length = 16
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based if crypto fails
		return fmt.Sprintf("servctl_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)[:length]
}

// GenerateDBPassword generates a database-safe password (alphanumeric only)
func GenerateDBPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("db%d", time.Now().UnixNano())
	}
	for i := range bytes {
		bytes[i] = chars[int(bytes[i])%len(chars)]
	}
	return string(bytes)
}

// DetectHostIP finds the primary IP address of the host
func DetectHostIP() (string, error) {
	// Method 1: Try to connect to external address to determine local IP
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 3*time.Second)
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String(), nil
	}

	// Method 2: Get all interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip IPv6 and loopback
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no suitable IP address found")
}

// ValidateIP checks if an IP address is valid and in private range
func ValidateIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}

	ipv4 := parsed.To4()
	if ipv4 == nil {
		return fmt.Errorf("IPv6 not supported, please use IPv4: %s", ip)
	}

	// Check if it's a private IP
	privateRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	}

	isPrivate := false
	for _, r := range privateRanges {
		if bytesInRange(ipv4, r.start.To4(), r.end.To4()) {
			isPrivate = true
			break
		}
	}

	if !isPrivate {
		return fmt.Errorf("IP %s is not in private range (10.x.x.x, 172.16-31.x.x, 192.168.x.x)", ip)
	}

	return nil
}

// bytesInRange checks if ip is between start and end
func bytesInRange(ip, start, end net.IP) bool {
	for i := 0; i < 4; i++ {
		if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}

// ValidateWebhookURL validates a Discord or Telegram webhook URL
func ValidateWebhookURL(url string) error {
	if url == "" {
		return nil // Empty is valid (optional)
	}

	// Discord webhook pattern
	discordPattern := regexp.MustCompile(`^https://discord\.com/api/webhooks/\d+/[\w-]+$`)
	// Slack-style pattern
	slackPattern := regexp.MustCompile(`^https://hooks\.slack\.com/services/[\w/]+$`)

	if discordPattern.MatchString(url) || slackPattern.MatchString(url) {
		return nil
	}

	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must use HTTPS")
	}

	return nil
}

// ValidatePassword checks password requirements
func ValidatePassword(password string, minLength int) error {
	if minLength == 0 {
		minLength = 8
	}
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters", minLength)
	}
	return nil
}

// IsIPInUse checks if an IP is already in use on the network
func IsIPInUse(ip string, timeout time.Duration) bool {
	// Quick ping check
	cmd := exec.Command("ping", "-c", "1", "-W", "1", ip)
	if err := cmd.Run(); err == nil {
		// Ping succeeded, IP might be in use
		return true
	}
	return false
}

// GetNetworkInfo returns information about network configuration
type NetworkInfo struct {
	CurrentIP  string
	Gateway    string
	Interface  string
	IsDHCP     bool
	DNSServers []string
}

// DetectNetworkInfo gathers current network configuration
func DetectNetworkInfo() (*NetworkInfo, error) {
	info := &NetworkInfo{
		IsDHCP: true, // Assume DHCP by default
	}

	// Get current IP
	ip, err := DetectHostIP()
	if err != nil {
		return nil, err
	}
	info.CurrentIP = ip

	// Try to get gateway
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err == nil {
		parts := strings.Fields(string(output))
		for i, part := range parts {
			if part == "via" && i+1 < len(parts) {
				info.Gateway = parts[i+1]
			}
			if part == "dev" && i+1 < len(parts) {
				info.Interface = parts[i+1]
			}
		}
	}

	return info, nil
}

// GetTimezoneOptions returns common timezone options
func GetTimezoneOptions() []string {
	return []string{
		"UTC",
		"America/New_York",
		"America/Los_Angeles",
		"America/Chicago",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Kolkata",
		"Asia/Dubai",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
}

// Validate performs full validation on a ServiceConfig
func (c *ServiceConfig) Validate() []error {
	var errors []error

	// Timezone
	if c.Timezone == "" {
		errors = append(errors, fmt.Errorf("timezone is required"))
	}

	// Host IP
	if c.HostIP != "" {
		if err := ValidateIP(c.HostIP); err != nil {
			errors = append(errors, err)
		}
	}

	// Database passwords
	if len(c.ImmichDBPassword) < 8 {
		errors = append(errors, fmt.Errorf("Immich database password must be at least 8 characters"))
	}
	if len(c.NextcloudDBPassword) < 8 {
		errors = append(errors, fmt.Errorf("Nextcloud database password must be at least 8 characters"))
	}

	// Nextcloud admin
	if c.NextcloudAdminUser == "" {
		errors = append(errors, fmt.Errorf("Nextcloud admin username is required"))
	}
	if len(c.NextcloudAdminPass) < 8 {
		errors = append(errors, fmt.Errorf("Nextcloud admin password must be at least 8 characters"))
	}

	// Webhook URLs
	if err := ValidateWebhookURL(c.DiscordWebhookURL); err != nil {
		errors = append(errors, fmt.Errorf("discord webhook: %w", err))
	}

	return errors
}

// AutoFillDefaults fills in any missing values with sensible defaults
func (c *ServiceConfig) AutoFillDefaults() {
	if c.Timezone == "" {
		c.Timezone = detectTimezone()
	}
	if c.PUID == 0 {
		c.PUID = 1000
	}
	if c.PGID == 0 {
		c.PGID = 1000
	}
	if c.DataRoot == "" {
		c.DataRoot = "/mnt/data"
	}
	if c.UploadPath == "" {
		c.UploadPath = c.DataRoot + "/gallery"
	}
	if c.ImmichDBPassword == "" {
		c.ImmichDBPassword = GenerateDBPassword()
	}
	if c.NextcloudDBPassword == "" {
		c.NextcloudDBPassword = GenerateDBPassword()
	}
	if c.NextcloudAdminUser == "" {
		c.NextcloudAdminUser = "admin"
	}
	if c.ImmichPort == 0 {
		c.ImmichPort = 2283
	}
	if c.NextcloudPort == 0 {
		c.NextcloudPort = 8080
	}
	if c.GlancesPort == 0 {
		c.GlancesPort = 61208
	}
}
