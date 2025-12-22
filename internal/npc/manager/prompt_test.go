package manager

import (
	"strings"
	"testing"
)

// TestEmotionToDescription tests all emotion types and value ranges.
// This ensures the emotion description mapping is complete and correct for all 5 levels.
func TestEmotionToDescription(t *testing.T) {
	m := NewNPCManager(nil, nil)

	tests := []struct {
		name        string
		value       int
		emotionType string
		expected    string
	}{
		// Trust levels - all 5 levels + boundary cases
		{"Trust 90", 90, "trust", "非常信任，願意分享秘密"},
		{"Trust 70", 70, "trust", "相當信任，願意合作"},
		{"Trust 50", 50, "trust", "中等信任，保持觀望"},
		{"Trust 30", 30, "trust", "不太信任，有所保留"},
		{"Trust 10", 10, "trust", "完全不信任，可能敵對"},
		{"Trust 80 (boundary)", 80, "trust", "非常信任，願意分享秘密"},
		{"Trust 60 (boundary)", 60, "trust", "相當信任，願意合作"},
		{"Trust 40 (boundary)", 40, "trust", "中等信任，保持觀望"},
		{"Trust 20 (boundary)", 20, "trust", "不太信任，有所保留"},
		{"Trust 0 (min)", 0, "trust", "完全不信任，可能敵對"},
		{"Trust 100 (max)", 100, "trust", "非常信任，願意分享秘密"},

		// Fear levels - all 5 levels + boundary cases
		{"Fear 90", 90, "fear", "極度恐懼，可能逃跑或崩潰"},
		{"Fear 70", 70, "fear", "非常害怕，行動遲疑"},
		{"Fear 50", 50, "fear", "有些害怕，小心翼翼"},
		{"Fear 30", 30, "fear", "略微緊張"},
		{"Fear 10", 10, "fear", "冷靜沉著"},
		{"Fear 80 (boundary)", 80, "fear", "極度恐懼，可能逃跑或崩潰"},
		{"Fear 60 (boundary)", 60, "fear", "非常害怕，行動遲疑"},
		{"Fear 40 (boundary)", 40, "fear", "有些害怕，小心翼翼"},
		{"Fear 20 (boundary)", 20, "fear", "略微緊張"},
		{"Fear 0 (min)", 0, "fear", "冷靜沉著"},
		{"Fear 100 (max)", 100, "fear", "極度恐懼，可能逃跑或崩潰"},

		// Stress levels - all 5 levels + boundary cases
		{"Stress 90", 90, "stress", "瀕臨崩潰，可能失控"},
		{"Stress 70", 70, "stress", "壓力很大，判斷力下降"},
		{"Stress 50", 50, "stress", "有些焦慮"},
		{"Stress 30", 30, "stress", "輕微壓力"},
		{"Stress 10", 10, "stress", "心態平穩"},
		{"Stress 80 (boundary)", 80, "stress", "瀕臨崩潰，可能失控"},
		{"Stress 60 (boundary)", 60, "stress", "壓力很大，判斷力下降"},
		{"Stress 40 (boundary)", 40, "stress", "有些焦慮"},
		{"Stress 20 (boundary)", 20, "stress", "輕微壓力"},
		{"Stress 0 (min)", 0, "stress", "心態平穩"},
		{"Stress 100 (max)", 100, "stress", "瀕臨崩潰，可能失控"},

		// Invalid emotion type
		{"Invalid type", 50, "invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.emotionToDescription(tt.value, tt.emotionType)
			if result != tt.expected {
				t.Errorf("emotionToDescription(%d, %s) = %q, want %q",
					tt.value, tt.emotionType, result, tt.expected)
			}
		})
	}
}

// TestMentalStateToDescription tests all mental state descriptions.
func TestMentalStateToDescription(t *testing.T) {
	m := NewNPCManager(nil, nil)

	tests := []struct {
		name     string
		state    MentalState
		expected string
	}{
		{"Normal", Normal, "正常(Normal) - 思緒清晰、判斷準確、行為穩定"},
		{"Anxious", Anxious, "焦慮(Anxious) - 容易緊張、決策猶豫、需要安撫"},
		{"Corrupted", Corrupted, "崩潰(Corrupted) - 精神失常、行為不可預測、可能暴力或自殘"},
		{"Invalid state", MentalState(99), "未知狀態"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.mentalStateToDescription(tt.state)
			if result != tt.expected {
				t.Errorf("mentalStateToDescription(%v) = %q, want %q",
					tt.state, result, tt.expected)
			}
		})
	}
}

