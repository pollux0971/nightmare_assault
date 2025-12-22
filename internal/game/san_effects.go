// Package game provides game-related types and logic for Nightmare Assault.
package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// init seeds the global random number generator for deterministic testing.
// Code Review Fix 7-4-3: Explicit seeding for reproducibility.
// Note: Go 1.20+ auto-seeds, but explicit seeding documents intent.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// SANState represents the player's mental state based on SAN value.
// Story 7.4 AC4: SAN 狀態效果分級
type SANState int

const (
	// SANStateClear represents 80-100 SAN (清醒 Clear)
	SANStateClear SANState = iota
	// SANStateAnxious represents 50-79 SAN (焦慮 Anxious)
	SANStateAnxious
	// SANStatePanic represents 20-49 SAN (恐慌 Panic)
	SANStatePanic
	// SANStateBreakdown represents 1-19 SAN (崩潰 Breakdown)
	SANStateBreakdown
	// SANStateInsane represents 0 SAN (瘋狂 Insane)
	SANStateInsane
)

// String returns the display name of the SAN state.
func (s SANState) String() string {
	switch s {
	case SANStateClear:
		return "清醒"
	case SANStateAnxious:
		return "焦慮"
	case SANStatePanic:
		return "恐慌"
	case SANStateBreakdown:
		return "崩潰"
	case SANStateInsane:
		return "瘋狂"
	default:
		return "未知"
	}
}

// Description returns the description of the SAN state effects.
func (s SANState) Description() string {
	switch s {
	case SANStateClear:
		return "思緒清晰，感知正常"
	case SANStateAnxious:
		return "心跳加速，開始冒冷汗"
	case SANStatePanic:
		return "視覺扭曲，感官欺騙"
	case SANStateBreakdown:
		return "嚴重幻覺，失去部分控制"
	case SANStateInsane:
		return "理智崩潰"
	default:
		return ""
	}
}

// GetSANState calculates the SAN state based on current SAN value.
// Story 7.4 AC4: SAN 狀態判定邏輯
func GetSANState(san int) SANState {
	switch {
	case san >= 80:
		return SANStateClear
	case san >= 50:
		return SANStateAnxious
	case san >= 20:
		return SANStatePanic
	case san >= 1:
		return SANStateBreakdown
	default:
		return SANStateInsane
	}
}

// DamageType represents the type of damage dealt to the player.
type DamageType int

const (
	// HP Damage Types (Story 7.4 AC2)
	DamageMinor    DamageType = iota // 輕傷：-10 ~ -20
	DamageModerate                    // 中傷：-30 ~ -50
	DamageSevere                      // 重傷：-60 ~ -80
	DamageFatal                       // 致命傷：-100

	// SAN Damage Types (Story 7.4 AC3)
	SANTeammateDeath  // 目睹隊友死亡：-15 ~ -25
	SANHorrorScene    // 遭遇恐怖場景：-5 ~ -15
	SANMonsterSighting // 看到怪物實體：-20 ~ -30
	SANCruelTruth     // 發現殘酷真相：-10 ~ -20
	SANRulePenalty    // 觸發規則懲罰：依規則設定
)

// String returns the display name of the damage type.
func (d DamageType) String() string {
	switch d {
	case DamageMinor:
		return "輕傷"
	case DamageModerate:
		return "中傷"
	case DamageSevere:
		return "重傷"
	case DamageFatal:
		return "致命傷"
	case SANTeammateDeath:
		return "目睹隊友死亡"
	case SANHorrorScene:
		return "遭遇恐怖場景"
	case SANMonsterSighting:
		return "看到怪物實體"
	case SANCruelTruth:
		return "發現殘酷真相"
	case SANRulePenalty:
		return "違反規則"
	default:
		return "未知"
	}
}

// CalculateDamage calculates the actual damage value for a given damage type.
// Returns a negative value for damage.
func CalculateDamage(damageType DamageType) int {
	switch damageType {
	// HP Damage
	case DamageMinor:
		return -(10 + rand.Intn(11)) // -10 to -20
	case DamageModerate:
		return -(30 + rand.Intn(21)) // -30 to -50
	case DamageSevere:
		return -(60 + rand.Intn(21)) // -60 to -80
	case DamageFatal:
		return -100

	// SAN Damage
	case SANTeammateDeath:
		return -(15 + rand.Intn(11)) // -15 to -25
	case SANHorrorScene:
		return -(5 + rand.Intn(11)) // -5 to -15
	case SANMonsterSighting:
		return -(20 + rand.Intn(11)) // -20 to -30
	case SANCruelTruth:
		return -(10 + rand.Intn(11)) // -10 to -20

	case SANRulePenalty:
		// Rule penalty default: -15 to -25 (can be overridden by specific rule config)
		// Story 7.4 AC3: 觸發規則懲罰
		return -(15 + rand.Intn(11)) // -15 to -25

	default:
		return 0
	}
}

// HealType represents the type of healing/recovery.
type HealType int

