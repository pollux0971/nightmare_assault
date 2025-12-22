// Package commands provides slash command implementations.
package commands

import (
	"fmt"
	"sort"
	"strings"
)

// Command represents a slash command that can be executed.
type Command interface {
	// Name returns the command name (without /).
	Name() string
	// Execute runs the command and returns output or error.
	Execute(args []string) (string, error)
	// Help returns a brief description of the command.
	Help() string
}

// Registry manages registered commands.
type Registry struct {
	commands map[string]Command
	aliases  map[string]string // alias -> canonical name
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register registers a command and its aliases.
func (r *Registry) Register(cmd Command) {
	name := strings.ToLower(cmd.Name())
	r.commands[name] = cmd

	// Register aliases if the command supports them
	if aliaser, ok := cmd.(interface{ Aliases() []string }); ok {
		for _, alias := range aliaser.Aliases() {
			r.aliases[strings.ToLower(alias)] = name
		}
	}
}

// Get retrieves a command by name or alias.
func (r *Registry) Get(name string) (Command, bool) {
	name = strings.ToLower(name)

	// Check direct command name
	if cmd, ok := r.commands[name]; ok {
		return cmd, ok
	}

	// Check aliases
	if canonicalName, ok := r.aliases[name]; ok {
		if cmd, ok := r.commands[canonicalName]; ok {
			return cmd, true
		}
	}

	return nil, false
}

// List returns all registered command names (sorted).
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SuggestCommand returns the closest matching command name for fuzzy matching.
// Uses simple Levenshtein distance for suggestions.
func (r *Registry) SuggestCommand(input string) string {
	input = strings.ToLower(input)
	bestMatch := ""
	bestDistance := 999

	// Check all command names
	for name := range r.commands {
		distance := levenshteinDistance(input, name)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = name
		}
	}

	// Check all aliases
	for alias := range r.aliases {
		distance := levenshteinDistance(input, alias)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = alias
		}
	}

	// Only suggest if the distance is reasonable (< 3 edits)
	if bestDistance <= 3 {
		return bestMatch
	}
	return ""
}

// FormatUnknownCommandError returns a formatted error message for unknown commands.
func (r *Registry) FormatUnknownCommandError(input string) string {
	suggestion := r.SuggestCommand(input)
	if suggestion != "" {
		return fmt.Sprintf(
			"未知指令: /%s\nUnknown command: /%s\n\n"+
				"你是否想要使用: /%s\nDid you mean: /%s?\n\n"+
				"使用 /help 查看所有可用指令\nUse /help to see all available commands",
			input, input, suggestion, suggestion)
	}
	return fmt.Sprintf(
		"未知指令: /%s\nUnknown command: /%s\n\n"+
			"使用 /help 查看所有可用指令\nUse /help to see all available commands",
		input, input)
}

// Parse parses a command string into name and arguments.
func Parse(input string) (name string, args []string) {
	input = strings.TrimSpace(input)

	// Remove leading /
	if len(input) > 0 && input[0] == '/' {
		input = input[1:]
	}

	// Split by spaces
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}

	name = strings.ToLower(parts[0])
	if len(parts) > 1 {
		args = parts[1:]
	}

	return name, args
}

// levenshteinDistance calculates the edit distance between two strings.
// Used for fuzzy command matching.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a 2D matrix for dynamic programming
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
