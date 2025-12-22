package knowledge

import (
	"context"
	"time"
)

// ContradictionType represents the severity category of a contradiction.
type ContradictionType string

const (
	// ContradictionMinor represents slight uncertainty or minor contradiction (severity 1-4)
	ContradictionMinor ContradictionType = "minor"
	// ContradictionModerate represents confusion requiring explanation (severity 5-7)
	ContradictionModerate ContradictionType = "moderate"
	// ContradictionMajor represents strong doubt or cognitive dissonance (severity 8-10)
	ContradictionMajor ContradictionType = "major"
)

// EmotionDelta represents a change in emotional state.
// This is a simplified version for the knowledge package.
// The full version exists in internal/npc/manager/emotion.go
type EmotionDelta struct {
	Trust  int `json:"trust"`  // Change in trust (-100 to +100)
	Fear   int `json:"fear"`   // Change in fear (-100 to +100)
	Stress int `json:"stress"` // Change in stress (-100 to +100)
}

// ContradictionResult contains the full analysis of a detected contradiction.
// It includes what contradicts, how severe it is, and suggested reactions.
type ContradictionResult struct {
	Type              ContradictionType `json:"type"`               // Minor, Moderate, or Major
	ExistingFact      *KnownFact        `json:"existing_fact"`      // The fact the entity already knows
	NewInfo           string            `json:"new_info"`           // The new contradictory information
	Severity          int               `json:"severity"`           // 1-10 severity score
	SuggestedDelta    EmotionDelta      `json:"suggested_delta"`    // Recommended emotional impact
	SuggestedReaction string            `json:"suggested_reaction"` // Suggested NPC response description
}

// negationPairs maps contradictory terms bidirectionally.
// When checking contradictions, we look for these opposing concepts.
var negationPairs = map[string]string{
	// Life and death
	"活著":  "死了",
	"死了":  "活著",
	"存活":  "死亡",
	"死亡":  "存活",
	"活着":  "死了", // Simplified Chinese variant
	"死掉":  "活著",
	"生還":  "死亡",
	"陣亡":  "活著",

	// Safety and danger
	"安全":  "危險",
	"危險":  "安全",
	"危险":  "安全", // Simplified Chinese
	"平安":  "危險",
	"險惡":  "安全",

	// Trust
	"可信":   "不可信",
	"不可信":  "可信",
	"可靠":   "不可靠",
	"不可靠":  "可靠",
	"值得信賴": "不值得信賴",

	// Existence
	"存在":  "不存在",
	"不存在": "存在",
	"有":   "沒有",
	"沒有":  "有",

	// States
	"開著": "關著",
	"關著": "開著",
	"打開": "關閉",
	"關閉": "打開",
	"開启": "关闭", // Simplified Chinese
	"关闭": "开启",

	// Presence
	"有人": "沒人",
	"沒人": "有人",
	"无人": "有人", // Simplified Chinese

	// Truth
	"真的": "假的",
	"假的": "真的",
	"真实": "虚假", // Simplified Chinese
	"虚假": "真实",

	// Light and dark
	"明亮": "黑暗",
	"黑暗": "明亮",
	"光明": "黑暗",
	"昏暗": "明亮",

	// Normal and abnormal
	"正常": "異常",
	"異常": "正常",
	"异常": "正常", // Simplified Chinese
}

