package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSaveManager(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)
	if manager == nil {
		t.Fatal("NewSaveManager returned nil")
	}

	if manager.SaveDir != saveDir {
		t.Errorf("Expected save dir %s, got %s", saveDir, manager.SaveDir)
	}
}

func TestSaveManagerSave(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Create game state to save
	gameState := NewSaveData()
	gameState.Metadata.SavedAt = time.Now()
	gameState.Player.HP = 80
	gameState.Player.SAN = 70
	gameState.Game.CurrentChapter = 2

	// Save to slot 1
	err := manager.Save(1, gameState)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	savePath := filepath.Join(saveDir, "save_1.json")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}

	// Verify checksum was set
	if gameState.Checksum == "" {
		t.Error("Checksum should be set after save")
	}
}

func TestSaveManagerSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, "deep", "nested", ".nightmare", "saves")

	manager := NewSaveManager(saveDir)
	gameState := NewSaveData()

	err := manager.Save(1, gameState)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		t.Error("Save directory was not created")
	}
}

func TestSaveManagerLoad(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Create and save game state
	original := NewSaveData()
	original.Metadata.SavedAt = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	original.Player.HP = 65
	original.Player.SAN = 45
	original.Player.Location = "basement"
	original.Game.CurrentChapter = 3
	original.Game.ChapterProgress = 0.5

	err := manager.Save(2, original)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load the save
	loaded, err := manager.Load(2)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data matches
	if loaded.Player.HP != original.Player.HP {
		t.Errorf("Player HP mismatch: expected %d, got %d", original.Player.HP, loaded.Player.HP)
	}

	if loaded.Player.SAN != original.Player.SAN {
		t.Errorf("Player SAN mismatch: expected %d, got %d", original.Player.SAN, loaded.Player.SAN)
	}

	if loaded.Game.CurrentChapter != original.Game.CurrentChapter {
		t.Errorf("Chapter mismatch: expected %d, got %d", original.Game.CurrentChapter, loaded.Game.CurrentChapter)
	}
}

func TestSaveManagerLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	_, err := manager.Load(1)
	if err == nil {
		t.Error("Expected error when loading non-existent save")
	}
}

func TestSaveManagerLoadCorrupted(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	if err := os.MkdirAll(saveDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create corrupted save file
	savePath := filepath.Join(saveDir, "save_1.json")
	if err := os.WriteFile(savePath, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}

	manager := NewSaveManager(saveDir)
	_, err := manager.Load(1)
	if err == nil {
		t.Error("Expected error when loading corrupted save")
	}
}

func TestSaveManagerLoadTamperedChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Save valid data
	original := NewSaveData()
	original.Player.HP = 100
	err := manager.Save(1, original)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the file (tamper with HP)
	savePath := filepath.Join(saveDir, "save_1.json")
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatal(err)
	}

	var loaded SaveData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}

	// Tamper with the data
	loaded.Player.HP = 999
	tamperedData, err := json.Marshal(loaded)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(savePath, tamperedData, 0600); err != nil {
		t.Fatal(err)
	}

	// Load should fail due to checksum mismatch
	_, err = manager.Load(1)
	if err == nil {
		t.Error("Expected error when loading tampered save")
	}
}

func TestSaveManagerValidateSlot(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewSaveManager(tmpDir)

	gameState := NewSaveData()

	// Invalid slots
	if err := manager.Save(0, gameState); err == nil {
		t.Error("Expected error for slot 0")
	}

	if err := manager.Save(4, gameState); err == nil {
		t.Error("Expected error for slot 4")
	}

	if err := manager.Save(-1, gameState); err == nil {
		t.Error("Expected error for negative slot")
	}
}

func TestSaveManagerAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Save first version
	v1 := NewSaveData()
	v1.Player.HP = 100
	if err := manager.Save(1, v1); err != nil {
		t.Fatal(err)
	}

	// Save second version
	v2 := NewSaveData()
	v2.Player.HP = 50
	if err := manager.Save(1, v2); err != nil {
		t.Fatal(err)
	}

	// Load and verify latest version
	loaded, err := manager.Load(1)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Player.HP != 50 {
		t.Errorf("Expected HP 50, got %d", loaded.Player.HP)
	}

	// Verify no temp files remain
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("Temp file should not remain: %s", entry.Name())
		}
	}
}

func TestSaveManagerGetSlotInfo(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Empty slot
	info, err := manager.GetSlotInfo(1)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsEmpty {
		t.Error("Expected slot to be empty")
	}

	// Save to slot
	gameState := NewSaveData()
	gameState.Metadata.SavedAt = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	gameState.Metadata.PlayTime = 3600 // 1 hour
	gameState.Game.CurrentChapter = 3
	if err := manager.Save(1, gameState); err != nil {
		t.Fatal(err)
	}

	// Check slot info
	info, err = manager.GetSlotInfo(1)
	if err != nil {
		t.Fatal(err)
	}

	if info.IsEmpty {
		t.Error("Expected slot to not be empty after save")
	}

	if info.Chapter != 3 {
		t.Errorf("Expected chapter 3, got %d", info.Chapter)
	}

	if info.PlayTime != 3600 {
		t.Errorf("Expected play time 3600, got %d", info.PlayTime)
	}
}

func TestSaveManagerGetAllSlotInfo(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Get all slots (all empty initially)
	allInfo, err := manager.GetAllSlotInfo()
	if err != nil {
		t.Fatal(err)
	}

	if len(allInfo) != 3 {
		t.Errorf("Expected 3 slots, got %d", len(allInfo))
	}

	for slotID, info := range allInfo {
		if !info.IsEmpty {
			t.Errorf("Slot %d should be empty", slotID)
		}
	}

	// Save to slot 2
	gameState := NewSaveData()
	gameState.Game.CurrentChapter = 5
	if err := manager.Save(2, gameState); err != nil {
		t.Fatal(err)
	}

	// Check again
	allInfo, err = manager.GetAllSlotInfo()
	if err != nil {
		t.Fatal(err)
	}

	if allInfo[2].IsEmpty {
		t.Error("Slot 2 should not be empty after save")
	}

	if !allInfo[1].IsEmpty || !allInfo[3].IsEmpty {
		t.Error("Slots 1 and 3 should still be empty")
	}
}

func TestSaveManagerDeleteSlot(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	// Save to slot
	gameState := NewSaveData()
	if err := manager.Save(1, gameState); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	savePath := filepath.Join(saveDir, "save_1.json")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatal("Save file should exist")
	}

	// Delete slot
	if err := manager.DeleteSlot(1); err != nil {
		t.Fatalf("DeleteSlot failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(savePath); !os.IsNotExist(err) {
		t.Error("Save file should be deleted")
	}
}

func TestSaveManagerDeleteNonExistentSlot(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewSaveManager(tmpDir)

	// Should not error when deleting non-existent slot
	err := manager.DeleteSlot(1)
	if err != nil {
		t.Errorf("DeleteSlot should not error for non-existent slot: %v", err)
	}
}

func TestSaveManagerSlotExists(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	manager := NewSaveManager(saveDir)

	if manager.SlotExists(1) {
		t.Error("Slot should not exist initially")
	}

	gameState := NewSaveData()
	if err := manager.Save(1, gameState); err != nil {
		t.Fatal(err)
	}

	if !manager.SlotExists(1) {
		t.Error("Slot should exist after save")
	}
}
