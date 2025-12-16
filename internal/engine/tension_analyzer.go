package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// TensionStats contains statistical analysis of tension history
type TensionStats struct {
	PeakValue    int     `json:"peak_value"`    // 最高張力值
	PeakBeat     int     `json:"peak_beat"`     // 最高張力發生的回合
	ValleyValue  int     `json:"valley_value"`  // 最低張力值
	ValleyBeat   int     `json:"valley_beat"`   // 最低張力發生的回合
	AverageValue float64 `json:"average_value"` // 平均張力值
	TotalBeats   int     `json:"total_beats"`   // 總回合數
	TotalChanges int     `json:"total_changes"` // 張力變化次數
}

// TensionAnalyzer provides analysis and visualization of tension history
type TensionAnalyzer struct {
	state *TensionState
}

// NewTensionAnalyzer creates a new analyzer for the given tension state
func NewTensionAnalyzer(state *TensionState) *TensionAnalyzer {
	return &TensionAnalyzer{
		state: state,
	}
}

// CalculateStats analyzes tension history and returns statistics
func (ta *TensionAnalyzer) CalculateStats() *TensionStats {
	history := ta.state.GetHistory()

	if len(history) == 0 {
		// No history, return stats based on current value
		currentValue := ta.state.GetValue()
		return &TensionStats{
			PeakValue:    currentValue,
			PeakBeat:     0,
			ValleyValue:  currentValue,
			ValleyBeat:   0,
			AverageValue: float64(currentValue),
			TotalBeats:   0,
			TotalChanges: 0,
		}
	}

	stats := &TensionStats{
		PeakValue:   history[0].NewValue,
		PeakBeat:    history[0].Beat,
		ValleyValue: history[0].NewValue,
		ValleyBeat:  history[0].Beat,
		TotalChanges: len(history),
	}

	sum := 0
	maxBeat := 0

	for _, entry := range history {
		// Track peak
		if entry.NewValue > stats.PeakValue {
			stats.PeakValue = entry.NewValue
			stats.PeakBeat = entry.Beat
		}

		// Track valley
		if entry.NewValue < stats.ValleyValue {
			stats.ValleyValue = entry.NewValue
			stats.ValleyBeat = entry.Beat
		}

		// Sum for average
		sum += entry.NewValue

		// Track max beat
		if entry.Beat > maxBeat {
			maxBeat = entry.Beat
		}
	}

	stats.AverageValue = float64(sum) / float64(len(history))
	stats.TotalBeats = maxBeat

	return stats
}

// GenerateASCIIChart creates a simple ASCII chart of tension over time
// Height is the number of rows in the chart (default 10)
func (ta *TensionAnalyzer) GenerateASCIIChart(height int) string {
	if height <= 0 {
		height = 10
	}

	history := ta.state.GetHistory()
	if len(history) == 0 {
		return "無張力歷史記錄"
	}

	var chart strings.Builder
	chart.WriteString("【張力歷史圖表】\n")
	chart.WriteString(fmt.Sprintf("總計 %d 筆記錄\n\n", len(history)))

	// Find min and max for scaling
	minVal, maxVal := ta.findMinMax(history)
	valueRange := maxVal - minVal
	if valueRange == 0 {
		valueRange = 1 // Prevent division by zero
	}

	// Draw chart from top to bottom
	for row := height; row >= 0; row-- {
		// Calculate the value threshold for this row
		threshold := minVal + (valueRange * row / height)

		// Y-axis label
		chart.WriteString(fmt.Sprintf("%3d |", threshold))

		// Plot points
		for _, entry := range history {
			if entry.NewValue >= threshold {
				chart.WriteString("█")
			} else {
				chart.WriteString(" ")
			}
		}

		chart.WriteString("\n")
	}

	// X-axis
	chart.WriteString("    +")
	chart.WriteString(strings.Repeat("-", len(history)))
	chart.WriteString("\n")

	// Beat labels (show every 10th beat or so)
	chart.WriteString("     ")
	for i, entry := range history {
		if i == 0 || entry.Beat%10 == 0 {
			chart.WriteString(fmt.Sprintf("%d", entry.Beat%10))
		} else {
			chart.WriteString(" ")
		}
	}
	chart.WriteString(" (Beat)\n\n")

	// Stats summary
	stats := ta.CalculateStats()
	chart.WriteString(fmt.Sprintf("峰值: %d (Beat %d) | ", stats.PeakValue, stats.PeakBeat))
	chart.WriteString(fmt.Sprintf("谷值: %d (Beat %d) | ", stats.ValleyValue, stats.ValleyBeat))
	chart.WriteString(fmt.Sprintf("平均: %.1f\n", stats.AverageValue))

	return chart.String()
}

// findMinMax finds the min and max tension values in history
func (ta *TensionAnalyzer) findMinMax(history []*TensionHistoryEntry) (int, int) {
	if len(history) == 0 {
		return 0, 100
	}

	minVal := history[0].NewValue
	maxVal := history[0].NewValue

	for _, entry := range history {
		if entry.NewValue < minVal {
			minVal = entry.NewValue
		}
		if entry.NewValue > maxVal {
			maxVal = entry.NewValue
		}
	}

	// Add some padding for better visualization
	padding := int(math.Max(5, float64(maxVal-minVal)/10))
	minVal = int(math.Max(0, float64(minVal-padding)))
	maxVal = minVal + padding

	return minVal, maxVal
}

// ExportJSON exports tension history and stats as JSON
func (ta *TensionAnalyzer) ExportJSON() (string, error) {
	data := struct {
		CurrentValue int                     `json:"current_value"`
		CurrentLevel TensionLevel            `json:"current_level"`
		Stats        *TensionStats           `json:"stats"`
		History      []*TensionHistoryEntry  `json:"history"`
	}{
		CurrentValue: ta.state.GetValue(),
		CurrentLevel: ta.state.GetLevel(),
		Stats:        ta.CalculateStats(),
		History:      ta.state.GetHistory(),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetRecentTrend analyzes the recent trend (last N entries)
// Returns "上升", "下降", or "穩定"
func (ta *TensionAnalyzer) GetRecentTrend(lastN int) string {
	history := ta.state.GetHistory()

	if len(history) < 2 {
		return "穩定"
	}

	// Get last N entries
	startIdx := 0
	if len(history) > lastN {
		startIdx = len(history) - lastN
	}
	recentHistory := history[startIdx:]

	// Calculate average change
	totalChange := 0
	for _, entry := range recentHistory {
		totalChange += entry.Delta
	}

	avgChange := float64(totalChange) / float64(len(recentHistory))

	if avgChange > 2 {
		return "上升"
	} else if avgChange < -2 {
		return "下降"
	}
	return "穩定"
}
