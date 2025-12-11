package commands

// QuitCommand exits the game.
type QuitCommand struct{}

// NewQuitCommand creates a new quit command.
func NewQuitCommand() *QuitCommand {
	return &QuitCommand{}
}

// Name returns the command name.
func (c *QuitCommand) Name() string {
	return "quit"
}

// Execute triggers quit confirmation.
func (c *QuitCommand) Execute(args []string) (string, error) {
	return "QUIT_REQUESTED", nil
}

// Help returns brief command description.
func (c *QuitCommand) Help() string {
	return "Exit to main menu"
}
