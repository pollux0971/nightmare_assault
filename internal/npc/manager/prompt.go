package manager

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
)

// emotionToDescription converts a numeric emotion value (0-100) to a descriptive Chinese text.
// This provides LLMs with meaningful context instead of raw numbers.
//
// The function maps emotion values to one of 5 descriptive levels for each emotion type:
// - trust: Describes the NPC's trust level towards the player
// - fear: Describes the NPC's fear level
// - stress: Describes the NPC's stress/anxiety level
//
// Returns an empty string for invalid emotion types.
func (m *NPCManager) emotionToDescription(value int, emotionType string) string {
	switch emotionType {
	case "trust":
		switch {
		case value >= 80:
			return "非常信任，願意分享秘密"
		case value >= 60:
			return "相當信任，願意合作"
		case value >= 40:
			return "中等信任，保持觀望"
		case value >= 20:
			return "不太信任，有所保留"
		default:
			return "完全不信任，可能敵對"
		}
	case "fear":
		switch {
		case value >= 80:
			return "極度恐懼，可能逃跑或崩潰"
		case value >= 60:
			return "非常害怕，行動遲疑"
		case value >= 40:
			return "有些害怕，小心翼翼"
		case value >= 20:
			return "略微緊張"
		default:
			return "冷靜沉著"
		}
	case "stress":
		switch {
		case value >= 80:
			return "瀕臨崩潰，可能失控"
		case value >= 60:
			return "壓力很大，判斷力下降"
		case value >= 40:
			return "有些焦慮"
		case value >= 20:
			return "輕微壓力"
		default:
			return "心態平穩"
		}
	}
	return ""
}

// mentalStateToDescription converts a MentalState enum to a descriptive Chinese text.
// The description includes both the state name (in English and Chinese) and behavioral implications.
//
// This helps LLMs understand how the NPC's mental state affects their behavior and decision-making.
func (m *NPCManager) mentalStateToDescription(state MentalState) string {
	switch state {
	case Normal:
		return "正常(Normal) - 思緒清晰、判斷準確、行為穩定"
	case Anxious:
		return "焦慮(Anxious) - 容易緊張、決策猶豫、需要安撫"
	case Corrupted:
		return "崩潰(Corrupted) - 精神失常、行為不可預測、可能暴力或自殘"
	default:
		return "未知狀態"
	}
}

// BuildNPCPrompt constructs a comprehensive LLM prompt for NPC dialogue and actions.
//
// The generated prompt includes the following sections (in order):
// 1. Basic Information - NPC name and physical appearance
// 2. Revealed Traits - Only traits that the player knows about (security-critical)
// 3. Behavioral Hints - Subtle clues about traits being hinted at
// 4. Current Emotional State - Trust, fear, stress with descriptive text
// 5. Mental/Psychological State - Current mental state with behavioral implications
// 6. Dialogue Style - Vocabulary preferences and speech quirks
//
// CRITICAL SECURITY FEATURE:
// This method NEVER includes hidden traits in the generated prompt. Only traits with
// status "revealed" or "hinting" are included. This prevents the LLM from accidentally
// leaking secret information about NPCs that the player hasn't discovered yet.
//
// Parameters:
//   - npcID: The unique identifier of the NPC
//
// Returns:
//   - A formatted markdown-style prompt string, or empty string if the NPC doesn't exist
//
// Example output:
//
//	## 角色：張醫生
//
//	外觀：40多歲男性，戴眼鏡，穿著白袍，神情疲憊
//
//	已知個性：理性、謹慎、有醫學專業
//
//	當前情緒：
//	- 信任程度：相當信任，願意合作
//	- 恐懼程度：非常害怕，行動遲疑
//	- 壓力程度：壓力很大，判斷力下降
//
//	心理狀態：焦慮(Anxious) - 容易緊張、決策猶豫、需要安撫
//
//	對話風格：
//	- 用詞：專業醫學術語混合口語
//	- 習慣：常說「從醫學角度來說」、「我見過太多這樣的案例」
func (m *NPCManager) BuildNPCPrompt(npcID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile := m.profiles[npcID]
	state := m.states[npcID]

	// Return empty string if NPC doesn't exist
	if profile == nil || state == nil {
		return ""
	}

	var sb strings.Builder

	// 1. Basic Information
	sb.WriteString(fmt.Sprintf("## 角色：%s\n\n", profile.Name))
	sb.WriteString(fmt.Sprintf("外觀：%s\n\n", profile.Appearance))

	// 2. Revealed Traits (SECURITY: Only show traits the player knows about)
	revealed := m.getRevealedTraits(npcID)
	if len(revealed) > 0 {
		sb.WriteString(fmt.Sprintf("已知個性：%s\n\n", strings.Join(revealed, "、")))
	}

	// 3. Behavioral Hints (Story 8.1: Phase-specific hints)
	hints := m.getPhaseSpecificHints(npcID, profile, state)
	if len(hints) > 0 {
		sb.WriteString(fmt.Sprintf("行為暗示：%s\n\n", strings.Join(hints, "；")))
	}

	// 4. Current Emotional State (with descriptive text for LLM understanding)
	sb.WriteString("當前情緒：\n")
	sb.WriteString(fmt.Sprintf("- 信任程度：%s\n", m.emotionToDescription(state.Emotion.Trust, "trust")))
	sb.WriteString(fmt.Sprintf("- 恐懼程度：%s\n", m.emotionToDescription(state.Emotion.Fear, "fear")))
	sb.WriteString(fmt.Sprintf("- 壓力程度：%s\n\n", m.emotionToDescription(state.Emotion.Stress, "stress")))

	// 5. Mental/Psychological State
	sb.WriteString(fmt.Sprintf("心理狀態：%s\n\n", m.mentalStateToDescription(state.MentalState)))

	// 6. Dialogue Style
	sb.WriteString("對話風格：\n")
	sb.WriteString(fmt.Sprintf("- 用詞：%s\n", profile.DialogueStyle.Vocabulary))
	if len(profile.DialogueStyle.Quirks) > 0 {
		sb.WriteString(fmt.Sprintf("- 習慣：%s\n", strings.Join(profile.DialogueStyle.Quirks, "、")))
	}

	return sb.String()
}

