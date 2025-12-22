package guardian

// GuardianConfig configures the Experience Guardian protection thresholds.
type GuardianConfig struct {
	// MaxConsecutiveDeaths is the number of consecutive deaths before activating protection.
	// Default: 2
	MaxConsecutiveDeaths int

	// LowStatStreakLimit is the number of consecutive turns with low HP/SAN before activating protection.
	// Default: 3
	LowStatStreakLimit int

	// LowHPThreshold is the HP value considered "low" for protection purposes.
	// Default: 20
	LowHPThreshold int

	// LowSANThreshold is the SAN value considered "low" for protection purposes.
	// Default: 30
	LowSANThreshold int

	// EnableDifficultyTuning enables difficulty adjustment based on player survival state.
	// Default: false
	EnableDifficultyTuning bool
}

// DefaultGuardianConfig returns the default Guardian configuration.
func DefaultGuardianConfig() GuardianConfig {
	return GuardianConfig{
		MaxConsecutiveDeaths:   2,
		LowStatStreakLimit:     3,
		LowHPThreshold:         20,
		LowSANThreshold:        30,
		EnableDifficultyTuning: false,
	}
}
