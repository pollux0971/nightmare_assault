package agents

import (
	"strings"
	"testing"
)

// TestShouldGenerateFakeOption 測試假選項生成判斷
// Story 9-6 AC1-AC2: 根據 SAN 值決定是否生成假選項
func TestShouldGenerateFakeOption(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	tests := []struct {
		name         string
		san          int
		expectNever  bool
		expectAlways bool
	}{
		{"SAN 100 - 永不生成", 100, true, false},
		{"SAN 50 - 永不生成", 50, true, false},
		{"SAN 30 - 永不生成邊界", 30, true, false},
		{"SAN 29 - 可能生成", 29, false, false},
		{"SAN 20 - 可能生成", 20, false, false},
		{"SAN 15 - 可能生成邊界", 15, false, false},
		{"SAN 14 - 更高機率生成", 14, false, false},
		{"SAN 5 - 更高機率生成", 5, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trueCount := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				if fg.ShouldGenerateFakeOption(tt.san) {
					trueCount++
				}
			}

			if tt.expectNever && trueCount > 0 {
				t.Errorf("Expected never to generate at SAN %d, but got %d/%d", tt.san, trueCount, iterations)
			}

			if tt.expectAlways && trueCount < iterations {
				t.Errorf("Expected always to generate at SAN %d, but got %d/%d", tt.san, trueCount, iterations)
			}

			// 檢查機率範圍
			if !tt.expectNever && !tt.expectAlways {
				probability := float64(trueCount) / float64(iterations)
				t.Logf("SAN %d: %.1f%% generation rate", tt.san, probability*100)

				// Story 9-6 AC1: SAN 15-29 應該約 30%
				if tt.san >= 15 && tt.san < 30 {
					if probability < 0.15 || probability > 0.45 {
						t.Errorf("Expected ~30%% at SAN %d, got %.1f%%", tt.san, probability*100)
					}
				}

				// Story 9-6 AC2: SAN < 15 應該約 50%
				if tt.san < 15 {
					if probability < 0.35 || probability > 0.65 {
						t.Errorf("Expected ~50%% at SAN %d, got %.1f%%", tt.san, probability*100)
					}
				}
			}
		})
	}
}

// TestGetFakeOptionCount 測試假選項數量
// Story 9-6 AC2: SAN < 15 時可能生成 2 個假選項
func TestGetFakeOptionCount(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	tests := []struct {
		name    string
		san     int
		wantMin int
		wantMax int
	}{
		{"SAN 100 - 無假選項", 100, 0, 0},
		{"SAN 30 - 無假選項邊界", 30, 0, 0},
		{"SAN 25 - 0-1 個假選項", 25, 0, 1},
		{"SAN 15 - 0-1 個假選項邊界", 15, 0, 1},
		{"SAN 10 - 0-2 個假選項", 10, 0, 2},
		{"SAN 5 - 0-2 個假選項", 5, 0, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts := make(map[int]int)
			iterations := 100

			for i := 0; i < iterations; i++ {
				count := fg.GetFakeOptionCount(tt.san)
				counts[count]++

				if count < tt.wantMin || count > tt.wantMax {
					t.Errorf("GetFakeOptionCount(%d) = %d, want range [%d, %d]",
						tt.san, count, tt.wantMin, tt.wantMax)
				}
			}

			t.Logf("SAN %d: count distribution = %v", tt.san, counts)

			// Story 9-6 AC2: SAN < 15 時應該有機會生成 2 個
			if tt.san < 15 && tt.wantMax == 2 {
				if _, hasTwo := counts[2]; !hasTwo {
					// 再測試一次確認
					foundTwo := false
					for i := 0; i < 50; i++ {
						if fg.GetFakeOptionCount(tt.san) == 2 {
							foundTwo = true
							break
						}
					}
					if !foundTwo {
						t.Errorf("Expected occasional count of 2 at SAN %d", tt.san)
					}
				}
			}
		})
	}
}

