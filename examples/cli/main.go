package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
)

func main() {
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

	agent, err := selectAgent(ctx, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nChatting with %s. Commands: /switch, /threads, /exit\n", agent.Name)

	var threadID string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch {
		case input == "/exit":
			fmt.Println("Goodbye!")
			return
		case input == "/switch":
			agent, err = selectAgent(ctx, client)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			threadID = ""
			fmt.Printf("\nSwitched to %s\n", agent.Name)
			continue
		case input == "/threads":
			listThreads(ctx, agent)
			continue
		}

		// Stream chat response
		fmt.Print("\nAgent: ")
		reader, err := agent.Stream(ctx, notionagents.ChatStreamParams{
			Message:  input,
			ThreadID: threadID,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			continue
		}

		var lastContent string
		for {
			chunk, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nStream error: %v\n", err)
				break
			}

			if chunk.Type == "message" && chunk.Role == "agent" {
				// Print incremental content (delta since last)
				if len(chunk.Content) > len(lastContent) {
					fmt.Print(chunk.Content[len(lastContent):])
					lastContent = chunk.Content
				}
			}
		}
		fmt.Println()

		if info := reader.ThreadInfo(); info != nil {
			threadID = info.ThreadID
		}
		reader.Close()
	}
}

func selectAgent(ctx context.Context, client *notionagents.Client) (*notionagents.Agent, error) {
	agents, err := notionagents.CollectAgents(ctx, client, nil)
	if err != nil {
		return nil, err
	}
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents found")
	}

	fmt.Println("\nAvailable agents:")
	for i, a := range agents {
		desc := ""
		if a.Description != nil {
			desc = " - " + *a.Description
		}
		fmt.Printf("  %d. %s%s\n", i+1, a.Name, desc)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nSelect agent (number): ")
		if !scanner.Scan() {
			return nil, fmt.Errorf("no input")
		}
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(agents) {
			fmt.Printf("Please enter a number between 1 and %d\n", len(agents))
			continue
		}
		selected := agents[n-1]
		agent := client.Agents.Agent(selected.ID)
		agent.Name = selected.Name
		agent.Instruction = selected.Instruction
		return agent, nil
	}
}

func listThreads(ctx context.Context, agent *notionagents.Agent) {
	resp, err := agent.ListThreads(ctx, &notionagents.ThreadListParams{PageSize: 10})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing threads: %v\n", err)
		return
	}
	if len(resp.Results) == 0 {
		fmt.Println("No threads found.")
		return
	}
	fmt.Println("\nRecent threads:")
	for _, t := range resp.Results {
		fmt.Printf("  [%s] %s (%s)\n", t.Status, t.Title, t.ID)
	}
}
