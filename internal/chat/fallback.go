package chat

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// NewFallbackManager creates a new FallbackManager with all template libraries loaded.
func NewFallbackManager() *FallbackManager {
	fm := &FallbackManager{
		templates: make([]DialogueTemplate, 0, 100),
	}

	// Load all template libraries
	fm.loadScientistTemplates()
	fm.loadGuardTemplates()
	fm.loadSurvivorTemplates()
	fm.loadGenericTemplates()

	return fm
}

// SelectTemplate selects the most appropriate dialogue template for the given context.
// This is the main entry point for fallback dialogue generation.
//
// Selection algorithm:
// 1. Determine target category from context flags
// 2. Filter candidates by archetype (prefer exact match, allow "Any")
// 3. Filter candidates by category
// 4. Filter candidates by conditions (Trust/Fear/Stress ranges, MentalState)
// 5. Randomly select from remaining candidates
// 6. Replace variables in selected template
// 7. Return final dialogue content
func (fm *FallbackManager) SelectTemplate(ctx FallbackContext) string {
	// Step 1: Determine target category from flags
	category := fm.determineCategoryFromContext(ctx)

	// Step 2-4: Filter candidates
	candidates := fm.filterTemplates(ctx, category)

	// Step 5: Select best match (random if multiple matches)
	if len(candidates) == 0 {
		// Fallback: try neutral category with relaxed conditions
		category = CategoryNeutral
		candidates = fm.filterTemplates(ctx, category)
	}

	if len(candidates) == 0 {
		// Ultimate fallback: use any generic neutral template
		for _, tmpl := range fm.templates {
			if tmpl.Archetype == "Any" && tmpl.Category == CategoryNeutral {
				candidates = append(candidates, tmpl)
			}
		}
	}

	if len(candidates) == 0 {
		// Should never happen if templates are properly loaded
		return "..."
	}

	// Randomly select from candidates
	selected := candidates[rand.Intn(len(candidates))]

	// Step 6: Replace variables
	return fm.replaceVariables(selected.Content, ctx)
}

// determineCategoryFromContext maps context flags to appropriate template category.
func (fm *FallbackManager) determineCategoryFromContext(ctx FallbackContext) TemplateCategory {
	// Priority order: Hallucination > Hostile > Revelation > Default
	if ctx.HasHallucination {
		// Hallucination triggers confusion or defensive responses
		if ctx.Emotion.Fear > 50 {
			return CategoryDefensive
		}
		return CategoryConfused
	}

	if ctx.HasHostile {
		// Hostile language triggers fearful or defensive responses
		if ctx.Emotion.Fear > 60 {
			return CategoryFearful
		}
		return CategoryDefensive
	}

	if ctx.HasRevelation {
		// Revelation triggers curiosity or agreement
		if ctx.Emotion.Trust > 50 {
			return CategoryAgree
		}
		return CategoryCurious
	}

	// Default: neutral response
	return CategoryNeutral
}

// filterTemplates filters templates based on archetype, category, and conditions.
func (fm *FallbackManager) filterTemplates(ctx FallbackContext, category TemplateCategory) []DialogueTemplate {
	candidates := make([]DialogueTemplate, 0)

	// First pass: exact archetype match + category match + conditions match
	for _, tmpl := range fm.templates {
		if tmpl.Archetype == ctx.Archetype && tmpl.Category == category {
			if fm.matchesConditions(tmpl.Conditions, ctx) {
				candidates = append(candidates, tmpl)
			}
		}
	}

	// If no exact archetype matches, try "Any" archetype
	if len(candidates) == 0 {
		for _, tmpl := range fm.templates {
			if tmpl.Archetype == "Any" && tmpl.Category == category {
				if fm.matchesConditions(tmpl.Conditions, ctx) {
					candidates = append(candidates, tmpl)
				}
			}
		}
	}

	return candidates
}

// matchesConditions checks if a template's conditions match the current context.
func (fm *FallbackManager) matchesConditions(cond TemplateConditions, ctx FallbackContext) bool {
	// Check Trust constraints
	if cond.MinTrust != nil && ctx.Emotion.Trust < *cond.MinTrust {
		return false
	}
	if cond.MaxTrust != nil && ctx.Emotion.Trust > *cond.MaxTrust {
		return false
	}

	// Check Fear constraints
	if cond.MinFear != nil && ctx.Emotion.Fear < *cond.MinFear {
		return false
	}
	if cond.MaxFear != nil && ctx.Emotion.Fear > *cond.MaxFear {
		return false
	}

	// Check Stress constraints
	if cond.MinStress != nil && ctx.Emotion.Stress < *cond.MinStress {
		return false
	}
	if cond.MaxStress != nil && ctx.Emotion.Stress > *cond.MaxStress {
		return false
	}

	// Check MentalState constraint
	if cond.RequiredMentalState != nil && ctx.MentalState != *cond.RequiredMentalState {
		return false
	}

	return true
}

