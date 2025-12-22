package knowledge

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GameEvent represents an event that occurred in the game world for propagation purposes.
// This is used by PropagateEvent to create facts that are automatically shared with witnesses.
type GameEvent struct {
	ID          string
	Type        string
	Description string
	Initiator   string // Entity ID that caused this event
	Location    string // Room ID where event occurred
	Beat        int    // Game beat/turn when this happened
	Importance  int    // 1-10 importance level
}

// RegisterFact registers a new fact in the global fact repository.
// This method is thread-safe and ensures all facts have unique IDs.
//
// If the fact's ID is empty, a unique ID is automatically generated.
// The fact is then stored in the global facts map for reference.
//
// Parameters:
//   - fact: The fact to register
//
// AC1: RegisterFact() 註冊新事實到全域庫
func (m *UpdateManager) RegisterFact(fact *Fact) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not provided
	if fact.ID == "" {
		fact.ID = generateFactID()
	}

	// Register in global facts
	m.globalFacts[fact.ID] = fact
}

// PropagateEvent propagates an event to all entities in the same room as witnesses.
// This represents events that are visibly occurring in a location - all present entities
// automatically learn about them with full confidence and witness status.
//
// The created fact is of type Event and is automatically shared with all entities
// in the event's location with:
//   - LearnMethod: Witness (they saw it happen)
//   - Confidence: 1.0 (complete certainty)
//   - PropagationDepth: 0 (original source)
//
// Parameters:
//   - event: The game event to propagate
//
// AC2: PropagateEvent() 傳播事件給同房間目擊者
func (m *UpdateManager) PropagateEvent(event *GameEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create fact from event
	fact := &Fact{
		ID:        event.ID,
		Content:   event.Description,
		Type:      Event,
		Source:    event.Initiator,
		CreatedAt: time.Now(),
		Location:  event.Location,
		Witnesses: []string{},
	}

	// Generate ID if not provided
	if fact.ID == "" {
		fact.ID = generateFactID()
	}

	// Find all entities in the same room
	witnesses := []string{}
	if occupants, exists := m.roomOccupants[event.Location]; exists {
		for entityID := range occupants {
			witnesses = append(witnesses, entityID)
		}
	}

	fact.Witnesses = witnesses

	// Register in global facts
	m.globalFacts[fact.ID] = fact

	// Propagate to all witnesses
	for _, entityID := range witnesses {
		m.addKnowledge(entityID, fact, Witness, 1.0, 0)
	}
}

// LearnFromDialogue propagates dialogue content to entities in the same room.
// When an entity speaks, other entities in the same room can hear and learn the information.
//
// Only entities in the same room as the speaker can learn from dialogue.
// Entities outside the room cannot hear and will not learn.
//
// The created fact is of type Dialogue and is shared with entities in the same room with:
//   - LearnMethod: Told (they heard someone say it)
//   - Confidence: 0.9 (very confident, but not witnessed)
//   - PropagationDepth: 1 (one degree removed from source)
//
// Parameters:
//   - listenerID: The entity ID that is listening
//   - speakerID: The entity ID that is speaking
//   - content: The dialogue content
//   - currentRoom: The room ID where the conversation is happening
//
// AC3: LearnFromDialogue() 從對話中學習（需同房間）
func (m *UpdateManager) LearnFromDialogue(listenerID, speakerID, content, currentRoom string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if listener and speaker are in the same room
	// If currentRoom is provided, verify both are in that room
	listenerRoom := m.getEntityRoomLocked(listenerID)
	speakerRoom := m.getEntityRoomLocked(speakerID)

	// Both must be in the same room
	if listenerRoom == "" || speakerRoom == "" || listenerRoom != speakerRoom {
		return // Cannot hear dialogue from different room
	}

	// If currentRoom is specified, verify it matches
	if currentRoom != "" && listenerRoom != currentRoom {
		return
	}

	// Create fact from dialogue
	fact := &Fact{
		ID:        generateFactID(),
		Content:   content,
		Type:      Dialogue,
		Source:    speakerID,
		CreatedAt: time.Now(),
		Location:  listenerRoom,
		Witnesses: []string{listenerID}, // Listener heard this
	}

	// Register in global facts
	m.globalFacts[fact.ID] = fact

	// Add to listener's knowledge
	m.addKnowledge(listenerID, fact, Told, 0.9, 1)
}

