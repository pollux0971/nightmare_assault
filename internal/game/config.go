// Package game provides game-related types and logic for Nightmare Assault.
package game

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// DifficultyLevel represents the game difficulty.
type DifficultyLevel int

const (
	DifficultyEasy DifficultyLevel = iota
	DifficultyHard
	DifficultyHell
)

// String returns the display name of the difficulty.
func (d DifficultyLevel) String() string {
	switch d {
	case DifficultyEasy:
		return "簡單"
	case DifficultyHard:
		return "困難"
	case DifficultyHell:
		return "地獄"
	default:
		return "未知"
	}
}

// Description returns the description of the difficulty.
func (d DifficultyLevel) Description() string {
	switch d {
	case DifficultyEasy:
		return "HP 消耗 0.5x，SAN 消耗 0.7x，敵人較弱，提示啟用"
	case DifficultyHard:
		return "HP 消耗 1.0x，SAN 消耗 1.0x，敵人標準，標準模式"
	case DifficultyHell:
		return "HP 消耗 1.5x，SAN 消耗 1.3x，敵人致命，永久死亡"
	default:
		return ""
	}
}

// HPDrainMultiplier returns the HP drain rate multiplier.
func (d DifficultyLevel) HPDrainMultiplier() float64 {
	switch d {
	case DifficultyEasy:
		return 0.5
	case DifficultyHard:
		return 1.0
	case DifficultyHell:
		return 1.5
	default:
		return 1.0
	}
}

// SANDrainMultiplier returns the SAN drain rate multiplier.
func (d DifficultyLevel) SANDrainMultiplier() float64 {
	switch d {
	case DifficultyEasy:
		return 0.7
	case DifficultyHard:
		return 1.0
	case DifficultyHell:
		return 1.3
	default:
		return 1.0
	}
}

// HintsEnabled returns whether hints are enabled at this difficulty.
func (d DifficultyLevel) HintsEnabled() bool {
	return d == DifficultyEasy
}

// IsPermadeath returns whether permadeath mode is active.
func (d DifficultyLevel) IsPermadeath() bool {
	return d == DifficultyHell
}

// GameLength represents the game length setting.
type GameLength int

const (
	LengthShort  GameLength = iota // ~15 min
	LengthMedium                   // ~30 min
	LengthLong                     // ~60 min
)

// String returns the display name of the game length.
func (l GameLength) String() string {
	switch l {
	case LengthShort:
		return "短篇"
	case LengthMedium:
		return "中篇"
	case LengthLong:
		return "長篇"
	default:
		return "未知"
	}
}

// Description returns the description of the game length.
func (l GameLength) Description() string {
	switch l {
	case LengthShort:
		return "約 15 分鐘，精簡劇情"
	case LengthMedium:
		return "約 30 分鐘，標準體驗"
	case LengthLong:
		return "約 60 分鐘，完整冒險"
	default:
		return ""
	}
}

// EstimatedMinutes returns the estimated play time in minutes.
func (l GameLength) EstimatedMinutes() int {
	switch l {
	case LengthShort:
		return 15
	case LengthMedium:
		return 30
	case LengthLong:
		return 60
	default:
		return 30
	}
}

// EventCount returns the approximate number of story events.
func (l GameLength) EventCount() int {
	switch l {
	case LengthShort:
		return 8
	case LengthMedium:
		return 15
	case LengthLong:
		return 25
	default:
		return 15
	}
}

// GameConfig represents the configuration for a new game.
type GameConfig struct {
	Theme      string          `json:"theme"`
	Difficulty DifficultyLevel `json:"difficulty"`
	Length     GameLength      `json:"length"`
	AdultMode  bool            `json:"adult_mode"`
	CreatedAt  time.Time       `json:"created_at"`
	frozen     bool            // Prevents modification after game starts
}

// Validation errors
var (
	ErrThemeTooShort      = errors.New("主題必須至少 3 個字元")
	ErrThemeTooLong       = errors.New("主題不能超過 100 個字元")
	ErrThemeInvalidChars  = errors.New("主題包含不允許的字元")
	ErrInvalidDifficulty  = errors.New("無效的難度設定")
	ErrInvalidLength      = errors.New("無效的遊戲長度")
	ErrConfigFrozen       = errors.New("遊戲配置已凍結，無法修改")
)

// Theme validation constants
const (
	ThemeMinLength = 3
	ThemeMaxLength = 100
)

// dangerousPatterns contains patterns that could be used for prompt injection
var dangerousPatterns = []string{
	"ignore",
	"forget",
	"disregard",
	"system:",
	"assistant:",
	"user:",
	"<|",
	"|>",
	"```",
	"[INST]",
	"[/INST]",
}

// NewGameConfig creates a new GameConfig with default values.
func NewGameConfig() *GameConfig {
	return &GameConfig{
		Theme:      "",
		Difficulty: DifficultyHard,
		Length:     LengthMedium,
		AdultMode:  false,
		CreatedAt:  time.Now(),
	}
}

