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
	// Handle subcommands before flag parsing.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			if err := runLogin(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "logout":
			if err := runLogout(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "status":
			if err := runStatus(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	agentFlag := flag.String("agent", "", "agent ID to connect to on startup")
	flag.Parse()

	token, err := resolveToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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
