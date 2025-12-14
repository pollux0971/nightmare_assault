package seed

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNewGlobalSeed_ValidInput tests creating a GlobalSeed with valid input.
func TestNewGlobalSeed_ValidInput(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Subtle hint", Keywords: []string{"shadow"}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Obvious clue", Keywords: []string{"shadow", "truth"}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Explicit revelation", Keywords: []string{"shadow", "truth", "reveal"}, BeatStart: 13, BeatEnd: 18},
	}

	seed, err := NewGlobalSeed("GS001", "The protagonist is being watched", "Entity monitoring protagonist", "tragic", clueChain)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if seed.ID != "GS001" {
		t.Errorf("Expected ID 'GS001', got '%s'", seed.ID)
	}
	if seed.Content != "The protagonist is being watched" {
		t.Errorf("Expected correct content, got '%s'", seed.Content)
	}
	if seed.LinkedTruth != "Entity monitoring protagonist" {
		t.Errorf("Expected correct LinkedTruth, got '%s'", seed.LinkedTruth)
	}
	if seed.LinkedEnding != "tragic" {
		t.Errorf("Expected LinkedEnding 'tragic', got '%s'", seed.LinkedEnding)
	}
	if seed.CurrentTier != 1 {
		t.Errorf("Expected CurrentTier to start at 1, got %d", seed.CurrentTier)
	}
	if len(seed.ClueChain) != 3 {
		t.Errorf("Expected ClueChain length 3, got %d", len(seed.ClueChain))
	}
	if len(seed.RelatedSeeds) != 0 {
		t.Errorf("Expected empty RelatedSeeds, got %d items", len(seed.RelatedSeeds))
	}
	if len(seed.RelatedRules) != 0 {
		t.Errorf("Expected empty RelatedRules, got %d items", len(seed.RelatedRules))
	}
	if seed.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set, got zero time")
	}
}

// TestNewGlobalSeed_InvalidClueChainLength tests validation of clue chain length.
func TestNewGlobalSeed_InvalidClueChainLength(t *testing.T) {
	tests := []struct {
		name      string
		clueChain []ClueTier
	}{
		{"Empty clue chain", []ClueTier{}},
		{"Too few tiers (2)", []ClueTier{
			{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 10},
		}},
		{"Too many tiers (4)", []ClueTier{
			{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 11, BeatEnd: 15},
			{Tier: 4, Content: "Clue 4", Keywords: []string{}, BeatStart: 16, BeatEnd: 20},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGlobalSeed("GS001", "Content", "Truth", "ending", tt.clueChain)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

// TestNewGlobalSeed_InvalidTierNumbers tests validation of tier numbering.
func TestNewGlobalSeed_InvalidTierNumbers(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 6, BeatEnd: 10}, // Should be Tier 2
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 11, BeatEnd: 15},
	}

	_, err := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)
	if err == nil {
		t.Error("Expected error for invalid tier numbering, got nil")
	}
}

// TestNewGlobalSeed_InvalidBeatRange tests validation of beat ranges.
func TestNewGlobalSeed_InvalidBeatRange(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 10, BeatEnd: 5}, // BeatStart > BeatEnd
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 10},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 11, BeatEnd: 15},
	}

	_, err := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)
	if err == nil {
		t.Error("Expected error for invalid beat range, got nil")
	}
}

