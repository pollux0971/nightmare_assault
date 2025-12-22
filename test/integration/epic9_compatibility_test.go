package integration

import (
	"context"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/nightmare-assault/nightmare-assault/internal/guardian"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

// TestEpic9_BackwardCompatibility tests that old code still works without Trinity
// 測試向後兼容性：不使用 Trinity 的舊代碼仍能運行
func TestEpic9_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("Orchestrator_Without_Trinity", func(t *testing.T) {
		// Create orchestrator without Trinity (using NewOrchestrator)
		orch := orchestrator.NewOrchestrator()

		// Verify it works
		AssertNotEqual(t, nil, orch, "Orchestrator should be created")
		AssertFalse(t, orch.HasTrinityRouter(), "Should not have Trinity router")

		// Old orchestrator should still function
		phase := orch.GetCurrentPhase()
		AssertEqual(t, orchestrator.PhaseGenesis, phase, "Should start in Genesis phase")

		t.Log("Orchestrator works without Trinity")
	})

	t.Run("MomentumController_Without_Guardian", func(t *testing.T) {
		// Create MomentumController without Guardian integration
		config := momentum.DefaultMomentumConfig()
		controller := momentum.NewMomentumController(config, nil)

		AssertNotEqual(t, nil, controller, "MomentumController should be created")

		// Should work without Guardian
		ctx := &momentum.NarrativeContext{
			CurrentBeat:              5,
			CurrentScene:             "test_scene",
			RiskLevel:                momentum.RiskMedium,
			IsPlotPoint:              false,
			NPCInitiatesConversation: false,
			PendingEvents:            make([]*momentum.GameEvent, 0),
			RecentChoices:            []string{"explore"},
			AutoResolvedBeats:        0,
		}

		shouldPause := controller.ShouldPauseForChoice(ctx)
		AssertTrue(t, shouldPause, "Should make pause decision without Guardian")

		t.Log("MomentumController works without Guardian")
	})

	t.Run("Guardian_Standalone", func(t *testing.T) {
		// Create Guardian without MomentumController integration
		config := guardian.DefaultGuardianConfig()
		guard := guardian.NewExperienceGuardian(config)

		AssertNotEqual(t, nil, guard, "Guardian should be created")

		// Should work independently
		AssertFalse(t, guard.IsProtectionActive(), "Protection should not be active initially")
		AssertEqual(t, 0, guard.GetConsecutiveDeaths(), "Should start with zero deaths")

		t.Log("Guardian works standalone")
	})
}

// TestEpic9_MixedMode tests coexistence of Trinity and non-Trinity code
// 測試混合模式：部分使用 Trinity，部分使用舊 provider
func TestEpic9_MixedMode(t *testing.T) {
	t.Parallel()

	t.Run("Trinity_And_Legacy_Coexist", func(t *testing.T) {
		// Create a Trinity router
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
		AssertNoError(t, err, "Failed to create Trinity router")

		// Create orchestrator with Trinity
		orchWithTrinity := orchestrator.NewOrchestratorWithTrinity(router)
		AssertTrue(t, orchWithTrinity.HasTrinityRouter(), "Should have Trinity")

		// Create orchestrator without Trinity
		orchWithoutTrinity := orchestrator.NewOrchestrator()
		AssertFalse(t, orchWithoutTrinity.HasTrinityRouter(), "Should not have Trinity")

		// Both should work
		AssertNotEqual(t, nil, orchWithTrinity, "Trinity orchestrator should work")
		AssertNotEqual(t, nil, orchWithoutTrinity, "Legacy orchestrator should work")

		t.Log("Trinity and legacy orchestrators can coexist")
	})

	t.Run("Partial_Trinity_Adoption", func(t *testing.T) {
		// Scenario: Some agents use Trinity, others use legacy providers
		// This would be implemented by having different orchestrator instances
		// or by selectively using Trinity routing

		// Create Trinity router
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
			FallbackEnabled: true,
			// Only certain agents use Trinity
			AgentTierOverrides: map[string]trinity.TierLevel{
				"judge": trinity.TierThinking, // Uses Trinity
				"npc":   trinity.TierThinking, // Uses Trinity
				// choice, narration, dream not specified - would use default or legacy
			},
		}

		router, err := trinity.NewTrinityRouter(routerConfig)
		AssertNoError(t, err, "Failed to create router")

		ctx := context.Background()
		messages := TestMessages("user", "Mixed mode test")

		// Route through Trinity for configured agents
		_, err = router.Route(ctx, "judge", messages)
		AssertNoError(t, err, "Trinity routing should work")

		_, err = router.Route(ctx, "npc", messages)
		AssertNoError(t, err, "Trinity routing should work")

		// Other agents would use default tier mapping
		_, err = router.Route(ctx, "choice", messages)
		AssertNoError(t, err, "Default routing should work")

		t.Log("Partial Trinity adoption works")
	})
}

