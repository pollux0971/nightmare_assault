package templates

import (
	"math/rand"
	"path/filepath"
	"sync"
)

// TemplateLibrary manages all game templates
type TemplateLibrary struct {
	loader         *TemplateLoader
	rules          map[RuleCategory][]*RuleTemplate
	scenes         map[SceneCategory][]*SceneTemplate
	npcs           []*NPCTemplate
	mu             sync.RWMutex
	isInitialized  bool
}

// NewTemplateLibrary creates a new template library
func NewTemplateLibrary(baseDir string) *TemplateLibrary {
	return &TemplateLibrary{
		loader: NewTemplateLoader(baseDir),
		rules:  make(map[RuleCategory][]*RuleTemplate),
		scenes: make(map[SceneCategory][]*SceneTemplate),
		npcs:   make([]*NPCTemplate, 0),
	}
}

// Initialize loads all templates from the templates directory
func (tl *TemplateLibrary) Initialize() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	// Load all rule templates
	ruleCategories := []RuleCategory{
		RuleCategorySensory,
		RuleCategorySpatial,
		RuleCategorySocial,
	}

	for _, category := range ruleCategories {
		filename := string(category) + ".yaml"
		err := tl.loadRuleCategory(category, filepath.Join("templates", "rules", filename))
		if err != nil {
			// Continue loading other categories even if one fails
			continue
		}
	}

	// Load all scene templates
	sceneCategories := []SceneCategory{
		SceneCategoryBiological,
		SceneCategoryTemporal,
		SceneCategoryDigital,
		SceneCategorySpatial,
	}

	for _, category := range sceneCategories {
		filename := string(category) + ".yaml"
		err := tl.loadSceneCategory(category, filepath.Join("templates", "scenes", filename))
		if err != nil {
			// Continue loading other categories even if one fails
			continue
		}
	}

	// Load NPC archetypes
	_ = tl.loadNPCArchetypes(filepath.Join("templates", "npcs", "archetypes.yaml"))

	tl.isInitialized = true
	return nil
}

// loadRuleCategory loads rules from a specific category file
func (tl *TemplateLibrary) loadRuleCategory(category RuleCategory, filePath string) error {
	var collection RuleTemplateCollection
	fullPath := filepath.Join(tl.loader.baseDir, filePath)

	err := tl.loader.LoadYAMLFile(fullPath, &collection)
	if err != nil {
		return err
	}

	tl.rules[category] = collection.Rules
	return nil
}

// loadSceneCategory loads scenes from a specific category file
func (tl *TemplateLibrary) loadSceneCategory(category SceneCategory, filePath string) error {
	var collection SceneTemplateCollection
	fullPath := filepath.Join(tl.loader.baseDir, filePath)

	err := tl.loader.LoadYAMLFile(fullPath, &collection)
	if err != nil {
		return err
	}

	tl.scenes[category] = collection.Scenes
	return nil
}

// loadNPCArchetypes loads NPC archetypes
func (tl *TemplateLibrary) loadNPCArchetypes(filePath string) error {
	var collection NPCTemplateCollection
	fullPath := filepath.Join(tl.loader.baseDir, filePath)

	err := tl.loader.LoadYAMLFile(fullPath, &collection)
	if err != nil {
		return err
	}

	tl.npcs = collection.NPCTypes
	return nil
}

// GetRulesByCategory returns all rules in a specific category
func (tl *TemplateLibrary) GetRulesByCategory(category RuleCategory) []*RuleTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	rules, exists := tl.rules[category]
	if !exists {
		return []*RuleTemplate{}
	}

	// Return a copy to prevent external modification
	result := make([]*RuleTemplate, len(rules))
	copy(result, rules)
	return result
}

// GetScenesByCategory returns all scenes in a specific category
func (tl *TemplateLibrary) GetScenesByCategory(category SceneCategory) []*SceneTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	scenes, exists := tl.scenes[category]
	if !exists {
		return []*SceneTemplate{}
	}

	// Return a copy to prevent external modification
	result := make([]*SceneTemplate, len(scenes))
	copy(result, scenes)
	return result
}

