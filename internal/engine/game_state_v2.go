package engine

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// GlobalSeed represents a main storyline foreshadowing element.
// This is a placeholder type - full implementation in Epic 2.
type GlobalSeed struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	// TODO: Add more fields in Epic 2 (LinkedTruth, LinkedEnding, etc.)
}

// LocalSeed represents a scene-specific foreshadowing element.
// This is a placeholder type - full implementation in Epic 2.
type LocalSeed struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	// TODO: Add more fields in Epic 2 (SceneID, PlantedAt, etc.)
}

// TensionState represents the tension/suspense state of the game.
// This is a placeholder type - full implementation in Epic 3.
type TensionState struct {
	Value int `json:"value"` // 0-100
	// TODO: Add more fields in Epic 3 (Level, History, etc.)
}

// ContextWindow represents the context management window.
// This is a placeholder type - full implementation in Epic 5.
type ContextWindow struct {
	Summary string `json:"summary"`
	// TODO: Add more fields in Epic 5 (RecentEntries, etc.)
}

// ActiveRule represents a currently active game rule.
// This is a placeholder type - full implementation in Epic 4.
type ActiveRule struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// TODO: Add more fields in Epic 4
}

// NPCState represents the state of an NPC teammate.
// This is a placeholder type - full implementation in Epic 6.
type NPCState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// TODO: Add more fields in Epic 6
}

// UsedTemplates tracks which templates have been used in generation.
// This is a placeholder type - full implementation in Epic 4.
type UsedTemplates struct {
	Rules  []string `json:"rules"`
	Scenes []string `json:"scenes"`
	// TODO: Add more fields in Epic 4
}

// GameStateV2 is the unified game state for v2.0 architecture.
// It centralizes all game data including HP, SAN, seeds, tension, context, etc.
type GameStateV2 struct {
	// 基礎欄位
	GameID      string   `json:"game_id"`
	CurrentBeat int      `json:"current_beat"`
	HP          int      `json:"hp"`
	SAN         int      `json:"san"`
	Inventory   []string `json:"inventory"`

	// v2.0 系統欄位
	GlobalSeeds []*GlobalSeed  `json:"global_seeds"`
	LocalSeeds  []*LocalSeed   `json:"local_seeds"`
	Tension     *TensionState  `json:"tension"`
	Context     *ContextWindow `json:"context"`

	// 場景與規則欄位
	CurrentScene string               `json:"current_scene"`
	ActiveRules  []*ActiveRule        `json:"active_rules"`
	NPCStates    map[string]*NPCState `json:"npc_states"`

	// 模板記錄
	UsedTemplates *UsedTemplates `json:"used_templates"`

	// 私有欄位用於線程安全
	mu          sync.RWMutex `json:"-"` // 不序列化
	currentBeat int          `json:"-"` // 內部計數器
	hp          int          `json:"-"` // 內部HP
	san         int          `json:"-"` // 內部SAN
}

// NewGameStateV2 creates a new v2.0 game state with default values.
func NewGameStateV2() *GameStateV2 {
	return &GameStateV2{
		GameID:      uuid.New().String(),
		CurrentBeat: 0,
		HP:          100,
		SAN:         100,
		Inventory:   make([]string, 0),

		GlobalSeeds: make([]*GlobalSeed, 0),
		LocalSeeds:  make([]*LocalSeed, 0),
		Tension: &TensionState{
			Value: 0,
		},
		Context: &ContextWindow{
			Summary: "",
		},

		CurrentScene: "",
		ActiveRules:  make([]*ActiveRule, 0),
		NPCStates:    make(map[string]*NPCState),

		UsedTemplates: &UsedTemplates{
			Rules:  make([]string, 0),
			Scenes: make([]string, 0),
		},

		currentBeat: 0,
		hp:          100,
		san:         100,
	}
}

// GetHP returns the current HP value (thread-safe).
func (g *GameStateV2) GetHP() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hp
}

// SetHP sets the HP value (thread-safe).
func (g *GameStateV2) SetHP(hp int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hp = hp
	g.HP = hp // 同步到公開欄位用於序列化
}

// GetSAN returns the current SAN value (thread-safe).
func (g *GameStateV2) GetSAN() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.san
}

// SetSAN sets the SAN value (thread-safe).
func (g *GameStateV2) SetSAN(san int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.san = san
	g.SAN = san // 同步到公開欄位用於序列化
}

// GetCurrentBeat returns the current beat number (thread-safe).
func (g *GameStateV2) GetCurrentBeat() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentBeat
}

// IncrementBeat increments the current beat counter (thread-safe).
func (g *GameStateV2) IncrementBeat() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.currentBeat++
	g.CurrentBeat = g.currentBeat // 同步到公開欄位用於序列化
}

// GetGlobalSeeds returns all global seeds (thread-safe).
func (g *GameStateV2) GetGlobalSeeds() []*GlobalSeed {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// 返回副本以避免外部修改
	seeds := make([]*GlobalSeed, len(g.GlobalSeeds))
	copy(seeds, g.GlobalSeeds)
	return seeds
}

// AddGlobalSeed adds a global seed (thread-safe).
func (g *GameStateV2) AddGlobalSeed(seed *GlobalSeed) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.GlobalSeeds = append(g.GlobalSeeds, seed)
}

// GetLocalSeeds returns all local seeds (thread-safe).
func (g *GameStateV2) GetLocalSeeds() []*LocalSeed {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// 返回副本以避免外部修改
	seeds := make([]*LocalSeed, len(g.LocalSeeds))
	copy(seeds, g.LocalSeeds)
	return seeds
}

// AddLocalSeed adds a local seed (thread-safe).
func (g *GameStateV2) AddLocalSeed(seed *LocalSeed) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.LocalSeeds = append(g.LocalSeeds, seed)
}

// MarshalJSON implements custom JSON marshaling.
// Syncs internal state to public fields before marshaling.
func (g *GameStateV2) MarshalJSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 同步內部狀態到公開欄位
	g.CurrentBeat = g.currentBeat
	g.HP = g.hp
	g.SAN = g.san

	// 創建臨時結構避免遞歸調用
	type Alias GameStateV2
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(g),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling.
// Syncs public fields to internal state after unmarshaling.
func (g *GameStateV2) UnmarshalJSON(data []byte) error {
	// 創建臨時結構避免遞歸調用
	type Alias GameStateV2
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(g),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 同步公開欄位到內部狀態
	g.currentBeat = g.CurrentBeat
	g.hp = g.HP
	g.san = g.SAN

	// 確保切片和映射已初始化
	if g.GlobalSeeds == nil {
		g.GlobalSeeds = make([]*GlobalSeed, 0)
	}
	if g.LocalSeeds == nil {
		g.LocalSeeds = make([]*LocalSeed, 0)
	}
	if g.ActiveRules == nil {
		g.ActiveRules = make([]*ActiveRule, 0)
	}
	if g.NPCStates == nil {
		g.NPCStates = make(map[string]*NPCState)
	}
	if g.Inventory == nil {
		g.Inventory = make([]string, 0)
	}

	return nil
}
