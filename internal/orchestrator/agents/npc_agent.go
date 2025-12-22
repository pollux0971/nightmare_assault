package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ==========================================================================
// Story 6-8: NPC Agent Implementation
// ==========================================================================

// NPCAgent generates NPC instances and dialogue
//
// Responsibilities:
//  1. Generate NPC instances based on archetypes (<5s using Smart Model)
//  2. Generate NPC dialogue (<500ms using Fast Model)
//  3. Link NPCs to Global Seeds for clue revelation
//  4. Calculate death timing for Sacrificial NPCs
//
// Design Philosophy:
//  - Use Smart Model for NPC generation (higher quality, longer timeout)
//  - Use Fast Model for dialogue (speed critical, <500ms)
//  - Template fallback for dialogue generation failures
type NPCAgent struct {
	*BaseAgentImpl
	rng *rand.Rand
}

// NewNPCAgent creates a new NPCAgent with BaseAgentImpl pattern
func NewNPCAgent(config AgentConfig) *NPCAgent {
	// Set defaults if not provided
	if config.Name == "" {
		config.Name = "NPCAgent"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	na := &NPCAgent{
		BaseAgentImpl: NewBaseAgentImpl(config),
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return na
}

// InvokeGenerate generates an NPC instance
//
// AC #1: Generate complete NPC with name, archetype, personality, appearance, backstory
// AC #3: Link to 0-2 Global Seeds based on archetype
// AC #4: Calculate death timing for N-01 Sacrificial
// AC #12: Generation time <5s using Smart Model
func (na *NPCAgent) InvokeGenerate(ctx context.Context, request *GenerateRequest) (*GenerateResponse, error) {
	// 1. Build generation prompt
	prompt := na.buildGeneratePrompt(request)

	// 2. Call LLM via BaseAgentImpl's InvokeWithRetry (Smart Model, 5s timeout)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := na.BaseAgentImpl.InvokeWithRetry(ctxWithTimeout, func(ctx context.Context) (any, error) {
		if na.Config.LLMClient == nil {
			// Return a marker for fallback
			return nil, fmt.Errorf("no LLM client available")
		}

		return na.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.7, // Creative but consistent
			"max_tokens":  1000,
		})
	})

	// 3. Generate or parse NPC
	var npc NPCInstance

	if err != nil {
		// Fallback to template-based generation (without calling generateFallbackNPC)
		archetypeInfo := GetArchetypeInfo(request.Archetype)
		name := na.generateName(request.Archetype, request.StoryContext.Theme)
		appearance := na.generateAppearance(request.Archetype, archetypeInfo.Traits)
		backstory := na.generateBackstory(request.Archetype, request.StoryContext.Theme)
		skills := na.generateSkills(request.Archetype)
		inventory := na.generateInventory(request.Archetype, request.StoryContext.Theme)
		secret := na.generateSecret(request.Archetype, request.StoryContext.Theme)

		npc = NPCInstance{
			ID:          na.generateNPCID(),
			Name:        name,
			Archetype:   request.Archetype,
			Personality: archetypeInfo.Traits[:3], // Use first 3 traits
			Appearance:  appearance,
			Backstory:   backstory,
			Skills:      skills,
			Inventory:   inventory,
			Secret:      secret,
			Status:      NPCStatusAlive,
		}
	} else {
		// Parse LLM response
		responseStr, ok := response.(string)
		if !ok {
			// Unexpected type, use fallback
			archetypeInfo := GetArchetypeInfo(request.Archetype)
			name := na.generateName(request.Archetype, request.StoryContext.Theme)
			appearance := na.generateAppearance(request.Archetype, archetypeInfo.Traits)
			backstory := na.generateBackstory(request.Archetype, request.StoryContext.Theme)
			skills := na.generateSkills(request.Archetype)
			inventory := na.generateInventory(request.Archetype, request.StoryContext.Theme)
			secret := na.generateSecret(request.Archetype, request.StoryContext.Theme)

			npc = NPCInstance{
				ID:          na.generateNPCID(),
				Name:        name,
				Archetype:   request.Archetype,
				Personality: archetypeInfo.Traits[:3],
				Appearance:  appearance,
				Backstory:   backstory,
				Skills:      skills,
				Inventory:   inventory,
				Secret:      secret,
				Status:      NPCStatusAlive,
			}
		} else {
			llmResp, err := na.parseGenerateResponse(responseStr)
			if err != nil {
				// Parse failed, use fallback
				archetypeInfo := GetArchetypeInfo(request.Archetype)
				name := na.generateName(request.Archetype, request.StoryContext.Theme)
				appearance := na.generateAppearance(request.Archetype, archetypeInfo.Traits)
				backstory := na.generateBackstory(request.Archetype, request.StoryContext.Theme)
				skills := na.generateSkills(request.Archetype)
				inventory := na.generateInventory(request.Archetype, request.StoryContext.Theme)
				secret := na.generateSecret(request.Archetype, request.StoryContext.Theme)

				npc = NPCInstance{
					ID:          na.generateNPCID(),
					Name:        name,
					Archetype:   request.Archetype,
					Personality: archetypeInfo.Traits[:3],
					Appearance:  appearance,
					Backstory:   backstory,
					Skills:      skills,
					Inventory:   inventory,
					Secret:      secret,
					Status:      NPCStatusAlive,
				}
			} else {
				// Use LLM response with fallback for missing fields
				skills := llmResp.Skills
				if len(skills) == 0 {
					skills = na.generateSkills(request.Archetype)
				}
				inventory := llmResp.Inventory
				if len(inventory) == 0 {
					inventory = na.generateInventory(request.Archetype, request.StoryContext.Theme)
				}
				secret := llmResp.Secret
				if secret == "" {
					secret = na.generateSecret(request.Archetype, request.StoryContext.Theme)
				}

				npc = NPCInstance{
					ID:          na.generateNPCID(),
					Name:        llmResp.Name,
					Archetype:   request.Archetype,
					Personality: llmResp.Personality,
					Appearance:  llmResp.Appearance,
					Backstory:   llmResp.Backstory,
					Skills:      skills,
					Inventory:   inventory,
					Secret:      secret,
					Status:      NPCStatusAlive,
				}
			}
		}
	}

	// 4. AC #3: Link to Global Seeds based on archetype (ALWAYS done)
	npc.LinkedSeeds = na.linkNPCToSeeds(request.Archetype, request.GlobalSeeds)

	// 5. AC #4: Calculate death timing for N-01 Sacrificial (ALWAYS done)
	if request.Archetype == NPCArchetypeSacrificial {
		npc.DeathTiming = na.calculateDeathTiming(request.PlotStructure)
	}

	return &GenerateResponse{NPC: npc}, nil
}

