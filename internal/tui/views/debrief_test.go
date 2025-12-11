package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func createTestDebriefData() *game.DebriefData {
	data := game.NewDebriefData()
	data.SetDeathInfo(game.NewDeathInfo(game.DeathTypeRule))
	data.DeathInfo.Chapter = 5
	data.DeathInfo.FinalHP = 0
	data.DeathInfo.FinalSAN = 45

	// Add triggered rule
	rule := game.NewRuleReveal("rule-1", "時間規則", "不要在夜晚開門", "即死")
	rule.DiscoveredClues = []string{"夜裡總有奇怪的聲音"}
	rule.MissedClues = []string{"門把在月光下泛著紅光"}
	rule.Explanation = "這個規則暗示了夜晚的危險"
	data.AddRuleReveal(rule)

	// Add clues
	clue1 := game.NewClueInfo("clue-1", "門把在月光下泛著紅光", "rule-1")
	clue1.Chapter = 3
	clue1.Context = "你走進走廊..."
	data.AddClue(clue1)

	clue2 := game.NewClueInfo("clue-2", "夜裡總有奇怪的聲音", "rule-1")
	clue2.Chapter = 2
	data.MarkClueDiscovered("clue-2")
	data.AddClue(clue2)

	// Add decision
	decision := game.NewDecisionPoint(4, []string{"打開門", "離開"}, 0)
	decision.IsSignificant = true
	decision.Consequence = "你觸發了規則"
	data.AddDecision(decision)

	return data
}

func TestNewDebriefModel(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	if model.data != data {
		t.Error("Data not set correctly")
	}
	if model.currentSection != SectionSummary {
		t.Errorf("Initial section = %v, want SectionSummary", model.currentSection)
	}
	if model.selectedOption != 0 {
		t.Errorf("Initial selectedOption = %d, want 0", model.selectedOption)
	}
}

func TestNewDebriefModelWithRollback(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyEasy
	data.AddCheckpoint(game.NewCheckpointInfo("cp-1", 3, 80, 60))

	model := NewDebriefModel(data)

	// Should have rollback option
	if len(model.options) != 3 {
		t.Errorf("Options count = %d, want 3 (with rollback)", len(model.options))
	}
	if model.options[0] != DebriefActionRollback {
		t.Error("First option should be DebriefActionRollback")
	}
}

func TestNewDebriefModelWithoutRollback(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyHard

	model := NewDebriefModel(data)

	// Should not have rollback option
	if len(model.options) != 2 {
		t.Errorf("Options count = %d, want 2 (without rollback)", len(model.options))
	}
	if model.options[0] != DebriefActionNewGame {
		t.Error("First option should be DebriefActionNewGame")
	}
}

func TestDebriefModelInit(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init should return nil for debrief view")
	}
}

func TestDebriefModelWindowSizeMsg(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	m, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = m.(DebriefModel)

	if model.width != 100 {
		t.Errorf("width = %d, want 100", model.width)
	}
	if model.height != 50 {
		t.Errorf("height = %d, want 50", model.height)
	}
}

func TestDebriefModelKeyNavigation(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	// Move down through sections
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DebriefModel)
	if model.currentSection != SectionRules {
		t.Errorf("After down, section = %v, want SectionRules", model.currentSection)
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DebriefModel)
	if model.currentSection != SectionClues {
		t.Errorf("After down, section = %v, want SectionClues", model.currentSection)
	}

	// Move up
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = m.(DebriefModel)
	if model.currentSection != SectionRules {
		t.Errorf("After up, section = %v, want SectionRules", model.currentSection)
	}
}

func TestDebriefModelTabCycle(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	// Tab through all sections
	sections := []DebriefSection{
		SectionRules, SectionClues, SectionDecisions, SectionOptions, SectionSummary,
	}

	for _, expected := range sections {
		m, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = m.(DebriefModel)
		if model.currentSection != expected {
			t.Errorf("After tab, section = %v, want %v", model.currentSection, expected)
		}
	}
}

func TestDebriefModelShiftTabCycle(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	// Start at SectionSummary

	// Shift+Tab should go to Options
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = m.(DebriefModel)
	if model.currentSection != SectionOptions {
		t.Errorf("After shift+tab from Summary, section = %v, want SectionOptions", model.currentSection)
	}
}

func TestDebriefModelOptionNavigation(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyHard // No rollback
	model := NewDebriefModel(data)
	model.currentSection = SectionOptions

	// Initial selection
	if model.selectedOption != 0 {
		t.Errorf("Initial selectedOption = %d, want 0", model.selectedOption)
	}

	// Move down
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DebriefModel)
	if model.selectedOption != 1 {
		t.Errorf("After down, selectedOption = %d, want 1", model.selectedOption)
	}

	// Move down again (should stay at 1 - last option)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DebriefModel)
	if model.selectedOption != 1 {
		t.Errorf("At bottom, selectedOption = %d, want 1", model.selectedOption)
	}

	// Move up
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = m.(DebriefModel)
	if model.selectedOption != 0 {
		t.Errorf("After up, selectedOption = %d, want 0", model.selectedOption)
	}
}

