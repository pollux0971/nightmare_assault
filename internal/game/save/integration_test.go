package save

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegration_SaveExitLoadFlow(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	// Simulate game session 1: Create and save game
	manager1 := NewSaveManager(saveDir)

	// Create game state
	gameState := NewSaveData()
	gameState.Metadata.SavedAt = time.Now()
	gameState.Metadata.PlayTime = 1800 // 30 minutes
	gameState.Metadata.Difficulty = "hard"
	gameState.Player.HP = 75
	gameState.Player.SAN = 60
	gameState.Player.Location = "basement"
	gameState.Player.KnownClues = []string{"bloody_key", "torn_note"}
	gameState.Game.CurrentChapter = 2
	gameState.Game.ChapterProgress = 0.4
	gameState.Game.TriggeredRules = []string{"rule_001"}

	// Save to slot 1
	err := manager1.Save(1, gameState)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	savePath := filepath.Join(saveDir, "save_1.json")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatal("Save file not created")
	}

	// Simulate "exit game" - create new manager (simulates restart)
	manager2 := NewSaveManager(saveDir)

	// Load from slot 1
	loaded, err := manager2.Load(1)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify all data was restored correctly
	if loaded.Player.HP != 75 {
		t.Errorf("HP mismatch: expected 75, got %d", loaded.Player.HP)
	}
	if loaded.Player.SAN != 60 {
		t.Errorf("SAN mismatch: expected 60, got %d", loaded.Player.SAN)
	}
	if loaded.Player.Location != "basement" {
		t.Errorf("Location mismatch: expected 'basement', got '%s'", loaded.Player.Location)
	}
	if loaded.Game.CurrentChapter != 2 {
		t.Errorf("Chapter mismatch: expected 2, got %d", loaded.Game.CurrentChapter)
	}
	if len(loaded.Player.KnownClues) != 2 {
		t.Errorf("KnownClues count mismatch: expected 2, got %d", len(loaded.Player.KnownClues))
	}
	if len(loaded.Game.TriggeredRules) != 1 {
		t.Errorf("TriggeredRules count mismatch: expected 1, got %d", len(loaded.Game.TriggeredRules))
	}
}

func TestIntegration_OverwriteConfirmFlow(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Save first version
	v1 := NewSaveData()
	v1.Player.HP = 100
	v1.Game.CurrentChapter = 1
	v1.Metadata.PlayTime = 600
	if err := manager.Save(1, v1); err != nil {
		t.Fatalf("First save failed: %v", err)
	}

	// Get slot info before overwrite
	info1, err := manager.GetSlotInfo(1)
	if err != nil {
		t.Fatalf("GetSlotInfo failed: %v", err)
	}
	if info1.IsEmpty {
		t.Fatal("Slot should not be empty after save")
	}
	if info1.Chapter != 1 {
		t.Errorf("Expected chapter 1, got %d", info1.Chapter)
	}

	// Save second version (overwrite)
	v2 := NewSaveData()
	v2.Player.HP = 50
	v2.Game.CurrentChapter = 3
	v2.Metadata.PlayTime = 3600
	if err := manager.Save(1, v2); err != nil {
		t.Fatalf("Second save (overwrite) failed: %v", err)
	}

	// Verify overwrite was successful
	info2, err := manager.GetSlotInfo(1)
	if err != nil {
		t.Fatalf("GetSlotInfo after overwrite failed: %v", err)
	}
	if info2.Chapter != 3 {
		t.Errorf("Expected chapter 3 after overwrite, got %d", info2.Chapter)
	}

	// Load and verify
	loaded, err := manager.Load(1)
	if err != nil {
		t.Fatalf("Load after overwrite failed: %v", err)
	}
	if loaded.Player.HP != 50 {
		t.Errorf("HP should be 50 after overwrite, got %d", loaded.Player.HP)
	}
}

func TestIntegration_ErrorScenarios(t *testing.T) {
	t.Run("LoadNonExistentSlot", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewSaveManager(tmpDir)

		_, err := manager.Load(1)
		if err == nil {
			t.Error("Expected error when loading non-existent slot")
		}
	})

	t.Run("LoadCorruptedFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
		if err := os.MkdirAll(saveDir, 0700); err != nil {
			t.Fatal(err)
		}

		// Create corrupted file
		corruptedPath := filepath.Join(saveDir, "save_1.json")
		if err := os.WriteFile(corruptedPath, []byte("not json {{{"), 0600); err != nil {
			t.Fatal(err)
		}

		manager := NewSaveManager(saveDir)
		_, err := manager.Load(1)
		if err == nil {
			t.Error("Expected error when loading corrupted file")
		}
	})

	t.Run("LoadTamperedChecksum", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
		manager := NewSaveManager(saveDir)

		// Save valid data
		gameState := NewSaveData()
		gameState.Player.HP = 100
		if err := manager.Save(1, gameState); err != nil {
			t.Fatal(err)
		}

		// Tamper with content (change HP value)
		savePath := filepath.Join(saveDir, "save_1.json")
		tamperedData := []byte(`{"version":1,"metadata":{"saved_at":"2024-01-01T00:00:00Z"},"player":{"hp":999},"checksum":"invalid"}`)
		if err := os.WriteFile(savePath, tamperedData, 0600); err != nil {
			t.Fatal(err)
		}

		// Load should fail
		_, loadErr := manager.Load(1)
		if loadErr == nil {
			t.Error("Expected error when loading tampered file")
		}
	})
}

