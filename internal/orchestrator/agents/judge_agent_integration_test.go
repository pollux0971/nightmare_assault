package agents

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJudgeAgentIntegration_RuleEngineOnly tests judgment using only rule engine
func TestJudgeAgentIntegration_RuleEngineOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewJudgeAgent(AgentConfig{
		Name:      "JudgeAgent",
		LLMClient: nil, // Rule engine only
		Timeout:   5 * time.Second,
	})

	tests := []struct {
		name            string
		choice          string
		rules           []JudgeHiddenRule
		difficulty      string
		expectImpact    ImpactLevel
		expectNextAction NextActionType
		expectViolations int
	}{
		{
			name:   "safe_choice_no_violation",
			choice: "我小心地檢查房間角落",
			rules: []JudgeHiddenRule{
				{
					ID:              "R-001",
					Name:            "倒影殺手",
					Type:            RuleTypeScene,
					TriggerKeywords: []string{"鏡子", "倒影"},
					Punishment:      RulePunishment{IsFatal: false, SANDamage: 40},
				},
			},
			difficulty:       "normal",
			expectImpact:     ImpactNone,
			expectNextAction: ActionContinueStory,
			expectViolations: 0,
		},
		{
			name:   "moderate_violation_scene_rule",
			choice: "我凝視著牆上的鏡子",
			rules: []JudgeHiddenRule{
				{
					ID:              "R-001",
					Name:            "倒影殺手",
					Type:            RuleTypeScene,
					TriggerKeywords: []string{"鏡子", "倒影", "凝視"},
					TriggerCondition: "凝視鏡子超過3秒",
					Punishment:      RulePunishment{IsFatal: false, SANDamage: 40},
				},
			},
			difficulty:       "normal",
			expectImpact:     ImpactModerate,
			expectNextAction: ActionApplyDamage,
			expectViolations: 1,
		},
		{
			name:   "lethal_violation_behavior_rule",
			choice: "我忍不住大聲呼喊求救",
			rules: []JudgeHiddenRule{
				{
					ID:              "R-002",
					Name:            "聲音禁忌",
					Type:            RuleTypeBehavior,
					TriggerKeywords: []string{"呼喊", "大聲", "尖叫"},
					TriggerCondition: "在黑暗中發出聲音",
					Punishment:      RulePunishment{IsFatal: true, HPDamage: 100},
				},
			},
			difficulty:       "hell",
			expectImpact:     ImpactLethal,
			expectNextAction: ActionTriggerDeath,
			expectViolations: 1,
		},
		{
			name:   "multiple_violations_priority_ordering",
			choice: "我對著鏡子大聲說話",
			rules: []JudgeHiddenRule{
				{
					ID:              "R-001",
					Name:            "倒影殺手",
					Type:            RuleTypeScene,
					TriggerKeywords: []string{"鏡子"},
					Punishment:      RulePunishment{IsFatal: false, SANDamage: 40},
				},
				{
					ID:              "R-002",
					Name:            "聲音禁忌",
					Type:            RuleTypeBehavior,
					TriggerKeywords: []string{"大聲", "說話"},
					Punishment:      RulePunishment{IsFatal: false, HPDamage: 30},
				},
			},
			difficulty:       "normal",
			expectImpact:     ImpactModerate,
			expectNextAction: ActionApplyDamage,
			expectViolations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &JudgeRequest{
				PlayerChoice: tt.choice,
				GameState: &GameStateSnapshot{
					HP:           100,
					SAN:          80,
					CurrentScene: "走廊",
					Difficulty:   tt.difficulty,
					RuleWarnings: make(map[string]int),
				},
				ActiveRules: tt.rules,
			}

			ctx := context.Background()
			start := time.Now()

			response, err := agent.InvokeJudge(ctx, request)
			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, response)

			// AC #6: Rule engine should be fast (< 50ms)
			assert.LessOrEqual(t, duration, 50*time.Millisecond,
				"Rule engine judgment should be under 50ms")

			// Verify judgment results
			assert.Equal(t, tt.expectImpact, response.ImpactLevel,
				"Impact level mismatch")
			assert.Equal(t, tt.expectNextAction, response.NextAction,
				"Next action mismatch")
			assert.Len(t, response.RulesViolated, tt.expectViolations,
				"Violation count mismatch")

			// Verify priority ordering for multiple violations
			if tt.expectViolations > 1 {
				for i := 0; i < len(response.RulesViolated)-1; i++ {
					currentPriority := GetRulePriority(response.RulesViolated[i].RuleType)
					nextPriority := GetRulePriority(response.RulesViolated[i+1].RuleType)
					assert.GreaterOrEqual(t, currentPriority, nextPriority,
						"Rules should be sorted by priority")
				}
			}

			t.Logf("Judgment completed in %v: %s → %s (Violations: %d)",
				duration, tt.choice, response.ImpactLevel, len(response.RulesViolated))
		})
	}
}