// GetHintingTraits returns a list of behavioral hints for traits in the "hinting" state.
//
// IMPORTANT: This is a compatibility wrapper for Story 1.8. The actual trait revelation
// logic is implemented in Story 1.6 (see reveal.go). This method works with the manager's
// stored profiles and states to extract hint strings for LLM prompts.
//
// For Story 1.8, this returns an empty slice as a placeholder since the full Trait structure
// with Hints field will be implemented in Story 1.6. When Story 1.6 is complete, this method
// will return actual hint strings extracted from TraitFull.Hints.
//
// Parameters:
//   - npcID: The unique identifier of the NPC
//
// Returns:
//   - A slice of hint strings for traits in "hinting" state
//   - Empty slice if no hints are available or NPC doesn't exist
func (m *NPCManager) GetHintingTraits(npcID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile := m.profiles[npcID]
	state := m.states[npcID]

	if profile == nil || state == nil {
		return []string{}
	}

	// Story 1.8 placeholder: Return empty slice
	// Story 1.6 will populate this with actual hints from TraitFull.Hints
	var hints []string

	for _, trait := range profile.Traits {
		if status, exists := state.TraitStates[trait.ID]; exists {
			if status == "hinting" {
				// Placeholder: When Story 1.6 is complete, this will extract
				// hints from the TraitFull structure
				// hints = append(hints, trait.Hints...)
			}
		}
	}

	return hints
}

// getRevealedTraits returns the content of traits that have been fully revealed to the player.
//
// SECURITY-CRITICAL: This method filters traits to ensure only "revealed" traits are included
// in LLM prompts. Hidden traits must never be leaked to prevent spoiling secrets.
//
// Parameters:
//   - npcID: The unique identifier of the NPC
//
// Returns:
//   - A slice of trait content strings for fully revealed traits
//   - Empty slice if no traits are revealed or NPC doesn't exist
func (m *NPCManager) getRevealedTraits(npcID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile := m.profiles[npcID]
	state := m.states[npcID]

	if profile == nil || state == nil {
		return []string{}
	}

	var revealed []string

	// Filter only traits with "revealed" status
	for _, trait := range profile.Traits {
		if status, exists := state.TraitStates[trait.ID]; exists {
			if status == "revealed" {
				revealed = append(revealed, trait.Content)
			}
		}
	}

	return revealed
}

