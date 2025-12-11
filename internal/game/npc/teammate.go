package npc

import (
	"time"
)

// NPCArchetype represents the archetype/role of an NPC teammate
type NPCArchetype string

const (
	// ArchetypeVictim - 受害者型：容易恐慌，需要保護
	ArchetypeVictim NPCArchetype = "victim"
	// ArchetypeUnreliable - 不可靠型：隱藏秘密，行為詭異
	ArchetypeUnreliable NPCArchetype = "unreliable"
	// ArchetypeLogic - 理性型：分析規則，提供邏輯推理
	ArchetypeLogic NPCArchetype = "logic"
	// ArchetypeIntuition - 直覺型：感知危險，提供預警
	ArchetypeIntuition NPCArchetype = "intuition"
	// ArchetypeInformer - 情報型：知道背景，提供線索
	ArchetypeInformer NPCArchetype = "informer"
	// ArchetypePossessed - 被附身型：已被影響，可能背叛
	ArchetypePossessed NPCArchetype = "possessed"
)

// PersonalityTraits represents personality characteristics
type PersonalityTraits struct {
	CoreTraits       []string `json:"core_traits"`
	BehaviorPatterns []string `json:"behavior_patterns"`
	SpeechStyle      string   `json:"speech_style"`
	FearResponse     string   `json:"fear_response"`
}

// CharacterSheet maintains character consistency
type CharacterSheet struct {
	Personality      PersonalityTraits `json:"personality"`
	EstablishedBehaviors []string      `json:"established_behaviors"`
	DialogueExamples []string          `json:"dialogue_examples"`
}

// TeammateStatus represents the current status of a teammate
type TeammateStatus struct {
	Alive        bool   `json:"alive"`
	Conscious    bool   `json:"conscious"`
	Condition    string `json:"condition"` // "healthy", "injured", "critical"
	Relationship int    `json:"relationship"` // 0-100, affects cooperation
}

// Item represents an item in inventory
type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Teammate represents an NPC teammate character
type Teammate struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Archetype      NPCArchetype      `json:"archetype"`
	Personality    PersonalityTraits `json:"personality"`
	Background     string            `json:"background"`
	Skills         []string          `json:"skills"`
	Status         TeammateStatus    `json:"status"`
	Location       string            `json:"location"` // Kept for backward compatibility
	HP             int               `json:"hp"`
	Inventory      []Item            `json:"inventory"`
	MemorySheet    CharacterSheet    `json:"memory_sheet"`

	// Story 4.4: Status Tracking Extensions
	LastSeen       time.Time        `json:"last_seen"`
	EmotionalState EmotionalState   `json:"emotional_state"`
	InjuryLevel    InjuryLevel      `json:"injury_level"`
	IsSeparated    bool             `json:"is_separated"`
	LastMessage    *TeammateMessage `json:"last_message,omitempty"`
}

// NewTeammate creates a new teammate with default values
func NewTeammate(id, name string, archetype NPCArchetype) *Teammate {
	return &Teammate{
		ID:        id,
		Name:      name,
		Archetype: archetype,
		HP:        100,
		Status: TeammateStatus{
			Alive:        true,
			Conscious:    true,
			Condition:    "healthy",
			Relationship: 50, // neutral starting relationship
		},
		Location:  "", // Initialize as empty, will be set when game starts
		Inventory: []Item{},
		Skills:    []string{},
		MemorySheet: CharacterSheet{
			Personality:          PersonalityTraits{},
			EstablishedBehaviors: []string{},
			DialogueExamples:     []string{},
		},
		// Story 4.4: Status Tracking defaults
		LastSeen:       time.Now(),
		EmotionalState: EmotionCalm,
		InjuryLevel:    InjuryNone,
		IsSeparated:    false,
		LastMessage:    nil,
	}
}