// TestJudgeAgentIntegration_WarningSystem tests the warning mechanism
func TestJudgeAgentIntegration_WarningSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewJudgeAgent(AgentConfig{
		Name:      "JudgeAgent",
		LLMClient: nil,
		Timeout:   5 * time.Second,
	})

	rule := JudgeHiddenRule{
		ID:              "R-001",
		Name:            "倒影殺手",
		Type:            RuleTypeScene,
		TriggerKeywords: []string{"鏡子"},
		Punishment:      RulePunishment{IsFatal: false, SANDamage: 40},
		MaxWarnings:     2,
	}

	tests := []struct {
		name              string
		difficulty        string
		initialWarnings   map[string]int
		expectSANDamage   int
		expectWarningsLeft int
	}{
		{
			name:              "easy_first_warning",
			difficulty:        "easy",
			initialWarnings:   map[string]int{},
			expectSANDamage:   -16, // 40 * 0.5 * 0.8 = 16
			expectWarningsLeft: 1,
		},
		{
			name:              "easy_second_warning",
			difficulty:        "easy",
			initialWarnings:   map[string]int{"R-001": 1},
			expectSANDamage:   -16, // Still half damage
			expectWarningsLeft: 0,
		},
		{
			name:              "easy_no_warnings_full_damage",
			difficulty:        "easy",
			initialWarnings:   map[string]int{"R-001": 0},
			expectSANDamage:   -32, // 40 * 0.8 = 32
			expectWarningsLeft: 0,
		},
		{
			name:              "hell_no_warnings",
			difficulty:        "hell",
			initialWarnings:   map[string]int{},
			expectSANDamage:   -48, // 40 * 1.2 = 48 (no warnings in hell)
			expectWarningsLeft: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &JudgeRequest{
				PlayerChoice: "我看向鏡子",
				GameState: &GameStateSnapshot{
					HP:           100,
					SAN:          80,
					Difficulty:   tt.difficulty,
					RuleWarnings: tt.initialWarnings,
				},
				ActiveRules: []JudgeHiddenRule{rule},
			}

			ctx := context.Background()
			response, err := agent.InvokeJudge(ctx, request)

			require.NoError(t, err)
			require.NotNil(t, response)

			// Check SAN damage
			assert.Equal(t, tt.expectSANDamage, response.SuggestedStateChanges.SAN,
				"SAN damage mismatch")

			// Check warnings remaining
			if warningsLeft, exists := response.SuggestedStateChanges.WarningsRemaining["R-001"]; exists {
				assert.Equal(t, tt.expectWarningsLeft, warningsLeft,
					"Warnings remaining mismatch")
			}

			t.Logf("Difficulty: %s, SAN Damage: %d, Warnings Left: %d",
				tt.difficulty, response.SuggestedStateChanges.SAN, tt.expectWarningsLeft)
		})
	}
}

