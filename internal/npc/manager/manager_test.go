package manager

import (
	"fmt"
	"sync"
	"testing"
)

// TestNewNPCManager tests the creation of a new NPCManager
func TestNewNPCManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *NPCManagerConfig
		expectNil   bool
		description string
	}{
		{
			name:        "with valid config",
			config:      DefaultNPCManagerConfig(),
			expectNil:   false,
			description: "should create manager with provided config",
		},
		{
			name:        "with nil config",
			config:      nil,
			expectNil:   false,
			description: "should create manager with default config",
		},
		{
			name: "with custom config",
			config: &NPCManagerConfig{
				TrustDecayRate:     1.0,
				FearDecayRate:      2.0,
				StressDecayRate:    1.5,
				BreakdownThreshold: 90,
				MinTrustForSecret:  80,
				HintDuration:       5,
			},
			expectNil:   false,
			description: "should create manager with custom config values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewNPCManager(nil, tt.config)

			if mgr == nil {
				t.Fatalf("NewNPCManager() returned nil")
			}

			// Verify maps are initialized
			if mgr.profiles == nil {
				t.Error("profiles map is nil")
			}
			if mgr.states == nil {
				t.Error("states map is nil")
			}
			if mgr.config == nil {
				t.Error("config is nil")
			}

			// Verify maps are empty
			if len(mgr.profiles) != 0 {
				t.Errorf("profiles map should be empty, got %d entries", len(mgr.profiles))
			}
			if len(mgr.states) != 0 {
				t.Errorf("states map should be empty, got %d entries", len(mgr.states))
			}

			// If custom config was provided, verify it was used
			if tt.config != nil {
				if mgr.config.TrustDecayRate != tt.config.TrustDecayRate {
					t.Errorf("TrustDecayRate = %v, want %v", mgr.config.TrustDecayRate, tt.config.TrustDecayRate)
				}
			}
		})
	}
}

// TestAddNPC tests adding NPCs to the manager
func TestAddNPC(t *testing.T) {
	tests := []struct {
		name        string
		profile     *NPCProfile
		wantErr     bool
		description string
	}{
		{
			name: "valid NPC",
			profile: &NPCProfile{
				ID:             "npc1",
				Name:           "Test NPC",
				Archetype:      "survivor",
				InitialEmotion: DefaultEmotionState(),
			},
			wantErr:     false,
			description: "should successfully add valid NPC",
		},
		{
			name:        "nil profile",
			profile:     nil,
			wantErr:     true,
			description: "should return error for nil profile",
		},
		{
			name: "empty ID",
			profile: &NPCProfile{
				ID:        "",
				Name:      "Invalid NPC",
				Archetype: "survivor",
			},
			wantErr:     true,
			description: "should return error for empty ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewNPCManager(nil, DefaultNPCManagerConfig())
			err := mgr.AddNPC(tt.profile)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddNPC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.profile != nil {
				// Verify profile was added
				profile := mgr.GetProfile(tt.profile.ID)
				if profile == nil {
					t.Error("profile not found after AddNPC")
				}

				// Verify state was initialized
				state := mgr.GetState(tt.profile.ID)
				if state == nil {
					t.Error("state not initialized after AddNPC")
				}

				// Verify state emotion matches profile initial emotion
				if state.Emotion != tt.profile.InitialEmotion {
					t.Errorf("state emotion = %v, want %v", state.Emotion, tt.profile.InitialEmotion)
				}

				// Verify state is alive
				if !state.IsAlive {
					t.Error("new NPC should be alive")
				}
			}
		})
	}
}

