package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

// BeatRangeConfig defines the beat ranges for each clue tier.
type BeatRangeConfig struct {
	Tier1Start int // Tier 1 start beat (default: 1)
	Tier1End   int // Tier 1 end beat (default: 5)
	Tier2Start int // Tier 2 start beat (default: 6)
	Tier2End   int // Tier 2 end beat (default: 12)
	Tier3Start int // Tier 3 start beat (default: 13)
	Tier3End   int // Tier 3 end beat (default: 18)
}

// DefaultBeatRangeConfig returns default beat ranges for standard game length.
func DefaultBeatRangeConfig() BeatRangeConfig {
	return BeatRangeConfig{
		Tier1Start: 1,
		Tier1End:   5,
		Tier2Start: 6,
		Tier2End:   12,
		Tier3Start: 13,
		Tier3End:   18,
	}
}

// SeedAgentConfig contains configuration for SeedAgent behavior.
type SeedAgentConfig struct {
	MaxRetries     int             // Maximum number of retry attempts (default: 3)
	RetryBackoff   int             // Base backoff time in seconds for exponential backoff (default: 1)
	EnableFallback bool            // Whether to use fallback seeds if all retries fail (default: true)
	BeatRanges     BeatRangeConfig // Beat ranges for clue tiers
}

// DefaultSeedAgentConfig returns a default configuration for SeedAgent.
func DefaultSeedAgentConfig() SeedAgentConfig {
	return SeedAgentConfig{
		MaxRetries:     3,
		RetryBackoff:   1,
		EnableFallback: true,
		BeatRanges:     DefaultBeatRangeConfig(),
	}
}

// SeedAgent is responsible for generating and managing foreshadowing seeds.
// It uses LLM to generate Global Seeds (main storyline) and Local Seeds (scene-specific).
type SeedAgent struct {
	provider api.Provider
	config   SeedAgentConfig
}

// NewSeedAgent creates a new SeedAgent with the given LLM provider and config.
func NewSeedAgent(provider api.Provider, config SeedAgentConfig) *SeedAgent {
	return &SeedAgent{
		provider: provider,
		config:   config,
	}
}

// GenerateGlobalParams contains parameters for generating Global Seeds.
type GenerateGlobalParams struct {
	WorldView  string // The world setting/view
	MainTheme  string // The main theme of the story
	Difficulty string // "easy", "medium", "hard"
}

// GlobalSeedJSON represents the JSON structure expected from LLM.
type GlobalSeedJSON struct {
	ID           string               `json:"id"`
	Content      string               `json:"content"`
	LinkedTruth  string               `json:"linked_truth"`
	LinkedEnding string               `json:"linked_ending"`
	ClueChain    []seed.ClueTier      `json:"clue_chain"`
	RelatedSeeds []string             `json:"related_seeds,omitempty"`
	RelatedRules []string             `json:"related_rules,omitempty"`
}

// GenerateGlobal generates Global Seeds based on the story parameters.
// The number of seeds depends on difficulty: Easy=3, Medium=4, Hard=5.
//
// Implements retry logic with exponential backoff and fallback to hardcoded seeds.
//
// Parameters:
//   - ctx: Context for cancellation
//   - params: Generation parameters (WorldView, MainTheme, Difficulty)
//
// Returns:
//   - []*seed.GlobalSeed: Generated global seeds
//   - error: If generation fails and fallback is disabled
func (sa *SeedAgent) GenerateGlobal(ctx context.Context, params GenerateGlobalParams) ([]*seed.GlobalSeed, error) {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled before generation: %w", err)
	}

	// Determine seed count based on difficulty
	seedCount := sa.getSeedCountByDifficulty(params.Difficulty)

	// Retry loop with exponential backoff
	var lastErr error
	for attempt := 0; attempt < sa.config.MaxRetries; attempt++ {
		// Check context cancellation between attempts
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled during retry %d: %w", attempt, err)
		}

		// Apply exponential backoff after first attempt
		if attempt > 0 {
			backoffDuration := time.Duration(attempt*attempt*sa.config.RetryBackoff) * time.Second
			select {
			case <-time.After(backoffDuration):
				// Backoff completed, continue to retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
			}
		}

		// Attempt to generate seeds
		seeds, err := sa.doGenerateGlobal(ctx, params, seedCount)
		if err == nil {
			// Success!
			return seeds, nil
		}

		// Store error for potential fallback
		lastErr = err
	}

	// All retries failed - use fallback if enabled
	if sa.config.EnableFallback {
		return sa.generateFallbackSeeds(params, seedCount), nil
	}

	return nil, fmt.Errorf("all %d retry attempts failed: %w", sa.config.MaxRetries, lastErr)
}

