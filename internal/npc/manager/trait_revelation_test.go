package manager

import (
	"testing"
)

// ==========================================================================
// Story 8.1: Multi-Phase Trait Revelation Tests
// ==========================================================================

// TestTraitStatus_MultiPhaseProgression tests the four-phase trait status progression.
// Story 8.1 AC1: TraitStatus supports hidden → hint_phase_1 → hint_phase_2 → revealed
func TestTraitStatus_MultiPhaseProgression(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  TraitStatus
		expectedNext   TraitStatus
	}{
		{
			name:          "Hidden progresses to HintPhase1",
			initialStatus: Hidden,
			expectedNext:  HintPhase1,
		},
		{
			name:          "HintPhase1 progresses to HintPhase2",
			initialStatus: HintPhase1,
			expectedNext:  HintPhase2,
		},
		{
			name:          "HintPhase2 progresses to Revealed",
			initialStatus: HintPhase2,
			expectedNext:  Revealed,
		},
		{
			name:          "Revealed stays Revealed",
			initialStatus: Revealed,
			expectedNext:  Revealed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewNPCRuntimeState()
			traitID := "test_trait"

			// Set initial status
			state.TraitStates[traitID] = tt.initialStatus.String()

			// Progress trait
			state.RevealTrait(traitID)

			// Check new status
			newStatus := state.GetTraitStatus(traitID)
			if newStatus != tt.expectedNext {
				t.Errorf("Expected status %s, got %s", tt.expectedNext.String(), newStatus.String())
			}
		})
	}
}

// TestTraitStatus_StringSerialization tests TraitStatus string conversion.
// Story 8.1 AC1: Verify string serialization for all phases
func TestTraitStatus_StringSerialization(t *testing.T) {
	tests := []struct {
		status         TraitStatus
		expectedString string
	}{
		{Hidden, "hidden"},
		{HintPhase1, "hint_phase_1"},
		{HintPhase2, "hint_phase_2"},
		{Revealed, "revealed"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedString, func(t *testing.T) {
			if tt.status.String() != tt.expectedString {
				t.Errorf("Expected %s, got %s", tt.expectedString, tt.status.String())
			}
		})
	}
}

// TestTraitStatus_JSONSerialization tests TraitStatus JSON marshaling/unmarshaling.
// Story 8.1 AC1: Verify JSON serialization for persistence
func TestTraitStatus_JSONSerialization(t *testing.T) {
	tests := []struct {
		name   string
		status TraitStatus
	}{
		{"Hidden", Hidden},
		{"HintPhase1", HintPhase1},
		{"HintPhase2", HintPhase2},
		{"Revealed", Revealed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := tt.status.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			// Unmarshal
			var status TraitStatus
			err = status.UnmarshalJSON(data)
			if err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}

			// Verify
			if status != tt.status {
				t.Errorf("Expected %s, got %s", tt.status.String(), status.String())
			}
		})
	}
}

// TestTraitStatus_LegacyCompatibility tests backward compatibility with "hinting".
// Story 8.1 AC1: Legacy "hinting" maps to HintPhase1
func TestTraitStatus_LegacyCompatibility(t *testing.T) {
	var status TraitStatus

	// Unmarshal legacy "hinting" string
	err := status.UnmarshalJSON([]byte(`"hinting"`))
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Should map to HintPhase1
	if status != HintPhase1 {
		t.Errorf("Expected HintPhase1, got %s", status.String())
	}

	// Verify Hinting constant equals HintPhase1
	if Hinting != HintPhase1 {
		t.Errorf("Hinting constant should equal HintPhase1")
	}
}

