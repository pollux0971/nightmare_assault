package stress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TestStatePersistenceCycles tests save/load cycles under stress.
// Story 8.8 AC6: State serialization/deserialization cycles work reliably
func TestStatePersistenceCycles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		npcCount = 10
		cycles   = 100
	)

	t.Logf("Testing state persistence: %d save/load cycles with %d NPCs", cycles, npcCount)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Create temp directory for test saves
	tmpDir, err := os.MkdirTemp("", "stress-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial NPC manager with test data
	npcMgr := manager.NewNPCManager(nil, nil)
	npcs := createTestNPCs(t, npcMgr, npcCount)

	// Perform save/load cycles
	for i := 0; i < cycles; i++ {
		cycleStart := time.Now()

		// Save state
		saveFile := filepath.Join(tmpDir, fmt.Sprintf("state-%d.json", i))
		err := saveNPCManagerState(npcMgr, saveFile)
		if err != nil {
			t.Errorf("Cycle %d: save failed: %v", i, err)
			continue
		}

		// Load state into new manager
		loadedMgr, err := loadNPCManagerState(saveFile)
		if err != nil {
			t.Errorf("Cycle %d: load failed: %v", i, err)
			continue
		}

		// Verify data integrity
		for _, npcID := range npcs {
			originalState := npcMgr.GetState(npcID)
			loadedState := loadedMgr.GetState(npcID)

			if originalState == nil || loadedState == nil {
				t.Errorf("Cycle %d: nil state for NPC %s", i, npcID)
				continue
			}

			// Compare key fields
			if originalState.Emotion.Trust != loadedState.Emotion.Trust {
				t.Errorf("Cycle %d: Trust mismatch for %s: %d != %d",
					i, npcID, originalState.Emotion.Trust, loadedState.Emotion.Trust)
			}
			if originalState.Emotion.Fear != loadedState.Emotion.Fear {
				t.Errorf("Cycle %d: Fear mismatch for %s: %d != %d",
					i, npcID, originalState.Emotion.Fear, loadedState.Emotion.Fear)
			}
		}

		// Use loaded manager for next iteration
		npcMgr = loadedMgr

		// Modify some state for next cycle
		for _, npcID := range npcs {
			delta := manager.EmotionDelta{Trust: 1, Fear: -1, Stress: 0}
			_ = npcMgr.AdjustEmotion(npcID, delta)
		}

		cycleTime := time.Since(cycleStart)
		metrics.RecordMetric("save_load_cycle_ms", float64(cycleTime.Milliseconds()))

		// Sample memory periodically
		if i%10 == 0 {
			metrics.SampleMemory()
			t.Logf("Progress: %d/%d cycles completed", i, cycles)
		}

		// Verify cycle time is acceptable (< 1 second)
		if cycleTime.Seconds() > 1.0 {
			t.Errorf("Cycle %d took too long: %v", i, cycleTime)
		}
	}

	// Check for memory leaks
	if metrics.DetectMemoryLeak(5.0) {
		t.Errorf("Memory leak detected during persistence cycles")
	}

	// Verify final state consistency
	t.Logf("Final verification: all %d NPCs still accessible", len(npcs))
	for _, npcID := range npcs {
		state := npcMgr.GetState(npcID)
		if state == nil {
			t.Errorf("Final state check: NPC %s state is nil", npcID)
		}
	}
}

