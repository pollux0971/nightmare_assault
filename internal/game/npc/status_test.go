package npc

import (
	"testing"
	"time"
)

func TestLocation(t *testing.T) {
	location := Location{
		Scene:            "大廳",
		DistanceToPlayer: 0,
	}

	if location.Scene == "" {
		t.Error("Scene should not be empty")
	}

	if location.DistanceToPlayer < 0 {
		t.Error("DistanceToPlayer should not be negative")
	}
}

func TestEmotionalState(t *testing.T) {
	emotions := []EmotionalState{
		EmotionCalm,
		EmotionAnxious,
		EmotionPanicked,
		EmotionRelieved,
		EmotionGrieving,
	}

	if len(emotions) != 5 {
		t.Errorf("Expected 5 emotional states, got %d", len(emotions))
	}
}

func TestInjuryLevel(t *testing.T) {
	levels := []InjuryLevel{
		InjuryNone,
		InjuryMinor,
		InjurySerious,
		InjuryCritical,
	}

	if len(levels) != 4 {
		t.Errorf("Expected 4 injury levels, got %d", len(levels))
	}

	if InjuryNone != 0 {
		t.Error("InjuryNone should be 0")
	}
}

func TestCalculateInjuryLevel(t *testing.T) {
	tests := []struct {
		hp   int
		want InjuryLevel
	}{
		{100, InjuryNone},
		{70, InjuryNone},
		{69, InjuryMinor},
		{50, InjuryMinor},
		{30, InjuryMinor},
		{29, InjurySerious},
		{15, InjurySerious},
		{14, InjuryCritical},
		{1, InjuryCritical},
	}

	for _, tt := range tests {
		got := CalculateInjuryLevel(tt.hp)
		if got != tt.want {
			t.Errorf("CalculateInjuryLevel(%d) = %v, want %v", tt.hp, got, tt.want)
		}
	}
}

func TestTeammateMessage(t *testing.T) {
	msg := TeammateMessage{
		Content:   "我在二樓發現了一本日記",
		Timestamp: time.Now(),
	}

	if msg.Content == "" {
		t.Error("Message content should not be empty")
	}

	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestBehaviorModifier(t *testing.T) {
	tests := []struct {
		hp            int
		minMoveSpeed  float64
		maxMoveSpeed  float64
		minReaction   float64
		maxReaction   float64
	}{
		{100, 1.0, 1.0, 1.0, 1.0},      // Healthy
		{50, 0.7, 0.9, 0.8, 1.0},       // Minor injury
		{20, 0.4, 0.6, 0.5, 0.7},       // Serious injury
		{10, 0.0, 0.2, 0.2, 0.4},       // Critical injury
	}

	for _, tt := range tests {
		modifier := GetBehaviorModifier(tt.hp)

		if modifier.MoveSpeed < tt.minMoveSpeed || modifier.MoveSpeed > tt.maxMoveSpeed {
			t.Errorf("HP %d: MoveSpeed = %f, want in range [%f, %f]",
				tt.hp, modifier.MoveSpeed, tt.minMoveSpeed, tt.maxMoveSpeed)
		}

		if modifier.Reaction < tt.minReaction || modifier.Reaction > tt.maxReaction {
			t.Errorf("HP %d: Reaction = %f, want in range [%f, %f]",
				tt.hp, modifier.Reaction, tt.minReaction, tt.maxReaction)
		}
	}
}

func TestUpdateTeammateLocation(t *testing.T) {
	teammate := NewTeammate("tm-001", "李明", ArchetypeLogic)

	newLocation := Location{
		Scene:            "廚房",
		DistanceToPlayer: 1,
	}

	teammate.UpdateLocation(newLocation)

	if teammate.Location != "廚房" {
		t.Errorf("Location = %s, want 廚房", teammate.Location)
	}

	if teammate.IsSeparated != true {
		t.Error("IsSeparated should be true when DistanceToPlayer > 0")
	}
}

func TestUpdateTeammateHP(t *testing.T) {
	teammate := NewTeammate("tm-001", "李明", ArchetypeLogic)

	teammate.UpdateHP(50)

	if teammate.HP != 50 {
		t.Errorf("HP = %d, want 50", teammate.HP)
	}

	// Injury level should update
	expectedInjury := CalculateInjuryLevel(50)
	if teammate.InjuryLevel != expectedInjury {
		t.Errorf("InjuryLevel = %v, want %v", teammate.InjuryLevel, expectedInjury)
	}

	// Status.Condition should update
	if teammate.HP < 30 && teammate.Status.Condition != "injured" {
		t.Errorf("Condition should be 'injured' when HP < 30")
	}

	// Test HP clamp to [0, 100]
	teammate.UpdateHP(150)
	if teammate.HP != 100 {
		t.Errorf("HP = %d, want 100 (clamped)", teammate.HP)
	}

	teammate.UpdateHP(-10)
	if teammate.HP != 0 {
		t.Errorf("HP = %d, want 0 (clamped)", teammate.HP)
	}

	if teammate.HP == 0 && teammate.Status.Alive {
		t.Error("Status.Alive should be false when HP = 0")
	}
}

func TestUpdateEmotionalState(t *testing.T) {
	teammate := NewTeammate("tm-001", "李明", ArchetypeLogic)

	teammate.UpdateEmotionalState(EmotionAnxious)

	if teammate.EmotionalState != EmotionAnxious {
		t.Errorf("EmotionalState = %v, want %v", teammate.EmotionalState, EmotionAnxious)
	}
}

func TestIsSeparated(t *testing.T) {
	teammate := NewTeammate("tm-001", "李明", ArchetypeLogic)

	// Initially not separated
	if teammate.IsSeparated {
		t.Error("Teammate should not be separated initially")
	}

	// Separated when distance > 0
	teammate.UpdateLocation(Location{
		Scene:            "二樓",
		DistanceToPlayer: 2,
	})

	if !teammate.IsSeparated {
		t.Error("Teammate should be separated when DistanceToPlayer > 0")
	}
}
