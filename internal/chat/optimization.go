package chat

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ParallelResponseGenerator generates NPC responses concurrently.
//
// Story 5-8 AC4: Parallel processing of multiple NPC responses
//
// Features:
// - Goroutine pool to control concurrency (max 3 concurrent)
// - Results returned in participant order
// - Timeout control per NPC (5 seconds)
// - Error handling (partial failures don't affect other NPCs)
type ParallelResponseGenerator struct {
	generator       *NPCResponseGenerator
	maxConcurrency  int
	npcTimeout      time.Duration
	orderPreserving bool
}

// responseResult represents the result of generating a single NPC response.
type responseResult struct {
	NPCID    string
	Response NPCResponse
	Err      error
	Index    int // Original index in participant list for ordering
}

// NewParallelResponseGenerator creates a new parallel response generator.
//
// Story 5-8 AC4: Initialize parallel response generator with concurrency control
//
// Parameters:
//   - generator: The NPCResponseGenerator to use for each NPC
//   - maxConcurrency: Maximum number of concurrent LLM calls (default: 3)
//   - npcTimeout: Timeout per NPC response (default: 5 seconds)
//   - orderPreserving: Whether to return responses in original order
func NewParallelResponseGenerator(
	generator *NPCResponseGenerator,
	maxConcurrency int,
	npcTimeout time.Duration,
	orderPreserving bool,
) *ParallelResponseGenerator {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // Default to 3 concurrent requests
	}

	if npcTimeout <= 0 {
		npcTimeout = 5 * time.Second // Default to 5s timeout
	}

	logger.Debug("ParallelResponseGenerator created", map[string]interface{}{
		"max_concurrency":  maxConcurrency,
		"npc_timeout":      npcTimeout.String(),
		"order_preserving": orderPreserving,
	})

	return &ParallelResponseGenerator{
		generator:       generator,
		maxConcurrency:  maxConcurrency,
		npcTimeout:      npcTimeout,
		orderPreserving: orderPreserving,
	}
}

// GenerateAllResponses generates responses for all NPCs in parallel.
//
// Story 5-8 AC4: Parallel processing with goroutine pool
//
// This method:
// 1. Creates a semaphore to limit concurrency
// 2. Launches goroutines for each NPC
// 3. Collects results via channel
// 4. Returns responses in participant order (if orderPreserving=true)
// 5. Handles timeouts and errors gracefully
//
// Parameters:
//   - ctx: Context for cancellation
//   - session: The chat session
//   - playerMessage: The player's message
//   - emotionChanges: Emotion deltas to apply
//   - flags: Chat flags from JudgeAgent
//
// Returns:
//   - []NPCResponse: List of responses (in participant order if orderPreserving=true)
//   - error: Only if all NPCs failed
func (pg *ParallelResponseGenerator) GenerateAllResponses(
	ctx context.Context,
	session *ChatSession,
	playerMessage string,
	emotionChanges map[string]manager.EmotionDelta,
	flags []ChatFlag,
) ([]NPCResponse, error) {
	// Get NPC participants
	npcParticipants := []ChatParticipant{}
	for _, p := range session.Participants {
		if !p.IsPlayer {
			npcParticipants = append(npcParticipants, p)
		}
	}

	if len(npcParticipants) == 0 {
		return []NPCResponse{}, nil
	}

	logger.Debug("Generating NPC responses in parallel", map[string]interface{}{
		"num_npcs":        len(npcParticipants),
		"max_concurrency": pg.maxConcurrency,
		"timeout":         pg.npcTimeout.String(),
	})

	// Create result channel
	results := make(chan responseResult, len(npcParticipants))

	// Create semaphore to limit concurrency
	sem := make(chan struct{}, pg.maxConcurrency)

	// Launch goroutines for each NPC
	var wg sync.WaitGroup
	for i, participant := range npcParticipants {
		wg.Add(1)
		go func(index int, p ChatParticipant) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }() // Release semaphore

			// Generate response for this NPC
			response, err := pg.generateSingleNPCResponse(
				ctx,
				p,
				playerMessage,
				session.MessageHistory,
				emotionChanges,
				flags,
			)

			results <- responseResult{
				NPCID:    p.ID,
				Response: response,
				Err:      err,
				Index:    index,
			}
		}(i, participant)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	responseMap := make(map[int]NPCResponse)
	errors := []error{}

	for result := range results {
		if result.Err != nil {
			logger.Warn("NPC response generation failed", map[string]interface{}{
				"npc_id": result.NPCID,
				"error":  result.Err.Error(),
			})
			errors = append(errors, result.Err)
			// Don't include failed responses
			continue
		}

		responseMap[result.Index] = result.Response
	}

	// If all NPCs failed, return error
	if len(responseMap) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all NPC response generation failed: %d errors", len(errors))
	}

	// Convert map to ordered slice
	responses := []NPCResponse{}
	if pg.orderPreserving {
		// Return in original participant order
		for i := 0; i < len(npcParticipants); i++ {
			if resp, ok := responseMap[i]; ok {
				responses = append(responses, resp)
			}
		}
	} else {
		// Return in any order (faster)
		for _, resp := range responseMap {
			responses = append(responses, resp)
		}
	}

	logger.Debug("Parallel NPC response generation completed", map[string]interface{}{
		"num_responses": len(responses),
		"num_errors":    len(errors),
	})

	return responses, nil
}