// TestGetCurrentClue tests retrieving the current clue tier.
func TestGetCurrentClue(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Tier 1 clue", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Tier 2 clue", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Tier 3 clue", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	// Test tier 1
	clue := seed.GetCurrentClue()
	if clue == nil {
		t.Fatal("Expected clue for tier 1, got nil")
	}
	if clue.Tier != 1 {
		t.Errorf("Expected Tier 1, got Tier %d", clue.Tier)
	}
	if clue.Content != "Tier 1 clue" {
		t.Errorf("Expected 'Tier 1 clue', got '%s'", clue.Content)
	}

	// Advance to tier 2
	seed.AdvanceTier()
	clue = seed.GetCurrentClue()
	if clue == nil {
		t.Fatal("Expected clue for tier 2, got nil")
	}
	if clue.Tier != 2 {
		t.Errorf("Expected Tier 2, got Tier %d", clue.Tier)
	}
	if clue.Content != "Tier 2 clue" {
		t.Errorf("Expected 'Tier 2 clue', got '%s'", clue.Content)
	}

	// Advance to tier 3
	seed.AdvanceTier()
	clue = seed.GetCurrentClue()
	if clue == nil {
		t.Fatal("Expected clue for tier 3, got nil")
	}
	if clue.Tier != 3 {
		t.Errorf("Expected Tier 3, got Tier %d", clue.Tier)
	}

	// Try to advance beyond tier 3 - should fail
	err := seed.AdvanceTier()
	if err == nil {
		t.Fatal("Expected error when advancing beyond tier 3")
	}
	// CurrentTier should remain at 3 (not incremented due to error)
	if seed.CurrentTier != 3 {
		t.Errorf("Expected CurrentTier to remain at 3 after failed advance, got %d", seed.CurrentTier)
	}
	// GetCurrentClue should still return tier 3
	clue = seed.GetCurrentClue()
	if clue == nil {
		t.Error("Expected tier 3 clue after failed advance")
	}
	if clue.Tier != 3 {
		t.Errorf("Expected Tier 3 after failed advance, got Tier %d", clue.Tier)
	}
}

// TestAdvanceTier tests tier advancement logic.
func TestAdvanceTier(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	// Tier 1 -> 2
	if err := seed.AdvanceTier(); err != nil {
		t.Errorf("Expected no error advancing from tier 1, got %v", err)
	}
	if seed.CurrentTier != 2 {
		t.Errorf("Expected CurrentTier 2, got %d", seed.CurrentTier)
	}
	if seed.LastRevealed.IsZero() {
		t.Error("Expected LastRevealed to be set")
	}

	// Tier 2 -> 3
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	firstReveal := seed.LastRevealed
	if err := seed.AdvanceTier(); err != nil {
		t.Errorf("Expected no error advancing from tier 2, got %v", err)
	}
	if seed.CurrentTier != 3 {
		t.Errorf("Expected CurrentTier 3, got %d", seed.CurrentTier)
	}
	if !seed.LastRevealed.After(firstReveal) {
		t.Error("Expected LastRevealed to be updated")
	}

	// Tier 3 -> Error (should NOT increment)
	if err := seed.AdvanceTier(); err == nil {
		t.Error("Expected error advancing beyond tier 3, got nil")
	}
	if seed.CurrentTier != 3 {
		t.Errorf("Expected CurrentTier to remain at 3 after failed advance, got %d", seed.CurrentTier)
	}
}

// TestIsReadyToReveal tests beat range checking for revelation timing.
func TestIsReadyToReveal(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	tests := []struct {
		beat     int
		expected bool
	}{
		{0, false},  // Before tier 1 range
		{1, true},   // Start of tier 1 range
		{3, true},   // Middle of tier 1 range
		{5, true},   // End of tier 1 range
		{6, false},  // After tier 1 range
		{10, false}, // Between tiers
	}

	for _, tt := range tests {
		result := seed.IsReadyToReveal(tt.beat)
		if result != tt.expected {
			t.Errorf("Beat %d: expected IsReadyToReveal=%v, got %v", tt.beat, tt.expected, result)
		}
	}

	// Advance to tier 2 and test
	seed.AdvanceTier()
	if !seed.IsReadyToReveal(6) {
		t.Error("Expected tier 2 to be ready at beat 6")
	}
	if !seed.IsReadyToReveal(12) {
		t.Error("Expected tier 2 to be ready at beat 12")
	}
	if seed.IsReadyToReveal(13) {
		t.Error("Expected tier 2 to not be ready at beat 13")
	}
}

