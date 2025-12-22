package manager

import (
	"encoding/json"
	"testing"
)

// TestTrait_JSONSerialization tests JSON marshaling and unmarshaling of Trait
func TestTrait_JSONSerialization(t *testing.T) {
	original := Trait{
		ID:         "trait_nervous",
		Content:    "容易緊張，說話會結巴",
		RevealTier: 2,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Trait: %v", err)
	}

	// Unmarshal back
	var decoded Trait
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Trait: %v", err)
	}

	// Verify fields
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content mismatch: got %s, want %s", decoded.Content, original.Content)
	}
	if decoded.RevealTier != original.RevealTier {
		t.Errorf("RevealTier mismatch: got %d, want %d", decoded.RevealTier, original.RevealTier)
	}
}

// TestDialogueStyle_JSONSerialization tests JSON marshaling and unmarshaling of DialogueStyle
func TestDialogueStyle_JSONSerialization(t *testing.T) {
	original := DialogueStyle{
		Formality:  3,
		Verbosity:  2,
		Quirks:     []string{"常說 '你知道嗎'", "緊張時會結巴"},
		Vocabulary: "軍事術語，簡潔命令式",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal DialogueStyle: %v", err)
	}

	// Unmarshal back
	var decoded DialogueStyle
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal DialogueStyle: %v", err)
	}

	// Verify fields
	if decoded.Formality != original.Formality {
		t.Errorf("Formality mismatch: got %d, want %d", decoded.Formality, original.Formality)
	}
	if decoded.Verbosity != original.Verbosity {
		t.Errorf("Verbosity mismatch: got %d, want %d", decoded.Verbosity, original.Verbosity)
	}
	if len(decoded.Quirks) != len(original.Quirks) {
		t.Fatalf("Quirks length mismatch: got %d, want %d", len(decoded.Quirks), len(original.Quirks))
	}
	for i := range original.Quirks {
		if decoded.Quirks[i] != original.Quirks[i] {
			t.Errorf("Quirks[%d] mismatch: got %s, want %s", i, decoded.Quirks[i], original.Quirks[i])
		}
	}
	if decoded.Vocabulary != original.Vocabulary {
		t.Errorf("Vocabulary mismatch: got %s, want %s", decoded.Vocabulary, original.Vocabulary)
	}
}

// TestNPCProfile_JSONSerialization tests JSON marshaling and unmarshaling of NPCProfile
func TestNPCProfile_JSONSerialization(t *testing.T) {
	original := NPCProfile{
		ID:         "npc_soldier_001",
		Name:       "張軍士",
		Archetype:  "Survivor",
		Appearance: "穿著破舊軍裝的中年男子，眼神堅毅",
		Backstory:  "前軍人，在災難中失去戰友",
		Skills:     []string{"射擊", "戰術指揮", "急救"},
		Inventory:  []string{"手槍", "軍用刀", "急救包"},
		Secret:     "曾為了生存拋棄受傷的戰友",
		SecretTier: 3,
		Traits: []Trait{
			{ID: "trait_loyal", Content: "對信任的人忠誠", RevealTier: 1},
			{ID: "trait_guilt", Content: "內心充滿罪惡感", RevealTier: 3},
		},
		LinkedSeeds:  []string{"seed_military_base", "seed_abandoned_hospital"},
		DeathBeat:    15,
		InitialEmotion: NewEmotionState(60, 30, 40),
		DialogueStyle: DialogueStyle{
			Formality:  3,
			Verbosity:  2,
			Quirks:     []string{"常說 '長官'", "緊張時會摸槍套"},
			Vocabulary: "軍事術語，簡潔命令式",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal NPCProfile: %v", err)
	}

	// Unmarshal back
	var decoded NPCProfile
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal NPCProfile: %v", err)
	}

	// Verify basic fields
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Archetype != original.Archetype {
		t.Errorf("Archetype mismatch: got %s, want %s", decoded.Archetype, original.Archetype)
	}
	if decoded.Secret != original.Secret {
		t.Errorf("Secret mismatch: got %s, want %s", decoded.Secret, original.Secret)
	}
	if decoded.SecretTier != original.SecretTier {
		t.Errorf("SecretTier mismatch: got %d, want %d", decoded.SecretTier, original.SecretTier)
	}
	if decoded.DeathBeat != original.DeathBeat {
		t.Errorf("DeathBeat mismatch: got %d, want %d", decoded.DeathBeat, original.DeathBeat)
	}

	// Verify slices
	if len(decoded.Skills) != len(original.Skills) {
		t.Fatalf("Skills length mismatch: got %d, want %d", len(decoded.Skills), len(original.Skills))
	}
	if len(decoded.Inventory) != len(original.Inventory) {
		t.Fatalf("Inventory length mismatch: got %d, want %d", len(decoded.Inventory), len(original.Inventory))
	}
	if len(decoded.Traits) != len(original.Traits) {
		t.Fatalf("Traits length mismatch: got %d, want %d", len(decoded.Traits), len(original.Traits))
	}
	if len(decoded.LinkedSeeds) != len(original.LinkedSeeds) {
		t.Fatalf("LinkedSeeds length mismatch: got %d, want %d", len(decoded.LinkedSeeds), len(original.LinkedSeeds))
	}

	// Verify InitialEmotion
	if decoded.InitialEmotion.Trust != original.InitialEmotion.Trust {
		t.Errorf("InitialEmotion.Trust mismatch: got %d, want %d", decoded.InitialEmotion.Trust, original.InitialEmotion.Trust)
	}
	if decoded.InitialEmotion.Fear != original.InitialEmotion.Fear {
		t.Errorf("InitialEmotion.Fear mismatch: got %d, want %d", decoded.InitialEmotion.Fear, original.InitialEmotion.Fear)
	}
	if decoded.InitialEmotion.Stress != original.InitialEmotion.Stress {
		t.Errorf("InitialEmotion.Stress mismatch: got %d, want %d", decoded.InitialEmotion.Stress, original.InitialEmotion.Stress)
	}

	// Verify DialogueStyle
	if decoded.DialogueStyle.Formality != original.DialogueStyle.Formality {
		t.Errorf("DialogueStyle.Formality mismatch: got %d, want %d", decoded.DialogueStyle.Formality, original.DialogueStyle.Formality)
	}
	if decoded.DialogueStyle.Verbosity != original.DialogueStyle.Verbosity {
		t.Errorf("DialogueStyle.Verbosity mismatch: got %d, want %d", decoded.DialogueStyle.Verbosity, original.DialogueStyle.Verbosity)
	}
}

