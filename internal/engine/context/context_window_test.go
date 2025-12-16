package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestNewContextWindow 測試建構函數
func TestNewContextWindow(t *testing.T) {
	cw := NewContextWindow()

	if cw == nil {
		t.Fatal("NewContextWindow() returned nil")
	}

	if cw.WindowSize != 5 {
		t.Errorf("Expected WindowSize = 5, got %d", cw.WindowSize)
	}

	if cw.Summary != "" {
		t.Errorf("Expected empty Summary, got %q", cw.Summary)
	}

	if len(cw.RecentEntries) != 0 {
		t.Errorf("Expected empty RecentEntries, got %d entries", len(cw.RecentEntries))
	}
}

// TestHistoryEntry_Serialization 測試 HistoryEntry 序列化
func TestHistoryEntry_Serialization(t *testing.T) {
	entry := HistoryEntry{
		Beat:           1,
		PlayerChoice:   "進入大廳",
		StoryContent:   "你看到一片血跡...",
		HPChange:       -10,
		SANChange:      -5,
		RulesTriggered: []string{"rule-001", "rule-002"},
		CluesFound:     []string{"clue-blood"},
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal HistoryEntry: %v", err)
	}

	// Unmarshal from JSON
	var decoded HistoryEntry
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal HistoryEntry: %v", err)
	}

	// Verify all fields
	if decoded.Beat != entry.Beat {
		t.Errorf("Beat mismatch: got %d, want %d", decoded.Beat, entry.Beat)
	}
	if decoded.PlayerChoice != entry.PlayerChoice {
		t.Errorf("PlayerChoice mismatch: got %q, want %q", decoded.PlayerChoice, entry.PlayerChoice)
	}
	if decoded.StoryContent != entry.StoryContent {
		t.Errorf("StoryContent mismatch: got %q, want %q", decoded.StoryContent, entry.StoryContent)
	}
	if decoded.HPChange != entry.HPChange {
		t.Errorf("HPChange mismatch: got %d, want %d", decoded.HPChange, entry.HPChange)
	}
	if decoded.SANChange != entry.SANChange {
		t.Errorf("SANChange mismatch: got %d, want %d", decoded.SANChange, entry.SANChange)
	}

	// Verify array contents (not just length)
	if len(decoded.RulesTriggered) != len(entry.RulesTriggered) {
		t.Fatalf("RulesTriggered length mismatch: got %d, want %d", len(decoded.RulesTriggered), len(entry.RulesTriggered))
	}
	for i, rule := range entry.RulesTriggered {
		if decoded.RulesTriggered[i] != rule {
			t.Errorf("RulesTriggered[%d] mismatch: got %q, want %q", i, decoded.RulesTriggered[i], rule)
		}
	}

	if len(decoded.CluesFound) != len(entry.CluesFound) {
		t.Fatalf("CluesFound length mismatch: got %d, want %d", len(decoded.CluesFound), len(entry.CluesFound))
	}
	for i, clue := range entry.CluesFound {
		if decoded.CluesFound[i] != clue {
			t.Errorf("CluesFound[%d] mismatch: got %q, want %q", i, decoded.CluesFound[i], clue)
		}
	}
}

// TestContextWindow_Serialization 測試 ContextWindow 序列化
func TestContextWindow_Serialization(t *testing.T) {
	cw := &ContextWindow{
		Summary: "Chapter 1: 進入廢棄醫院",
		RecentEntries: []HistoryEntry{
			{Beat: 1, PlayerChoice: "前進", StoryContent: "你進入大廳"},
			{Beat: 2, PlayerChoice: "調查", StoryContent: "你發現血跡"},
		},
		WindowSize: 5,
	}

	// Marshal to JSON
	data, err := json.Marshal(cw)
	if err != nil {
		t.Fatalf("Failed to marshal ContextWindow: %v", err)
	}

	// Unmarshal from JSON
	var decoded ContextWindow
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ContextWindow: %v", err)
	}

	// Verify fields
	if decoded.Summary != cw.Summary {
		t.Errorf("Summary mismatch: got %q, want %q", decoded.Summary, cw.Summary)
	}
	if decoded.WindowSize != cw.WindowSize {
		t.Errorf("WindowSize mismatch: got %d, want %d", decoded.WindowSize, cw.WindowSize)
	}
	if len(decoded.RecentEntries) != len(cw.RecentEntries) {
		t.Errorf("RecentEntries length mismatch: got %d, want %d", len(decoded.RecentEntries), len(cw.RecentEntries))
	}
}