// TestAddNPC_DuplicateID tests adding NPC with duplicate ID
func TestAddNPC_DuplicateID(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile1 := &NPCProfile{
		ID:             "npc1",
		Name:           "First NPC",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	}

	// Add first NPC - should succeed
	err := mgr.AddNPC(profile1)
	if err != nil {
		t.Fatalf("first AddNPC() failed: %v", err)
	}

	profile2 := &NPCProfile{
		ID:             "npc1", // Same ID
		Name:           "Second NPC",
		Archetype:      "researcher",
		InitialEmotion: DefaultEmotionState(),
	}

	// Add second NPC with same ID - should fail
	err = mgr.AddNPC(profile2)
	if err == nil {
		t.Error("AddNPC() should return error for duplicate ID")
	}

	// Verify first profile is still there and unchanged
	profile := mgr.GetProfile("npc1")
	if profile == nil {
		t.Fatal("original profile was lost")
	}
	if profile.Name != "First NPC" {
		t.Errorf("profile was modified, Name = %s, want %s", profile.Name, "First NPC")
	}
}

// TestGetProfile tests retrieving NPC profiles
func TestGetProfile(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Add test NPC
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	}
	mgr.AddNPC(profile)

	tests := []struct {
		name      string
		npcID     string
		wantNil   bool
		wantName  string
	}{
		{
			name:     "existing NPC",
			npcID:    "npc1",
			wantNil:  false,
			wantName: "Test NPC",
		},
		{
			name:    "non-existent NPC",
			npcID:   "npc999",
			wantNil: true,
		},
		{
			name:    "empty ID",
			npcID:   "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := mgr.GetProfile(tt.npcID)

			if (profile == nil) != tt.wantNil {
				t.Errorf("GetProfile() returned nil = %v, wantNil %v", profile == nil, tt.wantNil)
			}

			if !tt.wantNil && profile != nil {
				if profile.Name != tt.wantName {
					t.Errorf("GetProfile() Name = %s, want %s", profile.Name, tt.wantName)
				}
			}
		})
	}
}

// TestGetState tests retrieving NPC runtime states
func TestGetState(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Add test NPC
	customEmotion := NewEmotionState(60, 30, 20)
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		Archetype:      "survivor",
		InitialEmotion: customEmotion,
	}
	mgr.AddNPC(profile)

	tests := []struct {
		name    string
		npcID   string
		wantNil bool
	}{
		{
			name:    "existing NPC",
			npcID:   "npc1",
			wantNil: false,
		},
		{
			name:    "non-existent NPC",
			npcID:   "npc999",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := mgr.GetState(tt.npcID)

			if (state == nil) != tt.wantNil {
				t.Errorf("GetState() returned nil = %v, wantNil %v", state == nil, tt.wantNil)
			}

			if !tt.wantNil && state != nil {
				// Verify state matches profile's initial emotion
				if state.Emotion != customEmotion {
					t.Errorf("GetState() Emotion = %v, want %v", state.Emotion, customEmotion)
				}
				if !state.IsAlive {
					t.Error("new NPC should be alive")
				}
			}
		})
	}
}

// TestUpdateState tests updating NPC runtime states
func TestUpdateState(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Add test NPC
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	}
	mgr.AddNPC(profile)

	tests := []struct {
		name    string
		npcID   string
		state   *NPCRuntimeState
		wantErr bool
	}{
		{
			name:  "valid update",
			npcID: "npc1",
			state: &NPCRuntimeState{
				Emotion:     NewEmotionState(70, 10, 15),
				MentalState: Normal,
				IsAlive:     true,
			},
			wantErr: false,
		},
		{
			name:    "nil state",
			npcID:   "npc1",
			state:   nil,
			wantErr: true,
		},
		{
			name:  "non-existent NPC",
			npcID: "npc999",
			state: &NPCRuntimeState{
				Emotion:     DefaultEmotionState(),
				MentalState: Normal,
				IsAlive:     true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.UpdateState(tt.npcID, tt.state)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify state was updated
				state := mgr.GetState(tt.npcID)
				if state == nil {
					t.Fatal("state not found after update")
				}
				if state.Emotion != tt.state.Emotion {
					t.Errorf("state not updated, Emotion = %v, want %v", state.Emotion, tt.state.Emotion)
				}
			}
		})
	}
}

