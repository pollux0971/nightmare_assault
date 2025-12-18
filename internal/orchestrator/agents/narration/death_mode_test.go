package narration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.8 AC3: Death Narration Mode Tests
// ==========================================================================

func TestValidateDeathRequest(t *testing.T) {
	tests := []struct {
		name    string
		request *DeathNarrationRequest
		wantErr bool
	}{
		{
			name: "Valid death request",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				NPCArchetype:         agents.NPCArchetypeSacrificial,
				DeathBeat:            15,
				DeathReason:          "違反規則被殺",
				Location:             "走廊",
				PlayerChoice:         "鼓勵她進入房間",
				PlayerResponsibility: 0.9,
				Intimacy:             60,
				Adult18Plus:          false,
				CurrentHP:            70,
				CurrentSAN:           50,
			},
			wantErr: false,
		},
		{
			name:    "Nil request",
			request: nil,
			wantErr: true,
		},
		{
			name: "Empty NPC ID",
			request: &DeathNarrationRequest{
				NPCID:                "",
				NPCName:              "王小芳",
				DeathReason:          "違反規則",
				DeathBeat:            15,
				PlayerResponsibility: 0.5,
				Intimacy:             50,
			},
			wantErr: true,
		},
		{
			name: "Empty NPC name",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "",
				DeathReason:          "違反規則",
				DeathBeat:            15,
				PlayerResponsibility: 0.5,
				Intimacy:             50,
			},
			wantErr: true,
		},
		{
			name: "Empty death reason",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				DeathReason:          "",
				DeathBeat:            15,
				PlayerResponsibility: 0.5,
				Intimacy:             50,
			},
			wantErr: true,
		},
		{
			name: "Negative death beat",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				DeathReason:          "違反規則",
				DeathBeat:            -1,
				PlayerResponsibility: 0.5,
				Intimacy:             50,
			},
			wantErr: true,
		},
		{
			name: "Responsibility out of range (negative)",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				DeathReason:          "違反規則",
				DeathBeat:            15,
				PlayerResponsibility: -0.1,
				Intimacy:             50,
			},
			wantErr: true,
		},
		{
			name: "Responsibility out of range (>1)",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				DeathReason:          "違反規則",
				DeathBeat:            15,
				PlayerResponsibility: 1.5,
				Intimacy:             50,
			},
			wantErr: true,
		},
		{
			name: "Intimacy out of range (negative)",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				DeathReason:          "違反規則",
				DeathBeat:            15,
				PlayerResponsibility: 0.5,
				Intimacy:             -10,
			},
			wantErr: true,
		},
		{
			name: "Intimacy out of range (>100)",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				DeathReason:          "違反規則",
				DeathBeat:            15,
				PlayerResponsibility: 0.5,
				Intimacy:             150,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeathRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDeathRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildDeathPrompt(t *testing.T) {
	agent := &NarrationAgent{
		config: agents.AgentConfig{
			Name: "TestNarrationAgent",
		},
	}

	tests := []struct {
		name           string
		request        *DeathNarrationRequest
		expectContains []string // Strings that should appear in prompt
	}{
		{
			name: "Basic death prompt",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				NPCArchetype:         agents.NPCArchetypeSacrificial,
				NPCBackstory:         "一個無助的護士",
				DeathBeat:            15,
				DeathReason:          "違反規則被殺",
				Location:             "走廊",
				PlayerChoice:         "鼓勵她進入房間",
				PlayerResponsibility: 0.9,
				Intimacy:             60,
				Adult18Plus:          false,
				CurrentHP:            70,
				CurrentSAN:           50,
			},
			expectContains: []string{
				"王小芳",
				"犧牲者",
				"Beat: 15",
				"違反規則被殺",
				"玩家責任程度",
				"親密度",
				"普通模式",
			},
		},
		{
			name: "Death prompt with 18+ mode",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "李醫生",
				NPCArchetype:         agents.NPCArchetypeKnowledgeable,
				DeathBeat:            20,
				DeathReason:          "被怪物撕碎",
				Location:             "手術室",
				PlayerChoice:         "讓他獨自調查",
				PlayerResponsibility: 0.8,
				Intimacy:             80,
				Adult18Plus:          true,
				CurrentHP:            60,
				CurrentSAN:           40,
			},
			expectContains: []string{
				"李醫生",
				"18+ 成人模式",
				"詳細、具體、寫實",
				"被怪物撕碎",
			},
		},
		{
			name: "Death prompt with foreshadows",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "張護士",
				NPCArchetype:         agents.NPCArchetypeNeutral,
				DeathBeat:            18,
				DeathReason:          "中毒死亡",
				Location:             "藥房",
				PlayerChoice:         "觀察她的行動",
				PlayerResponsibility: 0.5,
				Intimacy:             50,
				Adult18Plus:          false,
				CurrentHP:            80,
				CurrentSAN:           60,
				Foreshadows: []orchestrator.DeathForeshadow{
					{
						Beat:    15,
						Type:    "clue",
						Content: "發現藥品被人動過手腳",
					},
					{
						Beat:    17,
						Type:    "environment",
						Content: "空氣中有奇怪的氣味",
					},
				},
			},
			expectContains: []string{
				"張護士",
				"先前的伏筆線索",
				"Beat 15",
				"發現藥品被人動過手腳",
				"Beat 17",
				"空氣中有奇怪的氣味",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := agent.buildDeathPrompt(tt.request)

			if prompt == "" {
				t.Errorf("buildDeathPrompt() returned empty prompt")
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(prompt, expected) {
					t.Errorf("buildDeathPrompt() prompt does not contain expected string: %q", expected)
				}
			}

			// Check basic structure
			if !strings.Contains(prompt, "JSON") {
				t.Errorf("buildDeathPrompt() prompt does not mention JSON format")
			}

			if !strings.Contains(prompt, "death_narrative") {
				t.Errorf("buildDeathPrompt() prompt does not mention death_narrative field")
			}

			if !strings.Contains(prompt, "last_words") {
				t.Errorf("buildDeathPrompt() prompt does not mention last_words field")
			}
		})
	}
}

