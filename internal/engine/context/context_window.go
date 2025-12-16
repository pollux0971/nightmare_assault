// Package context provides context window management for game history.
// It implements a sliding window mechanism with summary support for token optimization.
package context

import (
	"sync"
)

const (
	// MAX_CONTEXT_TOKENS is the maximum number of tokens allowed in context
	MAX_CONTEXT_TOKENS = 8000

	// MAX_HISTORY_BEFORE_SUMMARY is the maximum number of history entries before triggering summary
	// (corresponds to v2.0 MAX_RECENT_HISTORY)
	MAX_HISTORY_BEFORE_SUMMARY = 10

	// COMPRESSION_THRESHOLD is the threshold for triggering compression (v2.0 design value)
	COMPRESSION_THRESHOLD = 15

	// SUMMARY_TRIGGER is the number of new entries before generating a summary (Story 5.3)
	// Aligned with COMPRESSION_THRESHOLD from v2.0 design
	SUMMARY_TRIGGER = 15

	// DEFAULT_WINDOW_SIZE is the default sliding window size (number of recent entries passed to LLM)
	DEFAULT_WINDOW_SIZE = 5

	// TOTAL_SUMMARY_MAX_TOKENS is the maximum total token budget for accumulated summaries
	// When exceeded, old summaries should be compressed
	TOTAL_SUMMARY_MAX_TOKENS = 600

	// CHARS_PER_TOKEN is the estimated average characters per token (for Chinese text)
	// Used for simple token estimation: ~4 chars = 1 token
	CHARS_PER_TOKEN = 4

	// WARNING_THRESHOLD is the token usage percentage that triggers warnings (70%)
	WARNING_THRESHOLD = 0.7

	// EMERGENCY_THRESHOLD is the token usage percentage that triggers emergency optimization (90%)
	EMERGENCY_THRESHOLD = 0.9

	// MIN_WINDOW_SIZE is the minimum window size allowed during auto-optimization
	MIN_WINDOW_SIZE = 3
)

// HistoryEntry represents a single turn in the game history.
// It contains all information about player choice, story progression, and game state changes.
type HistoryEntry struct {
	Beat           int      `json:"beat"`            // Turn number
	PlayerChoice   string   `json:"player_choice"`   // Player's choice
	StoryContent   string   `json:"story_content"`   // Story text for this turn
	HPChange       int      `json:"hp_change"`       // HP change amount
	SANChange      int      `json:"san_change"`      // SAN change amount
	RulesTriggered []string `json:"rules_triggered"` // Triggered rule IDs
	CluesFound     []string `json:"clues_found"`     // Found clue IDs
}

// ContextWindow manages the game history context window and summary.
// It provides thread-safe access to recent entries and compressed summary text.
type ContextWindow struct {
	Summary             string         `json:"summary"`              // Summary text of game history
	AllEntries          []HistoryEntry `json:"all_entries"`          // Complete history (for summary generation)
	RecentEntries       []HistoryEntry `json:"recent_entries"`       // Recent N history entries (sliding window)
	WindowSize          int            `json:"window_size"`          // Window size (default 5)
	LastSummaryBeat     int            `json:"last_summary_beat"`    // Beat number of last summary generation
	IsSummaryInProgress bool           `json:"-"`                    // Summary generation in progress (not serialized)
	tokenCounter        TokenCounter   `json:"-"`                    // Token counter for optimization (not serialized)
	mu                  sync.RWMutex   // Mutex for concurrent access (not serialized)
}

// NewContextWindow creates a new ContextWindow with default settings.
// The window size is initialized to DEFAULT_WINDOW_SIZE (5).
func NewContextWindow() *ContextWindow {
	return &ContextWindow{
		Summary:       "",
		AllEntries:    []HistoryEntry{},
		RecentEntries: []HistoryEntry{},
		WindowSize:    DEFAULT_WINDOW_SIZE,
	}
}

// GetWindow returns a copy of the recent entries window.
// This method is thread-safe and nil-safe.
func (cw *ContextWindow) GetWindow() []HistoryEntry {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	// Nil safety: return empty slice instead of nil
	if cw.RecentEntries == nil {
		return []HistoryEntry{}
	}

	// Return a copy to prevent external modification
	window := make([]HistoryEntry, len(cw.RecentEntries))
	copy(window, cw.RecentEntries)
	return window
}

// GetSummary returns the current summary text.
// This method is thread-safe.
func (cw *ContextWindow) GetSummary() string {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.Summary
}

// AddEntry adds a new history entry and automatically updates the sliding window.
// This method is thread-safe.
func (cw *ContextWindow) AddEntry(entry HistoryEntry) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// 1. Add to complete history
	cw.AllEntries = append(cw.AllEntries, entry)

	// 2. Update sliding window (keep last WindowSize entries)
	cw.updateRecentEntries()

	return nil
}

