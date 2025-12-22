package knowledge

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkCachePut benchmarks cache put operations
func BenchmarkCachePut(b *testing.B) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 1000,
		TTL:     30 * time.Minute,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        7,
		Type:            "direct",
		Explanation:     "Test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put(fmt.Sprintf("statement_%d_a", i), fmt.Sprintf("statement_%d_b", i), result)
	}
}

// BenchmarkCacheGet benchmarks cache get operations
func BenchmarkCacheGet(b *testing.B) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 1000,
		TTL:     30 * time.Minute,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        7,
		Type:            "direct",
		Explanation:     "Test",
	}

	// Populate cache
	for i := 0; i < 100; i++ {
		cache.Put(fmt.Sprintf("statement_%d_a", i), fmt.Sprintf("statement_%d_b", i), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("statement_%d_a", i%100), fmt.Sprintf("statement_%d_b", i%100))
	}
}

// BenchmarkCacheHitRate benchmarks cache hit rate in realistic scenario
func BenchmarkCacheHitRate(b *testing.B) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 1000,
		TTL:     30 * time.Minute,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        7,
		Type:            "direct",
		Explanation:     "Test",
	}

	// Simulate realistic workload: 20 unique contradictions
	uniqueContradictions := 20

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 80% of requests hit same 20 contradictions
		idx := i % uniqueContradictions
		statementA := fmt.Sprintf("statement_%d_a", idx)
		statementB := fmt.Sprintf("statement_%d_b", idx)

		if cache.Get(statementA, statementB) == nil {
			cache.Put(statementA, statementB, result)
		}
	}

	stats := cache.GetStats()
	b.ReportMetric(stats.HitRate, "hit_rate_%")
	b.ReportMetric(float64(stats.CacheHits), "cache_hits")
	b.ReportMetric(float64(stats.CacheMisses), "cache_misses")
}

// BenchmarkLLMAnalyzer benchmarks LLM-based contradiction analysis
func BenchmarkLLMAnalyzer(b *testing.B) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 7,
			"type": "direct",
			"explanation": "Test"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Told,
		Confidence:  0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeContradiction(context.Background(), "Statement A", "Statement B", knownFact)
	}
}

// BenchmarkLLMAnalyzerWithCache benchmarks analyzer with cache enabled
func BenchmarkLLMAnalyzerWithCache(b *testing.B) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 7,
			"type": "direct",
			"explanation": "Test"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Told,
		Confidence:  0.7,
	}

	// Use same statements to trigger cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeContradiction(context.Background(), "Statement A", "Statement B", knownFact)
	}

	stats := analyzer.GetCacheStats()
	b.ReportMetric(stats.HitRate, "cache_hit_rate_%")
	b.ReportMetric(float64(mockProvider.callCount), "llm_calls")
}

// BenchmarkUpdateManagerWithLLM benchmarks UpdateManager with LLM analyzer
func BenchmarkUpdateManagerWithLLM(b *testing.B) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 7,
			"type": "direct",
			"explanation": "Test"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	config := DefaultUpdateManagerConfig()
	config.ContradictionAnalyzer = analyzer
	manager := NewUpdateManager(config)

	// Setup test data
	fact := &Fact{
		ID:      "fact_001",
		Content: "張醫生活著",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Witness,
		Confidence:       1.0,
		PropagationDepth: 0,
	}
	manager.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckContradiction("npc_001", "張醫生死了")
	}

	stats := manager.GetContradictionCacheStats()
	if stats != nil {
		b.ReportMetric(stats.HitRate, "cache_hit_rate_%")
	}
}

// BenchmarkUpdateManagerKeywordOnly benchmarks UpdateManager without LLM (keyword-only)
func BenchmarkUpdateManagerKeywordOnly(b *testing.B) {
	manager := NewUpdateManager(nil) // No analyzer

	// Setup test data
	fact := &Fact{
		ID:      "fact_001",
		Content: "張醫生活著",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Witness,
		Confidence:       1.0,
		PropagationDepth: 0,
	}
	manager.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckContradiction("npc_001", "張醫生死了")
	}
}

// BenchmarkCacheLRUEviction benchmarks LRU eviction performance
func BenchmarkCacheLRUEviction(b *testing.B) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 100, // Small cache to trigger frequent evictions
		TTL:     30 * time.Minute,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        7,
		Type:            "direct",
		Explanation:     "Test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add many unique entries to trigger evictions
		cache.Put(fmt.Sprintf("a_%d", i), fmt.Sprintf("b_%d", i), result)
	}

	stats := cache.GetStats()
	b.ReportMetric(float64(stats.Evictions), "evictions")
}

// BenchmarkParseAnalysisResponse benchmarks JSON parsing
func BenchmarkParseAnalysisResponse(b *testing.B) {
	analyzer := NewContradictionAnalyzer(nil)

	content := `{
		"is_contradictory": true,
		"severity": 7,
		"type": "indirect",
		"explanation": "Test explanation with some detail"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.parseAnalysisResponse(content)
	}
}

// BenchmarkComputeHash benchmarks hash computation
func BenchmarkComputeHash(b *testing.B) {
	cache := NewContradictionCache(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.computeHash("This is a longer statement that might be typical in the game", "This is another statement that contradicts the first")
	}
}
