// Package config provides configuration management for Nightmare Assault.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config represents the application configuration.
type Config struct {
	Version    string           `json:"version"`
	Language   string           `json:"language"` // UI language: en, zh-TW, zh-CN, ja
	Theme      string           `json:"theme"`    // Color theme
	Audio      AudioConfig      `json:"audio"`
	API        APIConfig        `json:"api"`
	Typewriter TypewriterConfig `json:"typewriter"`
	Debug      DebugConfig      `json:"debug"`
}

// DebugConfig contains debug-related settings.
type DebugConfig struct {
	Enabled    bool `json:"enabled"`     // Enable debug mode
	LogAPIKeys bool `json:"log_api_keys"` // Log API key info (masked)
	LogRequests bool `json:"log_requests"` // Log API requests/responses
}

// AudioConfig contains audio-related settings.
type AudioConfig struct {
	MasterVolume     int                   `json:"master_volume"`       // 0-100
	BGMEnabled       bool                  `json:"bgm_enabled"`
	SFXEnabled       bool                  `json:"sfx_enabled"`
	BGMVolume        int                   `json:"bgm_volume"`          // 0-100
	SFXVolume        int                   `json:"sfx_volume"`          // 0-100
	HeartbeatEnabled bool                  `json:"heartbeat_enabled"`
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
	ProviderID string `json:"provider_id"`
	BaseURL    string `json:"base_url,omitempty"`  // Custom base URL (optional, hardcoded in provider.go)
	Model      string `json:"model"`               // Model name
	MaxTokens  int    `json:"max_tokens"`          // Max tokens for generation (default: 100000)
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
			BGMVolume:        70,
			SFXVolume:        80,
			HeartbeatEnabled: true,
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
