// Package config provides configuration management for Nightmare Assault.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestReload tests the config hot reload functionality.
// Story 10-7 AC1: Reload() should successfully reload valid config
func TestReload(t *testing.T) {
	// Create temporary home directory structure
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".nightmare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	// Set temporary HOME for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create initial config
	initialConfig := DefaultConfig()
	initialConfig.Audio.MasterVolume = 50
	initialConfig.Typewriter.Speed = 30

	// Save initial config
	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify initial values
	if cfg.Audio.MasterVolume != 50 {
		t.Errorf("Expected initial master volume 50, got %d", cfg.Audio.MasterVolume)
	}
	if cfg.Typewriter.Speed != 30 {
		t.Errorf("Expected initial typewriter speed 30, got %d", cfg.Typewriter.Speed)
	}

	// Modify config file on disk
	modifiedConfig := DefaultConfig()
	modifiedConfig.Audio.MasterVolume = 80
	modifiedConfig.Typewriter.Speed = 60

	data, err = json.MarshalIndent(modifiedConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal modified config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write modified config: %v", err)
	}

	// Reload config
	if err := cfg.Reload(); err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	// Verify config was reloaded
	if cfg.Audio.MasterVolume != 80 {
		t.Errorf("Expected reloaded master volume 80, got %d", cfg.Audio.MasterVolume)
	}
	if cfg.Typewriter.Speed != 60 {
		t.Errorf("Expected reloaded typewriter speed 60, got %d", cfg.Typewriter.Speed)
	}
}

