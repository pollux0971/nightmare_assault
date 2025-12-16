package api

// ProviderInfo contains static information about a provider.
type ProviderInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BaseURL     string    `json:"base_url"`
	Format      APIFormat `json:"format"`
	Category    string    `json:"category"` // "official", "gateway", "local"
}

// ProviderConfig contains runtime configuration for a provider.
type ProviderConfig struct {
	ProviderID string    `json:"provider_id"`
	APIKey     string    `json:"api_key"`
	BaseURL    string    `json:"base_url"`    // Override default base URL
	Model      string    `json:"model"`       // Model name
	MaxTokens  int       `json:"max_tokens"`  // Max tokens for response
	Format     APIFormat `json:"format"`      // API format
}

// BuiltinProviders returns all builtin provider definitions.
func BuiltinProviders() []ProviderInfo {
	return []ProviderInfo{
		// === Official APIs ===
		{ID: "openai", Name: "OpenAI", Description: "GPT-4o, GPT-4 Turbo, o1", BaseURL: "https://api.openai.com/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "anthropic", Name: "Anthropic", Description: "Claude 3.5 Sonnet, Claude 3 Opus", BaseURL: "https://api.anthropic.com/v1", Format: FormatAnthropic, Category: "official"},
		{ID: "google", Name: "Google", Description: "Gemini Pro, Gemini Ultra", BaseURL: "https://generativelanguage.googleapis.com/v1beta", Format: FormatGoogle, Category: "official"},
		{ID: "mistral", Name: "Mistral", Description: "Mistral Large, Codestral", BaseURL: "https://api.mistral.ai/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "cohere", Name: "Cohere", Description: "Command R+", BaseURL: "https://api.cohere.ai/v1", Format: FormatCohere, Category: "official"},
		{ID: "xai", Name: "xAI", Description: "Grok-2", BaseURL: "https://api.x.ai/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "deepseek", Name: "DeepSeek", Description: "DeepSeek V3, DeepSeek Coder", BaseURL: "https://api.deepseek.com/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "zhipu", Name: "智譜 AI", Description: "GLM-4", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Format: FormatOpenAI, Category: "official"},
		{ID: "moonshot", Name: "Moonshot", Description: "Kimi", BaseURL: "https://api.moonshot.cn/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "baichuan", Name: "百川", Description: "Baichuan", BaseURL: "https://api.baichuan-ai.com/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "minimax", Name: "MiniMax", Description: "abab6.5", BaseURL: "https://api.minimax.chat/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "01ai", Name: "零一萬物", Description: "Yi-Large", BaseURL: "https://api.lingyiwanwu.com/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "reka", Name: "Reka", Description: "Reka Core", BaseURL: "https://api.reka.ai/v1", Format: FormatOpenAI, Category: "official"},
		{ID: "ai21", Name: "AI21", Description: "Jamba", BaseURL: "https://api.ai21.com/studio/v1", Format: FormatOpenAI, Category: "official"},

		// === Gateway / Aggregators ===
		{ID: "openrouter", Name: "OpenRouter", Description: "多模型聚合平台", BaseURL: "https://openrouter.ai/api/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "together", Name: "Together AI", Description: "開源模型平台", BaseURL: "https://api.together.xyz/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "groq", Name: "Groq", Description: "超快推理平台", BaseURL: "https://api.groq.com/openai/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "fireworks", Name: "Fireworks", Description: "高性能推理", BaseURL: "https://api.fireworks.ai/inference/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "perplexity", Name: "Perplexity", Description: "搜尋增強 AI", BaseURL: "https://api.perplexity.ai", Format: FormatOpenAI, Category: "gateway"},
		{ID: "deepinfra", Name: "Deepinfra", Description: "開源模型託管", BaseURL: "https://api.deepinfra.com/v1/openai", Format: FormatOpenAI, Category: "gateway"},
		{ID: "lepton", Name: "Lepton AI", Description: "AI 平台", BaseURL: "https://api.lepton.ai/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "novita", Name: "Novita AI", Description: "多模型平台", BaseURL: "https://api.novita.ai/v3/openai", Format: FormatOpenAI, Category: "gateway"},
		{ID: "siliconflow", Name: "SiliconFlow", Description: "矽流科技", BaseURL: "https://api.siliconflow.cn/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "cerebras", Name: "Cerebras", Description: "超快推理", BaseURL: "https://api.cerebras.ai/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "hyperbolic", Name: "Hyperbolic", Description: "開源模型", BaseURL: "https://api.hyperbolic.xyz/v1", Format: FormatOpenAI, Category: "gateway"},
		{ID: "sambanova", Name: "Sambanova", Description: "企業 AI", BaseURL: "https://api.sambanova.ai/v1", Format: FormatOpenAI, Category: "gateway"},

		// === Local / Self-hosted ===
		{ID: "ollama", Name: "Ollama", Description: "本地模型", BaseURL: "http://localhost:11434/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "lmstudio", Name: "LM Studio", Description: "本地 GUI", BaseURL: "http://localhost:1234/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "llamacpp", Name: "llama.cpp", Description: "本地推理", BaseURL: "http://localhost:8080/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "localai", Name: "LocalAI", Description: "本地 AI", BaseURL: "http://localhost:8080/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "jan", Name: "Jan", Description: "本地 AI 助手", BaseURL: "http://localhost:1337/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "vllm", Name: "vLLM", Description: "高性能推理", BaseURL: "http://localhost:8000/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "koboldcpp", Name: "Kobold.cpp", Description: "本地推理", BaseURL: "http://localhost:5001/v1", Format: FormatOpenAI, Category: "local"},
		{ID: "tabbyapi", Name: "TabbyAPI", Description: "本地 API", BaseURL: "http://localhost:5000/v1", Format: FormatOpenAI, Category: "local"},

		// === Custom ===
		{ID: "custom", Name: "自訂", Description: "自訂 API 端點", BaseURL: "", Format: FormatOpenAI, Category: "custom"},
	}
}

// GetProviderInfo returns provider info by ID.
func GetProviderInfo(id string) *ProviderInfo {
	for _, p := range BuiltinProviders() {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// GetProvidersByCategory returns providers filtered by category.
func GetProvidersByCategory(category string) []ProviderInfo {
	var result []ProviderInfo
	for _, p := range BuiltinProviders() {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// defaultModels maps provider IDs to their default model names.
var defaultModels = map[string]string{
	"openai":      "gpt-4o",
	"anthropic":   "claude-3-5-sonnet-20241022",
	"google":      "gemini-1.5-pro",
	"mistral":     "mistral-large-latest",
	"cohere":      "command-r-plus",
	"xai":         "grok-2",
	"deepseek":    "deepseek-chat",
	"openrouter":  "anthropic/claude-3.5-sonnet",
	"together":    "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
	"groq":        "llama-3.1-70b-versatile",
	"fireworks":   "accounts/fireworks/models/llama-v3p1-70b-instruct",
	"perplexity":  "llama-3.1-sonar-large-128k-online",
	"ollama":      "llama3.1",
	"lmstudio":    "local-model",
}

// modelHints provides suggested models for each provider.
var modelHints = map[string][]string{
	"openai": {
		"gpt-4o",
		"gpt-4-turbo",
		"gpt-4o-mini",
		"o1-preview",
	},
	"anthropic": {
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-haiku-20240307",
	},
	"google": {
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-2.0-flash-exp",
	},
	"openrouter": {
		"anthropic/claude-3.5-sonnet",
		"openai/gpt-4-turbo",
		"google/gemini-pro-1.5",
		"meta-llama/llama-3.1-70b-instruct",
	},
	"together": {
		"meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
		"mistralai/Mixtral-8x22B-Instruct-v0.1",
	},
	"groq": {
		"llama-3.1-70b-versatile",
		"llama-3.1-8b-instant",
		"mixtral-8x7b-32768",
	},
	"ollama": {
		"llama3.1",
		"mistral",
		"codellama",
	},
}

// GetDefaultModel returns the default model for a provider.
// Returns empty string if no default is defined.
func GetDefaultModel(providerID string) string {
	if model, ok := defaultModels[providerID]; ok {
		return model
	}
	return ""
}

// GetModelHints returns a list of suggested models for a provider.
// Returns empty slice if no hints are defined.
func GetModelHints(providerID string) []string {
	if hints, ok := modelHints[providerID]; ok {
		return hints
	}
	return nil
}
