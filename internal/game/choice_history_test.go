package game

import (
	"encoding/json"
	"testing"
	"time"
)

// ==========================================================================
// Story 7.3: Choice History Tests
// ==========================================================================

// TestChoiceHistory_RecordChoice tests recording choices to history.
// Story 7.3 AC5: Record choice content, timestamp, beat, state snapshot.
func TestChoiceHistory_RecordChoice(t *testing.T) {
	history := NewChoiceHistory()

	// Record a choice
	record := ChoiceRecord{
		ChoiceText:   "檢查房間裡的鏡子",
		BeatNumber:   1,
		IsFreeText:   false,
		HPAfter:      100,
		SANAfter:     95,
		TensionAfter: 10,
		Scene:        "bedroom",
	}

	history.RecordChoice(record)

	// Verify record was added
	if history.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", history.GetRecordCount())
	}

	// Verify timestamp was set
	lastRecord := history.GetLastRecord()
	if lastRecord == nil {
		t.Fatal("Expected last record to exist")
	}

	if lastRecord.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set automatically")
	}

	if lastRecord.ChoiceText != record.ChoiceText {
		t.Errorf("Expected choice text %s, got %s", record.ChoiceText, lastRecord.ChoiceText)
	}
}

// TestChoiceHistory_GetAllRecords tests retrieving all records.
func TestChoiceHistory_GetAllRecords(t *testing.T) {
	history := NewChoiceHistory()

	// Record multiple choices
	for i := 0; i < 5; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i + 1,
			IsFreeText:   false,
			HPAfter:      100 - (i * 5),
			SANAfter:     100 - (i * 3),
			TensionAfter: i * 10,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Get all records
	records := history.GetAllRecords()

	if len(records) != 5 {
		t.Errorf("Expected 5 records, got %d", len(records))
	}

	// Verify chronological order
	for i, record := range records {
		expectedBeat := i + 1
		if record.BeatNumber != expectedBeat {
			t.Errorf("Record %d: expected beat %d, got %d", i, expectedBeat, record.BeatNumber)
		}
	}

	// Verify it's a copy (modifications don't affect original)
	records[0].ChoiceText = "Modified"
	originalRecords := history.GetAllRecords()
	if originalRecords[0].ChoiceText == "Modified" {
		t.Error("Expected GetAllRecords to return a copy, but original was modified")
	}
}

// TestChoiceHistory_GetRecordsByBeatRange tests filtering by beat range.
// Story 7.3 AC5: Support query by beat range.
func TestChoiceHistory_GetRecordsByBeatRange(t *testing.T) {
	history := NewChoiceHistory()

	// Record choices at different beats
	for i := 1; i <= 10; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i,
			IsFreeText:   false,
			HPAfter:      100,
			SANAfter:     100,
			TensionAfter: 0,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Get records from beat 3 to 7
	records := history.GetRecordsByBeatRange(3, 7)

	if len(records) != 5 {
		t.Errorf("Expected 5 records (beats 3-7), got %d", len(records))
	}

	// Verify correct beats
	for i, record := range records {
		expectedBeat := 3 + i
		if record.BeatNumber != expectedBeat {
			t.Errorf("Record %d: expected beat %d, got %d", i, expectedBeat, record.BeatNumber)
		}
	}
}

// TestChoiceHistory_GetRecordsByScene tests filtering by scene.
func TestChoiceHistory_GetRecordsByScene(t *testing.T) {
	history := NewChoiceHistory()

	// Record choices in different scenes
	scenes := []string{"bedroom", "hallway", "bedroom", "kitchen", "bedroom"}
	for i, scene := range scenes {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i + 1,
			IsFreeText:   false,
			HPAfter:      100,
			SANAfter:     100,
			TensionAfter: 0,
			Scene:        scene,
		}
		history.RecordChoice(record)
	}

	// Get bedroom records
	bedroomRecords := history.GetRecordsByScene("bedroom")

	if len(bedroomRecords) != 3 {
		t.Errorf("Expected 3 bedroom records, got %d", len(bedroomRecords))
	}

	// Verify all are bedroom
	for _, record := range bedroomRecords {
		if record.Scene != "bedroom" {
			t.Errorf("Expected scene 'bedroom', got '%s'", record.Scene)
		}
	}
}

