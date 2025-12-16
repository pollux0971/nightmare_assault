package context

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/pkoukk/tiktoken-go"
)

// TokenCounter 定義 Token 計數接口
type TokenCounter interface {
	CountTokens(text string) int
	GetEncodingName() string // 例如 "cl100k_base" (GPT-4)
}

// TokenUsageReport Token 使用報告
type TokenUsageReport struct {
	SummaryTokens       int     `json:"summary_tokens"`        // 摘要 tokens
	RecentEntriesTokens int     `json:"recent_entries_tokens"` // 最近歷史 tokens
	TotalTokens         int     `json:"total_tokens"`          // 總計
	ModelLimit          int     `json:"model_limit"`           // 模型限制
	UsagePercentage     float64 `json:"usage_percentage"`      // 使用百分比
	IsWarning           bool    `json:"is_warning"`            // 是否警告 (>70%)
}

// =============================================================================
// TiktokenCounter - 精確 Token 計數器（使用 tiktoken-go）
// =============================================================================

// TiktokenCounter 使用 tiktoken-go 庫進行精確的 Token 計數
type TiktokenCounter struct {
	encoding     *tiktoken.Tiktoken
	encodingName string
}

// NewTiktokenCounter 創建基於 tiktoken 的 Token 計數器
// 支援的模型: "gpt-4", "gpt-4-turbo", "gpt-3.5-turbo", "text-davinci-003"
func NewTiktokenCounter(model string) (*TiktokenCounter, error) {
	// 根據模型選擇 encoding
	var encodingName string
	switch model {
	case "gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini":
		encodingName = "cl100k_base"
	case "gpt-3.5-turbo", "text-davinci-003":
		encodingName = "cl100k_base"
	default:
		// 默認使用 cl100k_base（最通用）
		encodingName = "cl100k_base"
	}

	// 獲取 encoding 實例
	enc, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding %s: %w", encodingName, err)
	}

	return &TiktokenCounter{
		encoding:     enc,
		encodingName: encodingName,
	}, nil
}

// CountTokens 計算文本的 Token 數量
func (tc *TiktokenCounter) CountTokens(text string) int {
	if text == "" {
		return 0
	}
	tokens := tc.encoding.Encode(text, nil, nil)
	return len(tokens)
}

// GetEncodingName 返回使用的 encoding 名稱
func (tc *TiktokenCounter) GetEncodingName() string {
	return tc.encodingName
}

// =============================================================================
// EstimateTokenCounter - 估算 Token 計數器（無外部依賴）
// =============================================================================

// EstimateTokenCounter 使用簡單估算方法計算 Token 數量
// 估算規則:
//   - 英文: 約 4 字元 = 1 token
//   - 中文: 約 2 字元 = 1 token
//   - 安全邊際: 向上取整 + 10% 緩衝
type EstimateTokenCounter struct{}

// NewEstimateTokenCounter 創建估算 Token 計數器
func NewEstimateTokenCounter() *EstimateTokenCounter {
	return &EstimateTokenCounter{}
}

// CountTokens 估算文本的 Token 數量
func (ec *EstimateTokenCounter) CountTokens(text string) int {
	if text == "" {
		return 0
	}

	// 統計中文和總字元數
	runeCount := utf8.RuneCountInString(text)
	chineseChars := countChineseChars(text)
	englishChars := runeCount - chineseChars

	// 估算公式:
	// - 中文: 2 字元 = 1 token
	// - 英文: 4 字元 = 1 token
	estimatedTokens := (chineseChars / 2) + (englishChars / 4)

	// 向上取整（如果有餘數）
	if chineseChars%2 != 0 {
		estimatedTokens++
	}
	if englishChars%4 != 0 {
		estimatedTokens++
	}

	// 加上 10% 安全邊際（寧可高估也不低估）
	withSafetyMargin := float64(estimatedTokens) * 1.1

	// 確保至少返回 1（如果有內容）
	result := int(withSafetyMargin)
	if result == 0 && runeCount > 0 {
		result = 1
	}

	return result
}

// GetEncodingName 返回估算方法描述
func (ec *EstimateTokenCounter) GetEncodingName() string {
	return "estimate_4char_en_2char_zh"
}

// countChineseChars 計算中文字元數量
// 使用 Unicode Han 字元範圍判斷
func countChineseChars(text string) int {
	count := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			count++
		}
	}
	return count
}
