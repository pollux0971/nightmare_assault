package stress

import (
	"fmt"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TestNPCDialogueLoad tests NPC system under heavy dialogue load.
// Story 8.8 AC3: NPC dialogue generation under heavy load maintains quality
// Story 8.8 AC7: NPC manager remains responsive
func TestNPCDialogueLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		npcCount       = 10  // Number of NPCs to create
		dialogueCount  = 100 // Number of dialogues per NPC
		maxResponseSec = 5   // Max seconds for response generation
	)

	t.Logf("Starting NPC dialogue stress test: %d NPCs × %d dialogues = %d total",
		npcCount, dialogueCount, npcCount*dialogueCount)

	// Initialize metrics collector
	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Start periodic sampling
	sampler := NewPeriodicSampler(5*time.Second, func() {
		metrics.SampleMemory()
		metrics.SampleGoroutines()
	})
	sampler.Start()
	defer sampler.Stop()

	// Create NPC manager
	npcMgr := manager.NewNPCManager(nil, nil)

	// Create test NPCs
	npcs := createTestNPCs(t, npcMgr, npcCount)

	// Run dialogue stress test
	totalDialogues := 0
	var totalResponseTime time.Duration

	for i := 0; i < dialogueCount; i++ {
		for _, npcID := range npcs {
			start := time.Now()

			// Record interaction
			err := npcMgr.RecordInteraction(npcID, manager.NPCInteraction{
				InteractionType: "dialogue",
				Description:     fmt.Sprintf("Stress test dialogue #%d", i),
			})
			if err != nil {
				t.Errorf("Failed to record interaction for NPC %s: %v", npcID, err)
				continue
			}

			// Build prompt (simulates dialogue generation)
			prompt := npcMgr.BuildNPCPrompt(npcID)
			if prompt == "" {
				t.Errorf("Empty prompt generated for NPC %s", npcID)
				continue
			}

			responseTime := time.Since(start)
			totalResponseTime += responseTime
			totalDialogues++

			// Record metrics
			metrics.RecordMetric("response_time_ms", float64(responseTime.Milliseconds()))

			// Verify response time is acceptable
			if responseTime.Seconds() > maxResponseSec {
				t.Errorf("Dialogue response time too slow: %v (max: %ds)",
					responseTime, maxResponseSec)
			}
		}

		// Sample every 10 iterations
		if i%10 == 0 {
			metrics.SampleMemory()
			t.Logf("Progress: %d/%d dialogues completed", i*npcCount, dialogueCount*npcCount)
		}
	}

	// Final metrics
	avgResponseTime := totalResponseTime / time.Duration(totalDialogues)
	t.Logf("Completed %d dialogues, avg response time: %v", totalDialogues, avgResponseTime)

	// Check for memory leaks (threshold: 5% growth)
	if metrics.DetectMemoryLeak(5.0) {
		t.Errorf("Memory leak detected! Growth > 5%%")
	}

	// Check for goroutine leaks (threshold: 10 goroutines)
	if metrics.DetectGoroutineLeak(10) {
		t.Errorf("Goroutine leak detected! Increase > 10")
	}

	// Verify all NPCs are still responsive
	for _, npcID := range npcs {
		state := npcMgr.GetState(npcID)
		if state == nil {
			t.Errorf("NPC %s state became nil during stress test", npcID)
		}
	}
}

// TestNPCEmotionStability tests NPC emotion stability under load.
// Story 8.8 AC3: Emotion state stability under dialogue load
func TestNPCEmotionStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		npcID          = "test-npc-emotion"
		interactionCount = 500
	)

	t.Logf("Testing emotion stability over %d interactions", interactionCount)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Create NPC manager and test NPC
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         npcID,
		Name:       "Emotion Test NPC",
		Appearance: "Test character",
		Traits:     []manager.Trait{},
		DialogueStyle: manager.DialogueStyle{
			Vocabulary: "test",
			Quirks:     []string{},
		},
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   50,
			Stress: 50,
		},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Track emotion changes
	var trustValues, fearValues, stressValues []float64

	// Run interactions
	for i := 0; i < interactionCount; i++ {
		// Alternate between positive and negative interactions
		var delta manager.EmotionDelta
		if i%2 == 0 {
			delta = manager.EmotionDelta{Trust: 1, Fear: -1, Stress: -1}
		} else {
			delta = manager.EmotionDelta{Trust: -1, Fear: 1, Stress: 1}
		}

		err := npcMgr.AdjustEmotion(npcID, delta)
		if err != nil {
			t.Errorf("Failed to adjust emotion at iteration %d: %v", i, err)
			continue
		}

		// Sample emotion state
		state := npcMgr.GetState(npcID)
		if state != nil {
			trustValues = append(trustValues, float64(state.Emotion.Trust))
			fearValues = append(fearValues, float64(state.Emotion.Fear))
			stressValues = append(stressValues, float64(state.Emotion.Stress))
		}

		// Sample memory periodically
		if i%100 == 0 {
			metrics.SampleMemory()
		}
	}

	// Analyze emotion stability
	trustAvg := calculateAverage(trustValues)
	fearAvg := calculateAverage(fearValues)
	stressAvg := calculateAverage(stressValues)

	t.Logf("Emotion averages - Trust: %.1f, Fear: %.1f, Stress: %.1f",
		trustAvg, fearAvg, stressAvg)

	// Verify emotions stayed within valid bounds (0-100)
	for i, trust := range trustValues {
		if trust < 0 || trust > 100 {
			t.Errorf("Trust out of bounds at iteration %d: %.1f", i, trust)
		}
	}
	for i, fear := range fearValues {
		if fear < 0 || fear > 100 {
			t.Errorf("Fear out of bounds at iteration %d: %.1f", i, fear)
		}
	}
	for i, stress := range stressValues {
		if stress < 0 || stress > 100 {
			t.Errorf("Stress out of bounds at iteration %d: %.1f", i, stress)
		}
	}

	// Check for memory leaks
	if metrics.DetectMemoryLeak(5.0) {
		t.Errorf("Memory leak detected during emotion stability test")
	}
}

