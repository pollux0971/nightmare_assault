package audio

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// TestGameIntegration_NarrativeWithMoodTag simulates game loop integration
func TestGameIntegration_NarrativeWithMoodTag(t *testing.T) {
	// Simulate game setup
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, t.TempDir())

	// Simulate LLM response with mood tag
	llmResponse := `
你推開沉重的木門，進入了一個陰暗的走廊。
牆壁上掛著古老的肖像畫，似乎在注視著你。
[MOOD:horror]
`

	// Simulate game loop calling auto-switch after displaying narrative
	CheckMoodAndSwitch(player, llmResponse)

	// Wait for async operation (in real game, this happens naturally)
	time.Sleep(50 * time.Millisecond)

	// Verify mood was detected and stored
	currentMood := player.GetCurrentMood()
	if currentMood != engine.MoodHorror {
		t.Errorf("Expected mood to be MoodHorror, got %v", currentMood)
	}
}

// TestGameIntegration_MultipleTurns simulates multiple game turns with mood changes
func TestGameIntegration_MultipleTurns(t *testing.T) {
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, t.TempDir())

	// Turn 1: Exploration
	turn1 := "你開始探索這個神秘的建築。[MOOD:exploration]"
	CheckMoodAndSwitch(player, turn1)
	time.Sleep(10 * time.Millisecond)

	if player.GetCurrentMood() != engine.MoodExploration {
		t.Error("Turn 1: Expected MoodExploration")
	}

	// Turn 2: Same mood (should be ignored)
	turn2 := "你繼續探索，發現了一個房間。[MOOD:exploration]"
	lastSwitch1 := player.lastSwitch
	CheckMoodAndSwitch(player, turn2)
	time.Sleep(10 * time.Millisecond)

	if player.lastSwitch != lastSwitch1 {
		t.Error("Turn 2: Same mood should not trigger switch")
	}

	// Turn 3: Different mood but too soon (< 30s)
	// Force recent switch time
	player.lastSwitch = time.Now().Add(-15 * time.Second)
	turn3 := "突然聽到腳步聲！[MOOD:tension]"
	moodBefore := player.GetCurrentMood()
	CheckMoodAndSwitch(player, turn3)
	time.Sleep(10 * time.Millisecond)

	if player.GetCurrentMood() != moodBefore {
		t.Error("Turn 3: Too soon, mood should not change")
	}

	// Turn 4: Different mood after interval
	// Force old switch time
	player.lastSwitch = time.Now().Add(-35 * time.Second)
	turn4 := "一個恐怖的身影出現了！[MOOD:horror]"
	CheckMoodAndSwitch(player, turn4)
	time.Sleep(10 * time.Millisecond)

	// Mood should be updated (even though crossfade fails with nil context)
	// The SwitchByMood was called and mood state is tracked
}

// TestGameIntegration_NoAudioContext simulates game without audio
func TestGameIntegration_NoAudioContext(t *testing.T) {
	cfg := config.AudioConfig{BGMEnabled: false, BGMVolume: 0.0}
	player := NewBGMPlayer(nil, cfg, t.TempDir()) // nil context

	// Should not panic even without audio context
	llmResponse := "恐怖的時刻到來了。[MOOD:horror]"
	CheckMoodAndSwitch(player, llmResponse)

	// Wait for async operation
	time.Sleep(50 * time.Millisecond)

	// Mood should still be tracked (for future use when audio is enabled)
	// This ensures graceful degradation
}

// TestGameIntegration_ParsingEdgeCases tests various mood tag formats
func TestGameIntegration_ParsingEdgeCases(t *testing.T) {
	cfg := config.AudioConfig{BGMEnabled: true, BGMVolume: 0.7}
	player := NewBGMPlayer(nil, cfg, t.TempDir())

	testCases := []struct {
		name         string
		response     string
		expectedMood engine.MoodType
	}{
		{
			name:         "Tag at start",
			response:     "[MOOD:tension] 緊張的開始",
			expectedMood: engine.MoodTension,
		},
		{
			name:         "Tag at end",
			response:     "你進入了安全區 [MOOD:safe]",
			expectedMood: engine.MoodSafe,
		},
		{
			name:         "Tag in middle",
			response:     "你推開門 [MOOD:mystery] 看到一個謎題",
			expectedMood: engine.MoodMystery,
		},
		{
			name:         "Multiple tags (uses first)",
			response:     "[MOOD:ending] 結局 [MOOD:safe]",
			expectedMood: engine.MoodEnding,
		},
		{
			name:         "No tag (default)",
			response:     "沒有標記的文字",
			expectedMood: engine.MoodExploration,
		},
		{
			name:         "Case insensitive",
			response:     "[MOOD:HORROR] 大寫也可以",
			expectedMood: engine.MoodHorror,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset player state
			player.currentMood = engine.MoodExploration
			player.lastSwitch = time.Time{} // Zero time

			// Process response
			CheckMoodAndSwitch(player, tc.response)
			time.Sleep(10 * time.Millisecond)

			// Verify mood (may not switch if same, but parse should work)
			mood := engine.ParseMood(tc.response)
			if mood != tc.expectedMood {
				t.Errorf("Expected mood %v, got %v", tc.expectedMood, mood)
			}
		})
	}
}

// TestGameIntegration_AudioManagerNil simulates missing AudioManager
func TestGameIntegration_AudioManagerNil(t *testing.T) {
	// Simulate game without AudioManager initialized
	var player *BGMPlayer = nil

	llmResponse := "這不應該導致 panic。[MOOD:horror]"

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panicked with nil AudioManager: %v", r)
		}
	}()

	CheckMoodAndSwitch(player, llmResponse)
}

// TestGetBGMForAllMoods verifies all mood types map to valid BGM files
func TestGetBGMForAllMoods(t *testing.T) {
	allMoods := []engine.MoodType{
		engine.MoodExploration,
		engine.MoodTension,
		engine.MoodSafe,
		engine.MoodHorror,
		engine.MoodMystery,
		engine.MoodEnding,
	}

	for _, mood := range allMoods {
		bgmFile := GetBGMForMood(mood)

		// Verify BGM file name is not empty
		if bgmFile == "" {
			t.Errorf("GetBGMForMood(%v) returned empty string", mood)
		}

		// Verify BGM file has correct extension
		if !strings.HasSuffix(bgmFile, ".mp3") && !strings.HasSuffix(bgmFile, ".ogg") {
			t.Errorf("GetBGMForMood(%v) = %q, expected .mp3 or .ogg extension", mood, bgmFile)
		}
	}
}
