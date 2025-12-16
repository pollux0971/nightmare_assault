package seed

import (
	"testing"
	"time"
)

// ============================================================================
// Task 5.1: LocalSeed Constructor Validation Tests (8-10 cases)
// ============================================================================

// TestNewLocalSeed_Success tests successful LocalSeed creation with valid parameters.
func TestNewLocalSeed_Success(t *testing.T) {
	seed, err := NewLocalSeed(
		"LS-hospital-01",
		"hospital_corridor",
		"Strange scratches on the wall",
		"Three parallel lines",
		"You notice strange scratches on the wall",
		10,
		5,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if seed.ID != "LS-hospital-01" {
		t.Errorf("Expected ID='LS-hospital-01', got '%s'", seed.ID)
	}

	if seed.SceneID != "hospital_corridor" {
		t.Errorf("Expected SceneID='hospital_corridor', got '%s'", seed.SceneID)
	}

	if seed.Content != "Strange scratches on the wall" {
		t.Errorf("Expected Content='Strange scratches on the wall', got '%s'", seed.Content)
	}

	if seed.Detail != "Three parallel lines" {
		t.Errorf("Expected Detail='Three parallel lines', got '%s'", seed.Detail)
	}

	if seed.PlantText != "You notice strange scratches on the wall" {
		t.Errorf("Expected PlantText='You notice strange scratches on the wall', got '%s'", seed.PlantText)
	}

	if seed.PlantedAt != 10 {
		t.Errorf("Expected PlantedAt=10, got %d", seed.PlantedAt)
	}

	if seed.MaxLifespan != 5 {
		t.Errorf("Expected MaxLifespan=5, got %d", seed.MaxLifespan)
	}

	if seed.Status != SeedStatusActive {
		t.Errorf("Expected Status=Active, got %s", seed.Status)
	}

	if seed.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set, got zero time")
	}
}

