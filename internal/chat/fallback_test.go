package chat

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TestFallbackManager_Initialization tests that FallbackManager loads all templates correctly
func TestFallbackManager_Initialization(t *testing.T) {
	fm := NewFallbackManager()

	if fm == nil {
		t.Fatal("NewFallbackManager() returned nil")
	}

	totalTemplates := fm.GetTemplateCount()
	if totalTemplates < 50 {
		t.Errorf("Expected at least 50 templates, got %d", totalTemplates)
	}

	// Verify templates for each archetype are loaded
	scientistTemplates := fm.GetTemplatesByArchetype("Scientist")
	if len(scientistTemplates) == 0 {
		t.Error("No Scientist templates loaded")
	}

	guardTemplates := fm.GetTemplatesByArchetype("Guard")
	if len(guardTemplates) == 0 {
		t.Error("No Guard templates loaded")
	}

	survivorTemplates := fm.GetTemplatesByArchetype("Survivor")
	if len(survivorTemplates) == 0 {
		t.Error("No Survivor templates loaded")
	}

	genericTemplates := fm.GetTemplatesByArchetype("Any")
	if len(genericTemplates) == 0 {
		t.Error("No generic (Any) templates loaded")
	}

	// Verify all 7 categories are covered
	categories := []TemplateCategory{
		CategoryAgree,
		CategoryDisagree,
		CategoryConfused,
		CategoryFearful,
		CategoryCurious,
		CategoryDefensive,
		CategoryNeutral,
	}

	for _, cat := range categories {
		templates := fm.GetTemplatesByCategory(cat)
		if len(templates) == 0 {
			t.Errorf("No templates found for category: %s", cat)
		}
	}

	// Validate all templates
	if err := fm.ValidateTemplates(); err != nil {
		t.Errorf("Template validation failed: %v", err)
	}

	t.Logf("Successfully loaded %d templates across all archetypes and categories", totalTemplates)
}

// TestSelectTemplate_HighTrust tests template selection with high trust NPC
func TestSelectTemplate_HighTrust(t *testing.T) {
	fm := NewFallbackManager()

	ctx := FallbackContext{
		NPCID:       "npc_scientist_01",
		NPCName:     "Dr. Chen",
		PlayerName:  "Player",
		Emotion:     manager.NewEmotionState(75, 20, 30), // High trust, low fear
		MentalState: manager.Normal,
		Archetype:   "Scientist",
		Flags:       []string{},
	}

	result := fm.SelectTemplate(ctx)

	if result == "" {
		t.Fatal("SelectTemplate returned empty string")
	}

	t.Logf("High trust template selected: %s", result)

	// High trust should tend toward agree/curious/neutral categories
	// We can't guarantee which specific template, but it should be coherent
	if strings.Contains(result, "{npc.name}") || strings.Contains(result, "{player.name}") {
		t.Error("Template variables were not replaced")
	}
}

// TestSelectTemplate_HighFear tests template selection with high fear NPC
func TestSelectTemplate_HighFear(t *testing.T) {
	fm := NewFallbackManager()

	ctx := FallbackContext{
		NPCID:       "npc_guard_01",
		NPCName:     "Marcus",
		PlayerName:  "Player",
		Emotion:     manager.NewEmotionState(30, 75, 65), // Low trust, high fear/stress
		MentalState: manager.Anxious,
		Archetype:   "Guard",
		Flags:       []string{"hostile"},
		HasHostile:  true,
	}

	result := fm.SelectTemplate(ctx)

	if result == "" {
		t.Fatal("SelectTemplate returned empty string")
	}

	t.Logf("High fear template selected: %s", result)

	// Should select fearful category due to high fear + hostile flag
	// Verify variables are replaced
	if strings.Contains(result, "{npc.name}") || strings.Contains(result, "{player.name}") {
		t.Error("Template variables were not replaced")
	}
}

