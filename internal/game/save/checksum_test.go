package save

import (
	"encoding/json"
	"testing"
	"time"
)

func TestComputeChecksum(t *testing.T) {
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 100
	save.Player.SAN = 80

	checksum, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}

	if checksum == "" {
		t.Error("Checksum should not be empty")
	}

	// SHA256 produces 64 hex characters
	if len(checksum) != 64 {
		t.Errorf("Expected 64 character checksum, got %d characters", len(checksum))
	}
}

func TestComputeChecksumDeterministic(t *testing.T) {
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 50

	checksum1, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("First ComputeChecksum failed: %v", err)
	}

	checksum2, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("Second ComputeChecksum failed: %v", err)
	}

	if checksum1 != checksum2 {
		t.Errorf("Checksums should be identical: %s vs %s", checksum1, checksum2)
	}
}

func TestComputeChecksumDifferentData(t *testing.T) {
	save1 := NewSaveData()
	save1.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save1.Player.HP = 100

	save2 := NewSaveData()
	save2.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save2.Player.HP = 99

	checksum1, err := ComputeChecksum(save1)
	if err != nil {
		t.Fatalf("First ComputeChecksum failed: %v", err)
	}

	checksum2, err := ComputeChecksum(save2)
	if err != nil {
		t.Fatalf("Second ComputeChecksum failed: %v", err)
	}

	if checksum1 == checksum2 {
		t.Error("Different data should produce different checksums")
	}
}

func TestVerifyChecksumValid(t *testing.T) {
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 100

	// Compute and set checksum
	checksum, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}
	save.Checksum = checksum

	// Verify should pass
	if err := VerifyChecksum(save); err != nil {
		t.Errorf("VerifyChecksum should pass for valid checksum: %v", err)
	}
}

func TestVerifyChecksumInvalid(t *testing.T) {
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 100

	// Set invalid checksum
	save.Checksum = "invalid_checksum_value"

	// Verify should fail
	err := VerifyChecksum(save)
	if err == nil {
		t.Error("VerifyChecksum should fail for invalid checksum")
	}
}

func TestVerifyChecksumTampered(t *testing.T) {
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 100

	// Compute and set checksum
	checksum, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}
	save.Checksum = checksum

	// Tamper with data
	save.Player.HP = 999

	// Verify should fail
	err = VerifyChecksum(save)
	if err == nil {
		t.Error("VerifyChecksum should fail for tampered data")
	}
}

func TestVerifyChecksumEmpty(t *testing.T) {
	save := NewSaveData()
	save.Checksum = ""

	// Empty checksum should fail verification
	err := VerifyChecksum(save)
	if err == nil {
		t.Error("VerifyChecksum should fail for empty checksum")
	}
}

func TestChecksumErrorMessage(t *testing.T) {
	save := NewSaveData()
	save.Checksum = "invalid"

	err := VerifyChecksum(save)
	if err == nil {
		t.Fatal("Expected error for invalid checksum")
	}

	// Error message should be user-friendly (in Chinese for this game)
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestSaveDataWithChecksum(t *testing.T) {
	// Test the full flow: create, compute checksum, serialize, deserialize, verify
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 75
	save.Player.SAN = 60
	save.Game.CurrentChapter = 3

	// Compute and set checksum
	checksum, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}
	save.Checksum = checksum

	// Serialize
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Deserialize
	var loaded SaveData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if err := VerifyChecksum(&loaded); err != nil {
		t.Errorf("VerifyChecksum should pass after round-trip: %v", err)
	}
}

func TestComputeChecksumExcludesChecksumField(t *testing.T) {
	save := NewSaveData()
	save.Metadata.SavedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	save.Player.HP = 100

	// Compute checksum with empty checksum field
	checksum1, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("First ComputeChecksum failed: %v", err)
	}

	// Set a different checksum value
	save.Checksum = "some_other_value"

	// Compute again - should be the same since checksum field is excluded
	checksum2, err := ComputeChecksum(save)
	if err != nil {
		t.Fatalf("Second ComputeChecksum failed: %v", err)
	}

	if checksum1 != checksum2 {
		t.Error("Checksum computation should exclude the checksum field itself")
	}
}

func TestIsCorruptedError(t *testing.T) {
	save := NewSaveData()
	save.Checksum = "invalid"

	err := VerifyChecksum(save)
	if err == nil {
		t.Fatal("Expected error")
	}

	// Check if it's a corrupted error type
	if !IsCorruptedError(err) {
		t.Error("Expected a corrupted error type")
	}
}

func TestChecksumFieldInSaveData(t *testing.T) {
	// Verify SaveData has Checksum field
	save := &SaveData{}
	save.Checksum = "test_checksum"

	if save.Checksum != "test_checksum" {
		t.Error("Checksum field should be settable")
	}
}