// TellNPC allows one entity to explicitly tell another entity about a fact they know.
// This represents deliberate information sharing and includes confidence decay and depth tracking.
//
// The confidence decays by 15% (multiplied by 0.85) each time a fact is told.
// The propagation depth increases by 1 to track how far removed the information is from source.
//
// Information distortion may occur based on configuration and random chance.
// Once MaxPropagationDepth is reached, the fact can no longer be shared.
//
// Parameters:
//   - tellerID: The entity ID that is sharing the information
//   - listenerID: The entity ID that is receiving the information
//   - factID: The ID of the fact being shared
//
// Returns:
//   - error if the teller doesn't know the fact, or if max depth is reached
//
// AC4: TellNPC() 主動告知 NPC 資訊（確信度衰減、傳播深度+1）
// AC6: 尊重 MaxPropagationDepth 限制
func (m *UpdateManager) TellNPC(tellerID, listenerID, factID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if teller knows this fact
	tellerKB := m.getKnowledgeBaseLocked(tellerID)
	if tellerKB == nil {
		return fmt.Errorf("teller %s has no knowledge base", tellerID)
	}

	knownFact, exists := tellerKB.KnownFacts[factID]
	if !exists {
		return fmt.Errorf("teller %s does not know fact %s", tellerID, factID)
	}

	// Get original fact
	originalFact := m.globalFacts[factID]
	if originalFact == nil {
		return fmt.Errorf("fact %s not found in global facts", factID)
	}

	// Calculate confidence decay (multiply by 0.85)
	newConfidence := knownFact.Confidence * 0.85

	// Increase propagation depth
	newDepth := knownFact.PropagationDepth + 1

	// Check max propagation depth
	if newDepth > m.config.MaxPropagationDepth {
		return fmt.Errorf("max propagation depth (%d) reached", m.config.MaxPropagationDepth)
	}

	// Check for information distortion
	// Story 8.3: Use intelligent distortion calculator if available
	isDistorted := knownFact.IsDistorted
	distortedContent := knownFact.DistortedContent

	// If not already distorted, check if it should become distorted now
	if !isDistorted && m.config.EnableDistortion {
		// Story 8.3: Use intelligent distortion calculator if available
		if m.distortionCalculator != nil {
			// Use smart distortion based on listener's state
			result, err := m.distortionCalculator.ApplyDistortion(listenerID, originalFact.Content, newDepth)
			if err == nil && result.ShouldDistort {
				isDistorted = true
				distortedContent = result.DistortedContent
			}
		} else {
			// Fallback to simple random distortion
			if rand.Float64() < m.config.DistortionRate {
				isDistorted = true
				// Use teller's distorted content if they have one, otherwise distort the original
				if knownFact.IsDistorted && knownFact.DistortedContent != "" {
					distortedContent = knownFact.DistortedContent
				} else {
					distortedContent = m.distortFact(originalFact.Content)
				}
			}
		}
	}

	// Add to listener's knowledge base
	m.ensureKnowledgeBaseLocked(listenerID)
	listenerKB := m.npcKnowledge[listenerID]
	if listenerKB == nil {
		// Player knowledge
		listenerKB = m.playerKnowledge
	}

	// Check if listener already knows this fact
	if existing, exists := listenerKB.KnownFacts[factID]; exists {
		// Update only if new confidence is higher
		if newConfidence > existing.Confidence {
			existing.Confidence = newConfidence
			existing.LearnMethod = Told
			existing.LearnedAt = time.Now()
			existing.LearnedFrom = tellerID
			existing.PropagationDepth = newDepth
			existing.IsDistorted = isDistorted
			existing.DistortedContent = distortedContent
		}
		return nil
	}

	// Add new knowledge
	listenerKB.KnownFacts[factID] = &KnownFact{
		FactID:           factID,
		LearnedAt:        time.Now(),
		LearnedFrom:      tellerID,
		LearnMethod:      Told,
		Confidence:       newConfidence,
		IsDistorted:      isDistorted,
		DistortedContent: distortedContent,
		PropagationDepth: newDepth,
	}
	listenerKB.LastUpdated = time.Now()

	return nil
}

