package audio

import (
	"log"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// CheckMoodAndSwitch checks LLM response text for mood tags and switches BGM accordingly
// Runs asynchronously using goroutine to avoid blocking game loop
// Silently handles errors (graceful degradation)
func CheckMoodAndSwitch(player *BGMPlayer, text string) {
	// Nil check
	if player == nil {
		return
	}

	// Run asynchronously
	go func() {
		// Parse mood from text
		mood := engine.ParseMood(text)

		// Attempt to switch BGM
		if err := player.SwitchByMood(mood); err != nil {
			log.Printf("[WARN] BGM auto-switch failed: %v", err)
			// Silent failure - don't interrupt game
		}
	}()
}
