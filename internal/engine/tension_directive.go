package engine

import "fmt"

// LengthRange represents the min/max length range for story content
type LengthRange struct {
	Min int `json:"min"` // Minimum word count
	Max int `json:"max"` // Maximum word count
}

// TensionDirective provides guidance to the LLM for narrative generation
// based on current tension level
type TensionDirective struct {
	Level             TensionLevel `json:"level"`              // Current tension level
	Instruction       string       `json:"instruction"`        // Main guidance for LLM
	AllowedElements   []string     `json:"allowed_elements"`   // Elements that can be used
	ForbiddenElements []string     `json:"forbidden_elements"` // Elements that should not be used
	LengthRange       LengthRange  `json:"length_range"`       // Word count range
}

// GenerateDirective creates a TensionDirective based on the current tension level
func GenerateDirective(level TensionLevel) *TensionDirective {
	switch level {
	case TensionLevelLow:
		return &TensionDirective{
			Level:       TensionLevelLow,
			Instruction: "鋪墊階段：詳細環境描寫、氛圍營造、線索埋設",
			AllowedElements: []string{
				"環境異常",
				"微妙違和",
				"不安感",
			},
			ForbiddenElements: []string{
				"直接攻擊",
				"追逐戰",
				"實體威脅",
			},
			LengthRange: LengthRange{
				Min: 500,
				Max: 800,
			},
		}

	case TensionLevelMedium:
		return &TensionDirective{
			Level:       TensionLevelMedium,
			Instruction: "懸疑階段：增加緊張感、間接威脅、NPC 異常",
			AllowedElements: []string{
				"聲音",
				"陰影",
				"間接威脅",
				"隊友緊張",
			},
			ForbiddenElements: []string{
				"直接衝突",
			},
			LengthRange: LengthRange{
				Min: 600,
				Max: 1000,
			},
		}

	case TensionLevelHigh:
		return &TensionDirective{
			Level:       TensionLevelHigh,
			Instruction: "高潮階段：直接衝突、生死抉擇、規則違反後果",
			AllowedElements: []string{
				"直接攻擊",
				"追逐",
				"規則觸發",
				"死亡威脅",
			},
			ForbiddenElements: []string{},
			LengthRange: LengthRange{
				Min: 800,
				Max: 1200,
			},
		}

	default:
		// Fallback to LOW
		return GenerateDirective(TensionLevelLow)
	}
}

// GenerateDirectiveFromValue creates a TensionDirective from a tension value
func GenerateDirectiveFromValue(value int) *TensionDirective {
	level := CalculateLevel(value)
	return GenerateDirective(level)
}

// FormatForPrompt formats the directive as a string suitable for LLM prompts
func (td *TensionDirective) FormatForPrompt() string {
	prompt := "【張力指令】\n"
	prompt += "等級：" + string(td.Level) + "\n"
	prompt += "指示：" + td.Instruction + "\n"

	if len(td.AllowedElements) > 0 {
		prompt += "允許元素：\n"
		for _, elem := range td.AllowedElements {
			prompt += "  - " + elem + "\n"
		}
	}

	if len(td.ForbiddenElements) > 0 {
		prompt += "禁止元素：\n"
		for _, elem := range td.ForbiddenElements {
			prompt += "  - " + elem + "\n"
		}
	}

	prompt += "篇幅範圍：" +
		fmt.Sprintf("%d-%d 字\n", td.LengthRange.Min, td.LengthRange.Max)

	return prompt
}
