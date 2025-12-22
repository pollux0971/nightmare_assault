package momentum

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// ContextBuilder builds NarrativeContext from GameStateV2.
// Story 6.6 AC5: buildContext() 從 GameStateV2 建構上下文
type ContextBuilder struct {
	riskEvaluator *RiskEvaluator
	plotDetector  *PlotDetector
}

// NewContextBuilder creates a new ContextBuilder instance.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		riskEvaluator: NewRiskEvaluator(),
		plotDetector:  NewPlotDetector(),
	}
}

// BuildContext constructs a NarrativeContext from GameStateV2.
// Story 6.6 AC1-AC4: 包含所有必要欄位
// Story 6.6 AC5: 從 GameStateV2 建構上下文
//
// This method:
// 1. Extracts basic state (CurrentBeat, CurrentScene)
// 2. Evaluates risk level using RiskEvaluator
// 3. Detects plot points using PlotDetector
// 4. Checks NPC initiation status
// 5. Extracts pending events
// 6. Records recent choices and auto-resolved beats
func (cb *ContextBuilder) BuildContext(state *engine.GameStateV2, autoResolvedBeats int) *NarrativeContext {
	if state == nil {
		return &NarrativeContext{
			RiskLevel:         RiskNone,
			IsPlotPoint:       false,
			PendingEvents:     []*GameEvent{},
			RecentChoices:     []string{},
			AutoResolvedBeats: 0,
		}
	}

	// AC1: 基本狀態欄位
	ctx := &NarrativeContext{
		CurrentBeat:  state.GetCurrentBeat(),
		CurrentScene: state.CurrentScene,
	}

	// AC1: 風險評估欄位 (integrate with RiskEvaluator)
	// Story 6.3 integration: RiskEvaluator.EvaluateContext() returns RiskAssessment
	assessment, _ := cb.riskEvaluator.EvaluateContext(state, state.CurrentScene)
	ctx.RiskLevel = assessment.Level
	ctx.RiskFactors = cb.extractRiskFactors(state)

	// AC2: 劇情標記欄位 (integrate with PlotDetector)
	// Story 6.4 integration: PlotDetector.IsPlotPoint() returns (bool, string)
	ctx.IsPlotPoint, ctx.PlotPointType = cb.plotDetector.IsPlotPoint(state)

	// AC3: NPC 互動欄位
	ctx.NPCInitiatesConversation, ctx.InitiatingNPC = cb.checkNPCInitiation(state)

	// AC4: 事件與歷史欄位
	ctx.PendingEvents = cb.extractPendingEvents(state)
	ctx.RecentChoices = cb.extractRecentChoices(state)
	ctx.AutoResolvedBeats = autoResolvedBeats

	return ctx
}

// extractRiskFactors extracts risk factors from GameStateV2.
// Story 6.6 Dev Notes: 從多個來源推斷風險因素
func (cb *ContextBuilder) extractRiskFactors(state *engine.GameStateV2) []string {
	factors := []string{}

	// Low HP is a risk factor
	hp := state.GetHP()
	if hp <= 25 {
		factors = append(factors, "low_hp")
	} else if hp <= 50 {
		factors = append(factors, "medium_hp")
	}

	// Low SAN is a risk factor
	san := state.GetSAN()
	if san <= 25 {
		factors = append(factors, "low_san")
	} else if san <= 50 {
		factors = append(factors, "medium_san")
	}

	// High tension is a risk factor
	if state.Tension != nil {
		tensionValue := state.Tension.GetValue()
		if tensionValue >= 80 {
			factors = append(factors, "high_tension")
		} else if tensionValue >= 60 {
			factors = append(factors, "medium_tension")
		}
	}

	// Active rules with warnings
	if state.RuleWarnings != nil && len(state.RuleWarnings) > 0 {
		for _, warnings := range state.RuleWarnings {
			if warnings > 0 {
				factors = append(factors, "rule_warnings")
				break
			}
		}
	}

	return factors
}

// checkNPCInitiation checks if any NPC is initiating conversation.
// Story 6.6 Dev Notes: 檢查 NPC 主動對話觸發條件
//
// Currently returns false as NPCManager integration is pending.
// Future implementation will check NPC trust/emotion thresholds.
func (cb *ContextBuilder) checkNPCInitiation(state *engine.GameStateV2) (bool, string) {
	if state.NPCManager == nil {
		return false, ""
	}

	// TODO: Epic 1/2 integration - check NPC state for conversation triggers
	// For now, this is a placeholder implementation
	// Future: Check NPC trust levels, emotion states, scheduled events

	return false, ""
}

// extractPendingEvents extracts pending game events from GameStateV2.
// Story 6.6 Dev Notes: 從多個來源推斷待處理事件
func (cb *ContextBuilder) extractPendingEvents(state *engine.GameStateV2) []*GameEvent {
	events := []*GameEvent{}

	// Check for NPC scheduled deaths
	// This is a placeholder - actual implementation depends on Epic 1 NPC system
	if state.NPCStates != nil {
		for npcID, npcState := range state.NPCStates {
			if npcState == nil {
				continue
			}
			// Placeholder: actual death detection logic will be in Epic 1
			_ = npcID
		}
	}

	// Check for seed harvest events
	currentBeat := state.GetCurrentBeat()
	for _, seed := range state.GetGlobalSeeds() {
		if seed == nil {
			continue
		}
		// Check if seed is ready for next tier revelation
		// A seed is "pending" if current tier has a clue and we're in its beat range
		currentClue := seed.GetCurrentClue()
		if currentClue != nil {
			// Check if we're in the reveal window for this tier
			if currentBeat >= currentClue.BeatStart && currentBeat <= currentClue.BeatEnd {
				// This seed is pending revelation - treat as major event
				events = append(events, &GameEvent{
					ID:          seed.ID,
					Type:        "seed_reveal_pending",
					Description: "Global seed ready for revelation",
					IsMajor:     true,
					Beat:        currentBeat,
					Initiator:   "seed_system",
				})
			}
		}
	}

	// Check for rule violation events
	if state.RuleWarnings != nil {
		for ruleID, warnings := range state.RuleWarnings {
			if warnings >= 3 {
				events = append(events, &GameEvent{
					ID:          ruleID,
					Type:        "rule_violation",
					Description: "Critical rule violation threshold reached",
					IsMajor:     true,
					Beat:        currentBeat,
					Initiator:   "rule_system",
				})
			}
		}
	}

	return events
}

// extractRecentChoices extracts recent player choices from GameStateV2.
// Story 6.6 AC4: RecentChoices 欄位
//
// Currently returns empty slice as choice history is not yet in GameStateV2.
// Future implementation will use ChoiceHistory from Epic 5.
func (cb *ContextBuilder) extractRecentChoices(state *engine.GameStateV2) []string {
	// TODO: Epic 5 integration - extract from ChoiceHistory
	// For now, return empty slice
	return []string{}
}

// BuildContextForController is a convenience method for MomentumController.
// It integrates with the controller's internal autoResolvedBeats counter.
func (mc *MomentumController) BuildContext(state *engine.GameStateV2, autoResolvedBeats int) *NarrativeContext {
	builder := NewContextBuilder()
	// Reuse the controller's existing evaluators to maintain state (e.g., PlotDetector milestones)
	builder.riskEvaluator = mc.riskEvaluator
	builder.plotDetector = mc.plotDetector
	return builder.BuildContext(state, autoResolvedBeats)
}
