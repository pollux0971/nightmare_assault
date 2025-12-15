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

// ============================================================================
// NewLocalSeedHarvestInstruction Error Path Tests (Code Review Fix #3)
// ============================================================================

// TestNewLocalSeedHarvestInstruction_NilSeed tests error when seed is nil.
func TestNewLocalSeedHarvestInstruction_NilSeed(t *testing.T) {
	_, err := NewLocalSeedHarvestInstruction(nil, 10)

	if err == nil {
		t.Fatal("Expected error for nil seed, got nil")
	}

	expectedMsg := "seed cannot be nil"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewLocalSeedHarvestInstruction_NegativeBeat tests error when currentBeat is negative.
func TestNewLocalSeedHarvestInstruction_NegativeBeat(t *testing.T) {
	seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)

	_, err := NewLocalSeedHarvestInstruction(seed, -5)

	if err == nil {
		t.Fatal("Expected error for negative beat, got nil")
	}

	expectedMsg := "invalid currentBeat: -5 (must be >= 0)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewLocalSeedHarvestInstruction_PrunedSeed tests error when seed is pruned.
func TestNewLocalSeedHarvestInstruction_PrunedSeed(t *testing.T) {
	seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)
	seed.Status = SeedStatusPruned

	_, err := NewLocalSeedHarvestInstruction(seed, 12)

	if err != ErrSeedNotReady {
		t.Errorf("Expected ErrSeedNotReady for pruned seed, got %v", err)
	}
}

// TestNewLocalSeedHarvestInstruction_HarvestedSeed tests error when seed already harvested.
func TestNewLocalSeedHarvestInstruction_HarvestedSeed(t *testing.T) {
	seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)
	seed.Status = SeedStatusHarvested

	_, err := NewLocalSeedHarvestInstruction(seed, 12)

	if err != ErrSeedNotReady {
		t.Errorf("Expected ErrSeedNotReady for harvested seed, got %v", err)
	}
}

// TestNewLocalSeedHarvestInstruction_Success tests successful instruction creation.
func TestNewLocalSeedHarvestInstruction_Success(t *testing.T) {
	seed, _ := NewLocalSeed("LS-hospital-01", "hospital", "Blood stains on wall", "Handprint", "You notice blood stains", 10, 5)

	instruction, err := NewLocalSeedHarvestInstruction(seed, 13)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if instruction.SeedID != "LS-hospital-01" {
		t.Errorf("Expected SeedID='LS-hospital-01', got '%s'", instruction.SeedID)
	}

	if instruction.ClueContent != "You notice blood stains" {
		t.Errorf("Expected ClueContent='You notice blood stains', got '%s'", instruction.ClueContent)
	}

	if instruction.MustInclude != "Handprint" {
		t.Errorf("Expected MustInclude='Handprint', got '%s'", instruction.MustInclude)
	}

	if !instruction.IsLocalSeed {
		t.Error("Expected IsLocalSeed=true, got false")
	}

	if instruction.Tier != 0 {
		t.Errorf("Expected Tier=0 for LocalSeed, got %d", instruction.Tier)
	}

	// Verify urgency-based priority (13 - 10 = 3 beats elapsed, 2 remaining → urgency 60 → priority 300)
	expectedPriority := 300 // urgency 60 * 5 = 300
	if instruction.Priority != expectedPriority {
		t.Errorf("Expected Priority=%d (urgency 60), got %d", expectedPriority, instruction.Priority)
	}
}

// TestNewLocalSeedHarvestInstruction_PriorityVariation tests different urgency levels.
func TestNewLocalSeedHarvestInstruction_PriorityVariation(t *testing.T) {
	tests := []struct {
		name             string
		plantedAt        int
		maxLifespan      int
		currentBeat      int
		expectedPriority int
		description      string
	}{
		{"expired", 10, 5, 16, 500, "Urgency 100 → Priority 500"},
		{"critical", 10, 5, 14, 450, "Urgency 90 → Priority 450"},
		{"high", 10, 5, 13, 300, "Urgency 60 → Priority 300"},
		{"medium", 10, 5, 12, 200, "Urgency 40 → Priority 200"},
		{"low", 10, 5, 11, 100, "Urgency 20 → Priority 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", tt.plantedAt, tt.maxLifespan)

			instruction, err := NewLocalSeedHarvestInstruction(seed, tt.currentBeat)

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if instruction.Priority != tt.expectedPriority {
				t.Errorf("%s: Expected Priority=%d, got %d", tt.description, tt.expectedPriority, instruction.Priority)
			}
		})
	}
}

// ============================================================================
// extractKeywords Tests (Code Review Fix #5)
// ============================================================================

// TestExtractKeywords tests keyword extraction from content.
func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedCount  int
		shouldContain  []string
		shouldNotExist []string
	}{
		{
			name:          "empty content",
			content:       "",
			expectedCount: 0,
		},
		{
			name:          "Chinese content with punctuation",
			content:       "牆上有奇怪的刮痕，三條平行線。",
			expectedCount: 2,
			shouldContain: []string{"牆上有奇怪的刮痕", "三條平行線"},
		},
		{
			name:          "English content with punctuation",
			content:       "Dark shadows move in the corner, watching you carefully.",
			expectedCount: 5,
			shouldContain: []string{"Dark", "shadows", "move", "in", "the"},
		},
		{
			name:          "Mixed Chinese and English",
			content:       "Blood stains on wall，handprint visible。",
			expectedCount: 5,
			shouldContain: []string{"Blood", "stains", "on", "wall", "handprint"},
		},
		{
			name:          "Single character segments filtered out",
			content:       "a b c d e f g h",
			expectedCount: 0, // All single chars filtered
		},
		{
			name:          "Long segments filtered (>10 chars)",
			content:       "verylongkeywordthatexceedstenlimit short",
			expectedCount: 1,
			shouldContain: []string{"short"},
		},
		{
			name:          "Limit to 5 keywords max",
			content:       "one two three four five six seven eight nine ten",
			expectedCount: 5, // Should only return first 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractKeywords(tt.content)

			if len(keywords) != tt.expectedCount {
				t.Errorf("Expected %d keywords, got %d: %v", tt.expectedCount, len(keywords), keywords)
			}

			for _, expected := range tt.shouldContain {
				found := false
				for _, kw := range keywords {
					if kw == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected keyword '%s' not found in %v", expected, keywords)
				}
			}

			for _, notExpected := range tt.shouldNotExist {
				for _, kw := range keywords {
					if kw == notExpected {
						t.Errorf("Unexpected keyword '%s' found in %v", notExpected, keywords)
					}
				}
			}
		})
	}
}

// TestExtractKeywords_EdgeCases tests edge cases.
func TestExtractKeywords_EdgeCases(t *testing.T) {
	// Test with only punctuation
	keywords := extractKeywords("，。、！？；：")
	if len(keywords) != 0 {
		t.Errorf("Expected 0 keywords for punctuation-only content, got %d", len(keywords))
	}

	// Test with whitespace only
	keywords = extractKeywords("     ")
	if len(keywords) != 0 {
		t.Errorf("Expected 0 keywords for whitespace-only content, got %d", len(keywords))
	}

	// Test unicode length calculation (Chinese chars)
	keywords = extractKeywords("測試")
	if len(keywords) != 1 || keywords[0] != "測試" {
		t.Errorf("Expected ['測試'], got %v", keywords)
	}
}
