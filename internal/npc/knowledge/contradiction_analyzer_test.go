package knowledge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProvider implements a mock LLM provider for testing
type MockLLMProvider struct {
	response      string
	err           error
	callCount     int
	lastMessages  []client.Message
	shouldTimeout bool
}

func (m *MockLLMProvider) Name() string { return "mock" }
func (m *MockLLMProvider) ModelInfo() client.ModelInfo {
	return client.ModelInfo{Provider: "mock", Model: "mock-model", MaxTokens: 4096}
}
func (m *MockLLMProvider) TestConnection(ctx context.Context) error { return nil }
func (m *MockLLMProvider) SendMessage(ctx context.Context, messages []client.Message) (*client.Response, error) {
	m.callCount++
	m.lastMessages = messages

	if m.shouldTimeout {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	if m.err != nil {
		return nil, m.err
	}

	return &client.Response{
		Content: m.response,
		Metadata: map[string]interface{}{
			"model": "mock-model",
		},
	}, nil
}
func (m *MockLLMProvider) Stream(ctx context.Context, messages []client.Message, callback func(chunk string)) error {
	return nil
}

// TestNewContradictionAnalyzer tests analyzer creation
func TestNewContradictionAnalyzer(t *testing.T) {
	t.Run("creates analyzer with default config", func(t *testing.T) {
		analyzer := NewContradictionAnalyzer(nil)
		require.NotNil(t, analyzer)
		assert.NotNil(t, analyzer.cache)
		assert.Equal(t, 5*time.Second, analyzer.timeout)
	})

	t.Run("creates analyzer with custom config", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		customCache := NewContradictionCache(&CacheConfig{MaxSize: 50, TTL: 10 * time.Minute})

		config := &ContradictionAnalyzerConfig{
			Provider: mockProvider,
			Cache:    customCache,
			Timeout:  10 * time.Second,
		}

		analyzer := NewContradictionAnalyzer(config)
		require.NotNil(t, analyzer)
		assert.Equal(t, mockProvider, analyzer.provider)
		assert.Equal(t, customCache, analyzer.cache)
		assert.Equal(t, 10*time.Second, analyzer.timeout)
	})
}

// TestAnalyzeContradiction_DirectContradictions tests direct contradiction detection
func TestAnalyzeContradiction_DirectContradictions(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 10,
			"type": "direct",
			"explanation": "明確相反：活著 vs 死了"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Witness,
		Confidence:  1.0,
	}

	t.Run("detects direct contradiction: alive vs dead", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"張醫生還活著",
			"張醫生死了",
			knownFact,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsContradictory)
		assert.Equal(t, 10, result.Severity)
		assert.Equal(t, "direct", result.Type)
		assert.Contains(t, result.Explanation, "活著")
	})
}

// TestAnalyzeContradiction_IndirectContradictions tests indirect contradiction detection
func TestAnalyzeContradiction_IndirectContradictions(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 7,
			"type": "indirect",
			"explanation": "間接矛盾：兇手逃走了 vs 兇手在房間裡"
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

	t.Run("detects indirect contradiction", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"我看到兇手逃走了",
			"兇手還在房間裡",
			knownFact,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsContradictory)
		assert.Equal(t, 7, result.Severity)
		assert.Equal(t, "indirect", result.Type)
	})
}

// TestAnalyzeContradiction_TemporalContradictions tests temporal contradiction detection
func TestAnalyzeContradiction_TemporalContradictions(t *testing.T) {
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

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Witness,
		Confidence:  0.9,
	}

	t.Run("detects temporal contradiction", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"我從未見過那個人",
			"我昨天和他聊天",
			knownFact,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsContradictory)
		assert.Equal(t, 8, result.Severity)
		assert.Equal(t, "temporal", result.Type)
	})
}