// CheckContradiction checks if new information contradicts what an entity already knows.
// It scans through the entity's knowledge base and returns the first contradiction found.
//
// Story 8.2: Enhanced with LLM-based semantic contradiction detection.
// If a ContradictionAnalyzer is configured, it uses semantic understanding.
// Otherwise, falls back to keyword-based detection.
//
// This method is thread-safe (uses RLock).
//
// Parameters:
//   - entityID: The entity whose knowledge to check
//   - newInfo: The new information to check for contradictions
//
// Returns:
//   - *ContradictionResult if a contradiction is found, nil otherwise
//
// AC1: CheckContradiction() 檢查新資訊與已知資訊的矛盾
func (m *UpdateManager) CheckContradiction(entityID, newInfo string) *ContradictionResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get entity's knowledge base
	kb := m.getKnowledgeBaseLocked(entityID)
	if kb == nil {
		return nil // No knowledge base means no contradictions
	}

	// Check each known fact for contradictions
	for factID, knownFact := range kb.KnownFacts {
		// Get the original fact from global repository
		originalFact := m.globalFacts[factID]
		if originalFact == nil {
			continue
		}

		// Get the content as the entity knows it (may be distorted)
		content := originalFact.Content
		if knownFact.IsDistorted && knownFact.DistortedContent != "" {
			content = knownFact.DistortedContent
		}

		// Story 8.2: Use LLM-based semantic analysis if analyzer is available
		var isContradictory bool
		var severity int
		var detectedType ContradictionType

		if m.contradictionAnalyzer != nil {
			// Use LLM semantic analysis
			result := m.checkContradictionWithLLM(content, newInfo, knownFact)
			if result != nil {
				isContradictory = result.IsContradictory
				severity = result.Severity
				_ = result.Type // Use the type from LLM analysis

				// Adjust severity based on known fact properties
				severity = m.adjustSeverityWithFactContext(severity, knownFact)
				detectedType = m.getContradictionType(severity)
			} else {
				// LLM analysis failed, fall back to keyword matching
				isContradictory = m.contradicts(content, newInfo)
				if isContradictory {
					severity = m.calculateContradictionSeverity(knownFact, newInfo)
					detectedType = m.getContradictionType(severity)
				}
			}
		} else {
			// Fall back to keyword-based detection
			isContradictory = m.contradicts(content, newInfo)
			if isContradictory {
				severity = m.calculateContradictionSeverity(knownFact, newInfo)
				detectedType = m.getContradictionType(severity)
			}
		}

		// If contradiction found, return result
		if isContradictory {
			return &ContradictionResult{
				Type:              detectedType,
				ExistingFact:      knownFact,
				NewInfo:           newInfo,
				Severity:          severity,
				SuggestedDelta:    m.getSuggestedEmotionDelta(severity),
				SuggestedReaction: m.getSuggestedReaction(severity),
			}
		}
	}

	return nil // No contradiction found
}

// contradicts checks if two pieces of information contradict each other
// based on predefined negation pairs (e.g., "alive" vs "dead").
//
// This is a simple keyword-based approach that will be enhanced with
// LLM-based semantic understanding in Story 8.2.
//
// Parameters:
//   - existing: The existing information
//   - new: The new information
//
// Returns:
//   - true if the information contradicts, false otherwise
//
// AC2: contradicts() 基於否定關係檢測矛盾
func (m *UpdateManager) contradicts(existing, new string) bool {
	// Check all negation pairs
	for positive, negative := range negationPairs {
		// Check if existing contains positive and new contains negative
		if containsIgnoreCase(existing, positive) && containsIgnoreCase(new, negative) {
			return true
		}
		// No need to check the reverse since negationPairs includes both directions
	}

	return false
}

// calculateContradictionSeverity calculates how severe a contradiction is (1-10).
// Higher severity occurs when:
//   - The existing fact has high confidence (>0.8)
//   - The existing fact was witnessed firsthand
//
// Base severity: 5
// High confidence (>0.8): +2
// Witnessed firsthand: +3
// Maximum: 10, Minimum: 1
//
// Parameters:
//   - knownFact: The existing known fact that is being contradicted
//   - newInfo: The new contradictory information (currently unused, but may be used for future enhancements)
//
// Returns:
//   - Severity score from 1-10
//
// AC3: calculateContradictionSeverity() 計算嚴重程度（1-10）
func (m *UpdateManager) calculateContradictionSeverity(knownFact *KnownFact, newInfo string) int {
	severity := 5 // Base severity

	// High confidence makes contradiction more severe
	if knownFact.Confidence > 0.8 {
		severity += 2
	}

	// Witnessing firsthand makes contradiction much more severe
	if knownFact.LearnMethod == Witness {
		severity += 3
	}

	// Clamp to valid range
	if severity > 10 {
		severity = 10
	}
	if severity < 1 {
		severity = 1
	}

	return severity
}

// getContradictionType converts a severity score into a contradiction type category.
//
// Severity ranges:
//   - 8-10: Major (strong doubt, cognitive dissonance)
//   - 5-7: Moderate (confusion, needs explanation)
//   - 1-4: Minor (slight uncertainty, may be ignored)
//
// Parameters:
//   - severity: The severity score (1-10)
//
// Returns:
//   - ContradictionType enum (Minor, Moderate, or Major)
//
// AC4: getContradictionType() 返回 minor/moderate/major
func (m *UpdateManager) getContradictionType(severity int) ContradictionType {
	if severity >= 8 {
		return ContradictionMajor
	}
	if severity >= 5 {
		return ContradictionModerate
	}
	return ContradictionMinor
}

