package game

import (
	"testing"
	"time"
)

func TestClueStatusString(t *testing.T) {
	tests := []struct {
		status   ClueStatus
		expected string
	}{
		{ClueStatusDiscovered, "已發現"},
		{ClueStatusMissed, "錯過"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("ClueStatus.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestNewClueInfo(t *testing.T) {
	clue := NewClueInfo("clue-1", "門把在月光下泛著紅光", "rule-1")

	if clue.ID != "clue-1" {
		t.Errorf("ID = %s, want clue-1", clue.ID)
	}
	if clue.Content != "門把在月光下泛著紅光" {
		t.Errorf("Content = %s, want 門把在月光下泛著紅光", clue.Content)
	}
	if clue.RuleID != "rule-1" {
		t.Errorf("RuleID = %s, want rule-1", clue.RuleID)
	}
	if clue.Status != ClueStatusMissed {
		t.Errorf("Status = %v, want ClueStatusMissed", clue.Status)
	}
	if clue.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewHallucinationLog(t *testing.T) {
	log := NewHallucinationLog("跟隨那個影子", 15, 3)

	if log.OptionText != "跟隨那個影子" {
		t.Errorf("OptionText = %s, want 跟隨那個影子", log.OptionText)
	}
	if log.SANValue != 15 {
		t.Errorf("SANValue = %d, want 15", log.SANValue)
	}
	if log.Chapter != 3 {
		t.Errorf("Chapter = %d, want 3", log.Chapter)
	}
	if log.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewDecisionPoint(t *testing.T) {
	options := []string{"打開門", "離開", "檢查門"}
	decision := NewDecisionPoint(2, options, 0)

	if decision.Chapter != 2 {
		t.Errorf("Chapter = %d, want 2", decision.Chapter)
	}
	if len(decision.Options) != 3 {
		t.Errorf("Options length = %d, want 3", len(decision.Options))
	}
	if decision.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", decision.SelectedIndex)
	}
	if decision.SelectedText != "打開門" {
		t.Errorf("SelectedText = %s, want 打開門", decision.SelectedText)
	}
}

func TestNewDecisionPointInvalidIndex(t *testing.T) {
	options := []string{"A", "B"}
	decision := NewDecisionPoint(1, options, 5) // Invalid index

	if decision.SelectedText != "" {
		t.Errorf("SelectedText should be empty for invalid index, got %s", decision.SelectedText)
	}
}

func TestNewCheckpointInfo(t *testing.T) {
	checkpoint := NewCheckpointInfo("cp-1", 3, 80, 60)

	if checkpoint.ID != "cp-1" {
		t.Errorf("ID = %s, want cp-1", checkpoint.ID)
	}
	if checkpoint.Chapter != 3 {
		t.Errorf("Chapter = %d, want 3", checkpoint.Chapter)
	}
	if checkpoint.HP != 80 {
		t.Errorf("HP = %d, want 80", checkpoint.HP)
	}
	if checkpoint.SAN != 60 {
		t.Errorf("SAN = %d, want 60", checkpoint.SAN)
	}
}

func TestNewRuleReveal(t *testing.T) {
	reveal := NewRuleReveal("rule-1", "時間規則", "不要在夜晚開門", "即死")

	if reveal.RuleID != "rule-1" {
		t.Errorf("RuleID = %s, want rule-1", reveal.RuleID)
	}
	if reveal.RuleType != "時間規則" {
		t.Errorf("RuleType = %s, want 時間規則", reveal.RuleType)
	}
	if reveal.TriggerCondition != "不要在夜晚開門" {
		t.Errorf("TriggerCondition = %s, want 不要在夜晚開門", reveal.TriggerCondition)
	}
	if reveal.ConsequenceType != "即死" {
		t.Errorf("ConsequenceType = %s, want 即死", reveal.ConsequenceType)
	}
	if len(reveal.DiscoveredClues) != 0 {
		t.Error("DiscoveredClues should be empty initially")
	}
	if len(reveal.MissedClues) != 0 {
		t.Error("MissedClues should be empty initially")
	}
}

func TestNewDebriefData(t *testing.T) {
	data := NewDebriefData()

	if data.DeathInfo != nil {
		t.Error("DeathInfo should be nil initially")
	}
	if len(data.TriggeredRules) != 0 {
		t.Error("TriggeredRules should be empty")
	}
	if len(data.AllClues) != 0 {
		t.Error("AllClues should be empty")
	}
	if len(data.HallucinationLogs) != 0 {
		t.Error("HallucinationLogs should be empty")
	}
	if len(data.KeyDecisions) != 0 {
		t.Error("KeyDecisions should be empty")
	}
	if len(data.Checkpoints) != 0 {
		t.Error("Checkpoints should be empty")
	}
}

func TestDebriefDataClueOperations(t *testing.T) {
	data := NewDebriefData()

	// Add clues
	clue1 := NewClueInfo("clue-1", "線索1", "rule-1")
	clue2 := NewClueInfo("clue-2", "線索2", "rule-1")
	clue3 := NewClueInfo("clue-3", "線索3", "rule-2")
	data.AddClue(clue1)
	data.AddClue(clue2)
	data.AddClue(clue3)

	if len(data.AllClues) != 3 {
		t.Errorf("AllClues count = %d, want 3", len(data.AllClues))
	}

	// Initially all missed
	missed := data.GetMissedClues()
	if len(missed) != 3 {
		t.Errorf("Missed clues count = %d, want 3", len(missed))
	}

	// Mark one discovered
	if !data.MarkClueDiscovered("clue-1") {
		t.Error("MarkClueDiscovered should return true")
	}

	discovered := data.GetDiscoveredClues()
	if len(discovered) != 1 {
		t.Errorf("Discovered clues count = %d, want 1", len(discovered))
	}

	missed = data.GetMissedClues()
	if len(missed) != 2 {
		t.Errorf("Missed clues count = %d, want 2", len(missed))
	}

	// Test GetCluesByRuleID
	rule1Clues := data.GetCluesByRuleID("rule-1")
	if len(rule1Clues) != 2 {
		t.Errorf("Rule-1 clues count = %d, want 2", len(rule1Clues))
	}
}

func TestDebriefDataMarkClueDiscoveredNotFound(t *testing.T) {
	data := NewDebriefData()

	if data.MarkClueDiscovered("nonexistent") {
		t.Error("MarkClueDiscovered should return false for nonexistent clue")
	}
}

func TestDebriefDataHallucinationLogs(t *testing.T) {
	data := NewDebriefData()

	log1 := NewHallucinationLog("選項1", 20, 1)
	log2 := NewHallucinationLog("選項2", 15, 2)
	data.AddHallucinationLog(log1)
	data.AddHallucinationLog(log2)

	if data.GetHallucinationCount() != 2 {
		t.Errorf("HallucinationCount = %d, want 2", data.GetHallucinationCount())
	}
}

func TestDebriefDataDecisions(t *testing.T) {
	data := NewDebriefData()

	decision1 := NewDecisionPoint(1, []string{"A", "B"}, 0)
	decision2 := NewDecisionPoint(2, []string{"C", "D"}, 1)
	decision2.IsSignificant = true
	decision3 := NewDecisionPoint(3, []string{"E", "F"}, 0)

	data.AddDecision(decision1)
	data.AddDecision(decision2)
	data.AddDecision(decision3)

	if len(data.KeyDecisions) != 3 {
		t.Errorf("KeyDecisions count = %d, want 3", len(data.KeyDecisions))
	}

	significant := data.GetSignificantDecisions()
	if len(significant) != 1 {
		t.Errorf("Significant decisions count = %d, want 1", len(significant))
	}
}

func TestDebriefDataCheckpoints(t *testing.T) {
	data := NewDebriefData()

	// Add checkpoints
	for i := 1; i <= 5; i++ {
		cp := NewCheckpointInfo("cp-"+string(rune('0'+i)), i, 100-i*10, 100-i*5)
		data.AddCheckpoint(cp)
	}

	// Should only keep last 3
	if len(data.Checkpoints) != 3 {
		t.Errorf("Checkpoints count = %d, want 3 (max)", len(data.Checkpoints))
	}

	latest := data.GetLatestCheckpoint()
	if latest == nil {
		t.Fatal("GetLatestCheckpoint should not return nil")
	}
	if latest.Chapter != 5 {
		t.Errorf("Latest checkpoint chapter = %d, want 5", latest.Chapter)
	}
}

func TestDebriefDataGetLatestCheckpointEmpty(t *testing.T) {
	data := NewDebriefData()

	if data.GetLatestCheckpoint() != nil {
		t.Error("GetLatestCheckpoint should return nil for empty checkpoints")
	}
}

func TestDebriefDataCanRollback(t *testing.T) {
	data := NewDebriefData()

	// No checkpoints, easy mode
	data.Difficulty = DifficultyEasy
	if data.CanRollback() {
		t.Error("CanRollback should be false without checkpoints")
	}

	// Add checkpoint
	data.AddCheckpoint(NewCheckpointInfo("cp-1", 1, 100, 100))
	if !data.CanRollback() {
		t.Error("CanRollback should be true with checkpoint on Easy")
	}

	// Change to hard mode
	data.Difficulty = DifficultyHard
	if data.CanRollback() {
		t.Error("CanRollback should be false on Hard mode")
	}

	// Hell mode
	data.Difficulty = DifficultyHell
	if data.CanRollback() {
		t.Error("CanRollback should be false on Hell mode")
	}
}

func TestDebriefDataClearCheckpoints(t *testing.T) {
	data := NewDebriefData()
	data.AddCheckpoint(NewCheckpointInfo("cp-1", 1, 100, 100))
	data.AddCheckpoint(NewCheckpointInfo("cp-2", 2, 90, 90))

	data.ClearCheckpoints()

	if len(data.Checkpoints) != 0 {
		t.Errorf("Checkpoints should be empty after clear, got %d", len(data.Checkpoints))
	}
}

func TestDebriefDataGetDeathSummary(t *testing.T) {
	data := NewDebriefData()

	// No death info
	if data.GetDeathSummary() != "死因不明" {
		t.Error("Should return '死因不明' for nil DeathInfo")
	}

	// HP death
	data.DeathInfo = NewDeathInfo(DeathTypeHP)
	summary := data.GetDeathSummary()
	if summary != "你的體力完全耗盡，無法再繼續前進。" {
		t.Errorf("HP death summary = %s", summary)
	}

	// SAN death
	data.DeathInfo = NewDeathInfo(DeathTypeSAN)
	summary = data.GetDeathSummary()
	if summary != "你的理智崩潰了，被恐懼和瘋狂吞噬。" {
		t.Errorf("SAN death summary = %s", summary)
	}

	// Rule death without rules
	data.DeathInfo = NewDeathInfo(DeathTypeRule)
	summary = data.GetDeathSummary()
	if summary != "你違反了隱藏的規則而遭受懲罰。" {
		t.Errorf("Rule death (no rules) summary = %s", summary)
	}

	// Rule death with rules
	data.AddRuleReveal(NewRuleReveal("rule-1", "時間規則", "不要在夜晚開門", "即死"))
	summary = data.GetDeathSummary()
	if summary != "你違反了隱藏的規則：「不要在夜晚開門」" {
		t.Errorf("Rule death (with rules) summary = %s", summary)
	}
}

func TestDebriefCollector(t *testing.T) {
	collector := NewDebriefCollector()

	// Test SetDifficulty
	collector.SetDifficulty(DifficultyHard)
	if collector.GetData().Difficulty != DifficultyHard {
		t.Error("Difficulty not set correctly")
	}

	// Test RecordClue
	clue := NewClueInfo("clue-1", "測試線索", "rule-1")
	collector.RecordClue(clue)
	if len(collector.GetData().AllClues) != 1 {
		t.Error("Clue not recorded")
	}

	// Test RecordClueDiscovered
	collector.RecordClueDiscovered("clue-1")
	if collector.GetData().AllClues[0].Status != ClueStatusDiscovered {
		t.Error("Clue not marked as discovered")
	}

	// Test RecordHallucination
	hallucination := NewHallucinationLog("幻覺選項", 10, 2)
	collector.RecordHallucination(hallucination)
	if collector.GetData().GetHallucinationCount() != 1 {
		t.Error("Hallucination not recorded")
	}

	// Test RecordDecision
	decision := NewDecisionPoint(1, []string{"A", "B"}, 0)
	collector.RecordDecision(decision)
	if len(collector.GetData().KeyDecisions) != 1 {
		t.Error("Decision not recorded")
	}

	// Test RecordCheckpoint
	checkpoint := NewCheckpointInfo("cp-1", 1, 100, 100)
	collector.RecordCheckpoint(checkpoint)
	if len(collector.GetData().Checkpoints) != 1 {
		t.Error("Checkpoint not recorded")
	}

	// Test RecordRuleViolation
	reveal := NewRuleReveal("rule-1", "時間規則", "不要開門", "即死")
	collector.RecordRuleViolation(reveal)
	if len(collector.GetData().TriggeredRules) != 1 {
		t.Error("Rule violation not recorded")
	}

	// Test RecordDeath
	death := NewDeathInfo(DeathTypeHP)
	collector.RecordDeath(death)
	if collector.GetData().DeathInfo == nil {
		t.Error("Death not recorded")
	}

	// Test Reset
	collector.Reset()
	data := collector.GetData()
	if len(data.AllClues) != 0 || len(data.KeyDecisions) != 0 {
		t.Error("Reset did not clear data")
	}
}

func TestClueInfoTimestamp(t *testing.T) {
	before := time.Now()
	clue := NewClueInfo("clue-1", "test", "rule-1")
	after := time.Now()

	if clue.Timestamp.Before(before) || clue.Timestamp.After(after) {
		t.Error("Timestamp should be set to current time")
	}
}
