// Package notionagents provides a Go client for the Notion Agents API.
//
// This SDK allows you to list agents, create and manage chat threads, and
// stream real-time responses from Notion Agents. It is built entirely on the
// Go standard library with zero external dependencies.
//
// # Getting started
//
// Create a client with your Notion API token:
//
//	client := notionagents.NewClient(notionagents.ClientOptions{
//	    Auth: "secret_...",
//	})
//
// # Async chat
//
// Start a chat, poll for completion, then fetch messages:
//
//	resp, _ := client.Agents.Agent(agentID).Chat(ctx, notionagents.ChatParams{
//	    Message: "Hello!",
//	})
//	thread := client.Agents.Agent(agentID).Thread(resp.ThreadID)
//	thread.Poll(ctx, nil)
//	messages, _ := thread.ListMessages(ctx, nil)
//
// # Streaming chat
//
// Stream responses in real time using the iterator-style [StreamReader]:
//
//	reader, _ := agent.Stream(ctx, notionagents.ChatStreamParams{
//	    Message: "Summarize my week",
//	})
//	defer reader.Close()
//	for {
//	    chunk, err := reader.Next()
//	    if err == io.EOF {
//	        break
//	    }
//	    // handle chunk
//	}
//
// Or use the channel-based [Agent.ChatStream] for concurrent consumption.
//
// # Pagination
//
// Auto-paginating iterators use Go 1.23 [iter.Seq2]:
//
//	for agent, err := range notionagents.IterAgents(ctx, client, nil) {
//	    // ...
//	}
package notionagents