// TestEpic9_MigrationPath tests the migration path from legacy to Trinity
// 測試遷移路徑：從舊系統逐步遷移到 Trinity
func TestEpic9_MigrationPath(t *testing.T) {
	t.Parallel()

	t.Run("Step1_Add_Trinity_Router", func(t *testing.T) {
		// Step 1: Add Trinity router to existing system
		// Create legacy orchestrator
		legacyOrch := orchestrator.NewOrchestrator()
		AssertFalse(t, legacyOrch.HasTrinityRouter(), "Legacy should not have Trinity")

		// Create Trinity router
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
		AssertNoError(t, err, "Should create Trinity router")

		// Upgrade to Trinity
		upgradedOrch := orchestrator.NewOrchestratorWithTrinity(router)
		AssertTrue(t, upgradedOrch.HasTrinityRouter(), "Upgraded should have Trinity")

		t.Log("Step 1: Trinity router added successfully")
	})

	t.Run("Step2_Configure_Agent_Tiers", func(t *testing.T) {
		// Step 2: Configure agent tier mappings
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
			FallbackEnabled: true,
			// Gradually migrate agents to appropriate tiers
			AgentTierOverrides: map[string]trinity.TierLevel{
				"judge":     trinity.TierThinking,
				"npc":       trinity.TierThinking,
				"choice":    trinity.TierReactive,
				"narration": trinity.TierReactive,
				"dream":     trinity.TierRapid,
			},
		}

		router, err := trinity.NewTrinityRouter(routerConfig)
		AssertNoError(t, err, "Should create configured router")

		// Verify configuration
		ctx := context.Background()
		messages := TestMessages("user", "Test")

		_, err = router.Route(ctx, "judge", messages)
		AssertNoError(t, err, "Judge should route to Thinking")

		_, err = router.Route(ctx, "dream", messages)
		AssertNoError(t, err, "Dream should route to Rapid")

		t.Log("Step 2: Agent tiers configured successfully")
	})

	t.Run("Step3_Enable_Fallback", func(t *testing.T) {
		// Step 3: Enable and test fallback mechanism
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
			FallbackEnabled:    true, // Enable fallback
			AgentTierOverrides: trinity.DefaultAgentTierMapping,
			RetryConfig:        client.DefaultRetryConfig(),
		}

		router, err := trinity.NewTrinityRouter(routerConfig)
		AssertNoError(t, err, "Should create router with fallback")

		// Verify fallback is enabled
		fallbackMetrics := router.GetFallbackMetrics()
		AssertNotEqual(t, nil, fallbackMetrics, "Fallback metrics should exist")

		t.Log("Step 3: Fallback enabled successfully")
	})

	t.Run("Step4_Monitor_Performance", func(t *testing.T) {
		// Step 4: Monitor performance with metrics
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
			AgentTierOverrides: trinity.DefaultAgentTierMapping,
		}

		router, err := trinity.NewTrinityRouter(routerConfig)
		AssertNoError(t, err, "Should create router")

		// Make some requests
		ctx := context.Background()
		messages := TestMessages("user", "Performance test")

		for i := 0; i < 10; i++ {
			router.Route(ctx, "judge", messages)
		}

		// Check metrics
		metrics := router.GetMetricsSummary()
		AssertTrue(t, metrics.ThinkingStats.TotalRequests > 0, "Should track requests")
		AssertTrue(t, metrics.ThinkingStats.SuccessRate > 0, "Should track success rate")

		t.Log("Step 4: Performance monitoring working")
		t.Logf("Metrics: Requests=%d, Success=%.2f%%",
			metrics.ThinkingStats.TotalRequests,
			metrics.ThinkingStats.SuccessRate*100)
	})
}

// TestEpic9_LegacyAPICompatibility tests that legacy API methods still work
// 測試 Legacy API 兼容性：舊的 API 方法仍然可用
func TestEpic9_LegacyAPICompatibility(t *testing.T) {
	t.Parallel()

	t.Run("NewOrchestrator_StillWorks", func(t *testing.T) {
		// Old constructor should still work
		orch := orchestrator.NewOrchestrator()
		AssertNotEqual(t, nil, orch, "NewOrchestrator should still work")

		// Old API should be available
		phase := orch.GetCurrentPhase()
		AssertEqual(t, orchestrator.PhaseGenesis, phase, "Should have correct phase")

		t.Log("Legacy NewOrchestrator() works")
	})

	t.Run("NewOrchestratorWithProvider_StillWorks", func(t *testing.T) {
		// Create a mock provider
		provider := NewMockProvider("Test response")

		// Old constructor with single provider should still work
		// Note: This creates a basic Trinity router internally
		orch, err := orchestrator.NewOrchestratorWithProvider(provider, "test-key")
		AssertNoError(t, err, "NewOrchestratorWithProvider should work")
		AssertNotEqual(t, nil, orch, "Orchestrator should be created")

		// Should have Trinity router (backward compatibility mode)
		AssertTrue(t, orch.HasTrinityRouter(), "Should have Trinity router in compatibility mode")

		t.Log("Legacy NewOrchestratorWithProvider() works")
	})

	t.Run("DefaultConfigs_StillWork", func(t *testing.T) {
		// Default configs should still be available
		momentumConfig := momentum.DefaultMomentumConfig()
		AssertNotEqual(t, nil, momentumConfig, "DefaultMomentumConfig should work")

		guardianConfig := guardian.DefaultGuardianConfig()
		AssertNotEqual(t, nil, guardianConfig, "DefaultGuardianConfig should work")

		// Create components with defaults
		controller := momentum.NewMomentumController(momentumConfig, nil)
		AssertNotEqual(t, nil, controller, "Should create with default config")

		guard := guardian.NewExperienceGuardian(guardianConfig)
		AssertNotEqual(t, nil, guard, "Should create with default config")

		t.Log("Legacy default configs work")
	})
}

