// Package config provides configuration management for Nightmare Assault.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// supportedLanguages defines the languages supported by the application.
// This is duplicated from i18n package to avoid circular dependency.
var supportedLanguages = []string{"zh-TW", "zh-CN", "en-US"}

// IsValidLanguage checks if a language code is supported.
func IsValidLanguage(lang string) bool {
	for _, supported := range supportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// Config represents the application configuration.
type Config struct {
	Version    string           `json:"version"`
	Language   string           `json:"language"` // UI language: en, zh-TW, zh-CN, ja
	Theme      string           `json:"theme"`    // Color theme
	Audio      AudioConfig      `json:"audio"`
	API        APIConfig        `json:"api"`
	Typewriter TypewriterConfig `json:"typewriter"`
	Debug      DebugConfig      `json:"debug"`
	Update     UpdateConfig     `json:"update"`   // Auto-update settings
	Trinity    TrinityConfig    `json:"trinity,omitempty"` // Trinity System configuration (Story 9-3)
}

// DebugConfig contains debug-related settings.
type DebugConfig struct {
	Enabled    bool `json:"enabled"`     // Enable debug mode
	LogAPIKeys bool `json:"log_api_keys"` // Log API key info (masked)
	LogRequests bool `json:"log_requests"` // Log API requests/responses
}

// UpdateConfig contains auto-update settings.
type UpdateConfig struct {
	Enabled         bool `json:"enabled"`           // Enable auto-update check on startup
	CheckInterval   int  `json:"check_interval"`    // Check interval in hours (default: 24)
	AllowPrerelease bool `json:"allow_prerelease"`  // Allow pre-release versions
}

// AudioConfig contains audio-related settings.
type AudioConfig struct {
	MasterVolume     int                   `json:"master_volume"`       // 0-100
	BGMEnabled       bool                  `json:"bgm_enabled"`
	SFXEnabled       bool                  `json:"sfx_enabled"`
	BGMVolume        int                   `json:"bgm_volume"`          // 0-100
	SFXVolume        int                   `json:"sfx_volume"`          // 0-100
	HeartbeatEnabled bool                  `json:"heartbeat_enabled"`
	BGMLoop          bool                  `json:"bgm_loop"`            // Loop BGM playback (Story 10-2 AC7)
	Platform         string                `json:"platform"`            // Detected platform: windows, linux, darwin
	PlatformSettings PlatformAudioSettings `json:"platform_settings"`   // Platform-specific audio optimizations
}

// PlatformAudioSettings contains platform-specific audio optimizations
type PlatformAudioSettings struct {
	BufferSize   int  `json:"buffer_size"`    // Audio buffer size (samples)
	SampleRate   int  `json:"sample_rate"`    // Sample rate (Hz)
	ChannelCount int  `json:"channel_count"`  // 1=mono, 2=stereo
	LowLatency   bool `json:"low_latency"`    // Low latency mode
}

// TypewriterConfig contains typewriter effect settings.
type TypewriterConfig struct {
	Enabled    bool `json:"enabled"`      // Enable/disable typewriter effect
	Speed      int  `json:"speed"`        // Characters per second (10-200)
	ShowCursor bool `json:"show_cursor"`  // Show typing cursor
}

// APIConfig contains API-related settings.
type APIConfig struct {
	Provider   ProviderSettings            `json:"provider"`     // Active provider configuration
	APIKeys    map[string]string          `json:"api_keys"`     // Encrypted API keys per provider
	LastTested map[string]time.Time       `json:"last_tested"`  // Last test time per provider

	// Deprecated: For backward compatibility during migration only
	Smart      *ProviderSettings          `json:"smart,omitempty"`
	Fast       *ProviderSettings          `json:"fast,omitempty"`
}

// ProviderSettings contains settings for the active API provider.
type ProviderSettings struct {
	ProviderID  string  `json:"provider_id"`
	BaseURL     string  `json:"base_url,omitempty"`  // Custom base URL (optional, hardcoded in provider.go)
	Model       string  `json:"model"`               // Model name
	MaxTokens   int     `json:"max_tokens"`          // Max tokens for generation (default: 100000)
	Temperature float64 `json:"temperature,omitempty"` // Temperature for generation (0.0-1.0)
}

// TrinityConfig contains Trinity System configuration (Story 9-3).
type TrinityConfig struct {
	// Enabled controls whether Trinity System is active.
	// When false, the system uses the legacy API.Provider configuration.
	Enabled bool `json:"enabled"`

	// Thinking tier provider (high-quality models for complex reasoning).
	Thinking ProviderSettings `json:"thinking,omitempty"`

	// Reactive tier provider (balanced models for interactive responses).
	Reactive ProviderSettings `json:"reactive,omitempty"`

	// Rapid tier provider (fast models for quick responses).
	Rapid ProviderSettings `json:"rapid,omitempty"`

	// AgentTierOverrides allows users to override the default agent-to-tier mapping.
	// Keys are agent names (e.g., "JudgeAgent"), values are tier names ("Thinking", "Reactive", "Rapid").
	AgentTierOverrides map[string]string `json:"agent_tier_overrides,omitempty"`

	// FallbackEnabled enables automatic fallback to lower tiers on failure.
	FallbackEnabled bool `json:"fallback_enabled"`

	// Guardian contains Guardian Layer configuration for player safety.
	Guardian GuardianSettings `json:"guardian,omitempty"`
}

// GuardianSettings contains Guardian Layer configuration.
type GuardianSettings struct {
	// Enabled controls whether Guardian Layer is active.
	Enabled bool `json:"enabled"`

	// LowHPThreshold triggers enhanced model when player HP is below this percentage (0-100).
	LowHPThreshold int `json:"low_hp_threshold,omitempty"`

	// LowSanThreshold triggers enhanced model when player Sanity is below this percentage (0-100).
	LowSanThreshold int `json:"low_san_threshold,omitempty"`

	// MaxConsecutiveDeaths triggers enhanced model after this many consecutive deaths.
	MaxConsecutiveDeaths int `json:"max_consecutive_deaths,omitempty"`
}

// detectPlatform returns the current platform
func detectPlatform() string {
	return runtime.GOOS // "windows", "linux", "darwin"
}

// DefaultPlatformSettings returns platform-optimized defaults
func DefaultPlatformSettings() PlatformAudioSettings {
	platform := detectPlatform()

	switch platform {
	case "windows":
		return PlatformAudioSettings{
			BufferSize:   2048,  // Larger buffer for Windows stability
			SampleRate:   48000,
			ChannelCount: 2,
			LowLatency:   false, // Avoid WASAPI issues
		}
	case "linux":
		return PlatformAudioSettings{
			BufferSize:   1024,  // Medium buffer for PulseAudio/ALSA
			SampleRate:   48000,
			ChannelCount: 2,
			LowLatency:   true,  // Linux audio is reliable
		}
	case "darwin":
		return PlatformAudioSettings{
			BufferSize:   512,   // Small buffer for CoreAudio
			SampleRate:   48000,
			ChannelCount: 2,
			LowLatency:   true,  // CoreAudio excels at low latency
		}
	default:
		return PlatformAudioSettings{
			BufferSize:   1024,
			SampleRate:   48000,
			ChannelCount: 2,
			LowLatency:   false,
		}
	}
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	// Note: Language will be set by caller using i18n.DetectSystemLanguage()
	// to avoid circular dependency between config and i18n packages
	return &Config{
		Version:  "1.0",
		Language: "en-US", // Will be overridden by system detection
		Theme:    "midnight", // Default theme
		Audio: AudioConfig{
			MasterVolume:     100,
			BGMEnabled:       true,
			SFXEnabled:       true,
			BGMVolume:        50,
			SFXVolume:        80,
			HeartbeatEnabled: true,
			BGMLoop:          true, // Enable loop by default (Story 10-2 AC7)
			Platform:         detectPlatform(),
			PlatformSettings: DefaultPlatformSettings(),
		},
		API: APIConfig{
			Provider: ProviderSettings{
				ProviderID: "",
				Model:      "",
				MaxTokens:  100000, // Default to 100k tokens
			},
			APIKeys:    make(map[string]string),
			LastTested: make(map[string]time.Time),
		},
		Typewriter: TypewriterConfig{
			Enabled:    true,
			Speed:      40, // Default 40 chars/second
			ShowCursor: true,
		},
		Debug: DebugConfig{
			Enabled:    false, // Disabled by default
			LogAPIKeys: false,
			LogRequests: false,
		},
		Update: UpdateConfig{
			Enabled:         true, // Enable auto-update check by default
			CheckInterval:   24,   // Check every 24 hours
			AllowPrerelease: false, // Stable releases only
		},
	}
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".nightmare", "config.json"), nil
}

