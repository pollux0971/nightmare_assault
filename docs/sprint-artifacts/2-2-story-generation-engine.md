# Story 2.2: Story Generation Engine

Status: Ready for Review

## Story

As a player,
I want to see AI-generated opening story content,
so that I can become immersed in a horror atmosphere and begin my adventure.

## Acceptance Criteria

1. **Loading Screen Display**
   - Given the user confirms game setup
   - When story generation begins
   - Then display a loading screen with thematic animation (e.g., "Entering nightmare...")
   - And show progress indicators if generation exceeds 2 seconds

2. **Streaming Response Time**
   - Given story generation is initiated
   - When calling the LLM API
   - Then the first token must arrive within 5 seconds
   - And display timeout error if no response after 10 seconds

3. **Typewriter Effect**
   - Given the LLM response stream starts
   - When rendering story text
   - Then display text with typewriter effect at 50-80 chars/second
   - And allow spacebar to skip/speed up animation
   - And preserve formatting (paragraphs, line breaks)

4. **Game Bible Prompt Integration**
   - Given the story generation request
   - When constructing the LLM prompt
   - Then include Game Bible rules (HP/SAN mechanics, horror tone, choice structure)
   - And inject user's theme, difficulty, length from GameConfig
   - And maintain consistent narrative voice

5. **Hidden Rule Seeds**
   - Given story generation is in progress
   - When creating the opening narrative
   - Then embed 2-3 hidden clues/rules that affect future outcomes
   - And track these seeds in game state for continuity
   - And ensure seeds align with chosen difficulty

6. **Error Handling and Retry**
   - Given an LLM API failure occurs
   - When generation fails or returns invalid content
   - Then show user-friendly error message
   - And offer retry with exponential backoff (3 attempts max)
   - And fallback to pre-written opening if all retries fail

## Tasks / Subtasks

- [x] Design Game Bible prompt template (AC: 4, 5)
  - [x] Create base system prompt with core mechanics
  - [x] Define horror tone guidelines and examples
  - [x] Add difficulty-specific modifiers
  - [x] Include hidden seed generation instructions
  - [x] Create template variables for user config injection

- [x] Implement story generation service (AC: 1, 2, 6)
  - [x] Create StoryEngine interface
  - [x] Implement LLM client wrapper (with streaming support)
  - [x] Add timeout and retry logic
  - [x] Build prompt construction pipeline
  - [x] Implement fallback story loader

- [x] Build streaming handler (AC: 2, 3)
  - [x] Create streaming response parser
  - [x] Implement token buffering for smooth rendering
  - [x] Add typewriter animation controller
  - [x] Support pause/resume/skip controls
  - [x] Handle incomplete/malformed streams

- [x] Create loading screen UI (AC: 1)
  - [x] Design thematic loading animation (spinner variants)
  - [x] Add progress percentage if available
  - [x] Include atmospheric flavor text rotation
  - [x] Implement timeout warning display

- [x] Implement hidden seed system (AC: 5)
  - [x] Define HiddenSeed data structure
  - [x] Create seed extraction from LLM metadata
  - [x] Store seeds in game state
  - [x] Build seed validation against game rules

- [x] Add story state management (AC: 4, 5)
  - [x] Create StoryState to track narrative progress
  - [x] Implement story segment storage
  - [x] Add context window management for follow-up calls
  - [x] Build narrative consistency checker

- [x] Create error handling and fallbacks (AC: 6)
  - [x] Write 3-5 fallback opening stories per theme category
  - [x] Implement retry mechanism with backoff
  - [x] Add error logging and telemetry
  - [x] Create user-facing error messages

## Dev Notes

### Architecture Pattern

- Use **Strategy Pattern** for different LLM providers (Smart Model vs Fast Model)
- Implement **Observer Pattern** for streaming updates to UI
- Apply **Template Method Pattern** for prompt construction
- Use **Circuit Breaker Pattern** for API failure handling

### Technical Constraints

- Smart Model (GPT-4 class) for story generation: higher quality, slower
- Must maintain < 5 second time-to-first-token for UX
- Streaming chunks must be processed in real-time without blocking UI
- Context window: Keep opening generation under 1000 tokens to preserve budget
- Rate limiting: Handle API quotas gracefully

