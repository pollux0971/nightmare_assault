package components

import (
	"testing"
	"time"
)

func TestNewSaveSlotList(t *testing.T) {
	slots := []SlotDisplayInfo{
		{SlotID: 1, IsEmpty: true},
		{SlotID: 2, IsEmpty: false, Chapter: 3, PlayTime: 3600, SavedAt: time.Now()},
		{SlotID: 3, IsEmpty: true},
	}

	list := NewSaveSlotList(slots, ModeSave)
	if list == nil {
		t.Fatal("NewSaveSlotList returned nil")
	}

	if list.Mode != ModeSave {
		t.Errorf("Expected mode ModeSave, got %v", list.Mode)
	}
}

func TestSaveSlotListModeLoad(t *testing.T) {
	slots := []SlotDisplayInfo{
		{SlotID: 1, IsEmpty: false, Chapter: 2, PlayTime: 1800, SavedAt: time.Now()},
		{SlotID: 2, IsEmpty: true},
		{SlotID: 3, IsEmpty: true},
	}

	list := NewSaveSlotList(slots, ModeLoad)
	if list.Mode != ModeLoad {
		t.Errorf("Expected mode ModeLoad, got %v", list.Mode)
	}
}

func TestSlotDisplayInfoFormat(t *testing.T) {
	// Empty slot
	empty := SlotDisplayInfo{SlotID: 1, IsEmpty: true}
	title := empty.Title()
	if title != "[1] 空" {
		t.Errorf("Expected '[1] 空', got '%s'", title)
	}

	// Used slot
	savedAt := time.Date(2024, 12, 10, 22, 30, 0, 0, time.UTC)
	used := SlotDisplayInfo{
		SlotID:   2,
		IsEmpty:  false,
		Chapter:  3,
		PlayTime: 5000, // 1h 23m
		SavedAt:  savedAt,
		Location: "廢棄醫院",
	}

	title = used.Title()
	if title == "" {
		t.Error("Title should not be empty for used slot")
	}

	desc := used.Description()
	if desc == "" {
		t.Error("Description should not be empty for used slot")
	}
}

func TestFormatPlayTimeDisplay(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "0m"},
		{60, "1m"},
		{3600, "1h 0m"},
		{3660, "1h 1m"},
		{5000, "1h 23m"},
		{7200, "2h 0m"},
	}

	for _, tt := range tests {
		result := FormatPlayTimeDisplay(tt.seconds)
		if result != tt.expected {
			t.Errorf("FormatPlayTimeDisplay(%d) = %s, want %s", tt.seconds, result, tt.expected)
		}
	}
}

func TestSaveSlotListSelectedSlot(t *testing.T) {
	slots := []SlotDisplayInfo{
		{SlotID: 1, IsEmpty: true},
		{SlotID: 2, IsEmpty: false, Chapter: 3},
		{SlotID: 3, IsEmpty: true},
	}

	list := NewSaveSlotList(slots, ModeSave)

	// Initial selection should be first slot
	selected := list.SelectedSlot()
	if selected.SlotID != 1 {
		t.Errorf("Expected initial selection to be slot 1, got %d", selected.SlotID)
	}
}

func TestSaveSlotListCanSelectEmpty(t *testing.T) {
	slots := []SlotDisplayInfo{
		{SlotID: 1, IsEmpty: true},
		{SlotID: 2, IsEmpty: false, Chapter: 3},
	}

	// In save mode, can select empty slots
	saveList := NewSaveSlotList(slots, ModeSave)
	slot := saveList.SelectedSlot()
	if !saveList.CanSelectSlot(slot) {
		t.Error("Should be able to select empty slot in save mode")
	}

	// In load mode, cannot select empty slots
	loadList := NewSaveSlotList(slots, ModeLoad)
	if loadList.CanSelectSlot(slots[0]) {
		t.Error("Should not be able to select empty slot in load mode")
	}
}

func TestConfirmDialogState(t *testing.T) {
	slots := []SlotDisplayInfo{
		{SlotID: 1, IsEmpty: false, Chapter: 2},
	}

	list := NewSaveSlotList(slots, ModeSave)

	if list.ShowingConfirmDialog {
		t.Error("Confirm dialog should not be shown initially")
	}

	list.ShowConfirmDialog()
	if !list.ShowingConfirmDialog {
		t.Error("Confirm dialog should be shown after ShowConfirmDialog()")
	}

	list.HideConfirmDialog()
	if list.ShowingConfirmDialog {
		t.Error("Confirm dialog should be hidden after HideConfirmDialog()")
	}
}

func TestSlotDisplayInfoFilterValue(t *testing.T) {
	slot := SlotDisplayInfo{SlotID: 2, IsEmpty: false, Chapter: 3, Location: "醫院"}
	filter := slot.FilterValue()
	if filter == "" {
		t.Error("FilterValue should not be empty")
	}
}
