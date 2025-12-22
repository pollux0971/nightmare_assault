package guardian

import (
	"fmt"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// TensionAdjustment represents a single tension adjustment event
type TensionAdjustment struct {
	Timestamp time.Time
	Delta     int
	Reason    string
	OldValue  int
	NewValue  int
}

// TensionManager manages tension curve adjustments and integrates with MomentumController
type TensionManager struct {
	config               GuardianConfig
	momentumController   *momentum.MomentumController
	adjustmentHistory    []TensionAdjustment
	minTensionThreshold  int // Minimum tension value to maintain
	consecutiveDeaths    int
	lowHPSANStreak       int

	// Story 9-11: Bidirectional sync with MomentumController
	lastKnownTension float64
	lastKnownPhase   string
	mu               sync.RWMutex
}

// TensionManagerConfig contains configuration for TensionManager
type TensionManagerConfig struct {
	MinTensionThreshold int // Default: 10 - prevents tension from dropping too low
}

// DefaultTensionManagerConfig returns the default TensionManager configuration
func DefaultTensionManagerConfig() TensionManagerConfig {
	return TensionManagerConfig{
		MinTensionThreshold: 10,
	}
}

// NewTensionManager creates a new TensionManager
func NewTensionManager(
	guardianConfig GuardianConfig,
	momentumController *momentum.MomentumController,
	config TensionManagerConfig,
) *TensionManager {
	return &TensionManager{
		config:              guardianConfig,
		momentumController:  momentumController,
		adjustmentHistory:   make([]TensionAdjustment, 0),
		minTensionThreshold: config.MinTensionThreshold,
		consecutiveDeaths:   0,
		lowHPSANStreak:      0,
	}
}

// GetCurrentPhase returns the current tension phase based on game state
func (tm *TensionManager) GetCurrentPhase(gameState *engine.GameStateV2) TensionPhase {
	if gameState == nil || gameState.Tension == nil {
		logger.Warn("TensionManager GetCurrentPhase called with nil gameState or Tension", nil)
		return PhaseRest
	}

	tension := gameState.Tension.GetValue()
	return calculatePhaseFromTension(tension)
}

// ShouldReduceTension determines if tension should be reduced based on Guardian triggers
func (tm *TensionManager) ShouldReduceTension(gameState *engine.GameStateV2) (bool, string) {
	if gameState == nil {
		logger.Warn("TensionManager ShouldReduceTension called with nil gameState", nil)
		return false, ""
	}

	// Check consecutive deaths
	if tm.consecutiveDeaths >= tm.config.MaxConsecutiveDeaths {
		logger.Info("TensionManager: Should reduce tension due to consecutive deaths", map[string]interface{}{
			"consecutive_deaths":     tm.consecutiveDeaths,
			"max_consecutive_deaths": tm.config.MaxConsecutiveDeaths,
		})
		return true, "consecutive_deaths"
	}

	// Check low HP/SAN streak
	if tm.lowHPSANStreak >= tm.config.LowStatStreakLimit {
		hp := gameState.GetHP()
		san := gameState.GetSAN()

		logger.Info("TensionManager: Should reduce tension due to low HP/SAN streak", map[string]interface{}{
			"low_hp_san_streak":   tm.lowHPSANStreak,
			"low_stat_streak_limit": tm.config.LowStatStreakLimit,
			"current_hp":          hp,
			"current_san":         san,
		})
		return true, "low_hp_san_streak"
	}

	return false, ""
}

// AdjustTension adjusts the tension value by the given delta
// Respects minimum threshold and records adjustment history
func (tm *TensionManager) AdjustTension(gameState *engine.GameStateV2, delta int, reason string) bool {
	if gameState == nil || gameState.Tension == nil {
		logger.Warn("TensionManager AdjustTension called with nil gameState or Tension", nil)
		return false
	}

	oldValue := gameState.Tension.GetValue()
	newValue := oldValue + delta

	// Apply minimum threshold - prevent tension from dropping too low
	if newValue < tm.minTensionThreshold {
		newValue = tm.minTensionThreshold
		logger.Debug("TensionManager: Tension adjustment clamped to minimum threshold", map[string]interface{}{
			"old_value":            oldValue,
			"requested_new_value":  oldValue + delta,
			"clamped_new_value":    newValue,
			"min_threshold":        tm.minTensionThreshold,
		})
	}

	// Apply the adjustment
	actualDelta := newValue - oldValue
	if actualDelta == 0 {
		logger.Debug("TensionManager: No tension adjustment needed (already at threshold)", map[string]interface{}{
			"current_value": oldValue,
			"reason":        reason,
		})
		return false
	}

	gameState.Tension.SetValue(newValue)

	// Record adjustment in history
	adjustment := TensionAdjustment{
		Timestamp: time.Now(),
		Delta:     actualDelta,
		Reason:    reason,
		OldValue:  oldValue,
		NewValue:  newValue,
	}
	tm.adjustmentHistory = append(tm.adjustmentHistory, adjustment)

	// Keep only last 50 adjustments to prevent unbounded growth
	if len(tm.adjustmentHistory) > 50 {
		tm.adjustmentHistory = tm.adjustmentHistory[len(tm.adjustmentHistory)-50:]
	}

	logger.Info("TensionManager: Tension adjusted", map[string]interface{}{
		"old_value": oldValue,
		"new_value": newValue,
		"delta":     actualDelta,
		"reason":    reason,
	})

	return true
}

// OnTurnEnd updates the TensionManager state at the end of each turn
func (tm *TensionManager) OnTurnEnd(gameState *engine.GameStateV2) {
	if gameState == nil {
		logger.Warn("TensionManager OnTurnEnd called with nil gameState", nil)
		return
	}

	// Update consecutive death counter
	isDead := gameState.IsDead()
	if isDead {
		tm.consecutiveDeaths++
		logger.Debug("TensionManager: Death detected", map[string]interface{}{
			"consecutive_deaths": tm.consecutiveDeaths,
		})
	} else {
		// Reset on survival
		if tm.consecutiveDeaths > 0 {
			logger.Debug("TensionManager: Player survived, resetting death counter", map[string]interface{}{
				"previous_consecutive_deaths": tm.consecutiveDeaths,
			})
			tm.consecutiveDeaths = 0
		}
	}

	// Update low HP/SAN streak
	hp := gameState.GetHP()
	san := gameState.GetSAN()
	hasLowHP := hp <= tm.config.LowHPThreshold
	hasLowSAN := san <= tm.config.LowSANThreshold

	if hasLowHP || hasLowSAN {
		tm.lowHPSANStreak++
		logger.Debug("TensionManager: Low HP/SAN detected", map[string]interface{}{
			"hp":                hp,
			"san":               san,
			"low_hp_san_streak": tm.lowHPSANStreak,
			"has_low_hp":        hasLowHP,
			"has_low_san":       hasLowSAN,
		})
	} else {
		// Reset streak on recovery
		if tm.lowHPSANStreak > 0 {
			logger.Debug("TensionManager: Player HP/SAN recovered, resetting streak", map[string]interface{}{
				"previous_streak": tm.lowHPSANStreak,
			})
			tm.lowHPSANStreak = 0
		}
	}

	// Check if tension should be reduced
	shouldReduce, reason := tm.ShouldReduceTension(gameState)
	if shouldReduce {
		// Calculate reduction amount based on current phase
		currentPhase := tm.GetCurrentPhase(gameState)
		reduction := tm.calculateTensionReduction(currentPhase, reason)

		if reduction < 0 {
			tm.AdjustTension(gameState, reduction, reason)
		}
	}
}

// calculateTensionReduction calculates appropriate tension reduction based on phase and reason
func (tm *TensionManager) calculateTensionReduction(phase TensionPhase, reason string) int {
	// Base reductions by reason
	var baseReduction int
	switch reason {
	case "consecutive_deaths":
		// More aggressive reduction for consecutive deaths
		baseReduction = -15
	case "low_hp_san_streak":
		// Moderate reduction for low HP/SAN
		baseReduction = -10
	default:
		baseReduction = -5
	}

	// Scale by phase - reduce more in higher tension phases
	var phaseMultiplier float64
	switch phase {
	case PhaseRest:
		phaseMultiplier = 0.5 // Reduce less in rest phase
	case PhaseBuildup:
		phaseMultiplier = 0.75
	case PhasePeak:
		phaseMultiplier = 1.0 // Full reduction in peak phase
	case PhaseRelease:
		phaseMultiplier = 1.2 // Slightly more in release phase
	default:
		phaseMultiplier = 1.0
	}

	finalReduction := int(float64(baseReduction) * phaseMultiplier)

	logger.Debug("TensionManager: Calculated tension reduction", map[string]interface{}{
		"reason":            reason,
		"phase":             phase.String(),
		"base_reduction":    baseReduction,
		"phase_multiplier":  phaseMultiplier,
		"final_reduction":   finalReduction,
	})

	return finalReduction
}

// GetAdjustmentHistory returns a copy of the adjustment history
func (tm *TensionManager) GetAdjustmentHistory() []TensionAdjustment {
	// Return a copy to prevent external modification
	history := make([]TensionAdjustment, len(tm.adjustmentHistory))
	copy(history, tm.adjustmentHistory)
	return history
}

// GetConsecutiveDeaths returns the current consecutive death count
func (tm *TensionManager) GetConsecutiveDeaths() int {
	return tm.consecutiveDeaths
}

// GetLowHPSANStreak returns the current low HP/SAN streak count
func (tm *TensionManager) GetLowHPSANStreak() int {
	return tm.lowHPSANStreak
}

// ResetProtectionState resets the tension manager state
func (tm *TensionManager) ResetProtectionState() {
	logger.Debug("TensionManager: Resetting protection state", map[string]interface{}{
		"previous_consecutive_deaths": tm.consecutiveDeaths,
		"previous_low_hp_san_streak":  tm.lowHPSANStreak,
	})

	tm.consecutiveDeaths = 0
	tm.lowHPSANStreak = 0
}

// IntegrateWithMomentumController updates MomentumController configuration based on current tension
// This ensures the momentum system respects the guardian's tension adjustments
func (tm *TensionManager) IntegrateWithMomentumController(gameState *engine.GameStateV2) {
	tm.mu.RLock()
	momentumController := tm.momentumController
	tm.mu.RUnlock()

	if momentumController == nil {
		logger.Debug("TensionManager: No MomentumController configured, skipping integration", nil)
		return
	}

	if gameState == nil || gameState.Tension == nil {
		logger.Warn("TensionManager: Cannot integrate with MomentumController - nil gameState or Tension", nil)
		return
	}

	currentPhase := tm.GetCurrentPhase(gameState)
	config := momentumController.GetConfig()
	if config == nil {
		logger.Warn("TensionManager: MomentumController has nil config", nil)
		return
	}

	// Adjust momentum settings based on tension phase
	// In lower tension phases, allow more auto-resolution
	// In higher tension phases, pause more frequently
	originalFrequency := config.Frequency
	originalPauseOnRisk := config.PauseOnRisk

	switch currentPhase {
	case PhaseRest:
		// Low tension - allow more auto-play
		config.Frequency = momentum.FrequencyLow
		config.PauseOnRisk = momentum.RiskHigh
	case PhaseBuildup:
		// Medium tension - balanced
		config.Frequency = momentum.FrequencyMedium
		config.PauseOnRisk = momentum.RiskMedium
	case PhasePeak, PhaseRelease:
		// High tension - more player control
		config.Frequency = momentum.FrequencyHigh
		config.PauseOnRisk = momentum.RiskLow
	}

	// Only log if settings changed
	if originalFrequency != config.Frequency || originalPauseOnRisk != config.PauseOnRisk {
		logger.Info("TensionManager: Updated MomentumController settings", map[string]interface{}{
			"tension_phase":       currentPhase.String(),
			"frequency_from":      originalFrequency.String(),
			"frequency_to":        config.Frequency.String(),
			"pause_on_risk_from":  originalPauseOnRisk.String(),
			"pause_on_risk_to":    config.PauseOnRisk.String(),
		})

		momentumController.SetConfig(config)
	}
}

// =============================================================================
// Story 9-11: Bidirectional Integration with MomentumController
// =============================================================================

// SyncFromMomentum synchronizes tension state from MomentumController
// This method only updates internal state and does NOT call back to MomentumController
// to avoid circular calls.
// AC2: Guardian receives automatic tension change notifications
func (tm *TensionManager) SyncFromMomentum(tensionValue float64, phase string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	oldTension := tm.lastKnownTension
	oldPhase := tm.lastKnownPhase

	tm.lastKnownTension = tensionValue
	tm.lastKnownPhase = phase

	// Log only if values changed significantly
	if oldPhase != phase || (tensionValue-oldTension) > 0.05 || (tensionValue-oldTension) < -0.05 {
		logger.Debug("TensionManager: Synced from MomentumController", map[string]interface{}{
			"old_tension": oldTension,
			"new_tension": tensionValue,
			"old_phase":   oldPhase,
			"new_phase":   phase,
		})
	}
}

// GetLastKnownTension returns the last known tension value from MomentumController
func (tm *TensionManager) GetLastKnownTension() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.lastKnownTension
}

