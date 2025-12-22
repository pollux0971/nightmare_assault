package views

import "github.com/nightmare-assault/nightmare-assault/internal/trinity"

// ProviderPreset defines a preset configuration for a provider
type ProviderPreset struct {
	ID            string  // Provider ID (e.g., "anthropic", "openai")
	DisplayName   string  // Display name (e.g., "Anthropic (Claude)")
	DefaultModel  string  // Default model for this provider
	DefaultTokens int     // Default max tokens
	DefaultTemp   float64 // Default temperature
	BaseURL       string  // Base URL (empty for official APIs)
}

// tierProviderPresets maps each tier level to its available provider presets
var tierProviderPresets = map[trinity.TierLevel][]ProviderPreset{
	trinity.TierThinking: thinkingTierPresets,
	trinity.TierReactive: reactiveTierPresets,
	trinity.TierRapid:    rapidTierPresets,
}

// thinkingTierPresets defines provider presets for Thinking Tier (高品質推理)
// Optimized for deep reasoning tasks with higher token limits and lower temperature
var thinkingTierPresets = []ProviderPreset{
	{
		ID:            "anthropic",
		DisplayName:   "Anthropic (Claude Opus 4)",
		DefaultModel:  "claude-opus-4-20250514",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "",
	},
	{
		ID:            "openai",
		DisplayName:   "OpenAI (o1-preview)",
		DefaultModel:  "o1-preview",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "",
	},
	{
		ID:            "google",
		DisplayName:   "Google (Gemini Pro)",
		DefaultModel:  "gemini-2.0-flash-thinking-exp",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "",
	},
	{
		ID:            "mistral",
		DisplayName:   "Mistral (Mistral Large)",
		DefaultModel:  "mistral-large-latest",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "",
	},
	{
		ID:            "openrouter",
		DisplayName:   "OpenRouter (Claude Opus)",
		DefaultModel:  "anthropic/claude-3-opus",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "https://openrouter.ai/api/v1",
	},
	{
		ID:            "deepseek",
		DisplayName:   "DeepSeek (V3)",
		DefaultModel:  "deepseek-chat",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "https://api.deepseek.com/v1",
	},
	{
		ID:            "ollama",
		DisplayName:   "Ollama (Qwen 2.5 72B)",
		DefaultModel:  "qwen2.5:72b",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "http://localhost:11434/v1",
	},
	{
		ID:            "custom",
		DisplayName:   "自訂 (Custom)",
		DefaultModel:  "",
		DefaultTokens: 16000,
		DefaultTemp:   0.4,
		BaseURL:       "",
	},
}

// reactiveTierPresets defines provider presets for Reactive Tier (平衡互動)
// Optimized for balanced interactions with medium token limits and moderate temperature
var reactiveTierPresets = []ProviderPreset{
	{
		ID:            "anthropic",
		DisplayName:   "Anthropic (Claude 3.5 Sonnet)",
		DefaultModel:  "claude-3-5-sonnet-20241022",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "",
	},
	{
		ID:            "openai",
		DisplayName:   "OpenAI (GPT-4o)",
		DefaultModel:  "gpt-4o",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "",
	},
	{
		ID:            "google",
		DisplayName:   "Google (Gemini Pro)",
		DefaultModel:  "gemini-1.5-pro",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "",
	},
	{
		ID:            "mistral",
		DisplayName:   "Mistral (Mistral Medium)",
		DefaultModel:  "mistral-medium-latest",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "",
	},
	{
		ID:            "openrouter",
		DisplayName:   "OpenRouter (Claude Sonnet)",
		DefaultModel:  "anthropic/claude-3.5-sonnet",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "https://openrouter.ai/api/v1",
	},
	{
		ID:            "deepseek",
		DisplayName:   "DeepSeek (Chat)",
		DefaultModel:  "deepseek-chat",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "https://api.deepseek.com/v1",
	},
	{
		ID:            "cohere",
		DisplayName:   "Cohere (Command R+)",
		DefaultModel:  "command-r-plus",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "",
	},
	{
		ID:            "zhipu",
		DisplayName:   "智譜 (GLM-4)",
		DefaultModel:  "glm-4",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "https://open.bigmodel.cn/api/paas/v4",
	},
	{
		ID:            "moonshot",
		DisplayName:   "Moonshot (Kimi)",
		DefaultModel:  "moonshot-v1-8k",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "https://api.moonshot.cn/v1",
	},
	{
		ID:            "ollama",
		DisplayName:   "Ollama (Llama 3.1 70B)",
		DefaultModel:  "llama3.1:70b",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "http://localhost:11434/v1",
	},
	{
		ID:            "custom",
		DisplayName:   "自訂 (Custom)",
		DefaultModel:  "",
		DefaultTokens: 8000,
		DefaultTemp:   0.7,
		BaseURL:       "",
	},
}

// rapidTierPresets defines provider presets for Rapid Tier (快速回應)
// Optimized for fast responses with lower token limits and higher temperature
var rapidTierPresets = []ProviderPreset{
	{
		ID:            "anthropic",
		DisplayName:   "Anthropic (Claude Haiku)",
		DefaultModel:  "claude-3-haiku-20240307",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "",
	},
	{
		ID:            "openai",
		DisplayName:   "OpenAI (GPT-4o-mini)",
		DefaultModel:  "gpt-4o-mini",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "",
	},
	{
		ID:            "google",
		DisplayName:   "Google (Gemini Flash)",
		DefaultModel:  "gemini-1.5-flash",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "",
	},
	{
		ID:            "groq",
		DisplayName:   "Groq (Llama 3.1 70B)",
		DefaultModel:  "llama-3.1-70b-versatile",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "https://api.groq.com/openai/v1",
	},
	{
		ID:            "openrouter",
		DisplayName:   "OpenRouter (GPT-4o-mini)",
		DefaultModel:  "openai/gpt-4o-mini",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "https://openrouter.ai/api/v1",
	},
	{
		ID:            "deepseek",
		DisplayName:   "DeepSeek (Chat)",
		DefaultModel:  "deepseek-chat",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "https://api.deepseek.com/v1",
	},
	{
		ID:            "ollama",
		DisplayName:   "Ollama (Llama 3.1)",
		DefaultModel:  "llama3.1",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "http://localhost:11434/v1",
	},
	{
		ID:            "lmstudio",
		DisplayName:   "LM Studio (本地模型)",
		DefaultModel:  "local-model",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "http://localhost:1234/v1",
	},
	{
		ID:            "custom",
		DisplayName:   "自訂 (Custom)",
		DefaultModel:  "",
		DefaultTokens: 4000,
		DefaultTemp:   0.9,
		BaseURL:       "",
	},
}

// GetPresetByIndex returns the provider preset for a given tier and index
// Returns the custom preset if index is out of bounds
func GetPresetByIndex(tier trinity.TierLevel, index int) ProviderPreset {
	presets := tierProviderPresets[tier]
	if index < 0 || index >= len(presets) {
		// Return custom preset as fallback
		return presets[len(presets)-1]
	}
	return presets[index]
}

// GetPresetCount returns the number of presets for a given tier
func GetPresetCount(tier trinity.TierLevel) int {
	return len(tierProviderPresets[tier])
}

// FindPresetIndex finds the index of a preset matching the given provider ID
// Returns -1 if not found (indicates custom mode)
func FindPresetIndex(tier trinity.TierLevel, providerID string) int {
	presets := tierProviderPresets[tier]
	for i, preset := range presets {
		if preset.ID == providerID {
			return i
		}
	}
	return -1
}