// TestContextWindow_ConcurrentAccess 測試並發安全性
func TestContextWindow_ConcurrentAccess(t *testing.T) {
	cw := NewContextWindow()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(beat int) {
			defer wg.Done()
			// This will test concurrent access - implementation needs mutex
			_ = beat
		}(i)
	}

	// Concurrent reads
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = cw.GetWindow()
			_ = cw.GetSummary()
		}()
	}

	wg.Wait()
}

// TestGetWindow 測試獲取窗口
func TestGetWindow(t *testing.T) {
	cw := NewContextWindow()
	cw.RecentEntries = []HistoryEntry{
		{Beat: 1, PlayerChoice: "A"},
		{Beat: 2, PlayerChoice: "B"},
	}

	window := cw.GetWindow()
	if len(window) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(window))
	}
}

// TestGetSummary 測試獲取摘要
func TestGetSummary(t *testing.T) {
	cw := NewContextWindow()
	cw.Summary = "Test Summary"

	summary := cw.GetSummary()
	if summary != "Test Summary" {
		t.Errorf("Expected 'Test Summary', got %q", summary)
	}
}

// TestContextWindow_NilSafety 測試 nil 安全性
func TestContextWindow_NilSafety(t *testing.T) {
	// Simulate incorrect manual construction (without NewContextWindow)
	cw := &ContextWindow{
		Summary:       "",
		RecentEntries: nil, // nil slice
		WindowSize:    5,
	}

	// GetWindow should not panic and return empty slice
	window := cw.GetWindow()
	if window == nil {
		t.Error("GetWindow() returned nil, expected empty slice")
	}
	if len(window) != 0 {
		t.Errorf("Expected empty window, got %d entries", len(window))
	}

	// GetSummary should work fine
	summary := cw.GetSummary()
	if summary != "" {
		t.Errorf("Expected empty summary, got %q", summary)
	}
}

// ============================================================================
// Story 5.2: Sliding Window Mechanism Tests
// ============================================================================

// TestAddEntry_SingleEntry 測試添加單個條目
func TestAddEntry_SingleEntry(t *testing.T) {
	cw := NewContextWindow()

	entry := HistoryEntry{
		Beat:         1,
		PlayerChoice: "進入大廳",
		StoryContent: "你看到血跡...",
	}

	err := cw.AddEntry(entry)
	if err != nil {
		t.Fatalf("AddEntry() failed: %v", err)
	}

	// 驗證 AllEntries
	allEntries := cw.GetAllEntries()
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 entry in AllEntries, got %d", len(allEntries))
	}
	if allEntries[0].Beat != 1 {
		t.Errorf("Expected Beat=1, got Beat=%d", allEntries[0].Beat)
	}

	// 驗證 RecentEntries
	window := cw.GetWindow()
	if len(window) != 1 {
		t.Errorf("Expected 1 entry in window, got %d", len(window))
	}
	if window[0].Beat != 1 {
		t.Errorf("Expected Beat=1 in window, got Beat=%d", window[0].Beat)
	}
}

