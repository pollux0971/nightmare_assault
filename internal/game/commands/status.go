package commands

import (
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// StatusCommand shows player status.
type StatusCommand struct {
	stats *game.PlayerStats
	turn  int
}

// NewStatusCommand creates a new status command.
func NewStatusCommand(stats *game.PlayerStats, turn int) *StatusCommand {
	return &StatusCommand{
		stats: stats,
		turn:  turn,
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

	status := fmt.Sprintf(`
═══════════════════════════════════════════════════
              PLAYER STATUS
═══════════════════════════════════════════════════

HP:  %s %3d/100
SAN: %s %3d/100

Sanity State: %s
Turn Count:   %d

═══════════════════════════════════════════════════
Press any key to return to game...
`, hpBar, c.stats.HP, sanBar, c.stats.SAN, sanState, c.turn)

	return status, nil
}

// Help returns brief command description.
func (c *StatusCommand) Help() string {
	return "View detailed player status"
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
