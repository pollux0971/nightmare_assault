// Package config provides configuration management for Nightmare Assault.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
}

// AudioConfig contains audio-related settings.
type AudioConfig struct {
	BGMEnabled bool    `json:"bgm_enabled"`
	SFXEnabled bool    `json:"sfx_enabled"`
	BGMVolume  float64 `json:"bgm_volume"` // 0.0 - 1.0
	SFXVolume  float64 `json:"sfx_volume"` // 0.0 - 1.0
}

// TypewriterConfig contains typewriter effect settings.
type TypewriterConfig struct {
	Enabled    bool `json:"enabled"`      // Enable/disable typewriter effect
	Speed      int  `json:"speed"`        // Characters per second (10-200)
	ShowCursor bool `json:"show_cursor"`  // Show typing cursor
}

// APIConfig contains API-related settings.
type APIConfig struct {
	Smart      ProviderSettings            `json:"smart"`       // Smart Model (GPT-4 class)
	Fast       ProviderSettings            `json:"fast"`        // Fast Model (GPT-3.5 class)
	APIKeys    map[string]string          `json:"api_keys"`    // Encrypted API keys per provider
	LastTested map[string]time.Time       `json:"last_tested"` // Last test time per provider
}

// ProviderSettings contains settings for a specific API role (smart/fast).
type ProviderSettings struct {
	ProviderID string `json:"provider_id"`
	BaseURL    string `json:"base_url,omitempty"`  // Custom base URL (optional)
	Model      string `json:"model"`               // Model name
	MaxTokens  int    `json:"max_tokens"`
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
			BGMEnabled: true,
			SFXEnabled: true,
			BGMVolume:  0.7,
			SFXVolume:  0.8,
		},
		API: APIConfig{
			Smart: ProviderSettings{
				ProviderID: "",
				Model:      "",
				MaxTokens:  4096,
			},
			Fast: ProviderSettings{
				ProviderID: "",
				Model:      "",
				MaxTokens:  2048,
			},
			APIKeys:    make(map[string]string),
			LastTested: make(map[string]time.Time),
		},
		Typewriter: TypewriterConfig{
			Enabled:    true,
			Speed:      40, // Default 40 chars/second
			ShowCursor: true,
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
	// Must have at least one API configured (smart or fast)
	return c.API.Smart.ProviderID != "" || c.API.Fast.ProviderID != ""
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
