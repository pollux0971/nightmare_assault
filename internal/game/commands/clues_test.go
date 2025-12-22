package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

func TestCluesCommand_Name(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewCluesCommand(gameState)

	if cmd.Name() != "clues" {
		t.Errorf("Expected name 'clues', got '%s'", cmd.Name())
	}
}

func TestCluesCommand_Help(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewCluesCommand(gameState)

	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
}

func TestCluesCommand_Execute_NoClues(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewCluesCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if !strings.Contains(output, "CLUES") || !strings.Contains(output, "線索") {
		t.Error("Output should contain clues header")
	}

	if !strings.Contains(output, "No clues") || !strings.Contains(output, "無線索") {
		t.Error("No clues should show appropriate message")
	}
}

func TestCluesCommand_Execute_WithGlobalSeeds(t *testing.T) {
	gameState := engine.NewGameStateV2()

	// Create global seeds with revealed clues
	gs1, err := seed.NewGlobalSeed(
		"GS001",
		"The hospital has a dark secret",
		"Hospital origins",
		"tragic",
		[]seed.ClueTier{
			{Tier: 1, Content: "Strange noises at night", BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Blood on the walls", BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Experiments on patients", BeatStart: 11, BeatEnd: 15},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create global seed: %v", err)
	}

	// Advance tier to reveal clue
	gs1.AdvanceTier()

	gameState.AddGlobalSeed(gs1)

	cmd := NewCluesCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check for dream clues category
	if !strings.Contains(output, "夢境線索") || !strings.Contains(output, "Dream Clues") {
		t.Error("Output should contain dream clues category")
	}

	// Check that revealed clue content is shown
	if !strings.Contains(output, "Strange noises at night") {
		t.Error("Output should contain revealed clue content")
	}
}

func TestCluesCommand_Execute_WithLocalSeeds(t *testing.T) {
	gameState := engine.NewGameStateV2()

	// Create local seeds (harvested clues)
	ls1 := &engine.LocalSeed{
		ID:          "LS001",
		Content:     "A torn diary page on the floor",
		PlantedBeat: 2,
		Urgency:     5,
		IsHarvested: true,
	}

	ls2 := &engine.LocalSeed{
		ID:          "LS002",
		Content:     "Scratches on the wall spelling 'HELP'",
		PlantedBeat: 3,
		Urgency:     7,
		IsHarvested: true,
	}

	gameState.LocalSeeds = []*engine.LocalSeed{ls1, ls2}

	cmd := NewCluesCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check for environmental clues category
	if !strings.Contains(output, "環境線索") || !strings.Contains(output, "Environmental Clues") {
		t.Error("Output should contain environmental clues category")
	}

	// Check that harvested clues are shown
	if !strings.Contains(output, "A torn diary page on the floor") {
		t.Error("Output should contain harvested local seed content")
	}
	if !strings.Contains(output, "Scratches on the wall spelling 'HELP'") {
		t.Error("Output should contain harvested local seed content")
	}
}

func TestCluesCommand_Execute_MixedClues(t *testing.T) {
	gameState := engine.NewGameStateV2()

	// Add global seed
	gs1, _ := seed.NewGlobalSeed(
		"GS001",
		"Mystery",
		"Truth",
		"mysterious",
		[]seed.ClueTier{
			{Tier: 1, Content: "Dream clue 1", BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Dream clue 2", BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Dream clue 3", BeatStart: 11, BeatEnd: 15},
		},
	)
	gs1.AdvanceTier()
	gameState.AddGlobalSeed(gs1)

	// Add local seed
	ls1 := &engine.LocalSeed{
		ID:          "LS001",
		Content:     "Environmental clue 1",
		IsHarvested: true,
	}
	gameState.LocalSeeds = []*engine.LocalSeed{ls1}

	cmd := NewCluesCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should show both categories
	if !strings.Contains(output, "夢境線索") {
		t.Error("Should show dream clues category")
	}
	if !strings.Contains(output, "環境線索") {
		t.Error("Should show environmental clues category")
	}

	// Should show count of 2 total clues
	if !strings.Contains(output, "2") {
		t.Error("Should show total clue count of 2")
	}
}

func TestCluesCommand_Execute_OnlyUnrevealedClues(t *testing.T) {
	gameState := engine.NewGameStateV2()

	// Create global seed but don't advance tier (not revealed)
	gs1, _ := seed.NewGlobalSeed(
		"GS001",
		"Mystery",
		"Truth",
		"mysterious",
		[]seed.ClueTier{
			{Tier: 1, Content: "Hidden clue", BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Hidden clue 2", BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Hidden clue 3", BeatStart: 11, BeatEnd: 15},
		},
	)
	gameState.AddGlobalSeed(gs1)

	// Add unharvested local seed
	ls1 := &engine.LocalSeed{
		ID:          "LS001",
		Content:     "Unharvested clue",
		IsHarvested: false,
	}
	gameState.LocalSeeds = []*engine.LocalSeed{ls1}

	cmd := NewCluesCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should show no clues message
	if !strings.Contains(output, "No clues") || !strings.Contains(output, "無線索") {
		t.Error("Should show no clues message when all clues are unrevealed")
	}

	// Should NOT show hidden clue content
	if strings.Contains(output, "Hidden clue") {
		t.Error("Should not show unrevealed clue content")
	}
	if strings.Contains(output, "Unharvested clue") {
		t.Error("Should not show unharvested clue content")
	}
}

func TestCluesCommand_Execute_NilGameState(t *testing.T) {
	cmd := NewCluesCommand(nil)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if !strings.Contains(output, "未初始化") {
		t.Error("Should show uninitialized message when gameState is nil")
	}
}

func TestCluesCommand_Execute_ResponseTime(t *testing.T) {
	gameState := engine.NewGameStateV2()

	// Add multiple clues
	for i := 0; i < 10; i++ {
		gs, _ := seed.NewGlobalSeed(
			"GS"+string(rune('0'+i)),
			"Mystery",
			"Truth",
			"mysterious",
			[]seed.ClueTier{
				{Tier: 1, Content: "Clue content", BeatStart: 1, BeatEnd: 5},
				{Tier: 2, Content: "Clue content 2", BeatStart: 6, BeatEnd: 10},
				{Tier: 3, Content: "Clue content 3", BeatStart: 11, BeatEnd: 15},
			},
		)
		gs.AdvanceTier()
		gameState.AddGlobalSeed(gs)
	}

	cmd := NewCluesCommand(gameState)

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

func TestCluesCommand_Execute_PreservesOriginalText(t *testing.T) {
	gameState := engine.NewGameStateV2()

	originalText := "The patient's notes mentioned 'shadows moving at midnight'"

	gs1, _ := seed.NewGlobalSeed(
		"GS001",
		"Mystery",
		"Truth",
		"mysterious",
		[]seed.ClueTier{
			{Tier: 1, Content: originalText, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Clue 3", BeatStart: 11, BeatEnd: 15},
		},
	)
	gs1.AdvanceTier()
	gameState.AddGlobalSeed(gs1)

	cmd := NewCluesCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// AC3: Preserve original text verbatim
	if !strings.Contains(output, originalText) {
		t.Error("Should preserve original clue text verbatim")
	}
}
