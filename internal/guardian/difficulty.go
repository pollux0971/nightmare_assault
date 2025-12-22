package guardian

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// DifficultyModifiers represents the current difficulty adjustments.
type DifficultyModifiers struct {
	// DamageReduction is the percentage reduction in damage output (0.0 to 0.3 = 0% to 30% reduction)
	DamageReduction float64

	// CheckBonus is the flat bonus added to skill checks (+0 to +15)
	CheckBonus int

	// EncounterFrequencyReduction is the percentage reduction in hostile encounters (0.0 to 0.5 = 0% to 50% reduction)
	EncounterFrequencyReduction float64

	// IsActive indicates whether modifiers are currently active
	IsActive bool
}

// DifficultyTuner manages dynamic difficulty adjustments based on player state.
type DifficultyTuner struct {
	guardian  *ExperienceGuardian
	modifiers DifficultyModifiers

	// config stores tuning parameters
	config DifficultyTuningConfig
}

// DifficultyTuningConfig configures the difficulty tuning behavior.
type DifficultyTuningConfig struct {
	// MaxDamageReduction is the maximum damage reduction percentage (default: 0.3 = 30%)
	MaxDamageReduction float64

	// MaxCheckBonus is the maximum bonus to skill checks (default: 15)
	MaxCheckBonus int

	// MaxEncounterReduction is the maximum encounter frequency reduction (default: 0.5 = 50%)
	MaxEncounterReduction float64

	// MinSurvivalRateForReduction is the survival rate below which reductions start (default: 0.5)
	MinSurvivalRateForReduction float64
}

// DefaultDifficultyTuningConfig returns the default difficulty tuning configuration.
func DefaultDifficultyTuningConfig() DifficultyTuningConfig {
	return DifficultyTuningConfig{
		MaxDamageReduction:          0.3,
		MaxCheckBonus:               15,
		MaxEncounterReduction:       0.5,
		MinSurvivalRateForReduction: 0.5,
	}
}

// NewDifficultyTuner creates a new difficulty tuner.
func NewDifficultyTuner(guardian *ExperienceGuardian) *DifficultyTuner {
	return &DifficultyTuner{
		guardian:  guardian,
		modifiers: DifficultyModifiers{},
		config:    DefaultDifficultyTuningConfig(),
	}
}

// NewDifficultyTunerWithConfig creates a new difficulty tuner with custom configuration.
func NewDifficultyTunerWithConfig(guardian *ExperienceGuardian, config DifficultyTuningConfig) *DifficultyTuner {
	return &DifficultyTuner{
		guardian:  guardian,
		modifiers: DifficultyModifiers{},
		config:    config,
	}
}

// AdjustDifficulty calculates and applies difficulty adjustments based on the current game state.
// This method should be called when Guardian protection is activated.
//
// Adjustment strategy:
// - If survival rate < 0.5 (struggling): Apply progressive difficulty reductions
// - If survival rate >= 0.5 (doing well): No reductions (standard difficulty)
//
// Returns the calculated modifiers.
func (dt *DifficultyTuner) AdjustDifficulty(gameState *engine.GameStateV2) DifficultyModifiers {
	if gameState == nil {
		logger.Warn("DifficultyTuner: AdjustDifficulty called with nil gameState", nil)
		return dt.modifiers
	}

	// Check if difficulty tuning is enabled
	if !dt.guardian.config.EnableDifficultyTuning {
		logger.Debug("DifficultyTuner: Difficulty tuning is disabled", nil)
		dt.modifiers = DifficultyModifiers{IsActive: false}
		return dt.modifiers
	}

	// Calculate survival rate
	hp := gameState.GetHP()
	san := gameState.GetSAN()
	survivalRate := (float64(hp)/100.0 + float64(san)/100.0) / 2.0

	// Only apply reductions if player is struggling
	if survivalRate >= dt.config.MinSurvivalRateForReduction {
		logger.Debug("DifficultyTuner: Player survival rate is healthy, no adjustments", map[string]interface{}{
			"survival_rate": survivalRate,
			"threshold":     dt.config.MinSurvivalRateForReduction,
		})
		dt.modifiers = DifficultyModifiers{IsActive: false}
		return dt.modifiers
	}

	// Calculate adjustment strength based on how far below threshold
	// If survival rate = 0.0, adjustment = 1.0 (maximum help)
	// If survival rate = 0.5, adjustment = 0.0 (no help)
	adjustmentStrength := (dt.config.MinSurvivalRateForReduction - survivalRate) / dt.config.MinSurvivalRateForReduction

	// Apply progressive difficulty reductions
	dt.modifiers = DifficultyModifiers{
		DamageReduction:             adjustmentStrength * dt.config.MaxDamageReduction,
		CheckBonus:                  int(adjustmentStrength * float64(dt.config.MaxCheckBonus)),
		EncounterFrequencyReduction: adjustmentStrength * dt.config.MaxEncounterReduction,
		IsActive:                    true,
	}

	logger.Info("DifficultyTuner: Difficulty adjustments applied", map[string]interface{}{
		"hp":                      hp,
		"san":                     san,
		"survival_rate":           survivalRate,
		"adjustment_strength":     adjustmentStrength,
		"damage_reduction":        dt.modifiers.DamageReduction,
		"check_bonus":             dt.modifiers.CheckBonus,
		"encounter_reduction":     dt.modifiers.EncounterFrequencyReduction,
		"consecutive_deaths":      dt.guardian.GetConsecutiveDeaths(),
		"low_hp_san_streak":       dt.guardian.GetLowHPSANStreak(),
		"protection_active":       dt.guardian.IsProtectionActive(),
	})

	return dt.modifiers
}

