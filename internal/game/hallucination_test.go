package game

import (
	"strings"
	"testing"
)

func TestNewHallucinationTracker(t *testing.T) {
	tracker := NewHallucinationTracker()

	if tracker == nil {
		t.Fatal("Expected HallucinationTracker to be created")
	}

	if len(tracker.Records) != 0 {
		t.Errorf("Expected empty tracker, got %d records", len(tracker.Records))
	}
}

func TestHallucinationTracker_AddRecord(t *testing.T) {
	tracker := NewHallucinationTracker()

	record := HallucinationRecord{
		TurnNumber:      5,
		SANValue:        15,
		OptionText:      "測試幻覺選項",
		RealOptions:     []string{"選項1", "選項2"},
		WasSelected:     false,
		ConsequenceDesc: "",
	}

	tracker.AddRecord(record)

	if len(tracker.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(tracker.Records))
	}

	if tracker.Records[0].OptionText != "測試幻覺選項" {
		t.Errorf("Expected option text to be saved")
	}
}

func TestHallucinationTracker_GetSelectedCount(t *testing.T) {
	tracker := NewHallucinationTracker()

	// Add 3 records, 2 selected
	tracker.AddRecord(HallucinationRecord{WasSelected: true})
	tracker.AddRecord(HallucinationRecord{WasSelected: false})
	tracker.AddRecord(HallucinationRecord{WasSelected: true})

	count := tracker.GetSelectedCount()

	if count != 2 {
		t.Errorf("Expected 2 selected, got %d", count)
	}
}

func TestHallucinationTracker_GetTotalCount(t *testing.T) {
	tracker := NewHallucinationTracker()

	tracker.AddRecord(HallucinationRecord{})
	tracker.AddRecord(HallucinationRecord{})
	tracker.AddRecord(HallucinationRecord{})

	count := tracker.GetTotalCount()

	if count != 3 {
		t.Errorf("Expected 3 total, got %d", count)
	}
}

func TestShouldInsertHallucination_HighSAN(t *testing.T) {
	// SAN >= 20: should never insert
	for san := 20; san <= 100; san += 10 {
		for i := 0; i < 100; i++ {
			if ShouldInsertHallucination(san) {
				t.Errorf("SAN=%d: should never insert hallucination", san)
			}
		}
	}
}

func TestShouldInsertHallucination_LowSAN(t *testing.T) {
	// SAN < 20: probabilistic insertion
	tests := []struct {
		san         int
		minRate     float64
		maxRate     float64
		description string
	}{
		{19, 0.03, 0.08, "SAN 19: ~5% rate"},
		{15, 0.20, 0.30, "SAN 15: ~25% rate"},
		{10, 0.45, 0.55, "SAN 10: ~50% rate"},
		{5, 0.70, 0.80, "SAN 5: ~75% rate"},
		{1, 0.90, 1.00, "SAN 1: ~95% rate"},
	}

	for _, tt := range tests {
		// Run 1000 trials to get statistically significant results
		insertCount := 0
		trials := 1000

		for i := 0; i < trials; i++ {
			if ShouldInsertHallucination(tt.san) {
				insertCount++
			}
		}

		rate := float64(insertCount) / float64(trials)

		if rate < tt.minRate || rate > tt.maxRate {
			t.Errorf("%s: rate=%.2f, want %.2f-%.2f", tt.description, rate, tt.minRate, tt.maxRate)
		}
	}
}

func TestInsertHallucinationOption_EmptyOptions(t *testing.T) {
	realOptions := []string{}
	hallucinationText := "幻覺選項"

	combined, index := InsertHallucinationOption(realOptions, hallucinationText)

	if len(combined) != 1 {
		t.Errorf("Expected 1 option, got %d", len(combined))
	}

	if combined[0] != hallucinationText {
		t.Errorf("Expected hallucination text, got %s", combined[0])
	}

	if index != 0 {
		t.Errorf("Expected index 0, got %d", index)
	}
}

func TestInsertHallucinationOption_RandomPosition(t *testing.T) {
	realOptions := []string{"選項1", "選項2", "選項3"}
	hallucinationText := "幻覺"

	// Run multiple times to check randomness
	positions := make(map[int]int)
	trials := 1000

	for i := 0; i < trials; i++ {
		_, index := InsertHallucinationOption(realOptions, hallucinationText)
		positions[index]++
	}

	// Should have inserted at various positions (0, 1, 2, or 3)
	if len(positions) < 3 {
		t.Errorf("Expected diverse positions, got only %v", positions)
	}

	// Each position should appear at least once in 1000 trials
	for i := 0; i <= len(realOptions); i++ {
		if positions[i] == 0 {
			t.Errorf("Position %d never used in %d trials", i, trials)
		}
	}
}

func TestInsertHallucinationOption_PreservesRealOptions(t *testing.T) {
	realOptions := []string{"A", "B", "C"}
	hallucinationText := "H"

	combined, hallucinationIndex := InsertHallucinationOption(realOptions, hallucinationText)

	// Should have 4 options total
	if len(combined) != 4 {
		t.Errorf("Expected 4 options, got %d", len(combined))
	}

	// Real options should still be present
	realCount := 0
	for _, opt := range combined {
		if opt == "A" || opt == "B" || opt == "C" {
			realCount++
		}
	}

	if realCount != 3 {
		t.Errorf("Expected 3 real options preserved, got %d", realCount)
	}

	// Hallucination should be at the specified index
	if combined[hallucinationIndex] != hallucinationText {
		t.Errorf("Expected hallucination at index %d, got %s", hallucinationIndex, combined[hallucinationIndex])
	}
}

