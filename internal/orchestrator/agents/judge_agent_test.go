package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewJudgeAgent tests JudgeAgent creation
func TestNewJudgeAgent(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		Name:       "TestJudgeAgent",
		Timeout:    10 * time.Second,
		MaxRetries: 2,
	})

	assert.NotNil(t, agent)
	assert.Equal(t, "TestJudgeAgent", agent.Config.Name)
	assert.Equal(t, 10*time.Second, agent.Config.Timeout)
	assert.Equal(t, 2, agent.Config.MaxRetries)
}

// TestNewJudgeAgent_WithDefaults tests default values
func TestNewJudgeAgent_WithDefaults(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	assert.NotNil(t, agent)
	assert.Equal(t, "JudgeAgent", agent.Config.Name)
	assert.Equal(t, 30*time.Second, agent.Config.Timeout)
	assert.Equal(t, 3, agent.Config.MaxRetries)
}

// TestCheckRuleViolation_SingleRule tests single rule violation
func TestCheckRuleViolation_SingleRule(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	rules := []JudgeHiddenRule{
		{
			ID:               "R-001",
			Name:             "倒影殺手",
			Type:             RuleTypeScene,
			TriggerKeywords:  []string{"鏡子", "倒影", "凝視"},
			TriggerCondition: "凝視鏡子超過3秒",
			Punishment: RulePunishment{
				IsFatal:   false,
				HPDamage:  0,
				SANDamage: 40,
			},
		},
	}

	tests := []struct {
		name           string
		choice         string
		expectViolated bool
	}{
		{
			name:           "violates_mirror_rule",
			choice:         "我決定凝視鏡子",
			expectViolated: true,
		},
		{
			name:           "violates_mirror_rule_with_reflection",
			choice:         "看著倒影中的自己",
			expectViolated: true,
		},
		{
			name:           "no_violation",
			choice:         "我決定走向門口",
			expectViolated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := agent.CheckRuleViolation(tt.choice, rules)

			if tt.expectViolated {
				assert.Len(t, violations, 1, "Should detect violation")
				assert.Equal(t, "R-001", violations[0].RuleID)
				assert.Equal(t, "倒影殺手", violations[0].RuleName)
				assert.Equal(t, 40, violations[0].SANDamage)
			} else {
				assert.Len(t, violations, 0, "Should not detect violation")
			}
		})
	}
}

// TestCheckRuleViolation_MultipleRules tests multiple rule violations
func TestCheckRuleViolation_MultipleRules(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	rules := []JudgeHiddenRule{
		{
			ID:               "R-001",
			Name:             "倒影殺手",
			Type:             RuleTypeScene,
			TriggerKeywords:  []string{"鏡子", "倒影"},
			TriggerCondition: "凝視鏡子",
			Punishment: RulePunishment{
				IsFatal:   false,
				HPDamage:  0,
				SANDamage: 40,
			},
		},
		{
			ID:               "R-002",
			Name:             "聲音禁忌",
			Type:             RuleTypeBehavior,
			TriggerKeywords:  []string{"呼喊", "大聲", "尖叫"},
			TriggerCondition: "發出大聲響",
			Punishment: RulePunishment{
				IsFatal:   true,
				HPDamage:  100,
				SANDamage: 0,
			},
		},
	}

	choice := "我凝視鏡子並大聲呼喊"
	violations := agent.CheckRuleViolation(choice, rules)

	assert.Len(t, violations, 2, "Should detect both violations")

	// AC #4: Check priority sorting (Scene > Behavior)
	assert.Equal(t, "R-001", violations[0].RuleID, "Scene rule should be first")
	assert.Equal(t, RuleTypeScene, violations[0].RuleType)

	assert.Equal(t, "R-002", violations[1].RuleID, "Behavior rule should be second")
	assert.Equal(t, RuleTypeBehavior, violations[1].RuleType)
}

