package seed

import (
	"fmt"
	"time"
)

// ClueTier represents a single tier in the progressive clue revelation system.
// Each Global Seed has 3 tiers that gradually reveal the foreshadowing from subtle to explicit.
type ClueTier struct {
	Tier      int      `json:"tier"`        // Tier level: 1 (subtle), 2 (obvious), or 3 (explicit)
	Content   string   `json:"content"`     // The clue content to be revealed
	Keywords  []string `json:"keywords"`    // Keywords that must be echoed in the narration
	BeatStart int      `json:"beat_start"`  // First beat where this clue can be revealed
	BeatEnd   int      `json:"beat_end"`    // Last beat where this clue can be revealed
}

// GlobalSeed represents a main storyline foreshadowing element.
// Global Seeds are planted in Phase 1 (Genesis) and progressively revealed throughout the game.
//
// Design Philosophy:
//   - Each Global Seed contains a 3-tier progressive revelation system
//   - Tier 1: Subtle, barely noticeable hints (early game)
//   - Tier 2: More obvious clues (mid game)
//   - Tier 3: Almost explicit revelations (late game)
//   - Seeds are linked to story truths and ending types
//   - Seeds can reference other seeds and rules to create narrative coherence
type GlobalSeed struct {
	ID           string     `json:"id"`             // Unique identifier (e.g., "GS001")
	Content      string     `json:"content"`        // Core foreshadowing content
	LinkedTruth  string     `json:"linked_truth"`   // The story truth this seed connects to
	LinkedEnding string     `json:"linked_ending"`  // The ending type this leads to (e.g., "tragic", "mysterious", "hopeful")
	CurrentTier  int        `json:"current_tier"`   // Current revelation tier (1-3), starts at 1
	ClueChain    []ClueTier `json:"clue_chain"`     // 3-tier progressive clue chain
	RelatedSeeds []string   `json:"related_seeds"`  // IDs of related Global Seeds
	RelatedRules []string   `json:"related_rules"`  // IDs of related hidden rules
	CreatedAt    time.Time  `json:"created_at"`     // When this seed was created
	LastRevealed time.Time  `json:"last_revealed"`  // When the last clue was revealed
}

// NewGlobalSeed creates a new Global Seed with the given parameters.
// The CurrentTier starts at 1 (first tier not yet revealed).
//
// Parameters:
//   - id: Unique identifier (e.g., "GS001")
//   - content: Core foreshadowing content
//   - linkedTruth: The story truth this seed connects to
//   - linkedEnding: The ending type (e.g., "tragic", "mysterious", "hopeful")
//   - clueChain: 3-tier progressive clue chain (must have exactly 3 tiers)
//
// Returns:
//   - *GlobalSeed: A new GlobalSeed instance
//   - error: If validation fails (e.g., clueChain doesn't have 3 tiers)
func NewGlobalSeed(id, content, linkedTruth, linkedEnding string, clueChain []ClueTier) (*GlobalSeed, error) {
	// Validate clue chain has exactly 3 tiers
	if len(clueChain) != 3 {
		return nil, fmt.Errorf("clue chain must have exactly 3 tiers, got %d", len(clueChain))
	}

	// Validate tier numbers are 1, 2, 3
	for i, clue := range clueChain {
		expectedTier := i + 1
		if clue.Tier != expectedTier {
			return nil, fmt.Errorf("tier %d should have Tier=%d, got Tier=%d", i, expectedTier, clue.Tier)
		}
		if clue.BeatStart > clue.BeatEnd {
			return nil, fmt.Errorf("tier %d: BeatStart (%d) cannot be greater than BeatEnd (%d)", clue.Tier, clue.BeatStart, clue.BeatEnd)
		}
	}

	return &GlobalSeed{
		ID:           id,
		Content:      content,
		LinkedTruth:  linkedTruth,
		LinkedEnding: linkedEnding,
		CurrentTier:  1, // Start at tier 1 (not yet revealed)
		ClueChain:    clueChain,
		RelatedSeeds: make([]string, 0),
		RelatedRules: make([]string, 0),
		CreatedAt:    time.Now().UTC(),
	}, nil
}

// GetCurrentClue returns the current clue tier based on CurrentTier.
// Returns nil if CurrentTier is invalid or if all tiers have been revealed.
//
// Returns:
//   - *ClueTier: The current clue tier, or nil if invalid/exhausted
func (gs *GlobalSeed) GetCurrentClue() *ClueTier {
	if gs.CurrentTier < 1 || gs.CurrentTier > len(gs.ClueChain) {
		return nil
	}
	// CurrentTier is 1-indexed, array is 0-indexed
	return &gs.ClueChain[gs.CurrentTier-1]
}