// ConfigDir returns the path to the config directory.
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".nightmare"), nil
}

// Load loads the configuration from the default path.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadFromPath(path)
}

// LoadFromPath loads the configuration from a specific path.
func LoadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Ensure maps are initialized
	if cfg.API.APIKeys == nil {
		cfg.API.APIKeys = make(map[string]string)
	}
	if cfg.API.LastTested == nil {
		cfg.API.LastTested = make(map[string]time.Time)
	}

	// Auto-detect platform if not set
	if cfg.Audio.Platform == "" {
		cfg.Audio.Platform = detectPlatform()
		cfg.Audio.PlatformSettings = DefaultPlatformSettings()
	}

	// Migrate from Smart/Fast to single Provider
	if cfg.MigrateToSingleProvider() {
		// Auto-save after migration
		cfg.Save()
	}

	// Override with environment variables if present
	cfg.LoadFromEnv()

	return &cfg, nil
}

// Save saves the configuration to the default path.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	return c.SaveToPath(path)
}

// SaveToPath saves the configuration to a specific path.
func (c *Config) SaveToPath(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// HasAPIKey checks if an API key is configured for a provider.
func (c *Config) HasAPIKey(providerID string) bool {
	_, ok := c.API.APIKeys[providerID]
	return ok
}

// IsConfigured checks if the minimum configuration is set.
func (c *Config) IsConfigured() bool {
	return c.API.Provider.ProviderID != ""
}

// MigrateToSingleProvider migrates from Smart/Fast dual config to single Provider config.
// Returns true if migration was performed.
func (c *Config) MigrateToSingleProvider() bool {
	// Check if migration is needed
	if c.API.Smart == nil && c.API.Fast == nil {
		return false
	}

	// Prefer Smart over Fast
	if c.API.Smart != nil && c.API.Smart.ProviderID != "" {
		c.API.Provider = *c.API.Smart
	} else if c.API.Fast != nil && c.API.Fast.ProviderID != "" {
		c.API.Provider = *c.API.Fast
	} else {
		// Initialize with defaults
		c.API.Provider = ProviderSettings{
			ProviderID: "",
			Model:      "",
			MaxTokens:  100000,
		}
	}

	// Ensure MaxTokens is at least 100,000 (user can modify later)
	if c.API.Provider.MaxTokens < 100000 {
		c.API.Provider.MaxTokens = 100000
	}

	// Clear deprecated fields
	c.API.Smart = nil
	c.API.Fast = nil

	return true
}

// Exists checks if the config file exists.
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// LoadFromEnv loads API configuration from environment variables.
// Environment variables take precedence over config file settings.
//
// Supported environment variables:
//   - NIGHTMARE_API_PROVIDER: Provider ID (openrouter, anthropic, openai, google, cohere)
//   - NIGHTMARE_API_MODEL: Model name
//   - NIGHTMARE_API_MAX_TOKENS: Maximum tokens (integer)
//   - NIGHTMARE_API_BASE_URL: Custom base URL (optional)
//   - OPENROUTER_API_KEY: OpenRouter API key
//   - ANTHROPIC_API_KEY: Anthropic API key
//   - OPENAI_API_KEY: OpenAI API key
//   - GOOGLE_API_KEY: Google AI API key
//   - COHERE_API_KEY: Cohere API key
//   - NIGHTMARE_DEBUG: Enable debug mode (true/false)
//   - NIGHTMARE_LOG_API_KEYS: Log API keys (masked) (true/false)
//   - NIGHTMARE_LOG_REQUESTS: Log API requests/responses (true/false)
func (c *Config) LoadFromEnv() {
	// Load provider settings
	if provider := os.Getenv("NIGHTMARE_API_PROVIDER"); provider != "" {
		c.API.Provider.ProviderID = provider
	}

	if model := os.Getenv("NIGHTMARE_API_MODEL"); model != "" {
		c.API.Provider.Model = model
	}

	if maxTokensStr := os.Getenv("NIGHTMARE_API_MAX_TOKENS"); maxTokensStr != "" {
		if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil && maxTokens > 0 {
			c.API.Provider.MaxTokens = maxTokens
		}
	}

	if baseURL := os.Getenv("NIGHTMARE_API_BASE_URL"); baseURL != "" {
		c.API.Provider.BaseURL = baseURL
	}

	// Load API keys from environment variables
	apiKeyEnvVars := map[string]string{
		"openrouter": "OPENROUTER_API_KEY",
		"anthropic":  "ANTHROPIC_API_KEY",
		"openai":     "OPENAI_API_KEY",
		"google":     "GOOGLE_API_KEY",
		"cohere":     "COHERE_API_KEY",
	}

	for providerID, envVar := range apiKeyEnvVars {
		if apiKey := os.Getenv(envVar); apiKey != "" {
			c.API.APIKeys[providerID] = apiKey
		}
	}

	// Load debug settings
	if debugStr := os.Getenv("NIGHTMARE_DEBUG"); debugStr != "" {
		c.Debug.Enabled = strings.ToLower(debugStr) == "true"
	}

	if logAPIKeysStr := os.Getenv("NIGHTMARE_LOG_API_KEYS"); logAPIKeysStr != "" {
		c.Debug.LogAPIKeys = strings.ToLower(logAPIKeysStr) == "true"
	}

	if logRequestsStr := os.Getenv("NIGHTMARE_LOG_REQUESTS"); logRequestsStr != "" {
		c.Debug.LogRequests = strings.ToLower(logRequestsStr) == "true"
	}
}

// GetAPIKey retrieves the API key for a given provider.
// Priority: Environment variable > Config file
func (c *Config) GetAPIKey(providerID string) string {
	// Check environment variable first (highest priority)
	envVarMap := map[string]string{
		"openrouter": "OPENROUTER_API_KEY",
		"anthropic":  "ANTHROPIC_API_KEY",
		"openai":     "OPENAI_API_KEY",
		"google":     "GOOGLE_API_KEY",
		"cohere":     "COHERE_API_KEY",
	}

	if envVar, ok := envVarMap[providerID]; ok {
		if apiKey := os.Getenv(envVar); apiKey != "" {
			return apiKey
		}
	}

	// Fall back to config file
	if apiKey, ok := c.API.APIKeys[providerID]; ok {
		return apiKey
	}

	return ""
}

// SetAPIKey sets the API key for a provider and saves to config file.
func (c *Config) SetAPIKey(providerID, apiKey string) error {
	if c.API.APIKeys == nil {
		c.API.APIKeys = make(map[string]string)
	}
	c.API.APIKeys[providerID] = apiKey
	return c.Save()
}

// Reload reloads the configuration from disk and applies it to the current instance.
// Story 10-7 AC1: Reload config from ~/.nightmare/config.json
// Story 10-7 AC2: On error, preserve original config and return detailed error
//
// Returns error if:
//   - Config file cannot be read
//   - Config file has invalid JSON format
//   - Config validation fails
func (c *Config) Reload() error {
	// Load fresh config from disk
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	newConfig, err := LoadFromPath(path)
	if err != nil {
		// Return error, original config is preserved
		return err
	}

	// Validate the new config before applying
	if err := validateConfig(newConfig); err != nil {
		// Return validation error, original config is preserved
		return err
	}

	// Apply the new config to the current instance
	*c = *newConfig

	return nil
}

// validateConfig validates the configuration for required fields and ranges.
func validateConfig(c *Config) error {
	if c == nil {
		return nil // Allow nil config for backward compatibility
	}

	// Validate language setting
	if c.Language != "" && !IsValidLanguage(c.Language) {
		return &ConfigError{Field: "language", Message: "unsupported language: " + c.Language}
	}

	// Validate audio settings
	if c.Audio.MasterVolume < 0 || c.Audio.MasterVolume > 100 {
		return &ConfigError{Field: "audio.master_volume", Message: "must be between 0 and 100"}
	}
	if c.Audio.BGMVolume < 0 || c.Audio.BGMVolume > 100 {
		return &ConfigError{Field: "audio.bgm_volume", Message: "must be between 0 and 100"}
	}
	if c.Audio.SFXVolume < 0 || c.Audio.SFXVolume > 100 {
		return &ConfigError{Field: "audio.sfx_volume", Message: "must be between 0 and 100"}
	}

	// Validate typewriter settings
	if c.Typewriter.Speed < 10 || c.Typewriter.Speed > 200 {
		return &ConfigError{Field: "typewriter.speed", Message: "must be between 10 and 200 chars/second"}
	}

	// Validate update settings
	if c.Update.CheckInterval < 1 {
		return &ConfigError{Field: "update.check_interval", Message: "must be at least 1 hour"}
	}

	// Validate Trinity configuration (Story 9-3 AC4)
	if err := ValidateTrinityConfig(&c.Trinity); err != nil {
		return err
	}

	return nil
}

// ValidateTrinityConfig validates Trinity configuration (Story 9-3 AC4).
func ValidateTrinityConfig(tc *TrinityConfig) error {
	if tc == nil || !tc.Enabled {
		return nil // Trinity disabled, no validation needed
	}

	// Validate that all three tiers have provider configuration
	if tc.Thinking.ProviderID == "" {
		return &ConfigError{Field: "trinity.thinking.provider_id", Message: "Thinking tier provider_id is required when Trinity is enabled"}
	}
	if tc.Thinking.Model == "" {
		return &ConfigError{Field: "trinity.thinking.model", Message: "Thinking tier model is required when Trinity is enabled"}
	}

	if tc.Reactive.ProviderID == "" {
		return &ConfigError{Field: "trinity.reactive.provider_id", Message: "Reactive tier provider_id is required when Trinity is enabled"}
	}
	if tc.Reactive.Model == "" {
		return &ConfigError{Field: "trinity.reactive.model", Message: "Reactive tier model is required when Trinity is enabled"}
	}

	if tc.Rapid.ProviderID == "" {
		return &ConfigError{Field: "trinity.rapid.provider_id", Message: "Rapid tier provider_id is required when Trinity is enabled"}
	}
	if tc.Rapid.Model == "" {
		return &ConfigError{Field: "trinity.rapid.model", Message: "Rapid tier model is required when Trinity is enabled"}
	}

	// Validate AgentTierOverrides - values must be valid tier names
	validTierNames := map[string]bool{
		"Thinking": true,
		"Reactive": true,
		"Rapid":    true,
	}

	for agentName, tierName := range tc.AgentTierOverrides {
		if !validTierNames[tierName] {
			return &ConfigError{
				Field:   "trinity.agent_tier_overrides." + agentName,
				Message: "invalid tier name '" + tierName + "', must be one of: Thinking, Reactive, Rapid",
			}
		}
	}

	// Validate Guardian thresholds
	if tc.Guardian.Enabled {
		if tc.Guardian.LowHPThreshold < 0 || tc.Guardian.LowHPThreshold > 100 {
			return &ConfigError{Field: "trinity.guardian.low_hp_threshold", Message: "must be between 0 and 100"}
		}
		if tc.Guardian.LowSanThreshold < 0 || tc.Guardian.LowSanThreshold > 100 {
			return &ConfigError{Field: "trinity.guardian.low_san_threshold", Message: "must be between 0 and 100"}
		}
		if tc.Guardian.MaxConsecutiveDeaths < 0 {
			return &ConfigError{Field: "trinity.guardian.max_consecutive_deaths", Message: "must be non-negative"}
		}
	}

	return nil
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string
	Message string
}

// Error returns the error message in English (i18n is handled at UI layer).
func (e *ConfigError) Error() string {
	return "Config error [" + e.Field + "]: " + e.Message
}