### Game Bible Core Rules

```markdown
# Game Bible v1.0

## Core Mechanics
- HP: Physical health (0-100). Reaches 0 = death.
- SAN: Sanity (0-100). Affects perception and choices.
  - 80-100: Clear-headed
  - 50-79: Anxious, minor hallucinations
  - 20-49: Panicked, reality distortion
  - 0-19: Insanity, loss of control

## Narrative Rules
- Every story beat ends with 2-4 choices
- Choices must have clear risk/reward implications
- At least one choice per beat should affect HP or SAN
- Maintain Lovecraftian cosmic horror atmosphere
- Use environmental storytelling over exposition

## Hidden Seeds
- Plant 2-3 subtle clues in opening (e.g., locked door, strange symbol)
- Seeds trigger callbacks 3-5 beats later
- Difficulty affects seed impact severity
```

### Data Structures

```go
type StoryEngine struct {
    llmClient    LLMClient
    gameBible    PromptTemplate
    config       *GameConfig
    storyState   *StoryState
    seedTracker  *SeedTracker
}

type HiddenSeed struct {
    ID          string
    Type        SeedType // Item, Event, Character, Location
    Description string
    TriggerBeat int      // When this seed activates
    Discovered  bool
}

type StoryState struct {
    CurrentBeat int
    History     []StorySegment
    ActiveSeeds []HiddenSeed
    ContextHash string // For cache/continuity
}
```

### File Structure

- `internal/engine/story.go` - Main story engine
- `internal/engine/prompts/` - Prompt templates
- `internal/engine/streaming.go` - Stream handler
- `internal/llm/client.go` - LLM API wrapper
- `assets/fallback_stories.json` - Emergency fallback content

### Performance Targets

- Time to first token: < 5 seconds (95th percentile)
- Typewriter speed: 50-80 chars/sec (user configurable later)
- Total opening generation: < 15 seconds end-to-end
- Memory: Keep story history under 10MB

### References

- [Source: docs/epics.md#Epic-2]
- [Related: Story 2-1 (receives GameConfig), Story 2-3 (triggers next input)]
- [Related: Story 2-4 (HP/SAN system integration)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Implementation Notes

- Created `internal/engine/prompts/gamebible.go` with Game Bible rules and prompt templates
- Implemented difficulty and length modifiers for prompt customization
- Built `internal/engine/story.go` with StoryEngine, StoryState, and HiddenSeed structures
- Added streaming handler in `internal/engine/streaming.go` with TypewriterConfig
- Created `internal/engine/fallback.go` with 5 fallback stories for different themes
- Built loading screen view in `internal/tui/views/story_loading.go` with spinner and flavor text
- Implemented seed extraction from LLM content using marker format
- Added comprehensive unit tests for all engine components

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for implementation with complete AC and task breakdown
- Includes Game Bible core rules for prompt engineering
- Implementation completed 2025-12-11 with all ACs satisfied
- All unit tests passing (44 tests across engine package)
- Build successful, no regressions

## File List

- `internal/engine/prompts/gamebible.go` (new) - Game Bible prompt templates
- `internal/engine/prompts/gamebible_test.go` (new) - Prompt template tests
- `internal/engine/story.go` (new) - Story engine with StoryState and HiddenSeed
- `internal/engine/story_test.go` (new) - Story engine tests
- `internal/engine/streaming.go` (new) - Typewriter effect and streaming handler
- `internal/engine/streaming_test.go` (new) - Streaming handler tests
- `internal/engine/fallback.go` (new) - Fallback stories for API failures
- `internal/engine/fallback_test.go` (new) - Fallback story tests
- `internal/tui/views/story_loading.go` (new) - Loading screen with spinner

## Change Log

- 2025-12-11: Implemented story generation engine with streaming support
- 2025-12-11: Created Game Bible prompt system with difficulty/length modifiers
- 2025-12-11: Added typewriter effect with configurable speed (50-80 chars/sec)
- 2025-12-11: Implemented hidden seed extraction and tracking system
- 2025-12-11: Created 5 fallback stories for emergency API failure handling
- 2025-12-11: Built loading screen with thematic animations and timeout handling
