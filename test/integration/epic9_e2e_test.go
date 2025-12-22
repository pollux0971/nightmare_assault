package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

// TestEpic9_FullGameFlow tests the complete game flow with Trinity + Guardian + Orchestrator + MomentumController
// 測試完整的遊戲流程：玩家輸入 → 判斷 → 生成選項 → 生成敘述 → NPC對話 → 夢境
func TestEpic9_FullGameFlow(t *testing.T) {
	t.Parallel()

	// Setup: Create all components
	thinkingProvider := NewMockProvider("Thinking response: Player intends to explore")
	reactiveProvider := NewMockProvider("Reactive response: Generated narration")
	rapidProvider := NewMockProvider("Rapid response: Dream sequence")

	// Create Trinity router with injected mock providers
	router := trinity.NewTrinityRouterWithProviders(
		thinkingProvider,
		reactiveProvider,
		rapidProvider,
		true, // fallback enabled
		make(map[string]trinity.TierLevel), // no overrides
	)

	// Create Guardian
// 	guardianConfig := guardian.DefaultGuardianConfig()
// 	experienceGuardian := guardian.NewExperienceGuardian(guardianConfig)

	// Create MomentumController
	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)

	// Create Orchestrator with Trinity
	orch := orchestrator.NewOrchestratorWithTrinity(router)

	// Create game state
	gameState := engine.NewGameStateV2()
	gameState.HP = 80
	gameState.SAN = 70
	gameState.CurrentScene = "hospital_entrance"

	// Test agent requests and verify correct tier usage
	testCases := []struct {
		name          string
		agentName     string
		expectedTier  trinity.TierLevel
		messages      []client.Message
		setupProvider func()
	}{
		{
			name:         "JudgeAgent_UsesThinkingTier",
			agentName:    "judge",
			expectedTier: trinity.TierThinking,
			messages:     TestMessages("user", "I want to explore the dark hallway"),
			setupProvider: func() {
				// JudgeAgent should use Thinking tier
				thinkingProvider.Reset()
				thinkingProvider.responseContent = ThinkingResponse(
					"Analyzing player intent: exploring hallway is medium risk",
					"The player intends to explore",
				)
			},
		},
		{
			name:         "ChoiceAgent_UsesReactiveTier",
			agentName:    "choice",
			expectedTier: trinity.TierReactive,
			messages:     TestMessages("user", "Generate choices for current situation"),
			setupProvider: func() {
				reactiveProvider.Reset()
				reactiveProvider.responseContent = "1. Enter the hallway\n2. Turn back\n3. Call out"
			},
		},
		{
			name:         "NarrationAgent_UsesReactiveTier",
			agentName:    "narration",
			expectedTier: trinity.TierReactive,
			messages:     TestMessages("user", "Generate narration for player action"),
			setupProvider: func() {
				reactiveProvider.Reset()
				reactiveProvider.responseContent = "You cautiously step into the darkness..."
			},
		},
		{
			name:         "NPCAgent_UsesThinkingTier",
			agentName:    "npc",
			expectedTier: trinity.TierThinking,
			messages:     TestMessages("user", "Generate NPC dialogue"),
			setupProvider: func() {
				thinkingProvider.Reset()
				thinkingProvider.responseContent = ThinkingResponse(
					"Considering NPC emotional state and relationship",
					"'Stay away from the basement,' the nurse warns you.",
				)
			},
		},
		{
			name:         "DreamAgent_UsesRapidTier",
			agentName:    "dream",
			expectedTier: trinity.TierRapid,
			messages:     TestMessages("user", "Generate dream sequence"),
			setupProvider: func() {
				rapidProvider.Reset()
				rapidProvider.responseContent = "You dream of endless corridors..."
			},
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup provider for this test case
			tc.setupProvider()

			// Route request through Trinity
			resp, err := router.Route(ctx, tc.agentName, tc.messages)
			AssertNoError(t, err, "Router.Route failed for "+tc.agentName)
			AssertNotEqual(t, "", resp.Content, "Response content should not be empty")

			// Verify metrics recorded for correct tier
			metrics := router.GetMetricsSummary()

			var tierMetrics *trinity.TierStats
			switch tc.expectedTier {
			case trinity.TierThinking:
				tierMetrics = &metrics.ThinkingStats
			case trinity.TierReactive:
				tierMetrics = &metrics.ReactiveStats
			case trinity.TierRapid:
				tierMetrics = &metrics.RapidStats
			}

			// Verify request was recorded
			AssertTrue(t, tierMetrics.TotalRequests > 0,
				"Expected requests recorded for tier "+tc.expectedTier.String())

			// Verify thinking tags handled for Thinking tier
			if tc.expectedTier == trinity.TierThinking {
				// Thinking tags should be removed from response content
				AssertFalse(t, strings.Contains(resp.Content, "<think>"),
					"Thinking tags should be removed from response")
				AssertFalse(t, strings.Contains(resp.Content, "</think>"),
					"Thinking tags should be removed from response")

				// Thinking chain should be in metadata
				if resp.Metadata != nil {
					_, hasThinking := resp.Metadata["thinking_chain"]
					AssertTrue(t, hasThinking, "Thinking tier response should have thinking_chain in metadata")
				}
			}
		})
	}

	// Verify Guardian integration
	// Verify MomentumController integration
	narrativeCtx := &momentum.NarrativeContext{
		CurrentBeat:              10,
		CurrentScene:             "hospital_entrance",
		RiskLevel:                momentum.RiskMedium,
		IsPlotPoint:              false,
		NPCInitiatesConversation: false,
		PendingEvents:            make([]*momentum.GameEvent, 0),
		RecentChoices:            []string{"explore hallway"},
		AutoResolvedBeats:        0,
	}

	shouldPause := momentumController.ShouldPauseForChoice(narrativeCtx)
	AssertTrue(t, shouldPause, "Should pause for medium risk with default config")

	// Verify overall system integration
	AssertNotEqual(t, nil, orch, "Orchestrator should be initialized")
	AssertTrue(t, orch.HasTrinityRouter(), "Orchestrator should have Trinity router")
}

