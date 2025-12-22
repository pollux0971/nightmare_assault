package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestStatusCommand_Name(t *testing.T) {
	stats := &game.PlayerStats{HP: 100, SAN: 100}
	cmd := NewStatusCommand(stats, 0)

	if cmd.Name() != "status" {
		t.Errorf("Expected name 'status', got '%s'", cmd.Name())
	}
}

func TestStatusCommand_Help(t *testing.T) {
	stats := &game.PlayerStats{HP: 100, SAN: 100}
	cmd := NewStatusCommand(stats, 0)

	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
}

func TestStatusCommand_Execute_BasicInfo(t *testing.T) {
	stats := &game.PlayerStats{
		HP:     75,
		SAN:    60,
		MaxHP:  100,
		MaxSAN: 100,
		State:  game.SanityAnxious,
	}
	cmd := NewStatusCommand(stats, 5)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check that output contains basic stats
	if !strings.Contains(output, "HP:") {
		t.Error("Output should contain HP information")
	}
	if !strings.Contains(output, "SAN:") {
		t.Error("Output should contain SAN information")
	}
	if !strings.Contains(output, "75/100") {
		t.Error("Output should contain HP value 75/100")
	}
	if !strings.Contains(output, "60/100") {
		t.Error("Output should contain SAN value 60/100")
	}
}

func TestStatusCommand_Execute_WithGameStateV2(t *testing.T) {
	stats := &game.PlayerStats{
		HP:     100,
		SAN:    100,
		MaxHP:  100,
		MaxSAN: 100,
		State:  game.SanityClearHeaded,
	}

	gameState := engine.NewGameStateV2()
	gameState.CurrentScene = "Abandoned Hospital"
	gameState.IncrementBeat()
	gameState.IncrementBeat()
	gameState.IncrementBeat()

	startTime := time.Now().Add(-5 * time.Minute) // Simulate 5 minutes of playtime
	cmd := NewStatusCommandV2(stats, 10, gameState, startTime)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check that output contains extended info
	if !strings.Contains(output, "Abandoned Hospital") {
		t.Error("Output should contain current location")
	}
	if !strings.Contains(output, "Beat") {
		t.Error("Output should contain beat information")
	}
	if !strings.Contains(output, "Playtime") {
		t.Error("Output should contain playtime")
	}
	if !strings.Contains(output, "Turn Count") {
		t.Error("Output should contain turn count")
	}
}

func TestStatusCommand_Execute_ResponseTime(t *testing.T) {
	stats := &game.PlayerStats{HP: 100, SAN: 100}
	gameState := engine.NewGameStateV2()
	cmd := NewStatusCommandV2(stats, 0, gameState, time.Now())

	start := time.Now()
	_, err := cmd.Execute([]string{})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// NFR-P02: Response time should be < 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("Command execution took %v, expected < 100ms", duration)
	}
}

func TestStatusCommand_Execute_PlaytimeFormat(t *testing.T) {
	stats := &game.PlayerStats{HP: 100, SAN: 100}
	gameState := engine.NewGameStateV2()

	// Test with specific playtime
	startTime := time.Now().Add(-1*time.Hour - 23*time.Minute - 45*time.Second)
	cmd := NewStatusCommandV2(stats, 0, gameState, startTime)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check playtime format (should be HH:MM:SS)
	if !strings.Contains(output, "01:23:") {
		t.Error("Playtime should be formatted as HH:MM:SS")
	}
}

func TestStatusCommand_Execute_WithoutGameState(t *testing.T) {
	stats := &game.PlayerStats{HP: 50, SAN: 30}
	cmd := NewStatusCommandV2(stats, 0, nil, time.Now())

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should still show basic stats
	if !strings.Contains(output, "HP:") {
		t.Error("Output should contain HP information")
	}

	// Should show Unknown for location when no gameState
	if !strings.Contains(output, "Unknown") {
		t.Error("Output should show Unknown location when gameState is nil")
	}
}