// InvokeBatchGenerate generates multiple NPC instances in a batch
//
// Story 7.6 AC #1: Phase 1 NPC Batch Generation
//   - Generate 2-4 NPCs (random count)
//   - Must include at least one N-01 Sacrificial (for teaching)
//   - Other NPCs randomly selected from different archetypes
//   - Generation time <10s using Smart Model
//
// Parameters:
//   - ctx: Context for cancellation
//   - count: Number of NPCs to generate (0 = random 2-4)
//   - archetypes: Specific archetypes to use (empty = random selection)
//   - storyContext: Story theme and scene
//   - globalSeeds: Available global seeds for linking
//   - plotStructure: Plot structure for death timing
//
// Returns:
//   - []NPCInstance: Generated NPC instances
//   - error: Error if batch generation fails
func (na *NPCAgent) InvokeBatchGenerate(
	ctx context.Context,
	count int,
	archetypes []NPCArchetype,
	storyContext StoryContext,
	globalSeeds []GlobalSeedInfo,
	plotStructure PlotStructure,
) ([]NPCInstance, error) {
	// 1. Determine count (2-4 NPCs)
	if count <= 0 {
		count = 2 + na.rng.Intn(3) // 2, 3, or 4
	}
	if count < 2 {
		count = 2
	}
	if count > 4 {
		count = 4
	}

	// 2. Determine archetypes
	selectedArchetypes := make([]NPCArchetype, 0, count)

	if len(archetypes) > 0 {
		// Use provided archetypes (exactly as provided, not more)
		// If archetypes provided, use them as-is
		selectedArchetypes = append(selectedArchetypes, archetypes...)
		// Update count to match provided archetypes length
		count = len(selectedArchetypes)
	} else {
		// AC #1: Must include at least one N-01 Sacrificial
		selectedArchetypes = append(selectedArchetypes, NPCArchetypeSacrificial)

		// AC #1: Randomly select other archetypes (avoid duplicates)
		for len(selectedArchetypes) < count {
			archetype := na.selectRandomArchetype(selectedArchetypes)
			selectedArchetypes = append(selectedArchetypes, archetype)
		}
	}

	// 3. Generate NPCs (with 10s timeout for entire batch)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	npcs := make([]NPCInstance, 0, count)
	for i, archetype := range selectedArchetypes {
		// Generate each NPC
		request := &GenerateRequest{
			Archetype:     archetype,
			StoryContext:  storyContext,
			GlobalSeeds:   globalSeeds,
			PlotStructure: plotStructure,
		}

		response, err := na.InvokeGenerate(ctxWithTimeout, request)
		if err != nil {
			// Log error but continue with fallback
			fmt.Printf("[NPCAgent] Warning: Failed to generate NPC %d (%s): %v\n", i+1, archetype, err)
			continue
		}

		npcs = append(npcs, response.NPC)
	}

	// 4. Ensure we have at least one NPC
	if len(npcs) == 0 {
		return nil, fmt.Errorf("failed to generate any NPCs")
	}

	return npcs, nil
}

// selectRandomArchetype selects a random archetype avoiding duplicates
func (na *NPCAgent) selectRandomArchetype(existing []NPCArchetype) NPCArchetype {
	allArchetypes := []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeHostile,
		NPCArchetypeNeutral,
		NPCArchetypeGuide,
		NPCArchetypeDeceiver,
	}

	// Try to find a non-duplicate archetype
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		archetype := allArchetypes[na.rng.Intn(len(allArchetypes))]

		// Check if already exists
		isDuplicate := false
		for _, existing := range existing {
			if existing == archetype {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			return archetype
		}
	}

	// If we can't find a non-duplicate, just return a random one
	return allArchetypes[na.rng.Intn(len(allArchetypes))]
}

// InvokeDialogue generates NPC dialogue
//
// AC #5: Generate dialogue 100-300 chars
// AC #6: Match dialogue style to archetype
// AC #7: Reveal clues for N-02 Knowledgeable based on tension
// AC #9: Adjust dialogue length/emotion based on tension
// AC #13: Generation time <500ms using Fast Model
func (na *NPCAgent) InvokeDialogue(ctx context.Context, request *DialogueRequest) (*DialogueResponse, error) {
	// 1. AC #10: Check if this is death dialogue
	if request.NPC.Status == NPCStatusDying || request.CurrentBeat == request.NPC.DeathTiming {
		return na.generateDeathDialogue(request), nil
	}

	// 2. Fast path: Use template fallback immediately when no LLM is configured
	// This ensures < 500ms response even without LLM (Story 7.7 AC #1)
	if na.Config.LLMClient == nil {
		return na.generateTemplateDialogue(request), nil
	}

	// 3. Build dialogue prompt
	prompt := na.buildDialoguePrompt(request)

	// 4. AC #13: Call Fast Model with 500ms timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	response, err := na.BaseAgentImpl.InvokeWithRetry(ctxWithTimeout, func(ctx context.Context) (any, error) {
		return na.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.8, // More creative for varied dialogue
			"max_tokens":  500,
		})
	})

	if err != nil {
		// Fallback to template dialogue on LLM failure
		return na.generateTemplateDialogue(request), nil
	}

	// 4. Parse response
	responseStr, ok := response.(string)
	if !ok {
		return na.generateTemplateDialogue(request), nil
	}

	llmResp, err := na.parseDialogueResponse(responseStr)
	if err != nil {
		return na.generateTemplateDialogue(request), nil
	}

	// Check if clue was revealed
	var seedRevealed *string
	if request.NPC.Archetype == NPCArchetypeKnowledgeable && len(request.NPC.LinkedSeeds) > 0 {
		seedID := request.NPC.LinkedSeeds[0]
		seedRevealed = &seedID
	}

	return &DialogueResponse{
		Dialogue:        llmResp.Dialogue,
		SeedRevealed:    seedRevealed,
		IsDeathDialogue: false,
	}, nil
}

