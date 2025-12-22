package agents

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// FakeOptionGenerator 負責生成假選項（幻覺選項）
// Story 9-6: 理智幻覺 - 假選項系統
//
// 設計原則：
//   - SAN < 30 時，30% 機率插入 1 個假選項
//   - SAN < 15 時，50% 機率插入 1-2 個假選項
//   - 假選項看起來合理但選擇後會造成 SAN/HP 損失
//   - 假選項有內部標記，選擇後觸發幻覺反應
//   - 假選項文字略微詭異但不明顯
type FakeOptionGenerator struct {
	rng *rand.Rand
}

// NewFakeOptionGenerator 創建新的假選項生成器
func NewFakeOptionGenerator(seed int64) *FakeOptionGenerator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &FakeOptionGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// FakeOptionCategory 假選項類別
type FakeOptionCategory int

const (
	// FakeCategoryFalseSafety 虛假安全：不存在的安全選項
	FakeCategoryFalseSafety FakeOptionCategory = iota
	// FakeCategoryInducedDanger 誘導危險：跟隨幻覺
	FakeCategoryInducedDanger
	// FakeCategoryAbsurdAction 荒謬行動：看似合理但危險
	FakeCategoryAbsurdAction
	// FakeCategoryParanoidDefense 偏執防禦：過度反應
	FakeCategoryParanoidDefense
)

// FakeOption 代表一個假選項
type FakeOption struct {
	Text        string             // 選項文字
	Category    FakeOptionCategory // 類別
	IsFake      bool               // 標記為假選項（內部使用）
	SANDamage   int                // 選擇後的 SAN 損失
	HPDamage    int                // 選擇後的 HP 損失
	Description string             // 內部描述（不顯示給玩家）
}

// ShouldGenerateFakeOption 判斷是否應該生成假選項
// Story 9-6 AC1-AC2: 根據 SAN 值判斷機率
func (fg *FakeOptionGenerator) ShouldGenerateFakeOption(san int) bool {
	if san >= 30 {
		return false
	}

	if san >= 15 {
		// SAN 15-29: 30% 機率
		return fg.rng.Float64() < 0.30
	}

	// SAN < 15: 50% 機率
	return fg.rng.Float64() < 0.50
}

// GetFakeOptionCount 獲取應該生成的假選項數量
// Story 9-6 AC2: SAN < 15 時可能生成 2 個假選項
func (fg *FakeOptionGenerator) GetFakeOptionCount(san int) int {
	if san >= 30 {
		return 0
	}

	if !fg.ShouldGenerateFakeOption(san) {
		return 0
	}

	if san >= 15 {
		// SAN 15-29: 只生成 1 個
		return 1
	}

	// SAN < 15: 生成 1-2 個
	if fg.rng.Float64() < 0.5 {
		return 2
	}
	return 1
}

// GenerateFakeOption 生成一個假選項
// Story 9-6 AC3-AC5: 生成合理但詭異的假選項
func (fg *FakeOptionGenerator) GenerateFakeOption(context OptionContext, variant int) FakeOption {
	// 根據場景類型選擇假選項類別
	category := fg.selectCategory(context.SceneType, variant)

	// 生成假選項
	return fg.generateByCategoryAndContext(category, context, variant)
}

// OptionContext 提供生成假選項所需的上下文
type OptionContext struct {
	SceneType     SceneType // 場景類型
	StoryContext  string    // 故事情境
	TensionLevel  int       // 張力等級
	PlayerSAN     int       // 玩家當前 SAN
	ExistingOptions []string // 已有的真實選項（用於避免重複）
}

// selectCategory 根據場景類型和變體選擇假選項類別
func (fg *FakeOptionGenerator) selectCategory(sceneType SceneType, variant int) FakeOptionCategory {
	// 根據場景類型映射到適合的類別
	categoryMap := map[SceneType][]FakeOptionCategory{
		SceneExplore: {
			FakeCategoryFalseSafety,
			FakeCategoryInducedDanger,
			FakeCategoryAbsurdAction,
		},
		SceneDialogue: {
			FakeCategoryAbsurdAction,
			FakeCategoryParanoidDefense,
			FakeCategoryFalseSafety,
		},
		SceneCombat: {
			FakeCategoryParanoidDefense,
			FakeCategoryAbsurdAction,
			FakeCategoryInducedDanger,
		},
		SceneEscape: {
			FakeCategoryInducedDanger,
			FakeCategoryFalseSafety,
			FakeCategoryParanoidDefense,
		},
	}

	categories, ok := categoryMap[sceneType]
	if !ok || len(categories) == 0 {
		// 默認使用所有類別
		categories = []FakeOptionCategory{
			FakeCategoryFalseSafety,
			FakeCategoryInducedDanger,
			FakeCategoryAbsurdAction,
			FakeCategoryParanoidDefense,
		}
	}

	// 使用變體和隨機來選擇類別
	index := (variant + fg.rng.Intn(len(categories))) % len(categories)
	return categories[index]
}