const (
	// HP Heal Types (Story 7.4 AC2)
	HealItem HealType = iota // 使用恢復道具
	HealMedical              // 獲得治療

	// SAN Heal Types (Story 7.4 AC5)
	SANRest          // 安全場景休息：+5 ~ +10
	SANPuzzleSolved  // 解決謎題或發現真相：+10 ~ +15
	SANMedicine      // 使用特殊道具 (安定劑)：+20
	SANPositiveInteraction // 與 NPC 正向互動：+5
)

// String returns the display name of the heal type.
func (h HealType) String() string {
	switch h {
	case HealItem:
		return "使用恢復道具"
	case HealMedical:
		return "獲得治療"
	case SANRest:
		return "安全場景休息"
	case SANPuzzleSolved:
		return "解決謎題"
	case SANMedicine:
		return "使用安定劑"
	case SANPositiveInteraction:
		return "正向互動"
	default:
		return "未知"
	}
}

// CalculateHeal calculates the actual heal value for a given heal type.
// Returns a positive value for healing.
func CalculateHeal(healType HealType) int {
	switch healType {
	// HP Heal - varies by item/treatment
	case HealItem:
		return 20 + rand.Intn(21) // +20 to +40
	case HealMedical:
		return 30 + rand.Intn(31) // +30 to +60

	// SAN Heal
	case SANRest:
		return 5 + rand.Intn(6) // +5 to +10
	case SANPuzzleSolved:
		return 10 + rand.Intn(6) // +10 to +15
	case SANMedicine:
		return 20 // Fixed +20
	case SANPositiveInteraction:
		return 5 // Fixed +5

	default:
		return 0
	}
}

// SANEffectProfile defines the visual and gameplay effects for a SAN state.
type SANEffectProfile struct {
	State                SANState
	TextCorruptionLevel  float64 // 0.0 - 1.0
	ColorShiftIntensity  float64 // 0.0 - 1.0
	UIStabilityLevel     int     // 0 = stable, 5 = maximum shake
	HallucinationChance  float64 // 0.0 - 1.0, chance to add hallucination options
	ForcedActionChance   float64 // 0.0 - 1.0, chance to force an action
	NarrativeModifier    string  // Modifier for narration agent prompt
}

// GetSANEffectProfile returns the effect profile for a given SAN state.
// Story 7.4 AC4: 根據 SAN 狀態套用對應效果
func GetSANEffectProfile(state SANState) SANEffectProfile {
	switch state {
	case SANStateClear:
		// 80-100: Normal, no effects
		return SANEffectProfile{
			State:                SANStateClear,
			TextCorruptionLevel:  0.0,
			ColorShiftIntensity:  0.0,
			UIStabilityLevel:     0,
			HallucinationChance:  0.0,
			ForcedActionChance:   0.0,
			NarrativeModifier:    "正常描述，清晰的場景與對話",
		}

	case SANStateAnxious:
		// 50-79: Mild anxiety effects
		return SANEffectProfile{
			State:                SANStateAnxious,
			TextCorruptionLevel:  0.1,
			ColorShiftIntensity:  0.2,
			UIStabilityLevel:     1,
			HallucinationChance:  0.0,
			ForcedActionChance:   0.0,
			NarrativeModifier:    "加入心跳聲、冷汗描述。文字顏色微妙變化。可能有焦慮影響的選項。",
		}

	case SANStatePanic:
		// 20-49: Panic with hallucinations
		return SANEffectProfile{
			State:                SANStatePanic,
			TextCorruptionLevel:  0.3,
			ColorShiftIntensity:  0.5,
			UIStabilityLevel:     3,
			HallucinationChance:  0.3,
			ForcedActionChance:   0.0,
			NarrativeModifier:    "文字部分扭曲或模糊。感官欺騙（看到不存在的東西）。可能出現「恐慌選項」(非理性行為)。顏色飽和度降低。",
		}

	case SANStateBreakdown:
		// 1-19: Severe breakdown
		return SANEffectProfile{
			State:                SANStateBreakdown,
			TextCorruptionLevel:  0.7,
			ColorShiftIntensity:  0.8,
			UIStabilityLevel:     5,
			HallucinationChance:  0.6,
			ForcedActionChance:   0.2,
			NarrativeModifier:    "嚴重文字扭曲。幻覺選項混入真實選項（不標記）。可能出現強制行為（失去部分控制權）。螢幕效果嚴重干擾。",
		}

	case SANStateInsane:
		// 0: Complete insanity - game over
		return SANEffectProfile{
			State:                SANStateInsane,
			TextCorruptionLevel:  1.0,
			ColorShiftIntensity:  1.0,
			UIStabilityLevel:     5,
			HallucinationChance:  1.0,
			ForcedActionChance:   1.0,
			NarrativeModifier:    "完全瘋狂。遊戲結束 Bad End (理智崩潰結局)。",
		}

	default:
		return GetSANEffectProfile(SANStateClear)
	}
}

