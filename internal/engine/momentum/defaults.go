// Package momentum 提供敘事動量控制系統
// Epic 6: Story 6.1 - MomentumController 基礎結構
package momentum

// DefaultMomentumConfig 返回預設的動量配置
// 適用於大部分玩家的平衡體驗
func DefaultMomentumConfig() *MomentumConfig {
	return &MomentumConfig{
		Frequency:      FrequencyMedium,
		AutoResolve:    true,
		MaxAutoBeats:   5,
		PauseOnRisk:    RiskMedium,
		PauseOnPlot:    true,
		PauseOnNPC:     true,
		PauseOnEvent:   true,
		PlayerOverride: true,
	}
}

// GetMomentumConfigForDifficulty 根據難度等級返回對應的動量配置
// 不同難度下的自動演繹策略不同
func GetMomentumConfigForDifficulty(difficulty string) *MomentumConfig {
	config := DefaultMomentumConfig()

	switch difficulty {
	case "easy", "Easy":
		// 簡單難度：更頻繁暫停，降低風險
		config.Frequency = FrequencyHigh
		config.MaxAutoBeats = 3
		config.PauseOnRisk = RiskLow

	case "normal", "Normal":
		// 普通難度：使用預設值
		// 已在 DefaultMomentumConfig() 設定

	case "hard", "Hard":
		// 困難難度：減少暫停，增加挑戰
		// Story 7-5 AC3: MaxAutoBeats=7, PauseOnRisk=High
		config.Frequency = FrequencyMedium
		config.MaxAutoBeats = 7
		config.PauseOnRisk = RiskHigh

	case "nightmare", "Nightmare":
		// 夢魘難度：很少暫停，只在高風險時暫停
		config.Frequency = FrequencyLow
		config.MaxAutoBeats = 10
		config.PauseOnRisk = RiskHigh

	case "hell", "Hell":
		// 地獄難度：最少暫停，只在致命風險暫停
		// Story 7-5 AC4: MaxAutoBeats=10, PauseOnRisk=Lethal
		config.Frequency = FrequencyLow
		config.MaxAutoBeats = 10
		config.PauseOnRisk = RiskLethal
		config.AutoResolve = true // 仍保留自動演繹，但限制更少

	default:
		// 未知難度：使用預設值
	}

	return config
}

// CinematicModeConfig 返回電影模式的配置
// 適合觀賞導向的玩家，自動演繹大部分內容
func CinematicModeConfig() *MomentumConfig {
	return &MomentumConfig{
		Frequency:      FrequencyLow,
		AutoResolve:    true,
		MaxAutoBeats:   15, // 允許長時間自動演繹
		PauseOnRisk:    RiskHigh,
		PauseOnPlot:    true,  // 仍在劇情點暫停
		PauseOnNPC:     true,  // 仍在 NPC 對話暫停
		PauseOnEvent:   true,  // 仍在重大事件暫停
		PlayerOverride: true,
	}
}

// InteractiveModeConfig 返回互動模式的配置
// 適合喜歡頻繁選擇的玩家
func InteractiveModeConfig() *MomentumConfig {
	return &MomentumConfig{
		Frequency:      FrequencyHigh,
		AutoResolve:    false, // 禁用自動演繹
		MaxAutoBeats:   0,
		PauseOnRisk:    RiskNone, // 任何情況都暫停
		PauseOnPlot:    true,
		PauseOnNPC:     true,
		PauseOnEvent:   true,
		PlayerOverride: true,
	}
}

// DifficultyToString converts game difficulty level to string
// Story 7-5: Helper for GameConfig integration
func DifficultyToString(difficulty int) string {
	switch difficulty {
	case 0: // DifficultyEasy
		return "easy"
	case 1: // DifficultyHard
		return "hard"
	case 2: // DifficultyHell
		return "hell"
	default:
		return "normal"
	}
}
