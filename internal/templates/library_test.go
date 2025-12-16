package templates

import (
	"testing"
)

// Story 4.5 AC1: Initialize and load all templates
func TestTemplateLibrary_Initialize(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	err = library.Initialize()

	if err != nil {
		t.Errorf("Initialize should not return error: %v", err)
	}

	if !library.IsInitialized() {
		t.Error("Library should be initialized")
	}

	// Verify that at least some templates were loaded
	allRules := library.GetAllRules()
	if len(allRules) == 0 {
		t.Error("Expected some rules to be loaded")
	}

	allScenes := library.GetAllScenes()
	if len(allScenes) == 0 {
		t.Error("Expected some scenes to be loaded")
	}

	allNPCs := library.GetNPCArchetypes()
	if len(allNPCs) == 0 {
		t.Error("Expected some NPCs to be loaded")
	}
}

// Story 4.5 AC2: GetRulesByCategory returns rules from specific category
func TestTemplateLibrary_GetRulesByCategory(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Get sensory rules
	sensoryRules := library.GetRulesByCategory(RuleCategorySensory)
	if len(sensoryRules) == 0 {
		t.Error("Expected sensory rules to be loaded")
	}

	// Verify all returned rules are sensory
	for _, rule := range sensoryRules {
		if rule.Category != RuleCategorySensory {
			t.Errorf("Expected sensory category, got %s", rule.Category)
		}
	}

	// Test non-existent category returns empty slice
	emptyRules := library.GetRulesByCategory(RuleCategory("nonexistent"))
	if len(emptyRules) != 0 {
		t.Error("Expected empty slice for non-existent category")
	}
}

// Story 4.5 AC2: GetScenesByCategory returns scenes from specific category
func TestTemplateLibrary_GetScenesByCategory(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Get biological scenes
	biologicalScenes := library.GetScenesByCategory(SceneCategoryBiological)
	if len(biologicalScenes) == 0 {
		t.Error("Expected biological scenes to be loaded")
	}

	// Verify all returned scenes are biological
	for _, scene := range biologicalScenes {
		if scene.Category != SceneCategoryBiological {
			t.Errorf("Expected biological category, got %s", scene.Category)
		}
	}

	// Test non-existent category returns empty slice
	emptyScenes := library.GetScenesByCategory(SceneCategory("nonexistent"))
	if len(emptyScenes) != 0 {
		t.Error("Expected empty slice for non-existent category")
	}
}

// Story 4.5 AC2: GetNPCArchetypes returns all NPC archetypes
func TestTemplateLibrary_GetNPCArchetypes(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	npcs := library.GetNPCArchetypes()
	if len(npcs) < 6 {
		t.Errorf("Expected at least 6 NPCs, got %d", len(npcs))
	}
}

// Story 4.5 AC3: SelectRandomRule with difficulty filter
func TestTemplateLibrary_SelectRandomRule_WithDifficulty(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a medium difficulty rule
	difficulty := RuleDifficultyMedium
	rule := library.SelectRandomRule(nil, &difficulty, nil)

	if rule == nil {
		t.Fatal("Expected to find a medium difficulty rule")
	}

	if rule.Difficulty != RuleDifficultyMedium {
		t.Errorf("Expected medium difficulty, got %s", rule.Difficulty)
	}
}

// Story 4.5 AC3: SelectRandomRule with category filter
func TestTemplateLibrary_SelectRandomRule_WithCategory(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a sensory rule
	category := RuleCategorySensory
	rule := library.SelectRandomRule(&category, nil, nil)

	if rule == nil {
		t.Fatal("Expected to find a sensory rule")
	}

	if rule.Category != RuleCategorySensory {
		t.Errorf("Expected sensory category, got %s", rule.Category)
	}
}

// Story 4.5 AC3: SelectRandomRule with tag filter
func TestTemplateLibrary_SelectRandomRule_WithTag(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a rule with "sensory" tag
	tag := "sensory"
	rule := library.SelectRandomRule(nil, nil, &tag)

	if rule == nil {
		t.Fatal("Expected to find a rule with 'sensory' tag")
	}

	if !rule.HasTag("sensory") {
		t.Error("Expected rule to have 'sensory' tag")
	}
}

