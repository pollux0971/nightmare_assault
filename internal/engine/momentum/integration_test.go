// Package momentum 提供敘事動量控制系統
// Epic 6: Story 6.7 - Epic 6 Comprehensive Integration Tests
package momentum

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_FullMomentumCycle tests the complete flow:
// BuildContext → EvaluateRisk → DetectPlot → ShouldPause
func TestIntegration_FullMomentumCycle(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(50)
	state.SetSAN(60)
	state.Tension.SetValue(40)
	state.CurrentScene = "hospital_room"

	// Advance to beat 15
	for i := 0; i < 15; i++ {
		state.IncrementBeat()
	}

	// Build context (Story 6.6)
	ctx := ctrl.BuildContext(state, 2)

	// Verify context was built correctly
	assert.Equal(t, 15, ctx.CurrentBeat)
	assert.Equal(t, "hospital_room", ctx.CurrentScene)
	assert.Equal(t, 2, ctx.AutoResolvedBeats)

	// Risk evaluation should have happened (Story 6.3)
	assert.NotNil(t, ctx.RiskLevel)
	assert.NotNil(t, ctx.RiskFactors)

	// Plot detection should have happened (Story 6.4)
	// At beat 15 with no special conditions, should not be a plot point
	assert.False(t, ctx.IsPlotPoint)

	// ShouldPause decision (Story 6.1, Story 6.5)
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	// With RiskNone and FrequencyMedium, should not pause
	assert.False(t, shouldPause)

	// Verify DetermineStopReason
	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonNone, stopReason)
}

// TestIntegration_LowRiskAutoResolveScenario tests the scenario from design doc:
// "低風險場景自動演繹" - should NOT pause
func TestIntegration_LowRiskAutoResolveScenario(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyMedium,
		AutoResolve:  true,
		MaxAutoBeats: 5,
		PauseOnRisk:  RiskMedium,
		PauseOnPlot:  true,
		PauseOnNPC:   true,
		PauseOnEvent: true,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	// Setup low-risk scenario
	state := engine.NewGameStateV2()
	state.SetHP(80)
	state.SetSAN(70)
	state.Tension.SetValue(20)
	state.CurrentScene = "安全房" // Safe scene
	state.IncrementBeat()

	ctx := ctrl.BuildContext(state, 0)

	// Verify low risk
	assert.True(t, ctx.RiskLevel < RiskMedium, "Expected low risk")

	// Should NOT pause (auto-resolve)
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.False(t, shouldPause, "Low risk scenario should auto-resolve")
}

// TestIntegration_HighRiskPauseScenario tests:
// "中風險場景暫停詢問" - should pause
func TestIntegration_HighRiskPauseScenario(t *testing.T) {
	config := DefaultMomentumConfig() // FrequencyMedium, PauseOnRisk: RiskMedium
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	// Setup high-risk scenario
	state := engine.NewGameStateV2()
	state.SetHP(25)         // Low HP → RiskHigh
	state.SetSAN(70)
	state.Tension.SetValue(85) // High tension → RiskHigh
	state.CurrentScene = "危險走廊"

	// Add multiple active rules
	for i := 0; i < 5; i++ {
		state.ActiveRules = append(state.ActiveRules, &engine.ActiveRule{
			ID:   "rule_" + string(rune('A'+i)),
			Name: "Test Rule",
		})
	}

	// First evaluate risk directly to verify setup
	evaluator := NewRiskEvaluator()
	assessment, err := evaluator.EvaluateContext(state, state.CurrentScene)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, assessment.Level, RiskMedium, "Expected medium or higher risk")

	// Now build context and manually set the evaluated risk
	ctx := ctrl.BuildContext(state, 0)
	ctx.RiskLevel = assessment.Level

	// Verify risk factors
	assert.NotEmpty(t, ctx.RiskFactors, "Expected risk factors")

	// Should pause
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "High risk scenario should pause")

	// Verify stop reason
	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonRiskLevel, stopReason)
}

