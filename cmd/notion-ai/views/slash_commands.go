package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SlashCommand defines a chat slash command.
type SlashCommand struct {
	Name        string
	Description string
}

// Commands is the registry of available slash commands.
var Commands = []SlashCommand{
	{Name: "/new", Description: "new thread"},
	{Name: "/threads", Description: "list threads"},
	{Name: "/agents", Description: "switch agent"},
	{Name: "/quit", Description: "exit"},
}

// SlashCompleter provides autocomplete for slash commands.
type SlashCompleter struct {
	Active      bool
	Matches     []SlashCommand
	SelectedIdx int
	CmdStyle    lipgloss.Style
	DescStyle   lipgloss.Style
	SelStyle    lipgloss.Style
}

// Update filters the command list based on the current input.
func (sc *SlashCompleter) Update(input string) {
	if !strings.HasPrefix(input, "/") || strings.Contains(input, " ") {
		sc.Active = false
		sc.Matches = nil
		sc.SelectedIdx = 0
		return
	}

	// Don't show dropdown for exact matches
	if sc.IsExactMatch(input) {
		sc.Active = false
		sc.Matches = nil
		sc.SelectedIdx = 0
		return
	}

	var matches []SlashCommand
	for _, cmd := range Commands {
		if strings.HasPrefix(cmd.Name, input) {
			matches = append(matches, cmd)
		}
	}

	sc.Matches = matches
	sc.Active = len(matches) > 0
	if sc.SelectedIdx >= len(matches) {
		sc.SelectedIdx = 0
	}
}

// MoveUp moves the selection up in the dropdown.
func (sc *SlashCompleter) MoveUp() {
	if sc.SelectedIdx > 0 {
		sc.SelectedIdx--
	}
}

// MoveDown moves the selection down in the dropdown.
func (sc *SlashCompleter) MoveDown() {
	if sc.SelectedIdx < len(sc.Matches)-1 {
		sc.SelectedIdx++
	}
}

// Selected returns the currently highlighted command, or nil if none.
func (sc *SlashCompleter) Selected() *SlashCommand {
	if !sc.Active || len(sc.Matches) == 0 {
		return nil
	}
	return &sc.Matches[sc.SelectedIdx]
}

// IsExactMatch returns true if the input exactly matches a command name.
func (sc *SlashCompleter) IsExactMatch(input string) bool {
	for _, cmd := range Commands {
		if cmd.Name == input {
			return true
		}
	}
	return false
}

// View renders the autocomplete dropdown.
func (sc *SlashCompleter) View(width int) string {
	if !sc.Active || len(sc.Matches) == 0 {
		return ""
	}

	var lines []string
	for i, cmd := range sc.Matches {
		name := sc.CmdStyle.Render(cmd.Name)
		desc := sc.DescStyle.Render(" " + cmd.Description)
		row := "  " + name + desc
		if i == sc.SelectedIdx {
			row = sc.SelStyle.Width(width - 2).Render(cmd.Name + " " + cmd.Description)
			row = "  " + row
		}
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n")
}

// HeightIfActive returns the height of the dropdown when active, 0 otherwise.
func (sc *SlashCompleter) HeightIfActive() int {
	if !sc.Active {
		return 0
	}
	return len(sc.Matches)
}
