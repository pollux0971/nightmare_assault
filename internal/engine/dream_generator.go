package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// DreamGenerator handles dream content generation using Smart Model
type DreamGenerator struct {
	client SmartModelClient
}

// SmartModelClient is an interface for AI model API clients
type SmartModelClient interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// NewDreamGenerator creates a new dream generator
func NewDreamGenerator(client SmartModelClient) *DreamGenerator {
	return &DreamGenerator{
		client: client,
	}
}

// GenerateOpeningDream generates an opening dream that hints at hidden rules
func (dg *DreamGenerator) GenerateOpeningDream(ctx context.Context, theme, rulesSummary, playerRole string) (string, error) {
	prompt := dg.buildOpeningDreamPrompt(theme, rulesSummary, playerRole)

	content, err := dg.client.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate opening dream: %w", err)
	}

	// Validate content length (200-400 characters as per AC1)
	if len(content) < 200 {
		// If too short, pad with atmospheric text
		content += "\n\n你感到一陣寒意，彷彿有什麼東西正在注視著你..."
	}

	return strings.TrimSpace(content), nil
}

// buildOpeningDreamPrompt constructs the prompt for opening dream generation
func (dg *DreamGenerator) buildOpeningDreamPrompt(theme, rulesSummary, playerRole string) string {
	return fmt.Sprintf(`你是恐怖遊戲的夢境設計師。生成一段開場夢境，預告即將發生的恐怖故事。

故事主題：%s
潛規則概要（隱藏）：%s
玩家角色：%s

要求：
1. 長度 200-400 字
2. 使用隱喻與象徵（不直接揭露規則）
3. 氛圍迷幻、邏輯略顯不連貫（符合夢境特性）
4. 包含 2-3 個意象（如鏡子、門、迷宮、影子）
5. 暗示規則但不明說（例如「鏡中的你做著相反的動作」暗示對立規則）
6. 結尾留懸念，引導進入正式故事
7. 使用第二人稱「你」來敘述

輸出格式：純夢境敘事文字，無額外說明。`, theme, rulesSummary, playerRole)
}

// CreateDreamRecord creates a DreamRecord from generated content
func (dg *DreamGenerator) CreateDreamRecord(
	dreamType game.DreamType,
	content string,
	relatedRuleID string,
	context game.DreamContext,
) game.DreamRecord {
	return game.DreamRecord{
		ID:            fmt.Sprintf("%s-%d", dreamType, time.Now().Unix()),
		Type:          dreamType,
		Timestamp:     time.Now(),
		Content:       content,
		RelatedRuleID: relatedRuleID,
		Context:       context,
	}
}

// ChapterDreamType represents different types of chapter dreams
type ChapterDreamType string

const (
	DreamTypeNightmare ChapterDreamType = "nightmare" // 噩夢: 重演恐怖時刻
	DreamTypeHint      ChapterDreamType = "hint"      // 提示: 線索關聯
	DreamTypeGrief     ChapterDreamType = "grief"     // 悲傷: 隊友死亡
	DreamTypeWarning   ChapterDreamType = "warning"   // 警告: 即將到來的危險
	DreamTypeRandom    ChapterDreamType = "random"    // 隨機: 一般情境
)

// ChapterDreamContext contains context for generating chapter dreams
type ChapterDreamContext struct {
	ChapterNum     int
	RecentEvents   string
	RuleHints      string
	PlayerSAN      int
	KnownClues     []string
	DeadTeammates  []string
	HighStress     bool
	RecentClue     bool
	TeammateDeaths bool
}

// GenerateChapterDream generates a chapter dream based on current game state and dream type
func (dg *DreamGenerator) GenerateChapterDream(ctx context.Context, dreamType ChapterDreamType, context ChapterDreamContext) (string, error) {
	prompt := dg.buildChapterDreamPrompt(dreamType, context)

	content, err := dg.client.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate chapter dream: %w", err)
	}

	// Ensure minimum length
	if len(content) < 100 {
		content += "\n\n你驚醒了，心跳加速..."
	}

	return strings.TrimSpace(content), nil
}

// buildChapterDreamPrompt constructs the prompt for chapter dream generation based on type
func (dg *DreamGenerator) buildChapterDreamPrompt(dreamType ChapterDreamType, context ChapterDreamContext) string {
	switch dreamType {
	case DreamTypeNightmare:
		return dg.buildNightmarePrompt(context)
	case DreamTypeHint:
		return dg.buildHintDreamPrompt(context)
	case DreamTypeGrief:
		return dg.buildGriefDreamPrompt(context)
	case DreamTypeWarning:
		return dg.buildWarningDreamPrompt(context)
	case DreamTypeRandom:
		return dg.buildRandomDreamPrompt(context)
	default:
		return dg.buildRandomDreamPrompt(context)
	}
}

