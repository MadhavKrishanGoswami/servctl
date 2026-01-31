package preflight

import (
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusPass, "PASS"},
		{StatusWarn, "WARN"},
		{StatusFail, "FAIL"},
		{StatusSkip, "SKIP"},
		{Status(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("Status.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCheckResultStructure(t *testing.T) {
	result := CheckResult{
		Name:    "Test Check",
		Status:  StatusPass,
		Message: "Test passed",
		Details: []string{"Detail 1", "Detail 2"},
	}

	if result.Name != "Test Check" {
		t.Errorf("CheckResult.Name = %v, want %v", result.Name, "Test Check")
	}
	if result.Status != StatusPass {
		t.Errorf("CheckResult.Status = %v, want %v", result.Status, StatusPass)
	}
	if len(result.Details) != 2 {
		t.Errorf("len(CheckResult.Details) = %v, want %v", len(result.Details), 2)
	}
}

func TestGetRequiredDependencies(t *testing.T) {
	deps := GetRequiredDependencies()

	if len(deps) == 0 {
		t.Error("GetRequiredDependencies() returned empty list")
	}

	// Check that we have the critical dependencies
	criticalDeps := []string{"curl", "docker", "lsblk", "mkfs.ext4"}
	foundDeps := make(map[string]bool)

	for _, dep := range deps {
		foundDeps[dep.Binary] = true
	}

	for _, critical := range criticalDeps {
		if !foundDeps[critical] {
			t.Errorf("Missing critical dependency: %s", critical)
		}
	}
}

func TestDependencyCriticality(t *testing.T) {
	deps := GetRequiredDependencies()

	blockerCount := 0
	highCount := 0
	recommendedCount := 0

	for _, dep := range deps {
		switch dep.Criticality {
		case "blocker":
			blockerCount++
		case "high":
			highCount++
		case "recommended":
			recommendedCount++
		}
	}

	if blockerCount == 0 {
		t.Error("Expected at least one blocker dependency")
	}

	t.Logf("Dependencies - Blockers: %d, High: %d, Recommended: %d",
		blockerCount, highCount, recommendedCount)
}

func TestHasBlockers(t *testing.T) {
	tests := []struct {
		name     string
		results  []CheckResult
		expected bool
	}{
		{
			name: "No blockers",
			results: []CheckResult{
				{Status: StatusPass},
				{Status: StatusWarn},
				{Status: StatusPass},
			},
			expected: false,
		},
		{
			name: "Has blocker",
			results: []CheckResult{
				{Status: StatusPass},
				{Status: StatusFail},
				{Status: StatusPass},
			},
			expected: true,
		},
		{
			name: "All pass",
			results: []CheckResult{
				{Status: StatusPass},
				{Status: StatusPass},
			},
			expected: false,
		},
		{
			name:     "Empty results",
			results:  []CheckResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasBlockers(tt.results); got != tt.expected {
				t.Errorf("HasBlockers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCountByStatus(t *testing.T) {
	results := []CheckResult{
		{Status: StatusPass},
		{Status: StatusPass},
		{Status: StatusWarn},
		{Status: StatusFail},
		{Status: StatusPass},
		{Status: StatusSkip},
	}

	counts := CountByStatus(results)

	if counts[StatusPass] != 3 {
		t.Errorf("CountByStatus[StatusPass] = %v, want %v", counts[StatusPass], 3)
	}
	if counts[StatusWarn] != 1 {
		t.Errorf("CountByStatus[StatusWarn] = %v, want %v", counts[StatusWarn], 1)
	}
	if counts[StatusFail] != 1 {
		t.Errorf("CountByStatus[StatusFail] = %v, want %v", counts[StatusFail], 1)
	}
	if counts[StatusSkip] != 1 {
		t.Errorf("CountByStatus[StatusSkip] = %v, want %v", counts[StatusSkip], 1)
	}
}

func TestOSInfoStructure(t *testing.T) {
	info := OSInfo{
		ID:              "ubuntu",
		VersionID:       "22.04",
		Name:            "Ubuntu",
		PrettyName:      "Ubuntu 22.04.1 LTS",
		VersionCodename: "jammy",
	}

	if info.ID != "ubuntu" {
		t.Errorf("OSInfo.ID = %v, want %v", info.ID, "ubuntu")
	}
	if info.VersionID != "22.04" {
		t.Errorf("OSInfo.VersionID = %v, want %v", info.VersionID, "22.04")
	}
}

// TestCheckOS tests that CheckOS returns appropriate result
// NOTE: This test requires /etc/os-release which only exists on Linux
func TestCheckOS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OS check in short mode (requires Linux)")
	}

	result := CheckOS()

	// Result should have a name
	if result.Name == "" {
		t.Error("CheckOS() returned empty name")
	}

	// Result should have a valid status
	validStatuses := map[Status]bool{
		StatusPass: true,
		StatusWarn: true,
		StatusFail: true,
		StatusSkip: true,
	}
	if !validStatuses[result.Status] {
		t.Errorf("CheckOS() returned invalid status: %v", result.Status)
	}

	t.Logf("CheckOS result: %s - %s", result.Status.String(), result.Message)
}

func TestCheckPrivileges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping privilege check in short mode (requires sudo)")
	}

	result := CheckPrivileges()

	if result.Name == "" {
		t.Error("CheckPrivileges() returned empty name")
	}

	// Should have some details
	t.Logf("CheckPrivileges result: %s - %s", result.Status.String(), result.Message)
}

func TestCheckConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connectivity check in short mode (requires network)")
	}

	result := CheckConnectivity()

	if result.Name == "" {
		t.Error("CheckConnectivity() returned empty name")
	}

	// Should have details about each test
	if len(result.Details) == 0 {
		t.Error("CheckConnectivity() returned no details")
	}

	t.Logf("CheckConnectivity result: %s - %s (details: %d)",
		result.Status.String(), result.Message, len(result.Details))
}

