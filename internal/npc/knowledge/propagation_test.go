package knowledge

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegisterFact tests the RegisterFact method
func TestRegisterFact(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("register fact with ID", func(t *testing.T) {
		fact := &Fact{
			ID:      "fact_001",
			Content: "張醫生死了",
			Type:    Event,
		}

		m.RegisterFact(fact)

		// Verify fact is registered
		registered := m.GetGlobalFact("fact_001")
		assert.NotNil(t, registered)
		assert.Equal(t, "fact_001", registered.ID)
		assert.Equal(t, "張醫生死了", registered.Content)
	})

	t.Run("register fact without ID generates ID", func(t *testing.T) {
		fact := &Fact{
			Content: "怪物出現了",
			Type:    Event,
		}

		m.RegisterFact(fact)

		// Verify fact has generated ID
		assert.NotEmpty(t, fact.ID)

		// Verify fact is registered
		registered := m.GetGlobalFact(fact.ID)
		assert.NotNil(t, registered)
		assert.Equal(t, "怪物出現了", registered.Content)
	})

	t.Run("thread safety", func(t *testing.T) {
		// Launch multiple goroutines to register facts concurrently
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				fact := &Fact{
					Content: "事實",
					Type:    Event,
				}
				m.RegisterFact(fact)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// All facts should be registered
		allFacts := m.GetAllFacts()
		assert.GreaterOrEqual(t, len(allFacts), 10)
	})
}

// TestPropagateEvent tests the PropagateEvent method
func TestPropagateEvent(t *testing.T) {
	m := NewUpdateManager(nil)

	// Setup room occupants
	m.SetEntityRoom("player", "lobby")
	m.SetEntityRoom("npc_001", "lobby")
	m.SetEntityRoom("npc_002", "corridor")

	t.Run("propagates to same room entities", func(t *testing.T) {
		event := &GameEvent{
			Description: "怪物出現了",
			Initiator:   "system",
			Location:    "lobby",
			Beat:        10,
			Importance:  8,
		}

		m.PropagateEvent(event)

		// Verify player knows about it
		playerKB := m.GetKnowledgeBase("player")
		require.NotNil(t, playerKB)
		assert.Len(t, playerKB.KnownFacts, 1)

		var factID string
		for fid := range playerKB.KnownFacts {
			factID = fid
			break
		}

		kf := playerKB.KnownFacts[factID]
		assert.Equal(t, Witness, kf.LearnMethod)
		assert.Equal(t, 1.0, kf.Confidence)
		assert.Equal(t, 0, kf.PropagationDepth)

		// Verify npc_001 knows about it
		npcKB := m.GetKnowledgeBase("npc_001")
		require.NotNil(t, npcKB)
		assert.Len(t, npcKB.KnownFacts, 1)
	})

	t.Run("does not propagate to different room", func(t *testing.T) {
		// npc_002 is in corridor, should not know about lobby event
		npcKB := m.GetKnowledgeBase("npc_002")
		if npcKB != nil {
			assert.Len(t, npcKB.KnownFacts, 0)
		}
	})

	t.Run("fact is registered globally", func(t *testing.T) {
		allFacts := m.GetAllFacts()
		assert.GreaterOrEqual(t, len(allFacts), 1)

		// Find the event fact
		var found bool
		for _, fact := range allFacts {
			if fact.Content == "怪物出現了" {
				found = true
				assert.Equal(t, Event, fact.Type)
				assert.Equal(t, "system", fact.Source)
				assert.Equal(t, "lobby", fact.Location)
				assert.Len(t, fact.Witnesses, 2) // player + npc_001
				break
			}
		}
		assert.True(t, found, "event fact should be in global facts")
	})
}

// TestLearnFromDialogue tests the LearnFromDialogue method
func TestLearnFromDialogue(t *testing.T) {
	m := NewUpdateManager(nil)

	m.SetEntityRoom("player", "lobby")
	m.SetEntityRoom("npc_001", "lobby")
	m.SetEntityRoom("npc_002", "corridor")

	t.Run("same room dialogue learning", func(t *testing.T) {
		// npc_001 hears player speak in lobby
		m.LearnFromDialogue("npc_001", "player", "我看到密道了", "lobby")

		// Verify npc_001 learned
		npcKB := m.GetKnowledgeBase("npc_001")
		require.NotNil(t, npcKB)
		assert.Len(t, npcKB.KnownFacts, 1)

		var factID string
		for fid := range npcKB.KnownFacts {
			factID = fid
			break
		}

		kf := npcKB.KnownFacts[factID]
		assert.Equal(t, Told, kf.LearnMethod)
		assert.Equal(t, 0.9, kf.Confidence)
		assert.Equal(t, 1, kf.PropagationDepth)
		assert.Equal(t, "player", kf.LearnedFrom)
	})

	t.Run("different room dialogue not heard", func(t *testing.T) {
		// npc_002 is in corridor, cannot hear lobby dialogue
		m.LearnFromDialogue("npc_002", "player", "秘密訊息", "lobby")

		// Verify npc_002 did not learn
		npcKB := m.GetKnowledgeBase("npc_002")
		if npcKB != nil {
			assert.Len(t, npcKB.KnownFacts, 0)
		}
	})

	t.Run("dialogue creates fact in global repository", func(t *testing.T) {
		allFacts := m.GetAllFacts()

		// Find dialogue fact
		var found bool
		for _, fact := range allFacts {
			if fact.Content == "我看到密道了" {
				found = true
				assert.Equal(t, Dialogue, fact.Type)
				assert.Equal(t, "player", fact.Source)
				break
			}
		}
		assert.True(t, found, "dialogue should create a global fact")
	})
}

