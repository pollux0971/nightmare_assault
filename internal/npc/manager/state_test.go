package manager

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMentalState_String(t *testing.T) {
	tests := []struct {
		name  string
		state MentalState
		want  string
	}{
		{"normal", Normal, "normal"},
		{"anxious", Anxious, "anxious"},
		{"corrupted", Corrupted, "corrupted"},
		{"unknown", MentalState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("MentalState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMentalState_JSON(t *testing.T) {
	tests := []struct {
		name  string
		state MentalState
		want  string
	}{
		{"normal", Normal, `"normal"`},
		{"anxious", Anxious, `"anxious"`},
		{"corrupted", Corrupted, `"corrupted"`},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.state)
			if err != nil {
				t.Fatalf("MentalState.MarshalJSON() error = %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("MentalState.MarshalJSON() = %v, want %v", string(data), tt.want)
			}
		})

		t.Run(tt.name+"_unmarshal", func(t *testing.T) {
			var state MentalState
			err := json.Unmarshal([]byte(tt.want), &state)
			if err != nil {
				t.Fatalf("MentalState.UnmarshalJSON() error = %v", err)
			}
			if state != tt.state {
				t.Errorf("MentalState.UnmarshalJSON() = %v, want %v", state, tt.state)
			}
		})
	}
}

func TestMentalState_UnmarshalJSON_Invalid(t *testing.T) {
	var state MentalState
	err := json.Unmarshal([]byte(`"invalid"`), &state)
	if err != nil {
		t.Fatalf("MentalState.UnmarshalJSON() should not error on invalid value, got %v", err)
	}
	if state != Normal {
		t.Errorf("MentalState.UnmarshalJSON() with invalid value should default to Normal, got %v", state)
	}
}

func TestRelationshipType_String(t *testing.T) {
	tests := []struct {
		name string
		rel  RelationshipType
		want string
	}{
		{"friendly", Friendly, "friendly"},
		{"neutral", Neutral, "neutral"},
		{"hostile", Hostile, "hostile"},
		{"fearful", Fearful, "fearful"},
		{"unknown", RelationshipType(999), "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rel.String(); got != tt.want {
				t.Errorf("RelationshipType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationshipType_JSON(t *testing.T) {
	tests := []struct {
		name string
		rel  RelationshipType
		want string
	}{
		{"friendly", Friendly, `"friendly"`},
		{"neutral", Neutral, `"neutral"`},
		{"hostile", Hostile, `"hostile"`},
		{"fearful", Fearful, `"fearful"`},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.rel)
			if err != nil {
				t.Fatalf("RelationshipType.MarshalJSON() error = %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("RelationshipType.MarshalJSON() = %v, want %v", string(data), tt.want)
			}
		})

		t.Run(tt.name+"_unmarshal", func(t *testing.T) {
			var rel RelationshipType
			err := json.Unmarshal([]byte(tt.want), &rel)
			if err != nil {
				t.Fatalf("RelationshipType.UnmarshalJSON() error = %v", err)
			}
			if rel != tt.rel {
				t.Errorf("RelationshipType.UnmarshalJSON() = %v, want %v", rel, tt.rel)
			}
		})
	}
}

func TestRelationshipType_UnmarshalJSON_Invalid(t *testing.T) {
	var rel RelationshipType
	err := json.Unmarshal([]byte(`"invalid"`), &rel)
	if err != nil {
		t.Fatalf("RelationshipType.UnmarshalJSON() should not error on invalid value, got %v", err)
	}
	if rel != Neutral {
		t.Errorf("RelationshipType.UnmarshalJSON() with invalid value should default to Neutral, got %v", rel)
	}
}

func TestCalculateRelationship(t *testing.T) {
	tests := []struct {
		name    string
		emotion EmotionState
		want    RelationshipType
	}{
		{
			name:    "friendly - high trust, low fear",
			emotion: EmotionState{Trust: 70, Fear: 20, Stress: 30},
			want:    Friendly,
		},
		{
			name:    "friendly - trust exactly 61, fear exactly 29",
			emotion: EmotionState{Trust: 61, Fear: 29, Stress: 50},
			want:    Friendly,
		},
		{
			name:    "fearful - low trust, high fear",
			emotion: EmotionState{Trust: 20, Fear: 70, Stress: 60},
			want:    Fearful,
		},
		{
			name:    "fearful - trust exactly 29, fear exactly 61",
			emotion: EmotionState{Trust: 29, Fear: 61, Stress: 50},
			want:    Fearful,
		},
		{
			name:    "hostile - low trust, low fear",
			emotion: EmotionState{Trust: 20, Fear: 20, Stress: 50},
			want:    Hostile,
		},
		{
			name:    "hostile - trust exactly 29, fear exactly 29",
			emotion: EmotionState{Trust: 29, Fear: 29, Stress: 50},
			want:    Hostile,
		},
		{
			name:    "neutral - medium trust and fear",
			emotion: EmotionState{Trust: 50, Fear: 50, Stress: 50},
			want:    Neutral,
		},
		{
			name:    "fearful - high trust, high fear (fear overrides)",
			emotion: EmotionState{Trust: 70, Fear: 70, Stress: 50},
			want:    Fearful,
		},
		{
			name:    "neutral - low trust, medium fear",
			emotion: EmotionState{Trust: 20, Fear: 40, Stress: 50},
			want:    Neutral,
		},
		{
			name:    "neutral - boundary case trust 30, fear 30",
			emotion: EmotionState{Trust: 30, Fear: 30, Stress: 50},
			want:    Neutral,
		},
		{
			name:    "friendly - boundary case trust 60, fear 30 (updated for Story 1.4)",
			emotion: EmotionState{Trust: 60, Fear: 30, Stress: 50},
			want:    Friendly,
		},
		{
			name:    "fearful - boundary case trust 30, fear 60 (updated for Story 1.4)",
			emotion: EmotionState{Trust: 30, Fear: 60, Stress: 50},
			want:    Fearful,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRelationship(tt.emotion)
			if got != tt.want {
				t.Errorf("CalculateRelationship(%+v) = %v, want %v",
					tt.emotion, got, tt.want)
			}
		})
	}
}

func TestNewNPCRuntimeState(t *testing.T) {
	state := NewNPCRuntimeState()

	if state == nil {
		t.Fatal("NewNPCRuntimeState() returned nil")
	}

	// Check default values
	if state.Emotion != DefaultEmotionState() {
		t.Errorf("NewNPCRuntimeState() emotion = %+v, want %+v",
			state.Emotion, DefaultEmotionState())
	}

	if state.MentalState != Normal {
		t.Errorf("NewNPCRuntimeState() mental_state = %v, want %v",
			state.MentalState, Normal)
	}

	if !state.IsAlive {
		t.Error("NewNPCRuntimeState() is_alive = false, want true")
	}

	if state.Relationship != Neutral {
		t.Errorf("NewNPCRuntimeState() relationship = %v, want %v",
			state.Relationship, Neutral)
	}

	if state.RelationshipScore != 0 {
		t.Errorf("NewNPCRuntimeState() relationship_score = %d, want 0",
			state.RelationshipScore)
	}

	if state.TraitStates == nil {
		t.Error("NewNPCRuntimeState() trait_states is nil, want empty map")
	}

	if len(state.TraitStates) != 0 {
		t.Errorf("NewNPCRuntimeState() trait_states length = %d, want 0",
			len(state.TraitStates))
	}

	if state.Interactions == nil {
		t.Error("NewNPCRuntimeState() interactions is nil, want empty slice")
	}

	if len(state.Interactions) != 0 {
		t.Errorf("NewNPCRuntimeState() interactions length = %d, want 0",
			len(state.Interactions))
	}

	if !state.LastInteraction.IsZero() {
		t.Error("NewNPCRuntimeState() last_interaction should be zero time")
	}
}

func TestNPCRuntimeState_AddInteraction(t *testing.T) {
	state := NewNPCRuntimeState()
	initialEmotion := state.Emotion

	// Add a friendly interaction
	state.AddInteraction("dialogue", "Had a friendly chat", EmotionDeltas["friendly_chat"])

	// Check interaction was added
	if len(state.Interactions) != 1 {
		t.Fatalf("AddInteraction() interactions length = %d, want 1",
			len(state.Interactions))
	}

	interaction := state.Interactions[0]
	if interaction.InteractionType != "dialogue" {
		t.Errorf("AddInteraction() interaction_type = %v, want dialogue",
			interaction.InteractionType)
	}

	if interaction.Description != "Had a friendly chat" {
		t.Errorf("AddInteraction() description = %v, want 'Had a friendly chat'",
			interaction.Description)
	}

	if interaction.EmotionDelta != EmotionDeltas["friendly_chat"] {
		t.Errorf("AddInteraction() emotion_delta = %+v, want %+v",
			interaction.EmotionDelta, EmotionDeltas["friendly_chat"])
	}

	if interaction.Timestamp.IsZero() {
		t.Error("AddInteraction() timestamp is zero")
	}

	// Check emotion was updated
	expectedEmotion := initialEmotion.Apply(EmotionDeltas["friendly_chat"])
	if state.Emotion != expectedEmotion {
		t.Errorf("AddInteraction() emotion = %+v, want %+v",
			state.Emotion, expectedEmotion)
	}

	// Check last interaction timestamp was updated
	if state.LastInteraction.IsZero() {
		t.Error("AddInteraction() last_interaction is zero")
	}

	if !state.LastInteraction.Equal(interaction.Timestamp) {
		t.Error("AddInteraction() last_interaction doesn't match interaction timestamp")
	}

	// Check relationship was recalculated
	expectedRelationship := CalculateRelationship(state.Emotion)
	if state.Relationship != expectedRelationship {
		t.Errorf("AddInteraction() relationship = %v, want %v",
			state.Relationship, expectedRelationship)
	}

	// Check relationship score was updated (using new formula from Story 1.4)
	expectedScore := CalculateRelationshipScore(state.Emotion)
	if state.RelationshipScore != expectedScore {
		t.Errorf("AddInteraction() relationship_score = %d, want %d",
			state.RelationshipScore, expectedScore)
	}
}

func TestNPCRuntimeState_AddInteraction_Multiple(t *testing.T) {
	state := NewNPCRuntimeState()

	// Add multiple interactions
	interactions := []struct {
		iType       string
		description string
		delta       EmotionDelta
	}{
		{"dialogue", "Friendly chat", EmotionDeltas["friendly_chat"]},
		{"help", "Helped the NPC", EmotionDeltas["help_npc"]},
		{"success", "Successfully completed task", EmotionDeltas["calm_down"]},
	}

	for i, inter := range interactions {
		state.AddInteraction(inter.iType, inter.description, inter.delta)

		if len(state.Interactions) != i+1 {
			t.Errorf("After interaction %d, interactions length = %d, want %d",
				i+1, len(state.Interactions), i+1)
		}
	}

	// Verify interactions are in order
	for i, inter := range interactions {
		if state.Interactions[i].InteractionType != inter.iType {
			t.Errorf("Interaction %d type = %v, want %v",
				i, state.Interactions[i].InteractionType, inter.iType)
		}
	}

	// Verify timestamps are chronologically ordered
	for i := 1; i < len(state.Interactions); i++ {
		if state.Interactions[i].Timestamp.Before(state.Interactions[i-1].Timestamp) {
			t.Errorf("Interaction %d timestamp is before interaction %d",
				i, i-1)
		}
	}
}

func TestNPCRuntimeState_AddInteraction_RelationshipChange(t *testing.T) {
	state := NewNPCRuntimeState()

	// Start with neutral state (Trust: 50, Fear: 25)
	if state.Relationship != Neutral {
		t.Fatalf("Initial relationship = %v, want Neutral", state.Relationship)
	}

	// Build trust to make relationship friendly
	// Need Trust > 60 && Fear < 30
	state.AddInteraction("help", "Help 1", EmotionDeltas["help_npc"])   // Trust +10, Fear -5
	state.AddInteraction("help", "Help 2", EmotionDeltas["help_npc"])   // Trust +10, Fear -5

	// Trust should be 70, Fear should be 15 -> Friendly
	if state.Relationship != Friendly {
		t.Errorf("After helping, relationship = %v, want Friendly (emotion: %+v)",
			state.Relationship, state.Emotion)
	}

	// Now threaten to make it fearful
	// Need Trust < 30 && Fear > 60
	state.AddInteraction("threat", "Threat 1", EmotionDeltas["threat"]) // Trust -10, Fear +15
	state.AddInteraction("threat", "Threat 2", EmotionDeltas["threat"]) // Trust -10, Fear +15
	state.AddInteraction("threat", "Threat 3", EmotionDeltas["threat"]) // Trust -10, Fear +15
	state.AddInteraction("threat", "Threat 4", EmotionDeltas["threat"]) // Trust -10, Fear +15
	state.AddInteraction("threat", "Threat 5", EmotionDeltas["threat"]) // Trust -10, Fear +15

	// Trust should be 20, Fear should be 90 (clamped to 100) -> Fearful
	if state.Relationship != Fearful {
		t.Errorf("After threats, relationship = %v, want Fearful (emotion: %+v)",
			state.Relationship, state.Emotion)
	}
}

func TestNPCRuntimeState_GetRecentInteractions(t *testing.T) {
	state := NewNPCRuntimeState()

	// Add 5 interactions
	for i := 0; i < 5; i++ {
		state.AddInteraction("test", "Interaction", EmotionDeltas["friendly_chat"])
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	tests := []struct {
		name string
		n    int
		want int
	}{
		{"get last 3", 3, 3},
		{"get all", 5, 5},
		{"get more than available", 10, 5},
		{"get zero", 0, 0},
		{"get negative", -1, 0},
		{"get one", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.GetRecentInteractions(tt.n)
			if len(got) != tt.want {
				t.Errorf("GetRecentInteractions(%d) length = %d, want %d",
					tt.n, len(got), tt.want)
			}

			// Verify we got the most recent ones
			if tt.want > 0 && tt.want <= 5 {
				// The last interaction should match
				lastGot := got[len(got)-1]
				lastAll := state.Interactions[len(state.Interactions)-1]
				if !lastGot.Timestamp.Equal(lastAll.Timestamp) {
					t.Error("GetRecentInteractions() did not return most recent interactions")
				}
			}
		})
	}
}

func TestNPCRuntimeState_GetRecentInteractions_EmptyState(t *testing.T) {
	state := NewNPCRuntimeState()

	result := state.GetRecentInteractions(5)
	if len(result) != 0 {
		t.Errorf("GetRecentInteractions(5) on empty state = %d interactions, want 0",
			len(result))
	}
}

func TestNPCRuntimeState_GetRecentInteractions_Independence(t *testing.T) {
	state := NewNPCRuntimeState()

	// Add some interactions
	state.AddInteraction("test1", "First", EmotionDeltas["friendly_chat"])
	state.AddInteraction("test2", "Second", EmotionDeltas["help_npc"])

	// Get recent interactions and modify the returned slice
	recent := state.GetRecentInteractions(2)
	recent[0].Description = "Modified"

	// Original should be unchanged
	if state.Interactions[0].Description == "Modified" {
		t.Error("Modifying returned slice affected the original state")
	}
}

func TestNPCRuntimeState_UpdateMentalState(t *testing.T) {
	tests := []struct {
		name    string
		emotion EmotionState
		want    MentalState
	}{
		{
			name:    "anxious - high stress but below corruption threshold (Story 1.5)",
			emotion: EmotionState{Trust: 30, Fear: 70, Stress: 80},
			want:    Anxious, // Story 1.5: Corrupted requires Stress >= 90
		},
		{
			name:    "anxious - boundary case stress 71, fear 61 (Story 1.5)",
			emotion: EmotionState{Trust: 30, Fear: 61, Stress: 71},
			want:    Anxious, // Story 1.5: Corrupted requires Stress >= 90
		},
		{
			name:    "anxious - high stress at threshold (Story 1.5: Stress >= 60)",
			emotion: EmotionState{Trust: 50, Fear: 40, Stress: 60},
			want:    Anxious,
		},
		{
			name:    "anxious - high fear at threshold (Story 1.5: Fear >= 70)",
			emotion: EmotionState{Trust: 50, Fear: 70, Stress: 40},
			want:    Anxious, // Changed from Fear: 60 to Fear: 70 to meet threshold
		},
		{
			name:    "normal - below anxious thresholds (Story 1.5: Stress < 60, Fear < 70)",
			emotion: EmotionState{Trust: 50, Fear: 40, Stress: 51},
			want:    Normal, // Changed from Anxious because Stress < 60 and Fear < 70
		},
		{
			name:    "anxious - high fear (Story 1.4: Fear >= 70)",
			emotion: EmotionState{Trust: 50, Fear: 70, Stress: 40},
			want:    Anxious,
		},
		{
			name:    "normal - low stress and fear",
			emotion: EmotionState{Trust: 70, Fear: 30, Stress: 40},
			want:    Normal,
		},
		{
			name:    "normal - boundary case stress 50, fear 50",
			emotion: EmotionState{Trust: 50, Fear: 50, Stress: 50},
			want:    Normal,
		},
		{
			name:    "not corrupted - high stress but not enough fear",
			emotion: EmotionState{Trust: 30, Fear: 60, Stress: 80},
			want:    Anxious, // Falls to anxious because fear > 50
		},
		{
			name:    "not corrupted - high fear but not enough stress",
			emotion: EmotionState{Trust: 30, Fear: 70, Stress: 70},
			want:    Anxious, // Falls to anxious because stress > 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewNPCRuntimeState()
			state.Emotion = tt.emotion
			state.UpdateMentalState()

			if state.MentalState != tt.want {
				t.Errorf("UpdateMentalState() with emotion %+v = %v, want %v",
					tt.emotion, state.MentalState, tt.want)
			}
		})
	}
}

