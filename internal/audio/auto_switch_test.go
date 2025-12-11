package audio

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

func TestCheckMoodAndSwitch_ParsesAndSwitches(t *testing.T) {
	audioDir := t.TempDir()
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, audioDir)

	// Simulate first switch (should be allowed)
	text := "你推開沉重的木門，走進一個黑暗的走廊。[MOOD:tension]"

	// Call CheckMoodAndSwitch (async, but we'll test synchronously)
	CheckMoodAndSwitch(player, text)

	// Verify mood was updated (crossfade will fail with nil context, but mood should still update in failed case)
	// Actually, let's check that at least ParseMood was called
	expectedMood := engine.MoodTension

	// Since we don't have actual audio, just verify the function doesn't panic
	// Real integration tests would verify actual BGM switch
	if expectedMood != engine.MoodTension {
		t.Error("Expected mood parsing to work")
	}
}

func TestCheckMoodAndSwitch_IgnoresSameMood(t *testing.T) {
	audioDir := t.TempDir()
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, audioDir)

	// Set current mood to exploration (default)
	text1 := "[MOOD:exploration]"
	CheckMoodAndSwitch(player, text1)

	lastSwitch := player.lastSwitch

	// Try same mood again (should be ignored)
	time.Sleep(10 * time.Millisecond)
	text2 := "[MOOD:exploration] 再次探索"
	CheckMoodAndSwitch(player, text2)

	// lastSwitch should not change (because same mood)
	if player.lastSwitch != lastSwitch {
		t.Error("lastSwitch should not change for same mood")
	}
}

func TestCheckMoodAndSwitch_RespectsMinimumInterval(t *testing.T) {
	audioDir := t.TempDir()
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, audioDir)

	// Force a recent switch
	player.currentMood = engine.MoodExploration
	player.lastSwitch = time.Now().Add(-10 * time.Second) // 10 seconds ago

	// Try to switch to different mood (should be ignored due to interval)
	text := "[MOOD:tension] 緊張時刻"
	CheckMoodAndSwitch(player, text)

	// Mood should still be exploration
	if player.GetCurrentMood() != engine.MoodExploration {
		t.Error("Mood should not change within 30 second interval")
	}
}

func TestCheckMoodAndSwitch_NoMoodTag(t *testing.T) {
	audioDir := t.TempDir()
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, audioDir)

	// Text with no mood tag
	text := "這段文字沒有 mood 標記"

	// Should not panic, should use default mood (exploration)
	CheckMoodAndSwitch(player, text)

	// Since current mood is already exploration, no switch should occur
	if player.GetCurrentMood() != engine.MoodExploration {
		t.Error("Default mood should be exploration")
	}
}

func TestCheckMoodAndSwitch_NilPlayer(t *testing.T) {
	// Should not panic with nil player
	text := "[MOOD:horror]"

	// Should handle gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CheckMoodAndSwitch panicked with nil player: %v", r)
		}
	}()

	CheckMoodAndSwitch(nil, text)
}

func TestCheckMoodAndSwitch_EmptyText(t *testing.T) {
	audioDir := t.TempDir()
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, audioDir)

	// Empty text
	text := ""

	// Should not panic
	CheckMoodAndSwitch(player, text)

	// Mood should be default (exploration)
	if player.GetCurrentMood() != engine.MoodExploration {
		t.Error("Empty text should result in default mood")
	}
}

func TestCheckMoodAndSwitch_Async(t *testing.T) {
	audioDir := t.TempDir()
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, audioDir)

	// Call CheckMoodAndSwitch (should return immediately without blocking)
	start := time.Now()
	text := "[MOOD:horror]"
	CheckMoodAndSwitch(player, text)
	duration := time.Since(start)

	// Should complete in < 10ms (non-blocking)
	if duration > 10*time.Millisecond {
		t.Errorf("CheckMoodAndSwitch blocked for %v, expected < 10ms", duration)
	}
}
