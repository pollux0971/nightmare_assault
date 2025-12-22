package knowledge

import (
	"testing"
)

// TestNewUpdateManager_DefaultConfig tests that NewUpdateManager properly initializes
// with default configuration when nil is passed.
// Verifies AC2: NewUpdateManager() 正確初始化所有欄位
func TestNewUpdateManager_DefaultConfig(t *testing.T) {
	mgr := NewUpdateManager(nil)

	if mgr == nil {
		t.Fatal("NewUpdateManager returned nil")
	}

	// Verify all fields are initialized
	if mgr.globalFacts == nil {
		t.Error("globalFacts map not initialized")
	}

	if mgr.npcKnowledge == nil {
		t.Error("npcKnowledge map not initialized")
	}

	if mgr.playerKnowledge == nil {
		t.Error("playerKnowledge not initialized")
	}

	if mgr.roomOccupants == nil {
		t.Error("roomOccupants map not initialized")
	}

	if mgr.config == nil {
		t.Error("config not initialized")
	}

	// Verify config has default values
	if !mgr.config.EnableDistortion {
		t.Error("expected EnableDistortion to be true by default")
	}

	if mgr.config.DistortionRate != 0.15 {
		t.Errorf("expected DistortionRate to be 0.15, got %f", mgr.config.DistortionRate)
	}

	if mgr.config.MaxPropagationDepth != 3 {
		t.Errorf("expected MaxPropagationDepth to be 3, got %d", mgr.config.MaxPropagationDepth)
	}
}

// TestNewUpdateManager_CustomConfig tests that NewUpdateManager properly uses
// provided custom configuration.
// Verifies AC2: NewUpdateManager() 正確初始化所有欄位
func TestNewUpdateManager_CustomConfig(t *testing.T) {
	customConfig := &UpdateManagerConfig{
		EnableDistortion:    false,
		DistortionRate:      0.25,
		MaxPropagationDepth: 5,
	}

	mgr := NewUpdateManager(customConfig)

	if mgr == nil {
		t.Fatal("NewUpdateManager returned nil")
	}

	// Verify config matches custom values
	if mgr.config.EnableDistortion {
		t.Error("expected EnableDistortion to be false")
	}

	if mgr.config.DistortionRate != 0.25 {
		t.Errorf("expected DistortionRate to be 0.25, got %f", mgr.config.DistortionRate)
	}

	if mgr.config.MaxPropagationDepth != 5 {
		t.Errorf("expected MaxPropagationDepth to be 5, got %d", mgr.config.MaxPropagationDepth)
	}
}

// TestUpdateManager_FieldsPresent tests that UpdateManager has all required fields.
// Verifies AC1: UpdateManager 類別具備 globalFacts、npcKnowledge、playerKnowledge、roomOccupants
func TestUpdateManager_FieldsPresent(t *testing.T) {
	mgr := NewUpdateManager(nil)

	// Test that we can access all required fields
	// (compilation will fail if fields don't exist)

	// globalFacts should be a map
	if mgr.globalFacts == nil {
		t.Error("globalFacts field is nil")
	}
	if len(mgr.globalFacts) != 0 {
		t.Error("globalFacts should be empty initially")
	}

	// npcKnowledge should be a map
	if mgr.npcKnowledge == nil {
		t.Error("npcKnowledge field is nil")
	}
	if len(mgr.npcKnowledge) != 0 {
		t.Error("npcKnowledge should be empty initially")
	}

	// playerKnowledge should be initialized
	if mgr.playerKnowledge == nil {
		t.Error("playerKnowledge field is nil")
	}

	// roomOccupants should be a map
	if mgr.roomOccupants == nil {
		t.Error("roomOccupants field is nil")
	}
	if len(mgr.roomOccupants) != 0 {
		t.Error("roomOccupants should be empty initially")
	}
}

