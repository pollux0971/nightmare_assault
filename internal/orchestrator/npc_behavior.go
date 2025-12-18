package orchestrator

import (
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.7: NPC Behavior Determination System
// ==========================================================================

// NPCAction represents an action that an NPC can take.
type NPCAction string

const (
	// NPCActionImpulsive represents impulsive action that might trigger rules (N-01 Sacrificial)
	NPCActionImpulsive NPCAction = "impulsive"

	// NPCActionCautious represents careful observation and rational analysis (N-02 Knowledgeable)
	NPCActionCautious NPCAction = "cautious"

	// NPCActionIntuitive represents relying on intuition and sensing danger early (N-03 Mystic)
	NPCActionIntuitive NPCAction = "intuitive"

	// NPCActionSelfPreserving represents self-preservation priority, possible betrayal (N-04 Betrayer)
	NPCActionSelfPreserving NPCAction = "self_preserving"

	// NPCActionFrozen represents paralyzed by fear, becoming a burden (N-05 Burden)
	NPCActionFrozen NPCAction = "frozen"

	// NPCActionExecutive represents strong execution without explanation (N-06 Silent)
	NPCActionExecutive NPCAction = "executive"

	// NPCActionPanic represents panic behavior when SAN is critically low
	NPCActionPanic NPCAction = "panic"

	// NPCActionCollapse represents mental collapse (SAN 1-19)
	NPCActionCollapse NPCAction = "collapse"
)

// Situation represents a dangerous or decision-requiring situation for NPC.
type Situation struct {
	Description    string // Description of the situation
	DangerLevel    int    // 0-100, higher means more dangerous
	RequiresChoice bool   // Whether NPC needs to make a choice
	Context        string // Additional context
}

// NPCBehaviorResult represents the result of NPC behavior determination.
type NPCBehaviorResult struct {
	Action      NPCAction // The action NPC takes
	Description string    // Description of the behavior
	Consequence string    // Consequence of the action (e.g., "觸發規則", "發現線索")
	LogEntry    string    // Entry to add to game log
}

// DetermineAction determines what action an NPC will take based on their archetype,
// SAN state, and the current situation.
//
// Story 7.7 AC #3: NPC Behavior Determination
//   - Based on Archetype.BehaviorPatterns
//   - Behavior should be consistent and predictable
//   - Record behavior to game log
//
// Story 7.7 AC #4: SAN State Influence on Behavior
//   - SAN 80-100: Normal archetype behavior
//   - SAN 50-79: More anxious, cautious
//   - SAN 20-49: Panic behavior
//   - SAN 1-19: Collapse behavior (depends on archetype)
func DetermineAction(npc *NPCProfile, npcSAN int, situation Situation) *NPCBehaviorResult {
	// Default SAN to 100 if not provided or invalid
	if npcSAN <= 0 {
		npcSAN = 100
	}

	// Check SAN state first - critical states override archetype
	// SAN 1-19: Collapse behavior (archetype-dependent)
	if npcSAN >= 1 && npcSAN <= 19 {
		return determineCollapseAction(npc, situation)
	}

	// SAN 20-49: Panic behavior (reduced rationality)
	if npcSAN >= 20 && npcSAN <= 49 {
		return determinePanicAction(npc, situation)
	}

	// SAN 50-79: Anxious behavior (archetype behavior but more cautious)
	if npcSAN >= 50 && npcSAN <= 79 {
		return determineAnxiousAction(npc, situation)
	}

	// SAN 80-100: Normal archetype behavior
	return determineNormalAction(npc, situation)
}

// determineNormalAction determines action based on archetype behavior patterns.
// SAN 80-100: Normal archetype-consistent behavior.
func determineNormalAction(npc *NPCProfile, situation Situation) *NPCBehaviorResult {
	switch npc.Archetype {
	case agents.NPCArchetypeSacrificial:
		// N-01 犧牲者：衝動行為，觸發規則
		return &NPCBehaviorResult{
			Action:      NPCActionImpulsive,
			Description: fmt.Sprintf("%s 恐懼地朝前衝去，沒有注意到警告標誌", npc.Name),
			Consequence: "可能觸發隱藏規則",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 做出衝動行為（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeKnowledgeable:
		// N-02 懷疑論者/知情者：謹慎觀察，理性分析
		return &NPCBehaviorResult{
			Action:      NPCActionCautious,
			Description: fmt.Sprintf("%s 仔細觀察周圍環境，分析情況", npc.Name),
			Consequence: "可能發現隱藏線索",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 進行謹慎分析（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeHostile:
		// N-03 靈感者：依靠直覺，提前感知危險
		return &NPCBehaviorResult{
			Action:      NPCActionIntuitive,
			Description: fmt.Sprintf("%s 突然停下腳步：「這裡...有什麼不對勁」", npc.Name),
			Consequence: "提前感知到危險",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 憑直覺察覺異常（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeNeutral:
		// N-04 背叛者：自保優先，可能背叛
		return &NPCBehaviorResult{
			Action:      NPCActionSelfPreserving,
			Description: fmt.Sprintf("%s 悄悄往後退了幾步，觀察情況", npc.Name),
			Consequence: "優先自保，可能在危險時背叛團隊",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 保持安全距離（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeGuide:
		// N-05 拖油瓶：僵直不動，拖累團隊
		return &NPCBehaviorResult{
			Action:      NPCActionFrozen,
			Description: fmt.Sprintf("%s 僵在原地，雙腿顫抖：「我...我動不了...」", npc.Name),
			Consequence: "無法行動，拖累團隊進度",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 被恐懼凍結（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeDeceiver:
		// N-06 沉默者：執行力強，但不解釋
		return &NPCBehaviorResult{
			Action:      NPCActionExecutive,
			Description: fmt.Sprintf("%s 二話不說，直接開始行動", npc.Name),
			Consequence: "有效執行但不解釋原因",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 沉默執行（%s）", npc.Name, situation.Description),
		}

	default:
		// Unknown archetype - default cautious behavior
		return &NPCBehaviorResult{
			Action:      NPCActionCautious,
			Description: fmt.Sprintf("%s 謹慎地觀察情況", npc.Name),
			Consequence: "保持警戒",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 保持警戒（%s）", npc.Name, situation.Description),
		}
	}
}

// determineAnxiousAction determines action when NPC is anxious (SAN 50-79).
// Archetype behavior becomes more cautious and nervous.
func determineAnxiousAction(npc *NPCProfile, situation Situation) *NPCBehaviorResult {
	switch npc.Archetype {
	case agents.NPCArchetypeSacrificial:
		return &NPCBehaviorResult{
			Action:      NPCActionImpulsive,
			Description: fmt.Sprintf("%s 驚慌失措地衝向出口，完全無視警告", npc.Name),
			Consequence: "高機率觸發規則",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 焦慮地做出衝動行為（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeKnowledgeable:
		return &NPCBehaviorResult{
			Action:      NPCActionCautious,
			Description: fmt.Sprintf("%s 焦慮地檢查周圍，手有些顫抖", npc.Name),
			Consequence: "分析能力下降",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 焦慮地進行觀察（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeHostile:
		return &NPCBehaviorResult{
			Action:      NPCActionIntuitive,
			Description: fmt.Sprintf("%s 不安地四處張望：「危險...到處都是危險...」", npc.Name),
			Consequence: "直覺變得過度敏感",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 焦慮地感知威脅（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeNeutral:
		return &NPCBehaviorResult{
			Action:      NPCActionSelfPreserving,
			Description: fmt.Sprintf("%s 警惕地看著所有人，隨時準備逃跑", npc.Name),
			Consequence: "團隊凝聚力下降，更可能背叛",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 焦慮地保持距離（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeGuide:
		return &NPCBehaviorResult{
			Action:      NPCActionFrozen,
			Description: fmt.Sprintf("%s 蹲在地上抱頭：「不要...不要過來...」", npc.Name),
			Consequence: "完全無法行動",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 焦慮地躲避（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeDeceiver:
		return &NPCBehaviorResult{
			Action:      NPCActionExecutive,
			Description: fmt.Sprintf("%s 快速行動，但明顯緊張不安", npc.Name),
			Consequence: "執行效率下降",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 緊張地執行行動（%s）", npc.Name, situation.Description),
		}

	default:
		return &NPCBehaviorResult{
			Action:      NPCActionCautious,
			Description: fmt.Sprintf("%s 焦慮地觀察情況", npc.Name),
			Consequence: "反應遲緩",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 焦慮（%s）", npc.Name, situation.Description),
		}
	}
}

// determinePanicAction determines action when NPC is panicking (SAN 20-49).
// Rationality is severely reduced, behavior becomes erratic.
func determinePanicAction(npc *NPCProfile, situation Situation) *NPCBehaviorResult {
	switch npc.Archetype {
	case agents.NPCArchetypeSacrificial:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 尖叫著亂跑，完全失去理智", npc.Name),
			Consequence: "必定觸發規則或造成危險",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 恐慌失控（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeKnowledgeable:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 語無倫次地喃喃自語，無法做出分析", npc.Name),
			Consequence: "失去分析能力",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 恐慌失去理智（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeHostile:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 驚恐地大喊：「它來了！它來了！」", npc.Name),
			Consequence: "直覺變成幻覺",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 陷入幻覺（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeNeutral:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 瘋狂地試圖逃跑，不顧一切", npc.Name),
			Consequence: "必定背叛團隊以求生",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 恐慌背叛（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeGuide:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 哭喊著：「救命！誰來救救我！」", npc.Name),
			Consequence: "完全崩潰，引來危險",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 恐慌哭喊（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeDeceiver:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 慌亂地四處逃竄，失去沉默的優勢", npc.Name),
			Consequence: "失去執行能力",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 恐慌逃竄（%s）", npc.Name, situation.Description),
		}

	default:
		return &NPCBehaviorResult{
			Action:      NPCActionPanic,
			Description: fmt.Sprintf("%s 陷入恐慌", npc.Name),
			Consequence: "無法理性行動",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 恐慌（%s）", npc.Name, situation.Description),
		}
	}
}

// determineCollapseAction determines action when NPC has collapsed (SAN 1-19).
// Final stage before complete mental breakdown, archetype-specific collapse behavior.
func determineCollapseAction(npc *NPCProfile, situation Situation) *NPCBehaviorResult {
	switch npc.Archetype {
	case agents.NPCArchetypeSacrificial:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 癱倒在地，眼神空洞，無法回應", npc.Name),
			Consequence: "完全失去行動能力，成為負擔",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 精神崩潰（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeKnowledgeable:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 緊抱著筆記本，喃喃重複：「都是假的...都是假的...」", npc.Name),
			Consequence: "知識體系崩潰，無法提供幫助",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 理性崩潰（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeHostile:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 陷入永恆的幻覺，看見不存在的恐怖", npc.Name),
			Consequence: "感知完全扭曲，可能攻擊友軍",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 感知崩潰（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeNeutral:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 瘋狂大笑，完全失去求生本能", npc.Name),
			Consequence: "自暴自棄，可能做出危險行為",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 意志崩潰（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeGuide:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 蜷縮成一團，無止境地顫抖哭泣", npc.Name),
			Consequence: "完全無法移動或交流",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 情感崩潰（%s）", npc.Name, situation.Description),
		}

	case agents.NPCArchetypeDeceiver:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 站立不動，如同雕像，對外界毫無反應", npc.Name),
			Consequence: "徹底沉默，失去所有功能",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 意識崩潰（%s）", npc.Name, situation.Description),
		}

	default:
		return &NPCBehaviorResult{
			Action:      NPCActionCollapse,
			Description: fmt.Sprintf("%s 精神崩潰", npc.Name),
			Consequence: "無法繼續行動",
			LogEntry:    fmt.Sprintf("[NPC行為] %s 崩潰（%s）", npc.Name, situation.Description),
		}
	}
}

// GetBehaviorPatternDescription returns a description of an archetype's behavior patterns.
// Useful for explaining NPC behavior to players.
func GetBehaviorPatternDescription(archetype agents.NPCArchetype) string {
	switch archetype {
	case agents.NPCArchetypeSacrificial:
		return "N-01 犧牲者：容易衝動行為，恐懼時會觸發危險規則"
	case agents.NPCArchetypeKnowledgeable:
		return "N-02 懷疑論者：謹慎觀察分析，理性處理問題"
	case agents.NPCArchetypeHostile:
		return "N-03 靈感者：依靠直覺行動，能提前感知危險"
	case agents.NPCArchetypeNeutral:
		return "N-04 背叛者：自保優先，危險時可能背叛團隊"
	case agents.NPCArchetypeGuide:
		return "N-05 拖油瓶：容易恐懼僵直，成為團隊負擔"
	case agents.NPCArchetypeDeceiver:
		return "N-06 沉默者：執行力強但不解釋，沉默行動"
	default:
		return "未知行為模式"
	}
}