// GetLastKnownPhase returns the last known phase from MomentumController
func (tm *TensionManager) GetLastKnownPhase() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.lastKnownPhase
}

// BindToMomentumController establishes bidirectional binding with MomentumController
// This is a convenience method that sets up both directions of the integration.
// If TensionManager was previously bound to another controller, it will be unbound first.
// AC1: Bidirectional references established
func (tm *TensionManager) BindToMomentumController(mc *momentum.MomentumController) error {
	if mc == nil {
		return fmt.Errorf("momentum controller cannot be nil")
	}

	tm.mu.Lock()
	// Unbind previous controller if any
	if tm.momentumController != nil {
		oldController := tm.momentumController
		tm.mu.Unlock()
		oldController.SetTensionManager(nil)
		tm.mu.Lock()
	}

	tm.momentumController = mc
	tm.mu.Unlock()

	// Set up bidirectional reference
	mc.SetTensionManager(tm)

	// Initial sync - read current state from MomentumController
	currentTension := mc.GetTension()
	currentPhase := mc.GetCurrentPhase()
	tm.SyncFromMomentum(currentTension, currentPhase)

	logger.Info("TensionManager: Bound to MomentumController", map[string]interface{}{
		"initial_tension": currentTension,
		"initial_phase":   currentPhase,
	})

	return nil
}

// AdjustMomentumTension adjusts tension in the MomentumController
// This is the method Guardian uses to actively change MomentumController's tension
// AC3: Guardian can adjust MomentumController tension
func (tm *TensionManager) AdjustMomentumTension(delta float64, reason string) error {
	tm.mu.RLock()
	momentumController := tm.momentumController
	tm.mu.RUnlock()

	if momentumController == nil {
		return fmt.Errorf("no momentum controller bound")
	}

	oldTension := momentumController.GetTension()

	// Call AdjustTension on MomentumController
	// This will automatically sync back to us via SyncFromMomentum
	momentumController.AdjustTension(delta)

	newTension := momentumController.GetTension()

	logger.Info("TensionManager: Adjusted MomentumController tension", map[string]interface{}{
		"reason":      reason,
		"old_tension": oldTension,
		"delta":       delta,
		"new_tension": newTension,
	})

	return nil
}
