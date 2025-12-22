package engine

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// GlobalSeed is now fully implemented in internal/engine/seed package (Epic 2, Story 2.1).
// Use seed.GlobalSeed instead of the old placeholder type.

// LocalSeed represents a scene-specific foreshadowing element.
// This is a placeholder type - full implementation in Epic 2, Story 2.3.
type LocalSeed struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	PlantedBeat int    `json:"planted_beat"`
	Urgency     int    `json:"urgency"`
	IsHarvested bool   `json:"is_harvested"`
	// TODO: Add more fields in Epic 2, Story 2.3 (SceneID, PlantedAt, etc.)
}

// DeepCopy creates a deep copy of the LocalSeed.
func (s *LocalSeed) DeepCopy() *LocalSeed {
	if s == nil {
		return nil
	}
	return &LocalSeed{
		ID:          s.ID,
		Content:     s.Content,
		PlantedBeat: s.PlantedBeat,
		Urgency:     s.Urgency,
		IsHarvested: s.IsHarvested,
	}
}

// TensionState is now fully implemented in tension.go (Epic 3)

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

// Story 2.6: Placeholder types for Epic 2 (Knowledge System) and Epic 6 (Momentum System)

// MomentumConfig represents the configuration for narrative momentum system.
// This is a placeholder type - full implementation in Epic 6, Story 6.2.
type MomentumConfig struct {
	Frequency      string `json:"frequency"`       // FrequencyLevel: "high"/"medium"/"low"
	AutoResolve    bool   `json:"auto_resolve"`    // Enable auto-resolve for low-risk actions
	MaxAutoBeats   int    `json:"max_auto_beats"`  // Maximum consecutive auto-resolved beats
	PauseOnRisk    string `json:"pause_on_risk"`   // RiskLevel: "none"/"low"/"medium"/"high"/"lethal"
	PauseOnPlot    bool   `json:"pause_on_plot"`   // Pause at plot points
	PauseOnNPC     bool   `json:"pause_on_npc"`    // Pause when NPC initiates conversation
	PauseOnEvent   bool   `json:"pause_on_event"`  // Pause on major events
	// TODO: Add more fields in Epic 6, Story 6.2
}

// Fact represents a piece of information in the knowledge system.
// This is a placeholder type - full implementation in Epic 2, Story 2.2.
type Fact struct {
	ID        string   `json:"id"`         // Unique fact ID
	Content   string   `json:"content"`    // The actual information
	Type      string   `json:"type"`       // FactType: "event"/"dialogue"/"discovery"/"rumor"/"secret"
	Source    string   `json:"source"`     // Who/what created this fact
	CreatedAt int      `json:"created_at"` // Beat number when fact was created
	Location  string   `json:"location"`   // Where the fact occurred
	Witnesses []string `json:"witnesses"`  // Entity IDs who witnessed this fact
	// TODO: Add more fields in Epic 2, Story 2.2 (Confidence, IsDistorted, etc.)
}

// ChatSession represents a complete chat conversation session.
// Story 5.6 AC1: ChatSession 包含完整對話記錄
// Note: This is a simplified version for storage. Full UI model is in views package.
type ChatSession struct {
	ID              string                 `json:"id"`                          // Unique session ID
	Participants    []string               `json:"participants"`                // Entity IDs of all participants
	Messages        []interface{}          `json:"messages"`                    // Message list (ChatMessage from views)
	StartBeat       int                    `json:"start_beat"`                  // Beat when chat started
	EndBeat         int                    `json:"end_beat"`                    // Beat when chat ended
	Summary         interface{}            `json:"summary,omitempty"`           // LLM-generated summary (ChatSummary from views)
	Interrupted     bool                   `json:"interrupted"`                 // Whether session was interrupted
	InterruptReason string                 `json:"interrupt_reason,omitempty"`  // Reason for interruption
	CreatedAt       string                 `json:"created_at"`                  // Creation timestamp (RFC3339 format)

	// Legacy/optional fields
	Location       string `json:"location,omitempty"`        // Where chat took place
	TurnsSpent     int    `json:"turns_spent,omitempty"`     // Chat turns consumed
}

