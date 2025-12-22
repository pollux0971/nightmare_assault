package knowledge

import (
	"encoding/json"
	"testing"
	"time"
)

// TestLearnMethod_String tests that all LearnMethod values have correct string representations.
// Verifies AC5: LearnMethod 枚舉 (witness/told/overheard/inferred)
func TestLearnMethod_String(t *testing.T) {
	tests := []struct {
		method   LearnMethod
		expected string
	}{
		{Witness, "witness"},
		{Told, "told"},
		{Overheard, "overheard"},
		{Inferred, "inferred"},
		{LearnMethod(999), "unknown"}, // Test unknown value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.method.String(); got != tt.expected {
				t.Errorf("LearnMethod.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLearnMethod_MarshalJSON tests JSON marshaling of LearnMethod.
// Verifies AC5: LearnMethod 枚舉正確序列化
func TestLearnMethod_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		method   LearnMethod
		expected string
	}{
		{"Witness", Witness, `"witness"`},
		{"Told", Told, `"told"`},
		{"Overheard", Overheard, `"overheard"`},
		{"Inferred", Inferred, `"inferred"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.method)
			if err != nil {
				t.Fatalf("Failed to marshal LearnMethod: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON() = %v, want %v", string(data), tt.expected)
			}
		})
	}
}

// TestLearnMethod_UnmarshalJSON tests JSON unmarshaling of LearnMethod.
// Verifies AC5: LearnMethod 枚舉正確反序列化
func TestLearnMethod_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected LearnMethod
	}{
		{"Witness", `"witness"`, Witness},
		{"Told", `"told"`, Told},
		{"Overheard", `"overheard"`, Overheard},
		{"Inferred", `"inferred"`, Inferred},
		{"Unknown", `"unknown"`, Witness}, // Unknown defaults to Witness
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lm LearnMethod
			err := json.Unmarshal([]byte(tt.json), &lm)
			if err != nil {
				t.Fatalf("Failed to unmarshal LearnMethod: %v", err)
			}
			if lm != tt.expected {
				t.Errorf("UnmarshalJSON() = %v, want %v", lm, tt.expected)
			}
		})
	}
}

// TestKnownFact_AllFields tests that KnownFact structure has all required fields.
// Verifies AC4: KnownFact 包含 LearnedAt/LearnedFrom/LearnMethod/Confidence/IsDistorted
func TestKnownFact_AllFields(t *testing.T) {
	now := time.Now()
	kf := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   now,
		LearnedFrom: "npc-1",
		LearnMethod: Told,
		Confidence:  0.8,
		IsDistorted: true,
	}

	// Verify field types and values
	var _ string = kf.FactID
	var _ time.Time = kf.LearnedAt
	var _ string = kf.LearnedFrom
	var _ LearnMethod = kf.LearnMethod
	var _ float64 = kf.Confidence
	var _ bool = kf.IsDistorted

	if kf.FactID != "fact-1" {
		t.Error("FactID field not accessible")
	}
	if kf.LearnedAt != now {
		t.Error("LearnedAt field not accessible")
	}
	if kf.LearnedFrom != "npc-1" {
		t.Error("LearnedFrom field not accessible")
	}
	if kf.LearnMethod != Told {
		t.Error("LearnMethod field not accessible")
	}
	if kf.Confidence != 0.8 {
		t.Error("Confidence field not accessible")
	}
	if !kf.IsDistorted {
		t.Error("IsDistorted field not accessible")
	}
}

// TestKnownFact_JSONSerialization tests that KnownFact can be serialized correctly.
// Verifies AC4: KnownFact 支援 JSON 序列化
func TestKnownFact_JSONSerialization(t *testing.T) {
	original := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "npc-2",
		LearnMethod: Overheard,
		Confidence:  0.6,
		IsDistorted: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal KnownFact: %v", err)
	}

	// Unmarshal back
	var restored KnownFact
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal KnownFact: %v", err)
	}

	// Verify all fields
	if restored.FactID != original.FactID {
		t.Errorf("FactID = %v, want %v", restored.FactID, original.FactID)
	}
	if restored.LearnedFrom != original.LearnedFrom {
		t.Errorf("LearnedFrom = %v, want %v", restored.LearnedFrom, original.LearnedFrom)
	}
	if restored.LearnMethod != original.LearnMethod {
		t.Errorf("LearnMethod = %v, want %v", restored.LearnMethod, original.LearnMethod)
	}
	if restored.Confidence != original.Confidence {
		t.Errorf("Confidence = %v, want %v", restored.Confidence, original.Confidence)
	}
	if restored.IsDistorted != original.IsDistorted {
		t.Errorf("IsDistorted = %v, want %v", restored.IsDistorted, original.IsDistorted)
	}
}

// TestBelief_AllFields tests that Belief structure has all required fields.
// Verifies AC6: Belief 包含 Content/BasedOn/Confidence
func TestBelief_AllFields(t *testing.T) {
	belief := &Belief{
		Content:    "The basement is dangerous",
		BasedOn:    []string{"fact-1", "fact-2"},
		Confidence: 0.9,
	}

	// Verify field types and values
	var _ string = belief.Content
	var _ []string = belief.BasedOn
	var _ float64 = belief.Confidence

	if belief.Content != "The basement is dangerous" {
		t.Error("Content field not accessible")
	}
	if len(belief.BasedOn) != 2 {
		t.Error("BasedOn field not accessible")
	}
	if belief.Confidence != 0.9 {
		t.Error("Confidence field not accessible")
	}
}

// TestBelief_JSONSerialization tests that Belief can be serialized correctly.
// Verifies AC6: Belief 支援 JSON 序列化
func TestBelief_JSONSerialization(t *testing.T) {
	original := &Belief{
		Content:    "NPCs are trustworthy",
		BasedOn:    []string{"fact-1", "fact-2", "fact-3"},
		Confidence: 0.75,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Belief: %v", err)
	}

	// Unmarshal back
	var restored Belief
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal Belief: %v", err)
	}

	// Verify all fields
	if restored.Content != original.Content {
		t.Errorf("Content = %v, want %v", restored.Content, original.Content)
	}
	if len(restored.BasedOn) != len(original.BasedOn) {
		t.Errorf("BasedOn length = %v, want %v", len(restored.BasedOn), len(original.BasedOn))
	}
	if restored.Confidence != original.Confidence {
		t.Errorf("Confidence = %v, want %v", restored.Confidence, original.Confidence)
	}
}

// TestKnowledgeBase_AllFields tests that KnowledgeBase has all required fields.
// Verifies AC3: KnowledgeBase 包含 OwnerID/KnownFacts/Beliefs/LastUpdated
func TestKnowledgeBase_AllFields(t *testing.T) {
	kb := &KnowledgeBase{
		OwnerID:     "npc-1",
		KnownFacts:  make(map[string]*KnownFact),
		Beliefs:     []*Belief{},
		LastUpdated: time.Now(),
	}

	// Verify field types and values
	var _ string = kb.OwnerID
	var _ map[string]*KnownFact = kb.KnownFacts
	var _ []*Belief = kb.Beliefs
	var _ time.Time = kb.LastUpdated

	if kb.OwnerID != "npc-1" {
		t.Error("OwnerID field not accessible")
	}
	if kb.KnownFacts == nil {
		t.Error("KnownFacts field not accessible")
	}
	if kb.Beliefs == nil {
		t.Error("Beliefs field not accessible")
	}
	if kb.LastUpdated.IsZero() {
		t.Error("LastUpdated field not accessible")
	}
}

// TestNewKnowledgeBase tests that NewKnowledgeBase initializes all fields correctly.
// Verifies AC3: KnowledgeBase 正確初始化
func TestNewKnowledgeBase(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	if kb == nil {
		t.Fatal("NewKnowledgeBase returned nil")
	}

	if kb.OwnerID != "npc-1" {
		t.Errorf("OwnerID = %v, want npc-1", kb.OwnerID)
	}

	if kb.KnownFacts == nil {
		t.Error("KnownFacts not initialized")
	}

	if len(kb.KnownFacts) != 0 {
		t.Error("KnownFacts should be empty initially")
	}

	if kb.Beliefs == nil {
		t.Error("Beliefs not initialized")
	}

	if len(kb.Beliefs) != 0 {
		t.Error("Beliefs should be empty initially")
	}

	if kb.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set")
	}

	if time.Since(kb.LastUpdated) > time.Second {
		t.Error("LastUpdated should be recent")
	}
}

// TestKnowledgeBase_AddFact tests adding facts to knowledge base.
// Verifies AC3: KnowledgeBase 支援添加事實
func TestKnowledgeBase_AddFact(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")
	initialTime := kb.LastUpdated

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Add a fact
	kf := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "player",
		LearnMethod: Told,
		Confidence:  0.8,
		IsDistorted: false,
	}
	kb.AddFact(kf)

	// Verify fact was added
	if len(kb.KnownFacts) != 1 {
		t.Errorf("KnownFacts length = %v, want 1", len(kb.KnownFacts))
	}

	// Verify fact is retrievable
	retrieved := kb.GetFact("fact-1")
	if retrieved == nil {
		t.Fatal("GetFact returned nil")
	}

	if retrieved.FactID != "fact-1" {
		t.Errorf("Retrieved fact ID = %v, want fact-1", retrieved.FactID)
	}

	// Verify LastUpdated was updated
	if !kb.LastUpdated.After(initialTime) {
		t.Error("LastUpdated was not updated after adding fact")
	}
}

// TestKnowledgeBase_AddFact_UpdateHigherConfidence tests that adding a fact
// with higher confidence updates the existing fact.
func TestKnowledgeBase_AddFact_UpdateHigherConfidence(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	// Add initial fact with low confidence
	kf1 := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "npc-2",
		LearnMethod: Told,
		Confidence:  0.5,
		IsDistorted: false,
	}
	kb.AddFact(kf1)

	// Add same fact with higher confidence
	kf2 := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "player",
		LearnMethod: Witness,
		Confidence:  0.9,
		IsDistorted: false,
	}
	kb.AddFact(kf2)

	// Should have only one fact
	if len(kb.KnownFacts) != 1 {
		t.Errorf("KnownFacts length = %v, want 1", len(kb.KnownFacts))
	}

	// Should have higher confidence version
	retrieved := kb.GetFact("fact-1")
	if retrieved.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", retrieved.Confidence)
	}
	if retrieved.LearnMethod != Witness {
		t.Errorf("LearnMethod = %v, want Witness", retrieved.LearnMethod)
	}
}

// TestKnowledgeBase_AddFact_KeepHigherConfidence tests that adding a fact
// with lower confidence does not update the existing fact.
func TestKnowledgeBase_AddFact_KeepHigherConfidence(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	// Add initial fact with high confidence
	kf1 := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "player",
		LearnMethod: Witness,
		Confidence:  0.9,
		IsDistorted: false,
	}
	kb.AddFact(kf1)

	// Try to add same fact with lower confidence
	kf2 := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "npc-2",
		LearnMethod: Told,
		Confidence:  0.5,
		IsDistorted: true,
	}
	kb.AddFact(kf2)

	// Should keep higher confidence version
	retrieved := kb.GetFact("fact-1")
	if retrieved.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9 (should keep higher)", retrieved.Confidence)
	}
	if retrieved.LearnMethod != Witness {
		t.Errorf("LearnMethod = %v, want Witness (should keep original)", retrieved.LearnMethod)
	}
	if retrieved.IsDistorted {
		t.Error("IsDistorted should be false (should keep original)")
	}
}

// TestKnowledgeBase_HasFact tests checking if knowledge base has a fact.
func TestKnowledgeBase_HasFact(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	if kb.HasFact("fact-1") {
		t.Error("HasFact(fact-1) = true, want false (empty KB)")
	}

	// Add a fact
	kf := &KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "player",
		LearnMethod: Witness,
		Confidence:  0.8,
		IsDistorted: false,
	}
	kb.AddFact(kf)

	if !kb.HasFact("fact-1") {
		t.Error("HasFact(fact-1) = false, want true")
	}

	if kb.HasFact("fact-2") {
		t.Error("HasFact(fact-2) = true, want false")
	}
}

// TestKnowledgeBase_AddBelief tests adding beliefs to knowledge base.
// Verifies AC3 & AC6: KnowledgeBase 支援添加信念
func TestKnowledgeBase_AddBelief(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")
	initialTime := kb.LastUpdated

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Add a belief
	belief := &Belief{
		Content:    "The player is trustworthy",
		BasedOn:    []string{"fact-1", "fact-2"},
		Confidence: 0.8,
	}
	kb.AddBelief(belief)

	// Verify belief was added
	if len(kb.Beliefs) != 1 {
		t.Errorf("Beliefs length = %v, want 1", len(kb.Beliefs))
	}

	if kb.Beliefs[0].Content != "The player is trustworthy" {
		t.Errorf("Belief content = %v, want 'The player is trustworthy'", kb.Beliefs[0].Content)
	}

	// Verify LastUpdated was updated
	if !kb.LastUpdated.After(initialTime) {
		t.Error("LastUpdated was not updated after adding belief")
	}
}

// TestKnowledgeBase_GetAllFactIDs tests getting all fact IDs.
func TestKnowledgeBase_GetAllFactIDs(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	// Empty KB
	ids := kb.GetAllFactIDs()
	if len(ids) != 0 {
		t.Errorf("GetAllFactIDs() length = %v, want 0", len(ids))
	}

	// Add facts
	kb.AddFact(&KnownFact{FactID: "fact-1", LearnedAt: time.Now(), LearnedFrom: "player", LearnMethod: Witness, Confidence: 0.8})
	kb.AddFact(&KnownFact{FactID: "fact-2", LearnedAt: time.Now(), LearnedFrom: "npc-1", LearnMethod: Told, Confidence: 0.6})
	kb.AddFact(&KnownFact{FactID: "fact-3", LearnedAt: time.Now(), LearnedFrom: "system", LearnMethod: Inferred, Confidence: 0.7})

	ids = kb.GetAllFactIDs()
	if len(ids) != 3 {
		t.Errorf("GetAllFactIDs() length = %v, want 3", len(ids))
	}

	// Verify all IDs are present (order doesn't matter)
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, expectedID := range []string{"fact-1", "fact-2", "fact-3"} {
		if !idMap[expectedID] {
			t.Errorf("Expected fact ID %v not found in GetAllFactIDs()", expectedID)
		}
	}
}

// TestKnowledgeBase_Copy tests deep copying of knowledge base.
func TestKnowledgeBase_Copy(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	// Add facts and beliefs
	kb.AddFact(&KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "player",
		LearnMethod: Witness,
		Confidence:  0.8,
		IsDistorted: false,
	})

	kb.AddBelief(&Belief{
		Content:    "Test belief",
		BasedOn:    []string{"fact-1"},
		Confidence: 0.7,
	})

	// Copy
	copied := kb.Copy()

	if copied == nil {
		t.Fatal("Copy returned nil")
	}

	// Verify basic fields
	if copied.OwnerID != kb.OwnerID {
		t.Error("OwnerID not copied correctly")
	}

	if len(copied.KnownFacts) != len(kb.KnownFacts) {
		t.Error("KnownFacts not copied correctly")
	}

	if len(copied.Beliefs) != len(kb.Beliefs) {
		t.Error("Beliefs not copied correctly")
	}

	// Verify deep copy - modifying copied should not affect original
	copied.AddFact(&KnownFact{
		FactID:      "fact-2",
		LearnedAt:   time.Now(),
		LearnedFrom: "system",
		LearnMethod: Inferred,
		Confidence:  0.5,
	})

	if len(kb.KnownFacts) != 1 {
		t.Error("Modifying copied KB affected original KnownFacts")
	}

	// Modify belief in copy
	copied.Beliefs[0].Content = "Modified belief"
	if kb.Beliefs[0].Content == "Modified belief" {
		t.Error("Modifying copied belief affected original")
	}

	// Modify BasedOn in copy
	copied.Beliefs[0].BasedOn[0] = "modified-fact"
	if kb.Beliefs[0].BasedOn[0] == "modified-fact" {
		t.Error("Modifying copied belief BasedOn affected original")
	}
}

// TestKnowledgeBase_JSONSerialization tests that KnowledgeBase can be serialized.
// Verifies AC3: KnowledgeBase 支援 JSON 序列化
func TestKnowledgeBase_JSONSerialization(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	kb.AddFact(&KnownFact{
		FactID:      "fact-1",
		LearnedAt:   time.Now(),
		LearnedFrom: "player",
		LearnMethod: Witness,
		Confidence:  0.9,
		IsDistorted: false,
	})

	kb.AddBelief(&Belief{
		Content:    "Test belief",
		BasedOn:    []string{"fact-1"},
		Confidence: 0.8,
	})

	// Marshal to JSON
	data, err := json.Marshal(kb)
	if err != nil {
		t.Fatalf("Failed to marshal KnowledgeBase: %v", err)
	}

	// Unmarshal back
	var restored KnowledgeBase
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal KnowledgeBase: %v", err)
	}

	// Verify fields
	if restored.OwnerID != kb.OwnerID {
		t.Errorf("OwnerID = %v, want %v", restored.OwnerID, kb.OwnerID)
	}

	if len(restored.KnownFacts) != len(kb.KnownFacts) {
		t.Errorf("KnownFacts length = %v, want %v", len(restored.KnownFacts), len(kb.KnownFacts))
	}

	if len(restored.Beliefs) != len(kb.Beliefs) {
		t.Errorf("Beliefs length = %v, want %v", len(restored.Beliefs), len(kb.Beliefs))
	}
}

// TestKnowledgeBase_String tests the String method.
func TestKnowledgeBase_String(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")
	str := kb.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	// Should be valid JSON
	var unmarshaled KnowledgeBase
	err := json.Unmarshal([]byte(str), &unmarshaled)
	if err != nil {
		t.Errorf("String() did not return valid JSON: %v", err)
	}
}

// TestKnowledgeBase_DifferentLearnMethods tests facts learned through different methods.
// Verifies AC5: LearnMethod 枚舉支援所有學習方式
func TestKnowledgeBase_DifferentLearnMethods(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	methods := []LearnMethod{Witness, Told, Overheard, Inferred}

	for i, method := range methods {
		kb.AddFact(&KnownFact{
			FactID:      string(rune('a' + i)),
			LearnedAt:   time.Now(),
			LearnedFrom: "source",
			LearnMethod: method,
			Confidence:  0.8,
			IsDistorted: false,
		})
	}

	if len(kb.KnownFacts) != 4 {
		t.Errorf("KnownFacts length = %v, want 4", len(kb.KnownFacts))
	}

	// Verify each method was preserved
	for i, method := range methods {
		factID := string(rune('a' + i))
		fact := kb.GetFact(factID)
		if fact == nil {
			t.Errorf("Fact %v not found", factID)
			continue
		}
		if fact.LearnMethod != method {
			t.Errorf("Fact %v LearnMethod = %v, want %v", factID, fact.LearnMethod, method)
		}
	}
}

// TestKnowledgeBase_MultipleBeliefs tests adding multiple beliefs.
// Verifies AC6: KnowledgeBase 支援多個信念
func TestKnowledgeBase_MultipleBeliefs(t *testing.T) {
	kb := NewKnowledgeBase("npc-1")

	beliefs := []*Belief{
		{Content: "Belief 1", BasedOn: []string{"fact-1"}, Confidence: 0.9},
		{Content: "Belief 2", BasedOn: []string{"fact-2", "fact-3"}, Confidence: 0.7},
		{Content: "Belief 3", BasedOn: []string{"fact-1", "fact-2", "fact-3"}, Confidence: 0.5},
	}

	for _, belief := range beliefs {
		kb.AddBelief(belief)
	}

	if len(kb.Beliefs) != 3 {
		t.Errorf("Beliefs length = %v, want 3", len(kb.Beliefs))
	}

	// Verify beliefs were added in order
	for i, belief := range beliefs {
		if kb.Beliefs[i].Content != belief.Content {
			t.Errorf("Belief[%d].Content = %v, want %v", i, kb.Beliefs[i].Content, belief.Content)
		}
	}
}

// TestKnowledgeBase_EmptyOwner tests that knowledge base can have empty owner.
func TestKnowledgeBase_EmptyOwner(t *testing.T) {
	kb := NewKnowledgeBase("")

	if kb == nil {
		t.Fatal("NewKnowledgeBase returned nil for empty owner")
	}

	if kb.OwnerID != "" {
		t.Error("OwnerID should be empty string")
	}
}

// TestKnownFact_DifferentConfidenceLevels tests various confidence levels.
// Verifies AC4: KnownFact.Confidence 支援 0.0-1.0 範圍
func TestKnownFact_DifferentConfidenceLevels(t *testing.T) {
	confidenceLevels := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, conf := range confidenceLevels {
		kf := &KnownFact{
			FactID:      "fact-1",
			LearnedAt:   time.Now(),
			LearnedFrom: "source",
			LearnMethod: Witness,
			Confidence:  conf,
			IsDistorted: false,
		}

		if kf.Confidence != conf {
			t.Errorf("Confidence = %v, want %v", kf.Confidence, conf)
		}
	}
}

// TestBelief_EmptyBasedOn tests belief with no supporting facts.
// Verifies AC6: Belief.BasedOn 可以為空
func TestBelief_EmptyBasedOn(t *testing.T) {
	belief := &Belief{
		Content:    "Pure intuition",
		BasedOn:    []string{},
		Confidence: 0.3,
	}

	if len(belief.BasedOn) != 0 {
		t.Errorf("BasedOn length = %v, want 0", len(belief.BasedOn))
	}
}
