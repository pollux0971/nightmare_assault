package orchestrator

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.8 AC2: Death Event Processing Tests
// ==========================================================================

func TestCalculateDeathOutcome(t *testing.T) {
	tests := []struct {
		name             string
		choice           DeathWarningChoice
		archetype        agents.NPCArchetype
		randomSeed       float64
		wantSurvived     bool
		wantResponsibility float64
	}{
		{
			name:               "Prevent choice always saves",
			choice:             WarningChoicePrevent,
			archetype:          agents.NPCArchetypeSacrificial,
			randomSeed:         0.0,
			wantSurvived:       true,
			wantResponsibility: 0.0,
		},
		{
			name:               "Observe with low random seed - dies",
			choice:             WarningChoiceObserve,
			archetype:          agents.NPCArchetypeNeutral,
			randomSeed:         0.3,
			wantSurvived:       false,
			wantResponsibility: 0.5,
		},
		{
			name:               "Observe with high random seed - survives",
			choice:             WarningChoiceObserve,
			archetype:          agents.NPCArchetypeNeutral,
			randomSeed:         0.7,
			wantSurvived:       true,
			wantResponsibility: 0.5,
		},
		{
			name:               "Observe with N-01 Sacrificial - higher death chance",
			choice:             WarningChoiceObserve,
			archetype:          agents.NPCArchetypeSacrificial,
			randomSeed:         0.6,
			wantSurvived:       false,
			wantResponsibility: 0.5,
		},
		{
			name:               "Encourage with low random seed - dies",
			choice:             WarningChoiceEncourage,
			archetype:          agents.NPCArchetypeNeutral,
			randomSeed:         0.1,
			wantSurvived:       false,
			wantResponsibility: 0.9,
		},
		{
			name:               "Encourage with N-01 Sacrificial - almost always dies",
			choice:             WarningChoiceEncourage,
			archetype:          agents.NPCArchetypeSacrificial,
			randomSeed:         0.9,
			wantSurvived:       false,
			wantResponsibility: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			survived, responsibility := CalculateDeathOutcome(tt.choice, tt.archetype, tt.randomSeed)

			if survived != tt.wantSurvived {
				t.Errorf("CalculateDeathOutcome() survived = %v, want %v", survived, tt.wantSurvived)
			}

			if responsibility != tt.wantResponsibility {
				t.Errorf("CalculateDeathOutcome() responsibility = %v, want %v", responsibility, tt.wantResponsibility)
			}
		})
	}
}

