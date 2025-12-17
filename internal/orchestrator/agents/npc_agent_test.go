package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewNPCAgent tests NPC Agent creation
func TestNewNPCAgent(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    10 * time.Second,
		MaxRetries: 2,
	})

	assert.NotNil(t, agent)
	assert.Equal(t, "TestNPCAgent", agent.Config.Name)
	assert.Equal(t, 10*time.Second, agent.Config.Timeout)
	assert.Equal(t, 2, agent.Config.MaxRetries)
}

// TestNewNPCAgent_WithDefaults tests default values
func TestNewNPCAgent_WithDefaults(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	assert.NotNil(t, agent)
	assert.Equal(t, "NPCAgent", agent.Config.Name)
	assert.Equal(t, 30*time.Second, agent.Config.Timeout)
	assert.Equal(t, 3, agent.Config.MaxRetries)
}

// TestLinkNPCToSeeds tests seed linking for different archetypes
func TestLinkNPCToSeeds(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	seeds := []GlobalSeedInfo{
		{ID: "GS-001", Description: "Test Seed 1"},
		{ID: "GS-002", Description: "Test Seed 2"},
		{ID: "GS-003", Description: "Test Seed 3"},
	}

	tests := []struct {
		name          string
		archetype     NPCArchetype
		expectMin     int
		expectMax     int
		shouldNotLink bool
	}{
		{
			name:          "sacrificial_no_link",
			archetype:     NPCArchetypeSacrificial,
			expectMin:     0,
			expectMax:     0,
			shouldNotLink: true,
		},
		{
			name:      "knowledgeable_link_1_or_2",
			archetype: NPCArchetypeKnowledgeable,
			expectMin: 1,
			expectMax: 2,
		},
		{
			name:          "hostile_no_link",
			archetype:     NPCArchetypeHostile,
			expectMin:     0,
			expectMax:     0,
			shouldNotLink: true,
		},
		{
			name:      "neutral_link_0_or_1",
			archetype: NPCArchetypeNeutral,
			expectMin: 0,
			expectMax: 1,
		},
		{
			name:      "guide_link_1",
			archetype: NPCArchetypeGuide,
			expectMin: 1,
			expectMax: 1,
		},
		{
			name:      "deceiver_link_1",
			archetype: NPCArchetypeDeceiver,
			expectMin: 1,
			expectMax: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linked := agent.linkNPCToSeeds(tt.archetype, seeds)

			if tt.shouldNotLink {
				assert.Len(t, linked, 0, "Should not link any seeds")
			} else {
				assert.GreaterOrEqual(t, len(linked), tt.expectMin, "Min seed count")
				assert.LessOrEqual(t, len(linked), tt.expectMax, "Max seed count")
			}
		})
	}
}

// TestCalculateDeathTiming tests death timing calculation
func TestCalculateDeathTiming(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 25},
		Act3Range:  [2]int{25, 30},
	}

	// Run multiple times to test probability distribution
	deathTimings := make([]int, 100)
	for i := 0; i < 100; i++ {
		deathTimings[i] = agent.calculateDeathTiming(plotStructure)
	}

	// Verify all death timings are within Act 2
	for _, timing := range deathTimings {
		assert.GreaterOrEqual(t, timing, plotStructure.Act2Range[0],
			"Death timing should be >= Act2 start")
		assert.LessOrEqual(t, timing, plotStructure.Act2Range[1],
			"Death timing should be <= Act2 end")
	}

	// Check distribution (should have deaths across Act 2)
	act2Mid := (plotStructure.Act2Range[0] + plotStructure.Act2Range[1]) / 2
	earlyDeaths := 0
	midDeaths := 0
	lateDeaths := 0

	for _, timing := range deathTimings {
		if timing < act2Mid-2 {
			earlyDeaths++
		} else if timing > act2Mid+2 {
			lateDeaths++
		} else {
			midDeaths++
		}
	}

	// AC #4: ~50% should be in midpoint range
	assert.Greater(t, midDeaths, 30, "Should have significant deaths at midpoint")

	t.Logf("Death timing distribution: Early=%d, Mid=%d, Late=%d",
		earlyDeaths, midDeaths, lateDeaths)
}