// TestJudgeAgentIntegration_DifficultyScaling tests difficulty multipliers
func TestJudgeAgentIntegration_DifficultyScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewJudgeAgent(AgentConfig{
		Name:      "JudgeAgent",
		LLMClient: nil,
		Timeout:   5 * time.Second,
	})

	rule := JudgeHiddenRule{
		ID:              "R-003",
		Name:            "物理規則",
		Type:            RuleTypeBehavior,
		TriggerKeywords: []string{"觸碰"},
		Punishment:      RulePunishment{IsFatal: false, HPDamage: 50, SANDamage: 0},
	}

	tests := []struct {
		difficulty      string
		expectHPDamage  int
		multiplier      float64
	}{
		{"easy", -20, 0.8},    // 50 * 0.5 * 0.8 = 20 (with warning)
		{"normal", -25, 1.0},  // 50 * 0.5 * 1.0 = 25
		{"hard", -25, 1.0},    // 50 * 0.5 * 1.0 = 25
		{"hell", -30, 1.2},    // 50 * 0.5 * 1.2 = 30 (but hell has no warnings, so full)
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			request := &JudgeRequest{
				PlayerChoice: "我觸碰了它",
				GameState: &GameStateSnapshot{
					HP:           100,
					SAN:          80,
					Difficulty:   tt.difficulty,
					RuleWarnings: map[string]int{}, // First violation
				},
				ActiveRules: []JudgeHiddenRule{rule},
			}

			ctx := context.Background()
			response, err := agent.InvokeJudge(ctx, request)

			require.NoError(t, err)
			require.NotNil(t, response)

			// For hell difficulty, recalculate expected damage (no warnings)
			expectedDamage := tt.expectHPDamage
			if tt.difficulty == "hell" {
				expectedDamage = int(float64(rule.Punishment.HPDamage) * tt.multiplier)
				expectedDamage = -expectedDamage
			}

			assert.Equal(t, expectedDamage, response.SuggestedStateChanges.HP,
				"HP damage mismatch for difficulty: %s", tt.difficulty)

			t.Logf("Difficulty: %s, HP Damage: %d (multiplier: %.1f)",
				tt.difficulty, response.SuggestedStateChanges.HP, tt.multiplier)
		})
	}
}

// TestJudgeAgentIntegration_WithLLMFallback tests LLM integration
func TestJudgeAgentIntegration_WithLLMFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load config to get real API provider
	cfg, err := config.Load()
	require.NoError(t, err)

	// Only run if API is configured
	if !cfg.IsConfigured() {
		t.Skip("API not configured, skipping LLM integration test")
	}

	apiKey := cfg.GetAPIKey(cfg.API.Provider.ProviderID)
	require.NotEmpty(t, apiKey, "API key must be set for integration tests")

	// Create real provider
	prov, err := api.NewProvider(api.ProviderConfig{
		ProviderID: cfg.API.Provider.ProviderID,
		APIKey:     apiKey,
		Model:      cfg.API.Provider.Model,
		MaxTokens:  cfg.API.Provider.MaxTokens,
		BaseURL:    cfg.API.Provider.BaseURL,
	})
	require.NoError(t, err)

	// Wrap provider in adapter
	llmClient := &providerAdapter{provider: prov}

	agent := NewJudgeAgent(AgentConfig{
		Name:      "JudgeAgent",
		LLMClient: llmClient,
		Timeout:   30 * time.Second,
	})

	request := &JudgeRequest{
		PlayerChoice: "我猶豫了一下，決定慢慢靠近那扇門",
		GameState: &GameStateSnapshot{
			HP:           100,
			SAN:          70,
			CurrentScene: "陰暗走廊",
			Difficulty:   "normal",
			RuleWarnings: make(map[string]int),
		},
		ActiveRules: []JudgeHiddenRule{
			{
				ID:               "R-004",
				Name:             "門後禁忌",
				Type:             RuleTypeScene,
				TriggerKeywords:  []string{"打開門", "推門", "開門"},
				TriggerCondition: "在半夜打開緊閉的門",
				Punishment:       RulePunishment{IsFatal: false, SANDamage: 30},
			},
		},
	}

	ctx := context.Background()
	start := time.Now()

	response, err := agent.InvokeJudge(ctx, request)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #6: Total response time should be < 500ms
	assert.LessOrEqual(t, duration, 500*time.Millisecond,
		"Judgment with LLM should be under 500ms")

	t.Logf("LLM judgment completed in %v", duration)
	t.Logf("Impact: %s, Next Action: %s", response.ImpactLevel, response.NextAction)
	t.Logf("Reasoning: %s", response.Reasoning)

	// Verify response structure
	assert.NotNil(t, response.ImpactLevel)
	assert.NotNil(t, response.NextAction)
}

