// Package game provides debrief data collection and management for Nightmare Assault.
package game

import (
	"time"
)

// ClueStatus represents whether a clue was discovered or missed.
type ClueStatus int

const (
	// ClueStatusMissed means the player didn't notice the clue.
	ClueStatusMissed ClueStatus = iota
	// ClueStatusDiscovered means the player noticed the clue.
	ClueStatusDiscovered
)

// String returns the display name of the clue status.
func (c ClueStatus) String() string {
	switch c {
	case ClueStatusDiscovered:
		return "已發現"
	case ClueStatusMissed:
		return "錯過"
	default:
		return "未知"
	}
}

// ClueInfo contains information about a clue in the game.
type ClueInfo struct {
	// ID is the unique identifier for this clue
	ID string `json:"id"`
	// Content is the clue text
	Content string `json:"content"`
	// Chapter where the clue appeared
	Chapter int `json:"chapter"`
	// Location in the narrative (paragraph/context)
	Location string `json:"location"`
	// RuleID is the rule this clue hints at
	RuleID string `json:"rule_id"`
	// Status indicates if discovered or missed
	Status ClueStatus `json:"status"`
	// Timestamp when the clue appeared
	Timestamp time.Time `json:"timestamp"`
	// Context is the surrounding narrative text
	Context string `json:"context,omitempty"`
}

// NewClueInfo creates a new clue info.
func NewClueInfo(id, content string, ruleID string) *ClueInfo {
	return &ClueInfo{
		ID:        id,
		Content:   content,
		RuleID:    ruleID,
		Status:    ClueStatusMissed, // Default to missed
		Timestamp: time.Now(),
	}
}

// HallucinationLog records when a player selected a hallucination option.
type HallucinationLog struct {
	// OptionText is the text of the hallucination option
	OptionText string `json:"option_text"`
	// SANValue is the player's SAN when they selected it
	SANValue int `json:"san_value"`
	// Chapter where this occurred
	Chapter int `json:"chapter"`
	// Timestamp of the selection
	Timestamp time.Time `json:"timestamp"`
	// Consequence describes what happened after selection
	Consequence string `json:"consequence"`
	// RealOption is what the real option would have been
	RealOption string `json:"real_option,omitempty"`
}

// NewHallucinationLog creates a new hallucination log entry.
func NewHallucinationLog(optionText string, sanValue int, chapter int) *HallucinationLog {
	return &HallucinationLog{
		OptionText: optionText,
		SANValue:   sanValue,
		Chapter:    chapter,
		Timestamp:  time.Now(),
	}
}

// DecisionPoint represents a key decision made during gameplay.
type DecisionPoint struct {
	// Chapter where the decision was made
	Chapter int `json:"chapter"`
	// Timestamp of the decision
	Timestamp time.Time `json:"timestamp"`
	// Options available to the player
	Options []string `json:"options"`
	// SelectedIndex is which option was chosen
	SelectedIndex int `json:"selected_index"`
	// SelectedText is the text of the chosen option
	SelectedText string `json:"selected_text"`
	// IsSignificant marks decisions that led to major consequences
	IsSignificant bool `json:"is_significant"`
	// Consequence describes what happened after this decision
	Consequence string `json:"consequence,omitempty"`
	// IsHallucination marks if this was a hallucination choice
	IsHallucination bool `json:"is_hallucination"`
}

// NewDecisionPoint creates a new decision point.
func NewDecisionPoint(chapter int, options []string, selectedIndex int) *DecisionPoint {
	selectedText := ""
	if selectedIndex >= 0 && selectedIndex < len(options) {
		selectedText = options[selectedIndex]
	}
	return &DecisionPoint{
		Chapter:       chapter,
		Timestamp:     time.Now(),
		Options:       options,
		SelectedIndex: selectedIndex,
		SelectedText:  selectedText,
	}
}

// CheckpointInfo represents a checkpoint for the easy mode rollback feature.
type CheckpointInfo struct {
	// ID is the unique identifier for this checkpoint
	ID string `json:"id"`
	// Chapter at the checkpoint
	Chapter int `json:"chapter"`
	// Timestamp when checkpoint was created
	Timestamp time.Time `json:"timestamp"`
	// HP at checkpoint
	HP int `json:"hp"`
	// SAN at checkpoint
	SAN int `json:"san"`
	// SavePath is the path to the checkpoint save file
	SavePath string `json:"save_path,omitempty"`
	// Description of the checkpoint location
	Description string `json:"description"`
}

// NewCheckpointInfo creates a new checkpoint info.
func NewCheckpointInfo(id string, chapter int, hp, san int) *CheckpointInfo {
	return &CheckpointInfo{
		ID:        id,
		Chapter:   chapter,
		Timestamp: time.Now(),
		HP:        hp,
		SAN:       san,
	}
}

