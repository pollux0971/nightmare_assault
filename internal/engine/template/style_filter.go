package template

import (
	"fmt"
	"log"
	"strings"
)

// StyleFilter performs theme adaptation and template matching.
//
// Story 7.1 AC #6: Template library adaptability check
// The filter ensures templates match the player's chosen theme:
//   - Perfect match (score ≥ 0.8): Use template directly
//   - Partial match (score ≥ 0.4): Keep core constraints, rewrite appearance/background
//   - No match (score < 0.4): Trigger improvisation mechanism (generate new entity following difficulty constraints)
//
// Key responsibilities:
//   - Tag-based matching between theme and templates
//   - Match score calculation
//   - Entity adaptation (rewriting for theme consistency)
//   - Entity improvisation (generating new entities when no match found)
//   - Quality checklist validation (ensuring entities have clear Constraints)
type StyleFilter struct {
	// Future: Add configuration options
}

// NewStyleFilter creates a new StyleFilter.
func NewStyleFilter() *StyleFilter {
	return &StyleFilter{}
}

// MatchScore represents the compatibility score between theme and template.
type MatchScore struct {
	Score       float64  // 0.0 - 1.0
	Reasoning   string   // Why this score
	MatchedTags []string // Tags that matched
	MissingTags []string // Tags from theme not found in template
}

// CalculateMatchScore calculates the match score between a theme and entity tags.
//
// Story 7.1 AC #6: Match score calculation
//   - Perfect match: score ≥ 0.8 (most theme tags present in entity)
//   - Partial match: score ≥ 0.4 (some theme tags present)
//   - No match: score < 0.4 (few or no theme tags present)
//
// Parameters:
//   - themeTags: Tags extracted from player's theme (e.g., ["hospital", "medical", "abandoned"])
//   - entityTags: Tags associated with the template entity (e.g., ["zombie", "physical", "blind"])
//
// Returns:
//   - *MatchScore: Score and detailed reasoning
func (sf *StyleFilter) CalculateMatchScore(themeTags []string, entityTags []string) *MatchScore {
	if len(themeTags) == 0 {
		return &MatchScore{
			Score:     0.5, // Default neutral score if no theme tags
			Reasoning: "No theme tags provided, using default score",
		}
	}

	// Normalize tags to lowercase for case-insensitive matching
	normalizedTheme := normalizeTags(themeTags)
	normalizedEntity := normalizeTags(entityTags)

	// Find matching tags
	matchedTags := []string{}
	for _, themeTag := range normalizedTheme {
		for _, entityTag := range normalizedEntity {
			if themeTag == entityTag {
				matchedTags = append(matchedTags, themeTag)
				break
			}
			// Also check for partial matches (e.g., "hospital" matches "medical")
			if strings.Contains(themeTag, entityTag) || strings.Contains(entityTag, themeTag) {
				matchedTags = append(matchedTags, themeTag)
				break
			}
		}
	}

	// Find missing tags
	missingTags := []string{}
	for _, themeTag := range normalizedTheme {
		found := false
		for _, matched := range matchedTags {
			if matched == themeTag {
				found = true
				break
			}
		}
		if !found {
			missingTags = append(missingTags, themeTag)
		}
	}

	// Calculate score
	matchRatio := float64(len(matchedTags)) / float64(len(normalizedTheme))

	// Apply weighting: exact matches count more than partial
	score := matchRatio

	var reasoning string
	if score >= 0.8 {
		reasoning = fmt.Sprintf("Perfect match: %d/%d theme tags matched", len(matchedTags), len(normalizedTheme))
	} else if score >= 0.4 {
		reasoning = fmt.Sprintf("Partial match: %d/%d theme tags matched", len(matchedTags), len(normalizedTheme))
	} else {
		reasoning = fmt.Sprintf("Poor match: only %d/%d theme tags matched", len(matchedTags), len(normalizedTheme))
	}

	return &MatchScore{
		Score:       score,
		Reasoning:   reasoning,
		MatchedTags: matchedTags,
		MissingTags: missingTags,
	}
}

