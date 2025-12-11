package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestNewGameSetupModel(t *testing.T) {
	m := NewGameSetupModel()

	if m.state != SetupThemeInput {
		t.Errorf("Initial state = %v, want %v", m.state, SetupThemeInput)
	}

	if m.config == nil {
		t.Error("Config should not be nil")
	}

	if m.IsDone() {
		t.Error("IsDone() should be false initially")
	}

	if m.IsCancelled() {
		t.Error("IsCancelled() should be false initially")
	}
}

func TestGameSetupModel_ThemeInput(t *testing.T) {
	m := NewGameSetupModel()

	// Simulate typing a valid theme
	// First, we need to update the text input
	m.themeInput.SetValue("廢棄醫院")

	// Press enter to proceed
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	if m.state != SetupDifficultySelect {
		t.Errorf("After valid theme, state = %v, want %v", m.state, SetupDifficultySelect)
	}

	if m.config.Theme != "廢棄醫院" {
		t.Errorf("Theme = %v, want %v", m.config.Theme, "廢棄醫院")
	}
}

func TestGameSetupModel_ThemeValidation(t *testing.T) {
	m := NewGameSetupModel()

	// Set an invalid theme (too short)
	m.themeInput.SetValue("ab")

	// Press enter
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	// Should still be in theme input state due to validation error
	if m.state != SetupThemeInput {
		t.Errorf("After invalid theme, state = %v, want %v", m.state, SetupThemeInput)
	}

	if m.themeError == "" {
		t.Error("themeError should be set for invalid theme")
	}
}

func TestGameSetupModel_DifficultySelect(t *testing.T) {
	m := NewGameSetupModel()
	m.themeInput.SetValue("valid theme")
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	// Now in difficulty select state
	if m.state != SetupDifficultySelect {
		t.Fatalf("Expected state = %v, got %v", SetupDifficultySelect, m.state)
	}

	// Select Hell difficulty (index 2)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newModel.(GameSetupModel)
	if m.selectedIndex != 2 {
		t.Errorf("After pressing 3, selectedIndex = %v, want 2", m.selectedIndex)
	}

	// Confirm selection
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	if m.state != SetupLengthSelect {
		t.Errorf("After difficulty selection, state = %v, want %v", m.state, SetupLengthSelect)
	}

	if m.config.Difficulty != game.DifficultyHell {
		t.Errorf("Difficulty = %v, want %v", m.config.Difficulty, game.DifficultyHell)
	}
}

func TestGameSetupModel_LengthSelect(t *testing.T) {
	m := NewGameSetupModel()
	m.config.Theme = "valid theme"
	m.state = SetupLengthSelect
	m.selectedIndex = 0

	// Select Long (index 2)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newModel.(GameSetupModel)

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	if m.state != SetupAdultModeToggle {
		t.Errorf("After length selection, state = %v, want %v", m.state, SetupAdultModeToggle)
	}

	if m.config.Length != game.LengthLong {
		t.Errorf("Length = %v, want %v", m.config.Length, game.LengthLong)
	}
}

func TestGameSetupModel_AdultModeToggle(t *testing.T) {
	m := NewGameSetupModel()
	m.config.Theme = "valid theme"
	m.state = SetupAdultModeToggle
	m.selectedIndex = 0 // Enable adult mode

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	if m.state != SetupSummary {
		t.Errorf("After adult mode toggle, state = %v, want %v", m.state, SetupSummary)
	}

	if !m.config.AdultMode {
		t.Error("AdultMode should be true when selecting index 0")
	}
}

func TestGameSetupModel_Summary_Confirm(t *testing.T) {
	m := NewGameSetupModel()
	m.config.Theme = "valid theme"
	m.state = SetupSummary
	m.selectedIndex = 0 // Confirm

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	if !m.confirmed {
		t.Error("confirmed should be true")
	}

	if !m.IsDone() {
		t.Error("IsDone() should be true after confirm")
	}

	// Execute the command and check the message
	if cmd != nil {
		msg := cmd()
		if doneMsg, ok := msg.(GameSetupDoneMsg); ok {
			if doneMsg.Cancelled {
				t.Error("GameSetupDoneMsg.Cancelled should be false")
			}
			if doneMsg.Config == nil {
				t.Error("GameSetupDoneMsg.Config should not be nil")
			}
		} else {
			t.Error("Command should return GameSetupDoneMsg")
		}
	}
}

