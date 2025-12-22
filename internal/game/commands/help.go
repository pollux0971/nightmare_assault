package commands

import (
	"fmt"
	"strings"
)

// HelpCommand provides help information.
type HelpCommand struct {
	registry *Registry
}

// NewHelpCommand creates a new help command.
func NewHelpCommand() *HelpCommand {
	return &HelpCommand{}
}

// NewHelpCommandWithRegistry creates a new help command with a registry reference.
func NewHelpCommandWithRegistry(registry *Registry) *HelpCommand {
	return &HelpCommand{
		registry: registry,
	}
}

// Name returns the command name.
func (c *HelpCommand) Name() string {
	return "help"
}

// Execute displays help information.
func (c *HelpCommand) Execute(args []string) (string, error) {
	// If specific command help is requested
	if len(args) > 0 && c.registry != nil {
		cmdName := strings.ToLower(args[0])
		cmd, exists := c.registry.Get(cmdName)
		if exists {
			return c.detailedHelp(cmd), nil
		}
		return fmt.Sprintf("Unknown command: %s\nUse /help to see all available commands.", cmdName), nil
	}

	// Build command list dynamically if registry is available
	commandList := c.buildCommandList()

	help := fmt.Sprintf(`
═══════════════════════════════════════════════════
              NIGHTMARE ASSAULT - HELP
═══════════════════════════════════════════════════

HOW TO PLAY:
  • Read the story text carefully
  • Make choices using number keys (1-4)
  • Press Enter to select default choice
  • Press 'f' to enter free text mode

SLASH COMMANDS:
%s

KEYBOARD SHORTCUTS:
  ESC      - Return to menu / Cancel
  ↑↓       - Scroll story text
  Ctrl+C   - Quit game

GAME MECHANICS:
  HP  - Health Points (0-100)
  SAN - Sanity (0-100)

  Sanity States:
    80-100: Clear-headed (normal)
    50-79:  Anxious (minor effects)
    20-49:  Panicked (hallucinations)
    0-19:   Insanity (loss of control)

TIP: Use /help <command> for detailed help on a specific command

═══════════════════════════════════════════════════
Press any key to return to game...
`, commandList)
	return help, nil
}

// buildCommandList creates a formatted list of all available commands.
func (c *HelpCommand) buildCommandList() string {
	if c.registry == nil {
		// Fallback to static list if no registry
		return `  /help         - Show this help screen
  /status       - View detailed player status
  /inventory    - View inventory items (aliases: /inv, /i)
  /clues        - View discovered clues
  /save <slot>  - Save game to slot
  /load <slot>  - Load game from slot
  /api          - Manage API configuration
  /team         - View team status
  /hint         - Get hints
  /rules        - View discovered rules
  /dreams       - Review dream sequences
  /speed        - Adjust text speed
  /theme        - Change visual theme
  /bgm          - Control background music
  /sfx          - Control sound effects
  /lang         - Change language
  /quit         - Exit to main menu`
	}

	// Get all commands from registry
	commandNames := c.registry.List()

	// Define command categories for better organization
	categories := map[string][]string{
		"Game State": {"status", "inventory", "clues", "team"},
		"Save/Load":  {"save", "load"},
		"Settings":   {"api", "speed", "theme", "bgm", "sfx", "lang"},
		"Assistance": {"help", "hint", "rules", "dreams"},
		"System":     {"quit"},
	}

	var output strings.Builder

	// Build categorized command list
	for _, category := range []string{"Game State", "Save/Load", "Assistance", "Settings", "System"} {
		commands := categories[category]
		hasCommands := false

		for _, cmdName := range commands {
			cmd, exists := c.registry.Get(cmdName)
			if exists {
				if !hasCommands {
					output.WriteString(fmt.Sprintf("\n  【%s】\n", category))
					hasCommands = true
				}
				output.WriteString(fmt.Sprintf("  /%s%s - %s\n",
					cmdName,
					strings.Repeat(" ", max(0, 12-len(cmdName))),
					cmd.Help()))
			}
		}
	}

	// Add any uncategorized commands
	categorizedCmds := make(map[string]bool)
	for _, cmds := range categories {
		for _, cmd := range cmds {
			categorizedCmds[cmd] = true
		}
	}

	uncategorized := make([]string, 0)
	for _, cmdName := range commandNames {
		if !categorizedCmds[cmdName] {
			uncategorized = append(uncategorized, cmdName)
		}
	}

	if len(uncategorized) > 0 {
		output.WriteString("\n  【Other】\n")
		for _, cmdName := range uncategorized {
			cmd, exists := c.registry.Get(cmdName)
			if exists {
				output.WriteString(fmt.Sprintf("  /%s%s - %s\n",
					cmdName,
					strings.Repeat(" ", max(0, 12-len(cmdName))),
					cmd.Help()))
			}
		}
	}

	return output.String()
}

// detailedHelp returns detailed help for a specific command.
func (c *HelpCommand) detailedHelp(cmd Command) string {
	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("              COMMAND: /%s\n", strings.ToUpper(cmd.Name())))
	output.WriteString("═══════════════════════════════════════════════════\n\n")

	output.WriteString(fmt.Sprintf("Description: %s\n\n", cmd.Help()))

	// Check if command has additional interfaces for more info
	if usager, ok := cmd.(interface{ Usage() string }); ok {
		output.WriteString(fmt.Sprintf("Usage: %s\n\n", usager.Usage()))
	}

	if describer, ok := cmd.(interface{ Description() string }); ok {
		output.WriteString(fmt.Sprintf("Details: %s\n\n", describer.Description()))
	}

	if aliaser, ok := cmd.(interface{ Aliases() []string }); ok {
		aliases := aliaser.Aliases()
		if len(aliases) > 0 {
			output.WriteString(fmt.Sprintf("Aliases: %s\n\n", strings.Join(aliases, ", ")))
		}
	}

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString("Press any key to return to game...\n")

	return output.String()
}

// Help returns brief command description.
func (c *HelpCommand) Help() string {
	return "Display help information and available commands"
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
