package trinity

import "testing"

func TestTierLevel_String(t *testing.T) {
	tests := []struct {
		tier     TierLevel
		expected string
	}{
		{TierThinking, "Thinking"},
		{TierReactive, "Reactive"},
		{TierRapid, "Rapid"},
		{TierLevel(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.expected {
				t.Errorf("TierLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetTierForAgent(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		overrides map[string]TierLevel
		expected  TierLevel
	}{
		{
			name:      "JudgeAgent → Thinking (default)",
			agentName: "JudgeAgent",
			overrides: nil,
			expected:  TierThinking,
		},
		{
			name:      "NarrationAgent → Reactive (default)",
			agentName: "NarrationAgent",
			overrides: nil,
			expected:  TierReactive,
		},
		{
			name:      "DreamAgent → Rapid (default)",
			agentName: "DreamAgent",
			overrides: nil,
			expected:  TierRapid,
		},
		{
			name:      "Unknown agent → Reactive (fallback)",
			agentName: "UnknownAgent",
			overrides: nil,
			expected:  TierReactive,
		},
		{
			name:      "User override: DreamAgent → Thinking",
			agentName: "DreamAgent",
			overrides: map[string]TierLevel{"DreamAgent": TierThinking},
			expected:  TierThinking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTierForAgent(tt.agentName, tt.overrides); got != tt.expected {
				t.Errorf("GetTierForAgent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseTierLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected TierLevel
		ok       bool
	}{
		{"Thinking", TierThinking, true},
		{"thinking", TierThinking, true},
		{"Reactive", TierReactive, true},
		{"reactive", TierReactive, true},
		{"Rapid", TierRapid, true},
		{"rapid", TierRapid, true},
		{"invalid", TierReactive, false},
		{"", TierReactive, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseTierLevel(tt.input)
			if ok != tt.ok {
				t.Errorf("ParseTierLevel(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("ParseTierLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDefaultAgentTierMapping(t *testing.T) {
	// Verify key agents are mapped correctly
	requiredMappings := map[string]TierLevel{
		"JudgeAgent":     TierThinking,
		"NPCAgent":       TierThinking,
		"NarrationAgent": TierReactive,
		"ChoiceAgent":    TierReactive,
		"DreamAgent":     TierRapid,
	}

	for agent, expected := range requiredMappings {
		if got, ok := DefaultAgentTierMapping[agent]; !ok {
			t.Errorf("DefaultAgentTierMapping missing %q", agent)
		} else if got != expected {
			t.Errorf("DefaultAgentTierMapping[%q] = %v, want %v", agent, got, expected)
		}
	}
}
