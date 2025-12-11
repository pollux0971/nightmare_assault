package npc

import (
	"testing"
	"time"
)

func TestDialogueEntry(t *testing.T) {
	entry := DialogueEntry{
		Timestamp: time.Now(),
		Speaker:   "李明",
		Content:   "這裡看起來很詭異",
	}

	if entry.Speaker != "李明" {
		t.Errorf("Expected speaker '李明', got '%s'", entry.Speaker)
	}

	if entry.Content == "" {
		t.Error("Content should not be empty")
	}
}

func TestClueRevelation(t *testing.T) {
	clue := ClueRevelation{
		ClueID:         "clue-001",
		RevelationType: "worry",
		Subtlety:       7,
	}

	if clue.ClueID != "clue-001" {
		t.Errorf("Expected ClueID 'clue-001', got '%s'", clue.ClueID)
	}

	validTypes := []string{"hint", "worry", "observation", "memory"}
	found := false
	for _, vt := range validTypes {
		if clue.RevelationType == vt {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("RevelationType '%s' not in valid types", clue.RevelationType)
	}

	if clue.Subtlety < 1 || clue.Subtlety > 10 {
		t.Errorf("Subtlety %d should be between 1-10", clue.Subtlety)
	}
}

func TestDialogueSystem_AddEntry(t *testing.T) {
	system := NewDialogueSystem()

	entry := DialogueEntry{
		Timestamp: time.Now(),
		Speaker:   "Player",
		Content:   "你好",
	}

	system.AddEntry(entry)

	if len(system.GetHistory()) != 1 {
		t.Errorf("Expected 1 dialogue entry, got %d", len(system.GetHistory()))
	}

	retrieved := system.GetHistory()[0]
	if retrieved.Speaker != "Player" {
		t.Errorf("Expected speaker 'Player', got '%s'", retrieved.Speaker)
	}
}

func TestDialogueSystem_GetHistory(t *testing.T) {
	system := NewDialogueSystem()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		system.AddEntry(DialogueEntry{
			Timestamp: time.Now(),
			Speaker:   "Player",
			Content:   "Message",
		})
	}

	history := system.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(history))
	}
}

func TestDialogueSystem_GetRecentHistory(t *testing.T) {
	system := NewDialogueSystem()

	// Add 10 entries
	for i := 0; i < 10; i++ {
		system.AddEntry(DialogueEntry{
			Timestamp: time.Now(),
			Speaker:   "Player",
			Content:   "Message",
		})
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Get last 3 entries
	recent := system.GetRecentHistory(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent entries, got %d", len(recent))
	}
}

func TestDialogueSystem_ClearHistory(t *testing.T) {
	system := NewDialogueSystem()

	system.AddEntry(DialogueEntry{
		Timestamp: time.Now(),
		Speaker:   "Player",
		Content:   "Test",
	})

	system.ClearHistory()

	if len(system.GetHistory()) != 0 {
		t.Error("History should be empty after ClearHistory()")
	}
}

// Edge case tests

func TestDialogueSystem_GetRecentHistory_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		totalEntries  int
		requestCount  int
		expectedCount int
	}{
		{
			name:          "Request more than available",
			totalEntries:  3,
			requestCount:  10,
			expectedCount: 3,
		},
		{
			name:          "Request zero entries",
			totalEntries:  5,
			requestCount:  0,
			expectedCount: 0,
		},
		{
			name:          "Request negative count",
			totalEntries:  5,
			requestCount:  -5,
			expectedCount: 0,
		},
		{
			name:          "Empty dialogue log",
			totalEntries:  0,
			requestCount:  5,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewDialogueSystem()

			// Add entries
			for i := 0; i < tt.totalEntries; i++ {
				system.AddEntry(DialogueEntry{
					Timestamp: time.Now(),
					Speaker:   "Player",
					Content:   "Message",
				})
			}

			recent := system.GetRecentHistory(tt.requestCount)
			if len(recent) != tt.expectedCount {
				t.Errorf("Expected %d entries, got %d", tt.expectedCount, len(recent))
			}
		})
	}
}

func TestDialogueSystem_DeepCopy_ClueRevealed(t *testing.T) {
	system := NewDialogueSystem()

	// Add entry with ClueRevealed
	clue := &ClueRevelation{
		ClueID:         "clue-001",
		RevelationType: "hint",
		Subtlety:       5,
	}

	entry := DialogueEntry{
		Timestamp:    time.Now(),
		Speaker:      "Teammate",
		Content:      "I found something",
		ClueRevealed: clue,
	}
	system.AddEntry(entry)

	// Get history and modify the clue
	history := system.GetHistory()
	if history[0].ClueRevealed == nil {
		t.Fatal("Expected ClueRevealed to be non-nil")
	}

	// Modify the returned clue (should not affect original due to deep copy)
	history[0].ClueRevealed.ClueID = "modified-clue"
	history[0].ClueRevealed.Subtlety = 10

	// Verify original is unchanged
	originalHistory := system.GetHistory()
	if originalHistory[0].ClueRevealed.ClueID != "clue-001" {
		t.Errorf("Deep copy failed: ClueID was modified to %s", originalHistory[0].ClueRevealed.ClueID)
	}
	if originalHistory[0].ClueRevealed.Subtlety != 5 {
		t.Errorf("Deep copy failed: Subtlety was modified to %d", originalHistory[0].ClueRevealed.Subtlety)
	}
}

func TestDialogueSystem_DeepCopy_GetRecentHistory(t *testing.T) {
	system := NewDialogueSystem()

	clue := &ClueRevelation{
		ClueID:         "clue-002",
		RevelationType: "worry",
		Subtlety:       7,
	}

	system.AddEntry(DialogueEntry{
		Timestamp:    time.Now(),
		Speaker:      "Teammate",
		Content:      "Something's wrong",
		ClueRevealed: clue,
	})

	// Get recent history
	recent := system.GetRecentHistory(1)
	if len(recent) != 1 || recent[0].ClueRevealed == nil {
		t.Fatal("Expected 1 entry with ClueRevealed")
	}

	// Modify returned clue
	recent[0].ClueRevealed.ClueID = "modified"

	// Verify original unchanged
	original := system.GetHistory()
	if original[0].ClueRevealed.ClueID != "clue-002" {
		t.Error("Deep copy failed in GetRecentHistory")
	}
}

func TestDialogueSystem_SetGetTeammates(t *testing.T) {
	system := NewDialogueSystem()

	teammates := []*Teammate{
		NewTeammate("tm-001", "Alice", ArchetypeLogic),
		NewTeammate("tm-002", "Bob", ArchetypeVictim),
	}

	system.SetTeammates(teammates)

	retrieved := system.GetTeammates()
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 teammates, got %d", len(retrieved))
	}
}

func TestDialogueSystem_EmptyTeammates(t *testing.T) {
	system := NewDialogueSystem()

	teammates := system.GetTeammates()
	if len(teammates) != 0 {
		t.Errorf("Expected empty teammates, got %d", len(teammates))
	}
}
