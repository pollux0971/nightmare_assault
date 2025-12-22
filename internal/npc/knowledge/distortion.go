package knowledge

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// NPCStateProvider provides access to NPC emotional state and profile information.
// This interface allows the DistortionCalculator to access NPC data without
// creating circular dependencies between knowledge and manager packages.
type NPCStateProvider interface {
	// GetNPCEmotion returns the emotional state for a given NPC ID.
	// Returns (trust, fear, stress) values (0-100) and error if NPC not found.
	GetNPCEmotion(npcID string) (trust, fear, stress int, err error)

	// GetNPCTraits returns the trait IDs for a given NPC.
	// Some traits (like "anxious", "paranoid") may increase distortion tendency.
	GetNPCTraits(npcID string) ([]string, error)
}

// DistortionCalculator implements intelligent information distortion logic.
// It determines how likely an NPC is to distort information based on:
// - NPC personality traits
// - Current emotional state (fear, stress, trust)
// - Propagation depth (how many times information has been passed)
// - Relationship with the player (trust level)
//
// Story 8.3 AC1: 更智慧的扭曲邏輯（基於 NPC 個性）
// Story 8.3 AC2: 扭曲程度與傳播深度相關
// Story 8.3 AC3: 恐懼/壓力高的 NPC 更容易扭曲
// Story 8.3 AC4: 玩家可透過高信任度獲得準確資訊
type DistortionCalculator struct {
	// stateProvider provides access to NPC emotional state and profile data
	stateProvider NPCStateProvider

	// config holds distortion calculation parameters
	config *DistortionConfig
}

// DistortionConfig contains tunable parameters for distortion calculation.
type DistortionConfig struct {
	// BaseDistortionRate is the baseline probability of distortion (0.0-1.0)
	// Default: 0.15 (15%)
	BaseDistortionRate float64

	// FearWeight determines how much fear affects distortion (0.0-1.0)
	// Higher values make fear more influential
	// Default: 0.4
	FearWeight float64

	// StressWeight determines how much stress affects distortion (0.0-1.0)
	// Higher values make stress more influential
	// Default: 0.3
	StressWeight float64

	// TrustProtection determines how much trust reduces distortion (0.0-1.0)
	// Higher values give more protection from distortion at high trust
	// Default: 0.5
	TrustProtection float64

	// DepthMultiplier increases distortion rate per propagation depth level
	// Each depth level multiplies the distortion rate by this value
	// Default: 1.25 (25% increase per level)
	DepthMultiplier float64

	// TraitModifiers maps trait IDs to distortion modifiers
	// Positive values increase distortion, negative values decrease it
	// Default modifiers:
	//   "anxious": +0.15
	//   "paranoid": +0.25
	//   "calm": -0.10
	//   "rational": -0.15
	TraitModifiers map[string]float64

	// HighTrustThreshold is the trust level above which distortion is significantly reduced
	// Default: 70
	HighTrustThreshold int

	// HighFearThreshold is the fear level above which distortion is significantly increased
	// Default: 60
	HighFearThreshold int

	// HighStressThreshold is the stress level above which distortion is significantly increased
	// Default: 60
	HighStressThreshold int
}

// DefaultDistortionConfig returns a DistortionConfig with balanced default values.
func DefaultDistortionConfig() *DistortionConfig {
	return &DistortionConfig{
		BaseDistortionRate:  0.15,
		FearWeight:          0.4,
		StressWeight:        0.3,
		TrustProtection:     0.5,
		DepthMultiplier:     1.25,
		HighTrustThreshold:  70,
		HighFearThreshold:   60,
		HighStressThreshold: 60,
		TraitModifiers: map[string]float64{
			"anxious":   0.15,
			"paranoid":  0.25,
			"calm":      -0.10,
			"rational":  -0.15,
			"nervous":   0.12,
			"composed":  -0.12,
			"unstable":  0.20,
			"reliable":  -0.18,
		},
	}
}

// NewDistortionCalculator creates a new DistortionCalculator with the given configuration.
// If config is nil, it uses DefaultDistortionConfig().
func NewDistortionCalculator(stateProvider NPCStateProvider, config *DistortionConfig) *DistortionCalculator {
	if config == nil {
		config = DefaultDistortionConfig()
	}

	return &DistortionCalculator{
		stateProvider: stateProvider,
		config:        config,
	}
}

