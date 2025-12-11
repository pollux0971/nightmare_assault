// Package effects provides visual and interactive horror effects for the Nightmare Assault TUI.
package effects

import (
	"math/rand"
)

// HorrorStyle defines the visual and interactive horror effects based on player's SAN value.
// Each field controls a different aspect of the horror experience.
type HorrorStyle struct {
	TextCorruption    float64 // 0.0 (none) - 1.0 (complete distortion)
	TypingBehavior    float64 // 0.0 (normal) - 0.2 (20% deletion rate)
	ColorShift        int     // 0-360 degree hue shift
	UIStability       int     // 0-5 pixel shake amplitude
	OptionReliability float64 // 0.0 (completely unreliable) - 1.0 (fully reliable) - for Story 6.2
}

// CalculateHorrorStyle returns the horror effect parameters based on the player's current SAN value.
// The effects escalate as SAN decreases, creating increasing psychological pressure.
//
// SAN Ranges:
//   - 80-100: Clear-headed (no effects)
//   - 60-79: Slightly anxious (minor visual disturbances)
//   - 40-59: Anxious (noticeable effects)
//   - 20-39: Panicked (severe distortions)
//   - 1-19: Insanity (extreme effects)
func CalculateHorrorStyle(san int) HorrorStyle {
	// Clamp SAN to valid range
	if san < 0 {
		san = 0
	}
	if san > 100 {
		san = 100
	}

	// SAN 80-100: Clear-headed - no effects
	if san >= 80 {
		return HorrorStyle{
			TextCorruption:    0.0,
			TypingBehavior:    0.0,
			ColorShift:        0,
			UIStability:       0,
			OptionReliability: 1.0,
		}
	}

	// SAN 60-79: Slightly anxious - minor effects
	if san >= 60 {
		return HorrorStyle{
			TextCorruption:    0.1,
			TypingBehavior:    0.0,
			ColorShift:        5 + rand.Intn(6), // 5-10
			UIStability:       0,
			OptionReliability: 0.95,
		}
	}

	// SAN 40-59: Anxious - noticeable effects
	if san >= 40 {
		return HorrorStyle{
			TextCorruption:    0.3,
			TypingBehavior:    0.0,
			ColorShift:        15 + rand.Intn(16), // 15-30
			UIStability:       1,
			OptionReliability: 0.85,
		}
	}

	// SAN 20-39: Panicked - severe effects
	if san >= 20 {
		return HorrorStyle{
			TextCorruption:    0.6,
			TypingBehavior:    0.05,
			ColorShift:        45 + rand.Intn(46), // 45-90
			UIStability:       2 + rand.Intn(2),   // 2-3
			OptionReliability: 0.7,
		}
	}

	// SAN 1-19: Insanity - extreme effects
	return HorrorStyle{
		TextCorruption:    0.9,
		TypingBehavior:    0.15,
		ColorShift:        120 + rand.Intn(61), // 120-180
		UIStability:       4 + rand.Intn(2),    // 4-5
		OptionReliability: 0.5,
	}
}
