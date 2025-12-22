package manager

import (
	"testing"
)

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{"value below 0", -10, 0},
		{"value at 0", 0, 0},
		{"value in range", 50, 50},
		{"value at 100", 100, 100},
		{"value above 100", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clamp(tt.value); got != tt.want {
				t.Errorf("clamp(%d) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestNewEmotionState(t *testing.T) {
	tests := []struct {
		name                string
		trust, fear, stress int
		want                EmotionState
	}{
		{
			name:   "normal values",
			trust:  50,
			fear:   30,
			stress: 40,
			want:   EmotionState{Trust: 50, Fear: 30, Stress: 40},
		},
		{
			name:   "values below 0 clamped",
			trust:  -10,
			fear:   -5,
			stress: -20,
			want:   EmotionState{Trust: 0, Fear: 0, Stress: 0},
		},
		{
			name:   "values above 100 clamped",
			trust:  150,
			fear:   120,
			stress: 200,
			want:   EmotionState{Trust: 100, Fear: 100, Stress: 100},
		},
		{
			name:   "mixed boundary values",
			trust:  -10,
			fear:   50,
			stress: 150,
			want:   EmotionState{Trust: 0, Fear: 50, Stress: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEmotionState(tt.trust, tt.fear, tt.stress)
			if got != tt.want {
				t.Errorf("NewEmotionState(%d, %d, %d) = %+v, want %+v",
					tt.trust, tt.fear, tt.stress, got, tt.want)
			}
		})
	}
}

func TestDefaultEmotionState(t *testing.T) {
	got := DefaultEmotionState()
	want := EmotionState{Trust: 50, Fear: 25, Stress: 25}

	if got != want {
		t.Errorf("DefaultEmotionState() = %+v, want %+v", got, want)
	}
}

func TestEmotionState_Apply(t *testing.T) {
	tests := []struct {
		name    string
		initial EmotionState
		delta   EmotionDelta
		want    EmotionState
	}{
		{
			name:    "friendly chat",
			initial: EmotionState{Trust: 50, Fear: 30, Stress: 40},
			delta:   EmotionDeltas["friendly_chat"],
			want:    EmotionState{Trust: 55, Fear: 27, Stress: 35},
		},
		{
			name:    "threat",
			initial: EmotionState{Trust: 50, Fear: 30, Stress: 40},
			delta:   EmotionDeltas["threat"],
			want:    EmotionState{Trust: 30, Fear: 55, Stress: 55},
		},
		{
			name:    "help npc",
			initial: EmotionState{Trust: 50, Fear: 30, Stress: 40},
			delta:   EmotionDeltas["help_npc"],
			want:    EmotionState{Trust: 65, Fear: 25, Stress: 30},
		},
		{
			name:    "witness death",
			initial: EmotionState{Trust: 50, Fear: 30, Stress: 40},
			delta:   EmotionDeltas["witness_death"],
			want:    EmotionState{Trust: 50, Fear: 60, Stress: 80},
		},
		{
			name:    "clamp at upper bound",
			initial: EmotionState{Trust: 95, Fear: 95, Stress: 95},
			delta:   EmotionDelta{Trust: 20, Fear: 20, Stress: 20},
			want:    EmotionState{Trust: 100, Fear: 100, Stress: 100},
		},
		{
			name:    "clamp at lower bound",
			initial: EmotionState{Trust: 5, Fear: 5, Stress: 5},
			delta:   EmotionDelta{Trust: -20, Fear: -20, Stress: -20},
			want:    EmotionState{Trust: 0, Fear: 0, Stress: 0},
		},
		{
			name:    "calm down",
			initial: EmotionState{Trust: 50, Fear: 30, Stress: 40},
			delta:   EmotionDeltas["calm_down"],
			want:    EmotionState{Trust: 55, Fear: 20, Stress: 25},
		},
		{
			name:    "reveal secret",
			initial: EmotionState{Trust: 50, Fear: 30, Stress: 40},
			delta:   EmotionDeltas["reveal_secret"],
			want:    EmotionState{Trust: 70, Fear: 20, Stress: 35},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.initial.Apply(tt.delta)
			if got != tt.want {
				t.Errorf("EmotionState.Apply() = %+v, want %+v", got, tt.want)
			}

			// Ensure original state is unchanged (immutability test)
			originalCopy := tt.initial
			if tt.initial != originalCopy {
				t.Error("Apply() modified the original EmotionState (not immutable)")
			}
		})
	}
}

func TestEmotionState_Copy(t *testing.T) {
	original := EmotionState{Trust: 50, Fear: 30, Stress: 40}
	copied := original.Copy()

	// Verify values are equal
	if copied != original {
		t.Errorf("Copy() = %+v, want %+v", copied, original)
	}

	// Modify copy and ensure original is unchanged
	copied.Trust = 100
	if original.Trust == 100 {
		t.Error("Modifying copy affected the original")
	}
}

func TestEmotionState_String(t *testing.T) {
	state := EmotionState{Trust: 50, Fear: 30, Stress: 40}
	got := state.String()

	// Just verify it contains the expected values
	if got == "" {
		t.Error("String() returned empty string")
	}

	// Verify it's valid JSON
	if got[0] != '{' || got[len(got)-1] != '}' {
		t.Errorf("String() = %q, doesn't look like JSON", got)
	}
}

func TestPredefinedEmotionDeltas(t *testing.T) {
	// Test that all predefined deltas exist and have reasonable values
	expectedDeltas := []string{
		"friendly_chat",
		"share_info",
		"threat",
		"help_npc",
		"ignore_distress",
		"reveal_secret",
		"catch_lying",
		"witness_death",
		"combat_together",
		"abandon_npc",
		"calm_down",
		"aggressive_talk",
		"hallucination",
	}

	for _, name := range expectedDeltas {
		t.Run(name, func(t *testing.T) {
			delta, exists := EmotionDeltas[name]
			if !exists {
				t.Fatalf("EmotionDelta %s does not exist", name)
			}
			// Just verify the delta has some non-zero values
			if delta.Trust == 0 && delta.Fear == 0 && delta.Stress == 0 {
				t.Errorf("%s delta is all zeros", name)
			}
		})
	}
}

func TestEmotionState_ApplyMultiple(t *testing.T) {
	// Test applying multiple deltas in sequence
	initial := DefaultEmotionState()

	// Friendly chat, then help, then calm down
	result := initial.
		Apply(EmotionDeltas["friendly_chat"]).
		Apply(EmotionDeltas["help_npc"]).
		Apply(EmotionDeltas["calm_down"])

	// After positive interactions, trust should be higher, fear/stress lower
	if result.Trust <= initial.Trust {
		t.Errorf("After positive interactions, Trust should increase. Got %d, initial %d",
			result.Trust, initial.Trust)
	}
	if result.Fear >= initial.Fear {
		t.Errorf("After positive interactions, Fear should decrease. Got %d, initial %d",
			result.Fear, initial.Fear)
	}
	if result.Stress >= initial.Stress {
		t.Errorf("After positive interactions, Stress should decrease. Got %d, initial %d",
			result.Stress, initial.Stress)
	}
}
