package agents

// ==========================================================================
// Story 6-8: NPC Agent Types
// ==========================================================================

// NPCArchetype represents the NPC archetype type
type NPCArchetype string

const (
	// NPCArchetypeSacrificial is N-01 Sacrificial (犧牲者)
	NPCArchetypeSacrificial NPCArchetype = "N-01"
	// NPCArchetypeKnowledgeable is N-02 Knowledgeable (知情者)
	NPCArchetypeKnowledgeable NPCArchetype = "N-02"
	// NPCArchetypeHostile is N-03 Hostile (敵對者)
	NPCArchetypeHostile NPCArchetype = "N-03"
	// NPCArchetypeNeutral is N-04 Neutral (中立者)
	NPCArchetypeNeutral NPCArchetype = "N-04"
	// NPCArchetypeGuide is N-05 Guide (引導者)
	NPCArchetypeGuide NPCArchetype = "N-05"
	// NPCArchetypeDeceiver is N-06 Deceiver (欺騙者)
	NPCArchetypeDeceiver NPCArchetype = "N-06"
)

// String returns the string representation of NPCArchetype
func (na NPCArchetype) String() string {
	return string(na)
}

// NPCStatus represents the NPC's current status
type NPCStatus string

const (
	// NPCStatusAlive indicates NPC is alive
	NPCStatusAlive NPCStatus = "alive"
	// NPCStatusDying indicates NPC is in death process
	NPCStatusDying NPCStatus = "dying"
	// NPCStatusDead indicates NPC is dead
	NPCStatusDead NPCStatus = "dead"
)

// String returns the string representation of NPCStatus
func (ns NPCStatus) String() string {
	return string(ns)
}

// GenerateRequest is the request for generating an NPC instance
//
// Parameters:
//   - Archetype: The NPC archetype (N-01 to N-06)
//   - StoryContext: Current story theme and scene
//   - GlobalSeeds: Available global seeds for linking
//   - PlotStructure: Plot structure for death timing calculation
type GenerateRequest struct {
	Archetype     NPCArchetype
	StoryContext  StoryContext
	GlobalSeeds   []GlobalSeedInfo
	PlotStructure PlotStructure
}

// StoryContext contains story theme and scene information
type StoryContext struct {
	Theme string // e.g., "hospital", "school", "village"
	Scene string // Current scene description
}

// GlobalSeedInfo contains simplified global seed information
type GlobalSeedInfo struct {
	ID          string
	Description string
	CoreTruth   string
	ClueChain   []ClueInfo
}

// ClueInfo contains clue information
type ClueInfo struct {
	Tier      int
	Content   string
	Revealed  bool
}

// PlotStructure contains plot structure information
type PlotStructure struct {
	TotalBeats int
	Act1Range  [2]int // [start, end]
	Act2Range  [2]int
	Act3Range  [2]int
}

// GenerateResponse is the response containing generated NPC instance
type GenerateResponse struct {
	NPC NPCInstance
}

// NPCInstance represents a generated NPC instance
//
// Contains:
//   - Name: NPC name (符合主題與場景)
//   - Archetype: NPC archetype (N-01 to N-06)
//   - Personality: 3-5 personality keywords
//   - Appearance: 50-100 character description
//   - Backstory: 100-200 character background
//   - LinkedSeeds: Linked global seed IDs (0-2)
//   - DeathTiming: Scheduled death beat (0 = no death)
//   - Status: Current status (alive/dying/dead)
//   - DeathBeat: Actual death beat
//   - DeathReason: Reason for death
type NPCInstance struct {
	ID          string
	Name        string
	Archetype   NPCArchetype
	Personality []string // 3-5 keywords
	Appearance  string   // 50-100 chars
	Backstory   string   // 100-200 chars
	LinkedSeeds []string // GlobalSeed IDs (0-2)
	DeathTiming int      // Scheduled death beat (0 = no death)
	Status      NPCStatus
	DeathBeat   int
	DeathReason string
}

// DialogueRequest is the request for generating NPC dialogue
//
// Parameters:
//   - NPC: The NPC instance
//   - PlayerQuestion: Player's question (optional)
//   - Context: Current scene context
//   - Tension: Current tension level (0-100)
//   - CurrentBeat: Current beat number
type DialogueRequest struct {
	NPC            NPCInstance
	PlayerQuestion string
	Context        string
	Tension        int
	CurrentBeat    int
}

// DialogueResponse is the response containing generated dialogue
//
// Contains:
//   - Dialogue: Generated dialogue text (100-300 chars)
//   - SeedRevealed: Global seed ID if clue was revealed
//   - IsDeathDialogue: True if this is death dialogue
type DialogueResponse struct {
	Dialogue        string
	SeedRevealed    *string // Global seed ID (if revealed)
	IsDeathDialogue bool
}

