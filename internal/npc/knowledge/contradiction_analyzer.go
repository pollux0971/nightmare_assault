package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ContradictionAnalysisResult represents the result of LLM-based contradiction analysis.
type ContradictionAnalysisResult struct {
	IsContradictory bool   `json:"is_contradictory"` // Whether the statements contradict
	Severity        int    `json:"severity"`          // Severity score 1-10
	Type            string `json:"type"`              // Type: direct|indirect|temporal|conditional
	Explanation     string `json:"explanation"`       // Brief explanation of the contradiction
}

// ContradictionAnalyzer uses LLM to analyze semantic contradictions between statements.
// It understands complex scenarios including indirect, temporal, and conditional contradictions.
//
// Story 8.2 AC1: 語義矛盾判斷
type ContradictionAnalyzer struct {
	provider client.Provider
	cache    *ContradictionCache
	timeout  time.Duration
}

// ContradictionAnalyzerConfig configures the contradiction analyzer.
type ContradictionAnalyzerConfig struct {
	Provider client.Provider
	Cache    *ContradictionCache
	Timeout  time.Duration
}

// NewContradictionAnalyzer creates a new ContradictionAnalyzer.
// If config is nil or cache is not provided, a default cache is created.
// If timeout is 0, defaults to 5 seconds.
func NewContradictionAnalyzer(config *ContradictionAnalyzerConfig) *ContradictionAnalyzer {
	if config == nil {
		config = &ContradictionAnalyzerConfig{}
	}

	cache := config.Cache
	if cache == nil {
		cache = NewContradictionCache(DefaultCacheConfig())
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &ContradictionAnalyzer{
		provider: config.Provider,
		cache:    cache,
		timeout:  timeout,
	}
}

// AnalyzeContradiction analyzes whether two statements contradict each other using LLM.
// It supports complex scenarios:
//   - Direct contradictions: "alive" vs "dead"
//   - Indirect contradictions: "I saw the killer escape" vs "the killer is in the room"
//   - Temporal contradictions: "I never met him" vs "I talked to him yesterday"
//   - Conditional contradictions: "If it rains I stay home" vs "I went out yesterday" + "it rained yesterday"
//
// Story 8.2 AC1: LLM 語義矛盾判斷
// Story 8.2 AC2: 矛盾嚴重程度評估 (1-10)
// Story 8.2 AC3: 複雜情境支援
func (a *ContradictionAnalyzer) AnalyzeContradiction(ctx context.Context, statementA, statementB string, knownFact *KnownFact) (*ContradictionAnalysisResult, error) {
	// Check cache first (AC4: 效能優化)
	if cached := a.cache.Get(statementA, statementB); cached != nil {
		logger.Debug("Contradiction cache hit", map[string]interface{}{
			"statement_a": statementA,
			"statement_b": statementB,
		})
		return cached, nil
	}

	// No LLM provider means we can't do semantic analysis
	if a.provider == nil {
		logger.Debug("No LLM provider for contradiction analysis, returning no contradiction", nil)
		return &ContradictionAnalysisResult{
			IsContradictory: false,
			Severity:        0,
			Type:            "unknown",
			Explanation:     "No LLM provider available for semantic analysis",
		}, nil
	}

	// Prepare context with timeout
	analyzeCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Build the prompt for LLM analysis
	prompt := a.buildAnalysisPrompt(statementA, statementB, knownFact)

	// Call LLM
	startTime := time.Now()
	messages := []client.Message{
		{Role: "user", Content: prompt},
	}

	response, err := a.provider.SendMessage(analyzeCtx, messages)
	if err != nil {
		logger.Error("LLM contradiction analysis failed", map[string]interface{}{
			"error":       err.Error(),
			"statement_a": statementA,
			"statement_b": statementB,
		})
		// Return a non-contradictory result on error to avoid blocking game flow
		return &ContradictionAnalysisResult{
			IsContradictory: false,
			Severity:        0,
			Type:            "error",
			Explanation:     fmt.Sprintf("LLM analysis failed: %v", err),
		}, err
	}

	elapsed := time.Since(startTime)
	logger.Debug("LLM contradiction analysis completed", map[string]interface{}{
		"elapsed_ms":  elapsed.Milliseconds(),
		"statement_a": statementA,
		"statement_b": statementB,
	})

	// Parse the response
	result, err := a.parseAnalysisResponse(response.Content)
	if err != nil {
		logger.Error("Failed to parse LLM contradiction response", map[string]interface{}{
			"error":    err.Error(),
			"content":  response.Content,
		})
		return &ContradictionAnalysisResult{
			IsContradictory: false,
			Severity:        0,
			Type:            "parse_error",
			Explanation:     fmt.Sprintf("Failed to parse response: %v", err),
		}, err
	}

	// Cache the result (AC4: 快取機制)
	a.cache.Put(statementA, statementB, result)

	return result, nil
}

// buildAnalysisPrompt constructs the LLM prompt for contradiction analysis.
// The prompt is designed to handle complex scenarios and return structured JSON.
//
// Story 8.2 AC1: 包含上下文感知的 Prompt
// Story 8.2 AC3: 複雜情境支援
func (a *ContradictionAnalyzer) buildAnalysisPrompt(statementA, statementB string, knownFact *KnownFact) string {
	// Build context information about the known fact
	contextInfo := ""
	if knownFact != nil {
		contextInfo = fmt.Sprintf(`
語義分析上下文：
- 陳述 A 的可信度：%.2f (0.0-1.0)
- 陳述 A 的學習方式：%s
- 陳述 A 的傳播深度：%d

`, knownFact.Confidence, knownFact.LearnMethod.String(), knownFact.PropagationDepth)
	}

	prompt := fmt.Sprintf(`你是一個專門分析語義矛盾的 AI 助手。請分析以下兩個陳述是否在語義上矛盾。

%s陳述 A：「%s」
陳述 B：「%s」

請仔細考慮以下四種矛盾類型：

1. **直接矛盾 (direct)**：字面意思明確相反
   例如：「他活著」vs「他死了」

2. **間接矛盾 (indirect)**：通過推論得出的矛盾
   例如：「我看到兇手逃走了」vs「兇手在房間裡」

3. **時序矛盾 (temporal)**：描述同一時段不同狀態，或時間順序上的矛盾
   例如：「我從未見過那個人」vs「我昨天和他聊天」

4. **條件矛盾 (conditional)**：在特定條件下的矛盾
   例如：「如果下雨我會留在家裡」vs「昨天下雨但我出去了」

請以 JSON 格式回應，格式如下：
{
  "is_contradictory": true 或 false,
  "severity": 1-10 的整數（1=細微差異，10=嚴重邏輯矛盾）,
  "type": "direct" 或 "indirect" 或 "temporal" 或 "conditional" 或 "none",
  "explanation": "簡要說明（50字以內）"
}

**評分標準：**
- 嚴重程度 1-3：細微差異或不確定的矛盾
- 嚴重程度 4-6：明顯的矛盾但可能有合理解釋
- 嚴重程度 7-9：清晰的矛盾，難以合理解釋
- 嚴重程度 10：絕對的邏輯矛盾，不可能同時為真

**重要提醒：**
- 只回傳 JSON 格式，不要包含任何其他文字
- 如果兩個陳述不矛盾，請設定 is_contradictory 為 false，type 為 "none"，severity 為 0
- 請考慮語義上的矛盾，而非僅僅是字面上的差異
`, contextInfo, statementA, statementB)

	return prompt
}

// parseAnalysisResponse parses the LLM response into a ContradictionAnalysisResult.
// It handles JSON parsing and validates the structure.
//
// Story 8.2 AC1: 解析 LLM 回應提取判斷結果
func (a *ContradictionAnalyzer) parseAnalysisResponse(content string) (*ContradictionAnalysisResult, error) {
	// Clean up the content - remove markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Try to find JSON object in the content
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("no valid JSON object found in response")
	}
	content = content[startIdx : endIdx+1]

	var result ContradictionAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate the result
	if result.Severity < 0 || result.Severity > 10 {
		result.Severity = 0 // Clamp to valid range
	}

	// Validate type
	validTypes := map[string]bool{
		"direct": true, "indirect": true, "temporal": true, "conditional": true, "none": true,
	}
	if !validTypes[result.Type] {
		result.Type = "unknown"
	}

	return &result, nil
}

// GetCacheStats returns cache statistics for monitoring.
// Story 8.2 AC4: 快取命中率監控
func (a *ContradictionAnalyzer) GetCacheStats() CacheStats {
	return a.cache.GetStats()
}

// ClearCache clears the contradiction cache.
// Useful for testing or when memory needs to be reclaimed.
func (a *ContradictionAnalyzer) ClearCache() {
	a.cache.Clear()
}