// TestGetRulePriority tests rule priority ordering
func TestGetRulePriority(t *testing.T) {
	tests := []struct {
		ruleType RuleType
		expected int
	}{
		{RuleTypeScene, 4},
		{RuleTypeTime, 3},
		{RuleTypeBehavior, 2},
		{RuleTypeState, 1},
	}

	for _, tt := range tests {
		t.Run(tt.ruleType.String(), func(t *testing.T) {
			priority := GetRulePriority(tt.ruleType)
			assert.Equal(t, tt.expected, priority)
		})
	}

	// Verify ordering
	assert.Greater(t, GetRulePriority(RuleTypeScene), GetRulePriority(RuleTypeTime))
	assert.Greater(t, GetRulePriority(RuleTypeTime), GetRulePriority(RuleTypeBehavior))
	assert.Greater(t, GetRulePriority(RuleTypeBehavior), GetRulePriority(RuleTypeState))
}

// TestDetermineImpactLevel tests impact level determination
func TestDetermineImpactLevel(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		name       string
		violations []RuleViolation
		expected   ImpactLevel
	}{
		{
			name:       "no_violations",
			violations: []RuleViolation{},
			expected:   ImpactNone,
		},
		{
			name: "minor_violation",
			violations: []RuleViolation{
				{Severity: ImpactMinor, IsFatal: false},
			},
			expected: ImpactMinor,
		},
		{
			name: "moderate_violation",
			violations: []RuleViolation{
				{Severity: ImpactModerate, IsFatal: false},
			},
			expected: ImpactModerate,
		},
		{
			name: "major_violation",
			violations: []RuleViolation{
				{Severity: ImpactMajor, IsFatal: false},
			},
			expected: ImpactMajor,
		},
		{
			name: "lethal_violation",
			violations: []RuleViolation{
				{Severity: ImpactLethal, IsFatal: true},
			},
			expected: ImpactLethal,
		},
		{
			name: "multiple_violations_take_highest",
			violations: []RuleViolation{
				{Severity: ImpactMinor, IsFatal: false},
				{Severity: ImpactMajor, IsFatal: false},
				{Severity: ImpactModerate, IsFatal: false},
			},
			expected: ImpactMajor,
		},
		{
			name: "fatal_overrides_all",
			violations: []RuleViolation{
				{Severity: ImpactMinor, IsFatal: false},
				{Severity: ImpactModerate, IsFatal: true},
			},
			expected: ImpactLethal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.DetermineImpactLevel(tt.violations)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCalculateStateChanges tests HP/SAN calculation
func TestCalculateStateChanges(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		name            string
		violations      []RuleViolation
		difficulty      string
		currentWarnings map[string]int
		expectHP        int
		expectSAN       int
	}{
		{
			name: "first_warning_easy_half_damage",
			violations: []RuleViolation{
				{
					RuleID:    "R-001",
					HPDamage:  20,
					SANDamage: 40,
					IsFatal:   false,
				},
			},
			difficulty:      "easy",
			currentWarnings: map[string]int{},
			expectHP:        -8,  // 20 * 0.5 * 0.8 = 8
			expectSAN:       -16, // 40 * 0.5 * 0.8 = 16
		},
		{
			name: "no_warning_normal_full_damage",
			violations: []RuleViolation{
				{
					RuleID:    "R-001",
					HPDamage:  20,
					SANDamage: 40,
					IsFatal:   false,
				},
			},
			difficulty:      "normal",
			currentWarnings: map[string]int{"R-001": 0},
			expectHP:        -20, // 20 * 1.0
			expectSAN:       -40, // 40 * 1.0
		},
		{
			name: "hell_difficulty_multiplier",
			violations: []RuleViolation{
				{
					RuleID:    "R-001",
					HPDamage:  20,
					SANDamage: 40,
					IsFatal:   false,
				},
			},
			difficulty:      "hell",
			currentWarnings: map[string]int{"R-001": 0},
			expectHP:        -24, // 20 * 1.2
			expectSAN:       -48, // 40 * 1.2
		},
		{
			name: "fatal_ignores_warnings",
			violations: []RuleViolation{
				{
					RuleID:    "R-002",
					HPDamage:  100,
					SANDamage: 0,
					IsFatal:   true,
				},
			},
			difficulty:      "easy",
			currentWarnings: map[string]int{"R-002": 2},
			expectHP:        -80, // 100 * 0.8 (fatal ignores warning, but difficulty applies)
			expectSAN:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := agent.CalculateStateChanges(
				tt.violations,
				tt.difficulty,
				tt.currentWarnings,
			)

			assert.Equal(t, tt.expectHP, changes.HP, "HP change mismatch")
			assert.Equal(t, tt.expectSAN, changes.SAN, "SAN change mismatch")
		})
	}
}

