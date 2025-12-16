package engine

import (
	"regexp"
	"strings"
)

// MoodType represents the mood/atmosphere of a game scene
type MoodType int

const (
	// MoodExploration represents exploration scenes (default)
	MoodExploration MoodType = iota
	// MoodTension represents tense scenes
	MoodTension
	// MoodChase represents intense chase/escape scenes
	MoodChase
	// MoodSafe represents safe/rest areas
	MoodSafe
	// MoodHorror represents horror revelation moments
	MoodHorror
	// MoodMystery represents puzzle/mystery scenes
	MoodMystery
	// MoodDream represents dream/surreal sequences
	MoodDream
	// MoodSanity represents sanity collapse/madness
	MoodSanity
	// MoodRitual represents ritual/occult scenes
	MoodRitual
	// MoodEnding represents death/ending scenes
	MoodEnding
)

// String returns the string representation of MoodType
func (m MoodType) String() string {
	switch m {
	case MoodExploration:
		return "exploration"
	case MoodTension:
		return "tension"
	case MoodChase:
		return "chase"
	case MoodSafe:
		return "safe"
	case MoodHorror:
		return "horror"
	case MoodMystery:
		return "mystery"
	case MoodDream:
		return "dream"
	case MoodSanity:
		return "sanity"
	case MoodRitual:
		return "ritual"
	case MoodEnding:
		return "ending"
	default:
		return "exploration"
	}
}

// ParseMood extracts mood tag from LLM response text
// Format: [MOOD:xxx] where xxx is one of: exploration, tension, chase, safe, horror, mystery, dream, sanity, ritual, ending
// Returns MoodExploration as default if no valid tag found
func ParseMood(text string) MoodType {
	// Regex to match [MOOD:xxx]
	re := regexp.MustCompile(`\[MOOD:(\w+)\]`)
	matches := re.FindStringSubmatch(text)

	// No mood tag found, return default
	if len(matches) < 2 {
		return MoodExploration
	}

	// Parse mood type (case-insensitive)
	moodStr := strings.ToLower(matches[1])
	switch moodStr {
	case "tension":
		return MoodTension
	case "chase":
		return MoodChase
	case "safe":
		return MoodSafe
	case "horror":
		return MoodHorror
	case "mystery":
		return MoodMystery
	case "dream":
		return MoodDream
	case "sanity":
		return MoodSanity
	case "ritual":
		return MoodRitual
	case "ending":
		return MoodEnding
	case "exploration":
		return MoodExploration
	default:
		return MoodExploration
	}
}
