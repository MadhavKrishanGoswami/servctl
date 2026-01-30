# Testing Guide for servctl

## Test Organization

servctl follows Go's standard testing conventions where test files (`*_test.go`) are 
co-located with the source files they test. This is the recommended Go practice.

```
internal/
├── preflight/
│   ├── preflight.go       # Implementation
│   └── preflight_test.go  # Unit tests
├── storage/
│   ├── discovery.go
│   ├── discovery_test.go
│   ├── classification.go
│   ├── classification_test.go
│   └── ...
└── tui/
    └── ...
```

## Test Categories

### 1. Unit Tests (Run Anywhere)
These tests work on **any platform** (Mac, Windows, Linux) and don't require special 
system access. They test pure logic, data structures, and algorithms.

**To run unit tests only:**
```bash
go test ./... -short
```

**What's tested:**
- String conversions (`DiskType.String()`, `Status.String()`)
- Data structure validation
- Classification logic
- Formatting functions (`formatBytes()`, `FormatRecommendationSummary()`)
- Dry-run operations (no actual disk changes)

### 2. Integration Tests (Require Linux VM)
These tests require a **Linux environment** (Ubuntu 22.04+) with actual system access.
They test real system interactions.

**To run integration tests:**
```bash
# On Ubuntu VM only
go test ./... -tags=integration
```

**What's tested:**
- Actual disk discovery (`lsblk` parsing)
- Real fstab modifications
- Docker status checks
- Network connectivity
- Privilege verification

### 3. Destructive Tests (Require Test VM with Sacrificial Disks)
These tests actually format disks and modify system configs. 
**NEVER run on production systems!**

**To run destructive tests:**
```bash
# On disposable test VM with extra disks only
go test ./... -tags=destructive
```

## Running Tests by Platform

### On macOS (Development Machine)
```bash
# Safe to run - unit tests only
go test ./internal/storage/... -v

# Will skip Linux-specific tests automatically
go test ./internal/preflight/... -v -short
```

### On Ubuntu VM (Full Testing)
```bash
# Run all tests including integration
go test ./... -v

# Include integration tests
go test ./... -v -tags=integration
```

### On Disposable Test VM (Full + Destructive)
```bash
# Run everything including destructive tests
go test ./... -v -tags=integration,destructive
```

## Test File Naming

| Pattern | Description |
|---------|-------------|
| `*_test.go` | Standard unit tests (run anywhere) |
| `*_integration_test.go` | Integration tests (Linux only) |
| `*_linux_test.go` | Linux-specific tests (auto-skipped on other OS) |

## Quick Commands

```bash
# Run all safe tests
make test

# Run tests with coverage
make test-coverage

# Run only unit tests (fast)
go test ./... -short -count=1

# Run specific package tests
go test ./internal/storage/... -v

# Run specific test function
go test ./internal/storage/... -run TestFormatBytes -v
```

## CI/CD Pipeline

In GitHub Actions, we run tests in stages:
1. **Unit tests** - On all platforms (ubuntu, macos, windows)
2. **Integration tests** - On Ubuntu runner only
3. **Destructive tests** - On self-hosted runner with test disks
