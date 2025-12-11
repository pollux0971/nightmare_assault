package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMainMenuModel(t *testing.T) {
	m := NewMainMenuModel("1.0.0", false)

	if m.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", m.version)
	}

	if m.hasSaveFiles {
		t.Error("Expected hasSaveFiles to be false")
	}
}

func TestMainMenuViewContainsTitle(t *testing.T) {
	m := NewMainMenuModel("1.0.0", false)
	view := m.View()

	if !strings.Contains(view, "NIGHTMARE ASSAULT") {
		t.Error("Expected title in view")
	}
}

func TestMainMenuViewContainsVersion(t *testing.T) {
	m := NewMainMenuModel("1.0.0", false)
	view := m.View()

	if !strings.Contains(view, "v1.0.0") {
		t.Errorf("Expected version in view, got: %s", view)
	}
}

func TestMainMenuViewContainsOptions(t *testing.T) {
	m := NewMainMenuModel("1.0.0", false)
	view := m.View()

	options := []string{"新遊戲", "繼續遊戲", "設定", "離開"}
	for _, opt := range options {
		if !strings.Contains(view, opt) {
			t.Errorf("Expected option '%s' in view", opt)
		}
	}
}

func TestMainMenuNavigation(t *testing.T) {
	m := NewMainMenuModel("1.0.0", true) // Enable continue to test normal navigation

	// Test down navigation
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := newModel.(MainMenuModel)
	if updated.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", updated.selectedIndex)
	}

	// Test up navigation
	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = newModel.(MainMenuModel)
	if updated.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", updated.selectedIndex)
	}
}

func TestMainMenuNumberSelect(t *testing.T) {
	m := NewMainMenuModel("1.0.0", true) // Enable continue game

	// Press '3' for settings
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	updated := newModel.(MainMenuModel)

	if cmd == nil {
		t.Error("Expected command after selecting settings")
	}

	// Verify selectedIndex changed
	if updated.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2, got %d", updated.selectedIndex)
	}
}

func TestMainMenuExitConfirmation(t *testing.T) {
	m := NewMainMenuModel("1.0.0", false)
	m.selectedIndex = 3 // Exit option
	// Need to update the list selection to match
	m.list.Select(3)

	// Press enter to select exit
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(MainMenuModel)

	if !updated.IsExitConfirming() {
		t.Error("Expected exit confirmation to be shown")
	}

	// Press 'n' to cancel
	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated = newModel.(MainMenuModel)

	if updated.IsExitConfirming() {
		t.Error("Expected exit confirmation to be cancelled")
	}
}

func TestMainMenuDisabledContinue(t *testing.T) {
	m := NewMainMenuModel("1.0.0", false) // No save files

	// Navigate to Continue (index 1)
	m.selectedIndex = 0
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := newModel.(MainMenuModel)

	// Should skip disabled "Continue Game" and go to Settings
	if updated.selectedIndex != 2 {
		t.Errorf("Expected to skip disabled item, selectedIndex should be 2, got %d", updated.selectedIndex)
	}
}