// TestIntegration_PlotPointTriggersPause tests:
// "劇情點觸發暫停" - plot point should pause even if low risk
func TestIntegration_PlotPointTriggersPause(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyLow, // Low frequency
		AutoResolve:  true,
		MaxAutoBeats: 10,
		PauseOnRisk:  RiskHigh, // High threshold
		PauseOnPlot:  true,
		PauseOnNPC:   false,
		PauseOnEvent: false,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	// Setup low-risk but plot point scenario
	state := engine.NewGameStateV2()
	state.SetHP(100)
	state.SetSAN(100)
	state.Tension.SetValue(20)

	// Advance to beat 30 (act transition)
	for i := 0; i < 30; i++ {
		state.IncrementBeat()
	}

	ctx := ctrl.BuildContext(state, 0)

	// Verify plot point detected
	assert.True(t, ctx.IsPlotPoint, "Expected plot point at beat 30")
	assert.Equal(t, PlotPointActTransition, ctx.PlotPointType)

	// Should pause despite low risk
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "Plot point should trigger pause")

	// Verify stop reason
	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonPlotPoint, stopReason)
}

// TestIntegration_NPCInitiationTriggersPause tests:
// "NPC 主動對話觸發聊天室"
func TestIntegration_NPCInitiationTriggersPause(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyLow,
		AutoResolve:  true,
		MaxAutoBeats: 10,
		PauseOnRisk:  RiskLethal, // Very high threshold
		PauseOnPlot:  false,
		PauseOnNPC:   true, // Enabled
		PauseOnEvent: false,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(100)
	state.SetSAN(100)

	ctx := ctrl.BuildContext(state, 0)
	// Manually set NPC initiation (since NPCManager integration is pending)
	ctx.NPCInitiatesConversation = true
	ctx.InitiatingNPC = "Dr. Zhang"

	// Should pause for NPC conversation
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "NPC initiation should trigger pause")

	// Verify stop reason
	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonNPCConversation, stopReason)
}

// TestIntegration_MaxAutoBeatsTriggersPause tests:
// "最大自動演繹回合限制"
func TestIntegration_MaxAutoBeatsTriggersPause(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyLow,
		AutoResolve:  true,
		MaxAutoBeats: 5, // Max 5 auto beats
		PauseOnRisk:  RiskHigh,
		PauseOnPlot:  false,
		PauseOnNPC:   false,
		PauseOnEvent: false,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(100)
	state.SetSAN(100)
	state.Tension.SetValue(10)

	// Simulate 5 auto-resolved beats
	ctx := ctrl.BuildContext(state, 5)

	// Should pause due to max auto beats
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "Max auto beats should trigger pause")

	// Verify stop reason
	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonMaxAutoBeats, stopReason)
}

// TestIntegration_FrequencyLevelEffects tests all three frequency levels
func TestIntegration_FrequencyLevelEffects(t *testing.T) {
	tests := []struct {
		name        string
		frequency   FrequencyLevel
		riskLevel   RiskLevel
		shouldPause bool
	}{
		// FrequencyHigh: Always pauses
		{"High freq, no risk", FrequencyHigh, RiskNone, true},
		{"High freq, low risk", FrequencyHigh, RiskLow, true},
		{"High freq, high risk", FrequencyHigh, RiskHigh, true},

		// FrequencyMedium: Pauses at RiskMedium+
		{"Medium freq, no risk", FrequencyMedium, RiskNone, false},
		{"Medium freq, low risk", FrequencyMedium, RiskLow, false},
		{"Medium freq, medium risk", FrequencyMedium, RiskMedium, true},
		{"Medium freq, high risk", FrequencyMedium, RiskHigh, true},

		// FrequencyLow: Pauses at RiskHigh+
		{"Low freq, no risk", FrequencyLow, RiskNone, false},
		{"Low freq, low risk", FrequencyLow, RiskLow, false},
		{"Low freq, medium risk", FrequencyLow, RiskMedium, false},
		{"Low freq, high risk", FrequencyLow, RiskHigh, true},
		{"Low freq, lethal risk", FrequencyLow, RiskLethal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MomentumConfig{
				Frequency:    tt.frequency,
				PauseOnRisk:  RiskLethal, // Set high to test frequency logic
				PauseOnPlot:  false,
				PauseOnNPC:   false,
				PauseOnEvent: false,
			}
			ctrl := NewMomentumController(config, &mockNarrationAgent{})

			ctx := &NarrativeContext{
				RiskLevel: tt.riskLevel,
			}

			shouldPause := ctrl.ShouldPauseForChoice(ctx)
			assert.Equal(t, tt.shouldPause, shouldPause,
				"Frequency %s with risk %s should pause=%v",
				tt.frequency.String(), tt.riskLevel.String(), tt.shouldPause)
		})
	}
}

