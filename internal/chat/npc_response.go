package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 4.4: NPC Response Generation System
// ==========================================================================

// NPCResponseGenerator generates NPC responses using LLM-based dialogue generation.
// This is the core implementation for Story 4.4.
//
// Story 4.4 AC1: generateNPCResponse() 使用 LLM 生成 NPC 回應
// Story 4.4 AC2: Prompt 包含 NPC 完整檔案（BuildFullNPCPrompt）
// Story 4.4 AC3: Prompt 包含近期對話歷史（最近 5-10 條）
// Story 4.4 AC4: Prompt 包含判定 Flags（影響回應態度）
// Story 4.4 AC5: 回應包含適當的情感反應
// Story 4.4 AC6: 使用 Reactive/Rapid 模型（< 2 秒延遲）
type NPCResponseGenerator struct {
	npcManager      *manager.NPCManager
	llmClient       agents.LLMClient
	fallbackManager *FallbackManager
	config          *ResponseGeneratorConfig
}

// ResponseGeneratorConfig contains configuration for NPC response generation.
type ResponseGeneratorConfig struct {
	// Model is the LLM model to use (default: "gpt-4o-mini" for rapid responses)
	Model string

	// MaxTokens limits the response length (default: 150)
	MaxTokens int

	// Temperature controls response creativity (default: 0.7)
	Temperature float64

	// Timeout for LLM calls (default: 2s for reactive responses)
	Timeout time.Duration

	// MaxHistoryMessages is the number of recent messages to include in prompt (default: 10)
	MaxHistoryMessages int

	// EnableFallback allows fallback to template-based responses on LLM failure
	EnableFallback bool
}

// DefaultResponseGeneratorConfig returns the default configuration.
//
// Story 4.4 AC6: 使用 Reactive/Rapid 模型（< 2 秒延遲）
func DefaultResponseGeneratorConfig() *ResponseGeneratorConfig {
	return &ResponseGeneratorConfig{
		Model:              "gpt-4o-mini", // Fast model for reactive responses
		MaxTokens:          150,            // Short responses for chat
		Temperature:        0.7,            // Balanced creativity
		Timeout:            2 * time.Second, // < 2s latency requirement
		MaxHistoryMessages: 10,             // Recent context (5-10 messages)
		EnableFallback:     true,           // Graceful degradation
	}
}

// NewNPCResponseGenerator creates a new NPCResponseGenerator.
func NewNPCResponseGenerator(
	npcManager *manager.NPCManager,
	llmClient agents.LLMClient,
	fallbackManager *FallbackManager,
	config *ResponseGeneratorConfig,
) *NPCResponseGenerator {
	if config == nil {
		config = DefaultResponseGeneratorConfig()
	}

	return &NPCResponseGenerator{
		npcManager:      npcManager,
		llmClient:       llmClient,
		fallbackManager: fallbackManager,
		config:          config,
	}
}

// GenerateNPCResponse generates an NPC's response to the player's message.
//
// Story 4.4 AC1: generateNPCResponse() 使用 LLM 生成 NPC 回應
// Story 4.4 AC2: Prompt 包含 NPC 完整檔案（BuildFullNPCPrompt）
// Story 4.4 AC3: Prompt 包含近期對話歷史（最近 5-10 條）
// Story 4.4 AC4: Prompt 包含判定 Flags（影響回應態度）
// Story 4.4 AC5: 回應包含適當的情感反應
//
// This is the main entry point for NPC response generation in the chat system.
func (g *NPCResponseGenerator) GenerateNPCResponse(
	ctx context.Context,
	npcID string,
	playerMessage string,
	conversationHistory []ChatMessage,
	flags []ChatFlag,
	currentEmotion manager.EmotionState,
) (NPCResponse, error) {
	logger.Debug("GenerateNPCResponse invoked", map[string]interface{}{
		"npc_id":         npcID,
		"player_message": playerMessage,
		"num_history":    len(conversationHistory),
		"num_flags":      len(flags),
	})

	// Validate inputs
	if npcID == "" {
		return NPCResponse{}, fmt.Errorf("npcID cannot be empty")
	}
	if playerMessage == "" {
		return NPCResponse{}, fmt.Errorf("playerMessage cannot be empty")
	}

	// Get NPC profile and state
	profile := g.npcManager.GetProfile(npcID)
	if profile == nil {
		return NPCResponse{}, fmt.Errorf("NPC not found: %s", npcID)
	}

	state := g.npcManager.GetState(npcID)
	if state == nil {
		return NPCResponse{}, fmt.Errorf("NPC state not found: %s", npcID)
	}

	// Apply timeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, g.config.Timeout)
	defer cancel()

	// Build prompt with full NPC context
	prompt := g.buildResponsePrompt(npcID, playerMessage, conversationHistory, flags, currentEmotion)

	// Call LLM for response generation
	response, err := g.callLLMForResponse(ctxWithTimeout, prompt)
	if err != nil {
		logger.Warn("LLM call failed for NPC response generation", map[string]interface{}{
			"npc_id": npcID,
			"error":  err.Error(),
		})

		// Fallback to template-based response if enabled
		if g.config.EnableFallback && g.fallbackManager != nil {
			return g.generateFallbackResponse(npcID, profile, state, flags, currentEmotion), nil
		}

		return NPCResponse{}, fmt.Errorf("LLM call failed and fallback disabled: %w", err)
	}

	// Parse LLM response
	npcResponse := NPCResponse{
		NPCID:        npcID,
		Content:      response,
		Emotion:      currentEmotion,
		Flags:        flags,
		UsedFallback: false,
	}

	logger.Debug("GenerateNPCResponse completed", map[string]interface{}{
		"npc_id":        npcID,
		"content_len":   len(response),
		"used_fallback": false,
	})

	return npcResponse, nil
}

