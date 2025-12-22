package manager

import (
	"testing"
	"time"
)

// TestRecordInteraction tests the RecordInteraction method.
// Story 4.6: AC4 - Emotion changes recorded to NPCInteraction history
func TestRecordInteraction(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	// Add test NPC
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 50, 50),
	}
	err := mgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("AddNPC failed: %v", err)
	}

	// Create interaction
	interaction := NPCInteraction{
		Timestamp:       time.Now(),
		InteractionType: "test_interaction",
		EmotionDelta:    EmotionDelta{Trust: 10, Fear: -5, Stress: 0},
		Description:     "Test interaction description",
	}

	// Record interaction
	err = mgr.RecordInteraction("npc1", interaction)
	if err != nil {
		t.Fatalf("RecordInteraction failed: %v", err)
	}

	// Verify interaction was recorded
	state := mgr.GetState("npc1")
	if state == nil {
		t.Fatal("State is nil")
	}
	if len(state.Interactions) != 1 {
		t.Fatalf("Expected 1 interaction, got %d", len(state.Interactions))
	}
	if state.Interactions[0].InteractionType != "test_interaction" {
		t.Errorf("Wrong interaction type: %s", state.Interactions[0].InteractionType)
	}
	if state.Interactions[0].EmotionDelta.Trust != 10 {
		t.Errorf("Wrong trust delta: %d", state.Interactions[0].EmotionDelta.Trust)
	}
}

// TestRecordInteraction_InvalidNPC tests error handling for invalid NPC ID.
func TestRecordInteraction_InvalidNPC(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	interaction := NPCInteraction{
		Timestamp:       time.Now(),
		InteractionType: "test",
		EmotionDelta:    EmotionDelta{},
		Description:     "Test",
	}

	err := mgr.RecordInteraction("nonexistent", interaction)
	if err == nil {
		t.Error("Expected error for non-existent NPC")
	}
}

// TestRecordInteraction_HistoryLimit tests that history is limited to 100 interactions.
func TestRecordInteraction_HistoryLimit(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 50, 50),
	}
	mgr.AddNPC(profile)

	// Add 150 interactions
	for i := 0; i < 150; i++ {
		interaction := NPCInteraction{
			Timestamp:       time.Now(),
			InteractionType: "test",
			EmotionDelta:    EmotionDelta{Trust: 1},
			Description:     "Test",
		}
		mgr.RecordInteraction("npc1", interaction)
	}

	// Verify only last 100 are kept
	state := mgr.GetState("npc1")
	if len(state.Interactions) != 100 {
		t.Errorf("Expected 100 interactions, got %d", len(state.Interactions))
	}
}
