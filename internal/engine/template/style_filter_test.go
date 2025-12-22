package template

import (
	"testing"
)

func TestCalculateMatchScore_PerfectMatch(t *testing.T) {
	sf := NewStyleFilter()

	themeTags := []string{"hospital", "medical", "abandoned"}
	entityTags := []string{"hospital", "medical", "abandoned", "zombie"}

	score := sf.CalculateMatchScore(themeTags, entityTags)

	if score.Score < 0.8 {
		t.Errorf("Expected perfect match (≥0.8), got %.2f", score.Score)
	}

	if len(score.MatchedTags) != 3 {
		t.Errorf("Expected 3 matched tags, got %d", len(score.MatchedTags))
	}

	if len(score.MissingTags) != 0 {
		t.Errorf("Expected 0 missing tags, got %d", len(score.MissingTags))
	}
}

func TestCalculateMatchScore_PartialMatch(t *testing.T) {
	sf := NewStyleFilter()

	// 2 out of 4 theme tags match = 0.5 (partial match)
	themeTags := []string{"hospital", "medical", "abandoned", "night"}
	entityTags := []string{"hospital", "medical", "zombie"}

	score := sf.CalculateMatchScore(themeTags, entityTags)

	if score.Score < 0.4 || score.Score >= 0.8 {
		t.Errorf("Expected partial match (0.4-0.8), got %.2f", score.Score)
	}

	if len(score.MatchedTags) == 0 {
		t.Error("Expected some matched tags, got none")
	}

	if len(score.MissingTags) == 0 {
		t.Error("Expected some missing tags, got none")
	}
}

func TestCalculateMatchScore_NoMatch(t *testing.T) {
	sf := NewStyleFilter()

	themeTags := []string{"cyberpunk", "digital", "tech"}
	entityTags := []string{"zombie", "undead", "physical"}

	score := sf.CalculateMatchScore(themeTags, entityTags)

	if score.Score >= 0.4 {
		t.Errorf("Expected no match (<0.4), got %.2f", score.Score)
	}

	if len(score.MissingTags) == 0 {
		t.Error("Expected missing tags for poor match")
	}
}

func TestCalculateMatchScore_EmptyThemeTags(t *testing.T) {
	sf := NewStyleFilter()

	themeTags := []string{}
	entityTags := []string{"zombie", "undead"}

	score := sf.CalculateMatchScore(themeTags, entityTags)

	// Should return default neutral score
	if score.Score != 0.5 {
		t.Errorf("Expected default score 0.5 for empty theme tags, got %.2f", score.Score)
	}
}

func TestCalculateMatchScore_CaseInsensitive(t *testing.T) {
	sf := NewStyleFilter()

	themeTags := []string{"Hospital", "MEDICAL"}
	entityTags := []string{"hospital", "medical"}

	score := sf.CalculateMatchScore(themeTags, entityTags)

	if score.Score < 0.8 {
		t.Errorf("Expected case-insensitive perfect match, got %.2f", score.Score)
	}

	if len(score.MatchedTags) != 2 {
		t.Errorf("Expected 2 matched tags (case-insensitive), got %d", len(score.MatchedTags))
	}
}

func TestAdaptEntity(t *testing.T) {
	sf := NewStyleFilter()

	entity := map[string]interface{}{
		"id":          "E-01",
		"name":        "Blind Zombie",
		"description": "A zombie that cannot see",
		"tags":        []string{"zombie", "undead"},
		"constraints": []string{
			"Cannot see anything",
			"Cannot open doors",
		},
	}

	theme := "Cyberpunk city"
	themeTags := []string{"cyberpunk", "tech"}

	adapted, err := sf.AdaptEntity(entity, theme, themeTags)
	if err != nil {
		t.Fatalf("AdaptEntity failed: %v", err)
	}

	// Check that constraints are preserved
	constraints, ok := adapted["constraints"].([]string)
	if !ok {
		t.Fatal("Adapted entity missing constraints")
	}
	if len(constraints) != 2 {
		t.Errorf("Expected 2 constraints, got %d", len(constraints))
	}

	// Check that tags are updated
	tags, ok := adapted["tags"].([]string)
	if !ok {
		t.Fatal("Adapted entity missing tags")
	}
	if len(tags) < 2 {
		t.Errorf("Expected at least 2 tags (original + theme), got %d", len(tags))
	}

	// Check that description is updated
	description, ok := adapted["description"].(string)
	if !ok {
		t.Fatal("Adapted entity missing description")
	}
	if description == entity["description"] {
		t.Error("Expected description to be updated")
	}
}

