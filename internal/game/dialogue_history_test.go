package game

import (
	"strings"
	"testing"
	"time"
)

// ==========================================================================
// Story 7.7: Dialogue History System Tests
// ==========================================================================

// TestNewDialogueHistory tests dialogue history creation
func TestNewDialogueHistory(t *testing.T) {
	dh := NewDialogueHistory()

	if dh == nil {
		t.Fatal("Expected non-nil dialogue history")
	}

	if dh.GetRecordCount() != 0 {
		t.Errorf("Expected 0 records, got %d", dh.GetRecordCount())
	}
}

// TestRecordDialogue tests recording dialogue
func TestRecordDialogue(t *testing.T) {
	dh := NewDialogueHistory()

	record := DialogueRecord{
		NPCName:    "測試NPC",
		NPCID:      "NPC-001",
		Dialogue:   "這裡很危險...",
		BeatNumber: 5,
		Scene:      "走廊",
		Tension:    60,
		SAN:        80,
	}

	dh.RecordDialogue(record)

	if dh.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", dh.GetRecordCount())
	}

	// Check if timestamp was set
	records := dh.GetAllRecords()
	if records[0].Timestamp.IsZero() {
		t.Error("Expected timestamp to be set automatically")
	}
}

// TestGetAllRecords tests retrieving all records
func TestGetAllRecords(t *testing.T) {
	dh := NewDialogueHistory()

	// Add multiple records
	for i := 0; i < 5; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	records := dh.GetAllRecords()

	if len(records) != 5 {
		t.Errorf("Expected 5 records, got %d", len(records))
	}

	// Verify records are in chronological order
	for i := 0; i < len(records); i++ {
		if records[i].BeatNumber != i {
			t.Errorf("Expected beat %d, got %d", i, records[i].BeatNumber)
		}
	}
}

// TestGetRecordsByBeatRange tests filtering by beat range
func TestGetRecordsByBeatRange(t *testing.T) {
	dh := NewDialogueHistory()

	// Add records at different beats
	for i := 0; i < 10; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i * 2, // 0, 2, 4, 6, 8, 10, 12, 14, 16, 18
		}
		dh.RecordDialogue(record)
	}

	// Get records from beat 4 to 10
	records := dh.GetRecordsByBeatRange(4, 10)

	expectedCount := 4 // beats 4, 6, 8, 10
	if len(records) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(records))
	}

	// Verify all beats are in range
	for _, record := range records {
		if record.BeatNumber < 4 || record.BeatNumber > 10 {
			t.Errorf("Beat %d is outside range [4, 10]", record.BeatNumber)
		}
	}
}

// TestGetRecordsByNPC tests filtering by NPC ID
func TestGetRecordsByNPC(t *testing.T) {
	dh := NewDialogueHistory()

	// Add records from different NPCs
	npcs := []string{"NPC-001", "NPC-002", "NPC-001", "NPC-003", "NPC-001"}
	for i, npcID := range npcs {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      npcID,
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	// Get records for NPC-001
	records := dh.GetRecordsByNPC("NPC-001")

	if len(records) != 3 {
		t.Errorf("Expected 3 records for NPC-001, got %d", len(records))
	}

	for _, record := range records {
		if record.NPCID != "NPC-001" {
			t.Errorf("Expected NPCID NPC-001, got %s", record.NPCID)
		}
	}
}

// TestGetRecordsByScene tests filtering by scene
func TestGetRecordsByScene(t *testing.T) {
	dh := NewDialogueHistory()

	scenes := []string{"走廊", "房間", "走廊", "地下室", "走廊"}
	for i, scene := range scenes {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i,
			Scene:      scene,
		}
		dh.RecordDialogue(record)
	}

	records := dh.GetRecordsByScene("走廊")

	if len(records) != 3 {
		t.Errorf("Expected 3 records in 走廊, got %d", len(records))
	}
}

// TestGetRecent tests getting recent records
func TestGetRecent(t *testing.T) {
	dh := NewDialogueHistory()

	// Add 10 records
	for i := 0; i < 10; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	// Get 3 most recent
	recent := dh.GetRecent(3)

	if len(recent) != 3 {
		t.Errorf("Expected 3 recent records, got %d", len(recent))
	}

	// Verify they are the last 3 (beats 7, 8, 9)
	expectedBeats := []int{7, 8, 9}
	for i, record := range recent {
		if record.BeatNumber != expectedBeats[i] {
			t.Errorf("Expected beat %d, got %d", expectedBeats[i], record.BeatNumber)
		}
	}
}

// TestGetLastRecord tests getting the last record
func TestGetLastRecord(t *testing.T) {
	dh := NewDialogueHistory()

	// Empty history
	lastRecord := dh.GetLastRecord()
	if lastRecord != nil {
		t.Error("Expected nil for empty history")
	}

	// Add records
	for i := 0; i < 5; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	lastRecord = dh.GetLastRecord()
	if lastRecord == nil {
		t.Fatal("Expected non-nil last record")
	}

	if lastRecord.BeatNumber != 4 {
		t.Errorf("Expected last record beat 4, got %d", lastRecord.BeatNumber)
	}
}