// LLMGenerateResponse is the LLM's raw response for NPC generation
type LLMGenerateResponse struct {
	Name        string   `json:"name"`
	Personality []string `json:"personality"`
	Appearance  string   `json:"appearance"`
	Backstory   string   `json:"backstory"`
}

// LLMDialogueResponse is the LLM's raw response for dialogue
type LLMDialogueResponse struct {
	Dialogue string `json:"dialogue"`
}

// ArchetypeInfo contains archetype definition information
type ArchetypeInfo struct {
	ID          string
	Name        string
	Description string
	Traits      []string
}

// GetArchetypeInfo returns archetype information
func GetArchetypeInfo(archetype NPCArchetype) ArchetypeInfo {
	switch archetype {
	case NPCArchetypeSacrificial:
		return ArchetypeInfo{
			ID:          "N-01",
			Name:        "犧牲者 (Sacrificial)",
			Description: "無助、恐懼的角色，註定死亡以製造恐怖氛圍",
			Traits:      []string{"無助", "恐懼", "善良", "脆弱"},
		}
	case NPCArchetypeKnowledgeable:
		return ArchetypeInfo{
			ID:          "N-02",
			Name:        "知情者 (Knowledgeable)",
			Description: "神秘、了解真相的角色，提供線索但模稜兩可",
			Traits:      []string{"神秘", "謹慎", "知情", "深邃"},
		}
	case NPCArchetypeHostile:
		return ArchetypeInfo{
			ID:          "N-03",
			Name:        "敵對者 (Hostile)",
			Description: "冷漠、威脅的角色，阻礙玩家前進",
			Traits:      []string{"冷漠", "威脅", "危險", "敵意"},
		}
	case NPCArchetypeNeutral:
		return ArchetypeInfo{
			ID:          "N-04",
			Name:        "中立者 (Neutral)",
			Description: "普通、日常的角色，提供對比與真實感",
			Traits:      []string{"普通", "疑惑", "不知情", "平凡"},
		}
	case NPCArchetypeGuide:
		return ArchetypeInfo{
			ID:          "N-05",
			Name:        "引導者 (Guide)",
			Description: "關心、提示的角色，給予方向性建議",
			Traits:      []string{"關心", "友善", "提示", "鼓勵"},
		}
	case NPCArchetypeDeceiver:
		return ArchetypeInfo{
			ID:          "N-06",
			Name:        "欺騙者 (Deceiver)",
			Description: "虛假友善的角色，提供誤導與陷阱",
			Traits:      []string{"虛偽", "誤導", "表面友善", "陷阱"},
		}
	default:
		return ArchetypeInfo{
			ID:          "Unknown",
			Name:        "Unknown",
			Description: "Unknown archetype",
			Traits:      []string{},
		}
	}
}

// GetDialogueStyle returns the dialogue style description for an archetype
func GetDialogueStyle(archetype NPCArchetype) string {
	switch archetype {
	case NPCArchetypeSacrificial:
		return "語氣無助、恐懼、求助，使用短句和破碎語言。表達驚恐與絕望，尋求幫助。"
	case NPCArchetypeKnowledgeable:
		return "語氣神秘、模稜兩可，暗示線索但不直說。使用隱喻與暗示，保持神秘感。"
	case NPCArchetypeHostile:
		return "語氣冷漠、威脅、危險，言語簡短有力。表達敵意與威脅，拒絕合作。"
	case NPCArchetypeNeutral:
		return "語氣日常、普通、疑惑，像正常人對話。表達不確定性與困惑。"
	case NPCArchetypeGuide:
		return "語氣關心、提示、鼓勵，給予方向性建議。表達關懷與支持，提供幫助。"
	case NPCArchetypeDeceiver:
		return "語氣虛假友善、誤導，表面關心實則陷阱。表面熱情但暗藏危機。"
	default:
		return "語氣自然、真實。"
	}
}

// GetDialogueLengthRange returns the dialogue length range based on tension
func GetDialogueLengthRange(tension int) [2]int {
	if tension < 30 {
		return [2]int{150, 300} // 低張力: 從容、詳細
	} else if tension < 60 {
		return [2]int{100, 200} // 中張力: 緊張但仍能溝通
	} else if tension < 80 {
		return [2]int{80, 150} // 高張力: 簡短、緊迫
	} else {
		return [2]int{50, 100} // 極高張力: 驚恐、斷續
	}
}
