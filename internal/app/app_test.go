package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/views"
)

func TestNew(t *testing.T) {
	version := "1.0.0"
	m := New(version)

	if m.version != version {
		t.Errorf("Expected version %s, got %s", version, m.version)
	}

	if m.ready {
		t.Error("Expected ready to be false initially")
	}

	if m.state != StateLoading {
		t.Errorf("Expected state StateLoading, got %v", m.state)
	}
}

func TestInit(t *testing.T) {
	m := New("1.0.0")
	cmd := m.Init()

	// Init returns nil (config loading happens in Update now)
	if cmd != nil {
		t.Error("Expected Init to return nil")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := New("1.0.0")

	// Simulate window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	newModel, _ := m.Update(msg)

	updatedModel := newModel.(Model)
	if updatedModel.width != 100 {
		t.Errorf("Expected width 100, got %d", updatedModel.width)
	}
	if updatedModel.height != 30 {
		t.Errorf("Expected height 30, got %d", updatedModel.height)
	}
	if !updatedModel.ready {
		t.Error("Expected ready to be true after window size message")
	}
}

func TestUpdateQuit(t *testing.T) {
	m := New("1.0.0")
	m.state = StateMainMenu // Only quit with 'q' in main menu

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command for 'q' key in main menu")
	}
}

func TestUpdateCtrlC(t *testing.T) {
	m := New("1.0.0")

	// Test Ctrl+C
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command for Ctrl+C")
	}
}

func TestViewNotReady(t *testing.T) {
	m := New("1.0.0")
	view := m.View()

	// Now uses Chinese
	if view != "載入中..." {
		t.Errorf("Expected '載入中...' when not ready, got %s", view)
	}
}

func TestViewSmallTerminal(t *testing.T) {
	m := New("1.0.0")
	m.ready = true
	m.width = 60  // Less than MinWidth (80)
	m.height = 20 // Less than MinHeight (24)

	view := m.View()

	// Now uses Chinese
	if !strings.Contains(view, "終端機太小") {
		t.Errorf("Expected warning about terminal size, got: %s", view)
	}
}

func TestViewNormalTerminal(t *testing.T) {
	m := New("1.0.0")
	m.ready = true
	m.width = 100
	m.height = 30
	m.state = StateMainMenu
	// Initialize main menu with version
	m.mainMenu = views.NewMainMenuModel("1.0.0", false)

	view := m.View()

	if !strings.Contains(view, "NIGHTMARE ASSAULT") {
		t.Errorf("Expected title in view, got: %s", view)
	}
	if !strings.Contains(view, "v1.0.0") {
		t.Errorf("Expected version in view, got: %s", view)
	}
	// Check for quit hint in Chinese
	if !strings.Contains(view, "離開") {
		t.Errorf("Expected quit hint in view, got: %s", view)
	}
}

func TestWidthHeight(t *testing.T) {
	m := New("1.0.0")
	m.width = 120
	m.height = 40

	if m.Width() != 120 {
		t.Errorf("Expected Width() to return 120, got %d", m.Width())
	}
	if m.Height() != 40 {
		t.Errorf("Expected Height() to return 40, got %d", m.Height())
	}
}

func TestState(t *testing.T) {
	m := New("1.0.0")

	if m.State() != StateLoading {
		t.Errorf("Expected initial state StateLoading, got %v", m.State())
	}

	m.state = StateMainMenu
	if m.State() != StateMainMenu {
		t.Errorf("Expected state StateMainMenu, got %v", m.State())
	}
}