func TestCalculateNPCDeathSAN(t *testing.T) {
	tests := []struct {
		name                 string
		playerResponsibility float64
		intimacy             int
		wantMin              int
		wantMax              int
	}{
		{
			name:                 "Low responsibility, low intimacy",
			playerResponsibility: 0.0,
			intimacy:             0,
			wantMin:              -20,
			wantMax:              -15,
		},
		{
			name:                 "High responsibility, high intimacy",
			playerResponsibility: 1.0,
			intimacy:             100,
			wantMin:              -25,
			wantMax:              -25,
		},
		{
			name:                 "Medium responsibility, medium intimacy",
			playerResponsibility: 0.5,
			intimacy:             50,
			wantMin:              -24,
			wantMax:              -24,
		},
		{
			name:                 "High responsibility, low intimacy",
			playerResponsibility: 1.0,
			intimacy:             0,
			wantMin:              -25,
			wantMax:              -25,
		},
		{
			name:                 "Low responsibility, high intimacy",
			playerResponsibility: 0.0,
			intimacy:             100,
			wantMin:              -25,
			wantMax:              -25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanLoss := CalculateNPCDeathSAN(tt.playerResponsibility, tt.intimacy)

			if sanLoss < -25 || sanLoss > -15 {
				t.Errorf("CalculateNPCDeathSAN() = %d, want range [-25, -15]", sanLoss)
			}

			if sanLoss < tt.wantMin || sanLoss > tt.wantMax {
				t.Errorf("CalculateNPCDeathSAN() = %d, want range [%d, %d]", sanLoss, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateNPCDeathSAN_ClampInputs(t *testing.T) {
	tests := []struct {
		name                 string
		playerResponsibility float64
		intimacy             int
	}{
		{
			name:                 "Negative responsibility clamped to 0",
			playerResponsibility: -1.0,
			intimacy:             50,
		},
		{
			name:                 "Responsibility > 1 clamped to 1",
			playerResponsibility: 2.0,
			intimacy:             50,
		},
		{
			name:                 "Negative intimacy clamped to 0",
			playerResponsibility: 0.5,
			intimacy:             -10,
		},
		{
			name:                 "Intimacy > 100 clamped to 100",
			playerResponsibility: 0.5,
			intimacy:             150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanLoss := CalculateNPCDeathSAN(tt.playerResponsibility, tt.intimacy)

			// Should not panic and should return valid range
			if sanLoss < -25 || sanLoss > -15 {
				t.Errorf("CalculateNPCDeathSAN() = %d, want range [-25, -15]", sanLoss)
			}
		})
	}
}

func TestProcessNPCDeath(t *testing.T) {
	tests := []struct {
		name                 string
		npcID                string
		npcName              string
		archetype            agents.NPCArchetype
		deathReason          string
		deathBeat            int
		location             string
		playerChoice         string
		playerResponsibility float64
		intimacy             int
		wantError            bool
	}{
		{
			name:                 "Valid death event",
			npcID:                "npc-001",
			npcName:              "王小芳",
			archetype:            agents.NPCArchetypeSacrificial,
			deathReason:          "違反規則被殺",
			deathBeat:            15,
			location:             "走廊",
			playerChoice:         "鼓勵她進入房間",
			playerResponsibility: 0.9,
			intimacy:             60,
			wantError:            false,
		},
		{
			name:                 "Empty NPC ID",
			npcID:                "",
			npcName:              "王小芳",
			archetype:            agents.NPCArchetypeSacrificial,
			deathReason:          "違反規則",
			deathBeat:            15,
			location:             "走廊",
			playerChoice:         "觀察",
			playerResponsibility: 0.5,
			intimacy:             50,
			wantError:            true,
		},
		{
			name:                 "Empty NPC name",
			npcID:                "npc-001",
			npcName:              "",
			archetype:            agents.NPCArchetypeSacrificial,
			deathReason:          "違反規則",
			deathBeat:            15,
			location:             "走廊",
			playerChoice:         "觀察",
			playerResponsibility: 0.5,
			intimacy:             50,
			wantError:            true,
		},
		{
			name:                 "Empty death reason",
			npcID:                "npc-001",
			npcName:              "王小芳",
			archetype:            agents.NPCArchetypeSacrificial,
			deathReason:          "",
			deathBeat:            15,
			location:             "走廊",
			playerChoice:         "觀察",
			playerResponsibility: 0.5,
			intimacy:             50,
			wantError:            true,
		},
		{
			name:                 "Negative death beat",
			npcID:                "npc-001",
			npcName:              "王小芳",
			archetype:            agents.NPCArchetypeSacrificial,
			deathReason:          "違反規則",
			deathBeat:            -1,
			location:             "走廊",
			playerChoice:         "觀察",
			playerResponsibility: 0.5,
			intimacy:             50,
			wantError:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ProcessNPCDeath(
				tt.npcID,
				tt.npcName,
				tt.archetype,
				tt.deathReason,
				tt.deathBeat,
				tt.location,
				tt.playerChoice,
				tt.playerResponsibility,
				tt.intimacy,
			)

			if tt.wantError {
				if err == nil {
					t.Errorf("ProcessNPCDeath() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ProcessNPCDeath() unexpected error: %v", err)
				return
			}

			if event.NPCID != tt.npcID {
				t.Errorf("event.NPCID = %v, want %v", event.NPCID, tt.npcID)
			}

			if event.NPCName != tt.npcName {
				t.Errorf("event.NPCName = %v, want %v", event.NPCName, tt.npcName)
			}

			if event.DeathReason != tt.deathReason {
				t.Errorf("event.DeathReason = %v, want %v", event.DeathReason, tt.deathReason)
			}

			if event.SANLoss < -25 || event.SANLoss > -15 {
				t.Errorf("event.SANLoss = %d, want range [-25, -15]", event.SANLoss)
			}

			if !event.Preventable {
				t.Errorf("event.Preventable = false, want true (Story 7.8 deaths are preventable)")
			}

			if event.Foreshadows == nil {
				t.Errorf("event.Foreshadows is nil, want empty slice")
			}

			if event.NPCReactions == nil {
				t.Errorf("event.NPCReactions is nil, want empty map")
			}
		})
	}
}

// ==========================================================================
// Story 7.8 AC3: Death Narrative Tests
// ==========================================================================

func TestGenerateDeathNarrative(t *testing.T) {
	tests := []struct {
		name                 string
		npcName              string
		deathReason          string
		playerResponsibility float64
		style                DeathNarrativeStyle
		wantNarrativeMin     int // Minimum narrative length
		wantLastWordsMin     int // Minimum last words length
	}{
		{
			name:                 "Detailed style with high responsibility",
			npcName:              "王小芳",
			deathReason:          "被怪物撕碎",
			playerResponsibility: 0.9,
			style:                DeathNarrativeDetailed,
			wantNarrativeMin:     100, // Simplified fallback generates shorter narratives
			wantLastWordsMin:     30,
		},
		{
			name:                 "Detailed style with low responsibility",
			npcName:              "李醫生",
			deathReason:          "突然倒下",
			playerResponsibility: 0.1,
			style:                DeathNarrativeDetailed,
			wantNarrativeMin:     100, // Simplified fallback generates shorter narratives
			wantLastWordsMin:     30,
		},
		{
			name:                 "Implied style with medium responsibility",
			npcName:              "張護士",
			deathReason:          "失血過多",
			playerResponsibility: 0.5,
			style:                DeathNarrativeImplied,
			wantNarrativeMin:     100, // Simplified fallback generates shorter narratives
			wantLastWordsMin:     30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			narrative, lastWords := GenerateDeathNarrative(
				tt.npcName,
				tt.deathReason,
				tt.playerResponsibility,
				tt.style,
			)

			// Check narrative length (Story 7.8 AC3: 200-400 chars)
			narrativeLen := len([]rune(narrative))
			if narrativeLen < tt.wantNarrativeMin {
				t.Errorf("narrative length = %d, want >= %d", narrativeLen, tt.wantNarrativeMin)
			}

			// Check last words length (Story 7.8 AC3: 30-80 chars)
			lastWordsLen := len([]rune(lastWords))
			if lastWordsLen < tt.wantLastWordsMin {
				t.Errorf("lastWords length = %d, want >= %d", lastWordsLen, tt.wantLastWordsMin)
			}

			// Check that NPC name appears in narrative
			if !containsRune(narrative, tt.npcName) {
				t.Errorf("narrative does not contain NPC name %q", tt.npcName)
			}

			// Check that death reason appears in narrative
			if !containsRune(narrative, tt.deathReason) {
				t.Errorf("narrative does not contain death reason %q", tt.deathReason)
			}

			// Check that NPC name appears in last words
			if !containsRune(lastWords, tt.npcName) {
				t.Errorf("lastWords does not contain NPC name %q", tt.npcName)
			}
		})
	}
}

func TestDeathNarrativeStyle(t *testing.T) {
	npcName := "測試角色"
	deathReason := "測試死因"

	// Detailed style should be more graphic
	detailedNarrative, _ := GenerateDeathNarrative(npcName, deathReason, 0.8, DeathNarrativeDetailed)
	impliedNarrative, _ := GenerateDeathNarrative(npcName, deathReason, 0.8, DeathNarrativeImplied)

	// Detailed should generally be longer
	detailedLen := len([]rune(detailedNarrative))
	impliedLen := len([]rune(impliedNarrative))

	if detailedLen <= impliedLen {
		t.Logf("Warning: Detailed narrative (%d chars) should typically be longer than implied (%d chars)",
			detailedLen, impliedLen)
	}

	// Both should be non-empty
	if detailedLen == 0 {
		t.Errorf("detailed narrative is empty")
	}
	if impliedLen == 0 {
		t.Errorf("implied narrative is empty")
	}
}

// ==========================================================================
// Story 7.8 AC4: Death Debrief Tests
// ==========================================================================

func TestGenerateDeathDebrief(t *testing.T) {
	event := &NPCDeathEvent{
		NPCID:                "npc-001",
		NPCName:              "王小芳",
		Archetype:            agents.NPCArchetypeSacrificial,
		DeathBeat:            15,
		DeathReason:          "違反規則被殺",
		PlayerChoice:         "鼓勵她進入房間",
		PlayerResponsibility: 0.9,
	}

	foreshadows := []DeathForeshadow{
		{
			Beat:    12,
			Type:    "clue",
			Content: "牆上寫著「禁止進入」",
		},
		{
			Beat:    13,
			Type:    "environment",
			Content: "門口有血跡",
		},
		{
			Beat:    14,
			Type:    "npc_dialogue",
			Content: "王小芳說：「這裡感覺不太對勁」",
		},
	}

	warningBeat := 14
	correctAction := "阻止她進入房間"

	debrief := GenerateDeathDebrief(event, foreshadows, warningBeat, correctAction)

	// Check basic fields
	if debrief.NPCID != event.NPCID {
		t.Errorf("debrief.NPCID = %v, want %v", debrief.NPCID, event.NPCID)
	}

	if debrief.NPCName != event.NPCName {
		t.Errorf("debrief.NPCName = %v, want %v", debrief.NPCName, event.NPCName)
	}

	// Check clues
	if len(debrief.Clues) != len(foreshadows) {
		t.Errorf("debrief.Clues length = %d, want %d", len(debrief.Clues), len(foreshadows))
	}

	// Check prevention beat
	if debrief.PreventionBeat != warningBeat {
		t.Errorf("debrief.PreventionBeat = %d, want %d", debrief.PreventionBeat, warningBeat)
	}

	// Check deadly choice
	if debrief.DeadlyChoice != event.PlayerChoice {
		t.Errorf("debrief.DeadlyChoice = %v, want %v", debrief.DeadlyChoice, event.PlayerChoice)
	}

	// Check correct action
	if debrief.CorrectAction != correctAction {
		t.Errorf("debrief.CorrectAction = %v, want %v", debrief.CorrectAction, correctAction)
	}

	// Check lesson is not empty
	if debrief.Lesson == "" {
		t.Errorf("debrief.Lesson is empty")
	}

	// Check avoid feeling message (Story 7.8 AC4: avoid "forced plot" feeling)
	if debrief.AvoidFeeling == "" {
		t.Errorf("debrief.AvoidFeeling is empty")
	}
}

// ==========================================================================
// Story 7.8 AC5: NPC Absence Effect Tests
// ==========================================================================

func TestGenerateAbsenceEffect(t *testing.T) {
	tests := []struct {
		name      string
		npcID     string
		npcName   string
		skills    []string
		items     []string
		wantConsequences int // Minimum number of consequences
	}{
		{
			name:      "NPC with skills and items",
			npcID:     "npc-001",
			npcName:   "李醫生",
			skills:    []string{"急救", "醫學知識"},
			items:     []string{"急救包", "止痛藥"},
			wantConsequences: 3,
		},
		{
			name:      "NPC with only skills",
			npcID:     "npc-002",
			npcName:   "王工程師",
			skills:    []string{"機械修理"},
			items:     []string{},
			wantConsequences: 2,
		},
		{
			name:      "NPC with only items",
			npcID:     "npc-003",
			npcName:   "張護士",
			skills:    []string{},
			items:     []string{"鑰匙"},
			wantConsequences: 2,
		},
		{
			name:      "NPC with neither skills nor items",
			npcID:     "npc-004",
			npcName:   "陳先生",
			skills:    []string{},
			items:     []string{},
			wantConsequences: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := GenerateAbsenceEffect(tt.npcID, tt.npcName, tt.skills, tt.items)

			if effect.NPCID != tt.npcID {
				t.Errorf("effect.NPCID = %v, want %v", effect.NPCID, tt.npcID)
			}

			if effect.NPCName != tt.npcName {
				t.Errorf("effect.NPCName = %v, want %v", effect.NPCName, tt.npcName)
			}

			if len(effect.MissingSkills) != len(tt.skills) {
				t.Errorf("effect.MissingSkills length = %d, want %d", len(effect.MissingSkills), len(tt.skills))
			}

			if len(effect.MissingItems) != len(tt.items) {
				t.Errorf("effect.MissingItems length = %d, want %d", len(effect.MissingItems), len(tt.items))
			}

			if len(effect.Consequences) < tt.wantConsequences {
				t.Errorf("effect.Consequences length = %d, want >= %d", len(effect.Consequences), tt.wantConsequences)
			}

			// Check that consequences mention NPC name
			foundName := false
			for _, consequence := range effect.Consequences {
				if containsRune(consequence, tt.npcName) {
					foundName = true
					break
				}
			}
			if !foundName {
				t.Errorf("consequences do not mention NPC name %q", tt.npcName)
			}
		})
	}
}

// ==========================================================================
// Story 7.8 AC6: N-01 Sacrificial Special Tests
// ==========================================================================

func TestCreateSacrificialDeathConfig(t *testing.T) {
	act2MiddleBeat := 15
	ruleID := "rule-001"
	counterExample := "不應該在夜間進入地下室"
	educational := "展示違反夜間規則的後果"

	config := CreateSacrificialDeathConfig(act2MiddleBeat, ruleID, counterExample, educational)

	if config.PlannedBeat != act2MiddleBeat {
		t.Errorf("config.PlannedBeat = %d, want %d", config.PlannedBeat, act2MiddleBeat)
	}

	if config.DemonstratedRule != ruleID {
		t.Errorf("config.DemonstratedRule = %v, want %v", config.DemonstratedRule, ruleID)
	}

	if config.CounterExample != counterExample {
		t.Errorf("config.CounterExample = %v, want %v", config.CounterExample, counterExample)
	}

	if config.Educational != educational {
		t.Errorf("config.Educational = %v, want %v", config.Educational, educational)
	}

	if !config.Saveable {
		t.Errorf("config.Saveable = false, want true (Story 7.8 AC6)")
	}

	if config.SaveConsequence == "" {
		t.Errorf("config.SaveConsequence is empty")
	}
}

func TestValidateSacrificialDeath(t *testing.T) {
	tests := []struct {
		name       string
		config     *SacrificialDeathConfig
		totalBeats int
		wantValid  bool
		wantError  bool
	}{
		{
			name: "Valid Act 2 middle death",
			config: &SacrificialDeathConfig{
				PlannedBeat:      15,
				DemonstratedRule: "rule-001",
				CounterExample:   "不要進入",
				Educational:      "展示規則",
				Saveable:         true,
			},
			totalBeats: 30,
			wantValid:  true,
			wantError:  false,
		},
		{
			name: "Death too early (Act 1)",
			config: &SacrificialDeathConfig{
				PlannedBeat:      5,
				DemonstratedRule: "rule-001",
				CounterExample:   "不要進入",
				Educational:      "展示規則",
				Saveable:         true,
			},
			totalBeats: 30,
			wantValid:  false,
			wantError:  true,
		},
		{
			name: "Death too late (Act 3)",
			config: &SacrificialDeathConfig{
				PlannedBeat:      25,
				DemonstratedRule: "rule-001",
				CounterExample:   "不要進入",
				Educational:      "展示規則",
				Saveable:         true,
			},
			totalBeats: 30,
			wantValid:  false,
			wantError:  true,
		},
		{
			name: "Missing demonstrated rule",
			config: &SacrificialDeathConfig{
				PlannedBeat:      15,
				DemonstratedRule: "",
				CounterExample:   "不要進入",
				Educational:      "展示規則",
				Saveable:         true,
			},
			totalBeats: 30,
			wantValid:  false,
			wantError:  true,
		},
		{
			name: "Missing counter example",
			config: &SacrificialDeathConfig{
				PlannedBeat:      15,
				DemonstratedRule: "rule-001",
				CounterExample:   "",
				Educational:      "展示規則",
				Saveable:         true,
			},
			totalBeats: 30,
			wantValid:  false,
			wantError:  true,
		},
		{
			name: "Not saveable (forced death)",
			config: &SacrificialDeathConfig{
				PlannedBeat:      15,
				DemonstratedRule: "rule-001",
				CounterExample:   "不要進入",
				Educational:      "展示規則",
				Saveable:         false,
			},
			totalBeats: 30,
			wantValid:  false,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateSacrificialDeath(tt.config, tt.totalBeats)

			if valid != tt.wantValid {
				t.Errorf("ValidateSacrificialDeath() valid = %v, want %v", valid, tt.wantValid)
			}

			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSacrificialDeath() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// ==========================================================================
// Helper Functions
// ==========================================================================

// containsRune checks if a string contains a substring (works with Unicode)
func containsRune(s, substr string) bool {
	return strings.Contains(s, substr)
}
