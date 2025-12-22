package momentum

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// Plot point type constants
const (
	PlotPointActTransition   = "act_transition"
	PlotPointSeedMilestone30 = "seed_milestone_30"
	PlotPointSeedMilestone60 = "seed_milestone_60"
	PlotPointNPCDeath        = "npc_death"
	PlotPointFirstViolation  = "first_violation"
)

// Act beat ranges (can be configured later)
const (
	Act1Start    = 0
	Act1End      = 30
	Act2Start    = 31
	Act2End      = 60
	Act3Start    = 61
	Act3End      = 100
	ActEndBuffer = 3 // Consider "near end" if within this many beats of act end
)

// PlotDetector detects critical plot points in the narrative.
// Story 6.4 AC1: PlotDetector 實作 IsPlotPoint() 方法
type PlotDetector struct {
	// Milestone tracking to prevent duplicate triggers
	milestonesReached map[string]bool
}

// NewPlotDetector creates a new PlotDetector instance.
// Story 6.4 Task 1.3: 實作 NewPlotDetector() 建構函數
func NewPlotDetector() *PlotDetector {
	return &PlotDetector{
		milestonesReached: make(map[string]bool),
	}
}

// IsPlotPoint checks if the current game state represents a plot point.
// Story 6.4 AC1: 返回 (bool, string) - 是否為劇情點及類型
func (pd *PlotDetector) IsPlotPoint(state *engine.GameStateV2) (bool, string) {
	if state == nil {
		return false, ""
	}

	currentBeat := state.GetCurrentBeat()

	// Priority 1: Check for NPC death
	if isNPCDeath, _ := pd.evaluateNPCDeath(state, currentBeat); isNPCDeath {
		return true, PlotPointNPCDeath
	}

	// Priority 2: Check for act transition
	if pd.isNearActEnd(currentBeat) {
		return true, PlotPointActTransition
	}

	// Priority 3: Check for seed milestones
	if isMilestone, milestoneType := pd.evaluateSeedProgress(state); isMilestone {
		return true, milestoneType
	}

	// Priority 4: Check for first rule violation
	if pd.evaluateRuleViolation(state) {
		return true, PlotPointFirstViolation
	}

	return false, ""
}

func (pd *PlotDetector) getCurrentAct(currentBeat int) int {
	if currentBeat >= Act1Start && currentBeat <= Act1End {
		return 1
	} else if currentBeat >= Act2Start && currentBeat <= Act2End {
		return 2
	} else if currentBeat >= Act3Start && currentBeat <= Act3End {
		return 3
	}
	return 0
}

func (pd *PlotDetector) isNearActEnd(currentBeat int) bool {
	if currentBeat >= Act1End-ActEndBuffer && currentBeat <= Act1End {
		return true
	}
	if currentBeat >= Act2End-ActEndBuffer && currentBeat <= Act2End {
		return true
	}
	if currentBeat >= Act3End-ActEndBuffer && currentBeat <= Act3End {
		return true
	}
	return false
}

func (pd *PlotDetector) evaluateSeedProgress(state *engine.GameStateV2) (bool, string) {
	progress := pd.calculateSeedProgress(state)

	if progress >= 60.0 && !pd.milestonesReached["seed_60"] {
		pd.milestonesReached["seed_60"] = true
		return true, PlotPointSeedMilestone60
	}

	if progress >= 30.0 && !pd.milestonesReached["seed_30"] {
		pd.milestonesReached["seed_30"] = true
		return true, PlotPointSeedMilestone30
	}

	return false, ""
}

func (pd *PlotDetector) calculateSeedProgress(state *engine.GameStateV2) float64 {
	seeds := state.GetGlobalSeeds()
	if len(seeds) == 0 {
		return 0.0
	}

	totalTiers := len(seeds) * 3
	revealedTiers := 0

	for _, seed := range seeds {
		if seed.CurrentTier > 3 {
			revealedTiers += 3
		} else {
			revealedTiers += (seed.CurrentTier - 1)
		}
	}

	if totalTiers == 0 {
		return 0.0
	}

	return (float64(revealedTiers) / float64(totalTiers)) * 100.0
}

func (pd *PlotDetector) evaluateNPCDeath(state *engine.GameStateV2, currentBeat int) (bool, string) {
	if state.NPCManager == nil || state.NPCManager.Profiles == nil {
		return false, ""
	}

	// TODO: Epic 1/2 Integration - NPC Death Detection
	// This function should check for NPCs with DeathBeat == currentBeat
	// Expected implementation:
	//   for npcID, profile := range state.NPCManager.Profiles {
	//       if profile.DeathBeat != nil && *profile.DeathBeat == currentBeat {
	//           if npcState := state.NPCStates[npcID]; npcState != nil && npcState.IsAlive {
	//               return true, npcID
	//           }
	//       }
	//   }
	//
	// Pending: NPC Profile structure needs DeathBeat *int field (Story 7-8)
	// See: internal/npc/manager/profile.go for NPC death system integration

	for _, npcState := range state.NPCStates {
		if npcState == nil {
			continue
		}
		// Placeholder: Will check npcState.DeathBeat when Epic 1/2 integration is complete
	}

	return false, ""
}

func (pd *PlotDetector) evaluateRuleViolation(state *engine.GameStateV2) bool {
	if state.RuleWarnings == nil || len(state.RuleWarnings) == 0 {
		return false
	}

	// TODO: Epic 2/3 Integration - Rule Violation Prediction
	// AC5: Detect if a rule is ABOUT TO BE violated for the first time (Warnings == 0)
	//
	// This requires implementing willViolateThisBeat() helper method that:
	// 1. Checks if rule conditions are about to be satisfied in current beat
	// 2. Simulates game state changes to predict violations
	// 3. Returns true only for rules that haven't been warned yet (Warnings == 0)
	//
	// Current implementation: Placeholder that always returns false
	// Pending: RulesEngine integration for rule condition evaluation
	// See Dev Notes: "willViolateThisBeat() 實作可能較複雜"

	for _, warnings := range state.RuleWarnings {
		if warnings == 0 {
			// Placeholder: Should check if this rule will be violated THIS beat
			// Requires rule condition evaluation logic from RulesEngine
		}
	}

	return false
}

func (pd *PlotDetector) Reset() {
	pd.milestonesReached = make(map[string]bool)
}

func (pd *PlotDetector) HasReachedMilestone(milestone string) bool {
	return pd.milestonesReached[milestone]
}

// Detect is an adapter method for compatibility with MomentumController.
func (pd *PlotDetector) Detect(ctx *NarrativeContext) bool {
	if ctx == nil {
		return false
	}
	return ctx.IsPlotPoint
}
