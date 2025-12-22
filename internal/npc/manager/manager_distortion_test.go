package manager

import (
	"testing"
)

// TestGetNPCEmotion tests the GetNPCEmotion method for distortion integration
func TestGetNPCEmotion(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	// Create test NPC
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(60, 40, 50),
	}

	err := mgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("failed to add NPC: %v", err)
	}

	// Test GetNPCEmotion
	trust, fear, stress, err := mgr.GetNPCEmotion("npc1")
	if err != nil {
		t.Fatalf("GetNPCEmotion failed: %v", err)
	}

	if trust != 60 {
		t.Errorf("expected trust 60, got %d", trust)
	}
	if fear != 40 {
		t.Errorf("expected fear 40, got %d", fear)
	}
	if stress != 50 {
		t.Errorf("expected stress 50, got %d", stress)
	}
}

// TestGetNPCEmotion_NonexistentNPC tests error handling
func TestGetNPCEmotion_NonexistentNPC(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	_, _, _, err := mgr.GetNPCEmotion("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent NPC, got nil")
	}
}

// TestGetNPCTraits tests the GetNPCTraits method for distortion integration
func TestGetNPCTraits(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	// Create test NPC with traits
	profile := &NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		Traits: []Trait{
			{ID: "anxious", Content: "Anxious personality"},
			{ID: "paranoid", Content: "Paranoid thoughts"},
			{ID: "rational", Content: "Rational thinking"},
		},
	}

	err := mgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("failed to add NPC: %v", err)
	}

	// Test GetNPCTraits
	traits, err := mgr.GetNPCTraits("npc1")
	if err != nil {
		t.Fatalf("GetNPCTraits failed: %v", err)
	}

	if len(traits) != 3 {
		t.Errorf("expected 3 traits, got %d", len(traits))
	}

	// Verify trait IDs are correct
	expectedTraits := map[string]bool{
		"anxious":  false,
		"paranoid": false,
		"rational": false,
	}

	for _, traitID := range traits {
		if _, exists := expectedTraits[traitID]; !exists {
			t.Errorf("unexpected trait ID: %s", traitID)
		}
		expectedTraits[traitID] = true
	}

	// Verify all expected traits were found
	for traitID, found := range expectedTraits {
		if !found {
			t.Errorf("expected trait %s not found", traitID)
		}
	}
}

// TestGetNPCTraits_NonexistentNPC tests error handling
func TestGetNPCTraits_NonexistentNPC(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	_, err := mgr.GetNPCTraits("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent NPC, got nil")
	}
}

// TestGetNPCTraits_NoTraits tests NPC with no traits
func TestGetNPCTraits_NoTraits(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	// Create test NPC with no traits
	profile := &NPCProfile{
		ID:     "npc1",
		Name:   "Test NPC",
		Traits: []Trait{},
	}

	err := mgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("failed to add NPC: %v", err)
	}

	// Test GetNPCTraits
	traits, err := mgr.GetNPCTraits("npc1")
	if err != nil {
		t.Fatalf("GetNPCTraits failed: %v", err)
	}

	if len(traits) != 0 {
		t.Errorf("expected 0 traits, got %d", len(traits))
	}
}

// TestGetNPCEmotion_AfterEmotionChange tests emotion retrieval after changes
func TestGetNPCEmotion_AfterEmotionChange(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	// Create test NPC
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 25, 25),
	}

	err := mgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("failed to add NPC: %v", err)
	}

	// Adjust emotion
	delta := EmotionDelta{Trust: 20, Fear: 30, Stress: 10}
	err = mgr.AdjustEmotion("npc1", delta)
	if err != nil {
		t.Fatalf("AdjustEmotion failed: %v", err)
	}

	// Get emotion after change
	trust, fear, stress, err := mgr.GetNPCEmotion("npc1")
	if err != nil {
		t.Fatalf("GetNPCEmotion failed: %v", err)
	}

	expectedTrust := 70  // 50 + 20
	expectedFear := 55   // 25 + 30
	expectedStress := 35 // 25 + 10

	if trust != expectedTrust {
		t.Errorf("expected trust %d, got %d", expectedTrust, trust)
	}
	if fear != expectedFear {
		t.Errorf("expected fear %d, got %d", expectedFear, fear)
	}
	if stress != expectedStress {
		t.Errorf("expected stress %d, got %d", expectedStress, stress)
	}
}