// TestTellNPC tests the TellNPC method
func TestTellNPC(t *testing.T) {
	m := NewUpdateManager(nil)

	// Create a fact and give it to npc_001 as witness
	fact := &Fact{
		ID:      "fact_001",
		Content: "張醫生死了",
		Type:    Event,
	}
	m.RegisterFact(fact)

	// npc_001 witnessed this fact
	m.mu.Lock()
	m.ensureKnowledgeBaseLocked("npc_001")
	m.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:           "fact_001",
		LearnedAt:        time.Now(),
		LearnMethod:      Witness,
		Confidence:       1.0,
		PropagationDepth: 0,
	}
	m.mu.Unlock()

	t.Run("tell NPC succeeds with confidence decay", func(t *testing.T) {
		// npc_001 tells npc_002
		err := m.TellNPC("npc_001", "npc_002", "fact_001")
		assert.NoError(t, err)

		// Verify npc_002 learned with decayed confidence
		npc2KB := m.GetKnowledgeBase("npc_002")
		require.NotNil(t, npc2KB)
		require.Contains(t, npc2KB.KnownFacts, "fact_001")

		kf := npc2KB.KnownFacts["fact_001"]
		assert.Equal(t, Told, kf.LearnMethod)
		assert.InDelta(t, 0.85, kf.Confidence, 0.001) // 1.0 * 0.85
		assert.Equal(t, 1, kf.PropagationDepth)
		assert.Equal(t, "npc_001", kf.LearnedFrom)
	})

	t.Run("confidence decays multiple times", func(t *testing.T) {
		// npc_002 tells npc_003
		err := m.TellNPC("npc_002", "npc_003", "fact_001")
		assert.NoError(t, err)

		// Verify npc_003 has further decayed confidence
		npc3KB := m.GetKnowledgeBase("npc_003")
		require.NotNil(t, npc3KB)
		require.Contains(t, npc3KB.KnownFacts, "fact_001")

		kf := npc3KB.KnownFacts["fact_001"]
		assert.InDelta(t, 0.7225, kf.Confidence, 0.001) // 0.85 * 0.85
		assert.Equal(t, 2, kf.PropagationDepth)
		assert.Equal(t, "npc_002", kf.LearnedFrom)
	})

	t.Run("max propagation depth limit", func(t *testing.T) {
		// Try to propagate beyond depth limit (default is 3)
		// npc_003 tries to tell npc_004
		err := m.TellNPC("npc_003", "npc_004", "fact_001")
		assert.NoError(t, err) // depth 2->3 should succeed

		// npc_004 tries to tell npc_005
		err = m.TellNPC("npc_004", "npc_005", "fact_001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max propagation depth")
	})

	t.Run("error when teller does not know fact", func(t *testing.T) {
		err := m.TellNPC("npc_999", "npc_002", "fact_001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no knowledge base")
	})

	t.Run("error when fact does not exist", func(t *testing.T) {
		err := m.TellNPC("npc_001", "npc_002", "nonexistent_fact")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not know fact")
	})
}

// TestTellNPCWithDistortion tests information distortion
func TestTellNPCWithDistortion(t *testing.T) {
	config := DefaultUpdateManagerConfig()
	config.EnableDistortion = true
	config.DistortionRate = 1.0 // 100% distortion for testing
	m := NewUpdateManager(config)

	// Create fact and give to npc_001
	fact := &Fact{
		ID:      "fact_distort",
		Content: "原始內容",
		Type:    Event,
	}
	m.RegisterFact(fact)

	m.mu.Lock()
	m.ensureKnowledgeBaseLocked("npc_001")
	m.npcKnowledge["npc_001"].KnownFacts["fact_distort"] = &KnownFact{
		FactID:           "fact_distort",
		LearnedAt:        time.Now(),
		LearnMethod:      Witness,
		Confidence:       1.0,
		PropagationDepth: 0,
	}
	m.mu.Unlock()

	t.Run("distortion occurs during telling", func(t *testing.T) {
		err := m.TellNPC("npc_001", "npc_002", "fact_distort")
		assert.NoError(t, err)

		npc2KB := m.GetKnowledgeBase("npc_002")
		require.NotNil(t, npc2KB)

		kf := npc2KB.KnownFacts["fact_distort"]
		// With 100% distortion rate, it should be distorted
		assert.True(t, kf.IsDistorted)
		assert.NotEmpty(t, kf.DistortedContent)
		assert.NotEqual(t, fact.Content, kf.DistortedContent)
	})
}

