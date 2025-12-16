package narration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTruncate tests the truncate utility function
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than maxLen",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "string equal to maxLen",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "string longer than maxLen",
			input:    "Hello World",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "maxLen zero",
			input:    "Hello",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "Chinese characters",
			input:    "你好世界",
			maxLen:   12,
			expected: "你好世界",
		},
		{
			name:     "Chinese characters truncated",
			input:    "你好世界，這是一個測試",
			maxLen:   12,
			expected: "你好世界...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result,
				"truncate('%s', %d) should return '%s'", tt.input, tt.maxLen, tt.expected)
		})
	}
}