// TestEpic9_FeatureFlagCompatibility tests gradual feature rollout
// 測試功能標誌兼容性：支持功能的逐步推出
func TestEpic9_FeatureFlagCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("FallbackEnabled_Flag", func(t *testing.T) {
		// Test with fallback enabled
		configEnabled := trinity.RouterConfig{
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

		routerEnabled, err := trinity.NewTrinityRouter(configEnabled)
		AssertNoError(t, err, "Router with fallback should work")
		AssertNotEqual(t, nil, routerEnabled.GetFallbackMetrics(), "Should have fallback metrics")

		// Test with fallback disabled
		configDisabled := configEnabled
		configDisabled.FallbackEnabled = false

		_, err = trinity.NewTrinityRouter(configDisabled)
		AssertNoError(t, err, "Router without fallback should work")

		t.Log("Fallback can be enabled/disabled via config")
	})

	t.Run("AgentTierOverrides_Optional", func(t *testing.T) {
		// Test with no overrides (uses defaults)
		configNoOverrides := trinity.RouterConfig{
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
			AgentTierOverrides: nil, // No overrides
		}

		routerDefaults, err := trinity.NewTrinityRouter(configNoOverrides)
		AssertNoError(t, err, "Router with default tiers should work")

		// Test with custom overrides
		configWithOverrides := configNoOverrides
		configWithOverrides.AgentTierOverrides = map[string]trinity.TierLevel{
			"choice": trinity.TierThinking, // Override default
		}

		routerCustom, err := trinity.NewTrinityRouter(configWithOverrides)
		AssertNoError(t, err, "Router with custom tiers should work")

		// Both should work
		AssertNotEqual(t, nil, routerDefaults, "Default router should work")
		AssertNotEqual(t, nil, routerCustom, "Custom router should work")

		t.Log("Agent tier overrides are optional")
	})
}

// TestEpic9_GradualAdoption tests gradual adoption of Trinity features
// 測試逐步採用：可以逐步啟用 Trinity 功能
func TestEpic9_GradualAdoption(t *testing.T) {
	t.Parallel()

	phases := []struct {
		name        string
		description string
		test        func(t *testing.T)
	}{
		{
			name:        "Phase_0_No_Trinity",
			description: "System works without Trinity",
			test: func(t *testing.T) {
				orch := orchestrator.NewOrchestrator()
				AssertNotEqual(t, nil, orch, "Should work without Trinity")
				AssertFalse(t, orch.HasTrinityRouter(), "Should not have Trinity")
			},
		},
		{
			name:        "Phase_1_Add_Trinity_Router",
			description: "Add Trinity router with minimal config",
			test: func(t *testing.T) {
				config := trinity.RouterConfig{
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
					FallbackEnabled:    false, // Start without fallback
					AgentTierOverrides: nil,   // Use defaults
				}

				router, err := trinity.NewTrinityRouter(config)
				AssertNoError(t, err, "Should create basic router")

				orch := orchestrator.NewOrchestratorWithTrinity(router)
				AssertTrue(t, orch.HasTrinityRouter(), "Should have Trinity")
			},
		},
		{
			name:        "Phase_2_Enable_Fallback",
			description: "Enable fallback for reliability",
			test: func(t *testing.T) {
				config := trinity.RouterConfig{
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
					FallbackEnabled:    true, // Enable fallback
					AgentTierOverrides: nil,
				}

				router, err := trinity.NewTrinityRouter(config)
				AssertNoError(t, err, "Should create router with fallback")
				AssertNotEqual(t, nil, router.GetFallbackMetrics(), "Should have fallback")
			},
		},
		{
			name:        "Phase_3_Optimize_Tiers",
			description: "Optimize agent tier mappings",
			test: func(t *testing.T) {
				config := trinity.RouterConfig{
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
					FallbackEnabled: true,
					// Optimized tier mapping
					AgentTierOverrides: map[string]trinity.TierLevel{
						"judge":     trinity.TierThinking,
						"npc":       trinity.TierThinking,
						"choice":    trinity.TierReactive,
						"narration": trinity.TierReactive,
						"dream":     trinity.TierRapid,
					},
				}

				router, err := trinity.NewTrinityRouter(config)
				AssertNoError(t, err, "Should create optimized router")
				AssertNotEqual(t, nil, router, "Router should be ready")
			},
		},
	}

	for _, phase := range phases {
		t.Run(phase.name, func(t *testing.T) {
			t.Logf("Testing: %s", phase.description)
			phase.test(t)
		})
	}

	t.Log("Gradual adoption path validated")
}
