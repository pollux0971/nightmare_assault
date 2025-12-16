package context

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ContextManager manages game context with summary, window, and token optimization
type ContextManager struct {
	window           *ContextWindow
	summaryGenerator SummaryGenerator
	tokenCounter     TokenCounter
	config           ContextConfig
	createdAt        time.Time
	lastUpdatedAt    time.Time
	summaryCount     int          // Number of summaries generated
	shrinkCount      int          // Number of window shrinks
	mu               sync.RWMutex
}

// ContextConfig configuration for ContextManager
type ContextConfig struct {
	WindowSize         int    `json:"window_size"`
	SummaryTrigger     int    `json:"summary_trigger"`
	TokenLimit         int    `json:"token_limit"`
	ModelName          string `json:"model_name"`
	EnableAutoOptimize bool   `json:"enable_auto_optimize"`
}

// DefaultContextConfig returns default configuration
func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		WindowSize:         DEFAULT_WINDOW_SIZE, // 5
		SummaryTrigger:     SUMMARY_TRIGGER,     // 15
		TokenLimit:         MAX_CONTEXT_TOKENS,  // 8000
		ModelName:          "gpt-4",
		EnableAutoOptimize: true,
	}
}

// ContextMetadata provides context usage statistics
type ContextMetadata struct {
	TotalEntries     int     `json:"total_entries"`
	WindowEntries    int     `json:"window_entries"`
	SummaryGenerated bool    `json:"summary_generated"`
	TotalTokens      int     `json:"total_tokens"`
	TokenUsageRatio  float64 `json:"token_usage_ratio"`
}

// ContextHealthStatus provides health check information
type ContextHealthStatus struct {
	IsHealthy     bool     `json:"is_healthy"`
	TokenUsageOK  bool     `json:"token_usage_ok"`
	WindowSizeOK  bool     `json:"window_size_ok"`
	SummaryStatus string   `json:"summary_status"`
	Issues        []string `json:"issues"`
}

// ContextStatistics provides detailed statistics
type ContextStatistics struct {
	TotalEntriesAdded     int       `json:"total_entries_added"`
	SummariesGenerated    int       `json:"summaries_generated"`
	WindowShrinks         int       `json:"window_shrinks"`
	AverageTokensPerEntry int       `json:"average_tokens_per_entry"`
	CreatedAt             time.Time `json:"created_at"`
	LastUpdatedAt         time.Time `json:"last_updated_at"`
}

// NewContextManager creates a new ContextManager with the given configuration
func NewContextManager(config ContextConfig) (*ContextManager, error) {
	// 1. Initialize TokenCounter
	var tokenCounter TokenCounter
	tc, err := NewTiktokenCounter(config.ModelName)
	if err != nil {
		// Fallback to estimate counter
		tokenCounter = NewEstimateTokenCounter()
	} else {
		tokenCounter = tc
	}

	// 2. Initialize ContextWindow
	window := NewContextWindow()
	window.WindowSize = config.WindowSize
	window.SetTokenCounter(tokenCounter)

	// 3. SummaryGenerator will be set by caller if needed
	// (requires LLM client which is external dependency)

	cm := &ContextManager{
		window:        window,
		tokenCounter:  tokenCounter,
		config:        config,
		createdAt:     time.Now(),
		lastUpdatedAt: time.Now(),
	}

	return cm, nil
}

// SetSummaryGenerator sets the summary generator (dependency injection)
func (cm *ContextManager) SetSummaryGenerator(gen SummaryGenerator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.summaryGenerator = gen
}

// GetOptimizedContext returns the current optimized context window
func (cm *ContextManager) GetOptimizedContext() *ContextWindow {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.window
}

// AddHistoryEntry adds a new history entry to the context
func (cm *ContextManager) AddHistoryEntry(entry HistoryEntry) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	err := cm.window.AddEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to add entry: %w", err)
	}

	cm.lastUpdatedAt = time.Now()

	// If auto-optimization is enabled, check and optimize
	if cm.config.EnableAutoOptimize {
		cm.checkAndOptimize()
	}

	return nil
}

// checkAndOptimize checks token usage and applies optimization if needed
// Must be called with write lock held
func (cm *ContextManager) checkAndOptimize() {
	// Check token usage
	tokens, limit, warningLevel := cm.window.CheckTokenUsage()

	switch warningLevel {
	case 2: // Emergency (>= 90%)
		// Shrink window
		if cm.window.ShrinkWindow() {
			cm.shrinkCount++
		}

	case 1: // Warning (>= 70%)
		// Trigger summary generation if needed
		if cm.window.ShouldGenerateSummary() && cm.summaryGenerator != nil {
			// Mark summary as needed
			// Actual generation should be done by Orchestrator (Epic 6)
			// For now, we just increment the count when we would generate
			cm.summaryCount++
		}

	case 0: // Normal
		// No optimization needed
	}

	_ = tokens
	_ = limit
}

// UpdateTension updates the current tension value (for future extensions)
func (cm *ContextManager) UpdateTension(tension float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	// Reserved for future tension integration
	cm.lastUpdatedAt = time.Now()
}

// Clear clears all context data (used for new game)
func (cm *ContextManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.window.Clear()
	cm.summaryCount = 0
	cm.shrinkCount = 0
	cm.lastUpdatedAt = time.Now()
}

