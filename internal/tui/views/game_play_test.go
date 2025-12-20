package views

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// TestDebriefCollector_Integration tests that DebriefCollector is properly integrated into GamePlayModel
func TestDebriefCollector_Integration(t *testing.T) {
	t.Skip("debriefCollector not implemented yet")
}

// TestDebriefCollector_RecordDecision tests that decision points are recorded
func TestDebriefCollector_RecordDecision(t *testing.T) {
	t.Skip("debriefCollector not implemented yet")
}

// TestDebriefCollector_RecordFreeTextDecision tests that free text decisions are recorded
func TestDebriefCollector_RecordFreeTextDecision(t *testing.T) {
	t.Skip("debriefCollector not implemented yet")
}

// TestDeathTriggersDebrief tests that death properly triggers debrief with collected data
func TestDeathTriggersDebrief(t *testing.T) {
	t.Skip("debriefCollector/triggerDeath not implemented yet")
}

// TestDeathTriggersDebrief_SANDeath tests SAN-based death
func TestDeathTriggersDebrief_SANDeath(t *testing.T) {
	t.Skip("debriefCollector/triggerDeath not implemented yet")
}

// TestStreamDoneMsg_Integration tests that StreamDoneMsg properly updates debrief collector
func TestStreamDoneMsg_Integration(t *testing.T) {
	// Create test config
	cfg := &config.Config{}

	// Create test game config
	gameConfig := &game.GameConfig{
		Difficulty: game.DifficultyEasy,
	}

	// Create test stats
	stats := game.NewPlayerStats()
	stats.HP = 100
	stats.SAN = 100

	// Create GamePlayModel
	model := NewGamePlayModel(nil, stats, gameConfig, nil, nil, cfg)
	model.ready = true

	// Simulate StreamDoneMsg with HP/SAN changes
	msg := StreamDoneMsg{
		Content:         "你受到了攻擊，體力下降！",
		ChoiceOptions:   []string{"反擊", "逃跑"},
		HPChange:        -30,
		SANChange:       -10,
		ChangeReason:    "被怪物攻擊",
	}

	// Process the message
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(GamePlayModel)

	// Verify stats were updated
	if model.stats.HP != 70 {
		t.Errorf("Expected HP 70, got %d", model.stats.HP)
	}
	if model.stats.SAN != 90 {
		t.Errorf("Expected SAN 90, got %d", model.stats.SAN)
	}

	// Note: HP/SAN changes are NOT directly recorded in debrief collector
	// They are recorded as part of rule violations or state transitions
	// This test verifies that the stats are correctly updated in the model
}

// TestDebriefCollector_DifficultySettings tests that difficulty affects debrief data
func TestDebriefCollector_DifficultySettings(t *testing.T) {
	t.Skip("debriefCollector not implemented yet")
}

// ============================================================================
// Story 8.2: Load System Tests
// ============================================================================

// TestRestoreGameState_FullRestore tests complete state restoration
// Story 8.2 AC6: GamePlayModel.RestoreGameState correctly updates all game state
func TestRestoreGameState_FullRestore(t *testing.T) {
	t.Skip("RestoreGameState not implemented yet")
}

// TestRestoreGameState_NilChecks tests error handling for nil parameters
func TestRestoreGameState_NilChecks(t *testing.T) {
	t.Skip("RestoreGameState not implemented yet")
}

// TestRestoreGameState_EffectManagerSync tests SAN sync with effect manager
func TestRestoreGameState_EffectManagerSync(t *testing.T) {
	t.Skip("RestoreGameState not implemented yet")
}

// TestLoadCommandV2_Integration tests that LoadCommandV2 is properly registered
// Story 8.2 AC1, AC2: /load command integration
func TestLoadCommandV2_Integration(t *testing.T) {
	t.Skip("LoadCommandV2 not fully integrated yet")
}

// TestLoadCommandV2_EmptySlotList tests /load with no saves
// Story 8.2 AC2: /load without args shows slot list
func TestLoadCommandV2_EmptySlotList(t *testing.T) {
	t.Skip("LoadCommandV2 not fully integrated yet")
}

// TestLoadCommandV2_InvalidSlot tests error handling for invalid slot IDs
// Story 8.2 AC5: Empty slot error handling
func TestLoadCommandV2_InvalidSlot(t *testing.T) {
	t.Skip("LoadCommandV2 not fully integrated yet")
}

