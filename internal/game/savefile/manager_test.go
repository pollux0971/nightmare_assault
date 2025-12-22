package savefile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// TestSaveV2_BasicSaveLoad tests basic save and load operations for V2 format.
// Story 8.1 AC: Save to ~/.nightmare/saves/save_N.json
func TestSaveV2_BasicSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	// Create test data
	settings := GameSettings{
		Theme:      "測試主題",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}

	state := engine.NewGameStateV2()
	state.SetHP(80)
	state.SetSAN(70)
	state.CurrentScene = "test_scene"

	bible := &StoryBibleData{
		WorldSetting: &WorldSetting{
			Location:   "測試場景",
			Atmosphere: "詭異",
			TimeFrame:  "現代",
			Background: "測試背景故事",
		},
	}

	saveFile := NewSaveFileV2(1, settings, state, bible)
	saveFile.Meta.PlayTime = 300 // 5 minutes

	// Save
	startTime := time.Now()
	err := SaveV2(saveDir, 1, saveFile)
	saveTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("SaveV2 failed: %v", err)
	}

	// NFR-P06: Response time < 500ms
	if saveTime > 500*time.Millisecond {
		t.Errorf("Save took %v, exceeds 500ms requirement", saveTime)
	}

	// Verify file exists
	savePath := filepath.Join(saveDir, "save_1.json")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}

	// Load
	startTime = time.Now()
	loaded, err := LoadV2(saveDir, 1)
	loadTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("LoadV2 failed: %v", err)
	}

	// NFR-P06: Response time < 500ms
	if loadTime > 500*time.Millisecond {
		t.Errorf("Load took %v, exceeds 500ms requirement", loadTime)
	}

	// Verify data integrity
	if loaded.Meta.SaveID != 1 {
		t.Errorf("Expected SaveID 1, got %d", loaded.Meta.SaveID)
	}

	if loaded.Meta.PlayTime != 300 {
		t.Errorf("Expected PlayTime 300, got %d", loaded.Meta.PlayTime)
	}

	if loaded.GameState.GetHP() != 80 {
		t.Errorf("Expected HP 80, got %d", loaded.GameState.GetHP())
	}

	if loaded.GameState.GetSAN() != 70 {
		t.Errorf("Expected SAN 70, got %d", loaded.GameState.GetSAN())
	}

	if loaded.Settings.Theme != "測試主題" {
		t.Errorf("Expected theme '測試主題', got '%s'", loaded.Settings.Theme)
	}

	if loaded.StoryBible.WorldSetting.Location != "測試場景" {
		t.Errorf("Expected setting '測試場景', got '%s'", loaded.StoryBible.WorldSetting.Location)
	}
}

// TestSaveV2_ChecksumVerification tests checksum validation.
// Story 8.1 AC: Use checksum verification (NFR-S05)
func TestSaveV2_ChecksumVerification(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	bible := &StoryBibleData{}

	saveFile := NewSaveFileV2(1, settings, state, bible)

	// Save
	err := SaveV2(saveDir, 1, saveFile)
	if err != nil {
		t.Fatalf("SaveV2 failed: %v", err)
	}

	// Verify checksum was set
	if saveFile.Checksum == "" {
		t.Error("Checksum should be set after save")
	}

	// Manually corrupt the save file
	savePath := filepath.Join(saveDir, "save_1.json")
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("Failed to read save file: %v", err)
	}

	var rawData map[string]interface{}
	json.Unmarshal(data, &rawData)

	// Corrupt game state
	if gameState, ok := rawData["game_state"].(map[string]interface{}); ok {
		gameState["hp"] = 999 // Corrupt HP value
	}

	corruptedData, _ := json.MarshalIndent(rawData, "", "  ")
	os.WriteFile(savePath, corruptedData, 0600)

	// Try to load corrupted file
	_, err = LoadV2(saveDir, 1)
	if err == nil {
		t.Error("LoadV2 should fail with corrupted checksum")
	}

	if !IsCorruptedError(err) {
		t.Errorf("Expected CorruptedError, got %T: %v", err, err)
	}
}

// TestSaveV2_MultipleSlots tests saving to multiple slots.
// Story 8.1 AC: Save to 3 save slots
func TestSaveV2_MultipleSlots(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	bible := &StoryBibleData{}

	// Save to all 3 slots with different HP values
	for slotID := 1; slotID <= 3; slotID++ {
		state := engine.NewGameStateV2()
		state.SetHP(50 + slotID*10)

		saveFile := NewSaveFileV2(slotID, settings, state, bible)
		saveFile.Meta.PlayTime = slotID * 100

		err := SaveV2(saveDir, slotID, saveFile)
		if err != nil {
			t.Fatalf("SaveV2 to slot %d failed: %v", slotID, err)
		}
	}

	// Verify all slots can be loaded independently
	for slotID := 1; slotID <= 3; slotID++ {
		loaded, err := LoadV2(saveDir, slotID)
		if err != nil {
			t.Fatalf("LoadV2 from slot %d failed: %v", slotID, err)
		}

		expectedHP := 50 + slotID*10
		if loaded.GameState.GetHP() != expectedHP {
			t.Errorf("Slot %d: expected HP %d, got %d", slotID, expectedHP, loaded.GameState.GetHP())
		}

		if loaded.Meta.PlayTime != slotID*100 {
			t.Errorf("Slot %d: expected PlayTime %d, got %d", slotID, slotID*100, loaded.Meta.PlayTime)
		}
	}
}