// GetNPCArchetypes returns all NPC archetypes
func (tl *TemplateLibrary) GetNPCArchetypes() []*NPCTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*NPCTemplate, len(tl.npcs))
	copy(result, tl.npcs)
	return result
}

// SelectRandomRule selects a random rule with optional filtering
func (tl *TemplateLibrary) SelectRandomRule(category *RuleCategory, difficulty *RuleDifficulty, tag *string) *RuleTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	// Collect all matching rules
	candidates := make([]*RuleTemplate, 0)

	// Determine which categories to search
	categoriesToSearch := []RuleCategory{}
	if category != nil {
		categoriesToSearch = append(categoriesToSearch, *category)
	} else {
		// Search all categories
		for cat := range tl.rules {
			categoriesToSearch = append(categoriesToSearch, cat)
		}
	}

	// Filter rules
	for _, cat := range categoriesToSearch {
		rules := tl.rules[cat]
		for _, rule := range rules {
			// Check difficulty filter
			if difficulty != nil && rule.Difficulty != *difficulty {
				continue
			}

			// Check tag filter
			if tag != nil && !rule.HasTag(*tag) {
				continue
			}

			candidates = append(candidates, rule)
		}
	}

	// Return random rule from candidates
	if len(candidates) == 0 {
		return nil
	}

	return candidates[rand.Intn(len(candidates))]
}

// SelectRandomScene selects a random scene with optional filtering
func (tl *TemplateLibrary) SelectRandomScene(category *SceneCategory, tag *string) *SceneTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	// Collect all matching scenes
	candidates := make([]*SceneTemplate, 0)

	// Determine which categories to search
	categoriesToSearch := []SceneCategory{}
	if category != nil {
		categoriesToSearch = append(categoriesToSearch, *category)
	} else {
		// Search all categories
		for cat := range tl.scenes {
			categoriesToSearch = append(categoriesToSearch, cat)
		}
	}

	// Filter scenes
	for _, cat := range categoriesToSearch {
		scenes := tl.scenes[cat]
		for _, scene := range scenes {
			// Check tag filter
			if tag != nil && !scene.HasTag(*tag) {
				continue
			}

			candidates = append(candidates, scene)
		}
	}

	// Return random scene from candidates
	if len(candidates) == 0 {
		return nil
	}

	return candidates[rand.Intn(len(candidates))]
}

// SelectRandomNPC selects a random NPC with optional filtering
func (tl *TemplateLibrary) SelectRandomNPC(archetype *NPCArchetype, knowledgeable *bool, trustworthy *bool) *NPCTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	// Collect all matching NPCs
	candidates := make([]*NPCTemplate, 0)

	for _, npc := range tl.npcs {
		// Check archetype filter
		if archetype != nil && npc.Archetype != *archetype {
			continue
		}

		// Check knowledgeable filter
		if knowledgeable != nil && npc.IsKnowledgeable() != *knowledgeable {
			continue
		}

		// Check trustworthy filter
		if trustworthy != nil && npc.IsTrustworthy() != *trustworthy {
			continue
		}

		candidates = append(candidates, npc)
	}

	// Return random NPC from candidates
	if len(candidates) == 0 {
		return nil
	}

	return candidates[rand.Intn(len(candidates))]
}

// ThemeBundle represents a theme-consistent combination of templates
type ThemeBundle struct {
	Scene *SceneTemplate
	Rules []*RuleTemplate
	NPCs  []*NPCTemplate
}