// TestGenerateFakeOption 測試假選項生成
// Story 9-6 AC3-AC5: 生成合理但詭異的假選項
func TestGenerateFakeOption(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneExplore,
		StoryContext: "你在一個黑暗的走廊中",
		TensionLevel: 70,
		PlayerSAN:    20,
	}

	tests := []struct {
		name    string
		variant int
	}{
		{"變體 0", 0},
		{"變體 1", 1},
		{"變體 2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := fg.GenerateFakeOption(context, tt.variant)

			// Story 9-6 AC4: 應該標記為假選項
			if !option.IsFake {
				t.Errorf("Expected option to be marked as fake")
			}

			// Story 9-6 AC5: 應該有文字
			if option.Text == "" {
				t.Errorf("Expected option to have text")
			}

			// Story 9-6 AC3: 應該有 SAN 損失
			if option.SANDamage >= 0 {
				t.Errorf("Expected negative SAN damage, got %d", option.SANDamage)
			}

			// 檢查文字長度（應該合理，不會太長）
			if len([]rune(option.Text)) > 20 {
				t.Errorf("Option text too long: %d runes", len([]rune(option.Text)))
			}

			// 應該有描述（內部使用）
			if option.Description == "" {
				t.Errorf("Expected option to have description")
			}

			t.Logf("Generated fake option: %s (SAN: %d, HP: %d, Category: %v)",
				option.Text, option.SANDamage, option.HPDamage, option.Category)
		})
	}
}

// TestGenerateFakeOption_DifferentCategories 測試不同類別的假選項
func TestGenerateFakeOption_DifferentCategories(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	sceneTypes := []SceneType{
		SceneExplore,
		SceneDialogue,
		SceneCombat,
		SceneEscape,
	}

	for _, sceneType := range sceneTypes {
		t.Run(sceneType.String(), func(t *testing.T) {
			context := OptionContext{
				SceneType:    sceneType,
				StoryContext: "測試場景",
				TensionLevel: 50,
				PlayerSAN:    15,
			}

			// 生成多個假選項，確保有變化
			categories := make(map[FakeOptionCategory]bool)
			for i := 0; i < 10; i++ {
				option := fg.GenerateFakeOption(context, i)
				categories[option.Category] = true
			}

			// 應該有多種類別
			if len(categories) < 2 {
				t.Logf("Warning: Only %d categories generated for %s", len(categories), sceneType)
			}
		})
	}
}

// TestGenerateFalseSafety 測試虛假安全選項
func TestGenerateFalseSafety(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneExplore,
		StoryContext: "測試",
		TensionLevel: 50,
		PlayerSAN:    20,
	}

	option := fg.generateFalseSafety(context, 0)

	if option.Category != FakeCategoryFalseSafety {
		t.Errorf("Expected FalseSafety category, got %v", option.Category)
	}

	if option.SANDamage >= -9 {
		t.Errorf("Expected at least -10 SAN damage, got %d", option.SANDamage)
	}
}

// TestGenerateInducedDanger 測試誘導危險選項
func TestGenerateInducedDanger(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneExplore,
		StoryContext: "測試",
		TensionLevel: 70,
		PlayerSAN:    10,
	}

	option := fg.generateInducedDanger(context, 0)

	if option.Category != FakeCategoryInducedDanger {
		t.Errorf("Expected InducedDanger category, got %v", option.Category)
	}

	// Story 9-6 AC3: 誘導危險應該造成更大傷害
	if option.SANDamage >= -14 {
		t.Errorf("Expected at least -15 SAN damage, got %d", option.SANDamage)
	}

	if option.HPDamage >= -9 {
		t.Errorf("Expected at least -10 HP damage, got %d", option.HPDamage)
	}
}

