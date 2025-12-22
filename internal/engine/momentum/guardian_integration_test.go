// Package momentum - Story 9-11: Guardian Integration Tests
package momentum

import (
	"sync"
	"testing"
	"time"
)

// mockTensionManager implements TensionManager interface for testing
type mockTensionManager struct {
	syncCalls       []syncCall
	mu              sync.Mutex
	syncFromMomentumFunc func(float64, string)
}

type syncCall struct {
	tensionValue float64
	phase        string
	timestamp    time.Time
}

func newMockTensionManager() *mockTensionManager {
	return &mockTensionManager{
		syncCalls: make([]syncCall, 0),
	}
}

func (m *mockTensionManager) SyncFromMomentum(tensionValue float64, phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.syncCalls = append(m.syncCalls, syncCall{
		tensionValue: tensionValue,
		phase:        phase,
		timestamp:    time.Now(),
	})

	if m.syncFromMomentumFunc != nil {
		m.syncFromMomentumFunc(tensionValue, phase)
	}
}

func (m *mockTensionManager) getSyncCalls() []syncCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]syncCall, len(m.syncCalls))
	copy(calls, m.syncCalls)
	return calls
}

func (m *mockTensionManager) clearSyncCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.syncCalls = make([]syncCall, 0)
}

// TestSetTensionManager tests setting and getting TensionManager
// AC1: MomentumController can hold TensionManager reference (optional)
func TestSetTensionManager(t *testing.T) {
	mc := NewMomentumController(nil, nil)

	// Initially no TensionManager
	if mc.HasTensionManager() {
		t.Error("Expected no TensionManager initially")
	}

	if tm := mc.GetTensionManager(); tm != nil {
		t.Error("Expected GetTensionManager to return nil initially")
	}

	// Set TensionManager
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	// Verify TensionManager is set
	if !mc.HasTensionManager() {
		t.Error("Expected HasTensionManager to return true")
	}

	if tm := mc.GetTensionManager(); tm != mockTM {
		t.Error("Expected GetTensionManager to return the set TensionManager")
	}
}

// TestSetTensionSyncToGuardian tests that SetTension automatically syncs to Guardian
// AC2: MomentumController.SetTension() automatically notifies Guardian
func TestSetTensionSyncToGuardian(t *testing.T) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	// Set tension
	mc.SetTension(0.5)

	// Verify sync was called
	calls := mockTM.getSyncCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 sync call, got %d", len(calls))
	}

	if calls[0].tensionValue != 0.5 {
		t.Errorf("Expected tension 0.5, got %f", calls[0].tensionValue)
	}

	if calls[0].phase != "Buildup" {
		t.Errorf("Expected phase Buildup, got %s", calls[0].phase)
	}

	// Verify MomentumController state
	if tension := mc.GetTension(); tension != 0.5 {
		t.Errorf("Expected tension 0.5, got %f", tension)
	}

	if phase := mc.GetCurrentPhase(); phase != "Buildup" {
		t.Errorf("Expected phase Buildup, got %s", phase)
	}
}

// TestAdjustTensionSyncToGuardian tests that AdjustTension automatically syncs to Guardian
// AC2: MomentumController.AdjustTension() automatically notifies Guardian
func TestAdjustTensionSyncToGuardian(t *testing.T) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	// Set initial tension
	mc.SetTension(0.3)
	mockTM.clearSyncCalls()

	// Adjust tension
	mc.AdjustTension(0.2)

	// Verify sync was called
	calls := mockTM.getSyncCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 sync call, got %d", len(calls))
	}

	if calls[0].tensionValue != 0.5 {
		t.Errorf("Expected tension 0.5, got %f", calls[0].tensionValue)
	}

	if calls[0].phase != "Buildup" {
		t.Errorf("Expected phase Buildup, got %s", calls[0].phase)
	}
}

