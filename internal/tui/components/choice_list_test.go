package components

import (
	"testing"
)

func TestNewChoiceList(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	list := NewChoiceList(choices)

	if list == nil {
		t.Fatal("NewChoiceList should not return nil")
	}
	if len(list.choices) != 3 {
		t.Errorf("Choices length = %d, want 3", len(list.choices))
	}
	if list.selected != 0 {
		t.Errorf("Initial selected = %d, want 0", list.selected)
	}
}

func TestChoiceList_GetSelected(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	list := NewChoiceList(choices)

	if list.GetSelected() != 0 {
		t.Errorf("GetSelected = %d, want 0", list.GetSelected())
	}
}

func TestChoiceList_SetSelected(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	list := NewChoiceList(choices)

	list.SetSelected(1)

	if list.GetSelected() != 1 {
		t.Errorf("After SetSelected(1), selected = %d, want 1", list.GetSelected())
	}

	// Invalid selection should be ignored
	list.SetSelected(10)
	if list.GetSelected() != 1 {
		t.Error("Invalid selection should be ignored")
	}

	list.SetSelected(-1)
	if list.GetSelected() != 1 {
		t.Error("Negative selection should be ignored")
	}
}

func TestChoiceList_GetChoice(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	list := NewChoiceList(choices)

	choice, ok := list.GetChoice(0)
	if !ok || choice != "Choice 1" {
		t.Errorf("GetChoice(0) = %q, %v, want 'Choice 1', true", choice, ok)
	}

	choice, ok = list.GetChoice(10)
	if ok {
		t.Error("GetChoice(10) should return false")
	}
}

func TestChoiceList_GetSelectedChoice(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	list := NewChoiceList(choices)

	choice, ok := list.GetSelectedChoice()
	if !ok || choice != "Choice 1" {
		t.Errorf("GetSelectedChoice = %q, %v, want 'Choice 1', true", choice, ok)
	}

	list.SetSelected(2)
	choice, ok = list.GetSelectedChoice()
	if !ok || choice != "Choice 3" {
		t.Errorf("GetSelectedChoice = %q, %v, want 'Choice 3', true", choice, ok)
	}
}

func TestChoiceList_Count(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	list := NewChoiceList(choices)

	if list.Count() != 3 {
		t.Errorf("Count = %d, want 3", list.Count())
	}

	emptyList := NewChoiceList([]string{})
	if emptyList.Count() != 0 {
		t.Errorf("Empty list Count = %d, want 0", emptyList.Count())
	}
}

func TestChoiceList_SetChoices(t *testing.T) {
	list := NewChoiceList([]string{"Old 1", "Old 2"})

	newChoices := []string{"New 1", "New 2", "New 3"}
	list.SetChoices(newChoices)

	if list.Count() != 3 {
		t.Errorf("After SetChoices, count = %d, want 3", list.Count())
	}
	if list.GetSelected() != 0 {
		t.Error("After SetChoices, selected should reset to 0")
	}

	choice, _ := list.GetChoice(0)
	if choice != "New 1" {
		t.Errorf("First choice = %q, want 'New 1'", choice)
	}
}

func TestChoiceList_SetHighlighted(t *testing.T) {
	list := NewChoiceList([]string{"Choice 1", "Choice 2"})

	list.SetHighlighted(true)
	if !list.highlighted {
		t.Error("SetHighlighted(true) should set highlighted to true")
	}

	list.SetHighlighted(false)
	if list.highlighted {
		t.Error("SetHighlighted(false) should set highlighted to false")
	}
}

func TestChoiceList_IsHighlighted(t *testing.T) {
	list := NewChoiceList([]string{"Choice 1", "Choice 2"})

	if list.IsHighlighted() {
		t.Error("Initial state should not be highlighted")
	}

	list.SetHighlighted(true)
	if !list.IsHighlighted() {
		t.Error("After SetHighlighted(true), should be highlighted")
	}
}

func TestChoiceList_Reset(t *testing.T) {
	list := NewChoiceList([]string{"Choice 1", "Choice 2", "Choice 3"})

	list.SetSelected(2)
	list.SetHighlighted(true)

	list.Reset()

	if list.GetSelected() != 0 {
		t.Error("After Reset, selected should be 0")
	}
	if list.IsHighlighted() {
		t.Error("After Reset, should not be highlighted")
	}
}

func TestChoiceList_EmptyChoices(t *testing.T) {
	list := NewChoiceList([]string{})

	if list.Count() != 0 {
		t.Error("Empty list should have count 0")
	}

	_, ok := list.GetSelectedChoice()
	if ok {
		t.Error("Empty list should not return a selected choice")
	}
}
