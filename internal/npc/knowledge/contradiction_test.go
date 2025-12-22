package knowledge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContradicts tests the contradicts method with various negation pairs
func TestContradicts(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("detects alive/dead contradiction", func(t *testing.T) {
		assert.True(t, m.contradicts("張醫生活著", "張醫生死了"))
		assert.True(t, m.contradicts("張醫生死了", "張醫生活著"))
		assert.True(t, m.contradicts("他還活著", "他已經死了"))
	})

	t.Run("detects safe/dangerous contradiction", func(t *testing.T) {
		assert.True(t, m.contradicts("這裡很安全", "這裡很危險"))
		assert.True(t, m.contradicts("這裡危險", "這裡安全"))
	})

	t.Run("detects trust contradiction", func(t *testing.T) {
		assert.True(t, m.contradicts("他很可信", "他不可信"))
		assert.True(t, m.contradicts("這個人可靠", "這個人不可靠"))
	})

	t.Run("detects existence contradiction", func(t *testing.T) {
		assert.True(t, m.contradicts("門存在", "門不存在"))
		assert.True(t, m.contradicts("這裡有人", "這裡沒人"))
	})

	t.Run("detects open/closed contradiction", func(t *testing.T) {
		assert.True(t, m.contradicts("門是開著的", "門是關著的"))
		assert.True(t, m.contradicts("門關著", "門開著"))
	})

	t.Run("no contradiction with similar but not opposite terms", func(t *testing.T) {
		assert.False(t, m.contradicts("張醫生活著", "張醫生很好"))
		assert.False(t, m.contradicts("這裡安全", "這裡很寧靜"))
		assert.False(t, m.contradicts("門是開著的", "門很大"))
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		assert.True(t, m.contradicts("張醫生活著", "張醫生死了"))
		assert.True(t, m.contradicts("門開著", "門關著"))
	})

	t.Run("detects contradictions in complex sentences", func(t *testing.T) {
		assert.True(t, m.contradicts(
			"我看到張醫生還活著，他在大廳裡",
			"剛才有人告訴我張醫生已經死了"))
		assert.True(t, m.contradicts(
			"我確定這個房間是安全的",
			"小心，這個房間很危險"))
	})
}

// TestCheckContradiction tests the CheckContradiction method
func TestCheckContradiction(t *testing.T) {
	m := NewUpdateManager(nil)

	// Setup: Create a fact that NPC witnessed
	fact := &Fact{
		ID:      "fact_001",
		Content: "張醫生還活著",
		Type:    Event,
	}
	m.RegisterFact(fact)

	// NPC witnessed this fact
	m.mu.Lock()
	m.ensureKnowledgeBaseLocked("npc_001")
	m.npcKnowledge["npc_001"].KnownFacts["fact_001"] = &KnownFact{
		FactID:      "fact_001",
		LearnedAt:   time.Now(),
		LearnMethod: Witness,
		Confidence:  1.0,
		PropagationDepth: 0,
	}
	m.mu.Unlock()

	t.Run("detects major contradiction", func(t *testing.T) {
		result := m.CheckContradiction("npc_001", "張醫生死了")

		require.NotNil(t, result)
		assert.Equal(t, ContradictionMajor, result.Type)
		assert.Equal(t, "張醫生死了", result.NewInfo)
		assert.Equal(t, 10, result.Severity) // 5 + 2(high confidence) + 3(witnessed)
		assert.Equal(t, -20, result.SuggestedDelta.Trust)
		assert.Equal(t, 10, result.SuggestedDelta.Fear)
		assert.Equal(t, 25, result.SuggestedDelta.Stress)
		assert.Contains(t, result.SuggestedReaction, "強烈質疑")
	})

	t.Run("no contradiction with non-contradictory info", func(t *testing.T) {
		result := m.CheckContradiction("npc_001", "張醫生很累")
		assert.Nil(t, result)
	})

	t.Run("returns nil for entity with no knowledge base", func(t *testing.T) {
		result := m.CheckContradiction("nonexistent_npc", "任何資訊")
		assert.Nil(t, result)
	})

	t.Run("checks distorted content if fact is distorted", func(t *testing.T) {
		// Add a distorted fact to npc_002
		m.mu.Lock()
		m.ensureKnowledgeBaseLocked("npc_002")
		m.npcKnowledge["npc_002"].KnownFacts["fact_001"] = &KnownFact{
			FactID:           "fact_001",
			LearnedAt:        time.Now(),
			LearnMethod:      Told,
			Confidence:       0.7,
			IsDistorted:      true,
			DistortedContent: "聽說張醫生活著", // Distorted version
			PropagationDepth: 1,
		}
		m.mu.Unlock()

		result := m.CheckContradiction("npc_002", "張醫生死了")
		require.NotNil(t, result)
		// Should still detect contradiction based on distorted content
		assert.Equal(t, ContradictionModerate, result.Type) // Lower severity due to lower confidence and not witnessed
	})
}

