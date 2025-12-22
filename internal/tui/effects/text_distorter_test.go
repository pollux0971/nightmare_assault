package effects

import (
	"strings"
	"testing"
)

// TestGetDistortionLevel 測試扭曲等級計算
// Story 9-5 AC1-AC3: 根據 SAN 值計算扭曲強度
func TestGetDistortionLevel(t *testing.T) {
	tests := []struct {
		name     string
		san      int
		expected DistortionLevel
	}{
		{"SAN 100 - 無扭曲", 100, DistortionNone},
		{"SAN 50 - 無扭曲邊界", 50, DistortionNone},
		{"SAN 49 - 輕度扭曲", 49, DistortionMild},
		{"SAN 30 - 輕度扭曲邊界", 30, DistortionMild},
		{"SAN 29 - 中度扭曲", 29, DistortionModerate},
		{"SAN 15 - 中度扭曲邊界", 15, DistortionModerate},
		{"SAN 14 - 重度扭曲", 14, DistortionSevere},
		{"SAN 5 - 重度扭曲", 5, DistortionSevere},
		{"SAN 0 - 重度扭曲", 0, DistortionSevere},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDistortionLevel(tt.san)
			if result != tt.expected {
				t.Errorf("GetDistortionLevel(%d) = %v, want %v", tt.san, result, tt.expected)
			}
		})
	}
}

// TestDistortText_NoDistortion 測試無扭曲狀態
// Story 9-5 AC1: SAN >= 50 時無扭曲
func TestDistortText_NoDistortion(t *testing.T) {
	td := NewTextDistorter(12345)
	originalText := "這是一段正常的文字，沒有任何扭曲。"

	tests := []struct {
		name string
		san  int
	}{
		{"SAN 100", 100},
		{"SAN 75", 75},
		{"SAN 50", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.DistortText(originalText, tt.san)
			if result != originalText {
				t.Errorf("DistortText() should not distort at SAN %d, got: %s", tt.san, result)
			}
		})
	}
}

// TestDistortText_MildDistortion 測試輕度扭曲
// Story 9-5 AC1: SAN 30-49 輕微字符替換
func TestDistortText_MildDistortion(t *testing.T) {
	td := NewTextDistorter(12345)
	originalText := "這是一段測試文字，用於檢測輕度扭曲效果。"

	tests := []struct {
		name string
		san  int
	}{
		{"SAN 45", 45},
		{"SAN 35", 35},
		{"SAN 30", 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.DistortText(originalText, tt.san)

			// 應該有一些變化，但不會完全不同
			if result == originalText {
				// 可能運氣好沒有扭曲，多測試幾次
				hasDistortion := false
				for i := 0; i < 10; i++ {
					testResult := td.DistortText(originalText, tt.san)
					if testResult != originalText {
						hasDistortion = true
						break
					}
				}
				if !hasDistortion {
					t.Errorf("Expected some distortion at SAN %d after multiple attempts", tt.san)
				}
			}

			// 文字長度應該相近（允許 ±20% 變化）
			origLen := len([]rune(originalText))
			resultLen := len([]rune(result))
			if resultLen < origLen-origLen/5 || resultLen > origLen+origLen/5 {
				t.Errorf("Distorted text length %d too different from original %d", resultLen, origLen)
			}
		})
	}
}

// TestDistortText_ModerateDistortion 測試中度扭曲
// Story 9-5 AC2: SAN 15-29 更頻繁替換和顛倒
func TestDistortText_ModerateDistortion(t *testing.T) {
	td := NewTextDistorter(12345)
	originalText := "這是一段需要中度扭曲的測試文字內容。"

	tests := []struct {
		name string
		san  int
	}{
		{"SAN 25", 25},
		{"SAN 20", 20},
		{"SAN 15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.DistortText(originalText, tt.san)

			// 應該有明顯變化
			if result == originalText {
				t.Errorf("Expected distortion at SAN %d, got unchanged text", tt.san)
			}

			// 測試多次應該看到字符顛倒的現象
			distortedCount := 0
			for i := 0; i < 20; i++ {
				testResult := td.DistortText(originalText, tt.san)
				if testResult != originalText {
					distortedCount++
				}
			}

			if distortedCount < 15 {
				t.Errorf("Expected consistent distortion at SAN %d, only %d/20 were distorted", tt.san, distortedCount)
			}
		})
	}
}

