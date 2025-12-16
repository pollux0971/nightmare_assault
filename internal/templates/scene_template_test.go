package templates

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// Story 4.3 AC1: SceneTemplate structure has three-stage progression
func TestSceneTemplate_ThreeStageStructure(t *testing.T) {
	scene := &SceneTemplate{
		ID:       "test_001",
		Name:     "Test Scene",
		Category: SceneCategoryBiological,
		ApplicableAreas: []string{"hospital", "clinic"},
		Stage1: SceneStage{
			Description: "Stage 1: Daily dissonance",
			Atmosphere:  []string{"dark", "eerie"},
			CommonProps: []string{"prop1"},
			Hazards:     []string{"hazard1"},
		},
		Stage2: SceneStage{
			Description: "Stage 2: Significant anomaly",
			Atmosphere:  []string{"oppressive", "disturbing"},
			CommonProps: []string{"prop2"},
			Hazards:     []string{"hazard2"},
		},
		Stage3: SceneStage{
			Description: "Stage 3: Law collapse",
			Atmosphere:  []string{"nightmare", "chaos"},
			CommonProps: []string{"prop3"},
			Hazards:     []string{"hazard3"},
		},
		Tags:        []string{"test", "sample"},
		Description: "A test scene for validation",
	}

	// Verify basic fields
	if scene.ID != "test_001" {
		t.Error("ID field not working")
	}
	if scene.Name != "Test Scene" {
		t.Error("Name field not working")
	}
	if scene.Category != SceneCategoryBiological {
		t.Error("Category field not working")
	}
	if len(scene.ApplicableAreas) != 2 {
		t.Error("ApplicableAreas field not working")
	}

	// Verify Stage 1
	if scene.Stage1.Description != "Stage 1: Daily dissonance" {
		t.Error("Stage1 Description not working")
	}
	if len(scene.Stage1.Atmosphere) != 2 {
		t.Error("Stage1 Atmosphere not working")
	}

	// Verify Stage 2
	if scene.Stage2.Description != "Stage 2: Significant anomaly" {
		t.Error("Stage2 Description not working")
	}
	if len(scene.Stage2.Atmosphere) != 2 {
		t.Error("Stage2 Atmosphere not working")
	}

	// Verify Stage 3
	if scene.Stage3.Description != "Stage 3: Law collapse" {
		t.Error("Stage3 Description not working")
	}
	if len(scene.Stage3.Atmosphere) != 2 {
		t.Error("Stage3 Atmosphere not working")
	}

	// Verify tags and description
	if len(scene.Tags) != 2 {
		t.Error("Tags field not working")
	}
	if scene.Description != "A test scene for validation" {
		t.Error("Description field not working")
	}
}

// Story 4.3 AC2: Load biological.yaml with three-stage scenes
func TestSceneTemplate_LoadBiologicalYAML(t *testing.T) {
	// Find the project root (where go.mod is)
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	templatesDir := filepath.Join(projectRoot, "templates")
	biologicalPath := filepath.Join(templatesDir, "scenes", "biological.yaml")

	// Read the file
	data, err := os.ReadFile(biologicalPath)
	if err != nil {
		t.Fatalf("Failed to read biological.yaml: %v", err)
	}

	// Parse YAML
	var collection SceneTemplateCollection
	err = yaml.Unmarshal(data, &collection)
	if err != nil {
		t.Fatalf("Failed to parse biological.yaml: %v", err)
	}

	// AC: Should contain at least 3 scenes
	if len(collection.Scenes) < 3 {
		t.Errorf("Expected at least 3 scenes in biological.yaml, got %d", len(collection.Scenes))
	}

	// AC: Category should be biological
	if collection.Category != SceneCategoryBiological {
		t.Errorf("Expected category 'biological', got '%s'", collection.Category)
	}

	// AC: Each scene should have three stages with descriptions and atmosphere
	for i, scene := range collection.Scenes {
		// Check Stage 1
		if scene.Stage1.Description == "" {
			t.Errorf("Scene %d (%s): Stage1 Description is empty", i, scene.ID)
		}
		if len(scene.Stage1.Atmosphere) == 0 {
			t.Errorf("Scene %d (%s): Stage1 Atmosphere is empty", i, scene.ID)
		}

		// Check Stage 2
		if scene.Stage2.Description == "" {
			t.Errorf("Scene %d (%s): Stage2 Description is empty", i, scene.ID)
		}
		if len(scene.Stage2.Atmosphere) == 0 {
			t.Errorf("Scene %d (%s): Stage2 Atmosphere is empty", i, scene.ID)
		}

		// Check Stage 3
		if scene.Stage3.Description == "" {
			t.Errorf("Scene %d (%s): Stage3 Description is empty", i, scene.ID)
		}
		if len(scene.Stage3.Atmosphere) == 0 {
			t.Errorf("Scene %d (%s): Stage3 Atmosphere is empty", i, scene.ID)
		}

		// Validate the scene
		err := scene.Validate()
		if err != nil {
			t.Errorf("Scene %d (%s) validation failed: %v", i, scene.ID, err)
		}
	}
}

