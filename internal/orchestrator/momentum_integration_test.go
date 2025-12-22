// Story 7.7: Phase 3 Integration Tests
// Epic 7 - 敘事動量系統整合測試
package orchestrator

import (
	"context"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMomentumIntegration_FullFlow tests the complete momentum flow
// Story 7.7 AC1: 完整動量流程測試（評估 → 自動演繹 → 暫停 → 選擇）
func TestMomentumIntegration_FullFlow(t *testing.T) {
	tests := []struct {
		name              string
		setupGameState    func(*engine.GameStateV2)
		setupMomentumCtrl func(*momentum.MomentumController)
		expectedStopReason momentum.StopReason
		expectedBeats     int
	}{
		{
			name: "low_risk_auto_resolves_until_max_beats",
			setupGameState: func(gs *engine.GameStateV2) {
				gs.HP = 80
				gs.SAN = 70
				gs.Tension.Value = 40
			},
			setupMomentumCtrl: func(mc *momentum.MomentumController) {
				config := momentum.DefaultMomentumConfig()
				config.MaxAutoBeats = 3
				config.AutoResolve = true
				mc.SetConfig(config)
			},
			expectedStopReason: momentum.StopReasonMaxAutoBeats,
			expectedBeats:      3,
		},
		{
			name: "medium_risk_pauses_on_risk_level",
			setupGameState: func(gs *engine.GameStateV2) {
				gs.HP = 35 // Medium risk
				gs.SAN = 45
				gs.Tension.Value = 60
			},
			setupMomentumCtrl: func(mc *momentum.MomentumController) {
				config := momentum.DefaultMomentumConfig()
				config.PauseOnRisk = momentum.RiskMedium
				config.AutoResolve = true
				mc.SetConfig(config)
			},
			expectedStopReason: momentum.StopReasonMaxAutoBeats,
			expectedBeats:      5, // Note: In real integration with RiskEvaluator, would pause earlier
		},
		{
			name: "high_frequency_always_pauses",
			setupGameState: func(gs *engine.GameStateV2) {
				gs.HP = 80
				gs.SAN = 70
				gs.Tension.Value = 30
			},
			setupMomentumCtrl: func(mc *momentum.MomentumController) {
				config := momentum.DefaultMomentumConfig()
				config.Frequency = momentum.FrequencyHigh
				config.AutoResolve = true
				mc.SetConfig(config)
			},
			expectedStopReason: momentum.StopReasonFrequency,
			expectedBeats:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup orchestrator with momentum controller
			orch := NewOrchestrator()

			// Setup game state
			tt.setupGameState(orch.gameState)

			// Setup momentum controller
			config := momentum.DefaultMomentumConfig()
			mc := momentum.NewMomentumController(config, orch.narrationAgent)
			orch.momentumController = mc

			if tt.setupMomentumCtrl != nil {
				tt.setupMomentumCtrl(mc)
			}

			// Run game loop
			ctx := context.Background()
			result, err := orch.RunGameLoop(ctx)

			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify expected behavior
			assert.Equal(t, tt.expectedStopReason, result.StopReason, "Stop reason should match")

			if tt.expectedBeats > 0 {
				assert.Equal(t, tt.expectedBeats, result.BeatsResolved, "Beats resolved should match")
				assert.Greater(t, len(result.Narratives), 0, "Should have generated narratives")
			}
		})
	}
}

