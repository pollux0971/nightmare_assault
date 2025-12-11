package effects

import (
	"strings"
	"testing"
)

func TestRenderShakingBorder_NoOffset(t *testing.T) {
	content := "Test Line 1\nTest Line 2\nTest Line 3"
	result := RenderShakingBorder(content, 0)

	if result != content {
		t.Errorf("Expected no change with offset=0, got %q", result)
	}
}

func TestRenderShakingBorder_WithOffset(t *testing.T) {
	content := "Test Line"
	result := RenderShakingBorder(content, 3)

	// Should have added some padding (random, so just check structure)
	lines := strings.Split(result, "\n")
	if len(lines) < 1 {
		t.Errorf("Expected at least one line in output")
	}

	// The result should end with the original content (after padding)
	if !strings.Contains(result, "Test Line") {
		t.Errorf("Expected original content to be preserved, got %q", result)
	}
}

func TestRenderShakingBorder_MultipleLines(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3"
	result := RenderShakingBorder(content, 2)

	// Should preserve all lines
	if !strings.Contains(result, "Line 1") {
		t.Errorf("Expected Line 1 to be preserved")
	}
	if !strings.Contains(result, "Line 2") {
		t.Errorf("Expected Line 2 to be preserved")
	}
	if !strings.Contains(result, "Line 3") {
		t.Errorf("Expected Line 3 to be preserved")
	}

	// Should have at least 3 line breaks
	lineCount := strings.Count(result, "\n")
	if lineCount < 3 {
		t.Errorf("Expected at least 3 line breaks, got %d", lineCount)
	}
}

func TestRenderShakingBorder_OffsetRange(t *testing.T) {
	content := "Test"

	// Run multiple times to test randomness behavior
	for i := 0; i < 10; i++ {
		result := RenderShakingBorder(content, 5)

		// Should always contain original content
		if !strings.Contains(result, "Test") {
			t.Errorf("Iteration %d: Expected content to be preserved", i)
		}

		// Result should not be dramatically different in length
		// (max padding is offset/2 + some random, times number of lines)
		maxExpectedLen := len(content) + 10 // generous buffer
		if len(result) > maxExpectedLen {
			t.Errorf("Iteration %d: Result too long: %d chars (max expected %d)", i, len(result), maxExpectedLen)
		}
	}
}

func TestCalculateShakeOffset_NoStability(t *testing.T) {
	result := CalculateShakeOffset(0)
	if result != 0 {
		t.Errorf("Expected 0 offset with stability=0, got %d", result)
	}
}

func TestCalculateShakeOffset_WithStability(t *testing.T) {
	stability := 5

	// Run multiple times to check random range
	for i := 0; i < 20; i++ {
		result := CalculateShakeOffset(stability)

		// Should be within [0, stability]
		if result < 0 || result > stability {
			t.Errorf("Iteration %d: CalculateShakeOffset(%d) = %d, want 0-%d", i, stability, result, stability)
		}
	}
}

func TestCalculateShakeOffset_Distribution(t *testing.T) {
	stability := 3
	results := make(map[int]int)

	// Run many times to check distribution
	for i := 0; i < 100; i++ {
		result := CalculateShakeOffset(stability)
		results[result]++
	}

	// Should produce multiple different values (not all the same)
	if len(results) < 2 {
		t.Errorf("Expected diverse offset values, got only %v", results)
	}

	// All values should be in valid range
	for offset := range results {
		if offset < 0 || offset > stability {
			t.Errorf("Invalid offset value: %d (stability=%d)", offset, stability)
		}
	}
}
