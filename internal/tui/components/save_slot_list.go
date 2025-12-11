// Package components provides reusable TUI components for Nightmare Assault.
package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SlotMode represents the mode of the save slot list (save or load).
type SlotMode int

const (
	ModeSave SlotMode = iota
	ModeLoad
)

// SlotDisplayInfo contains information to display about a save slot.
type SlotDisplayInfo struct {
	SlotID   int
	IsEmpty  bool
	Chapter  int
	PlayTime int // seconds
	SavedAt  time.Time
	Location string
}

// Title returns the display title for the slot.
func (s SlotDisplayInfo) Title() string {
	if s.IsEmpty {
		return fmt.Sprintf("[%d] 空", s.SlotID)
	}
	return fmt.Sprintf("[%d] 章節 %d - %s", s.SlotID, s.Chapter, s.locationOrDefault())
}

// Description returns the display description for the slot.
func (s SlotDisplayInfo) Description() string {
	if s.IsEmpty {
		return "尚無存檔資料"
	}
	return fmt.Sprintf("遊玩時間: %s | 存檔時間: %s",
		FormatPlayTimeDisplay(s.PlayTime),
		s.SavedAt.Format("2006-01-02 15:04"))
}

// FilterValue returns the filter value for the slot.
func (s SlotDisplayInfo) FilterValue() string {
	if s.IsEmpty {
		return fmt.Sprintf("slot %d empty", s.SlotID)
	}
	return fmt.Sprintf("slot %d chapter %d %s", s.SlotID, s.Chapter, s.Location)
}

func (s SlotDisplayInfo) locationOrDefault() string {
	if s.Location == "" {
		return "未知地點"
	}
	return s.Location
}

// FormatPlayTimeDisplay formats play time in seconds to a display string.
func FormatPlayTimeDisplay(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// SaveSlotList is a BubbleTea model for displaying and selecting save slots.
type SaveSlotList struct {
	list                 list.Model
	slots                []SlotDisplayInfo
	Mode                 SlotMode
	ShowingConfirmDialog bool
	confirmSlot          *SlotDisplayInfo
	width                int
	height               int
}

// slotItem wraps SlotDisplayInfo for the list model.
type slotItem struct {
	info SlotDisplayInfo
}

func (i slotItem) Title() string       { return i.info.Title() }
func (i slotItem) Description() string { return i.info.Description() }
func (i slotItem) FilterValue() string { return i.info.FilterValue() }

// NewSaveSlotList creates a new save slot list.
func NewSaveSlotList(slots []SlotDisplayInfo, mode SlotMode) *SaveSlotList {
	items := make([]list.Item, len(slots))
	for i, slot := range slots {
		items[i] = slotItem{info: slot}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true

	// Style the list items
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9D4EDD")).
		Bold(true).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#9D4EDD")).
		PaddingLeft(1)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B2CBF")).
		PaddingLeft(3)

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECF0F1")).
		PaddingLeft(2)

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7F8C8D")).
		PaddingLeft(2)

	l := list.New(items, delegate, 40, 15)
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	if mode == ModeSave {
		l.Title = "選擇存檔槽位"
	} else {
		l.Title = "選擇要載入的存檔"
	}

	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9D4EDD")).
		Bold(true).
		MarginBottom(1)

	return &SaveSlotList{
		list:   l,
		slots:  slots,
		Mode:   mode,
		width:  40,
		height: 15,
	}
}

// Init initializes the save slot list.
func (m *SaveSlotList) Init() tea.Cmd {
	return nil
}

// Update handles messages for the save slot list.
func (m *SaveSlotList) Update(msg tea.Msg) (*SaveSlotList, tea.Cmd) {
	if m.ShowingConfirmDialog {
		return m.updateConfirmDialog(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.SelectedSlot()
			if !m.CanSelectSlot(selected) {
				return m, nil
			}

			// Check if we need confirmation (overwriting existing save)
			if m.Mode == ModeSave && !selected.IsEmpty {
				m.ShowConfirmDialog()
				m.confirmSlot = &selected
				return m, nil
			}

			// Return slot selection message
			return m, func() tea.Msg {
				return SlotSelectedMsg{Slot: selected, Mode: m.Mode}
			}

		case "esc":
			return m, func() tea.Msg {
				return SlotCancelledMsg{}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *SaveSlotList) updateConfirmDialog(msg tea.Msg) (*SaveSlotList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "是":
			m.HideConfirmDialog()
			if m.confirmSlot != nil {
				slot := *m.confirmSlot
				m.confirmSlot = nil
				return m, func() tea.Msg {
					return SlotSelectedMsg{Slot: slot, Mode: m.Mode, Overwrite: true}
				}
			}

		case "n", "N", "否", "esc":
			m.HideConfirmDialog()
			m.confirmSlot = nil
		}
	}
	return m, nil
}

// View renders the save slot list.
func (m *SaveSlotList) View() string {
	if m.ShowingConfirmDialog && m.confirmSlot != nil {
		return m.renderConfirmDialog()
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9D4EDD")).
		Padding(1, 2)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7F8C8D")).
		MarginTop(1).
		Render("[Enter] 選擇 | [ESC] 取消")

	return boxStyle.Render(m.list.View() + "\n" + hint)
}

func (m *SaveSlotList) renderConfirmDialog() string {
	if m.confirmSlot == nil {
		return ""
	}

	slot := *m.confirmSlot
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(2, 4).
		Width(45)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true).
		Render("⚠️ 覆蓋確認")

	message := fmt.Sprintf("確定要覆蓋現有存檔？\n\n章節 %d - %s\n遊玩時間: %s\n存檔時間: %s",
		slot.Chapter,
		slot.locationOrDefault(),
		FormatPlayTimeDisplay(slot.PlayTime),
		slot.SavedAt.Format("2006-01-02 15:04"))

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7F8C8D")).
		MarginTop(1).
		Render("\n[Y] 是，覆蓋 | [N] 否，取消")

	return dialogStyle.Render(title + "\n\n" + message + hint)
}

// SelectedSlot returns the currently selected slot.
func (m *SaveSlotList) SelectedSlot() SlotDisplayInfo {
	if selected, ok := m.list.SelectedItem().(slotItem); ok {
		return selected.info
	}
	return SlotDisplayInfo{}
}

// CanSelectSlot checks if a slot can be selected based on mode.
func (m *SaveSlotList) CanSelectSlot(slot SlotDisplayInfo) bool {
	if m.Mode == ModeLoad && slot.IsEmpty {
		return false // Cannot load empty slot
	}
	return true
}

// ShowConfirmDialog shows the overwrite confirmation dialog.
func (m *SaveSlotList) ShowConfirmDialog() {
	m.ShowingConfirmDialog = true
}

// HideConfirmDialog hides the overwrite confirmation dialog.
func (m *SaveSlotList) HideConfirmDialog() {
	m.ShowingConfirmDialog = false
}

// SlotSelectedMsg is sent when a slot is selected.
type SlotSelectedMsg struct {
	Slot      SlotDisplayInfo
	Mode      SlotMode
	Overwrite bool
}

// SlotCancelledMsg is sent when slot selection is cancelled.
type SlotCancelledMsg struct{}
