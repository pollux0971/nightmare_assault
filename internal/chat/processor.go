package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ChatProcessor is the central coordinator for chat message processing.
// It orchestrates JudgeAgent judgment, flag handlers, UpdateManager propagation,
// and NPC response generation.
//
// Story 4-2 AC1: Contains all required dependencies
// Story 5-8: Enhanced with performance monitoring and caching
type ChatProcessor struct {
	npcManager      *manager.NPCManager
	updateManager   *knowledge.UpdateManager
	judgeAgent      *agents.JudgeAgent
	llmClient       agents.LLMClient
	handlerRegistry *HandlerRegistry
	config          *ChatProcessorConfig

	// Story 5-8: Performance optimization components
	metrics       *ChatMetrics      // Performance metrics tracking
	responseCache *ResponseCache    // Response caching (LRU + TTL)
}

// ChatProcessorConfig contains configuration for the ChatProcessor.
type ChatProcessorConfig struct {
	// MaxRetries for LLM calls (default: 3)
	MaxRetries int

	// Timeout for chat processing operations (default: 30s)
	Timeout time.Duration

	// EnableGracefulDegradation allows processing to continue even if some components fail
	EnableGracefulDegradation bool

	// Story 5-8 AC3: Cache configuration
	EnableCache       bool          // Enable response caching (default: true)
	CacheMaxSize      int           // Max cached entries (default: 100)
	CacheTTL          time.Duration // Cache TTL (default: 10 minutes)

	// Story 5-8 AC1: Performance monitoring
	EnableMetrics     bool          // Enable performance metrics (default: true)
}

// DefaultChatProcessorConfig returns a ChatProcessorConfig with sensible defaults.
func DefaultChatProcessorConfig() *ChatProcessorConfig {
	return &ChatProcessorConfig{
		MaxRetries:                3,
		Timeout:                   30 * time.Second,
		EnableGracefulDegradation: true,

		// Story 5-8: Performance optimization defaults
		EnableCache:   true,
		CacheMaxSize:  100,
		CacheTTL:      10 * time.Minute,
		EnableMetrics: true,
	}
}

