package manager

import (
	"math/rand"
)

// Mental state transition constants (per Story 1.5)
const (
	// BaseBreakdownChance is the base probability of breakdown
	BaseBreakdownChance = 0.05 // 5% base chance

	// StressMultiplier determines how much each point of stress above 90 increases breakdown chance
	StressMultiplier = 0.01 // 1% per stress point above 90

	// FearMultiplier determines how much each point of fear above 60 increases breakdown chance
	FearMultiplier = 0.005 // 0.5% per fear point above 60
)

// checkMentalStateTransition determines the next mental state based on current state and emotions.
// This implements the state machine pattern for NPC mental states.
//
// Transition rules (per Story 1.5):
// - Normal -> Anxious: Stress >= 60 OR Fear >= 70
// - Anxious -> Normal: Stress < 40 AND Fear < 50
// - Anxious -> Corrupted: Stress >= 90 AND rollBreakdown() succeeds
// - Corrupted: Irreversible (no transitions out)
func checkMentalStateTransition(current MentalState, emotion EmotionState) MentalState {
	switch current {
	case Normal:
		// Check for transition to Anxious
		if emotion.Stress >= 60 || emotion.Fear >= 70 {
			return Anxious
		}
		return Normal

	case Anxious:
		// Check for recovery to Normal
		if emotion.Stress < 40 && emotion.Fear < 50 {
			return Normal
		}

		// Check for breakdown to Corrupted
		if emotion.Stress >= 90 && rollBreakdown(emotion.Stress, emotion.Fear) {
			return Corrupted
		}
		return Anxious

	case Corrupted:
		// Corrupted state is irreversible - no transitions out
		return Corrupted

	default:
		return Normal
	}
}

// rollBreakdown performs a breakdown check for NPCs in Anxious state with high stress.
// Returns true if the NPC should transition to Corrupted state.
//
// Formula (per Story 1.5):
// chance := BaseBreakdownChance + (stress-90)*StressMultiplier + (fear-60)*FearMultiplier
// return rand.Float64() < chance
func rollBreakdown(stress, fear int) bool {
	// Calculate breakdown chance using the formula from Story 1.5
	chance := BaseBreakdownChance +
		float64(stress-90)*StressMultiplier +
		float64(fear-60)*FearMultiplier

	// If chance is negative or zero, no breakdown occurs
	if chance <= 0 {
		return false
	}

	// Perform random roll
	return rand.Float64() < chance
}
