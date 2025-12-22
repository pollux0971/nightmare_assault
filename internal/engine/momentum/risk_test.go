package momentum

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEvaluateContext_NilState tests that nil state is handled safely
func TestEvaluateContext_NilState(t *testing.T) {
	evaluator := NewRiskEvaluator()
	assessment, err := evaluator.EvaluateContext(nil, "test_scene")

	require.NoError(t, err)
	assert.Equal(t, RiskNone, assessment.Level)
	assert.Empty(t, assessment.Factors)
}

// TestEvaluateContext_HPSANCombinations tests various HP/SAN combinations
// AC3: HP/SAN 狀態影響風險等級
func TestEvaluateContext_HPSANCombinations(t *testing.T) {
	tests := []struct {
		name            string
		hp              int
		san             int
		expectedLevel   RiskLevel
		expectedFactors []string
	}{
		{
			name:            "Full health and sanity",
			hp:              100,
			san:             100,
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Moderate HP and SAN",
			hp:              50,
			san:             60,
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Medium HP, normal SAN",
			hp:              35,
			san:             70,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"medium_hp"},
		},
		{
			name:            "Low HP critical",
			hp:              15,
			san:             80,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_hp"},
		},
		{
			name:            "Normal HP, low SAN critical",
			hp:              80,
			san:             15,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_san"},
		},
		{
			name:            "Both HP and SAN critical",
			hp:              20,
			san:             25,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_hp", "medium_san"},
		},
		{
			name:            "Both HP and SAN medium",
			hp:              35,
			san:             30,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"medium_hp", "medium_san"},
		},
		{
			name:            "HP at exact threshold (20)",
			hp:              20,
			san:             100,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_hp"},
		},
		{
			name:            "SAN at exact threshold (40)",
			hp:              100,
			san:             40,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"medium_san"},
		},
		{
			name:            "HP zero (dead)",
			hp:              0,
			san:             50,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_hp"},
		},
		{
			name:            "SAN zero (insane)",
			hp:              50,
			san:             0,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_san"},
		},
	}

	evaluator := NewRiskEvaluator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.SetHP(tt.hp)
			state.SetSAN(tt.san)

			assessment, err := evaluator.EvaluateContext(state, "測試場景")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, assessment.Level, "Risk level mismatch")
			assert.Equal(t, len(tt.expectedFactors), len(assessment.Factors), "Factor count mismatch")

			// 驗證因素名稱
			factorNames := make([]string, len(assessment.Factors))
			for i, f := range assessment.Factors {
				factorNames[i] = f.Name
			}
			assert.ElementsMatch(t, tt.expectedFactors, factorNames, "Factor names mismatch")
		})
	}
}

// TestEvaluateContext_ActiveRules tests active rules count impact
// AC4: 活躍規則數量影響風險等級
func TestEvaluateContext_ActiveRules(t *testing.T) {
	tests := []struct {
		name            string
		ruleCount       int
		expectedLevel   RiskLevel
		expectedFactors []string
	}{
		{
			name:            "No active rules",
			ruleCount:       0,
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Few rules (2)",
			ruleCount:       2,
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Some rules (3)",
			ruleCount:       3,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"some_active_rules"},
		},
		{
			name:            "Some rules (4)",
			ruleCount:       4,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"some_active_rules"},
		},
		{
			name:            "Many rules (5)",
			ruleCount:       5,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"many_active_rules"},
		},
		{
			name:            "Many rules (7)",
			ruleCount:       7,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"many_active_rules"},
		},
	}

	evaluator := NewRiskEvaluator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			// 添加指定數量的規則
			for i := 0; i < tt.ruleCount; i++ {
				state.ActiveRules = append(state.ActiveRules, &engine.ActiveRule{
					ID:   "rule_" + string(rune('A'+i)),
					Name: "Test Rule",
				})
			}

			assessment, err := evaluator.EvaluateContext(state, "測試場景")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, assessment.Level)

			factorNames := make([]string, len(assessment.Factors))
			for i, f := range assessment.Factors {
				factorNames[i] = f.Name
			}
			assert.ElementsMatch(t, tt.expectedFactors, factorNames)
		})
	}
}

