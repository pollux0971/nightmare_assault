package context

import (
	"os"
	"strings"
	"testing"
)

// TestNewContextManager tests ContextManager initialization
func TestNewContextManager(t *testing.T) {
	config := ContextConfig{
		WindowSize:         5,
		SummaryTrigger:     15,
		TokenLimit:         8000,
		ModelName:          "gpt-4",
		EnableAutoOptimize: true,
	}

	cm, err := NewContextManager(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cm == nil {
		t.Fatal("Expected ContextManager to be initialized")
	}

	if cm.window == nil {
		t.Error("Expected window to be initialized")
	}

	if cm.tokenCounter == nil {
		t.Error("Expected tokenCounter to be initialized")
	}

	if cm.config.WindowSize != 5 {
		t.Errorf("Expected WindowSize 5, got %d", cm.config.WindowSize)
	}
}

// TestContextManager_GetOptimizedContext tests GetOptimizedContext method
func TestContextManager_GetOptimizedContext(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	window := cm.GetOptimizedContext()
	if window == nil {
		t.Fatal("Expected window to be returned")
	}
}

// TestContextManager_AddHistoryEntry tests adding history entries
func TestContextManager_AddHistoryEntry(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	entry := HistoryEntry{
		Beat:         1,
		PlayerChoice: "前進",
		StoryContent: "你進入了黑暗的走廊...",
	}

	err := cm.AddHistoryEntry(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	metadata := cm.GetContextMetadata()
	if metadata.TotalEntries != 1 {
		t.Errorf("Expected 1 entry, got %d", metadata.TotalEntries)
	}
}

// TestContextManager_FormatSummarySection tests summary formatting
func TestContextManager_FormatSummarySection(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// No summary yet
	formatted := cm.FormatSummarySection()
	if formatted != "" {
		t.Error("Expected empty string when no summary exists")
	}

	// Add summary
	cm.window.Summary = "[Chapter 1 Summary: 玩家進入廢棄醫院...]"

	formatted = cm.FormatSummarySection()
	if !strings.Contains(formatted, "前情提要") {
		t.Error("Expected formatted summary to contain '前情提要'")
	}

	if !strings.Contains(formatted, "Chapter 1") {
		t.Error("Expected formatted summary to contain summary content")
	}
}

// TestContextManager_FormatRecentHistory tests recent history formatting
func TestContextManager_FormatRecentHistory(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Empty history
	formatted := cm.FormatRecentHistory()
	if !strings.Contains(formatted, "最近歷史") {
		t.Error("Expected formatted history to contain '最近歷史'")
	}

	if !strings.Contains(formatted, "（無歷史）") {
		t.Error("Expected empty history message")
	}

	// Add entries
	for i := 1; i <= 3; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事內容",
			CluesFound:   []string{"clue1"},
		}
		cm.AddHistoryEntry(entry)
	}

	formatted = cm.FormatRecentHistory()
	if !strings.Contains(formatted, "回合 1") {
		t.Error("Expected beat number in formatted history")
	}

	if !strings.Contains(formatted, "玩家選擇") {
		t.Error("Expected player choice label")
	}

	if !strings.Contains(formatted, "發現線索") {
		t.Error("Expected clues label when clues are present")
	}
}

// TestContextManager_FormatCompleteContext tests complete context formatting
func TestContextManager_FormatCompleteContext(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add summary and entries
	cm.window.Summary = "[Chapter 1 Summary: Test summary]"
	for i := 1; i <= 2; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事",
		}
		cm.AddHistoryEntry(entry)
	}

	formatted := cm.FormatCompleteContext()

	if !strings.Contains(formatted, "前情提要") {
		t.Error("Expected summary section")
	}

	if !strings.Contains(formatted, "最近歷史") {
		t.Error("Expected recent history section")
	}

	if !strings.Contains(formatted, "Token 使用") {
		t.Error("Expected token usage comment")
	}
}