// RuleReveal contains information for revealing a rule during debrief.
type RuleReveal struct {
	// RuleID is the ID of the revealed rule
	RuleID string `json:"rule_id"`
	// RuleType is the type of rule (場景/時間/行為/對象/狀態)
	RuleType string `json:"rule_type"`
	// TriggerCondition describes what triggers the rule
	TriggerCondition string `json:"trigger_condition"`
	// ConsequenceType is the consequence (警告/傷害/即死)
	ConsequenceType string `json:"consequence_type"`
	// ConsequenceDetail describes the specific consequence
	ConsequenceDetail string `json:"consequence_detail"`
	// DiscoveredClues are clues the player found
	DiscoveredClues []string `json:"discovered_clues"`
	// MissedClues are clues the player missed
	MissedClues []string `json:"missed_clues"`
	// Explanation is LLM-generated text explaining the rule and clues
	Explanation string `json:"explanation,omitempty"`
	// ViolationCount is how many times the rule was violated
	ViolationCount int `json:"violation_count"`
}

// NewRuleReveal creates a new rule reveal.
func NewRuleReveal(ruleID, ruleType, trigger, consequence string) *RuleReveal {
	return &RuleReveal{
		RuleID:           ruleID,
		RuleType:         ruleType,
		TriggerCondition: trigger,
		ConsequenceType:  consequence,
		DiscoveredClues:  make([]string, 0),
		MissedClues:      make([]string, 0),
	}
}

// DebriefData contains all data needed for the death debrief screen.
type DebriefData struct {
	// DeathInfo contains death details
	DeathInfo *DeathInfo `json:"death_info"`
	// TriggeredRules are rules that were violated
	TriggeredRules []*RuleReveal `json:"triggered_rules"`
	// AllClues are all clues encountered during gameplay
	AllClues []*ClueInfo `json:"all_clues"`
	// HallucinationLogs are all hallucination choices made
	HallucinationLogs []*HallucinationLog `json:"hallucination_logs"`
	// KeyDecisions are significant decision points
	KeyDecisions []*DecisionPoint `json:"key_decisions"`
	// Checkpoints are available rollback points (easy mode only)
	Checkpoints []*CheckpointInfo `json:"checkpoints,omitempty"`
	// Difficulty level (affects available options)
	Difficulty DifficultyLevel `json:"difficulty"`
}

// NewDebriefData creates an empty debrief data container.
func NewDebriefData() *DebriefData {
	return &DebriefData{
		TriggeredRules:    make([]*RuleReveal, 0),
		AllClues:          make([]*ClueInfo, 0),
		HallucinationLogs: make([]*HallucinationLog, 0),
		KeyDecisions:      make([]*DecisionPoint, 0),
		Checkpoints:       make([]*CheckpointInfo, 0),
	}
}

// SetDeathInfo sets the death information.
func (d *DebriefData) SetDeathInfo(info *DeathInfo) {
	d.DeathInfo = info
}

// AddRuleReveal adds a revealed rule to the debrief.
func (d *DebriefData) AddRuleReveal(reveal *RuleReveal) {
	d.TriggeredRules = append(d.TriggeredRules, reveal)
}

// AddClue adds a clue to the tracking list.
func (d *DebriefData) AddClue(clue *ClueInfo) {
	d.AllClues = append(d.AllClues, clue)
}

// MarkClueDiscovered marks a clue as discovered by the player.
func (d *DebriefData) MarkClueDiscovered(clueID string) bool {
	for _, clue := range d.AllClues {
		if clue.ID == clueID {
			clue.Status = ClueStatusDiscovered
			return true
		}
	}
	return false
}

// GetMissedClues returns all missed clues.
func (d *DebriefData) GetMissedClues() []*ClueInfo {
	var missed []*ClueInfo
	for _, clue := range d.AllClues {
		if clue.Status == ClueStatusMissed {
			missed = append(missed, clue)
		}
	}
	return missed
}

// GetDiscoveredClues returns all discovered clues.
func (d *DebriefData) GetDiscoveredClues() []*ClueInfo {
	var discovered []*ClueInfo
	for _, clue := range d.AllClues {
		if clue.Status == ClueStatusDiscovered {
			discovered = append(discovered, clue)
		}
	}
	return discovered
}

// GetCluesByRuleID returns clues for a specific rule.
func (d *DebriefData) GetCluesByRuleID(ruleID string) []*ClueInfo {
	var clues []*ClueInfo
	for _, clue := range d.AllClues {
		if clue.RuleID == ruleID {
			clues = append(clues, clue)
		}
	}
	return clues
}