// TestNewLocalSeed_EmptyID tests error when ID is empty.
func TestNewLocalSeed_EmptyID(t *testing.T) {
	seed, err := NewLocalSeed("", "scene1", "content", "detail", "plant", 10, 5)

	if err == nil {
		t.Fatal("Expected error for empty ID, got nil")
	}

	if seed != nil {
		t.Error("Expected nil seed on error, got non-nil")
	}

	expectedMsg := "id cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewLocalSeed_EmptySceneID tests error when SceneID is empty.
func TestNewLocalSeed_EmptySceneID(t *testing.T) {
	seed, err := NewLocalSeed("LS-test-01", "", "content", "detail", "plant", 10, 5)

	if err == nil {
		t.Fatal("Expected error for empty SceneID, got nil")
	}

	if seed != nil {
		t.Error("Expected nil seed on error, got non-nil")
	}

	expectedMsg := "sceneID cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewLocalSeed_EmptyContent tests error when Content is empty.
func TestNewLocalSeed_EmptyContent(t *testing.T) {
	seed, err := NewLocalSeed("LS-test-01", "scene1", "", "detail", "plant", 10, 5)

	if err == nil {
		t.Fatal("Expected error for empty Content, got nil")
	}

	if seed != nil {
		t.Error("Expected nil seed on error, got non-nil")
	}

	expectedMsg := "content cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewLocalSeed_NegativePlantedAt tests error when PlantedAt is negative.
func TestNewLocalSeed_NegativePlantedAt(t *testing.T) {
	seed, err := NewLocalSeed("LS-test-01", "scene1", "content", "detail", "plant", -5, 5)

	if err == nil {
		t.Fatal("Expected error for negative PlantedAt, got nil")
	}

	if seed != nil {
		t.Error("Expected nil seed on error, got non-nil")
	}

	expectedMsg := "plantedAt cannot be negative, got -5"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewLocalSeed_DefaultMaxLifespan tests default MaxLifespan when 0 or negative provided.
func TestNewLocalSeed_DefaultMaxLifespan(t *testing.T) {
	tests := []struct {
		name            string
		inputLifespan   int
		expectedDefault int
	}{
		{"zero lifespan", 0, 5},
		{"negative lifespan", -3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, err := NewLocalSeed("LS-test-01", "scene1", "content", "detail", "plant", 10, tt.inputLifespan)

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if seed.MaxLifespan != tt.expectedDefault {
				t.Errorf("Expected MaxLifespan=%d (default), got %d", tt.expectedDefault, seed.MaxLifespan)
			}
		})
	}
}

// TestNewLocalSeed_AllFieldsCopied tests all fields are correctly copied.
func TestNewLocalSeed_AllFieldsCopied(t *testing.T) {
	seed, err := NewLocalSeed(
		"LS-library-03",
		"old_library",
		"Dusty book on forbidden rituals",
		"Blood stains on pages 13-15",
		"You find a dusty book with crimson stains",
		20,
		7,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all fields
	if seed.ID != "LS-library-03" {
		t.Errorf("ID mismatch: got %s", seed.ID)
	}
	if seed.SceneID != "old_library" {
		t.Errorf("SceneID mismatch: got %s", seed.SceneID)
	}
	if seed.Content != "Dusty book on forbidden rituals" {
		t.Errorf("Content mismatch: got %s", seed.Content)
	}
	if seed.Detail != "Blood stains on pages 13-15" {
		t.Errorf("Detail mismatch: got %s", seed.Detail)
	}
	if seed.PlantText != "You find a dusty book with crimson stains" {
		t.Errorf("PlantText mismatch: got %s", seed.PlantText)
	}
	if seed.PlantedAt != 20 {
		t.Errorf("PlantedAt mismatch: got %d", seed.PlantedAt)
	}
	if seed.MaxLifespan != 7 {
		t.Errorf("MaxLifespan mismatch: got %d", seed.MaxLifespan)
	}
	if seed.Status != SeedStatusActive {
		t.Errorf("Status mismatch: got %s", seed.Status)
	}
}

// TestNewLocalSeed_EmptyDetailAllowed tests that empty Detail is allowed (optional field).
func TestNewLocalSeed_EmptyDetailAllowed(t *testing.T) {
	seed, err := NewLocalSeed("LS-test-01", "scene1", "content", "", "plant", 10, 5)

	if err != nil {
		t.Fatalf("Expected no error for empty Detail, got %v", err)
	}

	if seed.Detail != "" {
		t.Errorf("Expected empty Detail, got '%s'", seed.Detail)
	}
}

// TestNewLocalSeed_EmptyPlantTextAllowed tests that empty PlantText is allowed (optional field).
func TestNewLocalSeed_EmptyPlantTextAllowed(t *testing.T) {
	seed, err := NewLocalSeed("LS-test-01", "scene1", "content", "detail", "", 10, 5)

	if err != nil {
		t.Fatalf("Expected no error for empty PlantText, got %v", err)
	}

	if seed.PlantText != "" {
		t.Errorf("Expected empty PlantText, got '%s'", seed.PlantText)
	}
}

// ============================================================================
// Task 5.2: Urgency Calculation Tests (6-8 cases)
// ============================================================================

// TestLocalSeed_CalculateUrgency tests urgency calculation for various remaining lifespans.
func TestLocalSeed_CalculateUrgency(t *testing.T) {
	tests := []struct {
		name            string
		plantedAt       int
		maxLifespan     int
		currentBeat     int
		expectedUrgency int
		description     string
	}{
		{
			"expired (0 remaining)",
			10, 5, 15,
			100,
			"MaxLifespan reached exactly",
		},
		{
			"expired (negative remaining)",
			10, 5, 18,
			100,
			"Past MaxLifespan by 3 beats",
		},
		{
			"critical (1 remaining)",
			10, 5, 14,
			90,
			"Only 1 beat left",
		},
		{
			"forced harvest (2 remaining)",
			10, 5, 13,
			60,
			"Forced harvest threshold (2 beats)",
		},
		{
			"medium (3 remaining)",
			10, 5, 12,
			40,
			"Medium urgency (3 beats)",
		},
		{
			"low (5 remaining)",
			10, 5, 10,
			20,
			"Just planted (full lifespan)",
		},
		{
			"low (4 remaining)",
			10, 5, 11,
			20,
			"Low urgency (4 beats)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", tt.plantedAt, tt.maxLifespan)
			urgency := seed.CalculateUrgency(tt.currentBeat)

			if urgency != tt.expectedUrgency {
				t.Errorf("%s: Expected urgency=%d, got %d",
					tt.description, tt.expectedUrgency, urgency)
			}
		})
	}
}

// TestLocalSeed_GetRemainingLifespan tests remaining lifespan calculation.
func TestLocalSeed_GetRemainingLifespan(t *testing.T) {
	tests := []struct {
		name         string
		plantedAt    int
		maxLifespan  int
		currentBeat  int
		expectedLeft int
	}{
		{"just planted", 10, 5, 10, 5},
		{"1 beat passed", 10, 5, 11, 4},
		{"2 beats passed", 10, 5, 12, 3},
		{"3 beats passed", 10, 5, 13, 2},
		{"4 beats passed", 10, 5, 14, 1},
		{"expired exactly", 10, 5, 15, 0},
		{"expired by 1", 10, 5, 16, -1},
		{"expired by 3", 10, 5, 18, -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", tt.plantedAt, tt.maxLifespan)
			remaining := seed.GetRemainingLifespan(tt.currentBeat)

			if remaining != tt.expectedLeft {
				t.Errorf("Expected remaining=%d, got %d", tt.expectedLeft, remaining)
			}
		})
	}
}

// TestLocalSeed_IsExpired tests expiration checking.
func TestLocalSeed_IsExpired(t *testing.T) {
	tests := []struct {
		name           string
		plantedAt      int
		maxLifespan    int
		currentBeat    int
		expectedExpiry bool
	}{
		{"not expired (just planted)", 10, 5, 10, false},
		{"not expired (1 beat left)", 10, 5, 14, false},
		{"expired (exactly at limit)", 10, 5, 15, true},
		{"expired (past limit)", 10, 5, 18, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", tt.plantedAt, tt.maxLifespan)
			expired := seed.IsExpired(tt.currentBeat)

			if expired != tt.expectedExpiry {
				t.Errorf("Expected IsExpired=%v, got %v", tt.expectedExpiry, expired)
			}
		})
	}
}