// TestAddKnowledge tests the addKnowledge internal method
func TestAddKnowledge(t *testing.T) {
	m := NewUpdateManager(nil)

	fact := &Fact{
		ID:      "fact_test",
		Content: "測試事實",
		Type:    Event,
		Source:  "system",
	}
	m.RegisterFact(fact)

	t.Run("adds new knowledge", func(t *testing.T) {
		m.mu.Lock()
		m.addKnowledge("npc_001", fact, Witness, 1.0, 0)
		m.mu.Unlock()

		kb := m.GetKnowledgeBase("npc_001")
		require.NotNil(t, kb)
		require.Contains(t, kb.KnownFacts, "fact_test")

		kf := kb.KnownFacts["fact_test"]
		assert.Equal(t, 1.0, kf.Confidence)
		assert.Equal(t, Witness, kf.LearnMethod)
	})

	t.Run("updates knowledge with higher confidence", func(t *testing.T) {
		// Add same fact with higher confidence
		m.mu.Lock()
		m.addKnowledge("npc_001", fact, Told, 1.0, 1) // Same confidence as before
		m.mu.Unlock()

		kb := m.GetKnowledgeBase("npc_001")
		kf := kb.KnownFacts["fact_test"]
		// Should NOT update because confidence is not strictly higher
		assert.Equal(t, Witness, kf.LearnMethod) // Should still be Witness
		assert.Equal(t, 1.0, kf.Confidence)

		// Now add with truly higher confidence (impossible since max is 1.0, so use different NPC)
		m.mu.Lock()
		m.addKnowledge("npc_002", fact, Witness, 0.8, 0)
		m.addKnowledge("npc_002", fact, Told, 0.95, 1) // Higher confidence
		m.mu.Unlock()

		kb2 := m.GetKnowledgeBase("npc_002")
		kf2 := kb2.KnownFacts["fact_test"]
		// Should update to Told since confidence is higher (0.95 > 0.8)
		assert.Equal(t, Told, kf2.LearnMethod)
		assert.Equal(t, 0.95, kf2.Confidence)
	})

	t.Run("does not update with lower confidence", func(t *testing.T) {
		// Try to add with lower confidence to npc_001 (has 1.0 confidence)
		m.mu.Lock()
		m.addKnowledge("npc_001", fact, Overheard, 0.5, 2)
		m.mu.Unlock()

		kb := m.GetKnowledgeBase("npc_001")
		kf := kb.KnownFacts["fact_test"]
		// Should remain Witness with confidence 1.0
		assert.Equal(t, Witness, kf.LearnMethod)
		assert.Equal(t, 1.0, kf.Confidence)
	})

	t.Run("creates knowledge base if not exists", func(t *testing.T) {
		m.mu.Lock()
		m.addKnowledge("npc_new", fact, Witness, 0.9, 0)
		m.mu.Unlock()

		kb := m.GetKnowledgeBase("npc_new")
		assert.NotNil(t, kb)
		assert.Contains(t, kb.KnownFacts, "fact_test")
	})
}