// NPCManagerState represents the serializable state of the NPC management system.
// This is a placeholder type - full implementation in Epic 2, Story 2.7.
// It contains the essential state needed for save/load functionality.
type NPCManagerState struct {
	Profiles map[string]interface{} `json:"profiles"` // NPC profiles (will be *manager.NPCProfile)
	States   map[string]interface{} `json:"states"`   // NPC runtime states (will be *manager.NPCRuntimeState)
	Config   interface{}            `json:"config"`   // Manager configuration (will be *manager.NPCManagerConfig)
	// TODO: Full integration with internal/npc/manager in Epic 2, Story 2.7
}

// GameStateV2 is the unified game state for v2.0 architecture.
// It centralizes all game data including HP, SAN, seeds, tension, context, etc.
//
// HP/SAN Design:
//   - Range: 0-100 (inclusive, enforced by StateManager in Epic 5)
//   - Initial values: HP=100, SAN=100
//   - Clamping: Values should be clamped to [0, 100] by StateManager when applying changes
//   - Thread-safety: All modifications are protected by sync.RWMutex
//   - State changes should flow through StateManager for centralized validation
//   - GameStateV2 itself does NOT enforce clamping - this is StateManager's responsibility
type GameStateV2 struct {
	// 基礎欄位
	GameID      string   `json:"game_id"`
	CurrentBeat int      `json:"current_beat"`
	HP          int      `json:"hp"`
	SAN         int      `json:"san"`
	Inventory   []string `json:"inventory"`

	// v2.0 系統欄位
	GlobalSeeds []*seed.GlobalSeed `json:"global_seeds"`
	LocalSeeds  []*LocalSeed       `json:"local_seeds"`
	Tension     *TensionState      `json:"tension"`
	Context     *ContextWindow     `json:"context"`

	// 場景與規則欄位
	CurrentScene string               `json:"current_scene"`
	ActiveRules  []*ActiveRule        `json:"active_rules"`
	NPCStates    map[string]*NPCState `json:"npc_states"`

	// Rule Warnings tracking (for Rule Hints system)
	RuleWarnings map[string]int `json:"rule_warnings"` // ruleID -> warning count

	// 模板記錄
	UsedTemplates *UsedTemplates `json:"used_templates"`

	// Story 7.7: NPC Dialogue History for /log command support
	// Records all NPC dialogues with metadata for review
	DialogueHistory *game.DialogueHistory `json:"-"` // Serialized separately

	// Story 9: Dream System
	// Records all dreams experienced during gameplay
	DreamLog *game.DreamLog `json:"-"` // Serialized separately

	// Story 2.6: Epic 2 & Epic 6 Extensions
	// MomentumConfig controls the narrative momentum system behavior
	// Full implementation in Epic 6
	MomentumConfig *MomentumConfig `json:"momentum_config"`

	// NPCManager holds the serializable state of the NPC management system
	// Full implementation in Epic 2, Story 2.7
	NPCManager *NPCManagerState `json:"npc_manager"`

	// GlobalFacts stores all facts in the global knowledge base
	// Full implementation in Epic 2, Story 2.2
	GlobalFacts []*Fact `json:"global_facts"`

	// ChatSessions records all chat conversation sessions
	// Full implementation in Epic 5, Story 5.6
	ChatSessions []*ChatSession `json:"chat_sessions"`

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

		GlobalSeeds: make([]*seed.GlobalSeed, 0),
		LocalSeeds:  make([]*LocalSeed, 0),
		Tension:     NewTensionState(),
		Context: &ContextWindow{
			Summary: "",
		},

		CurrentScene: "",
		ActiveRules:  make([]*ActiveRule, 0),
		NPCStates:    make(map[string]*NPCState),

		RuleWarnings: make(map[string]int),

		UsedTemplates: &UsedTemplates{
			Rules:  make([]string, 0),
			Scenes: make([]string, 0),
		},

		// Story 7.7: Initialize dialogue history
		DialogueHistory: game.NewDialogueHistory(),

		// Story 9: Initialize dream log
		DreamLog: game.NewDreamLog(),

		// Story 2.6: Initialize Epic 2 & Epic 6 fields
		MomentumConfig: &MomentumConfig{
			Frequency:    "medium",
			AutoResolve:  true,
			MaxAutoBeats: 5,
			PauseOnRisk:  "medium",
			PauseOnPlot:  true,
			PauseOnNPC:   true,
			PauseOnEvent: true,
		},
		NPCManager: &NPCManagerState{
			Profiles: make(map[string]interface{}),
			States:   make(map[string]interface{}),
			Config:   nil,
		},
		GlobalFacts:  make([]*Fact, 0),
		ChatSessions: make([]*ChatSession, 0),

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
	oldHP := g.hp
	g.hp = hp
	g.HP = hp // 同步到公開欄位用於序列化

	// Story 10-8 AC1: Log state change
	logger.Debug("GameState HP changed", map[string]interface{}{
		"old_hp": oldHP,
		"new_hp": hp,
		"delta": hp - oldHP,
	})
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
	oldSAN := g.san
	g.san = san
	g.SAN = san // 同步到公開欄位用於序列化

	// Story 10-8 AC1: Log state change
	logger.Debug("GameState SAN changed", map[string]interface{}{
		"old_san": oldSAN,
		"new_san": san,
		"delta": san - oldSAN,
	})
}

