// Package guardian - Story 9-11: MomentumController Integration Tests
package guardian

import (
	"sync"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
)

// TestBindToMomentumController tests the bidirectional binding
// AC1: Bidirectional binding established
func TestBindToMomentumController(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	// Bind TensionManager to MomentumController
	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	// Verify bidirectional references
	if !mc.HasTensionManager() {
		t.Error("Expected MomentumController to have TensionManager")
	}

	if mc.GetTensionManager() == nil {
		t.Error("Expected MomentumController.GetTensionManager() to return TensionManager")
	}

	// Verify initial sync occurred
	lastTension := tm.GetLastKnownTension()
	if lastTension != 0.0 {
		t.Errorf("Expected initial tension 0.0, got %f", lastTension)
	}

	lastPhase := tm.GetLastKnownPhase()
	if lastPhase != "Rest" {
		t.Errorf("Expected initial phase Rest, got %s", lastPhase)
	}
}

// TestBindToNilMomentumController tests error handling
func TestBindToNilMomentumController(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)

	err := tm.BindToMomentumController(nil)
	if err == nil {
		t.Error("Expected error when binding to nil MomentumController")
	}
}

// TestSyncFromMomentum tests tension synchronization
// AC2: Guardian receives automatic tension notifications
func TestSyncFromMomentum(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)

	// Sync tension
	tm.SyncFromMomentum(0.5, "Buildup")

	// Verify synced values
	if tension := tm.GetLastKnownTension(); tension != 0.5 {
		t.Errorf("Expected tension 0.5, got %f", tension)
	}

	if phase := tm.GetLastKnownPhase(); phase != "Buildup" {
		t.Errorf("Expected phase Buildup, got %s", phase)
	}

	// Sync again with different values
	tm.SyncFromMomentum(0.8, "Peak")

	if tension := tm.GetLastKnownTension(); tension != 0.8 {
		t.Errorf("Expected tension 0.8, got %f", tension)
	}

	if phase := tm.GetLastKnownPhase(); phase != "Peak" {
		t.Errorf("Expected phase Peak, got %s", phase)
	}
}

// TestAdjustMomentumTension tests Guardian adjusting MomentumController tension
// AC3: Guardian can adjust MomentumController tension
func TestAdjustMomentumTension(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	// Bind
	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	// Set initial tension in MomentumController
	mc.SetTension(0.5)

	// Guardian adjusts tension
	err = tm.AdjustMomentumTension(-0.2, "test_reduction")
	if err != nil {
		t.Fatalf("AdjustMomentumTension failed: %v", err)
	}

	// Verify MomentumController tension was adjusted
	newTension := mc.GetTension()
	if newTension != 0.3 {
		t.Errorf("Expected tension 0.3 after adjustment, got %f", newTension)
	}

	// Verify TensionManager received sync (automatic bidirectional sync)
	syncedTension := tm.GetLastKnownTension()
	if syncedTension != 0.3 {
		t.Errorf("Expected synced tension 0.3, got %f", syncedTension)
	}

	// Verify phase also updated
	expectedPhase := "Buildup"
	if phase := tm.GetLastKnownPhase(); phase != expectedPhase {
		t.Errorf("Expected phase %s, got %s", expectedPhase, phase)
	}
}

// TestAdjustMomentumTensionWithoutBinding tests error handling
func TestAdjustMomentumTensionWithoutBinding(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)

	// Try to adjust without binding
	err := tm.AdjustMomentumTension(-0.2, "test")
	if err == nil {
		t.Error("Expected error when adjusting tension without bound MomentumController")
	}
}

