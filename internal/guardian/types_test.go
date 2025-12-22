package guardian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGuardianConfig_CustomValues tests GuardianConfig with custom values
func TestGuardianConfig_CustomValues(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths:   5,
		LowStatStreakLimit:     7,
		LowHPThreshold:         15,
		LowSANThreshold:        25,
		EnableDifficultyTuning: true,
	}

	assert.Equal(t, 5, cfg.MaxConsecutiveDeaths, "MaxConsecutiveDeaths should be 5")
	assert.Equal(t, 7, cfg.LowStatStreakLimit, "LowStatStreakLimit should be 7")
	assert.Equal(t, 15, cfg.LowHPThreshold, "LowHPThreshold should be 15")
	assert.Equal(t, 25, cfg.LowSANThreshold, "LowSANThreshold should be 25")
	assert.True(t, cfg.EnableDifficultyTuning, "EnableDifficultyTuning should be true")
}

// TestGuardianConfig_ZeroValues tests GuardianConfig with zero values
func TestGuardianConfig_ZeroValues(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths:   0,
		LowStatStreakLimit:     0,
		LowHPThreshold:         0,
		LowSANThreshold:        0,
		EnableDifficultyTuning: false,
	}

	assert.Equal(t, 0, cfg.MaxConsecutiveDeaths, "MaxConsecutiveDeaths should be 0")
	assert.Equal(t, 0, cfg.LowStatStreakLimit, "LowStatStreakLimit should be 0")
	assert.Equal(t, 0, cfg.LowHPThreshold, "LowHPThreshold should be 0")
	assert.Equal(t, 0, cfg.LowSANThreshold, "LowSANThreshold should be 0")
	assert.False(t, cfg.EnableDifficultyTuning, "EnableDifficultyTuning should be false")
}

// TestGuardianConfig_DefaultValues tests DefaultGuardianConfig
func TestGuardianConfig_DefaultValues(t *testing.T) {
	cfg := DefaultGuardianConfig()

	assert.Equal(t, 2, cfg.MaxConsecutiveDeaths, "Default MaxConsecutiveDeaths should be 2")
	assert.Equal(t, 3, cfg.LowStatStreakLimit, "Default LowStatStreakLimit should be 3")
	assert.Equal(t, 20, cfg.LowHPThreshold, "Default LowHPThreshold should be 20")
	assert.Equal(t, 30, cfg.LowSANThreshold, "Default LowSANThreshold should be 30")
	assert.False(t, cfg.EnableDifficultyTuning, "Default EnableDifficultyTuning should be false")
}

// TestGuardianConfig_HighThresholds tests GuardianConfig with high thresholds
func TestGuardianConfig_HighThresholds(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths:   10,
		LowStatStreakLimit:     15,
		LowHPThreshold:         50,
		LowSANThreshold:        60,
		EnableDifficultyTuning: true,
	}

	assert.Equal(t, 10, cfg.MaxConsecutiveDeaths, "MaxConsecutiveDeaths should be 10")
	assert.Equal(t, 15, cfg.LowStatStreakLimit, "LowStatStreakLimit should be 15")
	assert.Equal(t, 50, cfg.LowHPThreshold, "LowHPThreshold should be 50")
	assert.Equal(t, 60, cfg.LowSANThreshold, "LowSANThreshold should be 60")
	assert.True(t, cfg.EnableDifficultyTuning, "EnableDifficultyTuning should be true")
}

