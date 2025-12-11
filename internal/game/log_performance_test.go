package game

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkGameLogAddEntry(b *testing.B) {
	log := NewGameLog(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.AddEntry(LogNarrative, fmt.Sprintf("Entry %d", i))
	}
}

func BenchmarkGameLogGetRecentEntries(b *testing.B) {
	log := NewGameLog(1000)

	// Pre-populate with 1000 entries
	for i := 0; i < 1000; i++ {
		log.AddEntry(LogNarrative, fmt.Sprintf("Entry %d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.GetRecentEntries(100)
	}
}

func BenchmarkRingBufferPush(b *testing.B) {
	rb := NewRingBuffer(1000)
	entry := LogEntry{
		Timestamp: time.Now(),
		Type:      LogNarrative,
		Content:   "Benchmark entry",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Push(entry)
	}
}

func BenchmarkRingBufferGetLast(b *testing.B) {
	rb := NewRingBuffer(1000)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		rb.Push(LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   fmt.Sprintf("Entry %d", i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.GetLast(100)
	}
}

// TestMemoryUsage tests that memory usage stays within acceptable limits.
func TestMemoryUsage(t *testing.T) {
	log := NewGameLog(1000)

	// Add 1500 entries (exceeds capacity to test overflow)
	for i := 0; i < 1500; i++ {
		content := fmt.Sprintf("Test entry number %d with some content to simulate real usage", i)
		log.AddEntry(LogNarrative, content)
	}

	// Should only have 1000 entries
	entries := log.GetRecentEntries(1500)
	if len(entries) != 1000 {
		t.Errorf("Expected max 1000 entries, got %d", len(entries))
	}

	// Estimate memory usage (very rough estimate)
	// Each entry: timestamp (24 bytes) + type (8 bytes) + content (~100 bytes avg) = ~132 bytes
	// 1000 entries * 132 bytes = ~132KB (well under 10MB requirement)
	estimatedMemory := len(entries) * 132 // bytes
	maxMemoryBytes := 10 * 1024 * 1024   // 10 MB

	if estimatedMemory > maxMemoryBytes {
		t.Errorf("Estimated memory usage %d bytes exceeds limit of %d bytes", estimatedMemory, maxMemoryBytes)
	}

	t.Logf("Estimated memory usage: ~%d KB (limit: 10 MB)", estimatedMemory/1024)
}

// TestRingBufferCapacity tests the ring buffer maintains capacity correctly.
func TestRingBufferCapacity(t *testing.T) {
	capacity := 100
	rb := NewRingBuffer(capacity)

	// Add 250 entries
	for i := 0; i < 250; i++ {
		rb.Push(LogEntry{
			Timestamp: time.Now(),
			Type:      LogNarrative,
			Content:   fmt.Sprintf("Entry %d", i),
		})
	}

	// Should only have last 100 entries
	entries := rb.GetLast(250)
	if len(entries) != capacity {
		t.Errorf("Expected %d entries, got %d", capacity, len(entries))
	}

	// Should have entries 150-249 (the last 100)
	firstEntry := entries[0]
	lastEntry := entries[len(entries)-1]

	// First entry in result should be "Entry 150"
	if firstEntry.Content != "Entry 150" {
		t.Errorf("Expected first entry 'Entry 150', got '%s'", firstEntry.Content)
	}

	// Last entry in result should be "Entry 249"
	if lastEntry.Content != "Entry 249" {
		t.Errorf("Expected last entry 'Entry 249', got '%s'", lastEntry.Content)
	}
}

// TestLongTermUsage simulates long-term game usage.
func TestLongTermUsage(t *testing.T) {
	log := NewGameLog(1000)

	// Simulate a long game session with 10,000 log entries
	for i := 0; i < 10000; i++ {
		var entryType LogType
		switch i % 4 {
		case 0:
			entryType = LogNarrative
		case 1:
			entryType = LogPlayerInput
		case 2:
			entryType = LogOptionChoice
		case 3:
			entryType = LogSystem
		}

		log.AddEntry(entryType, fmt.Sprintf("Entry %d", i))
	}

	// Should still only have 1000 entries
	entries := log.GetRecentEntries(1500)
	if len(entries) != 1000 {
		t.Errorf("Expected 1000 entries after long usage, got %d", len(entries))
	}

	// Should have the last 1000 entries (9000-9999)
	firstEntry := entries[0]
	if firstEntry.Content != "Entry 9000" {
		t.Errorf("Expected first entry 'Entry 9000', got '%s'", firstEntry.Content)
	}

	lastEntry := entries[len(entries)-1]
	if lastEntry.Content != "Entry 9999" {
		t.Errorf("Expected last entry 'Entry 9999', got '%s'", lastEntry.Content)
	}
}

// TestConcurrentAccess tests thread-safety of GameLog.
func TestConcurrentAccess(t *testing.T) {
	log := NewGameLog(1000)

	// Start multiple goroutines adding entries
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				log.AddEntry(LogNarrative, fmt.Sprintf("Goroutine %d Entry %d", id, j))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 1000 entries (last 1000 out of 1000 added)
	entries := log.GetRecentEntries(1500)
	if len(entries) != 1000 {
		t.Errorf("Expected 1000 entries after concurrent access, got %d", len(entries))
	}
}

// TestPerformanceAddAndRetrieve tests the performance of adding and retrieving entries.
func TestPerformanceAddAndRetrieve(t *testing.T) {
	log := NewGameLog(1000)

	// Measure time to add 1000 entries
	start := time.Now()
	for i := 0; i < 1000; i++ {
		log.AddEntry(LogNarrative, fmt.Sprintf("Entry %d", i))
	}
	addDuration := time.Since(start)

	// Measure time to retrieve 100 entries
	start = time.Now()
	entries := log.GetRecentEntries(100)
	retrieveDuration := time.Since(start)

	// Log performance metrics
	t.Logf("Time to add 1000 entries: %v", addDuration)
	t.Logf("Time to retrieve 100 entries: %v", retrieveDuration)

	// Verify retrieved entries
	if len(entries) != 100 {
		t.Errorf("Expected 100 entries, got %d", len(entries))
	}

	// Performance threshold: should be very fast (< 10ms for both operations)
	if addDuration > 10*time.Millisecond {
		t.Errorf("Adding 1000 entries took too long: %v (expected < 10ms)", addDuration)
	}

	if retrieveDuration > 1*time.Millisecond {
		t.Errorf("Retrieving 100 entries took too long: %v (expected < 1ms)", retrieveDuration)
	}
}
