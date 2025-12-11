# Story 2.3: Player Input Handling

Status: Ready for Review

## Story

As a player,
I want to interact with the game through choice selection or free text input,
so that my decisions can influence the story development.

## Acceptance Criteria

1. **Number Key Choice Selection**
   - Given the game presents 2-4 choices
   - When the player presses a number key (1-4)
   - Then the corresponding choice is selected within 100ms
   - And visual feedback highlights the selected choice
   - And the selection triggers story progression

2. **Free Text Input Mode**
   - Given the player wants to perform a custom action
   - When the player presses 'f' or types text without number prefix
   - Then enter free text input mode with input field
   - And allow multi-line input (max 200 characters)
   - And display character counter
   - And Enter submits the custom action

3. **Slash Command Parsing**
   - Given the player types text starting with '/'
   - When input is submitted
   - Then parse as command instead of story action
   - And execute corresponding command (see Story 2-6)
   - And show command feedback without advancing story

4. **Enter Key Default Selection**
   - Given choices are displayed
   - When the player presses Enter without number
   - Then automatically select choice #1 (default/safe option)
   - And show "[Default]" indicator on choice #1

5. **Input Validation and Sanitization**
   - Given any player input is submitted
   - When processing the input
   - Then validate length (3-200 chars for free text)
   - And sanitize for LLM injection safety
   - And block empty or whitespace-only input
   - And show inline error for invalid input

6. **Input Mode State Management**
   - Given the player is in any input mode
   - When switching between choice/free text/command modes
   - Then preserve the current story context
   - And clear previous input buffers
   - And update UI to reflect current mode
   - And ESC key cancels free text mode

## Tasks / Subtasks

- [x] Create input handler architecture (AC: 1-6)
  - [x] Define InputMode enum (ChoiceSelect, FreeText, Command)
  - [x] Implement InputHandler interface
  - [x] Build mode state machine with transitions
  - [x] Add mode-specific validation logic

- [x] Implement number key choice selection (AC: 1, 4)
  - [x] Capture keyboard events for 1-4 keys
  - [x] Map key to choice index
  - [x] Add visual highlight animation (<100ms)
  - [x] Trigger choice submission event
  - [x] Implement Enter key default selection

- [x] Build free text input component (AC: 2)
  - [x] Create text input widget with BubbleTea textarea
  - [x] Add character counter (current/max 200)
  - [x] Support multi-line input (2-3 lines max visible)
  - [x] Implement 'f' key to enter mode
  - [x] Add Enter to submit, ESC to cancel

- [x] Implement slash command detection (AC: 3)
  - [x] Create command parser for '/' prefix
  - [x] Extract command name and arguments
  - [x] Route to command handler (Story 2-6)
  - [x] Return command results to UI
  - [x] Prevent story progression on commands

- [x] Add input validation pipeline (AC: 5)
  - [x] Create validation chain (length, content, safety)
  - [x] Implement LLM injection sanitization (escape special chars)
  - [x] Add profanity filter (optional based on game rating)
  - [x] Build error message generator
  - [x] Add retry prompt on validation failure

- [x] Create input state manager (AC: 6)
  - [x] Track current InputMode in game state
  - [x] Implement mode transition guards
  - [x] Clear buffers on mode change
  - [x] Preserve story context across modes
  - [x] Handle ESC/back navigation

- [x] Build input UI components (AC: 1, 2, 4)
  - [x] Design choice list with number indicators
  - [x] Create free text input box with border
  - [x] Add mode indicator in status bar
  - [x] Implement visual feedback for selection
  - [x] Style default choice indicator

## Dev Notes

### Architecture Pattern

- Use **State Machine Pattern** for InputMode transitions
- Implement **Chain of Responsibility** for validation pipeline
- Apply **Command Pattern** for slash commands (deferred to Story 2-6)
- Use **Observer Pattern** to notify UI of input changes

### Technical Constraints

- Input response time: < 100ms for number key selection
- Free text max length: 200 characters (to preserve LLM context budget)
- Must support terminal resize during input without losing data
- All inputs must be non-blocking to UI render loop

### Input Sanitization Rules

```go
type InputSanitizer struct {
    MaxLength      int
    MinLength      int
    AllowedChars   *regexp.Regexp
    BlockedPatterns []string // LLM injection patterns
}

// Sanitization steps:
// 1. Trim whitespace
// 2. Check length bounds
// 3. Escape special chars: <, >, {, }, [, ]
// 4. Block patterns: "ignore previous", "system:", etc.
// 5. Normalize unicode
```

### Data Structures

```go
type InputMode int
const (
    InputModeChoice InputMode = iota
    InputModeFreeText
    InputModeCommand
)

type PlayerInput struct {
    Mode      InputMode
    Raw       string
    Sanitized string
    Timestamp time.Time
    Valid     bool
    Error     error
}

type InputHandler struct {
    currentMode InputMode
    buffer      string
    validator   *InputSanitizer
    eventBus    *EventBus
}
```

### File Structure

- `internal/tui/input/handler.go` - Main input handler
- `internal/tui/input/validator.go` - Validation logic
- `internal/tui/input/sanitizer.go` - Sanitization rules
- `internal/tui/components/choice_list.go` - Choice UI component
- `internal/tui/components/text_input.go` - Free text component

### Key Bindings

```
Number keys (1-4): Select choice
Enter:             Submit choice #1 (default) or submit free text
f:                 Enter free text mode
/:                 Command mode (auto-detected)
ESC:               Cancel free text/command, return to choice mode
Ctrl+C:            Quit game (global)
```

### Performance Targets

- Number key response: < 100ms (AC requirement)
- Input validation: < 50ms
- Mode transition: < 30ms
- UI feedback animation: 150-200ms

### References

- [Source: docs/epics.md#Epic-2]
- [Related: Story 2-2 (receives input for story progression)]
- [Related: Story 2-6 (slash command execution)]
- [Related: Story 2-5 (input area in layout)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Implementation Notes

- Created `internal/tui/input/handler.go` with InputHandler, PlayerInput, and InputMode
- Implemented `internal/tui/input/sanitizer.go` with LLM injection protection
- Built `internal/tui/components/choice_list.go` for number key selection with visual feedback
- Built `internal/tui/components/text_input.go` for multi-line free text input with character counter
- Added comprehensive unit tests for all input handling components
- All validation rules implemented: length check, blocked patterns, special char escaping

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for implementation with complete AC and task breakdown
- Includes key bindings and sanitization rules for security
- Implementation completed 2025-12-11 with all ACs satisfied
- All unit tests passing (input and components packages)
- Build successful, no regressions

## File List

- `internal/tui/input/handler.go` (new) - Input handler with mode state machine
- `internal/tui/input/handler_test.go` (new) - Handler tests
- `internal/tui/input/sanitizer.go` (new) - Input sanitization and validation
- `internal/tui/input/sanitizer_test.go` (new) - Sanitizer tests
- `internal/tui/components/choice_list.go` (new) - Choice selection UI component
- `internal/tui/components/choice_list_test.go` (new) - Choice list tests
- `internal/tui/components/text_input.go` (new) - Free text input UI component
- `internal/tui/components/text_input_test.go` (new) - Text input tests

## Change Log

- 2025-12-11: Implemented input handler with three modes (Choice, FreeText, Command)
- 2025-12-11: Created input sanitizer with LLM injection protection
- 2025-12-11: Built choice list component with number key selection
- 2025-12-11: Built free text input component with BubbleTea textarea integration
- 2025-12-11: Added comprehensive validation pipeline (length, patterns, escaping)
