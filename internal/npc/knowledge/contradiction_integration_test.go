package knowledge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateManager_WithLLMAnalyzer tests UpdateManager with LLM-based contradiction detection
func TestUpdateManager_WithLLMAnalyzer(t *testing.T) {
	// Create mock LLM provider
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 7,
			"type": "direct",
			"explanation": "明確矛盾：活著 vs 死了"
		}`,
	}

	// Create analyzer
	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	// Create UpdateManager with analyzer
	config := DefaultUpdateManagerConfig()
	config.ContradictionAnalyzer = analyzer
	manager := NewUpdateManager(config)

	// Setup test data
	fact := &Fact{
		ID:      "fact_001",
		Content: "張醫生還活著",
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

	t.Run("detects contradiction using LLM analyzer", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "張醫生死了")

		require.NotNil(t, result)
		assert.Equal(t, "張醫生死了", result.NewInfo)

		// LLM severity (7) + confidence bonus (2) + witness bonus (3) = 12, capped at 10
		assert.Equal(t, 10, result.Severity)
		assert.Equal(t, ContradictionMajor, result.Type)
	})

	t.Run("LLM is called for analysis", func(t *testing.T) {
		// Clear cache and reset count
		analyzer.ClearCache()
		mockProvider.callCount = 0
		manager.CheckContradiction("npc_001", "張醫生死了")
		assert.Equal(t, 1, mockProvider.callCount)
	})
}

// TestUpdateManager_LLMFallback tests fallback to keyword matching when LLM fails
func TestUpdateManager_LLMFallback(t *testing.T) {
	// Create mock LLM provider that fails
	mockProvider := &MockLLMProvider{
		err: context.DeadlineExceeded,
	}

	// Create analyzer
	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	// Create UpdateManager with analyzer
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

	t.Run("falls back to keyword matching when LLM fails", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "張醫生死了")

		require.NotNil(t, result)
		// Should still detect contradiction using keyword matching
		assert.Equal(t, 10, result.Severity) // 5 + 2 + 3 = 10
		assert.Equal(t, ContradictionMajor, result.Type)
	})
}

// TestUpdateManager_WithoutLLMAnalyzer tests UpdateManager without LLM analyzer (legacy mode)
func TestUpdateManager_WithoutLLMAnalyzer(t *testing.T) {
	// Create UpdateManager without analyzer
	manager := NewUpdateManager(nil)

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

	t.Run("uses keyword matching when no analyzer", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "張醫生死了")

		require.NotNil(t, result)
		assert.Equal(t, 10, result.Severity)
		assert.Equal(t, ContradictionMajor, result.Type)
	})
}

// TestUpdateManager_IndirectContradiction tests indirect contradiction detection
func TestUpdateManager_IndirectContradiction(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 6,
			"type": "indirect",
			"explanation": "間接矛盾：兇手逃走了但仍在房間"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	config := DefaultUpdateManagerConfig()
	config.ContradictionAnalyzer = analyzer
	manager := NewUpdateManager(config)

	fact := &Fact{
		ID:      "fact_001",
		Content: "我看到兇手逃走了",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Told,
		Confidence:       0.7,
		PropagationDepth: 1,
	}
	manager.mu.Unlock()

	t.Run("detects indirect contradiction", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "兇手還在房間裡")

		require.NotNil(t, result)
		assert.True(t, result.Severity > 0)
		// Indirect contradiction with low confidence - should be moderate
		assert.Contains(t, []ContradictionType{ContradictionModerate, ContradictionMajor}, result.Type)
	})
}

// TestUpdateManager_TemporalContradiction tests temporal contradiction detection
func TestUpdateManager_TemporalContradiction(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 8,
			"type": "temporal",
			"explanation": "時序矛盾：從未見過 vs 昨天聊天"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	config := DefaultUpdateManagerConfig()
	config.ContradictionAnalyzer = analyzer
	manager := NewUpdateManager(config)

	fact := &Fact{
		ID:      "fact_001",
		Content: "我從未見過那個人",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Witness,
		Confidence:       0.9,
		PropagationDepth: 0,
	}
	manager.mu.Unlock()

	t.Run("detects temporal contradiction", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "我昨天和他聊天")

		require.NotNil(t, result)
		// High severity temporal contradiction
		assert.Equal(t, ContradictionMajor, result.Type)
		assert.GreaterOrEqual(t, result.Severity, 8)
	})
}

// TestUpdateManager_ConditionalContradiction tests conditional contradiction detection
func TestUpdateManager_ConditionalContradiction(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 5,
			"type": "conditional",
			"explanation": "條件矛盾：下雨留家 vs 下雨出門"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	config := DefaultUpdateManagerConfig()
	config.ContradictionAnalyzer = analyzer
	manager := NewUpdateManager(config)

	fact := &Fact{
		ID:      "fact_001",
		Content: "如果下雨我會留在家裡",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Told,
		Confidence:       0.8,
		PropagationDepth: 1,
	}
	manager.mu.Unlock()

	t.Run("detects conditional contradiction", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "昨天下雨但我出去了")

		require.NotNil(t, result)
		// Conditional contradiction - moderate severity
		assert.Contains(t, []ContradictionType{ContradictionModerate, ContradictionMajor}, result.Type)
	})
}

// TestUpdateManager_CacheEfficiency tests cache hit rate
func TestUpdateManager_CacheEfficiency(t *testing.T) {
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

	fact := &Fact{
		ID:      "fact_001",
		Content: "測試內容A",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Told,
		Confidence:       0.7,
		PropagationDepth: 1,
	}
	manager.mu.Unlock()

	t.Run("achieves target cache hit rate", func(t *testing.T) {
		mockProvider.callCount = 0

		// Make repeated checks with same contradiction
		for i := 0; i < 10; i++ {
			manager.CheckContradiction("npc_001", "測試內容B")
		}

		// Should only call LLM once due to caching
		assert.Equal(t, 1, mockProvider.callCount)

		// Check cache stats
		stats := manager.GetContradictionCacheStats()
		require.NotNil(t, stats)

		// Hit rate should be high (9 hits out of 10 requests)
		assert.GreaterOrEqual(t, stats.HitRate, 60.0) // AC4: ≥ 60% hit rate
		assert.Equal(t, 9, stats.CacheHits)
		assert.Equal(t, 1, stats.CacheMisses)
	})
}

// TestUpdateManager_NoContradiction tests non-contradictory statements
func TestUpdateManager_NoContradiction(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": false,
			"severity": 0,
			"type": "none",
			"explanation": "沒有矛盾"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	config := DefaultUpdateManagerConfig()
	config.ContradictionAnalyzer = analyzer
	manager := NewUpdateManager(config)

	fact := &Fact{
		ID:      "fact_001",
		Content: "張醫生很累",
		Type:    Event,
	}
	manager.RegisterFact(fact)

	manager.mu.Lock()
	manager.ensureKnowledgeBaseLocked("npc_001")
	manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Told,
		Confidence:       0.7,
		PropagationDepth: 1,
	}
	manager.mu.Unlock()

	t.Run("returns nil for non-contradictory statements", func(t *testing.T) {
		result := manager.CheckContradiction("npc_001", "張醫生在休息")
		assert.Nil(t, result)
	})
}

// TestUpdateManager_SeverityAdjustment tests severity adjustment based on fact context
func TestUpdateManager_SeverityAdjustment(t *testing.T) {
	t.Run("high confidence increases severity", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			response: `{"is_contradictory": true, "severity": 5, "type": "direct", "explanation": "Test"}`,
		}

		analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
			Provider: mockProvider,
			Timeout:  1 * time.Second,
		})

		config := DefaultUpdateManagerConfig()
		config.ContradictionAnalyzer = analyzer
		manager := NewUpdateManager(config)

		fact := &Fact{
			ID:      "fact_001",
			Content: "測試A",
			Type:    Event,
		}
		manager.RegisterFact(fact)

		manager.mu.Lock()
		manager.ensureKnowledgeBaseLocked("npc_001")
		manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
			FactID:           "fact_001",
			LearnedAt:        time.Now(),
			LearnMethod:      Told,
			Confidence:       0.9, // High confidence
			PropagationDepth: 1,
		}
		manager.mu.Unlock()

		result := manager.CheckContradiction("npc_001", "測試B")
		require.NotNil(t, result)

		// Base 5 + confidence bonus 2 = 7
		assert.Equal(t, 7, result.Severity)
	})

	t.Run("witnessed increases severity", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			response: `{"is_contradictory": true, "severity": 5, "type": "direct", "explanation": "Test"}`,
		}

		analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
			Provider: mockProvider,
			Timeout:  1 * time.Second,
		})

		config := DefaultUpdateManagerConfig()
		config.ContradictionAnalyzer = analyzer
		manager := NewUpdateManager(config)

		fact := &Fact{
			ID:      "fact_001",
			Content: "測試A",
			Type:    Event,
		}
		manager.RegisterFact(fact)

		manager.mu.Lock()
		manager.ensureKnowledgeBaseLocked("npc_001")
		manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
			FactID:           "fact_001",
			LearnedAt:        time.Now(),
			LearnMethod:      Witness, // Witnessed
			Confidence:       0.7,
			PropagationDepth: 0,
		}
		manager.mu.Unlock()

		result := manager.CheckContradiction("npc_001", "測試B")
		require.NotNil(t, result)

		// Base 5 + witness bonus 3 = 8
		assert.Equal(t, 8, result.Severity)
	})

	t.Run("both high confidence and witnessed gives maximum", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			response: `{"is_contradictory": true, "severity": 6, "type": "direct", "explanation": "Test"}`,
		}

		analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
			Provider: mockProvider,
			Timeout:  1 * time.Second,
		})

		config := DefaultUpdateManagerConfig()
		config.ContradictionAnalyzer = analyzer
		manager := NewUpdateManager(config)

		fact := &Fact{
			ID:      "fact_001",
			Content: "測試A",
			Type:    Event,
		}
		manager.RegisterFact(fact)

		manager.mu.Lock()
		manager.ensureKnowledgeBaseLocked("npc_001")
		manager.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
			FactID:           "fact_001",
			LearnedAt:        time.Now(),
			LearnMethod:      Witness,
			Confidence:       0.95,
			PropagationDepth: 0,
		}
		manager.mu.Unlock()

		result := manager.CheckContradiction("npc_001", "測試B")
		require.NotNil(t, result)

		// Base 6 + confidence 2 + witness 3 = 11, capped at 10
		assert.Equal(t, 10, result.Severity)
	})
}
