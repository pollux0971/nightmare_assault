package orchestrator

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.8: NPC Death Integration Tests
// ==========================================================================

// TestNPCDeathIntegration_CompleteFlow tests the complete NPC death flow
//
// Story 7.8 AC1-AC6: Complete death flow integration
//   - Three-stage architecture (foreshadow, warning, consequence)
//   - Death event processing
//   - Death narrative generation
//   - Death debrief
//   - Absence effects
func TestNPCDeathIntegration_CompleteFlow(t *testing.T) {
	// Story 7.8 AC1: Three-Stage Architecture
	// Stage 1: Foreshadowing (Beat N-3 to N-1)
	foreshadows := []DeathForeshadow{
		{
			Beat:    12,
			Type:    "environment",
			Content: "地上有大量血跡，看起來像是有什麼被拖進去了",
		},
		{
			Beat:    13,
			Type:    "clue",
			Content: "牆上用血寫著「不要進去」四個大字",
		},
		{
			Beat:    14,
			Type:    "npc_dialogue",
			Content: "王小芳不安地說：「我總覺得這裡不太對勁……」",
		},
	}

	// Stage 2: Warning (Beat N = 15)
	warningBeat := 15

	// Player chooses to encourage (high risk choice)
	playerChoice := WarningChoiceEncourage

	// Stage 3: Calculate outcome
	npcArchetype := agents.NPCArchetypeSacrificial
	randomSeed := 0.5 // Will result in death for encourage + sacrificial (95% death chance)

	survived, playerResponsibility := CalculateDeathOutcome(playerChoice, npcArchetype, randomSeed)

	if survived {
		t.Errorf("Expected NPC to die with encourage choice and low random seed, but survived")
	}

	if playerResponsibility != 0.9 {
		t.Errorf("Expected high player responsibility (0.9) for encourage choice, got %.2f", playerResponsibility)
	}

	// Story 7.8 AC2: Death Event Processing
	deathEvent, err := ProcessNPCDeath(
		"npc-001",
		"王小芳",
		npcArchetype,
		"違反規則進入禁區被怪物撕碎",
		warningBeat,
		"地下室入口",
		string(playerChoice),
		playerResponsibility,
		60, // intimacy
	)

	if err != nil {
		t.Fatalf("ProcessNPCDeath failed: %v", err)
	}

	// Verify death event
	if deathEvent.NPCID != "npc-001" {
		t.Errorf("deathEvent.NPCID = %v, want npc-001", deathEvent.NPCID)
	}

	if deathEvent.SANLoss > -15 || deathEvent.SANLoss < -25 {
		t.Errorf("deathEvent.SANLoss = %d, want range [-25, -15]", deathEvent.SANLoss)
	}

	if !deathEvent.Preventable {
		t.Errorf("deathEvent.Preventable = false, want true (Story 7.8: deaths are preventable)")
	}

	// Story 7.8 AC3: Death Narrative Generation (using simplified function)
	narrative, lastWords := GenerateDeathNarrative(
		deathEvent.NPCName,
		deathEvent.DeathReason,
		deathEvent.PlayerResponsibility,
		DeathNarrativeDetailed, // 18+ mode for detailed narrative
	)

	// Verify narrative response
	if narrative == "" {
		t.Errorf("DeathNarrative is empty")
	}

	if lastWords == "" {
		t.Errorf("LastWords is empty")
	}

	narrativeLen := len([]rune(narrative))
	lastWordsLen := len([]rune(lastWords))

	// Story 7.8 AC4: Death Debrief
	debrief := GenerateDeathDebrief(
		deathEvent,
		foreshadows,
		warningBeat,
		"選擇阻止她進入地下室",
	)

	if debrief.NPCID != deathEvent.NPCID {
		t.Errorf("debrief.NPCID = %v, want %v", debrief.NPCID, deathEvent.NPCID)
	}

	if len(debrief.Clues) != len(foreshadows) {
		t.Errorf("debrief.Clues length = %d, want %d", len(debrief.Clues), len(foreshadows))
	}

	if debrief.Lesson == "" {
		t.Errorf("debrief.Lesson is empty")
	}

	if debrief.AvoidFeeling == "" {
		t.Errorf("debrief.AvoidFeeling is empty")
	}

	// Story 7.8 AC5: Absence Effects
	absenceEffect := GenerateAbsenceEffect(
		deathEvent.NPCID,
		deathEvent.NPCName,
		[]string{"急救", "醫療知識"},
		[]string{"急救包", "止痛藥"},
	)

	if absenceEffect.NPCID != deathEvent.NPCID {
		t.Errorf("absenceEffect.NPCID = %v, want %v", absenceEffect.NPCID, deathEvent.NPCID)
	}

	if len(absenceEffect.MissingSkills) != 2 {
		t.Errorf("absenceEffect.MissingSkills length = %d, want 2", len(absenceEffect.MissingSkills))
	}

	if len(absenceEffect.MissingItems) != 2 {
		t.Errorf("absenceEffect.MissingItems length = %d, want 2", len(absenceEffect.MissingItems))
	}

	if len(absenceEffect.Consequences) < 1 {
		t.Errorf("absenceEffect.Consequences is empty")
	}

	t.Logf("=== Complete Death Flow Test Summary ===")
	t.Logf("NPC: %s (ID: %s)", deathEvent.NPCName, deathEvent.NPCID)
	t.Logf("Death Beat: %d", deathEvent.DeathBeat)
	t.Logf("Player Responsibility: %.1f%%", playerResponsibility*100)
	t.Logf("SAN Loss: %d", deathEvent.SANLoss)
	t.Logf("Foreshadows: %d clues given", len(foreshadows))
	t.Logf("Narrative Length: %d runes", narrativeLen)
	t.Logf("Last Words Length: %d runes", lastWordsLen)
	t.Logf("Absence Effects: %d skills, %d items lost", len(absenceEffect.MissingSkills), len(absenceEffect.MissingItems))
}