// TestConcurrency tests thread safety of propagation methods
func TestPropagationConcurrency(t *testing.T) {
	m := NewUpdateManager(nil)

	// Setup rooms
	m.SetEntityRoom("player", "lobby")
	for i := 0; i < 10; i++ {
		m.SetEntityRoom(fmt.Sprintf("npc_%d", i), "lobby")
	}

	done := make(chan bool)

	// Concurrent event propagation
	for i := 0; i < 5; i++ {
		go func(idx int) {
			event := &GameEvent{
				Description: fmt.Sprintf("事件 %d", idx),
				Initiator:   "system",
				Location:    "lobby",
				Beat:        idx,
				Importance:  5,
			}
			m.PropagateEvent(event)
			done <- true
		}(i)
	}

	// Concurrent dialogue learning
	for i := 0; i < 5; i++ {
		go func(idx int) {
			m.LearnFromDialogue(fmt.Sprintf("npc_%d", idx), "player", fmt.Sprintf("對話 %d", idx), "lobby")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all facts were registered
	allFacts := m.GetAllFacts()
	assert.GreaterOrEqual(t, len(allFacts), 10)
}

// TestEnsureKnowledgeBase tests knowledge base creation
func TestEnsureKnowledgeBase(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("creates NPC knowledge base", func(t *testing.T) {
		m.mu.Lock()
		m.ensureKnowledgeBaseLocked("npc_test")
		m.mu.Unlock()

		kb := m.GetKnowledgeBase("npc_test")
		assert.NotNil(t, kb)
		assert.Equal(t, "npc_test", kb.OwnerID)
	})

	t.Run("player knowledge base exists by default", func(t *testing.T) {
		kb := m.GetKnowledgeBase("player")
		assert.NotNil(t, kb)
		assert.Equal(t, "player", kb.OwnerID)
	})
}

// TestDistortFact tests the simple distortion logic
func TestDistortFact(t *testing.T) {
	m := NewUpdateManager(nil)

	original := "張醫生死了"
	distorted := m.distortFact(original)

	// Should produce some variation
	assert.NotEmpty(t, distorted)
	// Simple distortion should contain original content
	assert.Contains(t, distorted, original)
}

// TestGetCurrentBeat tests the placeholder getCurrentBeat
func TestGetCurrentBeat(t *testing.T) {
	m := NewUpdateManager(nil)

	beat := m.getCurrentBeat()
	// Placeholder returns 0
	assert.Equal(t, 0, beat)
}

// TestGenerateFactID tests fact ID generation
func TestGenerateFactID(t *testing.T) {
	id1 := generateFactID()
	id2 := generateFactID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "IDs should be unique")
	assert.Contains(t, id1, "fact_")
}

// TestGetKnowledgeBase tests public knowledge base retrieval
func TestGetKnowledgeBase(t *testing.T) {
	m := NewUpdateManager(nil)

	// Add some knowledge
	fact := &Fact{
		ID:      "test_kb",
		Content: "測試",
		Type:    Event,
	}
	m.RegisterFact(fact)

	m.mu.Lock()
	m.addKnowledge("npc_001", fact, Witness, 1.0, 0)
	m.mu.Unlock()

	t.Run("returns copy of knowledge base", func(t *testing.T) {
		kb := m.GetKnowledgeBase("npc_001")
		require.NotNil(t, kb)
		assert.Contains(t, kb.KnownFacts, "test_kb")

		// Modifying copy should not affect original
		delete(kb.KnownFacts, "test_kb")

		kb2 := m.GetKnowledgeBase("npc_001")
		assert.Contains(t, kb2.KnownFacts, "test_kb", "original should be unmodified")
	})

	t.Run("returns nil for non-existent NPC", func(t *testing.T) {
		kb := m.GetKnowledgeBase("nonexistent")
		assert.Nil(t, kb)
	})
}

// TestGetGlobalFact tests global fact retrieval
func TestGetGlobalFact(t *testing.T) {
	m := NewUpdateManager(nil)

	fact := &Fact{
		ID:      "global_test",
		Content: "全域事實",
		Type:    Event,
	}
	m.RegisterFact(fact)

	t.Run("retrieves fact copy", func(t *testing.T) {
		retrieved := m.GetGlobalFact("global_test")
		require.NotNil(t, retrieved)
		assert.Equal(t, "global_test", retrieved.ID)
		assert.Equal(t, "全域事實", retrieved.Content)

		// Modifying copy should not affect original
		retrieved.Content = "修改後"

		retrieved2 := m.GetGlobalFact("global_test")
		assert.Equal(t, "全域事實", retrieved2.Content, "original should be unmodified")
	})

	t.Run("returns nil for non-existent fact", func(t *testing.T) {
		fact := m.GetGlobalFact("nonexistent")
		assert.Nil(t, fact)
	})
}

// TestGetAllFacts tests retrieving all global facts
func TestGetAllFacts(t *testing.T) {
	m := NewUpdateManager(nil)

	// Register multiple facts
	for i := 0; i < 5; i++ {
		fact := &Fact{
			ID:      fmt.Sprintf("fact_%d", i),
			Content: fmt.Sprintf("內容 %d", i),
			Type:    Event,
		}
		m.RegisterFact(fact)
	}

	facts := m.GetAllFacts()
	assert.Len(t, facts, 5)
}

// TestLoadFacts tests loading facts from serialization
func TestLoadFacts(t *testing.T) {
	m := NewUpdateManager(nil)

	facts := []*Fact{
		{ID: "load_1", Content: "事實1", Type: Event},
		{ID: "load_2", Content: "事實2", Type: Dialogue},
	}

	m.LoadFacts(facts)

	// Verify facts are loaded
	fact1 := m.GetGlobalFact("load_1")
	assert.NotNil(t, fact1)
	assert.Equal(t, "事實1", fact1.Content)

	fact2 := m.GetGlobalFact("load_2")
	assert.NotNil(t, fact2)
	assert.Equal(t, "事實2", fact2.Content)
}
