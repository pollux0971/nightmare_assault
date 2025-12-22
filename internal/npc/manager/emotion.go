package manager

import "encoding/json"

// EmotionState represents a three-dimensional emotional state for NPCs.
// Each dimension ranges from 0-100.
type EmotionState struct {
	Trust  int `json:"trust"`  // 信任度 (0-100)
	Fear   int `json:"fear"`   // 恐懼度 (0-100)
	Stress int `json:"stress"` // 壓力值 (0-100)
}

// EmotionDelta represents a change in emotional state.
type EmotionDelta struct {
	Trust  int `json:"trust"`
	Fear   int `json:"fear"`
	Stress int `json:"stress"`
}

// EmotionDeltas contains predefined emotion deltas for common interactions.
// Based on the proposal document, these represent standard emotional impacts
// that different player actions have on NPCs.
var EmotionDeltas = map[string]EmotionDelta{
	// friendly_chat: 友善對話，增加信任，減少恐懼與壓力
	"friendly_chat": {Trust: 5, Fear: -3, Stress: -5},

	// share_info: 分享資訊，增加信任，輕微減少壓力
	"share_info": {Trust: 8, Fear: 0, Stress: -2},

	// threat: 威脅行為，嚴重降低信任，大幅增加恐懼與壓力
	"threat": {Trust: -20, Fear: 25, Stress: 15},

	// help_npc: 幫助 NPC，大幅增加信任，減少恐懼與壓力
	"help_npc": {Trust: 15, Fear: -5, Stress: -10},

	// ignore_distress: 忽略 NPC 求助，降低信任，增加壓力
	"ignore_distress": {Trust: -10, Fear: 0, Stress: 5},

	// reveal_secret: 揭露秘密，建立深度信任，減少恐懼與壓力
	"reveal_secret": {Trust: 20, Fear: -10, Stress: -5},

	// catch_lying: 被揭穿謊言，嚴重後果
	"catch_lying": {Trust: -30, Fear: 10, Stress: 20},

	// witness_death: 目睹死亡，創傷性事件
	"witness_death": {Trust: 0, Fear: 30, Stress: 40},

	// combat_together: 並肩作戰，建立信任但增加壓力
	"combat_together": {Trust: 10, Fear: -5, Stress: 10},

	// abandon_npc: 拋棄 NPC，毀滅性影響
	"abandon_npc": {Trust: -40, Fear: 20, Stress: 30},

	// calm_down: 安撫 NPC，有效降低壓力與恐懼
	"calm_down": {Trust: 5, Fear: -10, Stress: -15},

	// aggressive_talk: 咄咄逼人對話，引發防禦反應
	"aggressive_talk": {Trust: -15, Fear: 15, Stress: 10},

	// hallucination: 幻覺相關事件，增加恐懼與壓力
	"hallucination": {Trust: -10, Fear: 20, Stress: 25},
}

// clamp ensures a value stays within the 0-100 range.
func clamp(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

// Apply applies an EmotionDelta to the current EmotionState and returns a new state.
// This method is immutable - it returns a new EmotionState rather than modifying the original.
func (e EmotionState) Apply(delta EmotionDelta) EmotionState {
	return EmotionState{
		Trust:  clamp(e.Trust + delta.Trust),
		Fear:   clamp(e.Fear + delta.Fear),
		Stress: clamp(e.Stress + delta.Stress),
	}
}

// Copy creates a deep copy of the EmotionState.
func (e EmotionState) Copy() EmotionState {
	return EmotionState{
		Trust:  e.Trust,
		Fear:   e.Fear,
		Stress: e.Stress,
	}
}

// String returns a string representation of the EmotionState.
func (e EmotionState) String() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// NewEmotionState creates a new EmotionState with the given values, clamped to 0-100.
func NewEmotionState(trust, fear, stress int) EmotionState {
	return EmotionState{
		Trust:  clamp(trust),
		Fear:   clamp(fear),
		Stress: clamp(stress),
	}
}

// DefaultEmotionState returns a neutral emotional state (50/25/25).
func DefaultEmotionState() EmotionState {
	return EmotionState{
		Trust:  50, // 中性信任
		Fear:   25, // 低恐懼
		Stress: 25, // 低壓力
	}
}
