package knowledge

// SetEntityRoom sets the room location for a given entity (NPC or player).
// This method is thread-safe and updates the roomOccupants map accordingly.
//
// If the entity was previously in another room, it is removed from that room.
// If the entity is already in the specified room, this is a no-op.
//
// Parameters:
//   - entityID: The unique identifier of the entity (NPC ID or "player")
//   - roomID: The unique identifier of the room
//
// AC1: SetEntityRoom() 設定實體所在房間
func (m *UpdateManager) SetEntityRoom(entityID, roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove entity from all rooms first
	for room, occupants := range m.roomOccupants {
		if occupants[entityID] {
			delete(occupants, entityID)
			// Clean up empty room map
			if len(occupants) == 0 {
				delete(m.roomOccupants, room)
			}
		}
	}

	// Add entity to new room
	if m.roomOccupants[roomID] == nil {
		m.roomOccupants[roomID] = make(map[string]bool)
	}
	m.roomOccupants[roomID][entityID] = true
}

// GetEntitiesInRoom returns a list of all entity IDs currently in the specified room.
// This method is thread-safe and returns a copy of the entity list.
//
// If the room doesn't exist or is empty, returns an empty slice.
//
// Parameters:
//   - roomID: The unique identifier of the room
//
// Returns:
//   - A slice of entity IDs (NPC IDs and/or "player")
//
// AC2: GetEntitiesInRoom() 取得房間中所有實體
func (m *UpdateManager) GetEntitiesInRoom(roomID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	occupants, exists := m.roomOccupants[roomID]
	if !exists || len(occupants) == 0 {
		return []string{}
	}

	// Create a copy of the entity IDs
	entities := make([]string, 0, len(occupants))
	for entityID := range occupants {
		entities = append(entities, entityID)
	}

	return entities
}

// IsInSameRoom checks if two entities are currently in the same room.
// This method is thread-safe.
//
// Parameters:
//   - entityID1: The first entity ID
//   - entityID2: The second entity ID
//
// Returns:
//   - true if both entities are in the same room, false otherwise
//
// AC3: IsInSameRoom() 檢查兩實體是否同房間
func (m *UpdateManager) IsInSameRoom(entityID1, entityID2 string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find which room each entity is in
	var room1, room2 string

	for roomID, occupants := range m.roomOccupants {
		if occupants[entityID1] {
			room1 = roomID
		}
		if occupants[entityID2] {
			room2 = roomID
		}
		// Early exit if both found
		if room1 != "" && room2 != "" {
			break
		}
	}

	// Both entities must be in the same room (and the room must not be empty string)
	return room1 != "" && room1 == room2
}

// GetNPCsInSameRoom returns a list of NPC IDs that are in the same room as the specified entity.
// This method is thread-safe and excludes the querying entity from the result.
//
// If the entity is not in any room, or is alone in the room, returns an empty slice.
// The returned list excludes "player" - only NPCs are returned.
//
// Parameters:
//   - entityID: The entity ID to query (typically "player" or an NPC ID)
//
// Returns:
//   - A slice of NPC IDs in the same room (excluding the querying entity and the player)
//
// AC4: GetNPCsInSameRoom() 取得同房間的 NPC 列表
func (m *UpdateManager) GetNPCsInSameRoom(entityID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find which room the entity is in
	var currentRoom string
	for roomID, occupants := range m.roomOccupants {
		if occupants[entityID] {
			currentRoom = roomID
			break
		}
	}

	// If entity is not in any room, return empty list
	if currentRoom == "" {
		return []string{}
	}

	// Get all occupants in the current room, excluding the querying entity and "player"
	occupants := m.roomOccupants[currentRoom]
	npcs := make([]string, 0)

	for occupantID := range occupants {
		// Exclude the querying entity and the player
		if occupantID != entityID && occupantID != "player" {
			npcs = append(npcs, occupantID)
		}
	}

	return npcs
}

// GetEntityRoom returns the room ID that the entity is currently in.
// This is a helper method for debugging and querying entity locations.
//
// Parameters:
//   - entityID: The entity ID to query
//
// Returns:
//   - The room ID where the entity is located, or empty string if not found
func (m *UpdateManager) GetEntityRoom(entityID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for roomID, occupants := range m.roomOccupants {
		if occupants[entityID] {
			return roomID
		}
	}

	return ""
}

// ClearRoom removes all entities from a specified room.
// This is useful for cleanup operations or scenario transitions.
//
// Parameters:
//   - roomID: The room ID to clear
func (m *UpdateManager) ClearRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.roomOccupants, roomID)
}

// GetAllRooms returns a list of all room IDs that currently have occupants.
// This is useful for debugging and scenario management.
//
// Returns:
//   - A slice of room IDs
func (m *UpdateManager) GetAllRooms() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms := make([]string, 0, len(m.roomOccupants))
	for roomID := range m.roomOccupants {
		rooms = append(rooms, roomID)
	}

	return rooms
}