// AdvanceTier advances the revelation tier by 1.
// Updates LastRevealed timestamp when advancing.
// Does NOT increment if already at maximum tier - returns error instead.
//
// Returns:
//   - error: If already at the maximum tier (3)
func (gs *GlobalSeed) AdvanceTier() error {
	if gs.CurrentTier >= len(gs.ClueChain) {
		return fmt.Errorf("already at maximum tier (%d), cannot advance further", gs.CurrentTier)
	}
	gs.CurrentTier++
	gs.LastRevealed = time.Now().UTC()
	return nil
}

// IsReadyToReveal checks if the current tier's clue is ready to be revealed at the given beat.
// A clue is ready if: BeatStart <= currentBeat <= BeatEnd
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - bool: true if the current clue can be revealed at this beat
func (gs *GlobalSeed) IsReadyToReveal(currentBeat int) bool {
	clue := gs.GetCurrentClue()
	if clue == nil {
		return false
	}
	return currentBeat >= clue.BeatStart && currentBeat <= clue.BeatEnd
}

// IsFullyRevealed checks if all tiers have been revealed.
//
// Returns:
//   - bool: true if CurrentTier > 3 (all tiers exhausted)
func (gs *GlobalSeed) IsFullyRevealed() bool {
	return gs.CurrentTier > len(gs.ClueChain)
}

// AddRelatedSeed adds a related Global Seed ID to the RelatedSeeds list.
// Prevents duplicate entries.
//
// Parameters:
//   - seedID: The ID of the related seed
func (gs *GlobalSeed) AddRelatedSeed(seedID string) {
	// Check for duplicates
	for _, id := range gs.RelatedSeeds {
		if id == seedID {
			return // Already exists
		}
	}
	gs.RelatedSeeds = append(gs.RelatedSeeds, seedID)
}

// AddRelatedRule adds a related rule ID to the RelatedRules list.
// Prevents duplicate entries.
//
// Parameters:
//   - ruleID: The ID of the related rule
func (gs *GlobalSeed) AddRelatedRule(ruleID string) {
	// Check for duplicates
	for _, id := range gs.RelatedRules {
		if id == ruleID {
			return // Already exists
		}
	}
	gs.RelatedRules = append(gs.RelatedRules, ruleID)
}

// GetRemainingBeats calculates how many beats remain until the current clue's BeatEnd.
// Returns -1 if the clue has already expired or is invalid.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - int: Number of beats remaining, or -1 if expired/invalid
func (gs *GlobalSeed) GetRemainingBeats(currentBeat int) int {
	clue := gs.GetCurrentClue()
	if clue == nil {
		return -1
	}
	remaining := clue.BeatEnd - currentBeat
	if remaining < 0 {
		return -1
	}
	return remaining
}

// DeepCopy creates a deep copy of the GlobalSeed.
// This prevents external modification of internal state.
//
// Returns:
//   - *GlobalSeed: A complete deep copy of the seed
func (gs *GlobalSeed) DeepCopy() *GlobalSeed {
	if gs == nil {
		return nil
	}

	// Copy ClueChain
	clueChainCopy := make([]ClueTier, len(gs.ClueChain))
	for i, clue := range gs.ClueChain {
		keywordsCopy := make([]string, len(clue.Keywords))
		copy(keywordsCopy, clue.Keywords)
		clueChainCopy[i] = ClueTier{
			Tier:      clue.Tier,
			Content:   clue.Content,
			Keywords:  keywordsCopy,
			BeatStart: clue.BeatStart,
			BeatEnd:   clue.BeatEnd,
		}
	}

	// Copy RelatedSeeds
	relatedSeedsCopy := make([]string, len(gs.RelatedSeeds))
	copy(relatedSeedsCopy, gs.RelatedSeeds)

	// Copy RelatedRules
	relatedRulesCopy := make([]string, len(gs.RelatedRules))
	copy(relatedRulesCopy, gs.RelatedRules)

	return &GlobalSeed{
		ID:           gs.ID,
		Content:      gs.Content,
		LinkedTruth:  gs.LinkedTruth,
		LinkedEnding: gs.LinkedEnding,
		CurrentTier:  gs.CurrentTier,
		ClueChain:    clueChainCopy,
		RelatedSeeds: relatedSeedsCopy,
		RelatedRules: relatedRulesCopy,
		CreatedAt:    gs.CreatedAt,
		LastRevealed: gs.LastRevealed,
	}
}
