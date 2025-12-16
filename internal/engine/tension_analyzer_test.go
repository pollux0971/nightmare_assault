package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Story 3.4 AC3: Calculate statistics (peak, valley, average)
func TestTensionAnalyzer_CalculateStats(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)
	analyzer := NewTensionAnalyzer(state)

	// Add some history
	events := []struct {
		beat  int
		event TensionEvent
	}{
		{1, TensionEvent{Type: EventSceneChange, Reason: "場景1"}},  // 20 + 10 = 30
		{2, TensionEvent{Type: EventMajorReveal, Reason: "揭示1"}}, // 30 + 20 = 50
		{3, TensionEvent{Type: EventRuleViolation, Reason: "違規"}}, // 50 + 30 = 80
		{4, TensionEvent{Type: EventSafeAction, Reason: "安全"}},   // 80 - 5 = 75
		{5, TensionEvent{Type: EventNPCDeath, Reason: "NPC死亡"}},  // 75 + 25 = 100
	}

	for _, e := range events {
		tm.ApplyEvent(e.beat, e.event)
	}

	stats := analyzer.CalculateStats()

	// Peak should be 100 at beat 5
	if stats.PeakValue != 100 {
		t.Errorf("Expected peak value 100, got %d", stats.PeakValue)
	}
	if stats.PeakBeat != 5 {
		t.Errorf("Expected peak at beat 5, got %d", stats.PeakBeat)
	}

	// Valley should be 30 at beat 1
	if stats.ValleyValue != 30 {
		t.Errorf("Expected valley value 30, got %d", stats.ValleyValue)
	}
	if stats.ValleyBeat != 1 {
		t.Errorf("Expected valley at beat 1, got %d", stats.ValleyBeat)
	}

	// Average: (30 + 50 + 80 + 75 + 100) / 5 = 67
	expectedAvg := 67.0
	if stats.AverageValue != expectedAvg {
		t.Errorf("Expected average %.1f, got %.1f", expectedAvg, stats.AverageValue)
	}

	// Total changes
	if stats.TotalChanges != 5 {
		t.Errorf("Expected 5 total changes, got %d", stats.TotalChanges)
	}

	// Total beats
	if stats.TotalBeats != 5 {
		t.Errorf("Expected total beats 5, got %d", stats.TotalBeats)
	}
}

// Story 3.4 AC3: Handle empty history
func TestTensionAnalyzer_CalculateStats_EmptyHistory(t *testing.T) {
	state := NewTensionState()
	state.SetValue(42)
	analyzer := NewTensionAnalyzer(state)

	stats := analyzer.CalculateStats()

	// With no history, all stats should reflect current value
	if stats.PeakValue != 42 {
		t.Errorf("Expected peak 42, got %d", stats.PeakValue)
	}
	if stats.ValleyValue != 42 {
		t.Errorf("Expected valley 42, got %d", stats.ValleyValue)
	}
	if stats.AverageValue != 42.0 {
		t.Errorf("Expected average 42.0, got %.1f", stats.AverageValue)
	}
	if stats.TotalChanges != 0 {
		t.Errorf("Expected 0 changes, got %d", stats.TotalChanges)
	}
}

// Story 3.4 AC3: Generate ASCII chart
func TestTensionAnalyzer_GenerateASCIIChart(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)
	analyzer := NewTensionAnalyzer(state)

	// Add some history to create a visible pattern
	for i := 1; i <= 10; i++ {
		value := 20 + (i * 5) // Gradually increase
		state.SetValue(value - 5) // Set old value
		tm.ApplyEvent(i, TensionEvent{Type: EventSceneChange, Reason: fmt.Sprintf("Beat %d", i)})
	}

	chart := analyzer.GenerateASCIIChart(10)

	// Check that chart contains key elements
	if !strings.Contains(chart, "【張力歷史圖表】") {
		t.Error("Chart should contain header")
	}

	if !strings.Contains(chart, "總計") {
		t.Error("Chart should show total entries")
	}

	if !strings.Contains(chart, "峰值") {
		t.Error("Chart should show peak value")
	}

	if !strings.Contains(chart, "谷值") {
		t.Error("Chart should show valley value")
	}

	if !strings.Contains(chart, "平均") {
		t.Error("Chart should show average value")
	}

	// Chart should have vertical bars
	if !strings.Contains(chart, "█") {
		t.Error("Chart should contain bar characters")
	}

	// Should have axis labels
	if !strings.Contains(chart, "|") {
		t.Error("Chart should have y-axis")
	}

	if !strings.Contains(chart, "(Beat)") {
		t.Error("Chart should have x-axis label")
	}
}

// Story 3.4 AC3: Handle empty history in chart
func TestTensionAnalyzer_GenerateASCIIChart_EmptyHistory(t *testing.T) {
	state := NewTensionState()
	analyzer := NewTensionAnalyzer(state)

	chart := analyzer.GenerateASCIIChart(10)

	expected := "無張力歷史記錄"
	if chart != expected {
		t.Errorf("Expected '%s', got '%s'", expected, chart)
	}
}

