package manager

// NPCManagerConfig contains configuration parameters for the NPCManager.
// These parameters control emotional decay rates, thresholds for special behaviors,
// and other gameplay mechanics related to NPCs.
type NPCManagerConfig struct {
	// TrustDecayRate is the amount of trust that decays per turn.
	// Default: 0.5 (trust decreases by 0.5 each turn)
	TrustDecayRate float64

	// FearDecayRate is the amount of fear that decays per turn.
	// Default: 1.0 (fear decreases by 1.0 each turn)
	FearDecayRate float64

	// StressDecayRate is the amount of stress that decays per turn.
	// Default: 0.5 (stress decreases by 0.5 each turn)
	StressDecayRate float64

	// BreakdownThreshold is the stress level at which NPCs may experience breakdown.
	// When stress >= this value, special breakdown checks are triggered.
	// Default: 80
	BreakdownThreshold int

	// MinTrustForSecret is the minimum trust level required for an NPC to share secrets.
	// Default: 75 (trust must be >= 75 for secret sharing)
	MinTrustForSecret int

	// HintDuration is the number of turns a hint remains active before being revealed.
	// Default: 3 (hints persist for 3 turns)
	HintDuration int
}

// DefaultNPCManagerConfig returns an NPCManagerConfig with sensible default values.
// These defaults are tuned for typical horror game pacing.
func DefaultNPCManagerConfig() *NPCManagerConfig {
	return &NPCManagerConfig{
		TrustDecayRate:     0.5,  // Gradual trust decay
		FearDecayRate:      1.0,  // Faster fear decay (fear is more volatile)
		StressDecayRate:    0.5,  // Gradual stress decay
		BreakdownThreshold: 80,   // High stress threshold before breakdown
		MinTrustForSecret:  75,   // High trust required for secrets
		HintDuration:       3,    // Hints last 3 turns
	}
}

// Validate checks if the configuration values are valid.
// Returns an error if any values are out of acceptable ranges.
func (c *NPCManagerConfig) Validate() error {
	// All decay rates should be non-negative
	if c.TrustDecayRate < 0 {
		return &ConfigError{Field: "TrustDecayRate", Reason: "must be non-negative"}
	}
	if c.FearDecayRate < 0 {
		return &ConfigError{Field: "FearDecayRate", Reason: "must be non-negative"}
	}
	if c.StressDecayRate < 0 {
		return &ConfigError{Field: "StressDecayRate", Reason: "must be non-negative"}
	}

	// Threshold values should be in 0-100 range
	if c.BreakdownThreshold < 0 || c.BreakdownThreshold > 100 {
		return &ConfigError{Field: "BreakdownThreshold", Reason: "must be between 0 and 100"}
	}
	if c.MinTrustForSecret < 0 || c.MinTrustForSecret > 100 {
		return &ConfigError{Field: "MinTrustForSecret", Reason: "must be between 0 and 100"}
	}

	// HintDuration should be positive
	if c.HintDuration <= 0 {
		return &ConfigError{Field: "HintDuration", Reason: "must be positive"}
	}

	return nil
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field  string
	Reason string
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	return "config error: " + e.Field + " " + e.Reason
}