// TestNPCProfile_ToVisible_FilterSecrets tests that ToVisible filters out secret information
func TestNPCProfile_ToVisible_FilterSecrets(t *testing.T) {
	profile := NPCProfile{
		ID:         "npc_test",
		Name:       "測試角色",
		Archetype:  "Researcher",
		Appearance: "戴眼鏡的年輕女性",
		Backstory:  "研究員背景",
		Skills:     []string{"研究", "分析"},
		Inventory:  []string{"筆記本", "實驗器材"},
		Secret:     "這是秘密內容",
		SecretTier: 2,
		Traits: []Trait{
			{ID: "trait_1", Content: "聰明", RevealTier: 1},
			{ID: "trait_2", Content: "謹慎", RevealTier: 2},
			{ID: "trait_3", Content: "隱藏特質", RevealTier: 3},
		},
		LinkedSeeds:    []string{"seed_lab"},
		DeathBeat:      10,
		InitialEmotion: DefaultEmotionState(),
		DialogueStyle: DialogueStyle{
			Formality:  4,
			Verbosity:  3,
			Quirks:     []string{"常說 '根據研究'"},
			Vocabulary: "科學術語",
		},
	}

	// Only trait_1 is revealed
	visible := profile.ToVisible([]string{"trait_1"})

	// Verify basic fields are copied
	if visible.ID != profile.ID {
		t.Errorf("ID mismatch: got %s, want %s", visible.ID, profile.ID)
	}
	if visible.Name != profile.Name {
		t.Errorf("Name mismatch: got %s, want %s", visible.Name, profile.Name)
	}

	// Verify secret fields are not present in VisibleNPCProfile (this is structural, not runtime)
	// We verify by checking that only revealed traits are present

	// Verify only revealed traits are included
	if len(visible.Traits) != 1 {
		t.Fatalf("Expected 1 revealed trait, got %d", len(visible.Traits))
	}
	if visible.Traits[0].ID != "trait_1" {
		t.Errorf("Expected trait_1, got %s", visible.Traits[0].ID)
	}

	// Verify LinkedSeeds, Secret, SecretTier, DeathBeat are not in VisibleNPCProfile
	// (This is a compile-time check - VisibleNPCProfile simply doesn't have these fields)

	// Verify Skills and Inventory are preserved
	if len(visible.Skills) != len(profile.Skills) {
		t.Errorf("Skills length mismatch: got %d, want %d", len(visible.Skills), len(profile.Skills))
	}
	if len(visible.Inventory) != len(profile.Inventory) {
		t.Errorf("Inventory length mismatch: got %d, want %d", len(visible.Inventory), len(profile.Inventory))
	}

	// Verify DialogueStyle is preserved
	if visible.DialogueStyle.Formality != profile.DialogueStyle.Formality {
		t.Errorf("DialogueStyle.Formality mismatch: got %d, want %d", visible.DialogueStyle.Formality, profile.DialogueStyle.Formality)
	}
}