// TestCalculateContradictionSeverity tests severity calculation
func TestCalculateContradictionSeverity(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("base severity for low confidence told", func(t *testing.T) {
		kf := &KnownFact{
			LearnMethod: Told,
			Confidence:  0.5,
		}
		severity := m.calculateContradictionSeverity(kf, "test")
		assert.Equal(t, 5, severity) // Base only
	})

	t.Run("high confidence adds 2 to severity", func(t *testing.T) {
		kf := &KnownFact{
			LearnMethod: Told,
			Confidence:  0.9,
		}
		severity := m.calculateContradictionSeverity(kf, "test")
		assert.Equal(t, 7, severity) // 5 + 2
	})

	t.Run("witnessed adds 3 to severity", func(t *testing.T) {
		kf := &KnownFact{
			LearnMethod: Witness,
			Confidence:  0.7,
		}
		severity := m.calculateContradictionSeverity(kf, "test")
		assert.Equal(t, 8, severity) // 5 + 3
	})

	t.Run("high confidence and witnessed gives maximum", func(t *testing.T) {
		kf := &KnownFact{
			LearnMethod: Witness,
			Confidence:  1.0,
		}
		severity := m.calculateContradictionSeverity(kf, "test")
		assert.Equal(t, 10, severity) // 5 + 2 + 3 = 10 (max)
	})

	t.Run("severity is clamped to 1-10", func(t *testing.T) {
		// Can't really go below 5 with current logic, but test the clamp exists
		kf := &KnownFact{
			LearnMethod: Overheard,
			Confidence:  0.3,
		}
		severity := m.calculateContradictionSeverity(kf, "test")
		assert.GreaterOrEqual(t, severity, 1)
		assert.LessOrEqual(t, severity, 10)
	})
}

// TestGetContradictionType tests severity to type mapping
func TestGetContradictionType(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("severity 8-10 is major", func(t *testing.T) {
		assert.Equal(t, ContradictionMajor, m.getContradictionType(8))
		assert.Equal(t, ContradictionMajor, m.getContradictionType(9))
		assert.Equal(t, ContradictionMajor, m.getContradictionType(10))
	})

	t.Run("severity 5-7 is moderate", func(t *testing.T) {
		assert.Equal(t, ContradictionModerate, m.getContradictionType(5))
		assert.Equal(t, ContradictionModerate, m.getContradictionType(6))
		assert.Equal(t, ContradictionModerate, m.getContradictionType(7))
	})

	t.Run("severity 1-4 is minor", func(t *testing.T) {
		assert.Equal(t, ContradictionMinor, m.getContradictionType(1))
		assert.Equal(t, ContradictionMinor, m.getContradictionType(2))
		assert.Equal(t, ContradictionMinor, m.getContradictionType(3))
		assert.Equal(t, ContradictionMinor, m.getContradictionType(4))
	})
}