// TestGetClueRecords tests filtering clue records
func TestGetClueRecords(t *testing.T) {
	dh := NewDialogueHistory()

	// Add mix of clue and non-clue records
	clueFlags := []bool{false, true, false, true, true}
	for i, hasClue := range clueFlags {
		record := DialogueRecord{
			NPCName:      "NPC",
			NPCID:        "NPC-001",
			Dialogue:     "對話內容",
			BeatNumber:   i,
			ClueRevealed: hasClue,
		}
		if hasClue {
			record.SeedID = "SEED-001"
		}
		dh.RecordDialogue(record)
	}

	clueRecords := dh.GetClueRecords()

	if len(clueRecords) != 3 {
		t.Errorf("Expected 3 clue records, got %d", len(clueRecords))
	}

	for _, record := range clueRecords {
		if !record.ClueRevealed {
			t.Error("Expected all records to have ClueRevealed=true")
		}
	}
}

// TestGetDeathDialogues tests filtering death dialogues
func TestGetDeathDialogues(t *testing.T) {
	dh := NewDialogueHistory()

	// Add mix of death and normal dialogues
	deathFlags := []bool{false, false, true, false, true}
	for i, isDeath := range deathFlags {
		record := DialogueRecord{
			NPCName:         "NPC",
			NPCID:           "NPC-001",
			Dialogue:        "對話內容",
			BeatNumber:      i,
			IsDeathDialogue: isDeath,
		}
		dh.RecordDialogue(record)
	}

	deathRecords := dh.GetDeathDialogues()

	if len(deathRecords) != 2 {
		t.Errorf("Expected 2 death dialogues, got %d", len(deathRecords))
	}

	for _, record := range deathRecords {
		if !record.IsDeathDialogue {
			t.Error("Expected all records to have IsDeathDialogue=true")
		}
	}
}

// TestGetStatsSummary tests dialogue statistics
func TestGetStatsSummary(t *testing.T) {
	dh := NewDialogueHistory()

	// Add diverse records
	records := []DialogueRecord{
		{NPCName: "NPC1", NPCID: "NPC-001", Dialogue: "對話1", ClueRevealed: true},
		{NPCName: "NPC2", NPCID: "NPC-002", Dialogue: "對話2", ClueRevealed: false},
		{NPCName: "NPC1", NPCID: "NPC-001", Dialogue: "對話3", IsDeathDialogue: true},
		{NPCName: "NPC3", NPCID: "NPC-003", Dialogue: "對話4", IsQuestion: true},
		{NPCName: "NPC2", NPCID: "NPC-002", Dialogue: "對話5", ClueRevealed: true},
	}

	for _, record := range records {
		dh.RecordDialogue(record)
	}

	summary := dh.GetStatsSummary()

	if summary.TotalDialogues != 5 {
		t.Errorf("Expected 5 total dialogues, got %d", summary.TotalDialogues)
	}

	if summary.ClueRevelations != 2 {
		t.Errorf("Expected 2 clue revelations, got %d", summary.ClueRevelations)
	}

	if summary.DeathDialogues != 1 {
		t.Errorf("Expected 1 death dialogue, got %d", summary.DeathDialogues)
	}

	if summary.QuestionResponses != 1 {
		t.Errorf("Expected 1 question response, got %d", summary.QuestionResponses)
	}

	// Check dialogues by NPC
	if summary.DialoguesByNPC["NPC1"] != 2 {
		t.Errorf("Expected 2 dialogues for NPC1, got %d", summary.DialoguesByNPC["NPC1"])
	}
}

// TestClear tests clearing history
func TestClear(t *testing.T) {
	dh := NewDialogueHistory()

	// Add records
	for i := 0; i < 5; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	if dh.GetRecordCount() != 5 {
		t.Error("Expected 5 records before clear")
	}

	dh.Clear()

	if dh.GetRecordCount() != 0 {
		t.Errorf("Expected 0 records after clear, got %d", dh.GetRecordCount())
	}
}

// TestToJSON tests JSON serialization
func TestDialogueHistory_ToJSON(t *testing.T) {
	dh := NewDialogueHistory()

	record := DialogueRecord{
		NPCName:    "測試NPC",
		NPCID:      "NPC-001",
		Dialogue:   "這是測試對話",
		BeatNumber: 5,
		Scene:      "走廊",
		Timestamp:  time.Now(),
	}
	dh.RecordDialogue(record)

	jsonData, err := dh.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// Verify JSON contains expected fields
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "測試NPC") {
		t.Error("JSON should contain NPC name")
	}
}

// TestFromJSON tests JSON deserialization
func TestDialogueHistory_FromJSON(t *testing.T) {
	// Create original history
	original := NewDialogueHistory()
	record := DialogueRecord{
		NPCName:    "測試NPC",
		NPCID:      "NPC-001",
		Dialogue:   "這是測試對話",
		BeatNumber: 5,
	}
	original.RecordDialogue(record)

	// Serialize
	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Deserialize to new history
	restored := NewDialogueHistory()
	err = restored.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if restored.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record after restore, got %d", restored.GetRecordCount())
	}

	restoredRecords := restored.GetAllRecords()
	if restoredRecords[0].NPCName != "測試NPC" {
		t.Error("NPC name not restored correctly")
	}
}

