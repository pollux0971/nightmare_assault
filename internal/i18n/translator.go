package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translator handles translation and localization
type Translator struct {
	locale       string
	translations map[string]interface{}
	fallback     map[string]interface{}
	mu           sync.RWMutex
}

var (
	globalTranslator *Translator
	once             sync.Once
)

// supportedLocales are the locales we support
var supportedLocales = []string{"zh-TW", "en-US"}

// New creates a new Translator with the specified locale
func New(locale string) (*Translator, error) {
	t := &Translator{
		locale: locale,
	}

	// Load translations for the specified locale
	if err := t.loadTranslations(locale); err != nil {
		return nil, fmt.Errorf("failed to load translations for %s: %w", locale, err)
	}

	// Load fallback (English)
	if locale != "en-US" {
		if err := t.loadFallback(); err != nil {
			return nil, fmt.Errorf("failed to load fallback translations: %w", err)
		}
	}

	return t, nil
}

// InitGlobal initializes the global translator
func InitGlobal(locale string) error {
	var err error
	once.Do(func() {
		globalTranslator, err = New(locale)
	})
	return err
}

// GetGlobal returns the global translator instance
func GetGlobal() *Translator {
	return globalTranslator
}

// SetGlobal sets the global translator instance
func SetGlobal(t *Translator) {
	globalTranslator = t
}

// loadTranslations loads translation file for a locale
func (t *Translator) loadTranslations(locale string) error {
	// Try embedded locales first, then external file
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(exePath)

	// Try multiple paths
	paths := []string{
		filepath.Join(execDir, "locales", fmt.Sprintf("%s.json", locale)),
		filepath.Join(execDir, "internal", "i18n", "locales", fmt.Sprintf("%s.json", locale)),
		fmt.Sprintf("internal/i18n/locales/%s.json", locale),
		fmt.Sprintf("locales/%s.json", locale),
	}

	var data []byte
	var lastErr error
	for _, path := range paths {
		data, lastErr = os.ReadFile(path)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return fmt.Errorf("could not find locale file for %s: %w", locale, lastErr)
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse locale file: %w", err)
	}

	t.mu.Lock()
	t.translations = translations
	t.mu.Unlock()

	return nil
}

// loadFallback loads English as fallback
func (t *Translator) loadFallback() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(exePath)

	paths := []string{
		filepath.Join(execDir, "locales", "en-US.json"),
		filepath.Join(execDir, "internal", "i18n", "locales", "en-US.json"),
		"internal/i18n/locales/en-US.json",
		"locales/en-US.json",
	}

	var data []byte
	var lastErr error
	for _, path := range paths {
		data, lastErr = os.ReadFile(path)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return fmt.Errorf("could not find fallback locale file: %w", lastErr)
	}

	var fallback map[string]interface{}
	if err := json.Unmarshal(data, &fallback); err != nil {
		return fmt.Errorf("failed to parse fallback locale file: %w", err)
	}

	t.mu.Lock()
	t.fallback = fallback
	t.mu.Unlock()

	return nil
}

// T translates a key with optional parameters
func (t *Translator) T(key string, params ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try current locale
	if val := t.lookup(key, t.translations); val != "" {
		return t.formatString(val, params...)
	}

	// Try fallback
	if val := t.lookup(key, t.fallback); val != "" {
		return t.formatString(val, params...)
	}

	// Return key if not found
	return fmt.Sprintf("[%s]", key)
}

// lookup looks up a nested key in translations map
func (t *Translator) lookup(key string, translations map[string]interface{}) string {
	parts := strings.Split(key, ".")
	var current interface{} = translations

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return ""
}

// formatString formats a string with parameters
func (t *Translator) formatString(template string, params ...interface{}) string {
	if len(params) == 0 {
		return template
	}

	// Simple parameter replacement: {0}, {1}, {2}, etc.
	result := template
	for i, param := range params {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(param))
	}

	// Named parameter replacement: {name}, {hp}, etc.
	if len(params)%2 == 0 {
		for i := 0; i < len(params); i += 2 {
			key := fmt.Sprint(params[i])
			val := fmt.Sprint(params[i+1])
			placeholder := fmt.Sprintf("{%s}", key)
			result = strings.ReplaceAll(result, placeholder, val)
		}
	}

	return result
}

// SetLocale changes the current locale
func (t *Translator) SetLocale(locale string) error {
	if !IsValidLocale(locale) {
		return fmt.Errorf("unsupported locale: %s", locale)
	}

	if err := t.loadTranslations(locale); err != nil {
		return err
	}

	t.mu.Lock()
	t.locale = locale
	t.mu.Unlock()

	return nil
}

// GetLocale returns the current locale
func (t *Translator) GetLocale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.locale
}

// IsValidLocale checks if a locale is supported
func IsValidLocale(locale string) bool {
	for _, supported := range supportedLocales {
		if supported == locale {
			return true
		}
	}
	return false
}

// GetSupportedLocales returns all supported locales
func GetSupportedLocales() []string {
	return supportedLocales
}

// DetectSystemLanguage detects the system language and returns appropriate locale
func DetectSystemLanguage() string {
	// Check LANG environment variable
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		return "en-US" // Default to English
	}

	// Extract language code (e.g., "zh_TW.UTF-8" -> "zh_TW")
	lang = strings.Split(lang, ".")[0]
	lang = strings.ReplaceAll(lang, "_", "-")

	// Map to supported locales
	if strings.HasPrefix(lang, "zh-TW") || strings.HasPrefix(lang, "zh-HK") {
		return "zh-TW"
	}
	if strings.HasPrefix(lang, "zh") {
		return "zh-TW" // Default Chinese to Traditional
	}

	return "en-US"
}
