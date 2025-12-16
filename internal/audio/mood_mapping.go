package audio

import "github.com/nightmare-assault/nightmare-assault/internal/engine"

// GetBGMForMood returns the BGM filename for a given mood type
func GetBGMForMood(mood engine.MoodType) string {
	scene := MoodToBGMScene(mood)
	return GetBGMFilename(scene)
}

// MoodToBGMScene converts a MoodType to corresponding BGMScene
func MoodToBGMScene(mood engine.MoodType) BGMScene {
	switch mood {
	case engine.MoodExploration:
		return BGMSceneExploration
	case engine.MoodTension:
		return BGMSceneTension
	case engine.MoodChase:
		return BGMSceneChase
	case engine.MoodSafe:
		return BGMSceneSafe
	case engine.MoodHorror:
		return BGMSceneHorror
	case engine.MoodMystery:
		return BGMSceneMystery
	case engine.MoodDream:
		return BGMSceneDream
	case engine.MoodSanity:
		return BGMSceneSanity
	case engine.MoodRitual:
		return BGMSceneRitual
	case engine.MoodEnding:
		return BGMSceneDeath
	default:
		return BGMSceneExploration
	}
}
