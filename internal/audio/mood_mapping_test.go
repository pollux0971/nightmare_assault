package audio

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

func TestGetBGMForMood_AllMoodTypes(t *testing.T) {
	tests := []struct {
		mood     engine.MoodType
		expected string
	}{
		{engine.MoodExploration, "ambient_exploration.mp3"},
		{engine.MoodTension, "tension_chase.mp3"},
		{engine.MoodSafe, "safe_rest.mp3"},
		{engine.MoodHorror, "horror_reveal.mp3"},
		{engine.MoodMystery, "mystery_puzzle.mp3"},
		{engine.MoodEnding, "ending_death.mp3"},
	}

	for _, tt := range tests {
		result := GetBGMForMood(tt.mood)
		if result != tt.expected {
			t.Errorf("GetBGMForMood(%v) = %q, expected %q", tt.mood, result, tt.expected)
		}
	}
}

func TestGetBGMForMood_MatchesBGMScene(t *testing.T) {
	// Verify mapping is consistent with existing BGMScene constants
	tests := []struct {
		mood     engine.MoodType
		bgmScene BGMScene
	}{
		{engine.MoodExploration, BGMSceneExploration},
		{engine.MoodTension, BGMSceneChase},
		{engine.MoodSafe, BGMSceneSafe},
		{engine.MoodHorror, BGMSceneHorror},
		{engine.MoodMystery, BGMSceneMystery},
		{engine.MoodEnding, BGMSceneDeath},
	}

	for _, tt := range tests {
		moodBGM := GetBGMForMood(tt.mood)
		sceneBGM := GetBGMFilename(tt.bgmScene)

		if moodBGM != sceneBGM {
			t.Errorf("Inconsistent mapping: mood %v -> %q, scene %v -> %q",
				tt.mood, moodBGM, tt.bgmScene, sceneBGM)
		}
	}
}

func TestMoodToBGMScene_AllMoodTypes(t *testing.T) {
	tests := []struct {
		mood     engine.MoodType
		expected BGMScene
	}{
		{engine.MoodExploration, BGMSceneExploration},
		{engine.MoodTension, BGMSceneChase},
		{engine.MoodSafe, BGMSceneSafe},
		{engine.MoodHorror, BGMSceneHorror},
		{engine.MoodMystery, BGMSceneMystery},
		{engine.MoodEnding, BGMSceneDeath},
	}

	for _, tt := range tests {
		result := MoodToBGMScene(tt.mood)
		if result != tt.expected {
			t.Errorf("MoodToBGMScene(%v) = %v, expected %v", tt.mood, result, tt.expected)
		}
	}
}
