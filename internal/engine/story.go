// Package engine provides the story generation engine for Nightmare Assault.
package engine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// Error definitions
var (
	ErrTimeout        = errors.New("故事生成超時")
	ErrNoProvider     = errors.New("未配置 LLM 提供者")
	ErrGenerationFail = errors.New("故事生成失敗")
	ErrMaxRetries     = errors.New("已達最大重試次數")
)

// SeedType represents the type of hidden seed.
type SeedType string

const (
	SeedTypeItem      SeedType = "Item"
	SeedTypeEvent     SeedType = "Event"
	SeedTypeCharacter SeedType = "Character"
	SeedTypeLocation  SeedType = "Location"
)

// HiddenSeed represents a hidden story element planted in the narrative.
type HiddenSeed struct {
	ID          string
	Type        SeedType
	Description string
	TriggerBeat int  // When this seed activates
	Discovered  bool // Whether player has discovered this seed
}

// StorySegment represents a single story beat/segment.
type StorySegment struct {
	Content     string
	Choices     []string
	HPChange    int
	SANChange   int
	Timestamp   time.Time
	SeedsPlanted []string // IDs of seeds planted in this segment
}

// StoryState tracks the current state of the story.
type StoryState struct {
	CurrentBeat  int
	History      []StorySegment
	ActiveSeeds  []HiddenSeed
	ContextHash  string
	TotalHP      int
	TotalSAN     int
	mu           sync.RWMutex
}

// NewStoryState creates a new story state with initial values.
func NewStoryState() *StoryState {
	return &StoryState{
		CurrentBeat: 0,
		History:     make([]StorySegment, 0),
		ActiveSeeds: make([]HiddenSeed, 0),
		TotalHP:     100,
		TotalSAN:    100,
	}
}

// AddSegment adds a new story segment to history.
func (s *StoryState) AddSegment(segment StorySegment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.History = append(s.History, segment)
	s.CurrentBeat++
}

// AddSeed adds a new hidden seed to track.
func (s *StoryState) AddSeed(seed HiddenSeed) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActiveSeeds = append(s.ActiveSeeds, seed)
}

// GetLastSegment returns the most recent story segment.
func (s *StoryState) GetLastSegment() *StorySegment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.History) == 0 {
		return nil
	}
	return &s.History[len(s.History)-1]
}

// GetContextSummary returns a summary for LLM context.
func (s *StoryState) GetContextSummary(maxSegments int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.History) == 0 {
		return ""
	}

	start := 0
	if len(s.History) > maxSegments {
		start = len(s.History) - maxSegments
	}

	var summary string
	for i := start; i < len(s.History); i++ {
		summary += s.History[i].Content + "\n\n"
	}
	return summary
}

// StoryEngineConfig holds configuration for the story engine.
type StoryEngineConfig struct {
	Provider           api.Provider
	GameConfig         *game.GameConfig
	TimeoutFirstToken  time.Duration
	TimeoutTotal       time.Duration
	MaxRetries         int
	RetryBaseDelay     time.Duration
}

// DefaultEngineConfig returns default engine configuration.
func DefaultEngineConfig() StoryEngineConfig {
	return StoryEngineConfig{
		TimeoutFirstToken: 5 * time.Second,
		TimeoutTotal:      30 * time.Second,
		MaxRetries:        3,
		RetryBaseDelay:    1 * time.Second,
	}
}

// StoryEngine handles story generation.
type StoryEngine struct {
	config     StoryEngineConfig
	storyState *StoryState
	mu         sync.Mutex
}

// NewStoryEngine creates a new story engine.
func NewStoryEngine(config StoryEngineConfig) *StoryEngine {
	return &StoryEngine{
		config:     config,
		storyState: NewStoryState(),
	}
}

// StreamCallback is called for each chunk of streamed content.
type StreamCallback func(chunk string)

// GenerationResult represents the result of story generation.
type GenerationResult struct {
	Content     string
	Choices     []string
	Seeds       []prompts.SeedInfo
	Error       error
	TimeTaken   time.Duration
	RetryCount  int
}

// GenerateOpening generates the opening story segment.
func (e *StoryEngine) GenerateOpening(ctx context.Context, callback StreamCallback) (*GenerationResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config.Provider == nil {
		return nil, ErrNoProvider
	}

	systemPrompt := prompts.BuildSystemPrompt(e.config.GameConfig)
	userPrompt := prompts.BuildOpeningPrompt(e.config.GameConfig)

	return e.generateWithRetry(ctx, systemPrompt, userPrompt, callback)
}

// GenerateContinuation generates the next story segment based on player choice.
func (e *StoryEngine) GenerateContinuation(ctx context.Context, choice string, callback StreamCallback) (*GenerationResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config.Provider == nil {
		return nil, ErrNoProvider
	}

	systemPrompt := prompts.BuildSystemPrompt(e.config.GameConfig)
	context := e.storyState.GetContextSummary(3) // Last 3 segments for context
	userPrompt := prompts.BuildContinuationPrompt(choice, context)

	return e.generateWithRetry(ctx, systemPrompt, userPrompt, callback)
}

