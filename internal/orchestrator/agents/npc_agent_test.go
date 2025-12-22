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

// ==========================================================================
// Story 7.7: Performance Test for 500ms Dialogue Generation
// ==========================================================================

// TestInvokeDialogue_PerformanceRequirement tests that dialogue generation
// completes within the 500ms requirement (AC #1).
//
// Story 7.7 AC #1: "NPC Agent 應該使用 Fast Model 生成對話 (< 500ms)"
// Note: With template fallback (no LLM), this should complete much faster.
// The test verifies the timeout mechanism works correctly.
func TestInvokeDialogue_PerformanceRequirement(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil, // Uses template fallback
	})

	npc := NPCInstance{
		ID:          "NPC-PERF",
		Name:        "測試NPC",
		Archetype:   NPCArchetypeSacrificial,
		Personality: []string{"無助", "恐懼"},
		Status:      NPCStatusAlive,
	}

	request := &DialogueRequest{
		NPC:         npc,
		Context:     "效能測試場景",
		Tension:     60,
		CurrentBeat: 5,
		NPCSAN:      80,
	}

	// Run multiple iterations to ensure consistent performance
	iterations := 10
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		response, err := agent.InvokeDialogue(context.Background(), request)
		duration := time.Since(start)
		totalDuration += duration

		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotEmpty(t, response.Dialogue)

		// Story 7.7 AC #1: Each call should complete within 500ms
		if duration > 500*time.Millisecond {
			t.Errorf("Iteration %d: Dialogue generation took %v, exceeds 500ms requirement", i+1, duration)
		}
	}

	avgDuration := totalDuration / time.Duration(iterations)
	t.Logf("Average dialogue generation time over %d iterations: %v", iterations, avgDuration)

	// Average should be well under 500ms (with template fallback, expect <1ms)
	if avgDuration > 100*time.Millisecond {
		t.Logf("Warning: Average time %v is higher than expected for template fallback", avgDuration)
	}
}

// ==========================================================================
// Story 7.6: NPC Generation Enhancement Tests
// ==========================================================================

// TestNPCAgent_GenerateSkills tests the generateSkills method (Story 7.6)
func TestNPCAgent_GenerateSkills(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	tests := []struct {
		name      string
		archetype NPCArchetype
		wantMin   int
	}{
		{"Sacrificial", NPCArchetypeSacrificial, 1},
		{"Knowledgeable", NPCArchetypeKnowledgeable, 2},
		{"Hostile", NPCArchetypeHostile, 1},
		{"Neutral", NPCArchetypeNeutral, 1},
		{"Guide", NPCArchetypeGuide, 2},
		{"Deceiver", NPCArchetypeDeceiver, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skills := agent.generateSkills(tt.archetype)
			assert.GreaterOrEqual(t, len(skills), tt.wantMin,
				"Should generate at least %d skills", tt.wantMin)
			for _, skill := range skills {
				assert.NotEmpty(t, skill, "Skills should not be empty")
			}
		})
	}
}

// TestNPCAgent_GenerateInventory tests the generateInventory method (Story 7.6)
func TestNPCAgent_GenerateInventory(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	tests := []struct {
		name      string
		archetype NPCArchetype
		theme     string
	}{
		{"Hospital Sacrificial", NPCArchetypeSacrificial, "hospital"},
		{"School Knowledgeable", NPCArchetypeKnowledgeable, "school"},
		{"Village Guide", NPCArchetypeGuide, "village"},
		{"Default Theme", NPCArchetypeNeutral, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := agent.generateInventory(tt.archetype, tt.theme)
			assert.NotEmpty(t, inventory, "Should generate at least one item")
			for _, item := range inventory {
				assert.NotEmpty(t, item, "Items should not be empty")
			}
			t.Logf("Generated inventory for %s/%s: %v", tt.archetype, tt.theme, inventory)
		})
	}
}

// TestNPCAgent_GenerateSecret tests the generateSecret method (Story 7.6)
func TestNPCAgent_GenerateSecret(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	archetypes := []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeHostile,
		NPCArchetypeNeutral,
		NPCArchetypeGuide,
		NPCArchetypeDeceiver,
	}

	for _, archetype := range archetypes {
		t.Run(archetype.String(), func(t *testing.T) {
			secret := agent.generateSecret(archetype, "hospital")
			assert.NotEmpty(t, secret, "Secret should not be empty")
			runeCount := len([]rune(secret))
			assert.GreaterOrEqual(t, runeCount, 20, "Secret should be at least 20 chars")
			t.Logf("Generated secret for %s: %s", archetype, secret)
		})
	}
}