// TestGetDifficultyMultiplier tests difficulty multipliers
func TestGetDifficultyMultiplier(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		difficulty string
		expected   float64
	}{
		{"easy", 0.8},
		{"normal", 1.0},
		{"hard", 1.0},
		{"hell", 1.2},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			result := agent.getDifficultyMultiplier(tt.difficulty)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetMaxWarnings tests warning limits by difficulty
func TestGetMaxWarnings(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		difficulty string
		expected   int
	}{
		{"easy", 2},
		{"normal", 1},
		{"hard", 1},
		{"hell", 0},
		{"unknown", 1},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			result := agent.getMaxWarnings(tt.difficulty)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGenerateDeathReason tests death reason generation
func TestGenerateDeathReason(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		name        string
		violation   RuleViolation
		choice      string
		expectEmpty bool
	}{
		{
			name: "fatal_violation_generates_reason",
			violation: RuleViolation{
				RuleID:   "R-002",
				RuleName: "聲音禁忌",
				IsFatal:  true,
				Reason:   "在黑暗中發出聲音",
			},
			choice:      "大聲呼喊",
			expectEmpty: false,
		},
		{
			name: "non_fatal_returns_empty",
			violation: RuleViolation{
				RuleID:   "R-001",
				RuleName: "倒影殺手",
				IsFatal:  false,
			},
			choice:      "凝視鏡子",
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := agent.GenerateDeathReason(tt.violation, tt.choice)

			if tt.expectEmpty {
				assert.Empty(t, reason)
			} else {
				assert.NotEmpty(t, reason)
				assert.Contains(t, reason, tt.violation.RuleName)
				assert.Contains(t, reason, tt.choice)
			}
		})
	}
}