// replaceVariables replaces template variables with actual values from context.
// Supported variables:
// - {npc.name} -> NPC's name
// - {player.name} -> Player's name
func (fm *FallbackManager) replaceVariables(content string, ctx FallbackContext) string {
	result := content

	// Replace {npc.name}
	result = strings.ReplaceAll(result, "{npc.name}", ctx.NPCName)

	// Replace {player.name}
	result = strings.ReplaceAll(result, "{player.name}", ctx.PlayerName)

	return result
}

// GetTemplateCount returns the total number of templates loaded.
// Used for testing and diagnostics.
func (fm *FallbackManager) GetTemplateCount() int {
	return len(fm.templates)
}

// GetTemplatesByArchetype returns all templates for a specific archetype.
// Used for testing and diagnostics.
func (fm *FallbackManager) GetTemplatesByArchetype(archetype string) []DialogueTemplate {
	result := make([]DialogueTemplate, 0)
	for _, tmpl := range fm.templates {
		if tmpl.Archetype == archetype {
			result = append(result, tmpl)
		}
	}
	return result
}

// GetTemplatesByCategory returns all templates for a specific category.
// Used for testing and diagnostics.
func (fm *FallbackManager) GetTemplatesByCategory(category TemplateCategory) []DialogueTemplate {
	result := make([]DialogueTemplate, 0)
	for _, tmpl := range fm.templates {
		if tmpl.Category == category {
			result = append(result, tmpl)
		}
	}
	return result
}

// Helper function for creating int pointers (used in template definitions)
func intPtr(v int) *int {
	return &v
}

// Helper function for creating MentalState pointers
func mentalStatePtr(m manager.MentalState) *manager.MentalState {
	return &m
}

// Template loading methods (implemented in separate files)
func (fm *FallbackManager) loadScientistTemplates() {
	fm.templates = append(fm.templates, getScientistTemplates()...)
}

func (fm *FallbackManager) loadGuardTemplates() {
	fm.templates = append(fm.templates, getGuardTemplates()...)
}

func (fm *FallbackManager) loadSurvivorTemplates() {
	fm.templates = append(fm.templates, getSurvivorTemplates()...)
}

func (fm *FallbackManager) loadGenericTemplates() {
	fm.templates = append(fm.templates, getGenericTemplates()...)
}

// BuildFallbackContext is a helper function to build a FallbackContext from NPC state.
// This simplifies integration with the NPC Manager.
func BuildFallbackContext(
	npcID string,
	npcName string,
	playerName string,
	archetype string,
	emotion manager.EmotionState,
	mentalState manager.MentalState,
	flags []string,
) FallbackContext {
	ctx := FallbackContext{
		NPCID:       npcID,
		NPCName:     npcName,
		PlayerName:  playerName,
		Emotion:     emotion,
		MentalState: mentalState,
		Archetype:   archetype,
		Flags:       flags,
	}

	// Parse flags for quick lookup
	for _, flag := range flags {
		switch strings.ToLower(flag) {
		case "hallucination":
			ctx.HasHallucination = true
		case "hostile":
			ctx.HasHostile = true
		case "revelation":
			ctx.HasRevelation = true
		}
	}

	return ctx
}

// ValidateTemplates checks all loaded templates for consistency.
// Returns error if any templates are invalid.
// Used for testing and diagnostics.
func (fm *FallbackManager) ValidateTemplates() error {
	for _, tmpl := range fm.templates {
		if tmpl.ID == "" {
			return fmt.Errorf("template with empty ID found")
		}
		if tmpl.Content == "" {
			return fmt.Errorf("template %s has empty content", tmpl.ID)
		}
		if tmpl.Archetype == "" {
			return fmt.Errorf("template %s has empty archetype", tmpl.ID)
		}
		if tmpl.Category == "" {
			return fmt.Errorf("template %s has empty category", tmpl.ID)
		}
	}
	return nil
}