// TestEvaluateContext_Tension tests tension level impact
// AC5: 張力等級影響風險等級
func TestEvaluateContext_Tension(t *testing.T) {
	tests := []struct {
		name            string
		tension         int
		expectedLevel   RiskLevel
		expectedFactors []string
	}{
		{
			name:            "Low tension (30)",
			tension:         30,
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Medium tension (55)",
			tension:         55,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"medium_tension"},
		},
		{
			name:            "High tension (85)",
			tension:         85,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"high_tension"},
		},
		{
			name:            "Tension at medium threshold (50)",
			tension:         50,
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"medium_tension"},
		},
		{
			name:            "Tension at high threshold (80)",
			tension:         80,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"high_tension"},
		},
		{
			name:            "Maximum tension (100)",
			tension:         100,
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"high_tension"},
		},
	}

	evaluator := NewRiskEvaluator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.Tension.SetValue(tt.tension)

			assessment, err := evaluator.EvaluateContext(state, "測試場景")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, assessment.Level)

			factorNames := make([]string, len(assessment.Factors))
			for i, f := range assessment.Factors {
				factorNames[i] = f.Name
			}
			assert.ElementsMatch(t, tt.expectedFactors, factorNames)
		})
	}
}

// TestEvaluateContext_NilTension tests handling of nil Tension
func TestEvaluateContext_NilTension(t *testing.T) {
	evaluator := NewRiskEvaluator()
	state := engine.NewGameStateV2()
	state.Tension = nil

	assessment, err := evaluator.EvaluateContext(state, "測試場景")
	require.NoError(t, err)
	assert.Equal(t, RiskNone, assessment.Level)
	assert.Empty(t, assessment.Factors)
}

// TestEvaluateContext_SceneDanger tests scene danger evaluation
// AC6: 場景危險度評估
func TestEvaluateContext_SceneDanger(t *testing.T) {
	tests := []struct {
		name            string
		scene           string
		expectedLevel   RiskLevel
		expectedFactors []string
	}{
		{
			name:            "Safe scene",
			scene:           "安全房",
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Safe scene (休息室)",
			scene:           "休息室",
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
		{
			name:            "Medium risk scene (走廊)",
			scene:           "陰暗的走廊",
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"dangerous_scene"},
		},
		{
			name:            "Medium risk scene (English corridor)",
			scene:           "Dark corridor",
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"dangerous_scene"},
		},
		{
			name:            "High risk scene (地下室)",
			scene:           "地下室",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"dangerous_scene"},
		},
		{
			name:            "High risk scene (basement)",
			scene:           "Basement Storage",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"dangerous_scene"},
		},
		{
			name:            "High risk scene (墓地)",
			scene:           "老舊墓地",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"dangerous_scene"},
		},
		{
			name:            "High risk scene (密室)",
			scene:           "神秘的密室",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"dangerous_scene"},
		},
		{
			name:            "Empty scene name",
			scene:           "",
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
	}

	evaluator := NewRiskEvaluator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()

			assessment, err := evaluator.EvaluateContext(state, tt.scene)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, assessment.Level)

			factorNames := make([]string, len(assessment.Factors))
			for i, f := range assessment.Factors {
				factorNames[i] = f.Name
			}
			assert.ElementsMatch(t, tt.expectedFactors, factorNames)
		})
	}
}

// TestEvaluateContext_MultipleFactors tests multiple risk factors aggregation
// 測試多重風險因素疊加 (最高風險決策)
func TestEvaluateContext_MultipleFactors(t *testing.T) {
	tests := []struct {
		name            string
		hp              int
		san             int
		tension         int
		ruleCount       int
		scene           string
		expectedLevel   RiskLevel
		expectedFactors []string
	}{
		{
			name:            "Multiple medium risks",
			hp:              35,
			san:             100,
			tension:         55,
			ruleCount:       3,
			scene:           "走廊",
			expectedLevel:   RiskMedium,
			expectedFactors: []string{"medium_hp", "medium_tension", "some_active_rules", "dangerous_scene"},
		},
		{
			name:            "One high risk dominates",
			hp:              15,
			san:             100,
			tension:         30,
			ruleCount:       2,
			scene:           "安全房",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_hp"},
		},
		{
			name:            "Mixed medium and high risks",
			hp:              35,
			san:             100,
			tension:         85,
			ruleCount:       4,
			scene:           "地下室",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"medium_hp", "high_tension", "some_active_rules", "dangerous_scene"},
		},
		{
			name:            "All high risks",
			hp:              15,
			san:             15,
			tension:         90,
			ruleCount:       6,
			scene:           "墓地",
			expectedLevel:   RiskHigh,
			expectedFactors: []string{"low_hp", "low_san", "high_tension", "many_active_rules", "dangerous_scene"},
		},
		{
			name:            "No risk factors",
			hp:              100,
			san:             100,
			tension:         30,
			ruleCount:       1,
			scene:           "休息室",
			expectedLevel:   RiskNone,
			expectedFactors: []string{},
		},
	}

	evaluator := NewRiskEvaluator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.SetHP(tt.hp)
			state.SetSAN(tt.san)
			state.Tension.SetValue(tt.tension)

			// 添加規則
			for i := 0; i < tt.ruleCount; i++ {
				state.ActiveRules = append(state.ActiveRules, &engine.ActiveRule{
					ID:   "rule_" + string(rune('A'+i)),
					Name: "Test Rule",
				})
			}

			assessment, err := evaluator.EvaluateContext(state, tt.scene)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, assessment.Level, "Risk level mismatch")
			assert.Equal(t, len(tt.expectedFactors), len(assessment.Factors), "Factor count mismatch")

			factorNames := make([]string, len(assessment.Factors))
			for i, f := range assessment.Factors {
				factorNames[i] = f.Name
			}
			assert.ElementsMatch(t, tt.expectedFactors, factorNames, "Factor names mismatch")
		})
	}
}