// TestDeleteNPC tests deleting NPCs
func TestDeleteNPC(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Add test NPCs
	mgr.AddNPC(&NPCProfile{
		ID:             "npc1",
		Name:           "NPC 1",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	})
	mgr.AddNPC(&NPCProfile{
		ID:             "npc2",
		Name:           "NPC 2",
		Archetype:      "researcher",
		InitialEmotion: DefaultEmotionState(),
	})

	tests := []struct {
		name    string
		npcID   string
		wantErr bool
	}{
		{
			name:    "delete existing NPC",
			npcID:   "npc1",
			wantErr: false,
		},
		{
			name:    "delete non-existent NPC",
			npcID:   "npc999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.DeleteNPC(tt.npcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteNPC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify NPC was deleted
				profile := mgr.GetProfile(tt.npcID)
				if profile != nil {
					t.Error("profile still exists after deletion")
				}
				state := mgr.GetState(tt.npcID)
				if state != nil {
					t.Error("state still exists after deletion")
				}
			}
		})
	}

	// Verify npc2 still exists
	profile := mgr.GetProfile("npc2")
	if profile == nil {
		t.Error("npc2 was incorrectly deleted")
	}
}

// TestListNPCIDs tests listing all NPC IDs
func TestListNPCIDs(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Test empty manager
	ids := mgr.ListNPCIDs()
	if len(ids) != 0 {
		t.Errorf("ListNPCIDs() on empty manager returned %d IDs, want 0", len(ids))
	}

	// Add NPCs
	npcs := []string{"npc1", "npc2", "npc3"}
	for _, id := range npcs {
		mgr.AddNPC(&NPCProfile{
			ID:             id,
			Name:           "NPC " + id,
			Archetype:      "survivor",
			InitialEmotion: DefaultEmotionState(),
		})
	}

	// Test with NPCs
	ids = mgr.ListNPCIDs()
	if len(ids) != len(npcs) {
		t.Errorf("ListNPCIDs() returned %d IDs, want %d", len(ids), len(npcs))
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}
	for _, expectedID := range npcs {
		if !idMap[expectedID] {
			t.Errorf("ListNPCIDs() missing ID %s", expectedID)
		}
	}
}

// TestGetConfig tests retrieving the manager configuration
func TestGetConfig(t *testing.T) {
	customConfig := &NPCManagerConfig{
		TrustDecayRate:     1.5,
		FearDecayRate:      2.5,
		StressDecayRate:    2.0,
		BreakdownThreshold: 85,
		MinTrustForSecret:  70,
		HintDuration:       4,
	}

	mgr := NewNPCManager(nil, customConfig)
	config := mgr.GetConfig()

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	if config.TrustDecayRate != customConfig.TrustDecayRate {
		t.Errorf("TrustDecayRate = %v, want %v", config.TrustDecayRate, customConfig.TrustDecayRate)
	}
	if config.FearDecayRate != customConfig.FearDecayRate {
		t.Errorf("FearDecayRate = %v, want %v", config.FearDecayRate, customConfig.FearDecayRate)
	}
	if config.BreakdownThreshold != customConfig.BreakdownThreshold {
		t.Errorf("BreakdownThreshold = %v, want %v", config.BreakdownThreshold, customConfig.BreakdownThreshold)
	}
}

// TestNPCManager_Concurrency tests thread safety of NPCManager
func TestNPCManager_Concurrency(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())
	var wg sync.WaitGroup

	// Number of concurrent operations
	numOps := 100

	// Concurrent writes (AddNPC)
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			profile := &NPCProfile{
				ID:             fmt.Sprintf("npc_%d", id),
				Name:           fmt.Sprintf("NPC %d", id),
				Archetype:      "survivor",
				InitialEmotion: DefaultEmotionState(),
			}
			mgr.AddNPC(profile)
		}(i)
	}

	// Concurrent reads (GetProfile)
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mgr.GetProfile(fmt.Sprintf("npc_%d", id))
		}(i)
	}

	// Concurrent state reads (GetState)
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mgr.GetState(fmt.Sprintf("npc_%d", id))
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Verify all NPCs were added
	ids := mgr.ListNPCIDs()
	if len(ids) != numOps {
		t.Errorf("concurrent AddNPC resulted in %d NPCs, want %d", len(ids), numOps)
	}
}

