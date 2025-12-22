package knowledge

import (
	"sort"
	"sync"
	"testing"
)

// TestSetEntityRoom tests the SetEntityRoom method
func TestSetEntityRoom(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*UpdateManager)
		entityID    string
		roomID      string
		verify      func(*testing.T, *UpdateManager)
		description string
	}{
		{
			name:     "set entity to new room",
			setup:    func(m *UpdateManager) {},
			entityID: "npc1",
			roomID:   "room1",
			verify: func(t *testing.T, m *UpdateManager) {
				room := m.GetEntityRoom("npc1")
				if room != "room1" {
					t.Errorf("expected entity in room1, got %s", room)
				}
				entities := m.GetEntitiesInRoom("room1")
				if len(entities) != 1 || entities[0] != "npc1" {
					t.Errorf("expected [npc1] in room1, got %v", entities)
				}
			},
			description: "should successfully place entity in new room",
		},
		{
			name: "move entity from one room to another",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entityID: "npc1",
			roomID:   "room2",
			verify: func(t *testing.T, m *UpdateManager) {
				// Entity should be in room2
				room := m.GetEntityRoom("npc1")
				if room != "room2" {
					t.Errorf("expected entity in room2, got %s", room)
				}
				// room1 should be empty
				entities1 := m.GetEntitiesInRoom("room1")
				if len(entities1) != 0 {
					t.Errorf("expected room1 to be empty, got %v", entities1)
				}
				// room2 should contain npc1
				entities2 := m.GetEntitiesInRoom("room2")
				if len(entities2) != 1 || entities2[0] != "npc1" {
					t.Errorf("expected [npc1] in room2, got %v", entities2)
				}
			},
			description: "should remove entity from old room and place in new room",
		},
		{
			name: "set player to room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entityID: "player",
			roomID:   "room1",
			verify: func(t *testing.T, m *UpdateManager) {
				entities := m.GetEntitiesInRoom("room1")
				// Should have both npc1 and player
				if len(entities) != 2 {
					t.Errorf("expected 2 entities in room1, got %d", len(entities))
				}
				// Check that both are present
				hasNPC := false
				hasPlayer := false
				for _, e := range entities {
					if e == "npc1" {
						hasNPC = true
					}
					if e == "player" {
						hasPlayer = true
					}
				}
				if !hasNPC || !hasPlayer {
					t.Errorf("expected both npc1 and player in room1, got %v", entities)
				}
			},
			description: "should handle player entity correctly",
		},
		{
			name: "set entity to same room (idempotent)",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entityID: "npc1",
			roomID:   "room1",
			verify: func(t *testing.T, m *UpdateManager) {
				room := m.GetEntityRoom("npc1")
				if room != "room1" {
					t.Errorf("expected entity in room1, got %s", room)
				}
				entities := m.GetEntitiesInRoom("room1")
				if len(entities) != 1 {
					t.Errorf("expected 1 entity in room1, got %d", len(entities))
				}
			},
			description: "setting entity to current room should be idempotent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(nil)
			tt.setup(mgr)
			mgr.SetEntityRoom(tt.entityID, tt.roomID)
			tt.verify(t, mgr)
		})
	}
}

// TestGetEntitiesInRoom tests the GetEntitiesInRoom method
func TestGetEntitiesInRoom(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*UpdateManager)
		roomID   string
		expected []string
	}{
		{
			name:     "empty room",
			setup:    func(m *UpdateManager) {},
			roomID:   "room1",
			expected: []string{},
		},
		{
			name: "room with single entity",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			roomID:   "room1",
			expected: []string{"npc1"},
		},
		{
			name: "room with multiple entities",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room1")
				m.SetEntityRoom("player", "room1")
			},
			roomID:   "room1",
			expected: []string{"npc1", "npc2", "player"},
		},
		{
			name: "non-existent room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			roomID:   "room999",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(nil)
			tt.setup(mgr)
			entities := mgr.GetEntitiesInRoom(tt.roomID)

			// Sort both slices for comparison
			sort.Strings(entities)
			sort.Strings(tt.expected)

			if len(entities) != len(tt.expected) {
				t.Errorf("expected %d entities, got %d", len(tt.expected), len(entities))
				return
			}

			for i := range entities {
				if entities[i] != tt.expected[i] {
					t.Errorf("expected entities %v, got %v", tt.expected, entities)
					break
				}
			}
		})
	}
}

