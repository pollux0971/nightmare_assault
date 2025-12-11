package save

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrorCategory represents the category of a save/load error.
type ErrorCategory int

const (
	CategoryUnknown ErrorCategory = iota
	CategoryPermission
	CategoryDiskFull
	CategoryNotFound
	CategoryCorrupted
	CategoryIO
)

// SaveError represents a save operation error with narrative context.
type SaveError struct {
	Narrative          string
	Cause              error
	RecoverySuggestion string
	Category           ErrorCategory
}

func (e *SaveError) Error() string {
	if e.RecoverySuggestion != "" {
		return fmt.Sprintf("%s（%s）", e.Narrative, e.RecoverySuggestion)
	}
	return e.Narrative
}

func (e *SaveError) Unwrap() error {
	return e.Cause
}

// LoadError represents a load operation error with narrative context.
type LoadError struct {
	Narrative          string
	Cause              error
	RecoverySuggestion string
	Category           ErrorCategory
}

func (e *LoadError) Error() string {
	if e.RecoverySuggestion != "" {
		return fmt.Sprintf("%s（%s）", e.Narrative, e.RecoverySuggestion)
	}
	return e.Narrative
}

func (e *LoadError) Unwrap() error {
	return e.Cause
}

// narrativeMessages maps error categories to narrative messages.
var narrativeMessages = map[ErrorCategory]struct {
	SaveNarrative string
	LoadNarrative string
	Recovery      string
}{
	CategoryPermission: {
		SaveNarrative: "記憶無法固化...你的思緒在虛空中流失。",
		LoadNarrative: "記憶被禁錮在無法觸及的領域...",
		Recovery:      "檢查磁碟權限",
	},
	CategoryDiskFull: {
		SaveNarrative: "虛空已滿，無法容納更多記憶...",
		LoadNarrative: "虛空已滿...",
		Recovery:      "需要至少 1MB 空間",
	},
	CategoryNotFound: {
		SaveNarrative: "這個存檔槽位消失在虛空中...",
		LoadNarrative: "這個存檔槽位是空的，沒有可以喚醒的記憶。",
		Recovery:      "請選擇其他槽位",
	},
	CategoryCorrupted: {
		SaveNarrative: "記憶正在崩解...",
		LoadNarrative: "這段記憶已被扭曲...無法還原現實。",
		Recovery:      "存檔檔案可能已損壞",
	},
	CategoryIO: {
		SaveNarrative: "現實的連結斷開了...",
		LoadNarrative: "無法讀取遠古的記憶...",
		Recovery:      "檔案存取錯誤",
	},
	CategoryUnknown: {
		SaveNarrative: "未知的力量阻止了記憶的保存...",
		LoadNarrative: "未知的力量阻止了記憶的喚醒...",
		Recovery:      "請稍後再試",
	},
}

// CategorizeError determines the error category based on the error type.
func CategorizeError(err error) ErrorCategory {
	if err == nil {
		return CategoryUnknown
	}

	// Check for specific error types
	if errors.Is(err, os.ErrPermission) {
		return CategoryPermission
	}

	if errors.Is(err, os.ErrNotExist) {
		return CategoryNotFound
	}

	// Check for corrupted error
	var corruptedErr *CorruptedError
	if errors.As(err, &corruptedErr) {
		return CategoryCorrupted
	}

	// Check error message for disk full indicators
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "no space") ||
		strings.Contains(errStr, "disk full") ||
		strings.Contains(errStr, "空間不足") {
		return CategoryDiskFull
	}

	// Check for I/O errors
	if strings.Contains(errStr, "i/o") ||
		strings.Contains(errStr, "input/output") {
		return CategoryIO
	}

	return CategoryUnknown
}

// NarrativeError converts an error to a narrative message.
func NarrativeError(err error) string {
	if err == nil {
		return ""
	}

	category := CategorizeError(err)
	msgs := narrativeMessages[category]

	// Use load narrative by default (more commonly displayed)
	narrative := msgs.LoadNarrative
	if narrative == "" {
		narrative = "發生了未知的錯誤..."
	}

	// Append technical info for debugging
	return fmt.Sprintf("%s（%v）", narrative, err)
}

// WrapSaveError wraps an error with save-specific narrative context.
func WrapSaveError(err error) *SaveError {
	if err == nil {
		return nil
	}

	category := CategorizeError(err)
	msgs := narrativeMessages[category]

	return &SaveError{
		Narrative:          msgs.SaveNarrative,
		Cause:              err,
		RecoverySuggestion: msgs.Recovery,
		Category:           category,
	}
}

// WrapLoadError wraps an error with load-specific narrative context.
func WrapLoadError(err error) *LoadError {
	if err == nil {
		return nil
	}

	category := CategorizeError(err)
	msgs := narrativeMessages[category]

	return &LoadError{
		Narrative:          msgs.LoadNarrative,
		Cause:              err,
		RecoverySuggestion: msgs.Recovery,
		Category:           category,
	}
}

// RecoverySuggestionFor returns a recovery suggestion for an error.
func RecoverySuggestionFor(err error) string {
	category := CategorizeError(err)
	return narrativeMessages[category].Recovery
}
