package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom status bar.
type StatusBar struct {
	AgentName string
	ThreadID  string
	Streaming bool
	Error     string
	Width     int
	KeyStyle  lipgloss.Style
	ValStyle  lipgloss.Style
	ErrStyle  lipgloss.Style
	BarStyle  lipgloss.Style
	CmdStyle  lipgloss.Style
}

// View renders the status bar.
func (s *StatusBar) View() string {
	left := ""
	if s.AgentName != "" {
		left = s.KeyStyle.Render("Agent:") + " " + s.ValStyle.Render(s.AgentName)
	}

	help := s.CmdStyle.Render("/new") + s.ValStyle.Render(" thread  ") +
		s.CmdStyle.Render("/agents") + s.ValStyle.Render(" switch")

	gap := s.Width - lipgloss.Width(left) - lipgloss.Width(help) - 4
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + help

	if s.Error != "" {
		return s.ErrStyle.Render(s.Error) + "\n" + s.BarStyle.Render(bar)
	}

	return s.BarStyle.Render(bar)
}