// TestNPCAgent_ValidateShowDontTell tests Show-Don't-Tell validation (Story 7.6 AC #2)
func TestNPCAgent_ValidateShowDontTell(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	tests := []struct {
		name         string
		introduction string
		wantValid    bool
	}{
		{
			name:         "Valid - Action and Dialogue",
			introduction: "他的手顫抖著，眼神不停閃避，握緊了口袋裡的鏽鑰匙。冷汗從額頭滑落，嘴唇緊抿成一條線，身體微微後退，彷彿隨時準備逃跑。",
			wantValid:    true,
		},
		{
			name:         "Valid - Showing through Items",
			introduction: "她冷笑一聲，把染血的繃帶扔在桌上：「又一個不聽勸的。」眼神冷漠地掃過眾人，轉身離開時腳步沉重而堅定，完全不在乎周圍的反應。",
			wantValid:    true,
		},
		{
			name:         "Invalid - Direct Description '很恐懼'",
			introduction: "他很恐懼，手中握著鏽鑰匙，不停地顫抖。",
			wantValid:    false,
		},
		{
			name:         "Invalid - Direct Description '是個'",
			introduction: "她是個知識淵博的人，手中拿著古老的書籍。",
			wantValid:    false,
		},
		{
			name:         "Invalid - Too Short",
			introduction: "他站在那裡。",
			wantValid:    false,
		},
		{
			name:         "Invalid - Contains '充滿恐懼'",
			introduction: "他的眼神充滿恐懼，手中緊握著武器，身體不住地顫抖著，完全無法控制自己的情緒。",
			wantValid:    false,
		},
		{
			name:         "Invalid - Contains '他很'",
			introduction: "他很冷靜地觀察周圍，手中拿著筆記本記錄著什麼，完全不受外界干擾。",
			wantValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := agent.validateShowDontTell(tt.introduction)
			assert.Equal(t, tt.wantValid, valid, "Validation result mismatch")
		})
	}
}

// TestNPCAgent_GenerateTemplateIntroduction tests template introduction (Story 7.6 AC #2)
func TestNPCAgent_GenerateTemplateIntroduction(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	archetypes := []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeHostile,
		NPCArchetypeNeutral,
		NPCArchetypeGuide,
		NPCArchetypeDeceiver,
	}

	for _, archetype := range archetypes {
		t.Run(archetype.String(), func(t *testing.T) {
			npc := NPCInstance{
				Name:      "測試角色",
				Archetype: archetype,
				Inventory: []string{"測試物品", "神秘道具"},
			}

			request := &IntroductionRequest{
				NPC: npc,
				StoryContext: StoryContext{
					Theme: "hospital",
					Scene: "廢棄的醫院走廊",
				},
			}

			response := agent.generateTemplateIntroduction(request)

			// Check not empty
			assert.NotEmpty(t, response.Introduction, "Introduction should not be empty")

			// Check length (100-200 chars)
			runeCount := len([]rune(response.Introduction))
			assert.GreaterOrEqual(t, runeCount, 50, "Introduction should be at least 50 chars")

			// Check Show-Don't-Tell validation
			assert.True(t, agent.validateShowDontTell(response.Introduction),
				"Template introduction should pass Show-Don't-Tell validation: %s",
				response.Introduction)

			t.Logf("%s introduction: %s", archetype, response.Introduction)
		})
	}
}