// TestTensionPhaseMappings tests phase transitions
// AC4: Phase transitions automatically sync to Guardian
func TestTensionPhaseMappings(t *testing.T) {
	tests := []struct {
		tension       float64
		expectedPhase string
	}{
		{0.0, "Rest"},
		{0.1, "Rest"},
		{0.24, "Rest"},
		{0.25, "Buildup"},
		{0.5, "Buildup"},
		{0.59, "Buildup"},
		{0.60, "Peak"},
		{0.8, "Peak"},
		{0.89, "Peak"},
		{0.90, "Release"},
		{1.0, "Release"},
	}

	for _, tt := range tests {
		mc := NewMomentumController(nil, nil)
		mockTM := newMockTensionManager()
		mc.SetTensionManager(mockTM)

		mc.SetTension(tt.tension)

		phase := mc.GetCurrentPhase()
		if phase != tt.expectedPhase {
			t.Errorf("Tension %f: expected phase %s, got %s", tt.tension, tt.expectedPhase, phase)
		}

		// Verify sync call
		calls := mockTM.getSyncCalls()
		if len(calls) != 1 {
			t.Errorf("Expected 1 sync call, got %d", len(calls))
			continue
		}

		if calls[0].phase != tt.expectedPhase {
			t.Errorf("Sync call: expected phase %s, got %s", tt.expectedPhase, calls[0].phase)
		}
	}
}

// TestTensionClamping tests that tension is clamped to [0.0, 1.0]
func TestTensionClamping(t *testing.T) {
	mc := NewMomentumController(nil, nil)

	// Test upper bound
	mc.SetTension(1.5)
	if tension := mc.GetTension(); tension != 1.0 {
		t.Errorf("Expected tension clamped to 1.0, got %f", tension)
	}

	// Test lower bound
	mc.SetTension(-0.5)
	if tension := mc.GetTension(); tension != 0.0 {
		t.Errorf("Expected tension clamped to 0.0, got %f", tension)
	}

	// Test adjust upper bound
	mc.SetTension(0.8)
	mc.AdjustTension(0.5)
	if tension := mc.GetTension(); tension != 1.0 {
		t.Errorf("Expected tension clamped to 1.0 after adjust, got %f", tension)
	}

	// Test adjust lower bound
	mc.SetTension(0.2)
	mc.AdjustTension(-0.5)
	if tension := mc.GetTension(); tension != 0.0 {
		t.Errorf("Expected tension clamped to 0.0 after adjust, got %f", tension)
	}
}

// TestTensionWithoutGuardian tests that MomentumController works without Guardian
// AC5: MomentumController works normally without TensionManager
func TestTensionWithoutGuardian(t *testing.T) {
	mc := NewMomentumController(nil, nil)

	// Should work without TensionManager
	mc.SetTension(0.5)
	if tension := mc.GetTension(); tension != 0.5 {
		t.Errorf("Expected tension 0.5, got %f", tension)
	}

	mc.AdjustTension(0.2)
	if tension := mc.GetTension(); tension != 0.7 {
		t.Errorf("Expected tension 0.7, got %f", tension)
	}

	if phase := mc.GetCurrentPhase(); phase != "Peak" {
		t.Errorf("Expected phase Peak, got %s", phase)
	}
}

// TestMultipleTensionChanges tests multiple tension adjustments
func TestMultipleTensionChanges(t *testing.T) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	// Multiple adjustments
	mc.SetTension(0.1)  // Rest (0.1)
	mc.AdjustTension(0.2) // Buildup (0.3)
	mc.AdjustTension(0.3) // Peak (0.6)
	mc.AdjustTension(0.3) // Release (0.9 - already Release phase!)
	mc.AdjustTension(0.2) // Release (1.0, clamped)

	// Verify all syncs were called
	calls := mockTM.getSyncCalls()
	if len(calls) != 5 {
		t.Fatalf("Expected 5 sync calls, got %d", len(calls))
	}

	expectedPhases := []string{"Rest", "Buildup", "Peak", "Release", "Release"}
	for i, expectedPhase := range expectedPhases {
		if calls[i].phase != expectedPhase {
			t.Errorf("Call %d: expected phase %s, got %s", i, expectedPhase, calls[i].phase)
		}
	}

	// Verify final state
	if phase := mc.GetCurrentPhase(); phase != "Release" {
		t.Errorf("Expected final phase Release, got %s", phase)
	}

	if tension := mc.GetTension(); tension != 1.0 {
		t.Errorf("Expected final tension 1.0, got %f", tension)
	}
}

// TestConcurrentTensionAccess tests concurrent access to tension
// AC6: Concurrent safety
func TestConcurrentTensionAccess(t *testing.T) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

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
					mc.SetTension(float64(j%100) / 100.0)
				} else {
					mc.AdjustTension(0.01)
				}

				// Read operations
				_ = mc.GetTension()
				_ = mc.GetCurrentPhase()
				_ = mc.HasTensionManager()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify no panics occurred and state is valid
	tension := mc.GetTension()
	if tension < 0.0 || tension > 1.0 {
		t.Errorf("Invalid tension after concurrent access: %f", tension)
	}

	phase := mc.GetCurrentPhase()
	validPhases := map[string]bool{"Rest": true, "Buildup": true, "Peak": true, "Release": true, "": true}
	if !validPhases[phase] {
		t.Errorf("Invalid phase after concurrent access: %s", phase)
	}
}