// TestIsFullyRevealed tests checking if all tiers have been revealed.
func TestIsFullyRevealed(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	// Tier 1: Not fully revealed
	if seed.IsFullyRevealed() {
		t.Error("Expected not fully revealed at tier 1")
	}

	// Tier 2: Not fully revealed
	seed.AdvanceTier()
	if seed.IsFullyRevealed() {
		t.Error("Expected not fully revealed at tier 2")
	}

	// Tier 3: Not fully revealed
	seed.AdvanceTier()
	if seed.IsFullyRevealed() {
		t.Error("Expected not fully revealed at tier 3")
	}

	// Try to advance beyond tier 3 - should fail and NOT set fully revealed
	err := seed.AdvanceTier()
	if err == nil {
		t.Fatal("Expected error when advancing beyond tier 3")
	}
	// Since advance failed, should NOT be fully revealed (still at tier 3)
	if seed.IsFullyRevealed() {
		t.Error("Expected NOT fully revealed after failed advance (still at tier 3)")
	}

	// Manually set CurrentTier to 4 to simulate fully revealed state
	seed.CurrentTier = 4
	if !seed.IsFullyRevealed() {
		t.Error("Expected fully revealed when CurrentTier > 3")
	}
}

// TestAddRelatedSeed tests adding related seed IDs.
func TestAddRelatedSeed(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	// Add first related seed
	seed.AddRelatedSeed("GS002")
	if len(seed.RelatedSeeds) != 1 {
		t.Errorf("Expected 1 related seed, got %d", len(seed.RelatedSeeds))
	}
	if seed.RelatedSeeds[0] != "GS002" {
		t.Errorf("Expected 'GS002', got '%s'", seed.RelatedSeeds[0])
	}

	// Add second related seed
	seed.AddRelatedSeed("GS003")
	if len(seed.RelatedSeeds) != 2 {
		t.Errorf("Expected 2 related seeds, got %d", len(seed.RelatedSeeds))
	}

	// Add duplicate - should be ignored
	seed.AddRelatedSeed("GS002")
	if len(seed.RelatedSeeds) != 2 {
		t.Errorf("Expected 2 related seeds after duplicate, got %d", len(seed.RelatedSeeds))
	}
}

// TestAddRelatedRule tests adding related rule IDs.
func TestAddRelatedRule(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	// Add first related rule
	seed.AddRelatedRule("R001")
	if len(seed.RelatedRules) != 1 {
		t.Errorf("Expected 1 related rule, got %d", len(seed.RelatedRules))
	}

	// Add second related rule
	seed.AddRelatedRule("R002")
	if len(seed.RelatedRules) != 2 {
		t.Errorf("Expected 2 related rules, got %d", len(seed.RelatedRules))
	}

	// Add duplicate - should be ignored
	seed.AddRelatedRule("R001")
	if len(seed.RelatedRules) != 2 {
		t.Errorf("Expected 2 related rules after duplicate, got %d", len(seed.RelatedRules))
	}
}

// TestGetRemainingBeats tests calculating remaining beats until expiration.
func TestGetRemainingBeats(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
	}

	seed, _ := NewGlobalSeed("GS001", "Content", "Truth", "ending", clueChain)

	tests := []struct {
		beat     int
		expected int
	}{
		{1, 4},   // 5 - 1 = 4 beats remaining
		{3, 2},   // 5 - 3 = 2 beats remaining
		{5, 0},   // 5 - 5 = 0 beats remaining
		{6, -1},  // Already expired
		{10, -1}, // Already expired
	}

	for _, tt := range tests {
		result := seed.GetRemainingBeats(tt.beat)
		if result != tt.expected {
			t.Errorf("Beat %d: expected %d remaining beats, got %d", tt.beat, tt.expected, result)
		}
	}
}