// TestIntegration_MultipleRiskFactorsCombine tests that multiple risk factors
// combine correctly to determine overall risk level
func TestIntegration_MultipleRiskFactorsCombine(t *testing.T) {
	tests := []struct {
		name          string
		hp            int
		san           int
		tension       int
		ruleCount     int
		scene         string
		expectedRisk  RiskLevel
		minFactors    int
	}{
		{
			name:         "All high risks",
			hp:           15,
			san:          15,
			tension:      90,
			ruleCount:    6,
			scene:        "地下室",
			expectedRisk: RiskHigh,
			minFactors:   4, // low_hp, low_san, high_tension, many_active_rules, dangerous_scene
		},
		{
			name:         "Multiple medium risks",
			hp:           35,
			san:          35,
			tension:      60,
			ruleCount:    3,
			scene:        "走廊",
			expectedRisk: RiskMedium,
			minFactors:   4, // medium_hp, medium_san, medium_tension, some_active_rules, dangerous_scene
		},
		{
			name:         "One high dominates multiple mediums",
			hp:           15,  // High risk
			san:          100,
			tension:      30,
			ruleCount:    0,
			scene:        "安全房",
			expectedRisk: RiskHigh,
			minFactors:   1, // low_hp
		},
		{
			name:         "No risk factors",
			hp:           100,
			san:          100,
			tension:      20,
			ruleCount:    0,
			scene:        "休息室",
			expectedRisk: RiskNone,
			minFactors:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.SetHP(tt.hp)
			state.SetSAN(tt.san)
			state.Tension.SetValue(tt.tension)

			for i := 0; i < tt.ruleCount; i++ {
				state.ActiveRules = append(state.ActiveRules, &engine.ActiveRule{
					ID:   "rule_" + string(rune('A'+i)),
					Name: "Test Rule",
				})
			}

			// Use RiskEvaluator directly
			evaluator := NewRiskEvaluator()
			assessment, err := evaluator.EvaluateContext(state, tt.scene)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedRisk, assessment.Level,
				"Risk level mismatch for scenario: %s", tt.name)
			assert.GreaterOrEqual(t, len(assessment.Factors), tt.minFactors,
				"Expected at least %d risk factors", tt.minFactors)
		})
	}
}