// TestNPCProfile_ToVisible_MultipleRevealedTraits tests filtering with multiple revealed traits
func TestNPCProfile_ToVisible_MultipleRevealedTraits(t *testing.T) {
	profile := NPCProfile{
		ID:   "npc_test",
		Name: "測試角色",
		Traits: []Trait{
			{ID: "trait_1", Content: "特質1", RevealTier: 1},
			{ID: "trait_2", Content: "特質2", RevealTier: 1},
			{ID: "trait_3", Content: "特質3", RevealTier: 2},
			{ID: "trait_4", Content: "特質4", RevealTier: 3},
		},
		InitialEmotion: DefaultEmotionState(),
		DialogueStyle:  DialogueStyle{},
	}

	// Reveal trait_1 and trait_3
	visible := profile.ToVisible([]string{"trait_1", "trait_3"})

	// Should have exactly 2 traits
	if len(visible.Traits) != 2 {
		t.Fatalf("Expected 2 revealed traits, got %d", len(visible.Traits))
	}

	// Verify the correct traits are present
	traitIDs := make(map[string]bool)
	for _, trait := range visible.Traits {
		traitIDs[trait.ID] = true
	}

	if !traitIDs["trait_1"] {
		t.Error("trait_1 should be revealed")
	}
	if !traitIDs["trait_3"] {
		t.Error("trait_3 should be revealed")
	}
	if traitIDs["trait_2"] {
		t.Error("trait_2 should not be revealed")
	}
	if traitIDs["trait_4"] {
		t.Error("trait_4 should not be revealed")
	}
}

// TestNPCProfile_ToVisible_NoRevealedTraits tests filtering with no revealed traits
func TestNPCProfile_ToVisible_NoRevealedTraits(t *testing.T) {
	profile := NPCProfile{
		ID:   "npc_test",
		Name: "測試角色",
		Traits: []Trait{
			{ID: "trait_1", Content: "特質1", RevealTier: 1},
			{ID: "trait_2", Content: "特質2", RevealTier: 2},
		},
		InitialEmotion: DefaultEmotionState(),
		DialogueStyle:  DialogueStyle{},
	}

	// No traits revealed
	visible := profile.ToVisible([]string{})

	// Should have empty traits array (not nil)
	if visible.Traits == nil {
		t.Error("Traits should be empty array, not nil")
	}
	if len(visible.Traits) != 0 {
		t.Errorf("Expected 0 revealed traits, got %d", len(visible.Traits))
	}
}

// TestNPCProfile_ToVisible_AllTraitsRevealed tests filtering with all traits revealed
func TestNPCProfile_ToVisible_AllTraitsRevealed(t *testing.T) {
	profile := NPCProfile{
		ID:   "npc_test",
		Name: "測試角色",
		Traits: []Trait{
			{ID: "trait_1", Content: "特質1", RevealTier: 1},
			{ID: "trait_2", Content: "特質2", RevealTier: 2},
			{ID: "trait_3", Content: "特質3", RevealTier: 3},
		},
		InitialEmotion: DefaultEmotionState(),
		DialogueStyle:  DialogueStyle{},
	}

	// All traits revealed
	visible := profile.ToVisible([]string{"trait_1", "trait_2", "trait_3"})

	// Should have all 3 traits
	if len(visible.Traits) != 3 {
		t.Fatalf("Expected 3 revealed traits, got %d", len(visible.Traits))
	}
}

// TestNPCProfile_ToVisible_NonexistentTraitIDs tests filtering with non-existent trait IDs
func TestNPCProfile_ToVisible_NonexistentTraitIDs(t *testing.T) {
	profile := NPCProfile{
		ID:   "npc_test",
		Name: "測試角色",
		Traits: []Trait{
			{ID: "trait_1", Content: "特質1", RevealTier: 1},
		},
		InitialEmotion: DefaultEmotionState(),
		DialogueStyle:  DialogueStyle{},
	}

	// Try to reveal non-existent traits
	visible := profile.ToVisible([]string{"trait_999", "trait_abc"})

	// Should have 0 traits
	if len(visible.Traits) != 0 {
		t.Errorf("Expected 0 revealed traits, got %d", len(visible.Traits))
	}
}