// TestEvaluateContext_BoundaryConditions tests edge cases
func TestEvaluateContext_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name            string
		hp              int
		san             int
		expectedLevel   RiskLevel
		minFactorCount  int
	}{
		{
			name:           "HP at 21 (just above low threshold)",
			hp:             21,
			san:            100,
			expectedLevel:  RiskMedium,
			minFactorCount: 1,
		},
		{
			name:           "HP at 41 (just above medium threshold)",
			hp:             41,
			san:            100,
			expectedLevel:  RiskNone,
			minFactorCount: 0,
		},
		{
			name:           "SAN at 21 (just above low threshold)",
			hp:             100,
			san:            21,
			expectedLevel:  RiskMedium,
			minFactorCount: 1,
		},
		{
			name:           "SAN at 41 (just above medium threshold)",
			hp:             100,
			san:            41,
			expectedLevel:  RiskNone,
			minFactorCount: 0,
		},
	}

	evaluator := NewRiskEvaluator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.SetHP(tt.hp)
			state.SetSAN(tt.san)

			assessment, err := evaluator.EvaluateContext(state, "測試場景")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, assessment.Level)
			assert.GreaterOrEqual(t, len(assessment.Factors), tt.minFactorCount)
		})
	}
}

// TestEvaluateSceneDanger tests the internal scene danger evaluation method
func TestEvaluateSceneDanger(t *testing.T) {
	evaluator := NewRiskEvaluator()
	state := engine.NewGameStateV2()

	tests := []struct {
		name     string
		scene    string
		expected RiskLevel
	}{
		{"Empty scene", "", RiskNone},
		{"Safe scene", "安全房", RiskNone},
		{"Corridor", "走廊", RiskMedium},
		{"Dark hallway", "dark hallway", RiskMedium},
		{"Basement", "basement", RiskHigh},
		{"地下室", "地下室", RiskHigh},
		{"Cemetery", "old cemetery", RiskHigh},
		{"Morgue", "hospital morgue", RiskHigh},
		{"Attic", "dusty attic", RiskMedium},
		{"Laboratory", "abandoned laboratory", RiskMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.evaluateSceneDanger(tt.scene, state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluate_BackwardCompatibility tests the old Evaluate method
func TestEvaluate_BackwardCompatibility(t *testing.T) {
	evaluator := NewRiskEvaluator()

	t.Run("Nil context", func(t *testing.T) {
		result := evaluator.Evaluate(nil)
		assert.Equal(t, RiskNone, result)
	})

	t.Run("With context", func(t *testing.T) {
		ctx := &NarrativeContext{
			RiskLevel: RiskHigh,
		}
		result := evaluator.Evaluate(ctx)
		assert.Equal(t, RiskHigh, result)
	})
}

// TestRiskAssessment_AddFactor tests the addFactor method
func TestRiskAssessment_AddFactor(t *testing.T) {
	assessment := &RiskAssessment{
		Level:   RiskNone,
		Factors: make([]RiskFactor, 0),
	}

	// Add first factor (Medium)
	assessment.addFactor("medium_hp", RiskMedium)
	assert.Equal(t, RiskMedium, assessment.Level)
	assert.Len(t, assessment.Factors, 1)
	assert.Equal(t, "medium_hp", assessment.Factors[0].Name)

	// Add second factor (High) - should update level
	assessment.addFactor("low_san", RiskHigh)
	assert.Equal(t, RiskHigh, assessment.Level)
	assert.Len(t, assessment.Factors, 2)

	// Add third factor (Low) - should not update level
	assessment.addFactor("some_factor", RiskLow)
	assert.Equal(t, RiskHigh, assessment.Level)
	assert.Len(t, assessment.Factors, 3)
}

// TestNewRiskEvaluator tests the constructor
func TestNewRiskEvaluator(t *testing.T) {
	evaluator := NewRiskEvaluator()
	assert.NotNil(t, evaluator)
}
