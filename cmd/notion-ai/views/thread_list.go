package views

import (
	"fmt"
	"io"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// threadItem wraps a ThreadListItem for the list model.
type threadItem struct {
	data notionagents.ThreadListItem
}

func (i threadItem) Title() string {
	title := i.data.Title
	if title == "" {
		title = "Untitled"
	}
	return title
}

func (i threadItem) Description() string {
	short := i.data.ID
	if len(short) > 12 {
		short = short[:12]
	}
	return fmt.Sprintf("[%s] %s...", i.data.Status, short)
}

func (i threadItem) FilterValue() string { return i.data.Title }

// threadDelegate renders thread list items.
type threadDelegate struct {
	accent lipgloss.Color
	muted  lipgloss.Color
}

func (d threadDelegate) Height() int                             { return 2 }
func (d threadDelegate) Spacing() int                            { return 1 }
func (d threadDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d threadDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(threadItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	statusIcon := "○"
	switch i.data.Status {
	case notionagents.ThreadStatusCompleted:
		statusIcon = "●"
	case notionagents.ThreadStatusFailed:
		statusIcon = "✕"
	}

	if index == m.Index() {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(d.accent).
			Render("▸ " + statusIcon + " " + title)
		desc = lipgloss.NewStyle().
			Foreground(d.accent).
			Render("    " + desc)
	} else {
		title = lipgloss.NewStyle().
			Render("  " + statusIcon + " " + title)
		desc = lipgloss.NewStyle().
			Foreground(d.muted).
			Render("    " + desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// ThreadList is the thread browser sub-model.
type ThreadList struct {
	List     list.Model
	Selected *notionagents.ThreadListItem
}

// NewThreadList creates a new thread list view.
func NewThreadList(threads []notionagents.ThreadListItem, width, height int, accent, muted lipgloss.Color) ThreadList {
	items := make([]list.Item, len(threads))
	for i, t := range threads {
		items[i] = threadItem{data: t}
	}

	delegate := threadDelegate{accent: accent, muted: muted}
	l := list.New(items, delegate, width, height)
	l.Title = "Threads"
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Padding(1, 0, 1, 2)

	return ThreadList{List: l}
}

// Update handles messages for the thread list.
func (t *ThreadList) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if item, ok := t.List.SelectedItem().(threadItem); ok {
				t.Selected = &item.data
			}
			return nil
		}
	}

	var cmd tea.Cmd
	t.List, cmd = t.List.Update(msg)
	return cmd
}

// View renders the thread list.
func (t *ThreadList) View() string {
	return t.List.View()
}

// SetSize updates list dimensions.
func (t *ThreadList) SetSize(w, h int) {
	t.List.SetSize(w, h)
}