// DistortionResult contains the result of distortion calculation.
type DistortionResult struct {
	// DistortionRate is the calculated probability of distortion (0.0-1.0)
	DistortionRate float64

	// ShouldDistort is true if distortion should occur (based on random roll)
	ShouldDistort bool

	// DistortedContent is the distorted version of the content (if ShouldDistort is true)
	DistortedContent string

	// Factors contains a breakdown of contributing factors for debugging
	Factors DistortionFactors
}

// DistortionFactors contains a breakdown of factors contributing to the distortion rate.
type DistortionFactors struct {
	BaseRate         float64 `json:"base_rate"`
	FearContribution float64 `json:"fear_contribution"`
	StressContribution float64 `json:"stress_contribution"`
	TrustReduction   float64 `json:"trust_reduction"`
	DepthMultiplier  float64 `json:"depth_multiplier"`
	TraitModifier    float64 `json:"trait_modifier"`
	FinalRate        float64 `json:"final_rate"`
}

// CalculateDistortionRate calculates the probability that an NPC will distort information.
//
// The calculation follows this formula:
// 1. Start with base distortion rate
// 2. Add fear contribution: (fear/100) * fearWeight
// 3. Add stress contribution: (stress/100) * stressWeight
// 4. Subtract trust reduction: (trust/100) * trustProtection
// 5. Apply trait modifiers (sum of all applicable trait modifiers)
// 6. Multiply by depth factor: rate * (depthMultiplier ^ propagationDepth)
// 7. Clamp result to [0.0, 1.0]
//
// Story 8.3 AC1: 更智慧的扭曲邏輯（基於 NPC 個性）
// Story 8.3 AC2: 扭曲程度與傳播深度相關
// Story 8.3 AC3: 恐懼/壓力高的 NPC 更容易扭曲
// Story 8.3 AC4: 玩家可透過高信任度獲得準確資訊
//
// Parameters:
//   - npcID: The NPC ID who is receiving/propagating the information
//   - propagationDepth: How many times the information has been passed (0 = original witness)
//
// Returns:
//   - float64: Distortion probability (0.0-1.0)
//   - DistortionFactors: Breakdown of contributing factors
//   - error: If NPC not found or other error
func (dc *DistortionCalculator) CalculateDistortionRate(npcID string, propagationDepth int) (float64, DistortionFactors, error) {
	// Initialize factors
	factors := DistortionFactors{
		BaseRate: dc.config.BaseDistortionRate,
	}

	// Get NPC emotional state
	trust, fear, stress, err := dc.stateProvider.GetNPCEmotion(npcID)
	if err != nil {
		return 0, factors, fmt.Errorf("failed to get NPC emotion: %w", err)
	}

	// Calculate fear contribution (0.0 to fearWeight)
	// Story 8.3 AC3: Higher fear increases distortion
	fearNormalized := float64(fear) / 100.0
	factors.FearContribution = fearNormalized * dc.config.FearWeight

	// Calculate stress contribution (0.0 to stressWeight)
	// Story 8.3 AC3: Higher stress increases distortion
	stressNormalized := float64(stress) / 100.0
	factors.StressContribution = stressNormalized * dc.config.StressWeight

	// Calculate trust reduction (0.0 to trustProtection)
	// Story 8.3 AC4: Higher trust reduces distortion
	trustNormalized := float64(trust) / 100.0
	factors.TrustReduction = trustNormalized * dc.config.TrustProtection

	// Get NPC traits and calculate trait modifier
	// Story 8.3 AC1: Personality-based distortion
	traits, err := dc.stateProvider.GetNPCTraits(npcID)
	if err != nil {
		// If traits can't be retrieved, continue with 0 modifier
		factors.TraitModifier = 0
	} else {
		factors.TraitModifier = dc.calculateTraitModifier(traits)
	}

	// Calculate depth multiplier
	// Story 8.3 AC2: Distortion increases with propagation depth
	factors.DepthMultiplier = math.Pow(dc.config.DepthMultiplier, float64(propagationDepth))

	// Combine all factors
	// Formula: (base + fear + stress - trust + trait) * depth
	rate := factors.BaseRate +
		factors.FearContribution +
		factors.StressContribution -
		factors.TrustReduction +
		factors.TraitModifier

	// Apply depth multiplier
	rate *= factors.DepthMultiplier

	// Clamp to [0.0, 1.0]
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}

	factors.FinalRate = rate

	return rate, factors, nil
}

