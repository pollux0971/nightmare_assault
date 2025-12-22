package guardian

import (
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// TensionPhase represents the tension curve phase of the game
type TensionPhase int

const (
	// PhaseRest represents low tension rest period (0-25)
	PhaseRest TensionPhase = iota
	// PhaseBuildup represents tension buildup period (25-60)
	PhaseBuildup
	// PhasePeak represents high tension peak period (60-90)
	PhasePeak
	// PhaseRelease represents tension release period (90-100)
	PhaseRelease
)

// String returns the string representation of TensionPhase
func (tp TensionPhase) String() string {
	switch tp {
	case PhaseRest:
		return "Rest"
	case PhaseBuildup:
		return "Buildup"
	case PhasePeak:
		return "Peak"
	case PhaseRelease:
		return "Release"
	default:
		return fmt.Sprintf("Unknown(%d)", tp)
	}
}

// updateTensionPhase updates the current tension phase based on game state
// Returns true if the phase changed, false otherwise
func (g *ExperienceGuardian) updateTensionPhase(gameState *engine.GameStateV2) bool {
	if gameState == nil || gameState.Tension == nil {
		logger.Warn("Guardian updateTensionPhase called with nil gameState or Tension", nil)
		return false
	}

	tension := gameState.Tension.GetValue()
	newPhase := calculatePhaseFromTension(tension)

	// Check if phase changed
	if newPhase != g.currentPhase {
		oldPhase := g.currentPhase
		g.currentPhase = newPhase

		logger.Info("Guardian: Tension phase changed", map[string]interface{}{
			"old_phase":      oldPhase.String(),
			"new_phase":      newPhase.String(),
			"tension":        tension,
		})

		return true
	}

	return false
}

// calculatePhaseFromTension determines the tension phase based on tension value
func calculatePhaseFromTension(tension int) TensionPhase {
	switch {
	case tension < 25:
		return PhaseRest
	case tension < 60:
		return PhaseBuildup
	case tension < 90:
		return PhasePeak
	default:
		return PhaseRelease
	}
}

// GetCurrentPhase returns the current tension phase
func (g *ExperienceGuardian) GetCurrentPhase() TensionPhase {
	return g.currentPhase
}