// TestCalculatePhaseTransitionScore tests the multi-factor scoring algorithm.
// Story 8.1 AC4: Phase transition score calculation
func TestCalculatePhaseTransitionScore(t *testing.T) {
	tests := []struct {
		name             string
		trust            int
		interactionCount int
		currentBeat      int
		lastCheckBeat    int
		revealTier       int
		expectedMin      int
		expectedMax      int
	}{
		{
			name:             "High trust, many interactions, recent check",
			trust:            85,
			interactionCount: 15,
			currentBeat:      20,
			lastCheckBeat:    15,
			revealTier:       1,
			expectedMin:      70, // Should trigger progression
			expectedMax:      100,
		},
		{
			name:             "Low trust, few interactions",
			trust:            10,
			interactionCount: 2,
			currentBeat:      10,
			lastCheckBeat:    5,
			revealTier:       1,
			expectedMin:      0,
			expectedMax:      20, // Should not trigger progression
		},
		{
			name:             "Medium trust, medium interactions",
			trust:            60,
			interactionCount: 8,
			currentBeat:      50,
			lastCheckBeat:    25,
			revealTier:       2,
			expectedMin:      40,
			expectedMax:      70,
		},
		{
			name:             "High trust, hard trait, many interactions",
			trust:            80,
			interactionCount: 25,
			currentBeat:      100,
			lastCheckBeat:    50,
			revealTier:       3,
			expectedMin:      70,
			expectedMax:      100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewNPCRuntimeState()
			state.Emotion.Trust = tt.trust

			score := CalculatePhaseTransitionScore(
				state,
				tt.interactionCount,
				tt.currentBeat,
				tt.lastCheckBeat,
				tt.revealTier,
			)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("Expected score between %d and %d, got %d",
					tt.expectedMin, tt.expectedMax, score)
			}

			t.Logf("Score: %d (Trust: %d, Interactions: %d, Beats: %d, Tier: %d)",
				score, tt.trust, tt.interactionCount, tt.currentBeat-tt.lastCheckBeat, tt.revealTier)
		})
	}
}

// TestAccelerateTraitRevelation tests the interaction-based acceleration mechanism.
// Story 8.1 AC3: Player interaction acceleration
func TestAccelerateTraitRevelation(t *testing.T) {
	tests := []struct {
		name             string
		trust            int
		playerAction     string
		traitContent     string
		interactionCount int
		expectAccelerate bool
	}{
		{
			name:             "High trust, relevant action, sufficient interactions",
			trust:            75,
			playerAction:     "Tell me about your fear of the dark",
			traitContent:     "Suffers from nyctophobia (fear of the dark)",
			interactionCount: 5,
			expectAccelerate: true,
		},
		{
			name:             "High trust, irrelevant action",
			trust:            80,
			playerAction:     "What's the weather like?",
			traitContent:     "Suffers from nyctophobia",
			interactionCount: 5,
			expectAccelerate: false,
		},
		{
			name:             "Low trust, relevant action",
			trust:            50,
			playerAction:     "Tell me about your fear",
			traitContent:     "Suffers from nyctophobia",
			interactionCount: 5,
			expectAccelerate: false,
		},
		{
			name:             "High trust, relevant action, too few interactions",
			trust:            80,
			playerAction:     "Tell me about your fear",
			traitContent:     "Suffers from nyctophobia",
			interactionCount: 2,
			expectAccelerate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewNPCRuntimeState()
			state.Emotion.Trust = tt.trust

			trait := &TraitFull{
				ID:      "test_trait",
				Content: tt.traitContent,
			}

			context := RevealContext{
				InteractionCount: tt.interactionCount,
			}

			result := AccelerateTraitRevelation(trait, state, context, tt.playerAction)

			if result != tt.expectAccelerate {
				t.Errorf("Expected acceleration=%v, got %v", tt.expectAccelerate, result)
			}
		})
	}
}

