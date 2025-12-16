package context

import (
	"fmt"
	"testing"
)

// BenchmarkFormatCompleteContext benchmarks the FormatCompleteContext method
func BenchmarkFormatCompleteContext(b *testing.B) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add entries
	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:           i,
			PlayerChoice:   fmt.Sprintf("選擇 %d", i),
			StoryContent:   fmt.Sprintf("故事內容 %d - 這是一段較長的故事文本用於測試性能", i),
			CluesFound:     []string{fmt.Sprintf("clue-%d", i)},
			RulesTriggered: []string{fmt.Sprintf("rule-%d", i)},
		}
		cm.AddHistoryEntry(entry)
	}

	// Add summary
	cm.window.Summary = "[Chapter 1 Summary: 玩家已經完成了第一章的探索，發現了重要線索。]"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.FormatCompleteContext()
	}
}

// BenchmarkAddHistoryEntry benchmarks adding history entries
func BenchmarkAddHistoryEntry(b *testing.B) {
	cm, _ := NewContextManager(ContextConfig{
		WindowSize:         5,
		SummaryTrigger:     15,
		TokenLimit:         8000,
		ModelName:          "gpt-4",
		EnableAutoOptimize: false, // Disable for fair benchmark
	})

	entry := HistoryEntry{
		Beat:         1,
		PlayerChoice: "測試選擇",
		StoryContent: "測試故事內容",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new manager each iteration to avoid state accumulation
		if i%100 == 0 {
			cm, _ = NewContextManager(DefaultContextConfig())
		}
		cm.AddHistoryEntry(entry)
	}
}

// BenchmarkGetOptimizedContext benchmarks context retrieval
func BenchmarkGetOptimizedContext(b *testing.B) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add some entries
	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事",
		}
		cm.AddHistoryEntry(entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.GetOptimizedContext()
	}
}

// BenchmarkGetContextMetadata benchmarks metadata retrieval
func BenchmarkGetContextMetadata(b *testing.B) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add entries
	for i := 1; i <= 10; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: "選擇",
			StoryContent: "故事內容包含一些文字",
		}
		cm.AddHistoryEntry(entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.GetContextMetadata()
	}
}

// BenchmarkSaveContext benchmarks context serialization
func BenchmarkSaveContext(b *testing.B) {
	cm, _ := NewContextManager(DefaultContextConfig())

	// Add entries
	for i := 1; i <= 20; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("選擇 %d", i),
			StoryContent: fmt.Sprintf("故事內容 %d", i),
		}
		cm.AddHistoryEntry(entry)
	}

	tempFile := "/tmp/bench_context.json"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.SaveContext(tempFile)
	}
}

// BenchmarkLoadContext benchmarks context deserialization
func BenchmarkLoadContext(b *testing.B) {
	// Prepare file
	cm1, _ := NewContextManager(DefaultContextConfig())
	for i := 1; i <= 20; i++ {
		entry := HistoryEntry{
			Beat:         i,
			PlayerChoice: fmt.Sprintf("選擇 %d", i),
			StoryContent: fmt.Sprintf("故事內容 %d", i),
		}
		cm1.AddHistoryEntry(entry)
	}

	tempFile := "/tmp/bench_context.json"
	cm1.SaveContext(tempFile)

	cm2, _ := NewContextManager(DefaultContextConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm2.LoadContext(tempFile)
	}
}