// TestAddEntry_WindowSliding 測試窗口滑動
func TestAddEntry_WindowSliding(t *testing.T) {
	cw := NewContextWindow()

	// 添加 10 個條目
	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("Choice %d", i),
			StoryContent: fmt.Sprintf("Story %d", i),
		}
		err := cw.AddEntry(entry)
		if err != nil {
			t.Fatalf("AddEntry(%d) failed: %v", i, err)
		}
	}

	// 驗證 AllEntries 有 10 個
	allEntries := cw.GetAllEntries()
	if len(allEntries) != 10 {
		t.Errorf("Expected 10 entries in AllEntries, got %d", len(allEntries))
	}

	// 驗證 RecentEntries 只有最後 5 個
	window := cw.GetWindow()
	if len(window) != 5 {
		t.Errorf("Expected 5 entries in window, got %d", len(window))
	}

	// 驗證窗口內容是 Beat 6-10
	expectedBeats := []int{6, 7, 8, 9, 10}
	for i, expectedBeat := range expectedBeats {
		if window[i].Beat != expectedBeat {
			t.Errorf("Window[%d]: expected Beat=%d, got Beat=%d", i, expectedBeat, window[i].Beat)
		}
	}

	// 驗證第一個條目 (Beat=1) 仍在 AllEntries
	if allEntries[0].Beat != 1 {
		t.Errorf("Expected first entry Beat=1, got Beat=%d", allEntries[0].Beat)
	}
}

// TestGetWindow_ReturnsCopy 測試返回副本
func TestGetWindow_ReturnsCopy(t *testing.T) {
	cw := NewContextWindow()

	// 添加 2 個條目
	cw.AddEntry(HistoryEntry{Beat: 1, PlayerChoice: "A"})
	cw.AddEntry(HistoryEntry{Beat: 2, PlayerChoice: "B"})

	// 獲取窗口
	window := cw.GetWindow()
	originalLength := len(window)

	// 修改返回的切片
	window[0].PlayerChoice = "MODIFIED"
	window = append(window, HistoryEntry{Beat: 999, PlayerChoice: "FAKE"})

	// 驗證原始數據未被修改
	newWindow := cw.GetWindow()
	if len(newWindow) != originalLength {
		t.Errorf("Original window length changed: expected %d, got %d", originalLength, len(newWindow))
	}
	if newWindow[0].PlayerChoice != "A" {
		t.Errorf("Original data was modified: expected 'A', got %q", newWindow[0].PlayerChoice)
	}
}

// TestEdgeCases 測試邊界情況
func TestEdgeCases(t *testing.T) {
	t.Run("Empty history", func(t *testing.T) {
		cw := NewContextWindow()
		window := cw.GetWindow()
		if len(window) != 0 {
			t.Errorf("Expected empty window, got %d entries", len(window))
		}
		allEntries := cw.GetAllEntries()
		if len(allEntries) != 0 {
			t.Errorf("Expected empty AllEntries, got %d entries", len(allEntries))
		}
	})

	t.Run("Exactly WindowSize entries", func(t *testing.T) {
		cw := NewContextWindow()
		// 添加剛好 5 個條目
		for i := 1; i <= 5; i++ {
			cw.AddEntry(HistoryEntry{Beat: i, PlayerChoice: fmt.Sprintf("C%d", i)})
		}

		window := cw.GetWindow()
		allEntries := cw.GetAllEntries()

		// 窗口應該等於全部歷史
		if len(window) != 5 {
			t.Errorf("Expected 5 entries in window, got %d", len(window))
		}
		if len(allEntries) != 5 {
			t.Errorf("Expected 5 entries in AllEntries, got %d", len(allEntries))
		}
		if window[0].Beat != 1 {
			t.Errorf("Expected first entry Beat=1, got %d", window[0].Beat)
		}
	})

	t.Run("Less than WindowSize entries", func(t *testing.T) {
		cw := NewContextWindow()
		// 只添加 3 個條目
		for i := 1; i <= 3; i++ {
			cw.AddEntry(HistoryEntry{Beat: i, PlayerChoice: fmt.Sprintf("C%d", i)})
		}

		window := cw.GetWindow()
		if len(window) != 3 {
			t.Errorf("Expected 3 entries in window, got %d", len(window))
		}
	})
}

