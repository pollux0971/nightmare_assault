package seed

import (
	"fmt"
	"log"
	"sync"
)

// MaxHarvestPerTurn limits the number of seeds that can be revealed in a single turn.
// This prevents narrative overload and ensures clues are naturally woven into the story.
const MaxHarvestPerTurn = 2

// SeedManager manages the lifecycle of Global Seeds throughout the game.
// It tracks all active Global Seeds, checks revelation timing, and generates
// harvest instructions for the Narration Agent.
//
// Thread-Safety: All methods are thread-safe using read/write mutex.
type SeedManager struct {
	globalSeeds      []*GlobalSeed
	lastRevealedBeat map[string]int // seedID -> beat number of last revelation
	mu               sync.RWMutex
}

// NewSeedManager creates a new SeedManager instance.
//
// Returns:
//   - *SeedManager: A new seed manager with empty seed list
func NewSeedManager() *SeedManager {
	return &SeedManager{
		globalSeeds:      make([]*GlobalSeed, 0),
		lastRevealedBeat: make(map[string]int),
	}
}

// AddGlobalSeed adds a Global Seed to the manager.
// Nil seeds are silently ignored to prevent panics.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - seed: The GlobalSeed to add
func (sm *SeedManager) AddGlobalSeed(seed *GlobalSeed) {
	if seed == nil {
		return // Silently ignore nil seeds
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.globalSeeds = append(sm.globalSeeds, seed)
}

// GetAllActiveGlobalSeeds returns a deep copy of all Global Seeds.
// Returns deep copies to prevent external modification of internal state.
//
// Thread-Safety: This method is thread-safe.
//
// Returns:
//   - []*GlobalSeed: Deep copies of all global seeds (never nil, empty slice if no seeds)
func (sm *SeedManager) GetAllActiveGlobalSeeds() []*GlobalSeed {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return deep copies to protect internal state
	copies := make([]*GlobalSeed, len(sm.globalSeeds))
	for i, seed := range sm.globalSeeds {
		copies[i] = seed.DeepCopy()
	}

	return copies
}

// CheckHarvest checks all Global Seeds and generates HarvestInstructions for seeds
// that are ready to be revealed at the current beat.
//
// Algorithm:
//  1. Iterate through all Global Seeds
//  2. Check if each seed is ready to reveal at currentBeat (BeatStart <= currentBeat <= BeatEnd)
//  3. Create HarvestInstruction for each ready seed
//  4. Sort instructions by priority (descending - highest priority first)
//  5. Return sorted list
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - []*HarvestInstruction: Sorted list of harvest instructions (never nil, empty if no seeds ready)
func (sm *SeedManager) CheckHarvest(currentBeat int) []*HarvestInstruction {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	instructions := make([]*HarvestInstruction, 0)

	// Check each seed for readiness
	for _, seed := range sm.globalSeeds {
		// Skip fully revealed seeds
		if seed.IsFullyRevealed() {
			log.Printf("[DEBUG] Skipped seed %s: fully revealed (tier %d)", seed.ID, seed.CurrentTier)
			continue
		}

		// Try to create harvest instruction
		instruction, err := NewHarvestInstruction(seed, currentBeat)
		if err != nil {
			// Seed not ready or no clue available - log and skip
			log.Printf("[DEBUG] Skipped seed %s at beat %d: %v", seed.ID, currentBeat, err)
			continue
		}

		instructions = append(instructions, instruction)
	}

	// Sort by priority (descending - highest first)
	sortByPriority(instructions)

	// Limit to MaxHarvestPerTurn to prevent narrative overload
	if len(instructions) > MaxHarvestPerTurn {
		log.Printf("[DEBUG] Limited harvest from %d to %d instructions at beat %d",
			len(instructions), MaxHarvestPerTurn, currentBeat)
		instructions = instructions[:MaxHarvestPerTurn]
	}

	if len(instructions) > 0 {
		log.Printf("[DEBUG] CheckHarvest at beat %d: returning %d instructions", currentBeat, len(instructions))
	}

	return instructions
}

// sortByPriority sorts harvest instructions by priority in descending order.
// Uses bubble sort for simplicity (fine for small slices of 3-5 seeds).
//
// Parameters:
//   - instructions: The slice to sort in-place
func sortByPriority(instructions []*HarvestInstruction) {
	n := len(instructions)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if instructions[j].Priority < instructions[j+1].Priority {
				// Swap
				instructions[j], instructions[j+1] = instructions[j+1], instructions[j]
			}
		}
	}
}

// MarkSeedRevealed marks a Global Seed as having revealed its current tier.
// This advances the seed to the next tier and records the revelation beat.
//
// Prevents duplicate revelations in the same beat - returns error if attempting
// to reveal the same seed twice in one beat.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - seedID: The ID of the seed to mark as revealed
//   - currentBeat: The beat number when revelation occurred
//
// Returns:
//   - error: If seed not found, or if already revealed in this beat
func (sm *SeedManager) MarkSeedRevealed(seedID string, currentBeat int) error {
	// Input validation
	if seedID == "" {
		return fmt.Errorf("seedID cannot be empty")
	}
	if currentBeat < 0 {
		return fmt.Errorf("invalid currentBeat: %d (must be >= 0)", currentBeat)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find the seed
	var targetSeed *GlobalSeed
	for _, seed := range sm.globalSeeds {
		if seed.ID == seedID {
			targetSeed = seed
			break
		}
	}

	if targetSeed == nil {
		return ErrSeedNotFound
	}

	// Check if already revealed in this beat
	if lastBeat, exists := sm.lastRevealedBeat[seedID]; exists && lastBeat == currentBeat {
		return ErrAlreadyRevealedThisBeat
	}

	// Advance the tier
	err := targetSeed.AdvanceTier()
	if err != nil {
		// Seed is already fully revealed
		return err
	}

	// Record the revelation beat
	sm.lastRevealedBeat[seedID] = currentBeat

	return nil
}

// GetGlobalSeedsProgress calculates the overall revelation progress of all Global Seeds.
// Returns a value between 0.0 (no seeds revealed) and 1.0 (all seeds fully revealed).
//
// Progress Calculation:
//   - Each seed contributes: CurrentTier / TotalTiers
//   - Overall progress: Average of all seed progress values
//   - Empty seed list returns 0.0
//
// Thread-Safety: This method is thread-safe.
//
// Returns:
//   - float64: Progress value in range [0.0, 1.0]
func (sm *SeedManager) GetGlobalSeedsProgress() float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.globalSeeds) == 0 {
		return 0.0
	}

	var totalProgress float64

	for _, seed := range sm.globalSeeds {
		// Each seed's progress = CurrentTier / TotalTiers
		totalTiers := len(seed.ClueChain)
		if totalTiers == 0 {
			continue
		}

		// CurrentTier starts at 1, fully revealed means CurrentTier == len(ClueChain)
		seedProgress := float64(seed.CurrentTier) / float64(totalTiers)
		totalProgress += seedProgress
	}

	return totalProgress / float64(len(sm.globalSeeds))
}