// buildResponsePrompt constructs the complete LLM prompt for NPC response generation.
//
// Story 4.4 AC2: Prompt 包含 NPC 完整檔案（BuildFullNPCPrompt）
// Story 4.4 AC3: Prompt 包含近期對話歷史（最近 5-10 條）
// Story 4.4 AC4: Prompt 包含判定 Flags（影響回應態度）
// Story 8.7: Multi-language support - prompts adapt to current language
func (g *NPCResponseGenerator) buildResponsePrompt(
	npcID string,
	playerMessage string,
	conversationHistory []ChatMessage,
	flags []ChatFlag,
	currentEmotion manager.EmotionState,
) string {
	var sb strings.Builder

	// Get current translator for language-aware prompts
	translator := i18n.GetGlobal()
	currentLocale := "zh-TW" // Default fallback
	if translator != nil {
		currentLocale = translator.GetLocale()
	}

	// Section 1: System Instructions (language-dependent)
	systemInstructions := g.getSystemInstructions(currentLocale)
	sb.WriteString(systemInstructions)
	sb.WriteString("\n\n")

	// Section 2: Full NPC Profile (Story 4.4 AC2)
	// Uses BuildFullNPCPrompt which includes knowledge base, emotion state, traits, etc.
	npcPrompt := g.npcManager.BuildFullNPCPrompt(npcID)
	sb.WriteString(npcPrompt)
	sb.WriteString("\n")

	// Section 3: Current Emotional State (Story 4.4 AC5)
	// This reinforces the emotion information already in BuildFullNPCPrompt
	sb.WriteString(g.getEmotionHeader(currentLocale))
	sb.WriteString("\n\n")
	sb.WriteString(g.formatEmotionState(currentLocale, currentEmotion))
	sb.WriteString("\n\n")

	// Section 4: Chat Flags (Story 4.4 AC4)
	// Flags affect the NPC's response attitude
	if len(flags) > 0 {
		sb.WriteString(g.getFlagsHeader(currentLocale))
		sb.WriteString("\n\n")
		sb.WriteString(g.getFlagsIntro(currentLocale))
		sb.WriteString("\n")
		for _, flag := range flags {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", flag, g.describeChatFlag(flag)))
		}
		sb.WriteString("\n")
	}

	// Section 5: Recent Conversation History (Story 4.4 AC3)
	// Include last 5-10 messages for context
	if len(conversationHistory) > 0 {
		sb.WriteString(g.getHistoryHeader(currentLocale))
		sb.WriteString("\n\n")

		// Limit to MaxHistoryMessages
		startIdx := 0
		if len(conversationHistory) > g.config.MaxHistoryMessages {
			startIdx = len(conversationHistory) - g.config.MaxHistoryMessages
		}

		playerLabel := g.getPlayerLabel(currentLocale)
		for i := startIdx; i < len(conversationHistory); i++ {
			msg := conversationHistory[i]
			speaker := msg.SenderID
			if speaker == "player" {
				speaker = playerLabel
			}
			sb.WriteString(fmt.Sprintf("- %s: %s\n", speaker, msg.Content))
		}
		sb.WriteString("\n")
	}

	// Section 6: Player's Current Message
	sb.WriteString(g.getPlayerMessageHeader(currentLocale))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", playerMessage))

	// Section 7: Response Instructions
	sb.WriteString(g.getResponseInstructions(currentLocale))

	return sb.String()
}

