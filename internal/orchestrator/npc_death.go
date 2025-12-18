// Package orchestrator provides NPC death-related types and logic for Nightmare Assault.
// Story 7.8: NPC Death & Emotion Design
package orchestrator

import (
	"fmt"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.8 AC2: NPC Death Event Processing
// ==========================================================================

// NPCDeathEvent represents a complete NPC death event with all context
//
// Story 7.8 AC2: Complete death event processing
//   - NPC identification and status update
//   - SAN damage calculation (-15 to -25)
//   - Death reason and narrative generation
//   - Event recording for game history
//   - Triggering other NPC reactions
type NPCDeathEvent struct {
	// NPC Information
	NPCID       string              `json:"npc_id"`
	NPCName     string              `json:"npc_name"`
	Archetype   agents.NPCArchetype `json:"archetype"`

	// Death Context
	DeathBeat   int       `json:"death_beat"`
	DeathReason string    `json:"death_reason"`
	Location    string    `json:"location"`
	Timestamp   time.Time `json:"timestamp"`

	// Player Responsibility
	PlayerChoice      string  `json:"player_choice"`       // Player's choice that led to death
	PlayerResponsibility float64 `json:"player_responsibility"` // 0.0-1.0, how responsible player is
	Preventable       bool    `json:"preventable"`         // Was death preventable?

	// Emotional Impact
	SANLoss      int    `json:"san_loss"`       // SAN damage dealt (-15 to -25)
	DeathNarrative string `json:"death_narrative"` // Generated death description (200-400 chars)
	LastWords    string `json:"last_words"`     // NPC's final words (30-80 chars)

	// Three-Stage Architecture tracking
	Foreshadows   []DeathForeshadow `json:"foreshadows"`   // Stage 1: Foreshadowing clues
	WarningBeat   int               `json:"warning_beat"`  // Stage 2: Warning beat
	WarningChoice string            `json:"warning_choice"` // Player's warning stage choice

	// Other NPC Reactions
	NPCReactions map[string]string `json:"npc_reactions"` // NPCID -> Reaction text
}

// DeathForeshadow represents a foreshadowing clue (Stage 1)
//
// Story 7.8 AC1: Stage 1 - Foreshadowing (Beat N-3 to N-1)
type DeathForeshadow struct {
	Beat    int    `json:"beat"`    // When this clue appeared
	Type    string `json:"type"`    // "clue", "environment", "npc_dialogue"
	Content string `json:"content"` // The foreshadow text
}

// DeathWarningChoice represents player choices at warning stage (Stage 2)
//
// Story 7.8 AC1: Stage 2 - Warning (Beat N)
type DeathWarningChoice string

const (
	WarningChoicePrevent  DeathWarningChoice = "prevent"  // Actively prevent NPC action
	WarningChoiceObserve  DeathWarningChoice = "observe"  // Stay neutral, observe
	WarningChoiceEncourage DeathWarningChoice = "encourage" // Encourage NPC to explore
)

// String returns the display name of the warning choice
func (d DeathWarningChoice) String() string {
	switch d {
	case WarningChoicePrevent:
		return "阻止"
	case WarningChoiceObserve:
		return "觀察"
	case WarningChoiceEncourage:
		return "鼓勵"
	default:
		return "未知"
	}
}

// CalculateDeathOutcome determines death outcome based on player choice
//
// Story 7.8 AC1: Stage 3 consequences logic
//   - Prevent: NPC survives, may feel grateful or resentful
//   - Observe: 50% chance of death
//   - Encourage: High chance of death, player clearly responsible
//
// Parameters:
//   - choice: Player's warning choice
//   - npcArchetype: NPC's archetype (affects survival probability)
//   - randomSeed: Random value 0.0-1.0 for probabilistic outcomes
//
// Returns:
//   - survived: Whether NPC survived
//   - playerResponsibility: 0.0-1.0 how responsible player is
func CalculateDeathOutcome(choice DeathWarningChoice, npcArchetype agents.NPCArchetype, randomSeed float64) (survived bool, playerResponsibility float64) {
	switch choice {
	case WarningChoicePrevent:
		// Prevent: Always survives
		return true, 0.0

	case WarningChoiceObserve:
		// Observe: 50% chance of death
		// N-01 Sacrificial has higher death chance (70%)
		deathChance := 0.5
		if npcArchetype == agents.NPCArchetypeSacrificial {
			deathChance = 0.7
		}
		survived = randomSeed > deathChance
		playerResponsibility = 0.5 // Partial responsibility
		return survived, playerResponsibility

	case WarningChoiceEncourage:
		// Encourage: High chance of death (80%)
		// N-01 Sacrificial almost always dies (95%)
		deathChance := 0.8
		if npcArchetype == agents.NPCArchetypeSacrificial {
			deathChance = 0.95
		}
		survived = randomSeed > deathChance
		playerResponsibility = 0.9 // High responsibility
		return survived, playerResponsibility

	default:
		// Unknown choice: treat as observe
		return randomSeed > 0.5, 0.5
	}
}

// CalculateNPCDeathSAN calculates SAN loss from NPC death
//
// Story 7.8 AC2: SAN loss calculation (-15 to -25)
func CalculateNPCDeathSAN(playerResponsibility float64, intimacy int) int {
	// Clamp inputs
	if playerResponsibility < 0.0 {
		playerResponsibility = 0.0
	}
	if playerResponsibility > 1.0 {
		playerResponsibility = 1.0
	}
	if intimacy < 0 {
		intimacy = 0
	}
	if intimacy > 100 {
		intimacy = 100
	}

	// Base SAN loss: -20
	baseLoss := -20

	// Responsibility modifier: 0 to -5
	// High responsibility = more guilt = more SAN loss (more negative)
	responsibilityModifier := -int(playerResponsibility * 5.0)

	// Intimacy modifier: 0 to -5
	intimacyModifier := -(intimacy / 20)

	totalLoss := baseLoss + responsibilityModifier + intimacyModifier

	// Clamp to [-25, -15]
	if totalLoss < -25 {
		totalLoss = -25
	}
	if totalLoss > -15 {
		totalLoss = -15
	}

	return totalLoss
}

// ProcessNPCDeath processes an NPC death event
func ProcessNPCDeath(
	npcID, npcName string,
	archetype agents.NPCArchetype,
	deathReason string,
	deathBeat int,
	location string,
	playerChoice string,
	playerResponsibility float64,
	intimacy int,
) (*NPCDeathEvent, error) {
	// Validate inputs
	if npcID == "" {
		return nil, fmt.Errorf("NPC ID cannot be empty")
	}
	if npcName == "" {
		return nil, fmt.Errorf("NPC name cannot be empty")
	}
	if deathReason == "" {
		return nil, fmt.Errorf("death reason cannot be empty")
	}
	if deathBeat < 0 {
		return nil, fmt.Errorf("death beat must be >= 0")
	}

	// Calculate SAN loss
	sanLoss := CalculateNPCDeathSAN(playerResponsibility, intimacy)

	// Create death event
	event := &NPCDeathEvent{
		NPCID:                npcID,
		NPCName:              npcName,
		Archetype:            archetype,
		DeathBeat:            deathBeat,
		DeathReason:          deathReason,
		Location:             location,
		Timestamp:            time.Now(),
		PlayerChoice:         playerChoice,
		PlayerResponsibility: playerResponsibility,
		Preventable:          true,
		SANLoss:              sanLoss,
		Foreshadows:          make([]DeathForeshadow, 0),
		NPCReactions:         make(map[string]string),
	}

	return event, nil
}

// GenerateDeathNarrative generates death narrative (simplified implementation)
func GenerateDeathNarrative(
	npcName string,
	deathReason string,
	playerResponsibility float64,
	style DeathNarrativeStyle,
) (narrative string, lastWords string) {
	if style == DeathNarrativeDetailed {
		narrative = fmt.Sprintf(
			`%s的死亡來得突然而殘酷。%s。你清楚地看到了每一個細節——痛苦扭曲的表情、絕望的掙扎、生命力一點點流失的過程。這一切都因為你的選擇而發生。`,
			npcName, deathReason,
		)

		if playerResponsibility > 0.7 {
			narrative += "\n\n你知道這是你的錯。如果你當時做出不同的選擇，%s或許還活著。這份罪惡感將永遠伴隨著你。"
			narrative = fmt.Sprintf(narrative, npcName)
		} else if playerResponsibility > 0.3 {
			narrative += "\n\n你本可以做得更好。你沒有全力去拯救%s，現在只能眼睜睜看著悲劇發生。"
			narrative = fmt.Sprintf(narrative, npcName)
		} else {
			narrative += "\n\n你無力阻止這一切。儘管你盡力了，但命運還是奪走了%s的生命。這種無力感讓你感到絕望。"
			narrative = fmt.Sprintf(narrative, npcName)
		}
	} else {
		narrative = fmt.Sprintf(
			`%s倒下了。%s。你轉過頭，不忍再看。一切發生得太快，等你反應過來時，已經太遲了。空氣中瀰漫著死亡的氣息，你的心跳加速，呼吸變得急促。`,
			npcName, deathReason,
		)

		if playerResponsibility > 0.7 {
			narrative += fmt.Sprintf("\n\n你的選擇導致了這個結果。這個念頭在你腦海中揮之不去。如果當時你做出不同的決定，%s也許還能活著。這份愧疚將永遠伴隨著你。", npcName)
		} else if playerResponsibility > 0.3 {
			narrative += fmt.Sprintf("\n\n你不確定自己是否做對了。或許還有別的辦法，或許%s的死亡本可避免。這個疑問將縈繞在你心頭，久久不散。", npcName)
		} else {
			narrative += fmt.Sprintf("\n\n你無能為力。有些事情就是無法改變。即使你已經盡力，但%s的命運似乎早已註定。你感到深深的無力與悲傷。", npcName)
		}
	}

	if playerResponsibility > 0.7 {
		lastWords = fmt.Sprintf(`「為什麼……為什麼你要……」%s的聲音微弱地傳來，隨後陷入了永恆的沉默。`, npcName)
	} else if playerResponsibility > 0.3 {
		lastWords = fmt.Sprintf(`「小心……不要……」%s想說些什麼，但已經說不出完整的話了。`, npcName)
	} else {
		lastWords = fmt.Sprintf(`「謝謝你……至少你……試過了……」%s艱難地說出最後的話語。`, npcName)
	}

	return narrative, lastWords
}

// DeathNarrativeStyle represents the narrative style for death description
type DeathNarrativeStyle string

const (
	DeathNarrativeDetailed DeathNarrativeStyle = "detailed" // 18+ mode
	DeathNarrativeImplied  DeathNarrativeStyle = "implied"  // Normal mode
)

// ==========================================================================
// Story 7.8 AC4: Death Debrief (Preventability Analysis)
// ==========================================================================

// DeathDebrief represents a post-death analysis showing preventability
//
// Story 7.8 AC4: Death debrief structure
//   - What clues warned about danger
//   - At which beat player could have prevented death
//   - Which choice led to death
//   - What the correct action was
type DeathDebrief struct {
	NPCID           string   `json:"npc_id"`
	NPCName         string   `json:"npc_name"`
	Clues           []string `json:"clues"`            // Clues that warned of danger
	PreventionBeat  int      `json:"prevention_beat"`  // Beat where player could prevent
	DeadlyChoice    string   `json:"deadly_choice"`    // Choice that led to death
	CorrectAction   string   `json:"correct_action"`   // What should have been done
	Lesson          string   `json:"lesson"`           // Lesson learned (gameplay/story)
	AvoidFeeling    string   `json:"avoid_feeling"`    // Guidance to avoid "forced plot" feeling
}

// GenerateDeathDebrief generates a death debrief analysis
//
// Story 7.8 AC4: Death debrief generation
//   - Clearly identify warning clues
//   - Show prevention opportunity
//   - Indicate deadly choice
//   - Suggest correct action
//   - Ensure "I could have saved them" feeling
//   - Avoid "forced death plot" feeling
//
// Parameters:
//   - event: The death event
//   - foreshadows: Foreshadowing clues from Stage 1
//   - warningBeat: Beat where warning was given
//   - correctAction: What player should have done
//
// Returns:
//   - *DeathDebrief: Complete debrief analysis
func GenerateDeathDebrief(
	event *NPCDeathEvent,
	foreshadows []DeathForeshadow,
	warningBeat int,
	correctAction string,
) *DeathDebrief {
	// Extract clue texts from foreshadows
	clues := make([]string, len(foreshadows))
	for i, f := range foreshadows {
		clues[i] = fmt.Sprintf("Beat %d: %s", f.Beat, f.Content)
	}

	// Generate lesson based on archetype
	lesson := generateDeathLesson(event.Archetype, event.DeathReason)

	// Generate avoidance guidance
	avoidFeeling := "這不是劇情殺。回顧整個過程，有多個時機可以改變結局。線索已經給出，選擇權始終在你手中。"

	debrief := &DeathDebrief{
		NPCID:          event.NPCID,
		NPCName:        event.NPCName,
		Clues:          clues,
		PreventionBeat: warningBeat,
		DeadlyChoice:   event.PlayerChoice,
		CorrectAction:  correctAction,
		Lesson:         lesson,
		AvoidFeeling:   avoidFeeling,
	}

	return debrief
}

// generateDeathLesson generates a lesson from NPC death
func generateDeathLesson(archetype agents.NPCArchetype, deathReason string) string {
	switch archetype {
	case agents.NPCArchetypeSacrificial:
		return fmt.Sprintf("N-01 犧牲者的死亡展示了規則的殘酷性。%s 這個死因揭示了某條隱藏規則的後果。", deathReason)
	case agents.NPCArchetypeKnowledgeable:
		return "知情者的死亡可能帶走了關鍵線索。你需要更加珍惜與 NPC 的互動機會。"
	case agents.NPCArchetypeHostile:
		return "即使是敵對者，他們的死亡也會留下影響。這個世界比你想像的更加複雜。"
	case agents.NPCArchetypeNeutral:
		return "中立者的死亡提醒你：在這個世界，沒有人是真正安全的。即使置身事外，危險依然存在。"
	case agents.NPCArchetypeGuide:
		return "引導者的逝去意味著你失去了重要的幫助。接下來的路會更加艱難。"
	case agents.NPCArchetypeDeceiver:
		return "欺騙者的死亡揭示了真相與謊言的代價。他們的離去可能是祝福，也可能是詛咒。"
	default:
		return "每個 NPC 的死亡都有其意義。要從中學習，避免重蹈覆轍。"
	}
}

// ==========================================================================
// Story 7.8 AC5: Death Aftermath (NPC Absence Effects)
// ==========================================================================

// NPCAbsenceEffect represents the effects of NPC absence after death
//
// Story 7.8 AC5: NPC absence effects
//   - Narrative mentions NPC absence
//   - Skills/items become unavailable
//   - Other NPCs mention the deceased
//   - Ending statistics track survival
type NPCAbsenceEffect struct {
	NPCID         string   `json:"npc_id"`
	NPCName       string   `json:"npc_name"`
	MissingSkills []string `json:"missing_skills"` // Skills no longer available
	MissingItems  []string `json:"missing_items"`  // Items NPC was holding
	Consequences  []string `json:"consequences"`   // Narrative consequences
}

// GenerateAbsenceEffect generates absence effects for a dead NPC
//
// Story 7.8 AC5: Generate absence effects
//   - Identify skills/items NPC had
//   - Generate narrative mentions
//   - Create difficulty increases for lost skills
//
// Parameters:
//   - npcID: ID of dead NPC
//   - npcName: Name of dead NPC
//   - skills: Skills the NPC had
//   - items: Items the NPC was holding
//
// Returns:
//   - *NPCAbsenceEffect: Absence effect data
func GenerateAbsenceEffect(npcID, npcName string, skills, items []string) *NPCAbsenceEffect {
	consequences := []string{
		fmt.Sprintf("%s不在了，你感到孤獨與無助。", npcName),
	}

	if len(skills) > 0 {
		consequences = append(consequences,
			fmt.Sprintf("你失去了 %s 的專業技能。某些情境將變得更加困難。", npcName),
		)
	}

	if len(items) > 0 {
		consequences = append(consequences,
			fmt.Sprintf("%s 身上的道具已經無法取得。你需要找到其他方式獲得它們。", npcName),
		)
	}

	return &NPCAbsenceEffect{
		NPCID:         npcID,
		NPCName:       npcName,
		MissingSkills: skills,
		MissingItems:  items,
		Consequences:  consequences,
	}
}

// ==========================================================================
// Story 7.8 AC6: N-01 Sacrificial Special Handling
// ==========================================================================

// SacrificialDeathConfig represents special config for N-01 Sacrificial death
//
// Story 7.8 AC6: N-01 Sacrificial special death handling
//   - Death timing in Act 2 middle (default)
//   - Death demonstrates a specific rule
//   - Death process shows wrong approach as counter-example
//   - Death is gruesome but educational
//   - Player can choose to save or not save
//   - Successful save may alter plot
type SacrificialDeathConfig struct {
	PlannedBeat      int    `json:"planned_beat"`      // Planned death beat (Act 2 middle)
	DemonstratedRule string `json:"demonstrated_rule"` // Rule ID this death demonstrates
	CounterExample   string `json:"counter_example"`   // Wrong approach shown
	Educational      string `json:"educational"`       // Educational purpose
	Saveable         bool   `json:"saveable"`          // Can player save them?
	SaveConsequence  string `json:"save_consequence"`  // What happens if saved
}

// CreateSacrificialDeathConfig creates N-01 death configuration
//
// Story 7.8 AC6: N-01 special death configuration
//   - Scheduled in Act 2 middle
//   - Demonstrates a rule violation
//   - Shows counter-example of correct behavior
//   - Educational but brutal
//
// Parameters:
//   - act2MiddleBeat: Beat at Act 2 midpoint
//   - ruleID: Rule to demonstrate
//   - counterExample: Wrong approach description
//   - educational: Educational purpose
//
// Returns:
//   - *SacrificialDeathConfig: Special death configuration
func CreateSacrificialDeathConfig(
	act2MiddleBeat int,
	ruleID string,
	counterExample string,
	educational string,
) *SacrificialDeathConfig {
	return &SacrificialDeathConfig{
		PlannedBeat:      act2MiddleBeat,
		DemonstratedRule: ruleID,
		CounterExample:   counterExample,
		Educational:      educational,
		Saveable:         true,
		SaveConsequence:  "如果玩家成功拯救 N-01，劇情將發生重大改變。N-01 將成為重要的盟友，並在後續提供關鍵幫助。",
	}
}

// ValidateSacrificialDeath validates if a death is appropriate for N-01
//
// Story 7.8 AC6: Validate N-01 death appropriateness
//   - Must occur in Act 2
//   - Must demonstrate a rule
//   - Must be preventable but challenging
//
// Parameters:
//   - config: Sacrificial death configuration
//   - totalBeats: Total beats in game
//
// Returns:
//   - bool: Whether configuration is valid
//   - error: Validation error if invalid
func ValidateSacrificialDeath(config *SacrificialDeathConfig, totalBeats int) (bool, error) {
	// Act 2 is approximately 40-60% of total beats
	act2Start := int(float64(totalBeats) * 0.3)
	act2End := int(float64(totalBeats) * 0.7)

	if config.PlannedBeat < act2Start || config.PlannedBeat > act2End {
		return false, fmt.Errorf("N-01 death must occur in Act 2 (beats %d-%d), got beat %d",
			act2Start, act2End, config.PlannedBeat)
	}

	if config.DemonstratedRule == "" {
		return false, fmt.Errorf("N-01 death must demonstrate a rule")
	}

	if config.CounterExample == "" {
		return false, fmt.Errorf("N-01 death must show a counter-example")
	}

	if !config.Saveable {
		return false, fmt.Errorf("N-01 death must be saveable (preventable)")
	}

	return true, nil
}

