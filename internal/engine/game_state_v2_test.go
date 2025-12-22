package engine

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

// Test Task 1: GameStateV2 struct definition
func TestNewGameStateV2(t *testing.T) {
	state := NewGameStateV2()

	if state == nil {
		t.Fatal("NewGameStateV2() returned nil")
	}

	// 檢查基礎欄位默認值
	if state.GetHP() != 100 {
		t.Errorf("Expected HP=100, got %d", state.GetHP())
	}

	if state.GetSAN() != 100 {
		t.Errorf("Expected SAN=100, got %d", state.GetSAN())
	}

	if state.GetCurrentBeat() != 0 {
		t.Errorf("Expected CurrentBeat=0, got %d", state.GetCurrentBeat())
	}

	// 檢查 GameID 不為空
	if state.GameID == "" {
		t.Error("GameID should not be empty")
	}

	// 檢查切片已初始化
	if state.GetGlobalSeeds() == nil {
		t.Error("GlobalSeeds should be initialized")
	}

	if state.GetLocalSeeds() == nil {
		t.Error("LocalSeeds should be initialized")
	}
}

// Test Story 7.4: TakeDamage method
func TestGameStateV2_TakeDamage(t *testing.T) {
	state := NewGameStateV2()

	tests := []struct {
		name        string
		hpDelta     int
		sanDelta    int
		reason      string
		expectedHP  int
		expectedSAN int
	}{
		{"Normal HP damage", -20, 0, "test", 80, 100},
		{"Normal SAN damage", 0, -30, "test", 80, 70},
		{"Both damage", -10, -15, "test", 70, 55},
		{"HP to zero", -70, 0, "test", 0, 55},
		{"SAN to zero", 0, -55, "test", 0, 0},
		{"Negative clamp HP", -50, 0, "test", 0, 0},
		{"Negative clamp SAN", 0, -50, "test", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newHP, newSAN := state.TakeDamage(tt.hpDelta, tt.sanDelta, tt.reason)

			if newHP != tt.expectedHP {
				t.Errorf("Expected HP=%d, got %d", tt.expectedHP, newHP)
			}

			if newSAN != tt.expectedSAN {
				t.Errorf("Expected SAN=%d, got %d", tt.expectedSAN, newSAN)
			}

			// Verify GetHP/GetSAN return the same values
			if state.GetHP() != tt.expectedHP {
				t.Errorf("GetHP()=%d, want %d", state.GetHP(), tt.expectedHP)
			}

			if state.GetSAN() != tt.expectedSAN {
				t.Errorf("GetSAN()=%d, want %d", state.GetSAN(), tt.expectedSAN)
			}
		})
	}
}

// Test Story 7.4: Heal method
func TestGameStateV2_Heal(t *testing.T) {
	state := NewGameStateV2()

	// First take damage
	state.TakeDamage(-50, -40, "initial damage")

	tests := []struct {
		name        string
		hpHeal      int
		sanHeal     int
		reason      string
		expectedHP  int
		expectedSAN int
	}{
		{"Normal HP heal", 20, 0, "test", 70, 60},
		{"Normal SAN heal", 0, 15, "test", 70, 75},
		{"Both heal", 10, 10, "test", 80, 85},
		{"HP over max", 50, 0, "test", 100, 85},
		{"SAN over max", 0, 50, "test", 100, 100},
		{"Heal at max", 10, 10, "test", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newHP, newSAN := state.Heal(tt.hpHeal, tt.sanHeal, tt.reason)

			if newHP != tt.expectedHP {
				t.Errorf("Expected HP=%d, got %d", tt.expectedHP, newHP)
			}

			if newSAN != tt.expectedSAN {
				t.Errorf("Expected SAN=%d, got %d", tt.expectedSAN, newSAN)
			}

			// Verify GetHP/GetSAN return the same values
			if state.GetHP() != tt.expectedHP {
				t.Errorf("GetHP()=%d, want %d", state.GetHP(), tt.expectedHP)
			}

			if state.GetSAN() != tt.expectedSAN {
				t.Errorf("GetSAN()=%d, want %d", state.GetSAN(), tt.expectedSAN)
			}
		})
	}
}

// Test Story 7.4: IsDead method
func TestGameStateV2_IsDead(t *testing.T) {
	tests := []struct {
		name     string
		hp       int
		expected bool
	}{
		{"Full HP", 100, false},
		{"Mid HP", 50, false},
		{"Low HP", 1, false},
		{"Zero HP", 0, true},
		{"Negative HP", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameStateV2()
			state.SetHP(tt.hp)

			result := state.IsDead()
			if result != tt.expected {
				t.Errorf("IsDead() with HP=%d: got %v, want %v", tt.hp, result, tt.expected)
			}
		})
	}
}

