package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// DreamGenerator is an interface for dream content generation.
type DreamGenerator interface {
	GenerateOpeningDream(ctx context.Context, theme, rulesSummary, playerRole string) (string, error)
	GenerateChapterDream(ctx context.Context, dreamType engine.ChapterDreamType, context engine.ChapterDreamContext) (string, error)
}

// DreamOrchestrator handles dream generation and triggering logic.
// Story 9-1: Opening dream generation during Genesis phase.
// Story 9-2: Chapter dream generation and triggering during Game Loop.
type DreamOrchestrator struct {
	dreamGenerator DreamGenerator
	storyBible     *StoryBible
	gameState      *engine.GameStateV2
}

// NewDreamOrchestrator creates a new dream orchestrator.
func NewDreamOrchestrator(
	dreamGenerator DreamGenerator,
	storyBible *StoryBible,
	gameState *engine.GameStateV2,
) *DreamOrchestrator {
	return &DreamOrchestrator{
		dreamGenerator: dreamGenerator,
		storyBible:     storyBible,
		gameState:      gameState,
	}
}

// GenerateOpeningDream generates the opening dream during Genesis phase.
// Story 9-1 AC:
//   - Generate opening dream (200-400 chars)
//   - Include 1-2 rule clues (clarity 0.2-0.4, very subtle)
//   - Save to Story Bible dreams[] array
//   - Mark related_rule_id
func (d *DreamOrchestrator) GenerateOpeningDream(ctx context.Context) (*DreamBlueprint, error) {
	log.Println("[DreamOrchestrator] Generating opening dream...")

	// Build rules summary from Story Bible
	rulesSummary := d.buildRulesSummary()

	// Get theme from Story Bible
	theme := d.storyBible.MainTheme
	if theme == "" && d.storyBible.WorldSetting != nil {
		theme = d.storyBible.WorldSetting.Location
	}

	// Generate opening dream content using DreamGenerator
	content, err := d.dreamGenerator.GenerateOpeningDream(ctx, theme, rulesSummary, "探險者")
	if err != nil {
		return nil, fmt.Errorf("failed to generate opening dream: %w", err)
	}

	// Select 1-2 rules to hint at
	relatedRuleIDs := d.selectRandomRules(1, 2)

	// Create dream blueprint
	blueprint := &DreamBlueprint{
		ID:             fmt.Sprintf("dream-opening-%d", time.Now().Unix()),
		Type:           "opening",
		Content:        content,
		RelatedRuleIDs: relatedRuleIDs,
		Clarity:        0.3, // Very subtle (0.2-0.4)
		TriggerBeat:    0,   // Shown before prologue
		TriggerSAN:     0,   // No SAN requirement for opening
		TriggerEvent:   "",
		Symbols:        []string{}, // TODO: Extract from generated content
		Atmosphere:     "uneasy",
		IsTriggered:    false,
	}

	log.Printf("[DreamOrchestrator] Opening dream generated: ID=%s, RuleIDs=%v, Clarity=%.2f",
		blueprint.ID, blueprint.RelatedRuleIDs, blueprint.Clarity)

	return blueprint, nil
}

// GenerateChapterDreams generates multiple chapter dreams during Genesis phase.
// Story 9-2 AC:
//   - Generate chapter dreams with progressive clarity
//   - Set trigger conditions (SAN < 50, specific beats, key plot points)
//   - Save to Story Bible dreams[] array
func (d *DreamOrchestrator) GenerateChapterDreams(ctx context.Context, count int) ([]*DreamBlueprint, error) {
	log.Printf("[DreamOrchestrator] Generating %d chapter dreams...", count)

	blueprints := make([]*DreamBlueprint, 0, count)

	// Get total beats from Story Bible
	totalBeats := d.storyBible.TotalBeats
	if totalBeats == 0 {
		totalBeats = 20 // Default
	}

	// Generate dreams with varying triggers and increasing clarity
	for i := 0; i < count; i++ {
		// Progressive clarity: starts at 0.4, increases to 0.8
		clarity := 0.4 + (float64(i) * 0.2)
		if clarity > 0.8 {
			clarity = 0.8
		}

		// Distribute trigger beats across the game
		triggerBeat := (totalBeats / (count + 1)) * (i + 1)

		// Vary trigger conditions
		var triggerSAN int
		var triggerEvent string

		switch i % 3 {
		case 0:
			// Low SAN trigger
			triggerSAN = 50
		case 1:
			// Beat-based trigger
			triggerSAN = 0
		case 2:
			// Event-based trigger
			triggerSAN = 0
			triggerEvent = "rule_violation"
		}

		// Select rules to hint at (1-2 rules)
		relatedRuleIDs := d.selectRandomRules(1, 2)

		blueprint := &DreamBlueprint{
			ID:             fmt.Sprintf("dream-chapter-%d-%d", i+1, time.Now().Unix()),
			Type:           "chapter",
			Content:        "", // Will be generated when triggered
			RelatedRuleIDs: relatedRuleIDs,
			Clarity:        clarity,
			TriggerBeat:    triggerBeat,
			TriggerSAN:     triggerSAN,
			TriggerEvent:   triggerEvent,
			Symbols:        []string{},
			Atmosphere:     d.determineAtmosphere(clarity),
			IsTriggered:    false,
		}

		blueprints = append(blueprints, blueprint)

		log.Printf("[DreamOrchestrator] Chapter dream %d generated: ID=%s, TriggerBeat=%d, TriggerSAN=%d, Clarity=%.2f",
			i+1, blueprint.ID, blueprint.TriggerBeat, blueprint.TriggerSAN, blueprint.Clarity)
	}

	return blueprints, nil
}

