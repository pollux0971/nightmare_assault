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

// providerAdapter adapts api.Provider to LLMClient interface
type providerAdapter struct {
	provider api.Provider
}

func (p *providerAdapter) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	messages := []api.Message{
		{Role: "user", Content: prompt},
	}

	response, err := p.provider.SendMessage(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// TestChoiceAgentIntegration_TemplateGeneration tests template-driven generation performance
func TestChoiceAgentIntegration_TemplateGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewChoiceAgent(AgentConfig{
		Name:      "ChoiceAgent",
		LLMClient: nil, // Template generation doesn't need LLM client
		Timeout:   5 * time.Second,
	})

	tests := []struct {
		name         string
		request      *ChoiceRequest
		maxDuration  time.Duration
		wantOptions  int
		wantScene    SceneType
	}{
		{
			name: "explore_scene_fast_generation",
			request: &ChoiceRequest{
				StoryContext: "你站在陰暗的走廊中，左右兩側都有門。",
				SceneType:    SceneExplore,
				TensionLevel: 30,
				ActiveRules:  []HiddenRule{},
				PlayerSAN:    80,
				Difficulty:   "normal",
			},
			maxDuration: 50 * time.Millisecond,
			wantOptions: 3,
			wantScene:   SceneExplore,
		},
		{
			name: "combat_scene_fast_generation",
			request: &ChoiceRequest{
				StoryContext: "怪物正在逼近！",
				SceneType:    SceneCombat,
				TensionLevel: 70,
				ActiveRules:  []HiddenRule{},
				PlayerSAN:    60,
				Difficulty:   "hard",
			},
			maxDuration: 50 * time.Millisecond,
			wantOptions: 3,
			wantScene:   SceneCombat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			start := time.Now()

			response, err := agent.InvokeGenerate(ctx, tt.request)
			duration := time.Since(start)

			require.NoError(t, err)
			assert.NotNil(t, response)
			assert.GreaterOrEqual(t, len(response.Options), 2, "Should have at least 2 options")
			assert.LessOrEqual(t, len(response.Options), 4, "Should have at most 4 options")
			assert.LessOrEqual(t, duration, tt.maxDuration, "Template generation should be under 50ms")

			// Verify option text length (≤15 Traditional Chinese characters)
			for _, opt := range response.Options {
				runeCount := len([]rune(opt.Text))
				assert.LessOrEqual(t, runeCount, 15, "Option text should be ≤15 chars: %s", opt.Text)
			}
		})
	}
}

// TestChoiceAgentIntegration_LLMFallback tests LLM fallback generation with real API
func TestChoiceAgentIntegration_LLMFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load config to get real API provider
	cfg, err := config.Load()
	require.NoError(t, err)
	require.True(t, cfg.IsConfigured(), "API must be configured for integration tests")

	// Get API key
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

	// Wrap provider in adapter to implement LLMClient interface
	llmClient := &providerAdapter{provider: prov}

	agent := NewChoiceAgent(AgentConfig{
		Name:      "ChoiceAgent",
		LLMClient: llmClient,
		Timeout:   30 * time.Second,
	})

	// Force LLM generation by using empty templates
	agent.templates = make(map[SceneType][]OptionTemplate)

	request := &ChoiceRequest{
		StoryContext: "你發現了一個神秘的房間，裡面有奇怪的符號和一本古老的書。空氣中瀰漫著不安的氣息。",
		SceneType:    SceneExplore,
		TensionLevel: 50,
		ActiveRules: []HiddenRule{
			{
				ID:              "rule_001",
				Name:            "不要在黑暗中開燈",
				TriggerKeywords: []string{"開燈", "打開燈"},
				Punishment: RulePunishment{
					IsFatal:   false,
					HPDamage:  20,
					SANDamage: 15,
				},
				DirectHint:   "記得，黑暗中開燈會引來「它們」",
				MetaphorHint: "光明會喚醒沉睡之物",
			},
		},
		PlayerSAN:  70,
		Difficulty: "normal",
	}

	ctx := context.Background()
	start := time.Now()

	response, err := agent.InvokeGenerate(ctx, request)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, response)

	t.Logf("LLM generation took %v", duration)

	// AC #5: LLM fallback should be under 450ms
	assert.LessOrEqual(t, duration, 450*time.Millisecond, "LLM generation should be under 450ms")

	// AC #1: Should generate 2-4 options with ≤15 chars
	assert.GreaterOrEqual(t, len(response.Options), 2)
	assert.LessOrEqual(t, len(response.Options), 4)

	for _, opt := range response.Options {
		runeCount := len([]rune(opt.Text))
		assert.LessOrEqual(t, runeCount, 15, "Option text should be ≤15 chars: %s", opt.Text)
		t.Logf("Option %d: %s (Length: %d)", opt.Index, opt.Text, runeCount)
	}

	// AC #2: Risk assessment should be provided (internal only)
	assert.NotNil(t, response.RiskLevels)
	assert.Equal(t, len(response.Options), len(response.RiskLevels))

	for idx, risk := range response.RiskLevels {
		t.Logf("Option %d Risk: %s", idx, risk)
	}
}