// TestAnalyzeContradiction_ConditionalContradictions tests conditional contradiction detection
func TestAnalyzeContradiction_ConditionalContradictions(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 6,
			"type": "conditional",
			"explanation": "條件矛盾：下雨留家 vs 昨天下雨出去了"
		}`,
	}

	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: mockProvider,
		Timeout:  1 * time.Second,
	})

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Told,
		Confidence:  0.8,
	}

	t.Run("detects conditional contradiction", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"如果下雨我會留在家裡",
			"昨天下雨但我出去了",
			knownFact,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsContradictory)
		assert.Equal(t, 6, result.Severity)
		assert.Equal(t, "conditional", result.Type)
	})
}

// TestAnalyzeContradiction_NoContradiction tests non-contradictory cases
func TestAnalyzeContradiction_NoContradiction(t *testing.T) {
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

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Told,
		Confidence:  0.7,
	}

	t.Run("returns false for non-contradictory statements", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"張醫生很累",
			"張醫生在休息",
			knownFact,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsContradictory)
		assert.Equal(t, 0, result.Severity)
		assert.Equal(t, "none", result.Type)
	})
}

// TestAnalyzeContradiction_CacheHit tests cache functionality
func TestAnalyzeContradiction_CacheHit(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{
			"is_contradictory": true,
			"severity": 8,
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

	t.Run("uses cache on second call", func(t *testing.T) {
		// First call - should call LLM
		result1, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"A",
			"B",
			knownFact,
		)
		require.NoError(t, err)
		require.NotNil(t, result1)
		assert.Equal(t, 1, mockProvider.callCount)

		// Second call - should use cache
		result2, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"A",
			"B",
			knownFact,
		)
		require.NoError(t, err)
		require.NotNil(t, result2)
		assert.Equal(t, 1, mockProvider.callCount) // Should not increase

		// Results should be identical
		assert.Equal(t, result1.IsContradictory, result2.IsContradictory)
		assert.Equal(t, result1.Severity, result2.Severity)
	})

	t.Run("cache works with reversed order", func(t *testing.T) {
		mockProvider.callCount = 0
		analyzer.ClearCache()

		// First call
		analyzer.AnalyzeContradiction(context.Background(), "X", "Y", knownFact)
		assert.Equal(t, 1, mockProvider.callCount)

		// Reversed order should hit cache
		analyzer.AnalyzeContradiction(context.Background(), "Y", "X", knownFact)
		assert.Equal(t, 1, mockProvider.callCount)
	})
}

// TestAnalyzeContradiction_LLMFailure tests LLM failure handling
func TestAnalyzeContradiction_LLMFailure(t *testing.T) {
	mockProvider := &MockLLMProvider{
		err: errors.New("LLM API error"),
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

	t.Run("returns error on LLM failure", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"A",
			"B",
			knownFact,
		)

		assert.Error(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsContradictory)
		assert.Equal(t, "error", result.Type)
	})
}

// TestAnalyzeContradiction_NoProvider tests behavior without LLM provider
func TestAnalyzeContradiction_NoProvider(t *testing.T) {
	analyzer := NewContradictionAnalyzer(&ContradictionAnalyzerConfig{
		Provider: nil,
		Timeout:  1 * time.Second,
	})

	knownFact := &KnownFact{
		FactID:      "test",
		LearnMethod: Told,
		Confidence:  0.7,
	}

	t.Run("returns non-contradictory when no provider", func(t *testing.T) {
		result, err := analyzer.AnalyzeContradiction(
			context.Background(),
			"A",
			"B",
			knownFact,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsContradictory)
		assert.Equal(t, "unknown", result.Type)
	})
}

// TestParseAnalysisResponse tests JSON parsing
func TestParseAnalysisResponse(t *testing.T) {
	analyzer := NewContradictionAnalyzer(nil)

	t.Run("parses valid JSON", func(t *testing.T) {
		content := `{
			"is_contradictory": true,
			"severity": 7,
			"type": "indirect",
			"explanation": "Test explanation"
		}`

		result, err := analyzer.parseAnalysisResponse(content)
		require.NoError(t, err)
		assert.True(t, result.IsContradictory)
		assert.Equal(t, 7, result.Severity)
		assert.Equal(t, "indirect", result.Type)
	})

	t.Run("handles JSON with markdown code blocks", func(t *testing.T) {
		content := "```json\n{\"is_contradictory\": true, \"severity\": 5, \"type\": \"direct\", \"explanation\": \"Test\"}\n```"

		result, err := analyzer.parseAnalysisResponse(content)
		require.NoError(t, err)
		assert.True(t, result.IsContradictory)
		assert.Equal(t, 5, result.Severity)
	})

	t.Run("handles JSON surrounded by text", func(t *testing.T) {
		content := "Here is the analysis:\n{\"is_contradictory\": false, \"severity\": 0, \"type\": \"none\", \"explanation\": \"No contradiction\"}\nEnd of analysis"

		result, err := analyzer.parseAnalysisResponse(content)
		require.NoError(t, err)
		assert.False(t, result.IsContradictory)
	})

	t.Run("clamps severity to valid range", func(t *testing.T) {
		content := `{"is_contradictory": true, "severity": 15, "type": "direct", "explanation": "Test"}`

		result, err := analyzer.parseAnalysisResponse(content)
		require.NoError(t, err)
		assert.Equal(t, 0, result.Severity) // Clamped to 0 (invalid value)
	})

	t.Run("validates type field", func(t *testing.T) {
		content := `{"is_contradictory": true, "severity": 5, "type": "invalid_type", "explanation": "Test"}`

		result, err := analyzer.parseAnalysisResponse(content)
		require.NoError(t, err)
		assert.Equal(t, "unknown", result.Type)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		content := "Not valid JSON at all"

		_, err := analyzer.parseAnalysisResponse(content)
		assert.Error(t, err)
	})

	t.Run("returns error for empty content", func(t *testing.T) {
		content := ""

		_, err := analyzer.parseAnalysisResponse(content)
		assert.Error(t, err)
	})
}

