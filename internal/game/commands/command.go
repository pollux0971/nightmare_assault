// Package commands provides slash command implementations.
package commands

import (
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
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register registers a command.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[strings.ToLower(name)]
	return cmd, ok
}

// List returns all registered command names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
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