func TestNPCRuntimeState_Copy(t *testing.T) {
	original := NewNPCRuntimeState()

	// Populate with data
	original.Emotion = EmotionState{Trust: 60, Fear: 30, Stress: 40}
	original.MentalState = Anxious
	original.IsAlive = true
	original.Relationship = Friendly
	original.RelationshipScore = 30
	original.TraitStates["brave"] = "active"
	original.TraitStates["paranoid"] = "dormant"
	original.AddInteraction("test", "Test interaction", EmotionDeltas["friendly_chat"])

	// Create copy
	copied := original.Copy()

	// Verify values are equal
	if copied.Emotion != original.Emotion {
		t.Errorf("Copy() emotion = %+v, want %+v", copied.Emotion, original.Emotion)
	}

	if copied.MentalState != original.MentalState {
		t.Errorf("Copy() mental_state = %v, want %v", copied.MentalState, original.MentalState)
	}

	if copied.IsAlive != original.IsAlive {
		t.Errorf("Copy() is_alive = %v, want %v", copied.IsAlive, original.IsAlive)
	}

	if copied.Relationship != original.Relationship {
		t.Errorf("Copy() relationship = %v, want %v", copied.Relationship, original.Relationship)
	}

	if copied.RelationshipScore != original.RelationshipScore {
		t.Errorf("Copy() relationship_score = %d, want %d",
			copied.RelationshipScore, original.RelationshipScore)
	}

	if len(copied.TraitStates) != len(original.TraitStates) {
		t.Errorf("Copy() trait_states length = %d, want %d",
			len(copied.TraitStates), len(original.TraitStates))
	}

	if len(copied.Interactions) != len(original.Interactions) {
		t.Errorf("Copy() interactions length = %d, want %d",
			len(copied.Interactions), len(original.Interactions))
	}

	// Verify it's a deep copy by modifying the copy
	copied.Emotion.Trust = 100
	if original.Emotion.Trust == 100 {
		t.Error("Modifying copy emotion affected original")
	}

	copied.MentalState = Corrupted
	if original.MentalState == Corrupted {
		t.Error("Modifying copy mental_state affected original")
	}

	copied.TraitStates["brave"] = "modified"
	if original.TraitStates["brave"] == "modified" {
		t.Error("Modifying copy trait_states affected original")
	}

	copied.Interactions[0].Description = "Modified"
	if original.Interactions[0].Description == "Modified" {
		t.Error("Modifying copy interactions affected original")
	}
}