// ==========================================================================
// Story 8.7: Language-aware prompt helpers
// ==========================================================================

// getSystemInstructions returns system instructions in the appropriate language
func (g *NPCResponseGenerator) getSystemInstructions(locale string) string {
	switch locale {
	case "zh-CN":
		return "你是一个互动式恐怖游戏中的 NPC。请根据以下角色设定与对话脉络，生成符合角色情感与个性的回应。"
	case "en-US":
		return "You are an NPC in an interactive horror game. Generate responses that match the character's emotions and personality based on the following character settings and dialogue context."
	default: // zh-TW
		return "你是一個互動式恐怖遊戲中的 NPC。請根據以下角色設定與對話脈絡，生成符合角色情感與個性的回應。"
	}
}

// getEmotionHeader returns the emotion section header in the appropriate language
func (g *NPCResponseGenerator) getEmotionHeader(locale string) string {
	switch locale {
	case "zh-CN":
		return "## 当前情绪状态"
	case "en-US":
		return "## Current Emotional State"
	default: // zh-TW
		return "## 當前情緒狀態"
	}
}

// formatEmotionState formats the emotion state in the appropriate language
func (g *NPCResponseGenerator) formatEmotionState(locale string, emotion manager.EmotionState) string {
	switch locale {
	case "zh-CN":
		return fmt.Sprintf("- 信任度: %d/100\n- 恐惧度: %d/100\n- 压力值: %d/100",
			emotion.Trust, emotion.Fear, emotion.Stress)
	case "en-US":
		return fmt.Sprintf("- Trust: %d/100\n- Fear: %d/100\n- Stress: %d/100",
			emotion.Trust, emotion.Fear, emotion.Stress)
	default: // zh-TW
		return fmt.Sprintf("- 信任度: %d/100\n- 恐懼度: %d/100\n- 壓力值: %d/100",
			emotion.Trust, emotion.Fear, emotion.Stress)
	}
}

// getFlagsHeader returns the flags section header in the appropriate language
func (g *NPCResponseGenerator) getFlagsHeader(locale string) string {
	switch locale {
	case "zh-CN":
		return "## 玩家消息特征（判定 Flags）"
	case "en-US":
		return "## Player Message Characteristics (Flags)"
	default: // zh-TW
		return "## 玩家訊息特徵（判定 Flags）"
	}
}

// getFlagsIntro returns the flags introduction text in the appropriate language
func (g *NPCResponseGenerator) getFlagsIntro(locale string) string {
	switch locale {
	case "zh-CN":
		return "玩家的消息包含以下特征，请在回应中反映出来："
	case "en-US":
		return "The player's message contains the following characteristics, please reflect them in your response:"
	default: // zh-TW
		return "玩家的訊息包含以下特徵，請在回應中反映出來："
	}
}

// getHistoryHeader returns the conversation history header in the appropriate language
func (g *NPCResponseGenerator) getHistoryHeader(locale string) string {
	switch locale {
	case "zh-CN":
		return "## 近期对话历史"
	case "en-US":
		return "## Recent Conversation History"
	default: // zh-TW
		return "## 近期對話歷史"
	}
}

// getPlayerLabel returns the player label in the appropriate language
func (g *NPCResponseGenerator) getPlayerLabel(locale string) string {
	switch locale {
	case "zh-CN":
		return "玩家"
	case "en-US":
		return "Player"
	default: // zh-TW
		return "玩家"
	}
}

// getPlayerMessageHeader returns the player message header in the appropriate language
func (g *NPCResponseGenerator) getPlayerMessageHeader(locale string) string {
	switch locale {
	case "zh-CN":
		return "## 玩家的消息"
	case "en-US":
		return "## Player's Message"
	default: // zh-TW
		return "## 玩家的訊息"
	}
}