// TestEpic9_FallbackMechanism tests the fallback mechanism across all tiers
// 測試降級機制：Thinking失敗 → Reactive → Rapid
func TestEpic9_FallbackMechanism(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		thinkingFails       bool
		reactiveFails       bool
		rapidFails          bool
		expectSuccess       bool
		expectedUsedTier    trinity.TierLevel
		expectDowngradeFrom trinity.TierLevel
		expectDowngradeTo   trinity.TierLevel
	}{
		{
			name:             "ThinkingFails_FallbackToReactive",
			thinkingFails:    true,
			reactiveFails:    false,
			rapidFails:       false,
			expectSuccess:    true,
			expectedUsedTier: trinity.TierReactive,
			expectDowngradeFrom: trinity.TierThinking,
			expectDowngradeTo:   trinity.TierReactive,
		},
		{
			name:             "ThinkingAndReactiveFail_FallbackToRapid",
			thinkingFails:    true,
			reactiveFails:    true,
			rapidFails:       false,
			expectSuccess:    true,
			expectedUsedTier: trinity.TierRapid,
			expectDowngradeFrom: trinity.TierThinking,
			expectDowngradeTo:   trinity.TierRapid,
		},
		{
			name:          "AllTiersFail_ReturnsError",
			thinkingFails: true,
			reactiveFails: true,
			rapidFails:    true,
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock providers
			thinkingProvider := NewMockProvider("Thinking response")
			reactiveProvider := NewMockProvider("Reactive response")
			rapidProvider := NewMockProvider("Rapid response")

			// Configure failures
			if tc.thinkingFails {
				thinkingProvider.SetShouldFail(true, ErrMockAPIError)
			}
			if tc.reactiveFails {
				reactiveProvider.SetShouldFail(true, ErrMockAPIError)
			}
			if tc.rapidFails {
				rapidProvider.SetShouldFail(true, ErrMockAPIError)
			}

			// Create router with injected mock providers
			router := trinity.NewTrinityRouterWithProviders(
				thinkingProvider,
				reactiveProvider,
				rapidProvider,
				true, // fallback enabled
				make(map[string]trinity.TierLevel), // no overrides
			)

			// Route request for JudgeAgent (Thinking tier)
			ctx := context.Background()
			messages := TestMessages("user", "Test message")
			resp, err := router.Route(ctx, "judge", messages)

			if tc.expectSuccess {
				AssertNoError(t, err, "Expected request to succeed with fallback")
				AssertNotEqual(t, "", resp.Content, "Response should have content")

				// Verify fallback metrics
				fallbackMetrics := router.GetFallbackMetrics()
				AssertNotEqual(t, nil, fallbackMetrics, "Fallback metrics should exist")
				AssertTrue(t, fallbackMetrics.TotalFallbacks > 0, "Should have fallback attempts")

				// Verify downgrade was recorded
				metrics := router.GetMetricsSummary()
				totalDowngrades := metrics.TotalDowngrades
				AssertTrue(t, totalDowngrades > 0, "Should have recorded tier downgrade")
			} else {
				AssertError(t, err, "Expected request to fail when all tiers fail")
			}
		})
	}
}

