package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
)

// LangCommand handles language switching
type LangCommand struct {
	config     *config.Config
	translator *i18n.Translator
}

// NewLangCommand creates a new lang command
func NewLangCommand(cfg *config.Config, translator *i18n.Translator) *LangCommand {
	return &LangCommand{
		config:     cfg,
		translator: translator,
	}
}

// Execute executes the lang command
func (c *LangCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		// Show current language
		currentLocale := c.translator.GetLocale()
		languageName := c.translator.T(fmt.Sprintf("languages.%s", currentLocale))
		supportedList := make([]string, 0)
		for _, locale := range i18n.GetSupportedLocales() {
			name := c.translator.T(fmt.Sprintf("languages.%s", locale))
			supportedList = append(supportedList, fmt.Sprintf("%s (%s)", name, locale))
		}

		return fmt.Sprintf("%s: %s (%s)\n\n%s:\n- %s",
			c.translator.T("commands.lang"),
			languageName,
			currentLocale,
			c.translator.T("settings.language"),
			strings.Join(supportedList, "\n- "),
		), nil
	}

	// Change language
	newLocale := args[0]

	// Validate locale
	if !i18n.IsValidLocale(newLocale) {
		errMsg := c.translator.T("errors.invalid_locale", newLocale)
		return "", fmt.Errorf("%s", errMsg)
	}

	// Get old locale for message
	oldLocale := c.translator.GetLocale()

	// Set new locale
	if err := c.translator.SetLocale(newLocale); err != nil {
		return "", fmt.Errorf("failed to switch language: %w", err)
	}

	// Update global translator
	i18n.SetGlobal(c.translator)

	// Save to config
	c.config.Language = newLocale
	if err := c.config.Save(); err != nil {
		// Log error but don't fail the command
		fmt.Printf("Warning: failed to save language preference: %v\n", err)
	}

	// Return confirmation message in NEW language
	languageName := c.translator.T(fmt.Sprintf("languages.%s", newLocale))
	message := c.translator.T("messages.language_changed", languageName)

	// Add note about new content
	if oldLocale != newLocale {
		message += "\n" + c.translator.T("messages.language_will_apply", languageName)
	}

	return message, nil
}

// Name returns the command name
func (c *LangCommand) Name() string {
	return "lang"
}

// Description returns the command description
func (c *LangCommand) Description() string {
	if c.translator != nil {
		return c.translator.T("commands.lang_desc")
	}
	return "Switch language (/lang zh-TW or /lang en-US)"
}

// Usage returns the command usage
func (c *LangCommand) Usage() string {
	return "/lang [zh-TW|en-US]"
}