// TestSaveV2_JSONStructure tests that JSON contains required fields.
// Story 8.1 AC: JSON should contain meta, settings, game_state, story_bible
func TestSaveV2_JSONStructure(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "測試",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	bible := &StoryBibleData{
		WorldSetting: &WorldSetting{
			Location: "測試場景",
		},
	}

	saveFile := NewSaveFileV2(1, settings, state, bible)

	err := SaveV2(saveDir, 1, saveFile)
	if err != nil {
		t.Fatalf("SaveV2 failed: %v", err)
	}

	// Read raw JSON
	savePath := filepath.Join(saveDir, "save_1.json")
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("Failed to read save file: %v", err)
	}

	var rawData map[string]interface{}
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify required top-level fields
	requiredFields := []string{"meta", "settings", "game_state", "story_bible", "checksum"}
	for _, field := range requiredFields {
		if _, exists := rawData[field]; !exists {
			t.Errorf("JSON missing required field: %s", field)
		}
	}

	// Verify meta structure
	meta, ok := rawData["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta field is not an object")
	}

	metaFields := []string{"save_id", "save_name", "created_at", "updated_at", "playtime", "game_version"}
	for _, field := range metaFields {
		if _, exists := meta[field]; !exists {
			t.Errorf("meta missing required field: %s", field)
		}
	}
}

// TestSaveV2_UpdateExistingSave tests updating an existing save.
func TestSaveV2_UpdateExistingSave(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	bible := &StoryBibleData{}

	// Initial save
	saveFile := NewSaveFileV2(1, settings, state, bible)
	saveFile.Meta.PlayTime = 100
	saveFile.Meta.CreatedAt = time.Now().Add(-1 * time.Hour)

	err := SaveV2(saveDir, 1, saveFile)
	if err != nil {
		t.Fatalf("Initial SaveV2 failed: %v", err)
	}

	createdAt := saveFile.Meta.CreatedAt
	time.Sleep(10 * time.Millisecond)

	// Update save
	state.SetHP(50)
	saveFile.Meta.PlayTime = 200

	err = SaveV2(saveDir, 1, saveFile)
	if err != nil {
		t.Fatalf("Update SaveV2 failed: %v", err)
	}

	// Load and verify
	loaded, err := LoadV2(saveDir, 1)
	if err != nil {
		t.Fatalf("LoadV2 failed: %v", err)
	}

	if loaded.GameState.GetHP() != 50 {
		t.Errorf("Expected updated HP 50, got %d", loaded.GameState.GetHP())
	}

	if loaded.Meta.PlayTime != 200 {
		t.Errorf("Expected updated PlayTime 200, got %d", loaded.Meta.PlayTime)
	}

	// CreatedAt should remain the same, UpdatedAt should be newer
	if !loaded.Meta.CreatedAt.Equal(createdAt) {
		t.Error("CreatedAt should not change on update")
	}

	if !loaded.Meta.UpdatedAt.After(createdAt) {
		t.Error("UpdatedAt should be after CreatedAt on update")
	}
}

// TestSaveV2_InvalidSlotID tests validation of slot IDs.
func TestSaveV2_InvalidSlotID(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	bible := &StoryBibleData{}

	invalidSlots := []int{0, -1, 4, 999}

	for _, slotID := range invalidSlots {
		saveFile := NewSaveFileV2(slotID, settings, state, bible)

		err := SaveV2(saveDir, slotID, saveFile)
		if err == nil {
			t.Errorf("SaveV2 should fail for invalid slot %d", slotID)
		}

		_, err = LoadV2(saveDir, slotID)
		if err == nil {
			t.Errorf("LoadV2 should fail for invalid slot %d", slotID)
		}
	}
}

// TestSaveV2_EmptySlot tests loading from an empty slot.
func TestSaveV2_EmptySlot(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	// Try to load from empty slot
	_, err := LoadV2(saveDir, 2)
	if err == nil {
		t.Error("LoadV2 should fail for empty slot")
	}

	// Error message should be user-friendly
	if err != nil && len(err.Error()) > 0 {
		// Just verify it's not a generic error
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	}
}

