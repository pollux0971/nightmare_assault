package prompts

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

func TestDifficultyModifier(t *testing.T) {
	tests := []struct {
		difficulty game.DifficultyLevel
		contains   string
	}{
		{game.DifficultyEasy, "Easy"},
		{game.DifficultyHard, "Hard"},
		{game.DifficultyHell, "Hell"},
	}

	for _, tt := range tests {
		result := DifficultyModifier(tt.difficulty)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("DifficultyModifier(%v) should contain %q", tt.difficulty, tt.contains)
		}
	}
}

func TestLengthModifier(t *testing.T) {
	tests := []struct {
		length   game.GameLength
		contains string
	}{
		{game.LengthShort, "Short"},
		{game.LengthMedium, "Medium"},
		{game.LengthLong, "Long"},
	}

	for _, tt := range tests {
		result := LengthModifier(tt.length)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("LengthModifier(%v) should contain %q", tt.length, tt.contains)
		}
	}
}

func TestAdultModeModifier(t *testing.T) {
	// Adult mode on
	result := AdultModeModifier(true)
	if !strings.Contains(result, "18+") {
		t.Error("AdultModeModifier(true) should contain '18+'")
	}

	// Adult mode off
	result = AdultModeModifier(false)
	if !strings.Contains(result, "Standard") {
		t.Error("AdultModeModifier(false) should contain 'Standard'")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"
	config.Difficulty = game.DifficultyHard
	config.Length = game.LengthMedium
	config.AdultMode = false

	result := BuildSystemPrompt(config)

	// Should contain game bible
	if !strings.Contains(result, "Game Bible") {
		t.Error("BuildSystemPrompt should contain Game Bible")
	}

	// Should contain difficulty
	if !strings.Contains(result, "Hard") {
		t.Error("BuildSystemPrompt should contain difficulty")
	}

	// Should contain length
	if !strings.Contains(result, "Medium") {
		t.Error("BuildSystemPrompt should contain length")
	}

	// Should contain content rating
	if !strings.Contains(result, "Standard") {
		t.Error("BuildSystemPrompt should contain content rating")
	}
}

func TestBuildOpeningPrompt(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	result := BuildOpeningPrompt(config)

	// Should contain theme
	if !strings.Contains(result, "廢棄醫院") {
		t.Error("BuildOpeningPrompt should contain theme")
	}

	// Should mention HP/SAN
	if !strings.Contains(result, "HP") || !strings.Contains(result, "SAN") {
		t.Error("BuildOpeningPrompt should mention HP and SAN")
	}
}

func TestBuildContinuationPrompt(t *testing.T) {
	choice := "進入房間"
	context := "你站在走廊上..."

	result := BuildContinuationPrompt(choice, context)

	// Should contain choice
	if !strings.Contains(result, choice) {
		t.Error("BuildContinuationPrompt should contain player choice")
	}

	// Should contain context
	if !strings.Contains(result, context) {
		t.Error("BuildContinuationPrompt should contain context")
	}
}

func TestExtractSeeds(t *testing.T) {
	content := `這是一段故事內容。
<!-- SEED:Item:一把生鏽的鑰匙 -->
你繼續前進。
<!-- SEED:Event:遠處傳來腳步聲 -->
結尾。`

	seeds := ExtractSeeds(content)

	if len(seeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(seeds))
	}

	if len(seeds) > 0 {
		if seeds[0].Type != "Item" {
			t.Errorf("First seed type = %v, want Item", seeds[0].Type)
		}
		if seeds[0].Description != "一把生鏽的鑰匙" {
			t.Errorf("First seed description = %v, want 一把生鏽的鑰匙", seeds[0].Description)
		}
	}

	if len(seeds) > 1 {
		if seeds[1].Type != "Event" {
			t.Errorf("Second seed type = %v, want Event", seeds[1].Type)
		}
	}
}

func TestExtractSeeds_NoSeeds(t *testing.T) {
	content := "這是一段沒有種子的故事內容。"
	seeds := ExtractSeeds(content)

	if len(seeds) != 0 {
		t.Errorf("Expected 0 seeds, got %d", len(seeds))
	}
}