// TestSelectTemplate_Confused tests template selection with hallucination flag
func TestSelectTemplate_Confused(t *testing.T) {
	fm := NewFallbackManager()

	ctx := FallbackContext{
		NPCID:            "npc_survivor_01",
		NPCName:          "Sarah",
		PlayerName:       "Player",
		Emotion:          manager.NewEmotionState(50, 45, 55), // Moderate emotions
		MentalState:      manager.Normal,
		Archetype:        "Survivor",
		Flags:            []string{"hallucination"},
		HasHallucination: true,
	}

	result := fm.SelectTemplate(ctx)

	if result == "" {
		t.Fatal("SelectTemplate returned empty string")
	}

	t.Logf("Confused template selected: %s", result)

	// Hallucination flag should trigger confused or defensive category
	// Verify it's a valid response
	if len(result) < 3 {
		t.Error("Template result is too short to be valid dialogue")
	}
}

// TestSelectTemplate_NoMatch tests fallback behavior when no exact match exists
func TestSelectTemplate_NoMatch(t *testing.T) {
	fm := NewFallbackManager()

	// Create context with extreme/unusual conditions
	ctx := FallbackContext{
		NPCID:       "npc_unknown_01",
		NPCName:     "Unknown",
		PlayerName:  "Player",
		Emotion:     manager.NewEmotionState(50, 50, 50), // Neutral emotions
		MentalState: manager.Normal,
		Archetype:   "UnknownArchetype", // Non-existent archetype
		Flags:       []string{},
	}

	result := fm.SelectTemplate(ctx)

	if result == "" {
		t.Fatal("SelectTemplate returned empty string even with no archetype match")
	}

	t.Logf("Fallback template selected: %s", result)

	// Should fall back to generic "Any" archetype templates
	// Verify it's a reasonable response
	if len(result) < 2 {
		t.Error("Fallback template is too short")
	}
}