// TestIntegration_ConfigSwitchesWork tests that all config switches work correctly
func TestIntegration_ConfigSwitchesWork(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(*MomentumConfig)
		setupCtx    func(*NarrativeContext)
		shouldPause bool
		stopReason  StopReason
	}{
		{
			name: "PauseOnNPC enabled",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnNPC = true
				c.PauseOnEvent = false
				c.PauseOnPlot = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.NPCInitiatesConversation = true
				ctx.InitiatingNPC = "NPC1"
				ctx.RiskLevel = RiskNone
			},
			shouldPause: true,
			stopReason:  StopReasonNPCConversation,
		},
		{
			name: "PauseOnNPC disabled",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnNPC = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.NPCInitiatesConversation = true
				ctx.RiskLevel = RiskNone
			},
			shouldPause: false,
			stopReason:  StopReasonNone,
		},
		{
			name: "PauseOnEvent enabled",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnEvent = true
				c.PauseOnNPC = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.PendingEvents = []*GameEvent{{IsMajor: true}}
				ctx.RiskLevel = RiskNone
			},
			shouldPause: true,
			stopReason:  StopReasonMajorEvent,
		},
		{
			name: "PauseOnEvent disabled",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnEvent = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.PendingEvents = []*GameEvent{{IsMajor: true}}
				ctx.RiskLevel = RiskNone
			},
			shouldPause: false,
			stopReason:  StopReasonNone,
		},
		{
			name: "PauseOnPlot enabled",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnPlot = true
				c.PauseOnNPC = false
				c.PauseOnEvent = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.IsPlotPoint = true
				ctx.PlotPointType = PlotPointActTransition
				ctx.RiskLevel = RiskNone
			},
			shouldPause: true,
			stopReason:  StopReasonPlotPoint,
		},
		{
			name: "PauseOnPlot disabled",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnPlot = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.IsPlotPoint = true
				ctx.RiskLevel = RiskNone
			},
			shouldPause: false,
			stopReason:  StopReasonNone,
		},
		{
			name: "PauseOnRisk threshold",
			setupConfig: func(c *MomentumConfig) {
				c.PauseOnRisk = RiskHigh
				c.PauseOnNPC = false
				c.PauseOnEvent = false
				c.PauseOnPlot = false
				c.Frequency = FrequencyLow
			},
			setupCtx: func(ctx *NarrativeContext) {
				ctx.RiskLevel = RiskHigh
			},
			shouldPause: true,
			stopReason:  StopReasonRiskLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultMomentumConfig()
			tt.setupConfig(config)

			ctrl := NewMomentumController(config, &mockNarrationAgent{})

			ctx := &NarrativeContext{}
			tt.setupCtx(ctx)

			shouldPause := ctrl.ShouldPauseForChoice(ctx)
			assert.Equal(t, tt.shouldPause, shouldPause)

			stopReason := ctrl.DetermineStopReason(ctx)
			if tt.shouldPause {
				assert.Equal(t, tt.stopReason, stopReason)
			}
		})
	}
}

// TestIntegration_ComponentInteractions tests that all components work together:
// RiskEvaluator ↔ PlotDetector ↔ ContextBuilder ↔ Controller
func TestIntegration_ComponentInteractions(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	// Create complex state
	state := engine.NewGameStateV2()
	state.SetHP(30)     // Will trigger medium_hp risk factor
	state.SetSAN(100)
	state.Tension.SetValue(75) // Will trigger medium_tension risk factor
	state.CurrentScene = "走廊" // Will trigger dangerous_scene risk factor

	// Add global seed for plot detection
	globalSeed, _ := seed.NewGlobalSeed(
		"GS001",
		"Test seed",
		"truth1",
		"tragic",
		[]seed.ClueTier{
			{Tier: 1, Content: "Clue 1", BeatStart: 0, BeatEnd: 10},
			{Tier: 2, Content: "Clue 2", BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Clue 3", BeatStart: 21, BeatEnd: 30},
		},
	)
	state.AddGlobalSeed(globalSeed)

	// Advance to beat 28 (near act end + approaching seed milestone)
	for i := 0; i < 28; i++ {
		state.IncrementBeat()
	}

	// Step 1: Build context (uses RiskEvaluator and PlotDetector internally)
	ctx := ctrl.BuildContext(state, 3)

	// Verify ContextBuilder output
	assert.Equal(t, 28, ctx.CurrentBeat)
	assert.Equal(t, "走廊", ctx.CurrentScene)
	assert.Equal(t, 3, ctx.AutoResolvedBeats)

	// Verify RiskEvaluator was called and factors extracted
	assert.GreaterOrEqual(t, len(ctx.RiskFactors), 2, "Should have multiple risk factors")
	assert.Contains(t, ctx.RiskFactors, "medium_hp")
	assert.Contains(t, ctx.RiskFactors, "medium_tension")

	// Verify PlotDetector was called
	assert.True(t, ctx.IsPlotPoint, "Should detect plot point at beat 28")
	assert.Equal(t, PlotPointActTransition, ctx.PlotPointType)

	// Step 2: Controller makes decision based on context
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "Should pause due to plot point")

	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonPlotPoint, stopReason)

	// Verify all components contributed to final decision
	t.Logf("Integration test passed with:")
	t.Logf("  - CurrentBeat: %d", ctx.CurrentBeat)
	t.Logf("  - RiskFactors: %v", ctx.RiskFactors)
	t.Logf("  - IsPlotPoint: %v (%s)", ctx.IsPlotPoint, ctx.PlotPointType)
	t.Logf("  - ShouldPause: %v", shouldPause)
	t.Logf("  - StopReason: %v", stopReason)
}