func TestCleanContent(t *testing.T) {
	content := `這是一段故事內容。
<!-- SEED:Item:一把生鏽的鑰匙 -->
你繼續前進。
<!-- SEED:Event:遠處傳來腳步聲 -->
結尾。`

	cleaned := CleanContent(content)

	// Should not contain seed markers
	if strings.Contains(cleaned, "<!-- SEED") {
		t.Error("CleanContent should remove seed markers")
	}

	// Should preserve actual content
	if !strings.Contains(cleaned, "這是一段故事內容") {
		t.Error("CleanContent should preserve story content")
	}
	if !strings.Contains(cleaned, "你繼續前進") {
		t.Error("CleanContent should preserve story content")
	}
}

func TestGameBibleContainsCoreMechanics(t *testing.T) {
	// Verify game bible contains essential elements
	requiredElements := []string{
		"HP",
		"SAN",
		"Narrative Rules",
		"Hidden Seeds",
		"Hidden Rules", // AC4: Game Bible includes hidden rules section
	}

	for _, elem := range requiredElements {
		if !strings.Contains(GameBible, elem) {
			t.Errorf("GameBible should contain %q", elem)
		}
	}
}

func TestBuildRulesPromptSection(t *testing.T) {
	// Test with nil rule set
	result := BuildRulesPromptSection(nil)
	if result != "" {
		t.Error("BuildRulesPromptSection(nil) should return empty string")
	}

	// Test with empty rule set
	emptyRuleSet := rules.NewRuleSet()
	result = BuildRulesPromptSection(emptyRuleSet)
	if result != "" {
		t.Error("BuildRulesPromptSection(empty) should return empty string")
	}

	// Test with rules
	ruleSet := rules.NewRuleSet()
	rule := rules.NewRule("test-1", rules.RuleTypeScenario)
	rule.Trigger = rules.Condition{Type: "location", Value: "basement", Operator: "equals"}
	rule.Consequence = rules.Outcome{Type: rules.ConsequenceDamage, HPDamage: 20}
	rule.Clues = []string{"地下室有詭異的氣息", "空氣中瀰漫著腐敗的味道"}
	rule.MaxViolations = 1
	ruleSet.Add(rule)

	result = BuildRulesPromptSection(ruleSet)

	// Should contain AI internal marker
	if !strings.Contains(result, "AI Internal") {
		t.Error("BuildRulesPromptSection should mark as AI internal")
	}

	// Should contain DO NOT reveal warning
	if !strings.Contains(result, "DO NOT reveal") {
		t.Error("BuildRulesPromptSection should warn not to reveal")
	}

	// Should contain rule type
	if !strings.Contains(result, "場景規則") {
		t.Error("BuildRulesPromptSection should contain rule type")
	}

	// Should contain trigger info
	if !strings.Contains(result, "basement") {
		t.Error("BuildRulesPromptSection should contain trigger value")
	}

	// Should contain clues
	if !strings.Contains(result, "地下室有詭異的氣息") {
		t.Error("BuildRulesPromptSection should contain clues")
	}
}

func TestBuildRulesPromptSectionConsequenceTypes(t *testing.T) {
	tests := []struct {
		conseqType rules.ConsequenceType
		expected   string
	}{
		{rules.ConsequenceWarning, "Warning"},
		{rules.ConsequenceDamage, "Damage"},
		{rules.ConsequenceInstantDeath, "Instant Death"},
	}

	for _, tt := range tests {
		ruleSet := rules.NewRuleSet()
		rule := rules.NewRule("test", rules.RuleTypeBehavior)
		rule.Consequence = rules.Outcome{Type: tt.conseqType, HPDamage: 10, SANDamage: 10}
		rule.Clues = []string{"test clue"}
		ruleSet.Add(rule)

		result := BuildRulesPromptSection(ruleSet)
		if !strings.Contains(result, tt.expected) {
			t.Errorf("BuildRulesPromptSection should contain %q for consequence type %v", tt.expected, tt.conseqType)
		}
	}
}

