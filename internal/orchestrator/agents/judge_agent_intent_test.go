package agents

import (
	"context"
	"testing"
	"time"
)

// ==========================================================================
// Story 7.3: Intent Classification Tests
// ==========================================================================

// TestClassifyIntent_ClearIntent tests intent classification with clear input
func TestClassifyIntent_ClearIntent(t *testing.T) {
	// Setup mock LLM client
	mockClient := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			// Return clear intent classification
			return `{
				"action": "檢查",
				"target": "鏡子",
				"is_ambiguous": false,
				"confidence": 0.95,
				"keywords": ["檢查", "鏡子"],
				"normalized_intent": "examine_mirror",
				"clarification_reason": "",
				"suggested_interpretations": [],
				"clarification_question": ""
			}`, nil
		},
	}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:           80,
		SAN:          70,
		CurrentScene: "房間",
		Difficulty:   "normal",
	}

	ctx := context.Background()
	intent, clarification, err := ja.ClassifyIntent(ctx, "我想檢查房間裡的鏡子", gameState)

	// Verify results
	if err != nil {
		t.Fatalf("ClassifyIntent failed: %v", err)
	}

	if clarification != nil {
		t.Errorf("Expected no clarification, got: %+v", clarification)
	}

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	// Check intent fields
	if intent.Action != "檢查" {
		t.Errorf("Expected action '檢查', got '%s'", intent.Action)
	}

	if intent.Target != "鏡子" {
		t.Errorf("Expected target '鏡子', got '%s'", intent.Target)
	}

	if intent.IsAmbiguous {
		t.Error("Expected IsAmbiguous=false, got true")
	}

	if intent.Confidence < 0.9 {
		t.Errorf("Expected high confidence, got %.2f", intent.Confidence)
	}

	if intent.NormalizedIntent != "examine_mirror" {
		t.Errorf("Expected normalized_intent 'examine_mirror', got '%s'", intent.NormalizedIntent)
	}
}

// TestClassifyIntent_AmbiguousIntent tests intent classification with ambiguous input
func TestClassifyIntent_AmbiguousIntent(t *testing.T) {
	// Setup mock LLM client
	mockClient := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			// Return ambiguous intent classification
			return `{
				"action": "檢查",
				"target": "?",
				"is_ambiguous": true,
				"confidence": 0.4,
				"keywords": ["看", "東西"],
				"normalized_intent": "examine_unknown",
				"clarification_reason": "目標不明確",
				"suggested_interpretations": ["檢查鏡子", "檢查桌子", "檢查門"],
				"clarification_question": "你想檢查哪個物品？"
			}`, nil
		},
	}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:           80,
		SAN:          70,
		CurrentScene: "房間",
		Difficulty:   "normal",
	}

	ctx := context.Background()
	intent, clarification, err := ja.ClassifyIntent(ctx, "看看那個東西", gameState)

	// Verify results
	if err != nil {
		t.Fatalf("ClassifyIntent failed: %v", err)
	}

	if clarification == nil {
		t.Fatal("Expected clarification, got nil")
	}

	// Check clarification fields
	if clarification.Reason != "目標不明確" {
		t.Errorf("Expected reason '目標不明確', got '%s'", clarification.Reason)
	}

	if clarification.Question != "你想檢查哪個物品？" {
		t.Errorf("Expected question '你想檢查哪個物品？', got '%s'", clarification.Question)
	}

	if len(clarification.SuggestedInterpretations) != 3 {
		t.Errorf("Expected 3 suggested interpretations, got %d", len(clarification.SuggestedInterpretations))
	}

	// Intent should still be returned
	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	if !intent.IsAmbiguous {
		t.Error("Expected IsAmbiguous=true, got false")
	}
}

// TestClassifyIntent_EmptyInput tests handling of empty input
func TestClassifyIntent_EmptyInput(t *testing.T) {
	mockClient := &MockLLMClient{}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:           80,
		SAN:          70,
		CurrentScene: "房間",
		Difficulty:   "normal",
	}

	ctx := context.Background()
	intent, clarification, err := ja.ClassifyIntent(ctx, "", gameState)

	// Verify results
	if err != nil {
		t.Fatalf("ClassifyIntent failed: %v", err)
	}

	if clarification == nil {
		t.Fatal("Expected clarification for empty input, got nil")
	}

	if clarification.Reason != "輸入為空" {
		t.Errorf("Expected reason '輸入為空', got '%s'", clarification.Reason)
	}

	// Intent should be nil for empty input
	if intent != nil {
		t.Errorf("Expected nil intent for empty input, got %+v", intent)
	}
}