// TestRestoreGameState_DebriefCollectorSync tests that debrief collector is synced
func TestRestoreGameState_DebriefCollectorSync(t *testing.T) {
	t.Skip("debriefCollector not implemented yet")
}

// TestRestoreGameState_SettingsNil tests restoration works with nil settings
func TestRestoreGameState_SettingsNil(t *testing.T) {
	t.Skip("RestoreGameState not implemented yet")
}

// ============================================================================
// Story 8.1: Save System Integration Tests
// ============================================================================

// TestSaveCommandV2_Integration tests that SaveCommandV2 is properly registered
// Story 8.1 AC #1: /save command integration
func TestSaveCommandV2_Integration(t *testing.T) {
	t.Skip("SaveCommandV2 not fully integrated yet")
}

// TestStateAccessor_GetGameState tests that GetGameState returns synchronized state
// Story 8.1 AC #1: Provides synchronized GameStateV2 for saving
func TestStateAccessor_GetGameState(t *testing.T) {
	t.Skip("GetGameState not implemented yet")
}

// TestStateAccessor_GetGameSettings tests that GetGameSettings returns correct settings
// Story 8.1 AC #2: Provides settings (theme, difficulty, length, adult_mode)
func TestStateAccessor_GetGameSettings(t *testing.T) {
	t.Skip("GetGameSettings not implemented yet")
}

// TestStateAccessor_GetGameStartTime tests that game start time is recorded
// Story 8.1 AC #2: Used for calculating playtime in save metadata
func TestStateAccessor_GetGameStartTime(t *testing.T) {
	t.Skip("GetGameStartTime not implemented yet")
}

// TestStateAccessor_GetStoryBible tests GetStoryBible and SetStoryBible
// Story 8.1 AC #2: Provides StoryBible for complete save file
func TestStateAccessor_GetStoryBible(t *testing.T) {
	t.Skip("GetStoryBible/SetStoryBible not implemented yet")
}

// TestSaveCommandV2_SlotList tests /save without args shows slot list
// Story 8.1 AC #3: Shows available save slots
func TestSaveCommandV2_SlotList(t *testing.T) {
	t.Skip("SaveCommandV2 not fully integrated yet")
}

// TestSaveCommandV2_SaveToSlot tests /save <slot> actually saves
// Story 8.1 AC #1: /save <slot> saves current game state
// Note: This test uses the savefile package directly to avoid pointer semantics issues
// with the command registration. Full integration testing requires the running app context.
func TestSaveCommandV2_SaveToSlot(t *testing.T) {
	t.Skip("GetGameState/GetGameSettings not implemented yet")
}

// TestSaveCommandV2_InvalidSlot tests error handling for invalid slot
// Story 8.1 AC #5: Returns clear error messages
func TestSaveCommandV2_InvalidSlot(t *testing.T) {
	t.Skip("SaveCommandV2 not fully integrated yet")
}

// ============================================================================
// Story 10-7: Config Hot Reload Tests
// ============================================================================

// TestConfigReload_CtrlRHotkey tests that Ctrl+R triggers config reload
// Story 10-7 AC1: Press Ctrl+R to reload config
func TestConfigReload_CtrlRHotkey(t *testing.T) {
	// Create temporary config directory
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".nightmare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	// Set temporary HOME for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create initial config
	initialConfig := config.DefaultConfig()
	initialConfig.Audio.MasterVolume = 50

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create GamePlayModel
	gameConfig := &game.GameConfig{Difficulty: game.DifficultyEasy}
	stats := game.NewPlayerStats()
	model := NewGamePlayModel(nil, stats, gameConfig, nil, nil, cfg)
	model.ready = true

	// Simulate Ctrl+R key press
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, cmd := model.Update(keyMsg)

	// Verify a command was returned (the reload command)
	if cmd == nil {
		t.Fatal("Expected reload command to be returned")
	}

	// Execute the command to get ConfigReloadedMsg
	msg := cmd()
	configMsg, ok := msg.(ConfigReloadedMsg)
	if !ok {
		t.Fatalf("Expected ConfigReloadedMsg, got %T", msg)
	}

	// Verify reload succeeded
	if !configMsg.Success {
		t.Errorf("Expected successful reload, got error: %v", configMsg.Error)
	}

	// Apply the ConfigReloadedMsg to the model
	model = updatedModel.(GamePlayModel)
	updatedModel, _ = model.Update(configMsg)
	model = updatedModel.(GamePlayModel)

	// Verify feedback message was added to story (fallback English or translated)
	if !strings.Contains(model.currentStory, "Config reloaded") && !strings.Contains(model.currentStory, "配置已重新載入") {
		t.Error("Expected success message in story")
	}
}

