package commands

import (
	"fmt"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// StatusCommand shows player status.
type StatusCommand struct {
	stats         *game.PlayerStats
	turn          int
	gameState     *engine.GameStateV2
	gameStartTime time.Time
}

// NewStatusCommand creates a new status command.
func NewStatusCommand(stats *game.PlayerStats, turn int) *StatusCommand {
	return &StatusCommand{
		stats:         stats,
		turn:          turn,
		gameStartTime: time.Now(), // Default to current time if not set
	}
}

// NewStatusCommandV2 creates a new status command with GameStateV2 integration.
func NewStatusCommandV2(stats *game.PlayerStats, turn int, gameState *engine.GameStateV2, startTime time.Time) *StatusCommand {
	return &StatusCommand{
		stats:         stats,
		turn:          turn,
		gameState:     gameState,
		gameStartTime: startTime,
	}
}

// Name returns the command name.
func (c *StatusCommand) Name() string {
	return "status"
}

// Execute displays player status.
func (c *StatusCommand) Execute(args []string) (string, error) {
	hpBar := makeBar(c.stats.HP, 100, 20)
	sanBar := makeBar(c.stats.SAN, 100, 20)
	sanState := c.stats.State.String()

	// Calculate playtime
	playtime := time.Since(c.gameStartTime)
	hours := int(playtime.Hours())
	minutes := int(playtime.Minutes()) % 60
	seconds := int(playtime.Seconds()) % 60
	playtimeStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	// Get location and chapter from GameStateV2 if available
	location := "Unknown"
	chapter := "Beat 0"
	if c.gameState != nil {
		if c.gameState.CurrentScene != "" {
			location = c.gameState.CurrentScene
		}
		beat := c.gameState.GetCurrentBeat()
		chapter = fmt.Sprintf("Beat %d", beat)
	}

	status := fmt.Sprintf(`
═══════════════════════════════════════════════════
              PLAYER STATUS
═══════════════════════════════════════════════════

HP:  %s %3d/100
SAN: %s %3d/100

Sanity State: %s
Current Location: %s
Current Chapter: %s
Turn Count: %d
Playtime: %s

═══════════════════════════════════════════════════
Press any key to return to game...
`, hpBar, c.stats.HP, sanBar, c.stats.SAN, sanState, location, chapter, c.turn, playtimeStr)

	return status, nil
}

// Help returns brief command description.
func (c *StatusCommand) Help() string {
	return "View detailed player status (HP, SAN, location, chapter, playtime)"
}

// makeBar creates a simple progress bar.
func makeBar(value, max, width int) string {
	filled := int(float64(value) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