// TestUpdateManagerConfig_AllFields tests that UpdateManagerConfig has all required fields.
// Verifies AC3: UpdateManagerConfig 包含 EnableDistortion/DistortionRate/MaxPropagationDepth
func TestUpdateManagerConfig_AllFields(t *testing.T) {
	config := &UpdateManagerConfig{
		EnableDistortion:    true,
		DistortionRate:      0.2,
		MaxPropagationDepth: 4,
	}

	// Verify all fields are accessible
	if !config.EnableDistortion {
		t.Error("EnableDistortion field not working")
	}

	if config.DistortionRate != 0.2 {
		t.Error("DistortionRate field not working")
	}

	if config.MaxPropagationDepth != 4 {
		t.Error("MaxPropagationDepth field not working")
	}
}

// TestDefaultUpdateManagerConfig tests that default config has sensible values.
// Verifies AC3: UpdateManagerConfig 包含 EnableDistortion/DistortionRate/MaxPropagationDepth
func TestDefaultUpdateManagerConfig(t *testing.T) {
	config := DefaultUpdateManagerConfig()

	if config == nil {
		t.Fatal("DefaultUpdateManagerConfig returned nil")
	}

	// Test EnableDistortion
	if !config.EnableDistortion {
		t.Error("expected EnableDistortion to be true by default")
	}

	// Test DistortionRate (should be 0.15)
	if config.DistortionRate != 0.15 {
		t.Errorf("expected DistortionRate to be 0.15, got %f", config.DistortionRate)
	}

	// Test MaxPropagationDepth (should be 3)
	if config.MaxPropagationDepth != 3 {
		t.Errorf("expected MaxPropagationDepth to be 3, got %d", config.MaxPropagationDepth)
	}
}

// TestUpdateManager_GetConfig tests that GetConfig returns the correct config.
func TestUpdateManager_GetConfig(t *testing.T) {
	customConfig := &UpdateManagerConfig{
		EnableDistortion:    false,
		DistortionRate:      0.5,
		MaxPropagationDepth: 10,
	}

	mgr := NewUpdateManager(customConfig)
	retrievedConfig := mgr.GetConfig()

	if retrievedConfig != customConfig {
		t.Error("GetConfig did not return the same config instance")
	}

	if retrievedConfig.EnableDistortion {
		t.Error("retrieved config has wrong EnableDistortion value")
	}

	if retrievedConfig.DistortionRate != 0.5 {
		t.Errorf("retrieved config has wrong DistortionRate: %f", retrievedConfig.DistortionRate)
	}

	if retrievedConfig.MaxPropagationDepth != 10 {
		t.Errorf("retrieved config has wrong MaxPropagationDepth: %d", retrievedConfig.MaxPropagationDepth)
	}
}

// TestUpdateManager_MapInitialization tests that all maps are properly initialized
// and can be used without panicking.
func TestUpdateManager_MapInitialization(t *testing.T) {
	mgr := NewUpdateManager(nil)

	// Test that we can add to globalFacts without panic
	fact := &Fact{}
	mgr.globalFacts["test-fact-1"] = fact
	if len(mgr.globalFacts) != 1 {
		t.Error("globalFacts map not working correctly")
	}

	// Test that we can add to npcKnowledge without panic
	kb := &KnowledgeBase{}
	mgr.npcKnowledge["npc-1"] = kb
	if len(mgr.npcKnowledge) != 1 {
		t.Error("npcKnowledge map not working correctly")
	}

	// Test that we can add to roomOccupants without panic
	mgr.roomOccupants["room-1"] = make(map[string]bool)
	mgr.roomOccupants["room-1"]["npc-1"] = true
	if len(mgr.roomOccupants) != 1 {
		t.Error("roomOccupants map not working correctly")
	}
}

