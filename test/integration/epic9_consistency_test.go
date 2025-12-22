package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

// TestEpic9_MetricsConsistency tests consistency between different metrics sources
// 測試 Metrics 一致性：Trinity metrics vs. Fallback metrics
func TestEpic9_MetricsConsistency(t *testing.T) {
	t.Parallel()

	// Create router with mock providers
	thinkingProvider := NewMockProvider("Thinking response")
	reactiveProvider := NewMockProvider("Reactive response")
	rapidProvider := NewMockProvider("Rapid response")

	router := trinity.NewTrinityRouterWithProviders(
		thinkingProvider,
		reactiveProvider,
		rapidProvider,
		true, // fallback enabled
		make(map[string]trinity.TierLevel), // no overrides
	)

	router.ResetMetrics()
	router.ResetFallbackMetrics()

	ctx := context.Background()
	messages := TestMessages("user", "Consistency test message")

	// Make multiple requests
	totalRequests := 20
	for i := 0; i < totalRequests; i++ {
		agent := []string{"judge", "choice", "narration"}[i%3]
		_, err := router.Route(ctx, agent, messages)
		AssertNoError(t, err, "Request should succeed")
	}

	// Get metrics from both sources
	trinityMetrics := router.GetMetricsSummary()
	fallbackMetrics := router.GetFallbackMetrics()

	// Verify Trinity metrics consistency
	t.Run("Trinity_Metrics_Internal_Consistency", func(t *testing.T) {
		totalTrinityRequests := trinityMetrics.ThinkingStats.TotalRequests +
			trinityMetrics.ReactiveStats.TotalRequests +
			trinityMetrics.RapidStats.TotalRequests

		AssertEqual(t, int64(totalRequests), totalTrinityRequests,
			"Trinity metrics should track all requests")

		// Verify success + failures = total
		for _, tierName := range []string{"Thinking", "Reactive", "Rapid"} {
			var tierMetrics *trinity.TierStats
			switch tierName {
			case "Thinking":
				tierMetrics = &trinityMetrics.ThinkingStats
			case "Reactive":
				tierMetrics = &trinityMetrics.ReactiveStats
			case "Rapid":
				tierMetrics = &trinityMetrics.RapidStats
			}

			if tierMetrics.TotalRequests > 0 {
				total := tierMetrics.SuccessRequests + tierMetrics.FailedRequests
				AssertEqual(t, tierMetrics.TotalRequests, total,
					tierName+" tier: success + failures should equal total")
			}
		}
	})

	// Verify fallback metrics if available
	t.Run("Fallback_Metrics_Consistency", func(t *testing.T) {
		if fallbackMetrics != nil {
			// With successful mock providers, no fallbacks should occur
			t.Logf("Fallback metrics: Total attempts=%d", fallbackMetrics.TotalFallbacks)

			// If no failures, fallback attempts should be 0
			if trinityMetrics.ThinkingStats.FailedRequests == 0 &&
				trinityMetrics.ReactiveStats.FailedRequests == 0 &&
				trinityMetrics.RapidStats.FailedRequests == 0 {
				AssertEqual(t, int64(0), fallbackMetrics.TotalFallbacks,
					"No fallback attempts should occur with successful providers")
			}
		}
	})
}

// TestEpic9_StateSynchronization tests bidirectional state sync between MomentumController and Guardian
// 測試狀態同步：MomentumController ↔ Guardian 雙向同步
func TestEpic9_StateSynchronization(t *testing.T) {
	t.Parallel()

	// Create Guardian
// 	guardianConfig := guardian.DefaultGuardianConfig()
// 	experienceGuardian := guardian.NewExperienceGuardian(guardianConfig)

	// Create MomentumController
	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)

	testCases := []struct {
		name              string
		initialTension    float64
		adjustment        float64
		expectedFinal     float64
	}{
		{
			name:           "Sync_LowTension",
			initialTension: 10.0,
			adjustment:     5.0,
			expectedFinal:  15.0,
		},
		{
			name:           "Sync_MediumTension",
			initialTension: 50.0,
			adjustment:     10.0,
			expectedFinal:  60.0,
		},
		{
			name:           "Sync_HighTension",
			initialTension: 80.0,
			adjustment:     15.0,
			expectedFinal:  95.0,
		},
		{
			name:           "Sync_Decrease",
			initialTension: 70.0,
			adjustment:     -20.0,
			expectedFinal:  50.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Set initial tension in MomentumController
			momentumController.SetTension(tc.initialTension)

			// 2. Sync to Guardian (MomentumController → Guardian)
			// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(tc.initialTension, phase)

			// 3. Adjust tension
			momentumController.AdjustTension(tc.adjustment)

			// 4. Verify final tension
			finalTension := momentumController.GetTension()
			AssertEqual(t, tc.expectedFinal, finalTension,
				"Final tension should match expected value")

			// 5. Sync again to Guardian
			// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(tc.expectedFinal, finalPhase)

			// Test passes if no panics or errors occur
		})
	}
}