// TestLocalSeed_ShouldForceHarvest tests forced harvest threshold (urgency >= 40).
func TestLocalSeed_ShouldForceHarvest(t *testing.T) {
	tests := []struct {
		name          string
		plantedAt     int
		maxLifespan   int
		currentBeat   int
		expectedForce bool
		description   string
	}{
		{"expired", 10, 5, 15, true, "Urgency 100 (expired)"},
		{"critical", 10, 5, 14, true, "Urgency 90 (1 beat left)"},
		{"high", 10, 5, 13, true, "Urgency 60 (2 beats left)"},
		{"threshold", 10, 5, 12, true, "Urgency 40 (3 beats left, threshold)"},
		{"low", 10, 5, 11, false, "Urgency 20 (4 beats left)"},
		{"fresh", 10, 5, 10, false, "Urgency 20 (just planted)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", tt.plantedAt, tt.maxLifespan)
			shouldForce := seed.ShouldForceHarvest(tt.currentBeat)

			if shouldForce != tt.expectedForce {
				t.Errorf("%s: Expected ShouldForceHarvest=%v, got %v",
					tt.description, tt.expectedForce, shouldForce)
			}
		})
	}
}

// ============================================================================
// Task 5.5: DeepCopy and Status Tests (4-6 cases)
// ============================================================================

// TestLocalSeed_DeepCopy tests deep copy functionality.
func TestLocalSeed_DeepCopy(t *testing.T) {
	original, _ := NewLocalSeed(
		"LS-test-01",
		"hospital",
		"Blood stains",
		"Handprint",
		"You see blood stains",
		10,
		5,
	)

	// Create deep copy
	copy := original.DeepCopy()

	// Verify all fields match
	if copy.ID != original.ID {
		t.Errorf("ID mismatch: original=%s, copy=%s", original.ID, copy.ID)
	}
	if copy.SceneID != original.SceneID {
		t.Errorf("SceneID mismatch")
	}
	if copy.Content != original.Content {
		t.Errorf("Content mismatch")
	}
	if copy.Detail != original.Detail {
		t.Errorf("Detail mismatch")
	}
	if copy.PlantText != original.PlantText {
		t.Errorf("PlantText mismatch")
	}
	if copy.PlantedAt != original.PlantedAt {
		t.Errorf("PlantedAt mismatch")
	}
	if copy.MaxLifespan != original.MaxLifespan {
		t.Errorf("MaxLifespan mismatch")
	}
	if copy.Status != original.Status {
		t.Errorf("Status mismatch")
	}
	if !copy.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt mismatch")
	}

	// Verify independence - modify copy, original should not change
	copy.Status = SeedStatusHarvested
	if original.Status == SeedStatusHarvested {
		t.Error("Modifying copy affected original - not a deep copy")
	}
}

