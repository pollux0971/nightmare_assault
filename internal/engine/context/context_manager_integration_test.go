package context

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestContextManager_FullWorkflow tests the complete workflow
func TestContextManager_FullWorkflow(t *testing.T) {
	cm, err := NewContextManager(DefaultContextConfig())
	if err != nil {
		t.Fatalf("Failed to create ContextManager: %v", err)
	}

	// Simulate 20 rounds of gameplay
	for i := 1; i <= 20; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("選擇 %d", i),
			StoryContent: fmt.Sprintf("故事文本 %d - 你進入了一個神秘的房間，發現了一些線索。", i),
			CluesFound:   []string{fmt.Sprintf("clue-%d", i)},
		}

		err := cm.AddHistoryEntry(entry)
		if err != nil {
			t.Fatalf("Failed to add entry %d: %v", i, err)
		}
	}

	// Verify metadata
	metadata := cm.GetContextMetadata()
	if metadata.TotalEntries != 20 {
		t.Errorf("Expected 20 total entries, got %d", metadata.TotalEntries)
	}

	// Window should only contain last 5 entries (default WindowSize)
	if metadata.WindowEntries != 5 {
		t.Errorf("Expected 5 window entries, got %d", metadata.WindowEntries)
	}

	// Verify token counting is working
	if metadata.TotalTokens == 0 {
		t.Error("Expected non-zero token count")
	}

	// Verify context formatting
	context := cm.FormatCompleteContext()
	if !strings.Contains(context, "最近歷史") {
		t.Error("Expected context to contain recent history section")
	}

	if !strings.Contains(context, "回合") {
		t.Error("Expected context to contain beat numbers")
	}

	// Verify we can get optimized context
	window := cm.GetOptimizedContext()
	if window == nil {
		t.Fatal("Expected window to be returned")
	}

	recentEntries := window.GetWindow()
	if len(recentEntries) != 5 {
		t.Errorf("Expected 5 recent entries, got %d", len(recentEntries))
	}

	// Verify most recent entry
	lastEntry := recentEntries[len(recentEntries)-1]
	if lastEntry.Beat != 20 {
		t.Errorf("Expected last entry to be beat 20, got %d", lastEntry.Beat)
	}
}

// TestContextManager_AgentIntegration simulates Agent usage
func TestContextManager_AgentIntegration(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add some history (simulating 5 rounds of gameplay)
	for i := 1; i <= 5; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("調查房間 %d", i),
			StoryContent: fmt.Sprintf("你發現了一些奇怪的痕跡，感到一陣寒意。房間 %d 似乎隱藏著秘密。", i),
			CluesFound:   []string{fmt.Sprintf("mysterious-mark-%d", i)},
			RulesTriggered: []string{},
		}

		cm.AddHistoryEntry(entry)
	}

	// Simulate Agent requesting context for prompt construction
	context := cm.FormatCompleteContext()

	// Verify context is meaningful
	if len(context) < 100 {
		t.Error("Context is too short, should have substantial content")
	}

	// Verify it contains history sections
	if !strings.Contains(context, "最近歷史") {
		t.Error("Context missing recent history section")
	}

	// Simulate constructing an LLM prompt
	prompt := fmt.Sprintf(`你是一個恐怖遊戲的敘事者。

%s

玩家接下來選擇：進入黑暗的走廊

請根據上述歷史，生成接下來的故事文本（200字以內）：`, context)

	// Verify prompt is valid
	if !strings.Contains(prompt, "最近歷史") {
		t.Error("Prompt should contain context")
	}

	if !strings.Contains(prompt, "進入黑暗的走廊") {
		t.Error("Prompt should contain player choice")
	}

	t.Logf("Generated prompt length: %d characters", len(prompt))
}

// TestContextManager_TokenOptimization tests token usage optimization
func TestContextManager_TokenOptimization(t *testing.T) {
	cm, _ := NewContextManager(ContextConfig{
		WindowSize:         5,
		SummaryTrigger:     15,
		TokenLimit:         1000, // Low limit to test optimization
		ModelName:          "gpt-4",
		EnableAutoOptimize: true,
	})

	// Add many entries with substantial content
	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "探索房間，仔細檢查每一個角落，尋找任何可疑的線索",
			StoryContent: strings.Repeat("你在房間中發現了許多奇怪的符號和古老的文字，牆壁上刻滿了神秘的圖案。", 5),
		}
		cm.AddHistoryEntry(entry)
	}

	metadata := cm.GetContextMetadata()

	// Check token usage ratio
	if metadata.TokenUsageRatio > 1.0 {
		t.Logf("Warning: Token usage exceeded limit (%d / %d = %.1f%%)",
			metadata.TotalTokens, cm.config.TokenLimit, metadata.TokenUsageRatio*100)
	}

	// Window should still be limited to 5 entries
	if metadata.WindowEntries > 5 {
		t.Errorf("Window should not exceed 5 entries, got %d", metadata.WindowEntries)
	}

	t.Logf("Token usage: %d / %d (%.1f%%)",
		metadata.TotalTokens, cm.config.TokenLimit, metadata.TokenUsageRatio*100)
}

