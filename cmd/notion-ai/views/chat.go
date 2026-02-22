package views

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Craft-themed loading phrases.
var craftPhrases = []string{
	"Thinking",
	"Crafting",
	"Creating",
	"Composing",
	"Shaping",
	"Forming",
	"Weaving",
	"Building",
	"Refining",
	"Imagining",
}

// phraseTickMsg fires on a timer to animate loading phrases.
type phraseTickMsg struct{}

// PhraseTickCmd returns a command that fires a phraseTickMsg after 600ms.
func PhraseTickCmd() tea.Cmd {
	return tea.Tick(600*time.Millisecond, func(time.Time) tea.Msg {
		return phraseTickMsg{}
	})
}

// ChatMessage represents a message displayed in the chat.
type ChatMessage struct {
	Role    string
	Content string
}

// Chat is the main chat view sub-model.
type Chat struct {
	Viewport     viewport.Model
	Input        textarea.Model
	Spinner      spinner.Model
	Messages     []ChatMessage
	Streaming    bool
	StreamBuf    string // accumulated streaming content (plain text during stream)
	InputFocused bool
	Width        int
	Height       int

	// Craft phrase animation
	phraseIndex int
	phraseFrame int

	// Slash command completer
	Completer SlashCompleter

	// Styles
	HumanLabel   lipgloss.Style
	AgentLabel   lipgloss.Style
	HumanMsg     lipgloss.Style
	AgentMsg     lipgloss.Style
	DividerStyle lipgloss.Style
	StreamStyle  lipgloss.Style
}

// NewChat creates a new chat view.
func NewChat(width, height int) Chat {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.CharLimit = 4096
	ta.SetWidth(width - 8) // account for border (2) + padding (2) + outer margin
	ta.SetHeight(1)
	ta.ShowLineNumbers = false

	// Clean, borderless input styling
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B"))
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#2EAADC")).Bold(true).SetString("> ")
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B"))
	ta.BlurredStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B")).SetString("> ")

	ta.Focus()
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	vp := viewport.New(width-4, height-8) // account for bordered input
	vp.SetContent("")

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Chat{
		Viewport:     vp,
		Input:        ta,
		Spinner:      sp,
		InputFocused: true,
		Width:        width,
		Height:       height,
	}
}

// Update handles messages for the chat view.
func (c *Chat) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if c.Streaming {
		var spinCmd tea.Cmd
		c.Spinner, spinCmd = c.Spinner.Update(msg)
		cmds = append(cmds, spinCmd)
	}

	switch msg.(type) {
	case phraseTickMsg:
		if c.Streaming {
			c.phraseFrame++
			if c.phraseFrame >= 3 {
				c.phraseFrame = 0
				c.phraseIndex = (c.phraseIndex + 1) % len(craftPhrases)
			}
			c.refreshViewport()
			cmds = append(cmds, PhraseTickCmd())
		}
		return tea.Batch(cmds...)
	}

	if c.InputFocused && !c.Streaming {
		var inputCmd tea.Cmd
		c.Input, inputCmd = c.Input.Update(msg)
		cmds = append(cmds, inputCmd)

		// Update completer after input changes
		c.Completer.Update(c.Input.Value())
	} else {
		var vpCmd tea.Cmd
		c.Viewport, vpCmd = c.Viewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	return tea.Batch(cmds...)
}

// SetSize updates the chat dimensions.
func (c *Chat) SetSize(w, h int) {
	c.Width = w
	c.Height = h
	// Border adds 2 lines (top + bottom), so input area = 1 (textarea) + 2 (border) = 3
	inputHeight := 3
	statusHeight := 1
	dropdownHeight := c.Completer.HeightIfActive()
	vpHeight := h - inputHeight - statusHeight - dropdownHeight
	if vpHeight < 3 {
		vpHeight = 3
	}
	c.Viewport.Width = w - 4
	c.Viewport.Height = vpHeight
	c.Input.SetWidth(w - 8) // border (2) + padding (2*1) + outer margin
}

// AppendMessage adds a message and refreshes the viewport.
func (c *Chat) AppendMessage(role, content string) {
	c.Messages = append(c.Messages, ChatMessage{Role: role, Content: content})
	c.refreshViewport()
}

