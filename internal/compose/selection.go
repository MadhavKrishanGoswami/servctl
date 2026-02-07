// Package compose handles Docker Compose generation and service configuration.
// This file implements interactive service configuration prompts.
package compose

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// PromptServiceConfig prompts user for service configuration
func PromptServiceConfig(reader *bufio.Reader, config *ServiceConfig) *ServiceConfig {
	fmt.Println("Service Configuration:")
	fmt.Println()

	// Host IP confirmation
	if config.HostIP != "" {
		fmt.Printf("  Host IP: %s (detected)\n", config.HostIP)
		fmt.Print("  Press Enter to keep, or enter new IP: ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if response != "" {
			if err := ValidateIP(response); err == nil {
				config.HostIP = response
			} else {
				fmt.Printf("  Invalid IP, keeping %s\n", config.HostIP)
			}
		}
	}
	fmt.Println()

	return config
}

// PromptPorts prompts user to customize service ports
func PromptPorts(reader *bufio.Reader, config *ServiceConfig) *ServiceConfig {
	fmt.Println("Service Ports (press Enter to keep defaults):")
	fmt.Println()

	ports := []struct {
		name    string
		current *int
		def     int
	}{
		{"Nextcloud", &config.NextcloudPort, 8080},
		{"Immich", &config.ImmichPort, 2283},
		{"Glances", &config.GlancesPort, 61208},
	}

	for _, p := range ports {
		fmt.Printf("  %s [%d]: ", p.name, *p.current)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if response != "" {
			if port, err := strconv.Atoi(response); err == nil && port > 0 && port < 65536 {
				*p.current = port
			}
		}
	}
	fmt.Println()

	return config
}

// RenderConfigPreview renders a preview of the service configuration
func RenderConfigPreview(config *ServiceConfig) string {
	var b strings.Builder

	b.WriteString("┌─────────────────────────────────────────┐\n")
	b.WriteString("│         Configuration Preview           │\n")
	b.WriteString("└─────────────────────────────────────────┘\n\n")

	b.WriteString(fmt.Sprintf("  Host IP:        %s\n", config.HostIP))
	b.WriteString(fmt.Sprintf("  Timezone:       %s\n", config.Timezone))
	b.WriteString(fmt.Sprintf("  Data Root:      %s\n", config.DataRoot))
	b.WriteString("\n")
	b.WriteString("  Service Ports:\n")
	b.WriteString(fmt.Sprintf("    • Nextcloud:  %d\n", config.NextcloudPort))
	b.WriteString(fmt.Sprintf("    • Immich:     %d\n", config.ImmichPort))
	b.WriteString(fmt.Sprintf("    • Glances:    %d\n", config.GlancesPort))
	b.WriteString("\n")

	return b.String()
}

// PromptConfigConfirmation prompts user to accept or customize the config
func PromptConfigConfirmation(reader *bufio.Reader, config *ServiceConfig) (*ServiceConfig, bool) {
	fmt.Print(RenderConfigPreview(config))
	fmt.Print("Press Enter to accept, 'c' to customize, or 's' to skip: ")

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "c":
		// Customize
		config = PromptServiceConfig(reader, config)
		config = PromptPorts(reader, config)
		return config, true
	case "s":
		return config, false
	default:
		return config, true
	}
}