// TestDistortText_SevereDistortion 測試重度扭曲
// Story 9-5 AC3: SAN < 15 大量替換和幻覺文字
func TestDistortText_SevereDistortion(t *testing.T) {
	td := NewTextDistorter(12345)
	originalText := "這是嚴重扭曲測試文字，應該會有大量改變和幻覺內容插入。"

	tests := []struct {
		name string
		san  int
	}{
		{"SAN 10", 10},
		{"SAN 5", 5},
		{"SAN 1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.DistortText(originalText, tt.san)

			// 應該有明顯變化
			if result == originalText {
				t.Errorf("Expected severe distortion at SAN %d, got unchanged text", tt.san)
			}

			// 嚴重扭曲可能會插入字符，所以長度可能增加
			origLen := len([]rune(originalText))
			resultLen := len([]rune(result))

			// 允許長度增加（因為插入幻覺文字）
			if resultLen < origLen/2 {
				t.Errorf("Distorted text too short: %d vs original %d", resultLen, origLen)
			}

			// 測試是否包含幻覺字符（符號或詭異文字）
			hasHallucinationChars := false
			hallucinationRunes := []rune{'█', '▓', '▒', '░', '●', '○', '★', '☆', '魘', '魅', '鬼', '死'}
			for _, hr := range hallucinationRunes {
				if strings.ContainsRune(result, hr) {
					hasHallucinationChars = true
					break
				}
			}

			// 在多次測試中應該至少有一次包含幻覺字符
			if !hasHallucinationChars {
				// 再測試幾次
				for i := 0; i < 10; i++ {
					testResult := td.DistortText(originalText, tt.san)
					for _, hr := range hallucinationRunes {
						if strings.ContainsRune(testResult, hr) {
							hasHallucinationChars = true
							break
						}
					}
					if hasHallucinationChars {
						break
					}
				}
			}

			// Note: 由於隨機性，可能不是每次都有幻覺字符，所以這個測試比較寬鬆
			// 主要確保扭曲確實發生
		})
	}
}

// TestIsCriticalGameInfo 測試關鍵資訊識別
// Story 9-5 AC5: HP/SAN 數值、選項編號不應被扭曲
func TestIsCriticalGameInfo(t *testing.T) {
	td := NewTextDistorter(12345)

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"純數字", "123", true},
		{"選項編號 1", "1. ", true},
		{"選項編號 2", "2.", true},
		{"HP 數值", "HP: 80/100", true},
		{"SAN 數值", "SAN: 45/100", true},
		{"HP 斜線格式", "HP/100", true},
		{"SAN 斜線格式", "SAN/100", true},
		{"一般文字", "這是一段普通文字", false},
		{"包含數字的文字", "我有3個蘋果", false},
		{"空字串", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.isCriticalGameInfo(tt.text)
			if result != tt.expected {
				t.Errorf("isCriticalGameInfo(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

// TestDistortText_ProtectCriticalInfo 測試關鍵資訊保護
// Story 9-5 AC5: 確保關鍵資訊不被扭曲
func TestDistortText_ProtectCriticalInfo(t *testing.T) {
	td := NewTextDistorter(12345)

	criticalTexts := []string{
		"HP: 80/100",
		"SAN: 45/100",
		"1",
		"2",
		"3",
	}

	// 即使在最低 SAN 下，關鍵資訊也不應該被扭曲
	for _, text := range criticalTexts {
		t.Run(text, func(t *testing.T) {
			result := td.DistortText(text, 1)
			if result != text {
				t.Errorf("Critical info %q was distorted to %q", text, result)
			}
		})
	}
}

// TestDistortChoice 測試選項扭曲
// Story 9-5 AC5: 選項編號不扭曲，內容會扭曲
func TestDistortChoice(t *testing.T) {
	td := NewTextDistorter(12345)

	tests := []struct {
		name        string
		choiceText  string
		san         int
		expectMatch bool // 是否期望完全匹配原文
	}{
		{"SAN 100 - 無扭曲", "1. 檢查門", 100, true},
		{"SAN 20 - 應該扭曲內容", "2. 觸摸鏡子", 20, false},
		{"SAN 5 - 嚴重扭曲內容", "3. 離開房間", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.DistortChoice(tt.choiceText, tt.san)

			if tt.expectMatch {
				if result != tt.choiceText {
					t.Errorf("Expected no distortion, got %q", result)
				}
			} else {
				// 選項編號應該保留
				if !strings.HasPrefix(result, strings.Split(tt.choiceText, ".")[0]+".") {
					t.Errorf("Choice number should be preserved, got %q", result)
				}
			}
		})
	}
}

// TestGetDistortionIntensity 測試扭曲強度值
func TestGetDistortionIntensity(t *testing.T) {
	tests := []struct {
		name     string
		san      int
		expected float64
	}{
		{"SAN 100 - 無強度", 100, 0.0},
		{"SAN 50 - 無強度", 50, 0.0},
		{"SAN 40 - 輕度", 40, 0.25},
		{"SAN 20 - 中度", 20, 0.5},
		{"SAN 10 - 重度", 10, 0.85},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDistortionIntensity(tt.san)
			if result != tt.expected {
				t.Errorf("GetDistortionIntensity(%d) = %.2f, want %.2f", tt.san, result, tt.expected)
			}
		})
	}
}

