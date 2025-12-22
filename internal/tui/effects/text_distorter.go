package effects

import (
	"math/rand"
	"strings"
	"unicode"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// TextDistorter 處理基於 SAN 值的文字扭曲效果
// Story 9-5: 理智幻覺 - 文字扭曲系統
//
// 設計原則：
//   - SAN < 50: 輕微文字扭曲（偶爾替換字符）
//   - SAN < 30: 中度扭曲（更頻繁的替換、顛倒文字）
//   - SAN < 15: 嚴重扭曲（大量替換、插入幻覺文字）
//   - 關鍵遊戲資訊（HP/SAN 數值、選項編號）不會被扭曲
//   - 扭曲效果是漸進式的，不會突然變化
type TextDistorter struct {
	rng *rand.Rand
}

// NewTextDistorter 創建新的文字扭曲器
func NewTextDistorter(seed int64) *TextDistorter {
	return &TextDistorter{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// DistortionLevel 定義文字扭曲的強度等級
type DistortionLevel int

const (
	// DistortionNone 無扭曲 (SAN >= 50)
	DistortionNone DistortionLevel = iota
	// DistortionMild 輕度扭曲 (SAN 30-49)
	DistortionMild
	// DistortionModerate 中度扭曲 (SAN 15-29)
	DistortionModerate
	// DistortionSevere 重度扭曲 (SAN < 15)
	DistortionSevere
)

// GetDistortionLevel 根據 SAN 值計算扭曲等級
// Story 9-5 AC1-AC3: SAN 值對應扭曲強度
func GetDistortionLevel(san int) DistortionLevel {
	switch {
	case san >= 50:
		return DistortionNone
	case san >= 30:
		return DistortionMild
	case san >= 15:
		return DistortionModerate
	default:
		return DistortionSevere
	}
}

// DistortText 對文字應用扭曲效果
// Story 9-5 AC1-AC4: 漸進式文字扭曲
// Story 9-5 AC5: 保護關鍵遊戲資訊
func (td *TextDistorter) DistortText(text string, san int) string {
	level := GetDistortionLevel(san)

	if level == DistortionNone {
		return text
	}

	// Story 9-5 AC5: 保護關鍵資訊（數字、選項編號）
	if td.isCriticalGameInfo(text) {
		return text
	}

	// 根據扭曲等級應用不同的扭曲策略
	switch level {
	case DistortionMild:
		return td.applyMildDistortion(text)
	case DistortionModerate:
		return td.applyModerateDistortion(text)
	case DistortionSevere:
		return td.applySevereDistortion(text)
	default:
		return text
	}
}

// isCriticalGameInfo 檢查文字是否為關鍵遊戲資訊
// Story 9-5 AC5: HP/SAN 數值、選項編號不應被扭曲
func (td *TextDistorter) isCriticalGameInfo(text string) bool {
	// 檢查是否包含純數字
	trimmed := strings.TrimSpace(text)

	// 純數字文字（如選項編號）
	if isNumericOnly(trimmed) {
		return true
	}

	// HP/SAN 數值格式
	if strings.Contains(text, "HP:") ||
	   strings.Contains(text, "SAN:") ||
	   strings.Contains(text, "HP/") ||
	   strings.Contains(text, "SAN/") {
		return true
	}

	// 選項編號格式 (如 "1. ", "2. ")
	if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
		if len(trimmed) >= 2 && (trimmed[1] == '.' || trimmed[1] == ')') {
			return true
		}
	}

	return false
}

// applyMildDistortion 應用輕度扭曲 (SAN 30-49)
// Story 9-5 AC1: 偶爾替換字符
func (td *TextDistorter) applyMildDistortion(text string) string {
	runes := []rune(text)
	result := make([]rune, len(runes))

	for i, r := range runes {
		// 5-10% 機率替換字符
		if !isWhitespaceOrPunctuation(r) && td.rng.Float64() < 0.08 {
			result[i] = td.getSimilarCharacter(r)
		} else {
			result[i] = r
		}
	}

	return string(result)
}

// applyModerateDistortion 應用中度扭曲 (SAN 15-29)
// Story 9-5 AC2: 更頻繁的替換、顛倒文字
func (td *TextDistorter) applyModerateDistortion(text string) string {
	runes := []rune(text)
	result := make([]rune, 0, len(runes))

	i := 0
	for i < len(runes) {
		r := runes[i]

		// 跳過空白和標點
		if isWhitespaceOrPunctuation(r) {
			result = append(result, r)
			i++
			continue
		}

		// 15-20% 機率進行字符替換
		if td.rng.Float64() < 0.18 {
			result = append(result, td.getSimilarCharacter(r))
			i++
			continue
		}

		// 10% 機率顛倒相鄰字符
		if i+1 < len(runes) && !isWhitespaceOrPunctuation(runes[i+1]) && td.rng.Float64() < 0.10 {
			result = append(result, runes[i+1])
			result = append(result, runes[i])
			i += 2
			continue
		}

		result = append(result, r)
		i++
	}

	return string(result)
}

// applySevereDistortion 應用重度扭曲 (SAN < 15)
// Story 9-5 AC3: 大量替換、插入幻覺文字
func (td *TextDistorter) applySevereDistortion(text string) string {
	runes := []rune(text)
	result := make([]rune, 0, len(runes)*2) // 預留更多空間用於插入

	i := 0
	for i < len(runes) {
		r := runes[i]

		// 跳過空白和標點
		if isWhitespaceOrPunctuation(r) {
			result = append(result, r)
			i++
			continue
		}

		// 30-40% 機率進行字符替換
		if td.rng.Float64() < 0.35 {
			result = append(result, td.getSimilarCharacter(r))
			i++
			continue
		}

		// 15% 機率插入幻覺文字
		if td.rng.Float64() < 0.15 {
			hallucinationChar := td.getHallucinationCharacter()
			result = append(result, hallucinationChar)
			// 有時候替換，有時候插入
			if td.rng.Float64() < 0.5 {
				i++ // 替換
			}
			continue
		}

		// 20% 機率顛倒相鄰多個字符
		if i+2 < len(runes) && td.rng.Float64() < 0.20 {
			// 顛倒 2-3 個字符
			reverseLen := 2
			if i+3 < len(runes) && td.rng.Float64() < 0.5 {
				reverseLen = 3
			}

			for j := reverseLen - 1; j >= 0; j-- {
				if i+j < len(runes) {
					result = append(result, runes[i+j])
				}
			}
			i += reverseLen
			continue
		}

		result = append(result, r)
		i++
	}

	return string(result)
}

// getSimilarCharacter 取得視覺上相似的字符
// 用於模擬視覺模糊或閱讀困難
func (td *TextDistorter) getSimilarCharacter(r rune) rune {
	// 中文字符
	if r >= 0x4E00 && r <= 0x9FFF {
		// 對於中文，使用視覺相似字
		similarGroups := map[rune][]rune{
			'看': {'着', '看', '観'},
			'門': {'門', '问', '闸'},
			'聽': {'聽', '听', '德'},
			'說': {'說', '说', '詁'},
			'走': {'走', '起', '赴'},
			'來': {'來', '来', '求'},
			'去': {'去', '云', '丢'},
			'人': {'人', '入', '个'},
			'手': {'手', '毛', '扌'},
			'心': {'心', '忄', '必'},
		}

		if similar, ok := similarGroups[r]; ok {
			return similar[td.rng.Intn(len(similar))]
		}

		// 否則隨機微調（相鄰字符）
		offset := td.rng.Intn(5) - 2
		return r + rune(offset)
	}

	// 英文字母
	if unicode.IsLetter(r) {
		similarMap := map[rune][]rune{
			'a': {'a', 'á', 'à', 'â', 'ä', '@'},
			'e': {'e', 'é', 'è', 'ê', 'ë', '3'},
			'i': {'i', 'í', 'ì', 'î', 'ï', '1', '!'},
			'o': {'o', 'ó', 'ò', 'ô', 'ö', '0'},
			'u': {'u', 'ú', 'ù', 'û', 'ü'},
			'n': {'n', 'ñ', 'm'},
			'c': {'c', 'ç', 'ć'},
			's': {'s', 'ś', 'š', '$'},
			'A': {'A', 'Á', 'À', 'Â', 'Ä', '@'},
			'E': {'E', 'É', 'È', 'Ê', 'Ë'},
			'I': {'I', 'Í', 'Ì', 'Î', 'Ï', '1'},
			'O': {'O', 'Ó', 'Ò', 'Ô', 'Ö', '0'},
			'U': {'U', 'Ú', 'Ù', 'Û', 'Ü'},
			'N': {'N', 'Ñ'},
			'C': {'C', 'Ç', 'Ć'},
			'S': {'S', 'Ś', 'Š', '$'},
		}

		if similar, ok := similarMap[r]; ok {
			return similar[td.rng.Intn(len(similar))]
		}
	}

	// 數字 (但這不應該被調用，因為數字應該被保護)
	if unicode.IsDigit(r) {
		return r
	}

	return r
}

// getHallucinationCharacter 產生幻覺字符
// Story 9-5 AC3: 插入幻覺文字
func (td *TextDistorter) getHallucinationCharacter() rune {
	hallucinationChars := []rune{
		'█', '▓', '▒', '░', // 方塊
		'▀', '▄', '■', '□', // 幾何形狀
		'◆', '◇', '●', '○', // 圓形
		'▲', '△', '▼', '▽', // 三角形
		'★', '☆', '※', '＊', // 星號
		'？', '！', '…', '～', // 標點（扭曲版本）
		'⋯', '⋮', '⋰', '⋱', // 點點點
		'ﾞ', 'ﾟ', '゛', '゜', // 日文標記
		'々', '〃', '〆', '〇', // 日文符號
	}

	// 也包括一些詭異的中文字
	hallucinationChinese := []rune{
		'孽', '魘', '魅', '魔', '鬼',
		'屍', '骸', '煞', '厲', '冤',
		'殺', '死', '亡', '滅', '絕',
		'血', '肉', '腐', '爛', '噬',
	}

	// 70% 機率使用符號，30% 機率使用詭異文字
	if td.rng.Float64() < 0.7 {
		return hallucinationChars[td.rng.Intn(len(hallucinationChars))]
	}
	return hallucinationChinese[td.rng.Intn(len(hallucinationChinese))]
}

// isWhitespaceOrPunctuation 檢查是否為空白或標點符號
func isWhitespaceOrPunctuation(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsPunct(r)
}

// isNumericOnly 檢查字串是否只包含數字
func isNumericOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// DistortNarration 對旁白文字應用扭曲
// Story 9-5 AC4: 漸進式扭曲，不突然變化
func (td *TextDistorter) DistortNarration(text string, san int) string {
	return td.DistortText(text, san)
}

// DistortDialogue 對對話文字應用扭曲
// NPC 對話可能會因玩家幻覺而扭曲
func (td *TextDistorter) DistortDialogue(text string, san int) string {
	return td.DistortText(text, san)
}

// DistortChoice 對選項文字應用扭曲
// Story 9-5 AC5: 選項編號不扭曲，但選項內容會扭曲
func (td *TextDistorter) DistortChoice(choiceText string, san int) string {
	// 分離選項編號和內容
	parts := strings.SplitN(choiceText, ".", 2)
	if len(parts) == 2 {
		// 有選項編號
		number := strings.TrimSpace(parts[0])
		content := strings.TrimSpace(parts[1])

		// 只扭曲內容，保留編號
		distortedContent := td.DistortText(content, san)
		return number + ". " + distortedContent
	}

	// 沒有選項編號，直接扭曲
	return td.DistortText(choiceText, san)
}

// GetDistortionIntensity 獲取當前扭曲強度（0.0-1.0）
// 用於 UI 顯示或其他系統參考
func GetDistortionIntensity(san int) float64 {
	level := GetDistortionLevel(san)
	switch level {
	case DistortionNone:
		return 0.0
	case DistortionMild:
		return 0.25
	case DistortionModerate:
		return 0.5
	case DistortionSevere:
		return 0.85
	default:
		return 0.0
	}
}

// DistortWithControlLevel 使用 ControlLevel 進行扭曲
// 整合現有的 game.ControlLevel 系統
func (td *TextDistorter) DistortWithControlLevel(text string, controlLevel game.ControlLevel) string {
	return td.DistortText(text, controlLevel.SAN)
}
