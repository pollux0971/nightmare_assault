package manager

import (
	"encoding/json"
	"testing"
)

// ==========================================================================
// Story 1.6: Trait Structure Tests
// ==========================================================================

func TestTraitStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status TraitStatus
		want   string
	}{
		{"Hidden status", Hidden, "hidden"},
		// Story 8.1: Hinting is now an alias for HintPhase1, returns "hint_phase_1"
		{"Hinting status", Hinting, "hint_phase_1"},
		{"Revealed status", Revealed, "revealed"},
		{"Invalid status", TraitStatus(999), "hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("TraitStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTraitStatus_JSON(t *testing.T) {
	tests := []struct {
		name   string
		status TraitStatus
	}{
		{"Hidden", Hidden},
		{"Hinting", Hinting},
		{"Revealed", Revealed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("Failed to marshal TraitStatus: %v", err)
			}

			// Unmarshal
			var decoded TraitStatus
			err = json.Unmarshal(data, &decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal TraitStatus: %v", err)
			}

			// Verify
			if decoded != tt.status {
				t.Errorf("JSON round-trip failed: got %v, want %v", decoded, tt.status)
			}
		})
	}
}

func TestTraitStatus_UnmarshalJSON_Invalid(t *testing.T) {
	var status TraitStatus
	err := json.Unmarshal([]byte(`"invalid"`), &status)
	if err != nil {
		t.Fatalf("UnmarshalJSON should not error on invalid string: %v", err)
	}
	if status != Hidden {
		t.Errorf("Invalid status should default to Hidden, got %v", status)
	}
}

func TestTriggerType_String(t *testing.T) {
	tests := []struct {
		name        string
		triggerType TriggerType
		want        string
	}{
		{"TrustLevel", TrustLevel, "trust_level"},
		{"FearLevel", FearLevel, "fear_level"},
		{"StressLevel", StressLevel, "stress_level"},
		{"Event", Event, "event"},
		{"InteractionCount", InteractionCount, "interaction_count"},
		{"TimeBased", TimeBased, "time_based"},
		{"Invalid", TriggerType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.triggerType.String()
			if got != tt.want {
				t.Errorf("TriggerType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTriggerType_JSON(t *testing.T) {
	tests := []struct {
		name        string
		triggerType TriggerType
	}{
		{"TrustLevel", TrustLevel},
		{"FearLevel", FearLevel},
		{"StressLevel", StressLevel},
		{"Event", Event},
		{"InteractionCount", InteractionCount},
		{"TimeBased", TimeBased},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.triggerType)
			if err != nil {
				t.Fatalf("Failed to marshal TriggerType: %v", err)
			}

			// Unmarshal
			var decoded TriggerType
			err = json.Unmarshal(data, &decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal TriggerType: %v", err)
			}

			// Verify
			if decoded != tt.triggerType {
				t.Errorf("JSON round-trip failed: got %v, want %v", decoded, tt.triggerType)
			}
		})
	}
}

func TestTriggerType_UnmarshalJSON_Invalid(t *testing.T) {
	var triggerType TriggerType
	err := json.Unmarshal([]byte(`"invalid"`), &triggerType)
	if err != nil {
		t.Fatalf("UnmarshalJSON should not error on invalid string: %v", err)
	}
	if triggerType != TrustLevel {
		t.Errorf("Invalid trigger type should default to TrustLevel, got %v", triggerType)
	}
}

func TestTraitFull_ToBasicTrait(t *testing.T) {
	fullTrait := TraitFull{
		ID:         "trait_paranoid",
		Content:    "極度偏執，不信任他人",
		RevealTier: 2,
		Triggers: []TraitTrigger{
			{Type: TrustLevel, Comparator: "<=", Threshold: 30},
		},
		Status: Hidden,
		Hints:  []string{"眼神充滿懷疑"},
	}

	basicTrait := fullTrait.ToBasicTrait()

	if basicTrait.ID != fullTrait.ID {
		t.Errorf("ToBasicTrait() ID mismatch: got %v, want %v", basicTrait.ID, fullTrait.ID)
	}
	if basicTrait.Content != fullTrait.Content {
		t.Errorf("ToBasicTrait() Content mismatch: got %v, want %v", basicTrait.Content, fullTrait.Content)
	}
	if basicTrait.RevealTier != fullTrait.RevealTier {
		t.Errorf("ToBasicTrait() RevealTier mismatch: got %v, want %v", basicTrait.RevealTier, fullTrait.RevealTier)
	}
}

func TestFromBasicTrait(t *testing.T) {
	basicTrait := Trait{
		ID:         "trait_brave",
		Content:    "勇敢無畏",
		RevealTier: 1,
	}

	fullTrait := FromBasicTrait(basicTrait)

	if fullTrait.ID != basicTrait.ID {
		t.Errorf("FromBasicTrait() ID mismatch: got %v, want %v", fullTrait.ID, basicTrait.ID)
	}
	if fullTrait.Content != basicTrait.Content {
		t.Errorf("FromBasicTrait() Content mismatch: got %v, want %v", fullTrait.Content, basicTrait.Content)
	}
	if fullTrait.RevealTier != basicTrait.RevealTier {
		t.Errorf("FromBasicTrait() RevealTier mismatch: got %v, want %v", fullTrait.RevealTier, basicTrait.RevealTier)
	}
	if fullTrait.Status != Hidden {
		t.Errorf("FromBasicTrait() Status should be Hidden, got %v", fullTrait.Status)
	}
	if len(fullTrait.Triggers) != 0 {
		t.Errorf("FromBasicTrait() Triggers should be empty, got %d items", len(fullTrait.Triggers))
	}
	if len(fullTrait.Hints) != 0 {
		t.Errorf("FromBasicTrait() Hints should be empty, got %d items", len(fullTrait.Hints))
	}
}

func TestTraitTrigger_JSON(t *testing.T) {
	trigger := TraitTrigger{
		Type:       TrustLevel,
		Threshold:  30,
		EventName:  "",
		Comparator: "<=",
	}

	// Marshal
	data, err := json.Marshal(trigger)
	if err != nil {
		t.Fatalf("Failed to marshal TraitTrigger: %v", err)
	}

	// Unmarshal
	var decoded TraitTrigger
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal TraitTrigger: %v", err)
	}

	// Verify
	if decoded.Type != trigger.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, trigger.Type)
	}
	if decoded.Threshold != trigger.Threshold {
		t.Errorf("Threshold mismatch: got %v, want %v", decoded.Threshold, trigger.Threshold)
	}
	if decoded.Comparator != trigger.Comparator {
		t.Errorf("Comparator mismatch: got %v, want %v", decoded.Comparator, trigger.Comparator)
	}
}