func TestBuildSystemPromptWithRules(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"
	config.Difficulty = game.DifficultyHard

	// Generate rules
	gen := rules.NewGenerator()
	ruleSet := gen.GenerateRules(config.Difficulty)

	result := BuildSystemPromptWithRules(config, ruleSet)

	// Should contain game bible
	if !strings.Contains(result, "Game Bible") {
		t.Error("BuildSystemPromptWithRules should contain Game Bible")
	}

	// Should contain hidden rules section (AC4)
	if !strings.Contains(result, "Active Hidden Rules") {
		t.Error("BuildSystemPromptWithRules should contain hidden rules section")
	}

	// Should contain rule clues
	clueFound := false
	for _, rule := range ruleSet.Rules {
		for _, clue := range rule.Clues {
			if strings.Contains(result, clue) {
				clueFound = true
				break
			}
		}
		if clueFound {
			break
		}
	}
	if !clueFound && ruleSet.Count() > 0 {
		t.Error("BuildSystemPromptWithRules should contain rule clues")
	}
}

func TestBuildSystemPromptWithRulesNilRuleSet(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	result := BuildSystemPromptWithRules(config, nil)

	// Should still contain game bible
	if !strings.Contains(result, "Game Bible") {
		t.Error("BuildSystemPromptWithRules should contain Game Bible even with nil rules")
	}

	// Should NOT contain hidden rules section
	if strings.Contains(result, "Active Hidden Rules") {
		t.Error("BuildSystemPromptWithRules should not contain hidden rules section when nil")
	}
}

func TestBuildOpeningPromptWithRules(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	// With rules
	ruleSet := rules.NewRuleSet()
	rule := rules.NewRule("test-1", rules.RuleTypeScenario)
	rule.Clues = []string{"test clue"}
	ruleSet.Add(rule)

	result := BuildOpeningPromptWithRules(config, ruleSet)

	// Should contain theme
	if !strings.Contains(result, "廢棄醫院") {
		t.Error("BuildOpeningPromptWithRules should contain theme")
	}

	// Should mention weaving clues (AC4)
	if !strings.Contains(result, "clues about the hidden rules") {
		t.Error("BuildOpeningPromptWithRules should mention rule clues")
	}

	// Should warn not to state rules explicitly
	if !strings.Contains(result, "Never state rules explicitly") {
		t.Error("BuildOpeningPromptWithRules should warn against explicit rule statements")
	}
}

func TestBuildOpeningPromptWithRulesNoRules(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	result := BuildOpeningPromptWithRules(config, nil)

	// Should contain theme
	if !strings.Contains(result, "廢棄醫院") {
		t.Error("BuildOpeningPromptWithRules should contain theme")
	}

	// Should NOT mention rule clues when no rules
	if strings.Contains(result, "clues about the hidden rules") {
		t.Error("BuildOpeningPromptWithRules should not mention rule clues when no rules")
	}
}

func TestBuildTeammatePromptSection(t *testing.T) {
	// Test with empty teammates
	result := BuildTeammatePromptSection(nil)
	if result != "" {
		t.Error("BuildTeammatePromptSection(nil) should return empty string")
	}

	result = BuildTeammatePromptSection([]*npc.Teammate{})
	if result != "" {
		t.Error("BuildTeammatePromptSection([]) should return empty string")
	}

	// Test with teammates
	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
		npc.NewTeammate("tm-002", "王芳", npc.ArchetypeVictim),
	}
	teammates[0].Background = "理性分析專家"
	teammates[0].Skills = []string{"邏輯推理", "記憶"}
	teammates[0].Personality.CoreTraits = []string{"rational", "calm"}

	result = BuildTeammatePromptSection(teammates)

	// Should contain AI internal marker
	if !strings.Contains(result, "AI Internal") {
		t.Error("BuildTeammatePromptSection should mark as AI internal")
	}

	// Should contain SHOW DON'T TELL reminder
	if !strings.Contains(result, "SHOW personality") {
		t.Error("BuildTeammatePromptSection should mention SHOW DON'T TELL")
	}

	// Should contain teammate names
	if !strings.Contains(result, "李明") {
		t.Error("BuildTeammatePromptSection should contain teammate name")
	}

	// Should contain archetype
	if !strings.Contains(result, string(npc.ArchetypeLogic)) {
		t.Error("BuildTeammatePromptSection should contain archetype")
	}

	// Should contain background
	if !strings.Contains(result, "理性分析專家") {
		t.Error("BuildTeammatePromptSection should contain background")
	}

	// Should contain skills
	if !strings.Contains(result, "邏輯推理") {
		t.Error("BuildTeammatePromptSection should contain skills")
	}
}