// updateRecentEntries updates the RecentEntries to contain the last WindowSize entries.
// Must be called with write lock held.
func (cw *ContextWindow) updateRecentEntries() {
	totalCount := len(cw.AllEntries)

	// Validate WindowSize and handle edge cases
	if totalCount == 0 || cw.WindowSize <= 0 {
		cw.RecentEntries = []HistoryEntry{}
		return
	}

	if totalCount <= cw.WindowSize {
		// Less than or equal to window size: keep all (defensive copy)
		cw.RecentEntries = make([]HistoryEntry, totalCount)
		copy(cw.RecentEntries, cw.AllEntries)
	} else {
		// More than window size: keep last N (defensive copy)
		start := totalCount - cw.WindowSize
		cw.RecentEntries = make([]HistoryEntry, cw.WindowSize)
		copy(cw.RecentEntries, cw.AllEntries[start:])
	}
}

// GetAllEntries returns a copy of all history entries.
// This method is thread-safe and is used for summary generation.
func (cw *ContextWindow) GetAllEntries() []HistoryEntry {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	if cw.AllEntries == nil {
		return []HistoryEntry{}
	}

	// Return a copy to prevent external modification
	allEntries := make([]HistoryEntry, len(cw.AllEntries))
	copy(allEntries, cw.AllEntries)
	return allEntries
}

// GetEntryCount returns the total number of history entries.
// This method is thread-safe.
func (cw *ContextWindow) GetEntryCount() int {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return len(cw.AllEntries)
}

// Clear removes all history entries and resets the window.
// This method is thread-safe.
func (cw *ContextWindow) Clear() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.AllEntries = []HistoryEntry{}
	cw.RecentEntries = []HistoryEntry{}
	cw.Summary = ""
	cw.LastSummaryBeat = 0
	cw.IsSummaryInProgress = false
}

// GetLastEntry returns the most recent history entry.
// Returns (entry, true) if history exists, or (zero-value, false) if empty.
// This method is thread-safe.
func (cw *ContextWindow) GetLastEntry() (HistoryEntry, bool) {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	if len(cw.AllEntries) == 0 {
		return HistoryEntry{}, false
	}

	return cw.AllEntries[len(cw.AllEntries)-1], true
}

// ShouldGenerateSummary checks if a summary should be generated.
// Returns true if:
// 1. Enough entries have been added since last summary (≥ SUMMARY_TRIGGER)
// 2. No summary generation is currently in progress
// This method is thread-safe.
// Note: This is a check method only - actual generation should be triggered by the caller (e.g., Orchestrator).
func (cw *ContextWindow) ShouldGenerateSummary() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	// Don't trigger if summary is already being generated
	if cw.IsSummaryInProgress {
		return false
	}

	// Calculate entries since last summary
	entriesSinceLastSummary := len(cw.AllEntries) - cw.LastSummaryBeat

	// Trigger if we've accumulated enough entries
	return entriesSinceLastSummary >= SUMMARY_TRIGGER
}

// GetEntriesToSummarize returns the entries that should be summarized.
// Returns entries from LastSummaryBeat onwards.
// This method is thread-safe and returns a defensive copy.
func (cw *ContextWindow) GetEntriesToSummarize() []HistoryEntry {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	start := cw.LastSummaryBeat
	if start < 0 {
		start = 0
	}

	if start >= len(cw.AllEntries) {
		return []HistoryEntry{}
	}

	// Return defensive copy
	entries := make([]HistoryEntry, len(cw.AllEntries)-start)
	copy(entries, cw.AllEntries[start:])
	return entries
}

// UpdateSummary updates the summary text and marks the current beat as summarized.
// This method is thread-safe and typically called after successful summary generation.
// It applies length control: if combined summary exceeds TOTAL_SUMMARY_MAX_TOKENS,
// it keeps only the new summary (simple truncation strategy).
func (cw *ContextWindow) UpdateSummary(newSummary string) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Merge with existing summary if present
	var merged string
	if cw.Summary != "" {
		merged = cw.Summary + "\n\n" + newSummary
	} else {
		merged = newSummary
	}

	// Check if merged summary exceeds token limit
	estimatedTokens := cw.estimateTokens(merged)
	if estimatedTokens > TOTAL_SUMMARY_MAX_TOKENS {
		// Truncation strategy: keep only the new summary
		// TODO: Future enhancement - compress old + new instead of truncation
		cw.Summary = newSummary
	} else {
		cw.Summary = merged
	}

	// Update last summary beat to current entry count
	cw.LastSummaryBeat = len(cw.AllEntries)
	cw.IsSummaryInProgress = false
}

// estimateTokens estimates the number of tokens in a text.
// Uses simple character-based estimation: tokens ≈ characters / CHARS_PER_TOKEN
// This is a rough estimate suitable for length control purposes.
func (cw *ContextWindow) estimateTokens(text string) int {
	return len(text) / CHARS_PER_TOKEN
}

// SetSummaryInProgress sets the summary generation in-progress flag.
// This prevents concurrent summary generation attempts.
// This method is thread-safe.
func (cw *ContextWindow) SetSummaryInProgress(inProgress bool) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.IsSummaryInProgress = inProgress
}

// =============================================================================
// Token Counting & Optimization (Story 5.4)
// =============================================================================