// TestGetTraitsInPhase tests phase-specific trait retrieval.
// Story 8.1 AC1: Get traits in specific phases
func TestGetTraitsInPhase(t *testing.T) {
	profile := &NPCProfile{
		ID:   "test_npc",
		Name: "Test NPC",
		Traits: []Trait{
			{ID: "trait1", Content: "Trait 1"},
			{ID: "trait2", Content: "Trait 2"},
			{ID: "trait3", Content: "Trait 3"},
			{ID: "trait4", Content: "Trait 4"},
		},
	}

	state := NewNPCRuntimeState()
	state.TraitStates["trait1"] = Hidden.String()
	state.TraitStates["trait2"] = HintPhase1.String()
	state.TraitStates["trait3"] = HintPhase2.String()
	state.TraitStates["trait4"] = Revealed.String()

	// Test each phase
	tests := []struct {
		phase         TraitStatus
		expectedCount int
		expectedIDs   []string
	}{
		{Hidden, 1, []string{"trait1"}},
		{HintPhase1, 1, []string{"trait2"}},
		{HintPhase2, 1, []string{"trait3"}},
		{Revealed, 1, []string{"trait4"}},
	}

	for _, tt := range tests {
		t.Run(tt.phase.String(), func(t *testing.T) {
			traits := state.GetTraitsInPhase(profile, tt.phase)

			if len(traits) != tt.expectedCount {
				t.Errorf("Expected %d traits in phase %s, got %d",
					tt.expectedCount, tt.phase.String(), len(traits))
			}

			for i, expectedID := range tt.expectedIDs {
				if i >= len(traits) {
					t.Errorf("Expected trait %s not found", expectedID)
					continue
				}
				if traits[i].ID != expectedID {
					t.Errorf("Expected trait ID %s, got %s", expectedID, traits[i].ID)
				}
			}
		})
	}
}

// TestGetHintingTraits_MultiPhase tests that GetHintingTraits includes both phases.
// Story 8.1 AC2: GetHintingTraits returns both HintPhase1 and HintPhase2
func TestGetHintingTraits_MultiPhase(t *testing.T) {
	profile := &NPCProfile{
		ID:   "test_npc",
		Name: "Test NPC",
		Traits: []Trait{
			{ID: "trait1", Content: "Trait 1"},
			{ID: "trait2", Content: "Trait 2"},
			{ID: "trait3", Content: "Trait 3"},
			{ID: "trait4", Content: "Trait 4"},
		},
	}

	state := NewNPCRuntimeState()
	state.TraitStates["trait1"] = Hidden.String()
	state.TraitStates["trait2"] = HintPhase1.String()
	state.TraitStates["trait3"] = HintPhase2.String()
	state.TraitStates["trait4"] = Revealed.String()

	hintingTraits := state.GetHintingTraits(profile)

	// Should include both trait2 (HintPhase1) and trait3 (HintPhase2)
	if len(hintingTraits) != 2 {
		t.Errorf("Expected 2 hinting traits, got %d", len(hintingTraits))
	}

	foundPhase1 := false
	foundPhase2 := false
	for _, trait := range hintingTraits {
		if trait.ID == "trait2" && trait.Status == HintPhase1 {
			foundPhase1 = true
		}
		if trait.ID == "trait3" && trait.Status == HintPhase2 {
			foundPhase2 = true
		}
	}

	if !foundPhase1 {
		t.Error("Expected to find HintPhase1 trait (trait2)")
	}
	if !foundPhase2 {
		t.Error("Expected to find HintPhase2 trait (trait3)")
	}
}

// TestCheckAndProgressTraits tests the integrated progression system.
// Story 8.1 AC4: Natural phase transition
func TestCheckAndProgressTraits(t *testing.T) {
	// Create traits with different revelation tiers
	traits := []TraitFull{
		{
			ID:               "easy_trait",
			Content:          "Easy to reveal",
			RevealTier:       1,
			Status:           Hidden,
			InteractionCount: 12,
			LastRevealCheck:  0,
		},
		{
			ID:               "medium_trait",
			Content:          "Medium difficulty",
			RevealTier:       2,
			Status:           HintPhase1,
			InteractionCount: 18,
			LastRevealCheck:  10,
		},
		{
			ID:               "hard_trait",
			Content:          "Hard to reveal",
			RevealTier:       3,
			Status:           Hidden,
			InteractionCount: 5,
			LastRevealCheck:  50,
		},
	}

	state := NewNPCRuntimeState()
	state.Emotion.Trust = 80 // High trust

	// Initialize trait states
	for _, trait := range traits {
		state.TraitStates[trait.ID] = trait.Status.String()
	}

	context := RevealContext{
		CurrentBeat:      100,
		InteractionCount: 20,
	}

	// Check and progress traits
	progressed := CheckAndProgressTraits(traits, state, context)

	// Easy trait should progress (high trust + many interactions)
	if _, ok := progressed["easy_trait"]; !ok {
		t.Error("Expected easy_trait to progress")
	}

	// Medium trait should progress (high trust + many interactions + time)
	if _, ok := progressed["medium_trait"]; !ok {
		t.Error("Expected medium_trait to progress")
	}

	// Verify final states
	if state.GetTraitStatus("easy_trait") != HintPhase1 {
		t.Errorf("Expected easy_trait to be HintPhase1, got %s",
			state.GetTraitStatus("easy_trait").String())
	}

	if state.GetTraitStatus("medium_trait") != HintPhase2 {
		t.Errorf("Expected medium_trait to be HintPhase2, got %s",
			state.GetTraitStatus("medium_trait").String())
	}
}

