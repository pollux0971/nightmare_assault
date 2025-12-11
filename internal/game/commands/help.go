package commands

// HelpCommand provides help information.
type HelpCommand struct{}

// NewHelpCommand creates a new help command.
func NewHelpCommand() *HelpCommand {
	return &HelpCommand{}
}

// Name returns the command name.
func (c *HelpCommand) Name() string {
	return "help"
}

// Execute displays help information.
func (c *HelpCommand) Execute(args []string) (string, error) {
	help := `
═══════════════════════════════════════════════════
              NIGHTMARE ASSAULT - HELP
═══════════════════════════════════════════════════

HOW TO PLAY:
  • Read the story text carefully
  • Make choices using number keys (1-4)
  • Press Enter to select default choice
  • Press 'f' to enter free text mode

SLASH COMMANDS:
  /help    - Show this help screen
  /status  - View detailed player status
  /quit    - Exit to main menu

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

═══════════════════════════════════════════════════
Press any key to return to game...
`
	return help, nil
}

// Help returns brief command description.
func (c *HelpCommand) Help() string {
	return "Display help information"
}