// ApplyTextDistortion applies text distortion effects based on SAN state.
// Story 7.4 AC4: 文字扭曲效果
func ApplyTextDistortion(text string, state SANState) string {
	profile := GetSANEffectProfile(state)

	if profile.TextCorruptionLevel == 0.0 {
		return text
	}

	// Simple distortion: randomly replace characters
	runes := []rune(text)
	distortedRunes := make([]rune, len(runes))

	for i, r := range runes {
		// Skip spaces and punctuation
		if r == ' ' || r == '\n' || r == ',' || r == '.' || r == '!' || r == '?' {
			distortedRunes[i] = r
			continue
		}

		// Randomly distort based on corruption level
		if rand.Float64() < profile.TextCorruptionLevel {
			// Replace with similar-looking character
			distortedRunes[i] = getDistortedChar(r)
		} else {
			distortedRunes[i] = r
		}
	}

	return string(distortedRunes)
}

// getDistortedChar returns a distorted version of the input character.
func getDistortedChar(r rune) rune {
	// For Chinese characters, slightly modify
	if r >= 0x4E00 && r <= 0x9FFF {
		// Offset by a small random amount
		offset := rand.Intn(5) - 2
		return r + rune(offset)
	}

	// For English letters, replace with similar-looking ones
	similarChars := map[rune][]rune{
		'a': {'@', 'á', 'à', 'â'},
		'e': {'é', 'è', 'ê', '3'},
		'i': {'í', 'ì', 'î', '1', '!'},
		'o': {'ó', 'ò', 'ô', '0'},
		'u': {'ú', 'ù', 'û'},
		'A': {'@', 'Á', 'À', 'Â'},
		'E': {'É', 'È', 'Ê', '3'},
		'I': {'Í', 'Ì', 'Î', '1', '!'},
		'O': {'Ó', 'Ò', 'Ô', '0'},
		'U': {'Ú', 'Ù', 'Û'},
	}

	if similar, ok := similarChars[r]; ok && len(similar) > 0 {
		return similar[rand.Intn(len(similar))]
	}

	return r
}

// GenerateHallucinationOption generates a hallucination option based on SAN state.
// Story 7.4 AC4: 幻覺選項生成邏輯
func GenerateHallucinationOption(state SANState, context string) string {
	profile := GetSANEffectProfile(state)

	if profile.HallucinationChance == 0.0 {
		return ""
	}

	// Check if we should generate a hallucination
	if rand.Float64() > profile.HallucinationChance {
		return ""
	}

	// Generate hallucination based on state
	hallucinations := []string{
		"你聽到有人在呼喊你的名字...",
		"牆壁上出現了一扇門（實際上不存在）",
		"地上有一把鑰匙在發光（幻覺）",
		"你看到一個熟悉的身影在遠處（幻覺）",
		"角落裡有什麼東西在動（但什麼都沒有）",
	}

	if state == SANStateBreakdown {
		hallucinations = append(hallucinations,
			"攻擊眼前的人（他們可能是敵人）",
			"吞下手中的不明物體（看起來像藥）",
			"跟隨那個聲音走（實際上是幻聽）",
		)
	}

	return hallucinations[rand.Intn(len(hallucinations))]
}

// ShouldForceAction determines if the player should be forced into an action.
// Story 7.4 AC4: 強制行為判定
func ShouldForceAction(state SANState) (bool, string) {
	profile := GetSANEffectProfile(state)

	if profile.ForcedActionChance == 0.0 {
		return false, ""
	}

	// Check if we should force an action
	if rand.Float64() > profile.ForcedActionChance {
		return false, ""
	}

	// Generate forced action based on state
	forcedActions := []string{
		"你無法控制地顫抖",
		"你突然尖叫出聲",
		"你不由自主地後退了幾步",
		"你的雙手開始不受控制",
	}

	return true, forcedActions[rand.Intn(len(forcedActions))]
}

// GetNarrativeStyle returns the narrative style prompt based on SAN state.
// Story 7.4 AC4: Narration Agent 整合
func GetNarrativeStyle(state SANState) string {
	profile := GetSANEffectProfile(state)
	return profile.NarrativeModifier
}

// FormatSANStateSummary returns a formatted summary of the current SAN state.
func FormatSANStateSummary(san int) string {
	state := GetSANState(san)
	profile := GetSANEffectProfile(state)

	var parts []string
	parts = append(parts, fmt.Sprintf("SAN: %d/100", san))
	parts = append(parts, fmt.Sprintf("狀態: %s", state.String()))
	parts = append(parts, fmt.Sprintf("描述: %s", state.Description()))

	if profile.TextCorruptionLevel > 0 {
		parts = append(parts, fmt.Sprintf("文字扭曲: %.0f%%", profile.TextCorruptionLevel*100))
	}
	if profile.HallucinationChance > 0 {
		parts = append(parts, fmt.Sprintf("幻覺機率: %.0f%%", profile.HallucinationChance*100))
	}

	return strings.Join(parts, " | ")
}