// AdaptEntity adapts a template entity to match the theme.
//
// Story 7.1 AC #6: Entity adaptation for partial matches (score ≥ 0.4)
// Keeps core Constraints (absolute limits that ensure player survival)
// Rewrites appearance and background to match theme
//
// Example:
//   Template: "Zombie" (blind, slow, can't open doors)
//   Theme: "Cyberpunk city"
//   Result: "Failed cyborg" (visual sensors destroyed, servo motors damaged, can't bypass electronic locks)
//
// Parameters:
//   - entity: Original template entity
//   - theme: Player's theme
//   - themeTags: Tags from the theme
//
// Returns:
//   - Adapted entity structure
//   - error: If adaptation fails
func (sf *StyleFilter) AdaptEntity(entity map[string]interface{}, theme string, themeTags []string) (map[string]interface{}, error) {
	adapted := make(map[string]interface{})

	// Copy all fields from original entity
	for k, v := range entity {
		adapted[k] = v
	}

	// Preserve core constraints (these are absolute limits)
	if constraints, ok := entity["constraints"]; ok {
		adapted["constraints"] = constraints
		log.Printf("[StyleFilter] Preserved constraints from original entity")
	}

	// Rewrite description to incorporate theme
	if description, ok := entity["description"].(string); ok {
		adapted["description"] = fmt.Sprintf("%s (adapted for theme: %s)", description, theme)
	}

	// Update tags to include theme tags
	if entityTags, ok := entity["tags"].([]string); ok {
		combinedTags := append(entityTags, themeTags...)
		adapted["tags"] = combinedTags
	} else {
		adapted["tags"] = themeTags
	}

	log.Printf("[StyleFilter] Adapted entity for theme: %s", theme)
	return adapted, nil
}

// ImproviseEntity generates a new entity when no suitable template matches.
//
// Story 7.1 AC #6: Improvisation mechanism for poor matches (score < 0.4)
// When no template matches the theme well, generate a new entity that:
//   - Follows the theme aesthetically
//   - Respects difficulty constraints (ability caps from difficulty tables)
//   - Has clear Constraints (absolute limits that provide counterplay)
//
// Example:
//   Theme: "Cyberpunk city digital ghost"
//   Difficulty: "easy"
//   Result: {
//     ID: "E-CUSTOM-01",
//     Name: "Rogue AI Enforcer",
//     Abilities: {Perception: "Infrared scanning", Movement: "Fast servo motors"},
//     Constraints: ["Cannot detect targets behind concrete walls", "Overheats after 30s pursuit"],
//     Counterplay: ["Hide behind walls", "Wait for overheat cooldown"]
//   }
//
// Parameters:
//   - theme: Player's theme
//   - themeTags: Tags extracted from theme
//   - difficulty: Difficulty level (easy/hard/hell)
//
// Returns:
//   - Improvised entity structure
//   - error: If improvisation fails
func (sf *StyleFilter) ImproviseEntity(theme string, themeTags []string, difficulty string) (map[string]interface{}, error) {
	entity := make(map[string]interface{})

	// Generate unique ID
	entity["id"] = "E-CUSTOM-01" // In real implementation, use UUID or counter

	// Generate name based on theme
	entity["name"] = fmt.Sprintf("Themed Entity (%s)", theme)

	// Add theme tags
	entity["tags"] = themeTags

	// Set difficulty
	entity["difficulty"] = difficulty

	// Generate abilities based on difficulty
	abilities := map[string]string{}
	switch difficulty {
	case "easy":
		abilities["perception"] = "Basic sensory capability"
		abilities["movement"] = "Slow movement"
		abilities["attack"] = "Low damage"
	case "hard":
		abilities["perception"] = "Enhanced sensory capability"
		abilities["movement"] = "Fast movement"
		abilities["attack"] = "Medium damage"
	case "hell":
		abilities["perception"] = "Advanced sensory capability"
		abilities["movement"] = "Very fast movement"
		abilities["attack"] = "High damage"
	default:
		abilities["perception"] = "Standard sensory capability"
		abilities["movement"] = "Normal movement"
		abilities["attack"] = "Standard damage"
	}
	entity["abilities"] = abilities

	// Generate constraints (CRITICAL: these provide player counterplay)
	constraints := []string{
		"Cannot detect perfectly still targets",
		"Cannot pass through solid barriers",
		"Requires line of sight for tracking",
	}
	entity["constraints"] = constraints

	// Generate counterplay strategies
	counterplay := []string{
		"Remain perfectly still when detected",
		"Use solid cover to break line of sight",
		"Create distractions to misdirect",
	}
	entity["counterplay"] = counterplay

	log.Printf("[StyleFilter] Improvised new entity for theme: %s, difficulty: %s", theme, difficulty)
	return entity, nil
}