// TestReload_InvalidJSON tests reload behavior with invalid JSON.
// Story 10-7 AC2: Should preserve original config on error
func TestReload_InvalidJSON(t *testing.T) {
	// Create temporary home directory structure
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".nightmare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	// Set temporary HOME for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create initial valid config
	initialConfig := DefaultConfig()
	initialConfig.Audio.MasterVolume = 50

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Store original value
	originalVolume := cfg.Audio.MasterVolume

	// Write invalid JSON to config file
	invalidJSON := []byte("{invalid json")
	if err := os.WriteFile(configPath, invalidJSON, 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Reload should fail
	err = cfg.Reload()
	if err == nil {
		t.Fatal("Expected error when reloading invalid JSON, got nil")
	}

	// Verify original config is preserved
	if cfg.Audio.MasterVolume != originalVolume {
		t.Errorf("Expected original volume %d to be preserved, got %d", originalVolume, cfg.Audio.MasterVolume)
	}
}

// TestReload_InvalidValues tests reload behavior with invalid config values.
// Story 10-7 AC2: Should preserve original config on validation error
func TestReload_InvalidValues(t *testing.T) {
	// Create temporary home directory structure
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".nightmare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	// Set temporary HOME for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create initial valid config
	initialConfig := DefaultConfig()
	initialConfig.Audio.MasterVolume = 50
	initialConfig.Typewriter.Speed = 40

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Store original values
	originalVolume := cfg.Audio.MasterVolume
	originalSpeed := cfg.Typewriter.Speed

	// Create config with invalid values
	invalidConfig := DefaultConfig()
	invalidConfig.Audio.MasterVolume = 150 // Invalid: > 100
	invalidConfig.Typewriter.Speed = 5     // Invalid: < 10

	data, err = json.MarshalIndent(invalidConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal invalid config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Reload should fail due to validation
	err = cfg.Reload()
	if err == nil {
		t.Fatal("Expected validation error when reloading invalid config, got nil")
	}

	// Verify error is ConfigError
	if _, ok := err.(*ConfigError); !ok {
		t.Errorf("Expected ConfigError, got %T: %v", err, err)
	}

	// Verify original config is preserved
	if cfg.Audio.MasterVolume != originalVolume {
		t.Errorf("Expected original volume %d to be preserved, got %d", originalVolume, cfg.Audio.MasterVolume)
	}
	if cfg.Typewriter.Speed != originalSpeed {
		t.Errorf("Expected original speed %d to be preserved, got %d", originalSpeed, cfg.Typewriter.Speed)
	}
}

// TestValidateConfig tests the config validation function.
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*Config)
		wantError bool
		errorField string
	}{
		{
			name:      "valid config",
			modify:    func(c *Config) {},
			wantError: false,
		},
		{
			name: "invalid master volume (negative)",
			modify: func(c *Config) {
				c.Audio.MasterVolume = -10
			},
			wantError: true,
			errorField: "audio.master_volume",
		},
		{
			name: "invalid master volume (> 100)",
			modify: func(c *Config) {
				c.Audio.MasterVolume = 150
			},
			wantError: true,
			errorField: "audio.master_volume",
		},
		{
			name: "invalid bgm volume (negative)",
			modify: func(c *Config) {
				c.Audio.BGMVolume = -5
			},
			wantError: true,
			errorField: "audio.bgm_volume",
		},
		{
			name: "invalid bgm volume (> 100)",
			modify: func(c *Config) {
				c.Audio.BGMVolume = 120
			},
			wantError: true,
			errorField: "audio.bgm_volume",
		},
		{
			name: "invalid sfx volume (negative)",
			modify: func(c *Config) {
				c.Audio.SFXVolume = -1
			},
			wantError: true,
			errorField: "audio.sfx_volume",
		},
		{
			name: "invalid sfx volume (> 100)",
			modify: func(c *Config) {
				c.Audio.SFXVolume = 101
			},
			wantError: true,
			errorField: "audio.sfx_volume",
		},
		{
			name: "invalid typewriter speed (too slow)",
			modify: func(c *Config) {
				c.Typewriter.Speed = 5
			},
			wantError: true,
			errorField: "typewriter.speed",
		},
		{
			name: "invalid typewriter speed (too fast)",
			modify: func(c *Config) {
				c.Typewriter.Speed = 250
			},
			wantError: true,
			errorField: "typewriter.speed",
		},
		{
			name: "invalid update check interval",
			modify: func(c *Config) {
				c.Update.CheckInterval = 0
			},
			wantError: true,
			errorField: "update.check_interval",
		},
		{
			name: "boundary: master volume = 0",
			modify: func(c *Config) {
				c.Audio.MasterVolume = 0
			},
			wantError: false,
		},
		{
			name: "boundary: master volume = 100",
			modify: func(c *Config) {
				c.Audio.MasterVolume = 100
			},
			wantError: false,
		},
		{
			name: "boundary: typewriter speed = 10",
			modify: func(c *Config) {
				c.Typewriter.Speed = 10
			},
			wantError: false,
		},
		{
			name: "boundary: typewriter speed = 200",
			modify: func(c *Config) {
				c.Typewriter.Speed = 200
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)

			err := validateConfig(cfg)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
					return
				}
				if configErr, ok := err.(*ConfigError); ok {
					if configErr.Field != tt.errorField {
						t.Errorf("Expected error field %s, got %s", tt.errorField, configErr.Field)
					}
				} else {
					t.Errorf("Expected ConfigError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestConfigError tests the ConfigError error message format.
func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "audio.master_volume",
		Message: "must be between 0 and 100",
	}

	expected := "Config error [audio.master_volume]: must be between 0 and 100"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

// TestIsValidLanguage tests the language validation function.
func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		lang  string
		valid bool
	}{
		{"zh-TW", true},
		{"zh-CN", true},
		{"en-US", true},
		{"ja-JP", false},
		{"fr-FR", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := IsValidLanguage(tt.lang)
			if result != tt.valid {
				t.Errorf("IsValidLanguage(%q) = %v, want %v", tt.lang, result, tt.valid)
			}
		})
	}
}

// TestReload_InvalidLanguage tests reload behavior with invalid language.
func TestReload_InvalidLanguage(t *testing.T) {
	// Create temporary home directory structure
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".nightmare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	// Set temporary HOME for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create initial valid config
	initialConfig := DefaultConfig()
	initialConfig.Language = "en-US"

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Store original language
	originalLang := cfg.Language

	// Create config with invalid language
	invalidConfig := DefaultConfig()
	invalidConfig.Language = "invalid-lang"

	data, err = json.MarshalIndent(invalidConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal invalid config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Reload should fail due to validation
	err = cfg.Reload()
	if err == nil {
		t.Fatal("Expected validation error when reloading invalid language, got nil")
	}

	// Verify error is ConfigError with language field
	if configErr, ok := err.(*ConfigError); ok {
		if configErr.Field != "language" {
			t.Errorf("Expected error field 'language', got %s", configErr.Field)
		}
	} else {
		t.Errorf("Expected ConfigError, got %T: %v", err, err)
	}

	// Verify original config is preserved
	if cfg.Language != originalLang {
		t.Errorf("Expected original language %q to be preserved, got %q", originalLang, cfg.Language)
	}
}