// Test Story 7.4: IsInsane method
func TestGameStateV2_IsInsane(t *testing.T) {
	tests := []struct {
		name     string
		san      int
		expected bool
	}{
		{"Full SAN", 100, false},
		{"Mid SAN", 50, false},
		{"Low SAN", 1, false},
		{"Zero SAN", 0, true},
		{"Negative SAN", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameStateV2()
			state.SetSAN(tt.san)

			result := state.IsInsane()
			if result != tt.expected {
				t.Errorf("IsInsane() with SAN=%d: got %v, want %v", tt.san, result, tt.expected)
			}
		})
	}
}

// Test Story 7.4: HP/SAN bounds enforcement
func TestGameStateV2_BoundsEnforcement(t *testing.T) {
	state := NewGameStateV2()

	// Test upper bounds
	state.TakeDamage(50, 50, "positive healing via TakeDamage")
	if state.GetHP() != 100 {
		t.Errorf("HP should be clamped to 100, got %d", state.GetHP())
	}
	if state.GetSAN() != 100 {
		t.Errorf("SAN should be clamped to 100, got %d", state.GetSAN())
	}

	// Test lower bounds
	state.TakeDamage(-200, -200, "excessive damage")
	if state.GetHP() != 0 {
		t.Errorf("HP should be clamped to 0, got %d", state.GetHP())
	}
	if state.GetSAN() != 0 {
		t.Errorf("SAN should be clamped to 0, got %d", state.GetSAN())
	}

	// Test heal from zero
	state.Heal(50, 50, "heal from zero")
	if state.GetHP() != 50 {
		t.Errorf("Expected HP=50 after heal, got %d", state.GetHP())
	}
	if state.GetSAN() != 50 {
		t.Errorf("Expected SAN=50 after heal, got %d", state.GetSAN())
	}
}

// Test Story 7.4: Thread safety for HP/SAN operations
func TestGameStateV2_HPSANThreadSafety(t *testing.T) {
	state := NewGameStateV2()
	var wg sync.WaitGroup

	// Spawn multiple goroutines modifying HP/SAN
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			state.TakeDamage(-1, -1, "concurrent damage")
		}()

		go func() {
			defer wg.Done()
			state.Heal(1, 1, "concurrent heal")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetHP()
			_ = state.GetSAN()
			_ = state.IsDead()
			_ = state.IsInsane()
		}()
	}

	wg.Wait()

	// Just verify no panic occurred and values are in valid range
	hp := state.GetHP()
	san := state.GetSAN()

	if hp < 0 || hp > 100 {
		t.Errorf("HP out of bounds after concurrent operations: %d", hp)
	}

	if san < 0 || san > 100 {
		t.Errorf("SAN out of bounds after concurrent operations: %d", san)
	}
}

// Test Task 2: 線程安全方法
func TestGameStateV2_ThreadSafety(t *testing.T) {
	state := NewGameStateV2()

	// 並發讀寫測試
	var wg sync.WaitGroup
	iterations := 100

	// 併發寫入 HP
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(val int) {
			defer wg.Done()
			state.SetHP(val)
		}(i)
	}

	// 併發讀取 HP
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = state.GetHP()
		}()
	}

	wg.Wait()
	// 測試通過表示沒有 race condition
}

func TestGameStateV2_SetGetHP(t *testing.T) {
	state := NewGameStateV2()

	testCases := []int{0, 50, 100, 150}
	for _, hp := range testCases {
		state.SetHP(hp)
		if got := state.GetHP(); got != hp {
			t.Errorf("SetHP(%d): got %d", hp, got)
		}
	}
}

func TestGameStateV2_SetGetSAN(t *testing.T) {
	state := NewGameStateV2()

	testCases := []int{0, 30, 70, 100}
	for _, san := range testCases {
		state.SetSAN(san)
		if got := state.GetSAN(); got != san {
			t.Errorf("SetSAN(%d): got %d", san, got)
		}
	}
}

func TestGameStateV2_IncrementBeat(t *testing.T) {
	state := NewGameStateV2()

	for i := 1; i <= 10; i++ {
		state.IncrementBeat()
		if got := state.GetCurrentBeat(); got != i {
			t.Errorf("After %d increments, got beat=%d", i, got)
		}
	}
}

// Test Task 3: 序列化支持
func TestGameStateV2_JSONSerialization(t *testing.T) {
	state := NewGameStateV2()
	state.SetHP(75)
	state.SetSAN(50)
	state.IncrementBeat()
	state.IncrementBeat()

	// Marshal to JSON
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal to new state
	var restored GameStateV2
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 驗證數據完整性
	if restored.GameID != state.GameID {
		t.Errorf("GameID mismatch: got %s, want %s", restored.GameID, state.GameID)
	}

	if restored.currentBeat != state.GetCurrentBeat() {
		t.Errorf("CurrentBeat mismatch: got %d, want %d", restored.currentBeat, state.GetCurrentBeat())
	}

	if restored.hp != state.GetHP() {
		t.Errorf("HP mismatch: got %d, want %d", restored.hp, state.GetHP())
	}

	if restored.san != state.GetSAN() {
		t.Errorf("SAN mismatch: got %d, want %d", restored.san, state.GetSAN())
	}
}

