package main

import (
	"context"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/brittonhayes/notion-agent-sdk-go/cmd/notion-ai/views"
)

type viewState int

const (
	viewLoading viewState = iota
	viewAgentSelect
	viewChat
	viewThreadList
)

// appModel is the root bubbletea model.
type appModel struct {
	// SDK
	client *notionagents.Client
	ctx    context.Context
	cancel context.CancelFunc

	// State
	currentView viewState
	width       int
	height      int
	err         string

	// Config
	agentFlag string // --agent flag value

	// Data
	agents      []notionagents.AgentData
	activeAgent *notionagents.Agent
	agentName   string
	threadID    string

	// Streaming
	streaming    bool
	streamCancel context.CancelFunc
	lastContent  string
	streamReader *notionagents.StreamReader

	// Sub-models
	loadSpinner spinner.Model
	agentSelect *views.AgentSelect
	chat        *views.Chat
	threadList  *views.ThreadList
	statusBar   *views.StatusBar

	// Rendering
	mdRenderer *markdownRenderer
	styles     styles
	theme      themeColors
}

func newAppModel(client *notionagents.Client, ctx context.Context, cancel context.CancelFunc, agentFlag string) appModel {
	theme := darkTheme()
	s := newStyles(theme)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = s.spinnerStyle

	return appModel{
		client:      client,
		ctx:         ctx,
		cancel:      cancel,
		currentView: viewLoading,
		agentFlag:   agentFlag,
		loadSpinner: sp,
		styles:      s,
		theme:       theme,
		mdRenderer:  newMarkdownRenderer(80),
		statusBar: &views.StatusBar{
			KeyStyle: s.statusKey,
			ValStyle: s.statusVal,
			ErrStyle: s.errorStyle,
			BarStyle: s.statusBar,
			CmdStyle: s.statusKey,
		},
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadSpinner.Tick,
		fetchAgentsCmd(m.ctx, m.client),
	)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.streamCancel != nil {
				m.streamCancel()
			}
			m.cancel()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mdRenderer = newMarkdownRenderer(msg.Width - 6)
		m.updateSizes()
		return m, nil

	case agentsLoadedMsg:
		m.agents = msg.agents
		m.err = ""
		// Auto-select agent based on --agent flag or personal agent
		if target := m.findAutoSelectAgent(); target != nil {
			return m.selectAgent(*target)
		}
		m.agentSelect = newAgentSelectView(&m, msg.agents)
		m.currentView = viewAgentSelect
		return m, nil

	case agentsErrorMsg:
		m.err = msg.err.Error()
		return m, nil

	case agentSelectedMsg:
		return m.selectAgent(msg.agent)

	case streamReaderReadyMsg:
		m.streamReader = msg.reader
		return m, readNextChunkCmd(msg.reader)

	case streamChunkMsg:
		return m.handleStreamChunk(msg)

	case streamDoneMsg:
		return m.finalizeStream(msg)

	case streamErrorMsg:
		m.streaming = false
		m.streamReader = nil
		if m.chat != nil {
			m.chat.Streaming = false
			m.chat.AppendMessage("agent", "Error: "+msg.err.Error())
		}
		m.statusBar.Streaming = false
		m.statusBar.Error = msg.err.Error()
		return m, nil

	case threadsLoadedMsg:
		m.err = ""
		m.threadList = newThreadListView(&m, msg.threads)
		m.currentView = viewThreadList
		return m, nil

	case threadsErrorMsg:
		m.err = msg.err.Error()
		return m, nil

	case messagesLoadedMsg:
		return m.loadThreadMessages(msg)

	case messagesErrorMsg:
		m.err = msg.err.Error()
		return m, nil
	}

	// Dispatch to current view
	switch m.currentView {
	case viewLoading:
		var cmd tea.Cmd
		m.loadSpinner, cmd = m.loadSpinner.Update(msg)
		return m, cmd

	case viewAgentSelect:
		return m.updateAgentSelect(msg)

	case viewChat:
		return m.updateChat(msg)

	case viewThreadList:
		return m.updateThreadList(msg)
	}

	return m, nil
}

func (m appModel) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.currentView {
	case viewLoading:
		return m.viewLoading()
	case viewAgentSelect:
		return m.viewAgentSelect()
	case viewChat:
		return m.viewChat()
	case viewThreadList:
		return m.viewThreadList()
	}

	return ""
}

// --- View renderers ---

