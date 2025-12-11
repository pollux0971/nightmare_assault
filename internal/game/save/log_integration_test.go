package save

import (
	"os"
	"testing"
	"time"
)

func TestSaveDataWithLogEntries(t *testing.T) {
	saveData := NewSaveData()

	// Add some log entries
	saveData.LogEntries = []LogEntry{
		{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   "Test narrative entry",
		},
		{
			Timestamp: time.Now(),
			Type:      LogPlayerInput,
			Content:   "/status",
		},
		{
			Timestamp: time.Now(),
			Type:      LogSystem,
			Content:   "HP: 100/100",
		},
	}

	if len(saveData.LogEntries) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(saveData.LogEntries))
	}
}

func TestSaveAndLoadWithLogEntries(t *testing.T) {
	manager := NewSaveManager(GetSaveDir())

	// Create save data with log entries
	saveData := NewSaveData()
	saveData.LogEntries = []LogEntry{
		{
			Timestamp: time.Date(2025, 12, 11, 15, 30, 0, 0, time.UTC),
			Type:      LogNarrative,
			Content:   "You enter the dark room",
		},
		{
			Timestamp: time.Date(2025, 12, 11, 15, 31, 0, 0, time.UTC),
			Type:      LogPlayerInput,
			Content:   "look around",
		},
		{
			Timestamp: time.Date(2025, 12, 11, 15, 32, 0, 0, time.UTC),
			Type:      LogOptionChoice,
			Content:   "1. Touch the statue",
		},
	}

	// Save to slot 1
	err := manager.Save(1, saveData)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Clean up
	defer os.Remove(GetSavePath(1))

	// Load from slot 1
	loadedData, err := manager.Load(1)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify log entries were preserved
	if len(loadedData.LogEntries) != 3 {
		t.Fatalf("Expected 3 log entries, got %d", len(loadedData.LogEntries))
	}

	// Check first entry
	if loadedData.LogEntries[0].Content != "You enter the dark room" {
		t.Errorf("Expected first entry 'You enter the dark room', got '%s'", loadedData.LogEntries[0].Content)
	}

	if loadedData.LogEntries[0].Type != LogNarrative {
		t.Errorf("Expected first entry type LogNarrative, got %v", loadedData.LogEntries[0].Type)
	}

	// Check second entry
	if loadedData.LogEntries[1].Content != "look around" {
		t.Errorf("Expected second entry 'look around', got '%s'", loadedData.LogEntries[1].Content)
	}

	if loadedData.LogEntries[1].Type != LogPlayerInput {
		t.Errorf("Expected second entry type LogPlayerInput, got %v", loadedData.LogEntries[1].Type)
	}

	// Check third entry
	if loadedData.LogEntries[2].Content != "1. Touch the statue" {
		t.Errorf("Expected third entry '1. Touch the statue', got '%s'", loadedData.LogEntries[2].Content)
	}

	if loadedData.LogEntries[2].Type != LogOptionChoice {
		t.Errorf("Expected third entry type LogOptionChoice, got %v", loadedData.LogEntries[2].Type)
	}
}

func TestSaveWithMaxLogEntries(t *testing.T) {
	manager := NewSaveManager(GetSaveDir())

	saveData := NewSaveData()

	// Add 1000 log entries (max capacity)
	for i := 0; i < 1000; i++ {
		saveData.LogEntries = append(saveData.LogEntries, LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   "Entry number",
		})
	}

	// Save
	err := manager.Save(2, saveData)
	if err != nil {
		t.Fatalf("Failed to save with 1000 entries: %v", err)
	}

	// Clean up
	defer os.Remove(GetSavePath(2))

	// Load and verify
	loadedData, err := manager.Load(2)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(loadedData.LogEntries) != 1000 {
		t.Errorf("Expected 1000 log entries, got %d", len(loadedData.LogEntries))
	}
}

func TestLogTypeSerialization(t *testing.T) {
	manager := NewSaveManager(GetSaveDir())

	saveData := NewSaveData()

	// Add one entry of each type
	saveData.LogEntries = []LogEntry{
		{Timestamp: time.Now(), Type: LogNarrative, Content: "Narrative"},
		{Timestamp: time.Now(), Type: LogPlayerInput, Content: "Input"},
		{Timestamp: time.Now(), Type: LogOptionChoice, Content: "Choice"},
		{Timestamp: time.Now(), Type: LogSystem, Content: "System"},
	}

	// Save and load
	err := manager.Save(3, saveData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	defer os.Remove(GetSavePath(3))

	loadedData, err := manager.Load(3)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify all types were preserved
	if loadedData.LogEntries[0].Type != LogNarrative {
		t.Errorf("Expected LogNarrative, got %v", loadedData.LogEntries[0].Type)
	}
	if loadedData.LogEntries[1].Type != LogPlayerInput {
		t.Errorf("Expected LogPlayerInput, got %v", loadedData.LogEntries[1].Type)
	}
	if loadedData.LogEntries[2].Type != LogOptionChoice {
		t.Errorf("Expected LogOptionChoice, got %v", loadedData.LogEntries[2].Type)
	}
	if loadedData.LogEntries[3].Type != LogSystem {
		t.Errorf("Expected LogSystem, got %v", loadedData.LogEntries[3].Type)
	}
}