// ExtractThemeTags extracts relevant tags from a theme string.
//
// This is a simple implementation that extracts keywords.
// In a production system, this might use NLP or LLM to extract semantic tags.
//
// Parameters:
//   - theme: Player's theme string (e.g., "廢棄醫院的午夜值班")
//
// Returns:
//   - []string: List of extracted tags (e.g., ["hospital", "medical", "abandoned", "night"])
func (sf *StyleFilter) ExtractThemeTags(theme string) []string {
	// Simple keyword extraction (in production, use LLM or NLP)
	keywords := map[string][]string{
		"醫院":   {"hospital", "medical"},
		"hospital": {"hospital", "medical"},
		"學校":   {"school", "education"},
		"school":   {"school", "education"},
		"旅館":   {"hotel", "lodging"},
		"hotel":    {"hotel", "lodging"},
		"溫泉":   {"hotspring", "bath"},
		"精神病院": {"asylum", "mental", "hospital"},
		"asylum":   {"asylum", "mental"},
		"廢棄":   {"abandoned", "ruined"},
		"abandoned": {"abandoned", "ruined"},
		"午夜":   {"midnight", "night"},
		"night":    {"night", "midnight"},
		"midnight": {"midnight", "night"},
		"森林":   {"forest", "nature"},
		"forest":   {"forest", "nature"},
		"城市":   {"city", "urban"},
		"city":     {"city", "urban"},
		"賽博": {"cyberpunk", "tech"},
		"cyber": {"cyberpunk", "tech"},
		"數位": {"digital", "virtual"},
		"digital": {"digital", "virtual"},
	}

	tags := []string{}
	themeLower := strings.ToLower(theme)

	for keyword, tagList := range keywords {
		if strings.Contains(themeLower, keyword) || strings.Contains(theme, keyword) {
			tags = append(tags, tagList...)
		}
	}

	// If no tags found, add a generic tag
	if len(tags) == 0 {
		tags = append(tags, "horror", "mystery")
	}

	// Remove duplicates
	uniqueTags := make(map[string]bool)
	result := []string{}
	for _, tag := range tags {
		if !uniqueTags[tag] {
			uniqueTags[tag] = true
			result = append(result, tag)
		}
	}

	log.Printf("[StyleFilter] Extracted %d tags from theme: %v", len(result), result)
	return result
}

// ValidateEntityConstraints validates that an entity has proper constraints.
//
// Story 7.1 AC #6: Quality checklist validation
// Every threat entity MUST have clear, absolute Constraints that provide counterplay.
// Without constraints, entities become unfair and unbeatable.
//
// Parameters:
//   - entity: Entity to validate
//
// Returns:
//   - error: If validation fails (missing or invalid constraints)
func (sf *StyleFilter) ValidateEntityConstraints(entity map[string]interface{}) error {
	constraints, ok := entity["constraints"]
	if !ok {
		return fmt.Errorf("entity missing 'constraints' field - must have absolute limits")
	}

	constraintList, ok := constraints.([]string)
	if !ok {
		return fmt.Errorf("constraints must be a list of strings")
	}

	if len(constraintList) == 0 {
		return fmt.Errorf("entity must have at least one constraint (absolute limit)")
	}

	// Validate that each constraint is meaningful (not empty)
	for i, constraint := range constraintList {
		if strings.TrimSpace(constraint) == "" {
			return fmt.Errorf("constraint %d is empty", i)
		}
	}

	log.Printf("[StyleFilter] Entity has %d valid constraints", len(constraintList))
	return nil
}

// Helper functions

// normalizeTags converts tags to lowercase and trims whitespace.
func normalizeTags(tags []string) []string {
	normalized := make([]string, len(tags))
	for i, tag := range tags {
		normalized[i] = strings.ToLower(strings.TrimSpace(tag))
	}
	return normalized
}