// calculateTraitModifier sums up the distortion modifiers for all NPC traits.
// Story 8.3 AC1: 更智慧的扭曲邏輯（基於 NPC 個性）
func (dc *DistortionCalculator) calculateTraitModifier(traits []string) float64 {
	modifier := 0.0

	for _, traitID := range traits {
		// Check if this trait has a modifier
		if traitMod, exists := dc.config.TraitModifiers[traitID]; exists {
			modifier += traitMod
		}
	}

	return modifier
}

// ApplyDistortion performs the full distortion process:
// 1. Calculates distortion rate based on NPC state
// 2. Performs random roll to determine if distortion occurs
// 3. If distortion occurs, generates distorted content
//
// Parameters:
//   - npcID: The NPC ID who is receiving/propagating the information
//   - factContent: The original fact content
//   - propagationDepth: How many times the information has been passed
//
// Returns:
//   - DistortionResult: Contains distortion rate, whether distortion occurred, and distorted content
//   - error: If calculation fails
func (dc *DistortionCalculator) ApplyDistortion(npcID string, factContent string, propagationDepth int) (DistortionResult, error) {
	result := DistortionResult{}

	// Calculate distortion rate
	rate, factors, err := dc.CalculateDistortionRate(npcID, propagationDepth)
	if err != nil {
		return result, err
	}

	result.DistortionRate = rate
	result.Factors = factors

	// Perform random roll
	roll := rand.Float64()
	result.ShouldDistort = roll < rate

	// If distortion should occur, generate distorted content
	if result.ShouldDistort {
		result.DistortedContent = dc.distortContent(factContent, npcID, factors)
	}

	return result, nil
}

// distortContent generates a distorted version of the fact content.
// The type and severity of distortion depends on the NPC's emotional state.
//
// Distortion types:
// - High fear (>= 60): Add fear-based qualifiers ("恐怖的", "可怕的")
// - High stress (>= 60): Add uncertainty markers ("可能", "似乎", "不確定")
// - Low trust (< 30): Add skeptical qualifiers ("據說", "有人說")
// - Default: Generic uncertainty markers
func (dc *DistortionCalculator) distortContent(content string, npcID string, factors DistortionFactors) string {
	// Get NPC emotional state for context-aware distortion
	trust, fear, stress, err := dc.stateProvider.GetNPCEmotion(npcID)
	if err != nil {
		// Fallback to simple distortion if emotion state unavailable
		return dc.applySimpleDistortion(content)
	}

	// High fear: Add fear-based qualifiers
	if fear >= dc.config.HighFearThreshold {
		return dc.applyFearDistortion(content)
	}

	// High stress: Add uncertainty markers
	if stress >= dc.config.HighStressThreshold {
		return dc.applyStressDistortion(content)
	}

	// Low trust: Add skeptical qualifiers
	if trust < 30 {
		return dc.applyLowTrustDistortion(content)
	}

	// Default: Generic uncertainty
	return dc.applySimpleDistortion(content)
}

// applyFearDistortion adds fear-based qualifiers to content.
func (dc *DistortionCalculator) applyFearDistortion(content string) string {
	prefixes := []string{
		"我感覺有些恐怖，",
		"令人不安的是，",
		"讓我害怕的是，",
		"可怕的是，",
	}
	suffixes := []string{
		"（但我很害怕）",
		"（這讓我很不安）",
		"（我不確定是不是真的）",
	}

	choice := rand.Intn(2)
	if choice == 0 {
		return prefixes[rand.Intn(len(prefixes))] + content
	}
	return content + suffixes[rand.Intn(len(suffixes))]
}

// applyStressDistortion adds stress-based uncertainty markers.
func (dc *DistortionCalculator) applyStressDistortion(content string) string {
	prefixes := []string{
		"可能",
		"似乎",
		"好像",
		"或許",
		"我覺得",
	}
	suffixes := []string{
		"（不太確定）",
		"（可能記錯了）",
		"（我有點混亂）",
		"（壓力太大記不清了）",
	}

	choice := rand.Intn(2)
	if choice == 0 {
		return prefixes[rand.Intn(len(prefixes))] + content
	}
	return content + suffixes[rand.Intn(len(suffixes))]
}