// generateByCategoryAndContext 根據類別和上下文生成假選項
func (fg *FakeOptionGenerator) generateByCategoryAndContext(category FakeOptionCategory, context OptionContext, variant int) FakeOption {
	switch category {
	case FakeCategoryFalseSafety:
		return fg.generateFalseSafety(context, variant)
	case FakeCategoryInducedDanger:
		return fg.generateInducedDanger(context, variant)
	case FakeCategoryAbsurdAction:
		return fg.generateAbsurdAction(context, variant)
	case FakeCategoryParanoidDefense:
		return fg.generateParanoidDefense(context, variant)
	default:
		return fg.generateFalseSafety(context, variant)
	}
}

// generateFalseSafety 生成虛假安全選項
// Story 9-6 AC5: 看起來是安全的出路，實際是幻覺
func (fg *FakeOptionGenerator) generateFalseSafety(context OptionContext, variant int) FakeOption {
	templates := []struct {
		text string
		desc string
	}{
		{"看到出口離開", "幻覺中的虛假出口"},
		{"進入安全房間", "不存在的安全區域"},
		{"找到救援信號", "幻聽的呼救聲"},
		{"發現隱藏通道", "視覺欺騙的假通道"},
		{"看見光明處", "幻覺中的光源"},
		{"聽到熟悉聲音", "幻聽的親人呼喚"},
		{"找到休息點", "虛假的安全感"},
		{"看到標記路線", "幻覺製造的指引"},
	}

	selected := templates[variant%len(templates)]

	// Story 9-6 AC3: 選擇後造成 SAN 損失
	sanDamage := -(10 + fg.rng.Intn(11)) // -10 to -20
	hpDamage := 0

	// 低 SAN 時可能也造成 HP 損失（撞到不存在的門等）
	if context.PlayerSAN < 10 {
		hpDamage = -(5 + fg.rng.Intn(11)) // -5 to -15
	}

	return FakeOption{
		Text:        selected.text,
		Category:    FakeCategoryFalseSafety,
		IsFake:      true,
		SANDamage:   sanDamage,
		HPDamage:    hpDamage,
		Description: selected.desc,
	}
}

// generateInducedDanger 生成誘導危險選項
// Story 9-6 AC5: 跟隨幻覺聲音/影像進入危險
func (fg *FakeOptionGenerator) generateInducedDanger(context OptionContext, variant int) FakeOption {
	templates := []struct {
		text string
		desc string
	}{
		{"跟隨呼喊聲", "幻聽引導進入危險"},
		{"追逐人影", "視覺幻覺的誘導"},
		{"靠近發光物", "幻覺中的誘人光芒"},
		{"跟隨腳步聲", "幻聽的陷阱"},
		{"觸碰閃光點", "危險的幻覺吸引"},
		{"走向歌聲", "幻聽的美妙歌聲"},
		{"追蹤香味", "幻覺的嗅覺誘導"},
		{"跟隨影子", "扭曲的視覺引導"},
	}

	selected := templates[variant%len(templates)]

	// Story 9-6 AC3: 造成更大的 SAN 和 HP 損失
	sanDamage := -(15 + fg.rng.Intn(11)) // -15 to -25
	hpDamage := -(10 + fg.rng.Intn(16))  // -10 to -25

	return FakeOption{
		Text:        selected.text,
		Category:    FakeCategoryInducedDanger,
		IsFake:      true,
		SANDamage:   sanDamage,
		HPDamage:    hpDamage,
		Description: selected.desc,
	}
}

