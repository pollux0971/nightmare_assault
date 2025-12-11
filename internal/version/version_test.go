package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion() should not return empty string")
	}
}

func TestGetInfo(t *testing.T) {
	info := GetInfo()

	// Check that all fields have values
	if info.Version == "" {
		t.Error("Info.Version should not be empty")
	}
	if info.Commit == "" {
		t.Error("Info.Commit should not be empty")
	}
	if info.BuildTime == "" {
		t.Error("Info.BuildTime should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("Info.GoVersion should not be empty")
	}

	// OS and Arch should come from runtime
	if info.OS != runtime.GOOS {
		t.Errorf("Info.OS = %s, want %s", info.OS, runtime.GOOS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("Info.Arch = %s, want %s", info.Arch, runtime.GOARCH)
	}
}

func TestString(t *testing.T) {
	s := String()

	// Should contain "Nightmare Assault"
	if !strings.Contains(s, "Nightmare Assault") {
		t.Errorf("String() should contain 'Nightmare Assault', got: %s", s)
	}

	// Should contain version
	if !strings.Contains(s, Version) {
		t.Errorf("String() should contain version '%s', got: %s", Version, s)
	}

	// Should contain commit info
	if !strings.Contains(s, "commit:") {
		t.Errorf("String() should contain 'commit:', got: %s", s)
	}
}

func TestDefaultValues(t *testing.T) {
	// Default values should be set (without build-time injection)
	tests := []struct {
		name  string
		value string
	}{
		{"Version", Version},
		{"Commit", Commit},
		{"BuildTime", BuildTime},
		{"GoVersion", GoVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s should have a default value", tt.name)
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	// PrintVersion writes to stdout, so we just verify it doesn't panic
	// and contains expected format when captured

	// Save and capture stdout
	old := Version
	Version = "test-v1.0.0"
	defer func() { Version = old }()

	// PrintVersion should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintVersion() panicked: %v", r)
		}
	}()

	// Just call to ensure no panic - actual output goes to stdout
	// In a real test we'd capture stdout, but this validates the function runs
	PrintVersion()
}
