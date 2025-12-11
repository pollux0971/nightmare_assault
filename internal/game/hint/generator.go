// Package hint provides hint generation for Nightmare Assault.
package hint

import (
	"fmt"
	"math/rand"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// HintType represents the category of hint to generate.
type HintType int

const (
	// HintTypeDirection suggests exploration direction.
	HintTypeDirection HintType = iota
	// HintTypeClue points to missed clues.
	HintTypeClue
	// HintTypeRule warns about upcoming rules.
	HintTypeRule
	// HintTypeItem hints about usable items.
	HintTypeItem
	// HintTypeAtmosphere provides general atmosphere hints.
	HintTypeAtmosphere
)

// String returns the display name of the hint type.
func (t HintType) String() string {
	switch t {
	case HintTypeDirection:
		return "探索方向"
	case HintTypeClue:
		return "線索提示"
	case HintTypeRule:
		return "危險警告"
	case HintTypeItem:
		return "道具暗示"
	case HintTypeAtmosphere:
		return "氛圍感知"
	default:
		return "未知"
	}
}

// HintContext contains information for hint generation.
type HintContext struct {
	// Type of hint to generate
	Type HintType
	// CurrentScene is the current location description
	CurrentScene string
	// RecentActions are the player's recent actions
	RecentActions []string
	// MissedClues are clues the player hasn't discovered
	MissedClues []string
	// UpcomingRules are rules that might trigger soon
	UpcomingRules []string
	// AvailableItems are items the player hasn't used
	AvailableItems []string
	// TurnsSinceMove is how long the player has been in this location
	TurnsSinceMove int
	// Chapter is the current chapter number
	Chapter int
}

// HintResult contains the generated hint and metadata.
type HintResult struct {
	// Text is the hint content
	Text string
	// Type is the hint category
	Type HintType
	// SANCost is how much SAN this hint cost
	SANCost int
	// Generated indicates if the hint was LLM-generated (vs fallback)
	Generated bool
}

// StateAnalyzer analyzes game state to determine hint context.
type StateAnalyzer struct{}

// NewStateAnalyzer creates a new state analyzer.
func NewStateAnalyzer() *StateAnalyzer {
	return &StateAnalyzer{}
}

// Analyze examines the game state and returns hint context.
func (sa *StateAnalyzer) Analyze(
	currentScene string,
	recentActions []string,
	turnsSinceMove int,
	chapter int,
	missedClues []string,
	upcomingRules []string,
	availableItems []string,
) *HintContext {
	ctx := &HintContext{
		CurrentScene:   currentScene,
		RecentActions:  recentActions,
		MissedClues:    missedClues,
		UpcomingRules:  upcomingRules,
		AvailableItems: availableItems,
		TurnsSinceMove: turnsSinceMove,
		Chapter:        chapter,
	}

	// Determine priority (AC5)
	ctx.Type = sa.determinePriority(ctx)

	return ctx
}

// IsStuck determines if the player appears stuck.
func (sa *StateAnalyzer) IsStuck(turnsSinceMove int) bool {
	return turnsSinceMove >= 5
}

// determinePriority decides what type of hint to provide (AC5 order).
func (sa *StateAnalyzer) determinePriority(ctx *HintContext) HintType {
	// 1. If player is stuck → direction
	if ctx.TurnsSinceMove >= 5 {
		return HintTypeDirection
	}

	// 2. If there are missed clues → clue
	if len(ctx.MissedClues) > 0 {
		return HintTypeClue
	}

	// 3. If there are upcoming rules → warning
	if len(ctx.UpcomingRules) > 0 {
		return HintTypeRule
	}

	// 4. If there are unused items → item
	if len(ctx.AvailableItems) > 0 {
		return HintTypeItem
	}

	// 5. Default → atmosphere
	return HintTypeAtmosphere
}

// Generator generates hints based on game state.
type Generator struct {
	analyzer       *StateAnalyzer
	baseCost       int
	costIncrement  int
	usageCount     map[int]int // chapter -> usage count
	currentChapter int
}

// NewGenerator creates a new hint generator.
func NewGenerator() *Generator {
	return &Generator{
		analyzer:      NewStateAnalyzer(),
		baseCost:      10,
		costIncrement: 5,
		usageCount:    make(map[int]int),
	}
}

// SetChapter sets the current chapter (for cost tracking).
func (g *Generator) SetChapter(chapter int) {
	if chapter != g.currentChapter {
		g.currentChapter = chapter
		// Reset usage count for new chapter
	}
}

// GetCurrentCost returns the current SAN cost for a hint.
func (g *Generator) GetCurrentCost(chapter int) int {
	usages := g.usageCount[chapter]
	if usages >= 3 {
		// Cap at 3 increments
		return g.baseCost + (2 * g.costIncrement) // 10 + 10 = 20 max
	}
	if usages >= 2 {
		return g.baseCost + g.costIncrement // 15
	}
	return g.baseCost // 10
}

// IncrementUsage records a hint usage for the chapter.
func (g *Generator) IncrementUsage(chapter int) {
	g.usageCount[chapter]++
}

// GetUsageCount returns how many hints have been used this chapter.
func (g *Generator) GetUsageCount(chapter int) int {
	return g.usageCount[chapter]
}

// ResetChapterUsage resets the usage count for a chapter.
func (g *Generator) ResetChapterUsage(chapter int) {
	delete(g.usageCount, chapter)
}

// CanAffordHint checks if the player has enough SAN.
func (g *Generator) CanAffordHint(currentSAN, chapter int) bool {
	return currentSAN >= g.GetCurrentCost(chapter)
}

// Generate creates a hint based on the provided context.
// If llmHint is empty, uses fallback hints.
func (g *Generator) Generate(ctx *HintContext, llmHint string) *HintResult {
	result := &HintResult{
		Type:      ctx.Type,
		SANCost:   g.GetCurrentCost(ctx.Chapter),
		Generated: llmHint != "",
	}

	if llmHint != "" {
		result.Text = llmHint
	} else {
		result.Text = g.getFallbackHint(ctx)
	}

	return result
}

// getFallbackHint returns a fallback hint if LLM fails.
func (g *Generator) getFallbackHint(ctx *HintContext) string {
	hints := g.getFallbackHintsForType(ctx.Type)
	if len(hints) == 0 {
		return "有些事情，只有親自探索才能發現..."
	}

	// Select a random hint (Go 1.20+ auto-seeds the global source)
	return hints[rand.Intn(len(hints))]
}

// getFallbackHintsForType returns fallback hints for a specific type.
func (g *Generator) getFallbackHintsForType(t HintType) []string {
	switch t {
	case HintTypeDirection:
		return []string{
			"或許換個方向探索會有不同的發現...",
			"這裡似乎沒有更多線索了，試著離開這裡吧...",
			"有時候，退一步反而能看清全貌...",
			"別在同一個地方徘徊太久，時間不等人...",
		}
	case HintTypeClue:
		return []string{
			"仔細觀察你周圍的環境，有些細節並非偶然...",
			"某些東西的位置似乎不太對勁...",
			"或許你錯過了什麼重要的東西...",
			"有些線索藏在顯眼的地方，卻容易被忽略...",
		}
	case HintTypeRule:
		return []string{
			"回想你聽到的聲音，它們可能在警告你什麼...",
			"這裡有些不成文的規矩，最好小心遵守...",
			"有些禁忌最好不要觸犯...",
			"注意周圍的異常，它們可能是警告的信號...",
		}
	case HintTypeItem:
		return []string{
			"有些東西最好不要碰，除非你確定它們的用途...",
			"你攜帶的東西或許比想像中更有用...",
			"工具在正確的時機使用才能發揮最大效用...",
			"別忘了你身上還有些東西...",
		}
	case HintTypeAtmosphere:
		return []string{
			"空氣中瀰漫著一股不安的氣息...",
			"你感覺到有什麼在暗處窺視著你...",
			"這個地方的氛圍讓人不寒而慄...",
			"信任你的直覺，它往往比理性更敏銳...",
		}
	default:
		return []string{
			"有些事情，只有親自探索才能發現...",
		}
	}
}

// BuildHintPrompt creates a prompt for LLM hint generation.
func BuildHintPrompt(ctx *HintContext) string {
	typeContext := ""
	switch ctx.Type {
	case HintTypeDirection:
		typeContext = "玩家似乎在原地徘徊太久，需要引導他們探索其他方向。"
	case HintTypeClue:
		if len(ctx.MissedClues) > 0 {
			typeContext = fmt.Sprintf("玩家錯過了一些線索：%v。暗示其中一個線索的存在。", ctx.MissedClues)
		} else {
			typeContext = "暗示玩家可能錯過了什麼重要的東西。"
		}
	case HintTypeRule:
		if len(ctx.UpcomingRules) > 0 {
			typeContext = "有隱藏的規則可能即將觸發。用模糊的方式警告玩家，但不能直接說出規則內容。"
		} else {
			typeContext = "暗示這個地方可能有些禁忌需要注意。"
		}
	case HintTypeItem:
		if len(ctx.AvailableItems) > 0 {
			typeContext = fmt.Sprintf("玩家有一些道具可以使用：%v。暗示其中一個道具可能有用。", ctx.AvailableItems)
		} else {
			typeContext = "暗示玩家手上的東西可能派得上用場。"
		}
	case HintTypeAtmosphere:
		typeContext = "提供一個關於當前場景氛圍的感知提示，增加恐怖感。"
	}

	recentActionsStr := "無"
	if len(ctx.RecentActions) > 0 {
		recentActionsStr = ""
		for i, action := range ctx.RecentActions {
			if i > 0 {
				recentActionsStr += ", "
			}
			recentActionsStr += action
			if i >= 2 { // Limit to 3 actions
				break
			}
		}
	}

	prompt := fmt.Sprintf(`你是一個恐怖遊戲的神秘嚮導。玩家請求提示。

**當前場景：**
%s

**玩家最近行動：**
%s

**提示類型：**
%s

**提示上下文：**
%s

**請生成一個 20-50 字的模糊暗示，符合以下原則：**
1. 不直接說出答案或規則內容
2. 使用隱喻、暗示、氛圍描述
3. 保持恐怖遊戲調性
4. 引導但不指明具體方向
5. 使用繁體中文

**輸出格式：**
只輸出提示文字，不要包含其他內容。`,
		ctx.CurrentScene,
		recentActionsStr,
		ctx.Type.String(),
		typeContext,
	)

	return prompt
}

// HintState tracks hint usage for a game session.
type HintState struct {
	// UsageByChapter tracks how many hints used per chapter
	UsageByChapter map[int]int `json:"usage_by_chapter"`
	// TotalUsage is the total hints used this session
	TotalUsage int `json:"total_usage"`
	// LastHintChapter is the chapter of the last hint
	LastHintChapter int `json:"last_hint_chapter"`
	// Enabled indicates if hints are available (not Hell mode)
	Enabled bool `json:"enabled"`
}

// NewHintState creates a new hint state.
func NewHintState(difficulty game.DifficultyLevel) *HintState {
	return &HintState{
		UsageByChapter: make(map[int]int),
		TotalUsage:     0,
		Enabled:        difficulty != game.DifficultyHell,
	}
}

// RecordUsage records a hint usage.
func (s *HintState) RecordUsage(chapter int) {
	s.UsageByChapter[chapter]++
	s.TotalUsage++
	s.LastHintChapter = chapter
}

// GetChapterUsage returns the usage count for a chapter.
func (s *HintState) GetChapterUsage(chapter int) int {
	return s.UsageByChapter[chapter]
}

// IsEnabled returns if hints are enabled.
func (s *HintState) IsEnabled() bool {
	return s.Enabled
}