// TestNPCManager_ConcurrentUpdates tests concurrent state updates
func TestNPCManager_ConcurrentUpdates(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Add initial NPC
	mgr.AddNPC(&NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	})

	var wg sync.WaitGroup
	numUpdates := 50

	// Concurrent state updates
	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(trust int) {
			defer wg.Done()
			state := NewNPCRuntimeState()
			state.Emotion = NewEmotionState(trust, 20, 30)
			mgr.UpdateState("npc1", state)
		}(i)
	}

	wg.Wait()

	// Verify state still exists and is valid
	state := mgr.GetState("npc1")
	if state == nil {
		t.Fatal("state was lost during concurrent updates")
	}
}

// TestNPCManager_MixedOperations tests mixed concurrent read/write operations
func TestNPCManager_MixedOperations(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())
	var wg sync.WaitGroup

	numOps := 50

	// Pre-populate with some NPCs
	for i := 0; i < 10; i++ {
		mgr.AddNPC(&NPCProfile{
			ID:             fmt.Sprintf("npc_%d", i),
			Name:           fmt.Sprintf("NPC %d", i),
			Archetype:      "survivor",
			InitialEmotion: DefaultEmotionState(),
		})
	}

	// Mix of operations
	for i := 0; i < numOps; i++ {
		// Add
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mgr.AddNPC(&NPCProfile{
				ID:             fmt.Sprintf("new_npc_%d", id),
				Name:           fmt.Sprintf("New NPC %d", id),
				Archetype:      "survivor",
				InitialEmotion: DefaultEmotionState(),
			})
		}(i)

		// Read profile
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mgr.GetProfile(fmt.Sprintf("npc_%d", id%10))
		}(i)

		// Read state
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mgr.GetState(fmt.Sprintf("npc_%d", id%10))
		}(i)

		// List
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.ListNPCIDs()
		}()
	}

	wg.Wait()

	// Verify manager is still functional
	ids := mgr.ListNPCIDs()
	if len(ids) < 10 {
		t.Errorf("manager has %d NPCs, expected at least 10", len(ids))
	}
}

// Benchmark tests
func BenchmarkAddNPC(b *testing.B) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		profile := &NPCProfile{
			ID:             fmt.Sprintf("npc_%d", i),
			Name:           "Benchmark NPC",
			Archetype:      "survivor",
			InitialEmotion: DefaultEmotionState(),
		}
		mgr.AddNPC(profile)
	}
}

func BenchmarkGetProfile(b *testing.B) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())
	mgr.AddNPC(&NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mgr.GetProfile("npc1")
	}
}

func BenchmarkUpdateState(b *testing.B) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())
	mgr.AddNPC(&NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		Archetype:      "survivor",
		InitialEmotion: DefaultEmotionState(),
	})
	state := NewNPCRuntimeState()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mgr.UpdateState("npc1", state)
	}
}

