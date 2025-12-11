package save

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// CorruptedError represents a save file corruption error.
type CorruptedError struct {
	Message string
}

func (e *CorruptedError) Error() string {
	return e.Message
}

// IsCorruptedError checks if an error is a CorruptedError.
func IsCorruptedError(err error) bool {
	var corruptedErr *CorruptedError
	return errors.As(err, &corruptedErr)
}

// ComputeChecksum computes the SHA256 checksum of a SaveData structure.
// The checksum field itself is excluded from the computation.
func ComputeChecksum(save *SaveData) (string, error) {
	// Create a copy to avoid modifying the original
	saveCopy := *save
	saveCopy.Checksum = "" // Exclude checksum field from computation

	// Serialize to JSON for consistent hashing
	data, err := json.Marshal(saveCopy)
	if err != nil {
		return "", fmt.Errorf("序列化存檔資料失敗：%w", err)
	}

	// Compute SHA256
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyChecksum verifies the checksum of a SaveData structure.
func VerifyChecksum(save *SaveData) error {
	if save.Checksum == "" {
		return &CorruptedError{
			Message: "存檔檔案損壞：缺少校驗碼",
		}
	}

	computed, err := ComputeChecksum(save)
	if err != nil {
		return &CorruptedError{
			Message: fmt.Sprintf("存檔檔案損壞：無法計算校驗碼：%v", err),
		}
	}

	if computed != save.Checksum {
		return &CorruptedError{
			Message: "存檔檔案損壞：校驗碼不符，檔案可能已被竄改或損壞",
		}
	}

	return nil
}

// SetChecksum computes and sets the checksum on a SaveData structure.
func SetChecksum(save *SaveData) error {
	checksum, err := ComputeChecksum(save)
	if err != nil {
		return err
	}
	save.Checksum = checksum
	return nil
}
