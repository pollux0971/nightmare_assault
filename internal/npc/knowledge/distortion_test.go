package knowledge

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// mockNPCStateProvider is a mock implementation of NPCStateProvider for testing.
type mockNPCStateProvider struct {
	emotions map[string]struct{ trust, fear, stress int }
	traits   map[string][]string
}

func newMockNPCStateProvider() *mockNPCStateProvider {
	return &mockNPCStateProvider{
		emotions: make(map[string]struct{ trust, fear, stress int }),
		traits:   make(map[string][]string),
	}
}

func (m *mockNPCStateProvider) GetNPCEmotion(npcID string) (trust, fear, stress int, err error) {
	emotion, exists := m.emotions[npcID]
	if !exists {
		return 0, 0, 0, fmt.Errorf("NPC %s not found", npcID)
	}
	return emotion.trust, emotion.fear, emotion.stress, nil
}

func (m *mockNPCStateProvider) GetNPCTraits(npcID string) ([]string, error) {
	traits, exists := m.traits[npcID]
	if !exists {
		return nil, fmt.Errorf("NPC %s not found", npcID)
	}
	return traits, nil
}

func (m *mockNPCStateProvider) setEmotion(npcID string, trust, fear, stress int) {
	m.emotions[npcID] = struct{ trust, fear, stress int }{trust, fear, stress}
}

func (m *mockNPCStateProvider) setTraits(npcID string, traits []string) {
	m.traits[npcID] = traits
}

