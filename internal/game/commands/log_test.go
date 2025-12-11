package commands

import (
	"testing"
)

func TestParseLogCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "Default count",
			input:     "/log",
			wantCount: DefaultLogCount,
			wantErr:   false,
		},
		{
			name:      "Default count without slash",
			input:     "log",
			wantCount: DefaultLogCount,
			wantErr:   false,
		},
		{
			name:      "Specific count",
			input:     "/log 20",
			wantCount: 20,
			wantErr:   false,
		},
		{
			name:      "Specific count without slash",
			input:     "log 50",
			wantCount: 50,
			wantErr:   false,
		},
		{
			name:      "Max count capped",
			input:     "/log 150",
			wantCount: MaxLogCount,
			wantErr:   false,
		},
		{
			name:      "Count equals max",
			input:     "/log 100",
			wantCount: 100,
			wantErr:   false,
		},
		{
			name:        "Invalid count - not a number",
			input:       "/log abc",
			wantCount:   0,
			wantErr:     true,
			errContains: "invalid count",
		},
		{
			name:        "Invalid count - negative",
			input:       "/log -5",
			wantCount:   0,
			wantErr:     true,
			errContains: "count must be positive",
		},
		{
			name:        "Invalid count - zero",
			input:       "/log 0",
			wantCount:   0,
			wantErr:     true,
			errContains: "count must be positive",
		},
		{
			name:        "Too many arguments",
			input:       "/log 20 30",
			wantCount:   0,
			wantErr:     true,
			errContains: "too many arguments",
		},
		{
			name:        "Not a log command",
			input:       "/save",
			wantCount:   0,
			wantErr:     true,
			errContains: "not a log command",
		},
		{
			name:        "Empty command",
			input:       "",
			wantCount:   0,
			wantErr:     true,
			errContains: "empty command",
		},
		{
			name:      "With extra whitespace",
			input:     "  /log  25  ",
			wantCount: 25,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, err := ParseLogCommand(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLogCommand() expected error, got nil")
					return
				}

				if tt.errContains != "" && err.Error() != "" {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("ParseLogCommand() error = %v, want error containing %v", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("ParseLogCommand() unexpected error: %v", err)
				return
			}

			if gotCount != tt.wantCount {
				t.Errorf("ParseLogCommand() count = %v, want %v", gotCount, tt.wantCount)
			}
		})
	}
}

func TestIsLogCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "Basic log command",
			input: "/log",
			want:  true,
		},
		{
			name:  "Log command without slash",
			input: "log",
			want:  true,
		},
		{
			name:  "Log command with count",
			input: "/log 20",
			want:  true,
		},
		{
			name:  "Not a log command",
			input: "/save",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
		{
			name:  "Just slash",
			input: "/",
			want:  false,
		},
		{
			name:  "With whitespace",
			input: "  /log  ",
			want:  true,
		},
		{
			name:  "Partial match",
			input: "/logger",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLogCommand(tt.input)
			if got != tt.want {
				t.Errorf("IsLogCommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultLogCount(t *testing.T) {
	if DefaultLogCount != 10 {
		t.Errorf("DefaultLogCount = %d, want 10", DefaultLogCount)
	}
}

func TestMaxLogCount(t *testing.T) {
	if MaxLogCount != 100 {
		t.Errorf("MaxLogCount = %d, want 100", MaxLogCount)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
