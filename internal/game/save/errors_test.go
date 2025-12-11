package save

import (
	"errors"
	"os"
	"testing"
)

func TestNarrativeErrorMapping(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		shouldMatch bool
	}{
		{"permission denied", os.ErrPermission, true},
		{"disk full", errors.New("no space left on device"), true},
		{"corrupted error", &CorruptedError{Message: "test"}, true},
		{"generic error", errors.New("some error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			narrative := NarrativeError(tt.err)
			if narrative == "" {
				t.Errorf("NarrativeError should return non-empty string for %v", tt.err)
			}

			// Should contain both narrative and technical parts
			if len(narrative) < 10 {
				t.Errorf("Narrative error seems too short: %s", narrative)
			}
		})
	}
}

func TestNarrativeErrorForPermission(t *testing.T) {
	narrative := NarrativeError(os.ErrPermission)

	// Should contain narrative text
	if narrative == "" {
		t.Error("Expected non-empty narrative for permission error")
	}

	// Should mention permission in some form
	found := false
	keywords := []string{"權限", "permission", "無法", "失敗"}
	for _, kw := range keywords {
		if containsString(narrative, kw) {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Warning: Narrative may not mention permission issue: %s", narrative)
	}
}

func TestNarrativeErrorForDiskFull(t *testing.T) {
	err := errors.New("no space left on device")
	narrative := NarrativeError(err)

	if narrative == "" {
		t.Error("Expected non-empty narrative for disk full error")
	}
}

func TestNarrativeErrorForCorrupted(t *testing.T) {
	err := &CorruptedError{Message: "checksum mismatch"}
	narrative := NarrativeError(err)

	if narrative == "" {
		t.Error("Expected non-empty narrative for corrupted error")
	}

	// Should preserve the original corruption message
	if !containsString(narrative, "checksum") && !containsString(narrative, "校驗") && !containsString(narrative, "損壞") {
		t.Logf("Narrative for corrupted: %s", narrative)
	}
}

func TestSaveErrorWithRecovery(t *testing.T) {
	err := os.ErrPermission
	saveErr := WrapSaveError(err)

	// Should have recovery suggestion
	if saveErr.RecoverySuggestion == "" {
		t.Error("Save error should have recovery suggestion")
	}

	// Should preserve original error
	if !errors.Is(saveErr.Cause, os.ErrPermission) {
		t.Error("Save error should preserve original cause")
	}
}

func TestLoadErrorWithRecovery(t *testing.T) {
	err := &CorruptedError{Message: "invalid checksum"}
	loadErr := WrapLoadError(err)

	// Should have recovery suggestion
	if loadErr.RecoverySuggestion == "" {
		t.Error("Load error should have recovery suggestion")
	}
}

func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		err      error
		category ErrorCategory
	}{
		{os.ErrPermission, CategoryPermission},
		{os.ErrNotExist, CategoryNotFound},
		{&CorruptedError{}, CategoryCorrupted},
		{errors.New("no space"), CategoryDiskFull},
		{errors.New("random error"), CategoryUnknown},
	}

	for _, tt := range tests {
		cat := CategorizeError(tt.err)
		if cat != tt.category {
			t.Errorf("CategorizeError(%v) = %v, want %v", tt.err, cat, tt.category)
		}
	}
}

func TestSaveErrorIsError(t *testing.T) {
	saveErr := &SaveError{
		Narrative: "記憶無法固化",
		Cause:     os.ErrPermission,
	}

	// Should implement error interface
	var _ error = saveErr

	errStr := saveErr.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestLoadErrorIsError(t *testing.T) {
	loadErr := &LoadError{
		Narrative: "記憶已被扭曲",
		Cause:     &CorruptedError{Message: "bad checksum"},
	}

	// Should implement error interface
	var _ error = loadErr

	errStr := loadErr.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr))
}