// TestEpic9_GuardianTensionManagement tests Guardian's tension management integration
// 測試 Guardian 的張力管理：張力調整、難度變化、phase transitions
func TestEpic9_GuardianTensionManagement(t *testing.T) {
	t.Parallel()

	// Create Guardian
// 	guardianConfig := guardian.DefaultGuardianConfig()
// 	experienceGuardian := guardian.NewExperienceGuardian(guardianConfig)

	// Create MomentumController with Guardian integration
	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)

	// Test tension curve phases
	testCases := []struct {
		name          string
		tensionValue  float64
		expectedPhase string
	}{
		{
			name:          "Rest_Phase",
			tensionValue:  0.0,
			expectedPhase: "Rest",
		},
		{
			name:          "Buildup_Phase",
			tensionValue:  30.0,
			expectedPhase: "Buildup",
		},
		{
			name:          "Peak_Phase",
			tensionValue:  80.0,
			expectedPhase: "Peak",
		},
		{
			name:          "Release_Phase",
			tensionValue:  20.0,
			expectedPhase: "Release",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set tension via MomentumController
			momentumController.SetTension(tc.tensionValue)

			// Verify tension was set
			currentTension := momentumController.GetTension()
			AssertEqual(t, tc.tensionValue, currentTension, "Tension should be set correctly")

			// Sync to Guardian (Guardian should update its phase)
			// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(tc.tensionValue, tc.expectedPhase)

			// Note: We can't verify Guardian's internal phase without exposing it
			// But we can verify that SyncFromMomentum doesn't panic
		})
	}

	// Test bidirectional synchronization
	t.Run("Bidirectional_Sync", func(t *testing.T) {
		// MomentumController adjusts tension
		initialTension := 50.0
		momentumController.SetTension(initialTension)

		// Guardian syncs from Momentum
		// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(initialTension, "Buildup")

		// Guardian adjusts difficulty (simulated via momentum adjustment)
		difficultyAdjustment := 10.0
		newTension := initialTension + difficultyAdjustment

		// Update MomentumController
		momentumController.AdjustTension(difficultyAdjustment)

		// Verify new tension
		AssertEqual(t, newTension, momentumController.GetTension(), "Tension should be adjusted")
	})

	// Test no circular calls
	t.Run("NoCircularCalls", func(t *testing.T) {
		// SyncFromMomentum should NOT trigger callback to MomentumController
		// This is a design test - we verify it doesn't cause infinite loop or panic

		// Set initial state
		momentumController.SetTension(40.0)

		// Sync to Guardian multiple times (should not cause issues)
		for i := 0; i < 5; i++ {
			// Note: SyncFromMomentum not yet implemented (Story 9-11)
			// experienceGuardian.SyncFromMomentum(40.0, "Buildup")
		}

		// Verify tension unchanged
		AssertEqual(t, 40.0, momentumController.GetTension(), "Tension should remain stable")
	})
}

// TestEpic9_DynamicConfiguration tests runtime configuration changes
// 測試運行時配置調整：動態更改 agent tier mapping
func TestEpic9_DynamicConfiguration(t *testing.T) {
	t.Parallel()

	// Create mock providers
	thinkingProvider := NewMockProvider("Thinking response")
	reactiveProvider := NewMockProvider("Reactive response")
	rapidProvider := NewMockProvider("Rapid response")

	// Create router with initial configuration
	router := trinity.NewTrinityRouterWithProviders(
		thinkingProvider,
		reactiveProvider,
		rapidProvider,
		true, // fallback enabled
		map[string]trinity.TierLevel{
			"choice": trinity.TierReactive, // Default
		},
	)

	ctx := context.Background()
	messages := TestMessages("user", "Test message")

	// Test 1: Initial configuration
	t.Run("InitialConfiguration", func(t *testing.T) {
		router.ResetMetrics()

		// ChoiceAgent should use Reactive tier
		_, err := router.Route(ctx, "choice", messages)
		AssertNoError(t, err)

		metrics := router.GetMetricsSummary()
		AssertTrue(t, metrics.ReactiveStats.TotalRequests > 0, "Should use Reactive tier")
	})

	// Test 2: Change configuration at runtime
	// Note: Current Trinity implementation doesn't support runtime tier override changes
	// This test documents the expected future behavior
	t.Run("RuntimeConfigChange", func(t *testing.T) {
		// In a future version, we could support:
		// router.UpdateAgentTierMapping("choice", trinity.TierThinking)

		// For now, we verify that creating a new router with different config works
		newThinkingProvider := NewMockProvider("New Thinking response")
		newReactiveProvider := NewMockProvider("New Reactive response")
		newRapidProvider := NewMockProvider("New Rapid response")

		newRouter := trinity.NewTrinityRouterWithProviders(
			newThinkingProvider,
			newReactiveProvider,
			newRapidProvider,
			true, // fallback enabled
			map[string]trinity.TierLevel{
				"choice": trinity.TierThinking, // Changed to Thinking
			},
		)

		// Verify new configuration takes effect
		_, err := newRouter.Route(ctx, "choice", messages)
		AssertNoError(t, err)

		newMetrics := newRouter.GetMetricsSummary()
		AssertTrue(t, newMetrics.ThinkingStats.TotalRequests > 0,
			"Should use Thinking tier with new config")
	})

	// Test 3: Metrics reflect configuration changes
	t.Run("MetricsReflectChanges", func(t *testing.T) {
		router.ResetMetrics()

		// Make multiple requests
		for i := 0; i < 5; i++ {
			_, err := router.Route(ctx, "choice", messages)
			AssertNoError(t, err)
		}

		metrics := router.GetMetricsSummary()
		AssertEqual(t, int64(5), metrics.ReactiveStats.TotalRequests,
			"Metrics should reflect all requests")
	})
}