// TestAdjustEmotion_Basic tests basic emotion adjustment
func TestAdjustEmotion_Basic(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 30, 40),
		Traits:         []Trait{},
	}

	mgr.AddNPC(profile)

	// Apply a delta
	delta := EmotionDelta{Trust: 10, Fear: -10, Stress: 5}
	err := mgr.AdjustEmotion("npc1", delta)
	if err != nil {
		t.Fatalf("Failed to adjust emotion: %v", err)
	}

	// Verify emotion updated
	state := mgr.GetState("npc1")
	if state.Emotion.Trust != 60 {
		t.Errorf("Expected Trust=60, got %d", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 20 {
		t.Errorf("Expected Fear=20, got %d", state.Emotion.Fear)
	}
	if state.Emotion.Stress != 45 {
		t.Errorf("Expected Stress=45, got %d", state.Emotion.Stress)
	}
}

// TestAdjustEmotion_Clamping tests emotion value clamping
func TestAdjustEmotion_Clamping(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(95, 5, 50),
		Traits:         []Trait{},
	}

	mgr.AddNPC(profile)

	// Apply delta that would exceed bounds
	delta := EmotionDelta{Trust: 20, Fear: -20, Stress: 100}
	err := mgr.AdjustEmotion("npc1", delta)
	if err != nil {
		t.Fatalf("Failed to adjust emotion: %v", err)
	}

	state := mgr.GetState("npc1")
	if state.Emotion.Trust != 100 {
		t.Errorf("Expected Trust clamped to 100, got %d", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 0 {
		t.Errorf("Expected Fear clamped to 0, got %d", state.Emotion.Fear)
	}
	if state.Emotion.Stress != 100 {
		t.Errorf("Expected Stress clamped to 100, got %d", state.Emotion.Stress)
	}
}

// TestAdjustEmotion_RelationshipUpdate tests relationship recalculation
func TestAdjustEmotion_RelationshipUpdate(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 50, 40),
		Traits:         []Trait{},
	}

	mgr.AddNPC(profile)

	// Adjust to friendly (Trust >= 60, Fear < 40)
	delta := EmotionDelta{Trust: 20, Fear: -30, Stress: 0}
	mgr.AdjustEmotion("npc1", delta)

	state := mgr.GetState("npc1")
	if state.Relationship != Friendly {
		t.Errorf("Expected Friendly relationship, got %v", state.Relationship)
	}

	// Verify relationship score also updated
	expectedScore := CalculateRelationshipScore(state.Emotion)
	if state.RelationshipScore != expectedScore {
		t.Errorf("Relationship score mismatch: expected %d, got %d", expectedScore, state.RelationshipScore)
	}
}

// TestAdjustEmotion_AllRelationshipTypes tests all relationship type transitions
func TestAdjustEmotion_AllRelationshipTypes(t *testing.T) {
	tests := []struct {
		name         string
		finalEmotion EmotionState
		expected     RelationshipType
	}{
		{
			name:         "Friendly",
			finalEmotion: NewEmotionState(70, 20, 30),
			expected:     Friendly,
		},
		{
			name:         "Fearful",
			finalEmotion: NewEmotionState(40, 70, 50),
			expected:     Fearful,
		},
		{
			name:         "Hostile",
			finalEmotion: NewEmotionState(20, 30, 40),
			expected:     Hostile,
		},
		{
			name:         "Neutral",
			finalEmotion: NewEmotionState(50, 50, 40),
			expected:     Neutral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

			profile := &NPCProfile{
				ID:             "npc1",
				Name:           "Test NPC",
				InitialEmotion: NewEmotionState(50, 50, 50),
				Traits:         []Trait{},
			}

			mgr.AddNPC(profile)

			// Calculate delta to reach target emotion
			delta := EmotionDelta{
				Trust:  tt.finalEmotion.Trust - 50,
				Fear:   tt.finalEmotion.Fear - 50,
				Stress: tt.finalEmotion.Stress - 50,
			}

			mgr.AdjustEmotion("npc1", delta)

			state := mgr.GetState("npc1")
			if state.Relationship != tt.expected {
				t.Errorf("Expected %v relationship, got %v", tt.expected, state.Relationship)
			}
		})
	}
}

// TestAdjustEmotion_MentalStateTransition_NormalToAnxious tests Normal->Anxious transition
func TestAdjustEmotion_MentalStateTransition_NormalToAnxious(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 50, 50),
		Traits:         []Trait{},
	}

	mgr.AddNPC(profile)

	// Initial state should be Normal
	state := mgr.GetState("npc1")
	if state.MentalState != Normal {
		t.Errorf("Expected initial state Normal, got %v", state.MentalState)
	}

	// Increase stress to trigger Anxious (Stress >= 60)
	delta := EmotionDelta{Trust: 0, Fear: 0, Stress: 15}
	mgr.AdjustEmotion("npc1", delta)

	state = mgr.GetState("npc1")
	if state.MentalState != Anxious {
		t.Errorf("Expected Anxious state at Stress=65, got %v", state.MentalState)
	}
}