// CheckDreamTriggers checks if any dreams should be triggered.
// Story 9-2 AC:
//   - Check SAN < 50
//   - Check specific beat numbers
//   - Check key plot points
//   - Return triggered dreams
func (d *DreamOrchestrator) CheckDreamTriggers(currentBeat int, currentSAN int, recentEvent string) []*DreamBlueprint {
	if d.storyBible == nil || len(d.storyBible.Dreams) == 0 {
		return nil
	}

	triggered := make([]*DreamBlueprint, 0)

	for _, blueprint := range d.storyBible.Dreams {
		// Skip already triggered dreams
		if blueprint.IsTriggered {
			continue
		}

		// Skip opening dream (handled separately)
		if blueprint.Type == "opening" {
			continue
		}

		// Check trigger conditions
		shouldTrigger := false

		// Condition 1: Beat-based trigger
		if blueprint.TriggerBeat > 0 && currentBeat >= blueprint.TriggerBeat {
			shouldTrigger = true
			log.Printf("[DreamOrchestrator] Beat trigger met: Dream=%s, Beat=%d", blueprint.ID, currentBeat)
		}

		// Condition 2: SAN-based trigger
		if blueprint.TriggerSAN > 0 && currentSAN < blueprint.TriggerSAN {
			shouldTrigger = true
			log.Printf("[DreamOrchestrator] SAN trigger met: Dream=%s, SAN=%d", blueprint.ID, currentSAN)
		}

		// Condition 3: Event-based trigger
		if blueprint.TriggerEvent != "" && blueprint.TriggerEvent == recentEvent {
			shouldTrigger = true
			log.Printf("[DreamOrchestrator] Event trigger met: Dream=%s, Event=%s", blueprint.ID, recentEvent)
		}

		if shouldTrigger {
			triggered = append(triggered, blueprint)
		}
	}

	return triggered
}

// GenerateDreamContent generates the actual dream content when triggered.
// Story 9-2 AC:
//   - Generate dream narrative (100-300 chars)
//   - Include progressive clues based on clarity
//   - Use DreamGenerator with appropriate context
func (d *DreamOrchestrator) GenerateDreamContent(
	ctx context.Context,
	blueprint *DreamBlueprint,
	recentEvents string,
) (string, error) {
	log.Printf("[DreamOrchestrator] Generating dream content for: %s", blueprint.ID)

	// If content already exists (e.g., opening dream), return it
	if blueprint.Content != "" {
		return blueprint.Content, nil
	}

	// Build rule hints based on related rules
	ruleHints := d.buildRuleHintsForDream(blueprint.RelatedRuleIDs, blueprint.Clarity)

	// Determine dream type based on context
	dreamType := d.determineDreamType(blueprint)

	// Build context for chapter dream
	chapterContext := engine.ChapterDreamContext{
		ChapterNum:     d.gameState.GetCurrentBeat(),
		RecentEvents:   recentEvents,
		RuleHints:      ruleHints,
		PlayerSAN:      d.gameState.GetSAN(),
		KnownClues:     []string{}, // TODO: Extract from game state
		DeadTeammates:  []string{}, // TODO: Extract from NPC states
		HighStress:     d.gameState.GetSAN() < 30,
		RecentClue:     false, // TODO: Track recent clue discoveries
		TeammateDeaths: false, // TODO: Track recent teammate deaths
	}

	// Generate content using DreamGenerator
	content, err := d.dreamGenerator.GenerateChapterDream(ctx, dreamType, chapterContext)
	if err != nil {
		return "", fmt.Errorf("failed to generate chapter dream: %w", err)
	}

	return content, nil
}