// buildNightmarePrompt creates prompt for nightmare dreams
func (dg *DreamGenerator) buildNightmarePrompt(context ChapterDreamContext) string {
	return fmt.Sprintf(`生成一段噩夢，反映玩家剛經歷的恐怖事件。

最近事件：%s
玩家當前 SAN：%d
已知線索：%v

要求：
1. 長度 100-200 字
2. 重演恐怖時刻但扭曲變形
3. 強化恐懼感（加入超現實元素）
4. 可能暗示即將到來的危險
5. 結尾突然驚醒
6. 使用第二人稱「你」來敘述

輸出格式：純夢境敘事文字。`, context.RecentEvents, context.PlayerSAN, context.KnownClues)
}

// buildHintDreamPrompt creates prompt for hint dreams
func (dg *DreamGenerator) buildHintDreamPrompt(context ChapterDreamContext) string {
	return fmt.Sprintf(`生成一段提示夢境，以隱喻方式呈現玩家已知線索。

已知線索：%v
潛規則提示：%s
當前章節：%d

要求：
1. 長度 100-200 字
2. 將線索轉化為象徵意象
3. 不直接揭露，但提供聯想方向
4. 保持夢境的超現實感
5. 玩家事後回想時能恍然大悟
6. 使用第二人稱「你」來敘述

輸出格式：純夢境敘事文字。`, context.KnownClues, context.RuleHints, context.ChapterNum)
}

// buildGriefDreamPrompt creates prompt for grief dreams
func (dg *DreamGenerator) buildGriefDreamPrompt(context ChapterDreamContext) string {
	return fmt.Sprintf(`生成一段悲傷夢境，處理隊友死亡的情感。

死亡的隊友：%v
當前章節：%d
最近事件：%s

要求：
1. 長度 100-200 字
2. 隊友回憶片段（溫暖/悲傷交織）
3. 情感告別（夢境中的對話或場景）
4. 可能暗示隊友的遺願或線索
5. 結尾帶有不捨感
6. 使用第二人稱「你」來敘述

輸出格式：純夢境敘事文字。`, context.DeadTeammates, context.ChapterNum, context.RecentEvents)
}

// buildWarningDreamPrompt creates prompt for warning dreams
func (dg *DreamGenerator) buildWarningDreamPrompt(context ChapterDreamContext) string {
	return fmt.Sprintf(`生成一段警告夢境，暗示即將到來的危險。

玩家當前 SAN：%d
潛規則提示：%s
當前章節：%d

要求：
1. 長度 100-200 字
2. 使用恐怖意象暗示危險
3. 不明確說明威脅，但營造緊張感
4. 可能包含警示符號（紅色、破碎、陰影）
5. 結尾留下不安感
6. 使用第二人稱「你」來敘述

輸出格式：純夢境敘事文字。`, context.PlayerSAN, context.RuleHints, context.ChapterNum)
}

// buildRandomDreamPrompt creates prompt for random dreams
func (dg *DreamGenerator) buildRandomDreamPrompt(context ChapterDreamContext) string {
	return fmt.Sprintf(`生成一段隨機夢境，與遊戲主題相關的超現實場景。

當前章節：%d
最近事件：%s

要求：
1. 長度 100-200 字
2. 超現實、奇幻的場景
3. 與主題相關但不直接關聯
4. 可能帶有一絲詭異感
5. 結尾平和或略帶疑惑
6. 使用第二人稱「你」來敘述

輸出格式：純夢境敘事文字。`, context.ChapterNum, context.RecentEvents)
}

// DetermineDreamProbability calculates dream trigger probability based on context
func DetermineDreamProbability(context ChapterDreamContext) (ChapterDreamType, float64) {
	// High stress after traumatic event: 70% nightmare
	if context.HighStress {
		return DreamTypeNightmare, 0.70
	}

	// Teammate death: 80% grief dream
	if context.TeammateDeaths {
		return DreamTypeGrief, 0.80
	}

	// Low SAN: 60% warning dream
	if context.PlayerSAN < 30 {
		return DreamTypeWarning, 0.60
	}

	// Recent clue discovered: 50% hint dream
	if context.RecentClue {
		return DreamTypeHint, 0.50
	}

	// Normal situation: 20% random dream
	return DreamTypeRandom, 0.20
}