// generateSingleNPCResponse generates a response for a single NPC with timeout.
//
// Story 5-8 AC4: Single NPC response with timeout control (5 seconds)
//
// This wraps the NPCResponseGenerator.GenerateNPCResponse with a timeout context.
func (pg *ParallelResponseGenerator) generateSingleNPCResponse(
	ctx context.Context,
	participant ChatParticipant,
	playerMessage string,
	conversationHistory []ChatMessage,
	emotionChanges map[string]manager.EmotionDelta,
	flags []ChatFlag,
) (NPCResponse, error) {
	// Create timeout context for this NPC
	ctxWithTimeout, cancel := context.WithTimeout(ctx, pg.npcTimeout)
	defer cancel()

	// Get current emotion (apply changes if any)
	currentEmotion := participant.Emotion
	if delta, exists := emotionChanges[participant.ID]; exists {
		currentEmotion = currentEmotion.Apply(delta)
	}

	// Generate response
	response, err := pg.generator.GenerateNPCResponse(
		ctxWithTimeout,
		participant.ID,
		playerMessage,
		conversationHistory,
		flags,
		currentEmotion,
	)

	if err != nil {
		// Check if it was a timeout
		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			logger.Warn("NPC response generation timed out", map[string]interface{}{
				"npc_id":  participant.ID,
				"timeout": pg.npcTimeout.String(),
			})
			return NPCResponse{}, fmt.Errorf("NPC %s response timed out after %v", participant.ID, pg.npcTimeout)
		}

		return NPCResponse{}, fmt.Errorf("NPC %s response failed: %w", participant.ID, err)
	}

	return response, nil
}

// OptimizedPromptBuilder builds optimized prompts with reduced token consumption.
//
// Story 5-8 AC2: Prompt optimization to reduce token usage
//
// Optimizations:
// - Limit conversation history to recent N messages (default: 5)
// - Compress NPC profile to essential information only
// - Use concise instruction language
// - Remove redundant context
type OptimizedPromptBuilder struct {
	maxHistoryMessages int
}

// NewOptimizedPromptBuilder creates a new optimized prompt builder.
func NewOptimizedPromptBuilder(maxHistoryMessages int) *OptimizedPromptBuilder {
	if maxHistoryMessages <= 0 {
		maxHistoryMessages = 5 // Default to 5 messages
	}

	return &OptimizedPromptBuilder{
		maxHistoryMessages: maxHistoryMessages,
	}
}

// BuildOptimizedPrompt builds an optimized prompt for NPC response generation.
//
// Story 5-8 AC2: Reduce prompt length while maintaining quality
//
// This is a simplified version of the full prompt that:
// - Only includes essential NPC info (name, archetype, emotion)
// - Limits history to last N messages
// - Uses concise instructions
// - Omits verbose context
func (pb *OptimizedPromptBuilder) BuildOptimizedPrompt(
	npcName string,
	npcArchetype string,
	currentEmotion manager.EmotionState,
	recentMessages []ChatMessage,
	playerMessage string,
) string {
	// Build compact prompt
	prompt := fmt.Sprintf("你是%s，%s。", npcName, npcArchetype)

	// Add emotion state (compact)
	prompt += fmt.Sprintf("情緒：信任%d 恐懼%d 壓力%d。",
		currentEmotion.Trust, currentEmotion.Fear, currentEmotion.Stress)

	// Add recent history (limited)
	if len(recentMessages) > 0 {
		prompt += "\n對話："
		start := len(recentMessages) - pb.maxHistoryMessages
		if start < 0 {
			start = 0
		}
		for _, msg := range recentMessages[start:] {
			speaker := msg.SenderID
			if speaker == "player" {
				speaker = "玩家"
			}
			prompt += fmt.Sprintf("\n%s:%s", speaker, msg.Content)
		}
	}

	// Add player message
	prompt += fmt.Sprintf("\n玩家:%s", playerMessage)

	// Add instruction (concise)
	prompt += "\n回應(簡短自然,<50字):"

	return prompt
}

// EstimateTokens roughly estimates the token count for a string.
//
// Story 5-8 AC2: Token estimation for monitoring
//
// This is a rough approximation:
// - English: ~4 chars per token
// - Chinese: ~2 chars per token
// - Mixed: average of both
func EstimateTokens(text string) int {
	// Simple heuristic: count chars and divide by average chars per token
	// For Chinese text, tokens are roughly 1.5-2 chars per token
	// For English text, tokens are roughly 4 chars per token
	// We'll use 3 as a middle ground

	charCount := len(text)
	return (charCount + 2) / 3 // Round up
}