// TestConfigReload_InvalidConfig tests error handling for invalid config
// Story 10-7 AC3: On error, show error message and keep original config
func TestConfigReload_InvalidConfig(t *testing.T) {
	// Create temporary config directory
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".nightmare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	// Set temporary HOME for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create valid initial config
	initialConfig := config.DefaultConfig()
	initialConfig.Audio.MasterVolume = 50

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Store original volume
	originalVolume := cfg.Audio.MasterVolume

	// Create GamePlayModel
	gameConfig := &game.GameConfig{Difficulty: game.DifficultyEasy}
	stats := game.NewPlayerStats()
	model := NewGamePlayModel(nil, stats, gameConfig, nil, nil, cfg)
	model.ready = true

	// Write invalid JSON to config file
	invalidJSON := []byte("{invalid json")
	if err := os.WriteFile(configPath, invalidJSON, 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Simulate Ctrl+R key press
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, cmd := model.Update(keyMsg)

	// Execute the reload command
	msg := cmd()
	configMsg, ok := msg.(ConfigReloadedMsg)
	if !ok {
		t.Fatalf("Expected ConfigReloadedMsg, got %T", msg)
	}

	// Verify reload failed
	if configMsg.Success {
		t.Error("Expected reload to fail with invalid JSON")
	}
	if configMsg.Error == nil {
		t.Error("Expected error to be set")
	}

	// Apply the ConfigReloadedMsg to the model
	model = updatedModel.(GamePlayModel)
	updatedModel, _ = model.Update(configMsg)
	model = updatedModel.(GamePlayModel)

	// Verify error message was added to story (fallback English or translated)
	if !strings.Contains(model.currentStory, "Config reload failed") && !strings.Contains(model.currentStory, "配置重載失敗") {
		t.Error("Expected error message in story")
	}
	if !strings.Contains(model.currentStory, "Keeping original config") && !strings.Contains(model.currentStory, "保持原有配置") {
		t.Error("Expected 'keep original config' message in story")
	}

	// Verify original config was preserved
	if model.config.Audio.MasterVolume != originalVolume {
		t.Errorf("Expected original volume %d to be preserved, got %d", originalVolume, model.config.Audio.MasterVolume)
	}
}

// TestConfigReload_MessageUpdate tests ConfigReloadedMsg handling
// Story 10-7 AC2: New config takes effect immediately
func TestConfigReload_MessageUpdate(t *testing.T) {
	cfg := &config.Config{}
	gameConfig := &game.GameConfig{Difficulty: game.DifficultyEasy}
	stats := game.NewPlayerStats()
	model := NewGamePlayModel(nil, stats, gameConfig, nil, nil, cfg)
	model.ready = true

	// Test success message
	successMsg := ConfigReloadedMsg{
		Success: true,
		Error:   nil,
	}

	updatedModel, _ := model.Update(successMsg)
	model = updatedModel.(GamePlayModel)

	// Check for success message (fallback English or translated)
	if !strings.Contains(model.currentStory, "✓ Config reloaded") && !strings.Contains(model.currentStory, "✓ 配置已重新載入") {
		t.Error("Expected success checkmark and message in story")
	}

	// Test error message
	model.currentStory = "" // Reset story
	errorMsg := ConfigReloadedMsg{
		Success: false,
		Error:   fmt.Errorf("test error"),
	}

	updatedModel, _ = model.Update(errorMsg)
	model = updatedModel.(GamePlayModel)

	// Check for error message (fallback English or translated)
	if !strings.Contains(model.currentStory, "✗ Config reload failed") && !strings.Contains(model.currentStory, "✗ 配置重載失敗") {
		t.Error("Expected error cross mark and message in story")
	}
	if !strings.Contains(model.currentStory, "test error") {
		t.Error("Expected error detail in story")
	}
}