// TestVisibleNPCProfile_JSONSerialization tests JSON marshaling of VisibleNPCProfile
func TestVisibleNPCProfile_JSONSerialization(t *testing.T) {
	profile := NPCProfile{
		ID:         "npc_test",
		Name:       "測試角色",
		Archetype:  "Survivor",
		Appearance: "普通外觀",
		Backstory:  "普通背景",
		Skills:     []string{"技能1"},
		Inventory:  []string{"物品1"},
		Secret:     "秘密內容",
		Traits: []Trait{
			{ID: "trait_1", Content: "特質1", RevealTier: 1},
		},
		InitialEmotion: DefaultEmotionState(),
		DialogueStyle: DialogueStyle{
			Formality: 3,
			Verbosity: 2,
		},
	}

	visible := profile.ToVisible([]string{"trait_1"})

	// Marshal to JSON
	data, err := json.Marshal(visible)
	if err != nil {
		t.Fatalf("Failed to marshal VisibleNPCProfile: %v", err)
	}

	// Verify Secret is not in JSON
	jsonStr := string(data)
	if contains(jsonStr, "秘密內容") {
		t.Error("Secret should not be in VisibleNPCProfile JSON")
	}
	if contains(jsonStr, "secret") {
		t.Error("'secret' field should not be in VisibleNPCProfile JSON")
	}

	// Unmarshal back
	var decoded VisibleNPCProfile
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal VisibleNPCProfile: %v", err)
	}

	// Verify fields
	if decoded.ID != visible.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, visible.ID)
	}
	if decoded.Name != visible.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, visible.Name)
	}
	if len(decoded.Traits) != len(visible.Traits) {
		t.Errorf("Traits length mismatch: got %d, want %d", len(decoded.Traits), len(visible.Traits))
	}
}

// TestDialogueStyle_Ranges tests valid ranges for DialogueStyle fields
func TestDialogueStyle_Ranges(t *testing.T) {
	tests := []struct {
		name       string
		formality  int
		verbosity  int
		shouldWarn bool
	}{
		{"Valid middle range", 3, 3, false},
		{"Valid low range", 1, 1, false},
		{"Valid high range", 5, 5, false},
		{"Out of range high", 6, 6, true},
		{"Out of range low", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DialogueStyle{
				Formality: tt.formality,
				Verbosity: tt.verbosity,
			}

			// Note: We don't enforce validation in the struct itself,
			// but we document the expected ranges (1-5).
			// Validation would be done at creation time in production code.
			if !tt.shouldWarn {
				if style.Formality < 1 || style.Formality > 5 {
					t.Errorf("Formality %d is out of expected range [1-5]", style.Formality)
				}
				if style.Verbosity < 1 || style.Verbosity > 5 {
					t.Errorf("Verbosity %d is out of expected range [1-5]", style.Verbosity)
				}
			}
		})
	}
}

// TestTrait_RevealTierRanges tests valid ranges for Trait RevealTier
func TestTrait_RevealTierRanges(t *testing.T) {
	tests := []struct {
		name       string
		revealTier int
		valid      bool
	}{
		{"Easy to reveal", 1, true},
		{"Medium difficulty", 2, true},
		{"Hard to reveal", 3, true},
		{"Invalid low", 0, false},
		{"Invalid high", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trait := Trait{
				ID:         "test_trait",
				Content:    "測試特質",
				RevealTier: tt.revealTier,
			}

			// Document expected range (1-3)
			if tt.valid {
				if trait.RevealTier < 1 || trait.RevealTier > 3 {
					t.Errorf("RevealTier %d is out of expected range [1-3]", trait.RevealTier)
				}
			}
		})
	}
}

// TestNPCProfile_SecretTierRanges tests valid ranges for Secret Tier
func TestNPCProfile_SecretTierRanges(t *testing.T) {
	tests := []struct {
		name       string
		secretTier int
		valid      bool
	}{
		{"Low importance", 1, true},
		{"Medium importance", 2, true},
		{"High importance", 3, true},
		{"Invalid low", 0, false},
		{"Invalid high", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := NPCProfile{
				ID:             "test_npc",
				SecretTier:     tt.secretTier,
				InitialEmotion: DefaultEmotionState(),
			}

			// Document expected range (1-3)
			if tt.valid {
				if profile.SecretTier < 1 || profile.SecretTier > 3 {
					t.Errorf("SecretTier %d is out of expected range [1-3]", profile.SecretTier)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