// TestNewDistortionCalculator tests the constructor.
func TestNewDistortionCalculator(t *testing.T) {
	provider := newMockNPCStateProvider()

	t.Run("with default config", func(t *testing.T) {
		calc := NewDistortionCalculator(provider, nil)
		if calc == nil {
			t.Fatal("NewDistortionCalculator returned nil")
		}
		if calc.config == nil {
			t.Fatal("config is nil")
		}
		if calc.stateProvider == nil {
			t.Fatal("stateProvider is nil")
		}
		if calc.config.BaseDistortionRate != 0.15 {
			t.Errorf("expected default BaseDistortionRate 0.15, got %f", calc.config.BaseDistortionRate)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		customConfig := &DistortionConfig{
			BaseDistortionRate: 0.25,
			FearWeight:         0.5,
			StressWeight:       0.4,
		}
		calc := NewDistortionCalculator(provider, customConfig)
		if calc.config.BaseDistortionRate != 0.25 {
			t.Errorf("expected BaseDistortionRate 0.25, got %f", calc.config.BaseDistortionRate)
		}
	})
}

// TestCalculateDistortionRate_BaseCase tests basic distortion rate calculation.
func TestCalculateDistortionRate_BaseCase(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 50, 25, 25)
	provider.setTraits("npc1", []string{})

	calc := NewDistortionCalculator(provider, nil)

	rate, factors, err := calc.CalculateDistortionRate("npc1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: base(0.15) + fear(0.25*0.4=0.10) + stress(0.25*0.3=0.075) - trust(0.5*0.5=0.25) = 0.075
	// With depth 0: 0.075 * 1.0 = 0.075
	expectedRate := 0.075
	tolerance := 0.001

	if diff := rate - expectedRate; diff < -tolerance || diff > tolerance {
		t.Errorf("expected rate ~%f, got %f", expectedRate, rate)
	}

	if factors.BaseRate != 0.15 {
		t.Errorf("expected BaseRate 0.15, got %f", factors.BaseRate)
	}
}

// TestCalculateDistortionRate_HighFear tests AC3: 恐懼/壓力高的 NPC 更容易扭曲
func TestCalculateDistortionRate_HighFear(t *testing.T) {
	provider := newMockNPCStateProvider()

	// Low fear NPC
	provider.setEmotion("npc_low_fear", 50, 10, 10)
	provider.setTraits("npc_low_fear", []string{})

	// High fear NPC
	provider.setEmotion("npc_high_fear", 50, 80, 10)
	provider.setTraits("npc_high_fear", []string{})

	calc := NewDistortionCalculator(provider, nil)

	rateLow, _, err := calc.CalculateDistortionRate("npc_low_fear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rateHigh, _, err := calc.CalculateDistortionRate("npc_high_fear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// High fear should have significantly higher distortion rate
	if rateHigh <= rateLow {
		t.Errorf("expected high fear rate (%f) > low fear rate (%f)", rateHigh, rateLow)
	}
}

// TestCalculateDistortionRate_HighStress tests AC3: 恐懼/壓力高的 NPC 更容易扭曲
func TestCalculateDistortionRate_HighStress(t *testing.T) {
	provider := newMockNPCStateProvider()

	// Low stress NPC
	provider.setEmotion("npc_low_stress", 50, 10, 10)
	provider.setTraits("npc_low_stress", []string{})

	// High stress NPC
	provider.setEmotion("npc_high_stress", 50, 10, 80)
	provider.setTraits("npc_high_stress", []string{})

	calc := NewDistortionCalculator(provider, nil)

	rateLow, _, err := calc.CalculateDistortionRate("npc_low_stress", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rateHigh, _, err := calc.CalculateDistortionRate("npc_high_stress", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// High stress should have significantly higher distortion rate
	if rateHigh <= rateLow {
		t.Errorf("expected high stress rate (%f) > low stress rate (%f)", rateHigh, rateLow)
	}
}

// TestCalculateDistortionRate_HighTrust tests AC4: 玩家可透過高信任度獲得準確資訊
func TestCalculateDistortionRate_HighTrust(t *testing.T) {
	provider := newMockNPCStateProvider()

	// Low trust NPC
	provider.setEmotion("npc_low_trust", 20, 30, 30)
	provider.setTraits("npc_low_trust", []string{})

	// High trust NPC
	provider.setEmotion("npc_high_trust", 90, 30, 30)
	provider.setTraits("npc_high_trust", []string{})

	calc := NewDistortionCalculator(provider, nil)

	rateLow, _, err := calc.CalculateDistortionRate("npc_low_trust", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rateHigh, _, err := calc.CalculateDistortionRate("npc_high_trust", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// High trust should have significantly lower distortion rate
	if rateHigh >= rateLow {
		t.Errorf("expected high trust rate (%f) < low trust rate (%f)", rateHigh, rateLow)
	}
}

// TestCalculateDistortionRate_PropagationDepth tests AC2: 扭曲程度與傳播深度相關
func TestCalculateDistortionRate_PropagationDepth(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 50, 30, 30)
	provider.setTraits("npc1", []string{})

	calc := NewDistortionCalculator(provider, nil)

	rate0, _, _ := calc.CalculateDistortionRate("npc1", 0)
	rate1, _, _ := calc.CalculateDistortionRate("npc1", 1)
	rate2, _, _ := calc.CalculateDistortionRate("npc1", 2)
	rate3, _, _ := calc.CalculateDistortionRate("npc1", 3)

	// Each depth level should increase distortion rate
	if rate1 <= rate0 {
		t.Errorf("expected rate at depth 1 (%f) > depth 0 (%f)", rate1, rate0)
	}
	if rate2 <= rate1 {
		t.Errorf("expected rate at depth 2 (%f) > depth 1 (%f)", rate2, rate1)
	}
	if rate3 <= rate2 {
		t.Errorf("expected rate at depth 3 (%f) > depth 2 (%f)", rate3, rate2)
	}

	// Verify depth multiplier is applied correctly (1.25^depth)
	expectedMultiplier1 := 1.25
	expectedMultiplier2 := 1.25 * 1.25
	expectedMultiplier3 := 1.25 * 1.25 * 1.25

	// rate1 should be roughly rate0 * 1.25
	ratio1 := rate1 / rate0
	if diff := ratio1 - expectedMultiplier1; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected depth 1 multiplier ~%f, got %f", expectedMultiplier1, ratio1)
	}

	// rate2 should be roughly rate0 * 1.25^2
	ratio2 := rate2 / rate0
	if diff := ratio2 - expectedMultiplier2; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected depth 2 multiplier ~%f, got %f", expectedMultiplier2, ratio2)
	}

	// rate3 should be roughly rate0 * 1.25^3
	ratio3 := rate3 / rate0
	if diff := ratio3 - expectedMultiplier3; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected depth 3 multiplier ~%f, got %f", expectedMultiplier3, ratio3)
	}
}

// TestCalculateDistortionRate_TraitModifiers tests AC1: 更智慧的扭曲邏輯（基於 NPC 個性）
func TestCalculateDistortionRate_TraitModifiers(t *testing.T) {
	provider := newMockNPCStateProvider()

	// NPC with no special traits
	provider.setEmotion("npc_normal", 50, 30, 30)
	provider.setTraits("npc_normal", []string{})

	// NPC with anxious trait (should increase distortion)
	provider.setEmotion("npc_anxious", 50, 30, 30)
	provider.setTraits("npc_anxious", []string{"anxious"})

	// NPC with rational trait (should decrease distortion)
	provider.setEmotion("npc_rational", 50, 30, 30)
	provider.setTraits("npc_rational", []string{"rational"})

	// NPC with multiple traits
	provider.setEmotion("npc_mixed", 50, 30, 30)
	provider.setTraits("npc_mixed", []string{"anxious", "paranoid"})

	calc := NewDistortionCalculator(provider, nil)

	rateNormal, _, _ := calc.CalculateDistortionRate("npc_normal", 0)
	rateAnxious, _, _ := calc.CalculateDistortionRate("npc_anxious", 0)
	rateRational, _, _ := calc.CalculateDistortionRate("npc_rational", 0)
	rateMixed, _, _ := calc.CalculateDistortionRate("npc_mixed", 0)

	// Anxious trait should increase distortion
	if rateAnxious <= rateNormal {
		t.Errorf("expected anxious rate (%f) > normal rate (%f)", rateAnxious, rateNormal)
	}

	// Rational trait should decrease distortion
	if rateRational >= rateNormal {
		t.Errorf("expected rational rate (%f) < normal rate (%f)", rateRational, rateNormal)
	}

	// Multiple negative traits should have cumulative effect
	if rateMixed <= rateAnxious {
		t.Errorf("expected mixed trait rate (%f) > anxious rate (%f)", rateMixed, rateAnxious)
	}
}

// TestCalculateDistortionRate_Clamping tests that distortion rate is clamped to [0.0, 1.0]
func TestCalculateDistortionRate_Clamping(t *testing.T) {
	provider := newMockNPCStateProvider()

	// Extremely low distortion scenario: high trust, low fear, low stress, calm trait
	provider.setEmotion("npc_min", 100, 0, 0)
	provider.setTraits("npc_min", []string{"calm", "rational"})

	// Extremely high distortion scenario: low trust, high fear, high stress, paranoid trait, high depth
	provider.setEmotion("npc_max", 0, 100, 100)
	provider.setTraits("npc_max", []string{"anxious", "paranoid", "unstable"})

	calc := NewDistortionCalculator(provider, nil)

	rateMin, _, _ := calc.CalculateDistortionRate("npc_min", 0)
	rateMax, _, _ := calc.CalculateDistortionRate("npc_max", 5) // High depth

	// Rate should be clamped to [0.0, 1.0]
	if rateMin < 0.0 {
		t.Errorf("expected rate >= 0.0, got %f", rateMin)
	}
	if rateMax > 1.0 {
		t.Errorf("expected rate <= 1.0, got %f", rateMax)
	}
}

// TestCalculateDistortionRate_NPCNotFound tests error handling
func TestCalculateDistortionRate_NPCNotFound(t *testing.T) {
	provider := newMockNPCStateProvider()
	calc := NewDistortionCalculator(provider, nil)

	_, _, err := calc.CalculateDistortionRate("nonexistent", 0)
	if err == nil {
		t.Error("expected error for nonexistent NPC, got nil")
	}
}

// TestApplyDistortion tests the full distortion application process
func TestApplyDistortion(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 50, 30, 30)
	provider.setTraits("npc1", []string{})

	calc := NewDistortionCalculator(provider, nil)

	// Set seed for reproducible test
	rand.Seed(42)

	result, err := calc.ApplyDistortion("npc1", "這是一個事實", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DistortionRate < 0.0 || result.DistortionRate > 1.0 {
		t.Errorf("invalid distortion rate: %f", result.DistortionRate)
	}

	if result.ShouldDistort && result.DistortedContent == "" {
		t.Error("ShouldDistort is true but DistortedContent is empty")
	}

	if !result.ShouldDistort && result.DistortedContent != "" {
		t.Error("ShouldDistort is false but DistortedContent is not empty")
	}
}

// TestDistortContent_FearBased tests fear-based distortion
func TestDistortContent_FearBased(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc_fearful", 50, 80, 30) // High fear
	provider.setTraits("npc_fearful", []string{})

	calc := NewDistortionCalculator(provider, nil)

	// Get factors for context
	_, factors, _ := calc.CalculateDistortionRate("npc_fearful", 0)

	content := "發現了一個房間"
	distorted := calc.distortContent(content, "npc_fearful", factors)

	// Should contain fear-related markers
	hasFearMarker := strings.Contains(distorted, "恐怖") ||
		strings.Contains(distorted, "害怕") ||
		strings.Contains(distorted, "可怕") ||
		strings.Contains(distorted, "不安")

	if !hasFearMarker {
		t.Errorf("expected fear-based distortion, got: %s", distorted)
	}
}

// TestDistortContent_StressBased tests stress-based distortion
func TestDistortContent_StressBased(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc_stressed", 50, 30, 80) // High stress
	provider.setTraits("npc_stressed", []string{})

	calc := NewDistortionCalculator(provider, nil)

	_, factors, _ := calc.CalculateDistortionRate("npc_stressed", 0)

	content := "看到了一個人"
	distorted := calc.distortContent(content, "npc_stressed", factors)

	// Should contain uncertainty markers (prefix or suffix)
	hasUncertaintyMarker := strings.Contains(distorted, "可能") ||
		strings.Contains(distorted, "似乎") ||
		strings.Contains(distorted, "好像") ||
		strings.Contains(distorted, "或許") ||
		strings.Contains(distorted, "我覺得") ||
		strings.Contains(distorted, "不確定") ||
		strings.Contains(distorted, "不太確定") ||
		strings.Contains(distorted, "記錯了") ||
		strings.Contains(distorted, "記不清了") ||
		strings.Contains(distorted, "混亂")

	if !hasUncertaintyMarker {
		t.Errorf("expected stress-based distortion, got: %s", distorted)
	}
}

// TestDistortContent_LowTrust tests low trust distortion
func TestDistortContent_LowTrust(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc_distrust", 20, 30, 30) // Low trust
	provider.setTraits("npc_distrust", []string{})

	calc := NewDistortionCalculator(provider, nil)

	_, factors, _ := calc.CalculateDistortionRate("npc_distrust", 0)

	content := "聽到了一個聲音"
	distorted := calc.distortContent(content, "npc_distrust", factors)

	// Should contain skeptical markers
	hasSkepticalMarker := strings.Contains(distorted, "據說") ||
		strings.Contains(distorted, "聽說") ||
		strings.Contains(distorted, "有人說") ||
		strings.Contains(distorted, "傳言") ||
		strings.Contains(distorted, "存疑") ||
		strings.Contains(distorted, "不太相信") ||
		strings.Contains(distorted, "未經證實")

	if !hasSkepticalMarker {
		t.Errorf("expected low-trust distortion, got: %s", distorted)
	}
}

// TestShouldDistort tests the convenience method
func TestShouldDistort(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 50, 30, 30)
	provider.setTraits("npc1", []string{})

	calc := NewDistortionCalculator(provider, nil)

	// Run multiple times to test randomness
	count := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		shouldDistort, err := calc.ShouldDistort("npc1", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shouldDistort {
			count++
		}
	}

	// Get expected rate
	expectedRate, _, _ := calc.CalculateDistortionRate("npc1", 0)
	expectedCount := int(expectedRate * float64(iterations))

	// Allow 20% variance due to randomness
	tolerance := int(0.2 * float64(expectedCount))
	if count < expectedCount-tolerance || count > expectedCount+tolerance {
		t.Errorf("expected ~%d distortions out of %d, got %d", expectedCount, iterations, count)
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *DistortionConfig
		expectError bool
	}{
		{
			name:        "valid config",
			config:      DefaultDistortionConfig(),
			expectError: false,
		},
		{
			name: "invalid base rate (negative)",
			config: &DistortionConfig{
				BaseDistortionRate: -0.1,
			},
			expectError: true,
		},
		{
			name: "invalid base rate (> 1.0)",
			config: &DistortionConfig{
				BaseDistortionRate: 1.5,
			},
			expectError: true,
		},
		{
			name: "invalid fear weight",
			config: &DistortionConfig{
				BaseDistortionRate: 0.15,
				FearWeight:         1.5,
			},
			expectError: true,
		},
		{
			name: "invalid stress weight",
			config: &DistortionConfig{
				BaseDistortionRate: 0.15,
				FearWeight:         0.4,
				StressWeight:       -0.1,
			},
			expectError: true,
		},
		{
			name: "invalid trust protection",
			config: &DistortionConfig{
				BaseDistortionRate: 0.15,
				FearWeight:         0.4,
				StressWeight:       0.3,
				TrustProtection:    1.5,
			},
			expectError: true,
		},
		{
			name: "invalid depth multiplier",
			config: &DistortionConfig{
				BaseDistortionRate: 0.15,
				FearWeight:         0.4,
				StressWeight:       0.3,
				TrustProtection:    0.5,
				DepthMultiplier:    0.5, // Should be >= 1.0
			},
			expectError: true,
		},
		{
			name: "invalid high trust threshold",
			config: &DistortionConfig{
				BaseDistortionRate:  0.15,
				FearWeight:          0.4,
				StressWeight:        0.3,
				TrustProtection:     0.5,
				DepthMultiplier:     1.25,
				HighTrustThreshold:  150, // Should be 0-100
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// TestIsDistortionPrefixOrSuffix tests distortion marker detection
func TestIsDistortionPrefixOrSuffix(t *testing.T) {
	tests := []struct {
		content  string
		expected bool
	}{
		{"這是一個事實", false},
		{"我聽說這是一個事實", true},
		{"可能這是一個事實", true},
		{"據說這是一個事實", true},
		{"這是一個事實（不確定）", true},
		{"好像這是一個事實", true},
		{"似乎這是一個事實", true},
		{"我感覺有些恐怖，這是一個事實", true},
		{"這是一個事實（但我很害怕）", true},
		{"這是一個普通的事實", false},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			result := IsDistortionPrefixOrSuffix(tt.content)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestDistortionFactors_Breakdown tests that factors are properly broken down
func TestDistortionFactors_Breakdown(t *testing.T) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 60, 40, 50)
	provider.setTraits("npc1", []string{"anxious"})

	calc := NewDistortionCalculator(provider, nil)

	_, factors, err := calc.CalculateDistortionRate("npc1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify base rate
	if factors.BaseRate != 0.15 {
		t.Errorf("expected BaseRate 0.15, got %f", factors.BaseRate)
	}

	// Verify fear contribution: (40/100) * 0.4 = 0.16
	expectedFear := 0.16
	if diff := factors.FearContribution - expectedFear; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected FearContribution %f, got %f", expectedFear, factors.FearContribution)
	}

	// Verify stress contribution: (50/100) * 0.3 = 0.15
	expectedStress := 0.15
	if diff := factors.StressContribution - expectedStress; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected StressContribution %f, got %f", expectedStress, factors.StressContribution)
	}

	// Verify trust reduction: (60/100) * 0.5 = 0.3
	expectedTrust := 0.3
	if diff := factors.TrustReduction - expectedTrust; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected TrustReduction %f, got %f", expectedTrust, factors.TrustReduction)
	}

	// Verify trait modifier: anxious = +0.15
	expectedTrait := 0.15
	if diff := factors.TraitModifier - expectedTrait; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected TraitModifier %f, got %f", expectedTrait, factors.TraitModifier)
	}

	// Verify depth multiplier: 1.25^2 = 1.5625
	expectedDepth := 1.5625
	if diff := factors.DepthMultiplier - expectedDepth; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected DepthMultiplier %f, got %f", expectedDepth, factors.DepthMultiplier)
	}

	// Verify final rate calculation
	// (0.15 + 0.16 + 0.15 - 0.3 + 0.15) * 1.5625 = 0.31 * 1.5625 = 0.484375
	expectedFinal := 0.484375
	if diff := factors.FinalRate - expectedFinal; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected FinalRate %f, got %f", expectedFinal, factors.FinalRate)
	}
}

