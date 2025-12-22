package game

import (
	"strings"
	"testing"
)

// TestGetSANState tests the SAN state calculation
func TestGetSANState(t *testing.T) {
	tests := []struct {
		name     string
		san      int
		expected SANState
	}{
		{"Clear - Max", 100, SANStateClear},
		{"Clear - Min", 80, SANStateClear},
		{"Anxious - Max", 79, SANStateAnxious},
		{"Anxious - Min", 50, SANStateAnxious},
		{"Panic - Max", 49, SANStatePanic},
		{"Panic - Min", 20, SANStatePanic},
		{"Breakdown - Max", 19, SANStateBreakdown},
		{"Breakdown - Min", 1, SANStateBreakdown},
		{"Insane", 0, SANStateInsane},
		{"Negative SAN", -10, SANStateInsane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSANState(tt.san)
			if result != tt.expected {
				t.Errorf("GetSANState(%d) = %v, want %v", tt.san, result, tt.expected)
			}
		})
	}
}

// TestSANStateString tests the String method
func TestSANStateString(t *testing.T) {
	tests := []struct {
		state    SANState
		expected string
	}{
		{SANStateClear, "清醒"},
		{SANStateAnxious, "焦慮"},
		{SANStatePanic, "恐慌"},
		{SANStateBreakdown, "崩潰"},
		{SANStateInsane, "瘋狂"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("SANState.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestSANStateDescription tests the Description method
func TestSANStateDescription(t *testing.T) {
	tests := []struct {
		state SANState
	}{
		{SANStateClear},
		{SANStateAnxious},
		{SANStatePanic},
		{SANStateBreakdown},
		{SANStateInsane},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			result := tt.state.Description()
			if result == "" {
				t.Errorf("SANState.Description() returned empty string for %v", tt.state)
			}
		})
	}
}

// TestDamageTypeString tests the String method for DamageType
func TestDamageTypeString(t *testing.T) {
	tests := []struct {
		damageType DamageType
		expected   string
	}{
		{DamageMinor, "輕傷"},
		{DamageModerate, "中傷"},
		{DamageSevere, "重傷"},
		{DamageFatal, "致命傷"},
		{SANTeammateDeath, "目睹隊友死亡"},
		{SANHorrorScene, "遭遇恐怖場景"},
		{SANMonsterSighting, "看到怪物實體"},
		{SANCruelTruth, "發現殘酷真相"},
		{SANRulePenalty, "違反規則"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.damageType.String()
			if result != tt.expected {
				t.Errorf("DamageType.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestCalculateDamage tests damage calculation for HP
func TestCalculateDamage(t *testing.T) {
	tests := []struct {
		name        string
		damageType  DamageType
		minExpected int
		maxExpected int
	}{
		{"Minor HP Damage", DamageMinor, -20, -10},
		{"Moderate HP Damage", DamageModerate, -50, -30},
		{"Severe HP Damage", DamageSevere, -80, -60},
		{"Fatal HP Damage", DamageFatal, -100, -100},
		{"SAN Teammate Death", SANTeammateDeath, -25, -15},
		{"SAN Horror Scene", SANHorrorScene, -15, -5},
		{"SAN Monster Sighting", SANMonsterSighting, -30, -20},
		{"SAN Cruel Truth", SANCruelTruth, -20, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiple times to cover randomness
			for i := 0; i < 10; i++ {
				result := CalculateDamage(tt.damageType)
				if result < tt.minExpected || result > tt.maxExpected {
					t.Errorf("CalculateDamage(%v) = %d, want between %d and %d",
						tt.damageType, result, tt.minExpected, tt.maxExpected)
				}
			}
		})
	}
}

// TestHealTypeString tests the String method for HealType
func TestHealTypeString(t *testing.T) {
	tests := []struct {
		healType HealType
		expected string
	}{
		{HealItem, "使用恢復道具"},
		{HealMedical, "獲得治療"},
		{SANRest, "安全場景休息"},
		{SANPuzzleSolved, "解決謎題"},
		{SANMedicine, "使用安定劑"},
		{SANPositiveInteraction, "正向互動"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.healType.String()
			if result != tt.expected {
				t.Errorf("HealType.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestCalculateHeal tests heal calculation
func TestCalculateHeal(t *testing.T) {
	tests := []struct {
		name        string
		healType    HealType
		minExpected int
		maxExpected int
	}{
		{"HP Item Heal", HealItem, 20, 40},
		{"HP Medical Heal", HealMedical, 30, 60},
		{"SAN Rest", SANRest, 5, 10},
		{"SAN Puzzle Solved", SANPuzzleSolved, 10, 15},
		{"SAN Medicine", SANMedicine, 20, 20},
		{"SAN Positive Interaction", SANPositiveInteraction, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiple times to cover randomness
			for i := 0; i < 10; i++ {
				result := CalculateHeal(tt.healType)
				if result < tt.minExpected || result > tt.maxExpected {
					t.Errorf("CalculateHeal(%v) = %d, want between %d and %d",
						tt.healType, result, tt.minExpected, tt.maxExpected)
				}
			}
		})
	}
}

// TestGetSANEffectProfile tests SAN effect profile retrieval
func TestGetSANEffectProfile(t *testing.T) {
	tests := []struct {
		name                     string
		state                    SANState
		expectedCorruptionMin    float64
		expectedCorruptionMax    float64
		expectedHallucinationMin float64
		expectedStability        int
	}{
		{"Clear State", SANStateClear, 0.0, 0.0, 0.0, 0},
		{"Anxious State", SANStateAnxious, 0.0, 0.2, 0.0, 1},
		{"Panic State", SANStatePanic, 0.2, 0.4, 0.2, 3},
		{"Breakdown State", SANStateBreakdown, 0.6, 0.8, 0.5, 5},
		{"Insane State", SANStateInsane, 1.0, 1.0, 1.0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := GetSANEffectProfile(tt.state)

			if profile.State != tt.state {
				t.Errorf("State = %v, want %v", profile.State, tt.state)
			}

			if profile.TextCorruptionLevel < tt.expectedCorruptionMin || profile.TextCorruptionLevel > tt.expectedCorruptionMax {
				t.Errorf("TextCorruptionLevel = %f, want between %f and %f",
					profile.TextCorruptionLevel, tt.expectedCorruptionMin, tt.expectedCorruptionMax)
			}

			if profile.HallucinationChance < tt.expectedHallucinationMin {
				t.Errorf("HallucinationChance = %f, want >= %f",
					profile.HallucinationChance, tt.expectedHallucinationMin)
			}

			if profile.UIStabilityLevel != tt.expectedStability {
				t.Errorf("UIStabilityLevel = %d, want %d",
					profile.UIStabilityLevel, tt.expectedStability)
			}

			if profile.NarrativeModifier == "" {
				t.Errorf("NarrativeModifier is empty for state %v", tt.state)
			}
		})
	}
}

// TestApplyTextDistortion tests text distortion functionality
func TestApplyTextDistortion(t *testing.T) {
	originalText := "這是一段測試文字。"

	tests := []struct {
		name  string
		state SANState
	}{
		{"Clear - No Distortion", SANStateClear},
		{"Anxious - Mild Distortion", SANStateAnxious},
		{"Panic - Moderate Distortion", SANStatePanic},
		{"Breakdown - Severe Distortion", SANStateBreakdown},
		{"Insane - Maximum Distortion", SANStateInsane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyTextDistortion(originalText, tt.state)

			// Clear state should have no distortion
			if tt.state == SANStateClear && result != originalText {
				t.Errorf("Clear state should not distort text, got %s", result)
			}

			// All states should return non-empty text
			if result == "" {
				t.Errorf("ApplyTextDistortion returned empty string")
			}

			// Text length should remain similar (accounting for UTF-8)
			if len([]rune(result)) < len([]rune(originalText))-5 {
				t.Errorf("Distorted text too short: %d runes vs original %d runes",
					len([]rune(result)), len([]rune(originalText)))
			}
		})
	}
}

// TestGenerateHallucinationOption tests hallucination generation
func TestGenerateHallucinationOption(t *testing.T) {
	context := "玩家在走廊中"

	tests := []struct {
		name          string
		state         SANState
		shouldGenerate bool
	}{
		{"Clear - No Hallucinations", SANStateClear, false},
		{"Anxious - No Hallucinations", SANStateAnxious, false},
		{"Panic - Possible Hallucinations", SANStatePanic, true},
		{"Breakdown - Likely Hallucinations", SANStateBreakdown, true},
		{"Insane - Always Hallucinations", SANStateInsane, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiple times due to randomness
			generatedCount := 0
			for i := 0; i < 20; i++ {
				result := GenerateHallucinationOption(tt.state, context)
				if result != "" {
					generatedCount++
				}
			}

			if !tt.shouldGenerate && generatedCount > 0 {
				t.Errorf("State %v should not generate hallucinations, but generated %d times",
					tt.state, generatedCount)
			}

			// For states that should generate, we expect at least some generation
			if tt.shouldGenerate && tt.state == SANStateInsane && generatedCount == 0 {
				t.Errorf("Insane state should always generate hallucinations")
			}
		})
	}
}

// TestShouldForceAction tests forced action logic
func TestShouldForceAction(t *testing.T) {
	tests := []struct {
		name        string
		state       SANState
		shouldForce bool
	}{
		{"Clear - No Forced Actions", SANStateClear, false},
		{"Anxious - No Forced Actions", SANStateAnxious, false},
		{"Panic - No Forced Actions", SANStatePanic, false},
		{"Breakdown - Possible Forced Actions", SANStateBreakdown, true},
		{"Insane - Always Forced Actions", SANStateInsane, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiple times due to randomness
			forcedCount := 0
			for i := 0; i < 20; i++ {
				forced, action := ShouldForceAction(tt.state)
				if forced {
					forcedCount++
					if action == "" {
						t.Errorf("Forced action should have description")
					}
				}
			}

			if !tt.shouldForce && forcedCount > 0 {
				t.Errorf("State %v should not force actions, but forced %d times",
					tt.state, forcedCount)
			}

			if tt.shouldForce && tt.state == SANStateInsane && forcedCount == 0 {
				t.Errorf("Insane state should force actions")
			}
		})
	}
}

// TestGetNarrativeStyle tests narrative style retrieval
func TestGetNarrativeStyle(t *testing.T) {
	tests := []struct {
		name  string
		state SANState
	}{
		{"Clear Narrative", SANStateClear},
		{"Anxious Narrative", SANStateAnxious},
		{"Panic Narrative", SANStatePanic},
		{"Breakdown Narrative", SANStateBreakdown},
		{"Insane Narrative", SANStateInsane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNarrativeStyle(tt.state)
			if result == "" {
				t.Errorf("GetNarrativeStyle(%v) returned empty string", tt.state)
			}
		})
	}
}

// TestFormatSANStateSummary tests SAN state summary formatting
func TestFormatSANStateSummary(t *testing.T) {
	tests := []struct {
		name string
		san  int
	}{
		{"Clear State", 100},
		{"Anxious State", 60},
		{"Panic State", 30},
		{"Breakdown State", 10},
		{"Insane State", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSANStateSummary(tt.san)

			// Check that summary contains key information
			if !strings.Contains(result, "SAN:") {
				t.Errorf("Summary should contain 'SAN:', got: %s", result)
			}

			if !strings.Contains(result, "狀態:") {
				t.Errorf("Summary should contain '狀態:', got: %s", result)
			}

			if result == "" {
				t.Errorf("FormatSANStateSummary(%d) returned empty string", tt.san)
			}
		})
	}
}

// TestSANEffectProfileIntegrity tests that effect profiles are consistent
func TestSANEffectProfileIntegrity(t *testing.T) {
	states := []SANState{
		SANStateClear,
		SANStateAnxious,
		SANStatePanic,
		SANStateBreakdown,
		SANStateInsane,
	}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			profile := GetSANEffectProfile(state)

			// Check ranges
			if profile.TextCorruptionLevel < 0.0 || profile.TextCorruptionLevel > 1.0 {
				t.Errorf("TextCorruptionLevel out of range [0, 1]: %f", profile.TextCorruptionLevel)
			}

			if profile.ColorShiftIntensity < 0.0 || profile.ColorShiftIntensity > 1.0 {
				t.Errorf("ColorShiftIntensity out of range [0, 1]: %f", profile.ColorShiftIntensity)
			}

			if profile.UIStabilityLevel < 0 || profile.UIStabilityLevel > 5 {
				t.Errorf("UIStabilityLevel out of range [0, 5]: %d", profile.UIStabilityLevel)
			}

			if profile.HallucinationChance < 0.0 || profile.HallucinationChance > 1.0 {
				t.Errorf("HallucinationChance out of range [0, 1]: %f", profile.HallucinationChance)
			}

			if profile.ForcedActionChance < 0.0 || profile.ForcedActionChance > 1.0 {
				t.Errorf("ForcedActionChance out of range [0, 1]: %f", profile.ForcedActionChance)
			}

			// Check that effects increase with severity
			if state != SANStateClear {
				clearProfile := GetSANEffectProfile(SANStateClear)
				if profile.TextCorruptionLevel < clearProfile.TextCorruptionLevel {
					t.Errorf("TextCorruptionLevel should increase with severity")
				}
			}
		})
	}
}
