package knowledge

import (
	"encoding/json"
	"time"
)

// LearnMethod represents how an entity learned a fact.
type LearnMethod int

const (
	// Witness represents direct observation of an event
	Witness LearnMethod = iota
	// Told represents being told by another entity
	Told
	// Overheard represents accidentally hearing information
	Overheard
	// Inferred represents deducing information from other facts
	Inferred
)

// String returns a string representation of the LearnMethod.
func (l LearnMethod) String() string {
	switch l {
	case Witness:
		return "witness"
	case Told:
		return "told"
	case Overheard:
		return "overheard"
	case Inferred:
		return "inferred"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for LearnMethod.
func (l LearnMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON implements json.Unmarshaler for LearnMethod.
func (l *LearnMethod) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "witness":
		*l = Witness
	case "told":
		*l = Told
	case "overheard":
		*l = Overheard
	case "inferred":
		*l = Inferred
	default:
		*l = Witness
	}
	return nil
}

// KnownFact represents a fact as known by an entity, with metadata about how it was learned.
type KnownFact struct {
	FactID           string      `json:"fact_id"`           // Reference to the Fact ID
	LearnedAt        time.Time   `json:"learned_at"`        // When the entity learned this fact
	LearnedFrom      string      `json:"learned_from"`      // Entity ID of who provided this info (empty if witnessed)
	LearnMethod      LearnMethod `json:"learn_method"`      // How the fact was learned
	Confidence       float64     `json:"confidence"`        // Confidence level (0.0-1.0)
	IsDistorted      bool        `json:"is_distorted"`      // Whether the fact has been distorted during propagation
	DistortedContent string      `json:"distorted_content"` // The distorted version of the content (if IsDistorted is true)
	PropagationDepth int         `json:"propagation_depth"` // How many times this fact has been passed between entities (0=witness)
}

// Belief represents an entity's personal interpretation or belief about something.
// Beliefs are derived from facts but include the entity's own interpretation.
type Belief struct {
	Content    string   `json:"content"`    // The belief statement
	BasedOn    []string `json:"based_on"`   // List of Fact IDs that support this belief
	Confidence float64  `json:"confidence"` // How strongly the entity believes this (0.0-1.0)
}

// KnowledgeBase represents an entity's personal knowledge and beliefs.
// Each NPC and the player has their own KnowledgeBase.
type KnowledgeBase struct {
	OwnerID     string                `json:"owner_id"`     // Entity ID (NPC or player) who owns this knowledge
	KnownFacts  map[string]*KnownFact `json:"known_facts"`  // Map of FactID -> KnownFact
	Beliefs     []*Belief             `json:"beliefs"`      // List of personal beliefs
	LastUpdated time.Time             `json:"last_updated"` // When this knowledge base was last modified
}

// NewKnowledgeBase creates a new KnowledgeBase for the given entity.
func NewKnowledgeBase(ownerID string) *KnowledgeBase {
	return &KnowledgeBase{
		OwnerID:     ownerID,
		KnownFacts:  make(map[string]*KnownFact),
		Beliefs:     []*Belief{},
		LastUpdated: time.Now(),
	}
}

// AddFact adds a new fact to the knowledge base. If the fact is already known, it updates it
// only if the new confidence is higher.
func (kb *KnowledgeBase) AddFact(knownFact *KnownFact) {
	existing, exists := kb.KnownFacts[knownFact.FactID]
	if exists {
		// Update if new confidence is higher
		if knownFact.Confidence > existing.Confidence {
			kb.KnownFacts[knownFact.FactID] = knownFact
			kb.LastUpdated = time.Now()
		}
	} else {
		// Add new fact
		kb.KnownFacts[knownFact.FactID] = knownFact
		kb.LastUpdated = time.Now()
	}
}

// HasFact checks if the entity knows a specific fact.
func (kb *KnowledgeBase) HasFact(factID string) bool {
	_, exists := kb.KnownFacts[factID]
	return exists
}

// GetFact retrieves a known fact by ID, or nil if not known.
func (kb *KnowledgeBase) GetFact(factID string) *KnownFact {
	return kb.KnownFacts[factID]
}

// AddBelief adds a new belief to the knowledge base.
func (kb *KnowledgeBase) AddBelief(belief *Belief) {
	kb.Beliefs = append(kb.Beliefs, belief)
	kb.LastUpdated = time.Now()
}

// GetAllFactIDs returns all fact IDs known to this entity.
func (kb *KnowledgeBase) GetAllFactIDs() []string {
	ids := make([]string, 0, len(kb.KnownFacts))
	for id := range kb.KnownFacts {
		ids = append(ids, id)
	}
	return ids
}

// Copy creates a deep copy of the KnowledgeBase.
func (kb *KnowledgeBase) Copy() *KnowledgeBase {
	// Copy known facts
	factsCopy := make(map[string]*KnownFact, len(kb.KnownFacts))
	for k, v := range kb.KnownFacts {
		factsCopy[k] = &KnownFact{
			FactID:           v.FactID,
			LearnedAt:        v.LearnedAt,
			LearnedFrom:      v.LearnedFrom,
			LearnMethod:      v.LearnMethod,
			Confidence:       v.Confidence,
			IsDistorted:      v.IsDistorted,
			DistortedContent: v.DistortedContent,
			PropagationDepth: v.PropagationDepth,
		}
	}

	// Copy beliefs
	beliefsCopy := make([]*Belief, len(kb.Beliefs))
	for i, b := range kb.Beliefs {
		basedOnCopy := make([]string, len(b.BasedOn))
		copy(basedOnCopy, b.BasedOn)
		beliefsCopy[i] = &Belief{
			Content:    b.Content,
			BasedOn:    basedOnCopy,
			Confidence: b.Confidence,
		}
	}

	return &KnowledgeBase{
		OwnerID:     kb.OwnerID,
		KnownFacts:  factsCopy,
		Beliefs:     beliefsCopy,
		LastUpdated: kb.LastUpdated,
	}
}

// String returns a JSON string representation of the KnowledgeBase.
func (kb *KnowledgeBase) String() string {
	data, _ := json.Marshal(kb)
	return string(data)
}