// TestIntegration_RealWorldScenario1_SafeExploration tests:
// "Player is exploring a safe area with full HP/SAN"
func TestIntegration_RealWorldScenario1_SafeExploration(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(100)
	state.SetSAN(100)
	state.Tension.SetValue(15)
	state.CurrentScene = "lobby"

	// Advance a few beats
	for i := 0; i < 10; i++ {
		state.IncrementBeat()
	}

	ctx := ctrl.BuildContext(state, 2)

	// Should be low risk
	assert.Equal(t, RiskNone, ctx.RiskLevel)
	assert.Empty(t, ctx.RiskFactors)

	// Should NOT pause (auto-resolve)
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.False(t, shouldPause, "Safe exploration should auto-resolve")
}

// TestIntegration_RealWorldScenario2_CriticalMoment tests:
// "Player at low HP, high tension, approaching critical plot point"
func TestIntegration_RealWorldScenario2_CriticalMoment(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(18)  // Critical HP
	state.SetSAN(25) // Low SAN
	state.Tension.SetValue(88)
	state.CurrentScene = "basement"

	// Add many active rules
	for i := 0; i < 6; i++ {
		state.ActiveRules = append(state.ActiveRules, &engine.ActiveRule{
			ID:   "rule_" + string(rune('A'+i)),
			Name: "Hidden Rule",
		})
	}

	// Advance to near act end
	for i := 0; i < 29; i++ {
		state.IncrementBeat()
	}

	// Evaluate risk
	evaluator := NewRiskEvaluator()
	assessment, err := evaluator.EvaluateContext(state, state.CurrentScene)
	require.NoError(t, err)

	ctx := ctrl.BuildContext(state, 0)
	ctx.RiskLevel = assessment.Level

	// Should be high risk
	assert.GreaterOrEqual(t, ctx.RiskLevel, RiskHigh)
	assert.NotEmpty(t, ctx.RiskFactors)

	// Should be plot point
	assert.True(t, ctx.IsPlotPoint)

	// Should DEFINITELY pause
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "Critical moment should pause")

	// Multiple valid stop reasons (plot point takes priority in DetermineStopReason)
	stopReason := ctrl.DetermineStopReason(ctx)
	// Plot point has priority over risk in the DetermineStopReason logic
	assert.Equal(t, StopReasonPlotPoint, stopReason)
}

// TestIntegration_RealWorldScenario3_NPCEncounter tests:
// "NPC approaches player for important conversation"
func TestIntegration_RealWorldScenario3_NPCEncounter(t *testing.T) {
	config := DefaultMomentumConfig()
	config.PauseOnNPC = true

	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(80)
	state.SetSAN(75)
	state.CurrentScene = "hallway"

	ctx := ctrl.BuildContext(state, 1)
	// Simulate NPC initiation
	ctx.NPCInitiatesConversation = true
	ctx.InitiatingNPC = "Dr. Zhang"

	// Should pause for NPC even if low risk
	shouldPause := ctrl.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "NPC encounter should pause")

	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonNPCConversation, stopReason)
}