// GetModifiers returns the current difficulty modifiers.
func (dt *DifficultyTuner) GetModifiers() DifficultyModifiers {
	return dt.modifiers
}

// ResetModifiers resets all difficulty modifiers to default (no adjustments).
func (dt *DifficultyTuner) ResetModifiers() {
	logger.Debug("DifficultyTuner: Resetting difficulty modifiers", map[string]interface{}{
		"previous_damage_reduction":    dt.modifiers.DamageReduction,
		"previous_check_bonus":         dt.modifiers.CheckBonus,
		"previous_encounter_reduction": dt.modifiers.EncounterFrequencyReduction,
		"was_active":                   dt.modifiers.IsActive,
	})

	dt.modifiers = DifficultyModifiers{
		DamageReduction:             0,
		CheckBonus:                  0,
		EncounterFrequencyReduction: 0,
		IsActive:                    false,
	}
}

// ApplyDamageModifier applies the damage reduction modifier to a damage value.
// Returns the adjusted damage amount (reduced if modifiers are active).
func (dt *DifficultyTuner) ApplyDamageModifier(damage int) int {
	if !dt.modifiers.IsActive {
		return damage
	}

	reduction := float64(damage) * dt.modifiers.DamageReduction
	adjustedDamage := int(float64(damage) - reduction)

	if adjustedDamage < 0 {
		adjustedDamage = 0
	}

	logger.Debug("DifficultyTuner: Applied damage modifier", map[string]interface{}{
		"original_damage":  damage,
		"reduction_amount": int(reduction),
		"adjusted_damage":  adjustedDamage,
	})

	return adjustedDamage
}

// ApplyCheckModifier applies the check bonus to a difficulty check.
// Returns the adjusted difficulty value (reduced if modifiers are active).
func (dt *DifficultyTuner) ApplyCheckModifier(difficulty int) int {
	if !dt.modifiers.IsActive {
		return difficulty
	}

	adjustedDifficulty := difficulty - dt.modifiers.CheckBonus

	if adjustedDifficulty < 1 {
		adjustedDifficulty = 1 // Minimum difficulty
	}

	logger.Debug("DifficultyTuner: Applied check modifier", map[string]interface{}{
		"original_difficulty": difficulty,
		"check_bonus":         dt.modifiers.CheckBonus,
		"adjusted_difficulty": adjustedDifficulty,
	})

	return adjustedDifficulty
}

// ShouldReduceEncounter determines if a hostile encounter should be skipped.
// Returns true if the encounter should be reduced (skipped) based on modifiers.
func (dt *DifficultyTuner) ShouldReduceEncounter(encounterChance float64) bool {
	if !dt.modifiers.IsActive {
		return false
	}

	// If encounter frequency is reduced by 50%, then 50% of encounters should be skipped
	shouldReduce := encounterChance < dt.modifiers.EncounterFrequencyReduction

	if shouldReduce {
		logger.Debug("DifficultyTuner: Encounter reduced", map[string]interface{}{
			"encounter_chance":     encounterChance,
			"reduction_threshold":  dt.modifiers.EncounterFrequencyReduction,
		})
	}

	return shouldReduce
}