// FormatSummarySection formats the summary as a prompt section
func (cm *ContextManager) FormatSummarySection() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.window.Summary == "" {
		return ""
	}

	return fmt.Sprintf(`## 前情提要
%s

`, cm.window.Summary)
}

// FormatRecentHistory formats recent history as dialogue format
func (cm *ContextManager) FormatRecentHistory() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	recentEntries := cm.window.GetWindow()

	if len(recentEntries) == 0 {
		return "## 最近歷史\n（無歷史）\n"
	}

	var sb strings.Builder
	sb.WriteString("## 最近歷史\n\n")

	for _, entry := range recentEntries {
		sb.WriteString(fmt.Sprintf("**回合 %d**\n", entry.Beat))
		sb.WriteString(fmt.Sprintf("- 玩家選擇：%s\n", entry.PlayerChoice))
		sb.WriteString(fmt.Sprintf("- 故事發展：%s\n", entry.StoryContent))

		// HP and SAN changes
		if entry.HPChange != 0 {
			sb.WriteString(fmt.Sprintf("- HP 變化：%+d\n", entry.HPChange))
		}
		if entry.SANChange != 0 {
			sb.WriteString(fmt.Sprintf("- SAN 變化：%+d\n", entry.SANChange))
		}

		if len(entry.CluesFound) > 0 {
			sb.WriteString(fmt.Sprintf("- 發現線索：%v\n", entry.CluesFound))
		}

		if len(entry.RulesTriggered) > 0 {
			sb.WriteString(fmt.Sprintf("- 觸發規則：%v\n", entry.RulesTriggered))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatCompleteContext formats complete context for LLM prompt
func (cm *ContextManager) FormatCompleteContext() string {
	var sb strings.Builder

	// 1. Summary section (if exists)
	summary := cm.FormatSummarySection()
	if summary != "" {
		sb.WriteString(summary)
	}

	// 2. Recent history
	sb.WriteString(cm.FormatRecentHistory())

	// 3. Token usage comment (for debugging)
	metadata := cm.GetContextMetadata()
	sb.WriteString(fmt.Sprintf("<!-- Token 使用：%d / %d (%.1f%%) -->\n",
		metadata.TotalTokens,
		cm.config.TokenLimit,
		metadata.TokenUsageRatio*100,
	))

	return sb.String()
}

// GetContextMetadata returns context metadata
func (cm *ContextManager) GetContextMetadata() ContextMetadata {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalEntries := cm.window.GetEntryCount()
	windowEntries := len(cm.window.GetWindow())
	summaryGenerated := cm.window.Summary != ""

	totalTokens := 0
	if cm.tokenCounter != nil {
		totalTokens = cm.window.CalculateTotalTokens()
	}

	tokenUsageRatio := 0.0
	if cm.config.TokenLimit > 0 {
		tokenUsageRatio = float64(totalTokens) / float64(cm.config.TokenLimit)
	}

	return ContextMetadata{
		TotalEntries:     totalEntries,
		WindowEntries:    windowEntries,
		SummaryGenerated: summaryGenerated,
		TotalTokens:      totalTokens,
		TokenUsageRatio:  tokenUsageRatio,
	}
}

// GetHealthStatus performs health check on context
func (cm *ContextManager) GetHealthStatus() ContextHealthStatus {
	metadata := cm.GetContextMetadata()

	status := ContextHealthStatus{
		IsHealthy:    true,
		TokenUsageOK: metadata.TokenUsageRatio < 0.9,
		WindowSizeOK: cm.window.WindowSize >= 3,
		Issues:       make([]string, 0),
	}

	if metadata.SummaryGenerated {
		status.SummaryStatus = "Generated"
	} else {
		status.SummaryStatus = "Not yet generated"
	}

	// Check for issues
	if metadata.TokenUsageRatio > 0.9 {
		status.IsHealthy = false
		status.Issues = append(status.Issues, "Token usage critical (>90%)")
	}

	if cm.window.WindowSize < 3 {
		status.IsHealthy = false
		status.Issues = append(status.Issues, "Window size too small (<3)")
	}

	return status
}

// GetStatistics returns detailed statistics
func (cm *ContextManager) GetStatistics() ContextStatistics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	metadata := cm.GetContextMetadata()

	avgTokens := 0
	if metadata.WindowEntries > 0 && metadata.TotalTokens > 0 {
		avgTokens = metadata.TotalTokens / metadata.WindowEntries
	}

	return ContextStatistics{
		TotalEntriesAdded:     metadata.TotalEntries,
		SummariesGenerated:    cm.summaryCount,
		WindowShrinks:         cm.shrinkCount,
		AverageTokensPerEntry: avgTokens,
		CreatedAt:             cm.createdAt,
		LastUpdatedAt:         cm.lastUpdatedAt,
	}
}

// SaveContext saves context to file
func (cm *ContextManager) SaveContext(path string) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.MarshalIndent(cm.window, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// LoadContext loads context from file
func (cm *ContextManager) LoadContext(path string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read context file: %w", err)
	}

	var window ContextWindow
	err = json.Unmarshal(data, &window)
	if err != nil {
		return fmt.Errorf("failed to unmarshal context: %w", err)
	}

	// Restore non-serialized fields
	window.SetTokenCounter(cm.tokenCounter)

	cm.window = &window
	cm.lastUpdatedAt = time.Now()

	return nil
}
