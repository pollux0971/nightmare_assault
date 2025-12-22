package knowledge

import (
	"testing"
)

// mockNPCManager is a simple mock implementation of NPCManagerInterface for integration testing.
type mockNPCManager struct {
	emotions map[string]struct{ trust, fear, stress int }
	traits   map[string][]string
}

func newMockNPCManager() *mockNPCManager {
	return &mockNPCManager{
		emotions: make(map[string]struct{ trust, fear, stress int }),
		traits:   make(map[string][]string),
	}
}

func (m *mockNPCManager) GetNPCEmotion(npcID string) (trust, fear, stress int, err error) {
	emotion, exists := m.emotions[npcID]
	if !exists {
		return 0, 0, 0, nil
	}
	return emotion.trust, emotion.fear, emotion.stress, nil
}

func (m *mockNPCManager) GetNPCTraits(npcID string) ([]string, error) {
	traits, exists := m.traits[npcID]
	if !exists {
		return []string{}, nil
	}
	return traits, nil
}

// TestDistortionIntegration_HighFearNPC demonstrates Story 8.3 AC3:
// 恐懼/壓力高的 NPC 更容易扭曲
func TestDistortionIntegration_HighFearNPC(t *testing.T) {
	// Setup
	npcMgr := newMockNPCManager()
	npcMgr.emotions["npc_fearful"] = struct{ trust, fear, stress int }{50, 80, 30}
	npcMgr.emotions["npc_calm"] = struct{ trust, fear, stress int }{50, 20, 20}
	npcMgr.traits["npc_fearful"] = []string{}
	npcMgr.traits["npc_calm"] = []string{}

	distortionCalc := NewDistortionCalculator(&npcManagerAdapter{npcMgr: npcMgr}, nil)

	config := &UpdateManagerConfig{
		EnableDistortion:    true,
		DistortionRate:      0.15,
		MaxPropagationDepth: 3,
		DistortionCalculator: distortionCalc,
	}

	mgr := NewUpdateManager(config)
	mgr.SetNPCManager(npcMgr)

	// Register fact
	fact := NewFact("fact1", "發現了一個房間", Event, "system", "room1", []string{})
	mgr.RegisterFact(fact)

	// Add NPCs to knowledge bases
	mgr.SetEntityRoom("npc_fearful", "room1")
	mgr.SetEntityRoom("npc_calm", "room1")

	// Both NPCs learn the fact as witnesses
	mgr.PropagateEvent(&GameEvent{
		ID:          "fact1",
		Type:        "discovery",
		Description: "發現了一個房間",
		Initiator:   "system",
		Location:    "room1",
		Beat:        1,
		Importance:  5,
	})

	// Now propagate to a third NPC through telling
	mgr.SetEntityRoom("npc_listener", "room2")
	npcMgr.emotions["npc_listener"] = struct{ trust, fear, stress int }{50, 30, 30}
	npcMgr.traits["npc_listener"] = []string{}

	// Fearful NPC tells listener (should have high distortion chance)
	fearfulKB := mgr.GetKnowledgeBase("npc_fearful")
	if fearfulKB == nil || !fearfulKB.HasFact("fact1") {
		t.Fatal("fearful NPC should know fact1")
	}

	// Move listener to same room
	mgr.SetEntityRoom("npc_listener", "room1")

	// Tell the fact
	err := mgr.TellNPC("npc_fearful", "npc_listener", "fact1")
	if err != nil {
		t.Fatalf("TellNPC failed: %v", err)
	}

	// Verify listener received the fact
	listenerKB := mgr.GetKnowledgeBase("npc_listener")
	if listenerKB == nil {
		t.Fatal("listener should have knowledge base")
	}

	if !listenerKB.HasFact("fact1") {
		t.Error("listener should know fact1")
	}

	// With high fear, there's a higher chance of distortion
	// We can't test randomness deterministically, but we can verify the system works
	knownFact := listenerKB.GetFact("fact1")
	if knownFact == nil {
		t.Fatal("listener should have known fact for fact1")
	}

	// Verify propagation depth increased
	if knownFact.PropagationDepth != 1 {
		t.Errorf("expected propagation depth 1, got %d", knownFact.PropagationDepth)
	}

	// Verify confidence decayed
	if knownFact.Confidence > 0.9 {
		t.Errorf("expected confidence decay, got %f", knownFact.Confidence)
	}
}

