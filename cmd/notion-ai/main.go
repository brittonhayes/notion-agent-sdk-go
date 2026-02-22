package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	agentFlag := flag.String("agent", "", "agent ID to connect to on startup")
	flag.Parse()

	token := os.Getenv("NOTION_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "NOTION_API_TOKEN environment variable is required")
		os.Exit(1)
	}

	opts := notionagents.ClientOptions{Auth: token}
	if baseURL := os.Getenv("NOTION_BASE_URL"); baseURL != "" {
		opts.BaseURL = baseURL
	}
	client := notionagents.NewClient(opts)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	model := newAppModel(client, ctx, cancel, *agentFlag)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