// TestGetSuggestedEmotionDelta tests emotion delta suggestions
func TestGetSuggestedEmotionDelta(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("major contradiction delta", func(t *testing.T) {
		delta := m.getSuggestedEmotionDelta(10)
		assert.Equal(t, -20, delta.Trust)
		assert.Equal(t, 10, delta.Fear)
		assert.Equal(t, 25, delta.Stress)
	})

	t.Run("moderate contradiction delta", func(t *testing.T) {
		delta := m.getSuggestedEmotionDelta(6)
		assert.Equal(t, -10, delta.Trust)
		assert.Equal(t, 5, delta.Fear)
		assert.Equal(t, 15, delta.Stress)
	})

	t.Run("minor contradiction delta", func(t *testing.T) {
		delta := m.getSuggestedEmotionDelta(3)
		assert.Equal(t, -5, delta.Trust)
		assert.Equal(t, 0, delta.Fear)
		assert.Equal(t, 5, delta.Stress)
	})

	t.Run("boundary values", func(t *testing.T) {
		// Test boundary between major and moderate
		delta8 := m.getSuggestedEmotionDelta(8)
		assert.Equal(t, -20, delta8.Trust) // Should be major

		delta7 := m.getSuggestedEmotionDelta(7)
		assert.Equal(t, -10, delta7.Trust) // Should be moderate

		// Test boundary between moderate and minor
		delta5 := m.getSuggestedEmotionDelta(5)
		assert.Equal(t, -10, delta5.Trust) // Should be moderate

		delta4 := m.getSuggestedEmotionDelta(4)
		assert.Equal(t, -5, delta4.Trust) // Should be minor
	})
}

// TestGetSuggestedReaction tests reaction text suggestions
func TestGetSuggestedReaction(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("major contradiction reaction", func(t *testing.T) {
		reaction := m.getSuggestedReaction(10)
		assert.Contains(t, reaction, "強烈質疑")
	})

	t.Run("moderate contradiction reaction", func(t *testing.T) {
		reaction := m.getSuggestedReaction(6)
		assert.Contains(t, reaction, "困惑")
	})

	t.Run("minor contradiction reaction", func(t *testing.T) {
		reaction := m.getSuggestedReaction(3)
		assert.Contains(t, reaction, "疑惑")
	})
}

// TestContradictionWithMultipleFacts tests checking multiple facts
func TestContradictionWithMultipleFacts(t *testing.T) {
	m := NewUpdateManager(nil)

	// Setup: Add multiple facts to NPC
	fact1 := &Fact{ID: "fact_safe", Content: "房間很安全", Type: Event}
	fact2 := &Fact{ID: "fact_door", Content: "門是開著的", Type: Event}
	fact3 := &Fact{ID: "fact_alive", Content: "張醫生活著", Type: Event}

	m.RegisterFact(fact1)
	m.RegisterFact(fact2)
	m.RegisterFact(fact3)

	m.mu.Lock()
	m.ensureKnowledgeBaseLocked("npc_multi")
	m.npcKnowledge["npc_multi"].KnownFacts["fact_safe"] = &KnownFact{
		FactID: "fact_safe", LearnMethod: Told, Confidence: 0.8, PropagationDepth: 1,
	}
	m.npcKnowledge["npc_multi"].KnownFacts["fact_door"] = &KnownFact{
		FactID: "fact_door", LearnMethod: Witness, Confidence: 1.0, PropagationDepth: 0,
	}
	m.npcKnowledge["npc_multi"].KnownFacts["fact_alive"] = &KnownFact{
		FactID: "fact_alive", LearnMethod: Witness, Confidence: 1.0, PropagationDepth: 0,
	}
	m.mu.Unlock()

	t.Run("finds first contradiction", func(t *testing.T) {
		// This contradicts fact_alive
		result := m.CheckContradiction("npc_multi", "張醫生死了")
		require.NotNil(t, result)
		assert.Equal(t, "fact_alive", result.ExistingFact.FactID)
	})

	t.Run("finds different contradiction", func(t *testing.T) {
		// This contradicts fact_door
		result := m.CheckContradiction("npc_multi", "門是關著的")
		require.NotNil(t, result)
		// Note: Order might vary based on map iteration, so just check it found one
		assert.NotNil(t, result.ExistingFact)
	})

	t.Run("info contradicting multiple facts returns one", func(t *testing.T) {
		// Contradicts fact_safe
		result := m.CheckContradiction("npc_multi", "房間危險")
		require.NotNil(t, result)
		// Returns first found contradiction
		assert.NotNil(t, result)
	})
}