// SelectThemeBundle selects theme-consistent templates for a coherent game experience
func (tl *TemplateLibrary) SelectThemeBundle(sceneCategory SceneCategory, ruleCount int, npcCount int) *ThemeBundle {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	bundle := &ThemeBundle{
		Rules: make([]*RuleTemplate, 0, ruleCount),
		NPCs:  make([]*NPCTemplate, 0, npcCount),
	}

	// Select a scene from the specified category
	scenes := tl.scenes[sceneCategory]
	if len(scenes) == 0 {
		return nil
	}
	bundle.Scene = scenes[rand.Intn(len(scenes))]

	// Select rules that match the scene's theme
	// For biological scenes, prefer sensory rules
	// For spatial scenes, prefer spatial rules
	// For temporal/digital scenes, mix of categories
	var preferredRuleCategory RuleCategory
	switch sceneCategory {
	case SceneCategoryBiological:
		preferredRuleCategory = RuleCategorySensory
	case SceneCategorySpatial:
		preferredRuleCategory = RuleCategorySpatial
	case SceneCategoryTemporal, SceneCategoryDigital:
		// Mix categories for these
		allCategories := []RuleCategory{RuleCategorySensory, RuleCategorySpatial, RuleCategorySocial}
		preferredRuleCategory = allCategories[rand.Intn(len(allCategories))]
	}

	// Get rules from preferred category
	preferredRules := tl.rules[preferredRuleCategory]
	if len(preferredRules) > 0 {
		// Randomly select rules without replacement
		selectedIndices := make(map[int]bool)
		for len(bundle.Rules) < ruleCount && len(selectedIndices) < len(preferredRules) {
			idx := rand.Intn(len(preferredRules))
			if !selectedIndices[idx] {
				selectedIndices[idx] = true
				bundle.Rules = append(bundle.Rules, preferredRules[idx])
			}
		}
	}

	// Select NPCs with varied archetypes
	if len(tl.npcs) > 0 {
		selectedIndices := make(map[int]bool)
		for len(bundle.NPCs) < npcCount && len(selectedIndices) < len(tl.npcs) {
			idx := rand.Intn(len(tl.npcs))
			if !selectedIndices[idx] {
				selectedIndices[idx] = true
				bundle.NPCs = append(bundle.NPCs, tl.npcs[idx])
			}
		}
	}

	return bundle
}

// GetAllRules returns all rules across all categories
func (tl *TemplateLibrary) GetAllRules() []*RuleTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	result := make([]*RuleTemplate, 0)
	for _, rules := range tl.rules {
		result = append(result, rules...)
	}
	return result
}

// GetAllScenes returns all scenes across all categories
func (tl *TemplateLibrary) GetAllScenes() []*SceneTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	result := make([]*SceneTemplate, 0)
	for _, scenes := range tl.scenes {
		result = append(result, scenes...)
	}
	return result
}

// GetRuleByID finds a rule by its ID
func (tl *TemplateLibrary) GetRuleByID(id string) *RuleTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	for _, rules := range tl.rules {
		for _, rule := range rules {
			if rule.ID == id {
				return rule
			}
		}
	}
	return nil
}

// GetSceneByID finds a scene by its ID
func (tl *TemplateLibrary) GetSceneByID(id string) *SceneTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	for _, scenes := range tl.scenes {
		for _, scene := range scenes {
			if scene.ID == id {
				return scene
			}
		}
	}
	return nil
}

// GetNPCByID finds an NPC by its ID
func (tl *TemplateLibrary) GetNPCByID(id string) *NPCTemplate {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	for _, npc := range tl.npcs {
		if npc.ID == id {
			return npc
		}
	}
	return nil
}

// HasErrors returns whether there were any errors during loading
func (tl *TemplateLibrary) HasErrors() bool {
	return tl.loader.HasErrors()
}

// GetErrors returns all errors that occurred during loading
func (tl *TemplateLibrary) GetErrors() []*LoadError {
	return tl.loader.GetErrors()
}

// GetErrorSummary returns a formatted summary of all errors
func (tl *TemplateLibrary) GetErrorSummary() string {
	return tl.loader.GetErrorSummary()
}

// IsInitialized returns whether the library has been initialized
func (tl *TemplateLibrary) IsInitialized() bool {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return tl.isInitialized
}