// TestDetermineNextAction tests next action routing
func TestDetermineNextAction(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		name        string
		impactLevel ImpactLevel
		expected    NextActionType
	}{
		{
			name:        "none_continues_story",
			impactLevel: ImpactNone,
			expected:    ActionContinueStory,
		},
		{
			name:        "minor_continues_story",
			impactLevel: ImpactMinor,
			expected:    ActionContinueStory,
		},
		{
			name:        "moderate_applies_damage",
			impactLevel: ImpactModerate,
			expected:    ActionApplyDamage,
		},
		{
			name:        "major_applies_damage",
			impactLevel: ImpactMajor,
			expected:    ActionApplyDamage,
		},
		{
			name:        "lethal_triggers_death",
			impactLevel: ImpactLethal,
			expected:    ActionTriggerDeath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.determineNextAction(tt.impactLevel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestInvokeJudge_NoViolation tests judgment with no violations
func TestInvokeJudge_NoViolation(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		LLMClient: nil, // No LLM client
	})

	request := &JudgeRequest{
		PlayerChoice: "我決定離開房間",
		GameState: &GameStateSnapshot{
			HP:           100,
			SAN:          80,
			CurrentScene: "走廊",
			Difficulty:   "normal",
			RuleWarnings: make(map[string]int),
		},
		ActiveRules: []JudgeHiddenRule{
			{
				ID:              "R-001",
				Name:            "倒影殺手",
				TriggerKeywords: []string{"鏡子", "倒影"},
				Punishment: RulePunishment{
					IsFatal:   false,
					SANDamage: 40,
				},
			},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeJudge(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, ImpactNone, response.ImpactLevel)
	assert.Len(t, response.RulesViolated, 0)
	assert.Equal(t, ActionContinueStory, response.NextAction)
	assert.Empty(t, response.DeathReason)
}

// TestInvokeJudge_ModerateViolation tests moderate violation judgment
func TestInvokeJudge_ModerateViolation(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		LLMClient: nil,
	})

	request := &JudgeRequest{
		PlayerChoice: "我凝視鏡子中的倒影",
		GameState: &GameStateSnapshot{
			HP:           100,
			SAN:          80,
			CurrentScene: "浴室",
			Difficulty:   "normal",
			RuleWarnings: make(map[string]int),
		},
		ActiveRules: []JudgeHiddenRule{
			{
				ID:               "R-001",
				Name:             "倒影殺手",
				Type:             RuleTypeScene,
				TriggerKeywords:  []string{"鏡子", "倒影", "凝視"},
				TriggerCondition: "凝視鏡子超過3秒",
				Punishment: RulePunishment{
					IsFatal:   false,
					HPDamage:  0,
					SANDamage: 40,
				},
			},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeJudge(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, ImpactModerate, response.ImpactLevel)
	assert.Len(t, response.RulesViolated, 1)
	assert.Equal(t, "R-001", response.RulesViolated[0].RuleID)
	assert.Equal(t, ActionApplyDamage, response.NextAction)
	assert.Empty(t, response.DeathReason)

	// Check state changes
	assert.Equal(t, 0, response.SuggestedStateChanges.HP)
	assert.Equal(t, -20, response.SuggestedStateChanges.SAN) // Half damage with warning
}

// TestInvokeJudge_LethalViolation tests lethal violation judgment
func TestInvokeJudge_LethalViolation(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		LLMClient: nil,
	})

	request := &JudgeRequest{
		PlayerChoice: "我大聲呼喊求救",
		GameState: &GameStateSnapshot{
			HP:           100,
			SAN:          80,
			CurrentScene: "黑暗走廊",
			Difficulty:   "hell",
			RuleWarnings: make(map[string]int),
		},
		ActiveRules: []JudgeHiddenRule{
			{
				ID:               "R-002",
				Name:             "聲音禁忌",
				Type:             RuleTypeBehavior,
				TriggerKeywords:  []string{"呼喊", "大聲", "尖叫"},
				TriggerCondition: "在黑暗中發出聲音",
				Punishment: RulePunishment{
					IsFatal:   true,
					HPDamage:  100,
					SANDamage: 0,
				},
			},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeJudge(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, ImpactLethal, response.ImpactLevel)
	assert.Len(t, response.RulesViolated, 1)
	assert.Equal(t, "R-002", response.RulesViolated[0].RuleID)
	assert.Equal(t, ActionTriggerDeath, response.NextAction)
	assert.NotEmpty(t, response.DeathReason)
	assert.Contains(t, response.DeathReason, "聲音禁忌")

	// Check state changes (hell difficulty, no warnings for fatal)
	assert.Equal(t, -120, response.SuggestedStateChanges.HP) // 100 * 1.2
}

// TestParseImpactLevel tests impact level string parsing
func TestParseImpactLevel(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		input    string
		expected ImpactLevel
	}{
		{"None", ImpactNone},
		{"none", ImpactNone},
		{"Minor", ImpactMinor},
		{"Moderate", ImpactModerate},
		{"Major", ImpactMajor},
		{"Lethal", ImpactLethal},
		{"unknown", ImpactNone},
		{"  Minor  ", ImpactMinor},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := agent.parseImpactLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==========================================================================
// H-7: Comprehensive Edge Case Tests
// ==========================================================================

// TestCalculateStateChanges_MultipleViolationsCumulative tests cumulative damage
func TestCalculateStateChanges_MultipleViolationsCumulative(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	tests := []struct {
		name            string
		violations      []RuleViolation
		difficulty      string
		currentWarnings map[string]int
		expectHP        int
		expectSAN       int
		description     string
	}{
		{
			name: "three_violations_cumulative_normal",
			violations: []RuleViolation{
				{RuleID: "R-001", HPDamage: 10, SANDamage: 20, IsFatal: false},
				{RuleID: "R-002", HPDamage: 15, SANDamage: 25, IsFatal: false},
				{RuleID: "R-003", HPDamage: 20, SANDamage: 30, IsFatal: false},
			},
			difficulty:      "normal",
			currentWarnings: map[string]int{},
			expectHP:        -22, // (10+15+20) * 0.5 (all have warnings)
			expectSAN:       -37, // (20+25+30) * 0.5
			description:     "Three violations with warnings should apply half damage each",
		},
		{
			name: "five_violations_mixed_warnings",
			violations: []RuleViolation{
				{RuleID: "R-001", HPDamage: 10, SANDamage: 10, IsFatal: false},
				{RuleID: "R-002", HPDamage: 10, SANDamage: 10, IsFatal: false},
				{RuleID: "R-003", HPDamage: 10, SANDamage: 10, IsFatal: false},
				{RuleID: "R-004", HPDamage: 10, SANDamage: 10, IsFatal: false},
				{RuleID: "R-005", HPDamage: 10, SANDamage: 10, IsFatal: false},
			},
			difficulty: "normal",
			currentWarnings: map[string]int{
				"R-001": 0, // No warning
				"R-002": 1, // Has warning
				"R-003": 0, // No warning
				// R-004 and R-005 uninitialized (will get warnings)
			},
			expectHP:  -35, // R-001(10) + R-002(5) + R-003(10) + R-004(5) + R-005(5)
			expectSAN: -35, // Same
			description: "Five violations with mixed warning states",
		},
		{
			name: "high_damage_hell_difficulty",
			violations: []RuleViolation{
				{RuleID: "R-001", HPDamage: 40, SANDamage: 60, IsFatal: false},
				{RuleID: "R-002", HPDamage: 50, SANDamage: 70, IsFatal: false},
			},
			difficulty:      "hell",
			currentWarnings: map[string]int{"R-001": 0, "R-002": 0},
			expectHP:        -108, // (40+50) * 1.2
			expectSAN:       -156, // (60+70) * 1.2
			description:     "High damage with hell multiplier should accumulate correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := agent.CalculateStateChanges(
				tt.violations,
				tt.difficulty,
				tt.currentWarnings,
			)

			assert.Equal(t, tt.expectHP, changes.HP, "HP: %s", tt.description)
			assert.Equal(t, tt.expectSAN, changes.SAN, "SAN: %s", tt.description)

			t.Logf("%s: HP=%d, SAN=%d", tt.name, changes.HP, changes.SAN)
		})
	}
}

// TestCalculateStateChanges_WarningTransition tests warning system transitions
func TestCalculateStateChanges_WarningTransition(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	violation := []RuleViolation{
		{RuleID: "R-001", HPDamage: 20, SANDamage: 40, IsFatal: false},
	}

	// Easy difficulty: 2 warnings
	// First violation (2 warnings remaining -> 1 warning remaining)
	changes1 := agent.CalculateStateChanges(violation, "easy", map[string]int{})
	assert.Equal(t, -8, changes1.HP, "First violation with 2 warnings: HP = 20 * 0.5 * 0.8")
	assert.Equal(t, -16, changes1.SAN, "First violation with 2 warnings: SAN = 40 * 0.5 * 0.8")
	assert.Equal(t, 1, changes1.WarningsRemaining["R-001"], "Should have 1 warning left")

	// Second violation (1 warning remaining -> 0 warnings remaining)
	changes2 := agent.CalculateStateChanges(violation, "easy", map[string]int{"R-001": 1})
	assert.Equal(t, -8, changes2.HP, "Second violation with 1 warning: HP = 20 * 0.5 * 0.8")
	assert.Equal(t, -16, changes2.SAN, "Second violation with 1 warning: SAN = 40 * 0.5 * 0.8")
	assert.Equal(t, 0, changes2.WarningsRemaining["R-001"], "Should have 0 warnings left")

	// Third violation (0 warnings remaining -> full damage)
	changes3 := agent.CalculateStateChanges(violation, "easy", map[string]int{"R-001": 0})
	assert.Equal(t, -16, changes3.HP, "Third violation with 0 warnings: HP = 20 * 0.8 (full damage)")
	assert.Equal(t, -32, changes3.SAN, "Third violation with 0 warnings: SAN = 40 * 0.8 (full damage)")
	assert.Equal(t, 0, changes3.WarningsRemaining["R-001"], "Should still have 0 warnings")

	t.Logf("Warning transition: 1st=%d HP, 2nd=%d HP, 3rd=%d HP (full damage)",
		changes1.HP, changes2.HP, changes3.HP)
}

// TestInvokeJudge_MixedFatalAndNonFatal tests multiple violations with mixed fatality
func TestInvokeJudge_MixedFatalAndNonFatal(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		LLMClient: nil,
	})

	request := &JudgeRequest{
		PlayerChoice: "我在鏡子前大聲呼喊",
		GameState: &GameStateSnapshot{
			HP:           100,
			SAN:          80,
			CurrentScene: "浴室",
			Difficulty:   "normal",
			RuleWarnings: make(map[string]int),
		},
		ActiveRules: []JudgeHiddenRule{
			{
				ID:               "R-001",
				Name:             "倒影殺手",
				Type:             RuleTypeScene,
				TriggerKeywords:  []string{"鏡子", "倒影"},
				TriggerCondition: "凝視鏡子",
				Punishment: RulePunishment{
					IsFatal:   false,
					HPDamage:  0,
					SANDamage: 40,
				},
			},
			{
				ID:               "R-002",
				Name:             "聲音禁忌",
				Type:             RuleTypeBehavior,
				TriggerKeywords:  []string{"呼喊", "大聲", "尖叫"},
				TriggerCondition: "發出大聲響",
				Punishment: RulePunishment{
					IsFatal:   true,
					HPDamage:  100,
					SANDamage: 0,
				},
			},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeJudge(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// AC #3: Fatal violation should override non-fatal
	assert.Equal(t, ImpactLethal, response.ImpactLevel, "Fatal violation should result in Lethal impact")
	assert.Len(t, response.RulesViolated, 2, "Should detect both violations")
	assert.Equal(t, ActionTriggerDeath, response.NextAction, "Should trigger death")
	assert.NotEmpty(t, response.DeathReason, "Should have death reason")

	// Check that death reason mentions the fatal rule
	assert.Contains(t, response.DeathReason, "聲音禁忌")

	t.Logf("Mixed violations result: %d violations, Impact=%s, DeathReason=%s",
		len(response.RulesViolated), response.ImpactLevel, response.DeathReason)
}

// TestInvokeJudge_ExtremeDamageValues tests very high damage values
func TestInvokeJudge_ExtremeDamageValues(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		LLMClient: nil,
	})

	tests := []struct {
		name        string
		hpDamage    int
		sanDamage   int
		difficulty  string
		description string
	}{
		{
			name:        "extreme_hp_damage_hell",
			hpDamage:    200,
			sanDamage:   0,
			difficulty:  "hell",
			description: "200 HP damage on hell should result in -240 HP",
		},
		{
			name:        "extreme_san_damage_hell",
			hpDamage:    0,
			sanDamage:   300,
			difficulty:  "hell",
			description: "300 SAN damage on hell should result in -360 SAN",
		},
		{
			name:        "both_extreme_easy",
			hpDamage:    150,
			sanDamage:   200,
			difficulty:  "easy",
			description: "Extreme damage on easy should still be very high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &JudgeRequest{
				PlayerChoice: "我觸發了超高傷害規則",
				GameState: &GameStateSnapshot{
					HP:           100,
					SAN:          80,
					CurrentScene: "危險區域",
					Difficulty:   tt.difficulty,
					RuleWarnings: map[string]int{"R-999": 0}, // No warnings
				},
				ActiveRules: []JudgeHiddenRule{
					{
						ID:               "R-999",
						Name:             "極限傷害",
						Type:             RuleTypeState,
						TriggerKeywords:  []string{"觸發"},
						TriggerCondition: "極限測試",
						Punishment: RulePunishment{
							IsFatal:   false,
							HPDamage:  tt.hpDamage,
							SANDamage: tt.sanDamage,
						},
					},
				},
			}

			ctx := context.Background()
			response, err := agent.InvokeJudge(ctx, request)

			require.NoError(t, err)
			assert.NotNil(t, response)

			// Log the actual damage values
			t.Logf("%s: %s", tt.name, tt.description)
			t.Logf("  Expected HP damage (approx): %d * multiplier", tt.hpDamage)
			t.Logf("  Actual HP change: %d", response.SuggestedStateChanges.HP)
			t.Logf("  Actual SAN change: %d", response.SuggestedStateChanges.SAN)

			// Verify damage is calculated (should be negative)
			if tt.hpDamage > 0 {
				assert.Less(t, response.SuggestedStateChanges.HP, 0, "HP should decrease")
			}
			if tt.sanDamage > 0 {
				assert.Less(t, response.SuggestedStateChanges.SAN, 0, "SAN should decrease")
			}
		})
	}
}

