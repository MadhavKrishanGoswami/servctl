# Contributing to servctl

Thank you for your interest in contributing! This guide will help you get started.

---

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Code Style](#code-style)
- [Testing](#testing)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

---

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to build something useful.

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Docker (for integration tests)
- Make (optional, for convenience commands)

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/madhav/servctl.git
cd servctl

# Install dependencies
go mod download

# Build locally
go build -o bin/servctl-local ./cmd/servctl

# Run tests
go test ./internal/... -short

# Verify everything works
./bin/servctl-local -version
```

---

## Development Setup

### Environment

```bash
# Set up pre-commit checks (optional)
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build Linux binary (for deployment) |
| `make build-local` | Build for current OS (development) |
| `make test-short` | Run unit tests |
| `make test-coverage` | Generate coverage report |
| `make docker-test` | Run tests in Ubuntu container |
| `make docker-shell` | Interactive Ubuntu shell |

### IDE Setup

**VS Code** (recommended):
```json
// .vscode/settings.json
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "gofmt",
  "go.testFlags": ["-v", "-short"]
}
```

---

## Project Structure

```
servctl/
├── cmd/servctl/           # CLI entry point
│   └── main.go            # Flag parsing and command routing
│
├── internal/              # Private application packages
│   ├── compose/           # Docker Compose generation
│   │   ├── compose.go     # File generation logic
│   │   └── selection.go   # User prompts and config
│   │
│   ├── directory/         # Directory structure
│   │   ├── directory.go   # Creation and permissions
│   │   └── selection.go   # Service selection prompts
│   │
│   ├── maintenance/       # Maintenance scripts
│   │   ├── maintenance.go # Script generation
│   │   └── selection.go   # Script selection prompts
│   │
│   ├── preflight/         # System checks
│   │   └── preflight.go   # Requirement validation
│   │
│   ├── report/            # Mission report
│   │   └── report.go      # Summary rendering
│   │
│   ├── storage/           # Disk management
│   │   ├── discovery.go   # Disk detection (lsblk)
│   │   ├── recommendation.go # Strategy generation
│   │   ├── operations.go  # Format, mount, RAID
│   │   ├── format.go      # Filesystem operations
│   │   └── power.go       # HDD spindown config
│   │
│   ├── tui/               # Terminal UI
│   │   ├── styles.go      # Lipgloss styles
│   │   └── *.go           # Component renderers
│   │
│   └── utils/             # Shared utilities
│       ├── logger.go      # File logging
│       └── helpers.go     # Common functions
│
├── templates/             # Template files
│   ├── docker-compose.yml.tmpl
│   └── env.tmpl
│
├── scripts/               # Development scripts
├── build/                 # Docker configs
└── docs/                  # Documentation
```

### Package Guidelines

| Package | Responsibility |
|---------|---------------|
| `compose` | Generate Docker Compose and .env files |
| `directory` | Create directories with proper permissions |
| `maintenance` | Generate bash scripts for cron jobs |
| `preflight` | Validate system requirements |
| `report` | Render final setup summary |
| `storage` | Discover, format, and configure disks |
| `tui` | Render beautiful terminal output |
| `utils` | Logging and shared helpers |

---

## Code Style

### Go Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `go fmt` before committing
- Use `go vet` to catch issues
- Maximum line length: 100 characters (soft limit)

### Naming

```go
// Exported functions: PascalCase with descriptive names
func DiscoverDisks() ([]Disk, error)

// Unexported helpers: camelCase
func parseBlockDevice(line string) Disk

// Structs: PascalCase, descriptive
type StorageStrategy struct {
    ID          StrategyID
    Name        string
    Description string
}

// Constants: PascalCase or grouped with type
const (
    StrategySimple StrategyID = iota
    StrategyMergerFS
    StrategyMirror
)
```

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to discover disks: %w", err)
}

// Return early on errors
result, err := doSomething()
if err != nil {
    return nil, err
}
// continue with result...
```

### Comments

```go
// Package storage provides disk discovery, formatting, and storage strategy
// management for servctl.
package storage

// DiscoverDisks scans the system for available block devices using lsblk.
// It returns a slice of Disk structs with detailed information about each device.
// On non-Linux systems, it returns simulated data for testing purposes.
func DiscoverDisks() ([]Disk, error) {
    // ...
}
```

---

## Testing

### Test Structure

```
internal/storage/
├── discovery.go
├── discovery_test.go      # Unit tests for discovery.go
├── operations.go
├── operations_test.go
├── error_test.go          # Error path tests
├── fuzz_test.go           # Fuzz tests
├── benchmark_test.go      # Performance tests
└── integration_test.go    # Linux-only integration tests
```

### Running Tests

```bash
# Quick unit tests (works on macOS)
go test ./internal/... -short -count=1

# Verbose output
go test ./internal/... -v -short

# Single package
go test ./internal/storage/... -v

# With coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run in Docker (full Linux environment)
make docker-test
```

### Test Patterns

```go
// Table-driven tests
func TestParseSize(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    uint64
        wantErr bool
    }{
        {"gigabytes", "100G", 107374182400, false},
        {"terabytes", "2T", 2199023255552, false},
        {"invalid", "abc", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseSize(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseSize() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("parseSize() = %v, want %v", got, tt.want)
            }
        })
    }
}

// Error path tests
func TestDiscoverDisks_InvalidJSON(t *testing.T) {
    // Test graceful handling of malformed lsblk output
}
```

### Integration Tests

Integration tests require Linux and the `integration` build tag:

```bash
# In Docker
go test ./internal/storage/... -tags=integration -v
```

---

## Commit Guidelines

### Commit Message Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no code change |
| `refactor` | Code restructuring |
| `test` | Adding or fixing tests |
| `chore` | Build process, deps, etc. |

### Examples

```bash
feat(storage): add ZFS mirror strategy
fix(preflight): handle missing docker command
docs(readme): add storage strategy table
test(compose): add password validation edge cases
refactor(tui): extract common styles to shared file
```

---

## Pull Request Process

### Before Submitting

1. **Run all checks:**
   ```bash
   go fmt ./...
   go vet ./...
   go test ./internal/... -short
   ```

2. **Update documentation** if adding features

3. **Add tests** for new functionality

4. **Keep changes focused** — one feature/fix per PR

### PR Template

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Tested manually on Ubuntu

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-reviewed the code
- [ ] Added/updated documentation
- [ ] No new warnings
```

### Review Process

1. Automated CI runs tests
2. Maintainer reviews code
3. Address feedback if any
4. Squash and merge

---

## Release Process

Releases are managed via GitHub Actions and tags:

```bash
# Create release
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes

---

## Questions?

- Open an [Issue](https://github.com/madhav/servctl/issues)
- Start a [Discussion](https://github.com/madhav/servctl/discussions)

Thank you for contributing! 🙏