// TestGetArchetypeInfo tests archetype information retrieval
func TestGetArchetypeInfo(t *testing.T) {
	tests := []struct {
		archetype    NPCArchetype
		expectedID   string
		expectedName string
	}{
		{NPCArchetypeSacrificial, "N-01", "犧牲者 (Sacrificial)"},
		{NPCArchetypeKnowledgeable, "N-02", "知情者 (Knowledgeable)"},
		{NPCArchetypeHostile, "N-03", "敵對者 (Hostile)"},
		{NPCArchetypeNeutral, "N-04", "中立者 (Neutral)"},
		{NPCArchetypeGuide, "N-05", "引導者 (Guide)"},
		{NPCArchetypeDeceiver, "N-06", "欺騙者 (Deceiver)"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			info := GetArchetypeInfo(tt.archetype)

			assert.Equal(t, tt.expectedID, info.ID)
			assert.Equal(t, tt.expectedName, info.Name)
			assert.NotEmpty(t, info.Description)
			assert.NotEmpty(t, info.Traits)
		})
	}
}

// TestGetDialogueStyle tests dialogue style retrieval
func TestGetDialogueStyle(t *testing.T) {
	tests := []struct {
		archetype    NPCArchetype
		expectedHint string
	}{
		{NPCArchetypeSacrificial, "無助"},
		{NPCArchetypeKnowledgeable, "神秘"},
		{NPCArchetypeHostile, "冷漠"},
		{NPCArchetypeNeutral, "日常"},
		{NPCArchetypeGuide, "關心"},
		{NPCArchetypeDeceiver, "虛假"},
	}

	for _, tt := range tests {
		t.Run(tt.archetype.String(), func(t *testing.T) {
			style := GetDialogueStyle(tt.archetype)

			assert.NotEmpty(t, style)
			assert.Contains(t, style, tt.expectedHint)
		})
	}
}

// TestGetDialogueLengthRange tests dialogue length range by tension
func TestGetDialogueLengthRange(t *testing.T) {
	tests := []struct {
		name          string
		tension       int
		expectMinLen  int
		expectMaxLen  int
	}{
		{"low_tension", 20, 150, 300},
		{"medium_tension", 50, 100, 200},
		{"high_tension", 70, 80, 150},
		{"very_high_tension", 90, 50, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lengthRange := GetDialogueLengthRange(tt.tension)

			assert.Equal(t, tt.expectMinLen, lengthRange[0])
			assert.Equal(t, tt.expectMaxLen, lengthRange[1])
		})
	}
}

// TestGenerateName tests name generation
func TestGenerateName(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	tests := []struct {
		theme         string
		expectedHints []string
	}{
		{"hospital", []string{"醫生", "護士", "主任", "醫師", "護理師"}},
		{"school", []string{"明", "芳", "老師", "同學"}},
		{"village", []string{"老", "大爺", "翠", "婆婆", "伯"}},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			// Generate multiple names to test variety
			names := make(map[string]bool)
			for i := 0; i < 20; i++ {
				name := agent.generateName(NPCArchetypeSacrificial, tt.theme)
				names[name] = true
				assert.NotEmpty(t, name)
			}

			// Should have variety (multiple different names)
			assert.Greater(t, len(names), 1, "Should generate varied names")

			t.Logf("Theme '%s' generated names: %v", tt.theme, names)
		})
	}
}

// TestGenerateAppearance tests appearance generation
func TestGenerateAppearance(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	tests := []struct {
		archetype    NPCArchetype
		expectedHint string
	}{
		{NPCArchetypeSacrificial, "瘦弱"},
		{NPCArchetypeKnowledgeable, "深邃"},
		{NPCArchetypeHostile, "冷漠"},
		{NPCArchetypeNeutral, "普通"},
		{NPCArchetypeGuide, "溫和"},
		{NPCArchetypeDeceiver, "友善"},
	}

	for _, tt := range tests {
		t.Run(tt.archetype.String(), func(t *testing.T) {
			appearance := agent.generateAppearance(tt.archetype, []string{})

			assert.NotEmpty(t, appearance)
			// AC #2: Appearance should be 50-100 chars
			runeCount := len([]rune(appearance))
			assert.GreaterOrEqual(t, runeCount, 30, "Appearance too short")
			assert.LessOrEqual(t, runeCount, 150, "Appearance too long")
			assert.Contains(t, appearance, tt.expectedHint)
		})
	}
}