// TestIsInSameRoom tests the IsInSameRoom method
func TestIsInSameRoom(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*UpdateManager)
		entity1   string
		entity2   string
		expected  bool
	}{
		{
			name: "two entities in same room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room1")
			},
			entity1:  "npc1",
			entity2:  "npc2",
			expected: true,
		},
		{
			name: "two entities in different rooms",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room2")
			},
			entity1:  "npc1",
			entity2:  "npc2",
			expected: false,
		},
		{
			name: "one entity not in any room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entity1:  "npc1",
			entity2:  "npc2",
			expected: false,
		},
		{
			name:     "both entities not in any room",
			setup:    func(m *UpdateManager) {},
			entity1:  "npc1",
			entity2:  "npc2",
			expected: false,
		},
		{
			name: "player and NPC in same room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("player", "room1")
				m.SetEntityRoom("npc1", "room1")
			},
			entity1:  "player",
			entity2:  "npc1",
			expected: true,
		},
		{
			name: "entity compared with itself",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entity1:  "npc1",
			entity2:  "npc1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(nil)
			tt.setup(mgr)
			result := mgr.IsInSameRoom(tt.entity1, tt.entity2)

			if result != tt.expected {
				t.Errorf("IsInSameRoom(%s, %s) = %v, want %v", tt.entity1, tt.entity2, result, tt.expected)
			}
		})
	}
}

// TestGetNPCsInSameRoom tests the GetNPCsInSameRoom method
func TestGetNPCsInSameRoom(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*UpdateManager)
		entityID string
		expected []string
	}{
		{
			name: "player with multiple NPCs in same room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("player", "room1")
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room1")
			},
			entityID: "player",
			expected: []string{"npc1", "npc2"},
		},
		{
			name: "NPC with other NPCs in same room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room1")
				m.SetEntityRoom("npc3", "room1")
			},
			entityID: "npc1",
			expected: []string{"npc2", "npc3"},
		},
		{
			name: "entity alone in room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entityID: "npc1",
			expected: []string{},
		},
		{
			name: "entity with only player in room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("player", "room1")
			},
			entityID: "npc1",
			expected: []string{},
		},
		{
			name:     "entity not in any room",
			setup:    func(m *UpdateManager) {},
			entityID: "npc1",
			expected: []string{},
		},
		{
			name: "NPCs in different rooms",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room2")
				m.SetEntityRoom("npc3", "room2")
			},
			entityID: "npc1",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(nil)
			tt.setup(mgr)
			npcs := mgr.GetNPCsInSameRoom(tt.entityID)

			// Sort both slices for comparison
			sort.Strings(npcs)
			sort.Strings(tt.expected)

			if len(npcs) != len(tt.expected) {
				t.Errorf("expected %d NPCs, got %d: %v", len(tt.expected), len(npcs), npcs)
				return
			}

			for i := range npcs {
				if npcs[i] != tt.expected[i] {
					t.Errorf("expected NPCs %v, got %v", tt.expected, npcs)
					break
				}
			}
		})
	}
}

// TestGetEntityRoom tests the GetEntityRoom method
func TestGetEntityRoom(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*UpdateManager)
		entityID string
		expected string
	}{
		{
			name: "entity in room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			entityID: "npc1",
			expected: "room1",
		},
		{
			name:     "entity not in any room",
			setup:    func(m *UpdateManager) {},
			entityID: "npc1",
			expected: "",
		},
		{
			name: "player in room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("player", "room1")
			},
			entityID: "player",
			expected: "room1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(nil)
			tt.setup(mgr)
			room := mgr.GetEntityRoom(tt.entityID)

			if room != tt.expected {
				t.Errorf("GetEntityRoom(%s) = %s, want %s", tt.entityID, room, tt.expected)
			}
		})
	}
}

// TestClearRoom tests the ClearRoom method
func TestClearRoom(t *testing.T) {
	mgr := NewUpdateManager(nil)

	// Setup: place entities in rooms
	mgr.SetEntityRoom("npc1", "room1")
	mgr.SetEntityRoom("npc2", "room1")
	mgr.SetEntityRoom("npc3", "room2")

	// Clear room1
	mgr.ClearRoom("room1")

	// Verify room1 is empty
	entities := mgr.GetEntitiesInRoom("room1")
	if len(entities) != 0 {
		t.Errorf("expected room1 to be empty, got %v", entities)
	}

	// Verify entities from room1 are not in any room
	room1 := mgr.GetEntityRoom("npc1")
	if room1 != "" {
		t.Errorf("expected npc1 to not be in any room, got %s", room1)
	}

	room2 := mgr.GetEntityRoom("npc2")
	if room2 != "" {
		t.Errorf("expected npc2 to not be in any room, got %s", room2)
	}

	// Verify room2 is unaffected
	entities2 := mgr.GetEntitiesInRoom("room2")
	if len(entities2) != 1 || entities2[0] != "npc3" {
		t.Errorf("expected room2 to contain [npc3], got %v", entities2)
	}
}