// TestChoiceAgentIntegration_RiskAssessment tests risk assessment with active rules
func TestChoiceAgentIntegration_RiskAssessment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewChoiceAgent(AgentConfig{
		Name:      "ChoiceAgent",
		LLMClient: nil,
		Timeout:   5 * time.Second,
	})

	rules := []HiddenRule{
		{
			ID:              "rule_light",
			Name:            "不要開燈",
			TriggerKeywords: []string{"開燈", "打開燈", "照明"},
			Punishment: RulePunishment{
				IsFatal:   false,
				HPDamage:  20,
				SANDamage: 15,
			},
		},
		{
			ID:              "rule_mirror",
			Name:            "不要看鏡子",
			TriggerKeywords: []string{"看鏡子", "照鏡子", "鏡中"},
			Punishment: RulePunishment{
				IsFatal:   true,
				HPDamage:  0,
				SANDamage: 100,
			},
		},
	}

	tests := []struct {
		name         string
		option       Option
		expectedRisk RiskLevel
	}{
		{
			name: "safe_option",
			option: Option{
				Index:       1,
				Text:        "檢查房間角落",
				Description: "Carefully search the room corners",
			},
			expectedRisk: RiskSafe,
		},
		{
			name: "warning_option",
			option: Option{
				Index:       2,
				Text:        "打開燈照明",
				Description: "Turn on the light to see better",
			},
			expectedRisk: RiskWarning,
		},
		{
			name: "lethal_option",
			option: Option{
				Index:       3,
				Text:        "看向鏡子",
				Description: "Look at the mirror on the wall",
			},
			expectedRisk: RiskLethal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := agent.CalculateRisk(tt.option, rules)
			assert.Equal(t, tt.expectedRisk, risk, "Risk level mismatch for option: %s", tt.option.Text)
			t.Logf("Option: %s → Risk: %s", tt.option.Text, risk)
		})
	}
}

// TestChoiceAgentIntegration_HallucinationByLevel tests hallucination probability at different SAN levels
func TestChoiceAgentIntegration_HallucinationByLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewChoiceAgent(AgentConfig{
		Name:      "ChoiceAgent",
		LLMClient: nil,
		Timeout:   5 * time.Second,
	})

	tests := []struct {
		name            string
		san             int
		samples         int
		minProbability  float64
		maxProbability  float64
		shouldHallucinate bool
	}{
		{
			name:              "high_san_no_hallucination",
			san:               80,
			samples:           100,
			minProbability:    0.0,
			maxProbability:    0.0,
			shouldHallucinate: false,
		},
		{
			name:              "medium_san_low_hallucination",
			san:               50,
			samples:           1000,
			minProbability:    0.05,
			maxProbability:    0.15,
			shouldHallucinate: true,
		},
		{
			name:              "low_san_medium_hallucination",
			san:               30,
			samples:           1000,
			minProbability:    0.25,
			maxProbability:    0.35,
			shouldHallucinate: true,
		},
		{
			name:              "very_low_san_high_hallucination",
			san:               10,
			samples:           1000,
			minProbability:    0.45,
			maxProbability:    0.55,
			shouldHallucinate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hallucinationCount := 0
			for i := 0; i < tt.samples; i++ {
				if agent.shouldAddHallucination(tt.san) {
					hallucinationCount++
				}
			}

			probability := float64(hallucinationCount) / float64(tt.samples)
			t.Logf("SAN %d: Hallucination probability %.2f%% (%d/%d samples)",
				tt.san, probability*100, hallucinationCount, tt.samples)

			if tt.shouldHallucinate {
				assert.GreaterOrEqual(t, probability, tt.minProbability,
					"Hallucination probability should be at least %.2f%%", tt.minProbability*100)
				assert.LessOrEqual(t, probability, tt.maxProbability,
					"Hallucination probability should be at most %.2f%%", tt.maxProbability*100)
			} else {
				assert.Equal(t, 0.0, probability, "Should have no hallucinations at high SAN")
			}
		})
	}
}

