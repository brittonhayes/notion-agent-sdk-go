package notionagents

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestThreadGet(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "/agents/agent-1/threads") {
			t.Errorf("path = %s, want to contain /agents/agent-1/threads", req.URL.Path)
		}
		if req.URL.Query().Get("id") != "thread-1" {
			t.Errorf("id param = %q, want %q", req.URL.Query().Get("id"), "thread-1")
		}

		return jsonResponse(200, ThreadListResponse{
			Object: "list",
			Results: []ThreadListItem{
				{Object: "thread", ID: "thread-1", Title: "My Thread", Status: ThreadStatusCompleted},
			},
		}), nil
	})

	thread := &Thread{ThreadID: "thread-1", AgentID: "agent-1", client: c}
	item, err := thread.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "thread-1" {
		t.Errorf("ID = %q, want %q", item.ID, "thread-1")
	}
	if item.Status != ThreadStatusCompleted {
		t.Errorf("Status = %q, want %q", item.Status, ThreadStatusCompleted)
	}
}

func TestThreadGetNotFound(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, ThreadListResponse{
			Object:  "list",
			Results: []ThreadListItem{},
		}), nil
	})

	thread := &Thread{ThreadID: "nonexistent", AgentID: "agent-1", client: c}
	_, err := thread.Get(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	tnf, ok := err.(*ThreadNotFoundError)
	if !ok {
		t.Fatalf("expected *ThreadNotFoundError, got %T", err)
	}
	if tnf.ThreadID != "nonexistent" {
		t.Errorf("ThreadID = %q, want %q", tnf.ThreadID, "nonexistent")
	}
}

func TestThreadListMessages(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "/threads/thread-1/messages") {
			t.Errorf("path = %s, want to contain /threads/thread-1/messages", req.URL.Path)
		}

		return jsonResponse(200, ThreadMessageListResponse{
			Object: "list",
			Type:   "thread_message",
			Results: []ThreadMessageItem{
				{Object: "thread_message", ID: "msg-1", Role: "human", Content: "Hello"},
				{Object: "thread_message", ID: "msg-2", Role: "assistant", Content: "Hi!"},
			},
			HasMore: false,
		}), nil
	})

	thread := &Thread{ThreadID: "thread-1", AgentID: "agent-1", client: c}
	resp, err := thread.ListMessages(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("results len = %d, want 2", len(resp.Results))
	}
	if resp.Results[0].Role != "human" {
		t.Errorf("Role = %q, want %q", resp.Results[0].Role, "human")
	}
}

func TestThreadListMessagesWithParams(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		q := req.URL.Query()
		if q.Get("verbose") != "true" {
			t.Errorf("verbose = %q, want %q", q.Get("verbose"), "true")
		}
		if q.Get("role") != "assistant" {
			t.Errorf("role = %q, want %q", q.Get("role"), "assistant")
		}
		if q.Get("page_size") != "5" {
			t.Errorf("page_size = %q, want %q", q.Get("page_size"), "5")
		}

		return jsonResponse(200, ThreadMessageListResponse{
			Object:  "list",
			Results: []ThreadMessageItem{},
		}), nil
	})

	thread := &Thread{ThreadID: "thread-1", AgentID: "agent-1", client: c}
	verbose := true
	_, err := thread.ListMessages(context.Background(), &ThreadMessageListParams{
		Verbose:  &verbose,
		Role:     "assistant",
		PageSize: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestThreadPollDelegatesToAgent(t *testing.T) {
	calls := 0
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(200, ThreadListResponse{
			Object: "list",
			Results: []ThreadListItem{
				{Object: "thread", ID: "t-1", Status: ThreadStatusCompleted},
			},
		}), nil
	})

	thread := &Thread{ThreadID: "t-1", AgentID: "agent-1", client: c}
	item, err := thread.Poll(context.Background(), &PollThreadOptions{
		MaxAttempts:    1,
		InitialDelayMs: 1,
		BaseDelayMs:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != ThreadStatusCompleted {
		t.Errorf("Status = %q, want %q", item.Status, ThreadStatusCompleted)
	}
}
