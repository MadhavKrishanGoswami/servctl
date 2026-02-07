package storage

import (
	"testing"
)

// =============================================================================
// Benchmark Tests - Measure performance of critical operations
// Run with: go test -bench=. -benchmem
// =============================================================================

// BenchmarkGenerateStrategies_SingleDisk benchmarks single disk strategy gen
func BenchmarkGenerateStrategies_SingleDisk(b *testing.B) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, IsAvailable: true},
	}
	info := SystemInfo{TotalRAM: 16 * 1024 * 1024 * 1024}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateStrategies(disks, info)
	}
}

// BenchmarkGenerateStrategies_MultiDisk benchmarks multi-disk strategy gen
func BenchmarkGenerateStrategies_MultiDisk(b *testing.B) {
	disks := []Disk{
		{Path: "/dev/sdb", Size: 4 * 1024 * 1024 * 1024 * 1024, Type: DiskTypeHDD, IsAvailable: true},
		{Path: "/dev/sdc", Size: 4 * 1024 * 1024 * 1024 * 1024, Type: DiskTypeHDD, IsAvailable: true},
		{Path: "/dev/nvme0n1", Size: 1 * 1024 * 1024 * 1024 * 1024, Type: DiskTypeNVMe, IsAvailable: true},
	}
	info := SystemInfo{TotalRAM: 32 * 1024 * 1024 * 1024}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateStrategies(disks, info)
	}
}

// BenchmarkScoreStrategies benchmarks strategy scoring
func BenchmarkScoreStrategies(b *testing.B) {
	strategies := []Strategy{
		{ID: StrategyMergerFS, Name: "MergerFS", Score: 65},
		{ID: StrategyMirror, Name: "Mirror", Score: 80},
		{ID: StrategyBackup, Name: "Backup", Score: 75},
		{ID: StrategyScratchVault, Name: "Scratch+Vault", Score: 70},
		{ID: StrategySpeedTiered, Name: "Speed-Tiered", Score: 85},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScoreStrategies(strategies)
	}
}

// BenchmarkFilterAvailableDisks_Bench benchmarks disk filtering
func BenchmarkFilterAvailableDisks_Bench(b *testing.B) {
	disks := make([]Disk, 20)
	for i := 0; i < 20; i++ {
		disks[i] = Disk{
			Path:        "/dev/sd" + string(rune('a'+i)),
			IsOSDisk:    i == 0,
			Removable:   i%5 == 0,
			IsAvailable: true,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterAvailableDisks(disks)
	}
}

// BenchmarkAreSizesSimilar benchmarks size comparison
func BenchmarkAreSizesSimilar(b *testing.B) {
	a := uint64(4 * 1024 * 1024 * 1024 * 1024)
	c := uint64(4*1024*1024*1024*1024 + 100*1024*1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AreSizesSimilar(a, c, 0.10)
	}
}

// BenchmarkGetDiskSpeedClass benchmarks speed classification
func BenchmarkGetDiskSpeedClass(b *testing.B) {
	disks := []Disk{
		{Type: DiskTypeNVMe},
		{Type: DiskTypeSSD},
		{Type: DiskTypeHDD},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range disks {
			GetDiskSpeedClass(d)
		}
	}
}

// BenchmarkIsHardwareRAID benchmarks RAID detection
func BenchmarkIsHardwareRAID(b *testing.B) {
	disk := Disk{Model: "PERC H730 Virtual Disk"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsHardwareRAID(disk)
	}
}

// BenchmarkDefaultStrategyConfig benchmarks config creation
func BenchmarkDefaultStrategyConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultStrategyConfig()
	}
}

// BenchmarkToConfigMap benchmarks config conversion
func BenchmarkToConfigMap(b *testing.B) {
	config := DefaultStrategyConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ToConfigMap()
	}
}

// BenchmarkRenderStrategyPreview benchmarks preview rendering
func BenchmarkRenderStrategyPreview(b *testing.B) {
	strategy := Strategy{
		ID:   StrategyMergerFS,
		Name: "Combined Pool (MergerFS)",
		Disks: []Disk{
			{Path: "/dev/sdb", SizeHuman: "4TB"},
			{Path: "/dev/sdc", SizeHuman: "4TB"},
			{Path: "/dev/sdd", SizeHuman: "2TB"},
		},
	}
	config := DefaultStrategyConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderStrategyPreview(strategy, config)
	}
}

// BenchmarkApplyStrategy_DryRun benchmarks strategy application
func BenchmarkApplyStrategy_DryRun(b *testing.B) {
	strategy := Strategy{
		ID:   StrategyPartition,
		Name: "Single Data Disk",
		Disks: []Disk{
			{Path: "/dev/sdb"},
		},
	}
	config := DefaultStrategyConfig().ToConfigMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyStrategy(strategy, config, true)
	}
}