// TakeDamage applies damage to HP or SAN and returns the new value.
// Story 7.4 AC2, AC3: HP 損失與 SAN 損失系統
// delta should be negative for damage, positive for healing
// Values are automatically clamped to [0, 100]
func (g *GameStateV2) TakeDamage(hp, san int, reason string) (newHP, newSAN int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	oldHP := g.hp
	oldSAN := g.san

	// Apply HP damage
	g.hp += hp
	if g.hp < 0 {
		g.hp = 0
	}
	if g.hp > 100 {
		g.hp = 100
	}
	g.HP = g.hp

	// Apply SAN damage
	g.san += san
	if g.san < 0 {
		g.san = 0
	}
	if g.san > 100 {
		g.san = 100
	}
	g.SAN = g.san

	// Story 10-8 AC1: Log damage
	logger.Debug("GameState damage applied", map[string]interface{}{
		"reason": reason,
		"hp_delta": hp,
		"san_delta": san,
		"old_hp": oldHP,
		"new_hp": g.hp,
		"old_san": oldSAN,
		"new_san": g.san,
	})

	return g.hp, g.san
}

// Heal restores HP or SAN and returns the new value.
// Story 7.4 AC2, AC5: HP 恢復與 SAN 恢復機制
// hp and san should be positive values to restore
// Values are automatically clamped to [0, 100]
func (g *GameStateV2) Heal(hp, san int, reason string) (newHP, newSAN int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Apply HP healing
	if hp > 0 {
		g.hp += hp
		if g.hp > 100 {
			g.hp = 100
		}
		g.HP = g.hp
	}

	// Apply SAN healing
	if san > 0 {
		g.san += san
		if g.san > 100 {
			g.san = 100
		}
		g.SAN = g.san
	}

	return g.hp, g.san
}

// IsDead returns true if HP is 0 or less.
// Story 7.4 AC2: HP 降到 0 時觸發死亡流程
func (g *GameStateV2) IsDead() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hp <= 0
}

// IsInsane returns true if SAN is 0.
// Story 7.4 AC4: SAN 0 (瘋狂 Insane) - 遊戲結束
func (g *GameStateV2) IsInsane() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.san <= 0
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
// Returns a deep copy to prevent external modification of internal state.
func (g *GameStateV2) GetGlobalSeeds() []*seed.GlobalSeed {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Deep copy to prevent external modification
	seeds := make([]*seed.GlobalSeed, len(g.GlobalSeeds))
	for i, s := range g.GlobalSeeds {
		seeds[i] = s.DeepCopy()
	}
	return seeds
}

