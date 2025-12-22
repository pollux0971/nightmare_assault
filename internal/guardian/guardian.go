package guardian

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ExperienceGuardian monitors player status and activates protection when needed.
// It tracks consecutive deaths and low HP/SAN streaks to prevent player frustration.
type ExperienceGuardian struct {
	config GuardianConfig

	// consecutiveDeaths tracks the number of consecutive deaths since last successful survival.
	consecutiveDeaths int

	// lowHPSANStreak tracks the number of consecutive turns with low HP or SAN.
	lowHPSANStreak int

	// lastDead tracks whether the player was dead in the last turn.
	// Used to detect consecutive deaths.
	lastDead bool

	// protectionActive tracks whether protection is currently active.
	protectionActive bool

	// currentPhase tracks the current tension curve phase.
	currentPhase TensionPhase
}

// NewExperienceGuardian creates a new Experience Guardian with the given configuration.
func NewExperienceGuardian(cfg GuardianConfig) *ExperienceGuardian {
	return &ExperienceGuardian{
		config:            cfg,
		consecutiveDeaths: 0,
		lowHPSANStreak:    0,
		lastDead:          false,
		protectionActive:  false,
		currentPhase:      PhaseRest,
	}
}

// OnTurnEnd updates the Guardian state at the end of each turn.
// This method should be called after each game turn to track player status.
func (g *ExperienceGuardian) OnTurnEnd(gameState *engine.GameStateV2) {
	if gameState == nil {
		logger.Warn("Guardian OnTurnEnd called with nil gameState", nil)
		return
	}

	// Check if player is dead
	isDead := gameState.IsDead()

	// Update consecutive death counter
	if isDead {
		if g.lastDead {
			// Player was dead last turn and is still dead - this is a consecutive death
			g.consecutiveDeaths++
			logger.Debug("Guardian: Consecutive death detected", map[string]interface{}{
				"consecutive_deaths": g.consecutiveDeaths,
			})
		} else {
			// First death
			g.consecutiveDeaths = 1
			logger.Debug("Guardian: First death detected", nil)
		}
		g.lastDead = true
	} else {
		// Player is alive
		if g.lastDead {
			// Player was dead but is now alive - reset death counter
			logger.Debug("Guardian: Player recovered from death, resetting death counter", nil)
		}
		g.lastDead = false
		g.consecutiveDeaths = 0
	}

	// Check player HP/SAN status
	g.checkPlayerStatus(gameState)

	// Update tension phase tracking
	g.updateTensionPhase(gameState)
}

// checkPlayerStatus checks if player has low HP or SAN and updates the low stat streak.
func (g *ExperienceGuardian) checkPlayerStatus(gameState *engine.GameStateV2) {
	hp := gameState.GetHP()
	san := gameState.GetSAN()

	// Check if player has low HP or SAN
	hasLowHP := hp <= g.config.LowHPThreshold
	hasLowSAN := san <= g.config.LowSANThreshold

	if hasLowHP || hasLowSAN {
		g.lowHPSANStreak++
		logger.Debug("Guardian: Low HP/SAN detected", map[string]interface{}{
			"hp":                hp,
			"san":               san,
			"low_hp_san_streak": g.lowHPSANStreak,
			"has_low_hp":        hasLowHP,
			"has_low_san":       hasLowSAN,
		})
	} else {
		// Reset streak if player recovers
		if g.lowHPSANStreak > 0 {
			logger.Debug("Guardian: Player HP/SAN recovered, resetting streak", map[string]interface{}{
				"previous_streak": g.lowHPSANStreak,
			})
		}
		g.lowHPSANStreak = 0
	}
}

// ShouldActivateProtection checks if Guardian protection should be activated.
// Returns true if any protection condition is met:
// - Consecutive deaths >= MaxConsecutiveDeaths
// - Low HP/SAN streak >= LowStatStreakLimit
func (g *ExperienceGuardian) ShouldActivateProtection(gameState *engine.GameStateV2) bool {
	if gameState == nil {
		logger.Warn("Guardian ShouldActivateProtection called with nil gameState", nil)
		return false
	}

	// Check consecutive death condition
	if g.consecutiveDeaths >= g.config.MaxConsecutiveDeaths {
		logger.Info("Guardian: Protection triggered by consecutive deaths", map[string]interface{}{
			"consecutive_deaths":     g.consecutiveDeaths,
			"max_consecutive_deaths": g.config.MaxConsecutiveDeaths,
		})
		g.protectionActive = true
		return true
	}

	// Check low HP/SAN streak condition
	if g.lowHPSANStreak >= g.config.LowStatStreakLimit {
		logger.Info("Guardian: Protection triggered by low HP/SAN streak", map[string]interface{}{
			"low_hp_san_streak":   g.lowHPSANStreak,
			"low_stat_streak_limit": g.config.LowStatStreakLimit,
			"hp":                  gameState.GetHP(),
			"san":                 gameState.GetSAN(),
		})
		g.protectionActive = true
		return true
	}

	return false
}

// ResetProtectionState resets the Guardian protection state.
// Should be called when protection is applied or when player recovers.
func (g *ExperienceGuardian) ResetProtectionState() {
	logger.Debug("Guardian: Resetting protection state", map[string]interface{}{
		"previous_consecutive_deaths": g.consecutiveDeaths,
		"previous_low_hp_san_streak":  g.lowHPSANStreak,
		"was_protection_active":       g.protectionActive,
	})

	g.consecutiveDeaths = 0
	g.lowHPSANStreak = 0
	g.lastDead = false
	g.protectionActive = false
}

// IsProtectionActive returns whether protection is currently active.
func (g *ExperienceGuardian) IsProtectionActive() bool {
	return g.protectionActive
}

// GetConsecutiveDeaths returns the current consecutive death count.
func (g *ExperienceGuardian) GetConsecutiveDeaths() int {
	return g.consecutiveDeaths
}

// GetLowHPSANStreak returns the current low HP/SAN streak count.
func (g *ExperienceGuardian) GetLowHPSANStreak() int {
	return g.lowHPSANStreak
}

// GetConfig returns the Guardian configuration.
func (g *ExperienceGuardian) GetConfig() GuardianConfig {
	return g.config
}
