package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// Placeholder constants for Story Bible fields when agent data is not yet available
//
// Issue #11 Fix: Define constants instead of hardcoded strings
const (
	PlaceholderHistory       = "TBD by Skeleton Mode"
	PlaceholderMystery       = "TBD by Skeleton Mode"
	PlaceholderRevelation    = "TBD by Skeleton Mode"
	PlaceholderWeirdElements = "TBD by Skeleton Mode"
	PlaceholderAtmosphere    = "TBD by Skeleton Mode"
)

// DifficultyLevel represents the game difficulty level (duplicated to avoid import cycle)
//
// Issue #1 Fix: Use strong types instead of strings for type safety
type DifficultyLevel int

const (
	DifficultyEasy DifficultyLevel = iota
	DifficultyHard
	DifficultyHell
)

// String returns the string representation for LLM prompts
func (d DifficultyLevel) String() string {
	switch d {
	case DifficultyEasy:
		return "easy"
	case DifficultyHard:
		return "hard"
	case DifficultyHell:
		return "hell"
	default:
		return "hard"
	}
}

// GameLength represents the game length (duplicated to avoid import cycle)
//
// Issue #1 Fix: Use strong types instead of strings for type safety
type GameLength int

const (
	LengthShort GameLength = iota
	LengthMedium
	LengthLong
)

// String returns the string representation for LLM prompts
func (l GameLength) String() string {
	switch l {
	case LengthShort:
		return "short"
	case LengthMedium:
		return "medium"
	case LengthLong:
		return "long"
	default:
		return "medium"
	}
}

// Phase1Config contains configuration for Phase 1: Genesis execution.
//
// Story 7.1 AC #3: Configuration required for Phase 1 execution
//
// This struct intentionally duplicates types from internal/game package to avoid
// import cycle issues. The orchestrator package cannot import internal/game because:
//   - internal/game imports internal/orchestrator/agents
//   - internal/orchestrator/agents imports internal/engine
//   - internal/engine/prompts/templates/base imports internal/game
//
// Issue #1 Fix: Changed from string types to strong enums for type safety
// Issue #8 Fix: Added documentation explaining why this conversion exists
//
// Conversion from game.GameConfig:
//   Use NewPhase1Config() helper function to convert game.GameConfig to Phase1Config.
//   This provides type-safe conversion without creating import cycles.
//
// Example:
//   gameConfig := game.NewGameConfig()
//   totalBeats := gameConfig.CalculateTotalBeats()
//   phase1Config, err := orchestrator.NewPhase1Config(
//       gameConfig.Theme,
//       gameConfig.Difficulty.String(),
//       gameConfig.Length.String(),
//       gameConfig.AdultMode,
//       totalBeats,
//   )
type Phase1Config struct {
	Theme       string          // Player's theme input (max 300 tokens)
	Difficulty  DifficultyLevel // Easy, Hard, Hell (with compile-time validation)
	Length      GameLength      // Short, Medium, Long (with compile-time validation)
	Adult18Plus bool            // 18+ mode toggle
	TotalBeats  int             // Calculated total beats based on difficulty + length
}

