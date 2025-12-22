package manager

import (
	"testing"
)

// ==========================================================================
// Story 1.6: Trait Revelation Logic Tests
// ==========================================================================

// Test compareValues function with different comparators
func TestCompareValues(t *testing.T) {
	tests := []struct {
		name       string
		actual     int
		threshold  int
		comparator string
		want       bool
	}{
		// >= operator
		{"Greater than with >=", 60, 50, ">=", true},
		{"Equal with >=", 50, 50, ">=", true},
		{"Less than with >=", 40, 50, ">=", false},

		// <= operator
		{"Less than with <=", 40, 50, "<=", true},
		{"Equal with <=", 50, 50, "<=", true},
		{"Greater than with <=", 60, 50, "<=", false},

		// == operator
		{"Equal with ==", 50, 50, "==", true},
		{"Not equal with ==", 60, 50, "==", false},

		// > operator
		{"Greater with >", 60, 50, ">", true},
		{"Equal with >", 50, 50, ">", false},
		{"Less with >", 40, 50, ">", false},

		// < operator
		{"Less with <", 40, 50, "<", true},
		{"Equal with <", 50, 50, "<", false},
		{"Greater with <", 60, 50, "<", false},

		// Invalid comparator
		{"Invalid comparator", 50, 50, "!=", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareValues(tt.actual, tt.threshold, tt.comparator)
			if got != tt.want {
				t.Errorf("compareValues(%d, %d, %q) = %v, want %v",
					tt.actual, tt.threshold, tt.comparator, got, tt.want)
			}
		})
	}
}

// Test containsEvent function
func TestContainsEvent(t *testing.T) {
	events := []string{"witness_death", "hallucination", "combat"}

	tests := []struct {
		name      string
		eventName string
		want      bool
	}{
		{"Event exists - exact match", "witness_death", true},
		{"Event exists - different case", "WITNESS_DEATH", true},
		{"Event exists - mixed case", "Hallucination", true},
		{"Event does not exist", "friendly_chat", false},
		{"Empty event name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsEvent(events, tt.eventName)
			if got != tt.want {
				t.Errorf("containsEvent(%v, %q) = %v, want %v",
					events, tt.eventName, got, tt.want)
			}
		})
	}
}