// TestGenerateAbsurdAction 測試荒謬行動選項
func TestGenerateAbsurdAction(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneExplore,
		StoryContext: "測試",
		TensionLevel: 60,
		PlayerSAN:    15,
	}

	option := fg.generateAbsurdAction(context, 0)

	if option.Category != FakeCategoryAbsurdAction {
		t.Errorf("Expected AbsurdAction category, got %v", option.Category)
	}

	if option.SANDamage >= 0 || option.HPDamage >= 0 {
		t.Errorf("Expected negative damage for absurd action")
	}
}

// TestGenerateParanoidDefense 測試偏執防禦選項
func TestGenerateParanoidDefense(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneCombat,
		StoryContext: "測試",
		TensionLevel: 80,
		PlayerSAN:    12,
	}

	option := fg.generateParanoidDefense(context, 0)

	if option.Category != FakeCategoryParanoidDefense {
		t.Errorf("Expected ParanoidDefense category, got %v", option.Category)
	}

	if option.SANDamage >= 0 || option.HPDamage >= 0 {
		t.Errorf("Expected negative damage for paranoid defense")
	}
}

// TestGenerateFakeOptions 測試批量生成假選項
// Story 9-6 AC2: 根據 SAN 生成適量的假選項
func TestGenerateFakeOptions(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	tests := []struct {
		name        string
		san         int
		expectCount int // -1 表示變動的
	}{
		{"SAN 50 - 無假選項", 50, 0},
		{"SAN 25 - 可能 0-1 個", 25, -1},
		{"SAN 10 - 可能 0-2 個", 10, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := OptionContext{
				SceneType:    SceneExplore,
				StoryContext: "測試場景",
				TensionLevel: 50,
				PlayerSAN:    tt.san,
			}

			options := fg.GenerateFakeOptions(context)

			if tt.expectCount >= 0 {
				if len(options) != tt.expectCount {
					t.Errorf("Expected %d fake options, got %d", tt.expectCount, len(options))
				}
			} else {
				// 變動的情況，測試多次
				counts := make(map[int]int)
				for i := 0; i < 50; i++ {
					opts := fg.GenerateFakeOptions(context)
					counts[len(opts)]++
				}
				t.Logf("SAN %d: fake option count distribution = %v", tt.san, counts)
			}

			// 檢查所有生成的選項
			for i, opt := range options {
				if !opt.IsFake {
					t.Errorf("Option %d should be marked as fake", i)
				}
			}
		})
	}
}

// TestApplyFakeOptionConsequence 測試假選項後果
// Story 9-6 AC4: 選擇假選項後觸發幻覺反應
func TestApplyFakeOptionConsequence(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	fakeOption := FakeOption{
		Text:        "測試假選項",
		Category:    FakeCategoryFalseSafety,
		IsFake:      true,
		SANDamage:   -15,
		HPDamage:    -10,
		Description: "測試描述",
	}

	currentHP := 80
	currentSAN := 60

	newHP, newSAN, message := fg.ApplyFakeOptionConsequence(fakeOption, currentHP, currentSAN)

	// 檢查數值變化
	expectedHP := currentHP + fakeOption.HPDamage
	expectedSAN := currentSAN + fakeOption.SANDamage

	if newHP != expectedHP {
		t.Errorf("Expected HP %d, got %d", expectedHP, newHP)
	}

	if newSAN != expectedSAN {
		t.Errorf("Expected SAN %d, got %d", expectedSAN, newSAN)
	}

	// 應該有訊息
	if message == "" {
		t.Errorf("Expected consequence message")
	}

	// 訊息應該包含傷害資訊
	if !strings.Contains(message, "HP") || !strings.Contains(message, "SAN") {
		t.Errorf("Message should contain damage info: %s", message)
	}

	t.Logf("Consequence message: %s", message)
}