// TestGetAllRooms tests the GetAllRooms method
func TestGetAllRooms(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*UpdateManager)
		expected []string
	}{
		{
			name:     "no rooms",
			setup:    func(m *UpdateManager) {},
			expected: []string{},
		},
		{
			name: "single room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
			},
			expected: []string{"room1"},
		},
		{
			name: "multiple rooms",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room2")
				m.SetEntityRoom("npc3", "room3")
			},
			expected: []string{"room1", "room2", "room3"},
		},
		{
			name: "multiple entities in same room",
			setup: func(m *UpdateManager) {
				m.SetEntityRoom("npc1", "room1")
				m.SetEntityRoom("npc2", "room1")
			},
			expected: []string{"room1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(nil)
			tt.setup(mgr)
			rooms := mgr.GetAllRooms()

			// Sort both slices for comparison
			sort.Strings(rooms)
			sort.Strings(tt.expected)

			if len(rooms) != len(tt.expected) {
				t.Errorf("expected %d rooms, got %d", len(tt.expected), len(rooms))
				return
			}

			for i := range rooms {
				if rooms[i] != tt.expected[i] {
					t.Errorf("expected rooms %v, got %v", tt.expected, rooms)
					break
				}
			}
		})
	}
}

// TestRoomManagementThreadSafety tests thread safety of room management
func TestRoomManagementThreadSafety(t *testing.T) {
	mgr := NewUpdateManager(nil)
	var wg sync.WaitGroup

	// Number of concurrent operations
	numGoroutines := 100

	// Test concurrent SetEntityRoom
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			entityID := "npc" + string(rune('0'+id%10))
			roomID := "room" + string(rune('0'+id%5))
			mgr.SetEntityRoom(entityID, roomID)
		}(i)
	}
	wg.Wait()

	// Test concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			roomID := "room" + string(rune('0'+id%5))
			_ = mgr.GetEntitiesInRoom(roomID)
			_ = mgr.GetAllRooms()
		}(i)
	}
	wg.Wait()

	// Test concurrent IsInSameRoom
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			entity1 := "npc" + string(rune('0'+id%10))
			entity2 := "npc" + string(rune('0'+(id+1)%10))
			_ = mgr.IsInSameRoom(entity1, entity2)
		}(i)
	}
	wg.Wait()

	// Test concurrent GetNPCsInSameRoom
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			entityID := "npc" + string(rune('0'+id%10))
			_ = mgr.GetNPCsInSameRoom(entityID)
		}(i)
	}
	wg.Wait()

	// If we reach here without deadlock or race conditions, the test passes
	t.Log("Thread safety test completed successfully")
}

// TestRoomOccupantsMapMaintenance tests that roomOccupants map is correctly maintained
func TestRoomOccupantsMapMaintenance(t *testing.T) {
	mgr := NewUpdateManager(nil)

	// Add entity to room
	mgr.SetEntityRoom("npc1", "room1")

	// Verify room exists
	mgr.mu.RLock()
	if mgr.roomOccupants["room1"] == nil {
		t.Error("room1 should exist in roomOccupants map")
	}
	mgr.mu.RUnlock()

	// Move entity to another room
	mgr.SetEntityRoom("npc1", "room2")

	// Verify room1 is cleaned up (empty rooms should be removed)
	mgr.mu.RLock()
	if mgr.roomOccupants["room1"] != nil {
		t.Error("room1 should be removed from roomOccupants map when empty")
	}
	if mgr.roomOccupants["room2"] == nil {
		t.Error("room2 should exist in roomOccupants map")
	}
	mgr.mu.RUnlock()

	// Add another entity to room2
	mgr.SetEntityRoom("npc2", "room2")

	// Remove npc1 from room2
	mgr.SetEntityRoom("npc1", "room3")

	// Verify room2 still exists (npc2 is still there)
	mgr.mu.RLock()
	if mgr.roomOccupants["room2"] == nil {
		t.Error("room2 should still exist (npc2 is there)")
	}
	if len(mgr.roomOccupants["room2"]) != 1 {
		t.Errorf("room2 should have 1 occupant, got %d", len(mgr.roomOccupants["room2"]))
	}
	mgr.mu.RUnlock()
}