// TestIntegration_RealWorldScenario4_ExtendedAutoResolve tests:
// "System auto-resolves multiple safe beats until max limit"
func TestIntegration_RealWorldScenario4_ExtendedAutoResolve(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyLow,
		AutoResolve:  true,
		MaxAutoBeats: 5,
		PauseOnRisk:  RiskHigh,
		PauseOnPlot:  false, // Disable to focus on MaxAutoBeats
		PauseOnNPC:   false,
		PauseOnEvent: false,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.SetHP(100)
	state.SetSAN(100)

	// Simulate 4 auto-resolved beats - should NOT pause
	ctx := ctrl.BuildContext(state, 4)
	assert.False(t, ctrl.ShouldPauseForChoice(ctx), "Should continue at 4 beats")

	// Simulate 5 auto-resolved beats - should pause
	ctx = ctrl.BuildContext(state, 5)
	assert.True(t, ctrl.ShouldPauseForChoice(ctx), "Should pause at max auto beats")

	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonMaxAutoBeats, stopReason)
}

// TestIntegration_ConfigPresets tests all configuration presets work correctly
func TestIntegration_ConfigPresets(t *testing.T) {
	tests := []struct {
		name        string
		configFunc  func() *MomentumConfig
		riskLevel   RiskLevel
		shouldPause bool
	}{
		{
			name:        "Easy difficulty - high frequency",
			configFunc:  func() *MomentumConfig { return GetMomentumConfigForDifficulty("easy") },
			riskLevel:   RiskLow,
			shouldPause: true, // Easy pauses on RiskLow
		},
		{
			name:        "Normal difficulty - medium frequency",
			configFunc:  func() *MomentumConfig { return GetMomentumConfigForDifficulty("normal") },
			riskLevel:   RiskLow,
			shouldPause: false, // Normal doesn't pause on RiskLow
		},
		{
			name:        "Nightmare difficulty - low frequency",
			configFunc:  func() *MomentumConfig { return GetMomentumConfigForDifficulty("nightmare") },
			riskLevel:   RiskMedium,
			shouldPause: false, // Nightmare only pauses on RiskHigh+
		},
		{
			name:        "Cinematic mode - minimal pauses",
			configFunc:  CinematicModeConfig,
			riskLevel:   RiskMedium,
			shouldPause: false, // Cinematic pauses only on RiskHigh
		},
		{
			name:        "Interactive mode - always pauses",
			configFunc:  InteractiveModeConfig,
			riskLevel:   RiskNone,
			shouldPause: true, // Interactive always pauses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.configFunc()
			ctrl := NewMomentumController(config, &mockNarrationAgent{})

			ctx := &NarrativeContext{
				RiskLevel:         tt.riskLevel,
				AutoResolvedBeats: 0,
			}

			shouldPause := ctrl.ShouldPauseForChoice(ctx)
			assert.Equal(t, tt.shouldPause, shouldPause,
				"Config %s with risk %s should pause=%v",
				tt.name, tt.riskLevel.String(), tt.shouldPause)
		})
	}
}

// TestIntegration_PriorityOrder tests that conditions are evaluated in correct priority order
func TestIntegration_PriorityOrder(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	// Create context with ALL triggers active
	ctx := &NarrativeContext{
		NPCInitiatesConversation: true,                          // Priority 1
		PendingEvents:            []*GameEvent{{IsMajor: true}}, // Priority 2
		IsPlotPoint:              true,                          // Priority 3
		RiskLevel:                RiskHigh,                      // Priority 4
		AutoResolvedBeats:        5,                             // Priority 5
	}

	// Should pause
	assert.True(t, ctrl.ShouldPauseForChoice(ctx))

	// Stop reason should be highest priority (NPC)
	stopReason := ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonNPCConversation, stopReason,
		"Highest priority condition should determine stop reason")

	// Test with NPC disabled
	config.PauseOnNPC = false
	stopReason = ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonMajorEvent, stopReason,
		"Next priority condition should determine stop reason")

	// Test with Event disabled
	config.PauseOnEvent = false
	stopReason = ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonPlotPoint, stopReason)

	// Test with Plot disabled
	config.PauseOnPlot = false
	stopReason = ctrl.DetermineStopReason(ctx)
	assert.Equal(t, StopReasonRiskLevel, stopReason)
}

