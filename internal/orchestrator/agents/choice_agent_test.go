package agents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==========================================================================
// Story 6-6: Choice Agent Tests
// ==========================================================================

// TestNewChoiceAgent tests ChoiceAgent creation
func TestNewChoiceAgent(t *testing.T) {
	mockLLM := &MockLLMClient{}

	config := AgentConfig{
		Name:       "ChoiceAgent",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockLLM,
	}

	agent := NewChoiceAgent(config)

	require.NotNil(t, agent)
	assert.Equal(t, "ChoiceAgent", agent.Config.Name)
	assert.Equal(t, 10*time.Second, agent.Config.Timeout)
	assert.NotNil(t, agent.templates)
}

// TestNewChoiceAgent_WithDefaults tests ChoiceAgent with default config
func TestNewChoiceAgent_WithDefaults(t *testing.T) {
	config := AgentConfig{
		LLMClient: &MockLLMClient{},
	}

	agent := NewChoiceAgent(config)

	require.NotNil(t, agent)
	assert.Equal(t, "ChoiceAgent", agent.Config.Name)
	assert.Equal(t, 30*time.Second, agent.Config.Timeout)
}

// ==========================================================================
// Type Tests
// ==========================================================================

// TestSceneType_String tests SceneType string representation
func TestSceneType_String(t *testing.T) {
	tests := []struct {
		sceneType SceneType
		expected  string
	}{
		{SceneExplore, "explore"},
		{SceneDialogue, "dialogue"},
		{SceneCombat, "combat"},
		{SceneEscape, "escape"},
		{SceneType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.sceneType.String())
		})
	}
}

// TestRiskLevel_String tests RiskLevel string representation
func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		risk     RiskLevel
		expected string
	}{
		{RiskSafe, "Safe"},
		{RiskWarning, "Warning"},
		{RiskDanger, "Danger"},
		{RiskLethal, "Lethal"},
		{RiskLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.risk.String())
		})
	}
}

// ==========================================================================
// Template Generation Tests (AC #5, #6)
// ==========================================================================

// TestGenerateFromTemplate_Success tests successful template generation
func TestGenerateFromTemplate_Success(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	request := &ChoiceRequest{
		SceneType:    SceneExplore,
		TensionLevel: 50,
		PlayerSAN:    80,
		Difficulty:   "normal",
	}

	options, err := agent.generateFromTemplate(request)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(options), 2, "Should generate at least 2 options")
	assert.LessOrEqual(t, len(options), 4, "Should generate at most 4 options")

	// AC #1: Check option length ≤ 15 chars
	for i, opt := range options {
		assert.LessOrEqual(t, len([]rune(opt.Text)), 15, "Option %d text should be ≤ 15 chars", i+1)
		assert.Equal(t, i+1, opt.Index, "Option %d should have correct index", i+1)
	}
}

// TestGenerateFromTemplate_AllSceneTypes tests all scene types
func TestGenerateFromTemplate_AllSceneTypes(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	sceneTypes := []SceneType{SceneExplore, SceneDialogue, SceneCombat, SceneEscape}

	for _, sceneType := range sceneTypes {
		t.Run(sceneType.String(), func(t *testing.T) {
			request := &ChoiceRequest{
				SceneType: sceneType,
				PlayerSAN: 80,
			}

			options, err := agent.generateFromTemplate(request)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(options), 2)
			assert.LessOrEqual(t, len(options), 4)
		})
	}
}

// TestFillTemplate tests template placeholder filling
func TestFillTemplate(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	tmpl := OptionTemplate{
		Text:     "檢查{object}",
		Variants: []string{"門", "窗"},
	}

	request := &ChoiceRequest{SceneType: SceneExplore}

	text := agent.fillTemplate(tmpl, request)

	// Should contain one of the variants
	assert.True(t, text == "檢查門" || text == "檢查窗", "Should contain variant: %s", text)
}

// ==========================================================================
// Risk Assessment Tests (AC #2)
// ==========================================================================