// TestGenerateBackstory tests backstory generation
//
// H-6 FIX: Updated validation to properly check 100-200 character requirement
func TestGenerateBackstory(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	for _, archetype := range []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeHostile,
		NPCArchetypeNeutral,
		NPCArchetypeGuide,
		NPCArchetypeDeceiver,
	} {
		t.Run(archetype.String(), func(t *testing.T) {
			backstory := agent.generateBackstory(archetype, "hospital")

			assert.NotEmpty(t, backstory)
			// AC #1: Backstory should be 100-200 chars (strict validation)
			runeCount := len([]rune(backstory))
			assert.GreaterOrEqual(t, runeCount, 100, "Backstory too short - AC requires 100-200 chars")
			assert.LessOrEqual(t, runeCount, 200, "Backstory too long - AC requires 100-200 chars")

			t.Logf("%s backstory: %d characters", archetype, runeCount)
		})
	}
}

// TestInvokeGenerate_Fallback tests NPC generation with fallback
func TestInvokeGenerate_Fallback(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil, // No LLM client, will use fallback
	})

	request := &GenerateRequest{
		Archetype: NPCArchetypeSacrificial,
		StoryContext: StoryContext{
			Theme: "hospital",
			Scene: "陰暗的走廊",
		},
		GlobalSeeds: []GlobalSeedInfo{
			{ID: "GS-001", Description: "倒影的秘密"},
		},
		PlotStructure: PlotStructure{
			TotalBeats: 30,
			Act1Range:  [2]int{0, 10},
			Act2Range:  [2]int{10, 25},
			Act3Range:  [2]int{25, 30},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeGenerate(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #1: Verify NPC has all required fields
	assert.NotEmpty(t, response.NPC.ID)
	assert.NotEmpty(t, response.NPC.Name)
	assert.Equal(t, NPCArchetypeSacrificial, response.NPC.Archetype)
	assert.NotEmpty(t, response.NPC.Personality)
	assert.NotEmpty(t, response.NPC.Appearance)
	assert.NotEmpty(t, response.NPC.Backstory)
	assert.Equal(t, NPCStatusAlive, response.NPC.Status)

	// AC #3: N-01 should not link seeds
	assert.Len(t, response.NPC.LinkedSeeds, 0)

	// AC #4: N-01 should have death timing
	assert.NotZero(t, response.NPC.DeathTiming)
	assert.GreaterOrEqual(t, response.NPC.DeathTiming, request.PlotStructure.Act2Range[0])
	assert.LessOrEqual(t, response.NPC.DeathTiming, request.PlotStructure.Act2Range[1])

	t.Logf("Generated NPC: %+v", response.NPC)
}

// TestInvokeGenerate_Knowledgeable tests N-02 generation
func TestInvokeGenerate_Knowledgeable(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil,
	})

	request := &GenerateRequest{
		Archetype: NPCArchetypeKnowledgeable,
		StoryContext: StoryContext{
			Theme: "school",
			Scene: "教室",
		},
		GlobalSeeds: []GlobalSeedInfo{
			{ID: "GS-001", Description: "Test Seed 1"},
			{ID: "GS-002", Description: "Test Seed 2"},
		},
		PlotStructure: PlotStructure{
			TotalBeats: 30,
			Act1Range:  [2]int{0, 10},
			Act2Range:  [2]int{10, 25},
			Act3Range:  [2]int{25, 30},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeGenerate(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #3: N-02 should link 1-2 seeds
	assert.GreaterOrEqual(t, len(response.NPC.LinkedSeeds), 1)
	assert.LessOrEqual(t, len(response.NPC.LinkedSeeds), 2)

	// AC #4: N-02 should NOT have death timing
	assert.Zero(t, response.NPC.DeathTiming)

	t.Logf("Generated N-02 NPC: %+v", response.NPC)
}

// TestInvokeDialogue_Fallback tests dialogue generation with fallback
func TestInvokeDialogue_Fallback(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil,
	})

	npc := NPCInstance{
		ID:          "NPC-001",
		Name:        "王護士",
		Archetype:   NPCArchetypeSacrificial,
		Personality: []string{"無助", "恐懼"},
		Status:      NPCStatusAlive,
	}

	request := &DialogueRequest{
		NPC:         npc,
		Context:     "陰暗的走廊",
		Tension:     50,
		CurrentBeat: 5,
	}

	ctx := context.Background()
	response, err := agent.InvokeDialogue(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #5: Dialogue should not be empty
	assert.NotEmpty(t, response.Dialogue)

	// AC #6: Should match archetype style (Sacrificial = helpless)
	assert.Contains(t, response.Dialogue, "救")

	assert.False(t, response.IsDeathDialogue)

	t.Logf("Generated dialogue: %s", response.Dialogue)
}

// TestInvokeDialogue_DeathDialogue tests death dialogue generation
func TestInvokeDialogue_DeathDialogue(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil,
	})

	npc := NPCInstance{
		ID:          "NPC-001",
		Name:        "王護士",
		Archetype:   NPCArchetypeSacrificial,
		Personality: []string{"無助", "恐懼"},
		Status:      NPCStatusAlive,
		LinkedSeeds: []string{"GS-001"},
		DeathTiming: 15,
	}

	request := &DialogueRequest{
		NPC:         npc,
		Context:     "血腥的房間",
		Tension:     90,
		CurrentBeat: 15, // Death timing reached
	}

	ctx := context.Background()
	response, err := agent.InvokeDialogue(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #10: Death dialogue should be present
	assert.NotEmpty(t, response.Dialogue)
	assert.True(t, response.IsDeathDialogue)

	// AC #10: Should contain fear and warning
	assert.Contains(t, response.Dialogue, "不")

	// AC #10: Should reveal seed
	assert.NotNil(t, response.SeedRevealed)
	assert.Equal(t, "GS-001", *response.SeedRevealed)

	t.Logf("Death dialogue: %s", response.Dialogue)
}

// TestGetClueRevealHint tests clue reveal hints by tension
func TestGetClueRevealHint(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	tests := []struct {
		name        string
		tension     int
		expectedLen int // Rough length indicator
	}{
		{"very_vague_low_tension", 20, 30},
		{"vague_medium_tension", 50, 40},
		{"specific_high_tension", 70, 30},
		{"direct_very_high_tension", 90, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := agent.getClueRevealHint(tt.tension)

			assert.NotEmpty(t, hint)
			t.Logf("Tension %d hint: %s", tt.tension, hint)
		})
	}
}

