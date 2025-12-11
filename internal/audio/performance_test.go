package audio

import (
	"runtime"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TestBGMPlayer_MemoryUsage tests memory usage
func TestBGMPlayer_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cfg := config.AudioConfig{
		BGMEnabled: true,
		BGMVolume:  0.7,
	}

	// Force GC before measurement
	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Create BGM player
	player := NewBGMPlayer(nil, cfg, t.TempDir())

	// Force GC after creation
	runtime.GC()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Use int64 to handle potential negative values from GC
	allocDiff := int64(memAfter.Alloc) - int64(memBefore.Alloc)
	t.Logf("Memory allocated: %d bytes (%.2f KB)", allocDiff, float64(allocDiff)/1024)

	// BGM player should use < 10MB memory (基礎占用)
	// Allow negative values (GC ran between measurements)
	maxMemory := int64(10 * 1024 * 1024) // 10MB
	if allocDiff > maxMemory {
		t.Errorf("BGMPlayer uses too much memory: %d bytes (max %d bytes)", allocDiff, maxMemory)
	}

	_ = player
}

// TestBGMPlayer_FadePerformance tests fade performance
func TestBGMPlayer_FadePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	player := NewBGMPlayer(nil, config.AudioConfig{BGMVolume: 0.8}, t.TempDir())

	// Test fade out performance
	start := time.Now()
	if err := player.FadeOut(1 * time.Second); err != nil {
		t.Fatalf("FadeOut failed: %v", err)
	}
	fadeOutDuration := time.Since(start)

	t.Logf("FadeOut duration: %v", fadeOutDuration)

	// Fade should complete in approximately 1 second (allow 10% overhead)
	maxDuration := 1100 * time.Millisecond
	if fadeOutDuration > maxDuration {
		t.Errorf("FadeOut took too long: %v (max %v)", fadeOutDuration, maxDuration)
	}

	// Test fade in performance
	player.SetTargetVolume(0.7)
	start = time.Now()
	if err := player.FadeIn(1 * time.Second); err != nil {
		t.Fatalf("FadeIn failed: %v", err)
	}
	fadeInDuration := time.Since(start)

	t.Logf("FadeIn duration: %v", fadeInDuration)

	if fadeInDuration > maxDuration {
		t.Errorf("FadeIn took too long: %v (max %v)", fadeInDuration, maxDuration)
	}
}

// TestBGMPlayer_VolumeChangeLatency tests volume change latency
func TestBGMPlayer_VolumeChangeLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	player := NewBGMPlayer(nil, config.AudioConfig{BGMVolume: 0.5}, t.TempDir())

	// Measure volume change latency
	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		player.SetVolume(float64(i%100) / 100.0)
	}

	totalDuration := time.Since(start)
	avgLatency := totalDuration / time.Duration(iterations)

	t.Logf("Volume change avg latency: %v (%d iterations)", avgLatency, iterations)

	// Volume change should be instant (< 1ms on average)
	maxAvgLatency := 1 * time.Millisecond
	if avgLatency > maxAvgLatency {
		t.Errorf("Volume change latency too high: %v (max %v)", avgLatency, maxAvgLatency)
	}
}

// TestAudioManager_InitializationPerformance tests initialization performance
func TestAudioManager_InitializationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cfg := config.DefaultConfig()
	manager := NewAudioManager(cfg.Audio)

	// Test async initialization doesn't block
	start := time.Now()
	manager.InitializeAsync()
	asyncDuration := time.Since(start)

	t.Logf("InitializeAsync duration: %v", asyncDuration)

	// Async init should return immediately (< 10ms)
	maxAsyncDuration := 10 * time.Millisecond
	if asyncDuration > maxAsyncDuration {
		t.Errorf("InitializeAsync blocked too long: %v (max %v)", asyncDuration, maxAsyncDuration)
	}

	// Wait a bit for actual initialization to complete
	time.Sleep(200 * time.Millisecond)
}

// TestBGMPlayer_ConcurrentAccess tests thread safety under concurrent access
func TestBGMPlayer_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	player := NewBGMPlayer(nil, config.AudioConfig{BGMVolume: 0.7}, t.TempDir())

	// Launch multiple goroutines accessing player concurrently
	done := make(chan bool)
	iterations := 100

	// Concurrent volume changes
	go func() {
		for i := 0; i < iterations; i++ {
			player.SetVolume(float64(i%100) / 100.0)
		}
		done <- true
	}()

	// Concurrent enable/disable
	go func() {
		for i := 0; i < iterations; i++ {
			if i%2 == 0 {
				player.Enable()
			} else {
				player.Disable()
			}
		}
		done <- true
	}()

	// Concurrent loop mode changes
	go func() {
		for i := 0; i < iterations; i++ {
			player.SetLoop(i%2 == 0)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < iterations; i++ {
			_ = player.Volume()
			_ = player.IsEnabled()
			_ = player.IsPlaying()
			_ = player.CurrentBGM()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	t.Log("Concurrent access test completed successfully")
}

// BenchmarkBGMPlayer_SetVolume benchmarks volume setting
func BenchmarkBGMPlayer_SetVolume(b *testing.B) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMVolume: 0.5}, b.TempDir())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.SetVolume(float64(i%100) / 100.0)
	}
}

// BenchmarkBGMPlayer_VolumeRead benchmarks volume reading
func BenchmarkBGMPlayer_VolumeRead(b *testing.B) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMVolume: 0.5}, b.TempDir())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = player.Volume()
	}
}

// BenchmarkGetBGMFilename benchmarks BGM filename lookup
func BenchmarkGetBGMFilename(b *testing.B) {
	scenes := []BGMScene{
		BGMSceneExploration,
		BGMSceneChase,
		BGMSceneSafe,
		BGMSceneHorror,
		BGMSceneMystery,
		BGMSceneDeath,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBGMFilename(scenes[i%len(scenes)])
	}
}