// getSuggestedEmotionDelta returns the recommended emotional impact based on severity.
//
// Emotion changes by severity:
//   - Major (8-10): Trust -20, Fear +10, Stress +25
//   - Moderate (5-7): Trust -10, Fear +5, Stress +15
//   - Minor (1-4): Trust -5, Fear +0, Stress +5
//
// These suggestions can be applied to NPC emotional state to reflect their
// reaction to contradictory information.
//
// Parameters:
//   - severity: The severity score (1-10)
//
// Returns:
//   - EmotionDelta with suggested changes
//
// AC5: getSuggestedEmotionDelta() 根據嚴重程度建議情感影響
func (m *UpdateManager) getSuggestedEmotionDelta(severity int) EmotionDelta {
	switch {
	case severity >= 8:
		// Major contradiction: Significant trust loss, increased fear and stress
		return EmotionDelta{
			Trust:  -20,
			Fear:   +10,
			Stress: +25,
		}
	case severity >= 5:
		// Moderate contradiction: Moderate trust loss and stress
		return EmotionDelta{
			Trust:  -10,
			Fear:   +5,
			Stress: +15,
		}
	default:
		// Minor contradiction: Small impact
		return EmotionDelta{
			Trust:  -5,
			Fear:   0,
			Stress: +5,
		}
	}
}

// getSuggestedReaction returns a text description of how an NPC might react
// to a contradiction of the given severity.
//
// Reaction descriptions:
//   - Major (8-10): "強烈質疑，可能認為對方在說謊或發瘋"
//   - Moderate (5-7): "困惑，要求解釋"
//   - Minor (1-4): "略微疑惑，但可能忽略"
//
// These are suggestions for the NPC generation system to guide appropriate responses.
//
// Parameters:
//   - severity: The severity score (1-10)
//
// Returns:
//   - String description of suggested reaction
//
// AC6: getSuggestedReaction() 建議 NPC 反應
func (m *UpdateManager) getSuggestedReaction(severity int) string {
	switch {
	case severity >= 8:
		return "強烈質疑，可能認為對方在說謊或發瘋"
	case severity >= 5:
		return "困惑，要求解釋"
	default:
		return "略微疑惑，但可能忽略"
	}
}

// checkContradictionWithLLM uses the LLM analyzer to check for semantic contradictions.
// Returns nil if the analysis fails or times out.
//
// Story 8.2 AC1: LLM 語義矛盾判斷
func (m *UpdateManager) checkContradictionWithLLM(existingContent, newInfo string, knownFact *KnownFact) *ContradictionAnalysisResult {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Analyze contradiction
	result, err := m.contradictionAnalyzer.AnalyzeContradiction(ctx, existingContent, newInfo, knownFact)
	if err != nil {
		// Log error but don't fail - return nil to trigger fallback
		return nil
	}

	return result
}

// adjustSeverityWithFactContext adjusts LLM-provided severity based on the known fact's properties.
// This considers:
//   - High confidence (>0.8): +2 to severity
//   - Witnessed firsthand: +3 to severity
//
// Story 8.2 AC2: 考慮信息來源可信度影響嚴重度
func (m *UpdateManager) adjustSeverityWithFactContext(llmSeverity int, knownFact *KnownFact) int {
	severity := llmSeverity

	// High confidence makes contradiction more severe
	if knownFact.Confidence > 0.8 {
		severity += 2
	}

	// Witnessing firsthand makes contradiction much more severe
	if knownFact.LearnMethod == Witness {
		severity += 3
	}

	// Clamp to valid range
	if severity > 10 {
		severity = 10
	}
	if severity < 1 && llmSeverity > 0 {
		severity = 1
	}

	return severity
}

// GetContradictionCacheStats returns cache statistics if analyzer is configured.
// Story 8.2 AC4: 快取命中率監控
func (m *UpdateManager) GetContradictionCacheStats() *CacheStats {
	if m.contradictionAnalyzer == nil {
		return nil
	}
	stats := m.contradictionAnalyzer.GetCacheStats()
	return &stats
}