// TestNPCDeathIntegration_SacrificialSpecial tests N-01 Sacrificial special handling
//
// Story 7.8 AC6: N-01 Sacrificial special death configuration
func TestNPCDeathIntegration_SacrificialSpecial(t *testing.T) {
	totalBeats := 30
	act2MiddleBeat := 15

	// Create sacrificial death config
	config := CreateSacrificialDeathConfig(
		act2MiddleBeat,
		"rule-night-001",
		"夜間進入禁區",
		"展示違反夜間規則的致命後果",
	)

	// Validate config
	valid, err := ValidateSacrificialDeath(config, totalBeats)
	if !valid {
		t.Errorf("ValidateSacrificialDeath failed: %v", err)
	}

	if config.PlannedBeat != act2MiddleBeat {
		t.Errorf("config.PlannedBeat = %d, want %d", config.PlannedBeat, act2MiddleBeat)
	}

	if !config.Saveable {
		t.Errorf("config.Saveable = false, want true (Story 7.8: sacrificial deaths must be preventable)")
	}

	if config.DemonstratedRule != "rule-night-001" {
		t.Errorf("config.DemonstratedRule = %v, want rule-night-001", config.DemonstratedRule)
	}

	if config.SaveConsequence == "" {
		t.Errorf("config.SaveConsequence is empty")
	}

	t.Logf("=== Sacrificial Death Config Test Summary ===")
	t.Logf("Planned Beat: %d (Act 2 middle)", config.PlannedBeat)
	t.Logf("Demonstrated Rule: %s", config.DemonstratedRule)
	t.Logf("Counter Example: %s", config.CounterExample)
	t.Logf("Saveable: %v", config.Saveable)
}

// TestThreeStageArchitecture tests the three-stage death architecture
//
// Story 7.8 AC1: Ensure three stages work together correctly
func TestThreeStageArchitecture(t *testing.T) {
	tests := []struct {
		name                 string
		playerChoice         DeathWarningChoice
		archetype            agents.NPCArchetype
		randomSeed           float64
		wantSurvived         bool
		wantResponsibility   float64
	}{
		{
			name:               "Prevent always saves",
			playerChoice:       WarningChoicePrevent,
			archetype:          agents.NPCArchetypeSacrificial,
			randomSeed:         0.0,
			wantSurvived:       true,
			wantResponsibility: 0.0,
		},
		{
			name:               "Observe - medium risk",
			playerChoice:       WarningChoiceObserve,
			archetype:          agents.NPCArchetypeNeutral,
			randomSeed:         0.7, // Above 50% death chance
			wantSurvived:       true,
			wantResponsibility: 0.5,
		},
		{
			name:               "Encourage - high risk",
			playerChoice:       WarningChoiceEncourage,
			archetype:          agents.NPCArchetypeSacrificial,
			randomSeed:         0.5,
			wantSurvived:       false, // 95% death chance for sacrificial
			wantResponsibility: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			survived, responsibility := CalculateDeathOutcome(tt.playerChoice, tt.archetype, tt.randomSeed)

			if survived != tt.wantSurvived {
				t.Errorf("survived = %v, want %v", survived, tt.wantSurvived)
			}

			if responsibility != tt.wantResponsibility {
				t.Errorf("responsibility = %.2f, want %.2f", responsibility, tt.wantResponsibility)
			}

			t.Logf("Choice: %s, Survived: %v, Responsibility: %.1f%%",
				tt.playerChoice.String(), survived, responsibility*100)
		})
	}
}