// TestCalculateRisk_AllLevels tests all risk levels
func TestCalculateRisk_AllLevels(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	tests := []struct {
		name         string
		optionText   string
		rules        []HiddenRule
		expectedRisk RiskLevel
	}{
		{
			name:       "Safe - no matching rules",
			optionText: "檢查門",
			rules:      []HiddenRule{},
			expectedRisk: RiskSafe,
		},
		{
			name:       "Warning - minor damage",
			optionText: "觸摸鏡子",
			rules: []HiddenRule{
				{
					ID:              "R001",
					TriggerKeywords: []string{"鏡子"},
					Punishment:      RulePunishment{HPDamage: 10, SANDamage: 10},
				},
			},
			expectedRisk: RiskWarning,
		},
		{
			name:       "Danger - major damage",
			optionText: "凝視鏡子",
			rules: []HiddenRule{
				{
					ID:              "R002",
					TriggerKeywords: []string{"鏡子"},
					Punishment:      RulePunishment{HPDamage: 30, SANDamage: 25},
				},
			},
			expectedRisk: RiskDanger,
		},
		{
			name:       "Lethal - fatal rule",
			optionText: "打破鏡子",
			rules: []HiddenRule{
				{
					ID:              "R003",
					TriggerKeywords: []string{"鏡子"},
					Punishment:      RulePunishment{IsFatal: true},
				},
			},
			expectedRisk: RiskLethal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := Option{Text: tt.optionText}
			risk := agent.CalculateRisk(option, tt.rules)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

// TestCalculateRisk_MultipleRules tests risk with multiple rules
func TestCalculateRisk_MultipleRules(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	option := Option{Text: "觸摸鏡子"}
	rules := []HiddenRule{
		{
			ID:              "R001",
			TriggerKeywords: []string{"觸摸"},
			Punishment:      RulePunishment{HPDamage: 5, SANDamage: 5},
		},
		{
			ID:              "R002",
			TriggerKeywords: []string{"鏡子"},
			Punishment:      RulePunishment{HPDamage: 30, SANDamage: 30},
		},
	}

	risk := agent.CalculateRisk(option, rules)

	// Should return highest risk (Danger from R002)
	assert.Equal(t, RiskDanger, risk)
}

// ==========================================================================
// Rule Hint Tests (AC #4)
// ==========================================================================

// TestAddRuleHint_DifficultyLevels tests hint subtlety by difficulty
func TestAddRuleHint_DifficultyLevels(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	option := Option{Text: "觸摸鏡子"}
	rules := []HiddenRule{
		{
			ID:              "R001",
			TriggerKeywords: []string{"鏡子"},
			DirectHint:      "看起來很危險",
			MetaphorHint:    "有些不對勁",
		},
	}

	tests := []struct {
		difficulty string
		expectHint bool
		hintText   string
	}{
		{"easy", true, "看起來很危險"},
		{"normal", true, "有些不對勁"},
		{"hard", false, ""},
		{"hell", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			result := agent.AddRuleHint(option, rules, tt.difficulty, RiskWarning)

			if tt.expectHint {
				assert.Contains(t, result, tt.hintText)
			} else {
				assert.Equal(t, option.Text, result)
			}
		})
	}
}

// ==========================================================================
// Hallucination Tests (AC #3)
// ==========================================================================

// TestShouldAddHallucination_SANLevels tests hallucination probability
func TestShouldAddHallucination_SANLevels(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	tests := []struct {
		san               int
		minProbability    float64
		maxProbability    float64
	}{
		{san: 80, minProbability: 0.0, maxProbability: 0.0}, // No hallucination
		{san: 50, minProbability: 0.05, maxProbability: 0.15}, // ~10%
		{san: 30, minProbability: 0.20, maxProbability: 0.40}, // ~30%
		{san: 10, minProbability: 0.40, maxProbability: 0.60}, // ~50%
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("SAN_%d", tt.san), func(t *testing.T) {
			// Run multiple times to test probability
			iterations := 1000
			halluCount := 0

			for i := 0; i < iterations; i++ {
				if agent.shouldAddHallucination(tt.san) {
					halluCount++
				}
			}

			probability := float64(halluCount) / float64(iterations)
			assert.GreaterOrEqual(t, probability, tt.minProbability, "Probability too low")
			assert.LessOrEqual(t, probability, tt.maxProbability, "Probability too high")
		})
	}
}

// TestGenerateHallucinationOption tests hallucination option generation
func TestGenerateHallucinationOption(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	request := &ChoiceRequest{
		SceneType: SceneExplore,
		PlayerSAN: 10,
	}

	option := agent.generateHallucinationOption(request)

	assert.NotEmpty(t, option.Text)
	assert.Contains(t, option.Description, "[幻覺]")
	assert.LessOrEqual(t, len([]rune(option.Text)), 15, "Hallucination text should be ≤ 15 chars")
}

// ==========================================================================
// Full Integration Tests (AC #1-#6)
// ==========================================================================

// TestInvokeGenerate_TemplateSuccess tests full template-driven generation
func TestInvokeGenerate_TemplateSuccess(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	request := &ChoiceRequest{
		StoryContext: "你站在一個陰暗的走廊",
		SceneType:    SceneExplore,
		TensionLevel: 50,
		ActiveRules: []HiddenRule{
			{
				ID:              "R001",
				TriggerKeywords: []string{"鏡子"},
				Punishment:      RulePunishment{HPDamage: 10, SANDamage: 10},
				DirectHint:      "看起來危險",
			},
		},
		PlayerSAN:  80,
		Difficulty: "normal",
	}

	response, err := agent.InvokeGenerate(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #1: Check 2-4 options
	assert.GreaterOrEqual(t, len(response.Options), 2)
	assert.LessOrEqual(t, len(response.Options), 4)

	// AC #1: Check option length
	for _, opt := range response.Options {
		assert.LessOrEqual(t, len([]rune(opt.Text)), 15)
	}

	// AC #2: Check risk levels present
	assert.NotNil(t, response.RiskLevels)
	assert.Equal(t, len(response.Options), len(response.RiskLevels))

	// AC #3: Check hallucination flags present
	assert.NotNil(t, response.HallucinationFlags)
}

// TestInvokeGenerate_LLMFallback tests LLM generation fallback
func TestInvokeGenerate_LLMFallback(t *testing.T) {
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, options map[string]any) (string, error) {
			return `{
				"options": [
					{"index": 1, "text": "檢查門", "description": "Check the door"},
					{"index": 2, "text": "離開", "description": "Leave"},
					{"index": 3, "text": "等待", "description": "Wait"}
				]
			}`, nil
		},
	}

	agent := NewChoiceAgent(AgentConfig{LLMClient: mockLLM})

	request := &ChoiceRequest{
		StoryContext: "測試情境",
		SceneType:    SceneType(999), // Invalid scene type to force LLM fallback
		TensionLevel: 50,
		ActiveRules:  []HiddenRule{},
		PlayerSAN:    80,
		Difficulty:   "normal",
	}

	response, err := agent.InvokeGenerate(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Options, 3)
}