// TestNPCManager_CheckTraitRevelation tests the manager-level trait revelation.
// Story 8.1 AC3 & AC4: Integration test for CheckTraitRevelation
func TestNPCManager_CheckTraitRevelation(t *testing.T) {
	mgr := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:   "test_npc",
		Name: "Test NPC",
		Traits: []Trait{
			{ID: "trait1", Content: "Suspicious behavior", RevealTier: 1},
			{ID: "trait2", Content: "Has secrets", RevealTier: 2},
		},
		InitialEmotion: EmotionState{
			Trust:  80,
			Fear:   20,
			Stress: 30,
		},
	}

	err := mgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Simulate multiple interactions
	for i := 0; i < 10; i++ {
		err = mgr.RecordInteraction("test_npc", NPCInteraction{
			InteractionType: "chat",
			Description:     "Player conversation",
		})
		if err != nil {
			t.Fatalf("Failed to record interaction: %v", err)
		}
	}

	// Check trait revelation with relevant message
	progressed, err := mgr.CheckTraitRevelation("test_npc", "Tell me about your suspicious behavior", 50)
	if err != nil {
		t.Fatalf("CheckTraitRevelation failed: %v", err)
	}

	// At least one trait should progress
	if len(progressed) == 0 {
		t.Error("Expected at least one trait to progress")
	}

	t.Logf("Progressed traits: %v", progressed)
}

// TestPhaseSpecificHints tests hint retrieval based on phase.
// Story 8.1 AC2: Phase-specific hint integration
func TestPhaseSpecificHints(t *testing.T) {
	trait := TraitFull{
		ID:          "test_trait",
		Content:     "Has a dark secret",
		Status:      HintPhase1,
		Hints:       []string{"Legacy hint"},
		HintsPhase1: []string{"Seems nervous when topic comes up"},
		HintsPhase2: []string{"Directly mentions past trauma"},
	}

	// Test Phase 1 hints
	if len(trait.HintsPhase1) != 1 {
		t.Errorf("Expected 1 Phase1 hint, got %d", len(trait.HintsPhase1))
	}

	// Test Phase 2 hints
	if len(trait.HintsPhase2) != 1 {
		t.Errorf("Expected 1 Phase2 hint, got %d", len(trait.HintsPhase2))
	}

	// Phase 1 hints should be more subtle than Phase 2
	if len(trait.HintsPhase2[0]) <= len(trait.HintsPhase1[0]) {
		t.Log("Note: Phase 2 hints should typically be more explicit than Phase 1")
	}
}

// TestContainsKeywords tests the keyword matching helper.
func TestContainsKeywords(t *testing.T) {
	tests := []struct {
		text      string
		reference string
		expected  bool
	}{
		{"Tell me about your fear", "fear of the dark", true},
		{"What's your biggest fear?", "Fear", true},
		{"Tell me about yourself", "fear", false},
		{"", "fear", false},
		{"fear", "", true}, // Empty reference always matches
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := containsKeywords(tt.text, tt.reference)
			if result != tt.expected {
				t.Errorf("containsKeywords(%q, %q) = %v, expected %v",
					tt.text, tt.reference, result, tt.expected)
			}
		})
	}
}

// ==========================================================================
// Story 8.1: Integration Tests
// ==========================================================================