// ExecutePhase1 executes Phase 1: Genesis with full configuration support.
//
// Story 7.1 AC #3: Phase 1 Genesis Flow
// This method coordinates all agents to generate the complete Story Bible:
//   1. Initialize GameStateV2 (HP=100, SAN=100, tension=0)
//   2. Load Template Library
//   3. Theme adaptation and template selection (Style Filter)
//   4. Narration Agent (Skeleton Mode) - Story structure planning
//   5. Seed Agent (Global Generator) - Main plot foreshadowing
//   6. NPC Agent (Generate Mode) - 2-4 teammates generation
//   7. Assemble and save Story Bible
//   8. Narration Agent (Opening Mode) - Opening narrative (800-1200 words)
//
// Story 7.1 AC #7: Initialize complete GameStateV2 including:
//   - Player state: HP=100, SAN=100, Inventory=[]
//   - Tension state: Tension=0, Level=LOW
//   - Scene state: CurrentScene="opening", VisitedScenes={}
//   - Progress tracking: CurrentBeat=1, TotalBeats=calculated value
//   - Global Seed progress: all seeds marked as is_revealed=false
//   - NPC state: all NPCs marked as Alive=true, SAN=100
//
// Performance: Must complete within 30 seconds (NFR-P01)
//
// Parameters:
//   - ctx: Context for timeout control (should have 30s deadline)
//   - config: Phase 1 configuration (theme, difficulty, length, adult mode, total beats)
//
// Returns:
//   - *StoryBible: Complete story bible for Phase 2/3
//   - *GenesisResult: Opening narrative and initial state
//   - error: Any error encountered during generation
func (o *Orchestrator) ExecutePhase1(ctx context.Context, config *Phase1Config) (*StoryBible, *GenesisResult, error) {
	startTime := time.Now()
	log.Printf("[Orchestrator] ExecutePhase1 started: theme=%s, difficulty=%s, length=%s, adult=%v, totalBeats=%d",
		config.Theme, config.Difficulty.String(), config.Length.String(), config.Adult18Plus, config.TotalBeats)

	// Validate config
	if config == nil {
		return nil, nil, fmt.Errorf("config cannot be nil")
	}
	if config.Theme == "" {
		// Use default theme if empty (Story 7.1 AC #2)
		config.Theme = "廢棄精神病院的午夜探險"
		log.Printf("[Orchestrator] Using default theme: %s", config.Theme)
	}
	if config.TotalBeats <= 0 {
		return nil, nil, fmt.Errorf("totalBeats must be > 0, got %d", config.TotalBeats)
	}

	// Step 1: Initialize GameStateV2 (AC #7)
	log.Println("[Orchestrator] Step 1/8: Initializing GameStateV2")
	o.mu.Lock()
	o.gameState = engine.NewGameStateV2()
	o.gameState.SetHP(100)
	o.gameState.SetSAN(100)
	o.gameState.Tension = engine.NewTensionState()
	o.gameState.Tension.Value = 0
	o.gameState.Tension.Level = engine.TensionLevelLow
	o.gameState.CurrentScene = "opening"
	// CurrentBeat will be incremented to 1 when game loop starts
	o.mu.Unlock()
	log.Printf("[Orchestrator] GameState initialized: HP=100, SAN=100, Tension=0, Scene=opening")

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled during initialization: %w", err)
	}

	// Step 2: Load Template Library (AC #3)
	log.Println("[Orchestrator] Step 2/8: Loading Template Library")
	templates := o.templateLib.SelectTemplates(config.Theme, config.Difficulty.String())
	if templates == nil {
		return nil, nil, fmt.Errorf("template selection failed: no suitable templates found")
	}
	log.Printf("[Orchestrator] Templates loaded: %d rules, %d scenes",
		len(templates.Rules), len(templates.Scenes))

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after template loading: %w", err)
	}

	// Step 3: Theme adaptation (AC #6 - Style Filter)
	log.Println("[Orchestrator] Step 3/8: Theme adaptation (Style Filter)")
	// TODO Story 7.1: Implement StyleFilter.SelectTemplates(theme, difficulty, templates)
	// For now, use templates as-is
	// Future: Calculate match score, adapt entities, or improvise new ones
	log.Println("[Orchestrator] Style Filter: using templates as-is (implementation pending)")

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after style filtering: %w", err)
	}

	// Step 4: Narration Agent (Skeleton Mode) - Generate story structure (AC #4)
	log.Println("[Orchestrator] Step 4/8: Generating story skeleton")
	skeleton, err := o.narrationAgent.GenerateSkeleton(ctx, SkeletonRequest{
		Theme:      config.Theme,
		Difficulty: config.Difficulty.String(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("skeleton generation failed: %w", err)
	}
	log.Printf("[Orchestrator] Skeleton generated: worldView=%s, theme=%s",
		skeleton.WorldView, skeleton.MainTheme)

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after skeleton generation: %w", err)
	}

	// Step 5: Seed Agent (Global Generator) - Generate main plot seeds (AC #4)
	log.Println("[Orchestrator] Step 5/8: Generating global seeds")
	seedResponse, err := o.seedAgent.InvokeGlobalGenerate(ctx, &agents.GlobalGenerateRequest{
		StoryBible: &agents.SeedStoryBible{
			Theme:       config.Theme,
			WorldView:   skeleton.WorldView,
			Difficulty:  config.Difficulty.String(),
			CoreTruth:   "", // TODO: Extract from skeleton when available
			GlobalSeeds: nil,
		},
		Difficulty:      config.Difficulty.String(),
		StoryLength:     config.Length.String(),
		PossibleEndings: nil, // TODO: Extract from skeleton when available
	})
	if err != nil {
		return nil, nil, fmt.Errorf("global seed generation failed: %w", err)
	}
	globalSeeds := seedResponse.GlobalSeeds
	log.Printf("[Orchestrator] Generated %d global seeds", len(globalSeeds))

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after seed generation: %w", err)
	}

	// Step 6: NPC Agent (Generate Mode) - Generate 2-4 teammates (AC #4)
	log.Println("[Orchestrator] Step 6/8: Generating NPCs")
	npcCount := GetNPCCountForDifficulty(config.Difficulty.String())
	npcProfiles, err := o.npcAgent.GenerateProfiles(ctx, NPCRequest{
		Skeleton: *skeleton,
		Count:    npcCount,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("NPC generation failed: %w", err)
	}
	log.Printf("[Orchestrator] Generated %d NPC profiles", len(npcProfiles))

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after NPC generation: %w", err)
	}

	// Step 7: Assemble Story Bible (AC #4)
	log.Println("[Orchestrator] Step 7/8: Assembling Story Bible")
	storyBible := &StoryBible{
		GameID:     o.gameState.GameID,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Difficulty: config.Difficulty.String(),
		TotalBeats: config.TotalBeats,

		// Legacy fields (for backward compatibility)
		WorldView: skeleton.WorldView,
		MainTheme: skeleton.MainTheme,
		Setting:   skeleton.Setting,

		// Core data
		GlobalSeeds:   globalSeeds,
		NPCProfiles:   npcProfiles,
		UsedTemplates: templates,

		// Story 7.1 enhanced fields
		// TODO: These will be properly populated when Skeleton mode returns structured data
		// Issue #11 Fix: Use constants instead of empty strings for placeholders
		WorldSetting: &WorldSetting{
			Location:      skeleton.Setting,
			History:       PlaceholderHistory,
			WeirdElements: []string{PlaceholderWeirdElements},
			Atmosphere:    PlaceholderAtmosphere,
			TimeFrame:     "",
			Background:    skeleton.WorldView,
		},
		CoreMystery: &CoreMystery{
			Question:   PlaceholderMystery,
			CoreTruth:  PlaceholderMystery,
			Revelation: PlaceholderRevelation,
			HiddenFrom: "",
		},
		StoryArc: &StoryArc{
			Act1End:       config.TotalBeats * 3 / 10,  // 30% of total beats
			Midpoint:      config.TotalBeats / 2,        // 50% of total beats
			Act2End:       config.TotalBeats * 8 / 10,   // 80% of total beats
			TurningPoints: []*TurningPoint{},
		},
		HiddenRules: []*HiddenRule{}, // TODO: Generate based on difficulty
		PossibleEndings: []*Ending{   // TODO: Extract from skeleton
			{
				ID:                     "ending-true",
				Name:                   "真結局",
				Type:                   "true",
				RequiredSeedPercentage: 0.8,
				Condition: &EndingCondition{
					MinSeedPercentage: 0.8,
				},
			},
			{
				ID:                     "ending-good",
				Name:                   "好結局",
				Type:                   "good",
				RequiredSeedPercentage: 0.4,
				Condition: &EndingCondition{
					MinSeedPercentage: 0.4,
				},
			},
			{
				ID:                     "ending-bad",
				Name:                   "壞結局",
				Type:                   "bad",
				RequiredSeedPercentage: 0.0,
				Condition: &EndingCondition{
					MinSeedPercentage: 0.0,
				},
			},
		},
	}

	// Save to orchestrator state
	o.mu.Lock()
	o.storyBible = storyBible
	o.mu.Unlock()

	log.Printf("[Orchestrator] Story Bible assembled: gameID=%s, totalBeats=%d, seeds=%d, npcs=%d",
		storyBible.GameID, storyBible.TotalBeats, len(globalSeeds), len(npcProfiles))

	// Step 7.5: Initialize Global Seed progress tracking (AC #7)
	log.Println("[Orchestrator] Step 7.5/8: Initializing seed progress tracking")
	o.mu.Lock()
	for _, seed := range globalSeeds {
		o.gameState.AddGlobalSeed(seed)
	}
	o.mu.Unlock()
	log.Printf("[Orchestrator] Added %d global seeds to game state (all marked as unrevealed)", len(globalSeeds))

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after story bible assembly: %w", err)
	}

	// Step 8: Generate Opening narrative (AC #5)
	log.Println("[Orchestrator] Step 8/8: Generating opening narrative")
	opening, err := o.narrationAgent.GenerateOpening(ctx, OpeningRequest{
		Skeleton: *o.storyBible,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("opening generation failed: %w", err)
	}
	log.Printf("[Orchestrator] Opening narrative generated (length: %d characters)", len(opening.Story))

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled after opening generation: %w", err)
	}

	// Transition to Game Loop phase
	o.mu.Lock()
	o.currentPhase = PhaseGameLoop
	o.mu.Unlock()

	elapsed := time.Since(startTime)
	log.Printf("[Orchestrator] ExecutePhase1 completed successfully in %.2f seconds", elapsed.Seconds())

	// NFR-P01: Verify performance requirement (< 30 seconds)
	if elapsed > 30*time.Second {
		log.Printf("[WARN] Phase 1 execution exceeded 30s target: %.2fs", elapsed.Seconds())
	}

	genesisResult := &GenesisResult{
		Story: opening.Story,
	}

	return storyBible, genesisResult, nil
}

