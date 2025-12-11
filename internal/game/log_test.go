package game

import (
	"testing"
	"time"
)

func TestLogType(t *testing.T) {
	// Test that LogType constants are defined
	types := []LogType{
		LogNarrative,
		LogPlayerInput,
		LogOptionChoice,
		LogSystem,
	}

	if len(types) != 4 {
		t.Errorf("Expected 4 log types, got %d", len(types))
	}
}

func TestLogEntry(t *testing.T) {
	now := time.Now()
	entry := LogEntry{
		Timestamp: now,
		Type:      LogNarrative,
		Content:   "Test content",
	}

	if entry.Timestamp != now {
		t.Errorf("Expected timestamp %v, got %v", now, entry.Timestamp)
	}

	if entry.Type != LogNarrative {
		t.Errorf("Expected type %v, got %v", LogNarrative, entry.Type)
	}

	if entry.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", entry.Content)
	}
}

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer(100)

	if rb == nil {
		t.Fatal("NewRingBuffer returned nil")
	}

	if rb.cap != 100 {
		t.Errorf("Expected capacity 100, got %d", rb.cap)
	}

	if rb.size != 0 {
		t.Errorf("Expected initial size 0, got %d", rb.size)
	}

	if rb.head != 0 {
		t.Errorf("Expected head 0, got %d", rb.head)
	}

	if rb.tail != 0 {
		t.Errorf("Expected tail 0, got %d", rb.tail)
	}
}

func TestRingBufferPush(t *testing.T) {
	rb := NewRingBuffer(3)

	entry1 := LogEntry{
		Timestamp: time.Now(),
		Type:      LogNarrative,
		Content:   "Entry 1",
	}

	rb.Push(entry1)

	if rb.size != 1 {
		t.Errorf("Expected size 1 after push, got %d", rb.size)
	}

	if rb.tail != 1 {
		t.Errorf("Expected tail 1 after push, got %d", rb.tail)
	}
}

func TestRingBufferPushOverflow(t *testing.T) {
	rb := NewRingBuffer(3)

	// Push 5 entries (exceeds capacity)
	for i := 0; i < 5; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   "Entry " + string(rune(i)),
		}
		rb.Push(entry)
	}

	if rb.size != 3 {
		t.Errorf("Expected size 3 after overflow, got %d", rb.size)
	}

	// Should have overwritten the oldest entries
	entries := rb.GetLast(3)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestRingBufferGetLast(t *testing.T) {
	rb := NewRingBuffer(10)

	// Push 5 entries
	for i := 0; i < 5; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   "Entry " + string(rune('0'+i)),
		}
		rb.Push(entry)
	}

	// Get last 3
	entries := rb.GetLast(3)

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Should get entries 2, 3, 4 in order
	if entries[0].Content != "Entry 2" {
		t.Errorf("Expected 'Entry 2', got '%s'", entries[0].Content)
	}
	if entries[1].Content != "Entry 3" {
		t.Errorf("Expected 'Entry 3', got '%s'", entries[1].Content)
	}
	if entries[2].Content != "Entry 4" {
		t.Errorf("Expected 'Entry 4', got '%s'", entries[2].Content)
	}
}

func TestRingBufferGetLastExceedsSize(t *testing.T) {
	rb := NewRingBuffer(10)

	// Push only 3 entries
	for i := 0; i < 3; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   "Entry " + string(rune('0'+i)),
		}
		rb.Push(entry)
	}

	// Request 10 entries
	entries := rb.GetLast(10)

	// Should only return 3
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestRingBufferGetLastZero(t *testing.T) {
	rb := NewRingBuffer(10)

	// Push some entries
	for i := 0; i < 3; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   "Entry " + string(rune('0'+i)),
		}
		rb.Push(entry)
	}

	// Request 0 entries
	entries := rb.GetLast(0)

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}

func TestRingBufferEmpty(t *testing.T) {
	rb := NewRingBuffer(10)

	entries := rb.GetLast(5)

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from empty buffer, got %d", len(entries))
	}
}

func TestNewGameLog(t *testing.T) {
	log := NewGameLog(100)

	if log == nil {
		t.Fatal("NewGameLog returned nil")
	}

	if log.entries == nil {
		t.Error("GameLog entries should not be nil")
	}
}

func TestGameLogAddEntry(t *testing.T) {
	log := NewGameLog(10)

	log.AddEntry(LogNarrative, "Test narrative")

	entries := log.GetRecentEntries(1)
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Content != "Test narrative" {
		t.Errorf("Expected content 'Test narrative', got '%s'", entries[0].Content)
	}

	if entries[0].Type != LogNarrative {
		t.Errorf("Expected type LogNarrative, got %v", entries[0].Type)
	}
}

func TestGameLogGetRecentEntries(t *testing.T) {
	log := NewGameLog(100)

	// Add 5 entries
	for i := 0; i < 5; i++ {
		log.AddEntry(LogNarrative, "Entry "+string(rune('0'+i)))
	}

	// Get recent 3
	entries := log.GetRecentEntries(3)

	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Should get entries 2, 3, 4
	if entries[0].Content != "Entry 2" {
		t.Errorf("Expected 'Entry 2', got '%s'", entries[0].Content)
	}
	if entries[1].Content != "Entry 3" {
		t.Errorf("Expected 'Entry 3', got '%s'", entries[1].Content)
	}
	if entries[2].Content != "Entry 4" {
		t.Errorf("Expected 'Entry 4', got '%s'", entries[2].Content)
	}
}

func TestGameLogCapacityLimit(t *testing.T) {
	log := NewGameLog(100)

	// Add 150 entries (exceeds capacity)
	for i := 0; i < 150; i++ {
		log.AddEntry(LogNarrative, "Entry "+string(rune(i)))
	}

	// Should only have 100 entries
	entries := log.GetRecentEntries(150)
	if len(entries) != 100 {
		t.Errorf("Expected max 100 entries, got %d", len(entries))
	}
}

func TestGameLogTypes(t *testing.T) {
	log := NewGameLog(100)

	// Add different types
	log.AddEntry(LogNarrative, "Narrative text")
	log.AddEntry(LogPlayerInput, "Player input")
	log.AddEntry(LogOptionChoice, "Option choice")
	log.AddEntry(LogSystem, "System message")

	entries := log.GetRecentEntries(4)

	if entries[0].Type != LogNarrative {
		t.Errorf("Expected LogNarrative, got %v", entries[0].Type)
	}
	if entries[1].Type != LogPlayerInput {
		t.Errorf("Expected LogPlayerInput, got %v", entries[1].Type)
	}
	if entries[2].Type != LogOptionChoice {
		t.Errorf("Expected LogOptionChoice, got %v", entries[2].Type)
	}
	if entries[3].Type != LogSystem {
		t.Errorf("Expected LogSystem, got %v", entries[3].Type)
	}
}
