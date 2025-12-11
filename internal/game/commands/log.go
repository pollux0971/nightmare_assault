package commands

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultLogCount is the default number of log entries to display.
	DefaultLogCount = 10
	// MaxLogCount is the maximum number of log entries to display.
	MaxLogCount = 100
)

// ParseLogCommand parses a /log command and returns the number of entries to display.
// Supports:
// - /log       -> returns DefaultLogCount (10)
// - /log 20    -> returns 20
// - /log 150   -> returns MaxLogCount (100, capped)
func ParseLogCommand(input string) (count int, err error) {
	input = strings.TrimSpace(input)

	// Remove leading slash if present
	input = strings.TrimPrefix(input, "/")

	parts := strings.Fields(input)

	if len(parts) == 0 {
		return 0, fmt.Errorf("empty command")
	}

	// Check if it's a log command
	if parts[0] != "log" {
		return 0, fmt.Errorf("not a log command")
	}

	// Default case: /log with no arguments
	if len(parts) == 1 {
		return DefaultLogCount, nil
	}

	// Parse count argument
	if len(parts) == 2 {
		parsedCount, parseErr := strconv.Atoi(parts[1])
		if parseErr != nil {
			return 0, fmt.Errorf("invalid count: %s", parts[1])
		}

		if parsedCount <= 0 {
			return 0, fmt.Errorf("count must be positive: %d", parsedCount)
		}

		// Cap at MaxLogCount
		if parsedCount > MaxLogCount {
			return MaxLogCount, nil
		}

		return parsedCount, nil
	}

	// Too many arguments
	return 0, fmt.Errorf("too many arguments")
}

// IsLogCommand checks if the input is a /log command.
func IsLogCommand(input string) bool {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "/")
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return false
	}

	return parts[0] == "log"
}