// TestApplyFakeOptionConsequence_MinimumValues 測試最小值限制
func TestApplyFakeOptionConsequence_MinimumValues(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	fakeOption := FakeOption{
		SANDamage: -50,
		HPDamage:  -60,
		Category:  FakeCategoryInducedDanger,
		IsFake:    true,
	}

	currentHP := 30
	currentSAN := 20

	newHP, newSAN, _ := fg.ApplyFakeOptionConsequence(fakeOption, currentHP, currentSAN)

	// 應該不低於 0
	if newHP < 0 {
		t.Errorf("HP should not be negative, got %d", newHP)
	}

	if newSAN < 0 {
		t.Errorf("SAN should not be negative, got %d", newSAN)
	}
}

// TestGenerateConsequenceMessage 測試後果訊息生成
func TestGenerateConsequenceMessage(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	categories := []FakeOptionCategory{
		FakeCategoryFalseSafety,
		FakeCategoryInducedDanger,
		FakeCategoryAbsurdAction,
		FakeCategoryParanoidDefense,
	}

	for _, category := range categories {
		t.Run(string(rune('0'+category)), func(t *testing.T) {
			option := FakeOption{
				Category:  category,
				SANDamage: -15,
				HPDamage:  -10,
			}

			message := fg.generateConsequenceMessage(option)

			if message == "" {
				t.Errorf("Expected non-empty message for category %v", category)
			}

			// 訊息應該包含傷害資訊
			if !strings.Contains(message, "HP") || !strings.Contains(message, "SAN") {
				t.Errorf("Message should contain damage info")
			}

			t.Logf("Message for %v: %s", category, message)
		})
	}
}

// TestConvertToOption 測試轉換為標準選項
func TestConvertToOption(t *testing.T) {
	fakeOpt := FakeOption{
		Text:        "測試假選項",
		Category:    FakeCategoryFalseSafety,
		IsFake:      true,
		SANDamage:   -15,
		HPDamage:    -10,
		Description: "內部描述",
	}

	option := ConvertToOption(fakeOpt, 3)

	if option.Index != 3 {
		t.Errorf("Expected index 3, got %d", option.Index)
	}

	if option.Text != fakeOpt.Text {
		t.Errorf("Expected text %q, got %q", fakeOpt.Text, option.Text)
	}

	if option.Description != fakeOpt.Description {
		t.Errorf("Expected description %q, got %q", fakeOpt.Description, option.Description)
	}
}

// TestMergeFakeOptionsWithReal 測試混合假選項與真實選項
// Story 9-6 AC5: 假選項應該無縫混入真實選項
func TestMergeFakeOptionsWithReal(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	realOptions := []Option{
		{Index: 1, Text: "真實選項 1"},
		{Index: 2, Text: "真實選項 2"},
		{Index: 3, Text: "真實選項 3"},
	}

	fakeOptions := []FakeOption{
		{Text: "假選項 1", IsFake: true, Category: FakeCategoryFalseSafety},
		{Text: "假選項 2", IsFake: true, Category: FakeCategoryInducedDanger},
	}

	merged := fg.MergeFakeOptionsWithReal(realOptions, fakeOptions)

	// 總數應該正確
	expectedCount := len(realOptions) + len(fakeOptions)
	if len(merged) != expectedCount {
		t.Errorf("Expected %d total options, got %d", expectedCount, len(merged))
	}

	// 檢查索引連續性
	for i, opt := range merged {
		expectedIndex := i + 1
		if opt.Index != expectedIndex {
			t.Errorf("Option %d has index %d, want %d", i, opt.Index, expectedIndex)
		}
	}

	// 應該包含所有真實選項的文字
	realTexts := make(map[string]bool)
	for _, opt := range realOptions {
		realTexts[opt.Text] = false
	}

	for _, opt := range merged {
		if _, exists := realTexts[opt.Text]; exists {
			realTexts[opt.Text] = true
		}
	}

	for text, found := range realTexts {
		if !found {
			t.Errorf("Real option text %q not found in merged options", text)
		}
	}

	t.Logf("Merged options:")
	for i, opt := range merged {
		t.Logf("  %d. %s", i+1, opt.Text)
	}
}

