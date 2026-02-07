// Package directory handles creation of the servctl directory structure.
// This file implements interactive service selection for directory creation.
package directory

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
)

// cleanPath removes trailing slashes and cleans the path
func cleanPath(p string) string {
	return filepath.Clean(strings.TrimSuffix(p, "/"))
}

// ServiceSelection represents which services to set up
type ServiceSelection struct {
	Nextcloud bool
	Immich    bool
	Databases bool
	Glances   bool
}

// DefaultServiceSelection returns all services enabled
func DefaultServiceSelection() ServiceSelection {
	return ServiceSelection{
		Nextcloud: true,
		Immich:    true,
		Databases: true,
		Glances:   true,
	}
}

// PromptServiceSelection prompts user to select which services to configure
func PromptServiceSelection(reader *bufio.Reader) ServiceSelection {
	selection := DefaultServiceSelection()

	fmt.Println("Select services to configure (Enter to keep all, or type numbers to toggle):")
	fmt.Println()

	renderSelection := func() {
		checkbox := func(enabled bool) string {
			if enabled {
				return "[x]"
			}
			return "[ ]"
		}
		fmt.Printf("  1. %s Nextcloud   - File sync & office suite\n", checkbox(selection.Nextcloud))
		fmt.Printf("  2. %s Immich      - Photo & video library\n", checkbox(selection.Immich))
		fmt.Printf("  3. %s Databases   - PostgreSQL & Redis\n", checkbox(selection.Databases))
		fmt.Printf("  4. %s Glances     - System monitoring\n", checkbox(selection.Glances))
		fmt.Println()
	}

	renderSelection()
	fmt.Print("Toggle (e.g., '1 3' to toggle Nextcloud & Databases), or Enter to continue: ")

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
			selection.Nextcloud = !selection.Nextcloud
		case "2":
			selection.Immich = !selection.Immich
		case "3":
			selection.Databases = !selection.Databases
		case "4":
			selection.Glances = !selection.Glances
		}
	}

	fmt.Println()
	fmt.Println("Selected configuration:")
	renderSelection()

	return selection
}

// GetDirectoriesForServices returns directories only for selected services
func GetDirectoriesForServices(sel ServiceSelection, homeDir, dataRoot string) []DirectorySpec {
	var dirs []DirectorySpec

	// Sanitize input paths to prevent double slashes
	homeDir = cleanPath(homeDir)
	dataRoot = cleanPath(dataRoot)

	// Always include core infrastructure directories
	dirs = append(dirs, DirectorySpec{
		Path:        filepath.Join(homeDir, "infra"),
		Type:        DirTypeUserSpace,
		Service:     "core",
		Description: "Infrastructure root",
		Mode:        0755,
	})
	dirs = append(dirs, DirectorySpec{
		Path:        filepath.Join(homeDir, "infra", "compose"),
		Type:        DirTypeUserSpace,
		Service:     "core",
		Description: "Docker Compose files",
		Mode:        0755,
	})
	dirs = append(dirs, DirectorySpec{
		Path:        filepath.Join(homeDir, "infra", "scripts"),
		Type:        DirTypeUserSpace,
		Service:     "core",
		Description: "Maintenance scripts",
		Mode:        0755,
	})
	dirs = append(dirs, DirectorySpec{
		Path:        filepath.Join(homeDir, "infra", "logs"),
		Type:        DirTypeUserSpace,
		Service:     "core",
		Description: "Log files",
		Mode:        0755,
	})

	// Data root
	dirs = append(dirs, DirectorySpec{
		Path:        dataRoot,
		Type:        DirTypeDataSpace,
		Service:     "core",
		Description: "Data storage root",
		Mode:        0755,
	})

	// Nextcloud directories
	if sel.Nextcloud {
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "nextcloud"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud root",
			Mode:        0755,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "nextcloud", "data"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud user data",
			Mode:        0770,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "nextcloud", "config"),
			Type:        DirTypeDataSpace,
			Service:     "nextcloud",
			Description: "Nextcloud configuration",
			Mode:        0770,
		})
	}

	// Immich directories
	if sel.Immich {
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "immich"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Immich root",
			Mode:        0755,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "immich", "upload"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Photo uploads",
			Mode:        0770,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "immich", "library"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Photo library",
			Mode:        0770,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "immich", "thumbs"),
			Type:        DirTypeDataSpace,
			Service:     "immich",
			Description: "Thumbnails cache",
			Mode:        0770,
		})
	}

	// Database directories
	if sel.Databases {
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "databases"),
			Type:        DirTypeDataSpace,
			Service:     "databases",
			Description: "Database storage",
			Mode:        0700,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "databases", "postgres"),
			Type:        DirTypeDataSpace,
			Service:     "databases",
			Description: "PostgreSQL data",
			Mode:        0700,
		})
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(dataRoot, "databases", "redis"),
			Type:        DirTypeDataSpace,
			Service:     "databases",
			Description: "Redis data",
			Mode:        0700,
		})
	}

	// Glances (monitoring) - no persistent data needed, just config
	if sel.Glances {
		dirs = append(dirs, DirectorySpec{
			Path:        filepath.Join(homeDir, "infra", "glances"),
			Type:        DirTypeUserSpace,
			Service:     "glances",
			Description: "Glances config",
			Mode:        0755,
		})
	}

	return dirs
}

// PromptCustomDataRoot prompts user to customize the data root path
func PromptCustomDataRoot(reader *bufio.Reader, defaultPath string) string {
	fmt.Printf("Data root path [%s]: ", defaultPath)

	response, err := reader.ReadString('\n')
	if err != nil {
		return defaultPath
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return defaultPath
	}

	return response
}

// CountSelectedServices returns the number of selected services
func (s ServiceSelection) CountSelectedServices() int {
	count := 0
	if s.Nextcloud {
		count++
	}
	if s.Immich {
		count++
	}
	if s.Databases {
		count++
	}
	if s.Glances {
		count++
	}
	return count
}

// SelectedNames returns names of selected services
func (s ServiceSelection) SelectedNames() []string {
	var names []string
	if s.Nextcloud {
		names = append(names, "Nextcloud")
	}
	if s.Immich {
		names = append(names, "Immich")
	}
	if s.Databases {
		names = append(names, "Databases")
	}
	if s.Glances {
		names = append(names, "Glances")
	}
	return names
}