func (e *StoryEngine) generateWithRetry(ctx context.Context, systemPrompt, userPrompt string, callback StreamCallback) (*GenerationResult, error) {
	var lastErr error
	startTime := time.Now()

	for attempt := 0; attempt < e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := e.config.RetryBaseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := e.generate(ctx, systemPrompt, userPrompt, callback)
		if err == nil {
			result.RetryCount = attempt
			result.TimeTaken = time.Since(startTime)
			return result, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}
	}

	return nil, errors.Join(ErrMaxRetries, lastErr)
}

func (e *StoryEngine) generate(ctx context.Context, systemPrompt, userPrompt string, callback StreamCallback) (*GenerationResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.config.TimeoutTotal)
	defer cancel()

	messages := []api.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	var content string
	var firstTokenReceived bool
	firstTokenTimer := time.NewTimer(e.config.TimeoutFirstToken)
	defer firstTokenTimer.Stop()

	// Channel to collect streamed content
	contentChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		err := e.config.Provider.Stream(ctx, messages, func(chunk string) {
			if !firstTokenReceived {
				firstTokenReceived = true
				firstTokenTimer.Stop()
			}
			contentChan <- chunk
			if callback != nil {
				callback(chunk)
			}
		})
		errChan <- err
		close(contentChan)
	}()

	// Wait for first token or timeout
	select {
	case <-firstTokenTimer.C:
		if !firstTokenReceived {
			cancel()
			return nil, ErrTimeout
		}
	case chunk := <-contentChan:
		content += chunk
		firstTokenReceived = true
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Collect remaining content
	for chunk := range contentChan {
		content += chunk
	}

	// Check for streaming error
	if err := <-errChan; err != nil {
		return nil, err
	}

	// Extract seeds from content
	seeds := prompts.ExtractSeeds(content)
	cleanContent := prompts.CleanContent(content)

	// Parse choices from content
	choices := parseChoices(cleanContent)

	result := &GenerationResult{
		Content: cleanContent,
		Choices: choices,
		Seeds:   seeds,
	}

	// Add seeds to story state
	for i, seed := range seeds {
		e.storyState.AddSeed(HiddenSeed{
			ID:          generateSeedID(e.storyState.CurrentBeat, i),
			Type:        SeedType(seed.Type),
			Description: seed.Description,
			TriggerBeat: e.storyState.CurrentBeat + 3 + (i % 3), // Trigger 3-5 beats later
			Discovered:  false,
		})
	}

	// Add segment to story state
	e.storyState.AddSegment(StorySegment{
		Content:   cleanContent,
		Choices:   choices,
		Timestamp: time.Now(),
	})

	return result, nil
}

// GetStoryState returns the current story state.
func (e *StoryEngine) GetStoryState() *StoryState {
	return e.storyState
}

// SetProvider sets the LLM provider.
func (e *StoryEngine) SetProvider(provider api.Provider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.Provider = provider
}

// SetGameConfig sets the game configuration.
func (e *StoryEngine) SetGameConfig(config *game.GameConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.GameConfig = config
}

// Reset resets the story engine for a new game.
func (e *StoryEngine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.storyState = NewStoryState()
}

// parseChoices extracts choices from story content.
func parseChoices(content string) []string {
	var choices []string
	lines := splitLines(content)

	inChoices := false
	for _, line := range lines {
		line = trimSpace(line)

		if inChoices {
			// Match numbered choices: 1. or 1) or 1、
			if isNumberedChoice(line) {
				choice := extractChoiceText(line)
				if choice != "" {
					choices = append(choices, choice)
				}
				continue
			} else if line == "" {
				// Empty line might end choices
				if len(choices) > 0 {
					break
				}
			}
		}

		// Detect choice section start (check after numbered choices)
		if containsChoiceHeader(line) {
			inChoices = true
		}
	}

	return choices
}

func containsChoiceHeader(line string) bool {
	// Chinese headers
	chineseHeaders := []string{"選擇", "選項"}
	for _, h := range chineseHeaders {
		if containsRunes(line, h) {
			return true
		}
	}
	// English headers (case insensitive)
	englishHeaders := []string{"choices", "options"}
	lower := toLower(line)
	for _, h := range englishHeaders {
		if contains(lower, h) {
			return true
		}
	}
	return false
}

func containsRunes(s, substr string) bool {
	sRunes := []rune(s)
	subRunes := []rune(substr)
	if len(sRunes) < len(subRunes) {
		return false
	}
	for i := 0; i <= len(sRunes)-len(subRunes); i++ {
		match := true
		for j := 0; j < len(subRunes); j++ {
			if sRunes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func isNumberedChoice(line string) bool {
	if len(line) < 2 {
		return false
	}
	// Check for digit followed by separator
	if line[0] >= '1' && line[0] <= '9' {
		if line[1] == '.' || line[1] == ')' {
			return true
		}
		// Check for Chinese separator (、is multi-byte)
		runes := []rune(line)
		if len(runes) >= 2 && runes[1] == '、' {
			return true
		}
	}
	return false
}

func extractChoiceText(line string) string {
	runes := []rune(line)
	if len(runes) < 2 {
		return ""
	}
	// Skip "1. " or "1) " or "1、"
	text := string(runes[2:])
	if len(text) > 0 && text[0] == ' ' {
		text = text[1:]
	}
	return trimSpace(text)
}

func generateSeedID(beat, index int) string {
	return string(rune('S')) + string(rune('0'+beat%10)) + string(rune('0'+index%10))
}

// Helper functions to avoid strings package dependency in hot path
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