func TestCheckHardware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping hardware check in short mode (requires Linux)")
	}

	result := CheckHardware()

	if result.Name == "" {
		t.Error("CheckHardware() returned empty name")
	}

	t.Logf("CheckHardware result: %s - %s", result.Status.String(), result.Message)
}

func TestCheckDependency(t *testing.T) {
	// Test with a dependency that should exist on most systems
	dep := Dependency{
		Name:        "Go",
		Binary:      "go",
		Package:     "golang",
		Criticality: "blocker",
		InstallCmd:  "apt install golang",
	}

	result := CheckDependency(dep)

	if result.Name == "" {
		t.Error("CheckDependency() returned empty name")
	}

	// Go should be installed since we're running tests
	if result.Status != StatusPass {
		t.Logf("Warning: 'go' binary not found. Status: %s", result.Status.String())
	}
}

func TestCheckDependencyMissing(t *testing.T) {
	// Test with a dependency that definitely shouldn't exist
	dep := Dependency{
		Name:        "NonExistent",
		Binary:      "nonexistent_binary_12345",
		Package:     "nonexistent",
		Criticality: "blocker",
		InstallCmd:  "apt install nonexistent",
	}

	result := CheckDependency(dep)

	if result.Status != StatusFail {
		t.Errorf("CheckDependency() for missing binary = %v, want %v",
			result.Status, StatusFail)
	}
}

func TestCheckAllDependencies(t *testing.T) {
	results := CheckAllDependencies()

	if len(results) == 0 {
		t.Error("CheckAllDependencies() returned empty results")
	}

	// Count results by status
	counts := CountByStatus(results)
	t.Logf("CheckAllDependencies: Pass=%d, Warn=%d, Fail=%d, Skip=%d",
		counts[StatusPass], counts[StatusWarn], counts[StatusFail], counts[StatusSkip])
}

// Benchmark tests
func BenchmarkCheckConnectivity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CheckConnectivity()
	}
}

func BenchmarkGetRequiredDependencies(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetRequiredDependencies()
	}
}

// Tests for auto-installation functions

func TestGetMissingDependencies(t *testing.T) {
	missing := GetMissingDependencies()

	// This is environment-dependent, so we just check the function runs
	t.Logf("Found %d missing dependencies", len(missing))

	// Verify returned deps have required fields
	for _, dep := range missing {
		if dep.Name == "" {
			t.Error("Missing dependency has empty Name")
		}
		if dep.Binary == "" {
			t.Error("Missing dependency has empty Binary")
		}
		if dep.InstallCmd == "" {
			t.Error("Missing dependency has empty InstallCmd")
		}
	}
}

func TestInstallResultStruct(t *testing.T) {
	dep := Dependency{
		Name:        "test",
		Binary:      "test",
		Package:     "test",
		Criticality: "recommended",
	}

	result := InstallResult{
		Dependency: dep,
		Success:    true,
		Error:      nil,
		Duration:   0,
	}

	if result.Dependency.Name != "test" {
		t.Errorf("InstallResult.Dependency.Name = %v, want test", result.Dependency.Name)
	}
	if !result.Success {
		t.Error("InstallResult.Success should be true")
	}
}

func TestInstallAllMissingDependenciesDryRun(t *testing.T) {
	// Dry run should not actually install anything
	results := InstallAllMissingDependencies(true)

	// All results should be successful in dry run
	for _, r := range results {
		if !r.Success {
			t.Errorf("Dry run should always succeed, failed for: %s", r.Dependency.Name)
		}
	}
}

func TestRunPreflightWithAutoFixDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	results, installResults, err := RunPreflightWithAutoFix(true)

	if err != nil {
		t.Errorf("RunPreflightWithAutoFix returned error: %v", err)
	}

	if len(results) == 0 {
		t.Error("RunPreflightWithAutoFix returned no check results")
	}

	t.Logf("Check results: %d, Install results: %d", len(results), len(installResults))
}

func TestSystemSetupDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Dry run should not make any changes
	err := SystemSetup(true)

	// This may fail on non-Ubuntu systems, which is expected
	if err != nil {
		t.Logf("SystemSetup dry run returned: %v (expected on non-Ubuntu)", err)
	}
}
