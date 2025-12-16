package context

import (
	"strings"
	"testing"
)

// TestContextWindow_SetGetTokenCounter 測試設置和獲取 TokenCounter
func TestContextWindow_SetGetTokenCounter(t *testing.T) {
	cw := NewContextWindow()

	// Initially nil
	if cw.GetTokenCounter() != nil {
		t.Error("Initial token counter should be nil")
	}

	// Set counter
	counter := NewEstimateTokenCounter()
	cw.SetTokenCounter(counter)

	// Verify
	if cw.GetTokenCounter() == nil {
		t.Error("Token counter should not be nil after SetTokenCounter")
	}
}

// TestContextWindow_CalculateTotalTokens_NoCounter 測試沒有 counter 時返回 0
func TestContextWindow_CalculateTotalTokens_NoCounter(t *testing.T) {
	cw := NewContextWindow()

	// Add entries
	cw.AddEntry(HistoryEntry{
		Beat:         1,
		PlayerChoice: "Enter the room",
		StoryContent: "You see a dark hallway.",
	})

	// Should return 0 when no counter
	tokens := cw.CalculateTotalTokens()
	if tokens != 0 {
		t.Errorf("Expected 0 tokens without counter, got %d", tokens)
	}
}

// TestContextWindow_CalculateTotalTokens_WithCounter 測試有 counter 時計算 tokens
func TestContextWindow_CalculateTotalTokens_WithCounter(t *testing.T) {
	cw := NewContextWindow()
	cw.SetTokenCounter(NewEstimateTokenCounter())

	// Add multiple entries
	for i := 0; i < 3; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			PlayerChoice: "Do something",
			StoryContent: "Something happens. More text here to increase token count.",
		})
	}

	// Add summary
	cw.UpdateSummary("This is a summary of the game so far.")

	tokens := cw.CalculateTotalTokens()
	if tokens <= 0 {
		t.Error("Total tokens should be > 0 with content")
	}
}

// TestContextWindow_GetTokenUsageReport_NoCounter 測試沒有 counter 的報告
func TestContextWindow_GetTokenUsageReport_NoCounter(t *testing.T) {
	cw := NewContextWindow()

	report := cw.GetTokenUsageReport()

	if report.TotalTokens != 0 {
		t.Errorf("Expected 0 total tokens, got %d", report.TotalTokens)
	}
	if report.ModelLimit != MAX_CONTEXT_TOKENS {
		t.Errorf("Expected model limit %d, got %d", MAX_CONTEXT_TOKENS, report.ModelLimit)
	}
}

// TestContextWindow_GetTokenUsageReport_WithCounter 測試有 counter 的完整報告
func TestContextWindow_GetTokenUsageReport_WithCounter(t *testing.T) {
	cw := NewContextWindow()
	cw.SetTokenCounter(NewEstimateTokenCounter())

	// Add entries
	cw.AddEntry(HistoryEntry{
		Beat:         1,
		PlayerChoice: "Choice one",
		StoryContent: "Story content one",
	})
	cw.AddEntry(HistoryEntry{
		Beat:         2,
		PlayerChoice: "Choice two",
		StoryContent: "Story content two",
	})

	// Add summary
	cw.UpdateSummary("Summary text here")

	report := cw.GetTokenUsageReport()

	// Verify report structure
	if report.SummaryTokens <= 0 {
		t.Error("Summary tokens should be > 0")
	}
	if report.RecentEntriesTokens <= 0 {
		t.Error("Recent entries tokens should be > 0")
	}
	if report.TotalTokens != report.SummaryTokens+report.RecentEntriesTokens {
		t.Errorf("Total (%d) should equal Summary (%d) + Entries (%d)",
			report.TotalTokens, report.SummaryTokens, report.RecentEntriesTokens)
	}
	if report.ModelLimit != MAX_CONTEXT_TOKENS {
		t.Errorf("Expected model limit %d, got %d", MAX_CONTEXT_TOKENS, report.ModelLimit)
	}
	if report.UsagePercentage < 0 || report.UsagePercentage > 1 {
		t.Errorf("Usage percentage should be 0-1, got %f", report.UsagePercentage)
	}
}