// TestBuildNPCPrompt tests the complete prompt building functionality.
// This is the main integration test that verifies all prompt sections are present.
func TestBuildNPCPrompt(t *testing.T) {
	m := NewNPCManager(nil, nil)

	// Create a realistic test NPC profile
	profile := &NPCProfile{
		ID:         "test-npc",
		Name:       "張醫生",
		Appearance: "40多歲男性，戴眼鏡，穿著白袍，神情疲憊",
		Traits: []Trait{
			{ID: "trait1", Content: "理性"},
			{ID: "trait2", Content: "謹慎"},
			{ID: "trait3", Content: "有醫學專業"},
			{ID: "secret", Content: "隱藏秘密"}, // This should NOT appear in prompt
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "專業醫學術語混合口語",
			Quirks:     []string{"常說「從醫學角度來說」", "我見過太多這樣的案例"},
		},
		InitialEmotion: EmotionState{
			Trust:  70,
			Fear:   65,
			Stress: 75,
		},
	}

	// Add NPC to manager
	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Mark some traits as revealed (but keep "secret" hidden)
	state := m.GetState("test-npc")
	state.TraitStates["trait1"] = "revealed"
	state.TraitStates["trait2"] = "revealed"
	state.TraitStates["trait3"] = "revealed"
	state.TraitStates["secret"] = "hidden" // Keep this hidden!
	state.MentalState = Anxious

	// Build the prompt
	prompt := m.BuildNPCPrompt("test-npc")

	// Verify all required sections are present
	requiredSections := []string{
		"## 角色：張醫生",
		"外觀：40多歲男性，戴眼鏡，穿著白袍，神情疲憊",
		"已知個性：理性、謹慎、有醫學專業",
		"當前情緒：",
		"- 信任程度：相當信任，願意合作",
		"- 恐懼程度：非常害怕，行動遲疑",
		"- 壓力程度：壓力很大，判斷力下降",
		"心理狀態：焦慮(Anxious)",
		"對話風格：",
		"- 用詞：專業醫學術語混合口語",
		"- 習慣：常說「從醫學角度來說」、我見過太多這樣的案例",
	}

	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Prompt missing required section: %q\nFull prompt:\n%s", section, prompt)
		}
	}

	// CRITICAL SECURITY CHECK: Verify hidden trait is NOT in prompt
	if strings.Contains(prompt, "隱藏秘密") {
		t.Errorf("SECURITY VIOLATION: Prompt contains hidden trait!\nPrompt:\n%s", prompt)
	}
}

// TestBuildNPCPrompt_NoHints tests prompt generation when there are no hints.
// This ensures the hints section is omitted when not applicable.
func TestBuildNPCPrompt_NoHints(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc2",
		Name:       "測試NPC",
		Appearance: "普通人",
		Traits:     []Trait{},
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	prompt := m.BuildNPCPrompt("npc2")

	// Should not contain hints section when no hints are present
	if strings.Contains(prompt, "行為暗示：") {
		t.Errorf("Prompt should not contain hints section when no hints present")
	}

	// Should not contain traits section when no traits are revealed
	if strings.Contains(prompt, "已知個性：") {
		t.Errorf("Prompt should not contain traits section when no traits revealed")
	}
}

// TestBuildNPCPrompt_NonExistentNPC tests error handling for non-existent NPCs.
func TestBuildNPCPrompt_NonExistentNPC(t *testing.T) {
	m := NewNPCManager(nil, nil)

	// Try to build prompt for non-existent NPC
	prompt := m.BuildNPCPrompt("non-existent")

	if prompt != "" {
		t.Errorf("Expected empty string for non-existent NPC, got: %q", prompt)
	}
}