// TestContextManager_SaveLoadContext tests serialization
func TestContextManager_SaveLoadContext(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add data
	for i := 1; i <= 5; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事內容",
		}
		cm.AddHistoryEntry(entry)
	}

	cm.window.Summary = "Test summary"

	// Save
	tempFile := "/tmp/test_context.json"
	defer os.Remove(tempFile)

	err := cm.SaveContext(tempFile)
	if err != nil {
		t.Fatalf("Expected no error saving, got: %v", err)
	}

	// Load into new manager
	cm2, _ := NewContextManager(DefaultContextConfig())
	err = cm2.LoadContext(tempFile)
	if err != nil {
		t.Fatalf("Expected no error loading, got: %v", err)
	}

	// Verify
	if cm2.window.Summary != cm.window.Summary {
		t.Error("Summary not restored correctly")
	}

	if len(cm2.window.AllEntries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(cm2.window.AllEntries))
	}
}

// TestContextManager_GetContextMetadata tests metadata retrieval
func TestContextManager_GetContextMetadata(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add entries
	for i := 1; i <= 7; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事內容包含一些文字",
		}
		cm.AddHistoryEntry(entry)
	}

	metadata := cm.GetContextMetadata()

	if metadata.TotalEntries != 7 {
		t.Errorf("Expected 7 entries, got %d", metadata.TotalEntries)
	}

	if metadata.WindowEntries != 5 { // Default window size
		t.Errorf("Expected 5 window entries, got %d", metadata.WindowEntries)
	}

	if metadata.TotalTokens == 0 {
		t.Error("Expected non-zero token count")
	}

	if metadata.TokenUsageRatio < 0 || metadata.TokenUsageRatio > 1 {
		t.Errorf("Invalid token usage ratio: %f", metadata.TokenUsageRatio)
	}
}

// TestContextManager_GetHealthStatus tests health monitoring
func TestContextManager_GetHealthStatus(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	status := cm.GetHealthStatus()

	if !status.IsHealthy {
		t.Error("Expected healthy status for new manager")
	}

	if !status.TokenUsageOK {
		t.Error("Expected token usage OK for new manager")
	}

	if !status.WindowSizeOK {
		t.Error("Expected window size OK")
	}

	if status.SummaryStatus != "Not yet generated" {
		t.Errorf("Expected 'Not yet generated', got '%s'", status.SummaryStatus)
	}
}

// TestContextManager_GetStatistics tests statistics retrieval
func TestContextManager_GetStatistics(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add entries
	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事",
		}
		cm.AddHistoryEntry(entry)
	}

	stats := cm.GetStatistics()

	if stats.TotalEntriesAdded != 10 {
		t.Errorf("Expected 10 entries, got %d", stats.TotalEntriesAdded)
	}

	if stats.AverageTokensPerEntry == 0 {
		t.Error("Expected non-zero average tokens")
	}
}

// TestContextManager_Clear tests clearing context
func TestContextManager_Clear(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add data
	for i := 1; i <= 5; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事",
		}
		cm.AddHistoryEntry(entry)
	}

	cm.Clear()

	metadata := cm.GetContextMetadata()
	if metadata.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", metadata.TotalEntries)
	}
}

// TestContextManager_ConcurrentAccess tests thread safety
func TestContextManager_ConcurrentAccess(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(beat int) {
			entry := HistoryEntry{
				Beat:         beat,
				PlayerChoice: "選擇",
				StoryContent: "故事",
			}
			cm.AddHistoryEntry(entry)
			done <- true
		}(i + 1)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = cm.GetOptimizedContext()
			_ = cm.FormatCompleteContext()
			_ = cm.GetContextMetadata()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	metadata := cm.GetContextMetadata()
	if metadata.TotalEntries != 10 {
		t.Errorf("Expected 10 entries, got %d", metadata.TotalEntries)
	}
}

