package orchestrator

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// TestMockSeedManager_MarkSeedRevealed tests the unused mock method for 100% coverage.
func TestMockSeedManager_MarkSeedRevealed(t *testing.T) {
	mgr := NewMockSeedManager()

	err := mgr.MarkSeedRevealed("seed-001", 5)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestMockRuleEngine_CheckViolation tests the unused mock method for 100% coverage.
func TestMockRuleEngine_CheckViolation(t *testing.T) {
	engine := NewMockRuleEngine()

	violated, violation := engine.CheckViolation("test action")
	if violated {
		t.Error("Expected no violation in mock")
	}

	if violation.RuleName != "" {
		t.Errorf("Expected empty violation, got: %+v", violation)
	}
}

// TestMockSeedAgent_GenerateGlobal_AllDifficulties tests all difficulty levels.
func TestMockSeedAgent_GenerateGlobal_AllDifficulties(t *testing.T) {
	agent := NewMockSeedAgent()

	difficulties := []struct {
		name          string
		expectedCount int
	}{
		{"easy", 3},
		{"medium", 4},
		{"hard", 5},
		{"unknown", 3}, // Default case
	}

	for _, tc := range difficulties {
		t.Run(tc.name, func(t *testing.T) {
			seeds, err := agent.GenerateGlobal(nil, agents.GenerateGlobalParams{
				WorldView:  "test",
				MainTheme:  "test",
				Difficulty: tc.name,
			})

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if len(seeds) != tc.expectedCount {
				t.Errorf("Expected %d seeds for difficulty '%s', got %d",
					tc.expectedCount, tc.name, len(seeds))
			}
		})
	}
}