// TestClone tests cloning dialogue history
func TestDialogueHistory_Clone(t *testing.T) {
	original := NewDialogueHistory()

	// Add records to original
	for i := 0; i < 3; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			NPCID:      "NPC-001",
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		original.RecordDialogue(record)
	}

	// Clone
	clone := original.Clone()

	if clone.GetRecordCount() != original.GetRecordCount() {
		t.Error("Clone should have same number of records")
	}

	// Modify original
	newRecord := DialogueRecord{
		NPCName:    "新NPC",
		NPCID:      "NPC-002",
		Dialogue:   "新對話",
		BeatNumber: 10,
	}
	original.RecordDialogue(newRecord)

	// Clone should not be affected
	if clone.GetRecordCount() != 3 {
		t.Error("Clone should not be affected by changes to original")
	}
}

// TestFormatForDisplay tests display formatting
func TestFormatForDisplay(t *testing.T) {
	dh := NewDialogueHistory()

	// Add records
	records := []DialogueRecord{
		{NPCName: "NPC1", Dialogue: "第一句對話", BeatNumber: 1},
		{NPCName: "NPC2", Dialogue: "第二句對話", BeatNumber: 2, ClueRevealed: true, SeedID: "SEED-001"},
		{NPCName: "NPC1", Dialogue: "臨終遺言", BeatNumber: 3, IsDeathDialogue: true},
	}

	for _, record := range records {
		dh.RecordDialogue(record)
	}

	// Format for display
	display := dh.FormatForDisplay(0)

	if !strings.Contains(display, "NPC 對話歷史") {
		t.Error("Display should contain header")
	}

	if !strings.Contains(display, "NPC1") {
		t.Error("Display should contain NPC names")
	}

	if !strings.Contains(display, "第一句對話") {
		t.Error("Display should contain dialogue content")
	}

	if !strings.Contains(display, "線索") {
		t.Error("Display should mark clue revelations")
	}

	if !strings.Contains(display, "臨終") {
		t.Error("Display should mark death dialogues")
	}
}

// TestFormatForDisplay_Empty tests display with no records
func TestFormatForDisplay_Empty(t *testing.T) {
	dh := NewDialogueHistory()

	display := dh.FormatForDisplay(0)

	if !strings.Contains(display, "尚未記錄") {
		t.Error("Empty display should show appropriate message")
	}
}

// TestFormatForDisplay_Limited tests limited display
func TestFormatForDisplay_Limited(t *testing.T) {
	dh := NewDialogueHistory()

	// Add 10 records
	for i := 0; i < 10; i++ {
		record := DialogueRecord{
			NPCName:    "NPC",
			Dialogue:   "對話內容",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	// Format with limit
	display := dh.FormatForDisplay(3)

	// Count beat markers to verify only 3 records shown
	beatCount := strings.Count(display, "[Beat")
	if beatCount != 3 {
		t.Errorf("Expected 3 records displayed, found %d beat markers", beatCount)
	}
}

// TestDialogueRecord_Structure tests the record structure
func TestDialogueRecord_Structure(t *testing.T) {
	record := DialogueRecord{
		NPCName:         "測試NPC",
		NPCID:           "NPC-001",
		Dialogue:        "測試對話內容",
		Timestamp:       time.Now(),
		BeatNumber:      5,
		Scene:           "走廊",
		Tension:         70,
		SAN:             80,
		IsQuestion:      true,
		ClueRevealed:    true,
		SeedID:          "SEED-001",
		IsDeathDialogue: false,
	}

	if record.NPCName == "" {
		t.Error("NPCName should not be empty")
	}

	if record.Dialogue == "" {
		t.Error("Dialogue should not be empty")
	}

	if record.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if record.BeatNumber < 0 {
		t.Error("BeatNumber should be non-negative")
	}

	if record.Tension < 0 || record.Tension > 100 {
		t.Error("Tension should be 0-100")
	}

	if record.SAN < 0 || record.SAN > 100 {
		t.Error("SAN should be 0-100")
	}
}

// BenchmarkRecordDialogue benchmarks dialogue recording
func BenchmarkRecordDialogue(b *testing.B) {
	dh := NewDialogueHistory()
	record := DialogueRecord{
		NPCName:    "BenchNPC",
		NPCID:      "NPC-001",
		Dialogue:   "基準測試對話",
		BeatNumber: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dh.RecordDialogue(record)
	}
}

// BenchmarkGetRecent benchmarks getting recent records
func BenchmarkGetRecent(b *testing.B) {
	dh := NewDialogueHistory()

	// Add 1000 records
	for i := 0; i < 1000; i++ {
		record := DialogueRecord{
			NPCName:    "BenchNPC",
			NPCID:      "NPC-001",
			Dialogue:   "基準測試對話",
			BeatNumber: i,
		}
		dh.RecordDialogue(record)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dh.GetRecent(10)
	}
}
