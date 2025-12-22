package rules

import (
	"fmt"
	"path/filepath"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/templates"
)

// RuleTemplateLibrary manages rule templates loaded from YAML files.
// Story 7.2: Integrates with template system to load rule templates.
type RuleTemplateLibrary struct {
	loader    *templates.TemplateLoader
	templates map[string]*templates.RuleTemplate // ID -> Template
	byCategory map[string][]*templates.RuleTemplate // Category -> Templates
	byDifficulty map[string][]*templates.RuleTemplate // Difficulty -> Templates
}

// NewRuleTemplateLibrary creates a new rule template library.
func NewRuleTemplateLibrary(templatesBaseDir string) *RuleTemplateLibrary {
	return &RuleTemplateLibrary{
		loader:       templates.NewTemplateLoader(templatesBaseDir),
		templates:    make(map[string]*templates.RuleTemplate),
		byCategory:   make(map[string][]*templates.RuleTemplate),
		byDifficulty: make(map[string][]*templates.RuleTemplate),
	}
}

// LoadAllRuleTemplates loads all rule templates from the templates/rules directory.
// Story 7.2: Loads sensory, spatial, and social rule templates.
func (lib *RuleTemplateLibrary) LoadAllRuleTemplates() error {
	// Load from templates/rules directory
	count := lib.loader.LoadDirectory("rules", func(filePath string) error {
		return lib.loadRuleTemplateFile(filePath)
	})

	if count == 0 {
		return fmt.Errorf("no rule template files loaded")
	}

	if lib.loader.HasErrors() {
		return fmt.Errorf("template loading errors: %s", lib.loader.GetErrorSummary())
	}

	return nil
}

// loadRuleTemplateFile loads a single rule template YAML file.
func (lib *RuleTemplateLibrary) loadRuleTemplateFile(filePath string) error {
	var collection templates.RuleTemplateCollection

	if err := lib.loader.LoadYAMLFile(filePath, &collection); err != nil {
		return err
	}

	// Validate and index templates
	for _, template := range collection.Rules {
		if err := template.Validate(); err != nil {
			return fmt.Errorf("invalid template %s in %s: %w", template.ID, filePath, err)
		}

		// Index by ID
		lib.templates[template.ID] = template

		// Index by category
		categoryKey := string(template.Category)
		lib.byCategory[categoryKey] = append(lib.byCategory[categoryKey], template)

		// Index by difficulty
		difficultyKey := string(template.Difficulty)
		lib.byDifficulty[difficultyKey] = append(lib.byDifficulty[difficultyKey], template)
	}

	return nil
}

// GetTemplate retrieves a template by ID.
func (lib *RuleTemplateLibrary) GetTemplate(id string) (*templates.RuleTemplate, error) {
	template, ok := lib.templates[id]
	if !ok {
		return nil, fmt.Errorf("template %s not found", id)
	}
	return template, nil
}

// GetTemplatesByCategory returns all templates in a category.
func (lib *RuleTemplateLibrary) GetTemplatesByCategory(category string) []*templates.RuleTemplate {
	return lib.byCategory[category]
}

// GetTemplatesByDifficulty returns all templates for a difficulty level.
func (lib *RuleTemplateLibrary) GetTemplatesByDifficulty(difficulty string) []*templates.RuleTemplate {
	return lib.byDifficulty[difficulty]
}

// FilterTemplates returns templates matching difficulty and optionally category.
// Story 7.2 AC1-AC3: Filter templates based on difficulty and category.
func (lib *RuleTemplateLibrary) FilterTemplates(difficulty game.DifficultyLevel, category string) []*templates.RuleTemplate {
	var filtered []*templates.RuleTemplate

	// Map game difficulty to template difficulty
	difficultyStr := mapGameDifficultyToTemplate(difficulty)

	// Get templates by difficulty
	candidates := lib.byDifficulty[difficultyStr]

	// Filter by category if specified
	if category != "" {
		for _, template := range candidates {
			if string(template.Category) == category {
				filtered = append(filtered, template)
			}
		}
		return filtered
	}

	return candidates
}

// mapGameDifficultyToTemplate maps game difficulty to template difficulty string.
func mapGameDifficultyToTemplate(difficulty game.DifficultyLevel) string {
	switch difficulty {
	case game.DifficultyEasy:
		return "easy"
	case game.DifficultyHard:
		return "medium" // Hard mode uses medium and hard templates
	case game.DifficultyHell:
		return "hard" // Hell mode uses hard and expert templates
	default:
		return "easy"
	}
}

