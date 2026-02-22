package views

import (
	"fmt"
	"io"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// agentItem wraps AgentData for the list model.
type agentItem struct {
	data notionagents.AgentData
}

func (i agentItem) Title() string {
	icon := "🤖"
	if i.data.Icon != nil && i.data.Icon.Emoji != nil {
		icon = *i.data.Icon.Emoji
	}
	name := i.data.Name
	if notionagents.IsPersonalAgent(i.data.ID) {
		name += " (Personal)"
	}
	return icon + " " + name
}

func (i agentItem) Description() string {
	if i.data.Description != nil {
		return *i.data.Description
	}
	return "No description"
}

func (i agentItem) FilterValue() string { return i.data.Name }

// agentDelegate renders agent list items with Notion styling.
type agentDelegate struct {
	accent lipgloss.Color
	muted  lipgloss.Color
}

func (d agentDelegate) Height() int                             { return 2 }
func (d agentDelegate) Spacing() int                            { return 1 }
func (d agentDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d agentDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(agentItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(d.accent).
			Render("▸ " + title)
		desc = lipgloss.NewStyle().
			Foreground(d.accent).
			Render("  " + desc)
	} else {
		title = lipgloss.NewStyle().
			Render("  " + title)
		desc = lipgloss.NewStyle().
			Foreground(d.muted).
			Render("  " + desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// AgentSelect is the agent picker sub-model.
type AgentSelect struct {
	List     list.Model
	Selected *notionagents.AgentData
}

// NewAgentSelect creates a new agent selection view.
func NewAgentSelect(agents []notionagents.AgentData, width, height int, accent, muted lipgloss.Color) AgentSelect {
	items := make([]list.Item, len(agents))
	for i, a := range agents {
		items[i] = agentItem{data: a}
	}

	delegate := agentDelegate{accent: accent, muted: muted}
	l := list.New(items, delegate, width, height)
	l.Title = "Select an Agent"
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Padding(1, 0, 1, 0)

	return AgentSelect{List: l}
}

// Update handles messages for the agent select view.
func (a *AgentSelect) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if item, ok := a.List.SelectedItem().(agentItem); ok {
				a.Selected = &item.data
			}
			return nil
		}
	}

	var cmd tea.Cmd
	a.List, cmd = a.List.Update(msg)
	return cmd
}

// View renders the agent select view.
func (a *AgentSelect) View() string {
	return a.List.View()
}

// SetSize updates the list dimensions.
func (a *AgentSelect) SetSize(w, h int) {
	a.List.SetSize(w, h)
}