// AddGlobalSeed adds a global seed (thread-safe).
func (g *GameStateV2) AddGlobalSeed(s *seed.GlobalSeed) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.GlobalSeeds = append(g.GlobalSeeds, s)
}

// GetLocalSeeds returns all local seeds (thread-safe).
// Returns a deep copy to prevent external modification of internal state.
func (g *GameStateV2) GetLocalSeeds() []*LocalSeed {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Deep copy to prevent external modification
	seeds := make([]*LocalSeed, len(g.LocalSeeds))
	for i, s := range g.LocalSeeds {
		seeds[i] = s.DeepCopy()
	}
	return seeds
}

// AddLocalSeed adds a local seed (thread-safe).
func (g *GameStateV2) AddLocalSeed(s *LocalSeed) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.LocalSeeds = append(g.LocalSeeds, s)
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
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate input
	if len(data) == 0 {
		return fmt.Errorf("empty JSON data")
	}

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
		g.GlobalSeeds = make([]*seed.GlobalSeed, 0)
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

	// Story 7.7: Initialize DialogueHistory if nil
	if g.DialogueHistory == nil {
		g.DialogueHistory = game.NewDialogueHistory()
	}

	// Story 9: Initialize DreamLog if nil
	if g.DreamLog == nil {
		g.DreamLog = game.NewDreamLog()
	}

	// Story 2.6: Initialize Epic 2 & Epic 6 fields if nil
	if g.MomentumConfig == nil {
		g.MomentumConfig = &MomentumConfig{
			Frequency:    "medium",
			AutoResolve:  true,
			MaxAutoBeats: 5,
			PauseOnRisk:  "medium",
			PauseOnPlot:  true,
			PauseOnNPC:   true,
			PauseOnEvent: true,
		}
	}
	if g.NPCManager == nil {
		g.NPCManager = &NPCManagerState{
			Profiles: make(map[string]interface{}),
			States:   make(map[string]interface{}),
			Config:   nil,
		}
	}
	if g.GlobalFacts == nil {
		g.GlobalFacts = make([]*Fact, 0)
	}
	if g.ChatSessions == nil {
		g.ChatSessions = make([]*ChatSession, 0)
	}

	return nil
}

// ==========================================================================
// Story 7.7: Dialogue History Integration
// ==========================================================================

// GetDialogueHistory returns the dialogue history (thread-safe).
// Used by /log command to display NPC dialogue history.
func (g *GameStateV2) GetDialogueHistory() *game.DialogueHistory {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.DialogueHistory
}

// RecordNPCDialogue records an NPC dialogue to history.
// Story 7.7 AC #6: Save NPC name, dialogue content, and current beat.
//
// Parameters:
//   - npcName: Name of the NPC speaking
//   - npcID: Unique ID of the NPC
//   - dialogue: The dialogue content
//   - scene: Current scene where dialogue occurred
//   - isQuestion: Whether this was a response to player question
//   - clueRevealed: Whether a clue was revealed in this dialogue
//   - seedID: Global seed ID if clue was revealed (empty string if none)
//   - isDeathDialogue: Whether this is death dialogue
func (g *GameStateV2) RecordNPCDialogue(
	npcName, npcID, dialogue, scene string,
	isQuestion, clueRevealed bool,
	seedID string,
	isDeathDialogue bool,
) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.DialogueHistory == nil {
		g.DialogueHistory = game.NewDialogueHistory()
	}

	// Get tension value safely
	tensionValue := 0
	if g.Tension != nil {
		tensionValue = g.Tension.GetValue()
	}

	record := game.DialogueRecord{
		NPCName:         npcName,
		NPCID:           npcID,
		Dialogue:        dialogue,
		BeatNumber:      g.currentBeat,
		Scene:           scene,
		Tension:         tensionValue,
		SAN:             g.san,
		IsQuestion:      isQuestion,
		ClueRevealed:    clueRevealed,
		SeedID:          seedID,
		IsDeathDialogue: isDeathDialogue,
	}

	g.DialogueHistory.RecordDialogue(record)
}