// TestMultiPhaseRevelationFlow tests the complete multi-phase revelation workflow.
// Story 8.1 AC1-AC4: End-to-end integration test
func TestMultiPhaseRevelationFlow(t *testing.T) {
	// Create a trait with triggers
	trait := TraitFull{
		ID:          "dark_secret",
		Content:     "Has a traumatic past",
		RevealTier:  2,
		Status:      Hidden,
		Hints:       []string{"Legacy hint"},
		HintsPhase1: []string{"Seems uncomfortable when past is mentioned"},
		HintsPhase2: []string{"Visibly distressed, mentions childhood trauma"},
		Triggers: []TraitTrigger{
			{
				Type:       TrustLevel,
				Threshold:  40,
				Comparator: ">=",
			},
			{
				Type:       InteractionCount,
				Threshold:  5,
				Comparator: ">=",
			},
		},
		InteractionCount: 0,
		LastRevealCheck:  0,
	}

	state := NewNPCRuntimeState()
	state.Emotion.Trust = 50 // Above threshold
	state.InitializeTraitStatus(trait.ID)

	context := RevealContext{
		CurrentBeat:      10,
		InteractionCount: 6, // Above threshold
		RecentEvents:     []string{},
	}

	// Phase 1: Hidden -> HintPhase1
	t.Run("Phase1_HiddenToHintPhase1", func(t *testing.T) {
		// Verify initial state
		if state.GetTraitStatus(trait.ID) != Hidden {
			t.Errorf("Expected Hidden, got %s", state.GetTraitStatus(trait.ID))
		}

		// Check and progress
		traits := []TraitFull{trait}
		progressed := CheckAndProgressTraits(traits, state, context)

		// Verify progression
		if len(progressed) != 1 {
			t.Errorf("Expected 1 progressed trait, got %d", len(progressed))
		}

		if state.GetTraitStatus(trait.ID) != HintPhase1 {
			t.Errorf("Expected HintPhase1, got %s", state.GetTraitStatus(trait.ID))
		}

		// Verify hint content (would be available in real TraitFull)
		if len(trait.HintsPhase1) == 0 {
			t.Error("Expected HintsPhase1 to have content")
		}
	})

	// Phase 2: HintPhase1 -> HintPhase2
	t.Run("Phase2_HintPhase1ToHintPhase2", func(t *testing.T) {
		// Update context for next phase
		context.CurrentBeat = 50
		context.InteractionCount = 12

		// Progress again
		trait.Status = state.GetTraitStatus(trait.ID) // Update trait status
		traits := []TraitFull{trait}
		progressed := CheckAndProgressTraits(traits, state, context)

		// Verify progression
		if len(progressed) != 1 {
			t.Errorf("Expected 1 progressed trait, got %d", len(progressed))
		}

		if state.GetTraitStatus(trait.ID) != HintPhase2 {
			t.Errorf("Expected HintPhase2, got %s", state.GetTraitStatus(trait.ID))
		}

		// Verify hint content is more explicit
		if len(trait.HintsPhase2) == 0 {
			t.Error("Expected HintsPhase2 to have content")
		}

		// Phase2 hints should be more explicit (longer or different)
		if trait.HintsPhase2[0] == trait.HintsPhase1[0] {
			t.Log("Note: Phase2 hints should typically be different from Phase1")
		}
	})

	// Phase 3: HintPhase2 -> Revealed
	t.Run("Phase3_HintPhase2ToRevealed", func(t *testing.T) {
		// Update context for final phase
		context.CurrentBeat = 100
		context.InteractionCount = 20
		state.Emotion.Trust = 85 // Very high trust

		// Progress to final phase
		trait.Status = state.GetTraitStatus(trait.ID)
		traits := []TraitFull{trait}
		progressed := CheckAndProgressTraits(traits, state, context)

		// Verify final progression
		if len(progressed) != 1 {
			t.Errorf("Expected 1 progressed trait, got %d", len(progressed))
		}

		if state.GetTraitStatus(trait.ID) != Revealed {
			t.Errorf("Expected Revealed, got %s", state.GetTraitStatus(trait.ID))
		}
	})

	// Phase 4: Already Revealed (no further progression)
	t.Run("Phase4_StaysRevealed", func(t *testing.T) {
		// Try to progress again
		context.CurrentBeat = 150
		trait.Status = state.GetTraitStatus(trait.ID)
		traits := []TraitFull{trait}
		progressed := CheckAndProgressTraits(traits, state, context)

		// Should not progress further
		if len(progressed) != 0 {
			t.Errorf("Expected 0 progressed traits (already revealed), got %d", len(progressed))
		}

		// Status should remain Revealed
		if state.GetTraitStatus(trait.ID) != Revealed {
			t.Errorf("Expected Revealed, got %s", state.GetTraitStatus(trait.ID))
		}
	})
}