// TestDistortionIntegration_HighTrustProtection demonstrates Story 8.3 AC4:
// 玩家可透過高信任度獲得準確資訊
func TestDistortionIntegration_HighTrustProtection(t *testing.T) {
	// Setup
	npcMgr := newMockNPCManager()
	npcMgr.emotions["npc_trusted"] = struct{ trust, fear, stress int }{90, 20, 20}
	npcMgr.traits["npc_trusted"] = []string{"reliable"}

	distortionCalc := NewDistortionCalculator(&npcManagerAdapter{npcMgr: npcMgr}, nil)

	config := &UpdateManagerConfig{
		EnableDistortion:    true,
		DistortionRate:      0.15,
		MaxPropagationDepth: 3,
		DistortionCalculator: distortionCalc,
	}

	mgr := NewUpdateManager(config)
	mgr.SetNPCManager(npcMgr)

	// Calculate distortion rate for high trust NPC
	rate, factors, err := distortionCalc.CalculateDistortionRate("npc_trusted", 0)
	if err != nil {
		t.Fatalf("CalculateDistortionRate failed: %v", err)
	}

	// High trust should significantly reduce distortion
	if rate > 0.2 {
		t.Errorf("expected low distortion rate with high trust, got %f", rate)
	}

	// Verify trust reduction is significant
	if factors.TrustReduction < 0.4 {
		t.Errorf("expected high trust reduction factor, got %f", factors.TrustReduction)
	}
}

// TestDistortionIntegration_PropagationDepth demonstrates Story 8.3 AC2:
// 扭曲程度與傳播深度相關
func TestDistortionIntegration_PropagationDepth(t *testing.T) {
	// Setup
	npcMgr := newMockNPCManager()
	npcMgr.emotions["npc1"] = struct{ trust, fear, stress int }{50, 30, 30}
	npcMgr.traits["npc1"] = []string{}

	distortionCalc := NewDistortionCalculator(&npcManagerAdapter{npcMgr: npcMgr}, nil)

	// Calculate distortion rates at different depths
	rate0, _, _ := distortionCalc.CalculateDistortionRate("npc1", 0)
	rate1, _, _ := distortionCalc.CalculateDistortionRate("npc1", 1)
	rate2, _, _ := distortionCalc.CalculateDistortionRate("npc1", 2)
	rate3, _, _ := distortionCalc.CalculateDistortionRate("npc1", 3)

	// Each depth should increase distortion rate
	if rate1 <= rate0 {
		t.Errorf("expected depth 1 rate (%f) > depth 0 rate (%f)", rate1, rate0)
	}
	if rate2 <= rate1 {
		t.Errorf("expected depth 2 rate (%f) > depth 1 rate (%f)", rate2, rate1)
	}
	if rate3 <= rate2 {
		t.Errorf("expected depth 3 rate (%f) > depth 2 rate (%f)", rate3, rate2)
	}

	// Depth multiplier should follow 1.25^depth pattern
	expectedMultiplier := 1.25
	actualMultiplier := rate1 / rate0
	tolerance := 0.01

	if diff := actualMultiplier - expectedMultiplier; diff < -tolerance || diff > tolerance {
		t.Errorf("expected depth multiplier ~%f, got %f", expectedMultiplier, actualMultiplier)
	}
}

