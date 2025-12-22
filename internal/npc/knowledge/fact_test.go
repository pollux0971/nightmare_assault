package knowledge

import (
	"encoding/json"
	"testing"
	"time"
)

// TestFactType_String tests that all FactType values have correct string representations.
// Verifies AC2: FactType 枚舉 (event/dialogue/discovery/rumor/secret)
func TestFactType_String(t *testing.T) {
	tests := []struct {
		factType FactType
		expected string
	}{
		{Event, "event"},
		{Dialogue, "dialogue"},
		{Discovery, "discovery"},
		{Rumor, "rumor"},
		{Secret, "secret"},
		{FactType(999), "unknown"}, // Test unknown value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.factType.String(); got != tt.expected {
				t.Errorf("FactType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFactType_MarshalJSON tests JSON marshaling of FactType.
// Verifies AC2: FactType 枚舉正確序列化
func TestFactType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		factType FactType
		expected string
	}{
		{"Event", Event, `"event"`},
		{"Dialogue", Dialogue, `"dialogue"`},
		{"Discovery", Discovery, `"discovery"`},
		{"Rumor", Rumor, `"rumor"`},
		{"Secret", Secret, `"secret"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.factType)
			if err != nil {
				t.Fatalf("Failed to marshal FactType: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON() = %v, want %v", string(data), tt.expected)
			}
		})
	}
}

// TestFactType_UnmarshalJSON tests JSON unmarshaling of FactType.
// Verifies AC2: FactType 枚舉正確反序列化
func TestFactType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected FactType
	}{
		{"Event", `"event"`, Event},
		{"Dialogue", `"dialogue"`, Dialogue},
		{"Discovery", `"discovery"`, Discovery},
		{"Rumor", `"rumor"`, Rumor},
		{"Secret", `"secret"`, Secret},
		{"Unknown", `"unknown"`, Event}, // Unknown defaults to Event
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FactType
			err := json.Unmarshal([]byte(tt.json), &ft)
			if err != nil {
				t.Fatalf("Failed to unmarshal FactType: %v", err)
			}
			if ft != tt.expected {
				t.Errorf("UnmarshalJSON() = %v, want %v", ft, tt.expected)
			}
		})
	}
}

// TestNewFact tests that NewFact creates a Fact with all required fields.
// Verifies AC1: Fact 包含 ID/Content/Type/Source/CreatedAt/Location/Witnesses
func TestNewFact(t *testing.T) {
	witnesses := []string{"npc-1", "npc-2", "player"}
	fact := NewFact("fact-1", "Test content", Event, "player", "room-1", witnesses)

	if fact == nil {
		t.Fatal("NewFact returned nil")
	}

	// Verify all fields are set correctly
	if fact.ID != "fact-1" {
		t.Errorf("ID = %v, want %v", fact.ID, "fact-1")
	}

	if fact.Content != "Test content" {
		t.Errorf("Content = %v, want %v", fact.Content, "Test content")
	}

	if fact.Type != Event {
		t.Errorf("Type = %v, want %v", fact.Type, Event)
	}

	if fact.Source != "player" {
		t.Errorf("Source = %v, want %v", fact.Source, "player")
	}

	if fact.Location != "room-1" {
		t.Errorf("Location = %v, want %v", fact.Location, "room-1")
	}

	// Verify CreatedAt is set and recent
	if fact.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero time")
	}
	if time.Since(fact.CreatedAt) > time.Second {
		t.Error("CreatedAt is not recent")
	}

	// Verify Witnesses
	if len(fact.Witnesses) != 3 {
		t.Errorf("Witnesses length = %v, want %v", len(fact.Witnesses), 3)
	}
	for i, w := range witnesses {
		if fact.Witnesses[i] != w {
			t.Errorf("Witnesses[%d] = %v, want %v", i, fact.Witnesses[i], w)
		}
	}
}

// TestNewFact_NilWitnesses tests that NewFact handles nil witnesses correctly.
// Verifies AC1: Fact 正確處理空見證者列表
func TestNewFact_NilWitnesses(t *testing.T) {
	fact := NewFact("fact-1", "Content", Event, "system", "room-1", nil)

	if fact == nil {
		t.Fatal("NewFact returned nil")
	}

	if fact.Witnesses == nil {
		t.Error("Witnesses should be initialized to empty slice, not nil")
	}

	if len(fact.Witnesses) != 0 {
		t.Errorf("Witnesses length = %v, want 0", len(fact.Witnesses))
	}
}

// TestFact_AllFields tests that Fact structure has all required fields with correct types.
// Verifies AC1: Fact 包含 ID/Content/Type/Source/CreatedAt/Location/Witnesses
func TestFact_AllFields(t *testing.T) {
	now := time.Now()
	fact := &Fact{
		ID:        "test-id",
		Content:   "test-content",
		Type:      Discovery,
		Source:    "npc-1",
		CreatedAt: now,
		Location:  "library",
		Witnesses: []string{"player", "npc-2"},
	}

	// Verify field types and values
	var _ string = fact.ID
	var _ string = fact.Content
	var _ FactType = fact.Type
	var _ string = fact.Source
	var _ time.Time = fact.CreatedAt
	var _ string = fact.Location
	var _ []string = fact.Witnesses

	if fact.ID != "test-id" {
		t.Error("ID field not accessible")
	}
	if fact.Content != "test-content" {
		t.Error("Content field not accessible")
	}
	if fact.Type != Discovery {
		t.Error("Type field not accessible")
	}
	if fact.Source != "npc-1" {
		t.Error("Source field not accessible")
	}
	if fact.CreatedAt != now {
		t.Error("CreatedAt field not accessible")
	}
	if fact.Location != "library" {
		t.Error("Location field not accessible")
	}
	if len(fact.Witnesses) != 2 {
		t.Error("Witnesses field not accessible")
	}
}

// TestFact_JSONSerialization tests that Fact can be serialized and deserialized correctly.
// Verifies AC1: Fact 支援 JSON 序列化
func TestFact_JSONSerialization(t *testing.T) {
	original := NewFact("fact-1", "A mysterious event occurred", Secret, "npc-1", "basement", []string{"npc-1", "npc-2"})

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Fact: %v", err)
	}

	// Unmarshal back
	var restored Fact
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal Fact: %v", err)
	}

	// Verify all fields
	if restored.ID != original.ID {
		t.Errorf("ID = %v, want %v", restored.ID, original.ID)
	}
	if restored.Content != original.Content {
		t.Errorf("Content = %v, want %v", restored.Content, original.Content)
	}
	if restored.Type != original.Type {
		t.Errorf("Type = %v, want %v", restored.Type, original.Type)
	}
	if restored.Source != original.Source {
		t.Errorf("Source = %v, want %v", restored.Source, original.Source)
	}
	if restored.Location != original.Location {
		t.Errorf("Location = %v, want %v", restored.Location, original.Location)
	}
	if len(restored.Witnesses) != len(original.Witnesses) {
		t.Errorf("Witnesses length = %v, want %v", len(restored.Witnesses), len(original.Witnesses))
	}
}

// TestFact_Copy tests that Copy creates a deep copy of the Fact.
func TestFact_Copy(t *testing.T) {
	original := NewFact("fact-1", "Original content", Event, "player", "room-1", []string{"npc-1"})
	copied := original.Copy()

	if copied == nil {
		t.Fatal("Copy returned nil")
	}

	// Verify all fields are equal
	if copied.ID != original.ID {
		t.Error("ID not copied correctly")
	}
	if copied.Content != original.Content {
		t.Error("Content not copied correctly")
	}
	if copied.Type != original.Type {
		t.Error("Type not copied correctly")
	}
	if copied.Source != original.Source {
		t.Error("Source not copied correctly")
	}
	if copied.Location != original.Location {
		t.Error("Location not copied correctly")
	}
	if !copied.CreatedAt.Equal(original.CreatedAt) {
		t.Error("CreatedAt not copied correctly")
	}

	// Verify deep copy of Witnesses slice
	if len(copied.Witnesses) != len(original.Witnesses) {
		t.Error("Witnesses not copied correctly")
	}

	// Modify copied witnesses - should not affect original
	copied.Witnesses[0] = "modified"
	if original.Witnesses[0] == "modified" {
		t.Error("Copy is not deep - modifying copied witnesses affected original")
	}
}

// TestFact_AddWitness tests adding witnesses to a fact.
func TestFact_AddWitness(t *testing.T) {
	fact := NewFact("fact-1", "Content", Event, "system", "room-1", []string{})

	// Add first witness
	fact.AddWitness("npc-1")
	if len(fact.Witnesses) != 1 {
		t.Errorf("Witnesses length = %v, want 1", len(fact.Witnesses))
	}
	if fact.Witnesses[0] != "npc-1" {
		t.Errorf("Witnesses[0] = %v, want npc-1", fact.Witnesses[0])
	}

	// Add second witness
	fact.AddWitness("player")
	if len(fact.Witnesses) != 2 {
		t.Errorf("Witnesses length = %v, want 2", len(fact.Witnesses))
	}

	// Add duplicate witness - should not add
	fact.AddWitness("npc-1")
	if len(fact.Witnesses) != 2 {
		t.Errorf("Witnesses length = %v, want 2 (duplicate should not be added)", len(fact.Witnesses))
	}
}

// TestFact_IsWitness tests checking if an entity is a witness.
func TestFact_IsWitness(t *testing.T) {
	fact := NewFact("fact-1", "Content", Event, "system", "room-1", []string{"npc-1", "player"})

	if !fact.IsWitness("npc-1") {
		t.Error("IsWitness(npc-1) = false, want true")
	}
	if !fact.IsWitness("player") {
		t.Error("IsWitness(player) = false, want true")
	}
	if fact.IsWitness("npc-2") {
		t.Error("IsWitness(npc-2) = true, want false")
	}
}

// TestFact_String tests the String method.
func TestFact_String(t *testing.T) {
	fact := NewFact("fact-1", "Content", Event, "system", "room-1", []string{"npc-1"})
	str := fact.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	// Should be valid JSON
	var unmarshaled Fact
	err := json.Unmarshal([]byte(str), &unmarshaled)
	if err != nil {
		t.Errorf("String() did not return valid JSON: %v", err)
	}
}

// TestFact_DifferentTypes tests creating facts with different types.
// Verifies AC2: FactType 枚舉支援所有類型
func TestFact_DifferentTypes(t *testing.T) {
	types := []FactType{Event, Dialogue, Discovery, Rumor, Secret}

	for _, factType := range types {
		fact := NewFact("fact-1", "Content", factType, "system", "room-1", nil)
		if fact.Type != factType {
			t.Errorf("Fact type = %v, want %v", fact.Type, factType)
		}
	}
}

// TestFact_EmptyContent tests that facts can have empty content.
func TestFact_EmptyContent(t *testing.T) {
	fact := NewFact("fact-1", "", Event, "system", "room-1", nil)

	if fact == nil {
		t.Fatal("NewFact returned nil for empty content")
	}

	if fact.Content != "" {
		t.Error("Content should be empty string")
	}
}

// TestFact_MultipleWitnesses tests that facts can have multiple witnesses.
// Verifies AC1: Fact.Witnesses 支援多個見證者
func TestFact_MultipleWitnesses(t *testing.T) {
	witnesses := []string{"player", "npc-1", "npc-2", "npc-3", "npc-4"}
	fact := NewFact("fact-1", "Content", Event, "system", "room-1", witnesses)

	if len(fact.Witnesses) != 5 {
		t.Errorf("Witnesses length = %v, want 5", len(fact.Witnesses))
	}

	for i, w := range witnesses {
		if fact.Witnesses[i] != w {
			t.Errorf("Witnesses[%d] = %v, want %v", i, fact.Witnesses[i], w)
		}
	}
}