// TestContextManager_SaveLoadWorkflow tests save/load in realistic scenario
func TestContextManager_SaveLoadWorkflow(t *testing.T) {
	// Create and populate context manager
	cm1, _ := NewContextManager(DefaultContextConfig())

	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("選擇 %d", i),
			StoryContent: fmt.Sprintf("故事內容 %d", i),
		}
		cm1.AddHistoryEntry(entry)
	}

	cm1.window.Summary = "[Chapter 1 Summary: 玩家探索了廢棄醫院的第一層，發現了多個房間中的線索。]"

	// Save to file
	tempFile := "/tmp/test_context_workflow.json"
	defer func() {
		_ = os.Remove(tempFile)
	}()

	err := cm1.SaveContext(tempFile)
	if err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	// Create new manager and load
	cm2, _ := NewContextManager(DefaultContextConfig())
	err = cm2.LoadContext(tempFile)
	if err != nil {
		t.Fatalf("Failed to load context: %v", err)
	}

	// Verify restoration
	metadata1 := cm1.GetContextMetadata()
	metadata2 := cm2.GetContextMetadata()

	if metadata2.TotalEntries != metadata1.TotalEntries {
		t.Errorf("Entry count mismatch: expected %d, got %d",
			metadata1.TotalEntries, metadata2.TotalEntries)
	}

	if cm2.window.Summary != cm1.window.Summary {
		t.Error("Summary not restored correctly")
	}

	// Verify we can continue using the loaded context
	entry := HistoryEntry{
		Beat:         11,
		PlayerChoice: "新選擇",
		StoryContent: "新故事",
	}
	err = cm2.AddHistoryEntry(entry)
	if err != nil {
		t.Errorf("Failed to add entry to loaded context: %v", err)
	}

	metadata3 := cm2.GetContextMetadata()
	if metadata3.TotalEntries != 11 {
		t.Errorf("Expected 11 entries after adding to loaded context, got %d", metadata3.TotalEntries)
	}
}

// TestContextManager_HealthMonitoring tests health monitoring features
func TestContextManager_HealthMonitoring(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Initially healthy
	status := cm.GetHealthStatus()
	if !status.IsHealthy {
		t.Error("New context manager should be healthy")
	}

	// Add entries
	for i := 1; i <= 5; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事",
		}
		cm.AddHistoryEntry(entry)
	}

	status = cm.GetHealthStatus()
	if !status.TokenUsageOK {
		t.Error("Token usage should be OK with only 5 entries")
	}

	// Check statistics
	stats := cm.GetStatistics()
	if stats.TotalEntriesAdded != 5 {
		t.Errorf("Expected 5 entries in stats, got %d", stats.TotalEntriesAdded)
	}

	if stats.AverageTokensPerEntry == 0 {
		t.Error("Expected non-zero average tokens per entry")
	}

	t.Logf("Statistics: %d entries, avg %d tokens/entry, %d summaries",
		stats.TotalEntriesAdded, stats.AverageTokensPerEntry, stats.SummariesGenerated)
}

// TestContextManager_ClearAndRestart tests clearing and restarting workflow
func TestContextManager_ClearAndRestart(t *testing.T) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add entries
	for i := 1; i <= 5; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事",
		}
		cm.AddHistoryEntry(entry)
	}

	metadata1 := cm.GetContextMetadata()
	if metadata1.TotalEntries != 5 {
		t.Error("Expected 5 entries before clear")
	}

	// Clear
	cm.Clear()

	metadata2 := cm.GetContextMetadata()
	if metadata2.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", metadata2.TotalEntries)
	}

	// Restart with new entries
	for i := 1; i <= 3; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "新選擇",
			StoryContent: "新故事",
		}
		cm.AddHistoryEntry(entry)
	}

	metadata3 := cm.GetContextMetadata()
	if metadata3.TotalEntries != 3 {
		t.Errorf("Expected 3 entries after restart, got %d", metadata3.TotalEntries)
	}
}
