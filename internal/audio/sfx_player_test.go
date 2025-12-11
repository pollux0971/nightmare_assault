package audio

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func TestNewSFXPlayer(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.8}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	if player == nil {
		t.Fatal("NewSFXPlayer returned nil")
	}

	if player.Volume() != 0.8 {
		t.Errorf("Volume() = %f, expected 0.8", player.Volume())
	}

	if !player.IsEnabled() {
		t.Error("IsEnabled() = false, expected true")
	}

	// Verify 4 channels initialized
	if len(player.channels) != 4 {
		t.Errorf("Expected 4 channels, got %d", len(player.channels))
	}
}

func TestSFXPlayer_Enable_Disable(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: false, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Initially disabled
	if player.IsEnabled() {
		t.Error("Player should be disabled initially")
	}

	// Enable
	player.Enable()
	if !player.IsEnabled() {
		t.Error("Player should be enabled after Enable()")
	}

	// Disable
	player.Disable()
	if player.IsEnabled() {
		t.Error("Player should be disabled after Disable()")
	}
}

func TestSFXPlayer_SetVolume(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.5}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Set new volume
	player.SetVolume(0.9)
	if player.Volume() != 0.9 {
		t.Errorf("Volume() = %f, expected 0.9", player.Volume())
	}

	// Clamp to 0.0
	player.SetVolume(-0.1)
	if player.Volume() != 0.0 {
		t.Errorf("Volume() = %f, expected 0.0 (clamped)", player.Volume())
	}

	// Clamp to 1.0
	player.SetVolume(1.5)
	if player.Volume() != 1.0 {
		t.Errorf("Volume() = %f, expected 1.0 (clamped)", player.Volume())
	}
}

func TestSFXPlayer_StopAll(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Mock some playing channels
	for i := 0; i < 4; i++ {
		player.channels[i].isPlaying = true
		player.channels[i].startTime = time.Now()
	}

	// Stop all
	player.StopAll()

	// Verify all stopped
	for i, ch := range player.channels {
		if ch.isPlaying {
			t.Errorf("Channel %d still playing after StopAll()", i)
		}
	}
}

func TestSFXPlayer_PlaySFX_NilContext(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Try to play with nil context (should not panic)
	err := player.PlaySFX("test.wav", SFXPriorityEnvironment)

	// Should return error gracefully
	if err == nil {
		t.Error("Expected error with nil audio context")
	}
}

func TestSFXPlayer_PlaySFX_Disabled(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: false, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Try to play when disabled (should silently skip)
	err := player.PlaySFX("test.wav", SFXPriorityEnvironment)

	// Should not return error (silent skip)
	if err != nil {
		t.Errorf("Expected nil error when disabled, got: %v", err)
	}
}

func TestSFXPriority_Constants(t *testing.T) {
	// Verify priority constants are defined correctly
	if SFXPriorityDeath != 100 {
		t.Errorf("SFXPriorityDeath = %d, expected 100", SFXPriorityDeath)
	}

	if SFXPriorityWarning != 80 {
		t.Errorf("SFXPriorityWarning = %d, expected 80", SFXPriorityWarning)
	}

	if SFXPriorityHeartbeat != 50 {
		t.Errorf("SFXPriorityHeartbeat = %d, expected 50", SFXPriorityHeartbeat)
	}

	if SFXPriorityEnvironment != 20 {
		t.Errorf("SFXPriorityEnvironment = %d, expected 20", SFXPriorityEnvironment)
	}
}

func TestSFXChannel_Initialization(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Verify all channels initialized
	for i, ch := range player.channels {
		if ch == nil {
			t.Errorf("Channel %d is nil", i)
		}

		if ch.isPlaying {
			t.Errorf("Channel %d should not be playing initially", i)
		}

		if ch.priority != 0 {
			t.Errorf("Channel %d priority = %d, expected 0", i, ch.priority)
		}
	}
}