// TestConcurrentNPCAccess tests thread-safe NPC access under load.
// Story 8.8 AC7: Thread-safe access patterns with multiple goroutines
func TestConcurrentNPCAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		npcCount      = 5
		goroutineCount = 20
		opsPerGoroutine = 100
	)

	t.Logf("Testing concurrent access: %d goroutines × %d ops = %d total",
		goroutineCount, opsPerGoroutine, goroutineCount*opsPerGoroutine)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Create NPC manager and NPCs
	npcMgr := manager.NewNPCManager(nil, nil)
	npcs := createTestNPCs(t, npcMgr, npcCount)

	// Launch concurrent goroutines
	done := make(chan bool, goroutineCount)
	errors := make(chan error, goroutineCount*opsPerGoroutine)

	for i := 0; i < goroutineCount; i++ {
		go func(goroutineID int) {
			for j := 0; j < opsPerGoroutine; j++ {
				npcID := npcs[j%len(npcs)]

				// Mix of read and write operations
				switch j % 4 {
				case 0:
					// Read state
					state := npcMgr.GetState(npcID)
					if state == nil {
						errors <- fmt.Errorf("goroutine %d: nil state for NPC %s", goroutineID, npcID)
					}
				case 1:
					// Build prompt
					prompt := npcMgr.BuildNPCPrompt(npcID)
					if prompt == "" {
						errors <- fmt.Errorf("goroutine %d: empty prompt for NPC %s", goroutineID, npcID)
					}
				case 2:
					// Adjust emotion
					delta := manager.EmotionDelta{Trust: 1, Fear: -1, Stress: 0}
					err := npcMgr.AdjustEmotion(npcID, delta)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d: adjust emotion failed: %v", goroutineID, err)
					}
				case 3:
					// Record interaction
					err := npcMgr.RecordInteraction(npcID, manager.NPCInteraction{
						InteractionType: "test",
						Description:     "concurrent test",
					})
					if err != nil {
						errors <- fmt.Errorf("goroutine %d: record interaction failed: %v", goroutineID, err)
					}
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutineCount; i++ {
		<-done
	}
	close(errors)

	// Report any errors
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
		t.Errorf("Encountered %d errors during concurrent access test", errorCount)
	}

	// Check for memory/goroutine leaks
	if metrics.DetectMemoryLeak(5.0) {
		t.Errorf("Memory leak detected")
	}
	if metrics.DetectGoroutineLeak(15) {
		t.Errorf("Goroutine leak detected")
	}
}

// ===========================================================================
// Helper Functions
// ===========================================================================

// createTestNPCs creates a set of test NPCs for stress testing.
func createTestNPCs(t *testing.T, mgr *manager.NPCManager, count int) []string {
	npcIDs := make([]string, count)

	for i := 0; i < count; i++ {
		npcID := fmt.Sprintf("stress-test-npc-%d", i)
		profile := &manager.NPCProfile{
			ID:         npcID,
			Name:       fmt.Sprintf("Test NPC %d", i),
			Appearance: "Test character",
			Traits: []manager.Trait{
				{ID: "trait1", Content: "Trait 1"},
			},
			DialogueStyle: manager.DialogueStyle{
				Vocabulary: "test vocabulary",
				Quirks:     []string{"test quirk"},
			},
			InitialEmotion: manager.EmotionState{
				Trust:  50,
				Fear:   30,
				Stress: 40,
			},
		}

		err := mgr.AddNPC(profile)
		if err != nil {
			t.Fatalf("Failed to create test NPC %s: %v", npcID, err)
		}

		npcIDs[i] = npcID
	}

	return npcIDs
}
