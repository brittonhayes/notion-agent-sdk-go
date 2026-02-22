# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Go SDK for the Notion Agents API. Provides typed client, streaming, pagination, and error handling as a reusable library. Prioritizes simplicity and zero dependencies.

## Commands

```bash
go build ./...                    # compile (including examples)
go vet ./...                      # static analysis
go fmt ./...                      # format
go test ./...                     # run tests
go test -v ./...                  # run tests (verbose)
go test ./testutil/...            # test utilities only
```

## Environment

Set `NOTION_API_TOKEN` env var for examples. The SDK itself takes the token as a `ClientOptions.Auth` parameter.

## Architecture

Multi-file single package (`notionagents`), organized by domain:

- **`client.go`** — `Client` struct, `NewClient()`, HTTP helpers (`doRequest`, `doJSON`), error classification
- **`types.go`** — All request/response structs, constants (`PersonalAgentID`, `DefaultBaseURL`, `DefaultVersion`), type aliases (`ThreadStatus`)
- **`agents.go`** — `AgentOperations` struct with `List()`, `Agent()`, `Personal()` methods
- **`agent.go`** — `Agent` struct with `Chat()`, `Stream()`, `ChatStream()`, `PollThread()`, `ListThreads()`, `GetThread()`, `Thread()`
- **`stream.go`** — `StreamReader` (iterator pattern for NDJSON), chunk accumulation, `ChatStream()` channel API
- **`thread.go`** — `Thread` struct with `Get()`, `Poll()`, `ListMessages()`
- **`pagination.go`** — Go 1.23 `iter.Seq2` iterators: `IterAgents`, `IterThreads`, `IterMessages` + `Collect*` helpers
- **`helpers.go`** — Utility functions: `IsPersonalAgent()`, `StripLangTags()`
- **`errors.go`** — Typed errors: `NotionAgentsError`, `AgentNotFoundError`, `ThreadNotFoundError`, `PollingTimeoutError`, `StreamError`
- **`doc.go`** — Package-level documentation
- **`testutil/`** — Mock HTTP client and response factories for testing
- **`examples/cli/`** — Interactive CLI demonstrating streaming chat, thread management

## Key Patterns

- **Streaming deltas**: The Notion streaming API sends cumulative content in each chunk. `StreamReader` tracks message state via a map and order slice.
- **Dual streaming API**: Both `StreamReader` (iterator via `Next()`) and `ChatStream()` (channels) are available.
- **Async polling**: `PollThread` uses exponential backoff with jitter (1s base, 10s cap, 60 max attempts by default).
- **Go 1.23 iterators**: Pagination uses `iter.Seq2[T, error]` for auto-paginating iteration.
- **API versioning**: `Notion-Version` header defaults to `2025-09-03`, overridable via `ClientOptions`.

## Coding Conventions

- **Stdlib only** — no external dependencies. All imports are from Go's standard library.
- **Error types** — use typed errors from `errors.go`. Classify by domain (`AgentNotFoundError`, etc.), not generic strings.
- **Unexported helpers** — internal HTTP methods (`doRequest`, `doJSON`) and request body types are unexported.
- **Nil-safe options** — all `*Params` and `*Options` arguments accept nil for defaults.