// ValidateTheme validates the theme string.
func ValidateTheme(theme string) error {
	theme = strings.TrimSpace(theme)
	length := utf8.RuneCountInString(theme)

	if length < ThemeMinLength {
		return ErrThemeTooShort
	}
	if length > ThemeMaxLength {
		return ErrThemeTooLong
	}

	// Check for dangerous patterns (prompt injection prevention)
	lowerTheme := strings.ToLower(theme)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerTheme, pattern) {
			return ErrThemeInvalidChars
		}
	}

	// Allow alphanumeric, spaces, and common punctuation in any language
	// Reject control characters and some special symbols
	for _, r := range theme {
		if r < 32 && r != '\t' {
			return ErrThemeInvalidChars
		}
	}

	return nil
}

// SanitizeTheme sanitizes the theme for safe use in LLM prompts.
func SanitizeTheme(theme string) string {
	theme = strings.TrimSpace(theme)

	// Remove any potential prompt injection patterns
	for _, pattern := range dangerousPatterns {
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(pattern))
		theme = re.ReplaceAllString(theme, "")
	}

	// Remove control characters
	var result strings.Builder
	for _, r := range theme {
		if r >= 32 || r == '\t' {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// SetTheme sets and validates the theme.
func (c *GameConfig) SetTheme(theme string) error {
	if c.frozen {
		return ErrConfigFrozen
	}
	if err := ValidateTheme(theme); err != nil {
		return err
	}
	c.Theme = SanitizeTheme(theme)
	return nil
}

// SetDifficulty sets the difficulty level.
func (c *GameConfig) SetDifficulty(d DifficultyLevel) error {
	if c.frozen {
		return ErrConfigFrozen
	}
	if d < DifficultyEasy || d > DifficultyHell {
		return ErrInvalidDifficulty
	}
	c.Difficulty = d
	return nil
}

// SetLength sets the game length.
func (c *GameConfig) SetLength(l GameLength) error {
	if c.frozen {
		return ErrConfigFrozen
	}
	if l < LengthShort || l > LengthLong {
		return ErrInvalidLength
	}
	c.Length = l
	return nil
}

// SetAdultMode sets the adult mode flag.
func (c *GameConfig) SetAdultMode(enabled bool) error {
	if c.frozen {
		return ErrConfigFrozen
	}
	c.AdultMode = enabled
	return nil
}

// Freeze makes the config immutable (called when game starts).
func (c *GameConfig) Freeze() {
	c.frozen = true
}

// IsFrozen returns whether the config is frozen.
func (c *GameConfig) IsFrozen() bool {
	return c.frozen
}

// Validate validates the entire configuration.
func (c *GameConfig) Validate() error {
	if err := ValidateTheme(c.Theme); err != nil {
		return err
	}
	if c.Difficulty < DifficultyEasy || c.Difficulty > DifficultyHell {
		return ErrInvalidDifficulty
	}
	if c.Length < LengthShort || c.Length > LengthLong {
		return ErrInvalidLength
	}
	return nil
}

// IsComplete checks if all required fields are set.
func (c *GameConfig) IsComplete() bool {
	return c.Theme != "" && c.Validate() == nil
}

// ToJSON serializes the config to JSON.
func (c *GameConfig) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// FromJSON deserializes the config from JSON.
func (c *GameConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}

// Clone creates a deep copy of the config.
func (c *GameConfig) Clone() *GameConfig {
	return &GameConfig{
		Theme:      c.Theme,
		Difficulty: c.Difficulty,
		Length:     c.Length,
		AdultMode:  c.AdultMode,
		CreatedAt:  c.CreatedAt,
	}
}

// GameConfigBuilder provides a builder pattern for GameConfig.
type GameConfigBuilder struct {
	config *GameConfig
	errors []error
}

// NewGameConfigBuilder creates a new builder.
func NewGameConfigBuilder() *GameConfigBuilder {
	return &GameConfigBuilder{
		config: NewGameConfig(),
		errors: make([]error, 0),
	}
}

// WithTheme sets the theme.
func (b *GameConfigBuilder) WithTheme(theme string) *GameConfigBuilder {
	if err := b.config.SetTheme(theme); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// WithDifficulty sets the difficulty.
func (b *GameConfigBuilder) WithDifficulty(d DifficultyLevel) *GameConfigBuilder {
	if err := b.config.SetDifficulty(d); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// WithLength sets the game length.
func (b *GameConfigBuilder) WithLength(l GameLength) *GameConfigBuilder {
	if err := b.config.SetLength(l); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// WithAdultMode sets the adult mode flag.
func (b *GameConfigBuilder) WithAdultMode(enabled bool) *GameConfigBuilder {
	_ = b.config.SetAdultMode(enabled)
	return b
}

// Build returns the built config and any errors.
func (b *GameConfigBuilder) Build() (*GameConfig, error) {
	if len(b.errors) > 0 {
		return nil, b.errors[0]
	}
	b.config.CreatedAt = time.Now()
	return b.config, nil
}

// Errors returns all accumulated errors.
func (b *GameConfigBuilder) Errors() []error {
	return b.errors
}