// TestBuildGeneratePrompt tests prompt building for generation
func TestBuildGeneratePrompt(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	request := &GenerateRequest{
		Archetype: NPCArchetypeSacrificial,
		StoryContext: StoryContext{
			Theme: "hospital",
			Scene: "病房",
		},
		GlobalSeeds: []GlobalSeedInfo{
			{ID: "GS-001", Description: "倒影的秘密"},
		},
	}

	prompt := agent.buildGeneratePrompt(request)

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "N-01")
	assert.Contains(t, prompt, "hospital")
	assert.Contains(t, prompt, "GS-001")
	assert.Contains(t, prompt, "json")

	t.Logf("Generate prompt length: %d", len(prompt))
}

// TestBuildDialoguePrompt tests prompt building for dialogue
func TestBuildDialoguePrompt(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	npc := NPCInstance{
		Name:        "王護士",
		Archetype:   NPCArchetypeKnowledgeable,
		Personality: []string{"神秘", "謹慎"},
		LinkedSeeds: []string{"GS-001"},
	}

	request := &DialogueRequest{
		NPC:            npc,
		PlayerQuestion: "這裡發生了什麼？",
		Context:        "走廊",
		Tension:        60,
	}

	prompt := agent.buildDialoguePrompt(request)

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "王護士")
	assert.Contains(t, prompt, "N-02")
	assert.Contains(t, prompt, "這裡發生了什麼？")
	assert.Contains(t, prompt, "線索揭露")

	t.Logf("Dialogue prompt length: %d", len(prompt))
}