// TestConcurrentAddEntry 測試並發添加
func TestConcurrentAddEntry(t *testing.T) {
	cw := NewContextWindow()

	var wg sync.WaitGroup
	goroutines := 10
	entriesPerGoroutine := 100

	// 10 個 goroutines 各添加 100 個條目
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < entriesPerGoroutine; i++ {
				entry := HistoryEntry{
					Beat:         goroutineID*entriesPerGoroutine + i,
					PlayerChoice: fmt.Sprintf("G%d-C%d", goroutineID, i),
					StoryContent: fmt.Sprintf("Story from goroutine %d", goroutineID),
				}
				cw.AddEntry(entry)
			}
		}(g)
	}

	wg.Wait()

	// 驗證最終有 1000 個條目
	allEntries := cw.GetAllEntries()
	expectedTotal := goroutines * entriesPerGoroutine
	if len(allEntries) != expectedTotal {
		t.Errorf("Expected %d total entries, got %d", expectedTotal, len(allEntries))
	}

	// 驗證窗口只有最後 5 個
	window := cw.GetWindow()
	if len(window) != 5 {
		t.Errorf("Expected 5 entries in window, got %d", len(window))
	}
}

// TestGetAllEntries 測試獲取完整歷史
func TestGetAllEntries(t *testing.T) {
	cw := NewContextWindow()

	// 添加 15 個條目
	for i := 1; i <= 15; i++ {
		cw.AddEntry(HistoryEntry{Beat: i, PlayerChoice: fmt.Sprintf("C%d", i)})
	}

	allEntries := cw.GetAllEntries()

	// 驗證有 15 個條目
	if len(allEntries) != 15 {
		t.Errorf("Expected 15 entries, got %d", len(allEntries))
	}

	// 驗證順序正確
	for i, entry := range allEntries {
		expectedBeat := i + 1
		if entry.Beat != expectedBeat {
			t.Errorf("AllEntries[%d]: expected Beat=%d, got Beat=%d", i, expectedBeat, entry.Beat)
		}
	}
}

// TestGetEntryCount 測試獲取條目數
func TestGetEntryCount(t *testing.T) {
	cw := NewContextWindow()

	if cw.GetEntryCount() != 0 {
		t.Errorf("Expected 0 entries initially, got %d", cw.GetEntryCount())
	}

	// 添加 7 個條目
	for i := 1; i <= 7; i++ {
		cw.AddEntry(HistoryEntry{Beat: i})
	}

	if cw.GetEntryCount() != 7 {
		t.Errorf("Expected 7 entries, got %d", cw.GetEntryCount())
	}
}

// TestClear 測試清空歷史
func TestClear(t *testing.T) {
	cw := NewContextWindow()

	// 添加一些條目
	for i := 1; i <= 10; i++ {
		cw.AddEntry(HistoryEntry{Beat: i})
	}

	// 清空
	cw.Clear()

	// 驗證清空成功
	if cw.GetEntryCount() != 0 {
		t.Errorf("Expected 0 entries after Clear(), got %d", cw.GetEntryCount())
	}

	window := cw.GetWindow()
	if len(window) != 0 {
		t.Errorf("Expected empty window after Clear(), got %d entries", len(window))
	}

	allEntries := cw.GetAllEntries()
	if len(allEntries) != 0 {
		t.Errorf("Expected empty AllEntries after Clear(), got %d entries", len(allEntries))
	}
}

// TestGetLastEntry 測試獲取最新條目
func TestGetLastEntry(t *testing.T) {
	cw := NewContextWindow()

	// 空歷史應該返回 nil
	lastEntry, ok := cw.GetLastEntry()
	if ok {
		t.Error("Expected ok=false for empty history")
	}

	// 添加條目
	cw.AddEntry(HistoryEntry{Beat: 1, PlayerChoice: "First"})
	cw.AddEntry(HistoryEntry{Beat: 2, PlayerChoice: "Second"})
	cw.AddEntry(HistoryEntry{Beat: 3, PlayerChoice: "Third"})

	// 應該返回最後一個
	lastEntry, ok = cw.GetLastEntry()
	if !ok {
		t.Fatal("Expected ok=true for non-empty history")
	}
	if lastEntry.Beat != 3 {
		t.Errorf("Expected last entry Beat=3, got Beat=%d", lastEntry.Beat)
	}
	if lastEntry.PlayerChoice != "Third" {
		t.Errorf("Expected PlayerChoice='Third', got %q", lastEntry.PlayerChoice)
	}
}