// TestMergeFakeOptionsWithReal_NoFakeOptions 測試無假選項時的合併
func TestMergeFakeOptionsWithReal_NoFakeOptions(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	realOptions := []Option{
		{Index: 1, Text: "選項 1"},
		{Index: 2, Text: "選項 2"},
	}

	merged := fg.MergeFakeOptionsWithReal(realOptions, nil)

	if len(merged) != len(realOptions) {
		t.Errorf("Expected %d options, got %d", len(realOptions), len(merged))
	}

	// 應該保持原樣
	for i, opt := range merged {
		if opt.Text != realOptions[i].Text {
			t.Errorf("Option %d text mismatch", i)
		}
	}
}

// TestGenerateInsertPositions 測試插入位置生成
func TestGenerateInsertPositions(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	realCount := 3
	fakeCount := 2

	positions := fg.generateInsertPositions(realCount, fakeCount)

	// 應該生成正確數量的位置
	if len(positions) != fakeCount {
		t.Errorf("Expected %d positions, got %d", fakeCount, len(positions))
	}

	// 所有位置應該在有效範圍內
	totalCount := realCount + fakeCount
	for _, pos := range positions {
		if pos < 0 || pos >= totalCount {
			t.Errorf("Position %d out of range [0, %d)", pos, totalCount)
		}
	}

	// 不應該有重複位置
	uniquePositions := make(map[int]bool)
	for _, pos := range positions {
		if uniquePositions[pos] {
			t.Errorf("Duplicate position %d", pos)
		}
		uniquePositions[pos] = true
	}
}

// TestFakeOptionTextQuality 測試假選項文字品質
// Story 9-6 AC5: 假選項文字應該略微詭異但不明顯
func TestFakeOptionTextQuality(t *testing.T) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneExplore,
		StoryContext: "測試場景",
		TensionLevel: 60,
		PlayerSAN:    15,
	}

	// 生成多個假選項並檢查文字品質
	for i := 0; i < 20; i++ {
		option := fg.GenerateFakeOption(context, i)

		// 文字不應該太長
		if len([]rune(option.Text)) > 15 {
			t.Errorf("Fake option text too long: %q (%d runes)",
				option.Text, len([]rune(option.Text)))
		}

		// 文字不應該太短
		if len([]rune(option.Text)) < 3 {
			t.Errorf("Fake option text too short: %q", option.Text)
		}

		// 不應該包含明顯的「假」、「幻覺」等關鍵字
		if strings.Contains(option.Text, "假") ||
		   strings.Contains(option.Text, "幻覺") ||
		   strings.Contains(option.Text, "錯誤") {
			t.Errorf("Fake option text too obvious: %q", option.Text)
		}
	}
}

// BenchmarkGenerateFakeOption 基準測試假選項生成性能
func BenchmarkGenerateFakeOption(b *testing.B) {
	fg := NewFakeOptionGenerator(12345)

	context := OptionContext{
		SceneType:    SceneExplore,
		StoryContext: "測試場景",
		TensionLevel: 60,
		PlayerSAN:    15,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fg.GenerateFakeOption(context, i%10)
	}
}

// BenchmarkMergeFakeOptionsWithReal 基準測試選項合併性能
func BenchmarkMergeFakeOptionsWithReal(b *testing.B) {
	fg := NewFakeOptionGenerator(12345)

	realOptions := []Option{
		{Index: 1, Text: "選項 1"},
		{Index: 2, Text: "選項 2"},
		{Index: 3, Text: "選項 3"},
	}

	fakeOptions := []FakeOption{
		{Text: "假選項 1", IsFake: true},
		{Text: "假選項 2", IsFake: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fg.MergeFakeOptionsWithReal(realOptions, fakeOptions)
	}
}
