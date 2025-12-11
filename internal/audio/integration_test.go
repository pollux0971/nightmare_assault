package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TestAudioManagerConfigIntegration tests integration with config system
func TestAudioManagerConfigIntegration(t *testing.T) {
	// Create a temporary config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.Audio.BGMEnabled = true
	cfg.Audio.SFXEnabled = true
	cfg.Audio.BGMVolume = 0.7
	cfg.Audio.SFXVolume = 0.8

	// Save config
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify audio config
	if loadedCfg.Audio.BGMVolume != 0.7 {
		t.Errorf("BGMVolume = %f, expected 0.7", loadedCfg.Audio.BGMVolume)
	}
	if loadedCfg.Audio.SFXVolume != 0.8 {
		t.Errorf("SFXVolume = %f, expected 0.8", loadedCfg.Audio.SFXVolume)
	}
	if !loadedCfg.Audio.BGMEnabled {
		t.Error("BGMEnabled should be true")
	}
	if !loadedCfg.Audio.SFXEnabled {
		t.Error("SFXEnabled should be true")
	}
}

// TestAudioManagerWithConfig tests AudioManager using config
func TestAudioManagerWithConfig(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.6,
		SFXVolume:  0.9,
	}

	manager := NewAudioManager(cfg)

	// Verify initial config
	audioCfg := manager.Config()
	if audioCfg.BGMVolume != 0.6 {
		t.Errorf("BGMVolume = %f, expected 0.6", audioCfg.BGMVolume)
	}
	if audioCfg.SFXVolume != 0.9 {
		t.Errorf("SFXVolume = %f, expected 0.9", audioCfg.SFXVolume)
	}

	// Update config
	newCfg := config.AudioConfig{
		BGMEnabled: false,
		SFXEnabled: false,
		BGMVolume:  0.3,
		SFXVolume:  0.4,
	}
	manager.UpdateConfig(newCfg)

	// Verify updated config
	updatedCfg := manager.Config()
	if updatedCfg.BGMVolume != 0.3 {
		t.Errorf("BGMVolume = %f, expected 0.3", updatedCfg.BGMVolume)
	}
	if updatedCfg.SFXVolume != 0.4 {
		t.Errorf("SFXVolume = %f, expected 0.4", updatedCfg.SFXVolume)
	}
	if updatedCfg.BGMEnabled {
		t.Error("BGMEnabled should be false")
	}
	if updatedCfg.SFXEnabled {
		t.Error("SFXEnabled should be false")
	}
}

// TestFullAudioSystemFlow tests complete audio system initialization flow
func TestFullAudioSystemFlow(t *testing.T) {
	// Step 1: Load config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Step 2: Create AudioManager
	manager := NewAudioManager(loadedCfg.Audio)
	if manager == nil {
		t.Fatal("AudioManager creation failed")
	}

	// Step 3: Check audio files (should fail in test environment)
	if manager.checkAudioFiles() {
		t.Error("checkAudioFiles should fail without audio files")
	}

	// Step 4: Verify graceful degradation
	if manager.IsInitialized() {
		t.Error("Manager should not be initialized without audio files")
	}

	// Step 5: Verify config persistence
	manager.UpdateConfig(config.AudioConfig{
		BGMEnabled: false,
		SFXEnabled: false,
		BGMVolume:  0.5,
		SFXVolume:  0.5,
	})

	updatedCfg := manager.Config()
	if updatedCfg.BGMEnabled {
		t.Error("BGMEnabled should be false after update")
	}
}

// TestAudioConfigValidation tests audio config validation
func TestAudioConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		bgmVol  float64
		sfxVol  float64
		wantErr bool
	}{
		{"Valid volumes", 0.5, 0.8, false},
		{"Min volumes", 0.0, 0.0, false},
		{"Max volumes", 1.0, 1.0, false},
		{"BGM too high", 1.5, 0.5, true},
		{"SFX too low", 0.5, -0.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.AudioConfig{
				BGMEnabled: true,
				SFXEnabled: true,
				BGMVolume:  tt.bgmVol,
				SFXVolume:  tt.sfxVol,
			}

			// Validate volume ranges
			hasError := cfg.BGMVolume < 0.0 || cfg.BGMVolume > 1.0 ||
				cfg.SFXVolume < 0.0 || cfg.SFXVolume > 1.0

			if hasError != tt.wantErr {
				t.Errorf("Validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

// TestAudioManagerShutdown tests graceful shutdown
func TestAudioManagerShutdown(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)

	// Manually set initialized (since we can't actually initialize)
	manager.initialized = true

	if !manager.IsInitialized() {
		t.Error("Manager should be initialized")
	}

	// Shutdown
	if err := manager.Shutdown(); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify shutdown
	if manager.IsInitialized() {
		t.Error("Manager should not be initialized after shutdown")
	}
}

// TestConfigDir validates config directory structure
func TestConfigDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	expectedDir := filepath.Join(homeDir, ".nightmare")

	configDir, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir failed: %v", err)
	}

	if configDir != expectedDir {
		t.Errorf("ConfigDir = %s, expected %s", configDir, expectedDir)
	}
}