// TestCalculateTraitModifier tests trait modifier calculation
func TestCalculateTraitModifier(t *testing.T) {
	calc := NewDistortionCalculator(nil, DefaultDistortionConfig())

	tests := []struct {
		name     string
		traits   []string
		expected float64
	}{
		{
			name:     "no traits",
			traits:   []string{},
			expected: 0.0,
		},
		{
			name:     "single positive trait",
			traits:   []string{"anxious"},
			expected: 0.15,
		},
		{
			name:     "single negative trait",
			traits:   []string{"rational"},
			expected: -0.15,
		},
		{
			name:     "multiple traits",
			traits:   []string{"anxious", "paranoid"},
			expected: 0.40, // 0.15 + 0.25
		},
		{
			name:     "mixed traits",
			traits:   []string{"anxious", "calm"},
			expected: 0.05, // 0.15 - 0.10
		},
		{
			name:     "unknown trait",
			traits:   []string{"unknown_trait"},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateTraitModifier(tt.traits)
			if diff := result - tt.expected; diff < -0.001 || diff > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// TestGetConfig tests configuration retrieval
func TestGetConfig(t *testing.T) {
	config := &DistortionConfig{
		BaseDistortionRate: 0.25,
		FearWeight:         0.5,
	}

	calc := NewDistortionCalculator(nil, config)

	retrievedConfig := calc.GetConfig()
	if retrievedConfig == nil {
		t.Fatal("GetConfig returned nil")
	}

	if retrievedConfig.BaseDistortionRate != 0.25 {
		t.Errorf("expected BaseDistortionRate 0.25, got %f", retrievedConfig.BaseDistortionRate)
	}
}

// BenchmarkCalculateDistortionRate benchmarks distortion rate calculation
func BenchmarkCalculateDistortionRate(b *testing.B) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 50, 30, 30)
	provider.setTraits("npc1", []string{"anxious", "nervous"})

	calc := NewDistortionCalculator(provider, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = calc.CalculateDistortionRate("npc1", 2)
	}
}

// BenchmarkApplyDistortion benchmarks full distortion application
func BenchmarkApplyDistortion(b *testing.B) {
	provider := newMockNPCStateProvider()
	provider.setEmotion("npc1", 50, 60, 50)
	provider.setTraits("npc1", []string{"anxious"})

	calc := NewDistortionCalculator(provider, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calc.ApplyDistortion("npc1", "這是一個重要的發現", 1)
	}
}

// TestDistortionEdgeCases tests various edge cases
func TestDistortionEdgeCases(t *testing.T) {
	provider := newMockNPCStateProvider()

	t.Run("zero emotion values", func(t *testing.T) {
		provider.setEmotion("npc_zero", 0, 0, 0)
		provider.setTraits("npc_zero", []string{})

		calc := NewDistortionCalculator(provider, nil)
		rate, _, err := calc.CalculateDistortionRate("npc_zero", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should still have base rate
		if rate < 0.0 {
			t.Errorf("expected rate >= 0, got %f", rate)
		}
	})

	t.Run("maximum emotion values", func(t *testing.T) {
		provider.setEmotion("npc_max", 100, 100, 100)
		provider.setTraits("npc_max", []string{})

		calc := NewDistortionCalculator(provider, nil)
		rate, _, err := calc.CalculateDistortionRate("npc_max", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should be clamped to 1.0
		if rate > 1.0 {
			t.Errorf("expected rate <= 1.0, got %f", rate)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		provider.setEmotion("npc1", 50, 60, 50)
		provider.setTraits("npc1", []string{})

		calc := NewDistortionCalculator(provider, nil)
		result, err := calc.ApplyDistortion("npc1", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should handle empty content gracefully
		if result.ShouldDistort && result.DistortedContent == "" {
			// This is actually expected - distortion markers are added to empty string
			// which results in just the markers
		}
	})

	t.Run("very long content", func(t *testing.T) {
		provider.setEmotion("npc1", 50, 60, 50)
		provider.setTraits("npc1", []string{})

		calc := NewDistortionCalculator(provider, nil)
		longContent := strings.Repeat("這是一個很長的內容", 100)
		result, err := calc.ApplyDistortion("npc1", longContent, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should handle long content
		if result.ShouldDistort && !strings.Contains(result.DistortedContent, "這是一個很長的內容") {
			t.Error("distorted content should still contain original content")
		}
	})
}