// Test evaluateTrigger with different trigger types
func TestEvaluateTrigger(t *testing.T) {
	state := &NPCRuntimeState{
		Emotion: EmotionState{
			Trust:  60,
			Fear:   40,
			Stress: 50,
		},
	}

	context := RevealContext{
		CurrentBeat:      10,
		RecentEvents:     []string{"witness_death", "hallucination"},
		InteractionCount: 5,
	}

	tests := []struct {
		name    string
		trigger TraitTrigger
		want    bool
	}{
		// Trust level triggers
		{
			"Trust level satisfied (>=)",
			TraitTrigger{Type: TrustLevel, Threshold: 50, Comparator: ">="},
			true,
		},
		{
			"Trust level not satisfied (>=)",
			TraitTrigger{Type: TrustLevel, Threshold: 70, Comparator: ">="},
			false,
		},

		// Fear level triggers
		{
			"Fear level satisfied (<=)",
			TraitTrigger{Type: FearLevel, Threshold: 50, Comparator: "<="},
			true,
		},
		{
			"Fear level not satisfied (<=)",
			TraitTrigger{Type: FearLevel, Threshold: 30, Comparator: "<="},
			false,
		},

		// Stress level triggers
		{
			"Stress level satisfied (==)",
			TraitTrigger{Type: StressLevel, Threshold: 50, Comparator: "=="},
			true,
		},
		{
			"Stress level not satisfied (==)",
			TraitTrigger{Type: StressLevel, Threshold: 60, Comparator: "=="},
			false,
		},

		// Interaction count triggers
		{
			"Interaction count satisfied (>=)",
			TraitTrigger{Type: InteractionCount, Threshold: 5, Comparator: ">="},
			true,
		},
		{
			"Interaction count not satisfied (>=)",
			TraitTrigger{Type: InteractionCount, Threshold: 10, Comparator: ">="},
			false,
		},

		// Time-based triggers
		{
			"Time-based satisfied (>=)",
			TraitTrigger{Type: TimeBased, Threshold: 10, Comparator: ">="},
			true,
		},
		{
			"Time-based not satisfied (>)",
			TraitTrigger{Type: TimeBased, Threshold: 10, Comparator: ">"},
			false,
		},

		// Event triggers
		{
			"Event trigger satisfied",
			TraitTrigger{Type: Event, EventName: "witness_death"},
			true,
		},
		{
			"Event trigger not satisfied",
			TraitTrigger{Type: Event, EventName: "friendly_chat"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateTrigger(tt.trigger, state, context)
			if got != tt.want {
				t.Errorf("evaluateTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test shouldReveal with multiple triggers (AND logic)
func TestShouldReveal(t *testing.T) {
	tests := []struct {
		name    string
		trait   TraitFull
		state   *NPCRuntimeState
		context RevealContext
		want    bool
	}{
		{
			name: "All triggers satisfied",
			trait: TraitFull{
				ID: "trait_paranoid",
				Triggers: []TraitTrigger{
					{Type: TrustLevel, Threshold: 30, Comparator: "<="},
					{Type: InteractionCount, Threshold: 5, Comparator: ">="},
				},
			},
			state: &NPCRuntimeState{
				Emotion: EmotionState{Trust: 25, Fear: 40, Stress: 50},
			},
			context: RevealContext{
				InteractionCount: 6,
			},
			want: true,
		},
		{
			name: "One trigger not satisfied",
			trait: TraitFull{
				ID: "trait_paranoid",
				Triggers: []TraitTrigger{
					{Type: TrustLevel, Threshold: 30, Comparator: "<="},
					{Type: InteractionCount, Threshold: 10, Comparator: ">="},
				},
			},
			state: &NPCRuntimeState{
				Emotion: EmotionState{Trust: 25, Fear: 40, Stress: 50},
			},
			context: RevealContext{
				InteractionCount: 5,
			},
			want: false,
		},
		{
			name: "No triggers defined",
			trait: TraitFull{
				ID:       "trait_brave",
				Triggers: []TraitTrigger{},
			},
			state: &NPCRuntimeState{
				Emotion: EmotionState{Trust: 50, Fear: 20, Stress: 30},
			},
			context: RevealContext{},
			want:    false,
		},
		{
			name: "Complex multi-trigger satisfied",
			trait: TraitFull{
				ID: "trait_complex",
				Triggers: []TraitTrigger{
					{Type: TrustLevel, Threshold: 60, Comparator: ">="},
					{Type: FearLevel, Threshold: 40, Comparator: "<="},
					{Type: Event, EventName: "combat"},
				},
			},
			state: &NPCRuntimeState{
				Emotion: EmotionState{Trust: 70, Fear: 30, Stress: 50},
			},
			context: RevealContext{
				RecentEvents: []string{"combat", "friendly_chat"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldReveal(&tt.trait, tt.state, tt.context)
			if got != tt.want {
				t.Errorf("shouldReveal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test parseTraitStatus function
func TestParseTraitStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   TraitStatus
	}{
		{"Hidden status", "hidden", Hidden},
		{"Hinting status", "hinting", Hinting},
		{"Revealed status", "revealed", Revealed},
		{"Invalid status defaults to Hidden", "invalid", Hidden},
		{"Empty status defaults to Hidden", "", Hidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTraitStatus(tt.status)
			if got != tt.want {
				t.Errorf("parseTraitStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// Test NPCRuntimeState.RevealTrait method
func TestNPCRuntimeState_RevealTrait(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		traitID        string
		expectedStatus string
	}{
		{
			// Story 8.1: Now progresses to HintPhase1 instead of Hinting
			name:           "Hidden to Hinting",
			initialStatus:  "hidden",
			traitID:        "trait_paranoid",
			expectedStatus: "hint_phase_1",
		},
		{
			// Story 8.1: Hinting (hint_phase_1) now progresses to hint_phase_2
			name:           "Hinting to Revealed",
			initialStatus:  "hinting",
			traitID:        "trait_paranoid",
			expectedStatus: "hint_phase_2",
		},
		{
			name:           "Revealed stays Revealed",
			initialStatus:  "revealed",
			traitID:        "trait_paranoid",
			expectedStatus: "revealed",
		},
		{
			// Story 8.1: Uninitialized progresses to HintPhase1
			name:           "Uninitialized to Hinting",
			initialStatus:  "",
			traitID:        "trait_new",
			expectedStatus: "hint_phase_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewNPCRuntimeState()
			if tt.initialStatus != "" {
				state.TraitStates[tt.traitID] = tt.initialStatus
			}

			state.RevealTrait(tt.traitID)

			got := state.TraitStates[tt.traitID]
			if got != tt.expectedStatus {
				t.Errorf("After RevealTrait(), status = %q, want %q", got, tt.expectedStatus)
			}
		})
	}
}

// Test NPCRuntimeState.GetTraitStatus method
func TestNPCRuntimeState_GetTraitStatus(t *testing.T) {
	state := NewNPCRuntimeState()
	state.TraitStates["trait_paranoid"] = "hinting"
	state.TraitStates["trait_brave"] = "revealed"

	tests := []struct {
		name    string
		traitID string
		want    TraitStatus
	}{
		{"Get Hinting status", "trait_paranoid", Hinting},
		{"Get Revealed status", "trait_brave", Revealed},
		{"Get non-existent trait defaults to Hidden", "trait_new", Hidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.GetTraitStatus(tt.traitID)
			if got != tt.want {
				t.Errorf("GetTraitStatus(%q) = %v, want %v", tt.traitID, got, tt.want)
			}
		})
	}
}

// Test NPCRuntimeState.InitializeTraitStatus method
func TestNPCRuntimeState_InitializeTraitStatus(t *testing.T) {
	state := NewNPCRuntimeState()
	state.TraitStates["existing_trait"] = "hinting"

	// Initialize new trait
	state.InitializeTraitStatus("new_trait")
	if status, exists := state.TraitStates["new_trait"]; !exists || status != "hidden" {
		t.Errorf("InitializeTraitStatus should set new trait to hidden, got %q", status)
	}

	// Should not overwrite existing trait
	state.InitializeTraitStatus("existing_trait")
	if status := state.TraitStates["existing_trait"]; status != "hinting" {
		t.Errorf("InitializeTraitStatus should not overwrite existing trait, got %q", status)
	}
}

// Test NPCRuntimeState.GetHintingTraits method
// Story 8.1: GetHintingTraits now includes both HintPhase1 and HintPhase2
func TestNPCRuntimeState_GetHintingTraits(t *testing.T) {
	state := NewNPCRuntimeState()
	// Story 8.1: Use hint_phase_1 instead of legacy "hinting"
	state.TraitStates["trait_paranoid"] = "hint_phase_1"
	state.TraitStates["trait_cautious"] = "hint_phase_2"
	state.TraitStates["trait_brave"] = "hidden"
	state.TraitStates["trait_skilled"] = "revealed"

	profile := &NPCProfile{
		Traits: []Trait{
			{ID: "trait_paranoid", Content: "極度偏執", RevealTier: 2},
			{ID: "trait_cautious", Content: "小心謹慎", RevealTier: 2},
			{ID: "trait_brave", Content: "勇敢無畏", RevealTier: 1},
			{ID: "trait_skilled", Content: "技術熟練", RevealTier: 1},
		},
	}

	hintingTraits := state.GetHintingTraits(profile)

	// Story 8.1: Should now return both Phase1 and Phase2 traits
	if len(hintingTraits) != 2 {
		t.Fatalf("Expected 2 hinting traits (Phase1 + Phase2), got %d", len(hintingTraits))
	}

	// Check that we have both phases represented
	foundPhase1 := false
	foundPhase2 := false
	for _, trait := range hintingTraits {
		if trait.ID == "trait_paranoid" && trait.Status == HintPhase1 {
			foundPhase1 = true
		}
		if trait.ID == "trait_cautious" && trait.Status == HintPhase2 {
			foundPhase2 = true
		}
	}

	if !foundPhase1 {
		t.Error("Expected to find trait_paranoid in HintPhase1")
	}
	if !foundPhase2 {
		t.Error("Expected to find trait_cautious in HintPhase2")
	}
}

// Test NPCRuntimeState.GetRevealedTraits method
func TestNPCRuntimeState_GetRevealedTraits(t *testing.T) {
	state := NewNPCRuntimeState()
	state.TraitStates["trait_paranoid"] = "hinting"
	state.TraitStates["trait_brave"] = "revealed"
	state.TraitStates["trait_skilled"] = "revealed"

	profile := &NPCProfile{
		Traits: []Trait{
			{ID: "trait_paranoid", Content: "極度偏執", RevealTier: 2},
			{ID: "trait_brave", Content: "勇敢無畏", RevealTier: 1},
			{ID: "trait_skilled", Content: "技術熟練", RevealTier: 1},
		},
	}

	revealedTraits := state.GetRevealedTraits(profile)

	if len(revealedTraits) != 2 {
		t.Fatalf("Expected 2 revealed traits, got %d", len(revealedTraits))
	}

	// Check that both revealed traits are in the result
	foundBrave := false
	foundSkilled := false
	for _, trait := range revealedTraits {
		if trait.ID == "trait_brave" {
			foundBrave = true
		}
		if trait.ID == "trait_skilled" {
			foundSkilled = true
		}
		if trait.Status != Revealed {
			t.Errorf("Expected status to be Revealed, got %v for trait %s", trait.Status, trait.ID)
		}
	}

	if !foundBrave || !foundSkilled {
		t.Errorf("Expected to find both trait_brave and trait_skilled in revealed traits")
	}
}

// Test CheckAndRevealTraits function
func TestCheckAndRevealTraits(t *testing.T) {
	state := NewNPCRuntimeState()
	state.Emotion = EmotionState{
		Trust:  25,
		Fear:   60,
		Stress: 70,
	}

	traits := []TraitFull{
		{
			ID: "trait_paranoid",
			Triggers: []TraitTrigger{
				{Type: TrustLevel, Threshold: 30, Comparator: "<="},
				{Type: InteractionCount, Threshold: 5, Comparator: ">="},
			},
			Status: Hidden,
		},
		{
			ID: "trait_fearful",
			Triggers: []TraitTrigger{
				{Type: FearLevel, Threshold: 50, Comparator: ">="},
			},
			Status: Hidden,
		},
		{
			ID: "trait_calm",
			Triggers: []TraitTrigger{
				{Type: StressLevel, Threshold: 30, Comparator: "<="},
			},
			Status: Hidden,
		},
	}

	context := RevealContext{
		InteractionCount: 6,
	}

	progressedTraits := CheckAndRevealTraits(traits, state, context)

	// Should progress trait_paranoid and trait_fearful (both meet conditions)
	// trait_calm should not progress (stress is 70, needs <= 30)
	if len(progressedTraits) != 2 {
		t.Errorf("Expected 2 traits to progress, got %d", len(progressedTraits))
	}

	// Check that the correct traits progressed
	foundParanoid := false
	foundFearful := false
	for _, id := range progressedTraits {
		if id == "trait_paranoid" {
			foundParanoid = true
		}
		if id == "trait_fearful" {
			foundFearful = true
		}
	}

	if !foundParanoid || !foundFearful {
		t.Errorf("Expected trait_paranoid and trait_fearful to progress, got %v", progressedTraits)
	}

	// Verify the status was updated
	if state.GetTraitStatus("trait_paranoid") != Hinting {
		t.Errorf("trait_paranoid should be Hinting, got %v", state.GetTraitStatus("trait_paranoid"))
	}

	if state.GetTraitStatus("trait_fearful") != Hinting {
		t.Errorf("trait_fearful should be Hinting, got %v", state.GetTraitStatus("trait_fearful"))
	}

	if state.GetTraitStatus("trait_calm") != Hidden {
		t.Errorf("trait_calm should still be Hidden, got %v", state.GetTraitStatus("trait_calm"))
	}
}

// Test CheckAndRevealTraits does not progress already revealed traits
func TestCheckAndRevealTraits_DoesNotProgressRevealed(t *testing.T) {
	state := NewNPCRuntimeState()
	state.Emotion = EmotionState{Trust: 25, Fear: 60, Stress: 70}
	state.TraitStates["trait_already_revealed"] = "revealed"

	traits := []TraitFull{
		{
			ID: "trait_already_revealed",
			Triggers: []TraitTrigger{
				{Type: TrustLevel, Threshold: 30, Comparator: "<="},
			},
			Status: Revealed,
		},
	}

	context := RevealContext{}

	progressedTraits := CheckAndRevealTraits(traits, state, context)

	// Should not progress already revealed trait
	if len(progressedTraits) != 0 {
		t.Errorf("Expected no traits to progress, got %d", len(progressedTraits))
	}

	// Status should still be revealed
	if state.GetTraitStatus("trait_already_revealed") != Revealed {
		t.Errorf("trait_already_revealed should remain Revealed")
	}
}

// Test full revelation flow: Hidden -> Hinting -> Revealed
func TestFullRevelationFlow(t *testing.T) {
	state := NewNPCRuntimeState()
	state.Emotion = EmotionState{Trust: 20, Fear: 50, Stress: 60}

	trait := TraitFull{
		ID: "trait_paranoid",
		Triggers: []TraitTrigger{
			{Type: TrustLevel, Threshold: 30, Comparator: "<="},
		},
		Status: Hidden,
		Hints:  []string{"眼神充滿懷疑"},
	}

	context := RevealContext{
		InteractionCount: 1,
	}

	// Initially hidden
	if state.GetTraitStatus(trait.ID) != Hidden {
		t.Errorf("Initial status should be Hidden")
	}

	// Story 8.1: Now requires 3 progressions to reach Revealed (4-phase system)
	// First progression: Hidden -> HintPhase1
	traits := []TraitFull{trait}
	CheckAndRevealTraits(traits, state, context)

	if state.GetTraitStatus(trait.ID) != HintPhase1 {
		t.Errorf("After first progression, status should be HintPhase1, got %v", state.GetTraitStatus(trait.ID))
	}

	// Second progression: HintPhase1 -> HintPhase2
	CheckAndRevealTraits(traits, state, context)

	if state.GetTraitStatus(trait.ID) != HintPhase2 {
		t.Errorf("After second progression, status should be HintPhase2, got %v", state.GetTraitStatus(trait.ID))
	}

	// Third progression: HintPhase2 -> Revealed
	CheckAndRevealTraits(traits, state, context)

	if state.GetTraitStatus(trait.ID) != Revealed {
		t.Errorf("After third progression, status should be Revealed, got %v", state.GetTraitStatus(trait.ID))
	}

	// Fourth progression: should remain Revealed
	CheckAndRevealTraits(traits, state, context)

	if state.GetTraitStatus(trait.ID) != Revealed {
		t.Errorf("After fourth progression, status should still be Revealed, got %v", state.GetTraitStatus(trait.ID))
	}
}
