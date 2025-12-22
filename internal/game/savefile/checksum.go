package savefile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/errors"
)

// ==========================================================================
// Story 8.1: Checksum functions for SaveFileV2
// ==========================================================================

// CorruptedError represents a save file corruption error.
// It wraps a FriendlyError for i18n support while maintaining backward compatibility.
type CorruptedError struct {
	Message     string
	SlotID      int
	FriendlyErr error // Optional: errors.SaveFileError for i18n support
}

func (e *CorruptedError) Error() string {
	return e.Message
}

// Unwrap returns the wrapped FriendlyError if available.
func (e *CorruptedError) Unwrap() error {
	return e.FriendlyErr
}

// NewCorruptedError creates a new CorruptedError with FriendlyError support.
func NewCorruptedError(slotID int, reason string) *CorruptedError {
	// Create the underlying FriendlyError for i18n support
	friendlyErr := errors.NewSaveFileCorruptedError(slotID, fmt.Errorf("%s", reason))
	return &CorruptedError{
		Message:     fmt.Sprintf("save file corrupted: %s", reason),
		SlotID:      slotID,
		FriendlyErr: friendlyErr,
	}
}

// IsCorruptedError checks if an error is a CorruptedError or a friendly SaveFileError with corrupted flag.
func IsCorruptedError(err error) bool {
	// Check for old CorruptedError
	if _, ok := err.(*CorruptedError); ok {
		return true
	}

	// Check if the error message contains corruption indicators
	errStr := err.Error()
	return strings.Contains(errStr, "corrupted") ||
		strings.Contains(errStr, "損壞") ||
		strings.Contains(errStr, "校驗") ||
		strings.Contains(errStr, "invalid JSON")
}

// ComputeChecksum computes the SHA256 checksum of a SaveFileV2 structure.
// The checksum field itself is excluded from the computation.
// Story 8.1 AC: Use checksum verification (NFR-S05)
func ComputeChecksum(save *SaveFileV2) (string, error) {
	// Create a copy to avoid modifying the original
	saveCopy := *save
	saveCopy.Checksum = "" // Exclude checksum field from computation

	// Serialize to JSON for consistent hashing
	data, err := json.Marshal(saveCopy)
	if err != nil {
		return "", fmt.Errorf("failed to serialize save data: %w", err)
	}

	// Compute SHA256
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyChecksum verifies the checksum of a SaveFileV2 structure.
// Story 8.1 AC: Use checksum verification (NFR-S05)
// Story 10.6: Uses FriendlyError for i18n-friendly error messages
func VerifyChecksum(save *SaveFileV2) error {
	slotID := 0 // Default slot, can be set from context if needed

	if save.Checksum == "" {
		return NewCorruptedError(slotID, "missing checksum")
	}

	computed, err := ComputeChecksum(save)
	if err != nil {
		return NewCorruptedError(slotID, fmt.Sprintf("cannot compute checksum: %v", err))
	}

	if computed != save.Checksum {
		return NewCorruptedError(slotID, "checksum mismatch - file may be tampered or corrupted")
	}

	return nil
}

// SetChecksum computes and sets the checksum on a SaveFileV2 structure.
// Story 8.1 AC: Use checksum verification (NFR-S05)
func SetChecksum(save *SaveFileV2) error {
	checksum, err := ComputeChecksum(save)
	if err != nil {
		return err
	}
	save.Checksum = checksum
	return nil
}