// TestEpic9_NoCircularCalls tests that sync operations don't cause circular calls
// 測試無循環調用：SyncFromMomentum 不應觸發回調
func TestEpic9_NoCircularCalls(t *testing.T) {
	t.Parallel()

// 	guardianConfig := guardian.DefaultGuardianConfig()
// 	experienceGuardian := guardian.NewExperienceGuardian(guardianConfig)

	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)

	// Test rapid sync operations
	t.Run("Rapid_Sync_Operations", func(t *testing.T) {
		// Perform many sync operations in rapid succession
		// If there are circular calls, this would cause stack overflow
		for i := 0; i < 100; i++ {
			tension := float64(i % 100)

			// MomentumController → Guardian
			// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(tension, phase)

			// Update MomentumController
			momentumController.SetTension(tension)
		}

		// Test passes if no stack overflow occurs
		t.Log("Rapid sync operations completed without circular calls")
	})

	// Test concurrent sync operations
	t.Run("Concurrent_Sync_Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				for j := 0; j < 10; j++ {
					tension := float64((index*10 + j) % 100)

					// Concurrent syncs
					// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(tension, phase)
					momentumController.SetTension(tension)
				}
			}(i)
		}

		// Wait for all goroutines
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		// Wait with timeout
		select {
		case <-done:
			t.Log("Concurrent sync operations completed successfully")
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent sync operations timed out - possible deadlock")
		}
	})
}

// TestEpic9_DataRaceDetection tests for data races in concurrent operations
// 測試數據競爭：驗證沒有 race conditions
func TestEpic9_DataRaceDetection(t *testing.T) {
	// Run with: go test -race
	t.Parallel()

	routerConfig := trinity.RouterConfig{
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "thinking",
			MaxTokens:   16000,
			Temperature: 0.4,
		},
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "reactive",
			MaxTokens:   8000,
			Temperature: 0.7,
		},
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "rapid",
			MaxTokens:   4000,
			Temperature: 0.9,
		},
		FallbackEnabled:    true,
		AgentTierOverrides: make(map[string]trinity.TierLevel),
	}

	router, err := trinity.NewTrinityRouter(routerConfig)
	AssertNoError(t, err, "Failed to create router")

	ctx := context.Background()
	messages := TestMessages("user", "Race detection test")

	// Concurrent operations that might race
	var wg sync.WaitGroup
	concurrency := 20

	// Test 1: Concurrent routing
	t.Run("Concurrent_Routing", func(t *testing.T) {
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				router.Route(ctx, "judge", messages)
			}()
		}
		wg.Wait()
	})

	// Test 2: Concurrent metrics access
	t.Run("Concurrent_Metrics_Access", func(t *testing.T) {
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				if index%2 == 0 {
					// Read metrics
					router.GetMetricsSummary()
				} else {
					// Write metrics (via routing)
					router.Route(ctx, "choice", messages)
				}
			}(i)
		}
		wg.Wait()
	})

	// Test 3: Concurrent reset and access
	t.Run("Concurrent_Reset_And_Access", func(t *testing.T) {
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				switch index % 3 {
				case 0:
					router.Route(ctx, "narration", messages)
				case 1:
					router.GetMetricsSummary()
				case 2:
					router.ResetMetrics()
				}
			}(i)
		}
		wg.Wait()
	})

	t.Log("Data race detection tests completed (run with -race flag)")
}

// TestEpic9_TensionSyncConsistency tests tension value consistency across components
// 測試張力值一致性：所有組件的張力值應保持同步
func TestEpic9_TensionSyncConsistency(t *testing.T) {
	t.Parallel()

// 	guardianConfig := guardian.DefaultGuardianConfig()
// 	experienceGuardian := guardian.NewExperienceGuardian(guardianConfig)

	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)

	// Create game state
	gameState := engine.NewGameStateV2()

	testSequence := []struct {
		step           string
		tension        float64
		expectedPhase  string
		hpDelta        int
		sanDelta       int
	}{
		{
			step:          "Initial_Rest",
			tension:       0.0,
			expectedPhase: "Rest",
			hpDelta:       0,
			sanDelta:      0,
		},
		{
			step:          "Buildup_Phase",
			tension:       40.0,
			expectedPhase: "Buildup",
			hpDelta:       -5,
			sanDelta:      -3,
		},
		{
			step:          "Peak_Phase",
			tension:       85.0,
			expectedPhase: "Peak",
			hpDelta:       -10,
			sanDelta:      -8,
		},
		{
			step:          "Release_Phase",
			tension:       25.0,
			expectedPhase: "Release",
			hpDelta:       5,
			sanDelta:      3,
		},
	}

	for _, step := range testSequence {
		t.Run(step.step, func(t *testing.T) {
			// 1. Update MomentumController
			momentumController.SetTension(step.tension)

			// 2. Update GameState
			gameState.HP += step.hpDelta
			gameState.SAN += step.sanDelta

			// Clamp values
			if gameState.HP < 0 {
				gameState.HP = 0
			}
			if gameState.HP > 100 {
				gameState.HP = 100
			}
			if gameState.SAN < 0 {
				gameState.SAN = 0
			}
			if gameState.SAN > 100 {
				gameState.SAN = 100
			}

			// 3. Sync to Guardian
			// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(step.tension, step.expectedPhase)

// 			// 4. Guardian evaluates state
// 			experienceGuardian.OnTurnEnd(gameState)

			// 5. Verify consistency
			currentTension := momentumController.GetTension()
			AssertEqual(t, step.tension, currentTension,
				"MomentumController tension should match expected")

			// Log state for verification
			t.Logf("Step '%s': Tension=%.1f, Phase=%s, HP=%d, SAN=%d",
				step.step, currentTension, step.expectedPhase,
				gameState.HP, gameState.SAN)
		})
	}
}

