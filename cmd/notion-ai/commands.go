package main

import (
	"context"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

// fetchAgentsCmd loads all agents from the API.
func fetchAgentsCmd(ctx context.Context, client *notionagents.Client) tea.Cmd {
	return func() tea.Msg {
		agents, err := notionagents.CollectAgents(ctx, client, nil)
		if err != nil {
			return agentsErrorMsg{err: err}
		}
		return agentsLoadedMsg{agents: agents}
	}
}

// fetchThreadsCmd loads threads for the given agent.
func fetchThreadsCmd(ctx context.Context, agent *notionagents.Agent) tea.Cmd {
	return func() tea.Msg {
		threads, err := notionagents.CollectThreads(ctx, agent, nil)
		if err != nil {
			return threadsErrorMsg{err: err}
		}
		return threadsLoadedMsg{threads: threads}
	}
}

// fetchMessagesCmd loads messages for a thread.
func fetchMessagesCmd(ctx context.Context, agent *notionagents.Agent, threadID string) tea.Cmd {
	return func() tea.Msg {
		thread := agent.Thread(threadID)
		msgs, err := notionagents.CollectMessages(ctx, thread, nil)
		if err != nil {
			return messagesErrorMsg{err: err}
		}
		return messagesLoadedMsg{threadID: threadID, messages: msgs}
	}
}