// TestReplaceVariables tests variable substitution in templates
func TestReplaceVariables(t *testing.T) {
	fm := NewFallbackManager()

	tests := []struct {
		name       string
		template   string
		ctx        FallbackContext
		expected   string
	}{
		{
			name:     "Replace NPC name",
			template: "I'm {npc.name}, and I need help.",
			ctx: FallbackContext{
				NPCName:    "Dr. Chen",
				PlayerName: "Alex",
			},
			expected: "I'm Dr. Chen, and I need help.",
		},
		{
			name:     "Replace player name",
			template: "Listen, {player.name}, we need to talk.",
			ctx: FallbackContext{
				NPCName:    "Marcus",
				PlayerName: "Alex",
			},
			expected: "Listen, Alex, we need to talk.",
		},
		{
			name:     "Replace both names",
			template: "{npc.name} trusts {player.name}.",
			ctx: FallbackContext{
				NPCName:    "Sarah",
				PlayerName: "Jordan",
			},
			expected: "Sarah trusts Jordan.",
		},
		{
			name:     "No variables",
			template: "This is a static message.",
			ctx: FallbackContext{
				NPCName:    "Test",
				PlayerName: "Player",
			},
			expected: "This is a static message.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.replaceVariables(tt.template, tt.ctx)
			if result != tt.expected {
				t.Errorf("replaceVariables() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestDetermineCategoryFromContext tests category determination from context flags
func TestDetermineCategoryFromContext(t *testing.T) {
	fm := NewFallbackManager()

	tests := []struct {
		name     string
		ctx      FallbackContext
		expected TemplateCategory
	}{
		{
			name: "Hallucination with low fear -> Confused",
			ctx: FallbackContext{
				HasHallucination: true,
				Emotion:          manager.NewEmotionState(50, 30, 40),
			},
			expected: CategoryConfused,
		},
		{
			name: "Hallucination with high fear -> Defensive",
			ctx: FallbackContext{
				HasHallucination: true,
				Emotion:          manager.NewEmotionState(50, 60, 50),
			},
			expected: CategoryDefensive,
		},
		{
			name: "Hostile with low fear -> Defensive",
			ctx: FallbackContext{
				HasHostile: true,
				Emotion:    manager.NewEmotionState(40, 50, 45),
			},
			expected: CategoryDefensive,
		},
		{
			name: "Hostile with high fear -> Fearful",
			ctx: FallbackContext{
				HasHostile: true,
				Emotion:    manager.NewEmotionState(30, 70, 60),
			},
			expected: CategoryFearful,
		},
		{
			name: "Revelation with high trust -> Agree",
			ctx: FallbackContext{
				HasRevelation: true,
				Emotion:       manager.NewEmotionState(65, 30, 35),
			},
			expected: CategoryAgree,
		},
		{
			name: "Revelation with low trust -> Curious",
			ctx: FallbackContext{
				HasRevelation: true,
				Emotion:       manager.NewEmotionState(40, 35, 40),
			},
			expected: CategoryCurious,
		},
		{
			name: "No flags -> Neutral",
			ctx: FallbackContext{
				Emotion: manager.NewEmotionState(50, 50, 50),
			},
			expected: CategoryNeutral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.determineCategoryFromContext(tt.ctx)
			if result != tt.expected {
				t.Errorf("determineCategoryFromContext() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMatchesConditions tests condition matching logic
func TestMatchesConditions(t *testing.T) {
	fm := NewFallbackManager()

	tests := []struct {
		name       string
		conditions TemplateConditions
		ctx        FallbackContext
		expected   bool
	}{
		{
			name: "Trust above minimum - match",
			conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
			ctx: FallbackContext{
				Emotion: manager.NewEmotionState(60, 30, 40),
			},
			expected: true,
		},
		{
			name: "Trust below minimum - no match",
			conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
			ctx: FallbackContext{
				Emotion: manager.NewEmotionState(40, 30, 40),
			},
			expected: false,
		},
		{
			name: "Fear within range - match",
			conditions: TemplateConditions{
				MinFear: intPtr(60),
				MaxFear: intPtr(80),
			},
			ctx: FallbackContext{
				Emotion: manager.NewEmotionState(50, 70, 50),
			},
			expected: true,
		},
		{
			name: "Fear outside range - no match",
			conditions: TemplateConditions{
				MinFear: intPtr(60),
				MaxFear: intPtr(80),
			},
			ctx: FallbackContext{
				Emotion: manager.NewEmotionState(50, 90, 50),
			},
			expected: false,
		},
		{
			name: "MentalState match",
			conditions: TemplateConditions{
				RequiredMentalState: mentalStatePtr(manager.Anxious),
			},
			ctx: FallbackContext{
				Emotion:     manager.NewEmotionState(50, 50, 50),
				MentalState: manager.Anxious,
			},
			expected: true,
		},
		{
			name: "MentalState mismatch",
			conditions: TemplateConditions{
				RequiredMentalState: mentalStatePtr(manager.Anxious),
			},
			ctx: FallbackContext{
				Emotion:     manager.NewEmotionState(50, 50, 50),
				MentalState: manager.Normal,
			},
			expected: false,
		},
		{
			name:       "No conditions - always match",
			conditions: TemplateConditions{},
			ctx: FallbackContext{
				Emotion:     manager.NewEmotionState(10, 90, 80),
				MentalState: manager.Corrupted,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.matchesConditions(tt.conditions, tt.ctx)
			if result != tt.expected {
				t.Errorf("matchesConditions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBuildFallbackContext tests the context builder helper
func TestBuildFallbackContext(t *testing.T) {
	emotion := manager.NewEmotionState(60, 40, 50)
	mentalState := manager.Normal

	ctx := BuildFallbackContext(
		"npc_01",
		"Dr. Chen",
		"Alex",
		"Scientist",
		emotion,
		mentalState,
		[]string{"hallucination", "hostile"},
	)

	if ctx.NPCID != "npc_01" {
		t.Errorf("NPCID = %s, want npc_01", ctx.NPCID)
	}
	if ctx.NPCName != "Dr. Chen" {
		t.Errorf("NPCName = %s, want Dr. Chen", ctx.NPCName)
	}
	if ctx.PlayerName != "Alex" {
		t.Errorf("PlayerName = %s, want Alex", ctx.PlayerName)
	}
	if ctx.Archetype != "Scientist" {
		t.Errorf("Archetype = %s, want Scientist", ctx.Archetype)
	}
	if !ctx.HasHallucination {
		t.Error("HasHallucination should be true")
	}
	if !ctx.HasHostile {
		t.Error("HasHostile should be true")
	}
	if ctx.HasRevelation {
		t.Error("HasRevelation should be false")
	}
}

// TestSelectTemplate_AlwaysReturnsValidResponse tests that SelectTemplate never fails
func TestSelectTemplate_AlwaysReturnsValidResponse(t *testing.T) {
	fm := NewFallbackManager()

	// Test with various extreme/edge case contexts
	contexts := []FallbackContext{
		// Extremely low trust, high fear
		{
			NPCID:       "test1",
			NPCName:     "NPC1",
			PlayerName:  "Player",
			Emotion:     manager.NewEmotionState(0, 100, 100),
			MentalState: manager.Corrupted,
			Archetype:   "Guard",
		},
		// Extremely high trust, low fear
		{
			NPCID:       "test2",
			NPCName:     "NPC2",
			PlayerName:  "Player",
			Emotion:     manager.NewEmotionState(100, 0, 0),
			MentalState: manager.Normal,
			Archetype:   "Scientist",
		},
		// All flags at once
		{
			NPCID:            "test3",
			NPCName:          "NPC3",
			PlayerName:       "Player",
			Emotion:          manager.NewEmotionState(50, 50, 50),
			MentalState:      manager.Anxious,
			Archetype:        "Survivor",
			HasHallucination: true,
			HasHostile:       true,
			HasRevelation:    true,
		},
		// Unknown archetype
		{
			NPCID:       "test4",
			NPCName:     "NPC4",
			PlayerName:  "Player",
			Emotion:     manager.NewEmotionState(50, 50, 50),
			MentalState: manager.Normal,
			Archetype:   "UnknownType",
		},
	}

	for i, ctx := range contexts {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			result := fm.SelectTemplate(ctx)
			if result == "" {
				t.Errorf("SelectTemplate returned empty string for context %d", i)
			}
			if len(result) < 2 {
				t.Errorf("SelectTemplate returned too short response: %s", result)
			}
			t.Logf("Context %d result: %s", i, result)
		})
	}
}

// TestTemplateDistribution tests that templates are well-distributed across categories
func TestTemplateDistribution(t *testing.T) {
	fm := NewFallbackManager()

	categories := []TemplateCategory{
		CategoryAgree,
		CategoryDisagree,
		CategoryConfused,
		CategoryFearful,
		CategoryCurious,
		CategoryDefensive,
		CategoryNeutral,
	}

	t.Log("Template distribution by category:")
	for _, cat := range categories {
		templates := fm.GetTemplatesByCategory(cat)
		t.Logf("  %s: %d templates", cat, len(templates))

		if len(templates) < 3 {
			t.Errorf("Category %s has too few templates (%d), should have at least 3", cat, len(templates))
		}
	}

	archetypes := []string{"Scientist", "Guard", "Survivor", "Any"}
	t.Log("\nTemplate distribution by archetype:")
	for _, arch := range archetypes {
		templates := fm.GetTemplatesByArchetype(arch)
		t.Logf("  %s: %d templates", arch, len(templates))

		if len(templates) < 5 {
			t.Errorf("Archetype %s has too few templates (%d), should have at least 5", arch, len(templates))
		}
	}

	t.Logf("\nTotal templates: %d", fm.GetTemplateCount())
}
