// Package game provides core game logic and state management.
package game

import (
	"math/rand"
	"time"
)

// HallucinationRecord represents a single hallucination event in the game.
// This is used for tracking and debrief display (AC8).
type HallucinationRecord struct {
	TurnNumber      int       // The turn number when this hallucination appeared
	SANValue        int       // Player's SAN value at the time
	OptionText      string    // The hallucinatory option text
	RealOptions     []string  // The actual valid options that were present
	WasSelected     bool      // Whether the player selected this hallucination
	ConsequenceDesc string    // Description of what happened if selected
	Timestamp       time.Time // When this occurred
}

// HallucinationTracker manages the collection of hallucination records.
type HallucinationTracker struct {
	Records []HallucinationRecord
}

// NewHallucinationTracker creates a new empty hallucination tracker.
func NewHallucinationTracker() *HallucinationTracker {
	return &HallucinationTracker{
		Records: make([]HallucinationRecord, 0),
	}
}

// AddRecord adds a new hallucination record to the tracker.
func (ht *HallucinationTracker) AddRecord(record HallucinationRecord) {
	ht.Records = append(ht.Records, record)
}

// GetRecords returns all hallucination records.
func (ht *HallucinationTracker) GetRecords() []HallucinationRecord {
	return ht.Records
}

// GetSelectedCount returns the number of hallucinations that were selected.
func (ht *HallucinationTracker) GetSelectedCount() int {
	count := 0
	for _, record := range ht.Records {
		if record.WasSelected {
			count++
		}
	}
	return count
}

// GetTotalCount returns the total number of hallucinations encountered.
func (ht *HallucinationTracker) GetTotalCount() int {
	return len(ht.Records)
}

// ShouldInsertHallucination determines if a hallucination option should be inserted.
// Based on AC1: probability = (20 - SAN) / 20 when SAN < 20.
//
// Returns true if a hallucination should be inserted based on current SAN.
func ShouldInsertHallucination(san int) bool {
	if san >= 20 {
		return false // No hallucinations when SAN >= 20
	}

	// Calculate probability: (20 - SAN) / 20
	// SAN 20 -> 0% probability
	// SAN 15 -> 25% probability
	// SAN 10 -> 50% probability
	// SAN 5  -> 75% probability
	// SAN 1  -> 95% probability
	probability := float64(20-san) / 20.0

	return rand.Float64() < probability
}

// HallucinationOption represents a hallucination option that can be inserted into choices.
type HallucinationOption struct {
	Text     string // The option text (appears identical to real options)
	Index    int    // The index where it was inserted in the options list
	SANValue int    // The SAN value when it was generated
}

// InsertHallucinationOption inserts a hallucination option into the real options list.
// Based on AC1: position is random, max 1 hallucination at a time.
//
// realOptions: The actual game options
// hallucinationText: The generated hallucination text
//
// Returns:
//   - The combined options list with hallucination inserted
//   - The index where the hallucination was inserted
func InsertHallucinationOption(realOptions []string, hallucinationText string) ([]string, int) {
	if len(realOptions) == 0 {
		// Edge case: no real options, just return the hallucination
		return []string{hallucinationText}, 0
	}

	// Random position for hallucination (AC1: random position, not always last)
	hallucinationIndex := rand.Intn(len(realOptions) + 1)

	// Insert hallucination at random position
	combined := make([]string, 0, len(realOptions)+1)
	combined = append(combined, realOptions[:hallucinationIndex]...)
	combined = append(combined, hallucinationText)
	combined = append(combined, realOptions[hallucinationIndex:]...)

	return combined, hallucinationIndex
}

// IsHallucinationIndex checks if the given selection index corresponds to a hallucination.
//
// selectedIndex: The index the player selected (0-based)
// hallucinationIndex: The index where the hallucination was inserted
//
// Returns true if the player selected the hallucination option.
func IsHallucinationIndex(selectedIndex, hallucinationIndex int) bool {
	return selectedIndex == hallucinationIndex
}

// HallucinationConsequence represents the result of selecting a hallucination.
type HallucinationConsequence struct {
	Description string // Narrative description of what happens
	SANLoss     int    // SAN loss amount (typically -5)
	IsDangerous bool   // Whether this leads to a dangerous situation
}

// GenerateHallucinationConsequence generates the consequence for selecting a hallucination.
// Based on AC4: immediate revelation + negative consequence (SAN -5 or danger).
func GenerateHallucinationConsequence(hallucinationText string) HallucinationConsequence {
	// For now, use template-based consequences
	// In full implementation, this would call Fast Model for dynamic generation

	revelations := []string{
		"你突然意識到...這不存在。",
		"當你伸手觸碰，它如煙霧般散去。",
		"你眨了眨眼，剛才看到的東西消失了。",
		"那個選項...從來不在那裡。",
		"你的手指穿過了空氣——什麼都沒有。",
	}

	revelation := revelations[rand.Intn(len(revelations))]

	// Randomly decide between SAN loss and dangerous situation
	isDangerous := rand.Float64() < 0.3 // 30% chance of dangerous situation

	consequence := HallucinationConsequence{
		Description: revelation,
		SANLoss:     5,
		IsDangerous: isDangerous,
	}

	if isDangerous {
		consequence.Description += " 當你恍神時，危險正在逼近..."
	} else {
		consequence.Description += " 你意識到自己的感官正在背叛你。"
	}

	return consequence
}

// HallucinationGenerator is responsible for generating hallucination option text.
// In full implementation, this would integrate with Fast Model LLM.
type HallucinationGenerator struct {
	// Future: Fast Model client
	// For now, use template-based generation
}

// NewHallucinationGenerator creates a new hallucination generator.
func NewHallucinationGenerator() *HallucinationGenerator {
	return &HallucinationGenerator{}
}

// Generate generates a hallucination option text.
// Based on AC3: "似是而非" - plausible but subtly wrong.
//
// context: Game context (scene, real options, player state)
//
// For now, this uses template-based generation. In full implementation,
// this would call Fast Model with a carefully crafted prompt.
func (hg *HallucinationGenerator) Generate(sceneDescription string, realOptions []string) string {
	// Template-based hallucination patterns
	// AC3: Should be plausible but involve non-existent items/people/exits

	templates := []string{
		"拿起桌上的手電筒",        // Item that doesn't exist
		"向左邊的門逃跑",         // Exit that doesn't exist
		"呼叫剛才遇到的同伴",       // Person who isn't there
		"使用口袋裡的火柴",        // Item player doesn't have
		"回到之前安全的房間",       // Location that changed
		"檢查窗戶是否能打開",       // Window that doesn't exist
		"撿起地上的鑰匙",         // Key that isn't there
		"打開你記得的那扇暗門",      // Secret door that's imagined
		"用手機照明",           // Phone you don't have/lost
		"躲進剛才看到的櫃子裡",      // Hiding spot that's gone
	}

	// Return random template
	// In full implementation, this would be replaced with LLM generation
	return templates[rand.Intn(len(templates))]
}

// GenerateWithContext generates a hallucination with full context.
// This is the main entry point for hallucination generation.
//
// Future implementation will use Fast Model with this prompt structure:
// - Current scene description
// - Real options available
// - Player psychological state
// - Generate option that is "似是而非" (plausible but false)
func (hg *HallucinationGenerator) GenerateWithContext(scene string, realOptions []string, playerSAN int) (string, error) {
	// For now, use simple generation
	// TODO: Replace with Fast Model integration when available
	hallucination := hg.Generate(scene, realOptions)

	// Ensure length is similar to real options (AC3: ±10 characters)
	// This validation would be done on LLM output in full implementation

	return hallucination, nil
}