// getResponseInstructions returns the response instructions in the appropriate language
func (g *NPCResponseGenerator) getResponseInstructions(locale string) string {
	switch locale {
	case "zh-CN":
		return `## 回应指示

请以这个角色的身份，根据上述信息生成一段简短的对话回应（1-3 句话）。
回应应该：
1. 符合角色的个性、情绪与对话风格
2. 反映当前的情感状态（信任、恐惧、压力）
3. 考虑对话历史的脉络
4. 对玩家消息的特征（Flags）作出适当反应
5. 只知道「已知信息」区块中列出的事实，对未知事项表现困惑

请直接输出角色的回应内容，不要加上「角色：」或其他前缀。
`
	case "en-US":
		return `## Response Instructions

As this character, generate a brief dialogue response (1-3 sentences) based on the above information.
The response should:
1. Match the character's personality, emotions, and dialogue style
2. Reflect the current emotional state (trust, fear, stress)
3. Consider the context of the conversation history
4. Respond appropriately to the player message characteristics (Flags)
5. Only know facts listed in the "Known Information" section, show confusion about unknown matters

Output the character's response directly without adding "Character:" or other prefixes.
`
	default: // zh-TW
		return `## 回應指示

請以這個角色的身份，根據上述資訊生成一段簡短的對話回應（1-3 句話）。
回應應該：
1. 符合角色的個性、情緒與對話風格
2. 反映當前的情感狀態（信任、恐懼、壓力）
3. 考慮對話歷史的脈絡
4. 對玩家訊息的特徵（Flags）作出適當反應
5. 只知道「已知資訊」區塊中列出的事實，對未知事項表現困惑

請直接輸出角色的回應內容，不要加上「角色：」或其他前綴。
`
	}
}

// describeChatFlag returns a human-readable description of a chat flag for the prompt.
//
// Story 4.4 AC4: Prompt 包含判定 Flags（影響回應態度）
// Story 8.7: Multi-language support for flag descriptions
func (g *NPCResponseGenerator) describeChatFlag(flag ChatFlag) string {
	// Get current locale
	translator := i18n.GetGlobal()
	currentLocale := "zh-TW" // Default fallback
	if translator != nil {
		currentLocale = translator.GetLocale()
	}

	// Return language-specific description
	switch currentLocale {
	case "zh-CN":
		return g.describeChatFlagZhCN(flag)
	case "en-US":
		return g.describeChatFlagEnUS(flag)
	default: // zh-TW
		return g.describeChatFlagZhTW(flag)
	}
}

// describeChatFlagZhTW returns Traditional Chinese descriptions
func (g *NPCResponseGenerator) describeChatFlagZhTW(flag ChatFlag) string {
	switch flag {
	case ChatFlagHallucination:
		return "玩家聲稱的事情與已知事實不符（可能是幻覺或記錯）"
	case ChatFlagHostile:
		return "玩家使用威脅性或攻擊性語言"
	case ChatFlagRevelation:
		return "玩家分享重要資訊或秘密"
	case ChatFlagPersuasion:
		return "玩家試圖說服或操縱你"
	case ChatFlagLie:
		return "玩家在說謊（刻意隱瞞或扭曲事實）"
	case ChatFlagContradiction:
		return "玩家的陳述與你已知的資訊矛盾"
	default:
		return string(flag)
	}
}

// describeChatFlagZhCN returns Simplified Chinese descriptions
func (g *NPCResponseGenerator) describeChatFlagZhCN(flag ChatFlag) string {
	switch flag {
	case ChatFlagHallucination:
		return "玩家声称的事情与已知事实不符（可能是幻觉或记错）"
	case ChatFlagHostile:
		return "玩家使用威胁性或攻击性语言"
	case ChatFlagRevelation:
		return "玩家分享重要信息或秘密"
	case ChatFlagPersuasion:
		return "玩家试图说服或操纵你"
	case ChatFlagLie:
		return "玩家在说谎（刻意隐瞒或扭曲事实）"
	case ChatFlagContradiction:
		return "玩家的陈述与你已知的信息矛盾"
	default:
		return string(flag)
	}
}

// describeChatFlagEnUS returns English descriptions
func (g *NPCResponseGenerator) describeChatFlagEnUS(flag ChatFlag) string {
	switch flag {
	case ChatFlagHallucination:
		return "Player's claim doesn't match known facts (possible hallucination or misremembering)"
	case ChatFlagHostile:
		return "Player uses threatening or aggressive language"
	case ChatFlagRevelation:
		return "Player shares important information or secrets"
	case ChatFlagPersuasion:
		return "Player attempts to persuade or manipulate you"
	case ChatFlagLie:
		return "Player is lying (deliberately concealing or distorting facts)"
	case ChatFlagContradiction:
		return "Player's statement contradicts your known information"
	default:
		return string(flag)
	}
}

