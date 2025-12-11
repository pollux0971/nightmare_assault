package i18n

import (
	"testing"
)

func TestDetectSystemLanguage(t *testing.T) {
	// This test can only verify it returns a valid locale
	lang := DetectSystemLanguage()

	if !IsValidLocale(lang) {
		t.Errorf("DetectSystemLanguage() returned invalid locale: %s", lang)
	}
}

func TestIsValidLocale(t *testing.T) {
	tests := []struct {
		locale string
		valid  bool
	}{
		{"zh-TW", true},
		{"en-US", true},
		{"fr-FR", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			result := IsValidLocale(tt.locale)
			if result != tt.valid {
				t.Errorf("IsValidLocale(%s) = %v, want %v", tt.locale, result, tt.valid)
			}
		})
	}
}

func TestGetSupportedLocales(t *testing.T) {
	locales := GetSupportedLocales()

	if len(locales) < 2 {
		t.Error("Expected at least 2 supported locales")
	}

	// Check for required locales
	hasZhTW := false
	hasEnUS := false
	for _, locale := range locales {
		if locale == "zh-TW" {
			hasZhTW = true
		}
		if locale == "en-US" {
			hasEnUS = true
		}
	}

	if !hasZhTW {
		t.Error("Missing zh-TW in supported locales")
	}
	if !hasEnUS {
		t.Error("Missing en-US in supported locales")
	}
}

func TestTranslatorLookup(t *testing.T) {
	tr := &Translator{
		translations: map[string]interface{}{
			"menu": map[string]interface{}{
				"new_game": "新遊戲",
				"quit":     "離開",
			},
			"simple": "簡單文字",
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"menu.new_game", "新遊戲"},
		{"menu.quit", "離開"},
		{"simple", "簡單文字"},
		{"nonexistent", ""},
		{"menu.nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := tr.lookup(tt.key, tr.translations)
			if result != tt.expected {
				t.Errorf("lookup(%s) = %s, want %s", tt.key, result, tt.expected)
			}
		})
	}
}

func TestTranslatorFormatString(t *testing.T) {
	tr := &Translator{}

	tests := []struct {
		name     string
		template string
		params   []interface{}
		expected string
	}{
		{
			name:     "no params",
			template: "Hello World",
			params:   nil,
			expected: "Hello World",
		},
		{
			name:     "positional params",
			template: "API failed: {0}",
			params:   []interface{}{"connection timeout"},
			expected: "API failed: connection timeout",
		},
		{
			name:     "multiple positional params",
			template: "{0} → {1}",
			params:   []interface{}{"v1.0.0", "v1.1.0"},
			expected: "v1.0.0 → v1.1.0",
		},
		{
			name:     "named params",
			template: "HP: {hp}/{max_hp}",
			params:   []interface{}{"hp", "50", "max_hp", "100"},
			expected: "HP: 50/100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.formatString(tt.template, tt.params...)
			if result != tt.expected {
				t.Errorf("formatString() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestTranslatorT(t *testing.T) {
	tr := &Translator{
		translations: map[string]interface{}{
			"menu": map[string]interface{}{
				"new_game": "新遊戲",
			},
			"errors": map[string]interface{}{
				"api_failed": "API 請求失敗: {0}",
			},
		},
		fallback: map[string]interface{}{
			"menu": map[string]interface{}{
				"settings": "Settings",
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		params   []interface{}
		expected string
	}{
		{
			name:     "simple translation",
			key:      "menu.new_game",
			params:   nil,
			expected: "新遊戲",
		},
		{
			name:     "translation with params",
			key:      "errors.api_failed",
			params:   []interface{}{"timeout"},
			expected: "API 請求失敗: timeout",
		},
		{
			name:     "fallback translation",
			key:      "menu.settings",
			params:   nil,
			expected: "Settings",
		},
		{
			name:     "missing translation",
			key:      "nonexistent.key",
			params:   nil,
			expected: "[nonexistent.key]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.T(tt.key, tt.params...)
			if result != tt.expected {
				t.Errorf("T(%s) = %s, want %s", tt.key, result, tt.expected)
			}
		})
	}
}

func TestTranslatorGetSetLocale(t *testing.T) {
	tr := &Translator{
		locale: "zh-TW",
	}

	// Test GetLocale
	if got := tr.GetLocale(); got != "zh-TW" {
		t.Errorf("GetLocale() = %s, want zh-TW", got)
	}
}