// RecordDreamToGameState records a triggered dream to the game state.
// Story 9-2 AC:
//   - Save dream to DreamLog
//   - Mark blueprint as triggered
func (d *DreamOrchestrator) RecordDreamToGameState(blueprint *DreamBlueprint, content string) {
	// Create dream record
	record := game.DreamRecord{
		ID:        blueprint.ID,
		Type:      game.DreamType(blueprint.Type),
		Timestamp: time.Now(),
		Content:   content,
		RelatedRuleID: func() string {
			if len(blueprint.RelatedRuleIDs) > 0 {
				return blueprint.RelatedRuleIDs[0]
			}
			return ""
		}(),
		Context: game.DreamContext{
			PlayerHP:     d.gameState.GetHP(),
			PlayerSAN:    d.gameState.GetSAN(),
			ChapterNum:   d.gameState.GetCurrentBeat(),
			KnownClues:   []string{}, // TODO: Extract from game state
			StoryTheme:   d.storyBible.MainTheme,
			RulesSummary: d.buildRulesSummary(),
		},
	}

	// Record to game state
	d.gameState.RecordDream(record)

	// Mark blueprint as triggered
	blueprint.IsTriggered = true

	log.Printf("[DreamOrchestrator] Dream recorded: ID=%s, Type=%s, Beat=%d",
		record.ID, record.Type, record.Context.ChapterNum)
}

// ==========================================================================
// Helper Methods
// ==========================================================================

// buildRulesSummary builds a summary of all hidden rules.
func (d *DreamOrchestrator) buildRulesSummary() string {
	if d.storyBible == nil || len(d.storyBible.HiddenRules) == 0 {
		return "未知規則"
	}

	summary := ""
	for i, rule := range d.storyBible.HiddenRules {
		if i > 0 {
			summary += "; "
		}
		summary += rule.Name
	}

	return summary
}

// selectRandomRules selects a random subset of rules.
func (d *DreamOrchestrator) selectRandomRules(min, max int) []string {
	if d.storyBible == nil || len(d.storyBible.HiddenRules) == 0 {
		return []string{}
	}

	// For now, select the first 1-2 rules
	// TODO: Implement proper random selection
	count := min
	if len(d.storyBible.HiddenRules) > max {
		count = max
	} else if len(d.storyBible.HiddenRules) < min {
		count = len(d.storyBible.HiddenRules)
	}

	ruleIDs := make([]string, 0, count)
	for i := 0; i < count && i < len(d.storyBible.HiddenRules); i++ {
		ruleIDs = append(ruleIDs, d.storyBible.HiddenRules[i].ID)
	}

	return ruleIDs
}

// buildRuleHintsForDream builds rule hints text based on clarity.
func (d *DreamOrchestrator) buildRuleHintsForDream(ruleIDs []string, clarity float64) string {
	if len(ruleIDs) == 0 {
		return ""
	}

	hints := ""
	for _, ruleID := range ruleIDs {
		rule := d.findRuleByID(ruleID)
		if rule != nil {
			if hints != "" {
				hints += "; "
			}
			// Adjust hint detail based on clarity
			if clarity < 0.4 {
				// Very subtle - just mention the theme
				hints += fmt.Sprintf("與%s相關", rule.Name)
			} else if clarity < 0.6 {
				// Moderate - provide some context
				hints += fmt.Sprintf("%s: %s", rule.Name, rule.Description[:min(len(rule.Description), 50)])
			} else {
				// Clear - provide detailed hint
				if len(rule.Hints) > 0 {
					hints += fmt.Sprintf("%s: %s", rule.Name, rule.Hints[0])
				} else {
					hints += fmt.Sprintf("%s: %s", rule.Name, rule.Description)
				}
			}
		}
	}

	return hints
}

// findRuleByID finds a rule by its ID.
func (d *DreamOrchestrator) findRuleByID(ruleID string) *HiddenRule {
	if d.storyBible == nil {
		return nil
	}

	for _, rule := range d.storyBible.HiddenRules {
		if rule.ID == ruleID {
			return rule
		}
	}

	return nil
}

// determineDreamType determines the appropriate dream type.
func (d *DreamOrchestrator) determineDreamType(blueprint *DreamBlueprint) engine.ChapterDreamType {
	// Check SAN level
	if d.gameState.GetSAN() < 30 {
		return engine.DreamTypeNightmare
	}

	// Check for event triggers
	if blueprint.TriggerEvent == "npc_death" {
		return engine.DreamTypeGrief
	}

	if blueprint.TriggerEvent == "rule_violation" {
		return engine.DreamTypeWarning
	}

	// Check clarity level
	if blueprint.Clarity >= 0.6 {
		return engine.DreamTypeHint
	}

	// Default to random
	return engine.DreamTypeRandom
}

// determineAtmosphere determines dream atmosphere based on clarity.
func (d *DreamOrchestrator) determineAtmosphere(clarity float64) string {
	if clarity < 0.4 {
		return "calm"
	} else if clarity < 0.7 {
		return "uneasy"
	}
	return "nightmare"
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
