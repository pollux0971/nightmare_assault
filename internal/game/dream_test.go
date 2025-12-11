package game

import (
	"testing"
	"time"
)

func TestNewDreamLog(t *testing.T) {
	log := NewDreamLog()
	if log == nil {
		t.Fatal("NewDreamLog() returned nil")
	}
	if log.DreamCount() != 0 {
		t.Errorf("Expected 0 dreams, got %d", log.DreamCount())
	}
}

func TestLogDream(t *testing.T) {
	log := NewDreamLog()

	dream := DreamRecord{
		ID:        "dream-1",
		Type:      DreamTypeOpening,
		Timestamp: time.Now(),
		Content:   "You see a mirror reflecting darkness...",
		Context: DreamContext{
			PlayerHP:  100,
			PlayerSAN: 100,
			ChapterNum: 1,
		},
	}

	log.LogDream(dream)

	if log.DreamCount() != 1 {
		t.Errorf("Expected 1 dream after logging, got %d", log.DreamCount())
	}

	lastDream := log.GetLastDream()
	if lastDream == nil {
		t.Fatal("GetLastDream() returned nil")
	}
	if lastDream.ID != "dream-1" {
		t.Errorf("Expected dream ID 'dream-1', got '%s'", lastDream.ID)
	}
}

func TestGetDreamsByType(t *testing.T) {
	log := NewDreamLog()

	// Add multiple dreams of different types
	log.LogDream(DreamRecord{
		ID:   "dream-1",
		Type: DreamTypeOpening,
	})
	log.LogDream(DreamRecord{
		ID:   "dream-2",
		Type: DreamTypeChapter,
	})
	log.LogDream(DreamRecord{
		ID:   "dream-3",
		Type: DreamTypeChapter,
	})

	openingDreams := log.GetDreamsByType(DreamTypeOpening)
	if len(openingDreams) != 1 {
		t.Errorf("Expected 1 opening dream, got %d", len(openingDreams))
	}

	chapterDreams := log.GetDreamsByType(DreamTypeChapter)
	if len(chapterDreams) != 2 {
		t.Errorf("Expected 2 chapter dreams, got %d", len(chapterDreams))
	}
}

func TestGetLastDream_Empty(t *testing.T) {
	log := NewDreamLog()

	lastDream := log.GetLastDream()
	if lastDream != nil {
		t.Error("Expected nil for empty log, got dream")
	}
}

func TestDreamCount(t *testing.T) {
	log := NewDreamLog()

	if log.DreamCount() != 0 {
		t.Errorf("Expected 0 dreams initially, got %d", log.DreamCount())
	}

	for i := 0; i < 5; i++ {
		log.LogDream(DreamRecord{
			ID:   string(rune('A' + i)),
			Type: DreamTypeChapter,
		})
	}

	if log.DreamCount() != 5 {
		t.Errorf("Expected 5 dreams after logging, got %d", log.DreamCount())
	}
}
