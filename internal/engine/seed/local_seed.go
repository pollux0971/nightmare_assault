package seed

import (
	"fmt"
	"time"
)

// SeedStatus represents the lifecycle status of a Local Seed.
// LocalSeeds transition through these states during their lifetime:
//   - Active: Seed is planted and awaiting harvest
//   - Harvested: Seed has been successfully revealed in narration
//   - Pruned: Seed was removed due to scene change or expiration
type SeedStatus string

const (
	// SeedStatusActive indicates the seed is planted and awaiting harvest.
	SeedStatusActive SeedStatus = "active"

	// SeedStatusHarvested indicates the seed has been successfully revealed.
	SeedStatusHarvested SeedStatus = "harvested"

	// SeedStatusPruned indicates the seed was removed (scene change or expiration).
	SeedStatusPruned SeedStatus = "pruned"
)

// LocalSeed represents a scene-specific foreshadowing element with time-based urgency.
//
// Design Philosophy (from v2.0 Architecture):
//   - LocalSeeds are scene-scoped: they only exist within a specific scene
//   - They have a MaxLifespan (default 5 beats) before expiring
//   - Urgency increases in 5 tiers (20/40/60/90/100) as the seed approaches expiration
//   - When scene changes, all active LocalSeeds for the old scene are pruned
//   - Urgency >= 40 triggers forced harvest to prevent missed foreshadowing
//
// Lifecycle States:
//   - Active: Planted and waiting to be harvested
//   - Harvested: Successfully revealed in narration
//   - Pruned: Removed due to scene change or expiration
type LocalSeed struct {
	ID          string     `json:"id"`           // Unique identifier (format: "LS-{scene}-{序號}")
	SceneID     string     `json:"scene_id"`     // Associated scene ID (e.g., "hospital_corridor")
	Content     string     `json:"content"`      // Foreshadowing content visible to player (e.g., "牆上有奇怪的刮痕")
	Detail      string     `json:"detail"`       // Key detail that must be echoed during harvest (e.g., "三條平行線")
	PlantText   string     `json:"plant_text"`   // Text used when planting the seed in narration
	PlantedAt   int        `json:"planted_at"`   // Beat number when this seed was planted
	MaxLifespan int        `json:"max_lifespan"` // Maximum beats before expiration (default 5)
	Status      SeedStatus `json:"status"`       // Current lifecycle status (active/harvested/pruned)
	CreatedAt   time.Time  `json:"created_at"`   // When this seed was created
}

// NewLocalSeed creates a new Local Seed with the given parameters.
//
// Validation rules:
//   - ID cannot be empty
//   - SceneID cannot be empty
//   - Content cannot be empty
//   - PlantedAt must be >= 0
//   - MaxLifespan must be > 0 (defaults to 5 if not provided)
//
// Parameters:
//   - id: Unique identifier (format: "LS-{scene}-{序號}")
//   - sceneID: Associated scene ID
//   - content: Foreshadowing content
//   - detail: Key detail for harvest validation
//   - plantText: Text for planting narration
//   - plantedAt: Beat number when planted
//   - maxLifespan: Maximum survival beats (use 0 for default of 5)
//
// Returns:
//   - *LocalSeed: A new LocalSeed instance with Active status
//   - error: If validation fails
func NewLocalSeed(id, sceneID, content, detail, plantText string, plantedAt, maxLifespan int) (*LocalSeed, error) {
	// Validate required fields
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	if sceneID == "" {
		return nil, fmt.Errorf("sceneID cannot be empty")
	}
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if plantedAt < 0 {
		return nil, fmt.Errorf("plantedAt cannot be negative, got %d", plantedAt)
	}

	// Default MaxLifespan to 5 if not provided
	if maxLifespan <= 0 {
		maxLifespan = 5
	}

	return &LocalSeed{
		ID:          id,
		SceneID:     sceneID,
		Content:     content,
		Detail:      detail,
		PlantText:   plantText,
		PlantedAt:   plantedAt,
		MaxLifespan: maxLifespan,
		Status:      SeedStatusActive,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// CalculateUrgency calculates the harvest urgency based on remaining lifespan.
// Returns an integer score (0-100) representing urgency level.
//
// Urgency Calculation Logic (5-tier system from authoritative design):
//   - Non-active seeds: urgency = 0   // Harvested/Pruned seeds have no urgency
//   - remaining <= 0:   urgency = 100 // Expired, highest urgency
//   - remaining == 1:   urgency = 90  // Last beat, extremely high
//   - remaining == 2:   urgency = 60  // Second-to-last, high (forced harvest threshold)
//   - remaining == 3:   urgency = 40  // Still some time, medium
//   - remaining >= 4:   urgency = 20  // Sufficient time, low
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - int: Urgency value in range [0, 100]
func (ls *LocalSeed) CalculateUrgency(currentBeat int) int {
	// Non-active seeds have no urgency (AC 5)
	if ls.Status != SeedStatusActive {
		return 0
	}

	remaining := ls.GetRemainingLifespan(currentBeat)

	if remaining <= 0 {
		return 100 // Expired, highest urgency
	}
	if remaining == 1 {
		return 90 // Last beat, extremely high
	}
	if remaining == 2 {
		return 60 // Second-to-last, high (forced harvest threshold)
	}
	if remaining == 3 {
		return 40 // Still some time, medium
	}
	return 20 // remaining >= 4, sufficient time, low
}

// GetRemainingLifespan calculates how many beats remain before the seed expires.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - int: Number of beats remaining (can be negative if expired)
func (ls *LocalSeed) GetRemainingLifespan(currentBeat int) int {
	return ls.MaxLifespan - (currentBeat - ls.PlantedAt)
}

// IsExpired checks if the seed has exceeded its MaxLifespan.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - bool: true if the seed has expired (remaining lifespan <= 0)
func (ls *LocalSeed) IsExpired(currentBeat int) bool {
	return ls.GetRemainingLifespan(currentBeat) <= 0
}

// ShouldForceHarvest determines if the seed urgency is high enough to force harvest.
// Seeds with urgency >= 40 should be prioritized to prevent missing foreshadowing.
// This corresponds to 3 or fewer beats remaining.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - bool: true if urgency >= 40 (3 or fewer beats remaining)
func (ls *LocalSeed) ShouldForceHarvest(currentBeat int) bool {
	return ls.CalculateUrgency(currentBeat) >= 40
}

// DeepCopy creates a deep copy of the LocalSeed.
// This prevents external modification of internal state.
//
// Returns:
//   - *LocalSeed: A complete deep copy of the seed
func (ls *LocalSeed) DeepCopy() *LocalSeed {
	if ls == nil {
		return nil
	}

	return &LocalSeed{
		ID:          ls.ID,
		SceneID:     ls.SceneID,
		Content:     ls.Content,
		Detail:      ls.Detail,
		PlantText:   ls.PlantText,
		PlantedAt:   ls.PlantedAt,
		MaxLifespan: ls.MaxLifespan,
		Status:      ls.Status,
		CreatedAt:   ls.CreatedAt,
	}
}

// PruneResult represents the result of pruning a LocalSeed.
// Contains information about the pruned seed for logging and narrative transitions.
type PruneResult struct {
	SeedID         string `json:"seed_id"`          // ID of the pruned seed
	SceneID        string `json:"scene_id"`         // Scene the seed belonged to
	Content        string `json:"content"`          // Original seed content (for reference)
	PruneReason    string `json:"prune_reason"`     // Reason for pruning ("scene_change" or "expired")
	TransitionText string `json:"transition_text"`  // Optional narrative text for transition
}