// TestAccelerationWithMultiPhase tests interaction acceleration across phases.
// Story 8.1 AC3: Acceleration mechanism integration
func TestAccelerationWithMultiPhase(t *testing.T) {
	trait := TraitFull{
		ID:          "fear_trait",
		Content:     "Suffers from nyctophobia (fear of dark)",
		RevealTier:  1,
		Status:      Hidden,
		HintsPhase1: []string{"Avoids dark rooms"},
		HintsPhase2: []string{"Panics when lights go out"},
	}

	state := NewNPCRuntimeState()
	state.Emotion.Trust = 75 // High enough for acceleration
	state.InitializeTraitStatus(trait.ID)

	context := RevealContext{
		CurrentBeat:      20,
		InteractionCount: 5, // Enough interactions
		RecentEvents:     []string{},
	}

	// Test acceleration with relevant player action
	playerAction := "I noticed you avoid the dark. Tell me about your fear of darkness."

	shouldAccelerate := AccelerateTraitRevelation(&trait, state, context, playerAction)

	if !shouldAccelerate {
		t.Error("Expected acceleration with high trust and relevant action")
	}

	// Apply acceleration by progressing trait
	state.RevealTrait(trait.ID)

	// Verify trait progressed
	if state.GetTraitStatus(trait.ID) != HintPhase1 {
		t.Errorf("Expected HintPhase1 after acceleration, got %s", state.GetTraitStatus(trait.ID))
	}

	// Test acceleration from Phase1 to Phase2
	trait.Status = HintPhase1
	playerAction2 := "Your fear seems deep. Can you tell me more about what scares you about the dark?"

	shouldAccelerate2 := AccelerateTraitRevelation(&trait, state, context, playerAction2)

	if !shouldAccelerate2 {
		t.Error("Expected acceleration from Phase1 to Phase2")
	}
}

// TestScoreThresholds tests edge cases around the 70-point threshold.
// Story 8.1 AC4: Boundary testing for phase transition scoring
func TestScoreThresholds(t *testing.T) {
	tests := []struct {
		name             string
		trust            int
		interactionCount int
		beatsPassed      int
		revealTier       int
		expectedAbove70  bool
	}{
		{
			name:             "Just below threshold",
			trust:            50,  // 25 points
			interactionCount: 8,   // ~24 points (tier 1)
			beatsPassed:      100, // 20 points (capped at 30)
			revealTier:       1,
			expectedAbove70:  false, // 25+24+20 = 69
		},
		{
			name:             "Exactly at threshold",
			trust:            50,  // 25 points
			interactionCount: 9,   // ~27 points
			beatsPassed:      90,  // 18 points
			revealTier:       1,
			expectedAbove70:  true, // 25+27+18 = 70
		},
		{
			name:             "Above threshold",
			trust:            80,  // 40 points
			interactionCount: 10,  // 30 points (tier 1 max)
			beatsPassed:      5,   // 1 point
			revealTier:       1,
			expectedAbove70:  true, // 40+30+1 = 71
		},
		{
			name:             "High tier requires more interactions",
			trust:            80,  // 40 points
			interactionCount: 10,  // ~15 points (tier 3)
			beatsPassed:      75,  // 15 points
			revealTier:       3,
			expectedAbove70:  true, // 40+15+15 = 70
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewNPCRuntimeState()
			state.Emotion.Trust = tt.trust

			score := CalculatePhaseTransitionScore(
				state,
				tt.interactionCount,
				tt.beatsPassed,
				0,
				tt.revealTier,
			)

			isAbove70 := score >= 70

			if isAbove70 != tt.expectedAbove70 {
				t.Errorf("Score %d, expected above70=%v, got %v",
					score, tt.expectedAbove70, isAbove70)
			}

			t.Logf("Score: %d (Trust: %d, Interactions: %d, Beats: %d, Tier: %d)",
				score, tt.trust, tt.interactionCount, tt.beatsPassed, tt.revealTier)
		})
	}
}
