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

		npc = NPCInstance{
			ID:          na.generateNPCID(),
			Name:        name,
			Archetype:   request.Archetype,
			Personality: archetypeInfo.Traits[:3], // Use first 3 traits
			Appearance:  appearance,
			Backstory:   backstory,
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

			npc = NPCInstance{
				ID:          na.generateNPCID(),
				Name:        name,
				Archetype:   request.Archetype,
				Personality: archetypeInfo.Traits[:3],
				Appearance:  appearance,
				Backstory:   backstory,
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

				npc = NPCInstance{
					ID:          na.generateNPCID(),
					Name:        name,
					Archetype:   request.Archetype,
					Personality: archetypeInfo.Traits[:3],
					Appearance:  appearance,
					Backstory:   backstory,
					Status:      NPCStatusAlive,
				}
			} else {
				npc = NPCInstance{
					ID:          na.generateNPCID(),
					Name:        llmResp.Name,
					Archetype:   request.Archetype,
					Personality: llmResp.Personality,
					Appearance:  llmResp.Appearance,
					Backstory:   llmResp.Backstory,
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

	// 2. Build dialogue prompt
	prompt := na.buildDialoguePrompt(request)

	// 3. AC #13: Call Fast Model with 500ms timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	response, err := na.BaseAgentImpl.InvokeWithRetry(ctxWithTimeout, func(ctx context.Context) (any, error) {
		if na.Config.LLMClient == nil {
			return nil, fmt.Errorf("no LLM client available")
		}

		return na.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.8, // More creative for varied dialogue
			"max_tokens":  500,
		})
	})

	if err != nil {
		// Fallback to template dialogue
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
	sb.WriteString("- Backstory: 100-200 字背景故事\n\n")

	sb.WriteString("**輸出格式（JSON）**：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"name\": \"王小芳\",\n")
	sb.WriteString("  \"personality\": [\"無助\", \"恐懼\", \"善良\"],\n")
	sb.WriteString("  \"appearance\": \"外貌描述（50-100 字）\",\n")
	sb.WriteString("  \"backstory\": \"背景故事（100-200 字）\"\n")
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
// AC #7: Clue revelation for N-02
// AC #9: Tension-based length adjustment
func (na *NPCAgent) buildDialoguePrompt(request *DialogueRequest) string {
	var sb strings.Builder

	sb.WriteString("你是「規則怪談」恐怖遊戲的 NPC 對話生成器。\n\n")

	sb.WriteString("**NPC 信息**：\n")
	sb.WriteString(fmt.Sprintf("- 姓名：%s\n", request.NPC.Name))
	sb.WriteString(fmt.Sprintf("- 原型：%s\n", request.NPC.Archetype))
	sb.WriteString(fmt.Sprintf("- 性格：%s\n\n", strings.Join(request.NPC.Personality, "、")))

	// AC #6: Dialogue style
	sb.WriteString("**對話風格**：\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", GetDialogueStyle(request.NPC.Archetype)))

	// AC #9: Dialogue length based on tension
	lengthRange := GetDialogueLengthRange(request.Tension)
	sb.WriteString(fmt.Sprintf("**對話長度**：%d-%d 字\n\n", lengthRange[0], lengthRange[1]))

	// AC #7: Clue revelation for N-02 Knowledgeable
	if request.NPC.Archetype == NPCArchetypeKnowledgeable && len(request.NPC.LinkedSeeds) > 0 {
		sb.WriteString("**線索揭露**：\n")
		clueHint := na.getClueRevealHint(request.Tension)
		sb.WriteString(fmt.Sprintf("%s\n\n", clueHint))
	}

	sb.WriteString("**輸出格式（JSON）**：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"dialogue\": \"NPC 對話內容（%d-%d 字）\"\n", lengthRange[0], lengthRange[1]))
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("**當前上下文**：\n")
	sb.WriteString(fmt.Sprintf("場景：%s\n", request.Context))
	sb.WriteString(fmt.Sprintf("張力水平：%d\n\n", request.Tension))

	if request.PlayerQuestion != "" {
		sb.WriteString("**玩家問題**：\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", request.PlayerQuestion))
	}

	sb.WriteString("請生成 NPC 對話。")

	return sb.String()
}

// getClueRevealHint returns clue reveal hint based on tension
//
// AC #7: Tension-based clue revelation
func (na *NPCAgent) getClueRevealHint(tension int) string {
	if tension < 30 {
		// AC #7: Very vague (0-30)
		return "非常隱晦地提示有些事不太對勁，不直接說明"
	} else if tension < 60 {
		// AC #7: Vague (30-60)
		return "模糊地提示注意某個現象或符號，但不說明具體內容"
	} else if tension < 80 {
		// AC #7: Specific (60-80)
		return "提供具體線索，但保留部分關鍵信息"
	} else {
		// AC #7: Direct (80-100)
		return "緊急情況下直接揭露重要線索，說出真相"
	}
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
func (na *NPCAgent) generateTemplateDialogue(request *DialogueRequest) *DialogueResponse {
	var dialogue string

	switch request.NPC.Archetype {
	case NPCArchetypeSacrificial:
		dialogue = "救...救救我...我不想死...這裡太可怕了..."
	case NPCArchetypeKnowledgeable:
		dialogue = "有些事...我不能說太多。你自己要小心，注意周圍的一切。"
	case NPCArchetypeHostile:
		dialogue = "你不該來這裡。現在...離開，否則後果自負。"
	case NPCArchetypeNeutral:
		dialogue = "我也不知道發生了什麼...這裡變得很奇怪，我只想離開。"
	case NPCArchetypeGuide:
		dialogue = "小心前面的路，那裡有危險。記住，不要相信所有看到的東西。"
	case NPCArchetypeDeceiver:
		dialogue = "別擔心，我會幫你的。那邊很安全，相信我，跟我走就對了。"
	default:
		dialogue = "..."
	}

	// Adjust for tension
	if request.Tension > 80 {
		dialogue = strings.ReplaceAll(dialogue, "。", "...") // Add suspense
	}

	return &DialogueResponse{
		Dialogue:        dialogue,
		SeedRevealed:    nil,
		IsDeathDialogue: false,
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
