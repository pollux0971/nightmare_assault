package knowledge

import (
	"encoding/json"
	"time"
)

// FactType represents the category of a fact.
type FactType int

const (
	// Event represents a fact about an event that occurred
	Event FactType = iota
	// Dialogue represents a fact from a conversation
	Dialogue
	// Discovery represents a fact about something discovered
	Discovery
	// Rumor represents unverified information or hearsay
	Rumor
	// Secret represents confidential or hidden information
	Secret
)

// String returns a string representation of the FactType.
func (f FactType) String() string {
	switch f {
	case Event:
		return "event"
	case Dialogue:
		return "dialogue"
	case Discovery:
		return "discovery"
	case Rumor:
		return "rumor"
	case Secret:
		return "secret"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for FactType.
func (f FactType) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

// UnmarshalJSON implements json.Unmarshaler for FactType.
func (f *FactType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "event":
		*f = Event
	case "dialogue":
		*f = Dialogue
	case "discovery":
		*f = Discovery
	case "rumor":
		*f = Rumor
	case "secret":
		*f = Secret
	default:
		*f = Event
	}
	return nil
}

// Fact represents a piece of information in the game world.
// Facts are stored in the global fact repository and can be learned by NPCs and players.
type Fact struct {
	ID        string    `json:"id"`         // Unique identifier for the fact
	Content   string    `json:"content"`    // The actual information content
	Type      FactType  `json:"type"`       // Category of the fact
	Source    string    `json:"source"`     // Who/what created this fact (entity ID or "system")
	CreatedAt time.Time `json:"created_at"` // When this fact was created
	Location  string    `json:"location"`   // Where this fact originated (room/area ID)
	Witnesses []string  `json:"witnesses"`  // List of entity IDs who witnessed this fact firsthand
}

// NewFact creates a new Fact with the given parameters.
func NewFact(id, content string, factType FactType, source, location string, witnesses []string) *Fact {
	if witnesses == nil {
		witnesses = []string{}
	}
	return &Fact{
		ID:        id,
		Content:   content,
		Type:      factType,
		Source:    source,
		CreatedAt: time.Now(),
		Location:  location,
		Witnesses: witnesses,
	}
}

// Copy creates a deep copy of the Fact.
func (f *Fact) Copy() *Fact {
	witnessesCopy := make([]string, len(f.Witnesses))
	copy(witnessesCopy, f.Witnesses)

	return &Fact{
		ID:        f.ID,
		Content:   f.Content,
		Type:      f.Type,
		Source:    f.Source,
		CreatedAt: f.CreatedAt,
		Location:  f.Location,
		Witnesses: witnessesCopy,
	}
}

// AddWitness adds a witness to the fact if not already present.
func (f *Fact) AddWitness(entityID string) {
	for _, w := range f.Witnesses {
		if w == entityID {
			return // Already a witness
		}
	}
	f.Witnesses = append(f.Witnesses, entityID)
}

// IsWitness checks if the given entity ID is a witness to this fact.
func (f *Fact) IsWitness(entityID string) bool {
	for _, w := range f.Witnesses {
		if w == entityID {
			return true
		}
	}
	return false
}

// String returns a JSON string representation of the Fact.
func (f *Fact) String() string {
	data, _ := json.Marshal(f)
	return string(data)
}