// TestContextWindow_GetTokenUsageReport_WarningThreshold 測試警告閾值
func TestContextWindow_GetTokenUsageReport_WarningThreshold(t *testing.T) {
	cw := NewContextWindow()

	// Use mock counter to control token count
	mockCounter := &MockTokenCounter{
		TokensPerCall: 1000, // Each call returns 1000 tokens
	}
	cw.SetTokenCounter(mockCounter)

	// Add content to trigger warning (need > 70% of 8000 = 5600 tokens)
	// With mock returning 1000 per call, we need 6 calls: 1 summary + 5 entries
	cw.UpdateSummary("Summary")
	for i := 0; i < 5; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			PlayerChoice: "A",
			StoryContent: "B",
		})
	}

	report := cw.GetTokenUsageReport()

	// Should trigger warning (6000 tokens / 8000 = 75%)
	if !report.IsWarning {
		t.Errorf("Expected warning at %.1f%% usage", report.UsagePercentage*100)
	}
}

// TestContextWindow_GetTokenUsageReport_NoWarning 測試不觸發警告
func TestContextWindow_GetTokenUsageReport_NoWarning(t *testing.T) {
	cw := NewContextWindow()

	// Use mock counter with low token count
	mockCounter := &MockTokenCounter{
		TokensPerCall: 100, // Each call returns 100 tokens
	}
	cw.SetTokenCounter(mockCounter)

	// Add small content
	cw.AddEntry(HistoryEntry{
		Beat:         1,
		PlayerChoice: "A",
		StoryContent: "B",
	})

	report := cw.GetTokenUsageReport()

	// Should NOT trigger warning (200 tokens / 8000 = 2.5%)
	if report.IsWarning {
		t.Errorf("Should not trigger warning at %.1f%% usage", report.UsagePercentage*100)
	}
}

// TestContextWindow_CalculateTotalTokens_ThreadSafe 測試並發安全
func TestContextWindow_CalculateTotalTokens_ThreadSafe(t *testing.T) {
	cw := NewContextWindow()
	cw.SetTokenCounter(NewEstimateTokenCounter())

	// Add initial content
	cw.AddEntry(HistoryEntry{
		Beat:         1,
		StoryContent: "Initial content",
	})

	// Run concurrent calculations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = cw.CalculateTotalTokens()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestContextWindow_GetTokenUsageReport_LargeContent 測試大量內容
func TestContextWindow_GetTokenUsageReport_LargeContent(t *testing.T) {
	cw := NewContextWindow()
	cw.SetTokenCounter(NewEstimateTokenCounter())

	// Create large content
	largeText := strings.Repeat("This is a long story content. ", 100)

	for i := 0; i < 10; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			PlayerChoice: "Choice " + string(rune(i)),
			StoryContent: largeText,
		})
	}

	cw.UpdateSummary(strings.Repeat("Summary sentence. ", 50))

	report := cw.GetTokenUsageReport()

	// Should have significant token count
	if report.TotalTokens < 100 {
		t.Errorf("Expected large token count, got %d", report.TotalTokens)
	}

	// Verify calculations
	if report.TotalTokens != report.SummaryTokens+report.RecentEntriesTokens {
		t.Error("Token calculation mismatch")
	}
}

// TestContextWindow_CheckTokenUsage_NoCounter 測試沒有 counter 的檢查
func TestContextWindow_CheckTokenUsage_NoCounter(t *testing.T) {
	cw := NewContextWindow()

	tokens, limit, level := cw.CheckTokenUsage()

	if tokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", tokens)
	}
	if limit != MAX_CONTEXT_TOKENS {
		t.Errorf("Expected limit %d, got %d", MAX_CONTEXT_TOKENS, limit)
	}
	if level != 0 {
		t.Errorf("Expected warning level 0, got %d", level)
	}
}

// TestContextWindow_CheckTokenUsage_NormalLevel 測試正常級別 (< 70%)
func TestContextWindow_CheckTokenUsage_NormalLevel(t *testing.T) {
	cw := NewContextWindow()
	mockCounter := &MockTokenCounter{
		TokensPerCall: 50, // Low token count
	}
	cw.SetTokenCounter(mockCounter)

	cw.AddEntry(HistoryEntry{
		Beat:         1,
		StoryContent: "Short text",
	})

	tokens, limit, level := cw.CheckTokenUsage()

	if tokens <= 0 {
		t.Error("Expected some tokens")
	}
	if limit != MAX_CONTEXT_TOKENS {
		t.Errorf("Expected limit %d, got %d", MAX_CONTEXT_TOKENS, limit)
	}
	if level != 0 {
		t.Errorf("Expected warning level 0 (normal), got %d", level)
	}
}