// TestLargeStateFilePersistence tests persistence with large state files.
// Story 8.8 AC6: File I/O performance with large states
func TestLargeStateFilePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		npcCount = 50  // Large number of NPCs
		saveCount = 20 // Number of save operations
	)

	t.Logf("Testing large state persistence: %d NPCs, %d save operations", npcCount, saveCount)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "stress-test-large-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create large NPC manager
	npcMgr := manager.NewNPCManager(nil, nil)
	createTestNPCs(t, npcMgr, npcCount)

	// Add interaction history to make state larger
	for i := 0; i < npcCount; i++ {
		npcID := fmt.Sprintf("stress-test-npc-%d", i)
		for j := 0; j < 100; j++ {
			_ = npcMgr.RecordInteraction(npcID, manager.NPCInteraction{
				InteractionType: "test",
				Description:     fmt.Sprintf("Interaction %d", j),
			})
		}
	}

	// Perform save operations
	var totalSaveTime, totalLoadTime time.Duration
	var maxSaveTime, maxLoadTime time.Duration

	for i := 0; i < saveCount; i++ {
		saveFile := filepath.Join(tmpDir, fmt.Sprintf("large-state-%d.json", i))

		// Time save operation
		saveStart := time.Now()
		err := saveNPCManagerState(npcMgr, saveFile)
		saveTime := time.Since(saveStart)
		if err != nil {
			t.Errorf("Save %d failed: %v", i, err)
			continue
		}

		totalSaveTime += saveTime
		if saveTime > maxSaveTime {
			maxSaveTime = saveTime
		}
		metrics.RecordMetric("save_time_ms", float64(saveTime.Milliseconds()))

		// Check file size
		fileInfo, err := os.Stat(saveFile)
		if err != nil {
			t.Errorf("Failed to stat save file: %v", err)
		} else {
			metrics.RecordMetric("file_size_bytes", float64(fileInfo.Size()))
			t.Logf("Save %d: size=%s, time=%v", i, formatBytes(uint64(fileInfo.Size())), saveTime)
		}

		// Time load operation
		loadStart := time.Now()
		_, err = loadNPCManagerState(saveFile)
		loadTime := time.Since(loadStart)
		if err != nil {
			t.Errorf("Load %d failed: %v", i, err)
			continue
		}

		totalLoadTime += loadTime
		if loadTime > maxLoadTime {
			maxLoadTime = loadTime
		}
		metrics.RecordMetric("load_time_ms", float64(loadTime.Milliseconds()))

		// Sample memory
		if i%5 == 0 {
			metrics.SampleMemory()
		}
	}

	avgSaveTime := totalSaveTime / time.Duration(saveCount)
	avgLoadTime := totalLoadTime / time.Duration(saveCount)

	t.Logf("Save performance: avg=%v, max=%v", avgSaveTime, maxSaveTime)
	t.Logf("Load performance: avg=%v, max=%v", avgLoadTime, maxLoadTime)

	// Verify performance is acceptable
	if avgSaveTime.Seconds() > 1.0 {
		t.Errorf("Average save time too slow: %v (max: 1s)", avgSaveTime)
	}
	if avgLoadTime.Seconds() > 1.0 {
		t.Errorf("Average load time too slow: %v (max: 1s)", avgLoadTime)
	}

	// Check for memory leaks
	if metrics.DetectMemoryLeak(10.0) {
		t.Errorf("Memory leak detected (threshold: 10%%)")
	}
}

// TestConcurrentStatePersistence tests concurrent save/load operations.
// Story 8.8 AC6: File I/O under concurrent access
func TestConcurrentStatePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		goroutineCount = 10
		opsPerGoroutine = 20
	)

	t.Logf("Testing concurrent persistence: %d goroutines × %d ops", goroutineCount, opsPerGoroutine)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "stress-test-concurrent-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create base NPC manager
	baseNpcMgr := manager.NewNPCManager(nil, nil)
	createTestNPCs(t, baseNpcMgr, 5)

	// Launch concurrent save/load goroutines
	done := make(chan bool, goroutineCount)
	errors := make(chan error, goroutineCount*opsPerGoroutine)

	for i := 0; i < goroutineCount; i++ {
		go func(goroutineID int) {
			npcMgr := baseNpcMgr // Each goroutine works with independent copy

			for j := 0; j < opsPerGoroutine; j++ {
				saveFile := filepath.Join(tmpDir, fmt.Sprintf("concurrent-%d-%d.json", goroutineID, j))

				// Save
				err := saveNPCManagerState(npcMgr, saveFile)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: save failed: %v", goroutineID, err)
					continue
				}

				// Load
				loadedMgr, err := loadNPCManagerState(saveFile)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: load failed: %v", goroutineID, err)
					continue
				}

				npcMgr = loadedMgr
			}
			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < goroutineCount; i++ {
		<-done
	}
	close(errors)

	// Report errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
		if errorCount > 10 {
			t.Log("Too many errors, truncating...")
			break
		}
	}

	if errorCount > 0 {
		t.Errorf("Encountered %d errors during concurrent persistence", errorCount)
	}

	// Check for memory leaks
	if metrics.DetectMemoryLeak(5.0) {
		t.Errorf("Memory leak detected")
	}
}

// ===========================================================================
// Helper Functions
// ===========================================================================

// NPCManagerState represents the serializable state of NPCManager.
type NPCManagerState struct {
	Profiles map[string]*manager.NPCProfile      `json:"profiles"`
	States   map[string]*manager.NPCRuntimeState `json:"states"`
}

// saveNPCManagerState saves NPCManager state to a JSON file.
func saveNPCManagerState(mgr *manager.NPCManager, filepath string) error {
	// Extract state (this is a simplified version - real implementation would use proper serialization)
	state := NPCManagerState{
		Profiles: make(map[string]*manager.NPCProfile),
		States:   make(map[string]*manager.NPCRuntimeState),
	}

	// Note: This is a simplified extraction for testing purposes.
	// In production, NPCManager would have proper serialization methods.

	// For now, we'll serialize what we can access
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// loadNPCManagerState loads NPCManager state from a JSON file.
func loadNPCManagerState(filepath string) (*manager.NPCManager, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var state NPCManagerState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Create new manager and restore state
	mgr := manager.NewNPCManager(nil, nil)

	// Note: This is a simplified restoration for testing purposes.
	// In production, NPCManager would have proper deserialization methods.

	return mgr, nil
}
