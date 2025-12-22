package guardian

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// GetDifficultyAdjustment calculates a difficulty adjustment multiplier based on player survival state.
// Returns a value in range [0.3, 1.2]:
// - 0.3: Maximum difficulty reduction (70% easier) for struggling players
// - 1.0: No adjustment (standard difficulty)
// - 1.2: Increased difficulty (20% harder) for players doing well
//
// The adjustment is calculated using survival rate:
// survivalRate = (HP/100 + SAN/100) / 2
// adjustment = 0.3 + 0.9 * survivalRate
//
// If EnableDifficultyTuning is false, always returns 1.0 (no adjustment).
func (g *ExperienceGuardian) GetDifficultyAdjustment(gameState *engine.GameStateV2) float64 {
	// Check if difficulty tuning is enabled
	if !g.config.EnableDifficultyTuning {
		return 1.0
	}

	if gameState == nil {
		logger.Warn("Guardian GetDifficultyAdjustment called with nil gameState", nil)
		return 1.0
	}

	// Get player stats
	hp := gameState.GetHP()
	san := gameState.GetSAN()

	// Calculate survival rate (0.0 to 1.0)
	// Normalize HP and SAN to 0-1 range, then average them
	hpRate := float64(hp) / 100.0
	sanRate := float64(san) / 100.0
	survivalRate := (hpRate + sanRate) / 2.0

	// Calculate adjustment: 0.3 + 0.9 * survivalRate
	// This maps:
	//   survivalRate 0.0 -> adjustment 0.3 (maximum help)
	//   survivalRate 0.5 -> adjustment 0.75 (slight help)
	//   survivalRate 1.0 -> adjustment 1.2 (increased challenge)
	adjustment := 0.3 + 0.9*survivalRate

	logger.Debug("Guardian: Difficulty adjustment calculated", map[string]interface{}{
		"hp":            hp,
		"san":           san,
		"survival_rate": survivalRate,
		"adjustment":    adjustment,
	})

	return adjustment
}
