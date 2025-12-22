package savefile

import (
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

// ==========================================================================
// Story 8.1: SaveFileV2 - New save format for v2.0 architecture
// ==========================================================================

// GameSettings represents game configuration in save file.
// This avoids circular dependency with internal/game package.
type GameSettings struct {
	Theme      string    `json:"theme"`
	Model      string    `json:"model"`
	Difficulty string    `json:"difficulty"` // "easy", "hard", "hell"
	Length     string    `json:"length"`     // "short", "medium", "long"
	AdultMode  bool      `json:"adult_mode"`
	CreatedAt  time.Time `json:"created_at"`
}

// StoryBibleData contains serializable story bible data.
// This is a minimal structure to avoid import cycles with orchestrator package.
// It contains all the same fields as orchestrator.StoryBible for JSON serialization.
type StoryBibleData struct {
	// Metadata
	GameID     string `json:"game_id"`
	CreatedAt  string `json:"created_at"`
	Difficulty string `json:"difficulty"`
	TotalBeats int    `json:"total_beats"`

	// Core Story Elements
	WorldView string `json:"world_view"`
	MainTheme string `json:"main_theme"`
	Setting   string `json:"setting"`

	// Enhanced Story Bible (Story 7.1)
	WorldSetting    *WorldSetting    `json:"world_setting"`
	CoreMystery     *CoreMystery     `json:"core_mystery"`
	StoryArc        *StoryArc        `json:"story_arc"`
	HiddenRules     []*HiddenRule    `json:"hidden_rules"`
	GlobalSeeds     []*seed.GlobalSeed `json:"global_seeds"`
	NPCProfiles     []*NPCProfile    `json:"npc_profiles"`
	PossibleEndings []*Ending        `json:"possible_endings"`
	UsedTemplates   *UsedTemplates   `json:"used_templates"`
}

// WorldSetting defines the game world configuration.
type WorldSetting struct {
	Location      string   `json:"location"`
	History       string   `json:"history"`
	WeirdElements []string `json:"weird_elements"`
	Atmosphere    string   `json:"atmosphere"`
	TimeFrame     string   `json:"time_frame"`
	Background    string   `json:"background"`
}

// CoreMystery defines the hidden truth.
type CoreMystery struct {
	Question   string `json:"question"`
	CoreTruth  string `json:"core_truth"`
	Revelation string `json:"revelation"`
	HiddenFrom string `json:"hidden_from"`
}

// StoryArc defines the three-act structure.
type StoryArc struct {
	Act1End       int             `json:"act1_end"`
	Midpoint      int             `json:"midpoint"`
	Act2End       int             `json:"act2_end"`
	TurningPoints []*TurningPoint `json:"turning_points"`
}

// TurningPoint represents a key moment.
type TurningPoint struct {
	Name        string `json:"name"`
	Beat        int    `json:"beat"`
	Description string `json:"description"`
}

// HiddenRule represents a game rule.
type HiddenRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Hints       []string `json:"hints"`
	Penalty     string   `json:"penalty"`
}

// NPCProfile represents an NPC.
type NPCProfile struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Archetype    string   `json:"archetype"`
	Personality  []string `json:"personality"`
	Appearance   string   `json:"appearance"`
	Backstory    string   `json:"backstory"`
	Skills       []string `json:"skills"`
	Inventory    []string `json:"inventory"`
	Secret       string   `json:"secret"`
	Introduction string   `json:"introduction"`
	LinkedSeeds  []string `json:"linked_seeds"`
	DeathTiming  int      `json:"death_timing"`
	Status       string   `json:"status"`
	DeathBeat    int      `json:"death_beat"`
	DeathReason  string   `json:"death_reason"`
	Description  string   `json:"description,omitempty"`
}

// Ending represents a possible game ending.
type Ending struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Type                   string           `json:"type"`
	Condition              *EndingCondition `json:"condition"`
	Description            string           `json:"description"`
	RequiredSeedPercentage float64          `json:"required_seed_percentage"`
}

// EndingCondition defines what triggers an ending.
type EndingCondition struct {
	MinSeedPercentage float64 `json:"min_seed_percentage"`
	MaxRuleViolations int     `json:"max_rule_violations,omitempty"`
	MinHP             int     `json:"min_hp,omitempty"`
	MinSAN            int     `json:"min_san,omitempty"`
}

// UsedTemplates tracks template usage.
type UsedTemplates struct {
	Rules  []string `json:"rules"`
	Scenes []string `json:"scenes"`
}

// SaveFileV2 represents the v2.0 save file format.
// Story 8.1 AC: JSON should contain meta, settings, game_state, story_bible
type SaveFileV2 struct {
	Meta       SaveMetadata        `json:"meta"`
	Settings   GameSettings        `json:"settings"`
	GameState  *engine.GameStateV2 `json:"game_state"`
	StoryBible *StoryBibleData     `json:"story_bible"`
	Checksum   string              `json:"checksum"`
}

// SaveMetadata contains metadata about the save file.
// Story 8.1 AC: save_id, save_name, created_at, updated_at, playtime, game_version
type SaveMetadata struct {
	SaveID      int       `json:"save_id"`       // 1-3 (slot number)
	SaveName    string    `json:"save_name"`     // User-defined name (optional)
	CreatedAt   time.Time `json:"created_at"`    // First save timestamp
	UpdatedAt   time.Time `json:"updated_at"`    // Last save timestamp
	PlayTime    int       `json:"playtime"`      // Total playtime in seconds
	GameVersion string    `json:"game_version"`  // Game version (e.g., "v2.0.0")
}

// NewSaveFileV2 creates a new SaveFileV2 with default values.
func NewSaveFileV2(slotID int, settings GameSettings, state *engine.GameStateV2, bible *StoryBibleData) *SaveFileV2 {
	now := time.Now()
	return &SaveFileV2{
		Meta: SaveMetadata{
			SaveID:      slotID,
			SaveName:    "",
			CreatedAt:   now,
			UpdatedAt:   now,
			PlayTime:    0,
			GameVersion: "v2.0.0",
		},
		Settings:   settings,
		GameState:  state,
		StoryBible: bible,
		Checksum:   "",
	}
}