// NewChatProcessor creates a new ChatProcessor with the given dependencies.
// If config is nil, it uses DefaultChatProcessorConfig().
//
// Story 4-2 AC1: Constructor with all required dependencies
// Story 5-8: Initialize metrics and cache if enabled
func NewChatProcessor(
	npcManager *manager.NPCManager,
	updateManager *knowledge.UpdateManager,
	judgeAgent *agents.JudgeAgent,
	llmClient agents.LLMClient,
	config *ChatProcessorConfig,
) *ChatProcessor {
	if config == nil {
		config = DefaultChatProcessorConfig()
	}

	cp := &ChatProcessor{
		npcManager:      npcManager,
		updateManager:   updateManager,
		judgeAgent:      judgeAgent,
		llmClient:       llmClient,
		handlerRegistry: NewHandlerRegistry(),
		config:          config,
	}

	// Story 5-8 AC1: Initialize metrics if enabled
	if config.EnableMetrics {
		cp.metrics = NewChatMetrics()
		logger.Debug("ChatProcessor metrics enabled", nil)
	}

	// Story 5-8 AC3: Initialize response cache if enabled
	if config.EnableCache {
		cache, err := NewResponseCache(config.CacheMaxSize, config.CacheTTL)
		if err != nil {
			logger.Warn("Failed to create response cache, caching disabled", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			cp.responseCache = cache
			logger.Debug("ChatProcessor response cache enabled", map[string]interface{}{
				"max_size": config.CacheMaxSize,
				"ttl":      config.CacheTTL.String(),
			})
		}
	}

	return cp
}

// ProcessPlayerMessage processes a player's chat message and returns a complete ProcessResult.
//
// Story 4-2 AC2: ProcessPlayerMessage() handles player messages and returns ProcessResult
// Story 4-2 AC3: Calls JudgeAgent.JudgeChat() for judgment
// Story 4-2 AC4: Propagates information to UpdateManager
// Story 4-2 AC5: Returns ProcessResult with NPCResponses/EmotionChanges/Flags
//
// Implementation Flow:
// 1. Validate input (non-empty message, valid session, at least 1 NPC)
// 2. Build JudgeChatRequest (participants, history, relevant facts)
// 3. Call judgeAgent.JudgeChat()
// 4. Process flags using HandlerRegistry.ExecuteAllHandlers()
// 5. Propagate player message as Fact to UpdateManager
// 6. Collect emotion changes from flag handlers
// 7. Generate NPC responses (STUB for Story 4-4)
// 8. Return complete ProcessResult
func (cp *ChatProcessor) ProcessPlayerMessage(ctx context.Context, session *ChatSession, message string, gameState *agents.GameStateSnapshot) (*ProcessResult, error) {
	logger.Debug("ChatProcessor.ProcessPlayerMessage invoked", map[string]interface{}{
		"session_id":       session.SessionID,
		"message":          message,
		"num_participants": len(session.Participants),
	})

	// Step 1: Validate input
	if err := cp.validateProcessRequest(session, message); err != nil {
		return &ProcessResult{
			Success: false,
			Error:   fmt.Sprintf("Validation failed: %v", err),
		}, err
	}

	// Step 2: Build JudgeChatRequest with participants, history, and relevant facts
	judgeRequest, err := cp.buildJudgeChatRequest(session, message, gameState)
	if err != nil {
		logger.Warn("Failed to build JudgeChatRequest", map[string]interface{}{
			"error": err.Error(),
		})
		return &ProcessResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to build judge request: %v", err),
		}, err
	}

	// Step 3: Call JudgeAgent.JudgeChat() for message analysis
	judgeResult, err := cp.judgeAgent.JudgeChat(ctx, judgeRequest)
	if err != nil {
		// AC: Graceful degradation - JudgeAgent failure uses empty flags
		logger.Warn("JudgeAgent.JudgeChat failed, using empty flags", map[string]interface{}{
			"error": err.Error(),
		})
		judgeResult = &agents.JudgeChatResult{
			Flags:      []agents.ChatFlag{},
			Confidence: 0.0,
			Reasoning:  fmt.Sprintf("Judge failed: %v", err),
		}
	}

	// Convert agents.ChatFlag to chat.ChatFlag for processing
	chatFlags := cp.convertToChatFlags(judgeResult.Flags)

	logger.Debug("JudgeAgent judgment completed", map[string]interface{}{
		"num_flags":  len(chatFlags),
		"flags":      cp.flagsToStrings(chatFlags),
		"confidence": judgeResult.Confidence,
	})

	// Step 4 & 5 & 6: Process flags, propagate information, collect emotion changes
	emotionChanges := make(map[string]manager.EmotionDelta)
	contradictions := []knowledge.ContradictionResult{}

	// Get all NPC participants
	npcParticipants := cp.getNPCParticipants(session)

	// Story 4-7 AC3: Check for contradictions for each NPC BEFORE processing flags
	// This allows contradiction detection to inform flag handling
	// We also build a map to associate contradictions with specific NPCs
	contradictionMap := make(map[string]*knowledge.ContradictionResult)
	for _, participant := range npcParticipants {
		if cp.updateManager != nil {
			// Check if player's message contradicts what this NPC knows
			contradiction := cp.updateManager.CheckContradiction(participant.ID, message)
			if contradiction != nil {
				contradictions = append(contradictions, *contradiction)
				contradictionMap[participant.ID] = contradiction

				// Story 4-7 AC3: Apply suggested emotion delta from contradiction
				if emotionChanges[participant.ID].Trust == 0 && emotionChanges[participant.ID].Fear == 0 && emotionChanges[participant.ID].Stress == 0 {
					// Convert knowledge.EmotionDelta to manager.EmotionDelta
					emotionChanges[participant.ID] = manager.EmotionDelta{
						Trust:  contradiction.SuggestedDelta.Trust,
						Fear:   contradiction.SuggestedDelta.Fear,
						Stress: contradiction.SuggestedDelta.Stress,
					}
				} else {
					// Accumulate with existing emotion changes
					existing := emotionChanges[participant.ID]
					emotionChanges[participant.ID] = manager.EmotionDelta{
						Trust:  existing.Trust + contradiction.SuggestedDelta.Trust,
						Fear:   existing.Fear + contradiction.SuggestedDelta.Fear,
						Stress: existing.Stress + contradiction.SuggestedDelta.Stress,
					}
				}

				// Apply emotion change to NPC via NPCManager
				if cp.npcManager != nil {
					if err := cp.npcManager.AdjustEmotion(participant.ID, emotionChanges[participant.ID]); err != nil {
						logger.Warn("Failed to apply contradiction emotion change", map[string]interface{}{
							"npc_id": participant.ID,
							"error":  err.Error(),
						})
					}
				}

				logger.Debug("Contradiction detected", map[string]interface{}{
					"npc_id":   participant.ID,
					"type":     contradiction.Type,
					"severity": contradiction.Severity,
					"message":  message,
				})
			}
		}
	}

	// Process flags for each NPC participant
	for _, participant := range npcParticipants {
		// Build flag handler context
		flagContext := FlagHandlerContext{
			NPCID:         participant.ID,
			PlayerMessage: message,
			JudgeResult:   nil, // Story 4-2 doesn't use full JudgeResponseV2
			GameState:     nil, // Will be provided in future stories if needed
			UpdateManager: cp.updateManager,
			NPCManager:    cp.npcManager,
		}

		// Execute all flag handlers for this NPC
		handlerResult := cp.handlerRegistry.ExecuteAllHandlers(chatFlags, flagContext)

		// Collect emotion changes (accumulate with contradiction changes)
		if handlerResult.EmotionDelta.Trust != 0 || handlerResult.EmotionDelta.Fear != 0 || handlerResult.EmotionDelta.Stress != 0 {
			if existing, exists := emotionChanges[participant.ID]; exists {
				// Accumulate with existing contradiction-based changes
				emotionChanges[participant.ID] = manager.EmotionDelta{
					Trust:  existing.Trust + handlerResult.EmotionDelta.Trust,
					Fear:   existing.Fear + handlerResult.EmotionDelta.Fear,
					Stress: existing.Stress + handlerResult.EmotionDelta.Stress,
				}
			} else {
				emotionChanges[participant.ID] = handlerResult.EmotionDelta
			}
		}

		// Collect contradictions from handlers (may have additional contradictions)
		contradictions = append(contradictions, handlerResult.Contradictions...)

		if !handlerResult.Success {
			logger.Warn("Flag handler execution had errors", map[string]interface{}{
				"npc_id": participant.ID,
				"error":  handlerResult.Error,
			})
		}
	}

	// Step 5: Propagate player message as Fact to UpdateManager
	if err := cp.propagatePlayerMessage(session, message); err != nil {
		// AC: UpdateManager failure logs error but continues
		logger.Warn("Failed to propagate player message", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Step 7: Generate NPC responses (STUB for Story 4-4)
	// Story 4-7 AC4: Pass contradictions to response generation for questioning
	npcResponses := cp.generateNPCResponsesWithContradictions(session, message, emotionChanges, contradictionMap)

	// Step 8: Return complete ProcessResult
	result := &ProcessResult{
		NPCResponses:   npcResponses,
		EmotionChanges: emotionChanges,
		Flags:          chatFlags,
		Contradictions: contradictions,
		Success:        true,
		Error:          "",
	}

	logger.Debug("ChatProcessor.ProcessPlayerMessage completed", map[string]interface{}{
		"num_npc_responses": len(result.NPCResponses),
		"num_emotions":      len(result.EmotionChanges),
		"num_flags":         len(result.Flags),
		"num_contradictions": len(result.Contradictions),
	})

	return result, nil
}

// validateProcessRequest validates the input parameters for ProcessPlayerMessage.
//
// Checks:
// - Message is non-empty
// - Session is valid
// - At least one NPC participant exists
func (cp *ChatProcessor) validateProcessRequest(session *ChatSession, message string) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	// Must have at least one NPC participant
	hasNPC := false
	for _, p := range session.Participants {
		if !p.IsPlayer {
			hasNPC = true
			break
		}
	}

	if !hasNPC {
		return fmt.Errorf("session must have at least one NPC participant")
	}

	return nil
}

// buildJudgeChatRequest constructs a JudgeChatRequest with full context.
//
// Includes:
// - Player message
// - All participants (with emotion states)
// - Recent conversation history
// - Game state
// - Relevant facts from UpdateManager
func (cp *ChatProcessor) buildJudgeChatRequest(session *ChatSession, message string, gameState *agents.GameStateSnapshot) (*agents.JudgeChatRequest, error) {
	// Convert ChatParticipant to agents.ChatParticipant
	agentParticipants := make([]agents.ChatParticipant, len(session.Participants))
	for i, p := range session.Participants {
		agentParticipants[i] = agents.ChatParticipant{
			ID:       p.ID,
			Name:     p.Name,
			IsPlayer: p.IsPlayer,
			Emotion: agents.EmotionState{
				Trust:  p.Emotion.Trust,
				Fear:   p.Emotion.Fear,
				Stress: p.Emotion.Stress,
			},
			Relationship: p.Relationship,
		}
	}

	// Convert ChatMessage to agents.ChatMessage for conversation history
	agentHistory := make([]agents.ChatMessage, len(session.MessageHistory))
	for i, msg := range session.MessageHistory {
		agentHistory[i] = agents.ChatMessage{
			Speaker:   msg.SenderID,
			Content:   msg.Content,
			Timestamp: fmt.Sprintf("%d", msg.Timestamp),
		}
	}

	// Get relevant facts from UpdateManager
	relevantFacts := cp.getRelevantFacts(session)

	return &agents.JudgeChatRequest{
		PlayerMessage:       message,
		Participants:        agentParticipants,
		ConversationHistory: agentHistory,
		GameState:           gameState,
		RelevantFacts:       relevantFacts,
	}, nil
}

// propagatePlayerMessage registers the player's message as a Fact and propagates it.
//
// Story 4-2 AC4: Propagates information to UpdateManager
// Story 4-7 AC1: Player messages automatically propagated to NPCs in same room
//
// This method:
// 1. Creates a Fact for the player's message
// 2. Registers it in the global fact repository
// 3. Uses LearnFromDialogue to propagate to all NPCs in the same room
func (cp *ChatProcessor) propagatePlayerMessage(session *ChatSession, message string) error {
	if cp.updateManager == nil {
		// Graceful degradation - UpdateManager is optional
		logger.Debug("UpdateManager not available, skipping fact propagation", nil)
		return nil
	}

	// Create a fact for the player's message
	fact := &knowledge.Fact{
		ID:        uuid.New().String(),
		Content:   message,
		Type:      knowledge.Dialogue,
		Source:    "player",
		CreatedAt: time.Now(),
		Location:  session.Location,
		Witnesses: cp.getWitnessIDs(session),
	}

	// Register fact in global repository
	cp.updateManager.RegisterFact(fact)

	// Story 4-7 AC1: Automatically propagate to all NPCs in same room
	// Use LearnFromDialogue which handles room checking and confidence levels
	for _, participant := range session.Participants {
		if !participant.IsPlayer {
			// LearnFromDialogue will verify room membership and create knowledge
			cp.updateManager.LearnFromDialogue(participant.ID, "player", message, session.Location)

			logger.Debug("Player message propagated to NPC", map[string]interface{}{
				"npc_id":   participant.ID,
				"fact_id":  fact.ID,
				"content":  message,
				"location": session.Location,
			})
		}
	}

	logger.Debug("Player message fact created and propagated", map[string]interface{}{
		"fact_id":       fact.ID,
		"content":       fact.Content,
		"location":      fact.Location,
		"num_witnesses": len(fact.Witnesses),
	})

	return nil
}

// generateNPCResponses generates NPC responses to the player's message.
//
// STUB for Story 4-4: Returns placeholder responses
// Story 4-4 will implement actual LLM-based response generation
// Story 4-7 AC2: NPC responses are automatically propagated to other participants
//
// DEPRECATED: Use generateNPCResponsesWithContradictions instead
func (cp *ChatProcessor) generateNPCResponses(session *ChatSession, message string, emotionChanges map[string]manager.EmotionDelta) []NPCResponse {
	return cp.generateNPCResponsesWithContradictions(session, message, emotionChanges, nil)
}

// generateNPCResponsesWithContradictions generates NPC responses to the player's message,
// incorporating contradiction context if any contradictions were detected.
//
// Story 4-4: STUB for actual LLM-based response generation
// Story 4-7 AC2: NPC responses are automatically propagated to other participants
// Story 4-7 AC4: When contradiction detected, NPC response includes questioning/doubt
//
// For each NPC:
// 1. Gets current emotion state and applies changes
// 2. Checks if this NPC has a contradiction with the player's message
// 3. If contradiction exists, enhances response with questioning based on severity
// 4. Generates response (placeholder for now, Story 4-4 will implement LLM)
// 5. Propagates response to other participants in the room
//
// Parameters:
//   - contradictionMap: Map from NPC ID to their contradiction (already computed in ProcessPlayerMessage)
func (cp *ChatProcessor) generateNPCResponsesWithContradictions(
	session *ChatSession,
	message string,
	emotionChanges map[string]manager.EmotionDelta,
	contradictionMap map[string]*knowledge.ContradictionResult,
) []NPCResponse {
	responses := []NPCResponse{}

	// Handle nil contradictionMap
	if contradictionMap == nil {
		contradictionMap = make(map[string]*knowledge.ContradictionResult)
	}

	for _, participant := range session.Participants {
		if participant.IsPlayer {
			continue
		}

		// Get current emotion state
		currentEmotion := participant.Emotion

		// Apply emotion changes if any
		if delta, exists := emotionChanges[participant.ID]; exists {
			currentEmotion = currentEmotion.Apply(delta)
		}

		// Story 4-7 AC4: Check if this NPC has a contradiction
		contradiction := contradictionMap[participant.ID]
		responseContent := "Processing..."
		responseFlags := []ChatFlag{}

		if contradiction != nil {
			// Enhance response based on contradiction severity
			responseContent = cp.generateContradictionResponse(contradiction, participant.Name)
			responseFlags = append(responseFlags, ChatFlagContradiction)

			logger.Debug("Generating contradiction response for NPC", map[string]interface{}{
				"npc_id":   participant.ID,
				"npc_name": participant.Name,
				"type":     contradiction.Type,
				"severity": contradiction.Severity,
			})
		}

		// STUB: Create response
		// Story 4-4 will implement actual LLM-based response generation
		response := NPCResponse{
			NPCID:        participant.ID,
			Content:      responseContent,
			Emotion:      currentEmotion,
			Flags:        responseFlags,
			UsedFallback: false,
		}

		responses = append(responses, response)

		// Story 4-7 AC2: Propagate NPC response to other participants in same room
		cp.propagateNPCResponse(session, participant.ID, response.Content)
	}

	return responses
}

// generateContradictionResponse generates a contextual response based on contradiction severity.
//
// Story 4-7 AC4: When contradiction detected, NPC response includes questioning/doubt
//
// Response templates by severity:
// - Major (8-10): Strong questioning, accusation of lying or insanity
// - Moderate (5-7): Confused, requesting explanation
// - Minor (1-4): Slight doubt, may overlook
//
// Parameters:
//   - contradiction: The detected contradiction result
//   - npcName: The name of the NPC for personalized response
//
// Returns:
//   - String response incorporating the contradiction context
func (cp *ChatProcessor) generateContradictionResponse(contradiction *knowledge.ContradictionResult, npcName string) string {
	// Story 4-7 AC4: Generate response based on contradiction severity and suggested reaction
	switch contradiction.Type {
	case knowledge.ContradictionMajor:
		// Major contradiction: Strong doubt, potential accusation
		return fmt.Sprintf("[%s 強烈質疑] 等等，你剛才說的和我知道的完全不一樣！%s", npcName, contradiction.SuggestedReaction)

	case knowledge.ContradictionModerate:
		// Moderate contradiction: Confusion, request explanation
		return fmt.Sprintf("[%s 感到困惑] 這和我之前聽說的不太一樣...你能解釋一下嗎？", npcName)

	case knowledge.ContradictionMinor:
		// Minor contradiction: Slight doubt but may overlook
		return fmt.Sprintf("[%s 略微疑惑] 嗯...我記得好像不是這樣的，不過也許我記錯了。", npcName)

	default:
		// Fallback
		return fmt.Sprintf("[%s] 你說的和我知道的有些出入...", npcName)
	}
}

// propagateNPCResponse propagates an NPC's response to all other entities in the same room.
//
// Story 4-7 AC2: NPC responses automatically propagated to other participants
//
// This method:
// 1. Creates a Fact for the NPC's response
// 2. Registers it in the global fact repository
// 3. Uses LearnFromDialogue to propagate to all other entities (NPCs + player) in the same room
func (cp *ChatProcessor) propagateNPCResponse(session *ChatSession, speakerID string, content string) error {
	if cp.updateManager == nil {
		// Graceful degradation - UpdateManager is optional
		logger.Debug("UpdateManager not available, skipping NPC response propagation", nil)
		return nil
	}

	// Create a fact for the NPC's response
	fact := &knowledge.Fact{
		ID:        uuid.New().String(),
		Content:   content,
		Type:      knowledge.Dialogue,
		Source:    speakerID,
		CreatedAt: time.Now(),
		Location:  session.Location,
		Witnesses: cp.getWitnessIDs(session),
	}

	// Register fact in global repository
	cp.updateManager.RegisterFact(fact)

	// Story 4-7 AC2: Automatically propagate to all entities in same room (excluding speaker)
	for _, participant := range session.Participants {
		// Skip the speaker themselves
		if participant.ID == speakerID {
			continue
		}

		// Propagate to both NPCs and player
		listenerID := participant.ID
		if participant.IsPlayer {
			listenerID = "player"
		}

		// LearnFromDialogue will verify room membership and create knowledge
		cp.updateManager.LearnFromDialogue(listenerID, speakerID, content, session.Location)

		logger.Debug("NPC response propagated to listener", map[string]interface{}{
			"speaker_id":  speakerID,
			"listener_id": listenerID,
			"fact_id":     fact.ID,
			"content":     content,
			"location":    session.Location,
		})
	}

	logger.Debug("NPC response fact created and propagated", map[string]interface{}{
		"fact_id":       fact.ID,
		"speaker_id":    speakerID,
		"content":       fact.Content,
		"location":      fact.Location,
		"num_witnesses": len(fact.Witnesses),
	})

	return nil
}

// Helper methods

// convertToChatFlags converts agents.ChatFlag to chat.ChatFlag
func (cp *ChatProcessor) convertToChatFlags(agentFlags []agents.ChatFlag) []ChatFlag {
	chatFlags := make([]ChatFlag, 0, len(agentFlags))
	for _, af := range agentFlags {
		switch af {
		case agents.FlagHallucination:
			chatFlags = append(chatFlags, ChatFlagHallucination)
		case agents.FlagHostile:
			chatFlags = append(chatFlags, ChatFlagHostile)
		case agents.FlagRevelation:
			chatFlags = append(chatFlags, ChatFlagRevelation)
		case agents.FlagContradiction:
			chatFlags = append(chatFlags, ChatFlagContradiction)
		case agents.FlagPersuasion:
			chatFlags = append(chatFlags, ChatFlagPersuasion)
		case agents.FlagLie:
			chatFlags = append(chatFlags, ChatFlagLie)
		}
	}
	return chatFlags
}

// flagsToStrings converts ChatFlags to string slice for logging
func (cp *ChatProcessor) flagsToStrings(flags []ChatFlag) []string {
	strs := make([]string, len(flags))
	for i, f := range flags {
		strs[i] = f.String()
	}
	return strs
}

// getNPCParticipants returns all NPC participants from the session
func (cp *ChatProcessor) getNPCParticipants(session *ChatSession) []ChatParticipant {
	npcs := []ChatParticipant{}
	for _, p := range session.Participants {
		if !p.IsPlayer {
			npcs = append(npcs, p)
		}
	}
	return npcs
}

// getWitnessIDs returns all participant IDs as witnesses
func (cp *ChatProcessor) getWitnessIDs(session *ChatSession) []string {
	ids := make([]string, len(session.Participants))
	for i, p := range session.Participants {
		ids[i] = p.ID
	}
	return ids
}

// getRelevantFacts retrieves relevant facts from UpdateManager for the session
func (cp *ChatProcessor) getRelevantFacts(session *ChatSession) []string {
	// For now, return empty slice
	// Story 4-2 doesn't require sophisticated fact retrieval
	// Future stories will implement context-aware fact selection
	return []string{}
}

// ==========================================================================
// Story 5-8: Performance Monitoring & Caching API
// ==========================================================================

// GetMetrics returns the current performance metrics.
//
// Story 5-8 AC1 & AC2: Provide metrics API for monitoring
//
// Returns nil if metrics are disabled.
func (cp *ChatProcessor) GetMetrics() *MetricsSnapshot {
	if cp.metrics == nil {
		return nil
	}

	snapshot := cp.metrics.GetMetrics()
	return &snapshot
}

// RecordTokenUsage records token usage for an LLM call.
//
// Story 5-8 AC2: Track token consumption
//
// This should be called after each LLM API call to track token usage.
func (cp *ChatProcessor) RecordTokenUsage(npcID string, inputTokens, outputTokens int) {
	if cp.metrics != nil {
		cp.metrics.RecordTokenUsage(npcID, inputTokens, outputTokens)
	}
}

// ClearCache clears all cached responses.
//
// Story 5-8 AC3: Cache management
//
// This is useful when NPCs undergo significant state changes.
func (cp *ChatProcessor) ClearCache() {
	if cp.responseCache != nil {
		cp.responseCache.Clear()
		logger.Debug("ChatProcessor cache cleared", nil)
	}
}

// InvalidateNPCCache clears all cached responses for a specific NPC.
//
// Story 5-8 AC3: Invalidate cache on significant NPC state changes
//
// Call this when an NPC's state changes significantly (e.g., major emotion change).
func (cp *ChatProcessor) InvalidateNPCCache(npcID string) {
	if cp.responseCache != nil {
		cp.responseCache.InvalidateNPC(npcID)
		logger.Debug("ChatProcessor NPC cache invalidated", map[string]interface{}{
			"npc_id": npcID,
		})
	}
}
