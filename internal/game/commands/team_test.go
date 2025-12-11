package commands

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

func TestTeamCommand_Name(t *testing.T) {
	cmd := &TeamCommand{}
	if cmd.Name() != "team" {
		t.Errorf("Expected command name 'team', got '%s'", cmd.Name())
	}
}

func TestTeamCommand_Help(t *testing.T) {
	cmd := &TeamCommand{}
	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
}

func TestTeamCommand_ExecuteNoTeammates(t *testing.T) {
	cmd := NewTeamCommand(nil)
	output, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute should not return error: %v", err)
	}

	if !strings.Contains(output, "沒有隊友") || !strings.Contains(output, "No teammates") {
		t.Errorf("Output should indicate no teammates, got: %s", output)
	}
}

func TestTeamCommand_ExecuteWithTeammates(t *testing.T) {
	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
		npc.NewTeammate("tm-002", "王芳", npc.ArchetypeVictim),
	}
	teammates[0].Location = "大廳"
	teammates[0].Inventory = []npc.Item{
		{Name: "手電筒", Description: "照明工具"},
	}
	teammates[1].Location = "走廊"
	teammates[1].HP = 80

	cmd := NewTeamCommand(teammates)
	output, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute should not return error: %v", err)
	}

	// Should contain teammate names
	if !strings.Contains(output, "李明") {
		t.Error("Output should contain teammate name '李明'")
	}
	if !strings.Contains(output, "王芳") {
		t.Error("Output should contain teammate name '王芳'")
	}

	// Should contain HP
	if !strings.Contains(output, "100") && !strings.Contains(output, "80") {
		t.Error("Output should contain HP values")
	}

	// Should contain locations
	if !strings.Contains(output, "大廳") {
		t.Error("Output should contain location '大廳'")
	}

	// Should contain inventory
	if !strings.Contains(output, "手電筒") {
		t.Error("Output should contain inventory item '手電筒'")
	}
}

func TestTeamCommand_ExecuteWithDeadTeammate(t *testing.T) {
	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
	}
	teammates[0].Status.Alive = false
	teammates[0].Status.Condition = "dead"

	cmd := NewTeamCommand(teammates)
	output, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute should not return error: %v", err)
	}

	// Should indicate dead status
	if !strings.Contains(output, "dead") && !strings.Contains(output, "已死亡") {
		t.Error("Output should indicate teammate is dead")
	}
}

func TestTeamCommand_ExecuteWithInjuredTeammate(t *testing.T) {
	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
	}
	teammates[0].HP = 30
	teammates[0].Status.Condition = "injured"

	cmd := NewTeamCommand(teammates)
	output, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute should not return error: %v", err)
	}

	// Should show low HP
	if !strings.Contains(output, "30") {
		t.Error("Output should show HP 30")
	}

	// Should indicate injured condition
	if !strings.Contains(output, "injured") && !strings.Contains(output, "受傷") {
		t.Error("Output should indicate injured condition")
	}
}
