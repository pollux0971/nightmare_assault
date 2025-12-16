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