// GetAllTemplates returns all loaded templates.
func (lib *RuleTemplateLibrary) GetAllTemplates() []*templates.RuleTemplate {
	templates := make([]*templates.RuleTemplate, 0, len(lib.templates))
	for _, template := range lib.templates {
		templates = append(templates, template)
	}
	return templates
}

// Count returns the total number of loaded templates.
func (lib *RuleTemplateLibrary) Count() int {
	return len(lib.templates)
}

// GetCategories returns all available categories.
func (lib *RuleTemplateLibrary) GetCategories() []string {
	categories := make([]string, 0, len(lib.byCategory))
	for category := range lib.byCategory {
		categories = append(categories, category)
	}
	return categories
}

// ConvertTemplateToHiddenRule converts a RuleTemplate to a HiddenRule.
// Story 7.2: Converts YAML template to runtime HiddenRule structure.
func ConvertTemplateToHiddenRule(template *templates.RuleTemplate, difficulty game.DifficultyLevel) *HiddenRule {
	rule := &HiddenRule{
		ID:               template.ID,
		Name:             template.Name,
		Category:         string(template.Category),
		Difficulty:       string(template.Difficulty),
		TriggerMedium:    template.TriggerMedium,
		TriggerCondition: extractTriggerCondition(template.TriggerMedium),
		FalseClue:        template.FalseClue,
		SurvivalRule:     template.SurvivalRule,
		Punishment: HiddenRulePunishment{
			HPDamage:     template.Punishment.HPDamage,
			SANDamage:    template.Punishment.SANDamage,
			IsFatal:      isRuleFatal(template, difficulty),
			CustomEffect: template.Punishment.Effect,
		},
		ClueHints:    convertClueHintsToTiered(template.ClueHints, difficulty),
		TimesHinted:  0,
		IsViolated:   false,
		RelatedRules: []string{},
	}

	return rule
}

// extractTriggerCondition extracts trigger keywords from trigger medium.
// This is a simple implementation - could be enhanced with more sophisticated parsing.
func extractTriggerCondition(triggerMedium string) string {
	// For now, return the trigger medium as-is
	// In a more sophisticated implementation, this could parse and extract keywords
	return triggerMedium
}

// isRuleFatal determines if a rule should be fatal based on template and difficulty.
// Story 7.2 AC2-AC3: Hard has 20-30% fatal, Hell has 50%+ fatal.
func isRuleFatal(template *templates.RuleTemplate, difficulty game.DifficultyLevel) bool {
	// If template punishment already defines fatality
	totalDamage := template.Punishment.HPDamage + template.Punishment.SANDamage

	// Rules that deal 80+ total damage are considered fatal
	if totalDamage >= 80 {
		return true
	}

	// Difficulty-based fatality (handled by generator)
	return false
}

// convertClueHintsToTiered converts simple clue hints to tiered clue hints.
// Story 7.2 AC5: Creates tiered clue hints with beat ranges.
func convertClueHintsToTiered(hints []string, difficulty game.DifficultyLevel) []ClueHint {
	if len(hints) == 0 {
		return []ClueHint{}
	}

	tieredHints := make([]ClueHint, 0, len(hints))

	// Distribute hints across tiers based on difficulty
	// Easy: All 3 tiers, Hard: Tiers 1-2, Hell: Tier 1 only
	maxTier := 3
	switch difficulty {
	case game.DifficultyHard:
		maxTier = 2
	case game.DifficultyHell:
		maxTier = 1
	}

	// Distribute hints across available tiers
	for i, hint := range hints {
		// Calculate tier (cycle through available tiers)
		tier := (i % maxTier) + 1

		// Calculate beat range based on tier
		// Tier 1: Beats 1-8 (early game)
		// Tier 2: Beats 9-16 (mid game)
		// Tier 3: Beats 17-24 (late game)
		beatRange := [2]int{1, 8}
		switch tier {
		case 1:
			beatRange = [2]int{1, 8}
		case 2:
			beatRange = [2]int{9, 16}
		case 3:
			beatRange = [2]int{17, 24}
		}

		tieredHints = append(tieredHints, ClueHint{
			Tier:      tier,
			BeatRange: beatRange,
			Hint:      hint,
			Revealed:  false,
		})
	}

	return tieredHints
}

// GetTemplatesPath returns the path to the rules templates directory.
func GetTemplatesPath(baseDir string) string {
	return filepath.Join(baseDir, "templates", "rules")
}