// TestGlobalSeed_JSONSerialization tests marshaling and unmarshaling GlobalSeed.
func TestGlobalSeed_JSONSerialization(t *testing.T) {
	// Create a seed with all fields populated
	original, err := NewGlobalSeed(
		"GS001",
		"Test foreshadowing content",
		"Test truth revelation",
		"tragic",
		[]ClueTier{
			{Tier: 1, Content: "Subtle hint", Keywords: []string{"shadow", "whisper"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Obvious clue", Keywords: []string{"shadow", "truth"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Full revelation", Keywords: []string{"shadow", "truth", "entity"}, BeatStart: 13, BeatEnd: 18},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create test seed: %v", err)
	}

	// Add related seeds and rules
	original.AddRelatedSeed("GS002")
	original.AddRelatedSeed("GS003")
	original.AddRelatedRule("RULE001")

	// Advance tier to populate LastRevealed
	_ = original.AdvanceTier()

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal to new seed
	var restored GlobalSeed
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify all fields are restored correctly
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", restored.ID, original.ID)
	}
	if restored.Content != original.Content {
		t.Errorf("Content mismatch: got %s, want %s", restored.Content, original.Content)
	}
	if restored.LinkedTruth != original.LinkedTruth {
		t.Errorf("LinkedTruth mismatch: got %s, want %s", restored.LinkedTruth, original.LinkedTruth)
	}
	if restored.LinkedEnding != original.LinkedEnding {
		t.Errorf("LinkedEnding mismatch: got %s, want %s", restored.LinkedEnding, original.LinkedEnding)
	}
	if restored.CurrentTier != original.CurrentTier {
		t.Errorf("CurrentTier mismatch: got %d, want %d", restored.CurrentTier, original.CurrentTier)
	}
	if len(restored.ClueChain) != len(original.ClueChain) {
		t.Errorf("ClueChain length mismatch: got %d, want %d", len(restored.ClueChain), len(original.ClueChain))
	}
	if len(restored.RelatedSeeds) != len(original.RelatedSeeds) {
		t.Errorf("RelatedSeeds length mismatch: got %d, want %d", len(restored.RelatedSeeds), len(original.RelatedSeeds))
	}
	if len(restored.RelatedRules) != len(original.RelatedRules) {
		t.Errorf("RelatedRules length mismatch: got %d, want %d", len(restored.RelatedRules), len(original.RelatedRules))
	}

	// Verify timestamp fields
	if restored.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after unmarshal")
	}
	if restored.LastRevealed.IsZero() {
		t.Error("LastRevealed should not be zero after unmarshal")
	}

	// Verify ClueChain details
	for i, clue := range original.ClueChain {
		if restored.ClueChain[i].Tier != clue.Tier {
			t.Errorf("ClueChain[%d].Tier mismatch: got %d, want %d", i, restored.ClueChain[i].Tier, clue.Tier)
		}
		if restored.ClueChain[i].Content != clue.Content {
			t.Errorf("ClueChain[%d].Content mismatch: got %s, want %s", i, restored.ClueChain[i].Content, clue.Content)
		}
		if len(restored.ClueChain[i].Keywords) != len(clue.Keywords) {
			t.Errorf("ClueChain[%d].Keywords length mismatch: got %d, want %d", i, len(restored.ClueChain[i].Keywords), len(clue.Keywords))
		}
	}
}

// TestGlobalSeed_JSONFormatStability tests that JSON format is stable and parseable.
func TestGlobalSeed_JSONFormatStability(t *testing.T) {
	seed, err := NewGlobalSeed(
		"GS001",
		"Content",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "C1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "C2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "C3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create seed: %v", err)
	}

	// Marshal to JSON
	data, err := json.Marshal(seed)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify JSON contains expected fields
	var rawJSON map[string]interface{}
	err = json.Unmarshal(data, &rawJSON)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Check all expected fields are present
	expectedFields := []string{
		"id", "content", "linked_truth", "linked_ending",
		"current_tier", "clue_chain", "related_seeds", "related_rules",
		"created_at", "last_revealed",
	}

	for _, field := range expectedFields {
		if _, exists := rawJSON[field]; !exists {
			t.Errorf("Expected field %q not found in JSON", field)
		}
	}
}

