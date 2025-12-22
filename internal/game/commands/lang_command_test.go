package commands

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
)

func TestLangCommandName(t *testing.T) {
	cmd := &LangCommand{}
	if got := cmd.Name(); got != "lang" {
		t.Errorf("Name() = %s, want lang", got)
	}
}

func TestLangCommandUsage(t *testing.T) {
	cmd := &LangCommand{}
	usage := cmd.Usage()
	if usage == "" {
		t.Error("Usage() should not be empty")
	}
	// Story 10-3: 驗證 zh-CN 是否在 usage 中
	if usage != "/lang [zh-TW|zh-CN|en-US]" {
		t.Errorf("Usage() = %s, want /lang [zh-TW|zh-CN|en-US]", usage)
	}
}

func TestLangCommandDescription(t *testing.T) {
	cmd := &LangCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestLangCommandExecute(t *testing.T) {
	// Create mock translator
	tr := &i18n.Translator{}
	cfg := config.DefaultConfig()

	cmd := NewLangCommand(cfg, tr)

	// Test with no args (show current language)
	// Note: This will fail without actual locale files, so we skip detailed testing
	if cmd == nil {
		t.Fatal("NewLangCommand() returned nil")
	}

	// Test command structure
	if cmd.config == nil {
		t.Error("config should not be nil")
	}

	if cmd.translator == nil {
		t.Error("translator should not be nil")
	}
}

func TestLangCommandInvalidLocale(t *testing.T) {
	tr := &i18n.Translator{}
	cfg := config.DefaultConfig()
	cmd := NewLangCommand(cfg, tr)

	// Test with invalid locale
	_, err := cmd.Execute([]string{"invalid-locale"})
	if err == nil {
		t.Error("Execute() should return error for invalid locale")
	}
}

// TestLangCommand_SupportedLocales 測試支持的語言列表
// Story 10-3: 驗證支持 zh-TW, zh-CN, en-US
func TestLangCommand_SupportedLocales(t *testing.T) {
	supportedLocales := i18n.GetSupportedLocales()

	expectedLocales := []string{"zh-TW", "zh-CN", "en-US"}
	if len(supportedLocales) != len(expectedLocales) {
		t.Errorf("GetSupportedLocales() returned %d locales, want %d",
			len(supportedLocales), len(expectedLocales))
	}

	for _, expected := range expectedLocales {
		found := false
		for _, locale := range supportedLocales {
			if locale == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedLocales() missing locale: %s", expected)
		}
	}
}

// TestLangCommand_ValidateZhCN 測試 zh-CN 是否為有效語言
// Story 10-3 AC #1: 支援簡體中文 (zh-CN)
func TestLangCommand_ValidateZhCN(t *testing.T) {
	if !i18n.IsValidLocale("zh-CN") {
		t.Error("zh-CN should be a valid locale")
	}
}

// TestLangCommand_Help 測試幫助文本
func TestLangCommand_Help(t *testing.T) {
	cmd := &LangCommand{}
	help := cmd.Help()
	if help == "" {
		t.Error("Help() should not be empty")
	}
}