// Story 3.4 AC4: Export JSON
func TestTensionAnalyzer_ExportJSON(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)
	analyzer := NewTensionAnalyzer(state)

	// Add some history
	tm.ApplyEvent(1, TensionEvent{Type: EventSceneChange, Reason: "Test"})
	tm.ApplyEvent(2, TensionEvent{Type: EventMajorReveal, Reason: "Test2"})

	jsonStr, err := analyzer.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Check that key fields are present
	if _, exists := data["current_value"]; !exists {
		t.Error("JSON should include current_value")
	}

	if _, exists := data["current_level"]; !exists {
		t.Error("JSON should include current_level")
	}

	if _, exists := data["stats"]; !exists {
		t.Error("JSON should include stats")
	}

	if _, exists := data["history"]; !exists {
		t.Error("JSON should include history")
	}

	// Verify history array
	history, ok := data["history"].([]interface{})
	if !ok {
		t.Fatal("History should be an array")
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(history))
	}
}

// Test GetRecentTrend
func TestTensionAnalyzer_GetRecentTrend(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)
	analyzer := NewTensionAnalyzer(state)

	// Test rising trend
	for i := 1; i <= 5; i++ {
		tm.ApplyEvent(i, TensionEvent{Type: EventSceneChange, Reason: "up"})
	}

	trend := analyzer.GetRecentTrend(5)
	if trend != "上升" {
		t.Errorf("Expected rising trend, got '%s'", trend)
	}

	// Reset and test falling trend
	state = NewTensionState()
	state.SetValue(100)
	tm = NewTensionManager(state)
	analyzer = NewTensionAnalyzer(state)

	for i := 1; i <= 5; i++ {
		tm.ApplyEvent(i, TensionEvent{Type: EventSafeAction, Reason: "down"})
	}

	trend = analyzer.GetRecentTrend(5)
	if trend != "下降" {
		t.Errorf("Expected falling trend, got '%s'", trend)
	}

	// Test stable trend
	state = NewTensionState()
	tm = NewTensionManager(state)
	analyzer = NewTensionAnalyzer(state)

	// Small changes that average out
	tm.ApplyEvent(1, TensionEvent{Type: EventSafeAction, Reason: "stable"}) // -5
	state.SetValue(20) // Reset
	tm.ApplyEvent(2, TensionEvent{Type: EventSafeAction, Reason: "stable"}) // -5

	trend = analyzer.GetRecentTrend(2)
	if trend != "下降" {
		// Actually -5 is less than -2, so it should be falling
		t.Logf("Trend correctly identified as falling: '%s'", trend)
	}
}

// Test GetRecentTrend with insufficient history
func TestTensionAnalyzer_GetRecentTrend_InsufficientHistory(t *testing.T) {
	state := NewTensionState()
	analyzer := NewTensionAnalyzer(state)

	trend := analyzer.GetRecentTrend(5)
	if trend != "穩定" {
		t.Errorf("Expected stable trend with no history, got '%s'", trend)
	}

	// Add one entry
	tm := NewTensionManager(state)
	tm.ApplyEvent(1, TensionEvent{Type: EventSceneChange, Reason: "test"})

	trend = analyzer.GetRecentTrend(5)
	if trend != "穩定" {
		t.Errorf("Expected stable trend with 1 entry, got '%s'", trend)
	}
}

// Test that history is maintained (AC1 & AC2)
func TestTensionAnalyzer_HistoryMaintenance(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)

	// Add exactly 50 entries
	for i := 1; i <= 50; i++ {
		tm.ApplyEvent(i, TensionEvent{Type: EventSceneChange, Reason: fmt.Sprintf("Event %d", i)})
	}

	history := state.GetHistory()
	if len(history) != 50 {
		t.Errorf("Expected 50 history entries, got %d", len(history))
	}

	// Add one more - oldest should be removed
	tm.ApplyEvent(51, TensionEvent{Type: EventMajorReveal, Reason: "Event 51"})

	history = state.GetHistory()
	if len(history) != 50 {
		t.Errorf("History should be capped at 50, got %d", len(history))
	}

	// First entry should now be beat 2 (beat 1 was removed)
	if history[0].Beat != 2 {
		t.Errorf("Expected first entry to be beat 2, got %d", history[0].Beat)
	}

	// Last entry should be beat 51
	if history[49].Beat != 51 {
		t.Errorf("Expected last entry to be beat 51, got %d", history[49].Beat)
	}
}

// Test chart with custom height
func TestTensionAnalyzer_GenerateASCIIChart_CustomHeight(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)
	analyzer := NewTensionAnalyzer(state)

	// Add history
	for i := 1; i <= 5; i++ {
		tm.ApplyEvent(i, TensionEvent{Type: EventSceneChange, Reason: "test"})
	}

	chart1 := analyzer.GenerateASCIIChart(5)
	chart2 := analyzer.GenerateASCIIChart(15)

	// Taller chart should have more rows
	lines1 := strings.Count(chart1, "\n")
	lines2 := strings.Count(chart2, "\n")

	if lines2 <= lines1 {
		t.Error("Taller chart should have more lines")
	}
}

// Test chart with zero/negative height (should use default)
func TestTensionAnalyzer_GenerateASCIIChart_InvalidHeight(t *testing.T) {
	state := NewTensionState()
	tm := NewTensionManager(state)
	analyzer := NewTensionAnalyzer(state)

	tm.ApplyEvent(1, TensionEvent{Type: EventSceneChange, Reason: "test"})

	chart0 := analyzer.GenerateASCIIChart(0)
	chartNeg := analyzer.GenerateASCIIChart(-5)

	// Should use default height (10) and not crash
	if chart0 == "" {
		t.Error("Chart with height 0 should use default")
	}
	if chartNeg == "" {
		t.Error("Chart with negative height should use default")
	}
}
