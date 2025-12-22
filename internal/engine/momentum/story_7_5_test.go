// Story 7-5 Test: Difficulty Config Integration
package momentum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetMomentumConfigForDifficulty_AC1_Easy tests Easy difficulty configuration
func TestGetMomentumConfigForDifficulty_AC1_Easy(t *testing.T) {
	config := GetMomentumConfigForDifficulty("easy")

	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, 3, config.MaxAutoBeats, "Easy: MaxAutoBeats should be 3")
	assert.Equal(t, RiskLow, config.PauseOnRisk, "Easy: PauseOnRisk should be Low")
}

// TestGetMomentumConfigForDifficulty_AC2_Normal tests Normal difficulty configuration
func TestGetMomentumConfigForDifficulty_AC2_Normal(t *testing.T) {
	config := GetMomentumConfigForDifficulty("normal")

	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, 5, config.MaxAutoBeats, "Normal: MaxAutoBeats should be 5")
	assert.Equal(t, RiskMedium, config.PauseOnRisk, "Normal: PauseOnRisk should be Medium")
}

// TestGetMomentumConfigForDifficulty_AC3_Hard tests Hard difficulty configuration
func TestGetMomentumConfigForDifficulty_AC3_Hard(t *testing.T) {
	config := GetMomentumConfigForDifficulty("hard")

	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, 7, config.MaxAutoBeats, "Hard: MaxAutoBeats should be 7")
	assert.Equal(t, RiskHigh, config.PauseOnRisk, "Hard: PauseOnRisk should be High")
}

// TestGetMomentumConfigForDifficulty_AC4_Hell tests Hell difficulty configuration
func TestGetMomentumConfigForDifficulty_AC4_Hell(t *testing.T) {
	config := GetMomentumConfigForDifficulty("hell")

	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, 10, config.MaxAutoBeats, "Hell: MaxAutoBeats should be 10")
	assert.Equal(t, RiskLethal, config.PauseOnRisk, "Hell: PauseOnRisk should be Lethal")
}

// TestGetMomentumConfigForDifficulty_CaseInsensitive tests case-insensitive difficulty strings
func TestGetMomentumConfigForDifficulty_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name       string
		difficulty string
		expected   int
	}{
		{"Easy lowercase", "easy", 3},
		{"Easy uppercase", "Easy", 3},
		{"Hard lowercase", "hard", 7},
		{"Hard uppercase", "Hard", 7},
		{"Hell lowercase", "hell", 10},
		{"Hell uppercase", "Hell", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetMomentumConfigForDifficulty(tt.difficulty)
			assert.Equal(t, tt.expected, config.MaxAutoBeats, "MaxAutoBeats mismatch for "+tt.difficulty)
		})
	}
}

// TestDifficultyToString tests the difficulty conversion helper
func TestDifficultyToString(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		expected   string
	}{
		{"Easy", 0, "easy"},
		{"Hard", 1, "hard"},
		{"Hell", 2, "hell"},
		{"Unknown", 99, "normal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DifficultyToString(tt.difficulty)
			assert.Equal(t, tt.expected, result, "Difficulty string mismatch")
		})
	}
}

// TestGetMomentumConfigForDifficulty_UnknownDefault tests unknown difficulty defaults to normal
func TestGetMomentumConfigForDifficulty_UnknownDefault(t *testing.T) {
	config := GetMomentumConfigForDifficulty("unknown")

	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, 5, config.MaxAutoBeats, "Unknown difficulty should default to Normal")
	assert.Equal(t, RiskMedium, config.PauseOnRisk, "Unknown difficulty should default to Normal")
}

// TestGetMomentumConfigForDifficulty_AllFieldsSet tests that all config fields are properly set
func TestGetMomentumConfigForDifficulty_AllFieldsSet(t *testing.T) {
	difficulties := []string{"easy", "normal", "hard", "hell"}

	for _, diff := range difficulties {
		t.Run(diff, func(t *testing.T) {
			config := GetMomentumConfigForDifficulty(diff)

			assert.NotNil(t, config, "Config should not be nil for "+diff)
			assert.True(t, config.PauseOnPlot, "PauseOnPlot should be true")
			assert.True(t, config.PauseOnNPC, "PauseOnNPC should be true")
			assert.True(t, config.PauseOnEvent, "PauseOnEvent should be true")
			assert.True(t, config.PlayerOverride, "PlayerOverride should be true")
		})
	}
}