// TestBuildNPCPrompt_AllEmotionLevels tests different emotion combinations.
// This ensures the emotion descriptions are correctly applied in real prompts.
func TestBuildNPCPrompt_AllEmotionLevels(t *testing.T) {
	m := NewNPCManager(nil, nil)

	emotionTests := []struct {
		name     string
		emotion  EmotionState
		expected []string
	}{
		{
			name:    "High trust, low fear",
			emotion: EmotionState{Trust: 85, Fear: 15, Stress: 25},
			expected: []string{
				"非常信任，願意分享秘密",
				"冷靜沉著",
				"輕微壓力",
			},
		},
		{
			name:    "Low trust, high fear",
			emotion: EmotionState{Trust: 15, Fear: 85, Stress: 90},
			expected: []string{
				"完全不信任，可能敵對",
				"極度恐懼，可能逃跑或崩潰",
				"瀕臨崩潰，可能失控",
			},
		},
		{
			name:    "Medium all",
			emotion: EmotionState{Trust: 50, Fear: 50, Stress: 50},
			expected: []string{
				"中等信任，保持觀望",
				"有些害怕，小心翼翼",
				"有些焦慮",
			},
		},
	}

	for i, tt := range emotionTests {
		profile := &NPCProfile{
			ID:         string(rune('a' + i)),
			Name:       tt.name,
			Appearance: "test",
			Traits:     []Trait{},
			DialogueStyle: DialogueStyle{
				Vocabulary: "test",
				Quirks:     []string{},
			},
			InitialEmotion: tt.emotion,
		}

		err := m.AddNPC(profile)
		if err != nil {
			t.Fatalf("Failed to add NPC: %v", err)
		}

		prompt := m.BuildNPCPrompt(profile.ID)

		for _, expectedText := range tt.expected {
			if !strings.Contains(prompt, expectedText) {
				t.Errorf("Test %q: prompt missing %q\nPrompt:\n%s",
					tt.name, expectedText, prompt)
			}
		}
	}
}