func TestGameStateV2_JSONDoesNotIncludeMutex(t *testing.T) {
	state := NewGameStateV2()

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 檢查 JSON 中不包含 mu 字段
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	if _, exists := raw["mu"]; exists {
		t.Error("JSON should not include 'mu' field")
	}
}

// Test Task 4: GlobalSeeds 和 LocalSeeds 操作
func TestGameStateV2_AddGlobalSeed(t *testing.T) {
	state := NewGameStateV2()

	testSeed, err := seed.NewGlobalSeed(
		"GS001",
		"測試伏筆",
		"測試真相",
		"tragic",
		[]seed.ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 13, BeatEnd: 18},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create test seed: %v", err)
	}

	state.AddGlobalSeed(testSeed)

	seeds := state.GetGlobalSeeds()
	if len(seeds) != 1 {
		t.Fatalf("Expected 1 global seed, got %d", len(seeds))
	}

	if seeds[0].ID != "GS001" {
		t.Errorf("Expected seed ID 'GS001', got '%s'", seeds[0].ID)
	}
}

func TestGameStateV2_AddLocalSeed(t *testing.T) {
	state := NewGameStateV2()

	seed := &LocalSeed{
		ID:      "LS001",
		Content: "場景伏筆",
	}

	state.AddLocalSeed(seed)

	seeds := state.GetLocalSeeds()
	if len(seeds) != 1 {
		t.Fatalf("Expected 1 local seed, got %d", len(seeds))
	}

	if seeds[0].ID != "LS001" {
		t.Errorf("Expected seed ID 'LS001', got '%s'", seeds[0].ID)
	}
}

// Test boundary conditions
func TestGameStateV2_BoundaryConditions(t *testing.T) {
	state := NewGameStateV2()

	// HP 邊界測試
	state.SetHP(-10)
	if hp := state.GetHP(); hp != -10 {
		t.Errorf("HP should allow negative values for death state, got %d", hp)
	}

	state.SetHP(0)
	if hp := state.GetHP(); hp != 0 {
		t.Errorf("HP should allow zero, got %d", hp)
	}

	// SAN 邊界測試
	state.SetSAN(-50)
	if san := state.GetSAN(); san != -50 {
		t.Errorf("SAN should allow negative values, got %d", san)
	}

	state.SetSAN(150)
	if san := state.GetSAN(); san != 150 {
		t.Errorf("SAN should allow values > 100, got %d", san)
	}
}

// Test deep copy protection (Issue #2)
func TestGameStateV2_DeepCopyProtection(t *testing.T) {
	gs := NewGameStateV2()

	// Test GlobalSeed deep copy
	testSeed, err := seed.NewGlobalSeed(
		"GS001",
		"Original content",
		"Original truth",
		"tragic",
		[]seed.ClueTier{
			{Tier: 1, Content: "Clue 1", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Clue 3", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create test seed: %v", err)
	}
	gs.AddGlobalSeed(testSeed)

	// 獲取副本並嘗試修改
	seeds := gs.GetGlobalSeeds()
	seeds[0].Content = "Modified content"
	if seeds[0].CurrentTier < 3 {
		_ = seeds[0].AdvanceTier()
	}

	// 驗證原始數據未被修改
	original := gs.GetGlobalSeeds()
	if original[0].Content != "Original content" {
		t.Errorf("Deep copy failed: GlobalSeed content was modified")
	}
	if original[0].CurrentTier != 1 {
		t.Errorf("Deep copy failed: GlobalSeed CurrentTier was modified")
	}

	// Test LocalSeed deep copy
	gs.AddLocalSeed(&LocalSeed{
		ID:      "LS001",
		Content: "Original local",
		Urgency: 50,
	})

	localSeeds := gs.GetLocalSeeds()
	localSeeds[0].Content = "Modified local"
	localSeeds[0].IsHarvested = true

	// 驗證原始數據未被修改
	originalLocal := gs.GetLocalSeeds()
	if originalLocal[0].Content != "Original local" {
		t.Errorf("Deep copy failed: LocalSeed content was modified")
	}
	if originalLocal[0].IsHarvested {
		t.Errorf("Deep copy failed: LocalSeed IsHarvested was modified")
	}
}

// Test UnmarshalJSON error cases (Issue #3)
func TestGameStateV2_UnmarshalJSON_ErrorCases(t *testing.T) {
	gs := NewGameStateV2()

	// Test empty data
	err := gs.UnmarshalJSON(nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}

	// Test invalid JSON
	err = gs.UnmarshalJSON([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// Test concurrent access (Issue #4)
func TestGameStateV2_ConcurrentAccess(t *testing.T) {
	gs := NewGameStateV2()

	var wg sync.WaitGroup
	iterations := 100

	// 並發讀寫 HP
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				gs.SetHP(j)
				gs.GetHP()
			}
		}()
	}

	// 並發讀寫 SAN
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				gs.SetSAN(j)
				gs.GetSAN()
			}
		}()
	}

	// 並發讀寫 Seeds
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				testSeed, _ := seed.NewGlobalSeed(
					fmt.Sprintf("GS%03d-%03d", id, j),
					"content",
					"truth",
					"tragic",
					[]seed.ClueTier{
						{Tier: 1, Content: "C1", Keywords: []string{}, BeatStart: 1, BeatEnd: 5},
						{Tier: 2, Content: "C2", Keywords: []string{}, BeatStart: 6, BeatEnd: 12},
						{Tier: 3, Content: "C3", Keywords: []string{}, BeatStart: 13, BeatEnd: 18},
					},
				)
				gs.AddGlobalSeed(testSeed)
				gs.GetGlobalSeeds()
			}
		}(i)
	}

	wg.Wait()
}