// TestClassifyIntent_LongInput tests handling of very long input
func TestClassifyIntent_LongInput(t *testing.T) {
	mockClient := &MockLLMClient{}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:           80,
		SAN:          70,
		CurrentScene: "房間",
		Difficulty:   "normal",
	}

	// Create input >200 chars
	longInput := "我想要仔細檢查房間裡的每一個角落和物品，包括鏡子、桌子、椅子、床、窗戶、門、衣櫃、書架、地板、天花板、牆壁、畫框、燈具、開關、插座、地毯、窗簾、枕頭、被子、床單、書本、筆、紙、鑰匙、手機、錢包、包包、衣服、鞋子、帽子、手套、圍巾、襪子、內衣、褲子、裙子、上衣、外套、背心、連衣裙、領帶、皮帶、眼鏡、手錶、戒指、項鍊、耳環、手鐲、腳鍊等等所有東西，希望能找到線索或者發現什麼異常的地方，同時要小心謹慎避免觸發任何危險的機關或者規則，確保自己的安全。"

	ctx := context.Background()
	intent, clarification, err := ja.ClassifyIntent(ctx, longInput, gameState)

	// Verify results
	if err != nil {
		t.Fatalf("ClassifyIntent failed: %v", err)
	}

	if clarification == nil {
		t.Fatal("Expected clarification for long input, got nil")
	}

	if clarification.Reason != "輸入過長，請簡化描述" {
		t.Errorf("Expected reason about long input, got '%s'", clarification.Reason)
	}

	// Intent should be nil for long input
	if intent != nil {
		t.Errorf("Expected nil intent for long input, got %+v", intent)
	}
}

