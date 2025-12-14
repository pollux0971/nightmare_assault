package engine

import (
	"encoding/json"
	"sync"
	"testing"
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

	seed := &GlobalSeed{
		ID:      "GS001",
		Content: "測試伏筆",
	}

	state.AddGlobalSeed(seed)

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