// callLLMForResponse calls the LLM client to generate the NPC response.
//
// Story 4.4 AC1: generateNPCResponse() 使用 LLM 生成 NPC 回應
// Story 4.4 AC6: 使用 Reactive/Rapid 模型（< 2 秒延遲）
func (g *NPCResponseGenerator) callLLMForResponse(ctx context.Context, prompt string) (string, error) {
	// Build LLM options
	options := map[string]any{
		"model":       g.config.Model,
		"max_tokens":  g.config.MaxTokens,
		"temperature": g.config.Temperature,
	}

	// Call LLM with timeout context
	response, err := g.llmClient.Generate(ctx, prompt, options)
	if err != nil {
		return "", fmt.Errorf("LLM generate failed: %w", err)
	}

	// Trim whitespace from response
	response = strings.TrimSpace(response)

	if response == "" {
		return "", fmt.Errorf("LLM returned empty response")
	}

	return response, nil
}

// generateFallbackResponse generates a fallback response using templates.
//
// This is used when LLM generation fails and EnableFallback is true.
// Implements graceful degradation pattern.
func (g *NPCResponseGenerator) generateFallbackResponse(
	npcID string,
	profile *manager.NPCProfile,
	state *manager.NPCRuntimeState,
	flags []ChatFlag,
	currentEmotion manager.EmotionState,
) NPCResponse {
	logger.Debug("Using fallback template for NPC response", map[string]interface{}{
		"npc_id": npcID,
	})

	// Build fallback context
	flagStrings := make([]string, len(flags))
	for i, f := range flags {
		flagStrings[i] = f.String()
	}

	fallbackContext := BuildFallbackContext(
		npcID,
		profile.Name,
		"玩家", // Default player name
		profile.Archetype,
		currentEmotion,
		state.MentalState,
		flagStrings,
	)

	// Select template
	content := g.fallbackManager.SelectTemplate(fallbackContext)

	return NPCResponse{
		NPCID:        npcID,
		Content:      content,
		Emotion:      currentEmotion,
		Flags:        flags,
		UsedFallback: true,
	}
}

// ==========================================================================
// Helper Types for LLM Response Parsing
// ==========================================================================

// LLMNPCResponseRaw is used for parsing structured LLM responses (if needed).
// Currently we use direct text responses, but this allows future enhancement
// to support JSON-formatted responses with emotion metadata.
type LLMNPCResponseRaw struct {
	Content  string                  `json:"content"`
	Emotion  *manager.EmotionState   `json:"emotion,omitempty"`
	Metadata map[string]interface{}  `json:"metadata,omitempty"`
}

// ParseLLMResponse parses a structured JSON response from the LLM.
// This is a utility function for future use if we switch to JSON responses.
func ParseLLMResponse(rawResponse string) (*LLMNPCResponseRaw, error) {
	var parsed LLMNPCResponseRaw
	if err := json.Unmarshal([]byte(rawResponse), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response JSON: %w", err)
	}

	if parsed.Content == "" {
		return nil, fmt.Errorf("LLM response missing content field")
	}

	return &parsed, nil
}

// ==========================================================================
// Integration with ChatProcessor
// ==========================================================================

// GenerateAllNPCResponses generates responses for all NPCs in the session.
//
// This is a convenience method that can be used by ChatProcessor to generate
// responses for all participating NPCs in one call.
func (g *NPCResponseGenerator) GenerateAllNPCResponses(
	ctx context.Context,
	session *ChatSession,
	playerMessage string,
	emotionChanges map[string]manager.EmotionDelta,
	flags []ChatFlag,
) ([]NPCResponse, error) {
	responses := make([]NPCResponse, 0)

	for _, participant := range session.Participants {
		if participant.IsPlayer {
			continue
		}

		// Get current emotion (apply changes if any)
		currentEmotion := participant.Emotion
		if delta, exists := emotionChanges[participant.ID]; exists {
			currentEmotion = currentEmotion.Apply(delta)
		}

		// Generate response for this NPC
		response, err := g.GenerateNPCResponse(
			ctx,
			participant.ID,
			playerMessage,
			session.MessageHistory,
			flags,
			currentEmotion,
		)

		if err != nil {
			logger.Warn("Failed to generate response for NPC", map[string]interface{}{
				"npc_id": participant.ID,
				"error":  err.Error(),
			})

			// Continue with other NPCs even if one fails
			continue
		}

		responses = append(responses, response)
	}

	return responses, nil
}