// TestChoiceHistory_GetRecent tests getting recent choices.
func TestChoiceHistory_GetRecent(t *testing.T) {
	history := NewChoiceHistory()

	// Record 10 choices
	for i := 1; i <= 10; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i,
			IsFreeText:   false,
			HPAfter:      100,
			SANAfter:     100,
			TensionAfter: 0,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Get 3 most recent
	recent := history.GetRecent(3)

	if len(recent) != 3 {
		t.Errorf("Expected 3 recent records, got %d", len(recent))
	}

	// Verify they are the last 3 (beats 8, 9, 10)
	expectedBeats := []int{8, 9, 10}
	for i, record := range recent {
		if record.BeatNumber != expectedBeats[i] {
			t.Errorf("Record %d: expected beat %d, got %d", i, expectedBeats[i], record.BeatNumber)
		}
	}

	// Test edge case: request more than available
	all := history.GetRecent(20)
	if len(all) != 10 {
		t.Errorf("Expected all 10 records, got %d", len(all))
	}

	// Test edge case: request 0
	none := history.GetRecent(0)
	if len(none) != 0 {
		t.Errorf("Expected 0 records, got %d", len(none))
	}
}

// TestChoiceHistory_GetLastRecord tests getting the last record.
func TestChoiceHistory_GetLastRecord(t *testing.T) {
	history := NewChoiceHistory()

	// Empty history
	if record := history.GetLastRecord(); record != nil {
		t.Error("Expected nil for empty history")
	}

	// Add records
	for i := 1; i <= 5; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i,
			IsFreeText:   false,
			HPAfter:      100,
			SANAfter:     100,
			TensionAfter: 0,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Get last record
	lastRecord := history.GetLastRecord()
	if lastRecord == nil {
		t.Fatal("Expected last record to exist")
	}

	if lastRecord.BeatNumber != 5 {
		t.Errorf("Expected last beat to be 5, got %d", lastRecord.BeatNumber)
	}
}

// TestChoiceHistory_GetStatsSummary tests stats summary generation.
// Story 7.3 AC5: Support stats for death debrief.
func TestChoiceHistory_GetStatsSummary(t *testing.T) {
	history := NewChoiceHistory()

	// Empty history
	emptySummary := history.GetStatsSummary()
	if emptySummary.TotalChoices != 0 {
		t.Errorf("Expected 0 total choices, got %d", emptySummary.TotalChoices)
	}

	// Record mixed choices (free text and predefined)
	choices := []struct {
		beat       int
		freeText   bool
		hp         int
		san        int
		tension    int
	}{
		{1, false, 100, 100, 0},
		{2, true, 95, 90, 10},
		{3, false, 90, 85, 20},
		{4, true, 80, 75, 35},
		{5, false, 70, 70, 50},
	}

	for _, choice := range choices {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   choice.beat,
			IsFreeText:   choice.freeText,
			HPAfter:      choice.hp,
			SANAfter:     choice.san,
			TensionAfter: choice.tension,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Get summary
	summary := history.GetStatsSummary()

	// Verify counts
	if summary.TotalChoices != 5 {
		t.Errorf("Expected 5 total choices, got %d", summary.TotalChoices)
	}

	if summary.FreeTextChoices != 2 {
		t.Errorf("Expected 2 free text choices, got %d", summary.FreeTextChoices)
	}

	if summary.PredefinedChoices != 3 {
		t.Errorf("Expected 3 predefined choices, got %d", summary.PredefinedChoices)
	}

	// Verify state changes
	expectedHPLost := 100 - 70 // From first to last
	if summary.TotalHPLost != expectedHPLost {
		t.Errorf("Expected %d HP lost, got %d", expectedHPLost, summary.TotalHPLost)
	}

	expectedSANLost := 100 - 70
	if summary.TotalSANLost != expectedSANLost {
		t.Errorf("Expected %d SAN lost, got %d", expectedSANLost, summary.TotalSANLost)
	}

	expectedTensionIncrease := 50 - 0
	if summary.TensionIncrease != expectedTensionIncrease {
		t.Errorf("Expected tension increase %d, got %d", expectedTensionIncrease, summary.TensionIncrease)
	}
}

// TestChoiceHistory_Clear tests clearing the history.
func TestChoiceHistory_Clear(t *testing.T) {
	history := NewChoiceHistory()

	// Add some records
	for i := 0; i < 5; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i + 1,
			IsFreeText:   false,
			HPAfter:      100,
			SANAfter:     100,
			TensionAfter: 0,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Clear
	history.Clear()

	// Verify empty
	if history.GetRecordCount() != 0 {
		t.Errorf("Expected 0 records after clear, got %d", history.GetRecordCount())
	}

	if lastRecord := history.GetLastRecord(); lastRecord != nil {
		t.Error("Expected nil last record after clear")
	}
}

// TestChoiceHistory_JSON tests JSON serialization.
func TestChoiceHistory_JSON(t *testing.T) {
	history := NewChoiceHistory()

	// Add records
	for i := 0; i < 3; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i + 1,
			IsFreeText:   i%2 == 0,
			HPAfter:      100 - (i * 10),
			SANAfter:     100 - (i * 5),
			TensionAfter: i * 15,
			Scene:        "scene",
			Timestamp:    time.Now(),
		}
		history.RecordChoice(record)
	}

	// Serialize to JSON
	jsonData, err := history.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var temp interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Deserialize to new history
	newHistory := NewChoiceHistory()
	if err := newHistory.FromJSON(jsonData); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Verify counts match
	if newHistory.GetRecordCount() != history.GetRecordCount() {
		t.Errorf("Expected %d records after deserialization, got %d",
			history.GetRecordCount(), newHistory.GetRecordCount())
	}

	// Verify data matches
	originalRecords := history.GetAllRecords()
	newRecords := newHistory.GetAllRecords()

	for i := range originalRecords {
		if originalRecords[i].ChoiceText != newRecords[i].ChoiceText {
			t.Errorf("Record %d: choice text mismatch", i)
		}
		if originalRecords[i].BeatNumber != newRecords[i].BeatNumber {
			t.Errorf("Record %d: beat number mismatch", i)
		}
		if originalRecords[i].HPAfter != newRecords[i].HPAfter {
			t.Errorf("Record %d: HP mismatch", i)
		}
	}
}