// Test extreme boundary values (Issue #4)
func TestGameStateV2_ExtremeBoundaryValues(t *testing.T) {
	gs := NewGameStateV2()

	// Note: GameStateV2 itself does NOT enforce clamping
	// This is StateManager's responsibility (to be implemented in Epic 5)
	// These tests verify that GameStateV2 accepts any value

	// HP extreme values
	gs.SetHP(-1000)
	if gs.GetHP() != -1000 {
		t.Errorf("Expected HP=-1000, got %d", gs.GetHP())
	}

	gs.SetHP(1000)
	if gs.GetHP() != 1000 {
		t.Errorf("Expected HP=1000, got %d", gs.GetHP())
	}

	// SAN extreme values
	gs.SetSAN(-1000)
	if gs.GetSAN() != -1000 {
		t.Errorf("Expected SAN=-1000, got %d", gs.GetSAN())
	}

	gs.SetSAN(1000)
	if gs.GetSAN() != 1000 {
		t.Errorf("Expected SAN=1000, got %d", gs.GetSAN())
	}
}

// ==============================================================================
// Story 2.6: GameStateV2 Extension Tests
// ==============================================================================

// TestNewGameStateV2_Story26_FieldInitialization verifies that all new fields
// added in Story 2.6 are properly initialized in the constructor.
// AC1: MomentumConfig field initialization
// AC2: NPCManager field initialization
// AC3: GlobalFacts field initialization
// AC4: ChatSessions field initialization
func TestNewGameStateV2_Story26_FieldInitialization(t *testing.T) {
	state := NewGameStateV2()

	// AC1: Verify MomentumConfig is initialized with default values
	if state.MomentumConfig == nil {
		t.Fatal("MomentumConfig should be initialized")
	}
	if state.MomentumConfig.Frequency != "medium" {
		t.Errorf("Expected MomentumConfig.Frequency='medium', got %s", state.MomentumConfig.Frequency)
	}
	if !state.MomentumConfig.AutoResolve {
		t.Error("Expected MomentumConfig.AutoResolve=true")
	}
	if state.MomentumConfig.MaxAutoBeats != 5 {
		t.Errorf("Expected MomentumConfig.MaxAutoBeats=5, got %d", state.MomentumConfig.MaxAutoBeats)
	}
	if state.MomentumConfig.PauseOnRisk != "medium" {
		t.Errorf("Expected MomentumConfig.PauseOnRisk='medium', got %s", state.MomentumConfig.PauseOnRisk)
	}
	if !state.MomentumConfig.PauseOnPlot {
		t.Error("Expected MomentumConfig.PauseOnPlot=true")
	}
	if !state.MomentumConfig.PauseOnNPC {
		t.Error("Expected MomentumConfig.PauseOnNPC=true")
	}
	if !state.MomentumConfig.PauseOnEvent {
		t.Error("Expected MomentumConfig.PauseOnEvent=true")
	}

	// AC2: Verify NPCManager is initialized with empty maps
	if state.NPCManager == nil {
		t.Fatal("NPCManager should be initialized")
	}
	if state.NPCManager.Profiles == nil {
		t.Error("NPCManager.Profiles should be initialized")
	}
	if len(state.NPCManager.Profiles) != 0 {
		t.Errorf("Expected NPCManager.Profiles to be empty, got %d items", len(state.NPCManager.Profiles))
	}
	if state.NPCManager.States == nil {
		t.Error("NPCManager.States should be initialized")
	}
	if len(state.NPCManager.States) != 0 {
		t.Errorf("Expected NPCManager.States to be empty, got %d items", len(state.NPCManager.States))
	}

	// AC3: Verify GlobalFacts is initialized as empty slice
	if state.GlobalFacts == nil {
		t.Fatal("GlobalFacts should be initialized")
	}
	if len(state.GlobalFacts) != 0 {
		t.Errorf("Expected GlobalFacts to be empty, got %d items", len(state.GlobalFacts))
	}

	// AC4: Verify ChatSessions is initialized as empty slice
	if state.ChatSessions == nil {
		t.Fatal("ChatSessions should be initialized")
	}
	if len(state.ChatSessions) != 0 {
		t.Errorf("Expected ChatSessions to be empty, got %d items", len(state.ChatSessions))
	}
}