// Test Validate method with three-stage structure
func TestSceneTemplate_Validate(t *testing.T) {
	// Valid scene with all three stages
	validScene := &SceneTemplate{
		ID:   "test_001",
		Name: "Test",
		Stage1: SceneStage{
			Description: "Stage 1",
			Atmosphere:  []string{"dark"},
		},
		Stage2: SceneStage{
			Description: "Stage 2",
			Atmosphere:  []string{"eerie"},
		},
		Stage3: SceneStage{
			Description: "Stage 3",
			Atmosphere:  []string{"nightmare"},
		},
	}

	err := validScene.Validate()
	if err != nil {
		t.Errorf("Valid scene should pass validation, got error: %v", err)
	}

	// Missing ID
	invalidScene1 := &SceneTemplate{
		Name: "Test",
		Stage1: SceneStage{
			Description: "Stage 1",
			Atmosphere:  []string{"dark"},
		},
		Stage2: SceneStage{
			Description: "Stage 2",
			Atmosphere:  []string{"eerie"},
		},
		Stage3: SceneStage{
			Description: "Stage 3",
			Atmosphere:  []string{"nightmare"},
		},
	}

	err = invalidScene1.Validate()
	if err == nil {
		t.Error("Scene without ID should fail validation")
	}

	// Missing Name
	invalidScene2 := &SceneTemplate{
		ID: "test_001",
		Stage1: SceneStage{
			Description: "Stage 1",
			Atmosphere:  []string{"dark"},
		},
		Stage2: SceneStage{
			Description: "Stage 2",
			Atmosphere:  []string{"eerie"},
		},
		Stage3: SceneStage{
			Description: "Stage 3",
			Atmosphere:  []string{"nightmare"},
		},
	}

	err = invalidScene2.Validate()
	if err == nil {
		t.Error("Scene without Name should fail validation")
	}

	// Missing Stage1 Description
	invalidScene3 := &SceneTemplate{
		ID:   "test_001",
		Name: "Test",
		Stage1: SceneStage{
			Atmosphere: []string{"dark"},
		},
		Stage2: SceneStage{
			Description: "Stage 2",
			Atmosphere:  []string{"eerie"},
		},
		Stage3: SceneStage{
			Description: "Stage 3",
			Atmosphere:  []string{"nightmare"},
		},
	}

	err = invalidScene3.Validate()
	if err == nil {
		t.Error("Scene without Stage1 Description should fail validation")
	}

	// Empty Stage2 Atmosphere
	invalidScene4 := &SceneTemplate{
		ID:   "test_001",
		Name: "Test",
		Stage1: SceneStage{
			Description: "Stage 1",
			Atmosphere:  []string{"dark"},
		},
		Stage2: SceneStage{
			Description: "Stage 2",
			Atmosphere:  []string{},
		},
		Stage3: SceneStage{
			Description: "Stage 3",
			Atmosphere:  []string{"nightmare"},
		},
	}

	err = invalidScene4.Validate()
	if err == nil {
		t.Error("Scene with empty Stage2 Atmosphere should fail validation")
	}
}

// Test GetStage method
func TestSceneTemplate_GetStage(t *testing.T) {
	scene := &SceneTemplate{
		Stage1: SceneStage{Description: "Stage 1"},
		Stage2: SceneStage{Description: "Stage 2"},
		Stage3: SceneStage{Description: "Stage 3"},
	}

	// Test getting each stage
	stage1 := scene.GetStage(1)
	if stage1 == nil || stage1.Description != "Stage 1" {
		t.Error("GetStage(1) failed")
	}

	stage2 := scene.GetStage(2)
	if stage2 == nil || stage2.Description != "Stage 2" {
		t.Error("GetStage(2) failed")
	}

	stage3 := scene.GetStage(3)
	if stage3 == nil || stage3.Description != "Stage 3" {
		t.Error("GetStage(3) failed")
	}

	// Test invalid stage number
	stageInvalid := scene.GetStage(4)
	if stageInvalid != nil {
		t.Error("GetStage(4) should return nil")
	}

	stageZero := scene.GetStage(0)
	if stageZero != nil {
		t.Error("GetStage(0) should return nil")
	}
}

