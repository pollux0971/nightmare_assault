package agents

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

// ==========================================================================
// Story 6-5: Seed Agent Types
// ==========================================================================

// SeedStoryBible is a simplified view of the story skeleton for SeedAgent
//
// This is a minimal version containing only the fields needed by SeedAgent.
// Avoids circular dependencies by not importing orchestrator or narration packages.
type SeedStoryBible struct {
	Theme       string  // Main theme (e.g., "廢棄醫院")
	WorldView   string  // World setting description
	Difficulty  string  // Difficulty level
	CoreTruth   string  // Core truth to be revealed
	GlobalSeeds []*seed.GlobalSeed // Global Seeds (when managing Local Seeds)
}

// SeedEnding represents a possible game ending for SeedAgent
//
// Simplified version of narration.Ending, contains only fields needed for seed-to-ending linking.
type SeedEnding struct {
	ID                     string  // Ending ID
	Name                   string  // Ending name (e.g., "True Ending")
	Type                   string  // Ending type ("true"/"good"/"bad")
	RequiredSeedPercentage float64 // Required Global Seed reveal percentage (0.0-1.0)
}

// SeedOperation represents the type of operation to perform on Local Seeds
//
// Operations:
//   - Skip: Do nothing this beat (default)
//   - Plant: Plant a new Local Seed
//   - Harvest: Reveal an existing Local Seed in the narrative
//   - Prune: Remove expired or irrelevant Local Seeds
type SeedOperation int

const (
	// SeedOpSkip indicates no operation should be performed
	SeedOpSkip SeedOperation = iota

	// SeedOpPlant indicates a new Local Seed should be planted
	SeedOpPlant

	// SeedOpHarvest indicates an existing Local Seed should be revealed
	SeedOpHarvest

	// SeedOpPrune indicates expired/irrelevant Local Seeds should be removed
	SeedOpPrune
)

// String returns the string representation of SeedOperation
func (op SeedOperation) String() string {
	switch op {
	case SeedOpSkip:
		return "Skip"
	case SeedOpPlant:
		return "Plant"
	case SeedOpHarvest:
		return "Harvest"
	case SeedOpPrune:
		return "Prune"
	default:
		return "Unknown"
	}
}

// ==========================================================================
// Global Generator Mode Types (Genesis Phase)
// ==========================================================================

// GlobalGenerateRequest is the request for generating Global Seeds
//
// Used in Genesis Phase to generate 3-5 main storyline seeds with 3-tier clue chains.
//
// Parameters:
//   - StoryBible: The story skeleton (world view, plot structure, endings)
//   - Difficulty: Difficulty level (easy/normal/hard/hell)
//   - StoryLength: Story length (short/medium/long) - affects seed count
//   - PossibleEndings: List of possible endings to link seeds to
type GlobalGenerateRequest struct {
	StoryBible      *SeedStoryBible
	Difficulty      string
	StoryLength     string
	PossibleEndings []SeedEnding
}

// GlobalGenerateResponse is the response from generating Global Seeds
//
// Contains:
//   - GlobalSeeds: Generated Global Seeds (3-5 seeds with 3-tier clue chains)
type GlobalGenerateResponse struct {
	GlobalSeeds []*seed.GlobalSeed
}

// ==========================================================================
// Local Manager Mode Types (Game Loop)
// ==========================================================================

// LocalManageRequest is the request for managing Local Seeds
//
// Used in Game Loop to dynamically manage scene-level seeds.
//
// Parameters:
//   - CurrentBeat: Current game beat number
//   - StoryBible: Story Bible for context
//   - TensionState: Current tension state (from Epic 3)
//   - PlayerHints: Number of hints player has accumulated
//   - ActiveLocalSeeds: Currently active Local Seeds
//   - CurrentContext: Current scene/narrative context
type LocalManageRequest struct {
	CurrentBeat      int
	StoryBible       *SeedStoryBible
	TensionState     *engine.TensionState
	PlayerHints      int
	ActiveLocalSeeds []*seed.LocalSeed
	CurrentContext   string
}

// LocalManageResponse is the response from managing Local Seeds
//
// Contains:
//   - Operation: The seed operation to perform (Skip/Plant/Harvest/Prune)
//   - TargetSeed: The specific seed to operate on (nil for Skip/Plant)
//   - PrunedSeeds: List of seeds that were pruned (empty if Operation != Prune)
type LocalManageResponse struct {
	Operation    SeedOperation
	TargetSeed   *seed.LocalSeed
	PrunedSeeds  []*seed.LocalSeed
	PlantContext string // Context for planting new seed (if Operation == Plant)
}
