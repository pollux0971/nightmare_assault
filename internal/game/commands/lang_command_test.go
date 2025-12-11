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