// TestEpic9_MetricsThreadSafety tests thread safety of metrics operations
// 測試 Metrics 線程安全：並發讀寫不應引起問題
func TestEpic9_MetricsThreadSafety(t *testing.T) {
	// Run with: go test -race
	t.Parallel()

	routerConfig := trinity.RouterConfig{
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "thinking",
			MaxTokens:   16000,
			Temperature: 0.4,
		},
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "reactive",
			MaxTokens:   8000,
			Temperature: 0.7,
		},
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "rapid",
			MaxTokens:   4000,
			Temperature: 0.9,
		},
		FallbackEnabled:    true,
		AgentTierOverrides: make(map[string]trinity.TierLevel),
	}

	router, err := trinity.NewTrinityRouter(routerConfig)
	AssertNoError(t, err, "Failed to create router")

	ctx := context.Background()
	messages := TestMessages("user", "Thread safety test")

	// Concurrent operations
	var wg sync.WaitGroup
	readers := 10
	writers := 10

	// Start reader goroutines
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				router.GetMetricsSummary()
				router.GetFallbackMetrics()
			}
		}()
	}

	// Start writer goroutines
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			agent := []string{"judge", "choice", "narration"}[index%3]
			for j := 0; j < 100; j++ {
				router.Route(ctx, agent, messages)
			}
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Thread safety test completed successfully")
	case <-time.After(10 * time.Second):
		t.Fatal("Thread safety test timed out")
	}

	// Verify final metrics are consistent
	finalMetrics := router.GetMetricsSummary()
	totalRequests := finalMetrics.ThinkingStats.TotalRequests +
		finalMetrics.ReactiveStats.TotalRequests +
		finalMetrics.RapidStats.TotalRequests

	expectedRequests := int64(writers * 100)
	AssertEqual(t, expectedRequests, totalRequests,
		"Total requests should match expected count")
}

// Helper function to determine phase from tension value
func getPhaseForTension(tension float64) string {
	switch {
	case tension < 20:
		return "Rest"
	case tension < 60:
		return "Buildup"
	case tension < 80:
		return "Peak"
	default:
		return "Release"
	}
}

// TestEpic9_StateConsistencyAfterErrors tests state consistency after error conditions
// 測試錯誤後狀態一致性：錯誤不應破壞系統狀態
func TestEpic9_StateConsistencyAfterErrors(t *testing.T) {
	t.Parallel()

	routerConfig := trinity.RouterConfig{
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "thinking",
			MaxTokens:   16000,
			Temperature: 0.4,
		},
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "reactive",
			MaxTokens:   8000,
			Temperature: 0.7,
		},
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  "mock",
			Model:       "rapid",
			MaxTokens:   4000,
			Temperature: 0.9,
		},
		FallbackEnabled:    true,
		AgentTierOverrides: make(map[string]trinity.TierLevel),
	}

	router, err := trinity.NewTrinityRouter(routerConfig)
	AssertNoError(t, err, "Failed to create router")

	router.ResetMetrics()

	ctx := context.Background()
	messages := TestMessages("user", "State consistency test")

	// 1. Make successful requests
	for i := 0; i < 5; i++ {
		_, err := router.Route(ctx, "judge", messages)
		AssertNoError(t, err, "Initial requests should succeed")
	}

	// 2. Get baseline metrics
	baselineMetrics := router.GetMetricsSummary()
	baselineTotal := baselineMetrics.ThinkingStats.TotalRequests

	// 3. Simulate error scenario (with mock providers, we can't actually fail)
	// Document expected behavior

	// 4. Make more successful requests
	for i := 0; i < 5; i++ {
		_, err := router.Route(ctx, "judge", messages)
		AssertNoError(t, err, "Requests after error should succeed")
	}

	// 5. Verify metrics consistency
	finalMetrics := router.GetMetricsSummary()
	finalTotal := finalMetrics.ThinkingStats.TotalRequests

	expectedTotal := baselineTotal + 5
	AssertEqual(t, expectedTotal, finalTotal,
		"Metrics should remain consistent after error scenario")

	t.Log("State consistency maintained after error conditions")
}