// TestInvokeGenerate_WithLowSAN tests hallucination insertion
func TestInvokeGenerate_WithLowSAN(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	// Run multiple times to increase chance of hallucination
	for i := 0; i < 10; i++ {
		request := &ChoiceRequest{
			SceneType:    SceneExplore,
			TensionLevel: 50,
			ActiveRules:  []HiddenRule{},
			PlayerSAN:    5, // Very low SAN = 50% hallucination chance
			Difficulty:   "normal",
		}

		response, err := agent.InvokeGenerate(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)

		// Check if any option is hallucination
		halluFound := false
		for idx, isHallu := range response.HallucinationFlags {
			if isHallu {
				halluFound = true
				// Verify hallucination option exists
				var foundOpt *Option
				for _, opt := range response.Options {
					if opt.Index == idx {
						foundOpt = &opt
						break
					}
				}
				assert.NotNil(t, foundOpt, "Hallucination flag set but option not found")
				break
			}
		}

		// With SAN=5, at least one iteration should have hallucination
		if halluFound {
			return
		}
	}
}

// ==========================================================================
// Performance Tests (AC #5)
// ==========================================================================

// TestTemplateGeneration_Performance tests template generation speed
func TestTemplateGeneration_Performance(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	request := &ChoiceRequest{
		SceneType: SceneExplore,
		PlayerSAN: 80,
	}

	start := time.Now()
	_, err := agent.generateFromTemplate(request)
	duration := time.Since(start)

	require.NoError(t, err)
	// AC #5: Template generation < 50ms
	assert.Less(t, duration, 50*time.Millisecond, "Template generation too slow: %v", duration)
}

// ==========================================================================
// Edge Cases
// ==========================================================================

// TestParseChoiceResponse_InvalidJSON tests error handling
func TestParseChoiceResponse_InvalidJSON(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	_, err := agent.parseChoiceResponse("{invalid json")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

// TestParseChoiceResponse_InvalidOptionCount tests option count validation
func TestParseChoiceResponse_InvalidOptionCount(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	tests := []struct {
		name     string
		response string
	}{
		{"too few", `{"options": [{"index": 1, "text": "test"}]}`},
		{"too many", `{"options": [
			{"index": 1, "text": "1"},
			{"index": 2, "text": "2"},
			{"index": 3, "text": "3"},
			{"index": 4, "text": "4"},
			{"index": 5, "text": "5"}
		]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := agent.parseChoiceResponse(tt.response)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid option count")
		})
	}
}

// TestGenerateFallbackOptions tests fallback generation
func TestGenerateFallbackOptions(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	request := &ChoiceRequest{
		SceneType:   SceneExplore,
		ActiveRules: []HiddenRule{},
		PlayerSAN:   80,
		Difficulty:  "normal",
	}

	response := agent.generateFallbackOptions(request)

	require.NotNil(t, response)
	assert.Len(t, response.Options, 3, "Fallback should provide 3 options")
	assert.NotNil(t, response.RiskLevels)
	assert.NotNil(t, response.HallucinationFlags)
}

// TestRandomSelect tests random selection utility
func TestRandomSelect(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	templates := []OptionTemplate{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "5"},
	}

	selected := agent.randomSelect(templates, 3)

	assert.Len(t, selected, 3)

	// Check uniqueness
	ids := make(map[string]bool)
	for _, tmpl := range selected {
		ids[tmpl.ID] = true
	}
	assert.Len(t, ids, 3, "Selected templates should be unique")
}

// TestRandomSelect_RequestMoreThanAvailable tests edge case
func TestRandomSelect_RequestMoreThanAvailable(t *testing.T) {
	agent := NewChoiceAgent(AgentConfig{LLMClient: &MockLLMClient{}})

	templates := []OptionTemplate{
		{ID: "1"},
		{ID: "2"},
	}

	selected := agent.randomSelect(templates, 5)

	// Should return all available
	assert.Len(t, selected, 2)
}