// TestAdjustEmotion_MentalStateTransition_AnxiousToNormal tests Anxious->Normal recovery
func TestAdjustEmotion_MentalStateTransition_AnxiousToNormal(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 60, 70), // Will be Anxious
		Traits:         []Trait{},
	}

	mgr.AddNPC(profile)

	state := mgr.GetState("npc1")
	if state.MentalState != Anxious {
		t.Fatalf("Expected initial Anxious state, got %v", state.MentalState)
	}

	// Reduce stress and fear to trigger recovery (Stress < 40 AND Fear < 50)
	delta := EmotionDelta{Trust: 0, Fear: -20, Stress: -35}
	mgr.AdjustEmotion("npc1", delta)

	state = mgr.GetState("npc1")
	if state.MentalState != Normal {
		t.Errorf("Expected Normal state after recovery, got %v", state.MentalState)
	}
}

// TestAdjustEmotion_NotFound tests error handling for nonexistent NPC
func TestAdjustEmotion_NotFound(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	delta := EmotionDelta{Trust: 10, Fear: 0, Stress: 0}
	err := mgr.AdjustEmotion("nonexistent", delta)
	if err == nil {
		t.Error("Expected error when adjusting emotion for nonexistent NPC")
	}
}

// TestAdjustEmotion_PredefinedDeltas tests using predefined emotion deltas
func TestAdjustEmotion_PredefinedDeltas(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 30, 40),
		Traits:         []Trait{},
	}

	mgr.AddNPC(profile)

	tests := []struct {
		name      string
		deltaKey  string
		expectErr bool
	}{
		{"friendly_chat", "friendly_chat", false},
		{"threat", "threat", false},
		{"help_npc", "help_npc", false},
		{"witness_death", "witness_death", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := EmotionDeltas[tt.deltaKey]
			initialState := mgr.GetState("npc1").Emotion.Copy()

			err := mgr.AdjustEmotion("npc1", delta)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error=%v, got %v", tt.expectErr, err)
			}

			if !tt.expectErr {
				state := mgr.GetState("npc1")

				// Verify the delta was applied correctly
				expectedTrust := clamp(initialState.Trust + delta.Trust)
				expectedFear := clamp(initialState.Fear + delta.Fear)
				expectedStress := clamp(initialState.Stress + delta.Stress)

				if state.Emotion.Trust != expectedTrust {
					t.Errorf("Trust mismatch: expected %d, got %d", expectedTrust, state.Emotion.Trust)
				}
				if state.Emotion.Fear != expectedFear {
					t.Errorf("Fear mismatch: expected %d, got %d", expectedFear, state.Emotion.Fear)
				}
				if state.Emotion.Stress != expectedStress {
					t.Errorf("Stress mismatch: expected %d, got %d", expectedStress, state.Emotion.Stress)
				}
			}
		})
	}
}

// TestAdjustEmotion_ConcurrentAccess tests thread safety
func TestAdjustEmotion_ConcurrentAccess(t *testing.T) {
	mgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	// Add test NPC
	profile := &NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: NewEmotionState(50, 50, 50),
		Traits:         []Trait{},
	}
	mgr.AddNPC(profile)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent emotion adjustments
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			delta := EmotionDelta{Trust: 1, Fear: -1, Stress: 0}
			mgr.AdjustEmotion("npc1", delta)
		}()
	}

	wg.Wait()

	// Verify NPC still exists and is in valid state
	state := mgr.GetState("npc1")
	if state == nil {
		t.Fatal("NPC state lost during concurrent access")
	}

	// All values should be within valid range
	if state.Emotion.Trust < 0 || state.Emotion.Trust > 100 {
		t.Errorf("Trust out of range: %d", state.Emotion.Trust)
	}
	if state.Emotion.Fear < 0 || state.Emotion.Fear > 100 {
		t.Errorf("Fear out of range: %d", state.Emotion.Fear)
	}
	if state.Emotion.Stress < 0 || state.Emotion.Stress > 100 {
		t.Errorf("Stress out of range: %d", state.Emotion.Stress)
	}
}
