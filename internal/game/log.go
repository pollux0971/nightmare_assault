package game

import (
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game/save"
)

// Re-export LogType and LogEntry from save package to avoid duplication
type LogType = save.LogType
type LogEntry = save.LogEntry

const (
	LogNarrative    = save.LogNarrative
	LogPlayerInput  = save.LogPlayerInput
	LogOptionChoice = save.LogOptionChoice
	LogSystem       = save.LogSystem
)

// RingBuffer is a circular buffer for storing log entries with a fixed capacity.
type RingBuffer struct {
	buffer []LogEntry
	head   int
	tail   int
	size   int
	cap    int
}

// NewRingBuffer creates a new ring buffer with the specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buffer: make([]LogEntry, capacity),
		head:   0,
		tail:   0,
		size:   0,
		cap:    capacity,
	}
}

// Push adds a new entry to the ring buffer.
func (r *RingBuffer) Push(entry LogEntry) {
	r.buffer[r.tail] = entry
	r.tail = (r.tail + 1) % r.cap

	if r.size < r.cap {
		r.size++
	} else {
		// Buffer is full, move head forward
		r.head = (r.head + 1) % r.cap
	}
}

// GetLast retrieves the last n entries from the ring buffer.
func (r *RingBuffer) GetLast(n int) []LogEntry {
	if n > r.size {
		n = r.size
	}

	if n == 0 {
		return []LogEntry{}
	}

	result := make([]LogEntry, n)
	for i := 0; i < n; i++ {
		idx := (r.tail - n + i + r.cap) % r.cap
		result[i] = r.buffer[idx]
	}
	return result
}

// GameLog manages the game's log entries.
type GameLog struct {
	entries *RingBuffer
	mu      sync.RWMutex
}

// NewGameLog creates a new game log with the specified capacity.
func NewGameLog(capacity int) *GameLog {
	return &GameLog{
		entries: NewRingBuffer(capacity),
	}
}

// AddEntry adds a new log entry.
func (g *GameLog) AddEntry(entryType LogType, content string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Type:      entryType,
		Content:   content,
	}
	g.entries.Push(entry)
}

// GetRecentEntries retrieves the n most recent entries.
func (g *GameLog) GetRecentEntries(n int) []LogEntry {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.entries.GetLast(n)
}