// TestBuildAnalysisPrompt tests prompt construction
func TestBuildAnalysisPrompt(t *testing.T) {
	analyzer := NewContradictionAnalyzer(nil)

	t.Run("builds prompt without known fact", func(t *testing.T) {
		prompt := analyzer.buildAnalysisPrompt("Statement A", "Statement B", nil)

		assert.Contains(t, prompt, "Statement A")
		assert.Contains(t, prompt, "Statement B")
		assert.Contains(t, prompt, "語義矛盾")
		assert.Contains(t, prompt, "direct")
		assert.Contains(t, prompt, "indirect")
		assert.Contains(t, prompt, "temporal")
		assert.Contains(t, prompt, "conditional")
	})

	t.Run("builds prompt with known fact context", func(t *testing.T) {
		knownFact := &KnownFact{
			FactID:      "test",
			LearnMethod: Witness,
			Confidence:  0.9,
			PropagationDepth: 1,
		}

		prompt := analyzer.buildAnalysisPrompt("Statement A", "Statement B", knownFact)

		assert.Contains(t, prompt, "0.90") // Confidence
		assert.Contains(t, prompt, "witness")
		assert.Contains(t, prompt, "1") // Propagation depth
	})

	t.Run("prompt includes all contradiction types", func(t *testing.T) {
		prompt := analyzer.buildAnalysisPrompt("A", "B", nil)

		assert.Contains(t, prompt, "直接矛盾")
		assert.Contains(t, prompt, "間接矛盾")
		assert.Contains(t, prompt, "時序矛盾")
		assert.Contains(t, prompt, "條件矛盾")
	})
}

// TestGetCacheStats tests cache statistics retrieval
func TestGetCacheStats(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{"is_contradictory": true, "severity": 5, "type": "direct", "explanation": "Test"}`,
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

	t.Run("returns cache stats", func(t *testing.T) {
		analyzer.ClearCache()
		mockProvider.callCount = 0

		// Make some calls
		analyzer.AnalyzeContradiction(context.Background(), "A", "B", knownFact)
		analyzer.AnalyzeContradiction(context.Background(), "A", "B", knownFact) // Cache hit
		analyzer.AnalyzeContradiction(context.Background(), "C", "D", knownFact)

		stats := analyzer.GetCacheStats()
		// Total requests = 3 (all Get calls to cache)
		// Cache hits = 1 (second "A", "B" call)
		// Cache misses = 2 (first "A", "B" and "C", "D")
		assert.Equal(t, 3, stats.TotalRequests)
		assert.Equal(t, 1, stats.CacheHits)
		assert.Equal(t, 2, stats.CacheMisses)
		assert.InDelta(t, 33.33, stats.HitRate, 0.1) // 1/3 = 33.33%
	})
}

// TestClearCache tests cache clearing
func TestClearCache(t *testing.T) {
	mockProvider := &MockLLMProvider{
		response: `{"is_contradictory": true, "severity": 5, "type": "direct", "explanation": "Test"}`,
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

	t.Run("clears cache", func(t *testing.T) {
		// Make a call
		analyzer.AnalyzeContradiction(context.Background(), "A", "B", knownFact)
		assert.Equal(t, 1, mockProvider.callCount)

		// Clear cache
		analyzer.ClearCache()

		// Make same call again - should call LLM again
		mockProvider.callCount = 0
		analyzer.AnalyzeContradiction(context.Background(), "A", "B", knownFact)
		assert.Equal(t, 1, mockProvider.callCount)
	})
}