func TestIntegration_CrossPlatformPaths(t *testing.T) {
	// Test that paths work correctly
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, "deep", "nested", ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	gameState := NewSaveData()
	gameState.Player.HP = 80

	// Save should create all directories
	if err := manager.Save(1, gameState); err != nil {
		t.Fatalf("Save with nested path failed: %v", err)
	}

	// Verify directory structure was created
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		t.Error("Save directory was not created")
	}

	// Load should work
	loaded, err := manager.Load(1)
	if err != nil {
		t.Fatalf("Load from nested path failed: %v", err)
	}

	if loaded.Player.HP != 80 {
		t.Errorf("HP mismatch: expected 80, got %d", loaded.Player.HP)
	}
}

func TestIntegration_AllSlotsManagement(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Initially all slots should be empty
	allInfo, err := manager.GetAllSlotInfo()
	if err != nil {
		t.Fatalf("GetAllSlotInfo failed: %v", err)
	}

	for slotID, info := range allInfo {
		if !info.IsEmpty {
			t.Errorf("Slot %d should be empty initially", slotID)
		}
	}

	// Save to slot 2
	gameState := NewSaveData()
	gameState.Game.CurrentChapter = 5
	if err := manager.Save(2, gameState); err != nil {
		t.Fatalf("Save to slot 2 failed: %v", err)
	}

	// Check slots again
	allInfo, err = manager.GetAllSlotInfo()
	if err != nil {
		t.Fatalf("GetAllSlotInfo after save failed: %v", err)
	}

	if allInfo[1].IsEmpty != true {
		t.Error("Slot 1 should still be empty")
	}
	if allInfo[2].IsEmpty != false {
		t.Error("Slot 2 should not be empty after save")
	}
	if allInfo[2].Chapter != 5 {
		t.Errorf("Slot 2 chapter mismatch: expected 5, got %d", allInfo[2].Chapter)
	}
	if allInfo[3].IsEmpty != true {
		t.Error("Slot 3 should still be empty")
	}

	// Delete slot 2
	if err := manager.DeleteSlot(2); err != nil {
		t.Fatalf("DeleteSlot failed: %v", err)
	}

	// Verify deletion
	if manager.SlotExists(2) {
		t.Error("Slot 2 should not exist after deletion")
	}
}

func TestIntegration_SaveDataIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Create comprehensive game state
	original := NewSaveData()
	original.Metadata.PlayTime = 7200
	original.Metadata.Difficulty = "nightmare"
	original.Metadata.StoryLength = "long"

	original.Player.HP = 42
	original.Player.SAN = 33
	original.Player.Location = "rooftop"
	original.Player.Inventory = []Item{
		{Name: "rusty_key", Description: "An old key covered in rust"},
		{Name: "flashlight", Description: "Battery is almost dead"},
		{Name: "diary_page", Description: "A torn page from someone's diary"},
	}
	original.Player.KnownClues = []string{"secret_passage", "hidden_room", "ghost_sighting"}

	original.Game.CurrentChapter = 4
	original.Game.ChapterProgress = 0.75
	original.Game.TriggeredRules = []string{"rule_death_look", "rule_mirror"}
	original.Game.DiscoveredRules = []string{"rule_death_look"}

	original.Teammates = []TeammateState{
		{Name: "Alice", Alive: true, HP: 60, Location: "rooftop", Relationship: 80},
		{Name: "Bob", Alive: false, HP: 0, Location: "basement", Relationship: -20},
	}

	original.Context.RecentSummary = "The group discovered a hidden passage..."
	original.Context.CurrentScene = "Wind howls across the rooftop..."
	original.Context.GameBible = "Core rules: Never look in mirrors after midnight..."

	// Save
	if err := manager.Save(1, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := manager.Load(1)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify everything was preserved
	if loaded.Metadata.PlayTime != original.Metadata.PlayTime {
		t.Errorf("PlayTime mismatch")
	}
	if loaded.Metadata.Difficulty != original.Metadata.Difficulty {
		t.Errorf("Difficulty mismatch")
	}
	if loaded.Player.HP != original.Player.HP {
		t.Errorf("Player HP mismatch")
	}
	if loaded.Player.SAN != original.Player.SAN {
		t.Errorf("Player SAN mismatch")
	}
	if len(loaded.Player.Inventory) != len(original.Player.Inventory) {
		t.Errorf("Inventory count mismatch")
	}
	if len(loaded.Teammates) != len(original.Teammates) {
		t.Errorf("Teammates count mismatch")
	}
	if loaded.Teammates[0].Name != "Alice" {
		t.Errorf("First teammate name mismatch")
	}
	if loaded.Teammates[1].Alive != false {
		t.Error("Bob should be dead")
	}
	if loaded.Context.RecentSummary != original.Context.RecentSummary {
		t.Errorf("RecentSummary mismatch")
	}
}