// TestBidirectionalSync tests that changes in MomentumController sync to Guardian
// AC2: Automatic synchronization
func TestBidirectionalSync(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	// Bind
	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	// Change tension in MomentumController
	mc.SetTension(0.7)

	// Verify Guardian received the update
	if tension := tm.GetLastKnownTension(); tension != 0.7 {
		t.Errorf("Expected Guardian to sync tension 0.7, got %f", tension)
	}

	if phase := tm.GetLastKnownPhase(); phase != "Peak" {
		t.Errorf("Expected Guardian to sync phase Peak, got %s", phase)
	}

	// Adjust tension in MomentumController
	mc.AdjustTension(0.15)

	// Verify Guardian received the update
	if tension := tm.GetLastKnownTension(); tension != 0.85 {
		t.Errorf("Expected Guardian to sync tension 0.85, got %f", tension)
	}

	if phase := tm.GetLastKnownPhase(); phase != "Peak" {
		t.Errorf("Expected Guardian to sync phase Peak, got %s", phase)
	}
}

// TestGuardianInitiatedChange tests Guardian adjusting tension
// AC3: Guardian adjustments work correctly
func TestGuardianInitiatedChange(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	// Bind
	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	// Set high tension
	mc.SetTension(0.9)

	// Guardian reduces tension (protection mechanism)
	err = tm.AdjustMomentumTension(-0.3, "guardian_protection")
	if err != nil {
		t.Fatalf("Guardian adjustment failed: %v", err)
	}

	// Verify both systems are in sync
	mcTension := mc.GetTension()
	tmTension := tm.GetLastKnownTension()

	if !floatEquals(mcTension, 0.6) {
		t.Errorf("Expected MomentumController tension ~0.6, got %f", mcTension)
	}

	if !floatEquals(tmTension, 0.6) {
		t.Errorf("Expected Guardian synced tension ~0.6, got %f", tmTension)
	}

	// Verify phase
	mcPhase := mc.GetCurrentPhase()
	tmPhase := tm.GetLastKnownPhase()

	if mcPhase != "Peak" {
		t.Errorf("Expected MomentumController phase Peak, got %s", mcPhase)
	}

	if tmPhase != "Peak" {
		t.Errorf("Expected Guardian synced phase Peak, got %s", tmPhase)
	}
}

// TestMultipleAdjustments tests multiple tension adjustments
func TestMultipleAdjustments(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	// Multiple adjustments
	adjustments := []struct {
		delta  float64
		reason string
	}{
		{0.2, "increase1"},
		{0.2, "increase2"},
		{-0.1, "decrease1"},
		{0.3, "increase3"},
	}

	for _, adj := range adjustments {
		err = tm.AdjustMomentumTension(adj.delta, adj.reason)
		if err != nil {
			t.Fatalf("Adjustment failed: %v", err)
		}
	}

	// Verify final state (0 + 0.2 + 0.2 - 0.1 + 0.3 = 0.6)
	expectedTension := 0.6
	mcTension := mc.GetTension()
	tmTension := tm.GetLastKnownTension()

	if !floatEquals(mcTension, expectedTension) {
		t.Errorf("Expected MomentumController tension ~%f, got %f", expectedTension, mcTension)
	}

	if !floatEquals(tmTension, expectedTension) {
		t.Errorf("Expected Guardian synced tension ~%f, got %f", expectedTension, tmTension)
	}
}

// TestConcurrentSync tests concurrent synchronization
// AC6: Concurrent safety
func TestConcurrentSync(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				if id%2 == 0 {
					// MomentumController side changes
					mc.SetTension(float64(j%100) / 100.0)
				} else {
					// Guardian side changes
					_ = tm.AdjustMomentumTension(0.01, "concurrent_test")
				}

				// Read operations
				_ = tm.GetLastKnownTension()
				_ = tm.GetLastKnownPhase()
			}
		}(i)
	}

	wg.Wait()

	// Verify no panics and state is consistent
	mcTension := mc.GetTension()
	tmTension := tm.GetLastKnownTension()

	if mcTension < 0.0 || mcTension > 1.0 {
		t.Errorf("Invalid MomentumController tension: %f", mcTension)
	}

	if tmTension < 0.0 || tmTension > 1.0 {
		t.Errorf("Invalid Guardian synced tension: %f", tmTension)
	}

	// They should be in sync (within floating point tolerance)
	if (mcTension - tmTension) > 0.01 || (mcTension - tmTension) < -0.01 {
		t.Errorf("Tensions out of sync: MC=%f, TM=%f", mcTension, tmTension)
	}
}