// NewPhase1Config creates a Phase1Config from individual parameters.
//
// Issue #1 Fix: Added helper function for converting game.GameConfig to Phase1Config
// This avoids import cycles while providing type-safe conversion.
//
// Parameters:
//   - theme: Player's theme input
//   - difficultyStr: Difficulty string ("easy", "hard", "hell")
//   - lengthStr: Length string ("short", "medium", "long")
//   - adult18Plus: 18+ mode toggle
//   - totalBeats: Calculated total beats
//
// Returns:
//   - *Phase1Config: Configuration ready for ExecutePhase1
//   - error: If invalid difficulty or length string
//
// Example usage from TUI:
//   import "github.com/nightmare-assault/nightmare-assault/internal/game"
//   gameConfig := game.NewGameConfig()
//   totalBeats := gameConfig.CalculateTotalBeats()
//   phase1Config, err := orchestrator.NewPhase1Config(
//       gameConfig.Theme,
//       gameConfig.Difficulty.String(),
//       gameConfig.Length.String(),
//       gameConfig.AdultMode,
//       totalBeats,
//   )
func NewPhase1Config(theme, difficultyStr, lengthStr string, adult18Plus bool, totalBeats int) (*Phase1Config, error) {
	// Parse difficulty
	var difficulty DifficultyLevel
	switch difficultyStr {
	case "easy", "簡單":
		difficulty = DifficultyEasy
	case "hard", "困難":
		difficulty = DifficultyHard
	case "hell", "地獄":
		difficulty = DifficultyHell
	default:
		return nil, fmt.Errorf("invalid difficulty: %s (expected: easy, hard, hell)", difficultyStr)
	}

	// Parse length
	var length GameLength
	switch lengthStr {
	case "short", "短篇":
		length = LengthShort
	case "medium", "中篇":
		length = LengthMedium
	case "long", "長篇":
		length = LengthLong
	default:
		return nil, fmt.Errorf("invalid length: %s (expected: short, medium, long)", lengthStr)
	}

	return &Phase1Config{
		Theme:       theme,
		Difficulty:  difficulty,
		Length:      length,
		Adult18Plus: adult18Plus,
		TotalBeats:  totalBeats,
	}, nil
}
