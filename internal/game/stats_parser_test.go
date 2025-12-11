package game

import (
	"testing"
)

func TestNewStatsParser(t *testing.T) {
	parser := NewStatsParser()

	if parser == nil {
		t.Fatal("NewStatsParser should not return nil")
	}
}

func TestStatsParser_ParseInlineText(t *testing.T) {
	parser := NewStatsParser()

	tests := []struct {
		name        string
		text        string
		expectHP    int
		expectSAN   int
		expectFound bool
	}{
		{
			name:        "Standard format",
			text:        "You take damage. HP: -15. Your sanity drops. SAN: -5.",
			expectHP:    -15,
			expectSAN:   -5,
			expectFound: true,
		},
		{
			name:        "Positive values",
			text:        "You feel better. HP: +10. Peace returns. SAN: +5.",
			expectHP:    10,
			expectSAN:   5,
			expectFound: true,
		},
		{
			name:        "HP only",
			text:        "Damage taken. HP: -20",
			expectHP:    -20,
			expectSAN:   0,
			expectFound: true,
		},
		{
			name:        "SAN only",
			text:        "Terror grips you. SAN: -30",
			expectHP:    0,
			expectSAN:   -30,
			expectFound: true,
		},
		{
			name:        "No markers",
			text:        "You explore the room carefully.",
			expectHP:    0,
			expectSAN:   0,
			expectFound: false,
		},
		{
			name:        "Large values",
			text:        "Critical hit! HP: -45. Overwhelming terror. SAN: -25.",
			expectHP:    -45,
			expectSAN:   -25,
			expectFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp, san, found := parser.ParseInlineText(tt.text)

			if found != tt.expectFound {
				t.Errorf("found = %v, want %v", found, tt.expectFound)
			}
			if hp != tt.expectHP {
				t.Errorf("HP = %d, want %d", hp, tt.expectHP)
			}
			if san != tt.expectSAN {
				t.Errorf("SAN = %d, want %d", san, tt.expectSAN)
			}
		})
	}
}

func TestStatsParser_ParseJSON(t *testing.T) {
	parser := NewStatsParser()

	tests := []struct {
		name        string
		json        string
		expectHP    int
		expectSAN   int
		expectFound bool
	}{
		{
			name:        "Valid JSON",
			json:        `{"narrative": "test", "stat_changes": {"hp": -15, "san": -5}}`,
			expectHP:    -15,
			expectSAN:   -5,
			expectFound: true,
		},
		{
			name:        "HP only",
			json:        `{"stat_changes": {"hp": -20}}`,
			expectHP:    -20,
			expectSAN:   0,
			expectFound: true,
		},
		{
			name:        "SAN only",
			json:        `{"stat_changes": {"san": -10}}`,
			expectHP:    0,
			expectSAN:   -10,
			expectFound: true,
		},
		{
			name:        "No stat_changes",
			json:        `{"narrative": "test"}`,
			expectHP:    0,
			expectSAN:   0,
			expectFound: false,
		},
		{
			name:        "Invalid JSON",
			json:        `{invalid}`,
			expectHP:    0,
			expectSAN:   0,
			expectFound: false,
		},
		{
			name:        "Empty stat_changes",
			json:        `{"stat_changes": {}}`,
			expectHP:    0,
			expectSAN:   0,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp, san, found := parser.ParseJSON(tt.json)

			if found != tt.expectFound {
				t.Errorf("found = %v, want %v", found, tt.expectFound)
			}
			if hp != tt.expectHP {
				t.Errorf("HP = %d, want %d", hp, tt.expectHP)
			}
			if san != tt.expectSAN {
				t.Errorf("SAN = %d, want %d", san, tt.expectSAN)
			}
		})
	}
}

func TestStatsParser_ParseKeywords(t *testing.T) {
	parser := NewStatsParser()

	tests := []struct {
		name      string
		text      string
		expectHP  int
		expectSAN int
	}{
		{
			name:      "Lose health",
			text:      "You lose significant health.",
			expectHP:  -15, // matches "lose significant health"
			expectSAN: 0,
		},
		{
			name:      "Take damage",
			text:      "You take heavy damage from the attack.",
			expectHP:  -15,
			expectSAN: 0,
		},
		{
			name:      "Terror",
			text:      "Terror grips your mind.",
			expectSAN: -15,
			expectHP:  0,
		},
		{
			name:      "Fear",
			text:      "Overwhelming fear washes over you.",
			expectSAN: -15, // matches "overwhelming fear"
			expectHP:  0,
		},
		{
			name:      "Both",
			text:      "You are wounded badly and terror fills your heart.",
			expectHP:  -15,
			expectSAN: -15,
		},
		{
			name:      "No keywords",
			text:      "You walk through the hallway.",
			expectHP:  0,
			expectSAN: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp, san := parser.ParseKeywords(tt.text)

			if hp != tt.expectHP {
				t.Errorf("HP = %d, want %d", hp, tt.expectHP)
			}
			if san != tt.expectSAN {
				t.Errorf("SAN = %d, want %d", san, tt.expectSAN)
			}
		})
	}
}

func TestStatsParser_Parse(t *testing.T) {
	parser := NewStatsParser()

	tests := []struct {
		name      string
		text      string
		expectHP  int
		expectSAN int
	}{
		{
			name:      "Inline format first",
			text:      "Story text. HP: -20. SAN: -10.",
			expectHP:  -20,
			expectSAN: -10,
		},
		{
			name:      "JSON format",
			text:      `{"stat_changes": {"hp": -15, "san": -8}}`,
			expectHP:  -15,
			expectSAN: -8,
		},
		{
			name:      "Fallback to keywords",
			text:      "You take heavy damage and feel terror.",
			expectHP:  -15,
			expectSAN: -15,
		},
		{
			name:      "No changes",
			text:      "You explore the room.",
			expectHP:  0,
			expectSAN: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp, san := parser.Parse(tt.text)

			if hp != tt.expectHP {
				t.Errorf("HP = %d, want %d", hp, tt.expectHP)
			}
			if san != tt.expectSAN {
				t.Errorf("SAN = %d, want %d", san, tt.expectSAN)
			}
		})
	}
}

func TestStatsParser_ValidateBounds(t *testing.T) {
	parser := NewStatsParser()

	tests := []struct {
		name     string
		hp       int
		san      int
		expectHP int
		expectSAN int
	}{
		{
			name:      "Within bounds",
			hp:        -20,
			san:       -10,
			expectHP:  -20,
			expectSAN: -10,
		},
		{
			name:      "HP too negative",
			hp:        -60,
			san:       -10,
			expectHP:  -50,
			expectSAN: -10,
		},
		{
			name:      "SAN too negative",
			hp:        -10,
			san:       -60,
			expectHP:  -10,
			expectSAN: -50,
		},
		{
			name:      "HP too positive",
			hp:        40,
			san:       10,
			expectHP:  30,
			expectSAN: 10,
		},
		{
			name:      "SAN too positive",
			hp:        10,
			san:       40,
			expectHP:  10,
			expectSAN: 30,
		},
		{
			name:      "Both exceed bounds",
			hp:        -100,
			san:       -100,
			expectHP:  -50,
			expectSAN: -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp, san := parser.ValidateBounds(tt.hp, tt.san)

			if hp != tt.expectHP {
				t.Errorf("HP = %d, want %d", hp, tt.expectHP)
			}
			if san != tt.expectSAN {
				t.Errorf("SAN = %d, want %d", san, tt.expectSAN)
			}
		})
	}
}
