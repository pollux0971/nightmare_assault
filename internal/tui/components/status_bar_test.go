package components

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestNewStatusBar(t *testing.T) {
	stats := game.NewPlayerStats()
	bar := NewStatusBar(stats, 1)

	if bar == nil {
		t.Fatal("NewStatusBar should not return nil")
	}
	if bar.stats != stats {
		t.Error("StatusBar should reference the provided stats")
	}
	if bar.turnCount != 1 {
		t.Errorf("Turn count = %d, want 1", bar.turnCount)
	}
}

func TestStatusBar_SetTurnCount(t *testing.T) {
	bar := NewStatusBar(game.NewPlayerStats(), 1)

	bar.SetTurnCount(5)

	if bar.turnCount != 5 {
		t.Errorf("After SetTurnCount(5), turnCount = %d, want 5", bar.turnCount)
	}
}

func TestStatusBar_SetGameMode(t *testing.T) {
	bar := NewStatusBar(game.NewPlayerStats(), 1)

	bar.SetGameMode("Playing")

	if bar.gameMode != "Playing" {
		t.Errorf("GameMode = %q, want 'Playing'", bar.gameMode)
	}
}

func TestStatusBar_UpdateStats(t *testing.T) {
	stats1 := game.NewPlayerStats()
	bar := NewStatusBar(stats1, 1)

	stats2 := &game.PlayerStats{HP: 50, SAN: 60}
	bar.UpdateStats(stats2)

	if bar.stats.HP != 50 {
		t.Errorf("After update, HP = %d, want 50", bar.stats.HP)
	}
	if bar.stats.SAN != 60 {
		t.Errorf("After update, SAN = %d, want 60", bar.stats.SAN)
	}
}

func TestStatusBar_GetSanityStateColor(t *testing.T) {
	tests := []struct {
		state    game.SanityState
		expected string
	}{
		{game.SanityClearHeaded, "10"}, // green
		{game.SanityAnxious, "11"},     // yellow
		{game.SanityPanicked, "9"},     // orange
		{game.SanityInsanity, "1"},     // red
	}

	bar := NewStatusBar(game.NewPlayerStats(), 1)

	for _, tt := range tests {
		color := bar.getSanityStateColor(tt.state)
		if color != tt.expected {
			t.Errorf("getSanityStateColor(%v) = %q, want %q", tt.state, color, tt.expected)
		}
	}
}

func TestStatusBar_GetHPBar(t *testing.T) {
	bar := NewStatusBar(game.NewPlayerStats(), 1)

	// Test full HP
	bar.stats.HP = 100
	hpBar := bar.getHPBar(20)
	if len(hpBar) == 0 {
		t.Error("HP bar should not be empty")
	}

	// Test half HP
	bar.stats.HP = 50
	hpBar = bar.getHPBar(20)
	if len(hpBar) == 0 {
		t.Error("HP bar should not be empty")
	}

	// Test low HP
	bar.stats.HP = 10
	hpBar = bar.getHPBar(20)
	if len(hpBar) == 0 {
		t.Error("HP bar should not be empty")
	}
}

func TestStatusBar_GetSANBar(t *testing.T) {
	bar := NewStatusBar(game.NewPlayerStats(), 1)

	// Test full SAN
	bar.stats.SAN = 100
	sanBar := bar.getSANBar(20)
	if len(sanBar) == 0 {
		t.Error("SAN bar should not be empty")
	}

	// Test low SAN
	bar.stats.SAN = 30
	sanBar = bar.getSANBar(20)
	if len(sanBar) == 0 {
		t.Error("SAN bar should not be empty")
	}
}

func TestStatusBar_View(t *testing.T) {
	stats := &game.PlayerStats{
		HP:    80,
		SAN:   70,
		State: game.SanityAnxious,
	}
	bar := NewStatusBar(stats, 5)
	bar.SetGameMode("Playing")

	view := bar.View()

	if len(view) == 0 {
		t.Error("View should not be empty")
	}
}

func TestStatusBar_SetWidth(t *testing.T) {
	bar := NewStatusBar(game.NewPlayerStats(), 1)

	bar.SetWidth(100)

	if bar.width != 100 {
		t.Errorf("Width = %d, want 100", bar.width)
	}
}

func TestStatusBar_Height(t *testing.T) {
	bar := NewStatusBar(game.NewPlayerStats(), 1)

	if bar.Height() != 3 {
		t.Errorf("Height = %d, want 3", bar.Height())
	}
}
