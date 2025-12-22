package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

func TestInventoryCommand_Name(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewInventoryCommand(gameState)

	if cmd.Name() != "inventory" {
		t.Errorf("Expected name 'inventory', got '%s'", cmd.Name())
	}
}

func TestInventoryCommand_Help(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewInventoryCommand(gameState)

	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
}

func TestInventoryCommand_Aliases(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewInventoryCommand(gameState)

	aliases := cmd.Aliases()
	if len(aliases) == 0 {
		t.Error("Inventory command should have aliases")
	}

	// Check for expected aliases
	hasInv := false
	hasI := false
	for _, alias := range aliases {
		if alias == "inv" {
			hasInv = true
		}
		if alias == "i" {
			hasI = true
		}
	}

	if !hasInv {
		t.Error("Should have 'inv' alias")
	}
	if !hasI {
		t.Error("Should have 'i' alias")
	}
}

func TestInventoryCommand_Execute_EmptyInventory(t *testing.T) {
	gameState := engine.NewGameStateV2()
	cmd := NewInventoryCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if !strings.Contains(output, "INVENTORY") || !strings.Contains(output, "背包") {
		t.Error("Output should contain inventory header")
	}

	if !strings.Contains(output, "Empty") || !strings.Contains(output, "無物品") {
		t.Error("Empty inventory should show 'Empty' or '無物品' message")
	}
}

func TestInventoryCommand_Execute_WithItems(t *testing.T) {
	gameState := engine.NewGameStateV2()
	gameState.Inventory = []string{
		"Old Key",
		"Torn Letter",
		"Flashlight",
		"Medical Kit",
	}

	cmd := NewInventoryCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check that all items are listed
	if !strings.Contains(output, "Old Key") {
		t.Error("Output should contain 'Old Key'")
	}
	if !strings.Contains(output, "Torn Letter") {
		t.Error("Output should contain 'Torn Letter'")
	}
	if !strings.Contains(output, "Flashlight") {
		t.Error("Output should contain 'Flashlight'")
	}
	if !strings.Contains(output, "Medical Kit") {
		t.Error("Output should contain 'Medical Kit'")
	}

	// Check item count
	if !strings.Contains(output, "4") {
		t.Error("Output should show item count of 4")
	}
}

func TestInventoryCommand_Execute_NilGameState(t *testing.T) {
	cmd := NewInventoryCommand(nil)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if !strings.Contains(output, "未初始化") {
		t.Error("Should show uninitialized message when gameState is nil")
	}
}

func TestInventoryCommand_Execute_ResponseTime(t *testing.T) {
	gameState := engine.NewGameStateV2()
	gameState.Inventory = []string{
		"Item1", "Item2", "Item3", "Item4", "Item5",
		"Item6", "Item7", "Item8", "Item9", "Item10",
	}

	cmd := NewInventoryCommand(gameState)

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

func TestInventoryCommand_Execute_ListFormatting(t *testing.T) {
	gameState := engine.NewGameStateV2()
	gameState.Inventory = []string{
		"First Item",
		"Second Item",
		"Third Item",
	}

	cmd := NewInventoryCommand(gameState)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check that items are numbered
	if !strings.Contains(output, "1.") {
		t.Error("Items should be numbered starting with 1")
	}
	if !strings.Contains(output, "2.") {
		t.Error("Items should be numbered (2)")
	}
	if !strings.Contains(output, "3.") {
		t.Error("Items should be numbered (3)")
	}
}