// TestContextWindow_RecentEntries_Independence 測試 RecentEntries 獨立性（修復 aliasing bug）
func TestContextWindow_RecentEntries_Independence(t *testing.T) {
	cw := NewContextWindow()

	// 添加少於 WindowSize 的條目
	cw.AddEntry(HistoryEntry{Beat: 1, PlayerChoice: "Original"})

	// 獲取窗口快照
	window := cw.GetWindow()
	originalChoice := window[0].PlayerChoice

	// 直接修改 AllEntries
	cw.AllEntries[0].PlayerChoice = "HACKED"

	// RecentEntries 和之前的快照應該不受影響
	if originalChoice != "Original" {
		t.Errorf("snapshot should be independent, got %q", originalChoice)
	}
	if cw.RecentEntries[0].PlayerChoice != "Original" {
		t.Errorf("RecentEntries should be independent from AllEntries, got %q", cw.RecentEntries[0].PlayerChoice)
	}
}

// TestContextWindow_SlicingAlsoCopies 測試切片操作也使用拷貝
func TestContextWindow_SlicingAlsoCopies(t *testing.T) {
	cw := NewContextWindow()

	// 添加超過 WindowSize 的條目
	for i := 1; i <= 5; i++ {
		cw.AddEntry(HistoryEntry{Beat: i, PlayerChoice: fmt.Sprintf("Choice%d", i)})
	}

	// RecentEntries 應該是 [1,2,3,4,5]
	if len(cw.RecentEntries) != 5 {
		t.Fatalf("Expected 5 entries in RecentEntries, got %d", len(cw.RecentEntries))
	}

	// 修改 AllEntries 中對應的位置
	cw.AllEntries[2].PlayerChoice = "HACKED3" // 原本的 Choice3

	// RecentEntries 不應受影響
	if cw.RecentEntries[2].PlayerChoice != "Choice3" {
		t.Errorf("RecentEntries should be a copy, not a slice reference, got %q", cw.RecentEntries[2].PlayerChoice)
	}
}

// TestContextWindow_WindowSizeZero 測試 WindowSize=0 的邊界情況
func TestContextWindow_WindowSizeZero(t *testing.T) {
	cw := NewContextWindow()
	cw.WindowSize = 0

	cw.AddEntry(HistoryEntry{Beat: 1})
	cw.AddEntry(HistoryEntry{Beat: 2})

	// WindowSize=0 應該返回空窗口，不應 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("WindowSize=0 caused panic: %v", r)
		}
	}()

	window := cw.GetWindow()
	if len(window) != 0 {
		t.Errorf("Expected empty window for WindowSize=0, got %d entries", len(window))
	}
}

// TestContextWindow_WindowSizeNegative 測試負數 WindowSize
func TestContextWindow_WindowSizeNegative(t *testing.T) {
	cw := NewContextWindow()
	cw.WindowSize = -5

	cw.AddEntry(HistoryEntry{Beat: 1})

	// 負數 WindowSize 應該返回空窗口，不應 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Negative WindowSize caused panic: %v", r)
		}
	}()

	window := cw.GetWindow()
	if len(window) != 0 {
		t.Errorf("Expected empty window for negative WindowSize, got %d entries", len(window))
	}
}