func TestDebriefModelSelectNewGame(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyHard
	model := NewDebriefModel(data)
	model.currentSection = SectionOptions
	model.selectedOption = 0 // NewGame

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(DebriefSelectMsg)
	if !ok {
		t.Fatal("Command should return DebriefSelectMsg")
	}
	if selectMsg.Action != DebriefActionNewGame {
		t.Errorf("Action = %v, want DebriefActionNewGame", selectMsg.Action)
	}
}

func TestDebriefModelSelectMenu(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyHard
	model := NewDebriefModel(data)
	model.currentSection = SectionOptions
	model.selectedOption = 1 // Menu

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(DebriefSelectMsg)
	if !ok {
		t.Fatal("Command should return DebriefSelectMsg")
	}
	if selectMsg.Action != DebriefActionMenu {
		t.Errorf("Action = %v, want DebriefActionMenu", selectMsg.Action)
	}
}

func TestDebriefModelSelectRollback(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyEasy
	data.AddCheckpoint(game.NewCheckpointInfo("cp-1", 3, 80, 60))
	model := NewDebriefModel(data)
	model.currentSection = SectionOptions
	model.selectedOption = 0 // Rollback (first option)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(DebriefSelectMsg)
	if !ok {
		t.Fatal("Command should return DebriefSelectMsg")
	}
	if selectMsg.Action != DebriefActionRollback {
		t.Errorf("Action = %v, want DebriefActionRollback", selectMsg.Action)
	}
}

func TestDebriefModelEscapeToMenu(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Escape should return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(DebriefSelectMsg)
	if !ok {
		t.Fatal("Command should return DebriefSelectMsg")
	}
	if selectMsg.Action != DebriefActionMenu {
		t.Errorf("Escape action = %v, want DebriefActionMenu", selectMsg.Action)
	}
}

func TestDebriefModelToggleExpansion(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.currentSection = SectionRules

	// Initial state - not expanded
	if model.expandedRules[0] {
		t.Error("Rule should not be expanded initially")
	}

	// Press Enter to expand
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(DebriefModel)

	if !model.expandedRules[0] {
		t.Error("Rule should be expanded after Enter")
	}

	// Press Enter again to collapse
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(DebriefModel)

	if model.expandedRules[0] {
		t.Error("Rule should be collapsed after second Enter")
	}
}

func TestDebriefModelExpandAll(t *testing.T) {
	data := createTestDebriefData()
	// Add more rules
	data.AddRuleReveal(game.NewRuleReveal("rule-2", "行為規則", "不要跑步", "傷害"))
	model := NewDebriefModel(data)
	model.currentSection = SectionRules

	// Press 'e' to expand all
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = m.(DebriefModel)

	if !model.expandedRules[0] || !model.expandedRules[1] {
		t.Error("All rules should be expanded after 'e'")
	}
}

func TestDebriefModelCollapseAll(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.currentSection = SectionRules
	model.expandedRules[0] = true

	// Press 'c' to collapse all
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = m.(DebriefModel)

	if model.expandedRules[0] {
		t.Error("All rules should be collapsed after 'c'")
	}
}

func TestDebriefModelViewNoData(t *testing.T) {
	model := NewDebriefModel(nil)
	model.width = 80
	model.height = 24

	view := model.View()

	if !strings.Contains(view, "沒有可用的覆盤資料") {
		t.Error("View should show 'no data' message when data is nil")
	}
}

func TestDebriefModelViewWithData(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.width = 80
	model.height = 40

	view := model.View()

	// Should contain title
	if !strings.Contains(view, "死亡覆盤") {
		t.Error("View should contain title '死亡覆盤'")
	}

	// Should contain section headers
	if !strings.Contains(view, "死因摘要") {
		t.Error("View should contain '死因摘要'")
	}
	if !strings.Contains(view, "觸發的規則") {
		t.Error("View should contain '觸發的規則'")
	}
	if !strings.Contains(view, "錯過的線索") {
		t.Error("View should contain '錯過的線索'")
	}
	if !strings.Contains(view, "關鍵決策點") {
		t.Error("View should contain '關鍵決策點'")
	}

	// Should contain options
	if !strings.Contains(view, "開始新遊戲") {
		t.Error("View should contain '開始新遊戲'")
	}
	if !strings.Contains(view, "返回主選單") {
		t.Error("View should contain '返回主選單'")
	}
}