// TestGuardianConfig_RealisticScenario tests realistic game configuration
func TestGuardianConfig_RealisticScenario(t *testing.T) {
	// Easy mode configuration
	easyConfig := GuardianConfig{
		MaxConsecutiveDeaths:   1, // Protect after 1 death
		LowStatStreakLimit:     2, // Protect after 2 low stat turns
		LowHPThreshold:         40, // Higher threshold
		LowSANThreshold:        50, // Higher threshold
		EnableDifficultyTuning: true,
	}

	assert.Equal(t, 1, easyConfig.MaxConsecutiveDeaths)
	assert.Equal(t, 2, easyConfig.LowStatStreakLimit)
	assert.Equal(t, 40, easyConfig.LowHPThreshold)
	assert.Equal(t, 50, easyConfig.LowSANThreshold)
	assert.True(t, easyConfig.EnableDifficultyTuning)

	// Hard mode configuration
	hardConfig := GuardianConfig{
		MaxConsecutiveDeaths:   5,  // Protect after 5 deaths
		LowStatStreakLimit:     10, // Protect after 10 low stat turns
		LowHPThreshold:         10, // Lower threshold
		LowSANThreshold:        15, // Lower threshold
		EnableDifficultyTuning: false,
	}

	assert.Equal(t, 5, hardConfig.MaxConsecutiveDeaths)
	assert.Equal(t, 10, hardConfig.LowStatStreakLimit)
	assert.Equal(t, 10, hardConfig.LowHPThreshold)
	assert.Equal(t, 15, hardConfig.LowSANThreshold)
	assert.False(t, hardConfig.EnableDifficultyTuning)
}

// TestGuardianConfig_ExtremeValues tests extreme configuration values
func TestGuardianConfig_ExtremeValues(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths:   1000,
		LowStatStreakLimit:     1000,
		LowHPThreshold:         100,
		LowSANThreshold:        100,
		EnableDifficultyTuning: true,
	}

	assert.Equal(t, 1000, cfg.MaxConsecutiveDeaths)
	assert.Equal(t, 1000, cfg.LowStatStreakLimit)
	assert.Equal(t, 100, cfg.LowHPThreshold)
	assert.Equal(t, 100, cfg.LowSANThreshold)
	assert.True(t, cfg.EnableDifficultyTuning)
}

// TestGuardianConfig_AsymmetricThresholds tests asymmetric HP/SAN thresholds
func TestGuardianConfig_AsymmetricThresholds(t *testing.T) {
	// HP focused protection
	hpFocused := GuardianConfig{
		MaxConsecutiveDeaths:   2,
		LowStatStreakLimit:     3,
		LowHPThreshold:         40, // High HP threshold
		LowSANThreshold:        15, // Low SAN threshold
		EnableDifficultyTuning: false,
	}

	assert.Greater(t, hpFocused.LowHPThreshold, hpFocused.LowSANThreshold, "HP threshold should be higher")

	// SAN focused protection
	sanFocused := GuardianConfig{
		MaxConsecutiveDeaths:   2,
		LowStatStreakLimit:     3,
		LowHPThreshold:         15, // Low HP threshold
		LowSANThreshold:        40, // High SAN threshold
		EnableDifficultyTuning: false,
	}

	assert.Greater(t, sanFocused.LowSANThreshold, sanFocused.LowHPThreshold, "SAN threshold should be higher")
}

// TestGuardianConfig_MinimalProtection tests minimal protection configuration
func TestGuardianConfig_MinimalProtection(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths:   100,
		LowStatStreakLimit:     100,
		LowHPThreshold:         1,
		LowSANThreshold:        1,
		EnableDifficultyTuning: false,
	}

	assert.Equal(t, 100, cfg.MaxConsecutiveDeaths, "Very high death threshold")
	assert.Equal(t, 100, cfg.LowStatStreakLimit, "Very high streak threshold")
	assert.Equal(t, 1, cfg.LowHPThreshold, "Very low HP threshold")
	assert.Equal(t, 1, cfg.LowSANThreshold, "Very low SAN threshold")
	assert.False(t, cfg.EnableDifficultyTuning)
}

// TestGuardianConfig_MaximalProtection tests maximal protection configuration
func TestGuardianConfig_MaximalProtection(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths:   1,
		LowStatStreakLimit:     1,
		LowHPThreshold:         90,
		LowSANThreshold:        90,
		EnableDifficultyTuning: true,
	}

	assert.Equal(t, 1, cfg.MaxConsecutiveDeaths, "Minimal death tolerance")
	assert.Equal(t, 1, cfg.LowStatStreakLimit, "Minimal streak tolerance")
	assert.Equal(t, 90, cfg.LowHPThreshold, "Very high HP threshold")
	assert.Equal(t, 90, cfg.LowSANThreshold, "Very high SAN threshold")
	assert.True(t, cfg.EnableDifficultyTuning)
}
