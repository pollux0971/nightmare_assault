package seed

import (
	"testing"
)

// TestNewHarvestInstruction tests creating harvest instructions from seeds.
func TestNewHarvestInstruction(t *testing.T) {
	seed, _ := NewGlobalSeed(
		"GS001",
		"Core seed content",
		"Hidden truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Subtle hint", Keywords: []string{"shadow", "whisper"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Obvious clue", Keywords: []string{"shadow", "truth"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Explicit reveal", Keywords: []string{"shadow", "truth", "entity"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	// Test successful instruction creation at beat 3 (within Tier 1 range)
	instruction, err := NewHarvestInstruction(seed, 3)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if instruction.SeedID != "GS001" {
		t.Errorf("Expected SeedID 'GS001', got '%s'", instruction.SeedID)
	}

	if instruction.ClueContent != "Subtle hint" {
		t.Errorf("Expected ClueContent 'Subtle hint', got '%s'", instruction.ClueContent)
	}

	if len(instruction.Keywords) != 2 {
		t.Errorf("Expected 2 keywords, got %d", len(instruction.Keywords))
	}

	if instruction.Keywords[0] != "shadow" || instruction.Keywords[1] != "whisper" {
		t.Errorf("Expected keywords ['shadow', 'whisper'], got %v", instruction.Keywords)
	}

	if instruction.MustInclude != "Hidden truth" {
		t.Errorf("Expected MustInclude 'Hidden truth' (LinkedTruth), got '%s'", instruction.MustInclude)
	}

	if instruction.Tier != 1 {
		t.Errorf("Expected Tier 1, got %d", instruction.Tier)
	}

	if instruction.Priority <= 0 {
		t.Errorf("Expected positive priority, got %d", instruction.Priority)
	}
}

// TestNewHarvestInstruction_NotReady tests error when seed not ready.
func TestNewHarvestInstruction_NotReady(t *testing.T) {
	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 10, BeatEnd: 15},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 16, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 21, BeatEnd: 25},
		},
	)

	// Try to create instruction at beat 5 (before Tier 1 starts at beat 10)
	instruction, err := NewHarvestInstruction(seed, 5)

	if err != ErrSeedNotReady {
		t.Errorf("Expected ErrSeedNotReady, got: %v", err)
	}

	if instruction != nil {
		t.Error("Expected nil instruction when seed not ready")
	}
}

// TestNewHarvestInstruction_FullyRevealed tests error when all tiers exhausted.
func TestNewHarvestInstruction_FullyRevealed(t *testing.T) {
	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	// Advance to tier 3
	_ = seed.AdvanceTier() // Now tier 2
	_ = seed.AdvanceTier() // Now tier 3

	// Manually set CurrentTier beyond max to simulate fully revealed state
	// (In production, this happens after successfully revealing tier 3)
	seed.CurrentTier = 4

	// Try to create instruction when fully revealed
	instruction, err := NewHarvestInstruction(seed, 15)

	if err != ErrNoClueAvailable {
		t.Errorf("Expected ErrNoClueAvailable, got: %v", err)
	}

	if instruction != nil {
		t.Error("Expected nil instruction when fully revealed")
	}
}

// TestCalculatePriority tests the priority calculation logic.
func TestCalculatePriority(t *testing.T) {
	tests := []struct {
		name         string
		tier         int
		beatStart    int
		beatEnd      int
		currentBeat  int
		relatedSeeds int
		description  string
	}{
		{
			name:         "Tier 1, early in window, no related seeds",
			tier:         1,
			beatStart:    1,
			beatEnd:      5,
			currentBeat:  2,
			relatedSeeds: 0,
			description:  "Should have highest tier bonus (+300)",
		},
		{
			name:         "Tier 2, mid window, some related seeds",
			tier:         2,
			beatStart:    6,
			beatEnd:      12,
			currentBeat:  9,
			relatedSeeds: 3,
			description:  "Should have medium tier bonus (+200) and related bonus (+15)",
		},
		{
			name:         "Tier 3, near end of window, many related seeds",
			tier:         3,
			beatStart:    13,
			beatEnd:      18,
			currentBeat:  17,
			relatedSeeds: 5,
			description:  "Should have low tier bonus (+100), high urgency, and related bonus (+25)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create seed with specific tier
			seed, _ := NewGlobalSeed(
				"GS001",
				"Test seed",
				"Truth",
				"mysterious",
				[]ClueTier{
					{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 5},
					{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 6, BeatEnd: 12},
					{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 13, BeatEnd: 18},
				},
			)

			// Advance to desired tier
			for i := 1; i < tt.tier; i++ {
				_ = seed.AdvanceTier()
			}

			// Add related seeds
			for i := 0; i < tt.relatedSeeds; i++ {
				seed.AddRelatedSeed("RELATED" + string(rune('A'+i)))
			}

			priority := calculatePriority(seed, tt.currentBeat)

			// Verify priority is positive
			if priority <= 0 {
				t.Errorf("%s: Expected positive priority, got %d", tt.description, priority)
			}

			// Log for manual verification
			t.Logf("%s: Priority = %d", tt.description, priority)
		})
	}
}

// TestCalculatePriority_UrgencyComparison tests that urgency affects priority.
func TestCalculatePriority_UrgencyComparison(t *testing.T) {
	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	// Priority at beat 2 (8 beats remaining until BeatEnd=10)
	priorityEarly := calculatePriority(seed, 2)

	// Priority at beat 9 (1 beat remaining until BeatEnd=10)
	priorityLate := calculatePriority(seed, 9)

	if priorityLate <= priorityEarly {
		t.Errorf("Expected priority near end of window (%d) > priority early in window (%d)", priorityLate, priorityEarly)
	}

	t.Logf("Early priority (8 beats remaining): %d", priorityEarly)
	t.Logf("Late priority (1 beat remaining): %d", priorityLate)
}

// TestCalculatePriority_TierComparison tests that lower tiers have higher priority.
func TestCalculatePriority_TierComparison(t *testing.T) {
	seed1, _ := NewGlobalSeed(
		"GS001",
		"Seed at tier 1",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	seed3, _ := NewGlobalSeed(
		"GS002",
		"Seed at tier 3",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	// Advance seed3 to tier 3
	_ = seed3.AdvanceTier()
	_ = seed3.AdvanceTier()

	// Compare at same beat within their respective windows
	priorityTier1 := calculatePriority(seed1, 5)  // Tier 1 at beat 5
	priorityTier3 := calculatePriority(seed3, 25) // Tier 3 at beat 25

	if priorityTier1 <= priorityTier3 {
		t.Errorf("Expected Tier 1 priority (%d) > Tier 3 priority (%d)", priorityTier1, priorityTier3)
	}

	t.Logf("Tier 1 priority: %d", priorityTier1)
	t.Logf("Tier 3 priority: %d", priorityTier3)
}
