# Story 2.1: New Game Setup Flow

Status: Ready for Review

## Story

As a player,
I want to configure new game parameters,
so that I can customize the story theme, difficulty, and length.

## Acceptance Criteria

1. **Story Theme Input**
   - Given the user starts a new game
   - When prompted for a story theme
   - Then the user can input a custom horror theme (e.g., "abandoned hospital", "cursed mansion")
   - And the theme must be between 3-100 characters

2. **Difficulty Selection**
   - Given the user has entered a theme
   - When presented with difficulty options
   - Then the user can select from: Easy (1), Hard (2), Hell (3)
   - And each difficulty shows impact description (HP/SAN drain rate, enemy strength)

3. **Length Selection**
   - Given the user has selected difficulty
   - When presented with length options
   - Then the user can select from: Short (~15 min), Medium (~30 min), Long (~60 min)
   - And the selection affects story arc complexity and event count

4. **18+ Mode Toggle**
   - Given the user has selected length
   - When prompted for content rating
   - Then the user can enable/disable 18+ mode (explicit gore, psychological horror)
   - And the default is OFF with clear warning for ON

5. **Configuration Summary**
   - Given all parameters are set
   - When the user confirms setup
   - Then display a summary screen with all choices
   - And allow edit before final confirmation

## Tasks / Subtasks

- [x] Create GameConfig structure (AC: 1-5)
  - [x] Define config struct with theme, difficulty, length, adult_mode fields
  - [x] Add validation methods for each field
  - [x] Implement JSON serialization for save/load

- [x] Implement setup flow state machine (AC: 1-5)
  - [x] Create SetupState enum (ThemeInput, DifficultySelect, LengthSelect, AdultModeToggle, Summary)
  - [x] Build state transition logic
  - [x] Handle back navigation between steps

- [x] Build TUI setup screens (AC: 1-5)
  - [x] Design theme input screen with text input component
  - [x] Create difficulty selection with radio buttons/number keys
  - [x] Create length selection screen
  - [x] Create 18+ toggle with warning modal
  - [x] Design summary screen with styled output

- [x] Add input validation and error handling (AC: 1, 5)
  - [x] Validate theme length and characters
  - [x] Sanitize user input for LLM safety
  - [x] Show inline error messages
  - [x] Prevent progression with invalid input

- [x] Create difficulty presets (AC: 2)
  - [x] Define Easy: HP drain 0.5x, SAN drain 0.7x, hints enabled
  - [x] Define Hard: HP drain 1.0x, SAN drain 1.0x, standard
  - [x] Define Hell: HP drain 1.5x, SAN drain 1.3x, permadeath mode

- [x] Implement setup persistence (AC: 5)
  - [x] Save configuration to game state
  - [x] Pass config to story generation engine
  - [x] Store config in save file for continue game

## Dev Notes

### Architecture Pattern

- Use **State Machine Pattern** for setup flow progression
- Implement **Builder Pattern** for GameConfig construction
- Apply **Validation Chain** for input verification

### Technical Constraints

- Setup flow must complete in < 60 seconds for UX
- All user inputs must be sanitized before LLM prompt injection
- Configuration must be immutable after game start
- Support Ctrl+C to cancel setup and return to main menu

### Data Structures

```go
type GameConfig struct {
    Theme      string
    Difficulty DifficultyLevel
    Length     GameLength
    AdultMode  bool
    CreatedAt  time.Time
}

type DifficultyLevel int
const (
    DifficultyEasy DifficultyLevel = iota
    DifficultyHard
    DifficultyHell
)

type GameLength int
const (
    LengthShort  GameLength = iota // ~15 min
    LengthMedium                    // ~30 min
    LengthLong                      // ~60 min
)
```

### File Structure

- `internal/game/setup.go` - Core setup logic
- `internal/game/config.go` - GameConfig definition
- `internal/tui/views/setup.go` - Setup TUI screens
- `internal/tui/components/input.go` - Reusable input components

### References

- [Source: docs/epics.md#Epic-2]
- [Related: Story 2-2 (receives config), Story 2-5 (layout standards)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Implementation Notes

- Created `internal/game/config.go` with GameConfig struct, DifficultyLevel, and GameLength enums
- Implemented comprehensive validation including LLM prompt injection prevention
- Built Builder pattern for type-safe GameConfig construction
- Created `internal/tui/views/game_setup.go` with state machine for 5-step setup flow
- Added StateGameSetup to app.go with full integration
- All difficulty presets implemented with HP/SAN multipliers, hints, and permadeath flags
- Theme sanitization removes dangerous patterns for LLM safety

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for implementation with complete AC and task breakdown
- Implementation completed 2025-12-11 with all ACs satisfied
- All unit tests passing (17 tests for config, 16 tests for game_setup view)
- Build successful, no regressions

## File List

- `internal/game/config.go` (new) - GameConfig structure with validation
- `internal/game/config_test.go` (new) - Unit tests for GameConfig
- `internal/tui/views/game_setup.go` (new) - Setup flow TUI view
- `internal/tui/views/game_setup_test.go` (new) - Unit tests for setup view
- `internal/app/app.go` (modified) - Added StateGameSetup and integration

## Change Log

- 2025-12-11: Implemented new game setup flow with 5-step wizard (theme, difficulty, length, 18+ mode, summary)
- 2025-12-11: Added GameConfig structure with validation and JSON serialization
- 2025-12-11: Created difficulty presets (Easy/Hard/Hell) with HP/SAN multipliers
- 2025-12-11: Added comprehensive unit tests for all new functionality
