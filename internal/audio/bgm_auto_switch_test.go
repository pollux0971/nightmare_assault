package audio

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

func TestBGMPlayer_CanSwitch_SameMood(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Set current mood to exploration
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-1 * time.Minute) // Long enough ago

	// Try to switch to same mood
	canSwitch := player.CanSwitch(engine.MoodExploration)
	if canSwitch {
		t.Error("CanSwitch should return false for same mood")
	}
}

func TestBGMPlayer_CanSwitch_MinimumInterval(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Set current mood and recent switch time
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-15 * time.Second) // Only 15 seconds ago

	// Try to switch to different mood (but too soon)
	canSwitch := player.CanSwitch(engine.MoodTension)
	if canSwitch {
		t.Error("CanSwitch should return false within 30 second interval")
	}
}

func TestBGMPlayer_CanSwitch_AfterMinimumInterval(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Set current mood and switch time > 30 seconds ago
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-35 * time.Second)

	// Try to switch to different mood (should be allowed)
	canSwitch := player.CanSwitch(engine.MoodTension)
	if !canSwitch {
		t.Error("CanSwitch should return true after 30 second interval")
	}
}

func TestBGMPlayer_CanSwitch_FirstSwitch(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// lastSwitch is zero time (never switched before)
	// currentMood is default (exploration)

	// Try to switch to different mood (should be allowed)
	canSwitch := player.CanSwitch(engine.MoodTension)
	if !canSwitch {
		t.Error("CanSwitch should return true for first switch")
	}
}

func TestBGMPlayer_SwitchByMood_IgnoresSameMood(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Set current mood
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-1 * time.Minute)

	// Track last switch time
	lastSwitchBefore := player.lastSwitch

	// Try to switch to same mood (should be silently ignored)
	err := player.SwitchByMood(engine.MoodExploration)
	if err != nil {
		t.Errorf("SwitchByMood should not return error for same mood, got: %v", err)
	}

	// lastSwitch should NOT be updated
	if player.lastSwitch != lastSwitchBefore {
		t.Error("lastSwitch should not be updated when switch is ignored")
	}
}

func TestBGMPlayer_SwitchByMood_IgnoresTooSoon(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Set current mood and recent switch
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-10 * time.Second)
	lastSwitchBefore := player.lastSwitch

	// Try to switch too soon (should be silently ignored)
	err := player.SwitchByMood(engine.MoodTension)
	if err != nil {
		t.Errorf("SwitchByMood should not return error when too soon, got: %v", err)
	}

	// lastSwitch should NOT be updated
	if player.lastSwitch != lastSwitchBefore {
		t.Error("lastSwitch should not be updated when switch is rejected")
	}
}

func TestBGMPlayer_SwitchByMood_NilContext(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Set conditions that allow switch
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-1 * time.Minute)

	// Try to switch with nil context (Crossfade will fail gracefully)
	err := player.SwitchByMood(engine.MoodTension)

	// Should not panic, may return error from Crossfade
	if err == nil {
		t.Log("SwitchByMood with nil context returned no error (silent failure expected)")
	} else {
		t.Logf("SwitchByMood with nil context returned error: %v (expected)", err)
	}
}

func TestBGMPlayer_GetCurrentMood(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Default mood should be exploration
	mood := player.GetCurrentMood()
	if mood != engine.MoodExploration {
		t.Errorf("GetCurrentMood() = %v, expected MoodExploration", mood)
	}

	// Set a different mood
	player.currentMood = engine.MoodHorror

	mood = player.GetCurrentMood()
	if mood != engine.MoodHorror {
		t.Errorf("GetCurrentMood() = %v, expected MoodHorror", mood)
	}
}

func TestBGMPlayer_CurrentMoodInitialization(t *testing.T) {
	player := NewBGMPlayer(nil, config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}, t.TempDir())

	// Verify default initialization
	if player.currentMood != engine.MoodExploration {
		t.Errorf("currentMood should initialize to MoodExploration, got %v", player.currentMood)
	}

	// Verify zero time for lastSwitch
	if !player.lastSwitch.IsZero() {
		t.Error("lastSwitch should be zero time initially")
	}
}