// TestDistortionIntegration_TraitModifiers demonstrates Story 8.3 AC1:
// 更智慧的扭曲邏輯（基於 NPC 個性）
func TestDistortionIntegration_TraitModifiers(t *testing.T) {
	// Setup
	npcMgr := newMockNPCManager()

	// NPC with anxious trait
	npcMgr.emotions["npc_anxious"] = struct{ trust, fear, stress int }{50, 30, 30}
	npcMgr.traits["npc_anxious"] = []string{"anxious"}

	// NPC with paranoid trait
	npcMgr.emotions["npc_paranoid"] = struct{ trust, fear, stress int }{50, 30, 30}
	npcMgr.traits["npc_paranoid"] = []string{"paranoid"}

	// NPC with rational trait
	npcMgr.emotions["npc_rational"] = struct{ trust, fear, stress int }{50, 30, 30}
	npcMgr.traits["npc_rational"] = []string{"rational"}

	// NPC with no special traits
	npcMgr.emotions["npc_normal"] = struct{ trust, fear, stress int }{50, 30, 30}
	npcMgr.traits["npc_normal"] = []string{}

	distortionCalc := NewDistortionCalculator(&npcManagerAdapter{npcMgr: npcMgr}, nil)

	rateAnxious, _, _ := distortionCalc.CalculateDistortionRate("npc_anxious", 0)
	rateParanoid, _, _ := distortionCalc.CalculateDistortionRate("npc_paranoid", 0)
	rateRational, _, _ := distortionCalc.CalculateDistortionRate("npc_rational", 0)
	rateNormal, _, _ := distortionCalc.CalculateDistortionRate("npc_normal", 0)

	// Anxious should have higher rate than normal
	if rateAnxious <= rateNormal {
		t.Errorf("expected anxious rate (%f) > normal rate (%f)", rateAnxious, rateNormal)
	}

	// Paranoid should have even higher rate (larger modifier)
	if rateParanoid <= rateAnxious {
		t.Errorf("expected paranoid rate (%f) > anxious rate (%f)", rateParanoid, rateAnxious)
	}

	// Rational should have lower rate than normal
	if rateRational >= rateNormal {
		t.Errorf("expected rational rate (%f) < normal rate (%f)", rateRational, rateNormal)
	}
}

// TestDistortionIntegration_EndToEnd demonstrates the complete distortion flow
func TestDistortionIntegration_EndToEnd(t *testing.T) {
	// Setup a complete scenario
	npcMgr := newMockNPCManager()

	// NPC1: Witness (calm, rational)
	npcMgr.emotions["npc1"] = struct{ trust, fear, stress int }{50, 10, 10}
	npcMgr.traits["npc1"] = []string{"rational"}

	// NPC2: First reteller (anxious, moderate fear)
	npcMgr.emotions["npc2"] = struct{ trust, fear, stress int }{40, 50, 40}
	npcMgr.traits["npc2"] = []string{"anxious"}

	// NPC3: Second reteller (paranoid, high fear)
	npcMgr.emotions["npc3"] = struct{ trust, fear, stress int }{30, 70, 60}
	npcMgr.traits["npc3"] = []string{"paranoid"}

	// Player: High trust with NPC1
	npcMgr.emotions["player"] = struct{ trust, fear, stress int }{80, 20, 20}
	npcMgr.traits["player"] = []string{}

	distortionCalc := NewDistortionCalculator(&npcManagerAdapter{npcMgr: npcMgr}, nil)

	config := &UpdateManagerConfig{
		EnableDistortion:    true,
		DistortionRate:      0.15,
		MaxPropagationDepth: 3,
		DistortionCalculator: distortionCalc,
	}

	mgr := NewUpdateManager(config)
	mgr.SetNPCManager(npcMgr)

	// Setup rooms
	mgr.SetEntityRoom("npc1", "room1")
	mgr.SetEntityRoom("npc2", "room1")
	mgr.SetEntityRoom("npc3", "room2")
	mgr.SetEntityRoom("player", "room2")

	// Event occurs: NPC1 and NPC2 witness it
	mgr.PropagateEvent(&GameEvent{
		ID:          "event1",
		Type:        "discovery",
		Description: "在走廊盡頭發現了一扇古老的門",
		Initiator:   "system",
		Location:    "room1",
		Beat:        1,
		Importance:  7,
	})

	// Verify witnesses learned the fact
	kb1 := mgr.GetKnowledgeBase("npc1")
	kb2 := mgr.GetKnowledgeBase("npc2")

	if !kb1.HasFact("event1") {
		t.Error("npc1 should know event1")
	}
	if !kb2.HasFact("event1") {
		t.Error("npc2 should know event1")
	}

	// NPC2 tells NPC3 (anxious -> paranoid, should increase distortion)
	mgr.SetEntityRoom("npc3", "room1")
	err := mgr.TellNPC("npc2", "npc3", "event1")
	if err != nil {
		t.Fatalf("TellNPC (npc2 -> npc3) failed: %v", err)
	}

	// NPC3 tells player (paranoid, high fear -> player with high trust)
	mgr.SetEntityRoom("player", "room1")
	err = mgr.TellNPC("npc3", "player", "event1")
	if err != nil {
		t.Fatalf("TellNPC (npc3 -> player) failed: %v", err)
	}

	// Verify propagation chain
	kb3 := mgr.GetKnowledgeBase("npc3")
	kbPlayer := mgr.GetKnowledgeBase("player")

	if !kb3.HasFact("event1") {
		t.Error("npc3 should know event1")
	}
	if !kbPlayer.HasFact("event1") {
		t.Error("player should know event1")
	}

	// Check propagation depths
	fact3 := kb3.GetFact("event1")
	factPlayer := kbPlayer.GetFact("event1")

	if fact3.PropagationDepth != 1 {
		t.Errorf("npc3 should have depth 1, got %d", fact3.PropagationDepth)
	}
	if factPlayer.PropagationDepth != 2 {
		t.Errorf("player should have depth 2, got %d", factPlayer.PropagationDepth)
	}

	// Check confidence decay (0.85 is the expected value for depth 1)
	if fact3.Confidence > 0.86 {
		t.Errorf("expected confidence decay for npc3, got %f", fact3.Confidence)
	}
	if factPlayer.Confidence >= fact3.Confidence {
		t.Errorf("expected additional confidence decay for player, got %f", factPlayer.Confidence)
	}
}

