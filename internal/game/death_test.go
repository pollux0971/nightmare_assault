package game

import (
	"testing"
)

func TestDeathTypeString(t *testing.T) {
	tests := []struct {
		deathType DeathType
		expected  string
	}{
		{DeathTypeHP, "體力耗盡"},
		{DeathTypeSAN, "理智崩潰"},
		{DeathTypeRule, "規則懲罰"},
	}

	for _, tt := range tests {
		if got := tt.deathType.String(); got != tt.expected {
			t.Errorf("DeathType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestDeathTypeEnglishName(t *testing.T) {
	tests := []struct {
		deathType DeathType
		expected  string
	}{
		{DeathTypeHP, "HP Depleted"},
		{DeathTypeSAN, "Sanity Collapse"},
		{DeathTypeRule, "Rule Violation"},
	}

	for _, tt := range tests {
		if got := tt.deathType.EnglishName(); got != tt.expected {
			t.Errorf("DeathType.EnglishName() = %v, want %v", got, tt.expected)
		}
	}
}

func TestNewDeathInfo(t *testing.T) {
	info := NewDeathInfo(DeathTypeSAN)

	if info.Type != DeathTypeSAN {
		t.Errorf("Type = %v, want DeathTypeSAN", info.Type)
	}
	if info.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestDeathInfoIsInsanity(t *testing.T) {
	infoSAN := NewDeathInfo(DeathTypeSAN)
	if !infoSAN.IsInsanity() {
		t.Error("DeathTypeSAN.IsInsanity() should return true")
	}

	infoHP := NewDeathInfo(DeathTypeHP)
	if infoHP.IsInsanity() {
		t.Error("DeathTypeHP.IsInsanity() should return false")
	}
}

func TestDeathInfoIsRuleViolation(t *testing.T) {
	infoRule := NewDeathInfo(DeathTypeRule)
	infoRule.TriggeringRuleID = "rule-123"
	if !infoRule.IsRuleViolation() {
		t.Error("DeathTypeRule.IsRuleViolation() should return true")
	}

	infoHP := NewDeathInfo(DeathTypeHP)
	if infoHP.IsRuleViolation() {
		t.Error("DeathTypeHP.IsRuleViolation() should return false")
	}
}

func TestNewDeathState(t *testing.T) {
	state := NewDeathState()

	if state.Deaths == nil {
		t.Error("Deaths should not be nil")
	}
	if len(state.Deaths) != 0 {
		t.Errorf("Deaths should be empty, got %d", len(state.Deaths))
	}
	if state.CurrentDeath != nil {
		t.Error("CurrentDeath should be nil")
	}
}

func TestDeathStateRecordDeath(t *testing.T) {
	state := NewDeathState()
	death1 := NewDeathInfo(DeathTypeHP)
	death2 := NewDeathInfo(DeathTypeSAN)

	state.RecordDeath(death1)
	if len(state.Deaths) != 1 {
		t.Errorf("Deaths count = %d, want 1", len(state.Deaths))
	}
	if state.CurrentDeath != death1 {
		t.Error("CurrentDeath should be death1")
	}

	state.RecordDeath(death2)
	if len(state.Deaths) != 2 {
		t.Errorf("Deaths count = %d, want 2", len(state.Deaths))
	}
	if state.CurrentDeath != death2 {
		t.Error("CurrentDeath should be death2")
	}
}

func TestDeathStateGetDeathCount(t *testing.T) {
	state := NewDeathState()

	if state.GetDeathCount() != 0 {
		t.Errorf("GetDeathCount() = %d, want 0", state.GetDeathCount())
	}

	state.RecordDeath(NewDeathInfo(DeathTypeHP))
	state.RecordDeath(NewDeathInfo(DeathTypeSAN))
	state.RecordDeath(NewDeathInfo(DeathTypeRule))

	if state.GetDeathCount() != 3 {
		t.Errorf("GetDeathCount() = %d, want 3", state.GetDeathCount())
	}
}

func TestDeathStateGetLastDeath(t *testing.T) {
	state := NewDeathState()

	if state.GetLastDeath() != nil {
		t.Error("GetLastDeath() should return nil for empty state")
	}

	death1 := NewDeathInfo(DeathTypeHP)
	death1.Chapter = 1
	state.RecordDeath(death1)

	death2 := NewDeathInfo(DeathTypeSAN)
	death2.Chapter = 2
	state.RecordDeath(death2)

	last := state.GetLastDeath()
	if last != death2 {
		t.Error("GetLastDeath() should return most recent death")
	}
	if last.Chapter != 2 {
		t.Errorf("Last death chapter = %d, want 2", last.Chapter)
	}
}

func TestDeathStateClearCurrentDeath(t *testing.T) {
	state := NewDeathState()
	state.RecordDeath(NewDeathInfo(DeathTypeHP))

	if state.CurrentDeath == nil {
		t.Error("CurrentDeath should not be nil after recording")
	}

	state.ClearCurrentDeath()

	if state.CurrentDeath != nil {
		t.Error("CurrentDeath should be nil after clearing")
	}

	// Deaths list should still contain the death
	if len(state.Deaths) != 1 {
		t.Errorf("Deaths count = %d, want 1 (history preserved)", len(state.Deaths))
	}
}

func TestDeathInfoWithFullData(t *testing.T) {
	info := NewDeathInfo(DeathTypeRule)
	info.Chapter = 5
	info.FinalHP = 0
	info.FinalSAN = 45
	info.TriggeringRuleID = "basement-rule"
	info.LastAction = "進入地下室"
	info.Location = "廢棄醫院"
	info.Narrative = "你打開了不該打開的門..."

	if info.Chapter != 5 {
		t.Errorf("Chapter = %d, want 5", info.Chapter)
	}
	if info.TriggeringRuleID != "basement-rule" {
		t.Errorf("TriggeringRuleID = %s, want basement-rule", info.TriggeringRuleID)
	}
	if info.Location != "廢棄醫院" {
		t.Errorf("Location = %s, want 廢棄醫院", info.Location)
	}
	if !info.IsRuleViolation() {
		t.Error("Should be a rule violation")
	}
}