// InvokeIntroduction generates Show-Don't-Tell introduction for NPC
//
// Story 7.6 AC #2: Show, Don't Tell principle
//   - Generate 100-200 char introduction
//   - Reveal personality through action/dialogue/item
//   - Explicitly forbid direct trait descriptions
//   - Use Fast Model (<500ms)
//   - Validate against forbidden phrases
func (na *NPCAgent) InvokeIntroduction(ctx context.Context, request *IntroductionRequest) (*IntroductionResponse, error) {
	// 1. Fast path: Use template fallback immediately when no LLM is configured
	if na.Config.LLMClient == nil {
		return na.generateTemplateIntroduction(request), nil
	}

	// 2. Build introduction prompt with strict Show-Don't-Tell instructions
	prompt := na.buildIntroductionPrompt(request)

	// 3. AC #13: Call Fast Model with 500ms timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	response, err := na.BaseAgentImpl.InvokeWithRetry(ctxWithTimeout, func(ctx context.Context) (any, error) {
		return na.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.8, // More creative for vivid descriptions
			"max_tokens":  300,
		})
	})

	if err != nil {
		// Fallback to template introduction on LLM failure
		return na.generateTemplateIntroduction(request), nil
	}

	// 3. Parse response
	responseStr, ok := response.(string)
	if !ok {
		return na.generateTemplateIntroduction(request), nil
	}

	llmResp, err := na.parseIntroductionResponse(responseStr)
	if err != nil {
		return na.generateTemplateIntroduction(request), nil
	}

	// 4. Validate Show-Don't-Tell compliance
	if !na.validateShowDontTell(llmResp.Introduction) {
		// If validation fails, regenerate with template
		return na.generateTemplateIntroduction(request), nil
	}

	return &IntroductionResponse{
		Introduction: llmResp.Introduction,
	}, nil
}

// generateNPCID generates a unique NPC ID
func (na *NPCAgent) generateNPCID() string {
	return fmt.Sprintf("NPC-%d", time.Now().UnixNano())
}

// linkNPCToSeeds links NPC to Global Seeds based on archetype
//
// AC #3: Archetype-based linking strategy
func (na *NPCAgent) linkNPCToSeeds(archetype NPCArchetype, globalSeeds []GlobalSeedInfo) []string {
	linkedSeeds := []string{}

	if len(globalSeeds) == 0 {
		return linkedSeeds
	}

	switch archetype {
	case NPCArchetypeSacrificial:
		// AC #3: N-01 不連結（死亡製造恐怖）
		return linkedSeeds

	case NPCArchetypeKnowledgeable:
		// AC #3: N-02 連結 1-2 個（線索來源）
		count := 1 + na.rng.Intn(2) // 1 or 2
		for i := 0; i < count && i < len(globalSeeds); i++ {
			linkedSeeds = append(linkedSeeds, globalSeeds[i].ID)
		}

	case NPCArchetypeHostile:
		// AC #3: N-03 不連結（障礙角色）
		return linkedSeeds

	case NPCArchetypeNeutral:
		// AC #3: N-04 可選連結 0-1 個（50% 概率）
		if na.rng.Float64() < 0.5 && len(globalSeeds) > 0 {
			linkedSeeds = append(linkedSeeds, globalSeeds[0].ID)
		}

	case NPCArchetypeGuide:
		// AC #3: N-05 連結 1 個（指引方向）
		if len(globalSeeds) > 0 {
			linkedSeeds = append(linkedSeeds, globalSeeds[0].ID)
		}

	case NPCArchetypeDeceiver:
		// AC #3: N-06 連結 1 個（假線索）
		if len(globalSeeds) > 0 {
			idx := na.rng.Intn(len(globalSeeds))
			linkedSeeds = append(linkedSeeds, globalSeeds[idx].ID)
		}
	}

	return linkedSeeds
}

// calculateDeathTiming calculates death timing for N-01 Sacrificial
//
// AC #4: Death timing based on three-act structure
func (na *NPCAgent) calculateDeathTiming(plotStructure PlotStructure) int {
	// Only for N-01 Sacrificial
	act2Start := plotStructure.Act2Range[0]
	act2End := plotStructure.Act2Range[1]
	act2Mid := (act2Start + act2End) / 2

	// AC #4: Probability-based selection
	randVal := na.rng.Float64()

	if randVal < 0.3 {
		// 30% probability: Act 2 early (25-40%)
		offset := (act2Mid - act2Start) / 2
		return act2Start + na.rng.Intn(offset+1)
	} else if randVal < 0.8 {
		// 50% probability: Act 2 midpoint (40-60%) - best dramatic timing
		offset := (act2End - act2Start) / 4
		return act2Mid - offset + na.rng.Intn(offset*2+1)
	} else {
		// 20% probability: Act 2 late (60-75%)
		offset := (act2End - act2Mid) * 3 / 4
		return act2Mid + na.rng.Intn(offset+1)
	}
}