// TestGetRevealedTraits tests the trait filtering logic.
func TestGetRevealedTraits(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:   "npc-traits",
		Name: "Test",
		Traits: []Trait{
			{ID: "t1", Content: "特質1"},
			{ID: "t2", Content: "特質2"},
			{ID: "t3", Content: "特質3"},
		},
		Appearance: "test",
		DialogueStyle: DialogueStyle{
			Vocabulary: "test",
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	state := m.GetState("npc-traits")
	state.TraitStates["t1"] = "revealed"
	state.TraitStates["t2"] = "hidden"
	state.TraitStates["t3"] = "revealed"

	revealed := m.getRevealedTraits("npc-traits")

	// Should only contain t1 and t3
	if len(revealed) != 2 {
		t.Errorf("Expected 2 revealed traits, got %d", len(revealed))
	}

	// Check contents
	foundT1 := false
	foundT3 := false
	for _, trait := range revealed {
		if trait == "特質1" {
			foundT1 = true
		}
		if trait == "特質3" {
			foundT3 = true
		}
		if trait == "特質2" {
			t.Errorf("Hidden trait should not be in revealed list")
		}
	}

	if !foundT1 || !foundT3 {
		t.Errorf("Missing expected revealed traits. Got: %v", revealed)
	}
}

// TestGetHintingTraits tests the hinting trait retrieval.
// Note: This is a placeholder test for Story 1.8, as full hint logic is in Story 1.6.
func TestGetHintingTraits(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:   "npc-hints",
		Name: "Test",
		Traits: []Trait{
			{ID: "h1", Content: "暗示特質1"},
			{ID: "h2", Content: "暗示特質2"},
		},
		Appearance: "test",
		DialogueStyle: DialogueStyle{
			Vocabulary: "test",
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	state := m.GetState("npc-hints")
	state.TraitStates["h1"] = "hinting"
	state.TraitStates["h2"] = "hidden"

	// This is a placeholder implementation for Story 1.8
	hints := m.GetHintingTraits("npc-hints")

	// Story 1.8 returns empty slice as placeholder
	// Story 1.6 will implement full hint extraction
	if len(hints) != 0 {
		t.Logf("Note: GetHintingTraits returned %d hints (placeholder implementation)", len(hints))
	}
}

// TestBuildNPCPrompt_AllMentalStates tests all mental states in generated prompts.
func TestBuildNPCPrompt_AllMentalStates(t *testing.T) {
	m := NewNPCManager(nil, nil)

	mentalStates := []struct {
		state    MentalState
		expected string
	}{
		{Normal, "正常(Normal)"},
		{Anxious, "焦慮(Anxious)"},
		{Corrupted, "崩潰(Corrupted)"},
	}

	for i, tt := range mentalStates {
		profile := &NPCProfile{
			ID:         string(rune('x' + i)),
			Name:       "Test",
			Appearance: "test",
			Traits:     []Trait{},
			DialogueStyle: DialogueStyle{
				Vocabulary: "test",
			},
			InitialEmotion: DefaultEmotionState(),
		}

		err := m.AddNPC(profile)
		if err != nil {
			t.Fatalf("Failed to add NPC: %v", err)
		}

		state := m.GetState(profile.ID)
		state.MentalState = tt.state

		prompt := m.BuildNPCPrompt(profile.ID)

		if !strings.Contains(prompt, tt.expected) {
			t.Errorf("Prompt for mental state %v missing %q\nPrompt:\n%s",
				tt.state, tt.expected, prompt)
		}
	}
}

// TestBuildNPCPrompt_Example demonstrates the full prompt output.
// This test logs the complete generated prompt for visual inspection.
func TestBuildNPCPrompt_Example(t *testing.T) {
	m := NewNPCManager(nil, nil)

	// Create a realistic NPC
	profile := &NPCProfile{
		ID:         "doctor-zhang",
		Name:       "張醫生",
		Appearance: "40多歲男性，戴眼鏡，穿著白袍，神情疲憊",
		Traits: []Trait{
			{ID: "trait1", Content: "理性"},
			{ID: "trait2", Content: "謹慎"},
			{ID: "trait3", Content: "有醫學專業"},
			{ID: "secret", Content: "曾經誤診導致病人死亡"}, // This should NOT appear
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "專業醫學術語混合口語",
			Quirks:     []string{"常說「從醫學角度來說」", "「我見過太多這樣的案例」"},
		},
		InitialEmotion: EmotionState{
			Trust:  70,
			Fear:   65,
			Stress: 75,
		},
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Reveal some traits
	state := m.GetState("doctor-zhang")
	state.TraitStates["trait1"] = "revealed"
	state.TraitStates["trait2"] = "revealed"
	state.TraitStates["trait3"] = "revealed"
	state.TraitStates["secret"] = "hidden" // Keep secret hidden!
	state.MentalState = Anxious

	// Build prompt
	prompt := m.BuildNPCPrompt("doctor-zhang")

	// Log the full prompt for demonstration
	t.Logf("\n=== Generated NPC Prompt ===\n%s\n=== End of Prompt ===", prompt)

	// Verify secret is not leaked
	if strings.Contains(prompt, "誤診") || strings.Contains(prompt, "病人死亡") {
		t.Error("SECURITY VIOLATION: Hidden secret leaked to prompt!")
	} else {
		t.Log("✓ Security check passed: Hidden secret not leaked")
	}
}

// ==========================================================================
// Story 8.1: Phase-Specific Hints Tests
// ==========================================================================

// TestGetPhaseSpecificHints_Phase1Only tests HintPhase1 hint retrieval.
// Story 8.1 AC2: Phase-specific hint integration
func TestGetPhaseSpecificHints_Phase1Only(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc-phase1",
		Name:       "測試NPC",
		Appearance: "普通人",
		Traits: []Trait{
			{ID: "trait1", Content: "隱藏特質A"},
			{ID: "trait2", Content: "隱藏特質B"},
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Set trait1 to HintPhase1
	state := m.GetState("npc-phase1")
	state.TraitStates["trait1"] = HintPhase1.String()
	state.TraitStates["trait2"] = Hidden.String()

	// Build prompt
	prompt := m.BuildNPCPrompt("npc-phase1")

	// Since traits don't have actual HintsPhase1 data in basic Trait structure,
	// this tests that the getPhaseSpecificHints function is called
	// The actual hint content would come from TraitFull structure in real usage

	// This test verifies the code path is executed without error
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}
}

// TestGetPhaseSpecificHints_Phase2Only tests HintPhase2 hint retrieval.
// Story 8.1 AC2: Phase-specific hint integration - more explicit hints
func TestGetPhaseSpecificHints_Phase2Only(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc-phase2",
		Name:       "測試NPC",
		Appearance: "普通人",
		Traits: []Trait{
			{ID: "trait1", Content: "隱藏特質A"},
			{ID: "trait2", Content: "隱藏特質B"},
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Set trait1 to HintPhase2
	state := m.GetState("npc-phase2")
	state.TraitStates["trait1"] = HintPhase2.String()
	state.TraitStates["trait2"] = Hidden.String()

	// Build prompt
	prompt := m.BuildNPCPrompt("npc-phase2")

	// This test verifies the HintPhase2 code path is executed
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}
}

// TestGetPhaseSpecificHints_MixedPhases tests both Phase1 and Phase2 hints together.
// Story 8.1 AC2: Multiple traits at different phases should combine hints
func TestGetPhaseSpecificHints_MixedPhases(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc-mixed",
		Name:       "測試NPC",
		Appearance: "普通人",
		Traits: []Trait{
			{ID: "trait1", Content: "特質A"},
			{ID: "trait2", Content: "特質B"},
			{ID: "trait3", Content: "特質C"},
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Set different phases for different traits
	state := m.GetState("npc-mixed")
	state.TraitStates["trait1"] = HintPhase1.String() // Phase 1
	state.TraitStates["trait2"] = HintPhase2.String() // Phase 2
	state.TraitStates["trait3"] = Hidden.String()     // Hidden (no hints)

	// Build prompt
	prompt := m.BuildNPCPrompt("npc-mixed")

	// Verify prompt was generated
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Both phase hints should be processed (code path coverage)
	// In real usage with TraitFull, actual hint content would be included
}

// TestGetPhaseSpecificHints_NoHintingTraits tests that no hints section appears
// when all traits are hidden or revealed.
// Story 8.1 AC2: Hints should only appear for hinting phases
func TestGetPhaseSpecificHints_NoHintingTraits(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc-no-hints",
		Name:       "測試NPC",
		Appearance: "普通人",
		Traits: []Trait{
			{ID: "trait1", Content: "特質A"},
			{ID: "trait2", Content: "特質B"},
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Set all traits to non-hinting states
	state := m.GetState("npc-no-hints")
	state.TraitStates["trait1"] = Hidden.String()
	state.TraitStates["trait2"] = Revealed.String()

	// Build prompt
	prompt := m.BuildNPCPrompt("npc-no-hints")

	// Should NOT contain hints section
	if strings.Contains(prompt, "行為暗示：") {
		t.Error("Prompt should not contain hints section when no traits are in hinting phase")
	}
}

// TestGetPhaseSpecificHints_AllPhases tests prompt with traits in all phases.
// Story 8.1 AC2: Comprehensive phase coverage
func TestGetPhaseSpecificHints_AllPhases(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc-all-phases",
		Name:       "測試NPC",
		Appearance: "普通人",
		Traits: []Trait{
			{ID: "trait1", Content: "特質A"},
			{ID: "trait2", Content: "特質B"},
			{ID: "trait3", Content: "特質C"},
			{ID: "trait4", Content: "特質D"},
		},
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Set traits to all different phases
	state := m.GetState("npc-all-phases")
	state.TraitStates["trait1"] = Hidden.String()
	state.TraitStates["trait2"] = HintPhase1.String()
	state.TraitStates["trait3"] = HintPhase2.String()
	state.TraitStates["trait4"] = Revealed.String()

	// Build prompt
	prompt := m.BuildNPCPrompt("npc-all-phases")

	// Verify revealed trait appears in prompt
	if !strings.Contains(prompt, "特質D") {
		t.Error("Revealed trait should appear in prompt")
	}

	// Verify hidden trait does NOT appear
	if strings.Contains(prompt, "特質A") {
		t.Error("Hidden trait should NOT appear in prompt")
	}

	// Verify prompt structure is valid
	if !strings.Contains(prompt, "## 角色：") {
		t.Error("Prompt should contain character header")
	}
}

// TestGetPhaseSpecificHints_EmptyProfile tests edge case of NPC with no traits.
// Story 8.1 AC2: Graceful handling of edge cases
func TestGetPhaseSpecificHints_EmptyProfile(t *testing.T) {
	m := NewNPCManager(nil, nil)

	profile := &NPCProfile{
		ID:         "npc-empty",
		Name:       "空白NPC",
		Appearance: "普通人",
		Traits:     []Trait{}, // No traits at all
		DialogueStyle: DialogueStyle{
			Vocabulary: "口語",
			Quirks:     []string{},
		},
		InitialEmotion: DefaultEmotionState(),
	}

	err := m.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Build prompt
	prompt := m.BuildNPCPrompt("npc-empty")

	// Should not crash and should not contain hints section
	if prompt == "" {
		t.Error("Expected non-empty prompt even with no traits")
	}

	if strings.Contains(prompt, "行為暗示：") {
		t.Error("Prompt should not contain hints section when no traits exist")
	}

	// Should still contain basic sections
	if !strings.Contains(prompt, "空白NPC") {
		t.Error("Prompt should contain NPC name")
	}
}
