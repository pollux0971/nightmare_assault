package audio

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TestNewBGMPlayer tests BGMPlayer creation
func TestNewBGMPlayer(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		BGMVolume:  0.7,
		SFXEnabled: true,
		SFXVolume:  0.8,
	}

	audioDir := t.TempDir()
	player := NewBGMPlayer(nil, cfg, audioDir)

	if player == nil {
		t.Fatal("NewBGMPlayer should not return nil")
	}

	if player.volume != 0.7 {
		t.Errorf("Volume = %f, expected 0.7", player.volume)
	}

	if !player.enabled {
		t.Error("Player should be enabled")
	}
}

// TestBGMPlayer_Enable tests enable/disable functionality
func TestBGMPlayer_Enable(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: false}, t.TempDir())

	if player.IsEnabled() {
		t.Error("Player should be disabled initially")
	}

	player.Enable()
	if !player.IsEnabled() {
		t.Error("Player should be enabled after Enable()")
	}

	player.Disable()
	if player.IsEnabled() {
		t.Error("Player should be disabled after Disable()")
	}
}

// TestBGMPlayer_SetVolume tests volume control
func TestBGMPlayer_SetVolume(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMVolume: 0.5}, t.TempDir())

	tests := []struct {
		name     string
		volume   float64
		expected float64
	}{
		{"Valid min", 0.0, 0.0},
		{"Valid mid", 0.5, 0.5},
		{"Valid max", 1.0, 1.0},
		{"Below min", -0.1, 0.0}, // Should clamp to 0.0
		{"Above max", 1.5, 1.0},  // Should clamp to 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player.SetVolume(tt.volume)
			actual := player.Volume()
			if actual != tt.expected {
				t.Errorf("Volume = %f, expected %f", actual, tt.expected)
			}
		})
	}
}

// TestBGMPlayer_CurrentBGM tests current BGM tracking
func TestBGMPlayer_CurrentBGM(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{}, t.TempDir())

	if player.CurrentBGM() != "" {
		t.Error("CurrentBGM should be empty initially")
	}
}

// TestBGMPlayer_IsPlaying tests playing state
func TestBGMPlayer_IsPlaying(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{}, t.TempDir())

	if player.IsPlaying() {
		t.Error("Player should not be playing initially")
	}
}

// TestBGMPlayer_Stop tests stop functionality
func TestBGMPlayer_Stop(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{}, t.TempDir())

	// Stop should not panic even if nothing is playing
	player.Stop()

	if player.IsPlaying() {
		t.Error("Player should not be playing after Stop()")
	}
}

// TestBGMPlayer_LastSwitchTime tests last switch time tracking
func TestBGMPlayer_LastSwitchTime(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{}, t.TempDir())

	lastSwitch := player.LastSwitchTime()
	// Should be zero time initially (never switched)
	if !lastSwitch.IsZero() {
		t.Error("LastSwitchTime should be zero initially (never switched)")
	}
}

// TestBGMSceneMapping tests BGM scene type mapping
func TestBGMSceneMapping(t *testing.T) {
	tests := []struct {
		scene    BGMScene
		filename string
	}{
		{BGMSceneExploration, "ambient_exploration"},
		{BGMSceneChase, "tension_chase"},
		{BGMSceneSafe, "safe_rest"},
		{BGMSceneHorror, "horror_reveal"},
		{BGMSceneMystery, "mystery_puzzle"},
		{BGMSceneDeath, "ending_death"},
	}

	for _, tt := range tests {
		t.Run(string(tt.scene), func(t *testing.T) {
			filename := GetBGMFilename(tt.scene)
			if !containsStr(filename, tt.filename) {
				t.Errorf("GetBGMFilename(%s) = %s, should contain %s", tt.scene, filename, tt.filename)
			}
		})
	}
}

// TestBGMPlayer_PlayWithInvalidFile tests error handling
func TestBGMPlayer_PlayWithInvalidFile(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true}, t.TempDir())

	err := player.Play("nonexistent.mp3")
	if err == nil {
		t.Error("Play should return error for non-existent file")
	}
}

// TestBGMPlayer_PlayWithValidFile tests playing with a valid file
func TestBGMPlayer_PlayWithValidFile(t *testing.T) {
	t.Skip("Skipping actual audio playback test - requires real oto context")

	// This test would require:
	// 1. Real oto context (can't create in test environment)
	// 2. Valid audio files
	// 3. Platform audio device availability
	//
	// Integration tests with real files will be done manually
}

// TestBGMPlayer_LoopMode tests loop playback mode
func TestBGMPlayer_LoopMode(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{}, t.TempDir())

	// Loop should be enabled by default
	if !player.IsLoopEnabled() {
		t.Error("Loop should be enabled by default")
	}

	player.SetLoop(false)
	if player.IsLoopEnabled() {
		t.Error("Loop should be disabled after SetLoop(false)")
	}

	player.SetLoop(true)
	if !player.IsLoopEnabled() {
		t.Error("Loop should be enabled after SetLoop(true)")
	}
}

// TestBGMPlayer_FadeOut tests fade out functionality
func TestBGMPlayer_FadeOut(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.8}, t.TempDir())

	// Fade out should complete without error even if not playing
	err := player.FadeOut(100 * time.Millisecond)
	if err != nil {
		t.Errorf("FadeOut failed: %v", err)
	}

	// After fade out, volume should be 0.0
	if player.Volume() != 0.0 {
		t.Errorf("Volume after FadeOut = %f, expected 0.0", player.Volume())
	}
}

// TestBGMPlayer_FadeIn tests fade in functionality
func TestBGMPlayer_FadeIn(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.0}, t.TempDir())

	// Set target volume
	targetVolume := 0.7
	player.SetTargetVolume(targetVolume)

	// Fade in should complete without error
	err := player.FadeIn(100 * time.Millisecond)
	if err != nil {
		t.Errorf("FadeIn failed: %v", err)
	}

	// After fade in, volume should reach target
	actualVolume := player.Volume()
	if actualVolume < targetVolume-0.01 || actualVolume > targetVolume+0.01 {
		t.Errorf("Volume after FadeIn = %f, expected ~%f", actualVolume, targetVolume)
	}
}

// TestBGMPlayer_Crossfade tests crossfade functionality
func TestBGMPlayer_Crossfade(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.8}, t.TempDir())

	// Crossfade should handle nil context gracefully
	err := player.Crossfade("new_bgm.mp3", 200*time.Millisecond)
	if err == nil {
		t.Error("Crossfade should return error with nil context")
	}
}

// TestBGMPlayer_FadeCurve tests fade curve calculation
func TestBGMPlayer_FadeCurve(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		expected float64 // Approximate expected value
	}{
		{"Start", 0.0, 0.0},
		{"Quarter", 0.25, 0.25},
		{"Half", 0.5, 0.5},
		{"ThreeQuarters", 0.75, 0.75},
		{"End", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linearFadeCurve(tt.progress)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("linearFadeCurve(%f) = %f, expected ~%f", tt.progress, result, tt.expected)
			}
		})
	}
}

// Helper function for string checking
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
