# Notion Agents SDK for Go

A Go client for interacting with Notion Agents via the Notion Agents API.

[![Go Reference](https://pkg.go.dev/badge/github.com/brittonhayes/notion-agent-sdk-go.svg)](https://pkg.go.dev/github.com/brittonhayes/notion-agent-sdk-go)

> **Disclaimer**: This is an **unofficial**, community-maintained SDK and is not affiliated with or endorsed by Notion. It is maintained on a best-effort basis.

Status: **Alpha**

> **Notion Agents API reference**: [developers.notion.com/reference/internal/list-agents](https://developers.notion.com/reference/internal/list-agents)

This SDK is a lightweight, idiomatic Go client built entirely on the standard library (`net/http`, `encoding/json`). Zero external dependencies.

## Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Quickstart (async)](#quickstart-async)
- [Quickstart (streaming)](#quickstart-streaming)
- [Concepts](#concepts)
  - [Agents: custom vs personal](#agents-custom-vs-personal)
  - [Threads and messages](#threads-and-messages)
- [Verbose output: content_parts](#verbose-output-content_parts)
- [API reference](#api-reference)
- [Errors](#errors)
- [Examples](#examples)

## Requirements

- **Go 1.23+** (uses `iter` package for pagination iterators)
- A Notion API token (internal integration secret or OAuth access token)
  - Notion getting started guide: [developers.notion.com](https://developers.notion.com/guides/get-started/getting-started)
- Alpha access & Custom Agents access

## Installation

```bash
go get github.com/brittonhayes/notion-agent-sdk-go
```

### Environment variables

Most examples assume your token is set as:

```bash
export NOTION_API_TOKEN="secret_..."
```

## Quickstart (async)

The async flow returns immediately with a `thread_id`, then you poll for completion and fetch messages separately.

```go
package main

import (
    "context"
    "fmt"
    "log"

    notionagents "github.com/brittonhayes/notion-agent-sdk-go"
)

func main() {
    client := notionagents.NewClient(notionagents.ClientOptions{
        Auth: "secret_...",
    })
    ctx := context.Background()

    // Pick an agent
    agents, err := client.Agents.List(ctx, &notionagents.AgentListParams{PageSize: 10})
    if err != nil {
        log.Fatal(err)
    }
    agent := client.Agents.Agent(agents.Results[0].ID)

    // Start a conversation (returns quickly with pending status)
    invocation, err := agent.Chat(ctx, notionagents.ChatParams{Message: "Hello!"})
    if err != nil {
        log.Fatal(err)
    }

    // Poll until the thread is completed or failed
    thread := agent.Thread(invocation.ThreadID)
    threadInfo, err := thread.Poll(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Thread status: %s\n", threadInfo.Status)

    // Fetch messages
    verbose := true
    messages, err := thread.ListMessages(ctx, &notionagents.ThreadMessageListParams{
        PageSize: 50,
        Verbose:  &verbose,
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, msg := range messages.Results {
        fmt.Printf("%s: %s\n", msg.Role, msg.Content)
    }
}
```

## Quickstart (streaming)

The streaming flow uses newline-delimited JSON (NDJSON) under the hood. The SDK exposes it via a `StreamReader` iterator or channel-based API.

### StreamReader (iterator)

```go
agent := client.Agents.Personal()

reader, err := agent.Stream(ctx, notionagents.ChatStreamParams{
    Message: "Summarize my week",
})
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

for {
    chunk, err := reader.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    if chunk.Type == "message" && chunk.Role == "agent" {
        fmt.Print(notionagents.StripLangTags(chunk.Content))
    }
}

// Access accumulated thread info after stream completes
info := reader.ThreadInfo()
fmt.Printf("\nThread: %s (%d messages)\n", info.ThreadID, len(info.Messages))
```

### Channels

```go
chunks, info, errc := agent.ChatStream(ctx, notionagents.ChatStreamParams{
    Message: "Hello",
})

for chunk := range chunks {
    if chunk.Type == "message" && chunk.Role == "agent" {
        fmt.Print(chunk.Content)
    }
}

// Check for errors
if err := <-errc; err != nil {
    log.Fatal(err)
}

// Get final thread info
if threadInfo := <-info; threadInfo != nil {
    fmt.Printf("\nThread: %s\n", threadInfo.ThreadID)
}
```

## Concepts

### Agents: custom vs personal

- **Custom agents** are user-created agents in a workspace. They appear in `client.Agents.List()`.
- The **personal agent** is Notion AI, addressed by a reserved UUID:
  - Use `client.Agents.Personal()`, or `client.Agents.Agent(notionagents.PersonalAgentID)`.
  - Note: internal integrations can't access the personal agent, since internal integrations are generally owned by workspace owners rather than any specific user. The personal agent won't appear in `client.Agents.List()`, and requests targeting it will fail with `object_not_found`.

### Threads and messages

A chat happens inside a **thread**:

- `agent.Chat()` creates/continues a thread and returns `ChatInvocationResponse{ThreadID, Status: "pending"}`.
- `thread.Poll()` checks the thread until it is `completed` or `failed`, using exponential backoff with jitter.
- `thread.ListMessages()` fetches messages from `GET /threads/:thread_id/messages`.

Polling returns **thread metadata** (status/title/creator/version). Messages are fetched separately via `ListMessages()`.

## Verbose output: `content_parts`

When available, agent messages include a structured representation in `ContentParts`.

- **Streaming**: `agent.Stream()` includes `content_parts` by default. Pass `Verbose: boolPtr(false)` to omit.
- **Message listing**: `thread.ListMessages(&ThreadMessageListParams{Verbose: boolPtr(true)})` includes `content_parts`.

Part types you may encounter:

- `text` - model text output
- `thinking` - model reasoning
- `tool_call` - tool invocation with optional results
- `follow_ups` - suggested follow-up actions
- `custom_agent_template_picker` - non-text UI state

If you don't need this level of detail, set `Verbose` to `false` and use `Content` only.

## API reference

Full documentation is available via `go doc`:

```bash
go doc github.com/brittonhayes/notion-agent-sdk-go
```

### Exports

| Export | Description |
|--------|-------------|
| `NewClient` | Create a new API client |
| `PersonalAgentID` | Reserved UUID for the personal agent |
| `StripLangTags` | Remove `<lang ...>` tags from agent output |
| `IsPersonalAgent` | Check if an ID is the personal agent |
| `IterAgents` / `CollectAgents` | Auto-paginating agent iterators |
| `IterThreads` / `CollectThreads` | Auto-paginating thread iterators |
| `IterMessages` / `CollectMessages` | Auto-paginating message iterators |

### Client

```go
client := notionagents.NewClient(notionagents.ClientOptions{
    Auth:          "secret_...",    // required
    BaseURL:       "",              // defaults to "https://api.notion.com"
    NotionVersion: "",              // defaults to "2025-09-03"
    HTTPClient:    nil,             // defaults to http.DefaultClient
})
```

### client.Agents (AgentOperations)

```go
// List all accessible agents
resp, err := client.Agents.List(ctx, &notionagents.AgentListParams{
    Name:        "",    // filter by name
    PageSize:    10,
    StartCursor: "",
})

// Get an agent handle by ID
agent := client.Agents.Agent(agentID)

// Get the personal agent handle
personal := client.Agents.Personal()
```

### Agent

```go
// Async chat
resp, err := agent.Chat(ctx, notionagents.ChatParams{
    Message:     "Hello!",
    ThreadID:    "",                        // optional: continue existing thread
    Attachments: []ChatAttachmentInput{},   // optional: file attachments
})

// Streaming chat (iterator)
reader, err := agent.Stream(ctx, notionagents.ChatStreamParams{
    Message:  "Hello!",
    ThreadID: "",
    Verbose:  nil,  // default true
    OnMessage: func(msg notionagents.StreamMessage) {
        // called on each message upsert
    },
})

// Streaming chat (channels)
chunks, info, errc := agent.ChatStream(ctx, params)

// Thread operations
thread := agent.Thread(threadID)
item, err := agent.GetThread(ctx, threadID)
item, err := agent.PollThread(ctx, threadID, nil)
resp, err := agent.ListThreads(ctx, &notionagents.ThreadListParams{...})
```

### Thread

```go
// Get thread metadata
item, err := thread.Get(ctx)

// Poll until completed/failed with exponential backoff
item, err := thread.Poll(ctx, &notionagents.PollThreadOptions{
    MaxAttempts:    60,     // default
    BaseDelayMs:   1000,   // default
    MaxDelayMs:    10000,  // default
    InitialDelayMs: 1000,  // default
    OnPending:      func(t notionagents.ThreadListItem, attempt int) {},
    OnThreadNotFound: func(attempt int) {},
})

// List messages
resp, err := thread.ListMessages(ctx, &notionagents.ThreadMessageListParams{
    Verbose:     nil,       // default true
    Role:        "agent",   // "user" or "agent"
    PageSize:    50,
    StartCursor: "",
})
```

### Pagination helpers

The SDK provides Go 1.23 iterators that automatically handle cursor-based pagination:

```go
// Iterate over all agents
for agent, err := range notionagents.IterAgents(ctx, client, nil) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(agent.Name)
}

// Or collect all into a slice
agents, err := notionagents.CollectAgents(ctx, client, nil)
threads, err := notionagents.CollectThreads(ctx, agent, nil)
messages, err := notionagents.CollectMessages(ctx, thread, nil)
```

## Errors

The SDK provides typed errors for common scenarios:

```go
var agentErr *notionagents.AgentNotFoundError
var threadErr *notionagents.ThreadNotFoundError
var pollErr *notionagents.PollingTimeoutError
var streamErr *notionagents.StreamError

if errors.As(err, &agentErr) {
    fmt.Printf("Agent not found: %s\n", agentErr.AgentID)
}
```

| Error | Description |
|-------|-------------|
| `NotionAgentsError` | Base error with Code and Msg fields |
| `AgentNotFoundError` | Agent is missing or inaccessible |
| `ThreadNotFoundError` | Thread cannot be found |
| `PollingTimeoutError` | `Poll()` exceeded max attempts |
| `StreamError` | Streaming failure (HTTP error, malformed response, etc.) |

Streaming can also produce error chunks (`chunk.Type == "error"`) with a machine-readable `Code` and `Message`; handle both patterns.

## Examples

See [`examples/cli/`](examples/cli/) for a complete interactive CLI tool that demonstrates:

- Client initialization from environment variables
- Listing and selecting agents
- Streaming chat with real-time output
- Thread continuity across messages
- Signal handling with `context`

To run:

```bash
export NOTION_API_TOKEN="secret_..."
cd examples/cli
go run .
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