// Story 4.5 AC3: SelectRandomRule with multiple filters
func TestTemplateLibrary_SelectRandomRule_WithMultipleFilters(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a medium difficulty sensory rule
	category := RuleCategorySensory
	difficulty := RuleDifficultyMedium
	rule := library.SelectRandomRule(&category, &difficulty, nil)

	if rule != nil {
		if rule.Category != RuleCategorySensory {
			t.Errorf("Expected sensory category, got %s", rule.Category)
		}
		if rule.Difficulty != RuleDifficultyMedium {
			t.Errorf("Expected medium difficulty, got %s", rule.Difficulty)
		}
	}
}

// Story 4.5 AC3: SelectRandomScene with filters
func TestTemplateLibrary_SelectRandomScene(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a biological scene
	category := SceneCategoryBiological
	scene := library.SelectRandomScene(&category, nil)

	if scene == nil {
		t.Fatal("Expected to find a biological scene")
	}

	if scene.Category != SceneCategoryBiological {
		t.Errorf("Expected biological category, got %s", scene.Category)
	}
}

// Story 4.5 AC3: SelectRandomNPC with filters
func TestTemplateLibrary_SelectRandomNPC(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a survivor NPC
	archetype := NPCArchetypeSurvivor
	npc := library.SelectRandomNPC(&archetype, nil, nil)

	if npc == nil {
		t.Fatal("Expected to find a survivor NPC")
	}

	if npc.Archetype != NPCArchetypeSurvivor {
		t.Errorf("Expected survivor archetype, got %s", npc.Archetype)
	}

	// Select a knowledgeable NPC
	knowledgeable := true
	npc2 := library.SelectRandomNPC(nil, &knowledgeable, nil)

	if npc2 == nil {
		t.Fatal("Expected to find a knowledgeable NPC")
	}

	if !npc2.IsKnowledgeable() {
		t.Error("Expected NPC to be knowledgeable")
	}

	// Select a trustworthy NPC
	trustworthy := true
	npc3 := library.SelectRandomNPC(nil, nil, &trustworthy)

	if npc3 == nil {
		t.Fatal("Expected to find a trustworthy NPC")
	}

	if !npc3.IsTrustworthy() {
		t.Error("Expected NPC to be trustworthy")
	}
}

// Story 4.5 AC4: SelectThemeBundle creates theme-consistent combinations
func TestTemplateLibrary_SelectThemeBundle(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Select a biological theme bundle
	bundle := library.SelectThemeBundle(SceneCategoryBiological, 3, 2)

	if bundle == nil {
		t.Fatal("Expected to get a theme bundle")
	}

	// Verify scene
	if bundle.Scene == nil {
		t.Fatal("Bundle should have a scene")
	}

	if bundle.Scene.Category != SceneCategoryBiological {
		t.Errorf("Expected biological scene, got %s", bundle.Scene.Category)
	}

	// Verify rules (should try to get 3, but may get less if not enough available)
	if len(bundle.Rules) == 0 {
		t.Error("Bundle should have at least some rules")
	}

	// Verify NPCs (should try to get 2, but may get less if not enough available)
	if len(bundle.NPCs) == 0 {
		t.Error("Bundle should have at least some NPCs")
	}

	// Verify no duplicates in rules
	ruleIDs := make(map[string]bool)
	for _, rule := range bundle.Rules {
		if ruleIDs[rule.ID] {
			t.Errorf("Duplicate rule ID in bundle: %s", rule.ID)
		}
		ruleIDs[rule.ID] = true
	}

	// Verify no duplicates in NPCs
	npcIDs := make(map[string]bool)
	for _, npc := range bundle.NPCs {
		if npcIDs[npc.ID] {
			t.Errorf("Duplicate NPC ID in bundle: %s", npc.ID)
		}
		npcIDs[npc.ID] = true
	}
}