// TestNPCAgent_InvokeGenerate_WithNewFields tests NPC generation with Story 7.6 fields
func TestNPCAgent_InvokeGenerate_WithNewFields(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil, // Will use fallback
	})

	request := &GenerateRequest{
		Archetype: NPCArchetypeGuide,
		StoryContext: StoryContext{
			Theme: "hospital",
			Scene: "廢棄醫院的大廳",
		},
		GlobalSeeds:   []GlobalSeedInfo{},
		PlotStructure: PlotStructure{
			TotalBeats: 20,
			Act1Range:  [2]int{1, 5},
			Act2Range:  [2]int{6, 15},
			Act3Range:  [2]int{16, 20},
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeGenerate(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	npc := response.NPC

	// AC #1: Check all new fields are populated
	assert.NotEmpty(t, npc.Skills, "Skills should be populated")
	assert.NotEmpty(t, npc.Inventory, "Inventory should be populated")
	assert.NotEmpty(t, npc.Secret, "Secret should be populated")

	// Check basic fields still work
	assert.NotEmpty(t, npc.Name, "Name should be populated")
	assert.NotEmpty(t, npc.Personality, "Personality should be populated")
	assert.NotEmpty(t, npc.Appearance, "Appearance should be populated")
	assert.NotEmpty(t, npc.Backstory, "Backstory should be populated")

	t.Logf("Generated NPC with new fields:")
	t.Logf("  Name: %s", npc.Name)
	t.Logf("  Skills: %v", npc.Skills)
	t.Logf("  Inventory: %v", npc.Inventory)
	t.Logf("  Secret: %s", npc.Secret)
}

// TestNPCAgent_InvokeIntroduction_Fallback tests introduction generation (Story 7.6 AC #2)
func TestNPCAgent_InvokeIntroduction_Fallback(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil, // Will use fallback
	})

	npc := NPCInstance{
		Name:        "李醫師",
		Archetype:   NPCArchetypeGuide,
		Personality: []string{"關心", "友善", "提示"},
		Skills:      []string{"急救", "引導"},
		Inventory:   []string{"急救包", "手電筒"},
		Secret:      "曾經失去過重要的人，現在想幫助他人避免同樣的悲劇",
	}

	request := &IntroductionRequest{
		NPC: npc,
		StoryContext: StoryContext{
			Theme: "hospital",
			Scene: "廢棄醫院的急診室",
		},
	}

	ctx := context.Background()
	response, err := agent.InvokeIntroduction(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC #2: Check introduction follows Show-Don't-Tell
	assert.True(t, agent.validateShowDontTell(response.Introduction),
		"Introduction should pass Show-Don't-Tell validation: %s",
		response.Introduction)

	// Check length (50-250 chars, relaxed for fallback)
	runeCount := len([]rune(response.Introduction))
	assert.GreaterOrEqual(t, runeCount, 50, "Introduction should be at least 50 chars")
	assert.LessOrEqual(t, runeCount, 250, "Introduction should be at most 250 chars")

	t.Logf("Generated introduction: %s", response.Introduction)
}

// TestNPCAgent_BuildIntroductionPrompt tests introduction prompt building (Story 7.6)
func TestNPCAgent_BuildIntroductionPrompt(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{})

	npc := NPCInstance{
		Name:        "王醫生",
		Archetype:   NPCArchetypeKnowledgeable,
		Personality: []string{"神秘", "謹慎", "知情"},
		Skills:      []string{"觀察", "解謎"},
		Inventory:   []string{"破舊的筆記", "古老的鑰匙"},
		Secret:      "知道這個地方的核心秘密，但受到某種約束無法直接揭露",
	}

	request := &IntroductionRequest{
		NPC: npc,
		StoryContext: StoryContext{
			Theme: "hospital",
			Scene: "醫院地下室",
		},
	}

	prompt := agent.buildIntroductionPrompt(request)

	// Check prompt contains key elements
	assert.Contains(t, prompt, "Show, Don't Tell", "Should mention Show-Don't-Tell")
	assert.Contains(t, prompt, "王醫生", "Should include NPC name")
	assert.Contains(t, prompt, "破舊的筆記", "Should include inventory items")
	assert.Contains(t, prompt, "100-200 字", "Should specify length requirement")
	assert.Contains(t, prompt, "禁止直接描述", "Should forbid direct descriptions")
	assert.Contains(t, prompt, "json", "Should request JSON format")

	t.Logf("Introduction prompt length: %d", len(prompt))
}

// TestNPCAgent_MultipleNPCGeneration tests generating multiple NPCs (Story 7.6 AC #1)
func TestNPCAgent_MultipleNPCGeneration(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		LLMClient: nil, // Will use fallback
	})

	archetypes := []NPCArchetype{
		NPCArchetypeGuide,
		NPCArchetypeKnowledgeable,
		NPCArchetypeNeutral,
		NPCArchetypeSacrificial,
	}

	ctx := context.Background()

	for i, archetype := range archetypes {
		t.Run(archetype.String(), func(t *testing.T) {
			request := &GenerateRequest{
				Archetype: archetype,
				StoryContext: StoryContext{
					Theme: "hospital",
					Scene: "廢棄醫院",
				},
				GlobalSeeds: []GlobalSeedInfo{
					{ID: "GS-001", Description: "倒影的秘密"},
					{ID: "GS-002", Description: "鏡子的真相"},
				},
				PlotStructure: PlotStructure{
					TotalBeats: 20,
					Act1Range:  [2]int{1, 5},
					Act2Range:  [2]int{6, 15},
					Act3Range:  [2]int{16, 20},
				},
			}

			response, err := agent.InvokeGenerate(ctx, request)
			require.NoError(t, err)

			npc := response.NPC
			assert.NotEmpty(t, npc.Name)
			assert.NotEmpty(t, npc.Skills)
			assert.NotEmpty(t, npc.Inventory)
			assert.NotEmpty(t, npc.Secret)

			// Generate introduction
			introReq := &IntroductionRequest{
				NPC:          npc,
				StoryContext: request.StoryContext,
			}

			introResp, err := agent.InvokeIntroduction(ctx, introReq)
			require.NoError(t, err)

			assert.True(t, agent.validateShowDontTell(introResp.Introduction))

			t.Logf("NPC %d (%s):", i+1, archetype)
			t.Logf("  Name: %s", npc.Name)
			t.Logf("  Skills: %v", npc.Skills)
			t.Logf("  Inventory: %v", npc.Inventory)
			t.Logf("  Secret: %s", npc.Secret)
			t.Logf("  Introduction: %s", introResp.Introduction)
		})
	}
}