func (m *appModel) viewLoading() string {
	msg := m.loadSpinner.View() + " Loading agents..."
	if m.err != "" {
		msg += "\n\n" + m.styles.errorStyle.Render("Error: "+m.err)
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}

func (m *appModel) viewAgentSelect() string {
	if m.agentSelect == nil {
		return ""
	}
	return m.agentSelect.View()
}

func (m *appModel) viewChat() string {
	if m.chat == nil {
		return ""
	}

	body := m.chat.View()
	input := m.chat.ViewInput(m.styles.inputArea, m.styles.inputAreaFocused)
	m.statusBar.Width = m.width
	status := m.statusBar.View()

	return body + "\n" + input + "\n" + status
}

func (m *appModel) viewThreadList() string {
	if m.threadList == nil {
		return ""
	}
	return m.threadList.View()
}

// --- Update dispatchers ---

func (m appModel) updateAgentSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.agentSelect == nil {
		return m, nil
	}

	cmd := m.agentSelect.Update(msg)
	if m.agentSelect.Selected != nil {
		selected := *m.agentSelect.Selected
		m.agentSelect.Selected = nil
		return m, func() tea.Msg { return agentSelectedMsg{agent: selected} }
	}
	return m, cmd
}

func (m appModel) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.chat == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "esc" && m.streaming:
			if m.streamCancel != nil {
				m.streamCancel()
			}
			m.streaming = false
			m.chat.Streaming = false
			m.chat.ResetPhrase()
			m.statusBar.Streaming = false
			if m.chat.StreamBuf != "" {
				m.chat.AppendMessage("agent", m.chat.StreamBuf)
				m.chat.StreamBuf = ""
			}
			return m, nil

		case msg.String() == "tab":
			// Tab completes slash command when completer is active
			if m.chat.Completer.Active {
				if sel := m.chat.Completer.Selected(); sel != nil {
					m.chat.Input.SetValue(sel.Name)
					m.chat.Input.CursorEnd()
					m.chat.Completer.Update(sel.Name)
				}
				return m, nil
			}
			m.chat.InputFocused = !m.chat.InputFocused
			if m.chat.InputFocused {
				m.chat.Input.Focus()
			} else {
				m.chat.Input.Blur()
			}
			return m, nil

		case (msg.String() == "up" || msg.String() == "k") && m.chat.Completer.Active && m.chat.InputFocused:
			m.chat.Completer.MoveUp()
			return m, nil

		case (msg.String() == "down" || msg.String() == "j") && m.chat.Completer.Active && m.chat.InputFocused:
			m.chat.Completer.MoveDown()
			return m, nil

		case msg.String() == "enter" && m.chat.InputFocused && !m.streaming:
			input := m.chat.GetInput()
			if input == "" {
				return m, nil
			}
			m.chat.ClearInput()
			return m.handleChatInput(input)
		}
	}

	cmd := m.chat.Update(msg)
	return m, cmd
}

func (m appModel) updateThreadList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.threadList == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.currentView = viewChat
			return m, nil
		case "n":
			m.threadID = ""
			m.statusBar.ThreadID = ""
			if m.chat != nil {
				m.chat.Messages = nil
				m.chat.StreamBuf = ""
				m.chat.RefreshStreaming()
			}
			m.currentView = viewChat
			return m, nil
		}
	}

	cmd := m.threadList.Update(msg)
	if m.threadList.Selected != nil {
		selected := *m.threadList.Selected
		m.threadList.Selected = nil
		m.threadID = selected.ID
		m.statusBar.ThreadID = selected.ID
		m.currentView = viewChat
		return m, fetchMessagesCmd(m.ctx, m.activeAgent, selected.ID)
	}
	return m, cmd
}

// --- Helpers ---

func (m *appModel) selectAgent(agent notionagents.AgentData) (tea.Model, tea.Cmd) {
	a := m.client.Agents.Agent(agent.ID)
	a.Name = agent.Name
	a.Instruction = agent.Instruction
	m.activeAgent = a
	m.agentName = agent.Name
	m.threadID = ""
	m.err = ""

	c := views.NewChat(m.width, m.height)
	c.HumanLabel = m.styles.humanLabel
	c.AgentLabel = m.styles.agentLabel
	c.HumanMsg = m.styles.humanMsg
	c.AgentMsg = m.styles.agentMsg
	c.DividerStyle = m.styles.divider
	c.StreamStyle = m.styles.streamingText
	c.Spinner.Style = m.styles.spinnerStyle
	c.Completer.CmdStyle = lipgloss.NewStyle().Foreground(m.theme.accent).Bold(true)
	c.Completer.DescStyle = lipgloss.NewStyle().Foreground(m.theme.muted)
	c.Completer.SelStyle = lipgloss.NewStyle().Background(m.theme.surface).Foreground(m.theme.accent)
	m.chat = &c

	m.statusBar.AgentName = agent.Name
	m.statusBar.ThreadID = ""
	m.statusBar.Error = ""
	m.statusBar.Streaming = false

	m.currentView = viewChat
	m.updateSizes()
	return *m, nil
}

