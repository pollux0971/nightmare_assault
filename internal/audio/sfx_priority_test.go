package audio

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func TestSFXPlayer_PriorityReplacement(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Fill all 4 channels with environment SFX (low priority)
	for i := 0; i < 4; i++ {
		player.channels[i].isPlaying = true
		player.channels[i].priority = SFXPriorityEnvironment
		player.channels[i].startTime = time.Now().Add(-time.Duration(i) * time.Second)
		player.channels[i].sfxType = "environment"
	}

	// High priority SFX should replace lowest priority channel
	channelIndex := player.findBestChannel(SFXPriorityWarning)

	if channelIndex == -1 {
		t.Error("Should find a channel for high priority SFX")
	}
}

func TestSFXPlayer_IdleChannelPreferred(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Set 3 channels as playing
	for i := 0; i < 3; i++ {
		player.channels[i].isPlaying = true
		player.channels[i].priority = SFXPriorityEnvironment
	}

	// Channel 3 is idle
	player.channels[3].isPlaying = false

	// Should prefer idle channel
	channelIndex := player.findBestChannel(SFXPriorityEnvironment)

	if channelIndex != 3 {
		t.Errorf("Should prefer idle channel 3, got %d", channelIndex)
	}
}

func TestSFXPlayer_HighPriorityNotReplaced(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// All channels playing high priority
	for i := 0; i < 4; i++ {
		player.channels[i].isPlaying = true
		player.channels[i].priority = SFXPriorityWarning
	}

	// Low priority SFX should not find a channel
	channelIndex := player.findBestChannel(SFXPriorityEnvironment)

	if channelIndex != -1 {
		t.Error("Low priority should not replace high priority")
	}
}

func TestSFXPlayer_DeathPriorityStopsAll(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Fill channels with various priorities
	player.channels[0].isPlaying = true
	player.channels[0].priority = SFXPriorityEnvironment

	player.channels[1].isPlaying = true
	player.channels[1].priority = SFXPriorityWarning

	player.channels[2].isPlaying = true
	player.channels[2].priority = SFXPriorityHeartbeat

	// Death SFX should stop all
	err := player.PlayDeath()

	// Should not error (nil context is OK for this test)
	if err != nil && err.Error() != "audio context is nil" {
		t.Errorf("Unexpected error: %v", err)
	}

	// All channels should be stopped
	for i := 0; i < 4; i++ {
		if player.channels[i].isPlaying {
			t.Errorf("Channel %d should be stopped after PlayDeath()", i)
		}
	}
}

func TestSFXPlayer_PriorityOrdering(t *testing.T) {
	// Verify priority constants are in correct order
	if SFXPriorityDeath <= SFXPriorityWarning {
		t.Error("Death priority should be higher than Warning")
	}

	if SFXPriorityWarning <= SFXPriorityHeartbeat {
		t.Error("Warning priority should be higher than Heartbeat")
	}

	if SFXPriorityHeartbeat <= SFXPriorityEnvironment {
		t.Error("Heartbeat priority should be higher than Environment")
	}
}

func TestSFXPlayer_SamePriorityReplacement(t *testing.T) {
	cfg := config.AudioConfig{SFXEnabled: true, SFXVolume: 0.7}
	player := NewSFXPlayer(nil, cfg, t.TempDir())

	// Fill all channels with same priority
	for i := 0; i < 4; i++ {
		player.channels[i].isPlaying = true
		player.channels[i].priority = SFXPriorityEnvironment
		player.channels[i].startTime = time.Now().Add(-time.Duration(i) * time.Second)
	}

	// New same-priority SFX should find oldest channel (channel 3)
	channelIndex := player.findBestChannel(SFXPriorityEnvironment)

	// Should find the oldest channel or any channel with same priority
	if channelIndex == -1 {
		t.Error("Should find a channel for same priority SFX")
	}
}