// addKnowledge is an internal helper method that adds a fact to an entity's knowledge base.
// This method assumes the mutex is already locked and should only be called from within
// other locked methods.
//
// If the entity already knows the fact, it only updates if the new confidence is higher.
// If the entity is unknown, a new knowledge base is automatically created.
//
// Parameters:
//   - entityID: The entity ID to add knowledge to
//   - fact: The fact to add
//   - method: How the fact was learned
//   - confidence: Confidence level (0.0-1.0)
//   - depth: Propagation depth (0 = original witness)
//
// AC5: addKnowledge() 添加知識到 KnowledgeBase（防重複）
func (m *UpdateManager) addKnowledge(entityID string, fact *Fact, method LearnMethod, confidence float64, depth int) {
	// Ensure knowledge base exists
	m.ensureKnowledgeBaseLocked(entityID)

	var kb *KnowledgeBase
	if entityID == "player" {
		kb = m.playerKnowledge
	} else {
		kb = m.npcKnowledge[entityID]
	}

	// Check if already known
	if existing, exists := kb.KnownFacts[fact.ID]; exists {
		// Update only if new confidence is higher
		if confidence > existing.Confidence {
			existing.Confidence = confidence
			existing.LearnMethod = method
			existing.LearnedAt = time.Now()
			existing.PropagationDepth = depth
		}
		return
	}

	// Add new knowledge
	kb.KnownFacts[fact.ID] = &KnownFact{
		FactID:           fact.ID,
		LearnedAt:        time.Now(),
		LearnedFrom:      fact.Source,
		LearnMethod:      method,
		Confidence:       confidence,
		IsDistorted:      false,
		DistortedContent: "",
		PropagationDepth: depth,
	}
	kb.LastUpdated = time.Now()
}

// ensureKnowledgeBaseLocked ensures that a knowledge base exists for the given entity.
// This is a locked version that assumes the mutex is already held.
// Creates a new knowledge base if one doesn't exist.
func (m *UpdateManager) ensureKnowledgeBaseLocked(entityID string) {
	if entityID == "player" {
		if m.playerKnowledge == nil {
			m.playerKnowledge = NewKnowledgeBase("player")
		}
		return
	}

	if m.npcKnowledge[entityID] == nil {
		m.npcKnowledge[entityID] = NewKnowledgeBase(entityID)
	}
}

// getKnowledgeBaseLocked retrieves a knowledge base for the given entity.
// This is a locked version that assumes the mutex is already held.
// Returns nil if the knowledge base doesn't exist.
func (m *UpdateManager) getKnowledgeBaseLocked(entityID string) *KnowledgeBase {
	if entityID == "player" {
		return m.playerKnowledge
	}
	return m.npcKnowledge[entityID]
}

// getEntityRoomLocked finds which room an entity is in.
// This is a locked version that assumes the mutex is already held.
// Returns empty string if entity is not in any room.
func (m *UpdateManager) getEntityRoomLocked(entityID string) string {
	for roomID, occupants := range m.roomOccupants {
		if occupants[entityID] {
			return roomID
		}
	}
	return ""
}

// distortFact creates a distorted version of fact content.
// This is a simple implementation that will be enhanced in Story 2.5.
//
// Current implementation: Simple text modification as placeholder
// Future: More sophisticated distortion based on NPC mental state
func (m *UpdateManager) distortFact(content string) string {
	// Simple distortion: add uncertainty markers
	// This will be replaced with more sophisticated logic in Story 2.5
	distortions := []string{
		"我聽說" + content,
		"可能" + content,
		"據說" + content,
		content + "（不確定）",
	}
	return distortions[rand.Intn(len(distortions))]
}

// getCurrentBeat returns the current game beat/turn number.
// This is a placeholder implementation that will be integrated with GameStateV2.
//
// Current implementation: Returns 0
// Future: Will retrieve from GameStateV2.CurrentBeat
func (m *UpdateManager) getCurrentBeat() int {
	// TODO: Integrate with GameStateV2
	return 0
}

// generateFactID generates a unique ID for a fact.
// Uses timestamp and random number to ensure uniqueness.
func generateFactID() string {
	return fmt.Sprintf("fact_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// GetKnowledgeBase retrieves a knowledge base for public access.
// This is the thread-safe public version.
func (m *UpdateManager) GetKnowledgeBase(entityID string) *KnowledgeBase {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if entityID == "player" {
		return m.playerKnowledge.Copy()
	}

	kb := m.npcKnowledge[entityID]
	if kb == nil {
		return nil
	}
	return kb.Copy()
}

// GetGlobalFact retrieves a fact from the global repository.
func (m *UpdateManager) GetGlobalFact(factID string) *Fact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fact := m.globalFacts[factID]
	if fact == nil {
		return nil
	}
	return fact.Copy()
}

// GetAllFacts returns all facts in the global repository.
// This is useful for serialization and debugging.
func (m *UpdateManager) GetAllFacts() []*Fact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	facts := make([]*Fact, 0, len(m.globalFacts))
	for _, fact := range m.globalFacts {
		facts = append(facts, fact.Copy())
	}
	return facts
}

// LoadFacts loads facts into the global repository.
// This is used for deserialization.
func (m *UpdateManager) LoadFacts(facts []*Fact) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, fact := range facts {
		m.globalFacts[fact.ID] = fact
	}
}

// Simple helper for case-insensitive contains check
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