func TestNPCRuntimeState_String(t *testing.T) {
	state := NewNPCRuntimeState()
	state.AddInteraction("test", "Test interaction", EmotionDeltas["friendly_chat"])

	jsonStr := state.String()

	if jsonStr == "" {
		t.Error("String() returned empty string")
	}

	// Verify it's valid JSON
	if jsonStr[0] != '{' || jsonStr[len(jsonStr)-1] != '}' {
		t.Errorf("String() = %q, doesn't look like JSON", jsonStr)
	}

	// Try to unmarshal it back
	var decoded NPCRuntimeState
	err := json.Unmarshal([]byte(jsonStr), &decoded)
	if err != nil {
		t.Errorf("String() produced invalid JSON: %v", err)
	}

	// Verify key fields are present
	if decoded.Emotion != state.Emotion {
		t.Error("String() JSON missing or incorrect emotion field")
	}

	if decoded.MentalState != state.MentalState {
		t.Error("String() JSON missing or incorrect mental_state field")
	}

	if decoded.IsAlive != state.IsAlive {
		t.Error("String() JSON missing or incorrect is_alive field")
	}
}

func TestNPCRuntimeState_FullWorkflow(t *testing.T) {
	// Simulate a complete gameplay scenario
	state := NewNPCRuntimeState()

	// 1. Initial state should be neutral
	if state.Relationship != Neutral {
		t.Errorf("Initial relationship = %v, want Neutral", state.Relationship)
	}

	if state.MentalState != Normal {
		t.Errorf("Initial mental_state = %v, want Normal", state.MentalState)
	}

	// 2. Build trust through positive interactions
	state.AddInteraction("dialogue", "Had a friendly chat", EmotionDeltas["friendly_chat"])
	state.AddInteraction("help", "Helped with a task", EmotionDeltas["help_npc"])
	state.AddInteraction("success", "Successfully completed quest together", EmotionDeltas["calm_down"])

	// Should now be friendly
	if state.Relationship != Friendly {
		t.Errorf("After positive interactions, relationship = %v, want Friendly (emotion: %+v)",
			state.Relationship, state.Emotion)
	}

	// Mental state should still be normal
	state.UpdateMentalState()
	if state.MentalState != Normal {
		t.Errorf("After positive interactions, mental_state = %v, want Normal", state.MentalState)
	}

	// 3. Introduce stress through failures and traumatic events
	// Note: After Story 1.4, threshold is Stress >= 60 OR Fear >= 70 for Anxious
	state.AddInteraction("failure", "Failed to protect NPC", EmotionDeltas["ignore_distress"])
	state.AddInteraction("trauma", "Witnessed something horrific", EmotionDeltas["witness_death"])
	state.AddInteraction("stress", "Under immense pressure", EmotionDeltas["threat"])
	state.AddInteraction("failure", "Another failure", EmotionDeltas["catch_lying"])

	// Update mental state - should become anxious (stress >= 60 or fear >= 70)
	state.UpdateMentalState()
	if state.MentalState != Anxious {
		t.Errorf("After failures, mental_state = %v, want Anxious (emotion: %+v)",
			state.MentalState, state.Emotion)
	}

	// 4. Verify interaction history
	interactions := state.GetRecentInteractions(3)
	if len(interactions) != 3 {
		t.Fatalf("GetRecentInteractions(3) = %d, want 3", len(interactions))
	}

	// Last interaction should be a failure
	if interactions[2].InteractionType != "failure" {
		t.Errorf("Last interaction type = %v, want 'failure'", interactions[2].InteractionType)
	}

	// 5. Test state persistence through copy
	copied := state.Copy()
	if copied.String() != state.String() {
		t.Error("Copied state differs from original")
	}
}