// TestCalculateStateChanges_ZeroDamage tests zero damage edge case
func TestCalculateStateChanges_ZeroDamage(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{})

	violations := []RuleViolation{
		{RuleID: "R-000", HPDamage: 0, SANDamage: 0, IsFatal: false},
	}

	changes := agent.CalculateStateChanges(violations, "normal", map[string]int{})

	assert.Equal(t, 0, changes.HP, "Zero damage should result in 0 HP change")
	assert.Equal(t, 0, changes.SAN, "Zero damage should result in 0 SAN change")
}

// TestInvokeJudge_RegexMatching tests regex-based rule matching
func TestInvokeJudge_RegexMatching(t *testing.T) {
	agent := NewJudgeAgent(AgentConfig{
		LLMClient: nil,
	})

	request := &JudgeRequest{
		PlayerChoice: "我數了123個數字",
		GameState: &GameStateSnapshot{
			HP:           100,
			SAN:          80,
			CurrentScene: "教室",
			Difficulty:   "normal",
			RuleWarnings: make(map[string]int),
		},
		ActiveRules: []JudgeHiddenRule{
			{
				ID:               "R-100",
				Name:             "數字禁忌",
				Type:             RuleTypeBehavior,
				TriggerKeywords:  []string{},
				TriggerRegex:     `\d{3,}`, // Match 3+ consecutive digits
				TriggerCondition: "數三個以上的數字",
				Punishment: RulePunishment{
					IsFatal:   false,
					HPDamage:  10,
					SANDamage: 30,
				},
			},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeJudge(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// Should detect violation via regex
	assert.Greater(t, len(response.RulesViolated), 0, "Should detect regex-based violation")
	assert.Equal(t, "R-100", response.RulesViolated[0].RuleID)

	t.Logf("Regex matching test: detected %d violations", len(response.RulesViolated))
}