func TestTraitFull_JSON(t *testing.T) {
	trait := TraitFull{
		ID:         "trait_paranoid",
		Content:    "極度偏執，不信任他人",
		RevealTier: 2,
		Triggers: []TraitTrigger{
			{Type: TrustLevel, Comparator: "<=", Threshold: 30},
			{Type: InteractionCount, Comparator: ">=", Threshold: 5},
		},
		Status: Hidden,
		Hints: []string{
			"他的眼神充滿懷疑",
			"他總是檢查門窗是否上鎖",
		},
	}

	// Marshal
	data, err := json.Marshal(trait)
	if err != nil {
		t.Fatalf("Failed to marshal TraitFull: %v", err)
	}

	// Unmarshal
	var decoded TraitFull
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal TraitFull: %v", err)
	}

	// Verify basic fields
	if decoded.ID != trait.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, trait.ID)
	}
	if decoded.Content != trait.Content {
		t.Errorf("Content mismatch: got %v, want %v", decoded.Content, trait.Content)
	}
	if decoded.RevealTier != trait.RevealTier {
		t.Errorf("RevealTier mismatch: got %v, want %v", decoded.RevealTier, trait.RevealTier)
	}
	if decoded.Status != trait.Status {
		t.Errorf("Status mismatch: got %v, want %v", decoded.Status, trait.Status)
	}

	// Verify triggers
	if len(decoded.Triggers) != len(trait.Triggers) {
		t.Errorf("Triggers length mismatch: got %v, want %v", len(decoded.Triggers), len(trait.Triggers))
	}

	// Verify hints
	if len(decoded.Hints) != len(trait.Hints) {
		t.Errorf("Hints length mismatch: got %v, want %v", len(decoded.Hints), len(trait.Hints))
	}
}

func TestRevealContext_JSON(t *testing.T) {
	context := RevealContext{
		CurrentBeat:      10,
		RecentEvents:     []string{"witness_death", "hallucination"},
		InteractionCount: 5,
	}

	// Marshal
	data, err := json.Marshal(context)
	if err != nil {
		t.Fatalf("Failed to marshal RevealContext: %v", err)
	}

	// Unmarshal
	var decoded RevealContext
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RevealContext: %v", err)
	}

	// Verify
	if decoded.CurrentBeat != context.CurrentBeat {
		t.Errorf("CurrentBeat mismatch: got %v, want %v", decoded.CurrentBeat, context.CurrentBeat)
	}
	if decoded.InteractionCount != context.InteractionCount {
		t.Errorf("InteractionCount mismatch: got %v, want %v", decoded.InteractionCount, context.InteractionCount)
	}
	if len(decoded.RecentEvents) != len(context.RecentEvents) {
		t.Errorf("RecentEvents length mismatch: got %v, want %v", len(decoded.RecentEvents), len(context.RecentEvents))
	}
}