func TestGameSetupModel_Summary_Edit(t *testing.T) {
	m := NewGameSetupModel()
	m.config.Theme = "valid theme"
	m.state = SetupSummary
	m.selectedIndex = 1 // Edit

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(GameSetupModel)

	if m.state != SetupThemeInput {
		t.Errorf("After edit selection, state = %v, want %v", m.state, SetupThemeInput)
	}
}

func TestGameSetupModel_BackNavigation(t *testing.T) {
	tests := []struct {
		name          string
		initialState  SetupState
		expectedState SetupState
	}{
		{"Difficulty to Theme", SetupDifficultySelect, SetupThemeInput},
		{"Length to Difficulty", SetupLengthSelect, SetupDifficultySelect},
		{"AdultMode to Length", SetupAdultModeToggle, SetupLengthSelect},
		{"Summary to AdultMode", SetupSummary, SetupAdultModeToggle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGameSetupModel()
			m.config.Theme = "valid theme"
			m.state = tt.initialState

			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			m = newModel.(GameSetupModel)

			if m.state != tt.expectedState {
				t.Errorf("After ESC, state = %v, want %v", m.state, tt.expectedState)
			}
		})
	}
}

func TestGameSetupModel_Cancel(t *testing.T) {
	m := NewGameSetupModel()

	// Press ESC in theme input to cancel
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(GameSetupModel)

	if !m.cancelled {
		t.Error("cancelled should be true after ESC in theme input")
	}

	if !m.IsCancelled() {
		t.Error("IsCancelled() should be true")
	}

	// Check command
	if cmd != nil {
		msg := cmd()
		if doneMsg, ok := msg.(GameSetupDoneMsg); ok {
			if !doneMsg.Cancelled {
				t.Error("GameSetupDoneMsg.Cancelled should be true")
			}
		}
	}
}

func TestGameSetupModel_CtrlC(t *testing.T) {
	m := NewGameSetupModel()

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = newModel.(GameSetupModel)

	if !m.cancelled {
		t.Error("cancelled should be true after Ctrl+C")
	}

	if cmd == nil {
		t.Error("Command should not be nil after Ctrl+C")
	}
}

func TestGameSetupModel_WindowSize(t *testing.T) {
	m := NewGameSetupModel()

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModel.(GameSetupModel)

	if m.width != 100 {
		t.Errorf("width = %v, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %v, want 50", m.height)
	}
}

func TestGameSetupModel_KeyNavigation(t *testing.T) {
	m := NewGameSetupModel()
	m.config.Theme = "valid theme"
	m.state = SetupDifficultySelect
	m.selectedIndex = 1

	// Test up navigation
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(GameSetupModel)
	if m.selectedIndex != 0 {
		t.Errorf("After up, selectedIndex = %v, want 0", m.selectedIndex)
	}

	// Test down navigation
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(GameSetupModel)
	if m.selectedIndex != 1 {
		t.Errorf("After down, selectedIndex = %v, want 1", m.selectedIndex)
	}

	// Test vim-style navigation
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(GameSetupModel)
	if m.selectedIndex != 0 {
		t.Errorf("After k, selectedIndex = %v, want 0", m.selectedIndex)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(GameSetupModel)
	if m.selectedIndex != 1 {
		t.Errorf("After j, selectedIndex = %v, want 1", m.selectedIndex)
	}
}

func TestGameSetupModel_View(t *testing.T) {
	m := NewGameSetupModel()

	// Test that View() doesn't panic for each state
	states := []SetupState{
		SetupThemeInput,
		SetupDifficultySelect,
		SetupLengthSelect,
		SetupAdultModeToggle,
		SetupSummary,
	}

	for _, state := range states {
		m.state = state
		view := m.View()
		if view == "" {
			t.Errorf("View() returned empty string for state %v", state)
		}
	}
}

func TestGameSetupModel_GetConfig(t *testing.T) {
	m := NewGameSetupModel()
	m.config.Theme = "test"

	config := m.GetConfig()
	if config.Theme != "test" {
		t.Errorf("GetConfig().Theme = %v, want test", config.Theme)
	}
}