// TestGameStateV2_Story26_JSONSerialization verifies that all new fields
// are correctly serialized to JSON.
// AC5: JSON serialization correctness
func TestGameStateV2_Story26_JSONSerialization(t *testing.T) {
	state := NewGameStateV2()

	// Add some test data to the new fields
	state.MomentumConfig.Frequency = "high"
	state.MomentumConfig.MaxAutoBeats = 10

	state.NPCManager.Profiles["npc1"] = map[string]interface{}{
		"id":   "npc1",
		"name": "Test NPC",
	}
	state.NPCManager.States["npc1"] = map[string]interface{}{
		"is_alive": true,
	}

	state.GlobalFacts = append(state.GlobalFacts, &Fact{
		ID:        "fact1",
		Content:   "Test fact content",
		Type:      "event",
		Source:    "player",
		CreatedAt: 1,
		Location:  "room1",
		Witnesses: []string{"player", "npc1"},
	})

	state.ChatSessions = append(state.ChatSessions, &ChatSession{
		ID:           "chat1",
		StartBeat:    1,
		EndBeat:      5,
		Participants: []string{"player", "npc1"},
		Messages:     []interface{}{},
		Summary:      "Test chat summary",
	})

	// Serialize to JSON
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	// Verify the JSON contains the expected fields
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		t.Fatalf("JSON parsing failed: %v", err)
	}

	// Check MomentumConfig
	momentumConfig, ok := jsonMap["momentum_config"].(map[string]interface{})
	if !ok {
		t.Fatal("momentum_config field not found in JSON")
	}
	if momentumConfig["frequency"] != "high" {
		t.Errorf("Expected frequency='high', got %v", momentumConfig["frequency"])
	}
	if momentumConfig["max_auto_beats"] != float64(10) {
		t.Errorf("Expected max_auto_beats=10, got %v", momentumConfig["max_auto_beats"])
	}

	// Check NPCManager
	npcManager, ok := jsonMap["npc_manager"].(map[string]interface{})
	if !ok {
		t.Fatal("npc_manager field not found in JSON")
	}
	profiles, ok := npcManager["profiles"].(map[string]interface{})
	if !ok {
		t.Fatal("npc_manager.profiles not found")
	}
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	// Check GlobalFacts
	globalFacts, ok := jsonMap["global_facts"].([]interface{})
	if !ok {
		t.Fatal("global_facts field not found in JSON")
	}
	if len(globalFacts) != 1 {
		t.Errorf("Expected 1 fact, got %d", len(globalFacts))
	}
	fact := globalFacts[0].(map[string]interface{})
	if fact["id"] != "fact1" {
		t.Errorf("Expected fact id='fact1', got %v", fact["id"])
	}
	if fact["content"] != "Test fact content" {
		t.Errorf("Expected fact content='Test fact content', got %v", fact["content"])
	}

	// Check ChatSessions
	chatSessions, ok := jsonMap["chat_sessions"].([]interface{})
	if !ok {
		t.Fatal("chat_sessions field not found in JSON")
	}
	if len(chatSessions) != 1 {
		t.Errorf("Expected 1 chat session, got %d", len(chatSessions))
	}
	session := chatSessions[0].(map[string]interface{})
	if session["id"] != "chat1" {
		t.Errorf("Expected session id='chat1', got %v", session["id"])
	}
	if session["summary"] != "Test chat summary" {
		t.Errorf("Expected summary='Test chat summary', got %v", session["summary"])
	}
}