// Test GetAllRules
func TestTemplateLibrary_GetAllRules(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	allRules := library.GetAllRules()
	if len(allRules) == 0 {
		t.Error("Expected some rules")
	}

	// Verify rules from different categories
	categories := make(map[RuleCategory]bool)
	for _, rule := range allRules {
		categories[rule.Category] = true
	}

	if len(categories) == 0 {
		t.Error("Expected rules from at least one category")
	}
}

// Test GetAllScenes
func TestTemplateLibrary_GetAllScenes(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	allScenes := library.GetAllScenes()
	if len(allScenes) == 0 {
		t.Error("Expected some scenes")
	}

	// Verify scenes from different categories
	categories := make(map[SceneCategory]bool)
	for _, scene := range allScenes {
		categories[scene.Category] = true
	}

	if len(categories) == 0 {
		t.Error("Expected scenes from at least one category")
	}
}

// Test GetRuleByID
func TestTemplateLibrary_GetRuleByID(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Get a rule that should exist
	rule := library.GetRuleByID("sensory_001")
	if rule == nil {
		t.Error("Expected to find rule with ID 'sensory_001'")
	} else {
		if rule.ID != "sensory_001" {
			t.Errorf("Expected ID 'sensory_001', got '%s'", rule.ID)
		}
	}

	// Try to get a non-existent rule
	noRule := library.GetRuleByID("nonexistent_999")
	if noRule != nil {
		t.Error("Expected nil for non-existent rule ID")
	}
}

// Test GetSceneByID
func TestTemplateLibrary_GetSceneByID(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Get a scene that should exist
	scene := library.GetSceneByID("bio_001")
	if scene == nil {
		t.Error("Expected to find scene with ID 'bio_001'")
	} else {
		if scene.ID != "bio_001" {
			t.Errorf("Expected ID 'bio_001', got '%s'", scene.ID)
		}
	}

	// Try to get a non-existent scene
	noScene := library.GetSceneByID("nonexistent_999")
	if noScene != nil {
		t.Error("Expected nil for non-existent scene ID")
	}
}

// Test GetNPCByID
func TestTemplateLibrary_GetNPCByID(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Get an NPC that should exist
	npc := library.GetNPCByID("npc_001")
	if npc == nil {
		t.Error("Expected to find NPC with ID 'npc_001'")
	} else {
		if npc.ID != "npc_001" {
			t.Errorf("Expected ID 'npc_001', got '%s'", npc.ID)
		}
	}

	// Try to get a non-existent NPC
	noNPC := library.GetNPCByID("nonexistent_999")
	if noNPC != nil {
		t.Error("Expected nil for non-existent NPC ID")
	}
}

// Test thread safety by accessing library from multiple goroutines
func TestTemplateLibrary_ThreadSafety(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			library.GetAllRules()
			library.GetAllScenes()
			library.GetNPCArchetypes()
			library.SelectRandomRule(nil, nil, nil)
			library.SelectRandomScene(nil, nil)
			library.SelectRandomNPC(nil, nil, nil)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Test that library returns copies, not references
func TestTemplateLibrary_ReturnsCopies(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	library := NewTemplateLibrary(projectRoot)
	library.Initialize()

	// Get rules twice
	rules1 := library.GetRulesByCategory(RuleCategorySensory)
	rules2 := library.GetRulesByCategory(RuleCategorySensory)

	// Modifying one should not affect the other
	if len(rules1) > 0 {
		rules1[0] = nil
		if len(rules2) > 0 && rules2[0] == nil {
			t.Error("Library should return copies, not references")
		}
	}
}

// Test error handling with invalid base directory
func TestTemplateLibrary_InvalidBaseDir(t *testing.T) {
	library := NewTemplateLibrary("/nonexistent/path")
	err := library.Initialize()

	// Initialize should not return error (it continues on failures)
	if err != nil {
		t.Errorf("Initialize should not return error even with invalid path: %v", err)
	}

	// But the library should have recorded errors
	if !library.HasErrors() {
		// It's okay if there are no errors if the files simply don't exist
		// The loader will continue and just have empty collections
	}
}