// Test HasTag
func TestSceneTemplate_HasTag(t *testing.T) {
	scene := &SceneTemplate{
		Tags: []string{"medical", "horror", "biological"},
	}

	if !scene.HasTag("medical") {
		t.Error("Should have 'medical' tag")
	}

	if !scene.HasTag("horror") {
		t.Error("Should have 'horror' tag")
	}

	if scene.HasTag("nonexistent") {
		t.Error("Should not have 'nonexistent' tag")
	}
}

// Test HasAtmosphere across all stages
func TestSceneTemplate_HasAtmosphere(t *testing.T) {
	scene := &SceneTemplate{
		Stage1: SceneStage{
			Atmosphere: []string{"dark", "eerie"},
		},
		Stage2: SceneStage{
			Atmosphere: []string{"oppressive"},
		},
		Stage3: SceneStage{
			Atmosphere: []string{"nightmare"},
		},
	}

	// Should find atmosphere from Stage 1
	if !scene.HasAtmosphere("dark") {
		t.Error("Should have 'dark' atmosphere from Stage1")
	}

	if !scene.HasAtmosphere("eerie") {
		t.Error("Should have 'eerie' atmosphere from Stage1")
	}

	// Should find atmosphere from Stage 2
	if !scene.HasAtmosphere("oppressive") {
		t.Error("Should have 'oppressive' atmosphere from Stage2")
	}

	// Should find atmosphere from Stage 3
	if !scene.HasAtmosphere("nightmare") {
		t.Error("Should have 'nightmare' atmosphere from Stage3")
	}

	// Should not find non-existent atmosphere
	if scene.HasAtmosphere("bright") {
		t.Error("Should not have 'bright' atmosphere")
	}
}

// Test GetAtmosphereKeywords across all stages
func TestSceneTemplate_GetAtmosphereKeywords(t *testing.T) {
	scene := &SceneTemplate{
		Stage1: SceneStage{
			Atmosphere: []string{"dark"},
		},
		Stage2: SceneStage{
			Atmosphere: []string{"eerie"},
		},
		Stage3: SceneStage{
			Atmosphere: []string{"nightmare"},
		},
	}

	keywords := scene.GetAtmosphereKeywords()
	// Should contain keywords from all three stages
	if keywords != "dark, eerie, nightmare" {
		t.Errorf("Expected 'dark, eerie, nightmare', got '%s'", keywords)
	}

	// Test with multiple keywords per stage
	scene2 := &SceneTemplate{
		Stage1: SceneStage{
			Atmosphere: []string{"dark", "cold"},
		},
		Stage2: SceneStage{
			Atmosphere: []string{"eerie"},
		},
		Stage3: SceneStage{
			Atmosphere: []string{},
		},
	}

	keywords2 := scene2.GetAtmosphereKeywords()
	if keywords2 != "dark, cold, eerie" {
		t.Errorf("Expected 'dark, cold, eerie', got '%s'", keywords2)
	}
}

// Test CountElements across all stages
func TestSceneTemplate_CountElements(t *testing.T) {
	scene := &SceneTemplate{
		Stage1: SceneStage{
			CommonProps: []string{"prop1", "prop2"},
			Hazards:     []string{"hazard1"},
		},
		Stage2: SceneStage{
			CommonProps: []string{"prop3"},
			Hazards:     []string{"hazard2", "hazard3"},
		},
		Stage3: SceneStage{
			CommonProps: []string{"prop4"},
			Hazards:     []string{"hazard4"},
		},
	}

	count := scene.CountElements()
	// 2+1 (Stage1) + 1+2 (Stage2) + 1+1 (Stage3) = 8
	if count != 8 {
		t.Errorf("Expected count 8, got %d", count)
	}

	// Test with empty stages
	scene2 := &SceneTemplate{
		Stage1: SceneStage{},
		Stage2: SceneStage{},
		Stage3: SceneStage{},
	}

	count2 := scene2.CountElements()
	if count2 != 0 {
		t.Errorf("Expected count 0, got %d", count2)
	}
}

// Test HasApplicableArea
func TestSceneTemplate_HasApplicableArea(t *testing.T) {
	scene := &SceneTemplate{
		ApplicableAreas: []string{"hospital", "clinic", "laboratory"},
	}

	if !scene.HasApplicableArea("hospital") {
		t.Error("Should have 'hospital' applicable area")
	}

	if !scene.HasApplicableArea("clinic") {
		t.Error("Should have 'clinic' applicable area")
	}

	if scene.HasApplicableArea("school") {
		t.Error("Should not have 'school' applicable area")
	}
}

// Test SceneCategory constants
func TestSceneCategory_Constants(t *testing.T) {
	if SceneCategoryBiological != "biological" {
		t.Error("SceneCategoryBiological constant incorrect")
	}
	if SceneCategoryTemporal != "temporal" {
		t.Error("SceneCategoryTemporal constant incorrect")
	}
	if SceneCategoryDigital != "digital" {
		t.Error("SceneCategoryDigital constant incorrect")
	}
	if SceneCategorySpatial != "spatial" {
		t.Error("SceneCategorySpatial constant incorrect")
	}
}