// SetTokenCounter sets the token counter for this context window.
// This enables token counting and auto-optimization features.
// If counter is nil, token counting will be disabled.
// This method is thread-safe.
func (cw *ContextWindow) SetTokenCounter(counter TokenCounter) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.tokenCounter = counter
}

// GetTokenCounter returns the current token counter.
// Returns nil if no counter is set.
// This method is thread-safe.
func (cw *ContextWindow) GetTokenCounter() TokenCounter {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.tokenCounter
}

// CalculateTotalTokens calculates the total tokens used by current context.
// This includes:
//   - Summary tokens
//   - RecentEntries tokens (PlayerChoice + StoryContent)
// Returns 0 if no token counter is set.
// This method is thread-safe.
func (cw *ContextWindow) CalculateTotalTokens() int {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	if cw.tokenCounter == nil {
		return 0
	}

	// 1. Calculate Summary tokens
	summaryTokens := 0
	if cw.Summary != "" {
		summaryTokens = cw.tokenCounter.CountTokens(cw.Summary)
	}

	// 2. Calculate RecentEntries tokens
	entriesTokens := 0
	for _, entry := range cw.RecentEntries {
		// Combine PlayerChoice and StoryContent
		text := entry.PlayerChoice + " " + entry.StoryContent
		entriesTokens += cw.tokenCounter.CountTokens(text)
	}

	return summaryTokens + entriesTokens
}

// GetTokenUsageReport returns a detailed token usage report.
// This provides breakdown of token usage and warning status.
// Returns a zero-value report if no token counter is set.
// This method is thread-safe.
func (cw *ContextWindow) GetTokenUsageReport() TokenUsageReport {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	// Return zero report if no counter
	if cw.tokenCounter == nil {
		return TokenUsageReport{
			ModelLimit: MAX_CONTEXT_TOKENS,
		}
	}

	// Calculate individual components
	summaryTokens := 0
	if cw.Summary != "" {
		summaryTokens = cw.tokenCounter.CountTokens(cw.Summary)
	}

	entriesTokens := 0
	for _, entry := range cw.RecentEntries {
		text := entry.PlayerChoice + " " + entry.StoryContent
		entriesTokens += cw.tokenCounter.CountTokens(text)
	}

	totalTokens := summaryTokens + entriesTokens
	usagePercentage := float64(totalTokens) / float64(MAX_CONTEXT_TOKENS)

	return TokenUsageReport{
		SummaryTokens:       summaryTokens,
		RecentEntriesTokens: entriesTokens,
		TotalTokens:         totalTokens,
		ModelLimit:          MAX_CONTEXT_TOKENS,
		UsagePercentage:     usagePercentage,
		IsWarning:           usagePercentage > WARNING_THRESHOLD,
	}
}

// CheckTokenUsage checks if token usage exceeds thresholds.
// Returns (totalTokens, limit, warningLevel):
//   - warningLevel 0: Normal (< 70%)
//   - warningLevel 1: Warning (>= 70%, < 90%)
//   - warningLevel 2: Emergency (>= 90%)
// Returns (0, MAX_CONTEXT_TOKENS, 0) if no token counter is set.
// This method is thread-safe and is intended to be called by Orchestrator.
func (cw *ContextWindow) CheckTokenUsage() (totalTokens int, limit int, warningLevel int) {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	// Return safe values if no counter
	if cw.tokenCounter == nil {
		return 0, MAX_CONTEXT_TOKENS, 0
	}

	// Calculate total tokens
	totalTokens = cw.calculateTotalTokensUnsafe()
	limit = MAX_CONTEXT_TOKENS
	usageRatio := float64(totalTokens) / float64(limit)

	// Determine warning level
	if usageRatio >= EMERGENCY_THRESHOLD {
		warningLevel = 2 // Emergency
	} else if usageRatio >= WARNING_THRESHOLD {
		warningLevel = 1 // Warning
	} else {
		warningLevel = 0 // Normal
	}

	return totalTokens, limit, warningLevel
}

// calculateTotalTokensUnsafe is an internal helper that calculates tokens without locking.
// Must be called with read or write lock held.
func (cw *ContextWindow) calculateTotalTokensUnsafe() int {
	if cw.tokenCounter == nil {
		return 0
	}

	summaryTokens := 0
	if cw.Summary != "" {
		summaryTokens = cw.tokenCounter.CountTokens(cw.Summary)
	}

	entriesTokens := 0
	for _, entry := range cw.RecentEntries {
		text := entry.PlayerChoice + " " + entry.StoryContent
		entriesTokens += cw.tokenCounter.CountTokens(text)
	}

	return summaryTokens + entriesTokens
}

// ShrinkWindow reduces the window size by 1, down to MIN_WINDOW_SIZE.
// This is an emergency optimization when token usage is too high.
// Returns true if window was shrunk, false if already at minimum.
// This method is thread-safe and intended to be called by Orchestrator.
func (cw *ContextWindow) ShrinkWindow() bool {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.WindowSize <= MIN_WINDOW_SIZE {
		return false // Already at minimum
	}

	cw.WindowSize--
	cw.updateRecentEntries()
	return true
}