// TestGameStateV2_Story26_JSONDeserialization verifies that all new fields
// are correctly deserialized from JSON and nil fields are properly initialized.
// AC5: JSON deserialization correctness
func TestGameStateV2_Story26_JSONDeserialization(t *testing.T) {
	// Create JSON data with all new fields
	jsonData := `{
		"game_id": "test-game-123",
		"current_beat": 5,
		"hp": 80,
		"san": 70,
		"inventory": [],
		"global_seeds": [],
		"local_seeds": [],
		"tension": {
			"base_value": 0,
			"accumulated": 0,
			"decay_rate": 5,
			"modifiers": []
		},
		"context": {
			"summary": ""
		},
		"current_scene": "",
		"active_rules": [],
		"npc_states": {},
		"rule_warnings": {},
		"used_templates": {
			"rules": [],
			"scenes": []
		},
		"momentum_config": {
			"frequency": "high",
			"auto_resolve": false,
			"max_auto_beats": 7,
			"pause_on_risk": "high",
			"pause_on_plot": false,
			"pause_on_npc": false,
			"pause_on_event": false
		},
		"npc_manager": {
			"profiles": {
				"npc1": {
					"id": "npc1",
					"name": "Test NPC"
				}
			},
			"states": {
				"npc1": {
					"is_alive": true
				}
			},
			"config": null
		},
		"global_facts": [
			{
				"id": "fact1",
				"content": "Test fact",
				"type": "event",
				"source": "player",
				"created_at": 3,
				"location": "hallway",
				"witnesses": ["player"]
			}
		],
		"chat_sessions": [
			{
				"id": "session1",
				"start_beat": 2,
				"end_beat": 4,
				"participants": ["player", "npc1"],
				"messages": [{}, {}, {}, {}, {}],
				"summary": "Brief conversation"
			}
		]
	}`

	state := &GameStateV2{}
	err := json.Unmarshal([]byte(jsonData), state)
	if err != nil {
		t.Fatalf("JSON deserialization failed: %v", err)
	}

	// Verify MomentumConfig was deserialized correctly
	if state.MomentumConfig == nil {
		t.Fatal("MomentumConfig should be deserialized")
	}
	if state.MomentumConfig.Frequency != "high" {
		t.Errorf("Expected frequency='high', got %s", state.MomentumConfig.Frequency)
	}
	if state.MomentumConfig.AutoResolve {
		t.Error("Expected AutoResolve=false")
	}
	if state.MomentumConfig.MaxAutoBeats != 7 {
		t.Errorf("Expected MaxAutoBeats=7, got %d", state.MomentumConfig.MaxAutoBeats)
	}

	// Verify NPCManager was deserialized correctly
	if state.NPCManager == nil {
		t.Fatal("NPCManager should be deserialized")
	}
	if len(state.NPCManager.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(state.NPCManager.Profiles))
	}

	// Verify GlobalFacts was deserialized correctly
	if state.GlobalFacts == nil {
		t.Fatal("GlobalFacts should be deserialized")
	}
	if len(state.GlobalFacts) != 1 {
		t.Errorf("Expected 1 fact, got %d", len(state.GlobalFacts))
	}
	if state.GlobalFacts[0].ID != "fact1" {
		t.Errorf("Expected fact ID='fact1', got %s", state.GlobalFacts[0].ID)
	}
	if state.GlobalFacts[0].Content != "Test fact" {
		t.Errorf("Expected content='Test fact', got %s", state.GlobalFacts[0].Content)
	}

	// Verify ChatSessions was deserialized correctly
	if state.ChatSessions == nil {
		t.Fatal("ChatSessions should be deserialized")
	}
	if len(state.ChatSessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(state.ChatSessions))
	}
	if state.ChatSessions[0].ID != "session1" {
		t.Errorf("Expected session ID='session1', got %s", state.ChatSessions[0].ID)
	}
	if len(state.ChatSessions[0].Messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(state.ChatSessions[0].Messages))
	}
}

// TestGameStateV2_Story26_NilFieldInitialization verifies that nil fields
// are properly initialized during deserialization.
// AC5: Nil field handling in UnmarshalJSON
func TestGameStateV2_Story26_NilFieldInitialization(t *testing.T) {
	// Create minimal JSON without the new fields
	jsonData := `{
		"game_id": "test-game-456",
		"current_beat": 1,
		"hp": 100,
		"san": 100,
		"inventory": []
	}`

	state := &GameStateV2{}
	err := json.Unmarshal([]byte(jsonData), state)
	if err != nil {
		t.Fatalf("JSON deserialization failed: %v", err)
	}

	// Verify all new fields are initialized even when missing from JSON
	if state.MomentumConfig == nil {
		t.Error("MomentumConfig should be initialized when nil")
	}
	if state.NPCManager == nil {
		t.Error("NPCManager should be initialized when nil")
	}
	if state.GlobalFacts == nil {
		t.Error("GlobalFacts should be initialized when nil")
	}
	if state.ChatSessions == nil {
		t.Error("ChatSessions should be initialized when nil")
	}

	// Verify default values are correct
	if state.MomentumConfig.Frequency != "medium" {
		t.Errorf("Expected default frequency='medium', got %s", state.MomentumConfig.Frequency)
	}
	if len(state.GlobalFacts) != 0 {
		t.Errorf("Expected empty GlobalFacts, got %d items", len(state.GlobalFacts))
	}
	if len(state.ChatSessions) != 0 {
		t.Errorf("Expected empty ChatSessions, got %d items", len(state.ChatSessions))
	}
}

// TestGameStateV2_Story26_RoundTripSerialization verifies that data
// survives a full serialize-deserialize cycle without loss.
// AC5: Round-trip JSON correctness
func TestGameStateV2_Story26_RoundTripSerialization(t *testing.T) {
	// Create a state with all new fields populated
	original := NewGameStateV2()
	original.MomentumConfig.Frequency = "low"
	original.MomentumConfig.MaxAutoBeats = 3

	original.GlobalFacts = append(original.GlobalFacts, &Fact{
		ID:        "fact1",
		Content:   "Round trip test",
		Type:      "discovery",
		Source:    "system",
		CreatedAt: 10,
		Location:  "library",
		Witnesses: []string{"player", "npc1", "npc2"},
	})

	original.ChatSessions = append(original.ChatSessions, &ChatSession{
		ID:           "session1",
		StartBeat:    5,
		EndBeat:      8,
		Participants: []string{"player", "npc1"},
		Messages:     make([]interface{}, 15), // 15 placeholder messages
		Summary:      "Important conversation",
	})

	// Serialize
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Deserialize
	restored := &GameStateV2{}
	err = json.Unmarshal(data, restored)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify all data matches
	if restored.MomentumConfig.Frequency != "low" {
		t.Errorf("MomentumConfig.Frequency lost in round trip")
	}
	if restored.MomentumConfig.MaxAutoBeats != 3 {
		t.Errorf("MomentumConfig.MaxAutoBeats lost in round trip")
	}

	if len(restored.GlobalFacts) != 1 {
		t.Fatalf("GlobalFacts count changed in round trip")
	}
	if restored.GlobalFacts[0].Content != "Round trip test" {
		t.Errorf("GlobalFacts content lost in round trip")
	}
	if len(restored.GlobalFacts[0].Witnesses) != 3 {
		t.Errorf("GlobalFacts witnesses lost in round trip")
	}

	if len(restored.ChatSessions) != 1 {
		t.Fatalf("ChatSessions count changed in round trip")
	}
	if restored.ChatSessions[0].Summary != "Important conversation" {
		t.Errorf("ChatSessions summary lost in round trip")
	}
	if len(restored.ChatSessions[0].Messages) != 15 {
		t.Errorf("ChatSessions message count lost in round trip, got %d", len(restored.ChatSessions[0].Messages))
	}
}