// TestContextWindow_EdgeCases_Comprehensive 綜合邊界測試
func TestContextWindow_EdgeCases_Comprehensive(t *testing.T) {
	tests := []struct {
		name       string
		windowSize int
		entries    int
		wantLen    int
	}{
		{"WindowSize=0, no entries", 0, 0, 0},
		{"WindowSize=0, has entries", 0, 5, 0},
		{"WindowSize=-1", -1, 5, 0},
		{"WindowSize=-100", -100, 5, 0},
		{"WindowSize=1, one entry", 1, 1, 1},
		{"WindowSize=5, exactly 5 entries", 5, 5, 5},
		{"WindowSize=5, less than 5 entries", 5, 3, 3},
		{"WindowSize=5, more than 5 entries", 5, 10, 5},
		{"WindowSize=1000, few entries", 1000, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := NewContextWindow()
			cw.WindowSize = tt.windowSize

			for i := 0; i < tt.entries; i++ {
				cw.AddEntry(HistoryEntry{Beat: i + 1})
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test panicked: %v", r)
				}
			}()

			window := cw.GetWindow()
			if len(window) != tt.wantLen {
				t.Errorf("Expected window length %d, got %d", tt.wantLen, len(window))
			}
		})
	}
}

// TestContextWindow_IndependenceSuite 完整獨立性測試套件
func TestContextWindow_IndependenceSuite(t *testing.T) {
	scenarios := []struct {
		name       string
		windowSize int
		entries    int
	}{
		{"exactly window size", 5, 5},
		{"less than window size", 5, 3},
		{"more than window size", 5, 10},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			cw := NewContextWindow()
			cw.WindowSize = sc.windowSize

			for i := 0; i < sc.entries; i++ {
				cw.AddEntry(HistoryEntry{
					Beat:         i + 1,
					PlayerChoice: fmt.Sprintf("Original_%d", i),
				})
			}

			// 獲取窗口
			window := cw.GetWindow()

			// 修改 AllEntries
			for i := range cw.AllEntries {
				cw.AllEntries[i].PlayerChoice = "HACKED"
			}

			// 驗證 window（之前的快照）不受影響
			for i, entry := range window {
				if entry.PlayerChoice == "HACKED" {
					t.Errorf("GetWindow() snapshot should be independent, index=%d", i)
				}
			}

			// 驗證 RecentEntries 也不受影響
			for i, entry := range cw.RecentEntries {
				if entry.PlayerChoice == "HACKED" {
					t.Errorf("RecentEntries should be independent, index=%d", i)
				}
			}
		})
	}
}

// ============================================================================
// Story 5.3: Smart Summary Generation Tests
// ============================================================================

// TestContextWindow_SummaryTrigger 測試摘要觸發機制
func TestContextWindow_SummaryTrigger(t *testing.T) {
	cw := NewContextWindow()

	// 初始狀態：不應觸發摘要
	if cw.ShouldGenerateSummary() {
		t.Error("Should not trigger summary with 0 entries")
	}

	// 添加 14 個條目：不應觸發
	for i := 1; i <= 14; i++ {
		cw.AddEntry(HistoryEntry{Beat: i, PlayerChoice: fmt.Sprintf("Choice %d", i)})
	}
	if cw.ShouldGenerateSummary() {
		t.Error("Should not trigger summary with 14 entries (threshold is 15)")
	}

	// 添加第 15 個條目：應該觸發
	cw.AddEntry(HistoryEntry{Beat: 15, PlayerChoice: "Choice 15"})
	if !cw.ShouldGenerateSummary() {
		t.Error("Should trigger summary at 15 entries")
	}
}

// TestContextWindow_SummaryInProgressPrevention 測試防止重複觸發
func TestContextWindow_SummaryInProgressPrevention(t *testing.T) {
	cw := NewContextWindow()

	// 添加 15 個條目
	for i := 1; i <= 15; i++ {
		cw.AddEntry(HistoryEntry{Beat: i})
	}

	// 設置摘要生成中
	cw.IsSummaryInProgress = true

	// 即使條目達到閾值，也不應觸發
	if cw.ShouldGenerateSummary() {
		t.Error("Should not trigger when summary is already in progress")
	}
}