func (m *appModel) handleChatInput(input string) (tea.Model, tea.Cmd) {
	// Slash commands
	switch {
	case input == "/quit":
		m.cancel()
		return *m, tea.Quit
	case input == "/new":
		m.threadID = ""
		m.chat.Messages = nil
		m.chat.StreamBuf = ""
		m.chat.RefreshStreaming()
		m.statusBar.ThreadID = ""
		return *m, nil
	case input == "/threads":
		if m.activeAgent != nil {
			return *m, fetchThreadsCmd(m.ctx, m.activeAgent)
		}
		return *m, nil
	case input == "/agents":
		m.agentSelect = newAgentSelectView(m, m.agents)
		m.currentView = viewAgentSelect
		return *m, nil
	}

	// Normal message
	m.chat.AppendMessage("human", input)
	m.streaming = true
	m.chat.Streaming = true
	m.chat.StreamBuf = ""
	m.lastContent = ""
	m.statusBar.Streaming = true
	m.statusBar.Error = ""

	streamCtx, streamCancel := context.WithCancel(m.ctx)
	m.streamCancel = streamCancel

	m.chat.ResetPhrase()

	return *m, tea.Batch(
		m.chat.Spinner.Tick,
		views.PhraseTickCmd(),
		startStreamCmd(streamCtx, m.activeAgent, input, m.threadID),
	)
}

func (m appModel) handleStreamChunk(msg streamChunkMsg) (tea.Model, tea.Cmd) {
	chunk := msg.chunk
	if chunk.Type == "message" && chunk.Role == "agent" {
		// Compute delta (content is cumulative)
		if len(chunk.Content) > len(m.lastContent) {
			delta := chunk.Content[len(m.lastContent):]
			m.chat.StreamBuf += delta
			m.lastContent = chunk.Content
		}
		m.chat.RefreshStreaming()
	}
	if chunk.Type == "started" && chunk.ThreadID != "" {
		m.threadID = chunk.ThreadID
		m.statusBar.ThreadID = chunk.ThreadID
	}
	return m, readNextChunkCmd(m.streamReader)
}

func (m appModel) finalizeStream(msg streamDoneMsg) (tea.Model, tea.Cmd) {
	m.streaming = false
	m.chat.Streaming = false
	m.chat.ResetPhrase()
	m.statusBar.Streaming = false
	m.streamReader = nil

	if msg.info != nil {
		m.threadID = msg.info.ThreadID
		m.statusBar.ThreadID = msg.info.ThreadID
	}

	// Render final markdown
	content := m.chat.StreamBuf
	if content != "" {
		rendered := m.mdRenderer.render(content)
		m.chat.AppendMessage("agent", rendered)
	}
	m.chat.StreamBuf = ""
	m.lastContent = ""

	return m, nil
}

func (m *appModel) loadThreadMessages(msg messagesLoadedMsg) (tea.Model, tea.Cmd) {
	if m.chat == nil {
		return *m, nil
	}
	m.chat.Messages = nil
	for _, item := range msg.messages {
		role := item.Role
		if role == "user" {
			role = "human"
		}
		content := item.Content
		if role == "agent" {
			content = m.mdRenderer.render(content)
		}
		m.chat.Messages = append(m.chat.Messages, views.ChatMessage{
			Role:    role,
			Content: content,
		})
	}
	m.chat.RefreshStreaming()
	return *m, nil
}

// findAutoSelectAgent returns the agent to auto-select on startup:
// --agent flag match > personal agent > nil (fall back to select screen).
func (m *appModel) findAutoSelectAgent() *notionagents.AgentData {
	for i := range m.agents {
		if m.agentFlag != "" && m.agents[i].ID == m.agentFlag {
			return &m.agents[i]
		}
	}
	if m.agentFlag == "" {
		for i := range m.agents {
			if notionagents.IsPersonalAgent(m.agents[i].ID) {
				return &m.agents[i]
			}
		}
	}
	return nil
}

func (m *appModel) updateSizes() {
	if m.agentSelect != nil {
		m.agentSelect.SetSize(m.width, m.height)
	}
	if m.chat != nil {
		m.chat.SetSize(m.width, m.height)
	}
	if m.threadList != nil {
		m.threadList.SetSize(m.width, m.height)
	}
	m.statusBar.Width = m.width
}

func newAgentSelectView(m *appModel, agents []notionagents.AgentData) *views.AgentSelect {
	as := views.NewAgentSelect(agents, m.width, m.height, m.theme.accent, m.theme.muted)
	return &as
}

func newThreadListView(m *appModel, threads []notionagents.ThreadListItem) *views.ThreadList {
	tl := views.NewThreadList(threads, m.width, m.height, m.theme.accent, m.theme.muted)
	return &tl
}
