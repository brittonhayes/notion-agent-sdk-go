package main

import (
	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
)

// agentsLoadedMsg is sent when the agent list has been fetched.
type agentsLoadedMsg struct {
	agents []notionagents.AgentData
}

// agentsErrorMsg is sent when agent fetching fails.
type agentsErrorMsg struct {
	err error
}

// agentSelectedMsg is sent when the user picks an agent.
type agentSelectedMsg struct {
	agent notionagents.AgentData
}

// streamReaderReadyMsg carries a new StreamReader to consume.
type streamReaderReadyMsg struct {
	reader *notionagents.StreamReader
}

// streamChunkMsg carries a single streaming chunk with cumulative content.
type streamChunkMsg struct {
	chunk notionagents.StreamChunk
}

// streamDoneMsg signals the stream has completed.
type streamDoneMsg struct {
	info *notionagents.ThreadInfo
}

// streamErrorMsg is sent when a streaming error occurs.
type streamErrorMsg struct {
	err error
}

// threadsLoadedMsg carries the fetched thread list.
type threadsLoadedMsg struct {
	threads []notionagents.ThreadListItem
}

// threadsErrorMsg is sent when thread fetching fails.
type threadsErrorMsg struct {
	err error
}

// messagesLoadedMsg carries thread messages for display.
type messagesLoadedMsg struct {
	threadID string
	messages []notionagents.ThreadMessageItem
}

// messagesErrorMsg is sent when message fetching fails.
type messagesErrorMsg struct {
	err error
}

// errMsg is a generic error message.
type errMsg struct {
	err error
}