// AddHallucinationLog adds a hallucination log entry.
func (d *DebriefData) AddHallucinationLog(log *HallucinationLog) {
	d.HallucinationLogs = append(d.HallucinationLogs, log)
}

// GetHallucinationCount returns the number of hallucination choices made.
func (d *DebriefData) GetHallucinationCount() int {
	return len(d.HallucinationLogs)
}

// AddDecision adds a decision point.
func (d *DebriefData) AddDecision(decision *DecisionPoint) {
	d.KeyDecisions = append(d.KeyDecisions, decision)
}

// GetSignificantDecisions returns only significant decisions.
func (d *DebriefData) GetSignificantDecisions() []*DecisionPoint {
	var significant []*DecisionPoint
	for _, dec := range d.KeyDecisions {
		if dec.IsSignificant {
			significant = append(significant, dec)
		}
	}
	return significant
}

// AddCheckpoint adds a checkpoint (easy mode only).
func (d *DebriefData) AddCheckpoint(checkpoint *CheckpointInfo) {
	d.Checkpoints = append(d.Checkpoints, checkpoint)
	// Keep only the last 3 checkpoints
	if len(d.Checkpoints) > 3 {
		d.Checkpoints = d.Checkpoints[len(d.Checkpoints)-3:]
	}
}

// GetLatestCheckpoint returns the most recent checkpoint.
func (d *DebriefData) GetLatestCheckpoint() *CheckpointInfo {
	if len(d.Checkpoints) == 0 {
		return nil
	}
	return d.Checkpoints[len(d.Checkpoints)-1]
}

// CanRollback returns true if rollback is available (easy mode + checkpoints exist).
func (d *DebriefData) CanRollback() bool {
	return d.Difficulty == DifficultyEasy && len(d.Checkpoints) > 0
}

// ClearCheckpoints clears all checkpoints.
func (d *DebriefData) ClearCheckpoints() {
	d.Checkpoints = make([]*CheckpointInfo, 0)
}

// GetDeathSummary returns a formatted death summary.
func (d *DebriefData) GetDeathSummary() string {
	if d.DeathInfo == nil {
		return "死因不明"
	}

	switch d.DeathInfo.Type {
	case DeathTypeHP:
		return "你的體力完全耗盡，無法再繼續前進。"
	case DeathTypeSAN:
		return "你的理智崩潰了，被恐懼和瘋狂吞噬。"
	case DeathTypeRule:
		if len(d.TriggeredRules) > 0 {
			return "你違反了隱藏的規則：「" + d.TriggeredRules[0].TriggerCondition + "」"
		}
		return "你違反了隱藏的規則而遭受懲罰。"
	default:
		return "你的冒險在此結束。"
	}
}

// DebriefCollector handles collecting debrief data during gameplay.
type DebriefCollector struct {
	data *DebriefData
}

// NewDebriefCollector creates a new debrief collector.
func NewDebriefCollector() *DebriefCollector {
	return &DebriefCollector{
		data: NewDebriefData(),
	}
}

// GetData returns the collected debrief data.
func (c *DebriefCollector) GetData() *DebriefData {
	return c.data
}

// SetDifficulty sets the difficulty level.
func (c *DebriefCollector) SetDifficulty(difficulty DifficultyLevel) {
	c.data.Difficulty = difficulty
}

// RecordClue records a clue appearing in the game.
func (c *DebriefCollector) RecordClue(clue *ClueInfo) {
	c.data.AddClue(clue)
}

// RecordClueDiscovered marks a clue as discovered.
func (c *DebriefCollector) RecordClueDiscovered(clueID string) {
	c.data.MarkClueDiscovered(clueID)
}

// RecordHallucination records a hallucination choice.
func (c *DebriefCollector) RecordHallucination(log *HallucinationLog) {
	c.data.AddHallucinationLog(log)
}

// RecordDecision records a decision point.
func (c *DebriefCollector) RecordDecision(decision *DecisionPoint) {
	c.data.AddDecision(decision)
}

// RecordCheckpoint records a checkpoint (easy mode).
func (c *DebriefCollector) RecordCheckpoint(checkpoint *CheckpointInfo) {
	c.data.AddCheckpoint(checkpoint)
}

// RecordRuleViolation records a rule violation for debrief.
func (c *DebriefCollector) RecordRuleViolation(reveal *RuleReveal) {
	c.data.AddRuleReveal(reveal)
}

// RecordDeath records the death information.
func (c *DebriefCollector) RecordDeath(info *DeathInfo) {
	c.data.SetDeathInfo(info)
}

// Reset clears all collected data for a new game.
func (c *DebriefCollector) Reset() {
	c.data = NewDebriefData()
}
