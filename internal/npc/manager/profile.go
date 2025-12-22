package manager

// Trait represents a basic character trait structure.
// Note: Complete Trait logic will be implemented in Story 1.6.
type Trait struct {
	ID         string `json:"id"`          // Unique identifier for the trait
	Content    string `json:"content"`     // Description of the trait
	RevealTier int    `json:"reveal_tier"` // 1=easy, 2=medium, 3=hard
}

// DialogueStyle defines how an NPC speaks and communicates.
type DialogueStyle struct {
	Formality  int      `json:"formality"`  // 1-5 (1=casual, 5=formal)
	Verbosity  int      `json:"verbosity"`  // 1-5 (1=brief, 5=verbose)
	Quirks     []string `json:"quirks"`     // Speech patterns, catchphrases
	Vocabulary string   `json:"vocabulary"` // Vocabulary preference (e.g., "military terms", "medical jargon")
}

// NPCProfile is the complete, immutable configuration for an NPC.
// This structure contains all information about an NPC, including secrets.
type NPCProfile struct {
	// Basic identity information
	ID         string `json:"id"`         // Unique NPC identifier
	Name       string `json:"name"`       // NPC name
	Archetype  string `json:"archetype"`  // NPC archetype (e.g., Survivor, Researcher, Child)
	Appearance string `json:"appearance"` // Physical description
	Backstory  string `json:"backstory"`  // Background story

	// Abilities and possessions
	Skills    []string `json:"skills"`    // List of skills
	Inventory []string `json:"inventory"` // List of items

	// Secret information (hidden from player)
	Secret     string `json:"secret"`      // NPC's secret
	SecretTier int    `json:"secret_tier"` // Secret importance level (1-3)

	// Traits and relationships
	Traits       []Trait      `json:"traits"`         // Character traits
	LinkedSeeds  []string     `json:"linked_seeds"`   // Associated seed IDs
	DeathBeat    int          `json:"death_beat"`     // Death trigger beat (0 = won't die)
	InitialEmotion EmotionState `json:"initial_emotion"` // Initial emotional state

	// Communication style
	DialogueStyle DialogueStyle `json:"dialogue_style"` // How the NPC speaks
}

// VisibleNPCProfile contains only player-visible information.
// This structure is used for UI display and protects secret information.
type VisibleNPCProfile struct {
	// Basic identity information
	ID         string `json:"id"`
	Name       string `json:"name"`
	Archetype  string `json:"archetype"`
	Appearance string `json:"appearance"`
	Backstory  string `json:"backstory"`

	// Abilities and possessions
	Skills    []string `json:"skills"`
	Inventory []string `json:"inventory"`

	// Only revealed traits
	Traits []Trait `json:"traits"`

	// Current emotional state (from RuntimeState, not InitialEmotion)
	// Note: Emotion display is handled by RuntimeState in Story 1.3

	// Communication style
	DialogueStyle DialogueStyle `json:"dialogue_style"`
}

// ToVisible converts an NPCProfile to a VisibleNPCProfile, filtering out secret information.
// Only traits with IDs in revealedTraitIDs will be included.
func (p NPCProfile) ToVisible(revealedTraitIDs []string) VisibleNPCProfile {
	// Create a set for O(1) lookup
	revealedSet := make(map[string]bool)
	for _, id := range revealedTraitIDs {
		revealedSet[id] = true
	}

	// Filter traits to only include revealed ones
	var visibleTraits []Trait
	for _, trait := range p.Traits {
		if revealedSet[trait.ID] {
			visibleTraits = append(visibleTraits, trait)
		}
	}

	// If no traits revealed, initialize empty slice instead of nil
	if visibleTraits == nil {
		visibleTraits = []Trait{}
	}

	return VisibleNPCProfile{
		ID:            p.ID,
		Name:          p.Name,
		Archetype:     p.Archetype,
		Appearance:    p.Appearance,
		Backstory:     p.Backstory,
		Skills:        p.Skills,
		Inventory:     p.Inventory,
		Traits:        visibleTraits,
		DialogueStyle: p.DialogueStyle,
	}
}