// TestGlobalSeed_JSONRoundTrip tests that data survives multiple marshal/unmarshal cycles.
func TestGlobalSeed_JSONRoundTrip(t *testing.T) {
	original, err := NewGlobalSeed(
		"GS999",
		"Round trip test",
		"Truth survives",
		"hopeful",
		[]ClueTier{
			{Tier: 1, Content: "First", Keywords: []string{"alpha", "beta"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Second", Keywords: []string{"gamma", "delta"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Third", Keywords: []string{"epsilon", "zeta"}, BeatStart: 13, BeatEnd: 18},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create seed: %v", err)
	}

	original.AddRelatedSeed("GS001")
	original.AddRelatedRule("RULE999")

	// Round trip 3 times
	current := original
	for i := 0; i < 3; i++ {
		data, err := json.Marshal(current)
		if err != nil {
			t.Fatalf("Round %d: Marshal failed: %v", i+1, err)
		}

		var next GlobalSeed
		err = json.Unmarshal(data, &next)
		if err != nil {
			t.Fatalf("Round %d: Unmarshal failed: %v", i+1, err)
		}

		current = &next
	}

	// Verify final state matches original
	if current.ID != original.ID {
		t.Errorf("After round trips: ID mismatch: got %s, want %s", current.ID, original.ID)
	}
	if current.Content != original.Content {
		t.Errorf("After round trips: Content mismatch: got %s, want %s", current.Content, original.Content)
	}
	if len(current.RelatedSeeds) != len(original.RelatedSeeds) {
		t.Errorf("After round trips: RelatedSeeds count changed: got %d, want %d", len(current.RelatedSeeds), len(original.RelatedSeeds))
	}
}

// TestDeepCopy tests deep copying a GlobalSeed.
func TestDeepCopy(t *testing.T) {
	clueChain := []ClueTier{
		{Tier: 1, Content: "Clue 1", Keywords: []string{"hint", "shadow"}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "Clue 2", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
		{Tier: 3, Content: "Clue 3", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
	}

	original, _ := NewGlobalSeed("GS001", "Original content", "Original truth", "tragic", clueChain)
	original.AddRelatedSeed("GS002")
	original.AddRelatedRule("R001")

	// Create deep copy
	copy := original.DeepCopy()

	// Verify copy is equal
	if copy.ID != original.ID {
		t.Errorf("Expected ID '%s', got '%s'", original.ID, copy.ID)
	}
	if copy.Content != original.Content {
		t.Errorf("Expected Content '%s', got '%s'", original.Content, copy.Content)
	}
	if copy.CurrentTier != original.CurrentTier {
		t.Errorf("Expected CurrentTier %d, got %d", original.CurrentTier, copy.CurrentTier)
	}

	// Modify copy and verify original is unchanged
	copy.Content = "Modified content"
	copy.CurrentTier = 2
	copy.ClueChain[0].Content = "Modified clue"
	copy.ClueChain[0].Keywords[0] = "modified"
	copy.RelatedSeeds[0] = "GS999"
	copy.RelatedRules[0] = "R999"

	if original.Content == "Modified content" {
		t.Error("Deep copy failed: original Content was modified")
	}
	if original.CurrentTier == 2 {
		t.Error("Deep copy failed: original CurrentTier was modified")
	}
	if original.ClueChain[0].Content == "Modified clue" {
		t.Error("Deep copy failed: original ClueChain content was modified")
	}
	if original.ClueChain[0].Keywords[0] == "modified" {
		t.Error("Deep copy failed: original Keywords were modified")
	}
	if original.RelatedSeeds[0] == "GS999" {
		t.Error("Deep copy failed: original RelatedSeeds was modified")
	}
	if original.RelatedRules[0] == "R999" {
		t.Error("Deep copy failed: original RelatedRules was modified")
	}
}

// TestDeepCopy_NilSeed tests deep copying a nil GlobalSeed.
func TestDeepCopy_NilSeed(t *testing.T) {
	var seed *GlobalSeed
	copy := seed.DeepCopy()
	if copy != nil {
		t.Error("Expected nil copy for nil seed")
	}
}
