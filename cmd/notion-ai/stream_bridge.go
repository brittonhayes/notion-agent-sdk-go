package main

import (
	"context"
	"io"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

// startStreamCmd opens a streaming connection and returns the reader.
func startStreamCmd(ctx context.Context, agent *notionagents.Agent, message, threadID string) tea.Cmd {
	return func() tea.Msg {
		reader, err := agent.Stream(ctx, notionagents.ChatStreamParams{
			Message:  message,
			ThreadID: threadID,
		})
		if err != nil {
			return streamErrorMsg{err: err}
		}
		return streamReaderReadyMsg{reader: reader}
	}
}

// readNextChunkCmd reads one chunk then returns itself for the next.
func readNextChunkCmd(reader *notionagents.StreamReader) tea.Cmd {
	return func() tea.Msg {
		chunk, err := reader.Next()
		if err == io.EOF {
			info := reader.ThreadInfo()
			reader.Close()
			return streamDoneMsg{info: info}
		}
		if err != nil {
			reader.Close()
			return streamErrorMsg{err: err}
		}
		return streamChunkMsg{chunk: chunk}
	}
}