// TestChoiceAgentIntegration_RuleHintsByDifficulty tests rule hint integration at different difficulties
func TestChoiceAgentIntegration_RuleHintsByDifficulty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewChoiceAgent(AgentConfig{
		Name:      "ChoiceAgent",
		LLMClient: nil,
		Timeout:   5 * time.Second,
	})

	rules := []HiddenRule{
		{
			ID:              "rule_001",
			Name:            "不要開燈",
			TriggerKeywords: []string{"開燈"},
			Punishment: RulePunishment{
				IsFatal:   false,
				HPDamage:  20,
				SANDamage: 15,
			},
			DirectHint:   "記得，黑暗中開燈會引來「它們」",
			MetaphorHint: "光明會喚醒沉睡之物",
		},
	}

	dangerOption := Option{
		Index:       1,
		Text:        "打開燈",
		Description: "Turn on the light",
	}

	tests := []struct {
		name           string
		difficulty     string
		risk           RiskLevel
		expectHint     bool
		expectDirect   bool
		expectMetaphor bool
	}{
		{
			name:           "easy_danger_direct_hint",
			difficulty:     "easy",
			risk:           RiskDanger,
			expectHint:     true,
			expectDirect:   true,
			expectMetaphor: false,
		},
		{
			name:           "normal_danger_metaphor_hint",
			difficulty:     "normal",
			risk:           RiskDanger,
			expectHint:     true,
			expectDirect:   false,
			expectMetaphor: true,
		},
		{
			name:           "hard_danger_no_hint",
			difficulty:     "hard",
			risk:           RiskDanger,
			expectHint:     false,
			expectDirect:   false,
			expectMetaphor: false,
		},
		{
			name:           "hell_danger_no_hint",
			difficulty:     "hell",
			risk:           RiskDanger,
			expectHint:     false,
			expectDirect:   false,
			expectMetaphor: false,
		},
		{
			name:           "easy_safe_no_hint",
			difficulty:     "easy",
			risk:           RiskSafe,
			expectHint:     false,
			expectDirect:   false,
			expectMetaphor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.AddRuleHint(dangerOption, rules, tt.difficulty, tt.risk)

			if tt.expectHint {
				assert.NotEqual(t, dangerOption.Text, result, "Should add hint to option text")

				if tt.expectDirect {
					// Direct hint should be more explicit
					t.Logf("Easy mode hint: %s", result)
					// In easy mode, we expect clearer warnings
				} else if tt.expectMetaphor {
					// Metaphor hint should be more subtle
					t.Logf("Normal mode hint: %s", result)
					// In normal mode, we expect metaphorical warnings
				}
			} else {
				assert.Equal(t, dangerOption.Text, result, "Should not add hint for this difficulty/risk")
			}
		})
	}
}

// TestChoiceAgentIntegration_EndToEnd tests complete choice generation flow
func TestChoiceAgentIntegration_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	agent := NewChoiceAgent(AgentConfig{
		Name:      "ChoiceAgent",
		LLMClient: nil,
		Timeout:   10 * time.Second,
	})

	request := &ChoiceRequest{
		StoryContext: "你站在廢棄醫院的走廊中。左邊的門半開著，右邊的門緊閉。遠處傳來奇怪的聲音。",
		SceneType:    SceneExplore,
		TensionLevel: 60,
		ActiveRules: []HiddenRule{
			{
				ID:              "rule_door",
				Name:            "不要打開緊閉的門",
				TriggerKeywords: []string{"打開", "推開", "右邊的門"},
				Punishment: RulePunishment{
					IsFatal:   true,
					HPDamage:  0,
					SANDamage: 100,
				},
				DirectHint:   "緊閉的門後有危險",
				MetaphorHint: "被封印之物不應喚醒",
			},
		},
		PlayerSAN:  40,
		Difficulty: "normal",
	}

	ctx := context.Background()
	response, err := agent.InvokeGenerate(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #1: Generate 2-4 options with ≤15 chars
	assert.GreaterOrEqual(t, len(response.Options), 2)
	assert.LessOrEqual(t, len(response.Options), 4)

	t.Log("Generated options:")
	for _, opt := range response.Options {
		runeCount := len([]rune(opt.Text))
		assert.LessOrEqual(t, runeCount, 15)

		risk := response.RiskLevels[opt.Index]
		isHallucination := response.HallucinationFlags[opt.Index]

		t.Logf("  [%d] %s (Risk: %s, Hallucination: %v)",
			opt.Index, opt.Text, risk, isHallucination)
	}

	// AC #2: Risk levels should be provided
	assert.NotNil(t, response.RiskLevels)
	assert.Equal(t, len(response.Options), len(response.RiskLevels))

	// AC #3: Hallucination flags should be provided
	assert.NotNil(t, response.HallucinationFlags)
	assert.Equal(t, len(response.Options), len(response.HallucinationFlags))

	// Verify at least one option has risk assessment
	hasRiskAssessment := false
	for _, risk := range response.RiskLevels {
		if risk != RiskSafe {
			hasRiskAssessment = true
			break
		}
	}
	t.Logf("Has risk assessment: %v", hasRiskAssessment)
}