// TestIntegration_EdgeCases tests various edge cases
func TestIntegration_EdgeCases(t *testing.T) {
	t.Run("Nil context", func(t *testing.T) {
		ctrl := NewMomentumController(DefaultMomentumConfig(), &mockNarrationAgent{})
		shouldPause := ctrl.ShouldPauseForChoice(nil)
		assert.True(t, shouldPause, "Nil context should safely pause")
	})

	t.Run("Nil state to BuildContext", func(t *testing.T) {
		ctrl := NewMomentumController(DefaultMomentumConfig(), &mockNarrationAgent{})
		ctx := ctrl.BuildContext(nil, 0)
		assert.NotNil(t, ctx)
		assert.Equal(t, RiskNone, ctx.RiskLevel)
	})

	t.Run("Zero max auto beats", func(t *testing.T) {
		config := &MomentumConfig{
			Frequency:    FrequencyMedium,
			AutoResolve:  true,
			MaxAutoBeats: 0, // Zero limit
			PauseOnRisk:  RiskHigh,
		}
		ctrl := NewMomentumController(config, &mockNarrationAgent{})
		ctx := &NarrativeContext{
			AutoResolvedBeats: 0,
			RiskLevel:         RiskNone,
		}
		// Should NOT pause due to MaxAutoBeats (0 >= 0 is true, but logic handles it)
		// Actually, with MaxAutoBeats=0 and AutoResolvedBeats=0, it WILL pause
		shouldPause := ctrl.ShouldPauseForChoice(ctx)
		assert.True(t, shouldPause, "Zero max auto beats with 0 resolved should pause")
	})

	t.Run("Negative HP/SAN", func(t *testing.T) {
		state := engine.NewGameStateV2()
		// GameStateV2 may clamp values, but test evaluator handles it
		evaluator := NewRiskEvaluator()
		// Set to minimum values
		state.SetHP(0)
		state.SetSAN(0)
		assessment, err := evaluator.EvaluateContext(state, "test")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, assessment.Level, RiskHigh, "Zero HP/SAN should be high risk")
	})

	t.Run("Empty scene name", func(t *testing.T) {
		state := engine.NewGameStateV2()
		evaluator := NewRiskEvaluator()
		assessment, err := evaluator.EvaluateContext(state, "")
		require.NoError(t, err)
		// Empty scene should not add danger
		hasSceneDanger := false
		for _, f := range assessment.Factors {
			if f.Name == "dangerous_scene" {
				hasSceneDanger = true
			}
		}
		assert.False(t, hasSceneDanger, "Empty scene should not be dangerous")
	})
}

// TestIntegration_StatePersistence tests that component state persists correctly
// (e.g., PlotDetector milestones)
func TestIntegration_StatePersistence(t *testing.T) {
	ctrl := NewMomentumController(DefaultMomentumConfig(), &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	// Add seed for 30% milestone
	state.GlobalSeeds = []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 2}}

	// First call: should detect milestone
	ctx1 := ctrl.BuildContext(state, 0)
	assert.True(t, ctx1.IsPlotPoint)
	assert.Equal(t, PlotPointSeedMilestone30, ctx1.PlotPointType)

	// Second call: milestone already reached, should NOT detect again
	ctx2 := ctrl.BuildContext(state, 0)
	// PlotDetector maintains milestone state
	if ctx2.IsPlotPoint && ctx2.PlotPointType == PlotPointSeedMilestone30 {
		t.Error("PlotDetector should not re-trigger same milestone")
	}

	// Verify milestone was recorded
	assert.True(t, ctrl.plotDetector.HasReachedMilestone("seed_30"))
}