// TestDistortionIntegration_DistortionContent demonstrates distortion content generation
func TestDistortionIntegration_DistortionContent(t *testing.T) {
	npcMgr := newMockNPCManager()

	// High fear NPC for fear-based distortion
	npcMgr.emotions["npc_fear"] = struct{ trust, fear, stress int }{50, 80, 30}
	npcMgr.traits["npc_fear"] = []string{}

	// High stress NPC for stress-based distortion
	npcMgr.emotions["npc_stress"] = struct{ trust, fear, stress int }{50, 30, 80}
	npcMgr.traits["npc_stress"] = []string{}

	// Low trust NPC for skeptical distortion
	npcMgr.emotions["npc_distrust"] = struct{ trust, fear, stress int }{20, 30, 30}
	npcMgr.traits["npc_distrust"] = []string{}

	distortionCalc := NewDistortionCalculator(&npcManagerAdapter{npcMgr: npcMgr}, nil)

	originalContent := "發現了一個秘密房間"

	// Test fear-based distortion
	resultFear, _ := distortionCalc.ApplyDistortion("npc_fear", originalContent, 1)
	if resultFear.ShouldDistort && resultFear.DistortedContent != "" {
		// Verify distortion occurred and content is different
		if resultFear.DistortedContent == originalContent {
			t.Error("distorted content should be different from original")
		}
		// With high fear and depth, distortion should be likely
		t.Logf("Fear distortion: %s", resultFear.DistortedContent)
	}

	// Test stress-based distortion
	resultStress, _ := distortionCalc.ApplyDistortion("npc_stress", originalContent, 1)
	if resultStress.ShouldDistort && resultStress.DistortedContent != "" {
		t.Logf("Stress distortion: %s", resultStress.DistortedContent)
	}

	// Test low-trust distortion
	resultDistrust, _ := distortionCalc.ApplyDistortion("npc_distrust", originalContent, 1)
	if resultDistrust.ShouldDistort && resultDistrust.DistortedContent != "" {
		t.Logf("Low-trust distortion: %s", resultDistrust.DistortedContent)
	}
}