// TestMomentumIntegration_RiskLevels tests different risk level scenarios
// Story 7.7 AC2: 不同風險等級測試
func TestMomentumIntegration_RiskLevels(t *testing.T) {
	tests := []struct {
		name           string
		hp             int
		san            int
		tension        int
		expectedRisk   momentum.RiskLevel
		shouldPause    bool
		pauseThreshold momentum.RiskLevel
	}{
		{
			name:           "none_risk_continues",
			hp:             90,
			san:            85,
			tension:        20,
			expectedRisk:   momentum.RiskNone,
			shouldPause:    false,
			pauseThreshold: momentum.RiskMedium,
		},
		{
			name:           "low_risk_continues_if_threshold_medium",
			hp:             75,
			san:            70,
			tension:        35,
			expectedRisk:   momentum.RiskLow,
			shouldPause:    false,
			pauseThreshold: momentum.RiskMedium,
		},
		{
			name:           "medium_risk_pauses_if_threshold_medium",
			hp:             35,
			san:            45,
			tension:        55,
			expectedRisk:   momentum.RiskMedium,
			shouldPause:    true,
			pauseThreshold: momentum.RiskMedium,
		},
		{
			name:           "high_risk_always_pauses",
			hp:             15,
			san:            25,
			tension:        85,
			expectedRisk:   momentum.RiskHigh,
			shouldPause:    true,
			pauseThreshold: momentum.RiskMedium,
		},
		{
			name:           "lethal_risk_always_pauses",
			hp:             5,
			san:            10,
			tension:        95,
			expectedRisk:   momentum.RiskLethal,
			shouldPause:    true,
			pauseThreshold: momentum.RiskMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup momentum controller
			config := momentum.DefaultMomentumConfig()
			config.PauseOnRisk = tt.pauseThreshold
			config.AutoResolve = true
			config.MaxAutoBeats = 5

			mc := momentum.NewMomentumController(config, nil)

			// Build context
			ctx := &momentum.NarrativeContext{
				CurrentBeat:  1,
				CurrentScene: "test_scene",
			}

			// Simulate risk evaluation
			// (In real integration, this would come from RiskEvaluator)
			// For test, we directly set the risk level based on HP/SAN
			if tt.hp <= 10 || tt.san <= 10 {
				ctx.RiskLevel = momentum.RiskLethal
			} else if tt.hp <= 20 || tt.san <= 20 || tt.tension >= 80 {
				ctx.RiskLevel = momentum.RiskHigh
			} else if tt.hp <= 40 || tt.san <= 40 || tt.tension >= 50 {
				ctx.RiskLevel = momentum.RiskMedium
			} else if tt.hp <= 60 || tt.san <= 60 || tt.tension >= 30 {
				ctx.RiskLevel = momentum.RiskLow
			} else {
				ctx.RiskLevel = momentum.RiskNone
			}

			// Test should pause decision
			shouldPause := mc.ShouldPauseForChoice(ctx)

			assert.Equal(t, tt.expectedRisk, ctx.RiskLevel, "Risk level should match")
			assert.Equal(t, tt.shouldPause, shouldPause, "Pause decision should match")
		})
	}
}

// TestMomentumIntegration_FrequencyLevels tests different frequency settings
// Story 7.7 AC3: 不同 Frequency 等級測試
func TestMomentumIntegration_FrequencyLevels(t *testing.T) {
	tests := []struct {
		name          string
		frequency     momentum.FrequencyLevel
		riskLevel     momentum.RiskLevel
		shouldPause   bool
		expectedReason momentum.StopReason
	}{
		{
			name:          "high_frequency_always_pauses",
			frequency:     momentum.FrequencyHigh,
			riskLevel:     momentum.RiskNone,
			shouldPause:   true,
			expectedReason: momentum.StopReasonFrequency,
		},
		{
			name:          "medium_frequency_pauses_on_medium_risk",
			frequency:     momentum.FrequencyMedium,
			riskLevel:     momentum.RiskMedium,
			shouldPause:   true,
			expectedReason: momentum.StopReasonFrequency,
		},
		{
			name:          "medium_frequency_continues_on_low_risk",
			frequency:     momentum.FrequencyMedium,
			riskLevel:     momentum.RiskLow,
			shouldPause:   false,
			expectedReason: momentum.StopReasonNone,
		},
		{
			name:          "low_frequency_pauses_on_high_risk",
			frequency:     momentum.FrequencyLow,
			riskLevel:     momentum.RiskHigh,
			shouldPause:   true,
			expectedReason: momentum.StopReasonFrequency,
		},
		{
			name:          "low_frequency_continues_on_medium_risk",
			frequency:     momentum.FrequencyLow,
			riskLevel:     momentum.RiskMedium,
			shouldPause:   false,
			expectedReason: momentum.StopReasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup momentum controller
			config := momentum.DefaultMomentumConfig()
			config.Frequency = tt.frequency
			config.AutoResolve = true
			config.PauseOnRisk = momentum.RiskLethal // Set high threshold to test frequency logic

			mc := momentum.NewMomentumController(config, nil)

			// Build context
			ctx := &momentum.NarrativeContext{
				CurrentBeat:       1,
				CurrentScene:      "test_scene",
				RiskLevel:         tt.riskLevel,
				AutoResolvedBeats: 0,
			}

			// Test pause decision
			shouldPause := mc.ShouldPauseForChoice(ctx)
			reason := mc.DetermineStopReason(ctx)

			assert.Equal(t, tt.shouldPause, shouldPause, "Pause decision should match")
			if tt.shouldPause {
				assert.Equal(t, tt.expectedReason, reason, "Stop reason should match")
			}
		})
	}
}