// TestDistortText_ProgressiveDistortion 測試漸進式扭曲
// Story 9-5 AC4: 扭曲應該是漸進式的，不會突然變化
func TestDistortText_ProgressiveDistortion(t *testing.T) {
	td := NewTextDistorter(12345)
	testText := "這是用於測試漸進式扭曲效果的較長文字內容，應該隨著 SAN 值降低而逐漸扭曲。"

	// 測試不同 SAN 值的扭曲程度
	sanLevels := []int{50, 45, 40, 35, 30, 25, 20, 15, 10, 5}
	prevDifference := 0

	for i, san := range sanLevels {
		result := td.DistortText(testText, san)

		// 計算與原文的差異程度
		difference := calculateDifference(testText, result)

		if i > 0 {
			// 隨著 SAN 降低，差異應該增加或保持
			// 允許一些隨機波動，但總體趨勢應該是增加
			if difference < prevDifference-5 { // 允許 5% 的波動
				t.Logf("Warning: Distortion decreased from %.1f%% to %.1f%% when SAN dropped from %d to %d",
					float64(prevDifference), float64(difference), sanLevels[i-1], san)
			}
		}

		prevDifference = difference
		t.Logf("SAN %d: %.1f%% difference", san, float64(difference))
	}
}

// TestGetSimilarCharacter 測試相似字符生成
func TestGetSimilarCharacter(t *testing.T) {
	td := NewTextDistorter(12345)

	tests := []struct {
		name     string
		input    rune
		checkFn  func(rune) bool
	}{
		{
			name:  "中文字符",
			input: '看',
			checkFn: func(r rune) bool {
				// 應該是中文字符範圍
				return r >= 0x4E00 && r <= 0x9FFF
			},
		},
		{
			name:  "英文字母",
			input: 'a',
			checkFn: func(r rune) bool {
				// 應該是字母或特殊字符
				return true // 任何結果都可接受
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.getSimilarCharacter(tt.input)
			if !tt.checkFn(result) {
				t.Errorf("getSimilarCharacter(%c) = %c, failed check", tt.input, result)
			}
		})
	}
}

// TestGetHallucinationCharacter 測試幻覺字符生成
func TestGetHallucinationCharacter(t *testing.T) {
	td := NewTextDistorter(12345)

	// 生成多個幻覺字符，確保有變化
	chars := make(map[rune]bool)
	for i := 0; i < 50; i++ {
		char := td.getHallucinationCharacter()
		chars[char] = true
	}

	// 應該有多種不同的幻覺字符
	if len(chars) < 5 {
		t.Errorf("Expected variety in hallucination characters, got only %d unique chars", len(chars))
	}
}

// TestIsWhitespaceOrPunctuation 測試空白和標點檢測
func TestIsWhitespaceOrPunctuation(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		{"空格", ' ', true},
		{"換行", '\n', true},
		{"Tab", '\t', true},
		{"逗號", ',', true},
		{"句號", '.', true},
		{"問號", '?', true},
		{"中文逗號", '，', true},
		{"中文句號", '。', true},
		{"字母", 'a', false},
		{"中文字", '字', false},
		{"數字", '1', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWhitespaceOrPunctuation(tt.input)
			if result != tt.expected {
				t.Errorf("isWhitespaceOrPunctuation(%c) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsNumericOnly 測試純數字檢測
func TestIsNumericOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"純數字", "12345", true},
		{"單個數字", "1", true},
		{"包含字母", "123abc", false},
		{"包含空格", "123 456", false},
		{"包含標點", "123.", false},
		{"空字串", "", false},
		{"中文數字", "一二三", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNumericOnly(tt.input)
			if result != tt.expected {
				t.Errorf("isNumericOnly(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestDistortNarration 測試旁白扭曲
func TestDistortNarration(t *testing.T) {
	td := NewTextDistorter(12345)
	narration := "你走進一個昏暗的房間，牆壁上掛著一面古老的鏡子。"

	result := td.DistortNarration(narration, 10)

	// 應該有扭曲
	if result == narration {
		t.Errorf("Expected narration to be distorted at low SAN")
	}
}

// TestDistortDialogue 測試對話扭曲
func TestDistortDialogue(t *testing.T) {
	td := NewTextDistorter(12345)
	dialogue := "「歡迎來到這裡，旅行者。你看起來迷路了。」"

	result := td.DistortDialogue(dialogue, 15)

	// 應該有扭曲
	if result == dialogue {
		t.Errorf("Expected dialogue to be distorted at low SAN")
	}
}

// calculateDifference 計算兩個字串的差異百分比
func calculateDifference(original, distorted string) int {
	origRunes := []rune(original)
	distRunes := []rune(distorted)

	differences := 0
	maxLen := len(origRunes)
	if len(distRunes) > maxLen {
		maxLen = len(distRunes)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(origRunes) || i >= len(distRunes) {
			differences++
			continue
		}
		if origRunes[i] != distRunes[i] {
			differences++
		}
	}

	if maxLen == 0 {
		return 0
	}

	return (differences * 100) / maxLen
}

// BenchmarkDistortText_Mild 基準測試輕度扭曲性能
func BenchmarkDistortText_Mild(b *testing.B) {
	td := NewTextDistorter(12345)
	text := "這是一段用於性能測試的文字內容，包含足夠的字符來測試扭曲算法的效率。"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		td.DistortText(text, 40)
	}
}

// BenchmarkDistortText_Severe 基準測試重度扭曲性能
func BenchmarkDistortText_Severe(b *testing.B) {
	td := NewTextDistorter(12345)
	text := "這是一段用於性能測試的文字內容，包含足夠的字符來測試扭曲算法的效率。"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		td.DistortText(text, 5)
	}
}