func TestIsHallucinationIndex(t *testing.T) {
	hallucinationIndex := 2

	tests := []struct {
		selectedIndex int
		expected      bool
	}{
		{0, false},
		{1, false},
		{2, true}, // Matches hallucination index
		{3, false},
	}

	for _, tt := range tests {
		result := IsHallucinationIndex(tt.selectedIndex, hallucinationIndex)
		if result != tt.expected {
			t.Errorf("IsHallucinationIndex(%d, %d) = %v, want %v",
				tt.selectedIndex, hallucinationIndex, result, tt.expected)
		}
	}
}

func TestGenerateHallucinationConsequence(t *testing.T) {
	hallucinationText := "拿起桌上的槍"

	consequence := GenerateHallucinationConsequence(hallucinationText)

	// Should have a description
	if consequence.Description == "" {
		t.Error("Expected non-empty description")
	}

	// Should include revelation text
	revelationPhrases := []string{"不存在", "消失", "什麼都沒有", "空氣", "從來不在"}
	hasRevelation := false
	for _, phrase := range revelationPhrases {
		if strings.Contains(consequence.Description, phrase) {
			hasRevelation = true
			break
		}
	}

	if !hasRevelation {
		t.Errorf("Expected revelation phrase in description: %s", consequence.Description)
	}

	// Should have SAN loss
	if consequence.SANLoss != 5 {
		t.Errorf("Expected SAN loss of 5, got %d", consequence.SANLoss)
	}
}

func TestGenerateHallucinationConsequence_DangerousVsNormal(t *testing.T) {
	// Run multiple times to check that both types can occur
	dangerousCount := 0
	normalCount := 0
	trials := 100

	for i := 0; i < trials; i++ {
		consequence := GenerateHallucinationConsequence("測試")

		if consequence.IsDangerous {
			dangerousCount++
			// Dangerous consequences should mention danger
			if !strings.Contains(consequence.Description, "危險") {
				t.Error("Dangerous consequence should mention danger")
			}
		} else {
			normalCount++
			// Normal consequences should mention sensory betrayal
			if !strings.Contains(consequence.Description, "感官") && !strings.Contains(consequence.Description, "背叛") {
				t.Logf("Normal consequence should mention sensory issues: %s", consequence.Description)
			}
		}
	}

	// Both types should occur in 100 trials (probabilistic test)
	if dangerousCount == 0 {
		t.Error("Expected some dangerous consequences in 100 trials")
	}
	if normalCount == 0 {
		t.Error("Expected some normal consequences in 100 trials")
	}
}

func TestNewHallucinationGenerator(t *testing.T) {
	gen := NewHallucinationGenerator()

	if gen == nil {
		t.Fatal("Expected HallucinationGenerator to be created")
	}
}

func TestHallucinationGenerator_Generate(t *testing.T) {
	gen := NewHallucinationGenerator()

	scene := "你在一個黑暗的房間裡"
	realOptions := []string{"開燈", "摸索前進", "大聲呼救"}

	hallucination := gen.Generate(scene, realOptions)

	// Should return non-empty text
	if hallucination == "" {
		t.Error("Expected non-empty hallucination text")
	}

	// Should be in Chinese (template-based for now)
	if len(hallucination) == 0 {
		t.Error("Expected hallucination to have content")
	}
}

func TestHallucinationGenerator_GenerateWithContext(t *testing.T) {
	gen := NewHallucinationGenerator()

	scene := "你在走廊上"
	realOptions := []string{"往前走", "返回"}
	playerSAN := 10

	hallucination, err := gen.GenerateWithContext(scene, realOptions, playerSAN)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hallucination == "" {
		t.Error("Expected non-empty hallucination")
	}
}

func TestHallucinationGenerator_Diversity(t *testing.T) {
	gen := NewHallucinationGenerator()

	scene := "測試場景"
	realOptions := []string{"選項A", "選項B"}

	// Generate multiple hallucinations
	hallucinations := make(map[string]bool)
	for i := 0; i < 50; i++ {
		h := gen.Generate(scene, realOptions)
		hallucinations[h] = true
	}

	// Should generate diverse options (multiple different templates)
	if len(hallucinations) < 3 {
		t.Errorf("Expected diverse hallucinations, got only %d unique options", len(hallucinations))
	}
}

func TestHallucinationRecord_Structure(t *testing.T) {
	record := HallucinationRecord{
		TurnNumber:      10,
		SANValue:        12,
		OptionText:      "幻覺選項",
		RealOptions:     []string{"真實1", "真實2"},
		WasSelected:     true,
		ConsequenceDesc: "你意識到這不存在",
	}

	// Verify all fields are accessible
	if record.TurnNumber != 10 {
		t.Error("TurnNumber not set correctly")
	}
	if record.SANValue != 12 {
		t.Error("SANValue not set correctly")
	}
	if record.OptionText != "幻覺選項" {
		t.Error("OptionText not set correctly")
	}
	if len(record.RealOptions) != 2 {
		t.Error("RealOptions not set correctly")
	}
	if !record.WasSelected {
		t.Error("WasSelected not set correctly")
	}
	if record.ConsequenceDesc == "" {
		t.Error("ConsequenceDesc not set correctly")
	}
}