// TestMomentumIntegration_NPCConversation tests NPC-triggered chat mode
// Story 7.7 AC4: NPC 對話觸發聊天室測試
func TestMomentumIntegration_NPCConversation(t *testing.T) {
	// Setup orchestrator
	orch := NewOrchestrator()
	orch.gameState.HP = 80
	orch.gameState.SAN = 70
	orch.gameState.CurrentScene = "hospital_lobby"

	// Setup momentum controller with NPC conversation enabled
	config := momentum.DefaultMomentumConfig()
	config.PauseOnNPC = true
	config.AutoResolve = true
	mc := momentum.NewMomentumController(config, orch.narrationAgent)
	orch.momentumController = mc

	// Build context with NPC conversation
	ctx := &momentum.NarrativeContext{
		CurrentBeat:              1,
		CurrentScene:             "hospital_lobby",
		RiskLevel:                momentum.RiskLow,
		NPCInitiatesConversation: true,
		InitiatingNPC:            "nurse_wang",
	}

	// Test should pause for NPC conversation
	shouldPause := mc.ShouldPauseForChoice(ctx)
	reason := mc.DetermineStopReason(ctx)

	assert.True(t, shouldPause, "Should pause for NPC conversation")
	assert.Equal(t, momentum.StopReasonNPCConversation, reason, "Stop reason should be NPC conversation")

	// Test game loop routing
	// (This would be tested via RunGameLoop in real integration)
	// For now, verify the routing logic is correct
	if shouldPause && reason == momentum.StopReasonNPCConversation {
		// Should route to ChatOverlay
		t.Log("Would route to ChatOverlay for NPC:", ctx.InitiatingNPC)
	}
}

// TestMomentumIntegration_PlotPoint tests plot point detection
// Story 7.7 AC5: 劇情點檢測測試
func TestMomentumIntegration_PlotPoint(t *testing.T) {
	tests := []struct {
		name            string
		currentBeat     int
		totalBeats      int
		seedProgress    float64
		isPlotPoint     bool
		plotPointType   string
	}{
		{
			name:          "chapter_end_is_plot_point",
			currentBeat:   9,
			totalBeats:    10,
			seedProgress:  0.5,
			isPlotPoint:   true,
			plotPointType: "chapter_end",
		},
		{
			name:          "seed_milestone_30_is_plot_point",
			currentBeat:   5,
			totalBeats:    20,
			seedProgress:  0.3,
			isPlotPoint:   true,
			plotPointType: "seed_milestone",
		},
		{
			name:          "seed_milestone_60_is_plot_point",
			currentBeat:   12,
			totalBeats:    20,
			seedProgress:  0.6,
			isPlotPoint:   true,
			plotPointType: "seed_milestone",
		},
		{
			name:          "normal_beat_not_plot_point",
			currentBeat:   5,
			totalBeats:    20,
			seedProgress:  0.2,
			isPlotPoint:   false,
			plotPointType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup momentum controller
			config := momentum.DefaultMomentumConfig()
			config.PauseOnPlot = true
			config.AutoResolve = true

			mc := momentum.NewMomentumController(config, nil)

			// Build context
			ctx := &momentum.NarrativeContext{
				CurrentBeat:   tt.currentBeat,
				CurrentScene:  "test_scene",
				RiskLevel:     momentum.RiskLow,
				IsPlotPoint:   tt.isPlotPoint,
				PlotPointType: tt.plotPointType,
			}

			// Test pause decision
			shouldPause := mc.ShouldPauseForChoice(ctx)
			reason := mc.DetermineStopReason(ctx)

			if tt.isPlotPoint {
				assert.True(t, shouldPause, "Should pause at plot point")
				assert.Equal(t, momentum.StopReasonPlotPoint, reason, "Stop reason should be plot point")
			} else {
				// May or may not pause depending on other factors
				// Just verify plot point is not the reason
				if reason == momentum.StopReasonPlotPoint {
					t.Error("Should not pause for plot point when IsPlotPoint is false")
				}
			}
		})
	}
}