// BuildFullNPCPrompt constructs a comprehensive LLM prompt that includes the NPC's
// knowledge base in addition to all information from BuildNPCPrompt.
//
// This method integrates with the UpdateManager to retrieve the NPC's known facts
// and formats them with confidence levels and sources. It also explicitly marks
// what the NPC doesn't know to prevent the LLM from inventing information.
//
// The generated prompt includes all sections from BuildNPCPrompt() plus:
// 7. Known Information - Facts the NPC has learned with confidence and source
// 8. Unknown Information - Explicit instruction that the NPC only knows listed facts
//
// If UpdateManager is nil, this method gracefully degrades to BuildNPCPrompt().
//
// Parameters:
//   - npcID: The unique identifier of the NPC
//
// Returns:
//   - A formatted markdown-style prompt string with knowledge base information
//   - Same as BuildNPCPrompt() if updateMgr is nil or NPC doesn't exist
//
// AC4: BuildFullNPCPrompt() 整合知識庫資訊
// AC5: 標記 NPC 不知道的事項
func (m *NPCManager) BuildFullNPCPrompt(npcID string) string {
	// 1. Get base prompt from BuildNPCPrompt
	basePrompt := m.BuildNPCPrompt(npcID)

	// If no base prompt (NPC doesn't exist), return empty
	if basePrompt == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(basePrompt)

	// 2. If no UpdateManager, return base prompt only (graceful degradation)
	if m.updateMgr == nil {
		return sb.String()
	}

	// 3. Get NPC's knowledge base
	kb := m.updateMgr.GetKnowledgeBase(npcID)

	// Add knowledge section
	sb.WriteString("\n## 已知資訊\n\n")

	if kb == nil || len(kb.KnownFacts) == 0 {
		sb.WriteString("（此角色目前沒有獲得任何重要資訊）\n")
	} else {
		// Iterate through known facts
		for factID, knownFact := range kb.KnownFacts {
			// Get the actual fact content from global facts
			globalFact := m.updateMgr.GetGlobalFact(factID)
			if globalFact == nil {
				continue // Skip if fact no longer exists
			}

			// Determine content (use distorted if available)
			content := globalFact.Content
			if knownFact.IsDistorted && knownFact.DistortedContent != "" {
				content = knownFact.DistortedContent
			}

			// Format confidence level
			confidenceStr := "確定"
			if knownFact.Confidence < 0.6 {
				confidenceStr = "不確定"
			} else if knownFact.Confidence < 0.8 {
				confidenceStr = "大概"
			}

			// Format source
			sourceStr := ""
			switch knownFact.LearnMethod {
			case knowledge.Witness:
				sourceStr = "（親眼目睹）"
			case knowledge.Told:
				if knownFact.LearnedFrom != "" && knownFact.LearnedFrom != npcID {
					sourceStr = fmt.Sprintf("（從%s得知）", knownFact.LearnedFrom)
				} else {
					sourceStr = "（聽說）"
				}
			case knowledge.Overheard:
				sourceStr = "（無意中聽到）"
			case knowledge.Inferred:
				sourceStr = "（推測）"
			}

			// Format the fact entry
			sb.WriteString(fmt.Sprintf("- [%s] %s %s\n",
				confidenceStr, content, sourceStr))
		}
	}

	// 4. Add "Unknown Information" section (CRITICAL for preventing hallucination)
	sb.WriteString("\n## 不知道的事項\n\n")
	sb.WriteString("重要：此角色**只知道**上述「已知資訊」區塊中列出的事實。\n")
	sb.WriteString("請勿讓此角色提及、暗示或表現出知道任何未列出的資訊。\n")
	sb.WriteString("如果玩家提及此角色不知道的事情，此角色應表現出困惑、好奇或要求說明。\n")

	return sb.String()
}

// ==========================================================================
// Story 8.1: Phase-Specific Hint Retrieval
// ==========================================================================

// getPhaseSpecificHints returns hints based on current trait revelation phases.
// Story 8.1 AC2: Progressive hint mechanism integrated into dialogue flow
//
// This method:
// 1. Identifies traits in HintPhase1 and HintPhase2
// 2. Retrieves phase-specific hints from TraitFull structure
// 3. Returns combined hints for natural integration into NPC dialogue
//
// Returns:
//   - []string: Hint strings organized by phase (Phase1 hints are more subtle)
func (m *NPCManager) getPhaseSpecificHints(npcID string, profile *NPCProfile, state *NPCRuntimeState) []string {
	var allHints []string

	// Iterate through all traits
	for _, trait := range profile.Traits {
		status, exists := state.TraitStates[trait.ID]
		if !exists {
			continue
		}

		// Convert basic Trait to TraitFull for hint access
		// In a real implementation, profile.Traits would be []TraitFull
		fullTrait := FromBasicTrait(trait)

		switch status {
		case HintPhase1.String():
			// Phase 1: Subtle, indirect hints
			if len(fullTrait.HintsPhase1) > 0 {
				allHints = append(allHints, fullTrait.HintsPhase1...)
			} else if len(fullTrait.Hints) > 0 {
				// Fallback to legacy Hints field
				allHints = append(allHints, fullTrait.Hints...)
			}

		case HintPhase2.String():
			// Phase 2: More explicit hints
			if len(fullTrait.HintsPhase2) > 0 {
				allHints = append(allHints, fullTrait.HintsPhase2...)
			} else if len(fullTrait.Hints) > 0 {
				// Fallback to legacy Hints field
				allHints = append(allHints, fullTrait.Hints...)
			}
		}
	}

	return allHints
}