// TestContextWindow_LastSummaryBeat 測試上次摘要回合追踪
func TestContextWindow_LastSummaryBeat(t *testing.T) {
	cw := NewContextWindow()

	// 添加 15 個條目並生成摘要
	for i := 1; i <= 15; i++ {
		cw.AddEntry(HistoryEntry{Beat: i})
	}

	// 模擬摘要完成
	cw.LastSummaryBeat = 15

	// 添加更多條目（16-29），不應觸發
	for i := 16; i <= 29; i++ {
		cw.AddEntry(HistoryEntry{Beat: i})
	}
	if cw.ShouldGenerateSummary() {
		t.Error("Should not trigger: only 14 entries since last summary")
	}

	// 添加第 30 個條目：應該觸發（15 entries since last summary）
	cw.AddEntry(HistoryEntry{Beat: 30})
	if !cw.ShouldGenerateSummary() {
		t.Error("Should trigger at beat 30 (15 entries since beat 15)")
	}
}

// TestContextWindow_UpdateSummary 測試摘要更新
func TestContextWindow_UpdateSummary(t *testing.T) {
	cw := NewContextWindow()

	// 首次摘要
	cw.UpdateSummary("[Chapter 1 Summary: First summary.]")

	if cw.Summary != "[Chapter 1 Summary: First summary.]" {
		t.Error("Summary should be set correctly")
	}
	if cw.LastSummaryBeat != 0 {
		t.Error("LastSummaryBeat should be 0 initially")
	}
	if cw.IsSummaryInProgress {
		t.Error("IsSummaryInProgress should be false after update")
	}

	// 添加更多條目
	for i := 1; i <= 15; i++ {
		cw.AddEntry(HistoryEntry{Beat: i})
	}

	// 第二次摘要（應該合併）
	cw.UpdateSummary("[Chapter 2 Summary: Second summary.]")

	expectedSummary := "[Chapter 1 Summary: First summary.]\n\n[Chapter 2 Summary: Second summary.]"
	if cw.Summary != expectedSummary {
		t.Errorf("Summary should be merged. Got: %q", cw.Summary)
	}
	if cw.LastSummaryBeat != 15 {
		t.Errorf("LastSummaryBeat should be 15, got %d", cw.LastSummaryBeat)
	}
}

// TestContextWindow_SummaryGeneration_Integration 測試完整摘要生成流程
func TestContextWindow_SummaryGeneration_Integration(t *testing.T) {
	cw := NewContextWindow()
	mockGen := &MockSummaryGenerator{
		SummaryToReturn: "[Chapter 1 Summary: 玩家進入廢棄醫院. 角色狀態: 全員存活. 已知線索: 血跡. 當前目標: 尋找出口.]",
	}

	// 添加 15 個條目
	for i := 1; i <= 15; i++ {
		cw.AddEntry(HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("Choice %d", i),
			StoryContent: fmt.Sprintf("Story %d", i),
			CluesFound:   []string{fmt.Sprintf("clue-%d", i)},
		})
	}

	// 檢查應該觸發摘要
	if !cw.ShouldGenerateSummary() {
		t.Fatal("Should trigger summary after 15 entries")
	}

	// 設置生成中標誌
	cw.SetSummaryInProgress(true)

	// 模擬異步生成摘要
	entriesForSummary := cw.GetAllEntries()
	summary, err := mockGen.GenerateSummary(context.Background(), entriesForSummary)
	if err != nil {
		t.Fatalf("Failed to generate summary: %v", err)
	}

	// 更新摘要
	cw.UpdateSummary(summary)

	// 驗證結果
	if cw.Summary == "" {
		t.Error("Summary should not be empty")
	}
	if cw.LastSummaryBeat != 15 {
		t.Errorf("LastSummaryBeat should be 15, got %d", cw.LastSummaryBeat)
	}
	if cw.IsSummaryInProgress {
		t.Error("IsSummaryInProgress should be false after completion")
	}
	if !strings.Contains(cw.Summary, "Chapter") {
		t.Error("Summary should contain Chapter format")
	}

	// 驗證 mock 被調用
	if mockGen.CallCount != 1 {
		t.Errorf("Expected 1 call to generator, got %d", mockGen.CallCount)
	}
}
