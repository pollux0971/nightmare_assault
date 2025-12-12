// Package engine provides the story generation engine for Nightmare Assault.
package engine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/builder"
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
		TimeoutFirstToken: 0, // No timeout - let slow models complete
		TimeoutTotal:      0, // No timeout - let slow models complete
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

// ProgressCallback is called with progress updates (0-100).
type ProgressCallback func(percent int, state EstimationState)

// GenerationResult represents the result of story generation.
type GenerationResult struct {
	Content     string
	Choices     []string
	Seeds       []builder.SeedInfo
	Error       error
	TimeTaken   time.Duration
	RetryCount  int
}

// GenerateOpening generates the opening story segment.
func (e *StoryEngine) GenerateOpening(ctx context.Context, streamCallback StreamCallback, progressCallback ProgressCallback) (*GenerationResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config.Provider == nil {
		return nil, ErrNoProvider
	}

	systemPrompt := prompts.BuildSystemPrompt(e.config.GameConfig)
	userPrompt := prompts.BuildOpeningPrompt(e.config.GameConfig)

	// Estimate duration based on prompt length
	promptLen := len(systemPrompt) + len(userPrompt)
	expectedDuration := EstimateDurationFromLength(EstimateFromPromptLength(promptLen))

	return e.generateWithRetry(ctx, systemPrompt, userPrompt, expectedDuration, streamCallback, progressCallback)
}

// GenerateContinuation generates the next story segment based on player choice.
func (e *StoryEngine) GenerateContinuation(ctx context.Context, choice string, streamCallback StreamCallback, progressCallback ProgressCallback) (*GenerationResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config.Provider == nil {
		return nil, ErrNoProvider
	}

	systemPrompt := prompts.BuildSystemPrompt(e.config.GameConfig)
	context := e.storyState.GetContextSummary(3) // Last 3 segments for context
	userPrompt := prompts.BuildContinuationPrompt(choice, context)

	// Estimate duration based on prompt length
	promptLen := len(systemPrompt) + len(userPrompt)
	expectedDuration := EstimateDurationFromLength(EstimateFromPromptLength(promptLen))

	return e.generateWithRetry(ctx, systemPrompt, userPrompt, expectedDuration, streamCallback, progressCallback)
}

func (e *StoryEngine) generateWithRetry(ctx context.Context, systemPrompt, userPrompt string, expectedDuration time.Duration, streamCallback StreamCallback, progressCallback ProgressCallback) (*GenerationResult, error) {
	var lastErr error
	startTime := time.Now()

	// Create progress estimator
	estimator := NewProgressEstimator(expectedDuration)

	// Send initial progress
	if progressCallback != nil {
		estimator.SetState(StateConnecting)
		progressCallback(estimator.GetProgress(), StateConnecting)
	}

	for attempt := 0; attempt < e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := e.config.RetryBaseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			// Reset estimator for retry
			estimator.Reset()
		}

		result, err := e.generate(ctx, systemPrompt, userPrompt, estimator, streamCallback, progressCallback)
		if err == nil {
			result.RetryCount = attempt
			result.TimeTaken = time.Since(startTime)
			// Send completion progress
			if progressCallback != nil {
				estimator.ForceComplete()
				progressCallback(100, StateCompleting)
			}
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

func (e *StoryEngine) generate(ctx context.Context, systemPrompt, userPrompt string, estimator *ProgressEstimator, streamCallback StreamCallback, progressCallback ProgressCallback) (*GenerationResult, error) {
	// Use context directly without timeout - let slow models complete
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	messages := []api.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	var content string
	var firstTokenReceived bool
	// No first token timeout - disabled for slow models
	var firstTokenTimer *time.Timer
	if e.config.TimeoutFirstToken > 0 {
		firstTokenTimer = time.NewTimer(e.config.TimeoutFirstToken)
		defer firstTokenTimer.Stop()
	}

	// Channel to collect streamed content
	contentChan := make(chan string, 100)
	errChan := make(chan error, 1)

	// Progress update ticker
	progressTicker := time.NewTicker(200 * time.Millisecond)
	defer progressTicker.Stop()

	// Start in generating state
	estimator.SetState(StateGenerating)
	if progressCallback != nil {
		progressCallback(estimator.GetProgress(), StateGenerating)
	}

	go func() {
		err := e.config.Provider.Stream(ctx, messages, func(chunk string) {
			if !firstTokenReceived {
				firstTokenReceived = true
				if firstTokenTimer != nil {
					firstTokenTimer.Stop()
				}
				// Switch to streaming state on first token
				estimator.SetState(StateStreaming)
			}
			contentChan <- chunk
			// Update stream bytes for progress estimation
			estimator.UpdateStreamBytes(len(content) + len(chunk))
			if streamCallback != nil {
				streamCallback(chunk)
			}
		})
		errChan <- err
		close(contentChan)
	}()

	// Wait for first token (no timeout for slow models)
	if firstTokenTimer != nil {
		// With timeout enabled
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
	} else {
		// No timeout - wait indefinitely for first token
		select {
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
	}

	// Collect remaining content with progress updates
	collecting := true
	for collecting {
		select {
		case chunk, ok := <-contentChan:
			if !ok {
				collecting = false
				break
			}
			content += chunk
		case <-progressTicker.C:
			// Send progress update
			if progressCallback != nil {
				progressCallback(estimator.GetProgress(), estimator.state)
			}
		}
	}

	// Check for streaming error
	if err := <-errChan; err != nil {
		return nil, err
	}

	// Parse structured output (JSON or legacy format)
	output, err := builder.ParseStructuredOutput(content)
	if err != nil {
		// If parsing fails, fall back to basic content
		output = &builder.StoryOutput{
			Story:   content,
			Choices: []string{},
			Seeds:   []builder.SeedInfo{},
		}
	}

	result := &GenerationResult{
		Content: output.Story,
		Choices: output.Choices,
		Seeds:   output.Seeds,
	}

	// Add seeds to story state
	for i, seed := range output.Seeds {
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
		Content:   output.Story,
		Choices:   output.Choices,
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

// generateSeedID creates a unique ID for a seed based on beat and index.
func generateSeedID(beat, index int) string {
	return string(rune('S')) + string(rune('0'+beat%10)) + string(rune('0'+index%10))
}
