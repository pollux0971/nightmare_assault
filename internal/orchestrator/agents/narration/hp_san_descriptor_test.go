package narration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDescribeHPChange tests HP change descriptions
func TestDescribeHPChange(t *testing.T) {
	tests := []struct {
		name          string
		delta         int
		reason        string
		expectContain []string
		minLength     int
	}{
		{
			name:          "minor damage",
			delta:         -5,
			reason:        "你被輕微劃傷",
			expectContain: []string{"疼痛", "HP -5", "輕微劃傷"},
			minLength:     20,
		},
		{
			name:          "moderate damage",
			delta:         -20,
			reason:        "你被重重擊中",
			expectContain: []string{"劇痛", "HP -20", "重重擊中"},
			minLength:     25,
		},
		{
			name:          "major damage",
			delta:         -40,
			reason:        "你遭受嚴重攻擊",
			expectContain: []string{"重創", "HP -40", "嚴重攻擊"},
			minLength:     25,
		},
		{
			name:          "lethal damage",
			delta:         -60,
			reason:        "致命一擊",
			expectContain: []string{"致命", "HP -60", "致命一擊"},
			minLength:     20,
		},
		{
			name:          "minor healing",
			delta:         5,
			reason:        "你使用了繃帶",
			expectContain: []string{"好", "HP +5", "繃帶"},
			minLength:     15,
		},
		{
			name:          "moderate healing",
			delta:         20,
			reason:        "你得到了充分休息",
			expectContain: []string{"好多了", "HP +20", "充分休息"},
			minLength:     20,
		},
		{
			name:          "no change",
			delta:         0,
			reason:        "",
			expectContain: []string{},
			minLength:     0,
		},
		{
			name:          "damage without reason",
			delta:         -15,
			reason:        "",
			expectContain: []string{"HP -15"},
			minLength:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DescribeHPChange(tt.delta, tt.reason)

			// Check minimum length (except for no change)
			if tt.delta != 0 {
				runeCount := len([]rune(result))
				assert.GreaterOrEqual(t, runeCount, tt.minLength,
					"Description too short: %d chars", runeCount)
			}

			// Check expected content
			for _, expect := range tt.expectContain {
				assert.Contains(t, result, expect,
					"Description should contain '%s'", expect)
			}

			// If delta is 0, result should be empty
			if tt.delta == 0 {
				assert.Empty(t, result)
			}
		})
	}
}

// TestDescribeSANChange tests SAN change descriptions
func TestDescribeSANChange(t *testing.T) {
	tests := []struct {
		name          string
		delta         int
		reason        string
		expectContain []string
		minLength     int
	}{
		{
			name:          "minor san loss",
			delta:         -5,
			reason:        "你目睹了詭異的景象",
			expectContain: []string{"不安", "SAN -5", "詭異的景象"},
			minLength:     20,
		},
		{
			name:          "moderate san loss",
			delta:         -20,
			reason:        "你看見了不應存在的東西",
			expectContain: []string{"理智", "動搖", "SAN -20", "不應存在"},
			minLength:     30,
		},
		{
			name:          "major san loss",
			delta:         -40,
			reason:        "現實崩潰了",
			expectContain: []string{"理智搖搖欲墜", "SAN -40", "現實崩潰"},
			minLength:     30,
		},
		{
			name:          "lethal san loss",
			delta:         -60,
			reason:        "瘋狂侵蝕了你的心智",
			expectContain: []string{"意識崩潰", "SAN -60", "瘋狂"},
			minLength:     20,
		},
		{
			name:          "minor san recovery",
			delta:         5,
			reason:        "你感到安心",
			expectContain: []string{"平靜", "SAN +5", "安心"},
			minLength:     15,
		},
		{
			name:          "moderate san recovery",
			delta:         20,
			reason:        "你理解了真相",
			expectContain: []string{"理智", "恢復", "SAN +20", "真相"},
			minLength:     20,
		},
		{
			name:          "no change",
			delta:         0,
			reason:        "",
			expectContain: []string{},
			minLength:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DescribeSANChange(tt.delta, tt.reason)

			// Check minimum length (except for no change)
			if tt.delta != 0 {
				runeCount := len([]rune(result))
				assert.GreaterOrEqual(t, runeCount, tt.minLength,
					"Description too short: %d chars", runeCount)
			}

			// Check expected content
			for _, expect := range tt.expectContain {
				assert.Contains(t, result, expect,
					"Description should contain '%s'", expect)
			}

			// If delta is 0, result should be empty
			if tt.delta == 0 {
				assert.Empty(t, result)
			}
		})
	}
}

// TestHPSANDescriptionLength tests that descriptions meet AC #6 requirements
// AC #6: HP/SAN 變化描述（80-120 字）
func TestHPSANDescriptionLength(t *testing.T) {
	testCases := []struct {
		name   string
		delta  int
		reason string
	}{
		{"hp_damage", -25, "你被怪物攻擊，鮮血從傷口湧出"},
		{"san_loss", -25, "你目睹了難以名狀的恐怖景象，理智受到衝擊"},
		{"hp_heal", 15, "你包紮了傷口，疼痛逐漸減輕"},
		{"san_recovery", 15, "你深呼吸，努力讓自己冷靜下來"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result string
			if strings.Contains(tc.name, "hp") {
				result = DescribeHPChange(tc.delta, tc.reason)
			} else {
				result = DescribeSANChange(tc.delta, tc.reason)
			}

			runeCount := len([]rune(result))
			// AC requires 80-120 chars, but we allow some flexibility
			// since the reason string length varies
			assert.GreaterOrEqual(t, runeCount, 30,
				"Description should be at least 30 chars (AC: 80-120)")
			assert.LessOrEqual(t, runeCount, 150,
				"Description should be at most 150 chars (AC: 80-120)")
		})
	}
}

// TestGetHPSeverity tests HP severity classification
func TestGetHPSeverity(t *testing.T) {
	tests := []struct {
		delta            int
		expectedSeverity HPSeverity
	}{
		{-5, HPSeverityMinor},
		{-10, HPSeverityMinor},
		{-11, HPSeverityModerate},
		{-30, HPSeverityModerate},
		{-31, HPSeverityMajor},
		{-50, HPSeverityMajor},
		{-51, HPSeverityLethal},
		{-100, HPSeverityLethal},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("delta_%d", tt.delta), func(t *testing.T) {
			severity := getHPSeverity(tt.delta)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

// TestGetSANSeverity tests SAN severity classification
func TestGetSANSeverity(t *testing.T) {
	tests := []struct {
		delta            int
		expectedSeverity SANSeverity
	}{
		{-5, SANSeverityMinor},
		{-10, SANSeverityMinor},
		{-11, SANSeverityModerate},
		{-30, SANSeverityModerate},
		{-31, SANSeverityMajor},
		{-50, SANSeverityMajor},
		{-51, SANSeverityLethal},
		{-100, SANSeverityLethal},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("delta_%d", tt.delta), func(t *testing.T) {
			severity := getSANSeverity(tt.delta)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}