// TestChoiceHistory_Clone tests cloning.
func TestChoiceHistory_Clone(t *testing.T) {
	history := NewChoiceHistory()

	// Add records
	for i := 0; i < 3; i++ {
		record := ChoiceRecord{
			ChoiceText:   "選擇",
			BeatNumber:   i + 1,
			IsFreeText:   false,
			HPAfter:      100,
			SANAfter:     100,
			TensionAfter: 0,
			Scene:        "scene",
		}
		history.RecordChoice(record)
	}

	// Clone
	clone := history.Clone()

	// Verify counts match
	if clone.GetRecordCount() != history.GetRecordCount() {
		t.Errorf("Expected %d records in clone, got %d",
			history.GetRecordCount(), clone.GetRecordCount())
	}

	// Modify clone
	clone.RecordChoice(ChoiceRecord{
		ChoiceText:   "新選擇",
		BeatNumber:   4,
		IsFreeText:   false,
		HPAfter:      100,
		SANAfter:     100,
		TensionAfter: 0,
		Scene:        "scene",
	})

	// Verify original is unchanged
	if history.GetRecordCount() != 3 {
		t.Errorf("Expected original to have 3 records, got %d", history.GetRecordCount())
	}

	if clone.GetRecordCount() != 4 {
		t.Errorf("Expected clone to have 4 records, got %d", clone.GetRecordCount())
	}
}

// TestChoiceHistory_ThreadSafety tests concurrent access.
func TestChoiceHistory_ThreadSafety(t *testing.T) {
	history := NewChoiceHistory()

	// Concurrent writes
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			record := ChoiceRecord{
				ChoiceText:   "選擇",
				BeatNumber:   index,
				IsFreeText:   false,
				HPAfter:      100,
				SANAfter:     100,
				TensionAfter: 0,
				Scene:        "scene",
			}
			history.RecordChoice(record)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify count
	if history.GetRecordCount() != 10 {
		t.Errorf("Expected 10 records after concurrent writes, got %d", history.GetRecordCount())
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = history.GetAllRecords()
			_ = history.GetLastRecord()
			_ = history.GetRecordCount()
			done <- true
		}()
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}
}