func TestValidateDeathResponse(t *testing.T) {
	agent := &NarrationAgent{
		config: agents.AgentConfig{
			Name: "TestNarrationAgent",
		},
	}

	tests := []struct {
		name     string
		response *DeathNarrationResponse
		wantErr  bool
	}{
		{
			name: "Valid death response",
			response: &DeathNarrationResponse{
				DeathNarrative: strings.Repeat("這是一段足夠長的死亡敘事內容。", 16), // ~224 chars (200-400 range)
				LastWords:      "這是遺言內容，應該在三十到八十字之間。這樣應該就夠長了吧呢。", // ~30 chars (30-80 range)
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: false,
		},
		{
			name: "Narrative too short",
			response: &DeathNarrationResponse{
				DeathNarrative: "太短",
				LastWords:      "這是遺言內容，應該在三十到八十字之間。",
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: true,
		},
		{
			name: "Narrative too long",
			response: &DeathNarrationResponse{
				DeathNarrative: strings.Repeat("這是一段非常非常長的死亡敘事內容。", 30), // >400 chars
				LastWords:      "這是遺言內容，應該在三十到八十字之間。這樣應該就夠長了吧呢。",
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: true,
		},
		{
			name: "Last words too short",
			response: &DeathNarrationResponse{
				DeathNarrative: strings.Repeat("這是一段足夠長的死亡敘事內容。", 16),
				LastWords:      "太短了", // < 30 chars
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: true,
		},
		{
			name: "Last words too long",
			response: &DeathNarrationResponse{
				DeathNarrative: strings.Repeat("這是一段足夠長的死亡敘事內容。", 16),
				LastWords:      strings.Repeat("這是一段非常非常長的遺言內容。", 7), // >80 chars
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: true,
		},
		{
			name: "Invalid emotional tone",
			response: &DeathNarrationResponse{
				DeathNarrative: strings.Repeat("這是一段足夠長的死亡敘事內容。", 16),
				LastWords:      "這是遺言內容，應該在三十到八十字之間。這樣應該就夠長了吧呢。",
				EmotionalTone:  "invalid_tone",
				SANLoss:        -20,
			},
			wantErr: true,
		},
		{
			name: "Empty narrative",
			response: &DeathNarrationResponse{
				DeathNarrative: "",
				LastWords:      "這是遺言內容，應該在三十到八十字之間。這樣應該就夠長了吧呢。",
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: true,
		},
		{
			name: "Empty last words",
			response: &DeathNarrationResponse{
				DeathNarrative: strings.Repeat("這是一段足夠長的死亡敘事內容。", 16),
				LastWords:      "",
				EmotionalTone:  "guilt",
				SANLoss:        -20,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agent.validateDeathResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDeathResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSimpleDeathNarrative(t *testing.T) {
	agent := &NarrationAgent{
		config: agents.AgentConfig{
			Name: "TestNarrationAgent",
		},
	}

	tests := []struct {
		name                 string
		request              *DeathNarrationRequest
		expectEmotionalTone  string
		expectMinNarrativeLen int
		expectMinLastWordsLen int
	}{
		{
			name: "High responsibility death",
			request: &DeathNarrationRequest{
				NPCID:                "npc-001",
				NPCName:              "王小芳",
				NPCArchetype:         agents.NPCArchetypeSacrificial,
				DeathBeat:            15,
				DeathReason:          "違反規則被殺",
				Location:             "走廊",
				PlayerChoice:         "鼓勵她進入房間",
				PlayerResponsibility: 0.9,
				Intimacy:             60,
				Adult18Plus:          false,
				CurrentHP:            70,
				CurrentSAN:           50,
			},
			expectEmotionalTone:  "guilt",
			expectMinNarrativeLen: 100,
			expectMinLastWordsLen: 20,
		},
		{
			name: "Low responsibility death",
			request: &DeathNarrationRequest{
				NPCID:                "npc-002",
				NPCName:              "李醫生",
				NPCArchetype:         agents.NPCArchetypeKnowledgeable,
				DeathBeat:            20,
				DeathReason:          "被怪物殺死",
				Location:             "手術室",
				PlayerChoice:         "試圖救他但失敗了",
				PlayerResponsibility: 0.1,
				Intimacy:             80,
				Adult18Plus:          false,
				CurrentHP:            60,
				CurrentSAN:           40,
			},
			expectEmotionalTone:  "helplessness",
			expectMinNarrativeLen: 100,
			expectMinLastWordsLen: 20,
		},
		{
			name: "Medium responsibility death",
			request: &DeathNarrationRequest{
				NPCID:                "npc-003",
				NPCName:              "張護士",
				NPCArchetype:         agents.NPCArchetypeNeutral,
				DeathBeat:            18,
				DeathReason:          "中毒死亡",
				Location:             "藥房",
				PlayerChoice:         "觀察她的行動",
				PlayerResponsibility: 0.5,
				Intimacy:             50,
				Adult18Plus:          false,
				CurrentHP:            80,
				CurrentSAN:           60,
			},
			expectEmotionalTone:  "tragedy",
			expectMinNarrativeLen: 100,
			expectMinLastWordsLen: 20,
		},
		{
			name: "18+ mode death",
			request: &DeathNarrationRequest{
				NPCID:                "npc-004",
				NPCName:              "陳先生",
				NPCArchetype:         agents.NPCArchetypeHostile,
				DeathBeat:            22,
				DeathReason:          "被撕成碎片",
				Location:             "地下室",
				PlayerChoice:         "讓他自生自滅",
				PlayerResponsibility: 0.8,
				Intimacy:             30,
				Adult18Plus:          true,
				CurrentHP:            50,
				CurrentSAN:           30,
			},
			expectEmotionalTone:  "guilt",
			expectMinNarrativeLen: 100,
			expectMinLastWordsLen: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := agent.GenerateSimpleDeathNarrative(tt.request)

			if response == nil {
				t.Fatalf("GenerateSimpleDeathNarrative() returned nil")
			}

			// Check emotional tone
			if response.EmotionalTone != tt.expectEmotionalTone {
				t.Errorf("EmotionalTone = %v, want %v", response.EmotionalTone, tt.expectEmotionalTone)
			}

			// Check narrative length
			narrativeLen := len([]rune(response.DeathNarrative))
			if narrativeLen < tt.expectMinNarrativeLen {
				t.Errorf("DeathNarrative length = %d, want >= %d", narrativeLen, tt.expectMinNarrativeLen)
			}

			// Check last words length
			lastWordsLen := len([]rune(response.LastWords))
			if lastWordsLen < tt.expectMinLastWordsLen {
				t.Errorf("LastWords length = %d, want >= %d", lastWordsLen, tt.expectMinLastWordsLen)
			}

			// Check SAN loss range
			if response.SANLoss < -25 || response.SANLoss > -15 {
				t.Errorf("SANLoss = %d, want range [-25, -15]", response.SANLoss)
			}

			// Check that NPC name appears in narrative
			if !strings.Contains(response.DeathNarrative, tt.request.NPCName) {
				t.Errorf("DeathNarrative does not contain NPC name %q", tt.request.NPCName)
			}

			// Check that NPC name appears in last words
			if !strings.Contains(response.LastWords, tt.request.NPCName) {
				t.Errorf("LastWords does not contain NPC name %q", tt.request.NPCName)
			}

			// Check that death reason appears in narrative
			if !strings.Contains(response.DeathNarrative, tt.request.DeathReason) {
				t.Errorf("DeathNarrative does not contain death reason %q", tt.request.DeathReason)
			}
		})
	}
}

func TestGetResponsibilityGuidance(t *testing.T) {
	tests := []struct {
		name             string
		responsibility   float64
		expectContains   []string
	}{
		{
			name:           "High responsibility",
			responsibility: 0.9,
			expectContains: []string{"玩家責任極高", "這是玩家的錯", "罪惡感"},
		},
		{
			name:           "Medium responsibility",
			responsibility: 0.5,
			expectContains: []string{"玩家責任中等", "本可避免的悲劇", "悲劇感"},
		},
		{
			name:           "Low responsibility",
			responsibility: 0.1,
			expectContains: []string{"玩家責任較低", "無法改變的命運", "無力感"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guidance := getResponsibilityGuidance(tt.responsibility)

			if guidance == "" {
				t.Errorf("getResponsibilityGuidance() returned empty string")
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(guidance, expected) {
					t.Errorf("guidance does not contain expected string: %q", expected)
				}
			}
		})
	}
}

// ==========================================================================
// Integration Tests (would require mock LLM client)
// ==========================================================================

// TestInvokeDeath_Integration tests the full death narration flow
// This test would require a mock LLM client, so it's skipped in unit tests
func TestInvokeDeath_Integration(t *testing.T) {
	t.Skip("Integration test requires mock LLM client")

	// Create mock LLM client
	mockLLM := &MockLLMClient{}

	// Create narration agent
	agent := NewNarrationAgent(agents.AgentConfig{
		Name:      "TestNarrationAgent",
		LLMClient: mockLLM,
		Timeout:   5 * time.Second,
	})

	// Create death request
	request := &DeathNarrationRequest{
		NPCID:                "npc-001",
		NPCName:              "王小芳",
		NPCArchetype:         agents.NPCArchetypeSacrificial,
		DeathBeat:            15,
		DeathReason:          "違反規則被殺",
		Location:             "走廊",
		PlayerChoice:         "鼓勵她進入房間",
		PlayerResponsibility: 0.9,
		Intimacy:             60,
		Adult18Plus:          false,
		CurrentHP:            70,
		CurrentSAN:           50,
	}

	// Invoke death narration
	ctx := context.Background()
	response, err := agent.InvokeDeath(ctx, request)

	if err != nil {
		t.Fatalf("InvokeDeath() unexpected error: %v", err)
	}

	if response == nil {
		t.Fatalf("InvokeDeath() returned nil response")
	}

	// Validate response
	if err := agent.validateDeathResponse(response); err != nil {
		t.Errorf("Response validation failed: %v", err)
	}
}

// MockLLMClient is a mock implementation of LLMClient for testing
type MockLLMClient struct{}

func (m *MockLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	// Return mock JSON response
	return `{
		"death_narrative": "` + strings.Repeat("這是模擬的死亡敘事內容。", 15) + `",
		"last_words": "這是模擬的遺言內容，應該在三十到八十字之間。這樣應該就夠長了。",
		"emotional_tone": "guilt"
	}`, nil
}