// TestLocalSeed_DeepCopy_NilSeed tests deep copy with nil seed.
func TestLocalSeed_DeepCopy_NilSeed(t *testing.T) {
	var seed *LocalSeed = nil
	copy := seed.DeepCopy()

	if copy != nil {
		t.Error("Expected nil copy from nil seed, got non-nil")
	}
}

// TestLocalSeed_StatusTransitions tests seed status lifecycle.
func TestLocalSeed_StatusTransitions(t *testing.T) {
	seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)

	// Initial status should be Active
	if seed.Status != SeedStatusActive {
		t.Errorf("Expected initial status=Active, got %s", seed.Status)
	}

	// Transition to Harvested
	seed.Status = SeedStatusHarvested
	if seed.Status != SeedStatusHarvested {
		t.Error("Failed to transition to Harvested status")
	}

	// Reset and transition to Pruned
	seed.Status = SeedStatusActive
	seed.Status = SeedStatusPruned
	if seed.Status != SeedStatusPruned {
		t.Error("Failed to transition to Pruned status")
	}
}

// TestLocalSeed_CreatedAtTimestamp tests that CreatedAt is set correctly.
func TestLocalSeed_CreatedAtTimestamp(t *testing.T) {
	before := time.Now().UTC()
	seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)
	after := time.Now().UTC()

	if seed.CreatedAt.Before(before) || seed.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp not within expected range")
	}
}

// TestLocalSeed_CalculateUrgency_EdgeCases tests edge cases in urgency calculation.
func TestLocalSeed_CalculateUrgency_EdgeCases(t *testing.T) {
	// Test with very long lifespan
	seed1, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 100)
	if urgency := seed1.CalculateUrgency(10); urgency != 20 {
		t.Errorf("Expected urgency=20 for long lifespan, got %d", urgency)
	}

	// Test with lifespan of 1
	seed2, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 1)
	if urgency := seed2.CalculateUrgency(10); urgency != 90 {
		t.Errorf("Expected urgency=90 for lifespan=1, got %d", urgency)
	}

	// Test expired by many beats
	seed3, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)
	if urgency := seed3.CalculateUrgency(100); urgency != 100 {
		t.Errorf("Expected urgency=100 for heavily expired seed, got %d", urgency)
	}
}

// TestLocalSeed_CalculateUrgency_NonActiveStatus tests AC 5: non-active seeds return urgency 0.
func TestLocalSeed_CalculateUrgency_NonActiveStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      SeedStatus
		description string
	}{
		{"harvested", SeedStatusHarvested, "Harvested seeds have no urgency"},
		{"pruned", SeedStatusPruned, "Pruned seeds have no urgency"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)
			// Manually change status (simulating harvest/prune operation)
			seed.Status = tt.status

			// Even if seed is expired (beat 20 > plantedAt 10 + lifespan 5),
			// non-active seeds should return 0
			urgency := seed.CalculateUrgency(20)

			if urgency != 0 {
				t.Errorf("%s: Expected urgency=0, got %d", tt.description, urgency)
			}
		})
	}

	// Verify active seeds still calculate urgency correctly
	t.Run("active_still_works", func(t *testing.T) {
		seed, _ := NewLocalSeed("LS-test", "scene1", "content", "detail", "plant", 10, 5)
		// Status is Active by default
		urgency := seed.CalculateUrgency(20) // Expired

		if urgency != 100 {
			t.Errorf("Active seed: Expected urgency=100 for expired, got %d", urgency)
		}
	})
}
