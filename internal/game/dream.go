package game

import (
	"time"
)

// DreamType represents the type of dream.
type DreamType string

const (
	DreamTypeOpening DreamType = "opening"
	DreamTypeChapter DreamType = "chapter"
)

// DreamRecord represents a single dream instance.
type DreamRecord struct {
	ID            string    `json:"id"`
	Type          DreamType `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	Content       string    `json:"content"`
	RelatedRuleID string    `json:"related_rule_id,omitempty"`
	Context       DreamContext `json:"context"`
}

// DreamContext stores the game context when the dream was generated.
type DreamContext struct {
	PlayerHP     int      `json:"player_hp"`
	PlayerSAN    int      `json:"player_san"`
	ChapterNum   int      `json:"chapter_num"`
	KnownClues   []string `json:"known_clues"`
	StoryTheme   string   `json:"story_theme"`
	RulesSummary string   `json:"rules_summary"`
}

// DreamLog tracks all dreams experienced in the game.
type DreamLog struct {
	Dreams []DreamRecord `json:"dreams"`
}

// NewDreamLog creates a new empty dream log.
func NewDreamLog() *DreamLog {
	return &DreamLog{
		Dreams: make([]DreamRecord, 0),
	}
}

// LogDream adds a dream record to the log.
func (dl *DreamLog) LogDream(dream DreamRecord) {
	dl.Dreams = append(dl.Dreams, dream)
}

// GetDreamsByType returns all dreams of a specific type.
func (dl *DreamLog) GetDreamsByType(dreamType DreamType) []DreamRecord {
	result := make([]DreamRecord, 0)
	for _, dream := range dl.Dreams {
		if dream.Type == dreamType {
			result = append(result, dream)
		}
	}
	return result
}

// GetLastDream returns the most recent dream, or nil if no dreams exist.
func (dl *DreamLog) GetLastDream() *DreamRecord {
	if len(dl.Dreams) == 0 {
		return nil
	}
	return &dl.Dreams[len(dl.Dreams)-1]
}

// DreamCount returns the total number of dreams.
func (dl *DreamLog) DreamCount() int {
	return len(dl.Dreams)
}