// Test YAML marshaling/unmarshaling with three-stage structure
func TestSceneTemplate_YAMLSerialization(t *testing.T) {
	original := &SceneTemplate{
		ID:              "test_yaml_001",
		Name:            "YAML Test Scene",
		Category:        SceneCategoryTemporal,
		ApplicableAreas: []string{"school", "clock tower"},
		Stage1: SceneStage{
			Description: "Clocks show different times",
			Atmosphere:  []string{"disorienting", "unstable"},
			CommonProps: []string{"clock", "calendar"},
			Hazards:     []string{"temporal anomaly"},
		},
		Stage2: SceneStage{
			Description: "See past versions of yourself",
			Atmosphere:  []string{"surreal", "uncanny"},
			CommonProps: []string{"mirror", "photo"},
			Hazards:     []string{"time loop"},
		},
		Stage3: SceneStage{
			Description: "Rapid aging and decay",
			Atmosphere:  []string{"horrifying", "inevitable"},
			CommonProps: []string{"bones", "dust"},
			Hazards:     []string{"temporal collapse"},
		},
		Tags:        []string{"test", "yaml", "time"},
		Description: "A test scene for YAML serialization",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored SceneTemplate
	err = yaml.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify fields
	if restored.ID != original.ID {
		t.Error("ID mismatch after serialization")
	}
	if restored.Name != original.Name {
		t.Error("Name mismatch after serialization")
	}
	if restored.Category != original.Category {
		t.Error("Category mismatch after serialization")
	}
	if len(restored.ApplicableAreas) != len(original.ApplicableAreas) {
		t.Error("ApplicableAreas length mismatch after serialization")
	}

	// Verify Stage 1
	if restored.Stage1.Description != original.Stage1.Description {
		t.Error("Stage1 Description mismatch after serialization")
	}
	if len(restored.Stage1.Atmosphere) != len(original.Stage1.Atmosphere) {
		t.Error("Stage1 Atmosphere length mismatch after serialization")
	}

	// Verify Stage 2
	if restored.Stage2.Description != original.Stage2.Description {
		t.Error("Stage2 Description mismatch after serialization")
	}

	// Verify Stage 3
	if restored.Stage3.Description != original.Stage3.Description {
		t.Error("Stage3 Description mismatch after serialization")
	}

	if restored.Description != original.Description {
		t.Error("Description mismatch after serialization")
	}
}

// Test loading SceneTemplateCollection with three-stage structure
func TestSceneTemplateCollection_Load(t *testing.T) {
	yamlContent := `
version: "1.0"
category: biological
scenes:
  - id: test_001
    name: Test Hospital
    category: biological
    applicable_areas:
      - hospital
      - clinic
    stage1:
      description: Mold patterns like blood vessels
      atmosphere:
        - dark
        - eerie
      common_props:
        - surgical table
      hazards:
        - mold
    stage2:
      description: Walls become warm like skin
      atmosphere:
        - warm
        - disturbing
      common_props:
        - bone-like doorknobs
      hazards:
        - living walls
    stage3:
      description: Room becomes a stomach
      atmosphere:
        - suffocating
        - digestive
      common_props:
        - gastric acid
      hazards:
        - acid pools
    tags:
      - medical
      - horror
    description: A test hospital scene
`

	var collection SceneTemplateCollection
	err := yaml.Unmarshal([]byte(yamlContent), &collection)
	if err != nil {
		t.Fatalf("Failed to unmarshal collection: %v", err)
	}

	if collection.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", collection.Version)
	}

	if collection.Category != SceneCategoryBiological {
		t.Errorf("Expected category 'biological', got '%s'", collection.Category)
	}

	if len(collection.Scenes) != 1 {
		t.Errorf("Expected 1 scene, got %d", len(collection.Scenes))
	}

	// Validate the scene
	scene := collection.Scenes[0]

	// Check basic fields
	if scene.ID != "test_001" {
		t.Error("Scene ID incorrect")
	}

	if len(scene.ApplicableAreas) != 2 {
		t.Error("ApplicableAreas length incorrect")
	}

	// Check all three stages
	if scene.Stage1.Description != "Mold patterns like blood vessels" {
		t.Error("Stage1 Description incorrect")
	}
	if scene.Stage2.Description != "Walls become warm like skin" {
		t.Error("Stage2 Description incorrect")
	}
	if scene.Stage3.Description != "Room becomes a stomach" {
		t.Error("Stage3 Description incorrect")
	}

	// Validate
	err = scene.Validate()
	if err != nil {
		t.Errorf("Scene validation failed: %v", err)
	}
}
