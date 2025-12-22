package manager

import "testing"

// TestDefaultNPCManagerConfig tests the default configuration values
func TestDefaultNPCManagerConfig(t *testing.T) {
	config := DefaultNPCManagerConfig()

	if config == nil {
		t.Fatal("DefaultNPCManagerConfig() returned nil")
	}

	// Verify default values
	tests := []struct {
		name     string
		got      interface{}
		want     interface{}
	}{
		{"TrustDecayRate", config.TrustDecayRate, 0.5},
		{"FearDecayRate", config.FearDecayRate, 1.0},
		{"StressDecayRate", config.StressDecayRate, 0.5},
		{"BreakdownThreshold", config.BreakdownThreshold, 80},
		{"MinTrustForSecret", config.MinTrustForSecret, 75},
		{"HintDuration", config.HintDuration, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestNPCManagerConfig_Validate tests configuration validation
func TestNPCManagerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *NPCManagerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			config:  DefaultNPCManagerConfig(),
			wantErr: false,
		},
		{
			name: "valid custom config",
			config: &NPCManagerConfig{
				TrustDecayRate:     1.0,
				FearDecayRate:      2.0,
				StressDecayRate:    1.5,
				BreakdownThreshold: 90,
				MinTrustForSecret:  80,
				HintDuration:       5,
			},
			wantErr: false,
		},
		{
			name: "negative TrustDecayRate",
			config: &NPCManagerConfig{
				TrustDecayRate:     -0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  75,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "TrustDecayRate",
		},
		{
			name: "negative FearDecayRate",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      -1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  75,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "FearDecayRate",
		},
		{
			name: "negative StressDecayRate",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    -0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  75,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "StressDecayRate",
		},
		{
			name: "BreakdownThreshold too low",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: -1,
				MinTrustForSecret:  75,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "BreakdownThreshold",
		},
		{
			name: "BreakdownThreshold too high",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 101,
				MinTrustForSecret:  75,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "BreakdownThreshold",
		},
		{
			name: "MinTrustForSecret too low",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  -1,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "MinTrustForSecret",
		},
		{
			name: "MinTrustForSecret too high",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  101,
				HintDuration:       3,
			},
			wantErr: true,
			errMsg:  "MinTrustForSecret",
		},
		{
			name: "zero HintDuration",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  75,
				HintDuration:       0,
			},
			wantErr: true,
			errMsg:  "HintDuration",
		},
		{
			name: "negative HintDuration",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.5,
				FearDecayRate:      1.0,
				StressDecayRate:    0.5,
				BreakdownThreshold: 80,
				MinTrustForSecret:  75,
				HintDuration:       -1,
			},
			wantErr: true,
			errMsg:  "HintDuration",
		},
		{
			name: "zero decay rates (valid edge case)",
			config: &NPCManagerConfig{
				TrustDecayRate:     0.0,
				FearDecayRate:      0.0,
				StressDecayRate:    0.0,
				BreakdownThreshold: 80,
				MinTrustForSecret:  75,
				HintDuration:       3,
			},
			wantErr: false,
		},
		{
			name: "boundary values (valid)",
			config: &NPCManagerConfig{
				TrustDecayRate:     100.0,
				FearDecayRate:      100.0,
				StressDecayRate:    100.0,
				BreakdownThreshold: 0,
				MinTrustForSecret:  0,
				HintDuration:       1,
			},
			wantErr: false,
		},
		{
			name: "max boundary values (valid)",
			config: &NPCManagerConfig{
				TrustDecayRate:     999.9,
				FearDecayRate:      999.9,
				StressDecayRate:    999.9,
				BreakdownThreshold: 100,
				MinTrustForSecret:  100,
				HintDuration:       1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				// Check if error message contains expected field name
				if tt.errMsg != "" {
					configErr, ok := err.(*ConfigError)
					if !ok {
						t.Errorf("expected ConfigError, got %T", err)
						return
					}
					if configErr.Field != tt.errMsg {
						t.Errorf("expected error for field %s, got %s", tt.errMsg, configErr.Field)
					}
				}
			}
		})
	}
}

// TestConfigError tests the ConfigError type
func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:  "TestField",
		Reason: "is invalid",
	}

	expectedMsg := "config error: TestField is invalid"
	if err.Error() != expectedMsg {
		t.Errorf("ConfigError.Error() = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestConfigError_ErrorInterface tests that ConfigError implements error interface
func TestConfigError_ErrorInterface(t *testing.T) {
	var err error = &ConfigError{
		Field:  "TestField",
		Reason: "test reason",
	}

	if err.Error() == "" {
		t.Error("ConfigError should implement error interface")
	}
}

// BenchmarkValidate benchmarks the Validate method
func BenchmarkValidate(b *testing.B) {
	config := DefaultNPCManagerConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}

// BenchmarkValidate_Invalid benchmarks validation with invalid config
func BenchmarkValidate_Invalid(b *testing.B) {
	config := &NPCManagerConfig{
		TrustDecayRate:     -1.0, // Invalid
		FearDecayRate:      1.0,
		StressDecayRate:    0.5,
		BreakdownThreshold: 80,
		MinTrustForSecret:  75,
		HintDuration:       3,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}