// TestContextWindow_CheckTokenUsage_WarningLevel 測試警告級別 (70-90%)
func TestContextWindow_CheckTokenUsage_WarningLevel(t *testing.T) {
	cw := NewContextWindow()
	// Mock: each call returns 1000 tokens
	// Need 6 calls to reach ~6000 tokens = 75% of 8000 (safely in warning zone)
	mockCounter := &MockTokenCounter{
		TokensPerCall: 1000,
	}
	cw.SetTokenCounter(mockCounter)

	cw.UpdateSummary("Summary")
	for i := 0; i < 5; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			StoryContent: "Content",
		})
	}

	tokens, limit, level := cw.CheckTokenUsage()

	usageRatio := float64(tokens) / float64(limit)
	if usageRatio < WARNING_THRESHOLD {
		t.Errorf("Expected usage >= 70%%, got %.1f%%", usageRatio*100)
	}
	if usageRatio >= EMERGENCY_THRESHOLD {
		t.Errorf("Expected usage < 90%%, got %.1f%%", usageRatio*100)
	}
	if level != 1 {
		t.Errorf("Expected warning level 1 (warning), got %d at %.1f%% usage", level, usageRatio*100)
	}
}

// TestContextWindow_CheckTokenUsage_EmergencyLevel 測試緊急級別 (>= 90%)
func TestContextWindow_CheckTokenUsage_EmergencyLevel(t *testing.T) {
	cw := NewContextWindow()
	// Mock: each call returns 1500 tokens
	// Need 6 calls to reach ~9000 tokens = 112% of 8000 (emergency)
	mockCounter := &MockTokenCounter{
		TokensPerCall: 1500,
	}
	cw.SetTokenCounter(mockCounter)

	cw.UpdateSummary("Summary")
	for i := 0; i < 5; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			StoryContent: "Content",
		})
	}

	tokens, limit, level := cw.CheckTokenUsage()

	usageRatio := float64(tokens) / float64(limit)
	if usageRatio < EMERGENCY_THRESHOLD {
		t.Errorf("Expected usage >= 90%%, got %.1f%%", usageRatio*100)
	}
	if level != 2 {
		t.Errorf("Expected warning level 2 (emergency), got %d at %.1f%% usage", level, usageRatio*100)
	}
}

// TestContextWindow_ShrinkWindow_Success 測試成功縮小窗口
func TestContextWindow_ShrinkWindow_Success(t *testing.T) {
	cw := NewContextWindow()
	initialSize := cw.WindowSize // Should be 5 (DEFAULT_WINDOW_SIZE)

	// Add entries
	for i := 0; i < 10; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			StoryContent: "Content",
		})
	}

	// Shrink window
	shrunk := cw.ShrinkWindow()

	if !shrunk {
		t.Error("Expected window to shrink successfully")
	}
	if cw.WindowSize != initialSize-1 {
		t.Errorf("Expected window size %d, got %d", initialSize-1, cw.WindowSize)
	}

	// Verify RecentEntries updated
	if len(cw.GetWindow()) != cw.WindowSize {
		t.Error("RecentEntries should match new window size")
	}
}

// TestContextWindow_ShrinkWindow_AtMinimum 測試已經最小窗口時不再縮小
func TestContextWindow_ShrinkWindow_AtMinimum(t *testing.T) {
	cw := NewContextWindow()

	// Shrink to minimum
	for cw.WindowSize > MIN_WINDOW_SIZE {
		cw.ShrinkWindow()
	}

	// Try to shrink further
	shrunk := cw.ShrinkWindow()

	if shrunk {
		t.Error("Should not shrink below MIN_WINDOW_SIZE")
	}
	if cw.WindowSize != MIN_WINDOW_SIZE {
		t.Errorf("Expected window size %d, got %d", MIN_WINDOW_SIZE, cw.WindowSize)
	}
}

// TestContextWindow_ShrinkWindow_ThreadSafe 測試 ShrinkWindow 並發安全
func TestContextWindow_ShrinkWindow_ThreadSafe(t *testing.T) {
	cw := NewContextWindow()

	// Add entries
	for i := 0; i < 10; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i + 1,
			StoryContent: "Content",
		})
	}

	// Concurrent shrinking
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			cw.ShrinkWindow()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should be at minimum (shrunk multiple times)
	if cw.WindowSize < MIN_WINDOW_SIZE {
		t.Errorf("Window size should not go below MIN_WINDOW_SIZE (%d), got %d", MIN_WINDOW_SIZE, cw.WindowSize)
	}
}