// ==========================================================================
// Story 5.6: ChatSession Query Methods Tests
// ==========================================================================

// TestGetChatHistory tests retrieving recent chat sessions.
// Story 5.6 AC4: GetChatHistory(limit int) returns most recent N sessions.
func TestGetChatHistory(t *testing.T) {
	state := NewGameStateV2()

	// Test empty state
	result := state.GetChatHistory(5)
	if result == nil {
		t.Fatal("GetChatHistory should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Errorf("GetChatHistory from empty state should return empty slice, got %d items", len(result))
	}

	// Add test sessions
	state.ChatSessions = []*ChatSession{
		{ID: "s1", StartBeat: 1, EndBeat: 2, Participants: []string{"player"}},
		{ID: "s2", StartBeat: 3, EndBeat: 4, Participants: []string{"player"}},
		{ID: "s3", StartBeat: 5, EndBeat: 6, Participants: []string{"player"}},
		{ID: "s4", StartBeat: 7, EndBeat: 8, Participants: []string{"player"}},
		{ID: "s5", StartBeat: 9, EndBeat: 10, Participants: []string{"player"}},
	}

	// Test limit less than total
	result = state.GetChatHistory(3)
	if len(result) != 3 {
		t.Fatalf("Expected 3 sessions, got %d", len(result))
	}
	// Should return last 3 (s3, s4, s5)
	if result[0].ID != "s3" || result[1].ID != "s4" || result[2].ID != "s5" {
		t.Errorf("GetChatHistory(3) returned wrong sessions: %v", []string{result[0].ID, result[1].ID, result[2].ID})
	}

	// Test limit equal to total
	result = state.GetChatHistory(5)
	if len(result) != 5 {
		t.Fatalf("Expected 5 sessions, got %d", len(result))
	}

	// Test limit greater than total
	result = state.GetChatHistory(10)
	if len(result) != 5 {
		t.Fatalf("Expected 5 sessions (all), got %d", len(result))
	}

	// Test limit <= 0 (should return all)
	result = state.GetChatHistory(0)
	if len(result) != 5 {
		t.Fatalf("Expected 5 sessions (all), got %d", len(result))
	}

	result = state.GetChatHistory(-1)
	if len(result) != 5 {
		t.Fatalf("Expected 5 sessions (all), got %d", len(result))
	}
}

// TestGetChatSessionByID tests retrieving a session by ID.
// Story 5.6 AC4: GetChatSessionByID(id string) returns session or nil.
func TestGetChatSessionByID(t *testing.T) {
	state := NewGameStateV2()

	// Test empty state
	result := state.GetChatSessionByID("nonexistent")
	if result != nil {
		t.Error("GetChatSessionByID on empty state should return nil")
	}

	// Add test sessions
	state.ChatSessions = []*ChatSession{
		{ID: "session_001", Participants: []string{"player", "npc_001"}},
		{ID: "session_002", Participants: []string{"player", "npc_002"}},
		{ID: "session_003", Participants: []string{"player"}},
	}

	// Test finding existing session
	result = state.GetChatSessionByID("session_002")
	if result == nil {
		t.Fatal("Should find session_002")
	}
	if result.ID != "session_002" {
		t.Errorf("Found wrong session: %s", result.ID)
	}

	// Test not finding nonexistent session
	result = state.GetChatSessionByID("nonexistent")
	if result != nil {
		t.Error("Should return nil for nonexistent session")
	}

	// Test empty ID
	result = state.GetChatSessionByID("")
	if result != nil {
		t.Error("Should return nil for empty ID")
	}
}

// TestGetChatSessionsByNPC tests retrieving sessions involving a specific NPC.
// Story 5.6 AC4: GetChatSessionsByNPC(npcID string) returns sessions with that NPC.
func TestGetChatSessionsByNPC(t *testing.T) {
	state := NewGameStateV2()

	// Test empty state
	result := state.GetChatSessionsByNPC("npc_001")
	if result == nil {
		t.Fatal("Should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Error("Should return empty slice for empty state")
	}

	// Add test sessions
	state.ChatSessions = []*ChatSession{
		{ID: "s1", Participants: []string{"player", "npc_001"}},
		{ID: "s2", Participants: []string{"player", "npc_002"}},
		{ID: "s3", Participants: []string{"player", "npc_001", "npc_002"}},
		{ID: "s4", Participants: []string{"player"}},
		{ID: "s5", Participants: []string{"player", "npc_001"}},
	}

	// Test finding sessions with npc_001
	result = state.GetChatSessionsByNPC("npc_001")
	if len(result) != 3 {
		t.Fatalf("Expected 3 sessions with npc_001, got %d", len(result))
	}
	expectedIDs := map[string]bool{"s1": true, "s3": true, "s5": true}
	for _, session := range result {
		if !expectedIDs[session.ID] {
			t.Errorf("Unexpected session in results: %s", session.ID)
		}
	}

	// Test finding sessions with npc_002
	result = state.GetChatSessionsByNPC("npc_002")
	if len(result) != 2 {
		t.Fatalf("Expected 2 sessions with npc_002, got %d", len(result))
	}

	// Test NPC not in any session
	result = state.GetChatSessionsByNPC("npc_999")
	if len(result) != 0 {
		t.Errorf("Expected 0 sessions for npc_999, got %d", len(result))
	}

	// Test empty NPC ID
	result = state.GetChatSessionsByNPC("")
	if len(result) != 0 {
		t.Errorf("Expected 0 sessions for empty NPC ID, got %d", len(result))
	}
}

// TestGetChatSessionsByBeatRange tests retrieving sessions within a beat range.
// Story 5.6 (Dev Notes): GetChatSessionsByBeatRange returns overlapping sessions.
func TestGetChatSessionsByBeatRange(t *testing.T) {
	state := NewGameStateV2()

	// Test empty state
	result := state.GetChatSessionsByBeatRange(1, 10)
	if result == nil {
		t.Fatal("Should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Error("Should return empty slice for empty state")
	}

	// Add test sessions
	state.ChatSessions = []*ChatSession{
		{ID: "s1", StartBeat: 1, EndBeat: 5, Participants: []string{"player"}},
		{ID: "s2", StartBeat: 10, EndBeat: 15, Participants: []string{"player"}},
		{ID: "s3", StartBeat: 20, EndBeat: 25, Participants: []string{"player"}},
		{ID: "s4", StartBeat: 30, EndBeat: 35, Participants: []string{"player"}},
		{ID: "s5", StartBeat: 12, EndBeat: 18, Participants: []string{"player"}}, // Overlaps with s2
	}

	// Test range that includes s1 and s2
	result = state.GetChatSessionsByBeatRange(1, 15)
	if len(result) != 3 {
		t.Fatalf("Expected 3 sessions in range [1-15], got %d", len(result))
	}
	// Should include s1 (1-5), s2 (10-15), s5 (12-18)
	expectedIDs := map[string]bool{"s1": true, "s2": true, "s5": true}
	for _, session := range result {
		if !expectedIDs[session.ID] {
			t.Errorf("Unexpected session in results: %s", session.ID)
		}
	}

	// Test range that includes only s3
	result = state.GetChatSessionsByBeatRange(20, 25)
	if len(result) != 1 {
		t.Fatalf("Expected 1 session in range [20-25], got %d", len(result))
	}
	if result[0].ID != "s3" {
		t.Errorf("Expected s3, got %s", result[0].ID)
	}

	// Test range with no sessions
	result = state.GetChatSessionsByBeatRange(40, 50)
	if len(result) != 0 {
		t.Errorf("Expected 0 sessions in range [40-50], got %d", len(result))
	}

	// Test range that partially overlaps s2
	result = state.GetChatSessionsByBeatRange(14, 16)
	if len(result) != 2 {
		t.Fatalf("Expected 2 sessions in range [14-16], got %d", len(result))
	}
	// Should include s2 (10-15) and s5 (12-18)
	expectedIDs = map[string]bool{"s2": true, "s5": true}
	for _, session := range result {
		if !expectedIDs[session.ID] {
			t.Errorf("Unexpected session in results: %s", session.ID)
		}
	}

	// Test entire range
	result = state.GetChatSessionsByBeatRange(0, 100)
	if len(result) != 5 {
		t.Fatalf("Expected all 5 sessions in range [0-100], got %d", len(result))
	}
}

// TestGetChatHistory_ThreadSafety tests thread-safe access to chat history.
func TestGetChatHistory_ThreadSafety(t *testing.T) {
	state := NewGameStateV2()

	// Add initial sessions
	for i := 0; i < 10; i++ {
		state.ChatSessions = append(state.ChatSessions, &ChatSession{
			ID:           fmt.Sprintf("s%d", i),
			StartBeat:    i * 10,
			EndBeat:      i*10 + 5,
			Participants: []string{"player"},
		})
	}

	// Concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := state.GetChatHistory(5)
			if len(result) != 5 {
				t.Errorf("Expected 5 sessions, got %d", len(result))
			}
		}()
	}

	wg.Wait()
}