func TestNPCInteraction_Fields(t *testing.T) {
	timestamp := time.Now()
	interaction := NPCInteraction{
		Timestamp:       timestamp,
		InteractionType: "dialogue",
		EmotionDelta:    EmotionDeltas["friendly_chat"],
		Description:     "Test description",
	}

	if !interaction.Timestamp.Equal(timestamp) {
		t.Error("NPCInteraction.Timestamp not set correctly")
	}

	if interaction.InteractionType != "dialogue" {
		t.Error("NPCInteraction.InteractionType not set correctly")
	}

	if interaction.EmotionDelta != EmotionDeltas["friendly_chat"] {
		t.Error("NPCInteraction.EmotionDelta not set correctly")
	}

	if interaction.Description != "Test description" {
		t.Error("NPCInteraction.Description not set correctly")
	}
}

func TestNPCInteraction_JSON(t *testing.T) {
	interaction := NPCInteraction{
		Timestamp:       time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC),
		InteractionType: "test",
		EmotionDelta:    EmotionDeltas["friendly_chat"],
		Description:     "Test interaction",
	}

	// Marshal to JSON
	data, err := json.Marshal(interaction)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var decoded NPCInteraction
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields
	if !decoded.Timestamp.Equal(interaction.Timestamp) {
		t.Error("JSON round-trip changed Timestamp")
	}

	if decoded.InteractionType != interaction.InteractionType {
		t.Error("JSON round-trip changed InteractionType")
	}

	if decoded.EmotionDelta != interaction.EmotionDelta {
		t.Error("JSON round-trip changed EmotionDelta")
	}

	if decoded.Description != interaction.Description {
		t.Error("JSON round-trip changed Description")
	}
}