// TestUpdateManager_ThreadSafety tests that the mutex is present and accessible.
// While we can't fully test thread-safety without implementing actual methods,
// we can verify that the mutex field exists.
func TestUpdateManager_ThreadSafety(t *testing.T) {
	mgr := NewUpdateManager(nil)

	// Test that we can lock and unlock (this verifies the mutex exists)
	mgr.mu.Lock()
	mgr.mu.Unlock()

	mgr.mu.RLock()
	mgr.mu.RUnlock()

	// If we get here without panic, the mutex is working
}

// TestUpdateManager_ConfigurationVariety tests various configuration combinations.
func TestUpdateManager_ConfigurationVariety(t *testing.T) {
	tests := []struct {
		name   string
		config *UpdateManagerConfig
	}{
		{
			name: "distortion disabled",
			config: &UpdateManagerConfig{
				EnableDistortion:    false,
				DistortionRate:      0.0,
				MaxPropagationDepth: 3,
			},
		},
		{
			name: "high distortion",
			config: &UpdateManagerConfig{
				EnableDistortion:    true,
				DistortionRate:      0.8,
				MaxPropagationDepth: 2,
			},
		},
		{
			name: "unlimited propagation",
			config: &UpdateManagerConfig{
				EnableDistortion:    true,
				DistortionRate:      0.1,
				MaxPropagationDepth: 100,
			},
		},
		{
			name: "no propagation",
			config: &UpdateManagerConfig{
				EnableDistortion:    false,
				DistortionRate:      0.0,
				MaxPropagationDepth: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewUpdateManager(tt.config)

			if mgr == nil {
				t.Fatal("NewUpdateManager returned nil")
			}

			if mgr.config.EnableDistortion != tt.config.EnableDistortion {
				t.Errorf("EnableDistortion mismatch: got %v, want %v",
					mgr.config.EnableDistortion, tt.config.EnableDistortion)
			}

			if mgr.config.DistortionRate != tt.config.DistortionRate {
				t.Errorf("DistortionRate mismatch: got %f, want %f",
					mgr.config.DistortionRate, tt.config.DistortionRate)
			}

			if mgr.config.MaxPropagationDepth != tt.config.MaxPropagationDepth {
				t.Errorf("MaxPropagationDepth mismatch: got %d, want %d",
					mgr.config.MaxPropagationDepth, tt.config.MaxPropagationDepth)
			}
		})
	}
}

// TestUpdateManager_IndependentInstances tests that multiple UpdateManager
// instances are independent and don't share state.
func TestUpdateManager_IndependentInstances(t *testing.T) {
	mgr1 := NewUpdateManager(nil)
	mgr2 := NewUpdateManager(nil)

	// Add a fact to mgr1
	mgr1.globalFacts["fact-1"] = &Fact{}

	// Verify mgr2 doesn't have it
	if len(mgr2.globalFacts) != 0 {
		t.Error("UpdateManager instances are sharing globalFacts map")
	}

	// Add knowledge to mgr1
	mgr1.npcKnowledge["npc-1"] = &KnowledgeBase{}

	// Verify mgr2 doesn't have it
	if len(mgr2.npcKnowledge) != 0 {
		t.Error("UpdateManager instances are sharing npcKnowledge map")
	}
}

// TestUpdateManager_EmptyConfiguration tests behavior with zero values.
func TestUpdateManager_EmptyConfiguration(t *testing.T) {
	config := &UpdateManagerConfig{
		EnableDistortion:    false,
		DistortionRate:      0.0,
		MaxPropagationDepth: 0,
	}

	mgr := NewUpdateManager(config)

	if mgr == nil {
		t.Fatal("NewUpdateManager returned nil with zero config")
	}

	// Should still initialize all fields
	if mgr.globalFacts == nil {
		t.Error("globalFacts not initialized with zero config")
	}

	if mgr.npcKnowledge == nil {
		t.Error("npcKnowledge not initialized with zero config")
	}

	if mgr.playerKnowledge == nil {
		t.Error("playerKnowledge not initialized with zero config")
	}

	if mgr.roomOccupants == nil {
		t.Error("roomOccupants not initialized with zero config")
	}
}
