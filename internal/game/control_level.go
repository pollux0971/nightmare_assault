// Package game provides game-related types and logic for Nightmare Assault.
package game

import "math/rand"

// ControlLevel represents the player's control over the game based on SAN.
// Story 7.5 AC6: Control deprivation is gradual, not sudden.
type ControlLevel struct {
	SAN                int     // Current SAN value
	ControlPercentage  float64 // 0.0 - 1.0, percentage of control retained
	VisualInterference float64 // 0.0 - 1.0, visual disturbance intensity
	TextDistortion     float64 // 0.0 - 1.0, text distortion level
	HallucinationRate  float64 // 0.0 - 1.0, probability of hallucination options
	CharDeletionRate   float64 // 0.0 - 1.0, probability of character deletion
}

// GetControlLevel calculates the control level based on current SAN.
// Story 7.5 AC6: Gradual control deprivation mapping
//
// Code Review Fix 7-4-1: Aligned boundaries with san_effects.go (Story 7.4 AC4)
// Control Levels (aligned with SANState boundaries):
//   - SAN 80-100: Clear (清醒) - 100% control, no effects
//   - SAN 50-79:  Anxious (焦慮) - 95% control, minor visual interference
//   - SAN 20-49:  Panic (恐慌) - 85% control, text distortion + visual interference
//   - SAN 1-19:   Breakdown (崩潰) - 60% control, char deletion + hallucinations
//   - SAN 0:      Insane (瘋狂) - Game Over
//
// Note: Control is never completely removed - player always can input.
func GetControlLevel(san int) ControlLevel {
	// Clamp SAN to valid range
	if san < 0 {
		san = 0
	}
	if san > 100 {
		san = 100
	}

	switch {
	case san >= 80:
		// SAN 80-100: Clear (清醒) - Full control, no effects
		// Matches SANStateClear from san_effects.go
		return ControlLevel{
			SAN:                san,
			ControlPercentage:  1.0,
			VisualInterference: 0.0,
			TextDistortion:     0.0,
			HallucinationRate:  0.0,
			CharDeletionRate:   0.0,
		}

	case san >= 50:
		// SAN 50-79: Anxious (焦慮) - 95% control
		// Matches SANStateAnxious from san_effects.go
		// Minor visual effects: heartbeat, cold sweat descriptions
		return ControlLevel{
			SAN:                san,
			ControlPercentage:  0.95,
			VisualInterference: 0.15,
			TextDistortion:     0.0,
			HallucinationRate:  0.0,
			CharDeletionRate:   0.0,
		}

	case san >= 20:
		// SAN 20-49: Panic (恐慌) - 85% control
		// Matches SANStatePanic from san_effects.go
		// Text distortion begins, visual disturbance increases
		// Possible "panic options" (irrational behavior)
		return ControlLevel{
			SAN:                san,
			ControlPercentage:  0.85,
			VisualInterference: 0.4,
			TextDistortion:     0.3,
			HallucinationRate:  0.0,
			CharDeletionRate:   0.0,
		}

	default:
		// SAN 1-19: Breakdown (崩潰) - 60% control
		// Matches SANStateBreakdown from san_effects.go
		// Severe effects: hallucinations, forced actions, char deletion
		return ControlLevel{
			SAN:                san,
			ControlPercentage:  0.60,
			VisualInterference: 0.8,
			TextDistortion:     0.6,
			HallucinationRate:  0.5, // 50% chance to add hallucination options
			CharDeletionRate:   0.15, // 15% character deletion rate (10-20% range)
		}
	}
}

// GetCharDeletionCount calculates how many characters should be deleted.
// Story 7.5 AC4: Randomly delete 10-20% of characters when SAN 1-19.
func (c ControlLevel) GetCharDeletionCount(textLength int) int {
	if c.CharDeletionRate == 0.0 || textLength == 0 {
		return 0
	}

	// Calculate deletion percentage: 10-20% based on CharDeletionRate
	// CharDeletionRate of 0.15 means 15% base rate
	minRate := 0.10
	maxRate := 0.20
	deletionRate := minRate + (c.CharDeletionRate * (maxRate - minRate) / 0.15)

	// Add randomness
	deletionRate += (rand.Float64() * 0.05) - 0.025 // ±2.5%

	// Ensure within bounds
	if deletionRate < minRate {
		deletionRate = minRate
	}
	if deletionRate > maxRate {
		deletionRate = maxRate
	}

	deleteCount := int(float64(textLength) * deletionRate)
	return deleteCount
}

// ShouldShowHallucinationOption determines if hallucination options should appear.
// Story 7.5 AC5: Random hallucination option insertion when SAN < 20.
func (c ControlLevel) ShouldShowHallucinationOption() bool {
	if c.HallucinationRate == 0.0 {
		return false
	}
	return rand.Float64() < c.HallucinationRate
}

// GetHallucinationCount returns the number of hallucination options to add.
// Story 7.5 AC5: Insert 1-2 hallucination options when SAN < 20.
func (c ControlLevel) GetHallucinationCount() int {
	if !c.ShouldShowHallucinationOption() {
		return 0
	}
	// Randomly return 1 or 2
	if rand.Float64() < 0.5 {
		return 1
	}
	return 2
}

// DescribeControlState returns a human-readable description of the control state.
func (c ControlLevel) DescribeControlState() string {
	switch {
	case c.ControlPercentage == 1.0:
		return "完全控制"
	case c.ControlPercentage >= 0.9:
		return "幾乎完全控制"
	case c.ControlPercentage >= 0.8:
		return "部分控制受限"
	case c.ControlPercentage >= 0.6:
		return "控制嚴重受限"
	default:
		return "幾乎失控"
	}
}

// DescribePlayerFeeling returns a description of what the player is experiencing.
// Code Review Fix 7-4-1: Aligned with san_effects.go SANState boundaries
func (c ControlLevel) DescribePlayerFeeling() string {
	switch {
	case c.SAN >= 80:
		return "思緒清晰，一切正常" // Clear
	case c.SAN >= 50:
		return "心跳加速，開始冒冷汗" // Anxious (matches SANStateAnxious.Description)
	case c.SAN >= 20:
		return "視覺扭曲，感官欺騙" // Panic (matches SANStatePanic.Description)
	default:
		return "嚴重幻覺，失去部分控制" // Breakdown (matches SANStateBreakdown.Description)
	}
}