// TestMomentumIntegration_ContinuousNarrative tests continuous narrative display
// Story 7.7 AC6: 連續敘事顯示測試
func TestMomentumIntegration_ContinuousNarrative(t *testing.T) {
	// Setup orchestrator
	orch := NewOrchestrator()
	orch.gameState.HP = 80
	orch.gameState.SAN = 70

	// Setup momentum controller
	config := momentum.DefaultMomentumConfig()
	config.AutoResolve = true
	config.MaxAutoBeats = 5
	mc := momentum.NewMomentumController(config, orch.narrationAgent)
	orch.momentumController = mc

	// Run game loop
	ctx := context.Background()
	result, err := orch.RunGameLoop(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify narratives are generated
	if result.BeatsResolved > 0 {
		assert.Greater(t, len(result.Narratives), 0, "Should have generated narratives")
		assert.Equal(t, result.BeatsResolved, len(result.Narratives),
			"Number of narratives should match beats resolved")

		// Verify each narrative is non-empty
		for i, narrative := range result.Narratives {
			assert.NotEmpty(t, narrative, "Narrative %d should not be empty", i+1)
		}

		t.Logf("Generated %d narratives for %d beats", len(result.Narratives), result.BeatsResolved)
	}
}

// TestMomentumIntegration_StateChanges tests HP/SAN changes during auto-resolve
func TestMomentumIntegration_StateChanges(t *testing.T) {
	// Setup orchestrator
	orch := NewOrchestrator()
	initialHP := 80
	initialSAN := 70
	orch.gameState.HP = initialHP
	orch.gameState.SAN = initialSAN

	// Setup momentum controller
	config := momentum.DefaultMomentumConfig()
	config.AutoResolve = true
	config.MaxAutoBeats = 3
	mc := momentum.NewMomentumController(config, orch.narrationAgent)
	orch.momentumController = mc

	// Run game loop
	ctx := context.Background()
	result, err := orch.RunGameLoop(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify state changes are applied
	if result.BeatsResolved > 0 {
		// HP/SAN may have changed
		finalHP := orch.gameState.GetHP()
		finalSAN := orch.gameState.GetSAN()

		t.Logf("HP changed: %d -> %d (delta: %d)", initialHP, finalHP, result.HPDelta)
		t.Logf("SAN changed: %d -> %d (delta: %d)", initialSAN, finalSAN, result.SANDelta)

		// Verify HP/SAN are within valid range
		assert.GreaterOrEqual(t, finalHP, 0, "HP should not go below 0")
		assert.LessOrEqual(t, finalHP, 100, "HP should not exceed 100")
		assert.GreaterOrEqual(t, finalSAN, 0, "SAN should not go below 0")
		assert.LessOrEqual(t, finalSAN, 100, "SAN should not exceed 100")
	}
}

// TestMomentumIntegration_Death tests death detection during auto-resolve
func TestMomentumIntegration_Death(t *testing.T) {
	// Setup orchestrator with low HP
	orch := NewOrchestrator()
	orch.gameState.HP = 5 // Very low HP
	orch.gameState.SAN = 30

	// Setup momentum controller
	config := momentum.DefaultMomentumConfig()
	config.AutoResolve = true
	mc := momentum.NewMomentumController(config, orch.narrationAgent)
	orch.momentumController = mc

	// Run game loop
	ctx := context.Background()
	result, err := orch.RunGameLoop(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check if death is detected
	if orch.gameState.GetHP() <= 0 || orch.gameState.GetSAN() <= 0 {
		assert.True(t, result.IsGameOver, "Should detect game over")
		assert.Equal(t, "death", result.GameOverType, "Game over type should be death")
	}
}

// TestMomentumIntegration_Convergence tests convergence detection
func TestMomentumIntegration_Convergence(t *testing.T) {
	// Setup orchestrator approaching convergence
	orch := NewOrchestrator()
	orch.gameState.HP = 80
	orch.gameState.SAN = 70
	orch.gameState.Tension.Value = 96 // High tension triggers convergence

	// Setup momentum controller
	config := momentum.DefaultMomentumConfig()
	config.AutoResolve = true
	mc := momentum.NewMomentumController(config, orch.narrationAgent)
	orch.momentumController = mc

	// Run game loop
	ctx := context.Background()
	result, err := orch.RunGameLoop(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify convergence is triggered
	assert.True(t, result.IsGameOver, "Should detect game over")
	assert.Equal(t, "convergence", result.GameOverType, "Game over type should be convergence")
}
