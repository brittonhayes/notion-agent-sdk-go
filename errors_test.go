package notionagents

import (
	"errors"
	"fmt"
	"testing"
)

func TestNotionAgentsErrorMessage(t *testing.T) {
	err := &NotionAgentsError{Msg: "something broke", Code: "internal_error"}
	want := "notion agents error [internal_error]: something broke"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAgentNotFoundErrorMessage(t *testing.T) {
	err := &AgentNotFoundError{AgentID: "abc-123"}
	want := "agent not found: abc-123"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestThreadNotFoundErrorMessage(t *testing.T) {
	err := &ThreadNotFoundError{ThreadID: "thr-456"}
	want := "thread not found: thr-456"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestPollingTimeoutErrorMessage(t *testing.T) {
	err := &PollingTimeoutError{Attempts: 60}
	want := "polling timed out after 60 attempts"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestStreamErrorMessage(t *testing.T) {
	err := &StreamError{Msg: "connection lost", Code: "stream_error"}
	want := "stream error [stream_error]: connection lost"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestErrorTypeAssertions(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"NotionAgentsError", &NotionAgentsError{Msg: "test", Code: "test"}},
		{"AgentNotFoundError", &AgentNotFoundError{AgentID: "a"}},
		{"ThreadNotFoundError", &ThreadNotFoundError{ThreadID: "t"}},
		{"PollingTimeoutError", &PollingTimeoutError{Attempts: 1}},
		{"StreamError", &StreamError{Msg: "test", Code: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify it implements error interface
			var e error = tt.err
			if e.Error() == "" {
				t.Error("Error() should not return empty string")
			}
		})
	}
}

func TestErrorTypeAs(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &AgentNotFoundError{AgentID: "a-1"})

	var anf *AgentNotFoundError
	if !errors.As(err, &anf) {
		t.Error("errors.As should find *AgentNotFoundError in wrapped error")
	}
	if anf.AgentID != "a-1" {
		t.Errorf("AgentID = %q, want %q", anf.AgentID, "a-1")
	}
}

func TestIsThreadNotFound(t *testing.T) {
	if !isThreadNotFound(&ThreadNotFoundError{ThreadID: "t"}) {
		t.Error("isThreadNotFound should return true for ThreadNotFoundError")
	}
	if isThreadNotFound(&NotionAgentsError{Msg: "other"}) {
		t.Error("isThreadNotFound should return false for other errors")
	}
}