// buildGeneratePrompt builds the LLM prompt for NPC generation
//
// AC #1, #2: Generate name, appearance, personality, backstory
func (na *NPCAgent) buildGeneratePrompt(request *GenerateRequest) string {
	var sb strings.Builder

	archetypeInfo := GetArchetypeInfo(request.Archetype)

	sb.WriteString("你是「規則怪談」恐怖遊戲的 NPC 設計師。\n\n")

	sb.WriteString("**任務**：根據 Archetype 模板生成 NPC 實例。\n\n")

	sb.WriteString(fmt.Sprintf("**Archetype**：%s - %s\n", archetypeInfo.ID, archetypeInfo.Name))
	sb.WriteString(fmt.Sprintf("**定義**：%s\n", archetypeInfo.Description))
	sb.WriteString(fmt.Sprintf("**特質**：%s\n\n", strings.Join(archetypeInfo.Traits, "、")))

	sb.WriteString("**輸出要求**：\n")
	sb.WriteString("- Name: 符合主題與場景的姓名（記憶深刻但不過於誇張）\n")
	sb.WriteString("- Personality: 3-5 個關鍵詞\n")
	sb.WriteString("- Appearance: 50-100 字外貌描述（包含衣著、髮型、特殊標記）\n")
	sb.WriteString("- Backstory: 100-200 字背景故事\n")
	sb.WriteString("- Skills: 2-3 個技能（如「急救」「開鎖」「戰鬥經驗」）\n")
	sb.WriteString("- Inventory: 1-3 個持有物品（如「生鏽的鑰匙」「破舊的日記」）\n")
	sb.WriteString("- Secret: 1 個隱藏秘密（50-100 字，與劇情或個性相關）\n\n")

	sb.WriteString("**輸出格式（JSON）**：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"name\": \"王小芳\",\n")
	sb.WriteString("  \"personality\": [\"無助\", \"恐懼\", \"善良\"],\n")
	sb.WriteString("  \"appearance\": \"外貌描述（50-100 字）\",\n")
	sb.WriteString("  \"backstory\": \"背景故事（100-200 字）\",\n")
	sb.WriteString("  \"skills\": [\"急救\", \"躲藏\"],\n")
	sb.WriteString("  \"inventory\": [\"破舊的護士服\", \"沾血的繃帶\"],\n")
	sb.WriteString("  \"secret\": \"她曾目睹醫院地下室的非法實驗，但因恐懼而選擇沉默\"\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("**故事上下文**：\n")
	sb.WriteString(fmt.Sprintf("主題：%s\n", request.StoryContext.Theme))
	sb.WriteString(fmt.Sprintf("場景：%s\n\n", request.StoryContext.Scene))

	if len(request.GlobalSeeds) > 0 {
		sb.WriteString("**Global Seeds 摘要**：\n")
		for _, seed := range request.GlobalSeeds {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", seed.ID, seed.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("請生成 %s NPC 實例。", archetypeInfo.Name))

	return sb.String()
}

// buildDialoguePrompt builds the LLM prompt for dialogue generation
//
// AC #6: Dialogue style mapping
// AC #7: Clue revelation for N-02 Knowledgeable
// AC #9: Tension-based length adjustment
// Story 7.7 AC #4: SAN state influence on dialogue style
// Story 7.7 AC #5: Clue revelation for N-02 Knowledgeable and N-03 Mystic/Inspirer
func (na *NPCAgent) buildDialoguePrompt(request *DialogueRequest) string {
	var sb strings.Builder

	sb.WriteString("你是「規則怪談」恐怖遊戲的 NPC 對話生成器。\n\n")

	sb.WriteString("**NPC 信息**：\n")
	sb.WriteString(fmt.Sprintf("- 姓名：%s\n", request.NPC.Name))
	sb.WriteString(fmt.Sprintf("- 原型：%s\n", request.NPC.Archetype))
	sb.WriteString(fmt.Sprintf("- 性格：%s\n", strings.Join(request.NPC.Personality, "、")))

	// Story 7.7 AC #4: Add NPC SAN state
	npcSAN := request.NPCSAN
	if npcSAN <= 0 {
		npcSAN = 100 // Default to full SAN if not provided
	}
	sb.WriteString(fmt.Sprintf("- SAN 值：%d/100\n\n", npcSAN))

	// AC #6: Dialogue style (adjusted by SAN)
	sb.WriteString("**對話風格**：\n")
	baseStyle := GetDialogueStyle(request.NPC.Archetype)
	sanStyle := na.getSANStyleModifier(npcSAN)
	sb.WriteString(fmt.Sprintf("%s\n%s\n\n", baseStyle, sanStyle))

	// AC #9: Dialogue length based on tension (adjusted by SAN)
	lengthRange := na.getDialogueLengthWithSAN(request.Tension, npcSAN)
	sb.WriteString(fmt.Sprintf("**對話長度**：%d-%d 字\n\n", lengthRange[0], lengthRange[1]))

	// Story 7.7 AC #5: Clue revelation for N-02 Knowledgeable and N-03 Mystic
	if na.shouldRevealClue(request.NPC.Archetype) && len(request.NPC.LinkedSeeds) > 0 {
		sb.WriteString("**線索揭露**：\n")
		clueHint := na.getClueRevealHint(request.Tension)
		sb.WriteString(fmt.Sprintf("%s\n", clueHint))
		sb.WriteString(fmt.Sprintf("相關線索：%s\n\n", strings.Join(request.NPC.LinkedSeeds, "、")))
	}

	sb.WriteString("**輸出格式（JSON）**：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"dialogue\": \"NPC 對話內容（%d-%d 字）\"\n", lengthRange[0], lengthRange[1]))
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("**當前上下文**：\n")
	sb.WriteString(fmt.Sprintf("場景：%s\n", request.Context))
	sb.WriteString(fmt.Sprintf("張力水平：%d\n", request.Tension))
	sb.WriteString(fmt.Sprintf("NPC SAN 狀態：%s\n\n", na.getSANStateName(npcSAN)))

	if request.PlayerQuestion != "" {
		sb.WriteString("**玩家問題**：\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", request.PlayerQuestion))
	}

	sb.WriteString("請生成符合 NPC 原型和 SAN 狀態的對話。")

	return sb.String()
}

// getClueRevealHint returns clue reveal hint based on tension
//
// Story 7.7 AC #5: Tension-based clue revelation
// - Tension 0-30: Very vague
// - Tension 30-60: Vague hints
// - Tension 60-80: Specific clues
// - Tension 80-100: Direct revelation (emergency)
func (na *NPCAgent) getClueRevealHint(tension int) string {
	if tension < 30 {
		// AC #5: Very vague (0-30)
		return "非常隱晦地提示有些事不太對勁，不直接說明（30-80 字）"
	} else if tension < 60 {
		// AC #5: Vague (30-60)
		return "模糊地提示注意某個現象或符號，但不說明具體內容（30-80 字）"
	} else if tension < 80 {
		// AC #5: Specific (60-80)
		return "提供具體線索，但保留部分關鍵信息（30-80 字）"
	} else {
		// AC #5: Direct (80-100)
		return "緊急情況下直接揭露重要線索，說出真相（30-80 字）"
	}
}

// getSANStyleModifier returns style modification based on NPC SAN state
//
// Story 7.7 AC #4: SAN State Influence on Dialogue
// - SAN 80-100 (正常): Normal archetype style
// - SAN 50-79 (焦慮): More tense, shorter sentences
// - SAN 20-49 (恐慌): Incoherent, repeating words
// - SAN 1-19 (崩潰): Trigger collapse behavior (archetype-dependent)
func (na *NPCAgent) getSANStyleModifier(san int) string {
	if san >= 80 {
		// SAN 80-100: Normal
		return "【SAN 正常】對話保持原型正常風格"
	} else if san >= 50 {
		// SAN 50-79: Anxious
		return "【SAN 焦慮】語氣更緊張，句子變短，有些結巴或停頓"
	} else if san >= 20 {
		// SAN 20-49: Panic
		return "【SAN 恐慌】開始語無倫次，重複某些詞，句子破碎不完整"
	} else {
		// SAN 1-19: Collapse
		return "【SAN 崩潰】觸發崩潰行為，依照原型表現極端反應（瘋狂、絕望、幻覺等）"
	}
}

// getSANStateName returns a readable SAN state name
func (na *NPCAgent) getSANStateName(san int) string {
	if san >= 80 {
		return "正常"
	} else if san >= 50 {
		return "焦慮"
	} else if san >= 20 {
		return "恐慌"
	} else if san >= 1 {
		return "崩潰"
	} else {
		return "未知"
	}
}

// getDialogueLengthWithSAN adjusts dialogue length based on tension and SAN
//
// Story 7.7 AC #4: Lower SAN = shorter, more fragmented dialogue
func (na *NPCAgent) getDialogueLengthWithSAN(tension int, san int) [2]int {
	// Get base length from tension
	baseLength := GetDialogueLengthRange(tension)

	// Adjust based on SAN state
	if san >= 80 {
		// Normal SAN: use base length
		return baseLength
	} else if san >= 50 {
		// Anxious: slightly shorter
		return [2]int{
			baseLength[0] * 80 / 100,
			baseLength[1] * 90 / 100,
		}
	} else if san >= 20 {
		// Panic: much shorter and fragmented
		return [2]int{
			baseLength[0] * 60 / 100,
			baseLength[1] * 70 / 100,
		}
	} else {
		// Collapse: very short, incoherent
		return [2]int{
			50,  // Minimum for collapse state
			100, // Maximum for collapse state
		}
	}
}

// shouldRevealClue determines if this NPC archetype should reveal clues
//
// Story 7.7 AC #5: N-02 Knowledgeable and N-03 Mystic/Inspirer reveal clues
func (na *NPCAgent) shouldRevealClue(archetype NPCArchetype) bool {
	return archetype == NPCArchetypeKnowledgeable || archetype == NPCArchetypeHostile
}

// parseGenerateResponse parses the LLM's JSON response for generation
func (na *NPCAgent) parseGenerateResponse(raw string) (*LLMGenerateResponse, error) {
	// Try to extract JSON from markdown code block
	jsonStr := na.extractJSON(raw)

	var result LLMGenerateResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return &result, nil
}

// parseDialogueResponse parses the LLM's JSON response for dialogue
func (na *NPCAgent) parseDialogueResponse(raw string) (*LLMDialogueResponse, error) {
	// Try to extract JSON from markdown code block
	jsonStr := na.extractJSON(raw)

	var result LLMDialogueResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return &result, nil
}

// extractJSON extracts JSON from markdown code block
func (na *NPCAgent) extractJSON(raw string) string {
	jsonStr := raw

	// Try to extract from ```json ... ```
	if start := strings.Index(raw, "```json"); start != -1 {
		jsonStr = raw[start+7:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	} else if start := strings.Index(raw, "```"); start != -1 {
		// Try to extract from ``` ... ```
		jsonStr = raw[start+3:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	}

	return strings.TrimSpace(jsonStr)
}

// generateFallbackNPC generates a fallback NPC using templates
func (na *NPCAgent) generateFallbackNPC(request *GenerateRequest) *GenerateResponse {
	archetypeInfo := GetArchetypeInfo(request.Archetype)

	name := na.generateName(request.Archetype, request.StoryContext.Theme)
	appearance := na.generateAppearance(request.Archetype, archetypeInfo.Traits)
	backstory := na.generateBackstory(request.Archetype, request.StoryContext.Theme)

	npc := NPCInstance{
		ID:          na.generateNPCID(),
		Name:        name,
		Archetype:   request.Archetype,
		Personality: archetypeInfo.Traits[:3], // Use first 3 traits
		Appearance:  appearance,
		Backstory:   backstory,
		Status:      NPCStatusAlive,
	}

	return &GenerateResponse{NPC: npc}
}

// generateName generates a name based on archetype and theme
//
// AC #2: Name符合主題與場景
func (na *NPCAgent) generateName(archetype NPCArchetype, theme string) string {
	var names []string

	switch theme {
	case "hospital", "醫院":
		names = []string{"王醫生", "李護士", "張主任", "陳醫師", "林護理師"}
	case "school", "學校":
		names = []string{"小明", "小芳", "張老師", "王同學", "李老師"}
	case "village", "鄉村":
		names = []string{"老王", "李大爺", "小翠", "張婆婆", "陳伯"}
	default:
		names = []string{"王先生", "李小姐", "張大哥", "陳姐", "林伯"}
	}

	return names[na.rng.Intn(len(names))]
}

// generateAppearance generates appearance description
//
// AC #2: Appearance 50-100 字，暗示 Archetype
func (na *NPCAgent) generateAppearance(archetype NPCArchetype, traits []string) string {
	var templates map[NPCArchetype]string = map[NPCArchetype]string{
		NPCArchetypeSacrificial:   "瘦弱的身軀，蒼白的臉色，眼神中充滿恐懼與無助。衣著凌亂，手腕上有不明傷痕。",
		NPCArchetypeKnowledgeable: "深邃的眼神，神秘的氣質。穿著整潔但略顯陳舊，總是若有所思地看著遠方。",
		NPCArchetypeHostile:       "冷漠的表情，危險的氣息。高大的身影投下陰影，讓人不由自主地感到威脅。",
		NPCArchetypeNeutral:       "普通的外貌，平凡的衣著。看起來就像任何一個普通人，毫不起眼。",
		NPCArchetypeGuide:         "溫和的笑容，關心的眼神。雖然看起來疲憊，但仍然願意伸出援手。",
		NPCArchetypeDeceiver:      "友善的表情，熱情的態度。但仔細觀察，會發現眼神中閃過一絲難以捉摸的東西。",
	}

	if template, ok := templates[archetype]; ok {
		return template
	}

	return "普通的外貌，沒有特別之處。"
}

// generateBackstory generates backstory
//
// AC #1: Backstory should be 100-200 chars
// H-6 FIX: Expanded all templates to meet 100-200 character requirement
func (na *NPCAgent) generateBackstory(archetype NPCArchetype, theme string) string {
	templates := map[NPCArchetype]string{
		NPCArchetypeSacrificial:   "一個不幸捲入這場噩夢的無辜者，過去平凡的生活已經一去不復返。每天都在恐懼中度過，不知道下一刻會發生什麼。曾經試圖逃離卻總是失敗，身心都受到了極大的折磨。現在只剩下恐懼與絕望，不知道明天是否還能活著，只能在黑暗中無助地等待著未知的命運降臨。",
		NPCArchetypeKnowledgeable: "似乎對這裡發生的一切有所了解，但總是欲言又止。曾經目睹過許多不可思議的事情，但每次想要說出真相時都會感到莫名的恐懼。過去的經歷讓他學會了保持沉默，因為知道太多的人往往活不長。那些試圖揭露真相的人都遭遇了不幸，這讓他更加謹慎，只在關鍵時刻才會透露一些隱晦的線索。",
		NPCArchetypeHostile:       "對外來者充滿敵意，認為所有人都是威脅。在這個地方生存了很長時間，見過太多人的死亡和背叛。過去的創傷讓他變得冷酷無情，只相信自己的生存法則，絕不會輕易相信任何人。曾經也有過同伴，但他們都因為信任他人而喪命，這段經歷徹底改變了他，讓他成為了一個孤獨而危險的存在。",
		NPCArchetypeNeutral:       "一個普通人，對周圍發生的怪事感到困惑與不安。本來只是路過這裡，卻不幸被困在這個詭異的地方。每天都在試圖尋找出路，但總是徒勞無功。只想盡快離開這裡，回到正常的生活，遠離這些恐怖的事情。對於這裡的規則一無所知，也不想深究，只希望能夠平安無事地活到明天。",
		NPCArchetypeGuide:         "曾經幫助過許多人逃離這個地方，雖然自己也傷痕累累。經歷了無數次生死考驗，逐漸摸清了一些規律和生存之道。仍然相信善良與希望，願意為他人指引方向，即使這可能讓自己陷入危險。見過太多人因為無知而死去，所以總是盡力提供幫助，希望能減少不必要的犧牲，讓更多人能夠活著走出這個噩夢。",
		NPCArchetypeDeceiver:      "表面上友善熱情，實則另有目的。曾經也是一個善良的人，但在這個地方的經歷改變了一切。過去的背叛讓他學會了偽裝和欺騙，現在只為自己的利益而行動，不惜犧牲他人來保全自己。深知這裡的某些秘密，並利用這些知識來操縱他人。在他看來，善良和信任只會帶來死亡，只有狡詐和自私才能讓自己活下去。",
	}

	if template, ok := templates[archetype]; ok {
		return template
	}

	return "一個神秘的人物，背景不明。沒有人知道他從哪裡來，也不知道他的真實目的是什麼。在這個詭異的地方已經待了很長時間，似乎對周圍的一切都很熟悉，但從不主動透露任何信息，總是保持著一種難以捉摸的距離感。"
}

// generateTemplateDialogue generates fallback dialogue using templates
//
// Story 7.7: Enhanced with SAN state influence
//
// Note: Template dialogues are intentionally shorter than AC #1's 100-300 character
// requirement for LLM-generated content. Templates serve as instant fallbacks when:
// - No LLM client is configured
// - LLM generation fails or times out
// - LLM response parsing fails
// The shorter templates ensure reliable performance under all conditions.
func (na *NPCAgent) generateTemplateDialogue(request *DialogueRequest) *DialogueResponse {
	npcSAN := request.NPCSAN
	if npcSAN <= 0 {
		npcSAN = 100 // Default to full SAN
	}

	var dialogue string

	// Generate dialogue based on archetype and SAN state
	if npcSAN >= 80 {
		// Normal SAN: standard archetype dialogue
		dialogue = na.getTemplateDialogueNormal(request.NPC.Archetype)
	} else if npcSAN >= 50 {
		// Anxious SAN: more tense dialogue
		dialogue = na.getTemplateDialogueAnxious(request.NPC.Archetype)
	} else if npcSAN >= 20 {
		// Panic SAN: fragmented dialogue
		dialogue = na.getTemplateDialoguePanic(request.NPC.Archetype)
	} else {
		// Collapse SAN: extreme collapse behavior
		dialogue = na.getTemplateDialogueCollapse(request.NPC.Archetype)
	}

	// Adjust for high tension
	if request.Tension > 80 {
		dialogue = strings.ReplaceAll(dialogue, "。", "...") // Add suspense
	}

	// Check if clue should be revealed
	var seedRevealed *string
	if na.shouldRevealClue(request.NPC.Archetype) && len(request.NPC.LinkedSeeds) > 0 {
		seedID := request.NPC.LinkedSeeds[0]
		seedRevealed = &seedID
	}

	return &DialogueResponse{
		Dialogue:        dialogue,
		SeedRevealed:    seedRevealed,
		IsDeathDialogue: false,
	}
}

// getTemplateDialogueNormal returns normal SAN dialogue templates
func (na *NPCAgent) getTemplateDialogueNormal(archetype NPCArchetype) string {
	switch archetype {
	case NPCArchetypeSacrificial:
		return "救...救救我...我不想死...這裡太可怕了..."
	case NPCArchetypeKnowledgeable:
		return "有些事...我不能說太多。你自己要小心，注意周圍的一切。"
	case NPCArchetypeHostile:
		return "這裡...有什麼在看著我們。我感覺到了...危險的氣息。"
	case NPCArchetypeNeutral:
		return "我也不知道發生了什麼...這裡變得很奇怪，我只想離開。"
	case NPCArchetypeGuide:
		return "小心前面的路，那裡有危險。記住，不要相信所有看到的東西。"
	case NPCArchetypeDeceiver:
		return "別擔心，我會幫你的。那邊很安全，相信我，跟我走就對了。"
	default:
		return "..."
	}
}

// getTemplateDialogueAnxious returns anxious SAN dialogue templates
func (na *NPCAgent) getTemplateDialogueAnxious(archetype NPCArchetype) string {
	switch archetype {
	case NPCArchetypeSacrificial:
		return "不...不要...我...我好怕...救命...救命啊..."
	case NPCArchetypeKnowledgeable:
		return "這個...不對勁...我知道...但是...我不能說...不能..."
	case NPCArchetypeHostile:
		return "危險...到處都是危險...它們...它們在靠近..."
	case NPCArchetypeNeutral:
		return "我...我得走了...不能再待了...這裡不對勁..."
	case NPCArchetypeGuide:
		return "等等...讓我想想...前面...前面很危險...我們..."
	case NPCArchetypeDeceiver:
		return "相信我...快...快跟我走...那邊...那邊安全..."
	default:
		return "..."
	}
}

// getTemplateDialoguePanic returns panic SAN dialogue templates
func (na *NPCAgent) getTemplateDialoguePanic(archetype NPCArchetype) string {
	switch archetype {
	case NPCArchetypeSacrificial:
		return "救命！救命！不要...不要...它來了...它來了！"
	case NPCArchetypeKnowledgeable:
		return "都是假的...都是假的...這不可能...不可能..."
	case NPCArchetypeHostile:
		return "看到了...我看到了...到處都是...到處..."
	case NPCArchetypeNeutral:
		return "逃！快逃！不管了...我要逃...要逃..."
	case NPCArchetypeGuide:
		return "我不知道...不知道了...哪裡...哪裡安全..."
	case NPCArchetypeDeceiver:
		return "走...快走...不...不對...往哪...往哪走..."
	default:
		return "..."
	}
}

// getTemplateDialogueCollapse returns collapse SAN dialogue templates
func (na *NPCAgent) getTemplateDialogueCollapse(archetype NPCArchetype) string {
	switch archetype {
	case NPCArchetypeSacrificial:
		return "（癱倒在地，眼神空洞，無法回應）"
	case NPCArchetypeKnowledgeable:
		return "假的...假的...假的...假的...（喃喃重複）"
	case NPCArchetypeHostile:
		return "（陷入永恆的幻覺，看見不存在的恐怖）"
	case NPCArchetypeNeutral:
		return "哈哈...哈哈哈...（瘋狂大笑）"
	case NPCArchetypeGuide:
		return "（蜷縮成一團，無止境地顫抖哭泣）"
	case NPCArchetypeDeceiver:
		return "（如同雕像，對外界毫無反應）"
	default:
		return "..."
	}
}

// generateDeathDialogue generates death dialogue for N-01 Sacrificial
//
// AC #10: Death dialogue with fear, warning, and emotional impact
func (na *NPCAgent) generateDeathDialogue(request *DialogueRequest) *DialogueResponse {
	var dialogue string

	// AC #10: Death dialogue structure
	fear := "不...不要...我不想死...這不可能..."
	warning := "不要相信他們...規則...記住規則..."
	death := fmt.Sprintf("%s倒下了，眼神中充滿絕望與恐懼...", request.NPC.Name)

	dialogue = fmt.Sprintf("%s\n\n%s\n\n%s", fear, warning, death)

	// If NPC has linked seeds, reveal important clue in death
	var seedRevealed *string
	if len(request.NPC.LinkedSeeds) > 0 {
		seedID := request.NPC.LinkedSeeds[0]
		seedRevealed = &seedID
		dialogue = fmt.Sprintf("%s\n\n（臨終前揭露了關於「%s」的重要線索）", dialogue, seedID)
	}

	return &DialogueResponse{
		Dialogue:        dialogue,
		SeedRevealed:    seedRevealed,
		IsDeathDialogue: true,
	}
}

// ==========================================================================
// Story 7.6: NPC Generation Enhancement - Skills, Inventory, Secret
// ==========================================================================

// generateSkills generates fallback skills based on archetype
func (na *NPCAgent) generateSkills(archetype NPCArchetype) []string {
	skillMap := map[NPCArchetype][]string{
		NPCArchetypeSacrificial:   {"躲藏", "求救"},
		NPCArchetypeKnowledgeable: {"觀察", "解謎", "記憶力"},
		NPCArchetypeHostile:       {"威嚇", "戰鬥"},
		NPCArchetypeNeutral:       {"日常技能", "溝通"},
		NPCArchetypeGuide:         {"急救", "引導", "鼓勵"},
		NPCArchetypeDeceiver:      {"說謊", "偽裝", "操縱"},
	}

	if skills, ok := skillMap[archetype]; ok {
		return skills
	}
	return []string{"基本技能"}
}

// generateInventory generates fallback inventory based on archetype and theme
func (na *NPCAgent) generateInventory(archetype NPCArchetype, theme string) []string {
	// Theme-specific items
	var themeItems []string
	switch theme {
	case "hospital", "醫院":
		themeItems = []string{"破舊的病歷", "生鏽的手術刀", "沾血的繃帶"}
	case "school", "學校":
		themeItems = []string{"舊課本", "破損的筆記", "褪色的照片"}
	case "village", "鄉村":
		themeItems = []string{"老舊的鑰匙", "祖傳的護身符", "發黃的信件"}
	default:
		themeItems = []string{"隨身物品", "舊照片", "神秘物件"}
	}

	// Archetype-specific items
	archetypeItems := map[NPCArchetype][]string{
		NPCArchetypeSacrificial:   {"撕裂的衣物", "求救信號"},
		NPCArchetypeKnowledgeable: {"神秘的筆記", "古老的文獻"},
		NPCArchetypeHostile:       {"生鏽的武器", "威脅工具"},
		NPCArchetypeNeutral:       {"日常用品"},
		NPCArchetypeGuide:         {"急救包", "照明工具"},
		NPCArchetypeDeceiver:      {"偽造的文件", "誘餌"},
	}

	// Combine theme and archetype items
	var inventory []string
	if len(themeItems) > 0 {
		inventory = append(inventory, themeItems[na.rng.Intn(len(themeItems))])
	}
	if items, ok := archetypeItems[archetype]; ok && len(items) > 0 {
		inventory = append(inventory, items[na.rng.Intn(len(items))])
	}

	if len(inventory) == 0 {
		inventory = []string{"神秘物品"}
	}

	return inventory
}

// generateSecret generates fallback secret based on archetype and theme
func (na *NPCAgent) generateSecret(archetype NPCArchetype, theme string) string {
	secretTemplates := map[NPCArchetype]string{
		NPCArchetypeSacrificial:   "曾經目睹過這裡發生的恐怖事件，但因恐懼而不敢說出真相，每晚都被噩夢纏繞",
		NPCArchetypeKnowledgeable: "知道這個地方的核心秘密，但受到某種約束無法直接揭露，只能用暗示引導他人",
		NPCArchetypeHostile:       "對這裡的人充滿怨恨，因為曾經遭受過背叛，現在只想讓他們付出代價",
		NPCArchetypeNeutral:       "無意間捲入這場噩夢，只想盡快逃離，不願捲入任何紛爭或秘密之中",
		NPCArchetypeGuide:         "曾經失去過重要的人，現在想幫助他人避免同樣的悲劇，即使代價是自己的生命",
		NPCArchetypeDeceiver:      "為了生存與某種邪惡力量達成了交易，必須犧牲他人來換取自己的安全",
	}

	if secret, ok := secretTemplates[archetype]; ok {
		return secret
	}

	return "隱藏著不為人知的秘密"
}

// ==========================================================================
// Story 7.6: Show-Don't-Tell Introduction
// ==========================================================================

// buildIntroductionPrompt builds the LLM prompt for Show-Don't-Tell introduction
//
// AC #2: Strict Show-Don't-Tell instructions with explicit forbidden phrases
func (na *NPCAgent) buildIntroductionPrompt(request *IntroductionRequest) string {
	var sb strings.Builder

	sb.WriteString("你是「規則怪談」恐怖遊戲的 NPC 描寫專家。\n\n")

	sb.WriteString("**任務**：使用 Show, Don't Tell 原則生成 NPC 首次登場描寫。\n\n")

	sb.WriteString("**NPC 信息**：\n")
	sb.WriteString(fmt.Sprintf("- 姓名：%s\n", request.NPC.Name))
	sb.WriteString(fmt.Sprintf("- 性格：%s\n", strings.Join(request.NPC.Personality, "、")))
	sb.WriteString(fmt.Sprintf("- 技能：%s\n", strings.Join(request.NPC.Skills, "、")))
	sb.WriteString(fmt.Sprintf("- 持有物品：%s\n", strings.Join(request.NPC.Inventory, "、")))
	sb.WriteString(fmt.Sprintf("- 秘密：%s\n\n", request.NPC.Secret))

	sb.WriteString("**Show, Don't Tell 原則（嚴格遵守）**：\n")
	sb.WriteString("✅ 正確示範：\n")
	sb.WriteString("- 「他的手顫抖著，眼神不停閃避，握緊了口袋裡的鏽鑰匙」（展示恐懼）\n")
	sb.WriteString("- 「她冷笑一聲，把染血的繃帶扔在桌上：『又一個不聽勸的』」（展示冷漠）\n")
	sb.WriteString("- 「他推了推眼鏡，從懷裡掏出一本破舊筆記，上面密密麻麻寫滿符號」（展示知識）\n\n")

	sb.WriteString("❌ 禁止直接描述（絕對不可出現）：\n")
	sb.WriteString("- 「他很恐懼」「她很冷靜」「他很神秘」「她很善良」\n")
	sb.WriteString("- 「他是個知識淵博的人」「她很無助」「他充滿敵意」\n")
	sb.WriteString("- 任何直接描述性格、情緒、特質的詞語\n\n")

	sb.WriteString("**必須透過以下方式展示性格**：\n")
	sb.WriteString("1. 具體動作：顫抖、逃避眼神、握拳、冷笑\n")
	sb.WriteString("2. 對話內容：簡短的話語（20 字內）\n")
	sb.WriteString("3. 持有物品：展示物品及其狀態\n\n")

	sb.WriteString("**輸出要求**：\n")
	sb.WriteString("- 100-200 字\n")
	sb.WriteString("- 必須包含動作、對話或物品中的至少兩項\n")
	sb.WriteString("- 絕對不可直接描述性格或情緒\n\n")

	sb.WriteString("**故事上下文**：\n")
	sb.WriteString(fmt.Sprintf("主題：%s\n", request.StoryContext.Theme))
	sb.WriteString(fmt.Sprintf("場景：%s\n\n", request.StoryContext.Scene))

	sb.WriteString("**輸出格式（JSON）**：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"introduction\": \"NPC 首次登場描寫（100-200 字，純 Show-Don't-Tell）\"\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("請生成符合 Show-Don't-Tell 原則的 NPC 登場描寫。")

	return sb.String()
}

// parseIntroductionResponse parses the LLM's JSON response for introduction
func (na *NPCAgent) parseIntroductionResponse(raw string) (*LLMIntroductionResponse, error) {
	// Try to extract JSON from markdown code block
	jsonStr := na.extractJSON(raw)

	var result LLMIntroductionResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return &result, nil
}

// validateShowDontTell validates that introduction follows Show-Don't-Tell principle
//
// AC #2: Check for forbidden direct trait descriptions
func (na *NPCAgent) validateShowDontTell(introduction string) bool {
	// Forbidden phrases that directly describe traits
	// Note: We check for these as complete words/phrases to avoid false positives
	forbiddenPhrases := []string{
		"很恐懼", "很害怕", "很冷靜", "很神秘", "很善良",
		"很無助", "很脆弱", "很危險", "很友善", "很冷漠",
		"充滿恐懼", "充滿敵意", "充滿善意", "充滿威脅", "充滿絕望",
		"是個", "是一個",
		"知識淵博", "經驗豐富", "勇敢無畏", "懦弱膽小",
		"顯得很", "看起來很",
	}

	// Check if any forbidden phrase exists
	for _, phrase := range forbiddenPhrases {
		if strings.Contains(introduction, phrase) {
			return false
		}
	}

	// Additional validation: introduction should not be too short
	// Relaxed to 50 characters since good Show-Don't-Tell can be concise
	runeCount := len([]rune(introduction))
	if runeCount < 50 {
		return false
	}

	return true
}

// generateTemplateIntroduction generates fallback Show-Don't-Tell introduction
//
// AC #2: Template-based introductions following Show-Don't-Tell principle
func (na *NPCAgent) generateTemplateIntroduction(request *IntroductionRequest) *IntroductionResponse {
	npc := request.NPC

	// Build introduction based on archetype using Show-Don't-Tell
	var introduction string

	switch npc.Archetype {
	case NPCArchetypeSacrificial:
		// Show fear through action and items
		item := "物品"
		if len(npc.Inventory) > 0 {
			item = npc.Inventory[0]
		}
		introduction = fmt.Sprintf("%s蜷縮在角落，雙手死死抓著%s，眼神不斷飄向門口。「救...救命...」聲音顫抖得幾乎聽不清，身體抖得像秋風中的落葉。",
			npc.Name, item)

	case NPCArchetypeKnowledgeable:
		// Show knowledge through items and mysterious behavior
		item := "一本破舊的筆記"
		if len(npc.Inventory) > 0 {
			item = npc.Inventory[0]
		}
		introduction = fmt.Sprintf("%s靠在牆邊，手中翻著%s，頭也不抬：「你來晚了。」指尖劃過書頁上的符號，眼神深邃得像看穿了一切。",
			npc.Name, item)

	case NPCArchetypeHostile:
		// Show hostility through action and dialogue
		item := "武器"
		if len(npc.Inventory) > 0 {
			item = npc.Inventory[0]
		}
		introduction = fmt.Sprintf("%s擋住去路，手中握著%s，冷笑著：「又來送死？」眼神裡的殺意毫不掩飾，雙拳緊握如蓄勢待發的猛獸。",
			npc.Name, item)

	case NPCArchetypeNeutral:
		// Show confusion through action and dialogue
		introduction = fmt.Sprintf("%s站在原地，眼神茫然地看著四周：「這...這是哪裡？」雙手不安地摩擦著，像是在確認自己是否還活著。",
			npc.Name)

	case NPCArchetypeGuide:
		// Show care through action and items
		item := "急救包"
		if len(npc.Inventory) > 0 {
			item = npc.Inventory[0]
		}
		introduction = fmt.Sprintf("%s快步走來，從背包掏出%s：「受傷了嗎？先處理一下。」疲憊的臉上仍帶著關切，動作熟練而溫和，眼中透露出經歷過太多生死的滄桑感。",
			npc.Name, item)

	case NPCArchetypeDeceiver:
		// Show deception through false friendliness and items
		item := "物品"
		if len(npc.Inventory) > 0 {
			item = npc.Inventory[0]
		}
		introduction = fmt.Sprintf("%s笑容滿面地迎上來，遞過%s：「別怕，我會幫你的。」但那笑容下，眼神閃過一絲難以察覺的算計。",
			npc.Name, item)

	default:
		introduction = fmt.Sprintf("%s靜靜地站在那裡，沒有說話。", npc.Name)
	}

	return &IntroductionResponse{
		Introduction: introduction,
	}
}