// generateAbsurdAction 生成荒謬行動選項
// Story 9-6 AC5: 看似合理但實際危險的行為
func (fg *FakeOptionGenerator) generateAbsurdAction(context OptionContext, variant int) FakeOption {
	templates := []struct {
		text string
		desc string
	}{
		{"服用地上藥片", "未知藥物的危險"},
		{"飲用瓶中液體", "不明液體"},
		{"打開神秘盒子", "潘朵拉之盒"},
		{"觸摸脈動物體", "危險的未知物"},
		{"戴上面具", "被詛咒的面具"},
		{"閱讀古書", "禁忌知識"},
		{"握住刀柄", "被污染的武器"},
		{"按下紅色按鈕", "未知機關"},
	}

	selected := templates[variant%len(templates)]

	// Story 9-6 AC3: 造成中等 SAN 損失和可能的 HP 損失
	sanDamage := -(12 + fg.rng.Intn(9))  // -12 to -20
	hpDamage := -(5 + fg.rng.Intn(16))   // -5 to -20

	return FakeOption{
		Text:        selected.text,
		Category:    FakeCategoryAbsurdAction,
		IsFake:      true,
		SANDamage:   sanDamage,
		HPDamage:    hpDamage,
		Description: selected.desc,
	}
}

// generateParanoidDefense 生成偏執防禦選項
// Story 9-6 AC5: 過度防禦反應
func (fg *FakeOptionGenerator) generateParanoidDefense(context OptionContext, variant int) FakeOption {
	templates := []struct {
		text string
		desc string
	}{
		{"攻擊可疑人", "攻擊無辜者"},
		{"砸碎鏡子", "破壞性衝動"},
		{"推倒櫃子", "過度反應的破壞"},
		{"撕毀文件", "破壞重要線索"},
		{"大聲尖叫", "引來真正的危險"},
		{"緊閉雙眼", "危險的逃避"},
		{"摀住耳朵", "忽視重要警告"},
		{"逃離現場", "錯誤的逃跑方向"},
	}

	selected := templates[variant%len(templates)]

	// Story 9-6 AC3: 造成中等 SAN 和 HP 損失
	sanDamage := -(10 + fg.rng.Intn(11)) // -10 to -20
	hpDamage := -(8 + fg.rng.Intn(13))   // -8 to -20

	return FakeOption{
		Text:        selected.text,
		Category:    FakeCategoryParanoidDefense,
		IsFake:      true,
		SANDamage:   sanDamage,
		HPDamage:    hpDamage,
		Description: selected.desc,
	}
}

// GenerateFakeOptions 生成多個假選項
// Story 9-6 AC2: 根據 SAN 生成 1-2 個假選項
func (fg *FakeOptionGenerator) GenerateFakeOptions(context OptionContext) []FakeOption {
	count := fg.GetFakeOptionCount(context.PlayerSAN)
	if count == 0 {
		return nil
	}

	options := make([]FakeOption, 0, count)
	for i := 0; i < count; i++ {
		option := fg.GenerateFakeOption(context, i)
		options = append(options, option)
	}

	return options
}

// ApplyFakeOptionConsequence 應用假選項的後果
// Story 9-6 AC4: 選擇假選項後的反應
func (fg *FakeOptionGenerator) ApplyFakeOptionConsequence(option FakeOption, currentHP, currentSAN int) (newHP, newSAN int, message string) {
	newHP = currentHP + option.HPDamage
	newSAN = currentSAN + option.SANDamage

	// 確保不低於 0
	if newHP < 0 {
		newHP = 0
	}
	if newSAN < 0 {
		newSAN = 0
	}

	// 生成反應訊息
	message = fg.generateConsequenceMessage(option)

	return newHP, newSAN, message
}