// TestNoCircularCalls verifies that syncing doesn't cause circular calls
func TestNoCircularCalls(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	// Set tension multiple times - should not cause infinite loops
	for i := 0; i < 10; i++ {
		mc.SetTension(float64(i) / 10.0)
		err = tm.AdjustMomentumTension(0.05, "test")
		if err != nil {
			t.Fatalf("Adjustment failed: %v", err)
		}
	}

	// If we get here without hanging, no circular calls occurred
	if tension := mc.GetTension(); tension < 0.0 || tension > 1.0 {
		t.Errorf("Invalid tension: %f", tension)
	}
}

// TestPhaseSync tests that phase changes are synced correctly
// AC4: Phase transitions sync to Guardian
func TestPhaseSync(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)

	err := tm.BindToMomentumController(mc)
	if err != nil {
		t.Fatalf("BindToMomentumController failed: %v", err)
	}

	testCases := []struct {
		tension       float64
		expectedPhase string
	}{
		{0.1, "Rest"},
		{0.3, "Buildup"},
		{0.7, "Peak"},
		{0.95, "Release"},
	}

	for _, tc := range testCases {
		mc.SetTension(tc.tension)

		mcPhase := mc.GetCurrentPhase()
		tmPhase := tm.GetLastKnownPhase()

		if mcPhase != tc.expectedPhase {
			t.Errorf("Tension %f: expected MomentumController phase %s, got %s",
				tc.tension, tc.expectedPhase, mcPhase)
		}

		if tmPhase != tc.expectedPhase {
			t.Errorf("Tension %f: expected Guardian synced phase %s, got %s",
				tc.tension, tc.expectedPhase, tmPhase)
		}
	}
}

// TestRebinding tests rebinding to a different MomentumController
func TestRebinding(t *testing.T) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)

	// Bind to first MomentumController
	mc1 := momentum.NewMomentumController(nil, nil)
	err := tm.BindToMomentumController(mc1)
	if err != nil {
		t.Fatalf("First binding failed: %v", err)
	}

	mc1.SetTension(0.5)

	// Bind to second MomentumController
	mc2 := momentum.NewMomentumController(nil, nil)
	err = tm.BindToMomentumController(mc2)
	if err != nil {
		t.Fatalf("Second binding failed: %v", err)
	}

	mc2.SetTension(0.8)

	// Verify Guardian is synced with second controller
	if tension := tm.GetLastKnownTension(); !floatEquals(tension, 0.8) {
		t.Errorf("Expected Guardian synced to second controller tension ~0.8, got %f", tension)
	}

	// Verify first controller is no longer bound
	// (Changes to mc1 should not affect tm)
	mc1.SetTension(0.2)

	// Guardian should still be synced with mc2
	if tension := tm.GetLastKnownTension(); !floatEquals(tension, 0.8) {
		t.Errorf("Expected Guardian to remain synced with second controller at ~0.8, got %f", tension)
	}
}

// Benchmark integration operations
func BenchmarkBindToMomentumController(b *testing.B) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm := NewTensionManager(config, nil, tmConfig)
		mc := momentum.NewMomentumController(nil, nil)
		_ = tm.BindToMomentumController(mc)
	}
}

func BenchmarkSyncFromMomentum(b *testing.B) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.SyncFromMomentum(float64(i%100)/100.0, "Peak")
	}
}

func BenchmarkAdjustMomentumTension(b *testing.B) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)
	_ = tm.BindToMomentumController(mc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.AdjustMomentumTension(0.01, "benchmark")
	}
}

func BenchmarkBidirectionalSync(b *testing.B) {
	config := DefaultGuardianConfig()
	tmConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(config, nil, tmConfig)
	mc := momentum.NewMomentumController(nil, nil)
	_ = tm.BindToMomentumController(mc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.SetTension(float64(i%100) / 100.0)
		_ = tm.GetLastKnownTension()
	}
}