func TestImproviseEntity(t *testing.T) {
	sf := NewStyleFilter()

	theme := "Cyberpunk city digital ghost"
	themeTags := []string{"cyberpunk", "digital", "tech"}
	difficulty := "easy"

	entity, err := sf.ImproviseEntity(theme, themeTags, difficulty)
	if err != nil {
		t.Fatalf("ImproviseEntity failed: %v", err)
	}

	// Check required fields
	if _, ok := entity["id"]; !ok {
		t.Error("Improvised entity missing ID")
	}
	if _, ok := entity["name"]; !ok {
		t.Error("Improvised entity missing name")
	}
	if _, ok := entity["tags"]; !ok {
		t.Error("Improvised entity missing tags")
	}
	if _, ok := entity["difficulty"]; !ok {
		t.Error("Improvised entity missing difficulty")
	}
	if _, ok := entity["abilities"]; !ok {
		t.Error("Improvised entity missing abilities")
	}
	if _, ok := entity["constraints"]; !ok {
		t.Error("Improvised entity missing constraints")
	}
	if _, ok := entity["counterplay"]; !ok {
		t.Error("Improvised entity missing counterplay")
	}

	// Validate constraints
	constraints, ok := entity["constraints"].([]string)
	if !ok {
		t.Fatal("Constraints not a string slice")
	}
	if len(constraints) == 0 {
		t.Error("Improvised entity must have at least one constraint")
	}
}

func TestExtractThemeTags(t *testing.T) {
	sf := NewStyleFilter()

	tests := []struct {
		name         string
		theme        string
		expectedTags []string
	}{
		{
			name:         "Chinese hospital theme",
			theme:        "廢棄醫院的午夜值班",
			expectedTags: []string{"abandoned", "ruined", "hospital", "medical", "midnight", "night"},
		},
		{
			name:         "English cyberpunk theme",
			theme:        "Cyberpunk city digital ghost",
			expectedTags: []string{"cyberpunk", "tech", "city", "urban", "digital", "virtual"},
		},
		{
			name:         "School theme",
			theme:        "學校的詭異規則",
			expectedTags: []string{"school", "education"},
		},
		{
			name:         "Generic theme with no keywords",
			theme:        "Something completely random",
			expectedTags: []string{"horror", "mystery"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := sf.ExtractThemeTags(tt.theme)

			if len(tags) == 0 {
				t.Error("Expected some tags, got none")
			}

			// Check that at least some expected tags are present
			foundCount := 0
			for _, expected := range tt.expectedTags {
				for _, tag := range tags {
					if tag == expected {
						foundCount++
						break
					}
				}
			}

			// Should find at least 50% of expected tags
			minExpected := len(tt.expectedTags) / 2
			if foundCount < minExpected {
				t.Errorf("Expected at least %d tags from %v, found %d: %v",
					minExpected, tt.expectedTags, foundCount, tags)
			}
		})
	}
}

func TestValidateEntityConstraints_Valid(t *testing.T) {
	sf := NewStyleFilter()

	entity := map[string]interface{}{
		"id":   "E-01",
		"name": "Test Entity",
		"constraints": []string{
			"Cannot see",
			"Cannot open doors",
		},
	}

	err := sf.ValidateEntityConstraints(entity)
	if err != nil {
		t.Errorf("Expected valid entity, got error: %v", err)
	}
}

func TestValidateEntityConstraints_Missing(t *testing.T) {
	sf := NewStyleFilter()

	entity := map[string]interface{}{
		"id":   "E-01",
		"name": "Test Entity",
		// No constraints field
	}

	err := sf.ValidateEntityConstraints(entity)
	if err == nil {
		t.Error("Expected error for missing constraints, got nil")
	}
}

func TestValidateEntityConstraints_Empty(t *testing.T) {
	sf := NewStyleFilter()

	entity := map[string]interface{}{
		"id":          "E-01",
		"name":        "Test Entity",
		"constraints": []string{}, // Empty constraints
	}

	err := sf.ValidateEntityConstraints(entity)
	if err == nil {
		t.Error("Expected error for empty constraints, got nil")
	}
}

func TestValidateEntityConstraints_WrongType(t *testing.T) {
	sf := NewStyleFilter()

	entity := map[string]interface{}{
		"id":          "E-01",
		"name":        "Test Entity",
		"constraints": "not a slice", // Wrong type
	}

	err := sf.ValidateEntityConstraints(entity)
	if err == nil {
		t.Error("Expected error for wrong type constraints, got nil")
	}
}

// Benchmark tests

func BenchmarkCalculateMatchScore(b *testing.B) {
	sf := NewStyleFilter()
	themeTags := []string{"hospital", "medical", "abandoned", "night"}
	entityTags := []string{"hospital", "zombie", "undead"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.CalculateMatchScore(themeTags, entityTags)
	}
}

func BenchmarkExtractThemeTags(b *testing.B) {
	sf := NewStyleFilter()
	theme := "廢棄醫院的午夜值班 - Abandoned hospital night shift"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.ExtractThemeTags(theme)
	}
}