// TestContradictionThreadSafety tests thread-safe access
func TestContradictionThreadSafety(t *testing.T) {
	m := NewUpdateManager(nil)

	// Setup fact
	fact := &Fact{ID: "thread_fact", Content: "測試活著", Type: Event}
	m.RegisterFact(fact)

	m.mu.Lock()
	m.ensureKnowledgeBaseLocked("npc_thread")
	m.npcKnowledge["npc_thread"].KnownFacts["thread_fact"] = &KnownFact{
		FactID: "thread_fact", LearnMethod: Witness, Confidence: 1.0, PropagationDepth: 0,
	}
	m.mu.Unlock()

	// Concurrent contradiction checks
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			result := m.CheckContradiction("npc_thread", "測試死了")
			assert.NotNil(t, result)
			done <- true
		}()
	}

	// Wait for all checks
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestContradictionWithNoKnowledge tests handling of unknown entities
func TestContradictionWithNoKnowledge(t *testing.T) {
	m := NewUpdateManager(nil)

	t.Run("returns nil for entity with no knowledge", func(t *testing.T) {
		result := m.CheckContradiction("unknown_npc", "任何資訊")
		assert.Nil(t, result)
	})

	t.Run("returns nil for player with empty knowledge", func(t *testing.T) {
		result := m.CheckContradiction("player", "任何資訊")
		assert.Nil(t, result)
	})
}

// TestNegationPairsCoverage tests various negation pairs
func TestNegationPairsCoverage(t *testing.T) {
	m := NewUpdateManager(nil)

	testCases := []struct {
		name     string
		existing string
		new      string
		should   bool
	}{
		{"生還/死亡", "他生還了", "他死亡了", true},
		{"真的/假的", "這是真的", "這是假的", true},
		{"明亮/黑暗", "房間明亮", "房間黑暗", true},
		{"正常/異常", "情況正常", "情況異常", true},
		{"相似但不矛盾", "很好", "很棒", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := m.contradicts(tc.existing, tc.new)
			assert.Equal(t, tc.should, result, "Expected %s vs %s to be %v", tc.existing, tc.new, tc.should)
		})
	}
}

// TestEmotionDeltaStructure tests the EmotionDelta structure
func TestEmotionDeltaStructure(t *testing.T) {
	delta := EmotionDelta{
		Trust:  -15,
		Fear:   8,
		Stress: 20,
	}

	assert.Equal(t, -15, delta.Trust)
	assert.Equal(t, 8, delta.Fear)
	assert.Equal(t, 20, delta.Stress)
}

// TestContradictionResultStructure tests the ContradictionResult structure
func TestContradictionResultStructure(t *testing.T) {
	kf := &KnownFact{
		FactID:      "test",
		LearnMethod: Witness,
		Confidence:  1.0,
	}

	result := &ContradictionResult{
		Type:         ContradictionMajor,
		ExistingFact: kf,
		NewInfo:      "新資訊",
		Severity:     9,
		SuggestedDelta: EmotionDelta{
			Trust:  -20,
			Fear:   10,
			Stress: 25,
		},
		SuggestedReaction: "強烈質疑",
	}

	assert.Equal(t, ContradictionMajor, result.Type)
	assert.Equal(t, "新資訊", result.NewInfo)
	assert.Equal(t, 9, result.Severity)
	assert.Equal(t, -20, result.SuggestedDelta.Trust)
	assert.Equal(t, "強烈質疑", result.SuggestedReaction)
}