func TestBuildSystemPromptWithTeammates(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"
	config.Difficulty = game.DifficultyHard

	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
	}

	result := BuildSystemPromptWithTeammates(config, nil, teammates)

	// Should contain game bible
	if !strings.Contains(result, "Game Bible") {
		t.Error("BuildSystemPromptWithTeammates should contain Game Bible")
	}

	// Should contain teammate section
	if !strings.Contains(result, "Active Teammates") {
		t.Error("BuildSystemPromptWithTeammates should contain teammates section")
	}

	// Should contain teammate name
	if !strings.Contains(result, "李明") {
		t.Error("BuildSystemPromptWithTeammates should contain teammate name")
	}
}

func TestBuildSystemPromptWithTeammatesAndRules(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	ruleSet := rules.NewRuleSet()
	rule := rules.NewRule("test-1", rules.RuleTypeScenario)
	rule.Clues = []string{"test clue"}
	ruleSet.Add(rule)

	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
	}

	result := BuildSystemPromptWithTeammates(config, ruleSet, teammates)

	// Should contain both rules and teammates
	if !strings.Contains(result, "Active Hidden Rules") {
		t.Error("BuildSystemPromptWithTeammates should contain rules section")
	}
	if !strings.Contains(result, "Active Teammates") {
		t.Error("BuildSystemPromptWithTeammates should contain teammates section")
	}
}

func TestBuildOpeningPromptWithTeammates(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	teammates := []*npc.Teammate{
		npc.NewTeammate("tm-001", "李明", npc.ArchetypeLogic),
		npc.NewTeammate("tm-002", "王芳", npc.ArchetypeVictim),
	}

	result := BuildOpeningPromptWithTeammates(config, nil, teammates)

	// Should contain theme
	if !strings.Contains(result, "廢棄醫院") {
		t.Error("BuildOpeningPromptWithTeammates should contain theme")
	}

	// Should mention introducing teammates
	if !strings.Contains(result, "Introduce the following teammates") {
		t.Error("BuildOpeningPromptWithTeammates should mention teammate introduction")
	}

	// Should list teammate names
	if !strings.Contains(result, "李明") {
		t.Error("BuildOpeningPromptWithTeammates should list teammate name")
	}
	if !strings.Contains(result, "王芳") {
		t.Error("BuildOpeningPromptWithTeammates should list teammate name")
	}

	// Should mention SHOW DON'T TELL
	if !strings.Contains(result, "SHOW their personality") {
		t.Error("BuildOpeningPromptWithTeammates should mention SHOW DON'T TELL")
	}

	// Should warn against direct trait description
	if !strings.Contains(result, "NEVER directly describe traits") {
		t.Error("BuildOpeningPromptWithTeammates should warn against direct trait description")
	}
}

func TestBuildOpeningPromptWithTeammatesNoTeammates(t *testing.T) {
	config := game.NewGameConfig()
	config.Theme = "廢棄醫院"

	result := BuildOpeningPromptWithTeammates(config, nil, nil)

	// Should NOT mention teammates when none provided
	if strings.Contains(result, "Introduce the following teammates") {
		t.Error("BuildOpeningPromptWithTeammates should not mention teammates when none provided")
	}
}

func TestGameBibleContainsTeammatesSection(t *testing.T) {
	// Verify game bible contains teammate section
	if !strings.Contains(GameBible, "NPC Teammates") {
		t.Error("GameBible should contain NPC Teammates section")
	}

	// Should mention SHOW DON'T TELL
	if !strings.Contains(GameBible, "SHOW DON'T TELL") {
		t.Error("GameBible should mention SHOW DON'T TELL")
	}

	// Should list all archetypes
	archetypes := []string{"Victim", "Unreliable", "Logic", "Intuition", "Informer", "Possessed"}
	for _, archetype := range archetypes {
		if !strings.Contains(GameBible, archetype) {
			t.Errorf("GameBible should contain archetype %q", archetype)
		}
	}
}

func TestBoolToText(t *testing.T) {
	result := boolToText(true, "yes", "no")
	if result != "yes" {
		t.Errorf("boolToText(true, yes, no) = %v, want yes", result)
	}

	result = boolToText(false, "yes", "no")
	if result != "no" {
		t.Errorf("boolToText(false, yes, no) = %v, want no", result)
	}
}