// UpdateLastAgent updates the last agent message (for finalized markdown).
func (c *Chat) UpdateLastAgent(content string) {
	for i := len(c.Messages) - 1; i >= 0; i-- {
		if c.Messages[i].Role == "agent" {
			c.Messages[i].Content = content
			break
		}
	}
	c.refreshViewport()
}

// currentPhrase returns the current craft phrase with animated dots.
func (c *Chat) currentPhrase() string {
	phrase := craftPhrases[c.phraseIndex%len(craftPhrases)]
	dots := strings.Repeat(".", c.phraseFrame+1)
	return phrase + dots
}

// ResetPhrase resets the phrase animation to the beginning.
func (c *Chat) ResetPhrase() {
	c.phraseIndex = 0
	c.phraseFrame = 0
}

// wrapText hard-wraps text to the given width, preserving existing newlines.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var out strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if lipgloss.Width(line) <= width {
			if out.Len() > 0 {
				out.WriteByte('\n')
			}
			out.WriteString(line)
			continue
		}
		words := strings.Fields(line)
		cur := ""
		for _, w := range words {
			if cur == "" {
				cur = w
			} else if lipgloss.Width(cur+" "+w) <= width {
				cur += " " + w
			} else {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				out.WriteString(cur)
				cur = w
			}
		}
		if cur != "" {
			if out.Len() > 0 {
				out.WriteByte('\n')
			}
			out.WriteString(cur)
		}
	}
	return out.String()
}

// refreshViewport rebuilds viewport content from messages.
func (c *Chat) refreshViewport() {
	// Indent matches the input area padding (2) so labels + text align with the prompt
	indent := "  "
	contentWidth := c.Width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}

	var lines []string
	for _, msg := range c.Messages {
		var label, text string
		if msg.Role == "human" {
			label = indent + c.HumanLabel.Render("You")
			text = indent + c.HumanMsg.Render(wrapText(msg.Content, contentWidth))
		} else {
			label = indent + c.AgentLabel.Render("Agent")
			text = indent + c.AgentMsg.Render(wrapText(msg.Content, contentWidth))
		}
		lines = append(lines, label)
		lines = append(lines, text)
		lines = append(lines, "")
	}

	if c.Streaming {
		lines = append(lines, indent+c.AgentLabel.Render("Agent")+" "+c.Spinner.View()+" "+c.currentPhrase())
		if c.StreamBuf != "" {
			lines = append(lines, indent+c.StreamStyle.Render(wrapText(c.StreamBuf, contentWidth)))
		}
	}

	c.Viewport.SetContent(strings.Join(lines, "\n"))
	c.Viewport.GotoBottom()
}

// RefreshStreaming updates the viewport with current streaming state.
func (c *Chat) RefreshStreaming() {
	c.refreshViewport()
}

// ViewHeader renders the chat header.
func (c *Chat) ViewHeader(agentName, threadID string, headerStyle, titleStyle, infoStyle lipgloss.Style) string {
	return headerStyle.Width(c.Width).Render(titleStyle.Render(agentName))
}

// ViewInput renders the input area with bordered styles.
func (c *Chat) ViewInput(blurredStyle, focusedStyle lipgloss.Style) string {
	style := blurredStyle
	if c.InputFocused {
		style = focusedStyle
	}

	if c.Streaming {
		muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B"))
		return style.Width(c.Width - 2).Render(muted.Render(c.currentPhrase()))
	}

	var parts []string

	// Autocomplete dropdown appears above the input
	if dropdown := c.Completer.View(c.Width - 4); dropdown != "" {
		parts = append(parts, dropdown)
	}

	parts = append(parts, c.Input.View())

	return style.Width(c.Width - 2).Render(strings.Join(parts, "\n"))
}

// View renders the entire chat view.
func (c *Chat) View() string {
	return c.Viewport.View()
}

// ClearInput empties the text input.
func (c *Chat) ClearInput() {
	c.Input.Reset()
	c.Completer.Update("")
}

// GetInput returns the current input text.
func (c *Chat) GetInput() string {
	return strings.TrimSpace(c.Input.Value())
}