// TestJudgeAgentIntegration_CompleteFlow tests complete judgment flow
func TestJudgeAgentIntegration_CompleteFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewJudgeAgent(AgentConfig{
		Name:      "JudgeAgent",
		LLMClient: nil,
		Timeout:   10 * time.Second,
	})

	// Scenario: Player gradually violates rules
	scenario := []struct {
		turn            int
		choice          string
		currentHP       int
		currentSAN      int
		currentWarnings map[string]int
		expectImpact    ImpactLevel
	}{
		{
			turn:            1,
			choice:          "我小心地環顧四周",
			currentHP:       100,
			currentSAN:      100,
			currentWarnings: make(map[string]int),
			expectImpact:    ImpactNone,
		},
		{
			turn:            2,
			choice:          "我看向牆上的鏡子",
			currentHP:       100,
			currentSAN:      100,
			currentWarnings: make(map[string]int),
			expectImpact:    ImpactModerate,
		},
		{
			turn:            3,
			choice:          "我再次凝視鏡子",
			currentHP:       100,
			currentSAN:      80,
			currentWarnings: map[string]int{"R-001": 1},
			expectImpact:    ImpactModerate,
		},
		{
			turn:            4,
			choice:          "我第三次看向鏡子",
			currentHP:       100,
			currentSAN:      64,
			currentWarnings: map[string]int{"R-001": 0},
			expectImpact:    ImpactModerate,
		},
	}

	rules := []JudgeHiddenRule{
		{
			ID:              "R-001",
			Name:            "倒影殺手",
			Type:            RuleTypeScene,
			TriggerKeywords: []string{"鏡子", "凝視"},
			TriggerCondition: "凝視鏡子",
			Punishment:      RulePunishment{IsFatal: false, SANDamage: 40},
		},
	}

	ctx := context.Background()

	for _, turn := range scenario {
		t.Run(turn.choice, func(t *testing.T) {
			request := &JudgeRequest{
				PlayerChoice: turn.choice,
				GameState: &GameStateSnapshot{
					HP:           turn.currentHP,
					SAN:          turn.currentSAN,
					CurrentScene: "浴室",
					Difficulty:   "normal",
					RuleWarnings: turn.currentWarnings,
					TurnNumber:   turn.turn,
				},
				ActiveRules: rules,
			}

			response, err := agent.InvokeJudge(ctx, request)

			require.NoError(t, err)
			require.NotNil(t, response)

			assert.Equal(t, turn.expectImpact, response.ImpactLevel,
				"Turn %d: Impact level mismatch", turn.turn)

			t.Logf("Turn %d: '%s' → %s (HP: %d, SAN: %d → SAN: %d)",
				turn.turn,
				turn.choice,
				response.ImpactLevel,
				turn.currentHP,
				turn.currentSAN,
				turn.currentSAN+response.SuggestedStateChanges.SAN)

			// Log violations
			if len(response.RulesViolated) > 0 {
				for _, v := range response.RulesViolated {
					t.Logf("  Violated: %s (%s)", v.RuleName, v.Severity)
				}
			}
		})
	}
}