// generateConsequenceMessage 生成假選項後果訊息
// Story 9-6 AC4: 清楚告知玩家這是幻覺
func (fg *FakeOptionGenerator) generateConsequenceMessage(option FakeOption) string {
	messages := map[FakeOptionCategory][]string{
		FakeCategoryFalseSafety: {
			"你朝著出口奔去，卻撞上了冰冷的牆壁。那裡什麼都沒有...",
			"你以為找到了安全的地方，但那只是你的幻覺。",
			"光明消失了，你意識到那只是你絕望中的幻想。",
			"你追逐的希望像泡沫一樣破滅了。",
		},
		FakeCategoryInducedDanger: {
			"你跟隨著聲音前進，突然發現自己走入了更深的黑暗...",
			"那個身影消失了，留下你獨自面對真正的危險。",
			"你意識到那不是真實的，但為時已晚。",
			"幻覺引導你走向了陷阱。",
		},
		FakeCategoryAbsurdAction: {
			"你拿起了那個東西，一股劇痛襲來！",
			"那不是你想像的那樣...一陣噁心感湧上。",
			"你的判斷出現了嚴重錯誤。",
			"這個決定讓你付出了慘痛的代價。",
		},
		FakeCategoryParanoidDefense: {
			"你的過度反應造成了更糟的情況...",
			"你攻擊了不該攻擊的對象，事態變得更加混亂。",
			"你的恐慌讓情況失控了。",
			"這個衝動的決定帶來了嚴重後果。",
		},
	}

	categoryMessages, ok := messages[option.Category]
	if !ok || len(categoryMessages) == 0 {
		return "你的判斷出現了錯誤...這是幻覺造成的。"
	}

	message := categoryMessages[fg.rng.Intn(len(categoryMessages))]

	// 添加傷害提示
	damageMsg := fmt.Sprintf("\n\n[幻覺後果] HP %d, SAN %d",
		option.HPDamage, option.SANDamage)

	return message + damageMsg
}

// IsFakeOption 檢查選項是否為假選項
// 提供給外部系統判斷
func IsFakeOption(option interface{}) bool {
	if fakeOpt, ok := option.(FakeOption); ok {
		return fakeOpt.IsFake
	}
	return false
}

// ConvertToOption 將 FakeOption 轉換為標準 Option
// 用於整合到現有的選項系統
func ConvertToOption(fakeOpt FakeOption, index int) Option {
	return Option{
		Index:       index,
		Text:        fakeOpt.Text,
		Description: fakeOpt.Description, // 內部標記，不顯示給玩家
	}
}

// MergeFakeOptionsWithReal 將假選項混入真實選項
// Story 9-6 AC5: 假選項應該無縫混入真實選項中
func (fg *FakeOptionGenerator) MergeFakeOptionsWithReal(realOptions []Option, fakeOptions []FakeOption) []Option {
	if len(fakeOptions) == 0 {
		return realOptions
	}

	// 計算總選項數
	totalCount := len(realOptions) + len(fakeOptions)
	merged := make([]Option, 0, totalCount)

	// 隨機決定假選項的插入位置
	insertPositions := fg.generateInsertPositions(len(realOptions), len(fakeOptions))

	realIdx := 0
	fakeIdx := 0
	optionNumber := 1

	for i := 0; i < totalCount; i++ {
		if containsInt(insertPositions, i) && fakeIdx < len(fakeOptions) {
			// 插入假選項
			fakeOpt := ConvertToOption(fakeOptions[fakeIdx], optionNumber)
			merged = append(merged, fakeOpt)
			fakeIdx++
		} else if realIdx < len(realOptions) {
			// 插入真實選項
			realOpt := realOptions[realIdx]
			realOpt.Index = optionNumber
			merged = append(merged, realOpt)
			realIdx++
		}
		optionNumber++
	}

	return merged
}

// generateInsertPositions 生成假選項的插入位置
func (fg *FakeOptionGenerator) generateInsertPositions(realCount, fakeCount int) []int {
	totalCount := realCount + fakeCount
	positions := make([]int, 0, fakeCount)

	// 生成隨機但不重複的位置
	usedPositions := make(map[int]bool)

	for i := 0; i < fakeCount; i++ {
		for {
			pos := fg.rng.Intn(totalCount)
			if !usedPositions[pos] {
				positions = append(positions, pos)
				usedPositions[pos] = true
				break
			}
		}
	}

	return positions
}

// containsInt 檢查切片是否包含元素
func containsInt(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// GetFakeOptionByControlLevel 根據 ControlLevel 生成假選項
// 整合現有的 game.ControlLevel 系統
func (fg *FakeOptionGenerator) GetFakeOptionByControlLevel(context OptionContext, controlLevel game.ControlLevel) []FakeOption {
	// 使用 ControlLevel 的 SAN 值
	context.PlayerSAN = controlLevel.SAN
	return fg.GenerateFakeOptions(context)
}