// TestInvokeJudgeWithIntent_ClearIntent tests full flow with clear intent
func TestInvokeJudgeWithIntent_ClearIntent(t *testing.T) {
	// Setup mock LLM client
	mockClient := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			// Return clear intent classification
			return `{
				"action": "攻擊",
				"target": "鏡子",
				"is_ambiguous": false,
				"confidence": 0.90,
				"keywords": ["打", "鏡子"],
				"normalized_intent": "attack_mirror",
				"clarification_reason": "",
				"suggested_interpretations": [],
				"clarification_question": ""
			}`, nil
		},
	}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	// Create a rule that triggers on "attack_mirror"
	activeRules := []JudgeHiddenRule{
		{
			ID:               "R-01",
			Name:             "不可破壞鏡子",
			Type:             RuleTypeBehavior,
			TriggerKeywords:  []string{"attack", "mirror", "打", "鏡子", "破壞"},
			TriggerCondition: "攻擊或破壞鏡子",
			Punishment: RulePunishment{
				IsFatal:   false,
				HPDamage:  30,
				SANDamage: 20,
			},
			MaxWarnings:  1,
			DirectHint:   "不要破壞鏡子",
			MetaphorHint: "反射的真相不可觸碰",
		},
	}

	gameState := &GameStateSnapshot{
		HP:            80,
		SAN:           70,
		CurrentScene:  "房間",
		Difficulty:    "normal",
		RuleWarnings:  make(map[string]int),
		PlayerItems:   []string{},
		TurnNumber:    5,
	}

	ctx := context.Background()
	response, err := ja.InvokeJudgeWithIntent(ctx, "我想打破鏡子", gameState, activeRules)

	// Verify results
	if err != nil {
		t.Fatalf("InvokeJudgeWithIntent failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Should have no clarification
	if response.ClarificationNeeded != nil {
		t.Errorf("Expected no clarification, got: %+v", response.ClarificationNeeded)
	}

	// Should have intent classification
	if response.IntentClassification == nil {
		t.Fatal("Expected intent classification, got nil")
	}

	// Should have detected rule violation
	if len(response.RulesViolated) == 0 {
		t.Error("Expected rule violation, got none")
	}

	// Should have impact level > None
	if response.ImpactLevel == ImpactNone {
		t.Error("Expected impact > None for rule violation")
	}

	// Should have state changes (HP/SAN damage)
	if response.SuggestedStateChanges.HP >= 0 {
		t.Error("Expected negative HP change (damage)")
	}

	if response.SuggestedStateChanges.SAN >= 0 {
		t.Error("Expected negative SAN change (damage)")
	}
}

// TestInvokeJudgeWithIntent_NeedsClarification tests flow with ambiguous intent
func TestInvokeJudgeWithIntent_NeedsClarification(t *testing.T) {
	// Setup mock LLM client
	mockClient := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			// Return ambiguous intent classification
			return `{
				"action": "使用",
				"target": "?",
				"is_ambiguous": true,
				"confidence": 0.3,
				"keywords": ["用", "它"],
				"normalized_intent": "use_unknown",
				"clarification_reason": "目標不明確，不知道要使用什麼物品",
				"suggested_interpretations": ["使用鑰匙", "使用手電筒", "使用道具"],
				"clarification_question": "你想使用什麼物品？"
			}`, nil
		},
	}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:            80,
		SAN:           70,
		CurrentScene:  "房間",
		Difficulty:    "normal",
		RuleWarnings:  make(map[string]int),
		PlayerItems:   []string{"鑰匙", "手電筒"},
		TurnNumber:    5,
	}

	ctx := context.Background()
	response, err := ja.InvokeJudgeWithIntent(ctx, "我想用它", gameState, []JudgeHiddenRule{})

	// Verify results
	if err != nil {
		t.Fatalf("InvokeJudgeWithIntent failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Should have clarification
	if response.ClarificationNeeded == nil {
		t.Fatal("Expected clarification, got nil")
	}

	// Should have intent classification
	if response.IntentClassification == nil {
		t.Fatal("Expected intent classification, got nil")
	}

	// Should have no rule violations (because clarification is needed first)
	if len(response.RulesViolated) != 0 {
		t.Error("Expected no rule violations when clarification is needed")
	}

	// Impact level should be None
	if response.ImpactLevel != ImpactNone {
		t.Errorf("Expected ImpactNone when clarification is needed, got %s", response.ImpactLevel)
	}

	// Next action should be ContinueStory
	if response.NextAction != ActionContinueStory {
		t.Errorf("Expected ActionContinueStory when clarification is needed, got %s", response.NextAction)
	}
}

// TestInvokeJudgeWithIntent_LowConfidence tests handling of low confidence classification
func TestInvokeJudgeWithIntent_LowConfidence(t *testing.T) {
	// Setup mock LLM client
	mockClient := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			// Return low confidence classification
			return `{
				"action": "做",
				"target": "某事",
				"is_ambiguous": false,
				"confidence": 0.5,
				"keywords": ["做", "某事"],
				"normalized_intent": "do_something",
				"clarification_reason": "意圖不夠明確，信心度較低",
				"suggested_interpretations": ["檢查環境", "移動到其他地方", "等待觀察"],
				"clarification_question": "請更具體描述你想做什麼？"
			}`, nil
		},
	}

	config := AgentConfig{
		Name:       "TestJudgeAgent",
		LLMClient:  mockClient,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:            80,
		SAN:           70,
		CurrentScene:  "房間",
		Difficulty:    "normal",
		RuleWarnings:  make(map[string]int),
		PlayerItems:   []string{},
		TurnNumber:    5,
	}

	ctx := context.Background()
	response, err := ja.InvokeJudgeWithIntent(ctx, "我想做某事", gameState, []JudgeHiddenRule{})

	// Verify results
	if err != nil {
		t.Fatalf("InvokeJudgeWithIntent failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Should have clarification (due to low confidence < 0.6)
	if response.ClarificationNeeded == nil {
		t.Fatal("Expected clarification for low confidence, got nil")
	}

	// Should have intent classification
	if response.IntentClassification == nil {
		t.Fatal("Expected intent classification, got nil")
	}

	// Confidence should be low
	if response.IntentClassification.Confidence >= 0.6 {
		t.Errorf("Expected low confidence (<0.6), got %.2f", response.IntentClassification.Confidence)
	}
}

// TestBuildIntentPrompt tests intent classification prompt building
func TestBuildIntentPrompt(t *testing.T) {
	config := AgentConfig{
		Name:       "TestJudgeAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ja := NewJudgeAgent(config)

	gameState := &GameStateSnapshot{
		HP:           60,
		SAN:          50,
		CurrentScene: "走廊",
		Difficulty:   "hard",
		PlayerItems:  []string{"鑰匙", "手電筒"},
		TurnNumber:   10,
	}

	prompt := ja.buildIntentPrompt("我想用鑰匙開門", gameState)

	// Verify prompt contains key elements
	testCases := []struct {
		expected string
		desc     string
	}{
		{"意圖解析", "should mention intent parsing"},
		{"動作】(action)", "should mention action extraction"},
		{"目標】(target)", "should mention target extraction"},
		{"is_ambiguous", "should mention ambiguity detection"},
		{"confidence", "should mention confidence scoring"},
		{"normalized_intent", "should mention intent normalization"},
		{"場景: 走廊", "should include current scene"},
		{"HP: 60", "should include HP"},
		{"SAN: 50", "should include SAN"},
		{"持有物品: 鑰匙, 手電筒", "should include player items"},
		{"我想用鑰匙開門", "should include player input"},
	}

	for _, tc := range testCases {
		if !containsString(prompt, tc.expected) {
			t.Errorf("Prompt %s - expected to contain '%s'", tc.desc, tc.expected)
		}
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