// applyLowTrustDistortion adds skeptical qualifiers.
func (dc *DistortionCalculator) applyLowTrustDistortion(content string) string {
	prefixes := []string{
		"據說",
		"聽說",
		"有人說",
		"傳言",
	}
	suffixes := []string{
		"（但我不太相信）",
		"（存疑）",
		"（未經證實）",
	}

	choice := rand.Intn(2)
	if choice == 0 {
		return prefixes[rand.Intn(len(prefixes))] + content
	}
	return content + suffixes[rand.Intn(len(suffixes))]
}

// applySimpleDistortion applies generic uncertainty markers.
func (dc *DistortionCalculator) applySimpleDistortion(content string) string {
	distortions := []string{
		"我聽說" + content,
		"可能" + content,
		"據說" + content,
		content + "（不確定）",
		"好像" + content,
		"似乎" + content,
	}
	return distortions[rand.Intn(len(distortions))]
}

// ShouldDistort is a convenience method that returns true if distortion should occur
// based on the calculated distortion rate and a random roll.
//
// This is a stateless method that can be used for quick distortion checks.
func (dc *DistortionCalculator) ShouldDistort(npcID string, propagationDepth int) (bool, error) {
	rate, _, err := dc.CalculateDistortionRate(npcID, propagationDepth)
	if err != nil {
		return false, err
	}

	return rand.Float64() < rate, nil
}

// GetConfig returns the calculator's configuration.
func (dc *DistortionCalculator) GetConfig() *DistortionConfig {
	return dc.config
}

// ValidateConfig validates the distortion configuration parameters.
// Returns an error if any parameter is out of valid range.
func ValidateConfig(config *DistortionConfig) error {
	if config.BaseDistortionRate < 0.0 || config.BaseDistortionRate > 1.0 {
		return fmt.Errorf("BaseDistortionRate must be between 0.0 and 1.0, got %f", config.BaseDistortionRate)
	}
	if config.FearWeight < 0.0 || config.FearWeight > 1.0 {
		return fmt.Errorf("FearWeight must be between 0.0 and 1.0, got %f", config.FearWeight)
	}
	if config.StressWeight < 0.0 || config.StressWeight > 1.0 {
		return fmt.Errorf("StressWeight must be between 0.0 and 1.0, got %f", config.StressWeight)
	}
	if config.TrustProtection < 0.0 || config.TrustProtection > 1.0 {
		return fmt.Errorf("TrustProtection must be between 0.0 and 1.0, got %f", config.TrustProtection)
	}
	if config.DepthMultiplier < 1.0 {
		return fmt.Errorf("DepthMultiplier must be >= 1.0, got %f", config.DepthMultiplier)
	}
	if config.HighTrustThreshold < 0 || config.HighTrustThreshold > 100 {
		return fmt.Errorf("HighTrustThreshold must be between 0 and 100, got %d", config.HighTrustThreshold)
	}
	if config.HighFearThreshold < 0 || config.HighFearThreshold > 100 {
		return fmt.Errorf("HighFearThreshold must be between 0 and 100, got %d", config.HighFearThreshold)
	}
	if config.HighStressThreshold < 0 || config.HighStressThreshold > 100 {
		return fmt.Errorf("HighStressThreshold must be between 0 and 100, got %d", config.HighStressThreshold)
	}
	return nil
}

// IsDistortionPrefixOrSuffix checks if the content contains known distortion markers.
// This can be used to detect if content has already been distorted.
func IsDistortionPrefixOrSuffix(content string) bool {
	distortionMarkers := []string{
		"我聽說", "可能", "據說", "（不確定）", "好像", "似乎",
		"我感覺有些恐怖", "令人不安的是", "讓我害怕的是", "可怕的是",
		"（但我很害怕）", "（這讓我很不安）", "（我不確定是不是真的）",
		"（不太確定）", "（可能記錯了）", "（我有點混亂）", "（壓力太大記不清了）",
		"聽說", "有人說", "傳言", "（但我不太相信）", "（存疑）", "（未經證實）",
		"或許", "我覺得",
	}

	for _, marker := range distortionMarkers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	return false
}