// doGenerateGlobal performs a single attempt at generating Global Seeds.
func (sa *SeedAgent) doGenerateGlobal(ctx context.Context, params GenerateGlobalParams, seedCount int) ([]*seed.GlobalSeed, error) {
	// Construct prompt
	prompt := sa.buildGlobalSeedPrompt(params, seedCount)

	// Call LLM
	messages := []api.Message{
		{Role: "system", Content: "You are a narrative design expert specializing in foreshadowing and story structure. You generate compelling, layered narrative seeds that create mystery and suspense."},
		{Role: "user", Content: prompt},
	}

	response, err := sa.provider.SendMessage(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Parse JSON response
	seeds, err := sa.parseGlobalSeedResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Validate we got the expected number of seeds
	if len(seeds) != seedCount {
		return nil, fmt.Errorf("expected %d seeds, got %d", seedCount, len(seeds))
	}

	return seeds, nil
}

// getSeedCountByDifficulty returns the number of Global Seeds to generate based on difficulty.
func (sa *SeedAgent) getSeedCountByDifficulty(difficulty string) int {
	switch strings.ToLower(difficulty) {
	case "easy":
		return 3
	case "medium":
		return 4
	case "hard":
		return 5
	default:
		return 4 // Default to medium
	}
}

// sanitizeInput sanitizes user input to prevent prompt injection attacks.
// Removes newlines, truncates long inputs, and filters suspicious patterns.
func sanitizeInput(input string) string {
	// Truncate to reasonable length (500 chars)
	if len(input) > 500 {
		input = input[:500]
	}

	// Replace newlines and carriage returns with spaces
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")

	// Remove common prompt injection patterns (case-insensitive)
	suspicious := []string{
		"ignore previous",
		"ignore all previous",
		"disregard previous",
		"forget previous",
		"new instructions:",
		"system:",
		"assistant:",
		"user:",
	}

	// Case-insensitive replacement
	lowerInput := strings.ToLower(input)
	for _, pattern := range suspicious {
		// Find all occurrences of the pattern (case-insensitive)
		patternLen := len(pattern)
		for {
			idx := strings.Index(lowerInput, pattern)
			if idx == -1 {
				break
			}
			// Replace in both original and lowercase versions
			replacement := strings.Repeat(" ", patternLen)
			input = input[:idx] + replacement + input[idx+patternLen:]
			lowerInput = lowerInput[:idx] + replacement + lowerInput[idx+patternLen:]
		}
	}

	// Trim excessive whitespace
	input = strings.TrimSpace(input)

	return input
}

// buildGlobalSeedPrompt constructs the LLM prompt for generating Global Seeds.
// Sanitizes all user-provided inputs to prevent prompt injection.
func (sa *SeedAgent) buildGlobalSeedPrompt(params GenerateGlobalParams, count int) string {
	// Sanitize all user inputs
	worldView := sanitizeInput(params.WorldView)
	mainTheme := sanitizeInput(params.MainTheme)
	difficulty := sanitizeInput(params.Difficulty)

	// Get beat ranges from config
	br := sa.config.BeatRanges

	return fmt.Sprintf(`Based on the following story skeleton, generate %d global seeds (main storyline foreshadowing elements):

World View: %s
Main Theme: %s
Difficulty: %s

For each global seed, provide:
1. A unique ID (format: "GS001", "GS002", etc.)
2. Core content of the foreshadowing (what the seed is about)
3. The truth it connects to (the revelation this seed leads to)
4. The ending type it contributes to (e.g., "tragic", "mysterious", "hopeful", "horrific")
5. A 3-tier clue chain with progressive revelation:
   - Tier 1 (beats %d-%d): Very subtle, barely noticeable hint
   - Tier 2 (beats %d-%d): More obvious clue that players should catch
   - Tier 3 (beats %d-%d): Almost explicit revelation that confirms the truth

Each tier should have:
- Tier number (1, 2, or 3)
- Content (the actual clue text)
- Keywords (2-4 keywords that must appear in the narrative when this clue is revealed)
- BeatStart and BeatEnd (the beat range when this clue can be revealed)

Requirements:
- Seeds should interconnect and reinforce each other
- Earlier tiers should be ambiguous and misleading
- Later tiers should confirm or subvert player expectations
- Keywords should be evocative and thematic
- The narrative should build suspense and dread

Output ONLY valid JSON in this exact format:
[
  {
    "id": "GS001",
    "content": "Brief description of the foreshadowing element",
    "linked_truth": "The truth this seed reveals",
    "linked_ending": "tragic",
    "clue_chain": [
      {
        "tier": 1,
        "content": "Very subtle hint that could be overlooked",
        "keywords": ["shadow", "whisper"],
        "beat_start": 1,
        "beat_end": 5
      },
      {
        "tier": 2,
        "content": "More obvious clue that confirms suspicions",
        "keywords": ["shadow", "truth", "watching"],
        "beat_start": 6,
        "beat_end": 12
      },
      {
        "tier": 3,
        "content": "Explicit revelation that ties everything together",
        "keywords": ["shadow", "truth", "reveal", "entity"],
        "beat_start": 13,
        "beat_end": 18
      }
    ]
  }
]

Generate %d seeds with this structure. Ensure they form a cohesive narrative web.`,
		count, worldView, mainTheme, difficulty,
		br.Tier1Start, br.Tier1End,
		br.Tier2Start, br.Tier2End,
		br.Tier3Start, br.Tier3End,
		count)
}

// parseGlobalSeedResponse parses the LLM JSON response into GlobalSeed structs.
func (sa *SeedAgent) parseGlobalSeedResponse(content string) ([]*seed.GlobalSeed, error) {
	// Clean the response - remove markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Parse JSON
	var seedsJSON []GlobalSeedJSON
	if err := json.Unmarshal([]byte(content), &seedsJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w\nFull LLM response:\n%s", err, content)
	}

	// Convert to seed.GlobalSeed structs
	seeds := make([]*seed.GlobalSeed, 0, len(seedsJSON))
	for i, sj := range seedsJSON {
		// Validate clue chain
		if len(sj.ClueChain) != 3 {
			return nil, fmt.Errorf("seed %d (%s): expected 3 clue tiers, got %d", i, sj.ID, len(sj.ClueChain))
		}

		// Create GlobalSeed
		gs, err := seed.NewGlobalSeed(
			sj.ID,
			sj.Content,
			sj.LinkedTruth,
			sj.LinkedEnding,
			sj.ClueChain,
		)
		if err != nil {
			return nil, fmt.Errorf("seed %d (%s): failed to create GlobalSeed: %w", i, sj.ID, err)
		}

		// Add related seeds and rules if provided
		for _, relatedSeed := range sj.RelatedSeeds {
			gs.AddRelatedSeed(relatedSeed)
		}
		for _, relatedRule := range sj.RelatedRules {
			gs.AddRelatedRule(relatedRule)
		}

		seeds = append(seeds, gs)
	}

	return seeds, nil
}

// generateFallbackSeeds creates simple, hardcoded Global Seeds when LLM generation fails.
// This ensures the game can always proceed even if the AI service is unavailable.
// Uses configured beat ranges for tier timing.
func (sa *SeedAgent) generateFallbackSeeds(params GenerateGlobalParams, count int) []*seed.GlobalSeed {
	seeds := make([]*seed.GlobalSeed, 0, count)

	// Get beat ranges from config
	br := sa.config.BeatRanges

	// Generate basic fallback seeds based on count
	for i := 0; i < count; i++ {
		gs, _ := seed.NewGlobalSeed(
			fmt.Sprintf("GS%03d", i+1),
			fmt.Sprintf("A mysterious element connected to %s (fallback seed %d)", params.MainTheme, i+1),
			fmt.Sprintf("A hidden truth about the %s world", params.MainTheme),
			"mysterious",
			[]seed.ClueTier{
				{
					Tier:      1,
					Content:   "Something seems unusual here...",
					Keywords:  []string{"mystery", "strange"},
					BeatStart: br.Tier1Start,
					BeatEnd:   br.Tier1End,
				},
				{
					Tier:      2,
					Content:   "The pattern becomes clearer now.",
					Keywords:  []string{"mystery", "pattern", "truth"},
					BeatStart: br.Tier2Start,
					BeatEnd:   br.Tier2End,
				},
				{
					Tier:      3,
					Content:   "The full truth is revealed at last.",
					Keywords:  []string{"mystery", "truth", "revelation"},
					BeatStart: br.Tier3Start,
					BeatEnd:   br.Tier3End,
				},
			},
		)
		seeds = append(seeds, gs)
	}

	return seeds
}