func TestDebriefModelViewExpandedRule(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.width = 80
	model.height = 50
	model.expandedRules[0] = true

	view := model.View()

	// Should show rule details
	if !strings.Contains(view, "不要在夜晚開門") {
		t.Error("View should contain rule trigger condition")
	}
	// Should show discovered clues
	if !strings.Contains(view, "已發現的線索") {
		t.Error("View should show discovered clues when expanded")
	}
	// Should show missed clues
	if !strings.Contains(view, "錯過的線索") {
		t.Error("View should show missed clues when expanded")
	}
}

func TestDebriefModelSetData(t *testing.T) {
	model := NewDebriefModel(nil)

	newData := createTestDebriefData()
	newData.Difficulty = game.DifficultyEasy
	newData.AddCheckpoint(game.NewCheckpointInfo("cp-1", 1, 100, 100))

	model.SetData(newData)

	if model.data != newData {
		t.Error("SetData should update data")
	}
	if len(model.options) != 3 {
		t.Errorf("SetData should rebuild options, got %d", len(model.options))
	}
	if model.options[0] != DebriefActionRollback {
		t.Error("Options should include rollback for easy mode with checkpoint")
	}
}

func TestDebriefModelSetSize(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	model.SetSize(120, 60)

	if model.width != 120 {
		t.Errorf("width = %d, want 120", model.width)
	}
	if model.height != 60 {
		t.Errorf("height = %d, want 60", model.height)
	}
}

func TestDebriefModelGetCurrentSection(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	if model.GetCurrentSection() != SectionSummary {
		t.Errorf("GetCurrentSection = %v, want SectionSummary", model.GetCurrentSection())
	}

	model.currentSection = SectionRules
	if model.GetCurrentSection() != SectionRules {
		t.Errorf("GetCurrentSection = %v, want SectionRules", model.GetCurrentSection())
	}
}

func TestDebriefModelGetSelectedOption(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)

	if model.GetSelectedOption() != 0 {
		t.Errorf("GetSelectedOption = %d, want 0", model.GetSelectedOption())
	}

	model.selectedOption = 1
	if model.GetSelectedOption() != 1 {
		t.Errorf("GetSelectedOption = %d, want 1", model.GetSelectedOption())
	}
}

func TestDebriefModelPageUpDown(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.width = 80
	model.height = 20
	model.maxScroll = 50

	// PageDown
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	model = m.(DebriefModel)
	if model.scrollOffset != 5 {
		t.Errorf("After PageDown, scrollOffset = %d, want 5", model.scrollOffset)
	}

	// PageUp
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	model = m.(DebriefModel)
	if model.scrollOffset != 0 {
		t.Errorf("After PageUp, scrollOffset = %d, want 0", model.scrollOffset)
	}
}

func TestDebriefModelUpFromOptions(t *testing.T) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.currentSection = SectionOptions
	model.selectedOption = 0

	// Press up when at first option - should go to decisions section
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = m.(DebriefModel)

	if model.currentSection != SectionDecisions {
		t.Errorf("Section = %v, want SectionDecisions", model.currentSection)
	}
}

func TestGetActionText(t *testing.T) {
	data := createTestDebriefData()
	data.Difficulty = game.DifficultyEasy
	data.AddCheckpoint(game.NewCheckpointInfo("cp-1", 3, 80, 60))
	model := NewDebriefModel(data)

	// Test rollback text with checkpoint
	text := model.getActionText(DebriefActionRollback)
	if !strings.Contains(text, "第3章") {
		t.Errorf("Rollback text should contain chapter, got: %s", text)
	}

	// Test new game text
	text = model.getActionText(DebriefActionNewGame)
	if text != "開始新遊戲" {
		t.Errorf("New game text = %s, want '開始新遊戲'", text)
	}

	// Test menu text
	text = model.getActionText(DebriefActionMenu)
	if text != "返回主選單" {
		t.Errorf("Menu text = %s, want '返回主選單'", text)
	}
}

// Benchmark for debrief view rendering
func BenchmarkDebriefRender(b *testing.B) {
	data := createTestDebriefData()
	model := NewDebriefModel(data)
	model.width = 80
	model.height = 40

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func BenchmarkDebriefRenderExpanded(b *testing.B) {
	data := createTestDebriefData()
	// Add more data
	for i := 0; i < 5; i++ {
		data.AddRuleReveal(game.NewRuleReveal("rule-"+string(rune('A'+i)), "規則", "條件", "後果"))
		clue := game.NewClueInfo("clue-extra-"+string(rune('0'+i)), "線索內容", "rule-"+string(rune('A'+i)))
		data.AddClue(clue)
	}

	model := NewDebriefModel(data)
	model.width = 80
	model.height = 60
	// Expand all
	for i := 0; i < 6; i++ {
		model.expandedRules[i] = true
		model.expandedClues[i] = true
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}
