// Package momentum 提供敘事動量控制系統
// Epic 6: Story 6.1 - MomentumController 基礎結構
package momentum

import (
	"sync"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// NarrationAgent 定義敘事代理介面
// 這是一個最小介面，避免循環依賴
// 實際實作在 internal/orchestrator/agents/narration
type NarrationAgent interface {
	// 這裡不需要定義具體方法，因為 Story 6.1 只建立骨架
	// 實際方法將在 Story 6.5 需要時定義
}

// Note: RiskEvaluator is fully implemented in risk.go (Story 6.3 completed).
// See risk.go for the complete implementation with all methods.

// Note: PlotDetector is fully implemented in plot.go (Story 6.4 completed).
// See plot.go for the complete implementation with all methods.

// TensionManager interface for Guardian integration (Story 9-11)
// This minimal interface avoids circular dependency with internal/guardian
type TensionManager interface {
	// SyncFromMomentum syncs tension state from MomentumController
	// This method should NOT call back to MomentumController to avoid circular calls
	SyncFromMomentum(tensionValue float64, phase string)
}

// MomentumController 控制敘事動量，決定何時暫停詢問玩家
type MomentumController struct {
	config         *MomentumConfig
	riskEvaluator  *RiskEvaluator
	plotDetector   *PlotDetector
	narrationAgent NarrationAgent
	logger         *logger.Logger

	// Guardian integration (Story 9-11) - optional
	tensionManager TensionManager
	tensionValue   float64
	currentPhase   string
	mu             sync.RWMutex
}

// NewMomentumController 創建新的動量控制器
// AC2: 接受 config 和 narrationAgent 參數，初始化所有欄位
func NewMomentumController(config *MomentumConfig, narrationAgent NarrationAgent) *MomentumController {
	// 如果未提供 config，使用預設值
	if config == nil {
		config = DefaultMomentumConfig()
	}

	// Note: logger.GetGlobal() may return nil if InitGlobal() hasn't been called yet.
	// This is acceptable as all logger usage is guarded with nil checks (defensive programming).
	// The controller will function correctly without logging if logger is nil.
	return &MomentumController{
		config:         config,
		riskEvaluator:  NewRiskEvaluator(),
		plotDetector:   NewPlotDetector(),
		narrationAgent: narrationAgent,
		logger:         logger.GetGlobal(),
		tensionValue:   0.0,
		currentPhase:   "Rest", // Initial phase for 0.0 tension
	}
}

// ShouldPauseForChoice 決定是否應暫停詢問玩家
// AC3: 根據多種條件決定是否暫停
// AC4: 按優先級順序檢查條件
func (m *MomentumController) ShouldPauseForChoice(ctx *NarrativeContext) bool {
	if ctx == nil {
		return true // 安全起見，未知狀態下暫停
	}

	// 優先級 1: NPC 主動對話 (最高優先級)
	if m.config.PauseOnNPC && ctx.NPCInitiatesConversation {
		if m.logger != nil {
			m.logger.Debug("Pausing for NPC conversation", map[string]interface{}{
				"npc": ctx.InitiatingNPC,
			})
		}
		return true
	}

	// 優先級 2: 重大事件
	if m.config.PauseOnEvent && m.hasMajorEvent(ctx) {
		if m.logger != nil {
			m.logger.Debug("Pausing for major event", map[string]interface{}{
				"events": len(ctx.PendingEvents),
			})
		}
		return true
	}

	// 優先級 3: 劇情點
	if m.config.PauseOnPlot && m.plotDetector.Detect(ctx) {
		if m.logger != nil {
			m.logger.Debug("Pausing for plot point", map[string]interface{}{
				"plot_type": ctx.PlotPointType,
			})
		}
		return true
	}

	// 優先級 4: 風險等級
	risk := m.riskEvaluator.Evaluate(ctx)
	if risk >= m.config.PauseOnRisk {
		if m.logger != nil {
			m.logger.Debug("Pausing for risk level", map[string]interface{}{
				"risk":           risk.String(),
				"pause_on_risk":  m.config.PauseOnRisk.String(),
			})
		}
		return true
	}

	// 優先級 5: 最大自動回合數
	if m.config.AutoResolve && ctx.AutoResolvedBeats >= m.config.MaxAutoBeats {
		if m.logger != nil {
			m.logger.Debug("Pausing for max auto beats", map[string]interface{}{
				"auto_beats":     ctx.AutoResolvedBeats,
				"max_auto_beats": m.config.MaxAutoBeats,
			})
		}
		return true
	}

	// 優先級 6: Frequency 等級決定
	return m.shouldPauseByFrequency(risk)
}

// hasMajorEvent 檢查是否有重大事件待處理
func (m *MomentumController) hasMajorEvent(ctx *NarrativeContext) bool {
	for _, event := range ctx.PendingEvents {
		if event.IsMajor {
			return true
		}
	}
	return false
}

// shouldPauseByFrequency 根據頻率等級決定是否暫停
func (m *MomentumController) shouldPauseByFrequency(risk RiskLevel) bool {
	switch m.config.Frequency {
	case FrequencyHigh:
		// 高頻率：總是暫停
		return true
	case FrequencyMedium:
		// 中頻率：中風險以上暫停
		return risk >= RiskMedium
	case FrequencyLow:
		// 低頻率：高風險以上暫停
		return risk >= RiskHigh
	default:
		// 預設：中等頻率
		return risk >= RiskMedium
	}
}

// DetermineStopReason 確定暫停原因 (用於日誌和 UI 顯示)
func (m *MomentumController) DetermineStopReason(ctx *NarrativeContext) StopReason {
	if ctx == nil {
		return StopReasonNone
	}

	// 按優先級順序檢查
	if m.config.PauseOnNPC && ctx.NPCInitiatesConversation {
		return StopReasonNPCConversation
	}

	if m.config.PauseOnEvent && m.hasMajorEvent(ctx) {
		return StopReasonMajorEvent
	}

	if m.config.PauseOnPlot && m.plotDetector.Detect(ctx) {
		return StopReasonPlotPoint
	}

	risk := m.riskEvaluator.Evaluate(ctx)
	if risk >= m.config.PauseOnRisk {
		return StopReasonRiskLevel
	}

	if m.config.AutoResolve && ctx.AutoResolvedBeats >= m.config.MaxAutoBeats {
		return StopReasonMaxAutoBeats
	}

	if m.shouldPauseByFrequency(risk) {
		return StopReasonFrequency
	}

	return StopReasonNone
}

// GetConfig 返回當前配置 (用於測試和檢視)
func (m *MomentumController) GetConfig() *MomentumConfig {
	return m.config
}

// SetConfig 更新配置 (允許動態調整)
func (m *MomentumController) SetConfig(config *MomentumConfig) {
	if config != nil {
		m.config = config
	}
}

// AutoResolve 自動演繹低風險行動直到觸發暫停條件
// Story 7-1 Implementation
//
// AC1: AutoResolve() 自動演繹低風險行動
// AC2: 返回 AutoResolveResult (Narratives/BeatsResolved/HPDelta/SANDelta)
// AC3: 迴圈生成敘事直到觸發暫停條件
// AC4: 每回合調用 NarrationAgent.InvokeAutoContent()
// AC5: 更新 GameStateV2 狀態（Beat/HP/SAN）
// AC6: 尊重 MaxAutoBeats 上限
// AC7: 記錄 StopReason/StopContext
//
// Parameters:
//   - ctx: NarrativeContext containing current game state
//
// Returns:
//   - *AutoResolveResult: Results of auto-resolution including narratives and state changes
func (m *MomentumController) AutoResolve(ctx *NarrativeContext) *AutoResolveResult {
	if ctx == nil {
		if m.logger != nil {
			m.logger.Error("AutoResolve called with nil context", nil)
		}
		return &AutoResolveResult{
			StopReason:  StopReasonNone,
			StopContext: nil,
		}
	}

	// Initialize result
	result := &AutoResolveResult{
		Narratives:    make([]string, 0),
		BeatsResolved: 0,
		HPDelta:       0,
		SANDelta:      0,
		PlantedSeeds:  make([]string, 0),
		RevealedClues: make([]string, 0),
	}

	// Check if AutoResolve is enabled
	if !m.config.AutoResolve {
		if m.logger != nil {
			m.logger.Debug("AutoResolve is disabled", nil)
		}
		result.StopReason = StopReasonNone
		result.StopContext = ctx
		return result
	}

	// AC3 & AC6: Loop until pause condition or MaxAutoBeats reached
	beatsGenerated := 0
	for beatsGenerated < m.config.MaxAutoBeats {
		// Update context with current auto-resolved beats count
		ctx.AutoResolvedBeats = beatsGenerated

		// Check if we should pause (AC3)
		if m.ShouldPauseForChoice(ctx) {
			reason := m.DetermineStopReason(ctx)
			if m.logger != nil {
				m.logger.Debug("AutoResolve paused", map[string]interface{}{
					"beats_generated": beatsGenerated,
					"stop_reason":     reason.String(),
				})
			}
			result.StopReason = reason
			result.StopContext = ctx
			break
		}

		// AC4: Call NarrationAgent.InvokeAutoContent() for each beat
		// Note: This is a placeholder. Actual implementation will need GameStateV2 and StoryBible
		// The orchestrator will need to provide these and call the narration agent
		if m.narrationAgent != nil {
			// TODO: Implement actual LLM call via NarrationAgent.InvokeAutoContent()
			// For now, we create a placeholder narrative
			narrative := m.generatePlaceholderNarrative(ctx, beatsGenerated)
			result.Narratives = append(result.Narratives, narrative)

			// Placeholder HP/SAN delta (actual values should come from NarrationAgent response)
			// AC5: Update state deltas
			result.HPDelta += 0   // Will be updated by actual narration
			result.SANDelta += -2 // Slight SAN decrease for progression
		}

		beatsGenerated++
		result.BeatsResolved++

		// Update beat count in context for next iteration
		ctx.CurrentBeat++
	}

	// AC6: Reached MaxAutoBeats limit
	if beatsGenerated >= m.config.MaxAutoBeats {
		if m.logger != nil {
			m.logger.Debug("AutoResolve reached MaxAutoBeats", map[string]interface{}{
				"max_auto_beats": m.config.MaxAutoBeats,
			})
		}
		result.StopReason = StopReasonMaxAutoBeats
		result.StopContext = ctx
	}

	// AC7: Return result with stop reason and context
	if m.logger != nil {
		m.logger.Info("AutoResolve completed", map[string]interface{}{
			"beats_resolved": result.BeatsResolved,
			"narratives":     len(result.Narratives),
			"hp_delta":       result.HPDelta,
			"san_delta":      result.SANDelta,
			"stop_reason":    result.StopReason.String(),
		})
	}

	return result
}

// generatePlaceholderNarrative generates a placeholder narrative for testing
// This will be replaced by actual NarrationAgent.InvokeAutoContent() call
func (m *MomentumController) generatePlaceholderNarrative(ctx *NarrativeContext, beatIndex int) string {
	// Placeholder implementation for testing
	// Real implementation will call NarrationAgent.InvokeAutoContent()
	return "你繼續前進，周圍的氣氛變得越來越詭異..."
}

// =============================================================================
// Guardian Integration (Story 9-11)
// =============================================================================

// SetTensionManager sets the optional Guardian TensionManager for bidirectional integration
// AC1: MomentumController can hold TensionManager reference (optional)
func (m *MomentumController) SetTensionManager(tm TensionManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tensionManager = tm
	if m.logger != nil {
		m.logger.Debug("MomentumController: TensionManager set", nil)
	}
}

// GetTensionManager returns the current TensionManager (may be nil)
func (m *MomentumController) GetTensionManager() TensionManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tensionManager
}