// TestConcurrentTensionManagerAccess tests concurrent SetTensionManager calls
func TestConcurrentTensionManagerAccess(t *testing.T) {
	mc := NewMomentumController(nil, nil)

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mockTM := newMockTensionManager()
				mc.SetTensionManager(mockTM)
				_ = mc.GetTensionManager()
				_ = mc.HasTensionManager()
			}
		}()
	}

	wg.Wait()

	// Verify final state is consistent
	if !mc.HasTensionManager() {
		t.Error("Expected TensionManager to be set after concurrent access")
	}
}

// TestNoCircularCalls tests that SyncFromMomentum doesn't cause circular calls
func TestNoCircularCalls(t *testing.T) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()

	// Set up a mock that would detect circular calls
	callCount := 0
	mockTM.syncFromMomentumFunc = func(tension float64, phase string) {
		callCount++
		if callCount > 1 {
			t.Error("Circular call detected: SyncFromMomentum called more than once")
		}
	}

	mc.SetTensionManager(mockTM)
	mc.SetTension(0.5)

	if callCount != 1 {
		t.Errorf("Expected exactly 1 sync call, got %d", callCount)
	}
}

// TestPhaseTransitionsSyncToGuardian tests that phase transitions trigger syncs
// AC4: Phase changes automatically sync to Guardian
func TestPhaseTransitionsSyncToGuardian(t *testing.T) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	// Start in Rest
	mc.SetTension(0.1)
	mockTM.clearSyncCalls()

	// Transition to Buildup
	mc.SetTension(0.4)
	calls := mockTM.getSyncCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 sync call for phase transition, got %d", len(calls))
	}
	if calls[0].phase != "Buildup" {
		t.Errorf("Expected phase Buildup, got %s", calls[0].phase)
	}

	mockTM.clearSyncCalls()

	// Transition to Peak
	mc.SetTension(0.7)
	calls = mockTM.getSyncCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 sync call for phase transition, got %d", len(calls))
	}
	if calls[0].phase != "Peak" {
		t.Errorf("Expected phase Peak, got %s", calls[0].phase)
	}

	mockTM.clearSyncCalls()

	// Transition to Release
	mc.SetTension(0.95)
	calls = mockTM.getSyncCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 sync call for phase transition, got %d", len(calls))
	}
	if calls[0].phase != "Release" {
		t.Errorf("Expected phase Release, got %s", calls[0].phase)
	}
}

// TestBackwardCompatibility tests that existing tests still pass
// AC5: Existing MomentumController tests still pass
func TestBackwardCompatibility(t *testing.T) {
	// Test that MomentumController still works as before
	config := DefaultMomentumConfig()
	mc := NewMomentumController(config, nil)

	// Test basic functionality
	ctx := &NarrativeContext{
		CurrentBeat:  1,
		CurrentScene: "test",
		RiskLevel:    RiskLow,
	}

	// Should work without tension functionality
	shouldPause := mc.ShouldPauseForChoice(ctx)
	_ = shouldPause // Don't care about the result, just that it doesn't panic

	// Test AutoResolve
	result := mc.AutoResolve(ctx)
	if result == nil {
		t.Error("Expected AutoResolve to return a result")
	}

	// Verify no panic occurs
	if t.Failed() {
		t.Error("Backward compatibility test failed - existing functionality broken")
	}
}

// Benchmark tension operations
func BenchmarkSetTension(b *testing.B) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.SetTension(float64(i%100) / 100.0)
	}
}

func BenchmarkAdjustTension(b *testing.B) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.AdjustTension(0.01)
	}
}

func BenchmarkGetTension(b *testing.B) {
	mc := NewMomentumController(nil, nil)
	mc.SetTension(0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mc.GetTension()
	}
}

func BenchmarkConcurrentTensionAccess(b *testing.B) {
	mc := NewMomentumController(nil, nil)
	mockTM := newMockTensionManager()
	mc.SetTensionManager(mockTM)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				mc.SetTension(0.5)
			} else {
				_ = mc.GetTension()
			}
			i++
		}
	})
}
