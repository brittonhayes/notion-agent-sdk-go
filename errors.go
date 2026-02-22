package notionagents

import "fmt"

// NotionAgentsError is the base error type for SDK errors.
type NotionAgentsError struct {
	Msg  string
	Code string
}

func (e *NotionAgentsError) Error() string {
	return fmt.Sprintf("notion agents error [%s]: %s", e.Code, e.Msg)
}

// AgentNotFoundError is returned when an agent cannot be found.
type AgentNotFoundError struct {
	AgentID string
}

func (e *AgentNotFoundError) Error() string {
	return fmt.Sprintf("agent not found: %s", e.AgentID)
}

// ThreadNotFoundError is returned when a thread cannot be found.
type ThreadNotFoundError struct {
	ThreadID string
}

func (e *ThreadNotFoundError) Error() string {
	return fmt.Sprintf("thread not found: %s", e.ThreadID)
}

// PollingTimeoutError is returned when thread polling exceeds max attempts.
type PollingTimeoutError struct {
	Attempts int
}

func (e *PollingTimeoutError) Error() string {
	return fmt.Sprintf("polling timed out after %d attempts", e.Attempts)
}

// StreamError is returned for streaming-related errors.
type StreamError struct {
	Msg  string
	Code string
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("stream error [%s]: %s", e.Code, e.Msg)
}