// HasTensionManager checks if a TensionManager is configured
// AC5: Supports pure MomentumController mode without Guardian
func (m *MomentumController) HasTensionManager() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tensionManager != nil
}

// SetTension sets the tension value and notifies Guardian if present
// AC2: Tension changes automatically sync to Guardian
func (m *MomentumController) SetTension(value float64) {
	m.mu.Lock()

	oldValue := m.tensionValue
	m.tensionValue = clamp(value, 0.0, 1.0)
	m.updatePhaseUnlocked()

	// Notify Guardian if present (AC2: automatic synchronization)
	// Note: This only syncs state, does NOT trigger adjustments (avoids circular calls)
	tensionManager := m.tensionManager
	currentPhase := m.currentPhase
	tensionValue := m.tensionValue

	m.mu.Unlock()

	// Call Guardian outside of lock to avoid deadlock
	if tensionManager != nil {
		tensionManager.SyncFromMomentum(tensionValue, currentPhase)
		if m.logger != nil {
			m.logger.Debug("MomentumController: Tension synced to Guardian", map[string]interface{}{
				"old_value": oldValue,
				"new_value": tensionValue,
				"phase":     currentPhase,
			})
		}
	}
}

// AdjustTension adjusts tension by delta and notifies Guardian if present
// AC2: Tension adjustments automatically sync to Guardian
func (m *MomentumController) AdjustTension(delta float64) {
	m.mu.Lock()

	oldValue := m.tensionValue
	m.tensionValue = clamp(m.tensionValue+delta, 0.0, 1.0)
	m.updatePhaseUnlocked()

	// Notify Guardian if present
	tensionManager := m.tensionManager
	currentPhase := m.currentPhase
	tensionValue := m.tensionValue

	m.mu.Unlock()

	// Call Guardian outside of lock to avoid deadlock
	if tensionManager != nil {
		tensionManager.SyncFromMomentum(tensionValue, currentPhase)
		if m.logger != nil {
			m.logger.Debug("MomentumController: Tension adjusted and synced to Guardian", map[string]interface{}{
				"old_value": oldValue,
				"delta":     delta,
				"new_value": tensionValue,
				"phase":     currentPhase,
			})
		}
	}
}

// GetTension returns the current tension value (0.0-1.0)
func (m *MomentumController) GetTension() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tensionValue
}

// GetCurrentPhase returns the current tension phase
// AC4: Phase changes automatically sync to Guardian
func (m *MomentumController) GetCurrentPhase() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentPhase
}

// updatePhaseUnlocked updates the tension phase based on current tension value
// Must be called with lock held
func (m *MomentumController) updatePhaseUnlocked() {
	oldPhase := m.currentPhase

	// Map tension (0.0-1.0) to Guardian phases
	switch {
	case m.tensionValue < 0.25:
		m.currentPhase = "Rest"
	case m.tensionValue < 0.60:
		m.currentPhase = "Buildup"
	case m.tensionValue < 0.90:
		m.currentPhase = "Peak"
	default:
		m.currentPhase = "Release"
	}

	if oldPhase != m.currentPhase && m.logger != nil {
		m.logger.Debug("MomentumController: Phase changed", map[string]interface{}{
			"old_phase":     oldPhase,
			"new_phase":     m.currentPhase,
			"tension_value": m.tensionValue,
		})
	}
}

// clamp clamps a value between min and max
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