// GetDialogueHistoryForDisplay returns formatted dialogue history for /log command.
// Story 7.7 AC #6: Support history review via /log command.
func (g *GameStateV2) GetDialogueHistoryForDisplay(maxRecords int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.DialogueHistory == nil {
		return "尚未記錄任何 NPC 對話。"
	}

	return g.DialogueHistory.FormatForDisplay(maxRecords)
}

// GetDialogueHistoryForDisplayPaged returns formatted dialogue history with pagination.
// Story 10-4 AC: 支援翻頁查看更早記錄
// page 1 = most recent, page 2 = older records, etc.
func (g *GameStateV2) GetDialogueHistoryForDisplayPaged(recordsPerPage, pageNumber int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.DialogueHistory == nil {
		return "尚未記錄任何 NPC 對話。"
	}

	return g.DialogueHistory.FormatForDisplayPaged(recordsPerPage, pageNumber)
}

// ==========================================================================
// Story 9: Dream System Integration
// ==========================================================================

// GetDreamLog returns the dream log (thread-safe).
// Used by /dreams command to display dream history.
func (g *GameStateV2) GetDreamLog() *game.DreamLog {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.DreamLog
}

// RecordDream records a dream to the dream log.
// Story 9 AC: Save dream content, related rules, and current beat.
func (g *GameStateV2) RecordDream(dream game.DreamRecord) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.DreamLog == nil {
		g.DreamLog = game.NewDreamLog()
	}

	g.DreamLog.LogDream(dream)
}

// ==========================================================================
// Story 5.6: ChatSession Storage and Query Methods
// ==========================================================================

// GetChatHistory returns the most recent N chat sessions.
// Story 5.6 AC4: 提供 GetChatHistory(limit int) 方法
// If limit <= 0 or limit > total sessions, returns all sessions.
func (g *GameStateV2) GetChatHistory(limit int) []*ChatSession {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.ChatSessions == nil {
		return []*ChatSession{}
	}

	totalSessions := len(g.ChatSessions)
	if limit <= 0 || limit > totalSessions {
		limit = totalSessions
	}

	// Return last N sessions (most recent)
	start := totalSessions - limit
	result := make([]*ChatSession, limit)
	copy(result, g.ChatSessions[start:])
	return result
}

// GetChatSessionByID returns a chat session by its ID.
// Story 5.6 AC4: 提供 GetChatSessionByID(id string) 方法
// Returns nil if not found.
func (g *GameStateV2) GetChatSessionByID(id string) *ChatSession {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.ChatSessions == nil {
		return nil
	}

	for _, session := range g.ChatSessions {
		if session != nil && session.ID == id {
			return session
		}
	}

	return nil
}

// GetChatSessionsByNPC returns all chat sessions involving a specific NPC.
// Story 5.6 AC4: 提供 GetChatSessionsByNPC(npcID string) 方法
// Returns empty slice if no sessions found.
func (g *GameStateV2) GetChatSessionsByNPC(npcID string) []*ChatSession {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.ChatSessions == nil {
		return []*ChatSession{}
	}

	result := []*ChatSession{}
	for _, session := range g.ChatSessions {
		if session == nil {
			continue
		}

		// Check if NPC is in participants list
		for _, participantID := range session.Participants {
			if participantID == npcID {
				result = append(result, session)
				break
			}
		}
	}

	return result
}

// GetChatSessionsByBeatRange returns chat sessions within a beat range.
// Story 5.6 (from Dev Notes): 提供按時間範圍查詢
// Returns sessions where EndBeat >= startBeat AND StartBeat <= endBeat (overlapping).
func (g *GameStateV2) GetChatSessionsByBeatRange(startBeat, endBeat int) []*ChatSession {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.ChatSessions == nil {
		return []*ChatSession{}
	}

	result := []*ChatSession{}
	for _, session := range g.ChatSessions {
		if session == nil {
			continue
		}

		// Check if session overlaps with the requested range
		if session.EndBeat >= startBeat && session.StartBeat <= endBeat {
			result = append(result, session)
		}
	}

	return result
}