// TestSaveV2_ComplexGameState tests saving complex game state.
func TestSaveV2_ComplexGameState(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "複雜測試場景",
		Difficulty: "hell",
		Length:     "long",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}

	state := engine.NewGameStateV2()
	state.SetHP(45)
	state.SetSAN(30)
	state.IncrementBeat()
	state.IncrementBeat()
	state.CurrentScene = "complex_scene"
	state.Inventory = []string{"鑰匙", "手電筒", "筆記"}

	// Add some NPCs
	state.NPCStates = map[string]*engine.NPCState{
		"npc_1": {ID: "npc_1", Name: "張三"},
		"npc_2": {ID: "npc_2", Name: "李四"},
	}

	bible := &StoryBibleData{
		WorldSetting: &WorldSetting{
			Location:   "廢棄醫院",
			Atmosphere: "恐怖",
			TimeFrame:  "1990年代",
			Background: "這是一個廢棄的醫院，曾經發生過不為人知的實驗...",
		},
		CoreMystery: &CoreMystery{
			CoreTruth:  "醫院進行非法人體實驗",
			HiddenFrom: "偽裝成精神病院",
			Revelation: "第三幕 Climax",
		},
	}

	saveFile := NewSaveFileV2(1, settings, state, bible)

	// Save
	err := SaveV2(saveDir, 1, saveFile)
	if err != nil {
		t.Fatalf("SaveV2 failed: %v", err)
	}

	// Load
	loaded, err := LoadV2(saveDir, 1)
	if err != nil {
		t.Fatalf("LoadV2 failed: %v", err)
	}

	// Verify complex state
	if loaded.GameState.GetCurrentBeat() != 2 {
		t.Errorf("Expected beat 2, got %d", loaded.GameState.GetCurrentBeat())
	}

	if len(loaded.GameState.Inventory) != 3 {
		t.Errorf("Expected 3 items, got %d", len(loaded.GameState.Inventory))
	}

	if len(loaded.GameState.NPCStates) != 2 {
		t.Errorf("Expected 2 NPCs, got %d", len(loaded.GameState.NPCStates))
	}

	if loaded.StoryBible.CoreMystery.CoreTruth != "醫院進行非法人體實驗" {
		t.Errorf("CoreTruth not preserved: %s", loaded.StoryBible.CoreMystery.CoreTruth)
	}
}

// TestGetSlotInfoV2 tests getting slot information.
func TestGetSlotInfoV2(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	// Empty slot
	info, err := GetSlotInfo(saveDir, 1)
	if err != nil {
		t.Fatalf("GetSlotInfoV2 failed: %v", err)
	}

	if !info.IsEmpty {
		t.Error("Empty slot should be reported as empty")
	}

	// Create a save
	settings := GameSettings{
		Theme:      "",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	state.CurrentScene = "test_scene"
	state.IncrementBeat()
	state.IncrementBeat()
	state.IncrementBeat()

	bible := &StoryBibleData{}

	saveFile := NewSaveFileV2(1, settings, state, bible)
	saveFile.Meta.PlayTime = 600 // 10 minutes

	err = SaveV2(saveDir, 1, saveFile)
	if err != nil {
		t.Fatalf("SaveV2 failed: %v", err)
	}

	// Get slot info
	info, err = GetSlotInfo(saveDir, 1)
	if err != nil {
		t.Fatalf("GetSlotInfoV2 failed: %v", err)
	}

	if info.IsEmpty {
		t.Error("Saved slot should not be reported as empty")
	}

	if info.Chapter != 3 {
		t.Errorf("Expected chapter 3, got %d", info.Chapter)
	}

	if info.PlayTime != 600 {
		t.Errorf("Expected playtime 600, got %d", info.PlayTime)
	}

	if info.Location != "test_scene" {
		t.Errorf("Expected location 'test_scene', got '%s'", info.Location)
	}
}

// BenchmarkSaveV2 benchmarks save performance.
// NFR-P06: Response time < 500ms
func BenchmarkSaveV2(b *testing.B) {
	tmpDir := b.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "benchmark",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	bible := &StoryBibleData{}

	saveFile := NewSaveFileV2(1, settings, state, bible)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := SaveV2(saveDir, 1, saveFile)
		if err != nil {
			b.Fatalf("SaveV2 failed: %v", err)
		}
	}
}

// BenchmarkLoadV2 benchmarks load performance.
// NFR-P06: Response time < 500ms
func BenchmarkLoadV2(b *testing.B) {
	tmpDir := b.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	

	settings := GameSettings{
		Theme:      "benchmark",
		Difficulty: "hard",
		Length:     "medium",
		Model:      "openai/gpt-4-turbo",
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
	state := engine.NewGameStateV2()
	bible := &StoryBibleData{}

	saveFile := NewSaveFileV2(1, settings, state, bible)

	err := SaveV2(saveDir, 1, saveFile)
	if err != nil {
		b.Fatalf("SaveV2 failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadV2(saveDir, 1)
		if err != nil {
			b.Fatalf("LoadV2 failed: %v", err)
		}
	}
}
